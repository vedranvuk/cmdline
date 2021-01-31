// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"fmt"
	"os"
)

// ErrCmdline is the base error of cmdline package.
var ErrCmdline = errors.New("cmdline")

// ErrConvert is the base string to go value conversion error.
var ErrConvert = fmt.Errorf("%w: convert", ErrCmdline)

// ErrRegister is the base Command or Parameter registration error.
var ErrRegister = fmt.Errorf("%w: register", ErrCmdline)

var (
	// ErrParse is the base error returned by Parse methods.
	ErrParse = fmt.Errorf("%w: parse error", ErrCmdline)

	// ErrInvalidArgument is returned when an invalid or malformed argument
	// was encountered.
	ErrInvalidArgument = fmt.Errorf("%w: invalid argument", ErrParse)
	// ErrNoArguments is returned if arguments were  empty and there were
	// defined commands or parameters.
	ErrNoArguments = fmt.Errorf("%w: no arguments", ErrParse)
	// ErrExtraArguments is returned when there were unparsed arguments left.
	ErrExtraArguments = fmt.Errorf("%w: extra arguments", ErrParse)

	// ErrNoDefinitions is returned if no Commands or Parameters were defined.
	ErrNoDefinitions = fmt.Errorf("%w: no definitions", ErrParse)
	// ErrCommandNotFound is returned when a Command or Parameter is not found.
	ErrCommandNotFound = fmt.Errorf("%w: command not found", ErrParse)
	// ErrParameterNotFound is returned when a Command or Parameter is not found.
	ErrParameterNotFound = fmt.Errorf("%w: parameter not found", ErrParse)
	// ErrParsed is returned when a parameter marked as parsed is parsed again.
	ErrParsed = fmt.Errorf("%w: parameter already parsed", ErrParse)
	// ErrValueRequired is returned when a parameter is missing a value.
	ErrValueRequired = fmt.Errorf("%w: parameter requires value", ErrParse)
	// ErrParameterRequired is returned if a required parameter was not parsed.
	ErrParameterRequired = fmt.Errorf("%w: required parameter not specified", ErrParse)
)

// ErrNoMatch is returned by Visit if there are no matches.
var ErrNoMatch = fmt.Errorf("%w: no matches", ErrCmdline)

// ErrNoHandler is returned when a Command does not have a handler set and
// it is required.
var ErrNoHandler = fmt.Errorf("%w: no handler", ErrCmdline)

// Parse parses commands from args and executes handlers of commands
// along the internally constructed match Chain. If a handler returns an error
// iteration is stopped and that error is returned.
// If all handlers return nil, result is nil.
// All handlers in the match chain may be nil.
//
// Context.Next will point to next Context in match Chain except for last
// Context whose Next() will return nil.
//
// If there are unparsed arguments left ErrExtraArguments is returned.
// If args are empty ErrNoArguments is returned.
// If there were no matches ErrNoMatch is returned.
// If an argument is invalid ErrInvalidArgument descendant is returned.
// If last command does not have a handler ErrNoHandler descendant is returned.
// If any other parse error occurs it is returned.
func Parse(commands *Commands, args ...string) error {
	var err error
	var chain = NewStaticChain()
	var arguments = NewArguments(args...)
	if err = commands.Parse(arguments, chain); err != nil {
		return err
	}
	var matchcount = chain.Length()
	if matchcount < 1 {
		return ErrNoMatch
	}
	var contexts = make([]context, 0, matchcount)
	var i int
	for i = 0; i < matchcount; i++ {
		contexts = append(contexts, context{
			Command: chain.Index(i),
		})
	}
	for i = 0; i < matchcount-1; i++ {
		contexts[i].next = &contexts[i+1]
	}
	contexts[matchcount-1].arguments = arguments.Slice()
	for i = 0; i < matchcount; i++ {
		if err = contexts[i].ExecuteHandler(); err != nil {
			if errors.Is(err, ErrNoHandler) {
				return nil
			}
			return err
		}
	}
	return nil
}

// ParseRaw parses commands from args and executes handlers of commands
// along the internally constructed match Chain. If a handler returns an error
// iteration is stopped and that error is returned.
// If all handlers return nil, result is nil.
// All handlers in the match chain except last commands' may be nil.
//
// Context.Next will point to next Context in match Chain except for last
// Context whose Next() will return nil.
//
// Unparsed arguments do not return ErrExtraArguments. Instead, they are passed
// to the last handler in match chain via Context.Extra().
//
// If args are empty ErrNoArguments is returned.
// If there were no matches ErrNoMatch is returned.
// If an argument is invalid ErrInvalidArgument descendant is returned.
// If last command does not have a handler ErrNoHandler descendant is returned.
// If any other parse error occurs it is returned.
func ParseRaw(commands *Commands, args ...string) error {
	var err error
	var chain = NewStaticChain()
	var arguments = NewArguments(args...)
	if err = commands.Parse(arguments, chain); err != nil {
		if !errors.Is(err, ErrExtraArguments) {
			return err
		}
	}
	var matchcount = chain.Length()
	if matchcount < 1 {
		return ErrNoMatch
	}
	var contexts = make([]context, 0, matchcount)
	var i int
	for i = 0; i < matchcount; i++ {
		contexts = append(contexts, context{
			Command: chain.Index(i),
		})
	}
	for i = 0; i < matchcount-1; i++ {
		contexts[i].next = &contexts[i+1]
	}
	contexts[matchcount-1].arguments = arguments.Slice()
	for i = 0; i < matchcount; i++ {
		if err = contexts[i].ExecuteHandler(); err != nil {
			if contexts[i].Command != chain.Last() && errors.Is(err, ErrNoHandler) {
				return nil
			}
			return err
		}
	}
	return nil
}

// ParseOS is like Parse but parses os.Args[1:].
func ParseOS(commands *Commands) error {
	return Parse(commands, os.Args[1:]...)
}

// ParseOSRaw is like ParseRaw but parses os.Args[1:].
func ParseOSRaw(commands *Commands) error {
	return ParseRaw(commands, os.Args[1:]...)
}
