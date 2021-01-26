// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/vedranvuk/strconvex"
)

// Parameter defines a Command parameter. It updates and maintains its state.
//
// Parameter is defined as Named or Indexed and is parsed differently depending.
// Named parameters are addressed by name on command line and can be specified
// in any order.
// e.g. "-v --name root --force".
// Indexed parameters are matched in order as they are defined.
// e.g. "value1 value2 value3".
// For details see Parameters.Parse.
//
// Parameter addiotionally defines:
// A help text which is displayed when parameter is printed.
// An interface to a Go value to be set from parsed parameter value.
type Parameter struct {
	// name is the parameter name.
	name string
	// help is the Param help text.
	help string
	// required specifies if this Param is required.
	required bool
	// indexed specifies if this param is an indexed param.
	indexed bool
	// value is a pointer to a Go value which is set
	// from parsed Param value if not nil and points to a
	// valid value.
	value interface{}

	// parsed indicates if Param was parsed from arguments.
	parsed bool
	// rawvalue is the raw parameter value argument, possibly empty.
	rawvalue string
}

// NewParameter returns a new Parameter initialized with given name and help,
// marked as required and/or indexed with value that receives argument value.
//
// Optional value must be a pointer to a Go value into which an optional
// Parameter argument string can be parsed according to rules defined in
// github.com/vedranvuk/strconvex.
func NewParameter(name, help string, required, indexed bool, value interface{}) *Parameter {
	return &Parameter{
		name:     name,
		help:     help,
		required: required,
		indexed:  indexed,
		value:    value,
	}
}

// Reset resets Parameter state.
func (p *Parameter) Reset() {
	p.rawvalue = ""
	p.parsed = false
}

// Name returns parameter name.
func (p *Parameter) Name() string { return p.name }

// Help returns parameter help.
func (p *Parameter) Help() string { return p.help }

// Required returns if the Parameter is defined as required.
func (p *Parameter) Required() bool { return p.required }

// Indexed returns if the parameter is defined as Indexed.
func (p *Parameter) Indexed() bool { return p.indexed }

// Parsed returns if the Parameter is marked as parsed.
func (p *Parameter) Parsed() bool { return p.parsed }

// Value returns the value registered with Parameter, which could be nil.
func (p *Parameter) Value() interface{} { return p.value }

// RawValue returns the raw argument string value parsed as Parameter value.
// Result is valid if Parsed returns true.
func (p *Parameter) RawValue() string { return p.rawvalue }

// Parse parses Parameter from arguments and optionally returns an error.
// Parameter will never be marked parsed in case of an error. On successfull
// Parse a nil error is returned and Parameter may or may not be marked parsed.
//
// If Parameter is already marked parsed returns ErrParsed descendant.
//
// If a conversion of value argument to Parameter target value fails an
// ErrConvert descendant is returned.
//
// Named and Indexed Parameters are parsed differently:
//
// Named Parameters
//
// If current argument is not a LongArgument or a ShortArgument Parameter is not
// parsed and nil is returned.
//
// If Parameter does not require a value, i.e. it is not marked as required and
// its target value is not set it is marked as parsed and Parse returns nil.
//
// If Parameter requires a value, i.e. it is either marked as required or has a
// target value set, arguments are advanced by one argument and that argument is
// read as a Parameter value. If there are no arguments left or argument is not
// TextArgument an ErrValueRequired descendant is returned.
//
// Indexed Parameters
//
// If Parameter is required and argument kind is not TextArgument returns an
// ErrValueRequired.
//
// If Parameter is optional and argument kind is not TextArgument returns nil
// and Parameter is not marked parsed.
//
func (p *Parameter) Parse(args Arguments) (err error) {
	if p.Parsed() {
		return fmt.Errorf("%w: '%s'", ErrParsed, p.name)
	}
	switch args.Kind() {
	case InvalidArgument:
		return ErrInvalidArgument
	case NoArgument:
		return ErrNoArguments
	case TextArgument, CombinedArgument:
		if p.Indexed() {
			goto checkValueArgument
		}
		return nil
	default:
		if p.Indexed() {
			return nil
		}
	}
	if args.Name() != p.name {
		return nil
	}
	if p.Required() || p.Value() != nil {
		if !args.Advance() {
			goto valueArgumentRequired
		}
		goto checkValueArgument
	}
	goto parseOK
checkValueArgument:
	switch args.Kind() {
	case InvalidArgument:
		return ErrInvalidArgument
	case TextArgument:
		goto parseValueArgument
	default:
		if p.Indexed() && !p.Required() {
			return nil
		}
		goto valueArgumentRequired
	}
parseValueArgument:
	if p.value != nil {
		if err = strconvex.StringToInterface(args.Name(), p.value); err != nil {
			return fmt.Errorf("%w: %v", ErrConvert, err)
		}
	}
	p.rawvalue = args.Raw()
parseOK:
	p.parsed = true
	args.Advance()
	return
valueArgumentRequired:
	return fmt.Errorf("%w: '%s'", ErrValueRequired, p.name)
}

