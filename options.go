// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

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
	//
	// There may only be a single variadic option in [Options].
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
//
// Several option types exist and define how option is parsed. For details see 
// [Kind] enum set for details.
//
// An Option can map to a variable whose value is set from the option value via
// [Option.Var] which must be a pointer to a supported type or any type that
// supports conversion from a string by implementing the [Value] interface.
//
// Supported types are:
// *bool, *string, *float32, *float64,
// *int, *int8, *int16, *1nt32, *int64,
// *uint, *uint8, *uint16, *u1nt32, *uint64
// *time.Duration, *[]string, and any type supporting Value interface.
//
// If an unsupported type was set as [Option.Var] [Parse] will return a
// conversion error.
type Option struct {

	// LongName is the long, more descriptive option name.
	//
	// It must contain no spaces and must be unique in Options as it is the
	// primary key for Option addressing.
	//
	// It is used as the Name for [Indexed] and [Variadic] options in 
	// registration methods. Those options cannot be addressed directly in 
	// arguments and so require no [Option.ShortName].
	//
	// An argument with a long prefix is matched against this property.
	LongName string

	// ShortName is the option short name.
	//
	// ShortName consists of a single alphanumeric. It is optional and if not
	// empty must be unique in all Options ShortName properties.
	//
	// [Indexed] and [Variadic] options do not use ShortName.
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

	// Values contains any string values passed to the option in arguments.
	//
	// How Values is parsed depends on [Option.Kind].
	Values

	// Var is an optional pointer to a variable that will be set from Option
	// argument(s).
	//
	// Only basic types are supported and a slice of string.
	Var any
}

// Reset resets the Option to initial state. It does not modify linked variable.
func (self *Option) Reset() {
	self.IsParsed = true
	clear(self.Values)
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
	return self.BooleanVar(longName, shortName, help, nil)
}

// Optional registers a new optional option in self and returns self.
func (self *Options) Optional(longName, shortName, help string) *Options {
	return self.OptionalVar(longName, shortName, help, nil)
}

// Required registers a new required option in self and returns self.
func (self *Options) Required(longName, shortName, help string) *Options {
	return self.RequiredVar(longName, shortName, help, nil)
}

// Repeated registers a new repeated option in self and returns self.
func (self *Options) Repeated(longName, shortName, help string) *Options {
	return self.RepeatedVar(longName, shortName, help, nil)
}

// Indexed registers a new indexed option in self and returns self.
func (self *Options) Indexed(name, help string) *Options {
	return self.IndexedVar(name, help, nil)
}

// Variadic registers a new variadic option in self and returns self.
func (self *Options) Variadic(name, help string) *Options {
	return self.VariadicVar(name, help, nil)
}

// Boolean registers a new Boolean option in self and returns self.
func (self *Options) BooleanVar(longName, shortName, help string, v any) *Options {
	return self.Register(&Option{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
		Kind:      Boolean,
		Var:       v,
	})
}

// Optional registers a new optional option in self and returns self.
func (self *Options) OptionalVar(longName, shortName, help string, v any) *Options {
	return self.Register(&Option{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
		Kind:      Optional,
		Var:       v,
	})
}

// Required registers a new required option in self and returns self.
func (self *Options) RequiredVar(longName, shortName, help string, v any) *Options {
	return self.Register(&Option{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
		Kind:      Required,
		Var:       v,
	})
}

// Repeated registers a new repeated option in self and returns self.
func (self *Options) RepeatedVar(longName, shortName, help string, v any) *Options {
	return self.Register(&Option{
		LongName:  longName,
		ShortName: shortName,
		Help:      help,
		Kind:      Repeated,
		Var:       v,
	})
}

// Indexed registers a new indexed option in self and returns self.
func (self *Options) IndexedVar(name, help string, v any) *Options {
	return self.Register(&Option{
		LongName:  name,
		ShortName: "",
		Help:      help,
		Kind:      Indexed,
		Var:       v,
	})
}

// Variadic registers a new variadic option in self and returns self.
func (self *Options) VariadicVar(name, help string, v any) *Options {
	return self.Register(&Option{
		LongName:  name,
		ShortName: "",
		Help:      help,
		Kind:      Variadic,
		Var:       v,
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

// Reset resets all options to initial parse state. It does not modify the
// linked variable.
func (self Options) Reset() {
	for _, option := range self {
		option.Reset()
	}
}
