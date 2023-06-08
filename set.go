package cmdline

import (
	"fmt"
)

const (
	// LongPrefix is the prefix which specifies long option name, e.g.--verbose
	LongPrefix = "--"
	// ShortPrefix is the prefix which specifies short option name, e.g. -v
	ShortPrefix = "-"
)


// New returns a new parse Set.
func New() *Set { return &Set{make(map[string]*Command)} }

// Set is a parse set that contains command and flag definitions and the
// post-parse state inspectable by handlers via their context.
type Set struct {
	cmds map[string]*Command
}

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


