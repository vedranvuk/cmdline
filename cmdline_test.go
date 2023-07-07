package cmdline

import (
	"fmt"
	"testing"
)

func TestOptions(t *testing.T) {

	var config = &Config{
		Args: []string{"items", "add", "-f"},
		Commands: Commands{
			&Command{
				Name: "help",
				Help: "Show help.",
				Handler: func(c Context) error {
					fmt.Printf("Help requested for: %s.\n", c.Value("topic"))
					return nil
				},
				Options: Options{
					&Variadic{
						Name: "topic",
						Help: "Help topic.",
					},
				},
			},
			&Command{
				Name: "items",
				Help: "Operate on items.",
				Handler: func(c Context) error {
					fmt.Printf("command: items\n")
					return nil
				},
				SubCommands: Commands{
					&Command{
						Name: "add",
						Help: "Add an item.",
						Handler: func(c Context) error {
							fmt.Printf("command: add (force: %t)\n", c.Parsed("force"))
							return nil
						},
						Options: Options{
							&Boolean{
								LongName:  "force",
								ShortName: "f",
								Help:      "Force it.",
							},
						},
					},
					&Command{
						Name: "remove",
						Help: "Remove an item.",
						Handler: func(c Context) error {
							fmt.Printf("command: remove\n")
							return nil
						},
					},
				},
			},
		},
	}

	if err := Parse(config); err != nil {
		t.Fatal(err)
	}
}
