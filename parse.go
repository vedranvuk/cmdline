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
func Parse(args []string, set *Set, globals *Options) (err error) {
	var t = newTokens(args, LongPrefix, ShortPrefix)
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

func (o *Options) parse(tok *tokens) error {
	var opt *option
	for {
		opt = nil
		// Parse option key and value, find option by key.
		var key, val, assignment = strings.Cut(tok.Text(), "=")
		if assignment {
			if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
				val = strings.TrimPrefix(strings.TrimSuffix(val, "\""), "\"")
			}
		}
		switch kind := tok.Kind(); kind {
		case argText:
			opt = o.getNextUnparsedIndexed()
		case argLong:
			if opt = o.get(key, ""); opt == nil {
				return fmt.Errorf("option --'%s' not registered", key)
			}
		case argShort:
			if opt = o.get("", key); opt == nil {
				return fmt.Errorf("option -'%s' not registered", key)
			}
		}

		if opt == nil {
			if opt = o.getVariadic(); opt == nil {
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
			opt.value = strings.Join(tok.FromCurrent(), " ")
		}

		opt.parsed = true

		if opt.kind == variadicOption {
			break
		}

		tok.Next()
	}

	for _, opt = range o.options {
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

func (o *Options) get(long, short string) *option {
	if long != "" {
		for _, v := range o.options {
			if v.long == long {
				return v
			}
		}
	}
	if short != "" {
		for _, v := range o.options {
			if v.short == short {
				return v
			}
		}
	}
	return nil
}

func (o *Options) getNextUnparsedIndexed() *option {
	for _, v := range o.options {
		if v.kind == indexedOption && v.parsed == false {
			return v
		}
	}
	return nil
}

func (o *Options) getVariadic() *option {
	for _, v := range o.options {
		if v.kind == variadicOption {
			return v
		}
	}
	return nil
}

func (s *Set) parse(t *tokens) error {
	switch kind, name := t.Kind(), t.Text(); kind {
	case argNone:
		return nil
	case argLong, argShort:
		return errors.New("expected command name, got option")
	case argText:
		cmd, ok := s.cmds[name]
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

type argKind int

const (
	argNone argKind = iota
	argLong
	argShort
	argText
)

type tokens struct {
	a     []string
	c     int
	i     int
	long  string
	short string
}

func newTokens(in []string, long, short string) *tokens {
	return &tokens{
		a:     in,
		c:     len(in),
		i:     0,
		long:  long,
		short: short,
	}
}

func (t *tokens) Raw() string {
	if t.Eof() {
		return ""
	}
	return t.a[t.i]
}

func (t *tokens) Kind() (kind argKind) {
	if t.Eof() {
		return argNone
	}
	kind = argText
	// in case of "-" as short and "--" as long, long wins.
	if strings.HasPrefix(t.Raw(), t.short) {
		kind = argShort
	}
	if strings.HasPrefix(t.Raw(), t.long) {
		kind = argLong
	}
	return
}

func (t *tokens) Text() string {
	switch k := t.Kind(); k {
	case argShort:
		return string(t.Raw()[len(t.short):])
	case argLong:
		return string(t.Raw()[len(t.long):])
	case argText:
		return t.Raw()
	}
	return ""
}

func (t *tokens) Next() *tokens {
	if t.Eof() {
		return t
	}
	t.i++
	return t
}

func (t *tokens) FromCurrent() []string { return t.a[t.i:] }

func (t *tokens) Eof() bool { return t.i >= t.c }