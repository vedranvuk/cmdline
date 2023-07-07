package cmdline

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Option defines an option in a slice of Commands.
type Option interface {
	// GetLongName returns the Option GetLongName by which Option is addressed from
	// Command Context. Option GetLongName must be unique in an Options.
	//
	// Options with both long and short names map their long name as the GetLongName.
	GetLongName() string
	// GetShortName returns the short form of the Option key. This is used only by
	// Boolean, Optional and Required options which can be addressed from
	// arguments with a shorter, single character option name.
	GetShortName() string
	// GetParsed returns true if the Option was parsed from the arguments.
	GetParsed() bool
	// GetRawValue returns the raw option value as a string. Returns an empty string if
	// option was not parsed, was not given an argument or is not applicable.
	GetRawValue() string
	// GetMappedValue returns the value mapped to this Option. May be nil if unmapped.
	GetMappedValue() any
}

// State represents the Option parse state.
type State struct {
	Parsed   bool
	RawValue string
}

// Options contains and manages a set of Options.
//
// An Options is used either as a set of Options for a Command or for Global
// Options which may be parsed before any Command invocation and are inspected
// separately instead of through a Command Context as they apply to no specific
// Command. For more info see Parse function.
type Options []Option

// Returns number of registered options in self.
func (self Options) Count() int { return len(self) }

// FindLong returns an Option with given longName or nil if not found.
func (self Options) FindLong(longName string) Option {
	for i := 0; i < len(self); i++ {
		if self[i].GetLongName() == longName {
			return self[i]
		}
	}
	return nil
}

// Get returns an Option with given shortName or nil if not found.
func (self Options) FindShort(shortName string) Option {
	for i := 0; i < len(self); i++ {
		if self[i].GetShortName() == shortName {
			return self[i]
		}
	}
	return nil
}

// Parsed implements Context.Parsed.
func (self Options) Parsed(longName string) bool {
	for _, v := range self {
		if v.GetLongName() == longName {
			return v.GetParsed()
		}
	}
	return false
}

// Parsed implements Context.Value.
func (self Options) Value(longName string) string {
	for _, v := range self {
		if v.GetLongName() == longName {
			return v.GetRawValue()
		}
	}
	return ""
}

// Register registers an Option in these Options where option must be one of
// the Option definition structs in this file. It returns self.
// Option parameter must be one of:
//
//	Boolean, Optional, Required, Indexed, Variadic
func (self Options) Register(option Option) Options {
	// TODO: Check short key dulicates.
	self = append(self, option)
	return self
}

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
	// MappedValue is an optional *bool that will be set to true if this option
	// gets parsed.
	MappedValue any

	// State embeds the base State properties.
	State
}

// Flag defines an option that is not required. It takes no arguments.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (self Options) Boolean(longName, shortName, help string) Options {
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
	// MappedValue is an optional pointer to one of supported variable types
	// that the option will get parsed into. For details see rawToMapped.
	MappedValue any

	// State embeds the base State properties.
	State
}

// Optional defines an option that is not required.
//
// It takes one argument that is described as type of value for the option when
// printing. Option is defined by long and short names and shows help when
// printed. Returns self.
func (self Options) Optional(longName, shortName, help string) Options {
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
	// MappedValue is an optional pointer to one of supported variable types
	// that the option will get parsed into. For details see rawToMapped.
	MappedValue any

	// State embeds the base State properties.
	State
}

// Required defines an option that is required. It takes one argument that is
// described as type of value for the option when printing.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (self Options) Required(longName, shortName, help string) Options {
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
	// MappedValue is an optional pointer to one of supported variable types
	// that the option will get parsed into. For details see rawToMapped.
	MappedValue any

	// State embeds the base State properties.
	State
}

// Indexed defines an option that is passed by index, i.e. the value for the
// option is not prefixed with a short or long option name. It takes one
// argument that is described as type of value for the option when printing.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (self Options) Indexed(name, help string) Options {
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
	// MappedValue is an optional pointer to one of supported variable types
	// that the option will get parsed into. For details see rawToMapped.
	MappedValue any

	// State embeds the base State properties.
	State
}

