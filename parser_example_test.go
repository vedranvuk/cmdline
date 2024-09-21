package cmdline_test

import (
	"fmt"

	"github.com/vedranvuk/cmdline"
)

func ExampleParse() {

	var verbose = false
	var force = false
	var count = 0
	var value = ""

	var config = &cmdline.Config{
		Arguments:             []string{"--verbose", "items", "add", "-f", "-c=9000", "--value=\"rofl\""},
		NoExecLastHandlerOnly: true,
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
			&cmdline.Command{
				Name: "help",
				Help: "Show help.",
				Handler: func(c cmdline.Context) error {
					fmt.Printf("Help requested for: %s.\n", c.Value("topic"))
					return nil
				},
				Options: cmdline.Options{
					&cmdline.Option{
						LongName: "topic",
						Help:     "Help topic.",
						Kind:     cmdline.Variadic,
					},
				},
			},
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
