package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/vedranvuk/bast"
	"github.com/vedranvuk/cmdline"
	"github.com/vedranvuk/strutils"
)

const (
	// DefaultTagName is the default name of a struct tag read by cmdline package.
	DefaultTagName = "cmdline"
	// DefaultOutputFile is the filename of the default generate output file.
	DefaultOutputFile = "cmdline.go"
)

// TagKey is a known cmdline key read from the cmdline tag in a struct field
// being bound to.
//
// It can appear in a struct tag or a doc comment of a struct being bound to.
// In a struct doc comment it can be specified multiple times.
// Some keys take values in key=value format.
type TagKey = string

const (
	// NameTag is used on a source struct and specifies the name of the command
	// that represents the struct being bound to.
	//
	// It is also the name of the generated variable of a struct manipulated by
	// the command options.
	//
	// It takes a single value in the key=value format that defines the command
	// name. E.g.: name=MyStruct.
	NameTag TagKey = "name"

	// HelpTag is used on a source struct and specifies the help text to be set
	// with the command that will represent the struct.
	//
	// It takes a single value in the key=value format that defines the command
	// help. E.g.: help=This is a help text for ca command representing a
	// bound struct.
	//
	// Help text cannnot span multiple lines, it is a one-line shown to user
	// when cmdline config help is requested.
	HelpTag TagKey = "help"

	// IgnoreTag is read from struct fields and specifies that the tagged field
	// should be excluded from generated command options.
	//
	// It takes no values.
	IgnoreTag TagKey = "ignore"

	// OptionalTag is read from struct fields and specifies that the tagged
	// field should use the Optional option.
	//
	// It takes no values.
	OptionalTag TagKey = "optional"

	// RequiredTag is read from struct fields and specifies that the tagged
	// field should use the Required option.
	//
	// It takes no values.
	RequiredTag TagKey = "required"
)

// AllTags are all supported tags.
var AllTags = []string{NameTag, HelpTag, IgnoreTag, OptionalTag, RequiredTag}

// cmdline:"name=generate"
// cmdline:"help=Generates commands from structs parsed from packages."
// cmdline:"help=."
type GenerateConfig struct {

	// TagName is the name of the tag read by cmdline from struct tags or
	// struct doc comments. If ommitted the default "cmdline" is read.
	TagName string `cmdline:"name=tagName" json:"tagName,omitempty"`

	// OutputFile is the output file that will contain generated commands.
	// It can be a full or relative path to a go file and if ommited a default
	// value "cmdline.go" is used.
	OutputFile string `cmdline:"name=outputFile,required" json:"outputFile,omitempty"`

	// PackageName is the name of the package generated file belongs to.
	PackageName string `cmdline:"packageName" json:"packageName,omitempty"`

	// Packages is a list of packages to parse. It is a list of relative or full
	// paths to go packages or import paths.
	Packages []string `cmdline:"name=packages" json:"packages,omitempty"`

	// PointerVars if true generates command variables as pointers.
	PointerVars bool `cmdline:"name=pointerVars" json:"pointerVars,omitempty"`

	// BastConfig is the bastard ast config.
	BastConfig *bast.Config `json:"-"`

	// state is the parse state.
	state struct {
		// Bast is the parsed Bast, nil until parsed.
		Bast *bast.Bast
		// maps package import path to package name.
		Imports map[string]string
		// Commands are the parsed Commands.
		Commands []Command
	} `json:"-"`
}

// DefaultGenerateConfig returns the default [GenerateConfig].
func DefaultGenerateConfig() (c *GenerateConfig) {
	c = new(GenerateConfig)
	c.BastConfig = bast.DefaultConfig()
	c.state.Imports = make(map[string]string)
	return
}

type (
	Command struct {
		Name       string   // command name.
		Help       string   // help text
		StructType string   // type name of source struct.
		Options    []Option // options to generate
	}

	Option struct {
		LongName  string   // option long name
		ShortName string   // option short name
		Help      []string // help text
		BasicType string
		Kind      cmdline.OptionKind
	}
)

// Signature returns the option type signature, with "cmdline."" selector
// prefix. Used from template that generates the go commands file.
func (self Option) Signature() string {
	switch self.Kind {
	case cmdline.OptionBoolean:
		return "cmdline.Boolean"
	case cmdline.OptionOptional:
		return "cmdline.Optional"
	case cmdline.OptionRequired:
		return "cmdline.Required"
	}
	return ""
}

