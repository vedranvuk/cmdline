// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"fmt"
)

// TopicMap maps topic names to topic text.
type TopicMap map[string]string

// HelpCommand is a utility function that returns a command that handles "help"
// using HelpHandler.
func HelpCommand(topicMap TopicMap) (out *Command) {

	const doc = `Show help on certain topic or a command.

Usage:
  help <help topic|command [subcommand]>`
	out = &Command{
		Name:                "help",
		Help:                "Prints out a help topic or a command usage.",
		Doc:                 doc,
		RequireSubExecution: false,
		Handler: func(c Context) error {

			var config *Config
			if config = c.Config(); config == nil {
				return errors.New("HelpCommand requires a config")
			}

			var vals = c.Values("topic")
			if len(vals) == 0 {
				fmt.Fprintf(config.GetOutput(), "%s\n\n", out.Doc)
				if len(topicMap) > 0 {
					fmt.Fprintf(config.GetOutput(), "Available topics are:\n\n")
					for topic := range topicMap {
						fmt.Fprintf(config.GetOutput(), "  %s\n", topic)
					}
					fmt.Fprintf(config.GetOutput(), "\n")
				}
				if config.Commands.Count() > 0 {
					fmt.Fprintf(config.GetOutput(), "Available commands are:\n\n")
					PrintCommandsNoOptions(config.GetOutput(), config, config.Commands, 1)
					fmt.Fprintf(config.GetOutput(), "\n")
				}
				return nil
			}

			if len(vals) == 1 {
				if help, ok := topicMap[vals[0]]; ok {
					fmt.Fprint(config.GetOutput(), help)
					return nil
				}
			}

			var cmd *Command
			for cmds := config.Commands; cmds != nil; {
				if cmd = cmds.Find(vals[0]); cmd == nil {
					return fmt.Errorf("Command '%s' not found.", vals[0])
				}
				vals = vals[1:]

				if len(vals) > 0 {
					cmds = cmd.SubCommands
					continue
				}

				if cmd.Doc == "" {
					fmt.Fprintf(config.GetOutput(), "%s\n", cmd.Help)
				} else {
					fmt.Fprintf(config.GetOutput(), "%s\n", cmd.Doc)
				}

				var (
					showOpts = cmd.Options.Count() > 0
					showCmds = cmd.SubCommands.Count() > 0
				)
				if showOpts || showCmds {
					fmt.Fprintf(config.GetOutput(), "\n")
				}
				if showOpts {
					fmt.Fprintf(config.GetOutput(), "Command options are:\n\n")
					PrintOptions(config.GetOutput(), config, cmd.Options, 2)
					fmt.Fprintf(config.GetOutput(), "\n")
				}
				if showCmds {
					fmt.Fprintf(config.GetOutput(), "Available sub commands are:\n\n")
					PrintCommandsGroup(config.GetOutput(), config, cmd.SubCommands, 2)
					fmt.Fprintf(config.GetOutput(), "\n")
				}
				return nil
			}

			return fmt.Errorf("Command '%s' not found.", vals[0])
		},
	}
	out.Options.Variadic("topic", "Help topic.")
	return
}