// Variadic defines an option that treats any and all arguments left to parse as
// arguments to self. Only one Variadic option may be defined on a command, it
// must be declared last i.e. no options may be defined after it and the command
// may not have command subsets.
//
// Any unparsed arguments at the time of invocation of this option's command
// handler are retrievable via Context.Value as a space delimited string array.
func (self Options) Variadic(name, help string) Options {
	return self.Register(&Variadic{
		Name: name,
		Help: help,
	})
}

// Following funcs implement the Option interface on all five Option types.

func (self Boolean) GetLongName() string  { return self.LongName }
func (self Optional) GetLongName() string { return self.LongName }
func (self Required) GetLongName() string { return self.LongName }
func (self Indexed) GetLongName() string  { return self.Name }
func (self Variadic) GetLongName() string { return self.Name }

func (self Boolean) GetShortName() string  { return self.ShortName }
func (self Optional) GetShortName() string { return self.ShortName }
func (self Required) GetShortName() string { return self.ShortName }
func (self Indexed) GetShortName() string  { return self.Name }
func (self Variadic) GetShortName() string { return self.Name }

func (self Boolean) GetParsed() bool  { return self.State.Parsed }
func (self Optional) GetParsed() bool { return self.State.Parsed }
func (self Required) GetParsed() bool { return self.State.Parsed }
func (self Indexed) GetParsed() bool  { return self.State.Parsed }
func (self Variadic) GetParsed() bool { return self.State.Parsed }

func (self Boolean) GetRawValue() string  { return self.State.RawValue }
func (self Optional) GetRawValue() string { return self.State.RawValue }
func (self Required) GetRawValue() string { return self.State.RawValue }
func (self Indexed) GetRawValue() string  { return self.State.RawValue }
func (self Variadic) GetRawValue() string { return self.State.RawValue }

func (self Boolean) GetMappedValue() any  { return self.MappedValue }
func (self Optional) GetMappedValue() any { return self.MappedValue }
func (self Required) GetMappedValue() any { return self.MappedValue }
func (self Indexed) GetMappedValue() any  { return self.MappedValue }
func (self Variadic) GetMappedValue() any { return self.MappedValue }

// Value defines a type that is capable of parsing a string into a value it
// represents.
type Value interface {
	// String must return a string representation of the type.
	String() string
	// Set must parse a string into self or return an error if it failed.
	Set(string) error
}

// rawToMapped converts option.State.RawValue to option.MappedValue if option's
// raw value is not nil.
//
// Returns nil on success of no value mapped.. Returns a non nil error on
// failed conversion only.
//
// option.MappedValue must be a pointer to a supported type or any type that
// supports conversion from a string by implementing the Value interface.
//
// Supported types are:
// *bool, *string, *float32, *float64,
// *int, *int8, *int16, *1nt32, *int64,
// *uint, *uint8, *uint16, *u1nt32, *uint64
// *time.Duration, any type supporting Value interface.
//
// If an unsupported type was set as option.MappedValue Parse will return a
// conversion error.
func (self Options) rawToMapped(option Option) (err error) {

	var (
		mapped = option.GetMappedValue()
		raw    = option.GetRawValue()
	)
	if mapped == nil {
		return nil
	}
	if _, ok := option.(*Boolean); !ok {
		if raw == "" {
			return nil
		}
	}
	switch p := option.(type) {
	case *Boolean:
		return setMappedValue(p.MappedValue, raw)
	case *Optional:
		return setMappedValue(p.MappedValue, raw)
	case *Required:
		return setMappedValue(p.MappedValue, raw)
	case *Indexed:
		return setMappedValue(p.MappedValue, raw)
	case *Variadic:
		return setMappedValue(p.MappedValue, raw)
	}
	return nil
}

