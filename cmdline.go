/*
Package cmdline implements a handler based command parsing from executable
arguments designed for hierarchical command based invocations, with options.

# Command line option syntax

The syntax mimics GNU-like long/short combination style of option names.

For example:

	--verbose	(long form)
	-v 			(short form)

If an option takes a value it must be assigned using "=". If it contains spaces
it must be enclosed in double quotes.

For example:

	--input-directory=/home/yourname
	--title="Adriatic Coasting"

Commands can be invoked by name, they can have options and have sub commands
that have options. So you can have things like:

	// Invoke an executable with global flag "-v" executing command "create"
	// with options "output-dir" taking a value and an indexed option
	// "My Project".
	tool.exe -v create --output-dir=/home/yourname/projects "My Project"

	// Invoke an executable with global flag "--color=auto" executing
	// sub-command "name" with indexed argument "newname" on command "change".
	ip.exe --color=auto change name "newname"

# Usage

Create a new command Set and register a command named "new" then add all four
option forms to the command as an example:

	set := New()
	set.Handle("new", func(c Context) error {
		fmt.Printf("Create a new project %s\n", c.Value("name"))
		return nil
	}).Options().
		Required("target-directory", "t", "Specify target directory.").
		Option("force-overwrite", "f", "Force overwriting target files.").
		IndexedRequired("name", "Name of the project.").
		Indexed("author", "Project author")

Create a set of global options that do not apply to any defined command.

	globals := Globals().
		Option("verbose", "v", "Be more verbose.").
		Option("help", "h", "Show help.")

Parse the program arguments into our command set and global options and handle
any errors returned.

	if err := Parse(os.Args[1:], set, globals); err != nil {
		if errors.Is(err, ErrNoArguments) {
			Usage()
		} else {
			fmt.Printf("invalid input: %v\n", err)
		}
		os.Exit(1)
	}
*/
package cmdline

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

var (
	// ErrNoArgs is returned by Parse if no arguments were given for parsing.
	ErrNoArgs = errors.New("no arguments")

	// ErrVariadic is returned by Parse to indicate the presence of a
	// variadic option in an option set.
	ErrVariadic = errors.New("variadic option")

	// ErrInvoked may be returned by a Command Handler to indicate that the
	// parsing should be stopped after this command, i.e., no sub commands will
	// be further parsed after all the Option arguments were parsed.
	//
	// Use this in situations where an invocation of a command takes over
	// complete program control and there is no need to parse sub commands.
	ErrInvoked = errors.New("command invoked")
)

// Handler is a Command invocation callback prototype. It carries a Command
// Context which allows inspection of the parse state.
//
// If the handler returns a nil error the parser continues the command chain
// execution.
//
// If the Handler returns a non-nil error the parser aborts the command chain
// execution and the error is pushed back to the Set.Parse() caller.
type Handler func(Context) error

// NopHandler is an no-op handler that just returns a nil error.
// It can be used as a placeholder to skip command implementation and continue
// the command chain execution.
var NopHandler = func(Context) error { return nil }

// Context is passed to the Command handler.
// It allows for inspection of the Command's Options.
type Context interface {
	// Parsed returns truth if an Option with specified Key was parsed.
	Parsed(string) bool
	// Value returns value of the flag with specified name.
	// Unparsed flags return an empty value. Use Parsed to check validity.
	Value(string) string
}

// Command defines a command invocable by name.
type Command struct {
	// Name is the name of the Command by which it is invoked from arguments.
	// Command name is required, must not be empty and must be unique in
	// Commands.
	Name string
	// ShortHelp is the short Command help. It should be a single line of text
	// that can be used as a short command description in the command listing.
	ShortHelp string
	// LongHelp is the long Command help. It can be a formatted string of any
	// length and should hold a long description of the Command's functionality.
	LongHelp string
	// Handler is the function to call when the Command gets invoked from
	// arguments during parsing.
	Handler Handler
	// SubCommands are this Command's sub commands. Command invocation can be
	// chained as described in the Parse method. SubCommands are optional.
	SubCommands *Commands
	// Options are this Command's options. Options are optional :|
	Options *Options
}

// GetOptions returns the Command's Options.
//
// If Command's Options are nil a new Options instance is allocated for the
// Command then returned.
func (self *Command) GetOptions() *Options {
	if self.Options == nil {
		self.Options = NewOptions()
	}
	return self.Options
}

