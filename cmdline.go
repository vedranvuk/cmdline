package cmdline

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	// ErrNoArgs is returned by Parse if no arguments were given for parsing.
	ErrNoArgs = errors.New("no arguments")
)

const (
	// DefaultLongPrefix is the default prefix for long option names.
	DefaultLongPrefix = "--"
	// DefaultShortPrefix is the default prefix for short option names.
	DefaultShortPrefix = "-"
)

// Config contains Command and global Option definitions, Arguments to parse
// and other options packed into a single struct for use as an argument to
// one of the Parse methods.
type Config struct {
	// Arguments are the arguments to parse. This is usually set to os.Args[1:].
	Arguments Arguments
	// Globals are the global Options, independant of any defined commands.
	Globals Options
	// GlobalsHandler is the handler for Globals.
	// It is optional and gets invoked before any commands are parsed.
	GlobalsHandler Handler
	// Commands is the root command set.
	Commands Commands
	// Usage is a function to call when no arguments are given to Parse.
	// If unset, invokes the built in Usage func.
	// Parse functions will still return ErrNoArgs.
	Usage func()
	// NoFailOnUnparsedRequired if true, will not return an error if a
	// defined Required or Indexed option was not parsed from arguments.
	// Defaults to false.
	NoFailOnUnparsedRequired bool
	// LongPrefix is the long Option prefix to use. It is optional and is
	// defaulted to DefaultLongPrefix by Parse() if left empty.
	LongPrefix string
	// ShortPrefix is the short Option prefix to use. It is optional and is
	// defaulted to DefaultShortPrefix by Parse() if left empty.
	ShortPrefix string

	// context is the context given to Config.Parse and is set at that time.
	// If nil context was given, Config.Parse sets it to context.Background().
	context context.Context
}

// Parse parses config.Arguments into config.Globals then config.Commands.
// It returns nil on success or an error if one occured.
// It invokes Config.Parse with a context.Background().
// See Config.Parse for more details.
func Parse(config *Config) error { return config.Parse(context.Background()) }

// Parse parses config.Arguments into config.Globals then config.Commands.
// It returns nil on success or an error if one occured.
// It invokes COnfig.Parse with the specified ctx.
// See Config.Parse for more details.
func ParseCtx(ctx context.Context, config *Config) error { return config.Parse(ctx) }

// Usage prints the default autogenerated usage text to Stdout.
// It is called in the case of no arguments if no Config.Usage is set and may
// be called manually.
func (self *Config) PrintUsage() {
	var exe = filepath.Base(os.Args[0])
	fmt.Printf("Usage: %s [global-options] [...command [...command-option]]\n", exe)
	fmt.Printf("Type '%s help' for more help.\n", exe)
}

// Parse parses self.Arguments into self.Globals then self.Commands and their
// Options, recursively. Both Globals and Commands are optional and if none are
// defined in the Config, Parse will return nil.
//
// Parse will first parse Config.Globals and call the GlobalsHandler if set
// after they have been parsed. If no GlobalsHandler is set, Globals may be
// inspected manually from the Config.
//
// Parse then continues matching the following argument to a Command,
// parsing Options for that Command and then calls the Command's Handler.
// If the Command contains sub Commands and there are unparsed arguments left,
// it continues parsing arguments into that Command's sub Commands.
//
// If an undefined Command or Option was specified, either due to a typo or
// malformatted arguments Parse will return a descriptive error.
//
// If no arguments were given Parse will call Config.Usage if set or print a
// default autogenerated usage text if not and in both cases return ErrNoArgs.
//
// As Commands and options can be defined declaratively, there is no way to
// check for name duplicates at runtime so a validation is performed before the
// parse operation. Once the validation passes, i.e. Parse is called once and
// no validation errors are returned the Config definition is considered
// validated and well formatted.
//
// Following validation rules are enforced before the parse process and will
// return a descriptive error if any validation fails:
//
// * There may be no duplicate Command instance names per their Commands group.
// * There may be no duplicate Option instance names within their Options group.
// For more details see Command, Option, ValidateOptions and ValidateCommands.
//
// * If the Config.Globals Options set contains a Variadic Option which consumes
// all following arguments as value to self, there may be no Command instances
// defined in Commands.
//
// * If a Commands set at the root Config.Commands or any sub level contains a
// Variadic Option definition, it may not have any sub commands as Variadic
// Option consumes all remaining arguments as its values and stops further
// sub Command parsing.
func (self *Config) Parse(ctx context.Context) (err error) {

	// Verify and store context.
	if self.context = ctx; self.context == nil {
		self.context = context.Background()
	}

	// No arguments case.
	// Call Usage or print default text.
	if len(self.Arguments) == 0 {
		if self.Usage != nil {
			self.Usage()
			return ErrNoArgs
		}
		self.PrintUsage()
		return ErrNoArgs
	}

	// Validation.
	if self.Commands.Count() > 0 && optionsHaveVariadicOption(self.Globals) {
		return errors.New("validation failed: globals contain a variadic option with command definitions present")
	}
	if err = ValidateOptions(self.Globals); err != nil {
		return
	}
	if err = ValidateCommands(self.Commands); err != nil {
		return
	}

	// Verify and set defaults.
	if self.LongPrefix == "" {
		self.LongPrefix = DefaultLongPrefix
	}
	if self.ShortPrefix == "" {
		self.ShortPrefix = DefaultShortPrefix
	}

	// Parse Globals.
	if err = self.Globals.parse(self); err != nil {
		return
	}

	// Parse Commands.
	var wrapper = &contextWrapper{
		self.context,
		self.Globals,
	}
	if self.GlobalsHandler != nil {
		if err = self.GlobalsHandler(wrapper); err != nil {
			return
		}
	}

	return self.Commands.parse(self)
}
