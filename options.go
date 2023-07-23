package cmdline

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Option abstracts an option in a slice of Options.
type Option interface {
	// GetLongName returns the Option LongName.
	// LongName represents the long Option name format.
	// Option GetLongName must be unique in its parent Options.
	// Option with both LongName and ShortName maps its LongName to GetLongName.
	// Option with Name maps Name to both GetLongName and GetShortName.
	GetLongName() string
	// GetShortName returns the Option ShortName.
	// ShortName represents the short Option name format.
	// Option GetLongName must be unique in its parent Options.
	// Option with both LongName and ShortName maps its LongName to GetLongName.
	// Option with Name maps Name to both GetLongName and GetShortName.
	GetShortName() string
	// GetIsParsed returns Option.State.IsParsed. It indicates if the Option was
	// parsed from arguments. For Repeated Option it indicates that the Option
	// was parsed at least once.
	GetIsParsed() bool
	// GetRawValues returns the raw option value as a string. Returns an empty
	// slice if Option was not parsed, was not given a argument as a value or
	// the Option takes no values (i.e. Boolean Option).
	GetRawValues() RawValues
	// GetMappedValue returns Option.MappedValue, the variable mapped to this
	// Option. May be nil if unmapped.
	GetMappedValue() any
}

// State represents the Option parse state and is embedded at the top level
// in all supported Option types.
type State struct {
	// IsParsed will be set to true if the Option gets parsed from arguments.
	IsParsed bool
	// RawValues will contain any arguments given as a value to the Option.
	RawValues []string
}

// Options contains and manages a set of Options.
//
// Options set is used to define a set of Options for a Command of the global
// Options not applicable to any specific Command - or all.
//
// Options are never sorted and the order in which Options are declared is
// importan; Print lists them in the oder they were declared and Indexed
// options are matched to indexes of their arguments.
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

// IsParsed implements Context.IsParseded.
func (self Options) IsParsed(longName string) bool {
	for _, v := range self {
		if v.GetLongName() == longName {
			return v.GetIsParsed()
		}
	}
	return false
}

