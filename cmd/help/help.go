package help

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"sort"
	"strings"

	"github.com/mitchellh/cli"
)

func CliHelpFunc(usage string) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		out := new(bytes.Buffer)
		out.WriteString(usage)

		maxKeyLen := 0
		keys := make([]string, 0, len(commands))
		for key := range commands {
			if len(key) > maxKeyLen {
				maxKeyLen = len(key)
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			commandFunc := commands[key]
			command, err := commandFunc()
			if err != nil {
				log.Printf("[ERR] cli: Command '%s' failed to load: %s", key, err)
				continue
			}

			key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen-len(key)))
			out.WriteString(fmt.Sprintf("  %s  %s\n", key, command.Synopsis()))
		}

		return strings.TrimRight(out.String(), "\n")
	}
}

func CommandHelp(usage string, flags *flag.FlagSet) string {
	out := new(bytes.Buffer)
	out.WriteString(usage)
	flags.VisitAll(func(f *flag.Flag) {
		printFlag(out, f)
	})

	return strings.TrimRight(out.String(), "\n")
}

func printFlag(w io.Writer, f *flag.Flag) {
	name, usage := flag.UnquoteUsage(f)
	if name != "" {
		fmt.Fprintf(w, "  -%s=<%s>\n", f.Name, name)
	} else {
		fmt.Fprintf(w, "  -%s\n", f.Name)
	}
	fmt.Fprintf(w, "  \t%s\n\n", usage)
}
