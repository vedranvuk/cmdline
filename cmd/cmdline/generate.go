package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/format"
	"io/fs"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/vedranvuk/bast"
	"github.com/vedranvuk/cmdline"
	"github.com/vedranvuk/strutils"
)

//go:embed generate.tmpl
var resources embed.FS

// FS returns the embedded resources as a file system.
func FS() embed.FS { return resources }

const (
	// DefaultTagKey is the default key of a tag value parsed by cmdline.
	DefaultTagKey = "cmdline"

	// DefaultOutputFile is the default base name of an output go file that
	// will contain generated code.
	DefaultOutputFile = "cmdline.go"
)

// GenerateConfig is the configuration for a command that generates
// [cmdline.Commands] from structs.
//
// For each source struct a command is generated with options that reflect
// fields in the source struct.
//
// cmdline:"name=generate"
// cmdline:"help=Generates commands from structs parsed from packages."
type GenerateConfig struct {

	// TagName is the name of the tag read by cmdline from struct tags or
	// doc comments.
	//
	// Default: DefaultNameTag.
	TagName string `cmdline:"name=tag-name" json:"tagName,omitempty"`

	// OutputFile is the output file that will contain generated commands.
	// It can be a full or relative path to a go file and if ommited a default
	// value "cmdline.go" is used.
	OutputFile string `cmdline:"name=output-file,required" json:"outputFile,omitempty"`

	// PackageName is the name of the package generated file belongs to.
	PackageName string `cmdline:"package-name" json:"packageName,omitempty"`

	// Packages is a list of packages to parse. It is a list of relative or full
	// paths to go packages or import paths.
	Packages []string `cmdline:"name=packages" json:"packages,omitempty"`

	// HelpFromTag if true Adds option help from HelpTag.
	HelpFromTag bool `cmdline:"name=help-from-tag" json:"helpFromTag,omitempty"`

	// HelpFromDocs if true adds option help from srtuct field docs.
	HelpFromDocs bool `cmdline:"name=help-from-docs" json:"helpFromDocs,omitempty"`

	// Print prints the output to stdout.
	//
	// Default: true
	Print bool `cmdline:"name=print-to-stdout" json:"print"`

	// NoWrite if true disables writing to output file.
	//
	// Default: false
	NoWrite bool `cmdline:"name=no-write"`

	// BastConfig is the bastard ast config.
	BastConfig *bast.Config `json:"-,"`

	// Model is the parsed model.
	Model `json:"-"`
}

// DefaultGenerateConfig returns the default [GenerateConfig].
func DefaultGenerateConfig() (c *GenerateConfig) {
	c = new(GenerateConfig)
	c.TagName = DefaultTagKey
	c.OutputFile = DefaultOutputFile
	c.HelpFromTag = true
	c.HelpFromDocs = true
	c.Print = true
	c.NoWrite = false
	c.BastConfig = bast.DefaultConfig()
	c.BastConfig.HaltOnTypeCheckErrors = false
	c.Model.ImportMap = make(ImportMap)
	return
}

// TagKey is a known key read from the a cmdline tag value.
//
// It can appear in a struct tag or a doc comment of a struct being bound to.
// In a struct doc comment it can be specified multiple times.
// Some keys take values in key=value format.
type TagKey = string

const (
	// NameKey is used on a source struct and specifies the name of the command
	// that represents the struct being bound to.
	//
	// It is also the name of the generated variable of a struct manipulated by
	// the command options.
	//
	// It takes a single value in the key=value format that defines the command
	// name. E.g.: name=MyStruct.
	NameKey TagKey = "name"

	// HelpKey is used on a source struct and specifies the help text to be set
	// with the command that will represent the struct.
	//
	// It takes a single value in the key=value format that defines the command
	// help. E.g.: help=This is a help text for ca command representing a
	// bound struct.
	//
	// Help text cannnot span multiple lines, it is a one-line shown to user
	// when cmdline config help is requested.
	HelpKey TagKey = "help"

	// IgnoreKey is read from struct fields and specifies that the tagged field
	// should be excluded from generated command options.
	//
	// It takes no values.
	IgnoreKey TagKey = "ignore"

	// OptionalKey is read from struct fields and specifies that the tagged
	// field should use the Optional option.
	//
	// It takes no values.
	OptionalKey TagKey = "optional"

	// RequiredKey is read from struct fields and specifies that the tagged
	// field should use the Required option.
	//
	// It takes no values.
	RequiredKey TagKey = "required"
)

// AllTags is a slice of  all supported cmdline tags.
var AllTags = []string{NameKey, HelpKey, IgnoreKey, OptionalKey, RequiredKey}

