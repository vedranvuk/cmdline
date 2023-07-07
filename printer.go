package cmdline

import (
	"fmt"
	"io"
)

// PrintConfig prints Globals and Commands to w from config.
func PrintConfig(w io.Writer, config *Config) {
	if config.Globals.Count() > 0 {
		io.WriteString(w, fmt.Sprintf("Global options:\n"))
		PrintOptions(w, config, config.Globals, 1)
	}
	if config.Commands.Count() > 0 {
		io.WriteString(w, "Commands:\n")
		PrintCommands(w, config, config.Commands, 1)
	}
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
	for _, option := range options {
		PrintOption(w, config, option, indent)
	}
}

// PrintOption prints option to w idented with ident tabs using config.
func PrintOption(w io.Writer, config *Config, option Option, indent int) {
	io.WriteString(w, getIndent(indent))
	switch o := option.(type) {
	case *Boolean:
		io.WriteString(w, fmt.Sprintf("%s%s\t%s%s\t%s\n",
			config.ShortPrefix, o.ShortName, config.LongPrefix, o.LongName, o.Help))
	case *Optional:
		io.WriteString(w, fmt.Sprintf("%s%s=value\t%s%s=value\t%s\n",
			config.ShortPrefix, o.ShortName, config.LongPrefix, o.LongName, o.Help))
	case *Required:
		io.WriteString(w, fmt.Sprintf("%s%s=value\t%s%s=value\t%s\n",
			config.ShortPrefix, o.ShortName, config.LongPrefix, o.LongName, o.Help))
	case *Indexed:
		io.WriteString(w, fmt.Sprintf("%s <value>\t%s\n", o.Name, o.Help))
	case *Variadic:
		io.WriteString(w, fmt.Sprintf("%s ...argument\t%s\n", o.Name, o.Help))
	}
}

// getIndent returns depth number of tabs used for indentation.
func getIndent(depth int) (result string) {
	for i := 0; i < depth; i++ {
		result += "\t"
	}
	return
}