func setMappedValue(v any, s string) (err error) {
	switch p := v.(type) {
	case *bool:
		*p = true
	case *string:
		*p = s
	case *int:
		var v int64
		if v, err = strconv.ParseInt(s, 10, 0); err == nil {
			*p = int(v)
		}
	case *uint:
		var v uint64
		if v, err = strconv.ParseUint(s, 10, 0); err == nil {
			*p = uint(v)
		}
	case *int8:
		var v int64
		if v, err = strconv.ParseInt(s, 10, 8); err == nil {
			*p = int8(v)
		}
	case *uint8:
		var v uint64
		if v, err = strconv.ParseUint(s, 10, 8); err == nil {
			*p = uint8(v)
		}
	case *int16:
		var v int64
		if v, err = strconv.ParseInt(s, 10, 16); err == nil {
			*p = int16(v)
		}
	case *uint16:
		var v uint64
		if v, err = strconv.ParseUint(s, 10, 16); err == nil {
			*p = uint16(v)
		}
	case *int32:
		var v int64
		if v, err = strconv.ParseInt(s, 10, 32); err == nil {
			*p = int32(v)
		}
	case *uint32:
		var v uint64
		if v, err = strconv.ParseUint(s, 10, 32); err == nil {
			*p = uint32(v)
		}
	case *int64:
		*p, err = strconv.ParseInt(s, 10, 64)
	case *uint64:
		*p, err = strconv.ParseUint(s, 10, 64)
	case *float32:
		var v float64
		if v, err = strconv.ParseFloat(s, 64); err == nil {
			*p = float32(v)
		}
	case *float64:
		*p, err = strconv.ParseFloat(s, 64)
	case *time.Duration:
		*p, err = time.ParseDuration(s)
	default:
		if v, ok := p.(Value); ok {
			err = v.Set(s)
		} else {
			return errors.New("unsupported mapped value")
		}
	}
	return
}

// parse parses self from args or returns an error.
func (self Options) parse(config *Config) (err error) {
	if self.Count() == 0 {
		return nil
	}
	var opt Option
For:
	for {
		opt = nil

		var key, val, assignment = strings.Cut(config.Arguments.Text(config), "=")
		key = strings.TrimSpace(key)
		if assignment {
			val = strings.TrimSpace(val)
			if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
				val = strings.TrimPrefix(strings.TrimSuffix(val, "\""), "\"")
			}
		}

		switch kind := config.Arguments.Kind(config); kind {
		case TextArgument:
			opt = self.getNextUnparsedIndexed()
		case LongArgument:
			if opt = self.FindLong(key); opt == nil {
				return fmt.Errorf("option %s%s not registered", config.LongPrefix, key)
			}
		case ShortArgument:
			if opt = self.FindShort(key); opt == nil {
				return fmt.Errorf("option %s%s not registered", config.ShortPrefix, key)
			}
		}

		if opt == nil {
			if opt = self.getVariadic(); opt == nil {
				break
			}
		}

		switch o := opt.(type) {
		case *Boolean:
			o.Parsed = true
		case *Optional:
			o.RawValue = val
			o.Parsed = true
		case *Required:
			if !assignment {
				return fmt.Errorf("required option '%s' requires a value", o.GetLongName())
			}
			o.RawValue = val
			o.Parsed = true
		case *Indexed:
			o.RawValue = key
			o.Parsed = true
		case *Variadic:
			o.RawValue = strings.Join(config.Arguments, " ")
			o.Parsed = true
			config.Arguments.End()
			break For
		}

		if err = self.rawToMapped(opt); err != nil {
			return fmt.Errorf("invalid value '%s' for option '%s': %w", opt.GetMappedValue(), opt.GetLongName(), err)
		}

		config.Arguments.Next()
	}

	for _, opt = range self {
		if !opt.GetParsed() {
			if _, ok := opt.(*Required); ok {
				return fmt.Errorf("required option '%s' not parsed", opt.GetLongName())
			}
			if _, ok := opt.(*Indexed); ok {
				return fmt.Errorf("indexed option '%s' not parsed", opt.GetLongName())
			}
		}
	}

	return
}

func (self Options) getNextUnparsedIndexed() Option {
	for _, v := range self {
		if _, ok := v.(Indexed); ok && !v.GetParsed() {
			return v
		}
	}
	return nil
}

func (self Options) getVariadic() Option {
	for _, v := range self {
		if _, ok := v.(*Variadic); ok {
			return v
		}
	}
	return nil
}
