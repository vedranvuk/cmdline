// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

// HelpHandler is a utility handler that prints the current configuration.
var HelpHandler = func(c Context) error {

	var vals []string
	if vals = c.Values("topic"); len(vals) == 0 {
		// TODO Print out a synopsis.
	}

	if config := c.Config(); config != nil {
		config.PrintUsage()
		PrintConfig(config.GetOutput(), c.Config())
	}

	return nil
}

// TopicMap maps topic names to topic text.
type TopicMap map[string]string

// HelpCommand is a utility function that returns a command that handles "help"
// using HelpHandler.
func HelpCommand(topicMap TopicMap) (out *Command) {
	out = &Command{
		Name:                "help",
		Help:                "Prints out the command usage.",
		RequireSubExecution: false,
		Handler: func(c Context) error {
			
			var vals []string
			if vals = c.Values("topic"); len(vals) == 0 {
				// TODO Print out a synopsis.
			}

			if config := c.Config(); config != nil {
				config.PrintUsage()
				PrintConfig(config.GetOutput(), c.Config())
			}

			return nil
		},
	}
	out.Options.Variadic("topic", "Help topic.")
	return
}
