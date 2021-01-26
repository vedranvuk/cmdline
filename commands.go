// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"fmt"
)

// Command is a command definition.
type Command struct {
	parent     *Commands   // parent is parent Commands.
	name, help string      // name and help is command name and help text.
	handler    Handler     // handler is the command handler.
	parameters *Parameters // parameters are this Command's Parameters.
	commands   *Commands   // commands are this Command's sub Commands.
}

// NewCommand returns a new Command instance with specified optional help and
// handler optional if raw is false.
// If raw is true command's handler will receive unparsed parameters for custom
// handling.
func NewCommand(parent *Commands, name, help string, handler Handler) *Command {
	var command = &Command{
		parent:  parent,
		name:    name,
		help:    help,
		handler: handler,
	}
	command.parameters = NewParameters(command)
	command.commands = NewCommands(command)
	return command
}

// Parent returns the Command's parent Commands.
func (c *Command) Parent() *Commands { return c.parent }

// Execute executes handler with specified context and returns its result.
// If no handler is registered an ErrNoHandler descendant is returned.
func (c *Command) Execute(ctx Context) error {
	if c.handler == nil {
		return fmt.Errorf("%w: %s", ErrNoHandler, c.name)
	}
	return c.handler(ctx)
}

// Reset resets command parameters and  sub commands recursively.
func (c *Command) Reset() {
	c.parameters.Reset()
	c.commands.Reset()
}

// Name returns Command name.
func (c *Command) Name() string { return c.name }

// Help returns Command help.
func (c *Command) Help() string { return c.help }

// Handler returns Command Handler.
func (c *Command) Handler() Handler { return c.handler }

// Parameters returns command Parameters.
func (c *Command) Parameters() *Parameters { return c.parameters }

// Commands returns Command's Commands.
func (c *Command) Commands() *Commands { return c.commands }

// nameToCommand is a map of command name to *Command.
type nameToCommand map[string]*Command

// Commands holds a set of Commands with a unique name.
type Commands struct {
	// parent is this Command's parent.
	// If it is a Parser these Commands are the root Commands.
	// If it is a Command these are Command's sub commands.
	parent *Command
	// commandmap is a map of command names to *Command definitions.
	commandmap nameToCommand
	// nameindexes is a slice of command names in order as they were defined.
	nameindexes []string
}

// NewCommands returns a new Commands instance with specified parent which can
// be nil.
func NewCommands(parent *Command) *Commands {
	return &Commands{
		parent:     parent,
		commandmap: make(nameToCommand),
	}
}

// Parent returns the parent Command.
func (c *Commands) Parent() *Command { return c.parent }

// Reset resets all registered commands and their subs recursively.
func (c *Commands) Reset() {
	var cmd *Command
	for _, cmd = range c.commandmap {
		cmd.Reset()
	}
}

// CommandCount returns number of registered commands.
func (c *Commands) Length() int { return len(c.commandmap) }

// Print prints Commands as a structured text suitable for terminal display.
func (c *Commands) Print() string { return PrintCommands(c) }

// AddCommand registers a new Command under specified name and help text that
// invokes handler when parsed from arguments.
//
// Command with an empty name allows for passing just parameters to Commands set
// when parsing and is executed in parallel with another named command in
// same Commands set.
//
// Order of registration is important. When printed, Commands are listed in the
// order they were registered instead of name sorted.
//
// If an error occurs Command will not be registered and resulting Command 
// will be nil with error being an ErrRegister descendant.
func (c *Commands) Add(name, help string, handler Handler) (*Command, error) {
	var ok bool
	if _, ok = c.commandmap[name]; ok {
		if name == "" {
			return nil, fmt.Errorf("%w: duplicate empty command", ErrRegister)
		}
		return nil, fmt.Errorf("%w: duplicate command: '%s'", ErrRegister, name)
	}
	if c.parent != nil {
		if c.parent.Parameters().HasOptionalIndexedParameters() {
			return nil, fmt.Errorf("%w: command with raw parameters cannot have sub commands", ErrRegister)
		}
	}
	var cmd = NewCommand(c, name, help, handler)
	c.commandmap[name] = cmd
	c.nameindexes = append(c.nameindexes, name)
	return cmd, nil
}

