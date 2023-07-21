# cmdline

Cmdline is handler based command line parser using a declarative approach. Defaults to _new_ GNU style, supports mapping to variables, commands and subcommands and various option types.

Package is in experimental stage.

## Table Of Contents

- [Options](#Options)
  - [Boolean](#boolean)
  - [Optional](#optional)
  - [Required](#required)
  - [Repeated](#repeated)
  - [Indexed](#indexed)
  - [Variadic](#variadic)

## Options

### Boolean

A Boolean is an Option that takes no value. If it is given in arguments it will be marked as parsed and its state retrievable via Handler Context.IsParsed.

```Go
var config = &cmdline.Config{
	Args: []string{"--boolean"}
	Globals: &cmdline.Options{
		&cmdline.Boolean{
			LongName: "boolean",
		},
	},
	GlobalsHandler: func(c cmdline.Context) error {
		if c.IsParsed("boolean") {
			fmt.Println("boolean was specified in command line arguments.")
		}
		return nil
	}

	cmdline.Parse(config)

	// Output:
	// boolean was specified in command line arguments.
}
```

### Optional

A Optional option is an option that takes a single value and is not required to be given in arguments, i.e. it will not raise an error if not specified.

```Go
var config = &cmdline.Config{
	Args: []string{"--optional=some_value"}
	Globals: &cmdline.Options{
		&cmdline.Optional{
			LongName: "optional",
		},
	},
	GlobalsHandler: func(c cmdline.Context) error {
		if c.IsParsed("optional") {
			fmt.Printf("Optional option value: %v\n", c.RawValues("optional").First())
		}
		return nil
	}

	cmdline.Parse(config)

	// Output:
	// Optional option value: some_value
}
```

### Required

A required option is an option that must be specified in command line arguments. it takes a single value and may not be repeated.

```Go
var config = &cmdline.Config{
	Args: []string{""}
	Globals: &cmdline.Options{
		&cmdline.Required{
			LongName: "required",
		},
	},

	if err := cmdline.Parse(config); err != nil {
		fmt.Printf("%v\n", err)
	}

	// Output:
	// required option 'required' not parsed.
}
```

### Repeated

A repeated option is an option that can be specified zero or more times. Each time it is given in arguments it takes a single value.

```Go
var config = &cmdline.Config{
	Args: []string{"--repeated=1", "--repeated=two=two", "--repeated=\"three=3\""}
	Globals: &cmdline.Options{
		&cmdline.Repeated{
			LongName: "repeated",
		},
	},
	GlobalsHandler: func(c cmdline.Context) error {
		fmt.Printf("%v\n", c.RawValues("repeated"))
		return nil
	}

	cmdline.Parse(config)

	// Output:
	// [1 two=two three=3]
}
```

### Indexed

An indexed option is an option that takes a single value and is not addressed by long or short nyme. Instead, arguments passed to a command are matched by index as they were specified to an index in order as Indexed commands were defined in Command Options.

```Go
var config = &cmdline.Config{
	Args: []string{"1", "2", "3"}
	Globals: &cmdline.Options{
		&cmdline.Indexed{
			LongName: "one",
		},
		&cmdline.Indexed{
			LongName: "two",
		},
		&cmdline.Indexed{
			LongName: "three",
		},
	},
	GlobalsHandler: func(c cmdline.Context) error {
		fmt.Printf("%v\n", c.RawValues("one").First())
		fmt.Printf("%v\n", c.RawValues("two").First())
		fmt.Printf("%v\n", c.RawValues("three").First())
		return nil
	}

	cmdline.Parse(config)

	// Output:
	// 1
	// 2
	// 3
}
```

### Variadic

A variadic option is an option that takes any unparsed arguments as values.

```Go
var config = &cmdline.Config{
	Args: []string{"1", "2", "3"}
	Globals: &cmdline.Options{
		&cmdline.Variadic{
			Name: "variadic",
		},
	},
	GlobalsHandler: func(c cmdline.Context) error {
		fmt.Printf("%v\n", c.RawValues("variadic"))
		return nil
	}

	cmdline.Parse(config)

	// Output:
	// [1 2 3]
}
```

## Usage

```Go
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
		Arguments:   []string{"--verbose", "items", "add", "-f", "-c=9000", "--value=\"rofl\""},
		Globals: cmdline.Options{
			&cmdline.Boolean{
				LongName:    "verbose",
				ShortName:   "v",
				Help:        "Be verbose.",
				MappedValue: &verbose,
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
					&cmdline.Variadic{
						Name: "topic",
						Help: "Help topic.",
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
							&cmdline.Boolean{
								LongName:    "force",
								ShortName:   "f",
								Help:        "Force it.",
								MappedValue: &force,
							},
							&cmdline.Optional{
								LongName:    "count",
								ShortName:   "c",
								Help:        "Give a count.",
								MappedValue: &count,
							},
							&cmdline.Optional{
								LongName:    "value",
								ShortName:   "v",
								Help:        "Give a value.",
								MappedValue: &value,
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
```
