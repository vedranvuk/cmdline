// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

// Handler is a prototype of a function that handles the event of a
// Command being parsed from command line arguments. A command handler.
//
// CommandFuncs of all Commands in the chain parsed on command line are
// visited during Parser.Parse(). Only the last matched command is marked
// as executed and can be discerned from visited CommandFuncs using
// Context.Executed().
//
// If a Handler returns a non-nil error further calling of handlers of
// Commands parsed from command line arguments is aborted and the error is
// propagated to Parse method and returned.
type Handler = func(Context) error

// Context is a Command Handler context.
type Context interface {
	// Name returns the name of Command that registered this handler.
	Name() string
	// Print prints the calling Command definition and any of its Commands as
	// structured text suitable for terminal display.
	Print() string
	// Parsed returns true if command's parameter under specified name is defined
	// and parsed from command line and false otherwise.
	Parsed(string) bool
	// Value returns the raw string argument given to a parameter under
	// specified name. If parameter was not parsed or is not registered
	// an empty string is returned.
	RawValue(string) string
	// Next returns next related context.
	// Depends on context usage.
	// Can be nil.
	Next() Context
	// Extra returns any arguments to Handler, depending on Parse function.
	Extra() []string
}

// NewContext returns a new Context instance.
func NewContext(command *Command, next *context, arguments []string) Context {
	return &context{
		Command:   command,
		next:      next,
		arguments: arguments,
	}
}

// context implements Context.
type context struct {
	*Command           // cmd is handler's command.
	next      *context // next points to next matched command's context.
	arguments []string // arguments to pass to handler.
}

// Arguments implements Context.Arguments.
func (c *context) Extra() []string { return c.arguments }

// Name implements Context.Name.
func (c *context) Name() string { return c.Command.Name() }

// Print implements Context.Print.
func (c *context) Print() string { return c.Commands().Print() }

// Value implements Context.Value.
func (c *context) RawValue(name string) string {
	var param, err = c.Parameters().Get(name)
	if err != nil {
		return ""
	}
	if !param.Parsed() {
		return ""
	}
	return param.RawValue()
}

// Parsed implements Context.Parsed.
func (c *context) Parsed(name string) bool {
	var param, err = c.Parameters().Get(name)
	if err != nil {
		return false
	}
	return param.Parsed()
}

// Next implements Context.Next.
func (c *context) Next() Context { return c.next }

// ExecuteHandler calls the context's command handler and propagates its result.
// If command has no handler result is ErrNoHandler.
func (c *context) ExecuteHandler() error { return c.Execute(c) }
