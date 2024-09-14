package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

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

	// Bast is the bastard ast config.
	Bast *bast.Config `cmdline:"name=bast" json:"bast,omitempty"`

	// bast is the parsed bast, nil until parsed.
	bast *bast.Bast `cmdline:"ignore" json:"-"`

	// commands are te generated commands.
	commands []*cmdline.Command `cmdline:"ignore" json:"-"`

	// structTypeNames are the names of strict types used to generate commands,
	// order matched by commands.
	structTypeNames []string `cmdline:"ignore" json:"-"`

	// imports is a list of imports to add to generated file.
	imports []string `cmdline:"ignore" json:"-"`
}

// DefaultGenerateConfig returns the default [GenerateConfig].
func DefaultGenerateConfig() *GenerateConfig {
	return &GenerateConfig{
		Bast: bast.DefaultConfig(),
	}
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

	if self.bast, err = bast.Load(self.Bast, self.Packages...); err != nil {
		return
	}

	for _, s := range self.bast.AllStructs() {
		if err = self.parseStruct(s); err != nil {
			return
		}
	}

	if err = self.generateOutput(); err != nil {
		return
	}

	return nil
}

func (self *GenerateConfig) parseStruct(s *bast.Struct) (err error) {

	var tag = strutils.Tag{
		Keys:              AllTags,
		TagName:           self.TagName,
		ErrorOnUnknownKey: true,
	}

	if err = tag.ParseDocs(s.Doc); err != nil {
		return
	}

	if tag.Exists(IgnoreTag) {
		return nil
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

	var c = &cmdline.Command{
		Name: name,
		Help: strings.Join(tag.Values[HelpTag], "\n"),
	}

	for _, f := range s.Fields.Values() {
		if err = self.parseField(f, c); err != nil {
			return
		}
	}

	self.commands = append(self.commands, c)
	self.structTypeNames = append(self.structTypeNames, s.Name)

	return nil
}

func (self *GenerateConfig) parseField(f *bast.Field, c *cmdline.Command) (err error) {

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

	var opt cmdline.Option
	var bt = self.bast.ResolveBasicType(f.Type)
	switch bt {
	case "":
		log.Printf("Cannot determine basic type for %s, skipping.\n", f.Type)
	case "bool":
		opt = &cmdline.Boolean{
			LongName:  name,
			ShortName: "",
			Help:      strings.Join(tag.Values[HelpTag], "\n"),
		}
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"string", "[]string":
		if optional {
			opt = &cmdline.Optional{
				LongName:  name,
				ShortName: "",
				Help:      strings.Join(tag.Values[HelpTag], "\n"),
			}
		}
		if required {
			opt = &cmdline.Required{
				LongName:  name,
				ShortName: "",
				Help:      strings.Join(tag.Values[HelpTag], "\n"),
			}
		}
	default:
		log.Printf("Unknown basic type: %s\n", bt)
	}

	c.Options = append(c.Options, opt)

	return nil
}

func (self *GenerateConfig) generateOutput() (err error) {
	// var sb = &strings.Builder{}
	var sb = os.Stdout

	// Header
	fmt.Fprintf(sb, "package %s\n", self.PackageName)
	fmt.Fprintln(sb)
	fmt.Fprintf(sb, "import (\n")
	for _, v := range self.imports {
		fmt.Fprintf(sb, "\t%s\n", v)
	}
	fmt.Fprintf(sb, ")\n")
	fmt.Fprintln(sb)
	// Variables
	fmt.Fprintf(sb, "var (\n")
	for i, c := range self.commands {
		fmt.Fprintf(sb, "\t%s = ", c.Name)
		if self.PointerVars {
			fmt.Fprint(sb, "*")
		}
		fmt.Fprintf(sb, "%s\n", self.structTypeNames[i])
	}
	fmt.Fprintf(sb, ")\n")
	fmt.Fprintln(sb)
	// Commands
	fmt.Fprintf(sb, "var (\n")
	for i, c := range self.commands {
		fmt.Fprintf(sb, "\t%s = ", c.Name)
		if self.PointerVars {
			fmt.Fprint(sb, "*")
		}
		fmt.Fprintf(sb, "%s\n", self.structTypeNames[i])
	}
	return nil
}
