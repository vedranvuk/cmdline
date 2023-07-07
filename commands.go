package cmdline

import (
	"errors"
	"fmt"
)

// Handler is a Command invocation callback prototype. It carries a Command
// Context which allows inspection of the parse state.
//
// If the handler returns a nil error the parser continues the command chain
// execution.
//
// If the Handler returns a non-nil error the parser aborts the command chain
// execution and the error is pushed back to the Set.Parse() caller.
type Handler func(Context) error

// Context is passed to the Command handler.
// It allows for inspection of the Command's Options.
type Context interface {
	// Parsed returns true if an Option with specified LongName/Name was parsed.
	Parsed(string) bool
	// Value returns the string value that was passed to the Option under
	// specified Name/LongName. Unparsed Options and options that take no
	// arguments return an empty string.
	Value(string) string
}

// NopHandler is an no-op handler that just returns a nil error.
// It can be used as a placeholder to skip command implementation and continue
// the command chain execution.
var NopHandler = func(Context) error { return nil }

// Command defines a command invocable by name.
type Command struct {
	// Name is the name of the Command by which it is invoked from arguments.
	// Command name is required, must not be empty and must be unique in
	// Commands.
	Name string
	// Help is the short Command help text.
	Help string
	// Handler is the function to call when the Command gets invoked from
	// arguments during parsing.
	Handler Handler
	// SubCommands are this Command's sub commands. Command invocation can be
	// chained as described in the Parse method. SubCommands are optional.
	SubCommands Commands
	// Options are this Command's options. Options are optional :|
	Options Options
}

// Commands holds a set of Commands.
type Commands []*Command

// Count returns the number of defined commands in self.
func (self Commands) Count() int { return len(self) }

// Find returns a Command from self by name or nil if not found.
func (self Commands) Find(name string) *Command {
	for i := 0; i < len(self); i++ {
		if self[i].Name == name {
			return self[i]
		}
	}
	return nil
}

// Handle is a helper that registers a new Command from arguments.
// It returns the newly registered command.
func (self Commands) Handle(name, help string, h Handler) (c *Command) {
	c = &Command{
		Name:    name,
		Help:    help,
		Handler: h,
	}
	self.Register(c)
	return
}

// Register is a helper that registers a fully defined Command and returns self.
func (self Commands) Register(command *Command) Commands {
	self = append(self, command)
	return self
}

// parse parses self from config or returns an error.
func (self Commands) parse(config *Config) (err error) {
	switch kind, name := config.Arguments.Kind(config), config.Arguments.Text(config); kind {
	case NoArgument:
		return nil
	case LongArgument, ShortArgument:
		return errors.New("expected command name, got option")
	case TextArgument:
		var cmd = self.Find(name)
		if cmd == nil {
			return fmt.Errorf("command '%s' not registered", name)
		}
		config.Arguments.Next()
		if err = cmd.Options.parse(config); err != nil {
			return
		}
		if err = cmd.Handler(cmd.Options); err != nil {
			return
		}
		return cmd.SubCommands.parse(config)
	}
	return nil
}