// Generate model.

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
		Bast *bast.Bast
		ImportMap
		Commands
	}

	// Commands is a slice of commands to be generated.
	Commands []Command

	// Command defines a cmdline.Command to be generated. It is generated from a
	// source struct.
	Command struct {
		// Name is the command name. It is generated from the source struct type
		// name or specified via cmdline tag in struct doc comments.
		Name string

		// Help text is the COmmand help text generated from source struct
		// doc comments.
		Help string // help text

		// SourceStructType is the name of the struct type from which the
		// Command is generated..
		SourceStructType string

		// SourceStructPackageName is the base name of the package in which
		// Source struct is defined. Used as selector prefix in generated
		// source.
		SourceStructPackageName string

		// Options to generate.
		Options []Option
	}

	// Option defines a cmdline.Option to generate in a command. It is generated
	// from a source struct field.
	Option struct {
		// LongName is the long name for the Option.
		LongName string
		// ShortName is the short name for the option.
		ShortName string
		// Help is the option help text.
		Help []string
		// BasicType is the determined basic type of the field for which Option
		// is generated.
		BasicType string
		// Kind is the [cmdline.Option] kind to generate.
		Kind cmdline.OptionKind
	}
)

func (self Command) Signature() string {
	return self.SourceStructPackageName + "." + self.SourceStructType
}

// Signature returns the option type signature, with "cmdline." selector
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
		self.TagName = DefaultTagKey
	}
	if self.OutputFile == "" {
		self.OutputFile = DefaultOutputFile
	}
	if len(self.Packages) == 0 {
		return errors.New("no packages specified")
	}

	if self.Model.Bast, err = bast.Load(self.BastConfig, self.Packages...); err != nil {
		return
	}

	if self.Model.ImportMap == nil {
		self.Model.ImportMap = make(ImportMap)
	}
	self.Model.ImportMap["github.com/vedranvuk/cmdline"] = ""

	for _, s := range self.Model.Bast.AllStructs() {

		var tag = strutils.Tag{
			KnownPairKeys:     AllTags,
			TagKey:            self.TagName,
			ErrorOnUnknownKey: true,
		}

		for _, line := range self.uncommentDocs(s.Doc) {
			if err = tag.Parse(line); err != nil {
				if err != strutils.ErrTagNotFound {
					return
				}
			}
		}

		var c = Command{
			Name:                    s.Name,
			Help:                    strings.Join(tag.Values[HelpKey], "\n"),
			SourceStructType:        s.Name,
			SourceStructPackageName: s.GetPackage().Name,
		}

		if tag.Exists(IgnoreKey) {
			continue
		}

		var name = s.Name
		if tag.Exists(NameKey) {
			if name = tag.First(NameKey); name == "" {
				err = errors.New("invalid name tag, no value")
			}
		}

		var (
			optional = tag.Exists(OptionalKey)
			required = tag.Exists(RequiredKey)
		)
		if optional && required {
			err = errors.New("optional and required tag keys are mutually exclusive")
			return
		}

		self.Model.ImportMap[s.GetPackage().Path] = s.GetPackage().Name
		if err = self.parseStruct(s, "", &c); err != nil {
			return
		}

		self.Model.Commands = append(self.Model.Commands, c)
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

	var imp = f.ImportSpecBySelectorExpr(f.Type)
	if imp != nil {
		var _, name, valid = strings.Cut(f.Type, ".")
		if valid {
			if s := self.Model.Bast.PkgStruct(imp.Path, name); s != nil {
				self.Model.ImportMap[imp.Path] = ""
				return self.parseStruct(s, path, c)
			}
		}
	}

	var tag = strutils.Tag{
		KnownPairKeys:     AllTags,
		TagKey:            self.TagName,
		ErrorOnUnknownKey: true,
	}

	if err = tag.Parse(f.Tag); err != nil {
		if err != strutils.ErrTagNotFound {
			return
		}
	}
	for _, line := range self.uncommentDocs(f.Doc) {
		if err = tag.Parse(line); err != nil {
			if err != strutils.ErrTagNotFound {
				return
			}
		}
	}

	var name = f.Name
	if tag.Exists(NameKey) {
		if name = tag.First(NameKey); name == "" {
			return errors.New("invalid name tag, no value")
		}
	}

	var (
		optional = tag.Exists(OptionalKey)
		required = tag.Exists(RequiredKey)
	)
	if optional && required {
		err = errors.New("optional and required tag keys are mutually exclusive")
		return
	}
	if !(optional && required) {
		optional = true
	}

	var opt Option
	switch opt.BasicType = self.Model.Bast.ResolveBasicType(f.Type); opt.BasicType {
	case "":
		log.Printf("Cannot determine basic type for field %s, skipping.\n", f.Type)
	case "bool":
		opt = Option{
			LongName:  name,
			ShortName: "",
			Help:      self.makeHelp(tag.Values[HelpKey], f.Doc),
			Kind:      cmdline.OptionBoolean,
		}
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"string", "[]string":
		if optional {
			opt = Option{
				LongName:  name,
				ShortName: "",
				Help:      self.makeHelp(tag.Values[HelpKey], f.Doc),
				Kind:      cmdline.OptionOptional,
			}
		}
		if required {
			opt = Option{
				LongName:  name,
				ShortName: "",
				Help:      self.makeHelp(tag.Values[HelpKey], f.Doc),
				Kind:      cmdline.OptionRequired,
			}
		}
	default:
		log.Printf("Unknown basic type: %s\n", opt.BasicType)
	}

	c.Options = append(c.Options, opt)

	return nil
}

