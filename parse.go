package cmdline

import (
	"errors"
	"fmt"
	"strings"
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
	if config.Globals != nil {
		if err = config.Globals.parse(args); err != nil {
			return
		}
	}
	if config.Commands != nil {
		return config.Commands.parse(args)
	}
	return nil
}

// Config is the configuration given to Parse.
type Config struct {
	// Args is the arguments to parse. This is usually set to os.Args[1:].
	Args []string
	// Commands is the CommandSet to parse. Optional.
	Commands *CommandSet
	// Globals is the global OptionSet to parse. Optional.
	Globals *OptionSet
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

// parse parses self from args or returns an error.
func (self *OptionSet) parse(args *arguments) error {
	var opt Option
For:
	for {
		opt = nil

		// Parse out and clean option key and its optional value.
		var key, val, assignment = strings.Cut(args.Text(), "=")
		key = strings.TrimSpace(key)
		if assignment {
			val = strings.TrimSpace(val)
			// naive dequoting: remove leading and trailing quotes if paired.
			if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
				val = strings.TrimPrefix(strings.TrimSuffix(val, "\""), "\"")
			}
			val = strings.TrimSpace(val)
		}

		switch kind := args.Kind(); kind {
		case argText:
			opt = self.getNextUnparsedIndexed()
		case argLong:
			if opt = self.get(key, ""); opt == nil {
				return fmt.Errorf("option --'%s' not registered", key)
			}
		case argShort:
			if opt = self.get("", key); opt == nil {
				return fmt.Errorf("option -'%s' not registered", key)
			}
		}

		if opt == nil {
			if opt = self.getVariadic(); opt == nil {
				break
			}
		}

		switch o := opt.(type) {
		case *Boolean:
			o.parsed = true
		case *Optional:
			o.value = val
			o.parsed = true
		case *Required:
			if !assignment {
				return fmt.Errorf("required option '%s' requires a value", o.Key())
			}
			o.value = val
			o.parsed = true
		case *Indexed:
			o.value = key
			o.parsed = true
		case *Variadic:
			o.value = strings.Join(args.FromCurrent(), " ")
			o.parsed = true
			break For
		}

		args.Next()
	}

	// Check required options are parsed.
	for _, opt = range self.options {
		if !opt.Parsed() {
			if _, ok := opt.(*Required); ok {
				return fmt.Errorf("required option '%s' not parsed", opt.Key())
			}
			if _, ok := opt.(*Indexed); ok {
				return fmt.Errorf("indexed option '%s' not parsed", opt.Key())
			}
		}
	}

	return nil
}

func (self *OptionSet) get(long, short string) Option {
	if long != "" {
		for _, v := range self.options {
			if v.Key() == long {
				return v
			}
		}
	}
	if short != "" {
		for _, v := range self.options {
			if v.ShortKey() == short {
				return v
			}
		}
	}
	return nil
}

func (self *OptionSet) getNextUnparsedIndexed() Option {
	for _, v := range self.options {
		if _, ok := v.(Indexed); ok && !v.Parsed() {
			return v
		}
	}
	return nil
}

func (self *OptionSet) getVariadic() Option {
	for _, v := range self.options {
		if _, ok := v.(*Variadic); ok {
			return v
		}
	}
	return nil
}

// parse parses t into s.
func (self *CommandSet) parse(t *arguments) error {
	switch kind, name := t.Kind(), t.Text(); kind {
	case argNone:
		return nil
	case argLong, argShort:
		return errors.New("expected command name, got option")
	case argText:
		cmd, ok := self.cmds[name]
		if !ok {
			return fmt.Errorf("command '%s' not registered", name)
		}
		t.Next()
		if err := cmd.opts.parse(t); err != nil {
			return err
		}
		if err := cmd.h(cmd.opts); err != nil {
			return err
		}
		if cmd.Sub().Count() > 0 {
			return cmd.Sub().parse(t)
		}
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
