// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"context"
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

// NopHandler is an no-op handler that just returns a nil error.
// It can be used as a placeholder to skip Command Handler implementation and
// allow for the command chain execution to continue.
var NopHandler = func(Context) error { return nil }

// HelpHandler is a utility handler that prints the current configuration.
var HelpHandler = func(c Context) error {

	if config := c.Config(); config != nil {
		config.PrintUsage()
		PrintConfig(config.GetOutput(), c.Config())
	}

	return nil
}

// HelpCommand is a utility function that returns a command that handles "help"
// using HelpHandler.
func HelpCommand() *Command {
	return &Command{
		Name:                "help",
		Help:                "Prints out the command usage.",
		RequireSubExecution: false,
		Handler:             HelpHandler,
	}
}

// Context is passed to the Command handler that allows inspection of
// Command's Options states. It wraps the standard context passed to Config.Parse
// possibly via Parse or ParseCtx.
type Context interface {

	// Context embeds the standard Context.
	//
	// It might carry a timeout, deadline or values available to invoked
	// Command Handler. It will be the context given to ParseCtx or
	// context.Background if Parse was used.
	context.Context

	// Parsed returns true if an Option with specified LongName was parsed.
	Parsed(string) bool

	// Values returns an array of strings that were passed to the
	// Option under specified LongName.
	//
	// Unparsed Options and options that take no arguments return nil.
	Values(string) Values

	// Config returns the config that is being used to parse.
	Config() *Config

	// Command returns the owner Command of this handler.
	//
	// If this handler is the global options handler result will be nil.
	Command() *Command

	// ParentCommand returns the parent command of this handler's command.
	//
	// It may return nil if this command has no parent command or if this
	// handler is the global options handler.
	ParentCommand() *Command

	// Options returns this Command's Options.
	Options() Options
}

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

	// RequireSubExecution if true, will raise an error if none of this
	// Command's SubCommands were executed. The setting is ignored if Command
	// has no SubCommands defined.
	// Defaults to false.
	RequireSubExecution bool

	// ExclusivityGroups are the exclusivity groups for this Command's Options.
	// If more than one Option from an ExclusivityGroup is passed in arguments
	// Parse/ParseCtx will return an error.
	//
	// If no ExclusivityGroups are defined no checking is performed.
	ExclusivityGroups ExclusivityGroups

	// executed is true if the command was parsed from arguments.
	executed bool
}

// ExclusivityGroup defines a group of option names which are mutually
// exclusive and may not be passed together to a command at the same time.
type ExclusivityGroup []string

// ExclusivityGroups is a group of ExclusivityGroup used to define more than
// one ExclusivityGroup.
type ExclusivityGroups []ExclusivityGroup

// Commands holds a set of Commands.
// Commands are never sorted and the order in which they are declared is
// important to the Print function which prints the Commands in the same order.
type Commands []*Command

// Count returns the number of defined commands in self.
func (self Commands) Count() int { return len(self) }

// Register is a helper that registers a fully defined Command and returns self.
func (self *Commands) Register(command *Command) Commands {
	*self = append(*self, command)
	return *self
}

// Handle is a helper that registers a new Command from arguments.
// It returns the newly registered command.
func (self *Commands) Handle(name, help string, h Handler) (c *Command) {
	c = &Command{
		Name:    name,
		Help:    help,
		Handler: h,
	}
	self.Register(c)
	return
}

// Find returns a Command from self by name or nil if not found.
func (self Commands) Find(name string) *Command {
	for i := 0; i < len(self); i++ {
		if self[i].Name == name {
			return self[i]
		}
	}
	return nil
}

// AnyExecuted returns true if any commands in Commands was executed.
func (self Commands) AnyExecuted() bool {
	for _, command := range self {
		if command.executed {
			return true
		}
	}
	return false
}

// Reset resets all command and command otpion states recursively.
func (self Commands) Reset() {
	for _, command := range self {
		command.executed = false
		command.Options.Reset()
		command.SubCommands.Reset()
	}
}

// VisitCommand is a prototype of a function called for each Command visited.
// It must return true to continue enumeration.
type VisitCommand func(c *Command) bool

// Walk calls f for each command in self, recursively.
// Order is top-down, i.e. parents first.
func (self Commands) Walk(f VisitCommand) { walkCommands(self, f, true, true) }

// WalkExecuted calls f for each executed command in self, recursively.
// Order is top-down, i.e. parents first.
func (self Commands) WalkExecuted(f VisitCommand) { walkCommands(self, f, true, false) }

// WalkNotExecuted calls f for each not executed command in self, recursively.
// Order is top-down, i.e. parents first.
func (self Commands) WalkNotExecuted(f VisitCommand) { walkCommands(self, f, false, true) }

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
