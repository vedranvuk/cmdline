// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"encoding"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vedranvuk/strutils"
)

// Args is a slice of strings containing arguments to parse and implements
// a few helpers to tokenize, scan and process the slice.
//
// It is used by [Commands.parse] and [Options.parse].
type Args []string

// Kind returns the current argument kind.
func (self Args) Kind(config *Config) (kind Argument) {
	if self.Eof() {
		return NoArgument
	}
	kind = TextArgument
	// in case of "-" as short and "--" as long, long wins.
	if strings.HasPrefix(self.First(), config.ShortPrefix) {
		kind = ShortArgument
	}
	if strings.HasPrefix(self.First(), config.LongPrefix) {
		kind = LongArgument
	}
	return
}

// Argument defines the kind of argument being parsed.
type Argument int

const (
	// NoArgument indicates no argument. It's returned if Arguments are empty.
	NoArgument Argument = iota
	// LongArgument indicates a token with a long option prefix.
	LongArgument
	// ShortArgument indicates a token with a short option prefix.
	ShortArgument
	// TextArgument indicates a text token with no prefix.
	TextArgument
)

// First returns first element of the slice, unmodified.
// If slice is empty returns an empty string.
func (self Args) First() string {
	if self.Eof() {
		return ""
	}
	return self[0]
}

// Text returns first element of the slice stripped of option prefixes.
// If slice is empty returns an empty string.
func (self Args) Text(config *Config) string {
	switch k := self.Kind(config); k {
	case ShortArgument:
		return string(self.First()[len(config.ShortPrefix):])
	case LongArgument:
		return string(self.First()[len(config.LongPrefix):])
	case TextArgument:
		return self.First()
	}
	return ""
}

// Next discards the first element of the slice.
func (self *Args) Next() *Args {
	if self.Eof() {
		return self
	}
	*self = (*self)[1:]
	return self
}

// Clear clears the slice.
func (self *Args) Clear() { *self = Args{} }

// Eof returns true if there are no more elements in the slice.
func (self Args) Eof() bool { return len(self) == 0 }

// parse parses [Commands] from config.Args or returns an error.
func (self Commands) parse(config *Config) (err error) {
	switch kind, name := config.Args.Kind(config), config.Args.Text(config); kind {
	case NoArgument:
		return nil
	case LongArgument, ShortArgument:
		return errors.New("expected command name, got option")
	case TextArgument:
		var cmd = self.Find(name)
		if cmd == nil {
			return fmt.Errorf("command '%s' not registered", name)
		}
		config.Args.Next()
		if err = cmd.Options.parse(config); err != nil {
			return
		}
		if err = validateCommandExclusivityGroups(cmd); err != nil {
			return
		}
		cmd.executed = true
		config.chain = append(config.chain, cmd)
		if err = cmd.SubCommands.parse(config); err != nil {
			return
		}
		if cmd.RequireSubExecution && cmd.SubCommands.Count() > 0 && !cmd.SubCommands.AnyExecuted() {
			return fmt.Errorf("command '%s' requires execution of one of its subcommands", cmd.Name)
		}
	}
	return nil
}

