// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vedranvuk/strutils"
)

// Kind specifies the kind of an Option.
//
// It defines the Option behaviour and how it parses its arguments.
type Kind int

const (
	// Invalid is an invalid, undefined or unknown type of option.
	Invalid Kind = iota

	// Boolean is a boolean Option.
	//
	// A Boolean option is an option that takes no arguments. It is marked as
	// parsed if it was specified in command arguments.
	Boolean

	// Optional is an optional Option.
	//
	// It will not return an error during parse if it was not specified in
	// command arguments.
	//
	// It takes a single value.
	Optional

	// Required is a required Option.
	//
	// A required option is an option which raises an error if not parsed from
	// arguments. It takes a value in the same way as [Optional].
	Required

	// Repeated is a repeated Option.
	//
	// Repeated option is an optional option that can be specified one or more
	// times. Each value give to an invocation of this option is appended to the
	// Option's RawValues slice.
	Repeated

	// Indexed is an indexed Option.
	//
	// An indexed option is an option that is not qualified by name when passed
	// to a Command but is instead matched by index of being defined in
	// Command's Options to index of an argument to the Command.
	//
	// Indexed options are parsed after named Options and their values must be
	// given after any and all named options that are to be set. Index of the
	// argument is counted from 0 after all named arguments were given.
	//
	// In this example the "indexedOptionValue" is the value for the first
	// Indexed Option defined in Options:
	//
	// myprog.exe somecommand --namedOption=namedOptionValue indexedOptionValue
	//
	// Indexed options may be declared at any point in between other types of
	// options in Options but the order in which Indexed options are declared
	// is important and defines the index of the Indexed command.
	//
	// An Indexed option is required and will raise an error if not parsed.
	Indexed

	// Variadic is a variadic Option.
	//
	// A Variadic Option is an option that takes any and all unparsed arguments
	// as an single argument to this Option.
	//
	// For that reason a Variadic Option's Parent Options may have no
	// Sub-Options.
	//
	// Variadic option is parsed after all other options and consumes all
	// unused arguments as arguments to self.
	Variadic
)

// String implements [fmt.Stringer] on Kind.
func (self Kind) String() string {
	switch self {
	case Boolean:
		return "Boolean"
	case Optional:
		return "Optional"
	case Required:
		return "Required"
	case Repeated:
		return "Repeated"
	case Indexed:
		return "Indexed"
	case Variadic:
		return "Variadic"
	default:
		return "[INVALID]"
	}
}

// Option defines an option.
type Option struct {

	// LongName is the long, more descriptive option name.
	//
	// It must contain no spaces and must be unique in Options as it is the
	// primary key for Option addressing.
	//
	// An argument with a long prefix is matched against this property.
	LongName string

	// ShortName is the option short name.
	//
	// ShortName consists of a single alphanumeric. It is optional and if not
	// empty must be unique in all Options ShortName properties.
	//
	// An argument with a short prefix is matched against this property.
	ShortName string

	// Help is the option help text.
	//
	// It should be a short, single line description of the option.
	Help string

	// IsParsed indicates if the Option was parsed from arguments.
	//
	// For Repeated Options it indicates that the Option was parsed at least
	// once.
	IsParsed bool

	// Kind is the kind of option which determines how the Option parses
	// its arguments.
	//
	// See [Kind] for details.
	Kind

	// Values contains any string values passed to the option as arguments.
	//
	// How Values is parsed depends on [Option.Kind].
	Values

	// Var is an optional pointer to a variable that will be set from Option
	// argument(s).
	//
	// Only basic types are supported and a slice of string.
	Var any
}

// Values is a helper alias for a slice of strings representing arguments
// passed to an Option. It implements several utilities for retrieving values.
type Values []string

// Count returns number of items in Values.
func (self Values) Count() int { return len(self) }

// IsEmpty returns true if Values are empty.
func (self Values) IsEmpty() bool { return len(self) == 0 }

// First returns the first element in Values or an empty string if empty.
func (self Values) First() string {
	if len(self) > 0 {
		return self[0]
	}
	return ""
}

// Value defines a type that is capable of parsing a string into a value it
// represents.
//
// Options support parsing aruments in types that satisfy the Value interface.
// A Value can be set as a target Var in an Option.
type Value interface {
	// String must return a string representation of the type.
	String() string
	// Set must parse a string into self or return an error if it failed.
	Set(Values) error
}

// Options contains and manages a set of Options.
//
// Options are never sorted and the order in which Options are declared is
// importan; Print lists them in the oder they were declared and Indexed
// options are matched to indexes of their arguments.
type Options []*Option

