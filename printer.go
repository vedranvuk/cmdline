package cmdline

import (
	"fmt"
	"io"
)

// PrintOptions prints options to w idented with ident tabs.
func PrintOptions(w io.Writer, options *Options, indent int) {
	if options == nil {
		return
	}
	if len(options.options) > 0 {
		for _, option := range options.options {
			PrintOption(w, option, indent)
		}
		io.WriteString(w, "\n")
	}
}

// PrintOption prints option to w idented with ident tabs.
func PrintOption(w io.Writer, option Option, indent int) {
	if option == nil {
		return
	}
	io.WriteString(w, getIndent(indent))
	switch o := option.(type) {
	case *Boolean:
		io.WriteString(w, fmt.Sprintf("-%s\t--%s\t%s\n", o.ShortName, o.LongName, ""))
	case *Optional:
		io.WriteString(w, fmt.Sprintf("-%s=value\t--%s=value\t%s\n", o.ShortName, o.LongName, ""))
	case *Required:
		io.WriteString(w, fmt.Sprintf("-%s=value\t--%s=value\t%s\n", o.ShortName, o.LongName, ""))
	case *Indexed:
		io.WriteString(w, fmt.Sprintf("%s <value>\t%s\n", o.Name, ""))
	case *Variadic:
		io.WriteString(w, fmt.Sprintf("%s ...value\t%s\n", o.Name, ""))
	}
}

// PrintOptions prints commands to w idented with ident tabs.
func PrintCommands(w io.Writer, commands *Commands, indent int) {
	if commands == nil {
		return
	}
	for _, command := range commands.commands {
		PrintCommand(w, command, indent)
	}
}

// PrintOption prints command to w idented with ident tabs.
func PrintCommand(w io.Writer, command *Command, indent int) {
	if command == nil {
		return
	}
	io.WriteString(w, getIndent(indent))
	io.WriteString(w, fmt.Sprintf("%s\t%s\n", command.Name, ""))
	if command.Options != nil && len(command.Options.options) > 0 {
		PrintOptions(w, command.Options, indent+1)
	}
	io.WriteString(w, "\n")
	if command.SubCommands != nil && command.SubCommands.Count() > 0 {
		io.WriteString(w, "\n")
		PrintCommands(w, command.SubCommands, indent+1)
	}
}

// PrintConfig prints Globals and Commands to w from config.
func PrintConfig(w io.Writer, config *Config) {
	if config.Globals != nil && len(config.Globals.options) > 0 {
		io.WriteString(w, fmt.Sprintf("Global options:\n\n"))
		PrintOptions(w, config.Globals, 1)
	}
	if config.Commands != nil && config.Commands.Count() > 0 {
		io.WriteString(w, "Commands:\n\n")
		PrintCommands(w, config.Commands, 1)
	}
}

func getIndent(depth int) (result string) {
	for i := 0; i < depth; i++ {
		result += "\t"
	}
	return
}