// generateOutput generates output go file with command definitions.
func (self *GenerateConfig) generateOutput() (err error) {

	var buf []byte
	if buf, err = fs.ReadFile(FS(), "generate.tmpl"); err != nil {
		return
	}

	var (
		t *template.Template
		m = parse.ParseComments | parse.SkipFuncCheck
	)
	if t, err = parseTemplateWithMode("cmdline", string(buf), m); err != nil {
		return fmt.Errorf("parse output template: %w", err)
	}

	var bb = bytes.NewBuffer(nil)
	if err = t.Execute(bb, self); err != nil {
		return fmt.Errorf("execute output template: %w", err)
	}
	fmt.Print(bb.String())

	var source []byte
	if source, err = format.Source(bb.Bytes()); err != nil {
		return fmt.Errorf("format output: %w", err)
	}

	if self.Print {
		if _, err = fmt.Print(string(source)); err != nil {
			return fmt.Errorf("print to stdout: %w", err)
		}
	}

	if !self.NoWrite {
		var file *os.File
		if file, err = os.OpenFile(
			self.OutputFile,
			os.O_CREATE|os.O_TRUNC|os.O_RDWR,
			os.ModePerm,
		); err != nil {
			return
		}
		defer file.Close()
		if _, err = file.Write(source); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
	}

	return nil
}

// parseTemplateWithMode parses a template with parse mode.
// Downside is that built in funcmap is not added.
func parseTemplateWithMode(name, text string, mode parse.Mode) (*template.Template, error) {
	var (
		t       = parse.New(name)
		treeSet = make(map[string]*parse.Tree)
	)
	t.Mode = mode
	var tree, err = t.Parse(text, "{{", "}}", treeSet)
	if err != nil {
		return nil, err
	}
	tmpl := template.New(name)
	return tmpl.AddParseTree(name, tree)
}

// uncommentDocs removes double-slash comment and trims leading and trailing
// space prefix from each in in and returns it.
func (self *GenerateConfig) uncommentDocs(in []string) (out []string) {
	out = make([]string, 0, len(in))
	for _, line := range in {
		out = append(out, strings.TrimSpace(strings.TrimPrefix(line, "//")))
	}
	return
}

// helpFromDoc generates help from tag and doc comment.
func (self *GenerateConfig) makeHelp(tag, doc []string) (out []string) {
	const col = 80
	var lt, ld, l = len(tag), len(doc), 0
	if !self.HelpFromTag {
		lt = 0
	}
	if !self.HelpFromDocs {
		ld = 0
	}
	if lt > 0 && ld > 0 {
		l = 1
	}
	l += lt + ld
	out = make([]string, 0, l)
	for _, line := range tag {
		out = append(out, line)
	}
	if lt > 0 && ld > 0 {
		out = append(out, "")
	}
	for _, line := range doc {
		line = strings.TrimSpace(strings.TrimPrefix(line, "//"))
		if strings.HasPrefix(line, "go:") {
			continue
		}
		if strings.HasPrefix(line, self.TagName+":") {
			continue
		}
		out = append(out, line)
	}
	return strutils.WrapText(strings.Join(out, " "), col, false)
}

// LazyStructCopy copies values from src fields that have a coresponding field
// in dst to that field in dst. Fields must have same name and type. Tags are
// ignored. src and dest must be of struct type and addressable.
func LazyStructCopy(src, dst interface{}) error {

	var (
		dstErr = errors.New("destination must be a pointer to a struct")
		srcv   = reflect.Indirect(reflect.ValueOf(src))
		dstv   = reflect.ValueOf(dst)
	)

	if srcv.Kind() != reflect.Struct {
		return errors.New("source must be a struct")
	}

	if dstv.Kind() != reflect.Pointer {
		return dstErr
	}
	dstv = reflect.Indirect(dstv)
	if dstv.Kind() != reflect.Struct {
		return dstErr
	}

	for i := 0; i < srcv.NumField(); i++ {
		var (
			name = srcv.Type().Field(i).Name
			tgt  = dstv.FieldByName(name)
		)
		if !tgt.IsValid() {
			continue
		}
		if tgt.Kind() != srcv.Field(i).Kind() {
			continue
		}
		if name == "_" {
			continue
		}
		if name[0] >= 97 && name[0] <= 122 {
			continue
		}
		tgt.Set(srcv.Field(i))
	}

	return nil
}
