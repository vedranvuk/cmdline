// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// PrintConfig prints Globals and Commands to w from config.
func PrintConfig(w io.Writer, config *Config) {
	var wr = newTabWriter(w)
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
	var wr = newTabWriter(w)
	for _, option := range options {
		if option.Kind != Indexed {
			continue
		}
		PrintOption(wr, config, option, indent)
	}
	for _, option := range options {
		if option.Kind != Boolean {
			continue
		}
		PrintOption(wr, config, option, indent)
	}
	for _, option := range options {
		if option.Kind != Optional {
			continue
		}
		PrintOption(wr, config, option, indent)
	}
	for _, option := range options {
		if option.Kind != Required {
			continue
		}
		PrintOption(wr, config, option, indent)
	}
	for _, option := range options {
		if option.Kind != Repeated {
			continue
		}
		PrintOption(wr, config, option, indent)
	}
	for _, option := range options {
		if option.Kind != Variadic {
			continue
		}
		PrintOption(wr, config, option, indent)
	}
	wr.Flush()
}

// PrintOption prints option to w idented with ident tabs using config.
func PrintOption(w io.Writer, config *Config, option *Option, indent int) {
	io.WriteString(w, getIndent(indent))
	switch option.Kind {
	case Boolean:
		io.WriteString(w, fmt.Sprintf("%s\t%s\n", h(config, option.LongName, option.ShortName, false), option.Help))
	case Optional:
		io.WriteString(w, fmt.Sprintf("%s\t%s\n", h(config, option.LongName, option.ShortName, true), option.Help))
	case Required:
		io.WriteString(w, fmt.Sprintf("%s\t%s\n", h(config, option.LongName, option.ShortName, true), option.Help))
	case Repeated:
		io.WriteString(w, fmt.Sprintf("%s\t%s\n", h(config, option.LongName, option.ShortName, true), option.Help))
	case Indexed:
		io.WriteString(w, fmt.Sprintf("\t<%s>\t%s\n", option.LongName, option.Help))
	case Variadic:
		io.WriteString(w, fmt.Sprintf("... \t%s\t%s\n", option.LongName, option.Help))
	}
}

func h(config *Config, longname, shortname string, value bool) (result string) {
	if shortname != "" {
		result = fmt.Sprintf("%s%s\t%s%s", config.ShortPrefix, shortname, config.LongPrefix, longname)
	} else {
		result = fmt.Sprintf("\t%s%s", config.LongPrefix, longname)
	}
	if value {
		result = result + " <value>"
	}
	return
}

// getIndent returns depth number of tabs used for indentation.
func getIndent(depth int) (result string) {
	for i := 0; i < depth; i++ {
		result += "  "
	}
	return
}

func newTabWriter(output io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(output, 2, 2, 2, 32, 0)
}
