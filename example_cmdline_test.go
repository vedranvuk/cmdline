package cmdline_test

import (
	"fmt"

	"github.com/vedranvuk/cmdline"
)

func ExampleParse() {

}

func ExampleBasic() {

	// Create a new command set to contain our commands.
	set := cmdline.NewCommands()

	// Register a "help" command.
	set.Handle("help", "Shows help.", "", func(c cmdline.Context) error {
		fmt.Printf("help command invoked\n")
		return nil
	})

	// Register a "create" command with two options,
	// first one required, second one optional.
	set.Handle("create", "Creates a thing.", "", func(c cmdline.Context) error {
		fmt.Printf("create command invoked\n")
		// Check if "input-dir" option was parsed
		// and if it was, retrieve and print its value.
		if c.Parsed("input-dir") {
			fmt.Printf("input-dir: %s\n", c.Value("input-dir"))
		}
		return nil
	}).GetOptions().
		Required("input-dir", "i", "Input directory.", "", "string").
		Boolean("output-dir", "o", "Output directory.", "")

	// Create global options that may preceede commands.
	globals := cmdline.NewOptions()
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
	set := cmdline.NewCommands()
	set.Handle("one", "Command one.", "", func(c cmdline.Context) error {
		fmt.Printf("one\n")
		return nil
	}).GetSubCommands().Handle("two", "Command two.", "", func(c cmdline.Context) error {
		fmt.Printf("two\n")
		return nil
	}).GetSubCommands().Handle("three", "Command three.", "", func(c cmdline.Context) error {
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
