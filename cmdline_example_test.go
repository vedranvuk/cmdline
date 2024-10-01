// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline_test

import (
	"fmt"

	"github.com/vedranvuk/cmdline"
)

func ExampleParse1() {

	var (
		verbose = false
		force   = false
		count   = 0
		value   = ""
	)

	var config = cmdline.Default("--verbose", "items", "add", "-f", "-c=9000", "--value=\"rofl\"")
	config.UseAssignment = true
	config.ExecAllHandlers = true

	config.Globals.BooleanVar("verbose", "v", "Be verbose.", &verbose)
	config.GlobalsHandler = func(c cmdline.Context) error {
		fmt.Printf("verbose requested.\n")
		return nil
	}

	config.Commands.Register(cmdline.HelpCommand(nil))

	{
		var items = config.Commands.Handle("items", "Operate on items.",
			func(c cmdline.Context) error {
				fmt.Printf("command: items\n")
				return nil
			},
		)

		items.SubCommands.Handle("add", "Add an item.",
			func(c cmdline.Context) error {
				fmt.Printf("command: add (force: %t) (count: %t)\n", c.Parsed("force"), c.Parsed("count"))
				return nil
			},
		).Options.
			BooleanVar("force", "f", "Force it.", &force).
			OptionalVar("count", "c", "Give a count.", &count).
			OptionalVar("value", "v", "Give a value.", &value)

		items.SubCommands.Handle("remove", "Remove an item.", func(c cmdline.Context) error {
			fmt.Printf("command: remove\n")
			return nil
		})
	}

	if err := cmdline.Parse(config); err != nil {
		panic(err)
	}

	// Output:
	// verbose requested.
	// command: items
	// command: add (force: true) (count: true)
}

func ExampleParse2() {

	var (
		verbose = false
		force   = false
		count   = 0
		value   = ""
	)

	var config = &cmdline.Config{
		Args:            []string{"--verbose", "items", "add", "-f", "-c=9000", "--value=\"rofl\""},
		UseAssignment:   true,
		ExecAllHandlers: true,
		Globals: cmdline.Options{
			&cmdline.Option{
				LongName:  "verbose",
				ShortName: "v",
				Help:      "Be verbose.",
				Var:       &verbose,
				Kind:      cmdline.Boolean,
			},
		},
		GlobalsHandler: func(c cmdline.Context) error {
			fmt.Printf("verbose requested.\n")
			return nil
		},
		Commands: cmdline.Commands{
			cmdline.HelpCommand(nil),
			&cmdline.Command{
				Name: "items",
				Help: "Operate on items.",
				Handler: func(c cmdline.Context) error {
					fmt.Printf("command: items\n")
					return nil
				},
				SubCommands: cmdline.Commands{
					&cmdline.Command{
						Name: "add",
						Help: "Add an item.",
						Handler: func(c cmdline.Context) error {
							fmt.Printf("command: add (force: %t) (count: %t)\n", c.Parsed("force"), c.Parsed("count"))
							return nil
						},
						Options: cmdline.Options{
							&cmdline.Option{
								LongName:  "force",
								ShortName: "f",
								Help:      "Force it.",
								Var:       &force,
								Kind:      cmdline.Boolean,
							},
							&cmdline.Option{
								LongName:  "count",
								ShortName: "c",
								Help:      "Give a count.",
								Var:       &count,
								Kind:      cmdline.Optional,
							},
							&cmdline.Option{
								LongName:  "value",
								ShortName: "v",
								Help:      "Give a value.",
								Var:       &value,
								Kind:      cmdline.Optional,
							},
						},
					},
					&cmdline.Command{
						Name: "remove",
						Help: "Remove an item.",
						Handler: func(c cmdline.Context) error {
							fmt.Printf("command: remove\n")
							return nil
						},
					},
				},
			},
		},
	}

	if err := cmdline.Parse(config); err != nil {
		panic(err)
	}

	// Output:
	// verbose requested.
	// command: items
	// command: add (force: true) (count: true)
}
