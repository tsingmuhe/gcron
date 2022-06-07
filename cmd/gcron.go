package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"

	"github.com/tsingmuhe/gcron/cmd/agent"
	"github.com/tsingmuhe/gcron/cmd/help"
)

const Name = "gcron"
const Version = "v1.1.1"
const Usage = `Open source distributed job scheduling system.

Usage: gcron [--help] [--version] <command> [options] [<args>]

Available commands:
`

func Execute() int {
	cmds := registeredCommands()

	var names []string
	for c := range cmds {
		names = append(names, c)
	}

	c := &cli.CLI{
		Args:         os.Args[1:],
		Commands:     cmds,
		Name:         Name,
		Version:      Version,
		Autocomplete: true,
		HelpFunc:     help.CliHelpFunc(Usage),
		HelpWriter:   os.Stdout,
		ErrorWriter:  os.Stderr,
	}

	exitCode, err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %v\n", err)
		return 1
	}
	return exitCode
}

func registeredCommands() map[string]cli.CommandFactory {
	ui := &cli.BasicUi{Writer: os.Stdout}

	return map[string]cli.CommandFactory{
		"agent": func() (cli.Command, error) {
			return &agent.Command{
				Ui: ui,
			}, nil
		},
	}
}