// Returns number of registered options in self.
func (self Options) Count() int { return len(self) }

// Register adds a new option to these Options and returns self.
func (self *Options) Register(option *Option) *Options {
	// TODO: Check short key dulicates.
	*self = append(*self, option)
	return self
}

// Boolean registers a new Boolean option in self and returns self.
func (self *Options) Boolean(longName, shortName, help string) *Options {
	return self.Register(&Option{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
		Kind:      Boolean,
	})
}

// Optional registers a new optional option in self and returns self.
func (self *Options) Optional(longName, shortName, help string) *Options {
	return self.Register(&Option{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
		Kind:      Optional,
	})
}

// Required registers a new required option in self and returns self.
func (self *Options) Required(longName, shortName, help string) *Options {
	return self.Register(&Option{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
		Kind:      Required,
	})
}

// Repeated registers a new repeated option in self and returns self.
func (self *Options) Repeated(longName, shortName, help string) *Options {
	return self.Register(&Option{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
		Kind:      Repeated,
	})
}

// Indexed registers a new indexed option in self and returns self.
func (self *Options) Indexed(name, help string) *Options {
	return self.Register(&Option{
		LongName:  name,
		ShortName: "",
		Help:      help,
		Kind:      Indexed,
	})
}

// Variadic registers a new variadic option in self and returns self.
func (self *Options) Variadic(name, help string) *Options {
	return self.Register(&Option{
		LongName:  name,
		ShortName: "",
		Help:      help,
		Kind:      Variadic,
	})
}

// FindLong returns an Option with given longName or nil if not found.
func (self Options) FindLong(longName string) *Option {
	for i := 0; i < len(self); i++ {
		if self[i].LongName == longName {
			return self[i]
		}
	}
	return nil
}

// Get returns an Option with given shortName or nil if not found.
func (self Options) FindShort(shortName string) *Option {
	for i := 0; i < len(self); i++ {
		if self[i].ShortName == shortName {
			return self[i]
		}
	}
	return nil
}

// IsParsed implements Context.IsParsed.
func (self Options) Parsed(longName string) bool {
	for _, v := range self {
		if v.LongName == longName {
			return v.IsParsed
		}
	}
	return false
}

// Parsed implements Context.RawValues.
func (self Options) Values(longName string) Values {
	for _, v := range self {
		if v.LongName == longName {
			return v.Values
		}
	}
	return nil
}

