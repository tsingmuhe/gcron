package gcron

import (
	"sync"
)

type Agent struct {
	config *Config

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
	return nil
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
