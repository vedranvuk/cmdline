package cmdline

import "fmt"

// Handler is a command handler. It receives a context to inspect parse state.
// If the Handler returns an error the execution chain is aborted and the error
// pushed back to the Set.Parse() caller.
type Handler func(Context) error

// NoHandler is a placeholder handler that does nothing.
var NoHandler = func(Context) error { return nil }

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
	help string
	set  *CommandSet
	opts *OptionSet
}

// Options returns the commands options.
func (self *Command) Options() *OptionSet { return self.opts }

// Sub returns this command's subset of commands which can be invoked on this
// command.
func (self *Command) Sub() *CommandSet {
	if self.set == nil {
		self.set = NewCommandSet()
	}
	return self.set
}

// CommandSet is a parse set that contains command and flag definitions and the
// post-parse state inspectable by handlers via their context.
type CommandSet struct {
	cmds map[string]*Command
}

// NewCommandSet returns a new parse Set.
func NewCommandSet() *CommandSet { return &CommandSet{make(map[string]*Command)} }

// Handle registers a command handler f under specified name and returns the
// newly defined command.
func (self *CommandSet) Handle(name, help string, h Handler) (c *Command) {
	if name == "" {
		panic("command name must not be empty")
	}
	if _, exists := self.cmds[name]; exists {
		panic(fmt.Sprintf("command '%s' already registered", name))
	}
	if h == nil {
		panic(fmt.Sprintf("command '%s' nil registering nil handler", name))
	}
	c = &Command{h: h, help: help, opts: &OptionSet{}}
	self.cmds[name] = c
	return
}

// Count returns number of defined commands.
func (self *CommandSet) Count() int { return len(self.cmds) }
