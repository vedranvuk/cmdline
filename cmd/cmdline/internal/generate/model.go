// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package generate

import (
	"fmt"
	"strings"

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

		// CmdName is the name of generated command.
		//
		// If not specified defaults to lowercase name of source struct
		// immediately followed with "Cmd".
		CmdName string

		// VarName is the name for the generated Command variable specified
		// via tag.
		VarName string

		// NoDeclareVar is specified via struct tag and specifies that the
		// variable for the command should not be declared.
		NoDeclareVar bool

		// HandlerName is the name of the handler function.
		// If not specified a handler name is generated from keyword "handle"
		// immediately followed by Command name.
		HandlerName string

		// GenerateHandler if true generates a handler stub for the generated
		// command.
		GenerateHandler bool

		// Help text is the Command help text generated from source struct
		// doc comments.
		Help string

		// SourceStructType is the name of the struct type from which the
		// Command is generated.
		SourceStructType string

		// SourceStructPackageName is the base name of the package in which
		// Source struct is defined.
		//
		// Currently it will always be the base name of the package path where
		// the source struct is defined.
		SourceStructPackageName string

		// Options to generate.
		Options []*Option
	}

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

		// FieldName is the source field name.
		FieldName string

		// FieldPath is the path through the nested structs to the struct field
		// in the source struct.
		FieldPath string

		// BasicType is the determined basic type of the field for which Option
		// is generated.
		BasicType string

		// Kind is the [cmdline.Option] kind to generate.
		Kind cmdline.Kind
	}
)

// AnyStructVars returns true if there are any structs for receiving command
// option values to declare.
func (self Model) AnyStructVars() (b bool) {
	for _, command := range self.Commands {
		if b = !command.NoDeclareVar; b {
			break
		}
	}
	return
}

// GetCommandName returns the command name.
func (self Command) GetCommandName() string {
	if self.CmdName != "" {
		return self.CmdName
	}
	return strings.ToLower(self.SourceStructType) + "Cmd"
}

// StructSelector returns a selector expression string that adresses the
// source struct in the package where it is defined.
// E.g.: "models.Struct"
func (self Command) StructSelector() string {
	return self.SourceStructPackageName + "." + self.SourceStructType
}

// GetStructVarName returns the variable name.
//
// It returns [Command.VarName] if not empty, otherwise concats [Command.Name] and
// "Var".
func (self Command) GetStructVarName() string {
	if self.VarName != "" {
		return self.VarName
	}
	return self.Name + "Var"
}

// GetHandlerName returns the handler name.
// If [Command.HandlerName] is not empty returns it oterwise the command name
// immediately followed by "Handler".
func (self Command) GetHandlerName() string {
	if self.HandlerName != "" {
		return self.HandlerName
	}
	return self.GetCommandName() + "Handler"
}

// Signature returns option cmdline registration function signature.
func (self Option) Signature() string { return self.Kind.String() }

// Declaration returns the option declaration in format
func (self Option) Declaration(cmd *Command) string {
	switch self.Kind {
	case cmdline.Boolean:
		return fmt.Sprintf("BooleanVar(%s, %s, %s, %s)", self.LongName, self.ShortName, self.Help, "")
	default:
		return ""
	}
}