package cmdline

import "fmt"

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

// Set is a parse set that contains command and flag definitions and the
// post-parse state inspectable by handlers via their context.
type Set struct {
	cmds map[string]*Command
}

// New returns a new parse Set.
func New() *Set { return &Set{make(map[string]*Command)} }

// Handle registers a command handler f under specified name and returns the
// newly defined command.
func (s *Set) Handle(name string, h Handler) (c *Command) {
	if name == "" {
		panic("command name must not be empty")
	}
	if _, exists := s.cmds[name]; exists {
		panic(fmt.Sprintf("command '%s' already registered", name))
	}
	if h == nil {
		panic(fmt.Sprintf("command '%s' nil registering nil handler", name))
	}
	c = &Command{h: h, opts: &Options{}}
	s.cmds[name] = c
	return
}

// Count returns number of defined commands.
func (s *Set) Count() int { return len(s.cmds) }
