// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package generate

import (
	"fmt"

	"github.com/vedranvuk/cmdline"
)

// This file defines the model exposed to the output file template.

// PairKey is a known pair key read from the a cmdline tag value.
//
// It can appear in source struct doc comments, source struct field docs or 
// source struct field tags.
//
// It can appear multiple times in a doc comment in which case values for same 
// tags are concatenated. Some keys take values in key=value format.
type PairKey = string

const (
	// IncludeKey is a placeholder key that can be used for when a command is
	// to be generated but there is no need to specify any other options for
	// generated code.
	//
	// By default, structs that have no cmdline tags are skipped so a struct can
	// be tagged with this key to be included and use all default options.
	//
	// It takes no value.
	IncludeKey PairKey = "include"

	// IgnoreKey is read from struct fields and specifies that the tagged field
	// should be excluded when generating command options from fields.
	//
	// It takes no values.
	IgnoreKey PairKey = "ignore"

	// HelpKey is used on a source struct and specifies the help text to be set
	// with the command that will represent the struct.
	//
	// It takes a single value in the key=value format that defines the command
	// help. E.g.: help=This is a help text for ca command representing a
	// bound struct.
	//
	// Help text cannnot span multiple lines, it is a one-line shown to user
	// when cmdline config help is requested.
	HelpKey PairKey = "help"

	// NameKey is used on a source struct and specifies the name of the command
	// that represents the struct being bound to.
	//
	// This name is the name by which source struct commands and their field
	// options are addressed from the command line.
	//
	// It takes a single value in the key=value format that defines the command
	// name. E.g.: name=MyStruct.
	NameKey PairKey = "name"

	// ShortNameKey is used on source struct fields and explicitly sets the 
	// option short name.
	ShortNameKey PairKey = "shortName"

	// CommandNameKey specifies the name for the generated command.
	CommandNameKey PairKey = "commandName"

	// TargetNameKey names variable of the output struct command options will
	// write to.
	//
	// This may name a variable declared in some other package file that the
	// generated command options can adress and write from arguments or name
	// the variable that will be generated in the output file so some other file
	// in the package can address it.
	//
	// If unspecified, name is generated from the command name such that the
	// command name is appended with "Var" suffix, e.g. "CommandVar".
	TargetNameKey PairKey = "targetName"

	// HandlerNameKey specifies the name for the command handler.
	//
	// If not specified defaults to name of generated command immediatelly
	// followed with "Handler."
	HandlerNameKey PairKey = "handlerName"

	// GenTargetKey specifies that the variable for the command should be declared.
	//
	// This is useful if the variable is already declared in some
	// other file in the package.
	//
	// Generated commands will still address the target variable defined by
	// [TargetNameKey].
	//
	// It takes no values.
	GenTargetKey PairKey = "genTarget"

	// GenHandlerKey if specified will generate the command handler stub.
	//
	// It takes no values.
	GenHandlerKey PairKey = "genHandler"

	// OptionalKey is read from struct fields and specifies that the tagged
	// field should use the Optional option.
	//
	// It takes no values.
	OptionalKey PairKey = "optional"

	// RequiredKey is read from struct fields and specifies that the tagged
	// field should use the Required option.
	//
	// It takes no values.
	RequiredKey PairKey = "required"
)

// knownPairKeys is a slice of all supported cmdline tag pair keys.
var knownPairKeys = []string{
	IncludeKey, NameKey, ShortNameKey, TargetNameKey, GenTargetKey, CommandNameKey, GenHandlerKey,
	HandlerNameKey, HelpKey, IgnoreKey, OptionalKey, RequiredKey,
}

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

		Imports []string

		// Options to generate.
		Options Options
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
			for _, path := range command.Imports {
				paths[path] =  struct{}{}
			}
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

// AddImport adds an additonal import.
func (self *Command) AddImport(path string) {
	for _, imp := range self.Imports {
		if imp == path {
			return
		}
	}
	self.Imports = append(self.Imports, path)
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
