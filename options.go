package cmdline

import (
	"fmt"
)

// Option defines an option.
type Option interface {
	// Key returns the Option Key by which Option is addressed from
	// Command Context. Option Key must be unique in an OptionSet.
	//
	// Options with both long and short names map their long name as the Key.
	Key() string
	// Parsed returns true if the Option was parsed from the arguments.
	Parsed() bool
	// Value returns parsed Option Value. Result may be an empty string if the
	// Option was not parsed or takes no argument.
	Value() string
}

// OptionSet contains and manages a set of Options.
//
// An OptionSet is used either as a set of Options for a Command or for Global
// Options which may be parsed before any Command invocation and are inspected
// separately instead of through a Command Context as they apply to no specific
// Command. For more info see Parse function.
type OptionSet struct {
	options []Option
}

// NewOptionSet returns a new OptionSet.
func NewOptionSet() *OptionSet { return &OptionSet{} }

// Boolean defines a boolean option.
//
// A Boolean option is an option that takes no argument and is simply marked as
// parsed if found in arguments.
type Boolean struct {
	// LongName is the long, more descriptive option name. It must contain no
	// spaces and must be unique in an OptionSet. It is used as the key to
	// retrieve this option state from its Command context.
	// An argument with a long prefix is matched against this property.
	LongName string
	// ShortName is the short option name consisting of a single alphanumeric.
	// An argument with a short prefix is matched against this property.
	ShortName string
	// ShortHelp is a one-line help text displayed alongside the option when
	// printing option list for a command. It is optional but should be provided
	// for a brief overview of the option purpose.
	ShortHelp string
	// LongHelp is a longer help format and is displayed when help is requested
	// for a specific option.
	LongHelp string

	// option embeds the base option properties.
	option
}

// Flag defines an option that is not required. It takes no arguments.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (self *OptionSet) Boolean(longName, shortName, shortHelp, longHelp string) *OptionSet {
	return self.Register(&Boolean{
		LongName:  longName,
		ShortName: shortName,
		ShortHelp: shortHelp,
		LongHelp:  longHelp,
	})
}

// Optional defines an optional option.
// An optional option is an option which is not required and raises no error if
// not parsed from arguments. Optional option takes a single argument as value
// which can be retrieved from Command context using Value("option_name")
type Optional struct {
	// LongName is the long, more descriptive option name. It must contain no
	// spaces and must be unique in an OptionSet. It is used as the key to
	// retrieve this option state from its Command context.
	// An argument with a long prefix is matched against this property.
	LongName string
	// ShortName is the short option name consisting of a single alphanumeric.
	// An argument with a short prefix is matched against this property.
	ShortName string
	// ShortHelp is a one-line help text displayed alongside the option when
	// printing option list for a command. It is optional but should be provided
	// for a brief overview of the option purpose.
	ShortHelp string
	// LongHelp is a longer help format and is displayed when help is requested
	// for a specific option.
	LongHelp string
	// Argument describes the type of argument value the option expects.
	// This is purely cosmetical and is used to give a hint to the user when
	// specifying the argument to the option.
	//
	// The value of the Argument field is otherwise unimportant and may contain
	// any text.
	Argument string

	// option embeds the base option properties.
	option
}

// Flag defines an option that is not required. It takes one argument that is
// described as type of value for the option when printing.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (self *OptionSet) Optional(longName, shortName, shortHelp, longHelp, argument string) *OptionSet {
	return self.Register(&Optional{
		LongName:  longName,
		ShortName: shortName,
		ShortHelp: shortHelp,
		LongHelp:  longHelp,
		Argument:  argument,
	})
}

// Required defines a required option.
// A required option is an option which raises an error if not found in
// arguments. Required option takes a single argument as value which can be
// retrieved from Command context using Value("option_name")
type Required struct {
	// LongName is the long, more descriptive option name. It must contain no
	// spaces and must be unique in an OptionSet. It is used as the key to
	// retrieve this option state from its Command context.
	// An argument with a long prefix is matched against this property.
	LongName string
	// ShortName is the short option name consisting of a single alphanumeric.
	// An argument with a short prefix is matched against this property.
	ShortName string
	// ShortHelp is a one-line help text displayed alongside the option when
	// printing option list for a command. It is optional but should be provided
	// for a brief overview of the option purpose.
	ShortHelp string
	// LongHelp is a longer help format and is displayed when help is requested
	// for a specific option.
	LongHelp string
	// Argument describes the type of argument value the option expects.
	// This is purely cosmetical and is used to give a hint to the user when
	// specifying the argument to the option.
	//
	// The value of the Argument field is otherwise unimportant and may contain
	// any text.
	Argument string

	// option embeds the base option properties.
	option
}

// Required defines an option that is required. It takes one argument that is
// described as type of value for the option when printing.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (self *OptionSet) Required(longName, shortName, shortHelp, longHelp, argument string) *OptionSet {
	return self.Register(&Required{
		LongName:  longName,
		ShortName: shortName,
		ShortHelp: shortHelp,
		LongHelp:  longHelp,
		Argument:  argument,
	})
}

