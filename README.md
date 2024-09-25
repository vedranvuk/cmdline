# cmdline

Cmdline is handler based command line parser. It supports Commands with SubCommands, multiple Option types and mapping to variables.

Package is in experimental stage.

## Examples

Simplest example with a few prefixed options parsed from os arguments.

```Go

func parse() {

	var config = DefaultOs()

	config.Globals.Optional("outfile", "o", "Specify output file.")

	config.GlobalsHandler = func(c cmdline.Context) error {
		fmt.Printf("outfile is: %s\n", c.Values("outfile").First())
		return nil
	}

	if err := config.Parse(): err != nil && err != cmdline.ErrNoArgs {
		log.Fatal(err)
	}
}
```

An example with global options and a few commands with subcommands.

```Go
func parse() {

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

	config.Commands.Handle("help", "Show help.", cmdline.HelpHandler).Options.
		Variadic("topic", "Help topic.")

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
```

Same functionality as in the previous example but in a declarative way.

```Go
func parse() {

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
```

## License

MIT
