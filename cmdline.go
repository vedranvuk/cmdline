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
// Parse methods.
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
	Usage func()
	// FailOnUnparsedRequiredOption if true, will return an error if a
	// Required or Indexed option was not parsed from arguments.
	// Default: true
	FailOnUnparsedRequiredOption bool
	// LongPrefix is the long Option prefix to use. It is optional and is
	// defaulted to DefaultLongPrefix by Parse() if left empty.
	LongPrefix string
	// ShortPrefix is the short Option prefix to use. It is optional and is
	// defaulted to DefaultShortPrefix by Parse() if left empty.
	ShortPrefix string

	// context is the context passed to all commands being executed.
	context context.Context
}

// Parse parses config.Arguments into config.Globals then config.Commands.
// It returns nil on success or an error if one occured.
func Parse(config *Config) error { return config.Parse(context.Background()) }

// Parse parses config.Arguments into config.Globals then config.Commands.
// It returns nil on success or an error if one occured.
func ParseCtx(ctx context.Context, config *Config) error { return config.Parse(ctx) }

// Parse parses self.Arguments into self.Globals then self.Commands.
// It returns nil on success or an error if one occured.
func (self *Config) Parse(ctx context.Context) (err error) {
	self.context = ctx
	if len(self.Arguments) == 0 {
		if self.Usage != nil {
			self.Usage()
			return
		}
		var exe = filepath.Base(os.Args[0])
		fmt.Printf("Usage: %s [global-options] [...command [...command-option]]\n", exe)
		fmt.Printf("Type '%s help' for more help.\n", exe)
		return nil
	}
	if err = ValidateOptions(self.Globals); err != nil {
		return
	}
	if err = ValidateCommands(self.Commands); err != nil {
		return
	}
	if self.LongPrefix == "" {
		self.LongPrefix = DefaultLongPrefix
	}
	if self.ShortPrefix == "" {
		self.ShortPrefix = DefaultShortPrefix
	}
	if err = self.Globals.parse(self); err != nil {
		return
	}
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