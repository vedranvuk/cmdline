package cmdline

// Handler is a command handler. It receives a context to inspect parse state.
// If the Handler returns an error the execution chan is aborted and the error
// pushed back to the Set.Parse() caller.
type Handler func(Context) error

// Context is passed to the Command handler.
type Context interface {
	// Parsed returns truth if a flag with specified long name was parsed.
	Parsed(string) bool
	// Value returns value of the flag with specified name.
	// Unparsed flags return an empty value. Use Parsed to check validity.
	Value(string) string
}

// Command defines an invokeable command.
type Command struct {
	h    Handler
	set  *Set
	opts *Options
}

// Options returns the commands options.
func (c *Command) Options() *Options { return c.opts }

// Sub returns this command's subset of commands which can be invoked on this
// command.
func (c *Command) Sub() *Set {
	if c.set == nil {
		c.set = New()
	}
	return c.set
}
