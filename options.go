package cmdline

import "fmt"

// Globals returns a new Options set to be used for global options in Parse().
func Globals() *Options { return &Options{} }

// Options contain a list of options.
type Options struct {
	options []*option
}

// Flag defines an option that is not required. It takes no arguments.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (o *Options) Boolean(long, short, help string) *Options {
	return o.option(booleanOption, long, short, "", help)
}

// Flag defines an option that is not required. It takes one argument that is
// described as type of value for the option when printing.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (o *Options) Optional(long, short, argument, help string) *Options {
	return o.option(optionalOption, long, short, argument, help)
}

// Required defines an option that is required. It takes one argument that is
// described as type of value for the option when printing.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (o *Options) Required(long, short, argument, help string) *Options {
	return o.option(requiredOption, long, short, argument, help)
}

// Indexed defines an option that is passed by index, i.e. the value for the
// option is not prefixed with a short or long option name. It takes one
// argument that is described as type of value for the option when printing.
// Option is defined by long and short names and shows help when printed.
// Returns self.
func (o *Options) Indexed(name, argument, help string) *Options {
	return o.option(indexedOption, name, "", argument, help)
}

// Variadic defines an option that treats any and all arguments left to parse as
// arguments to self. Only one Variadic option may be defined on a command, it
// must be declared last i.e. no options may be defined after it and the command
// may not have command subsets.
//
// Any unparsed arguments at the time of invocation of this option's command
// handler are retrievable via Context.Value as a space delimited string array.
func (o *Options) Variadic(name string, help string) *Options {
	return o.option(variadicOption, name, "", "", help)
}

// Parsed implements Context.Parsed.
func (o *Options) Parsed(name string) bool {
	for _, v := range o.options {
		if v.long == name {
			return v.parsed
		}
	}
	return false
}

// Parsed implements Context.Value.
func (o *Options) Value(name string) string {
	for _, v := range o.options {
		if v.long == name {
			return v.value
		}
	}
	return ""
}

func (o *Options) option(kind optionKind, long, short, argument, help string) *Options {
	if len(o.options) > 0 && o.options[len(o.options)-1].kind == variadicOption {
		panic("no options may be defined after a variadic option")
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
	o.options = append(o.options, &option{kind, false, short, long, argument, help, ""})
	return o
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

	short, long, argument, help, value string
}
