package gcron

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/serf"
)

type Agent struct {
	config *Config

	logger hclog.Logger

	conf    *serf.Config
	eventCh chan serf.Event
	serf    *serf.Serf

	// shutdownCh is used for shutdowns
	shutdown     bool
	shutdownCh   chan struct{}
	shutdownLock sync.Mutex
}

func NewAgent(config *Config, options ...AgentOption) *Agent {
	agent := &Agent{
		config:     config,
		shutdownCh: make(chan struct{}),
	}

	for _, option := range options {
		option(agent)
	}

	return agent
}

func (a *Agent) Start() error {
	a.logger = hclog.New(&hclog.LoggerOptions{
		Name:  "agent",
		Level: hclog.LevelFromString("DEBUG"),
	})

	// Initialize rand with current time
	rand.Seed(time.Now().UnixNano())

	err := a.setupSerf()
	if err != nil {
		return fmt.Errorf("can not setup serf, %s", err)
	}

	return nil
}

func (a *Agent) setupSerf() error {
	config := a.config

	bindIP, bindPort, err := config.AddrParts(config.BindAddr)
	if err != nil {
		return err
	}

	serfConfig := serf.DefaultConfig()
	switch config.Profile {
	case "lan":
		serfConfig.MemberlistConfig = memberlist.DefaultLANConfig()
	case "wan":
		serfConfig.MemberlistConfig = memberlist.DefaultWANConfig()
	case "local":
		serfConfig.MemberlistConfig = memberlist.DefaultLocalConfig()
	default:
		return fmt.Errorf("Unknown profile: %s", config.Profile)
	}

	serfConfig.MemberlistConfig.BindAddr = bindIP
	serfConfig.MemberlistConfig.BindPort = bindPort
	serfConfig.NodeName = config.NodeName
	serfConfig.Tags = config.GetTags()
	serfConfig.CoalescePeriod = 3 * time.Second
	serfConfig.QuiescentPeriod = time.Second
	serfConfig.UserCoalescePeriod = 3 * time.Second
	serfConfig.UserQuiescentPeriod = time.Second
	if config.ReconnectTimeout != 0 {
		serfConfig.ReconnectTimeout = config.ReconnectTimeout
	}

	a.eventCh = make(chan serf.Event, 2048)
	serfConfig.EventCh = a.eventCh

	a.logger.Info("Serf agent starting")
	serf, err := serf.Create(serfConfig)
	if err != nil {
		return err
	}
	a.serf = serf

	// Start event loop
	go a.eventLoop()
	return nil
}

func (a *Agent) eventLoop() {
	serfShutdownCh := a.serf.ShutdownCh()

	a.logger.Info("agent: Listen for events")
	for {
		select {
		case e := <-a.eventCh:
			a.logger.Info("agent: Received event", "event", e.String())

			if me, ok := e.(serf.MemberEvent); ok {
				for _, member := range me.Members {
					a.logger.Debug("agent: Member event", "node", a.config.NodeName, "member", member.Name, "event", e.EventType())
				}
			}
		case <-serfShutdownCh:
			a.logger.Warn("agent: Serf shutdown detected, quitting")
			return
		}
	}
}

func (a *Agent) ShutdownCh() <-chan struct{} {
	return a.shutdownCh
}

func (a *Agent) Shutdown() error {
	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()

	if a.shutdown {
		return nil
	}

	a.shutdown = true
	close(a.shutdownCh)
	return nil
}