// Generate generates the go source code containing cmdline.Command definitions
// that modify
//
// It skips the structs that have no cmdline tags. Structs that are to be used
// as generate source must have the NameTag at minimum.
func (self GenerateConfig) Generate() (err error) {

	if self.TagName == "" {
		self.TagName = DefaultTagName
	}
	if self.OutputFile == "" {
		self.OutputFile = DefaultOutputFile
	}
	if len(self.Packages) == 0 {
		return errors.New("no packages specified")
	}

	if self.state.Bast, err = bast.Load(self.BastConfig, self.Packages...); err != nil {
		return
	}

	self.state.Imports = make(map[string]string)

	for _, s := range self.state.Bast.AllStructs() {

		var tag = strutils.Tag{
			Keys:              AllTags,
			TagName:           self.TagName,
			ErrorOnUnknownKey: true,
		}

		if err = tag.ParseDocs(s.Doc); err != nil {
			return
		}

		var c = Command{
			Name:       s.Name,
			Help:       strings.Join(tag.Values[HelpTag], "\n"),
			StructType: s.Name,
		}

		if tag.Exists(IgnoreTag) {
			continue
		}

		var name = tag.First(NameTag)
		if name == "" {
			err = errors.New("invalid name tag, no value")
			return
		}

		var (
			optional = tag.Exists(OptionalTag)
			required = tag.Exists(RequiredTag)
		)
		if optional && required {
			err = errors.New("optional and required tag keys are mutually exclusive")
			return
		}

		self.state.Imports[s.GetPackage().Path] = s.GetPackage().Name
		if err = self.parseStruct(s, "", &c); err != nil {
			return
		}

		self.state.Commands = append(self.state.Commands, c)
	}

	if err = self.generateOutput(); err != nil {
		return
	}

	return nil
}

// parseStruct parses a struct definition into a command.
func (self *GenerateConfig) parseStruct(s *bast.Struct, path string, c *Command) (err error) {

	for _, f := range s.Fields.Values() {
		if err = self.parseField(f, path, c); err != nil {
			return
		}
	}

	return nil
}

// parseField parses a struct field into a command option.
func (self *GenerateConfig) parseField(f *bast.Field, path string, c *Command) (err error) {

	if path != "" {
		path += "."
	}
	if path += f.Name; f.Name == "" {
		path += f.Type
	}

	var imp = f.PackageImportBySelectorExpr(f.Type)
	if imp != nil {
		var _, name, valid = strings.Cut(f.Type, ".")
		if valid {
			if s := self.state.Bast.PkgStruct(imp.Path, name); s != nil {
				self.state.Imports[imp.Path] = ""
				return self.parseStruct(s, path, c)
			}
		}
	}

	var tag = strutils.Tag{
		Keys:              AllTags,
		TagName:           self.TagName,
		ErrorOnUnknownKey: true,
	}

	if err = tag.ParseStructTag(f.Tag); err != nil {
		return
	}

	var name = tag.First(NameTag)
	if name == "" {
		err = errors.New("invalid name tag, no value")
		return
	}

	var (
		optional = tag.Exists(OptionalTag)
		required = tag.Exists(RequiredTag)
	)
	if optional && required {
		err = errors.New("optional and required tag keys are mutually exclusive")
		return
	}

	var opt Option
	var bt = self.state.Bast.ResolveBasicType(f.Type)
	switch bt {
	case "":
		log.Printf("Cannot determine basic type for %s, skipping.\n", f.Type)
	case "bool":
		opt = Option{
			LongName:  name,
			ShortName: "",
			Help:      tag.Values[HelpTag],
		}
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"string", "[]string":
		if optional {
			opt = Option{
				LongName:  name,
				ShortName: "",
				Help:      tag.Values[HelpTag],
				Kind:      cmdline.OptionOptional,
			}
		}
		if required {
			opt = Option{
				LongName:  name,
				ShortName: "",
				Help:      tag.Values[HelpTag],
				Kind:      cmdline.OptionRequired,
			}
		}
	default:
		log.Printf("Unknown basic type: %s\n", bt)
	}

	c.Options = append(c.Options, opt)

	return nil
}

// generateOutput generates output go file with command definitions.
func (self *GenerateConfig) generateOutput() (err error) {

	var t *template.Template
	if t, err = template.New("cmdline").Parse(fileTemplate); err != nil {
		return
	}

	var file *os.File
	if file, err = os.OpenFile(self.OutputFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm); err != nil {
		return
	}
	defer file.Close()

	return t.Execute(file, self.state)
}

// fileTemplate is a template for the output go file.
const fileTemplate = `
package {{.PackageName}}

import (
	{{template "Imports" .}}
)

var (
	{{template "StructVars" .}}
)

var (
	{{template "Commands" .}}
)

{{define "Imports"}}{{range .Imports}}
	{{.}}
{{end}}
{{end}}

{{define "StructVars"}}{{range .Commands}}
	{{.Name}}Var = {{.StructType}}
{{end}}
{{end}}

{{define "Commands"}}{{range $command := .Commands}}
	{{$command.Name}}Cmd = &cmdline.Command{
		Name: {{.Name}},
		Help: {{.Help}},
		Options: []cmdline.Option{{range $command.Options}}
			{{template "Option" .}}
		{{end}}
		}
	{{end}}
	}
{{end}}

{{define "Option"}}
			&{{.Signature}}{
				LongName: {{.LongName}},
				ShortName: {{.ShortName}},
				Help: {{.Help}},
				MappedValue: &{{.LongName}},
			},
{{end}}
`
