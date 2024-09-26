// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
Package cmdline implements a handler based command parsing from executable
arguments designed for hierarchical command based invocations, with options.

# Command line option syntax

The syntax mimics GNU-like long/short combination style of option names.

For example:

	--verbose	(long form)
	-v 			(short form)

If an option takes a value it must be assigned using "=". If it contains spaces
it must be enclosed in double quotes.

For example:

	--input-directory=/home/yourname
	--title="Adriatic Coasting"

Commands can be invoked by name, they can have options and have sub commands
that have options. So you can have things like:

	// Invoke an executable with global flag "-v" executing command "create"
	// with options "output-dir" taking a value and an indexed option
	// "My Project".
	tool.exe -v create --output-dir=/home/yourname/projects "My Project"

	// Invoke an executable with global flag "--color=auto" executing
	// sub-command "name" with indexed argument "newname" on command "change".
	ip.exe --color=auto change name "newname"

# Usage

Create a new command Set and register a command named "new" then add all four
option forms to the command as an example:

	set := New()
	set.Handle("new", func(c Context) error {
		fmt.Printf("Create a new project %s\n", c.Value("name"))
		return nil
	}).Options().
		Required("target-directory", "t", "Specify target directory.").
		Option("force-overwrite", "f", "Force overwriting target files.").
		IndexedRequired("name", "Name of the project.").
		Indexed("author", "Project author")

Create a set of global options that do not apply to any defined command.

	globals := Globals().
		Option("verbose", "v", "Be more verbose.").
		Option("help", "h", "Show help.")

Parse the program arguments into our command set and global options and handle
any errors returned.

	if err := Parse(os.Args[1:], set, globals); err != nil {
		if errors.Is(err, ErrNoArguments) {
			Usage()
		} else {
			fmt.Printf("invalid input: %v\n", err)
		}
		os.Exit(1)
	}
*/
package cmdline
