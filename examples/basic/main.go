package main

import "github.com/vedranvuk/cmdline"

func cmdGlobal(ctx cmdline.Context) error {
	return nil
}

func cmdTest(ctx cmdline.Context) error {
	return nil
}

func main() {
	var commands = cmdline.NewCommands(nil)

	commands.MustAdd("", "Global flags", cmdGlobal).Parameters().
		MustAddNamed("verbose", "v", "Verbose output.", false, nil).Parent().Parent().
		MustAdd("test", "", cmdTest)

	var err error
	if err = cmdline.ParseRaw(commands, "--verbose", "test"); err != nil {
		panic(err)
	}
}
