package cmdline

import (
	"context"
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

// Context is passed to the Command handler that allows inspection of
// Command's Options states. It wraps the standard context passed to Config.Parse
// possibly via Parse or ParseCtx.
type Context interface {
	// Context embeds the standard Context.
	// It might carry a timeout, deadline or values available to invoked
	// Command Handler. It will be the context given to ParseCtx or
	// context.Background if Command
	context.Context
	// IsParsed returns true if an Option with specified name was parsed.
	//
	// Options with both Long and Short names use Long names to match against
	// the option name given to this method.
	IsParsed(string) bool
	// RawValues returns an array of raw string values that were passed to the
	// Option under specified Name/LongName. Unparsed Options and options that
	// take no arguments return an empty string.
	RawValues(string) RawValues
	// GetOptions returns this Commands' Options.
	GetOptions() Options
}

// RawValues is a helper alias for a slice of strings representing arguments
// passed to an Option. It implements several utilities for retrieving values.
type RawValues []string

// Count returns number of items in self.
func (self RawValues) Count() int { return len(self) }

// IsEmpty returns true if RawValues are empty.
func (self RawValues) IsEmpty() bool { return len(self) == 0 }

// First returns the first value in self or an empty string if empty.
func (self RawValues) First() string {
	if len(self) > 0 {
		return self[0]
	}
	return ""
}

// NopHandler is an no-op handler that just returns a nil error.
// It can be used as a placeholder to skip Command Handler implementation and
// allow for the command chain execution to continue.
var NopHandler = func(Context) error { return nil }

// Command defines a command invocable by name.
type Command struct {
	// Name is the name of the Command by which it is invoked from arguments.
	// Command name is required, must not be empty and must be unique in
	// Commands.
	Name string
	// Help is the short Command help text that should prefferably fit
	// the width of a standard terminal.
	Help string
	// Handler is the function to call when the Command gets invoked from
	// arguments during parsing.
	Handler Handler
	// SubCommands are this Command's sub commands. Command invocation can be
	// chained as described in the Parse method. SubCommands are optional.
	SubCommands Commands
	// Options are this Command's options. Options are optional :|
	Options Options

	// ExclusivityGroups are the exclusivity groups for this Command's Options.
	// If more than one Option from an ExclusivityGroup is passed in arguments
	// Parse/ParseCtx will return an error.
	//
	// If no ExclusivityGroups are defined no checking is performed.
	ExclusivityGroups ExclusivityGroups

	// executed is true if the command was parsed from arguments.
	executed bool
}

// Commands holds a set of Commands.
// Commands are never sorted and the order in which they are declared is
// important to the Print function which prints the Commands in the same order.
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

// VisitCommandFunc is a prototype of a function called for each Command visited.
// It must return true to continue enumeration.
type VisitCommandFunc func(c *Command) bool

// Walk calls f for each command in self, recursively.
func (self Commands) Walk(f VisitCommandFunc) { walkCommands(self, f, true, true) }

// WalkExecuted calls f for each executed command in self, recursively.
func (self Commands) WalkExecuted(f VisitCommandFunc) { walkCommands(self, f, true, false) }

// WalkNotExecuted calls f for each not executed command in self, recursively.
func (self Commands) WalkNotExecuted(f VisitCommandFunc) { walkCommands(self, f, false, true) }

// walk walks c calling f for each.
func walkCommands(c Commands, f func(c *Command) bool, executed, notexecuted bool) {
	for _, cmd := range c {
		if cmd.executed && executed || !cmd.executed && notexecuted {
			if !f(cmd) {
				break
			}
		}
		if cmd.SubCommands.Count() > 0 {
			walkCommands(cmd.SubCommands, f, executed, notexecuted)
		}
	}
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
		if err = validateCommandExclusivityGroups(cmd); err != nil {
			return
		}
		var wrapper = &contextWrapper{
			config.context,
			cmd.Options,
		}
		cmd.executed = true
		if err = cmd.Handler(wrapper); err != nil {
			return
		}
		return cmd.SubCommands.parse(config)
	}
	return nil
}

// contextWrapper wraps the standard Context and Options to imlement
// cmdline.Context.
type contextWrapper struct {
	context.Context
	Options
}
