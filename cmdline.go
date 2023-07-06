package cmdline

import (
	"errors"
	"strings"
)

var (
	// ErrNoArgs is returned by Parse if no arguments were given for parsing.
	ErrNoArgs = errors.New("no arguments")
	// ErrVariadic is returned by Parse to indicate the presence of a
	// variadic option in an option set.
	ErrVariadic = errors.New("variadic option")
	// ErrNoHelp is returned by the standard HelpCommand if no help
	// text is found for the query.
	ErrNoHelp = errors.New("no help for given query")

	// ErrInvoked may be returned by a Command Handler to indicate that the
	// parsing should be stopped after this command, i.e., no sub commands will
	// be further parsed after all the Option arguments were parsed.
	//
	// Use this in situations where an invocation of a command takes over
	// complete program control and there is no need to parse sub commands.
	ErrInvoked = errors.New("command invoked")
)

// Config is the configuration given to Parse.
type Config struct {
	// Args is the arguments to parse. This is usually set to os.Args[1:].
	Args []string
	// Commands is the Commands to parse. Optional.
	Commands Commands
	// Globals is the global Options to parse. Optional.
	Globals Options
	// GlobalsHandler is the handler for Globals.
	// It gets invoked before any Commands handlers. Optional.
	GlobalsHandler Handler
	// LongPrefix is the long Option prefix to use. Optional.
	// Defaults to DefaultLongPrefix if empty.
	LongPrefix string
	// ShortPrefix is the short Option prefix to use. Optional.
	// Defaults to DefaultShortPrefix if empty.
	ShortPrefix string
}

const (
	// DefaultLongPrefix is the default prefix which specifies long option name, e.g.--verbose
	DefaultLongPrefix = "--"
	// DefaultShortPrefix is the default prefix which specifies short option name, e.g. -v
	DefaultShortPrefix = "-"
)

// Parse parses args into specified command set and global options.
// The command set must contain definition of commands to invoke if parsed from
// args and globals flags contain Options that can be parsed before any command
// invocation in args and can be inspected directly, post-parse.
//
// Both the command set and globals are optional and can be nil. If both are nil
// parse will return an error.
// By specifying only the globals Parse behaves much like std/flag package.
//
// Returns one of errors declared in this package, an error from a command
// handler that broke the parse chain or nil if no errors occured.
//
// Args is usually os.Args[1:].
//
// Both set and globals are optional and can be nil but one must not be nil.
//
// .
// Globals will receive options that were specified in args before command
// invocations. It may be in nil in which case an option in args before a
// command invocation will produce an error.
func Parse(config *Config) (err error) {
	if config.LongPrefix == "" {
		config.LongPrefix = DefaultLongPrefix
	}
	if config.ShortPrefix == "" {
		config.ShortPrefix = DefaultShortPrefix
	}
	if len(config.Args) == 0 {
		return ErrNoArgs
	}
	var args = newArguments(config.Args, config.LongPrefix, config.ShortPrefix)
	if err = config.Globals.parse(args); err != nil {
		return
	}
	// if config.GlobalsHandler != nil {
	// 	if err = config.GlobalsHandler(config.Globals); err != nil {
	// 		return err
	// 	}
	// }
	if config.Commands != nil {
		return config.Commands.parse(args)
	}
	return nil
}

// argKind defines the kind of argument being parsed.
type argKind int

const (
	// argNone indicates no argument.
	argNone argKind = iota
	// argLong indicates an argument with a long option prefix.
	argLong
	// argShort indicates an argument with a short option prefix.
	argShort
	// argText indicates a text argument with no prefix.
	argText
)

// arguments wraps a slice of arguments to maintain a state for argument
// itearation and inspection tools.
type arguments struct {
	a     []string
	c     int
	i     int
	long  string
	short string
}

// newArguments wraps in into arguments, sets long and short prefixes to
// recognize and returns it.
func newArguments(in []string, long, short string) *arguments {
	return &arguments{
		a:     in,
		c:     len(in),
		i:     0,
		long:  long,
		short: short,
	}
}

// Raw returns unmodified current argument as given in input slice.
func (self *arguments) Raw() string {
	if self.Eof() {
		return ""
	}
	return self.a[self.i]
}

// Kind returns the current argument kind.
func (self *arguments) Kind() (kind argKind) {
	if self.Eof() {
		return argNone
	}
	kind = argText
	// in case of "-" as short and "--" as long, long wins.
	if strings.HasPrefix(self.Raw(), self.short) {
		kind = argShort
	}
	if strings.HasPrefix(self.Raw(), self.long) {
		kind = argLong
	}
	return
}

// Text returns the current argument as text-only, stripped of prefix, if any.
func (self *arguments) Text() string {
	switch k := self.Kind(); k {
	case argShort:
		return string(self.Raw()[len(self.short):])
	case argLong:
		return string(self.Raw()[len(self.long):])
	case argText:
		return self.Raw()
	}
	return ""
}

// Next advances current argument pointer to the next argument in the wrapped
// arguments and returns self. If no arguments are left to advance to the
// function is a no-op. Use Eof() to check if the arguments are exhausted.
func (self *arguments) Next() *arguments {
	if self.Eof() {
		return self
	}
	self.i++
	return self
}

// FromCurrent returns a slice of wrapped arguments starting from and including
// the current argument. If at EOF an empty slice is returned.
func (self *arguments) FromCurrent() []string { return self.a[self.i:] }

// Eof returns true if current argument index is past argument count.
func (self *arguments) Eof() bool { return self.i >= self.c }

// Count returns the argument count.
func (self *arguments) Count() int { return len(self.a) }
