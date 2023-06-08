package cmdline_test

import (
	"fmt"

	"github.com/vedranvuk/flag"
)

func ExampleBasic() {

	// Create a new command set to contain our commands.
	set := flag.New()

	// Register a "help" command.
	set.Handle("help", func(c flag.Context) error {
		fmt.Printf("help command invoked\n")
		return nil
	})

	// Register a "create" command with two oJptions,
	// first one required, second one optional.
	set.Handle("create", func(c flag.Context) error {
		fmt.Printf("create command invoked\n")
		// Check if "input-dir" option was parsed
		// and if it was, retrieve and print its value.
		if c.Parsed("input-dir") {
			fmt.Printf("input-dir: %s\n", c.Value("input-dir"))
		}
		return nil
	}).Options().
		Required("input-dir", "i", "string", "Input directory.").
		Boolean("output-dir", "o", "Output directory.")

	// Create global options that may preceede commands.
	globals := flag.Globals()
	globals.Boolean("verbose", "v", "Be verbose.")

	// Parse some arguments.
	args := []string{"-v", "create", "-i=/home/myname"}
	if err := flag.Parse(args, set, globals); err != nil {
		panic(err)
	}

	// Output:
	// create command invoked
	// input-dir: /home/myname
}
