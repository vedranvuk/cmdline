package cmdline

import (
	"errors"
	"fmt"
)

// ValidateOptions validates that options have unique and non-empty long names
// in the set and that short names, if not empty, are unique as well.
// Returns nil on success.
func ValidateOptions(options Options) error {
	for _, option := range options {
		switch option.(type) {
		case *Boolean, *Optional, *Required, *Indexed, *Variadic:
		default:
			return errors.New("invalid option type, must be a pointer to one of supported option types")
		}
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

// ValidateCommands validates that commands (and their sub command sets) have
// non-empty and unique names in their respective sets. Returns nil on success.
func ValidateCommands(commands Commands) (err error) {
	for _, command := range commands {
		if command.Name == "" {
			return errors.New("a command with an empty name is defined")
		}
		if command.Handler == nil {
			return fmt.Errorf("command '%s' has no handler assigned", command.Name)
		}
		for _, other := range commands {
			if other != command && other.Name == command.Name {
				return fmt.Errorf("duplicate command name: %s", command.Name)
			}
		}
		if err = ValidateOptions(command.Options); err != nil {
			return
		}
		if err = ValidateCommands(command.SubCommands); err != nil {
			return
		}
	}
	return nil
}