// Indexed defines an indexed option.
//
// An indexed option is an option that is not qualified by name when passed to
// a Command but is instead matched by index of being defined in an OptionSet
// and given as an argument to the Command.
//
// As Indexed Option is given unqualified in the arguments and the argument
// given for the option is the Option Value.
//
// In this example the "indexedOptionValue" is the value for the first Indexed
// Option defined in an OptionSet:
//
// myprox.exe somecommand --namedOption=namedOptionValue indexedOptionValue
//
// Indexed options may be declared at any point in between other types of
// options in an OptionSet but the order in which Indexed options are declared
// is important and defines the index of the Indexed command.
//
// An Indexed option is a required option and will raise an error if not parsed.
//
// Indexed options are parsed after named Options and must be specified after
// all named options to set for a Command in the arguments.
type Indexed struct {
	// LongName is the long, more descriptive option name. It must contain no
	// spaces and must be unique in an OptionSet. It is used as the key to
	// retrieve this option state from its Command context.
	Name string
	// ShortHelp is a one-line help text displayed alongside the option when
	// printing option list for a command. It is optional but should be provided
	// for a brief overview of the option purpose.
	ShortHelp string
	// LongHelp is a longer help format and is displayed when help is requested
	// for a specific option.
	LongHelp string
	// Argument describes the type of argument value the option expects.
	// This is purely cosmetical and is used to give a hint to the user when
	// specifying the argument to the option.
	//
	// The value of the Argument field is otherwise unimportant and may contain
	// any text.
	Argument string

	// option embeds the base option properties.
	option
}

// Indexed defines an option that is passed by index, i.e. the value for the
// option is not prefixed with a short or long option name. It takes one
// argument that is described as type of value for the option when printing.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (self *OptionSet) Indexed(name, shortHelp, longHelp, argument string) *OptionSet {
	return self.Register(&Indexed{
		Name:      name,
		ShortHelp: shortHelp,
		LongHelp:  longHelp,
		Argument:  argument,
	})
}

// Variadic defines a variadic option.
//
// A Variadic Option is an option that takes any and all unparsed arguments as
// an single argument to this Option.
//
// For that reason a Variadic Option may have no Sub-Commands.
//
// The arguments consumed by this Option are retrievable via Command Context by
// Name of this Option and are returned as a space delimited string.
type Variadic struct {
	// LongName is the long, more descriptive option name. It must contain no
	// spaces and must be unique in an OptionSet. It is used as the key to
	// retrieve this option state from its Command context.
	Name string
	// ShortHelp is a one-line help text displayed alongside the option when
	// printing option list for a command. It is optional but should be provided
	// for a brief overview of the option purpose.
	ShortHelp string
	// LongHelp is a longer help format and is displayed when help is requested
	// for a specific option.
	LongHelp string
	// Arguments describes the type of argument values the option expects.
	// This is purely cosmetical and is used to give a hint to the user when
	// specifying arguments to the option.
	//
	// The value of the Arguments field is otherwise unimportant and may contain
	// any text.
	Arguments string

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
func (self *OptionSet) Variadic(name, shortHelp, longHelp, arguments string) *OptionSet {
	return self.Register(&Variadic{
		Name:      name,
		ShortHelp: shortHelp,
		LongHelp:  longHelp,
		Arguments: arguments,
	})
}

// Following funcs implement the Option interface on implemented Option types.

func (self *Boolean) Key() string  { return self.LongName }
func (self *Optional) Key() string { return self.LongName }
func (self *Required) Key() string { return self.LongName }
func (self *Indexed) Key() string  { return self.Name }
func (self *Variadic) Key() string { return self.Name }

func (self *Boolean) Parsed() bool  { return self.option.parsed }
func (self *Optional) Parsed() bool { return self.option.parsed }
func (self *Required) Parsed() bool { return self.option.parsed }
func (self *Indexed) Parsed() bool  { return self.option.parsed }
func (self *Variadic) Parsed() bool { return self.option.parsed }

func (self *Boolean) Value() string  { return self.option.value }
func (self *Optional) Value() string { return self.option.value }
func (self *Required) Value() string { return self.option.value }
func (self *Indexed) Value() string  { return self.option.value }
func (self *Variadic) Value() string { return self.option.value }

// Register registers an Option with the CommandSet where option must be one of
// the Option definition structs in this file. It returns self.
// Option parameter must be one of:
//
//	Boolean, Optional, Required, Indexed, Variadic
func (self *OptionSet) Register(option Option) *OptionSet {
	switch o := option.(type) {
	case *Boolean:
		self.validateKey(o.LongName)
	case *Optional:
		self.validateKey(o.LongName)
	case *Required:
		self.validateKey(o.LongName)
	case *Indexed:
		self.validateKey(o.Name)
	case *Variadic:
		self.validateKey(o.Name)
	default:
		panic("unsupported option type")
	}

	if kind == requiredOption && argument == "" {
		panic("required option requires an argument")
	}
	if kind == indexedOption && argument == "" {
		panic("indexed option requires an argument")
	}
	for _, f := range o.options {
		if long == "" {
			panic("long option name must not be empty")
		}
		if f.long == long {
			panic(fmt.Sprintf("opiton long form '%s' already registered", long))
		}
		if f.short == short && short != "" {
			panic(fmt.Sprintf("option short form '%s' already registered", short))
		}
		if f.kind == variadicOption && kind == variadicOption {
			panic("option set already contains variadic option")
		}
	}
	return self
}

// validateKey panics if key is empty or non-unique in this OptionSet.
func (self *OptionSet) validateKey(key string) {
	if key == "" {
		panic("empty option key")
	}
	for _, option := range self.options {
		if option.Key() == key {
			panic(fmt.Sprintf("duplicate option key: %s", key))
		}
	}
}

// Parsed implements Context.Parsed.
func (o *OptionSet) Parsed(name string) bool {
	for _, v := range o.options {
		if v.Key() == name {
			return v.Parsed()
		}
	}
	return false
}

// Parsed implements Context.Value.
func (o *OptionSet) Value(name string) string {
	for _, v := range o.options {
		if v.Key() == name {
			return v.Value()
		}
	}
	return ""
}

type optionKind int

const (
	booleanOption optionKind = iota
	optionalOption
	requiredOption
	indexedOption
	variadicOption
)

type option struct {
	kind   optionKind
	parsed bool
	value  string
}
