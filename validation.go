// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"fmt"
)

// ValidateOptions validates that Option instances within options have unique
// and non-empty long names and that short names, if not empty, are unique as
// well. Returns nil on success.
func ValidateOptions(options Options) error {
	for _, option := range options {
		switch option.(type) {
		case *Boolean, *Optional, *Required, *Indexed, *Repeated, *Variadic:
		default:
			return errors.New("validation failed: invalid option type, must be a pointer to one of supported option types")
		}
		if option.GetLongName() == "" {
			return errors.New("validation failed: an option with an empty long name is defined")
		}
		for _, other := range options {
			if other != option && other.GetLongName() == option.GetLongName() {
				return fmt.Errorf("validation failed: duplicate option long name: %s", option.GetLongName())
			}
		}
		if option.GetShortName() != "" {
			for _, other := range options {
				if other != option && other.GetShortName() != "" {
					if other.GetShortName() == option.GetShortName() {
						return fmt.Errorf("validation failed: duplicate option short name: %s", option.GetLongName())
					}
				}
			}
		}
	}
	return nil
}

// optionsHaveVariadicOption returns true if options contain at least one
// Variadic Option.
func optionsHaveVariadicOption(options Options) bool {
	for _, opt := range options {
		if _, isVariadic := opt.(*Variadic); isVariadic {
			return true
		}
	}
	return false
}

// ValidateCommands validates that Command instances within commands have
// non-empty and unique names. It validates commands and their SubCommands in
// the same manner recursively. Returns nil on success.
func ValidateCommands(commands Commands) (err error) {
	for _, command := range commands {
		if command.Name == "" {
			return errors.New("validation failed: a command with an empty name is defined")
		}
		if command.Handler == nil {
			return fmt.Errorf("validation failed: command '%s' has no handler assigned", command.Name)
		}
		if command.SubCommands.Count() > 0 {
			for _, opt := range command.Options {
				if _, isVariadic := opt.(*Variadic); isVariadic {
					return fmt.Errorf("validation failed: command '%s' contains a variadic option and may have no sub-commands", command.Name)
				}
			}
		}
		for _, other := range commands {
			if other != command && other.Name == command.Name {
				return fmt.Errorf("validation failed: duplicate command name: %s", command.Name)
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

// validateCommandExclusivityGroups calls validateExclusivityGroups for command.
func validateCommandExclusivityGroups(command *Command) (err error) {
	if err = validateExclusivityGroups(command.ExclusivityGroups, command.Options); err != nil {
		return fmt.Errorf("command '%s' %w", command.Name, err)
	}
	return
}

// validateCommandExclusivityGroups returns nil if parsed options do not satisfy
// any of the defined groups or an error otherwise.
func validateExclusivityGroups(groups ExclusivityGroups, options Options) error {
	var conflict string
	for _, group := range groups {
		conflict = ""
		for _, name := range group {
			if options.IsParsed(name) {
				if conflict != "" {
					return fmt.Errorf("options '%s' and '%s' are mutually exclusive", conflict, name)
				}
				conflict = name
			}
		}
	}
	return nil
}
