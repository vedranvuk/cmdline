// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import "strings"

// NewArguments is an alias for NewUnixArguments.
func NewArguments(args ...string) Arguments { return NewUnixArguments(args...) }

// NewUnixArguments returns Arguments which recognize a freely interpreted
// Unix/GNU style argument prefixes (with some extensions) as defined:
//
// Two dash "--verbose" for LongArgument parameters (dash-separated string).
// One dash "-v" for ShortArgument parameters (one character after prefix).
// One dash "-Syuv" for CombinedArgument parameters (word after prefix).
// Non prefixed arguments are recognized as TextArgument.
// Spaces in all types of arguments except TextArgument are an InvalidArgument.
// No arguments in Arguments returns NoArgument.
// Args initialize Arguments to given slice of strings.
func NewUnixArguments(args ...string) Arguments {
	var p = &unixarguments{
		basearguments{
			args,
			InvalidArgument,
			"",
			"",
			nil,
		},
	}
	p.parse = p.ParseArgument
	p.parse()
	return p
}

// NewDOSArguments returns Arguments which recognize DOS style argument
// prefixes as defined:
// One frontslash "/" for both LongArgument and ShortArgument.
// CombinedArgument is not recognized.
// Spaces in all types of arguments except TextArgument are an InvalidArgument.
// No arguments in Arguments returns NoArgument.
// Args initialize Arguments to given slice of strings.
func NewDOSArguments(args ...string) Arguments {
	var p = &dosarguments{
		basearguments{
			args,
			InvalidArgument,
			"",
			"",
			nil,
		},
	}
	p.parse = p.ParseArgument
	p.parse()
	return p
}

// Arguments manages a slice of arguments.
type Arguments interface {
	// Length returns current number of elements in arguments.
	Length() int
	// Raw returns current argument in its raw, unparsed form.
	// If Arguments are empty returns an empty string.
	Raw() string
	// Kind returns current argument kind.
	// If there are no elements in Arguments result will be NoArgument.
	Kind() Argument
	// Name returns name part of the argument.
	// Result will be current argument stripped of any prefixes
	// and value assignments, depending on argument kind.
	Name() string
	// Value returns the value part of argument.
	// Result will be assignment value of a AssignmentArgument.
	// If current Argument is not an AssignmentArgument, result
	// will be an empty string.
	Value() string
	// Advance discards current argument and sets cursor to nest one.
	// If Arguments advanced to an argument result is true.
	// If there were no arguments left returns false.
	Advance() bool
	// Slice returns Arguments as a string slice.
	Slice() []string
}

// Argument defines a kind of argument as recognized by types in this package.
// Argument is determined from a string depending on Arguments implementation
type Argument int

const (
	// NoArgument represents no argument or a no argument state.
	NoArgument Argument = iota
	// InvalidArgument represents an Argument that has invalid formatting.
	InvalidArgument
	// TextArgument represents a single, possibly space separated text argument.
	TextArgument
	// ShortArgument represents a single char argument directly preffixed
	// with "-".
	ShortArgument
	// LongArgument represents a single word argument directly preffixed 
	// with "--".
	LongArgument
	// CombinedArgument represents a single word argument directly preffixed
	// with "-".
	CombinedArgument
	// AssignmentArgument represents an assignment argument.
	AssignmentArgument
)

// String implements stringer on Argument.
func (arg Argument) String() (s string) {
	switch arg {
	case InvalidArgument:
		s = "Invalid Argument"
	case NoArgument:
		s = "No Argument"
	case TextArgument:
		s = "Text Argument"
	case ShortArgument:
		s = "Short Argument"
	case LongArgument:
		s = "Long Argument"
	case CombinedArgument:
		s = "Combined Argument"
	case AssignmentArgument:
		s = "Assignment Argument"
	default:
		s = "!Undefined! Argument"
	}
	return
}

// basearguments implements base Arguments functionality.
type basearguments struct {
	args  []string
	kind  Argument
	name  string
	value string
	parse func() // implemented by wrappers.
}

// Raw implements Arguments.Raw.
func (ba *basearguments) Raw() string {
	if ba.Length() > 0 {
		return ba.args[0]
	}
	return ""
}

// Name implements Arguments.Name.
func (ba *basearguments) Name() string { return ba.name }

// Value implements Arguments.Value.
func (ba *basearguments) Value() string { return ba.value }

// Length implements Arguments.Length.
func (ba *basearguments) Length() int { return len(ba.args) }

// Current implements Arguments.Current.
func (ba *basearguments) Kind() Argument { return ba.kind }

// Advance implements Arguments.Advance.
func (ba *basearguments) Advance() bool {
	if ba.Length() == 0 {
		return false
	}
	ba.args = ba.args[1:]
	if ba.parse != nil {
		ba.parse()
	}
	return true
}

// Slice implements Arguments.Slice.
func (ba *basearguments) Slice() []string { return ba.args }

// unixarguments implements Arguments as defined by NewUnixArguments.
type unixarguments struct{ basearguments }

// ParseArgument help.
func (ua *unixarguments) ParseArgument() {
	ua.kind = NoArgument
	ua.name = ""
	ua.value = ""
	if ua.Length() == 0 {
		return
	}
	var arglen = len(ua.args[0])
	if arglen == 0 {
		ua.kind = InvalidArgument
		return
	}
	var i int
	for i = 0; i < arglen; i++ {
		if ua.args[0][i] != '-' {
			break
		}
		if ua.kind == NoArgument {
			ua.kind = ShortArgument
			continue
		}
		if ua.kind == ShortArgument {
			ua.kind = LongArgument
			continue
		}
		if ua.kind == LongArgument {
			ua.kind = InvalidArgument
			return
		}
	}
	if i >= arglen {
		ua.kind = InvalidArgument
		return
	}
	ua.name = ua.args[0][i:]
	arglen = len(ua.name)
	if ua.kind == NoArgument {
		ua.kind = TextArgument
	}
	if ua.kind == ShortArgument && arglen > 1 {
		ua.kind = CombinedArgument
	}
	if ua.kind != TextArgument {
		for i = 0; i < arglen; i++ {
			if ua.name[i] == ' ' {
				ua.kind = InvalidArgument
				return
			}
		}
	}
	var kv = strings.SplitN(ua.name, "=", 2)
	if len(kv) > 1 {
		if len(kv[0]) == 0 {
			ua.kind = InvalidArgument
			ua.name = ""
			return
		}
		ua.kind = AssignmentArgument
		ua.name = kv[0]
		ua.value = kv[1]
		return
	}
	return
}

// doasarguments implements Arguments as defined by NewDOSArguments.
type dosarguments struct{ basearguments }

func (args *dosarguments) ParseArgument() {
	// TODO Implement dosarguments.ParseArgument()
	return
}