// GetSubCommands returns this Command's subset of commands which can be invoked
// on this command.
//
// If the Command's sub-commands are nil a new Commands instance is allocated
// for the Command then returned.
func (self *Command) GetSubCommands() *Commands {
	if self.SubCommands == nil {
		self.SubCommands = NewCommands()
	}
	return self.SubCommands
}

// Parsed implements Context.Parsed for handlers of commands with no Options.
// It is a no-op placeholder that returns false.
func (self *Command) Parsed(string) bool { return false }

// Value implements Context.Value for handlers of commands with no Options.
// It is a no-op that returns an empty string.
func (self *Command) Value(string) string { return "" }

// Commands holds a set of Commands.
type Commands struct {
	cmds map[string]*Command
}

// NewCommands returns a new Commands instance.
func NewCommands() *Commands { return &Commands{make(map[string]*Command)} }

// Handle registers a new Command from specified name, shortHelp, longHelp and
// Handler h and returns it. If the registration fails the function panics.
func (self *Commands) Handle(name, shortHelp, longHelp string, h Handler) (c *Command) {
	c = &Command{
		Name:      name,
		ShortHelp: shortHelp,
		LongHelp:  longHelp,
		Handler:   h,
	}
	self.Register(c)
	return
}

// Register registers a fully defined Command and returns self. If the
// registration fails the function panics.
func (self *Commands) Register(command *Command) *Commands {
	if command.Name == "" {
		panic("command name is empty")
	}
	if _, exists := self.cmds[command.Name]; exists {
		panic(fmt.Sprintf("command '%s' already registered", command.Name))
	}
	if command.Handler == nil {
		panic(fmt.Sprintf("command '%s' nil registering nil handler", command.Name))
	}
	self.cmds[command.Name] = command
	return self
}

// Count returns the number of defined commands in these Commands.
func (self *Commands) Count() int { return len(self.cmds) }

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
}

// NewOptions returns a new Options instance.
func NewOptions() *Options { return &Options{} }

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
func (self *Options) Boolean(longName, shortName, shortHelp, longHelp string) *Options {
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
	// spaces and must be unique in an Options. It is used as the key to
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
func (self *Options) Optional(longName, shortName, shortHelp, longHelp, argument string) *Options {
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
	// spaces and must be unique in an Options. It is used as the key to
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
func (self *Options) Required(longName, shortName, shortHelp, longHelp, argument string) *Options {
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
// a Command but is instead matched by index of being defined in an Options
// and given as an argument to the Command.
//
// As Indexed Option is given unqualified in the arguments and the argument
// given for the option is the Option Value.
//
// In this example the "indexedOptionValue" is the value for the first Indexed
// Option defined in an Options:
//
// myprox.exe somecommand --namedOption=namedOptionValue indexedOptionValue
//
// Indexed options may be declared at any point in between other types of
// options in an Options but the order in which Indexed options are declared
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
func (self *Options) Indexed(name, shortHelp, longHelp, argument string) *Options {
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
	// spaces and must be unique in an Options. It is used as the key to
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
func (self *Options) Variadic(name, shortHelp, longHelp, arguments string) *Options {
	return self.Register(&Variadic{
		Name:      name,
		ShortHelp: shortHelp,
		LongHelp:  longHelp,
		Arguments: arguments,
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

func (self Boolean) Value() string  { return self.option.value }
func (self Optional) Value() string { return self.option.value }
func (self Required) Value() string { return self.option.value }
func (self Indexed) Value() string  { return self.option.value }
func (self Variadic) Value() string { return self.option.value }

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
		if o.Argument == "" {
			panic("required option requires an argument")
		}
	case *Indexed:
		self.validateKey(o.Name)
		if o.Argument == "" {
			panic("indexed option requires an argument")
		}
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
	value  string
}

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
	// Commands is the Commands to parse. Optional.
	Commands *Commands
	// Globals is the global Options to parse. Optional.
	Globals *Options
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

func (self *Commands) parse(t *arguments) error {
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
		if cmd.Options != nil {
			if err := cmd.Options.parse(t); err != nil {
				return err
			}
			if err := cmd.Handler(cmd.Options); err != nil {
				return err
			}
		} else {
			if err := cmd.Handler(cmd); err != nil {
				return err
			}
		}
		if cmd.SubCommands != nil && cmd.SubCommands.Count() > 0 {
			return cmd.GetSubCommands().parse(t)
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

// Usage prints
func Usage(set *Commands, globals *Options) {

}

// Print prints set and optional globals which can be nil into w.
func Print(w io.Writer, set *Commands, globals *Options) {

}
