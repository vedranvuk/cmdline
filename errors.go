package cmdline

import "errors"

var (
	// ErrNoArgs is returned by Parse if no arguments were given for parsing.
	ErrNoArgs = errors.New("no arguments")

	// ErrVariadic is returned by Parse to indicate the presence of a
	// variadic option in an option set.
	ErrVariadic = errors.New("variadic option")

	// ErrInvoked may be returned by a Command Handler to indicate that the
	// parsing should be stopped after this command, i.e., no sub commands will 
	// be further parsed after all the Option arguments were parsed.
	//
	// Use this in situations where an invocation of a command takes over 
	// complete program control and there is no need to parse sub commands.
	ErrInvoked = errors.New("command invoked")
)
