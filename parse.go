package cmdline

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// LongPrefix is the prefix which specifies long option name, e.g.--verbose
	LongPrefix = "--"
	// ShortPrefix is the prefix which specifies short option name, e.g. -v
	ShortPrefix = "-"
)

var (
	// ErrNoArgs is returned when no arguments were given for parsing.
	ErrNoArgs = errors.New("no arguments")

	// ErrVariadic is returned by options parser to indicate the presence of a
	// variadic option in an option set.
	ErrVariadic = errors.New("variadic option")
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
func Parse(args []string, set *CommandSet, globals *OptionSet) (err error) {
	var t = newArguments(args, LongPrefix, ShortPrefix)
	if globals != nil {
		if err = globals.parse(t); err != nil {
			return
		}
	}
	if set != nil {
		return set.parse(t)
	}
	return nil
}

// parse parses self from args or returns an error.
func (self *OptionSet) parse(args *arguments) error {
	var opt *option
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

		switch opt.kind {
		case requiredOption:
			if !assignment {
				return fmt.Errorf("required option '%s' requires a value", opt.long)
			}
			fallthrough
		case optionalOption:
			opt.value = val
		case indexedOption:
			opt.value = key
		case variadicOption:
			opt.value = strings.Join(args.FromCurrent(), " ")
		}

		opt.parsed = true

		if opt.kind == variadicOption {
			break
		}

		args.Next()
	}

	for _, opt = range self.options {
		if !opt.parsed {
			if opt.kind == requiredOption {
				return fmt.Errorf("required option '%s' not parsed", opt.long)
			}
			if opt.kind == indexedOption {
				return fmt.Errorf("indexed option '%s' not parsed", opt.long)
			}
		}
	}

	return nil
}

func (self *OptionSet) get(long, short string) *option {
	if long != "" {
		for _, v := range self.options {
			if v.long == long {
				return v
			}
		}
	}
	if short != "" {
		for _, v := range self.options {
			if v.short == short {
				return v
			}
		}
	}
	return nil
}

func (self *OptionSet) getNextUnparsedIndexed() *option {
	for _, v := range self.options {
		if v.kind == indexedOption && v.parsed == false {
			return v
		}
	}
	return nil
}

func (self *OptionSet) getVariadic() *option {
	for _, v := range self.options {
		if v.kind == variadicOption {
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