// Parameters define a set of Command Parameters unique by name.
//
type Parameters struct {
	// parent is the reference to owner *Command.
	parent *Command
	// longparams is a map of long param name to *Param.
	longparams nameToParameter
	// shortparams is a map of short param name to *Param.
	shortparams nameToParameter
	// longtoshort maps a long param name to short param name.
	longtoshort longToShort
	// longindexes hold long param names in order as they are registered.
	longindexes []string
}

// nameToParameter maps a param name to *Param.
type nameToParameter map[string]*Parameter

// longToShort maps a long param name to short param name.
type longToShort map[string]string

// NewParameters returns a new Parameters instance for specified command which
// can be nil.
func NewParameters(parent *Command) *Parameters {
	return &Parameters{
		parent:     parent,
		longparams:  make(nameToParameter),
		shortparams: make(nameToParameter),
		longtoshort: make(longToShort),
		longindexes: []string{},
	}
}

// Reset resets all Parameters.
func (p *Parameters) Reset() {
	var param *Parameter
	for _, param = range p.longparams {
		param.Reset()
	}
}

// Length returns number of defined parameters.
func (p *Parameters) Length() int { return len(p.longindexes) }

// Command returns parent Command.
func (p *Parameters) Command() *Command { return p.parent }

// HasOptionalIndexedParameters returns if Parameters contain one or more
// defined indexed Parameters that are optional.
//
// If it has, Command owning these Parameters may not have sub commands.
func (p *Parameters) HasOptionalIndexedParameters() bool {
	var param *Parameter
	for _, param = range p.longparams {
		if param.indexed && !param.required {
			return true
		}
	}
	return false
}

// AddNamed registers a new prefixed Param in these Parameters.
//
// Long param name is required, short is optional and can be empty, as is help.
//
// If required is specified value must be a pointer to a supported Go value
// which will be updated to a value parsed from an argument following param.
// If a required Param or its' value is not found in command line args an error
// is returned.
//
// If Param is not marked as required, specifying a value parameter is optional
// but dictates that:
// If nil, a value for the Param will not be parsed from args.
// If valid, the parser will parse the argument following the Param into it.
//
// Registration order is important. Prefixed params must be registered before
// raw params and are printed in order of registration.
//
// If an error occurs Param is not registered.
func (p *Parameters) AddNamed(name, shortname, help string, required bool, value interface{}) error {
	return p.addParam(name, shortname, help, required, false, value)
}

// MustAddNamed is like AddParam except the function panics on error.
// Returns a Command that the param was added to.
func (p *Parameters) MustAddNamed(name, shortname, help string, required bool, value interface{}) *Command {
	var err error
	if err = p.AddNamed(name, shortname, help, required, value); err != nil {
		panic(err)
	}
	return p.parent
}

// AddIndexed registers a raw Param under specified name which must be unique
// in long Parameters names. Raw params can only be defined after prefixed
// params or other raw params. Calls to AddParam after AddIndexed will error.
//
// Registering a raw parameter for a command will disable command's ability to
// have sub commands registered as its invocation would be ambiguous with raw
// parameters during Parameters parsing. If the command already has sub commands
// registered the function will error.
//
// Parsed arguments are applied to registered raw Parameters in order as they are
// defined. If value is a pointer to a valid Go value argument will be converted
// to that Go value. Specifying a value is optional and if nil, parsed argument
// will not be parsed into the value.
//
// Marking a raw param as required does not imply that value must not be nil
// as is in prefixed params. Required flag solely returns a parse error if
// required raw param was not parsed and value is set only if non-nil.
//
// A single non-required raw Param is allowed and it must be the last one.
//
// If an error occurs it is returned and the Param is not registered.
func (p *Parameters) AddIndexed(name, help string, required bool, value interface{}) error {
	return p.addParam(name, "", help, required, true, value)
}

// MustAddIndexed is like AddRawParam except the function panics on error.
// Returns a Command that the param was added to.
func (p *Parameters) MustAddIndexed(name, help string, required bool, value interface{}) *Command {
	var err error
	if err = p.AddIndexed(name, help, required, value); err != nil {
		panic(err)
	}
	return p.parent
}

// Get returns the Parameter under specified name and nil if found.
// If parameter is not found error will be a ErrNotFound descendant.
func (p *Parameters) Get(name string) (*Parameter, error) {
	var param *Parameter
	var exists bool
	if param, exists = p.longparams[name]; !exists {
		return nil, fmt.Errorf("%w: '%s'", ErrCommandNotFound, name)
	}
	return param, nil
}

// MustGet is like Get but panics if parameter is not found.
func (p *Parameters) MustGet(name string) *Parameter {
	var parameter *Parameter
	var err error
	if parameter, err = p.Get(name); err != nil {
		panic(err)
	}
	return parameter
}

