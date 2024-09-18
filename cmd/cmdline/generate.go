package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io/fs"
	"log"
	"reflect"
	"strings"
	"text/template"
	"text/template/parse"

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

	// HelpFromTag if true Adds option help from HelpTag.
	HelpFromTag bool `cmdline:"name=helpFromTag" json:"helpFromTag,omitempty"`

	// HelpFromDocs if true adds option help from srtuct field docs.
	HelpFromDocs bool `cmdline:"name=helpFromDocs" json:"helpFromDocs,omitempty"`

	// BastConfig is the bastard ast config.
	BastConfig *bast.Config `json:"-"`

	// Model is the parsed model.
	Model `json:"-"`
}

// DefaultGenerateConfig returns the default [GenerateConfig].
func DefaultGenerateConfig() (c *GenerateConfig) {
	c = new(GenerateConfig)
	c.TagName = DefaultTagName
	c.OutputFile = DefaultOutputFile
	c.HelpFromTag = true
	c.HelpFromDocs = true
	c.BastConfig = bast.DefaultConfig()
	c.Model.ImportMap = make(ImportMap)
	return
}

// Generate model.

type (
	ImportPath  = string
	PackageName = string
	ImportMap   = map[ImportPath]PackageName

	Model struct {
		Bast *bast.Bast
		ImportMap
		Commands
	}

	Commands []Command

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

	if self.Model.Bast, err = bast.Load(self.BastConfig, self.Packages...); err != nil {
		return
	}

	if self.Model.ImportMap == nil {
		self.Model.ImportMap = make(ImportMap)
	}
	self.Model.ImportMap["github.com/vedranvuk/cmdline"] = ""

	for _, s := range self.Model.Bast.AllStructs() {

		var tag = strutils.Tag{
			Keys:              AllTags,
			TagName:           self.TagName,
			ErrorOnUnknownKey: true,
		}

		if err = tag.ParseDocs(s.Doc); err != nil {
			if err != strutils.ErrTagNotFound {
				return
			}
		}

		var c = Command{
			Name:       s.Name,
			Help:       strings.Join(tag.Values[HelpTag], "\n"),
			StructType: s.Name,
		}

		if tag.Exists(IgnoreTag) {
			continue
		}

		var name = s.Name
		if tag.Exists(NameTag) {
			if name = tag.First(NameTag); name == "" {
				err = errors.New("invalid name tag, no value")
			}
		}

		var (
			optional = tag.Exists(OptionalTag)
			required = tag.Exists(RequiredTag)
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

	var imp = f.PackageImportBySelectorExpr(f.Type)
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
		Keys:              AllTags,
		TagName:           self.TagName,
		ErrorOnUnknownKey: true,
	}

	if err = tag.ParseStructTag(f.Tag); err != nil {
		if err != strutils.ErrTagNotFound {
			return
		}
	}
	if err = tag.ParseDocs(f.Doc); err != nil {
		if err != strutils.ErrTagNotFound {
			return
		}
	}

	var name = f.Name
	if tag.Exists(NameTag) {
		if name = tag.First(NameTag); name == "" {
			return errors.New("invalid name tag, no value")
		}
	}

	var (
		optional = tag.Exists(OptionalTag)
		required = tag.Exists(RequiredTag)
	)
	if optional && required {
		err = errors.New("optional and required tag keys are mutually exclusive")
		return
	}
	if !(optional && required) {
		optional = true
	}

	var opt Option
	var bt = self.Model.Bast.ResolveBasicType(f.Type)
	switch bt {
	case "":
		log.Printf("Cannot determine basic type for %s, skipping.\n", f.Type)
	case "bool":
		opt = Option{
			LongName:  name,
			ShortName: "",
			Help:      self.makeHelp(tag.Values[HelpTag], f.Doc),
		}
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"string", "[]string":
		if optional {
			opt = Option{
				LongName:  name,
				ShortName: "",
				Help:      self.makeHelp(tag.Values[HelpTag], f.Doc),
				Kind:      cmdline.OptionOptional,
			}
		}
		if required {
			opt = Option{
				LongName:  name,
				ShortName: "",
				Help:      self.makeHelp(tag.Values[HelpTag], f.Doc),
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

	var buf []byte
	if buf, err = fs.ReadFile(FS(), "commands.tmpl"); err != nil {
		return
	}

	var (
		t *template.Template
		m = parse.ParseComments | parse.SkipFuncCheck
	)
	if t, err = parseTemplateWithMode("cmdline", string(buf), m); err != nil {
		return
	}

	var bb = bytes.NewBuffer(nil)
	if err = t.Execute(bb, self); err != nil {
		return
	}

	var source []byte
	if source, err = format.Source(bb.Bytes()); err != nil {
		return
	}
	
	fmt.Print(string(source))
	return nil

	// var file = os.Stdout
	// var file *os.File
	// if file, err = os.OpenFile(self.OutputFile, os.O_CREATE|os.O_TRUNC|os.O_RDWR, os.ModePerm); err != nil {
	// 	return
	// }
	// defer file.Close()

	// return t.Execute(file, self)
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

// helpFromDoc generates help from doc comment.
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
	out = make([]string, 0, lt+ld)
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
