package cmdline

import (
	"fmt"
	"strings"
)

// Option defines an option for a Command.
type Option interface {
	// Key returns the Option Key by which Option is addressed from
	// Command Context. Option Key must be unique in an Options.
	//
	// Options with both long and short names map their long name as the Key.
	Key() string
	// ShortKey returns the short form of the Option key. This is used only by
	// Boolean, Optional and Required options which can be addressed from
	// arguments with a shorter, single character option name.
	ShortKey() string
	// Parsed returns true if the Option was parsed from the arguments.
	Parsed() bool
	// Value returns parsed Option Value. Result may be an empty string if the
	// Option was not parsed or takes no argument(s).
	Value() string
}

// Options contains and manages a set of Options.
//
// An Options is used either as a set of Options for a Command or for Global
// Options which may be parsed before any Command invocation and are inspected
// separately instead of through a Command Context as they apply to no specific
// Command. For more info see Parse function.
type Options struct {
	options []Option
	values  map[string]any
}

// NewOptions returns a new Options instance.
func NewOptions() *Options { return &Options{nil, make(map[string]any)} }

// Boolean defines a boolean option.
//
// A Boolean option is an option that takes no argument and is simply marked as
// parsed if found in arguments.
type Boolean struct {
	// LongName is the long, more descriptive option name. It must contain no
	// spaces and must be unique in an Options. It is used as the key to
	// retrieve this option state from its Command context.
	// An argument with a long prefix is matched against this property.
	LongName string
	// ShortName is the short option name consisting of a single alphanumeric.
	// An argument with a short prefix is matched against this property.
	ShortName string
	// Help is the short help text.
	Help string

	// option embeds the base option properties.
	option
}

// Flag defines an option that is not required. It takes no arguments.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (self *Options) Boolean(longName, shortName, help string) *Options {
	return self.Register(&Boolean{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
	})
}

// Optional defines an optional option.
//
// An optional option is an option which is not required and raises no error if
// not parsed from arguments. Optional option takes a single argument as value
// which can be retrieved from Command context using Value("option_name")
type Optional struct {
	// LongName is the long, more descriptive option name. It must contain no
	// spaces and must be unique in an Options. It is used as the key to
	// retrieve this option state from its Command context.
	// An argument with a long prefix is matched against this property.
	LongName string
	// ShortName is the short option name consisting of a single alphanumeric.
	// An argument with a short prefix is matched against this property.
	ShortName string
	// Help is the short help text.
	Help string

	// option embeds the base option properties.
	option
}

// Optional defines an option that is not required.
//
// It takes one argument that is described as type of value for the option when
// printing. Option is defined by long and short names and shows help when
// printed. Returns self.
func (self *Options) Optional(longName, shortName, help string) *Options {
	return self.Register(&Optional{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
	})
}

// Required defines a required option.
//
// A required option is an option which raises an error if not parsed from
// arguments. Required option takes a single argument as value which can be
// retrieved from Command context using Value("option_name")
type Required struct {
	// LongName is the long, more descriptive option name. It must contain no
	// spaces and must be unique in an Options. It is used as the key to
	// retrieve this option state from its Command context.
	// An argument with a long prefix is matched against this property.
	LongName string
	// ShortName is the short option name consisting of a single alphanumeric.
	// An argument with a short prefix is matched against this property.
	ShortName string
	// Help is the short help text.
	Help string

	// option embeds the base option properties.
	option
}

// Required defines an option that is required. It takes one argument that is
// described as type of value for the option when printing.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (self *Options) Required(longName, shortName, help string) *Options {
	return self.Register(&Required{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
	})
}

// Indexed defines an indexed option.
//
// An indexed option is an option that is not qualified by name when passed to
// a Command but is instead matched by index of being defined in Command's
// Options and given as an argument to the Command.
//
// As Indexed Option is given unqualified in the arguments and the argument
// given for the option is the Option Value.
//
// In this example the "indexedOptionValue" is the value for the first Indexed
// Option defined in Options:
//
// myprog.exe somecommand --namedOption=namedOptionValue indexedOptionValue
//
// Indexed options may be declared at any point in between other types of
// options in Options but the order in which Indexed options are declared
// is important and defines the index of the Indexed command.
//
// An Indexed option is a required option and will raise an error if not parsed.
//
// Indexed options are parsed after named Options and must be specified after
// all named options to set for a Command in the arguments.
type Indexed struct {
	// LongName is the long, more descriptive option name. It must contain no
	// spaces and must be unique in an Options. It is used as the key to
	// retrieve this option state from its Command context.
	Name string
	// Help is the short help text.
	Help string

	// option embeds the base option properties.
	option
}

// Indexed defines an option that is passed by index, i.e. the value for the
// option is not prefixed with a short or long option name. It takes one
// argument that is described as type of value for the option when printing.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (self *Options) Indexed(name, help string) *Options {
	return self.Register(&Indexed{
		Name: name,
		Help: help,
	})
}