// parse parses self from args or returns an error.
func (self Options) parse(config *Config) (err error) {

	if self.Count() == 0 {
		return nil
	}

	var (
		opt        *Option
		key, val   string
		assignment bool
	)

	for !config.Arguments.Eof() {

		// Parse key and val.
		if config.NoAssignment {
			key = strings.TrimSpace(config.Arguments.Text(config))
			val = ""
			assignment = false
		} else {
			key, val, assignment = strings.Cut(config.Arguments.Text(config), "=")
			key = strings.TrimSpace(key)
			if assignment && val != "" {
				val, _ = strutils.UnquoteDouble(strings.TrimSpace(val))
			}
		}

		// Try to detect the option by argument kind.
		// If not prefixed see if theres defined and yet unparsed indexed.
		// If prefixed see if its Boolean, Optional, Required or Repeated.
		switch kind := config.Arguments.Kind(config); kind {
		case TextArgument:
			if config.NoAssignment && opt != nil {
				switch opt.Kind {
				case Optional, Required, Repeated:
				default:
					return fmt.Errorf("option '%s' requires a value", opt.LongName)

				}
			} else {
				if o := self.getFirstUnparsedIndexed(); o != nil {
					opt = o
				}
			}
		case LongArgument:
			if !config.NoIndexedFirst {
				if fui := self.getFirstUnparsedIndexed(); fui != nil {
					return fmt.Errorf("indexed argument '%s' not parsed", fui.LongName)
				}
			}

			if config.NoAssignment && opt != nil {
				return fmt.Errorf("option '%s' requires a value", opt.LongName)
			}

			if opt = self.FindLong(key); opt == nil {
				return fmt.Errorf("unknown option '%s'", key)
			}

			switch opt.Kind {
			case Boolean:
			case Optional, Required, Repeated:
				if config.NoAssignment {
					config.Arguments.Next()
					continue
				}
			default:
				return fmt.Errorf("option '%s' exists, but is not named", opt.LongName)

			}
		case ShortArgument:
			if !config.NoIndexedFirst {
				if fui := self.getFirstUnparsedIndexed(); fui != nil {
					return fmt.Errorf("indexed argument '%s' not parsed", fui.LongName)
				}
			}

			if config.NoAssignment && opt != nil {
				return fmt.Errorf("option '%s' requires a value", opt.LongName)
			}

			if opt = self.FindShort(key); opt == nil {
				return fmt.Errorf("unknown option '%s'", key)
			}

			switch opt.Kind {
			case Boolean:
			case Optional, Required, Repeated:
				if config.NoAssignment {
					config.Arguments.Next()
					continue
				}
			default:
				return fmt.Errorf("option '%s' exists, but is not named", opt.LongName)

			}
		}

		// No options matched so far, see if there's a Variadic.
		if opt == nil {
			for _, v := range self {
				if v.Kind == Variadic {
					opt = v
					break
				}
			}
			if opt == nil {
				break
			}
		}

		// Fail if non *Repeatable option and parsed multiple times.
		if opt.Kind != Repeated {
			if opt.IsParsed {
				return fmt.Errorf("option %s specified multiple times", opt.LongName)
			}
		}

		// Sets the Option as parsed and sets raw value(s).
		switch opt.Kind {
		case Boolean:
			if config.NoAssignment {
				opt.IsParsed = true
			} else {
				if assignment {
					return fmt.Errorf("option '%s' cannot be assigned a value", opt.LongName)
				}
				opt.IsParsed = true
			}
		case Optional:
			if config.NoAssignment {
				opt.Values = append(opt.Values, key)
				opt.IsParsed = true
			} else {
				if !assignment || val == "" {
					return fmt.Errorf("option '%s' requires a value", opt.LongName)
				}
				opt.Values = append(opt.Values, val)
				opt.IsParsed = true
			}
		case Required:
			if config.NoAssignment {
				opt.Values = append(opt.Values, key)
				opt.IsParsed = true
			} else {
				if !assignment || val == "" {
					return fmt.Errorf("option '%s' requires a value", opt.LongName)
				}
				opt.Values = append(opt.Values, val)
				opt.IsParsed = true
			}
		case Repeated:
			if config.NoAssignment {
				opt.Values = append(opt.Values, key)
				opt.IsParsed = true
			} else {
				if !assignment || val == "" {
					return fmt.Errorf("option '%s' requires a value", opt.LongName)
				}
				opt.Values = append(opt.Values, val)
				opt.IsParsed = true
			}
		case Indexed:
			if config.NoAssignment {
				opt.Values = append(opt.Values, key)
				opt.IsParsed = true
			} else {
				opt.Values = append(opt.Values, key)
				opt.IsParsed = true
			}
		case Variadic:
			opt.Values = append(opt.Values, config.Arguments...)
			opt.IsParsed = true
			config.Arguments.End()
		}

		if err = self.setVar(opt); err != nil {
			return fmt.Errorf("invalid value '%s' for option '%s': %w", opt.Var, opt.LongName, err)
		}

		opt = nil
		config.Arguments.Next()
	}

	if !config.NoFailOnUnparsedRequired {
		for _, opt = range self {
			if !opt.IsParsed {
				if opt.Kind == Required {
					return fmt.Errorf("required option '%s' not parsed", opt.LongName)
				}
				if opt.Kind == Indexed {
					return fmt.Errorf("indexed option '%s' not parsed", opt.LongName)
				}
			}
		}
	}

	return
}

// getFirstUnparsedIndexed returns the first Indexed Option that is not parsed.
// Returns nil if none found.
func (self Options) getFirstUnparsedIndexed() *Option {
	for _, option := range self {
		if option.Kind == Indexed {
			if !option.IsParsed {
				return option
			}
		}
	}
	return nil
}

// setVar converts option.State.RawValue to option.MappedValue if option's
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
func (self Options) setVar(option *Option) (err error) {

	if option.Var == nil {
		return nil
	}

	if option.Kind != Boolean {
		if option.Values.Count() < 1 {
			return nil
		}
	}

	switch option.Kind {
	case Boolean, Optional, Required, Indexed, Variadic:
		return convertToVar(option.Var, option.Values)
	case Repeated:
		return convertToVar(option.Var, option.Values[len(option.Values)-1:])
	default:
		return errors.New("invalid OptionKind")
	}
}

// convertToVar sets v which must be a pointer to a supported type from raw
// or returns an error if conversion error occured.
func convertToVar(v any, raw Values) (err error) {
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
