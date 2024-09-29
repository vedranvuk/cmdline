// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package generate

import (
	"fmt"

	"github.com/vedranvuk/cmdline"
)

type (
	// ImportPath is a package import path.
	ImportPath = string
	// PackageName is a base name of a package.
	PackageName = string
	// ImportMap is a map of package import paths to package base names.
	ImportMap = map[ImportPath]PackageName

	// Model is the top level structure that holds the data from which to
	// generate the output go source file containing generated commands.
	Model struct {
		// ImportMap holds a list of imports to add to the generated file.
		ImportMap
		Commands
	}

	// Commands is a slice of commands to be generated.
	Commands []Command

	// Command defines a cmdline.Command to be generated. It is generated from a
	// source struct.
	Command struct {
		// Name is the command name as it appears in the command line interface.
		//
		// If not specified it defaults to source struct name.
		//
		Name string

		// Help text is the Command help text generated from source struct
		// doc comments.
		Help string

		// CommandName is the name of generated command.
		//
		// If not specified defaults to lowercase name of source struct
		// immediately followed with "Cmd".
		CommandName string

		// TargetName is the name for the generated Command variable specified
		// via tag.
		TargetName string

		// HandlerName is the name of the handler function.
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
		//
		// Currently it will always be the base name of the package path where
		// the source struct is defined.
		SourcePackageName string

		// SourcePackagePath is the path of the package that contains the
		// struct.
		SourcePackagePath string

		// Options to generate.
		Options
	}

	// Options is a slice of *Option.
	Options []*Option

	// Option defines a cmdline.Option to generate in a command. It is generated
	// from a source struct field.
	Option struct {
		// LongName is the long name for the Option.
		// Set from the source field name or custom name from tag.
		LongName string

		// ShortName is the short name for the option.
		// Auto generated.
		ShortName string

		// Help is the option help text.
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
		if b = !command.GenTarget; b {
			break
		}
	}
	return
}

func (self Model) Imports() []string {
	for _, command := range self.Commands {
		for _, option := range command.Options {
			_ = option
		}
	}
	return nil
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