// Variadic defines a variadic option.
//
// A Variadic Option is an option that takes any and all unparsed arguments as
// an single argument to this Option.
//
// For that reason a Variadic Option's Parent Commands may have no Sub-Commands.
//
// The arguments consumed by this Option are retrievable via Command Context by
// Name of this Option and are returned as a space delimited string.
type Variadic struct {
	// LongName is the long, more descriptive option name. It must contain no
	// spaces and must be unique in an Options. It is used as the key to
	// retrieve this option state from its Command context.
	Name string
	// Help is the short help text.
	Help string

	// option embeds the base option properties.
	option
}

// Variadic defines an option that treats any and all arguments left to parse as
// arguments to self. Only one Variadic option may be defined on a command, it
// must be declared last i.e. no options may be defined after it and the command
// may not have command subsets.
//
// Any unparsed arguments at the time of invocation of this option's command
// handler are retrievable via Context.Value as a space delimited string array.
func (self *Options) Variadic(name, help string) *Options {
	return self.Register(&Variadic{
		Name: name,
		Help: help,
	})
}

// Following funcs implement the Option interface on all five Option types.

func (self Boolean) Key() string  { return self.LongName }
func (self Optional) Key() string { return self.LongName }
func (self Required) Key() string { return self.LongName }
func (self Indexed) Key() string  { return self.Name }
func (self Variadic) Key() string { return self.Name }

func (self Boolean) ShortKey() string  { return self.ShortName }
func (self Optional) ShortKey() string { return self.ShortName }
func (self Required) ShortKey() string { return self.ShortName }
func (self Indexed) ShortKey() string  { return "" }
func (self Variadic) ShortKey() string { return "" }

func (self Boolean) Parsed() bool  { return self.option.parsed }
func (self Optional) Parsed() bool { return self.option.parsed }
func (self Required) Parsed() bool { return self.option.parsed }
func (self Indexed) Parsed() bool  { return self.option.parsed }
func (self Variadic) Parsed() bool { return self.option.parsed }

func (self Boolean) Value() string  { return self.option.raw }
func (self Optional) Value() string { return self.option.raw }
func (self Required) Value() string { return self.option.raw }
func (self Indexed) Value() string  { return self.option.raw }
func (self Variadic) Value() string { return self.option.raw }

// Register registers an Option in these Options where option must be one of
// the Option definition structs in this file. It returns self.
// Option parameter must be one of:
//
//	Boolean, Optional, Required, Indexed, Variadic
func (self *Options) Register(option Option) *Options {
	// TODO: Check short key dulicates.
	switch o := option.(type) {
	case *Boolean:
		self.validateKey(o.LongName)
		self.validateShortKey(o.ShortName)
	case *Optional:
		self.validateKey(o.LongName)
		self.validateShortKey(o.ShortName)
	case *Required:
		self.validateKey(o.LongName)
		self.validateShortKey(o.ShortName)
	case *Indexed:
		self.validateKey(o.Name)
	case *Variadic:
		self.validateKey(o.Name)
		for _, f := range self.options {
			if _, variadic := f.(*Variadic); variadic {
				panic("option set already contains a variadic option")
			}
		}
	default:
		panic("unsupported option type")
	}
	self.options = append(self.options, option)
	return self
}

// validateKey panics if key is empty or non-unique in these Options.
func (self *Options) validateKey(key string) {
	if key == "" {
		panic("empty option key")
	}
	for _, option := range self.options {
		if option.Key() == key {
			panic(fmt.Sprintf("duplicate option key: %s", key))
		}
	}
}

// validateKey panics if short key is empty or non-unique in these Options.
func (self *Options) validateShortKey(key string) {
	if key == "" {
		panic("empty option short key")
	}
	if len(key) > 1 {
		panic("short keys may be one character strings")
	}
	for _, option := range self.options {
		if k := option.ShortKey(); k != "" && k == key {
			panic(fmt.Sprintf("duplicate option short key: %s", key))
		}
	}
}

// Parsed implements Context.Parsed.
func (self *Options) Parsed(name string) bool {
	for _, v := range self.options {
		if v.Key() == name {
			return v.Parsed()
		}
	}
	return false
}

// Parsed implements Context.Value.
func (self *Options) Value(name string) string {
	for _, v := range self.options {
		if v.Key() == name {
			return v.Value()
		}
	}
	return ""
}

type option struct {
	parsed bool
	raw    string
	value  any
}

// parse parses self from args or returns an error.
func (self *Options) parse(args *arguments) error {
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
			o.raw = val
			o.parsed = true
		case *Required:
			if !assignment {
				return fmt.Errorf("required option '%s' requires a value", o.Key())
			}
			o.raw = val
			o.parsed = true
		case *Indexed:
			o.raw = key
			o.parsed = true
		case *Variadic:
			o.raw = strings.Join(args.FromCurrent(), " ")
			o.parsed = true
			break For
		}

		if err := self.rawToMapped(opt); err != nil {
			return fmt.Errorf("invalid value '%s' for option '%s': %w", opt.Value(), opt.Key(), err)
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

func (self *Options) get(long, short string) Option {
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

func (self *Options) getNextUnparsedIndexed() Option {
	for _, v := range self.options {
		if _, ok := v.(Indexed); ok && !v.Parsed() {
			return v
		}
	}
	return nil
}

func (self *Options) getVariadic() Option {
	for _, v := range self.options {
		if _, ok := v.(*Variadic); ok {
			return v
		}
	}
	return nil
}
