// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package generate

import (
	"fmt"

	"github.com/vedranvuk/cmdline"
)

// This file defines the model exposed to the output file template.

type (
	// Model is the top level structure that holds the data from which to
	// generate the output go source file containing generated commands.
	Model struct {
		// Commands is a slice of commands to be generated.
		Commands
	}

	// Commands is a slice of commands to be generated.
	Commands []Command

	// Command defines a [cmdline.Command] to be generated. It is generated 
	// from struct that is known as the "source" in the model.
	Command struct {
		// Name is the command name as it appears in the command line interface.
		//
		// If not specified via tag defaults to source struct name.
		Name string

		// Help text is the [cmdline.Command] help text.
		//
		// If [Config.HelpFromTag] is true value of [HelpKey] is added as help.
		// If [Config.HelpFromDocs] is true source struct doc comment is added
		// as help.
		Help string

		// CommandName is the name of generated command.
		//
		// If not specified defaults to camelcased name of source struct
		// immediately followed by "Cmd".
		CommandName string

		// TargetName is the name for the generated Command variable specified
		// via tag.
		//
		// If not specified defaults to camelcased name of source immediately 
		// followed by "Var".
		TargetName string

		// HandlerName is the name of the handler function.
		//
		// If not specified a handler name is generated from keyword "handle"
		// immediately followed by Command name.
		HandlerName string

		// GenTarget is specified via struct tag and specifies that the
		// variable for the command should not be declared.
		GenTarget bool

		// GenHandler if true generates a handler stub for the generated
		// command.
		GenHandler bool

		// SourceType is the name of the struct type from which the
		// Command is generated.
		SourceType string

		// SourcePackageName is the base name of the package in which
		// Source struct is defined.
		SourcePackageName string

		// SourcePackagePath is the path of the package that contains the
		// struct.
		SourcePackagePath string

		// Options to generate.
		Options
	}

	// Options is a slice of *Option.
	Options []*Option

	// Option defines a [cmdline.Option] in the generated [cmdline.Command]. 
	// It is generated from a source struct field.
	Option struct {
		// LongName is the long name for the Option.
		// Set from the source field name or custom name from tag.
		LongName string

		// ShortName is the short name for the option.
		// Auto generated.
		ShortName string

		// Help text is the [cmdline.Option] help text.
		//
		// If [Config.HelpFromTag] is true value of [HelpKey] is added as help.
		// If [Config.HelpFromDocs] is true source struct doc comment is added
		// as help.
		Help string

		// SourceFieldName is the source field name.
		SourceFieldName string

		// SourceFieldPath is the path through the nested structs to the struct field
		// in the source struct.
		SourceFieldPath string

		// SourceBasicType is the determined basic type of the field for which Option
		// is generated.
		SourceBasicType string

		// Kind is the [cmdline.Option] kind to generate.
		Kind cmdline.Kind
	}
)

// AnyTargets returns true if there are any structs for receiving command
// option values to declare.
func (self Model) AnyTargets() (b bool) {
	for _, command := range self.Commands {
		if b = command.GenTarget; b {
			break
		}
	}
	return
}

// Imports returns a slice of imports to include in the generated output file.
func (self Model) Imports() (out []string) {
	var paths = make(map[string]struct{})
	for _, command := range self.Commands {
		if command.GenTarget {
			paths[command.SourcePackagePath] = struct{}{}
		}
	}
	out = make([]string, 0, len(paths) + 1)
	out = append(out, "\"github.com/vedranvuk/cmdline\"")
	for key := range paths {
		out = append(out, fmt.Sprintf("\"%s\"", key))
	}
	return
}

// TargetSelector returns a selector expression string that adresses the
// source struct in the package where it is defined.
// E.g.: "models.Struct"
func (self Command) TargetSelector() string {
	return self.SourcePackageName + "." + self.SourceType
}

// Count returns number of [Option] in [Options].
func (self Options) Count() int { return len(self) }

func (self Options) IsLast(index int) bool { return index == len(self)-1 }

// Signature returns option cmdline registration function signature.
func (self Option) Signature() string { return self.Kind.String() }

// Declaration returns the option declaration in format
func (self Option) Declaration(cmd *Command) string {
	switch self.Kind {
	case cmdline.Boolean:
		return fmt.Sprintf("BooleanVar(\"%s\", \"%s\", \"%s\", &%s)", self.LongName, self.ShortName, self.Help, cmd.TargetName+"."+self.SourceFieldPath)
	case cmdline.Optional:
		return fmt.Sprintf("OptionalVar(\"%s\", \"%s\", \"%s\", &%s)", self.LongName, self.ShortName, self.Help, cmd.TargetName+"."+self.SourceFieldPath)
	case cmdline.Required:
		return fmt.Sprintf("RequiredVar(\"%s\", \"%s\", \"%s\", &%s)", self.LongName, self.ShortName, self.Help, cmd.TargetName+"."+self.SourceFieldPath)
	case cmdline.Repeated:
		return fmt.Sprintf("RepeatedVar(\"%s\", \"%s\", &%s)", self.LongName, self.Help, cmd.TargetName+"."+self.SourceFieldPath)
	case cmdline.Variadic:
		return fmt.Sprintf("VariadicVar(\"%s\", \"%s\", &%s)", self.LongName, self.Help, cmd.TargetName+"."+self.SourceFieldPath)
	default:
		return ""
	}
}