// Parse parses states of defined Parameters from arguments which it updates.
// Any arguments parsed into a Parameter or its value are advanced.
//
// Returns ErrInvalidArgument if next argument read is invalid.
// Returns ErrNoDefinitions if there are no defined parameters.
// Returns ErrNotFound descendant if parameter is not found.
// Returns ErrParameterRequired descendant if a required parameter was not parsed.
// Returns ErrParse descendant on any other error.
// Returns nil if all required parameters were parsed and there were no errors.
//
// Parse parses Named parameters first until it encounters a TextArgument which
// could be a command name or an argument to a defined Indexed parameter. If
// there are defined Indexed parameters, Parse parses that argument as Indexed
// parameter's value.
func (p *Parameters) Parse(args Arguments) error {
	var argcount = args.Length()
	var paramcount int = p.Length()
	if paramcount == 0 {
		return ErrNoDefinitions
	}
	var err error
	var param *Parameter
	var exists bool
	var i int
	for i = 0; i < paramcount; {
		switch args.Kind() {
		case InvalidArgument:
			return ErrInvalidArgument
		case NoArgument:
			goto checkRequired
		case TextArgument:
			// Start of raw params, skip prefixed.
			for i < paramcount {
				if param = p.longparams[p.longindexes[i]]; !param.indexed {
					i++
					continue
				}
				break
			}
			// No defined raw params, assume sub command name.
			if i >= paramcount {
				goto checkRequired
			}
			i++
		case ShortArgument:
			if param, exists = p.shortparams[args.Name()]; !exists {
				return fmt.Errorf("%w: short '%s'", ErrParameterNotFound, args.Name())
			}
			i++
		case LongArgument:
			if param, exists = p.longparams[args.Name()]; !exists {
				return fmt.Errorf("%w: long '%s'", ErrParameterNotFound, args.Name())
			}
			i++
		case CombinedArgument:
			// Parse all combined args and continue.
			var shorts = strings.Split(args.Name(), "")
			var short string
			for _, short = range shorts {
				if param, exists = p.shortparams[short]; !exists {
					return fmt.Errorf("%w: short '%s'", ErrParameterNotFound, short)
				}
				if param.value != nil {
					return fmt.Errorf("%w: short parameter '%s' requires argument, cannot combine", ErrParse, short)
				}
				// Param is specified multiple times.
				if param.parsed {
					return fmt.Errorf("%w: combined parameter '%s' specified multiple times", ErrParse, short)
				}
				param.parsed = true
				i++
			}
			args.Advance()
			continue
		}
		// Pass arguments to Parameter.
		if err = param.Parse(args); err != nil {
			if errors.Is(err, ErrNoArguments) {
				break
			}
			return err
		}
	}
checkRequired:
	// Check all required params were parsed.
	var arg string
	for arg, param = range p.longparams {
		if param.required && !param.parsed {
			return fmt.Errorf("%w: '%s'", ErrParameterRequired, arg)
		}
	}
	if argcount == 0 {
		return ErrNoArguments
	}
	return nil
}

// addParam is the generalized parameter registration method.
func (p *Parameters) addParam(name, shortname, help string, required, indexed bool, value interface{}) error {
	// Long name must not be empty and short name must be max one char long.
	// TODO Name must be ASCII and may contain hyphen but may not start with it.
	if name == "" || len(shortname) > 1 {
		return fmt.Errorf("%w: invalid name", ErrRegister)
	}
	// No long duplicates.
	var ok bool
	if _, ok = p.longparams[name]; ok {
		return fmt.Errorf("%w: duplicate long parameter name '%s'", ErrRegister, name)
	}
	// No short duplicates if not empty.
	if _, ok = p.shortparams[shortname]; ok && shortname != "" {
		return fmt.Errorf("%w: duplicate short parameter name '%s'", ErrRegister, shortname)
	}
	// Disallow adding optional raw parameters if command expects sub commands.
	if p.parent != nil && p.parent.Commands().Length() > 0 && indexed && !required {
		return fmt.Errorf("%w: cannot register optional raw parameter on a command with sub commands", ErrRegister)
	}
	// Indexed parameters can only be registered after named parameters.
	// Optional indexed parameters can only be registered after any required
	// indexed parameters.
	var param *Parameter
	if p.Length() > 0 {
		if param = p.longparams[p.longindexes[len(p.longindexes)-1]]; param != nil && param.indexed {
			if !indexed {
				return fmt.Errorf("%w: cannot register named parameter after indexed parameter", ErrRegister)
			}
			if !param.required {
				if !required {
					return fmt.Errorf("%w: cannot register multiple optional parameters", ErrRegister)
				}
				return fmt.Errorf("%w: cannot register required after optional parameter", ErrRegister)
			}
		}
	}
	// Required named parameters need a valid Go value.
	if value == nil {
		if !indexed && required {
			return fmt.Errorf("%w: value required", ErrRegister)
		}
	} else {
		// Value must be a valid pointer to a Go value.
		if v := reflect.ValueOf(value); !v.IsValid() || v.Kind() != reflect.Ptr {
			return fmt.Errorf("%w: invalid value", ErrRegister)
		}
	}
	// Register a new parameter.
	param = NewParameter(name, help, required, indexed, value)
	p.longparams[name] = param
	if shortname != "" {
		p.shortparams[shortname] = param
	}
	p.longtoshort[name] = shortname
	p.longindexes = append(p.longindexes, name)
	return nil
}
