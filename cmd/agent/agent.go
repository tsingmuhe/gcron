package agent

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mitchellh/cli"

	"github.com/tsingmuhe/gcron/gcron"
)

const (
	synopsis = "Runs a gcron agent"
	usage    = `Usage: gcron agent [options]

options:
  -config-file=<FILE>	Load configuration from FILE
        
`
)

const (
	gracefulTimeout = 3 * time.Second
)

type Command struct {
	Ui cli.Ui

	args []string
}

func (c *Command) Synopsis() string {
	return synopsis
}

func (c *Command) Help() string {
	return strings.TrimSpace(usage)
}

func (c *Command) Run(args []string) int {
	c.Ui = &cli.PrefixedUi{
		OutputPrefix: "==> ",
		InfoPrefix:   "    ",
		ErrorPrefix:  "==> ",
		Ui:           c.Ui,
	}
	c.args = args

	config := c.readConfig()
	if config == nil {
		return 1
	}

	agent := c.setupAgent(config)
	if agent == nil {
		return 1
	}

	// Wait for exit
	return c.handleSignals(agent)
}

func (c *Command) readConfig() *gcron.Config {
	cmdFlags := flag.NewFlagSet("agent", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }

	var configFile string
	cmdFlags.StringVar(&configFile, "config-file", "", "file to read config from")

	if err := cmdFlags.Parse(c.args); err != nil {
		return nil
	}

	config := gcron.DefaultConfig()
	if len(configFile) > 0 {
		fileConfig, err := gcron.ReadConfigPaths(configFile)
		if err != nil {
			c.Ui.Error(err.Error())
			return nil
		}
		config = gcron.MergeConfig(config, fileConfig)
	}

	return config
}

func (c *Command) setupAgent(config *gcron.Config) *gcron.Agent {
	agent := gcron.NewAgent(config)

	c.Ui.Output("Starting Serf agent...")
	if err := agent.Start(); err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to setup the gcron agent: %v", err))
		return nil
	}

	return agent
}

func (c *Command) handleSignals(agent *gcron.Agent) int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// Wait for a signal
WAIT:
	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-agent.ShutdownCh():
		// Agent is already shutdown!
		return 0
	}

	c.Ui.Output(fmt.Sprintf("Caught signal: %v", sig))

	// Check if this is a SIGHUP
	if sig == syscall.SIGHUP {
		//config = c.handleReload(config, agent)
		goto WAIT
	}

	// Fail fast if not doing a graceful leave
	if sig != syscall.SIGTERM && sig != os.Interrupt {
		return 1
	}

	// Attempt a graceful leave
	gracefulCh := make(chan struct{})
	c.Ui.Output("Gracefully shutting down agent...")

	go func() {
		if err := agent.Shutdown(); err != nil {
			c.Ui.Error(fmt.Sprintf("Error: %s", err))
			return
		}

		close(gracefulCh)
	}()

	// Wait for leave or another signal
	select {
	case <-signalCh:
		return 1
	case <-time.After(gracefulTimeout):
		return 1
	case <-gracefulCh:
		return 0
	}
}