// MustAddCommand is like AddCommand except the function panics on error.
// Returns added *Command.
func (c *Commands) MustAdd(name, help string, f Handler) *Command {
	var cmd, err = c.Add(name, help, f)
	if err != nil {
		panic(err)
	}
	return cmd
}

// GetCommand returns a *Command by name if found and truth if found.
func (c *Commands) Get(name string) (cmd *Command, ok bool) {
	cmd, ok = c.commandmap[name]
	return
}

// MustGetCommand is like GetCommand but panics if Command is not found.
func (c *Commands) MustGet(name string) *Command {
	var cmd *Command
	var ok bool
	if cmd, ok = c.commandmap[name]; ok {
		return cmd
	}
	panic(fmt.Sprintf("commandline: command '%s' not found", name))
}

// Parse parses registered commands from arguments and stores matches in
// optional chain. If an error occurs it is returned and Commands were not
// fully and successfully parsed.
//
// Parse expects next argument to be a TextArgument specifying command name.
// If it is a parameter Commands checks if an unnamed command is registered
// and passes arguments to that command's Parameters. 
//
// Result may be one of the following sorted by precedence:
// 	ErrNoDefinitions if no commands are defined.
// 	ErrNoArguments if args are empty.
// 	ErrInvalidArgument if next argument is InvalidArgument.
// 	ErrCommandNotFound if command is not found.
// 	ErrExtraArguments if there are unparsed arguments left.
// 	ErrParse descendants describing specific parse errors.
func (c *Commands) Parse(args Arguments, chain Chain) error {
	if c.Length() == 0 {
		return ErrNoDefinitions
	}
	var err error
	var cmd *Command
	var ok, global bool
	switch args.Kind() {
	case InvalidArgument:
		return fmt.Errorf("%w: '%s'", ErrInvalidArgument, args.Raw())
	case NoArgument:
		return ErrNoArguments
	case TextArgument:
		if cmd, ok = c.commandmap[args.Name()]; !ok {
			return fmt.Errorf("%w: '%s'", ErrCommandNotFound, args.Name())
		}
	default:
		if cmd, ok = c.commandmap[""]; !ok {
			return ErrCommandNotFound
		}
		global = true
	}
	// A command was matched at this point.
	// Append command to chain if given.
	if chain != nil {
		chain.Add(cmd)
	}
	// If an empty command is registered and argument was a param
	// threat it as a param to that command. Don't skip it so it can
	// be parsed py Paremeters.
	if !global {
		args.Advance()
	}
	// Parse Parameters.
	// If all required parameters parsed returns nil.
	// If no registered parameters returns ErrNoDefinitions.
	// If no arguments left in state returns ErrNoArguments.
	// If parameter not found returns ErrNotFound.
	// If paremeter repeats returns ErrDuplicateParameter.
	// Returns other parse specific errors.
	if err = cmd.Parameters().Parse(args); err != nil {
		if !errors.Is(err, ErrNoArguments) && !errors.Is(err, ErrNoDefinitions) {
			return err
		}
	}
	// Repeat parse on these Commands so that next command
	// invocation in arguments is passed back to these Commands.
	// Empty command is executed in parallel to other commands.
	if global {
		err = c.Parse(args, chain)
	} else {
		err = cmd.Commands().Parse(args, chain)
	}
	// Pass control to contained Commands to continue chaining.
	// They will return ErrNotFound or a descendant if
	// No commands or parameters were matched.
	if err != nil {
		if errors.Is(err, ErrNoArguments) || errors.Is(err, ErrNoDefinitions) {
			if args.Length() > 0 {
				return ErrExtraArguments
			}
			return nil
		}
	}
	return err
}
