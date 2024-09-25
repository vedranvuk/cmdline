// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"context"
	"errors"
	"fmt"
	"io"
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

// defaultOutput is the default output [Config] writes text output to.
var defaultOutput = os.Stdout

// Config is the command line parser configuration.
//
// It contains Command and global Option definitions, Arguments to parse
// and other options packed as an argument to one of the Parse methods.
//
// See [Config.Parse] on details how [Config] is used.
type Config struct {

	// Args are the arguments to parse. This is usually set to os.Args[1:].
	Args Args

	// Usage is a function to call when no arguments are given to Parse.
	//
	// If unset, invokes the built in [Config.PrintUsage]. [Parse] will still
	// return [ErrNoArgs].
	Usage func()

	// Output is the output for printing usage.
	//
	// It is nil by default in which case all the output goes to os.Stdout.
	Output io.Writer

	// Globals are the global Options, independant of any defined commands.
	//
	// They are parsed from arguments that precede any command invocation.
	// Their state can be inspected either from [Config.GlobalsHandler] or
	// after parsing by inspecting the Globals directly.
	Globals Options

	// GlobalsHandler is an optional handler for Globals.
	//
	// It is invoked if any global options get parsed and before any command
	// handlers are invoked.
	//
	// If it returns an error no commands will be parsed and the error
	// propagated back to the caller.
	GlobalsHandler Handler

	// GlobalExclusivityGroups are exclusivity groups for Globals.
	GlobalExclusivityGroups ExclusivityGroups

	// Commands is the root command set.
	Commands Commands

	// LongPrefix is the long Option prefix to use. It is optional and is
	// defaulted to DefaultLongPrefix by Parse() if left empty.
	LongPrefix string

	// ShortPrefix is the short Option prefix to use. It is optional and is
	// defaulted to DefaultShortPrefix by Parse() if left empty.
	ShortPrefix string

	// NoFailOnUnparsedRequired if true, will not return an error if a
	// defined Required or Indexed option was not parsed from arguments.
	//
	// Default: false
	NoFailOnUnparsedRequired bool

	// UseAssignment requires that [Option] value must be given using an
	// assignment operator such that option name is immediately followed by an
	// assignment operator then immediately with the option value.
	// E.g: '--key=value' instead of '--key value'.
	//
	// Default: false
	UseAssignment bool

	// IndexedFirst requires that Indexed Options are specified before any
	// other type of [Option] in [Options].
	//
	// If disabled, arguments to Indexed options may be specifies in between
	// other types of [Option] but in order as they are defined in [Options].
	//
	// Default: false
	IndexedFirst bool

	// ExecAllHandlers specifies that handlers of all commands in the execution
	// chain parsed from Args will be executed in order as specified.
	//
	// First [Handler] in the chain that returns an error stops the chain
	// execution and the error is passed back to the caller.
	//
	// If false only the handler of the last command in the chain is invoked.
	//
	// Default: false
	ExecAllHandlers bool

	// context is the context given to Config.Parse and is set at that time.
	// If nil context was given, Config.Parse sets it to context.Background().
	context context.Context
	// chain is the chain of commands to execute determined by parse.
	chain []*Command
}

// Default returns a new default [Config] starting with args.
func Default(args ...string) *Config {
	return &Config{
		Args:        args,
		LongPrefix:  DefaultLongPrefix,
		ShortPrefix: DefaultShortPrefix,
	}
}

// DefaultOs returns [Default] with os.Args[1:]... as arguments. 
func DefaultOs() *Config { return Default(os.Args[1:]...) }

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

// Parse parses self.Arguments into self.Globals then self.Commands and their
// Options, recursively. Both Globals and Commands are optional and if none are
// defined in the Config, Parse will return nil for any arguments except no
// arguments where it returns ErrNoArgs.
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
// If a Command handler returns an error the parse process is immediately
// aborted and the error propagated back to the Parse/ParseCtx caller.
//
// As Commands and options can be defined declaratively there is no way to
// check for name duplicates at runtime so a validation is performed before the
// parse operation. Once the validation passes, i.e. Parse is called once and
// no validation errors are returned the Config definition is considered
// validated and well formatted. This also ensures validity of any config
// modifications at runtime.
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
	if len(self.Args) == 0 {
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

	var w *wrapper

	// Process Globals
	if err = self.Globals.parse(self); err != nil {
		return
	}
	if err = validateExclusivityGroups(self.GlobalExclusivityGroups, self.Globals); err != nil {
		return
	}
	w = &wrapper{
		self.context,
		self,
		nil,
		nil,
		self.Globals,
	}
	if self.GlobalsHandler != nil {
		if err = self.GlobalsHandler(w); err != nil {
			return
		}
	}

	// Process Commands
	if err = self.Commands.parse(self); err != nil {
		return
	}
	if self.Commands.Count() == 0 || len(self.chain) == 0 {
		return nil
	}
	var (
		parent *Command
		last   = len(self.chain) - 1
	)
	for index, current := range self.chain {
		if self.ExecAllHandlers || index == last {
			w = &wrapper{
				self.context,
				self,
				current,
				parent,
				current.Options,
			}
			if err = current.Handler(w); err != nil {
				return
			}
		}
		parent = current
	}

	return nil
}

// Reset resets the state of all Commands and Options including Globals defined
// in self, recursively. After calling Reset the Config is ready to be parsed.
func (self *Config) Reset() {
	self.Globals.Reset()
	self.Commands.Reset()
}

// Usage prints the default autogenerated usage text to Stdout.
// It is called in the case of no arguments if no Config.Usage is set and may
// be called manually.
func (self *Config) PrintUsage() {

	var (
		out     string
		subs    = false
		program = filepath.Base(os.Args[0])
	)

	out = "Usage:\n\n"

	if self.Globals.Count() > 0 {
		out += fmt.Sprintf("  %s [global options]", program)
	} else {
		out += fmt.Sprintf("  %s", program)
	}

	for _, c := range self.Commands {
		if c.SubCommands.Count() > 0 {
			subs = true
			break
		}
	}

	if subs {
		out += " [command [subcommand...] [options]]\n"
	} else {
		out += " [command [options]]\n"
	}

	out += "\n"
	fmt.Fprint(self.output(), out)
}

// output returns the output to write to.
func (self *Config) output() io.Writer {
	if self.Output != nil {
		return self.Output
	}
	return defaultOutput
}

// wrapper implements [Context].
type wrapper struct {
	context.Context
	config  *Config
	command *Command
	parent  *Command
	options Options
}

// Parsed implements [Context.Parsed].
func (self *wrapper) Parsed(longName string) bool { return self.options.Parsed(longName) }

// Config implements [Context.Config].
func (self *wrapper) Config() *Config { return self.config }

// Command implements [Context.Command].
func (self *wrapper) Command() *Command { return self.command }

// ParentCommand implements [Context.ParentCommand].
func (self *wrapper) ParentCommand() *Command { return self.parent }

// Options implements [Context.Options].
func (self *wrapper) Options() Options { return self.options }

// Values implements [Context.Values].
func (self *wrapper) Values(longName string) Values {
	for _, option := range self.options {
		if option.LongName == longName {
			return option.Values
		}
	}
	return nil
}