// parse parses [Options] from config.Args or returns an error.
func (self Options) parse(config *Config) (err error) {

	if self.Count() == 0 {
		return nil
	}

	var (
		opt        *Option
		key, val   string
		assignment bool
		combined   string
	)

	for !config.Args.Eof() {

		// Parse key and val.
		if config.UseAssignment {
			key, val, assignment = strings.Cut(config.Args.Text(config), "=")
			key = strings.TrimSpace(key)
			if assignment && val != "" {
				val, _ = strutils.UnquoteDouble(strings.TrimSpace(val))
			}
		} else {
			key = strings.TrimSpace(config.Args.Text(config))
			val = ""
			assignment = false
		}

		// Try to detect the option by argument kind.
		// If not prefixed see if theres defined and yet unparsed indexed.
		// If prefixed see if its Boolean, Optional, Required or Repeated.
		switch kind := config.Args.Kind(config); kind {
		case TextArgument:
			if !config.UseAssignment && opt != nil {
				switch opt.Kind {
				case Optional, Required, Repeated:
				default:
					return fmt.Errorf("option '%s' requires a value", opt.LongName)

				}
			} else {
				if o := self.getFirstUnparsedIndexed(); o != nil {
					opt = o
				}
			}
		case LongArgument:
			if config.IndexedFirst {
				if fui := self.getFirstUnparsedIndexed(); fui != nil {
					return fmt.Errorf("indexed argument '%s' not parsed", fui.LongName)
				}
			}

			if !config.UseAssignment && opt != nil {
				return fmt.Errorf("option '%s' requires a value", opt.LongName)
			}

			if opt = self.FindLong(key); opt == nil {
				return fmt.Errorf("unknown option '%s'", key)
			}

			switch opt.Kind {
			case Boolean:
			case Optional, Required, Repeated:
				if !config.UseAssignment {
					config.Args.Next()
					continue
				}
			default:
				return fmt.Errorf("option '%s' exists, but is not named", opt.LongName)

			}
		case ShortArgument:

			if config.IndexedFirst {
				if fui := self.getFirstUnparsedIndexed(); fui != nil {
					return fmt.Errorf("indexed argument '%s' not parsed", fui.LongName)
				}
			}

			if !config.UseAssignment && opt != nil {
				return fmt.Errorf("option '%s' requires a value", opt.LongName)
			}

			// Set up for combined booleans parsing.
			if len(key) > 1 {
				combined = key
				for _, k := range combined {
					if opt = self.FindShort(string(k)); opt == nil {
						return fmt.Errorf("combined argument %s refers to unknown option %s", combined, string(k))
					}
					if opt.Kind != Boolean {
						return fmt.Errorf("combined argument %s may contain boolean options only", combined)
					}
				}
				opt = self.FindShort(combined[:1])
				combined = combined[1:]
				goto ParseOption
			}

			if opt = self.FindShort(key); opt == nil {
				return fmt.Errorf("unknown option '%s'", key)
			}

			switch opt.Kind {
			case Boolean:
			case Optional, Required, Repeated:
				if !config.UseAssignment {
					config.Args.Next()
					continue
				}
			default:
				return fmt.Errorf("option '%s' exists, but is not named", opt.LongName)

			}
		}

		// No options matched so far, see if there's a Variadic.
		if opt == nil {
			for _, v := range self {
				if v.Kind == Variadic {
					opt = v
					break
				}
			}
			if opt == nil {
				break
			}
		}

		// Fail if non *Repeatable option and parsed multiple times.
		if opt.Kind != Repeated {
			if opt.IsParsed {
				return fmt.Errorf("option %s specified multiple times", opt.LongName)
			}
		}

	ParseOption:

		// Set Option as parsed.
		switch opt.Kind {
		case Boolean:
			if !config.UseAssignment {
				opt.IsParsed = true
			} else {
				if assignment {
					return fmt.Errorf("option '%s' cannot be assigned a value", opt.LongName)
				}
				opt.IsParsed = true
			}
		case Optional:
			if !config.UseAssignment {
				opt.Values = append(opt.Values, key)
				opt.IsParsed = true
			} else {
				if !assignment || val == "" {
					return fmt.Errorf("option '%s' requires a value", opt.LongName)
				}
				opt.Values = append(opt.Values, val)
				opt.IsParsed = true
			}
		case Required:
			if !config.UseAssignment {
				opt.Values = append(opt.Values, key)
				opt.IsParsed = true
			} else {
				if !assignment || val == "" {
					return fmt.Errorf("option '%s' requires a value", opt.LongName)
				}
				opt.Values = append(opt.Values, val)
				opt.IsParsed = true
			}
		case Repeated:
			if !config.UseAssignment {
				opt.Values = append(opt.Values, key)
				opt.IsParsed = true
			} else {
				if !assignment || val == "" {
					return fmt.Errorf("option '%s' requires a value", opt.LongName)
				}
				opt.Values = append(opt.Values, val)
				opt.IsParsed = true
			}
		case Indexed:
			if !config.UseAssignment {
				opt.Values = append(opt.Values, key)
				opt.IsParsed = true
			} else {
				opt.Values = append(opt.Values, key)
				opt.IsParsed = true
			}
		case Variadic:
			opt.Values = append(opt.Values, config.Args...)
			opt.IsParsed = true
			config.Args.Clear()
		}

		// Set [Option.Var] value.
		if err = self.setVar(opt); err != nil {
			return fmt.Errorf("invalid Var '%v' for option '%s': %w", opt.Var, opt.LongName, err)
		}

		// Combined booleans loop.
		if len(combined) > 0 {
			opt = self.FindShort(combined[:1])
			combined = combined[1:]
			goto ParseOption
		}

		opt = nil
		config.Args.Next()
	}

	for _, opt = range self {
		if !opt.IsParsed {
			if opt.Kind == Required {
				return fmt.Errorf("required option '%s' not parsed", opt.LongName)
			}
			if opt.Kind == Indexed {
				return fmt.Errorf("indexed option '%s' not parsed", opt.LongName)
			}
		}
	}

	return
}

