package cmdline

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestOptions(t *testing.T) {

	var verbose = false
	var force = false
	var count = 0
	var value = ""

	var config = &Config{
		Arguments:   []string{"--verbose", "items", "add", "-f", "-c=9000", "--value=\"rofl\""},
		LongPrefix:  DefaultLongPrefix,
		ShortPrefix: DefaultShortPrefix,
		Globals: Options{
			&Boolean{
				LongName:    "verbose",
				ShortName:   "v",
				Help:        "Be verbose.",
				MappedValue: &verbose,
			},
		},
		GlobalsHandler: func(c Context) error {
			fmt.Printf("verbose requested.\n")
			return nil
		},
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
							fmt.Printf("command: add (force: %t) (count: %t)\n", c.Parsed("force"), c.Parsed("count"))
							return nil
						},
						Options: Options{
							&Boolean{
								LongName:    "force",
								ShortName:   "f",
								Help:        "Force it.",
								MappedValue: &force,
							},
							&Optional{
								LongName:    "count",
								ShortName:   "c",
								Help:        "Give a count.",
								MappedValue: &count,
							},
							&Optional{
								LongName:    "value",
								ShortName:   "v",
								Help:        "Give a value.",
								MappedValue: &value,
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

	PrintConfig(os.Stdout, config)

	if err := Parse(config); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("verbose: %t\n", verbose)
	fmt.Printf("force: %t\n", force)
	fmt.Printf("count: %d\n", count)
	fmt.Printf("value: %s\n", value)
}
