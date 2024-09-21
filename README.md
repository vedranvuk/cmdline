# cmdline

Cmdline is handler based command line parser. It supports Commands with SubCommands, multiple Option types and mapping to variables.

Package is in experimental stage.

## Table Of Contents

- [Examples](#examples)
- [Options](#options)
  - [Boolean](#boolean)
  - [Optional](#optional)
  - [Required](#required)
  - [Repeated](#repeated)
  - [Indexed](#indexed)
  - [Variadic](#variadic)

## Examples

### Global options

Simplest declarative example with one global Option (i.e. not tied to any command) named `verbose` (long name) or `v` (short name).

```Go
cmdline.Parse(&cmdline.Config{
	Args: os.Args[1:],
	Globals: &cmdline.Options{
		&cmdline.Boolean{
			LongName: "help",
			ShortName:"h",
			Help: "Show help.",
		},
	},
})
```

Passing `--help` or `-h` in arguments will set the Option as parsed.

Options may be declared using helper methods and can be chained.

```Go
var config = &cmdline.Config{
	Args: os.Args[1:],
}
config.Globals.
	Boolean("help", "h", "Show help.").
	Boolean("verbose", "v", "Be verbose.")
cmdline.Parse(config)
```

Without a `Config.GlobalsHandler` set, the Option can be examined from the config, post-parse.

```Go
var config = &cmdline.Config{
	Args: os.Args[1:],
	Globals: &cmdline.Options{
		&cmdline.Boolean{
			LongName: "help",
			ShortName:"h",
			Help: "Show help.",
		},
	},
}

cmdline.Parse(config)

fmt.Printf("help was parsed: %t\n", config.Globals.FindByLongName("help").IsParsed())

// Output:
// Help was parsed: true
```

A Handler can be set in the config to inspect global Options during parsing.

```Go
cmdline.Parse(&cmdline.Config{
	Args: os.Args[1:],
	Globals: &cmdline.Options{
		&cmdline.Boolean{
			LongName: "help",
			ShortName:"h",
			Help: "Show help.",
		},
	},
	GlobalsHandler: func(c cmdline.Context) error {
		if c.IsParsed("help") {
			fmt.Println("help was requested.")
		}
		return nil
	}
})

// Output:
// Help was requested.
```

### A single command with a handler

```Go
cmdline.Parse(&cmdline.Config{
	Args: []string{"execute"},
	Commands: &cmdline.Commands{
		{
			Name: "execute",
			Handler: func(c cmdline.Context) error {
				fmt.Println("execute executed")
				return nil
			},
		},
	},
})

// Output:
// execute executed
```

If a Command handler returns a non-nil error the parsing will be aborted and the error propagated back to `cmdline.Parse` caller.

### One global option and one command with one option

This example shows a config with one Boolean global Option and one Command with one Required Option. The arguments to parse are predefined in this example.

```Go
if err := cmdline.Parse(&cmdline.Config{
	Args: []string{"-v", "create", "--name=MyProject"},
	Globals: &cmdline.Options{
		&cmdline.Boolean{
			LongName: "verbose",
			ShortName: "v",
			Help: "Be verbose.",
		},
	},
	Commands: &cmdline.Commands{
		{
			Name: "create",
			Help: "Create a project``.",
			Handler: func(c cmdline.Context) error {
				fmt.Printf("Project name: &s\n", c.RawValues("name").First())
				return nil
			},
			Options: &cmdline.Options{
				&cmdline.Required{
					LongName: "name",
					ShortName: "n",
					Help: "Specify project name.",
				},
			},
		},
	},
}); err != nil {
	log.Fatal(err)
}

// Output:
// Project name: MyProject
```

Invoking the program with arguments `myprogram -v create --name=MyProject` will set `verbose` Option as parsed, invoke the `create` Command, set its' Option `name` as parsed and assign value `MyProject` to it which is then retrievable from its Command Handler or via `Config`, post-parse.

### SubCommands

A Commands set can have sub commands.

```Go
cmdline.Parse(&cmdline.Config{
	Args: []string{"-v", "ip", "link", "wa0gtfo6", "up"}
	Globals: &cmdline.Options{
		&cmdline.Boolean{
			LongName: "verbose",
			ShortName: "v",
		},
	},
	Commands: &cmdline.Commands{
		{
			Name: "ip",
			SubCommands: &cmdline.Commands{
				{
					Name: "link",
					Options: &cmdline.Options{
						&cmdline.Indexed{
							Name: "interface-name",
						},
					},
					SubCommands: &cmdline.Commands{
						{
							Name: "up",
							Handler: func(c cmdline.Context) error {
								return nil
							},
						},
					},
				},
			},
		},
	},
})
```

In the example above, a global option `verbose` is parsed then a Command `ip` is invoked, then its SubCommand `link` is invoked and its option `link` is receiving a value then finally the `up` Command is invoked.

Handlers of these Commands are called in the order as they are parsed from arguments.

### Mapped values

An Option can map to a variable deriving from one of supported core types. The string argument given to the Option will be parsed into the mapped variable.

Supported types are: `*bool, *int...*int64, *uint...*uint64, *float32/64, *string, *[]string` and any type that implements the `cmdline.Value` interface.

```Go
// Variable to receive the parsed value.
var name string

cmdline.Parse(&cmdline.Config{
	Args: []string{"--name=MyProject"},
	Globals: &cmdline.Options{
		&cmdline.Required{
			LongName: "name",
			ShortName:"n",
			Help: "Specify name.",
			// Variable mapping, must be a pointer to a supported type.
			MappedValue: &name,
		},
	},
})
fmt.Println("%v\n", name)

// Output:
// MyProject
```

## Options

Several option types are available:

### Boolean

A Boolean is an Option that takes no value. If an argument that addresses it is found in arguments the option will be set as parsed, effectively acting like a bool.

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

A Optional Option is an option that takes a single value and is not required to be specified in arguments, i.e. it will not raise an error if not specified.

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

A required Option is like Optional Option except it returns an error via `cmdline.Parse` if it is not specified in arguments.

```Go
var config = &cmdline.Config{
	Args: []string{}
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

A Repeated Option is an option that can be specified zero or more times. Each time it is given in arguments it takes a single value. All raw string values given to the Option as arguments are retrievable via its respective Command or Globals Handler `Context.RawValues()` method.

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

An Indexed Option is an option that takes a single value and is not addressed by its name. Instead, arguments passed to a command are matched by index as they were specified to an index of the order asthey were defined in its parent Options.

Indexed Options can appear anywhere in between other Option types in arguments and the order in which Indexed Options appear among other Indexed Options is taken as the index matched to the order in which they were defined.

Same goes for the order in which they are defined; their index is calculated from the order they were defined among other Indexed Options. Other Option types have no indexing property.

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

A Variadic Option is an option that takes any currently unparsed arguments as its values.

This creates ambiguity with SubCommands as Variadic Option is not named and not prefixed with any tokens. It simply consumes anything left in the arguments and stops the parse process.

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

## License

MIT