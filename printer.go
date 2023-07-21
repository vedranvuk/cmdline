package cmdline

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// PrintConfig prints Globals and Commands to w from config.
func PrintConfig(w io.Writer, config *Config) {
	var wr = newTabWriter(w, 0)
	if config.Globals.Count() > 0 {
		io.WriteString(wr, fmt.Sprintf("Global options:\n"))
		PrintOptions(wr, config, config.Globals, 1)
	}
	if config.Commands.Count() > 0 {
		io.WriteString(wr, "Commands:\n")
		PrintCommands(wr, config, config.Commands, 1)
	}
	wr.Flush()
}

// PrintOptions prints commands to w idented with ident tabs using config.
func PrintCommands(w io.Writer, config *Config, commands Commands, indent int) {
	for _, command := range commands {
		PrintCommand(w, config, command, indent)
	}
}

// PrintOption prints command to w idented with ident tabs using config.
func PrintCommand(w io.Writer, config *Config, command *Command, indent int) {
	io.WriteString(w, getIndent(indent))
	io.WriteString(w, fmt.Sprintf("%s\t%s\n", command.Name, command.Help))
	if command.Options.Count() > 0 {
		PrintOptions(w, config, command.Options, indent+1)
	}
	if command.SubCommands.Count() > 0 {
		PrintCommands(w, config, command.SubCommands, indent+1)
	}
}

// PrintOptions prints options to w idented with ident tabs using config.
func PrintOptions(w io.Writer, config *Config, options Options, indent int) {
	var wr = newTabWriter(w, indent)
	for _, option := range options {
		PrintOption(wr, config, option, indent)
	}
	wr.Flush()
}

// PrintOption prints option to w idented with ident tabs using config.
func PrintOption(w io.Writer, config *Config, option Option, indent int) {
	io.WriteString(w, getIndent(indent))
	switch o := option.(type) {
	case *Boolean:
		io.WriteString(w, fmt.Sprintf("%s%s\t%s%s\t%s\n",
			config.ShortPrefix, o.ShortName, config.LongPrefix, o.LongName, o.Help))
	case *Optional:
		io.WriteString(w, fmt.Sprintf("%s%s=<value>\t%s%s=<value>\t%s\n",
			config.ShortPrefix, o.ShortName, config.LongPrefix, o.LongName, o.Help))
	case *Required:
		io.WriteString(w, fmt.Sprintf("%s%s=<value>\t%s%s=<value>\t%s\n",
			config.ShortPrefix, o.ShortName, config.LongPrefix, o.LongName, o.Help))
	case *Repeated:
		io.WriteString(w, fmt.Sprintf("... %s%s=<value>\t... %s%s=<value>\t%s\n",
			config.ShortPrefix, o.ShortName, config.LongPrefix, o.LongName, o.Help))
	case *Indexed:
		io.WriteString(w, fmt.Sprintf("<%s>\t%s\n", o.Name, o.Help))
	case *Variadic:
		io.WriteString(w, fmt.Sprintf("... %s\t%s\n", o.Name, o.Help))
	}
}

// getIndent returns depth number of tabs used for indentation.
func getIndent(depth int) (result string) {
	for i := 0; i < depth; i++ {
		result += "\t"
	}
	return
}

func newTabWriter(output io.Writer, indent int) *tabwriter.Writer {
	return tabwriter.NewWriter(output, 2, 2, 4, 32, 0)
}
