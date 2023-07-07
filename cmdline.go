package cmdline

import (
	"errors"
	"fmt"
)

var (
	// ErrNoArgs is returned by Parse if no arguments were given for parsing.
	ErrNoArgs = errors.New("no arguments")
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
	// LongPrefix is the long Option prefix to use. It is optional and is
	// defaulted to DefaultLongPrefix by Parse() if left empty.
	LongPrefix string
	// ShortPrefix is the short Option prefix to use. It is optional and is
	// defaulted to DefaultShortPrefix by Parse() if left empty.
	ShortPrefix string
}

const (
	// DefaultLongPrefix is the default prefix for long option names.
	DefaultLongPrefix = "--"
	// DefaultShortPrefix is the default prefix for short option names.
	DefaultShortPrefix = "-"
)

// Parse parses config.Arguments into config.Globals then config.Commands.
// It returns nil on success or an error if one occured.
func Parse(config *Config) (err error) {
	if len(config.Arguments) == 0 {
		return ErrNoArgs
	}
	if err = validateOptions(config.Globals); err != nil {
		return
	}
	if err = validateCommands(config.Commands); err != nil {
		return
	}
	if config.LongPrefix == "" {
		config.LongPrefix = DefaultLongPrefix
	}
	if config.ShortPrefix == "" {
		config.ShortPrefix = DefaultShortPrefix
	}
	if err = config.Globals.parse(config); err != nil {
		return
	}
	if config.GlobalsHandler != nil {
		if err = config.GlobalsHandler(config.Globals); err != nil {
			return
		}
	}
	return config.Commands.parse(config)
}

// validateOptions validates that options have unique and non-empty long names
// in the set and that short names, if not empty, are unique as well.
// Returns nil on success.
func validateOptions(options Options) error {
	for _, option := range options {
		if option.GetLongName() == "" {
			return errors.New("an option with an empty long is defined")
		}
		for _, other := range options {
			if other != option && other.GetLongName() == option.GetLongName() {
				return fmt.Errorf("duplicate option long name: %s", option.GetLongName())
			}
		}
		if option.GetShortName() != "" {
			for _, other := range options {
				if other != option && other.GetShortName() != "" {
					if other.GetShortName() == option.GetShortName() {
						return fmt.Errorf("duplicate option long name: %s", option.GetLongName())
					}
				}
			}
		}
	}
	return nil
}

// validateCommands validates that commands (and their sub command sets) have
// non-empty and unique names in their respective sets. Returns nil on success.
func validateCommands(commands Commands) (err error) {
	for _, command := range commands {
		if command.Name == "" {
			return errors.New("a command with an empty name is defined")
		}
		for _, other := range commands {
			if other != command && other.Name == command.Name {
				return fmt.Errorf("duplicate command name: %s", command.Name)
			}
		}
		if err = validateOptions(command.Options); err != nil {
			return
		}
		if err = validateCommands(command.SubCommands); err != nil {
			return
		}
	}
	return nil
}