// getFirstUnparsedIndexed returns the first Indexed Option that is not parsed.
// Returns nil if none found.
func (self Options) getFirstUnparsedIndexed() *Option {
	for _, option := range self {
		if option.Kind == Indexed {
			if !option.IsParsed {
				return option
			}
		}
	}
	return nil
}

// setVar converts option.State.RawValue to option.MappedValue if option's
// raw value is not nil.
//
// Returns nil on success of no value mapped.. Returns a non nil error on
// failed conversion only.
//
// option.MappedValue must be a pointer to a supported type or any type that
// supports conversion from a string by implementing the Value interface.
//
// Supported types are:
// *bool, *string, *float32, *float64,
// *int, *int8, *int16, *1nt32, *int64,
// *uint, *uint8, *uint16, *u1nt32, *uint64
// *time.Duration, *[]string, and any type supporting Value interface.
//
// If an unsupported type was set as option.MappedValue Parse will return a
// conversion error.
func (self Options) setVar(option *Option) (err error) {

	if option.Var == nil {
		return nil
	}

	if option.Kind != Boolean {
		if option.Values.Count() < 1 {
			return nil
		}
	}

	switch option.Kind {
	case Boolean, Optional, Required, Indexed, Variadic:
		return convertToVar(option.Var, option.Values)
	case Repeated:
		return convertToVar(option.Var, option.Values[len(option.Values)-1:])
	default:
		return errors.New("invalid OptionKind")
	}
}

// convertToVar sets v which must be a pointer to a supported type from raw
// or returns an error if conversion error occured.
func convertToVar(v any, raw Values) (err error) {

	// TODO Accept pointer to a pointer (null-like behaviour).

	if tu, ok := v.(encoding.TextUnmarshaler); ok {
		return tu.UnmarshalText([]byte(raw.First()))
	}

	switch p := v.(type) {
	case *bool:
		*p = true
	case *string:
		*p = raw.First()
	case *int:
		var v int64
		if v, err = strconv.ParseInt(raw.First(), 10, 0); err == nil {
			*p = int(v)
		}
	case *uint:
		var v uint64
		if v, err = strconv.ParseUint(raw.First(), 10, 0); err == nil {
			*p = uint(v)
		}
	case *int8:
		var v int64
		if v, err = strconv.ParseInt(raw.First(), 10, 8); err == nil {
			*p = int8(v)
		}
	case *uint8:
		var v uint64
		if v, err = strconv.ParseUint(raw.First(), 10, 8); err == nil {
			*p = uint8(v)
		}
	case *int16:
		var v int64
		if v, err = strconv.ParseInt(raw.First(), 10, 16); err == nil {
			*p = int16(v)
		}
	case *uint16:
		var v uint64
		if v, err = strconv.ParseUint(raw.First(), 10, 16); err == nil {
			*p = uint16(v)
		}
	case *int32:
		var v int64
		if v, err = strconv.ParseInt(raw.First(), 10, 32); err == nil {
			*p = int32(v)
		}
	case *uint32:
		var v uint64
		if v, err = strconv.ParseUint(raw.First(), 10, 32); err == nil {
			*p = uint32(v)
		}
	case *int64:
		*p, err = strconv.ParseInt(raw.First(), 10, 64)
	case *uint64:
		*p, err = strconv.ParseUint(raw.First(), 10, 64)
	case *float32:
		var v float64
		if v, err = strconv.ParseFloat(raw.First(), 64); err == nil {
			*p = float32(v)
		}
	case *float64:
		*p, err = strconv.ParseFloat(raw.First(), 64)
	case *[]string:
		*p = append(*p, raw...)
	case *time.Duration:
		*p, err = time.ParseDuration(raw.First())
	case *time.Time:
		*p, err = time.Parse(time.RFC3339, raw.First())
	default:
		if v, ok := p.(Value); ok {
			err = v.Set(raw)
		} else {
			return errors.New("expected pointer to supported type")
		}
	}
	return
}
