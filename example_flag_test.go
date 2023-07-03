package cmdline_test

import (
	"fmt"

	"github.com/vedranvuk/cmdline"
)

func ExampleBasic() {

	// Create a new command set to contain our commands.
	set := cmdline.NewCommandSet()

	// Register a "help" command.
	set.Handle("help", "Shows help.", func(c cmdline.Context) error {
		fmt.Printf("help command invoked\n")
		return nil
	})

	// Register a "create" command with two oJptions,
	// first one required, second one optional.
	set.Handle("create", "Creates a thing.", func(c cmdline.Context) error {
		fmt.Printf("create command invoked\n")
		// Check if "input-dir" option was parsed
		// and if it was, retrieve and print its value.
		if c.Parsed("input-dir") {
			fmt.Printf("input-dir: %s\n", c.Value("input-dir"))
		}
		return nil
	}).Options().
		Required("input-dir", "i", "Input directory.", "", "string").
		Boolean("output-dir", "o", "Output directory.", "")

	// Create global options that may preceede commands.
	globals := cmdline.NewOptionSet()
	globals.Boolean("verbose", "v", "Be verbose.", "")

	// Parse some arguments.
	args := []string{"-v", "create", "-i=/home/myname"}
	if err := cmdline.Parse(&cmdline.Config{
		Args:     args,
		Globals:  globals,
		Commands: set,
	}); err != nil {
		panic(err)
	}

	// Output:
	// create command invoked
	// input-dir: /home/myname
}

func ExampleSubCommands() {
	set := cmdline.NewCommandSet()
	set.Handle("one", "Command one.", func(c cmdline.Context) error {
		fmt.Printf("one\n")
		return nil
	}).Sub().Handle("two", "Command two.", func(c cmdline.Context) error {
		fmt.Printf("two\n")
		return nil
	}).Sub().Handle("three", "Command three.", func(c cmdline.Context) error {
		fmt.Printf("three\n")
		return nil
	})
	args := []string{"one", "two", "three"}
	cmdline.Parse(&cmdline.Config{
		Args:     args,
		Commands: set,
	})

	// Output:
	// one
	// two
	// three
}