// Parsed implements Context.RawValues.
func (self Options) RawValues(longName string) RawValues {
	for _, v := range self {
		if v.GetLongName() == longName {
			return v.GetRawValues()
		}
	}
	return nil
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

// Optional defines an option that is not required and will not generate an
// error if config.NoFailOnUnparsedRequired is unset. It optionally takes one
// argument in the form of '--option=value' but can be specified without a
// value in the form of '--option',.
//
// In either case, if the option was given in arguments calling IsParsed for
// the Option will return true. If no assignemt was made ('--option')
// or no value was given on assignment ('--option=) the PawValues for the
// Option will return an empty array.
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

// Repeated option is an optional option that can be specified multiple times.
type Repeated struct {
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

// Repeated defines a repeatable option.
func (self Options) Repeated(longName, shortName, help string) Options {
	return self.Register(&Repeated{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
	})
}

// Variadic defines a variadic option.
//
// A Variadic Option is an option that takes any and all unparsed arguments as
// an single argument to this Option.
//
// For that reason a Variadic Option's Parent Options may have no Sub-Options.
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

func (self *Boolean) GetLongName() string  { return self.LongName }
func (self *Optional) GetLongName() string { return self.LongName }
func (self *Required) GetLongName() string { return self.LongName }
func (self *Repeated) GetLongName() string { return self.LongName }
func (self *Indexed) GetLongName() string  { return self.Name }
func (self *Variadic) GetLongName() string { return self.Name }

func (self *Boolean) GetShortName() string  { return self.ShortName }
func (self *Optional) GetShortName() string { return self.ShortName }
func (self *Required) GetShortName() string { return self.ShortName }
func (self *Repeated) GetShortName() string { return self.ShortName }
func (self *Indexed) GetShortName() string  { return self.Name }
func (self *Variadic) GetShortName() string { return self.Name }

func (self *Boolean) GetIsParsed() bool  { return self.State.IsParsed }
func (self *Optional) GetIsParsed() bool { return self.State.IsParsed }
func (self *Required) GetIsParsed() bool { return self.State.IsParsed }
func (self *Repeated) GetIsParsed() bool { return self.State.IsParsed }
func (self *Indexed) GetIsParsed() bool  { return self.State.IsParsed }
func (self *Variadic) GetIsParsed() bool { return self.State.IsParsed }

func (self *Boolean) GetRawValues() RawValues  { return self.State.RawValues }
func (self *Optional) GetRawValues() RawValues { return self.State.RawValues }
func (self *Required) GetRawValues() RawValues { return self.State.RawValues }
func (self *Repeated) GetRawValues() RawValues { return self.State.RawValues }
func (self *Indexed) GetRawValues() RawValues  { return self.State.RawValues }
func (self *Variadic) GetRawValues() RawValues { return self.State.RawValues }

func (self *Boolean) GetMappedValue() any  { return self.MappedValue }
func (self *Optional) GetMappedValue() any { return self.MappedValue }
func (self *Required) GetMappedValue() any { return self.MappedValue }
func (self *Repeated) GetMappedValue() any { return self.MappedValue }
func (self *Indexed) GetMappedValue() any  { return self.MappedValue }
func (self *Variadic) GetMappedValue() any { return self.MappedValue }

// Value defines a type that is capable of parsing a string into a value it
// represents.
type Value interface {
	// String must return a string representation of the type.
	String() string
	// Set must parse a string into self or return an error if it failed.
	Set(RawValues) error
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
// *time.Duration, *[]string, and any type supporting Value interface.
//
// If an unsupported type was set as option.MappedValue Parse will return a
// conversion error.
func (self Options) rawToMapped(option Option) (err error) {

	var (
		mapped = option.GetMappedValue()
		raw    = option.GetRawValues()
	)
	if mapped == nil {
		return nil
	}
	if _, ok := option.(*Boolean); !ok {
		if len(raw) == 0 {
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
	case *Repeated:
		return setMappedValue(p.MappedValue, raw[len(raw)-1:])
	case *Variadic:
		return setMappedValue(p.MappedValue, raw)
	}
	return nil
}

func setMappedValue(v any, raw RawValues) (err error) {
	switch p := v.(type) {
	case *bool:
		*p = true
	case *string:
		*p = raw.First()
	case *int:
		var v int64
		if v, err = strconv.ParseInt(raw.First(), 10, 0); err == nil {
			*p = int(v)
		}
	case *uint:
		var v uint64
		if v, err = strconv.ParseUint(raw.First(), 10, 0); err == nil {
			*p = uint(v)
		}
	case *int8:
		var v int64
		if v, err = strconv.ParseInt(raw.First(), 10, 8); err == nil {
			*p = int8(v)
		}
	case *uint8:
		var v uint64
		if v, err = strconv.ParseUint(raw.First(), 10, 8); err == nil {
			*p = uint8(v)
		}
	case *int16:
		var v int64
		if v, err = strconv.ParseInt(raw.First(), 10, 16); err == nil {
			*p = int16(v)
		}
	case *uint16:
		var v uint64
		if v, err = strconv.ParseUint(raw.First(), 10, 16); err == nil {
			*p = uint16(v)
		}
	case *int32:
		var v int64
		if v, err = strconv.ParseInt(raw.First(), 10, 32); err == nil {
			*p = int32(v)
		}
	case *uint32:
		var v uint64
		if v, err = strconv.ParseUint(raw.First(), 10, 32); err == nil {
			*p = uint32(v)
		}
	case *int64:
		*p, err = strconv.ParseInt(raw.First(), 10, 64)
	case *uint64:
		*p, err = strconv.ParseUint(raw.First(), 10, 64)
	case *float32:
		var v float64
		if v, err = strconv.ParseFloat(raw.First(), 64); err == nil {
			*p = float32(v)
		}
	case *float64:
		*p, err = strconv.ParseFloat(raw.First(), 64)
	case *[]string:
		*p = append(*p, raw...)
	case *time.Duration:
		*p, err = time.ParseDuration(raw.First())
	default:
		if v, ok := p.(Value); ok {
			err = v.Set(raw)
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
	for !config.Arguments.Eof() {
		opt = nil

		var key, val, assignment = strings.Cut(config.Arguments.Text(config), "=")
		key = strings.TrimSpace(key)
		if assignment && val != "" {
			val = strings.TrimSpace(val)
			if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
				val = strings.TrimPrefix(strings.TrimSuffix(val, "\""), "\"")
			}
		}

		switch kind := config.Arguments.Kind(config); kind {
		case TextArgument:
			for _, v := range self {
				if _, ok := v.(*Indexed); ok && !v.GetIsParsed() {
					opt = v
					break
				}
			}
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
			for _, v := range self {
				if _, ok := v.(*Variadic); ok {
					opt = v
					break
				}
			}
			if opt == nil {
				break
			}
		}

		// Fail if non *Repeatable option and parsed multiple times.
		if _, ok := opt.(*Repeated); !ok {
			if opt.GetIsParsed() {
				return fmt.Errorf("option %s specified multiple times", opt.GetLongName())
			}
		}

		switch o := opt.(type) {
		case *Boolean:
			if assignment {
				return fmt.Errorf("option '%s' cannot be assigned a value", o.GetLongName())
			}
			o.IsParsed = true
		case *Optional:
			if assignment && val != "" {
				o.RawValues = append(o.RawValues, val)
			}
			o.IsParsed = true
		case *Required:
			if !assignment || val == "" {
				return fmt.Errorf("option '%s' requires a value", o.GetLongName())
			}
			o.RawValues = append(o.RawValues, val)
			o.IsParsed = true
		case *Repeated:
			if !assignment || val == "" {
				return fmt.Errorf("option '%s' requires a value", o.GetLongName())
			}
			o.RawValues = append(o.RawValues, val)
			o.IsParsed = true
		case *Indexed:
			o.RawValues = append(o.RawValues, key)
			o.IsParsed = true
		case *Variadic:
			o.RawValues = append(o.RawValues, config.Arguments...)
			o.IsParsed = true
			config.Arguments.End()
		}

		if err = self.rawToMapped(opt); err != nil {
			return fmt.Errorf("invalid value '%s' for option '%s': %w", opt.GetMappedValue(), opt.GetLongName(), err)
		}

		config.Arguments.Next()
	}

	if config.NoFailOnUnparsedRequired {
		for _, opt = range self {
			if !opt.GetIsParsed() {
				if _, ok := opt.(*Required); ok {
					return fmt.Errorf("required option '%s' not parsed", opt.GetLongName())
				}
				if _, ok := opt.(*Indexed); ok {
					return fmt.Errorf("indexed option '%s' not parsed", opt.GetLongName())
				}
			}
		}
	}

	return
}
