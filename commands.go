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
	// Parsed returns truth if an Option with specified Key was parsed.
	Parsed(string) bool
	// Value returns value of the Option with specified name.
	// Unparsed Options return an empty value.
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

// Parsed implements Context.Parsed for handlers of commands with no Options.
// It is a no-op placeholder that returns false.
func (self Command) Parsed(string) bool { return false }

// Value implements Context.Value for handlers of commands with no Options.
// It is a no-op that returns an empty string.
func (self Command) Value(string) string { return "" }

// Commands holds a set of Commands.
type Commands []*Command

// Handle registers a new Command from specified name, shortHelp, longHelp and
// Handler h and returns it. If the registration fails the function panics.
func (self Commands) Handle(name, help string, h Handler) (c *Command) {
	c = &Command{
		Name:    name,
		Help:    help,
		Handler: h,
	}
	self.Register(c)
	return
}

// Register registers a fully defined Command and returns self. If the
// registration fails the function panics.
func (self Commands) Register(command *Command) Commands {
	if command.Name == "" {
		panic("command name is empty")
	}
	for i := 0; i < len(self); i++ {
		if self[i].Name == command.Name {
			panic(fmt.Sprintf("command '%s' already registered", command.Name))
		}
	}
	if command.Handler == nil {
		panic(fmt.Sprintf("command '%s' nil registering nil handler", command.Name))
	}
	self = append(self, command)
	return self
}

// Count returns the number of defined commands in these Commands.
func (self Commands) Count() int { return len(self) }

// Get returns a Command by name or nil if not found.
func (self Commands) Get(name string) *Command {
	for i := 0; i < len(self); i++ {
		if self[i].Name == name {
			return self[i]
		}
	}
	return nil
}

func (self Commands) parse(t *arguments) error {
	switch kind, name := t.Kind(), t.Text(); kind {
	case argNone:
		return nil
	case argLong, argShort:
		return errors.New("expected command name, got option")
	case argText:
		cmd := self.Get(name)
		if cmd == nil {
			return fmt.Errorf("command '%s' not registered", name)
		}
		t.Next()
		if cmd.Options != nil {
			if err := cmd.Options.parse(t); err != nil {
				return err
			}
			if err := cmd.Handler(cmd.Options); err != nil {
				return err
			}
		} else {
			if err := cmd.Handler(cmd); err != nil {
				return err
			}
		}
		if cmd.SubCommands != nil && cmd.SubCommands.Count() > 0 {
			return cmd.SubCommands.parse(t)
		}
	}
	return nil
}
