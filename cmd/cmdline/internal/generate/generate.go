// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package generate

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/format"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/vedranvuk/bast"
	"github.com/vedranvuk/cmdline"
	"github.com/vedranvuk/strutils"
)

//go:embed generate.declarative.tmpl
var resources embed.FS

// FS returns the embedded resources as a file system.
func FS() embed.FS { return resources }

const (
	// DefaultTagKey is the default key of a tag value parsed by cmdline.
	DefaultTagKey = "cmdline"

	// DefaultOutputFile is the default base name of an output go file that
	// will contain generated code.
	DefaultOutputFile = "cmdline.go"

	// DefaultConfigFileName is the default cmdline config file name.
	DefaultConfigFileName = "cmdline.json"
)

// PairKey is a known pair key read from the a cmdline tag value.
//
// It can appear in a struct tag or a doc comment of a struct being bound to.
// In a struct doc comment it can be specified multiple times.
// Some keys take values in key=value format.
type PairKey = string

const (
	// IncludeKey is a placeholder key for when a command line command is to be
	// generated for the struct but there is no need to specify any other
	// properties for the generated code.
	//
	// By default, structs that have no cmdline tags are skipped.
	IncludeKey PairKey = "include"

	// NameKey is used on a source struct and specifies the name of the command
	// that represents the struct being bound to.
	//
	// It is also the name of the generated variable of a struct manipulated by
	// the command options.
	//
	// It takes a single value in the key=value format that defines the command
	// name. E.g.: name=MyStruct.
	NameKey PairKey = "name"

	// CmdNameKey specifies the name for the generated command.
	CmdNameKey PairKey = "cmdName"

	// VarNameKey names the generated command variable name.
	//
	// If undpecified name is generated from the command name such that the
	// command name is appended with "Var" suffix, e.g. "CommandVar".
	VarNameKey PairKey = "varName"

	// NoDeclareVarKey specifies that the variable for the command should not be
	// declared. This is useful if the variable is already declared in some
	// other file in the package.
	NoDeclareVarKey PairKey = "noDeclareVar"

	// HandlerNameKey specifies the name for the command handler.
	//
	// If not specified defaults to name of generated command immediatelly
	// followed with "Handler."
	HandlerNameKey PairKey = "handlerName"

	// GenHandlerKey if specified generates the handler stub for the command.
	GenHandlerKey PairKey = "genHandler"

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

	// IgnoreKey is read from struct fields and specifies that the tagged field
	// should be excluded from generated command options.
	//
	// It takes no values.
	IgnoreKey PairKey = "ignore"

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
	IncludeKey, NameKey, VarNameKey, NoDeclareVarKey, CmdNameKey, GenHandlerKey,
	HandlerNameKey, HelpKey, IgnoreKey, OptionalKey, RequiredKey,
}

// Config is the [Generate] configuration.
//
// A [Config] with defaulted field values is returned by [Default].
type Config struct {

	// Packages is a list of packages to parse. It is a list of relative or full
	// paths to go packages or import paths.
	//
	// If no packages are specified parses all packages in the current
	// directory recursively.
	//
	// Default: "./..."
	Packages []string `cmdline:"name=packages" json:"packages,omitempty"`

	// OutputFile is the output file that will contain generated commands.
	//
	// It can be a full or relative path to a go file.
	//
	// Default "cmdline.go"
	OutputFile string `cmdline:"name=output-file,required" json:"outputFile,omitempty"`

	// PackageName is the name of the package generated file belongs to.
	//
	// Defaults to base name of the current directory.
	PackageName string `cmdline:"package-name" json:"packageName,omitempty"`

	// TagKey is the name of the tag key whose value is read by cmdline from
	// struct tags or doc comments.
	//
	// Default: "cmdline"
	TagKey string `cmdline:"name=tag-key" json:"tagName,omitempty"`

	// HelpFromTag if true Adds option help from HelpTag.
	//
	// Default: true
	HelpFromTag bool `cmdline:"name=help-from-tag" json:"helpFromTag,omitempty"`

	// HelpFromDocs if true adds option help from srtuct field docs.
	//
	// Default: true
	HelpFromDocs bool `cmdline:"name=help-from-docs" json:"helpFromDocs,omitempty"`

	// ErrorOnUnsupportedField if true throws an error during parse if an
	// unsupported field was found in a source struct.
	//
	// Default: false
	ErrorOnUnsupportedField bool `cmdline:"name=error-on-unsupported-field" json:"errorOnUnsupportedField,omitempty"`

	// Print prints the output to stdout.
	//
	// Default: true
	Print bool `cmdline:"name=print-to-stdout" json:"print"`

	// NoWrite if true disables writing to output file.
	//
	// Default: false
	NoWrite bool `cmdline:"name=no-write"`

	// BastConfig is the bastard ast config.
	BastConfig *bast.Config `json:"-"`

	// Model is the parsed model.
	Model `json:"-"`

	// bast is the parsed bast.
	bast *bast.Bast `json:"-"`
}

// Default returns the default [Config].
func Default() (c *Config) {
	c = new(Config)
	c.TagKey = DefaultTagKey
	c.OutputFile = DefaultOutputFile
	c.HelpFromTag = true
	c.HelpFromDocs = true
	c.Print = true
	c.NoWrite = false
	c.ErrorOnUnsupportedField = false
	c.BastConfig = bast.DefaultConfig()
	c.BastConfig.TypeCheckingErrors = false
	c.Model.ImportMap = make(ImportMap)
	return
}

// Generate generates the go source code containing cmdline.Command definitions.
//
// It skips the structs that have no cmdline tags. Structs that are to be used
// as generate source must have the NameTag at minimum.
//
// The struct has to have at least one cmdline tag to be parsed.
func Generate(config *Config) (err error) {

	if config.TagKey == "" {
		config.TagKey = DefaultTagKey
	}
	if config.OutputFile == "" {
		config.OutputFile = DefaultOutputFile
	}
	if len(config.Packages) == 0 {
		config.Packages = append(config.Packages, "./...")
	}
	if config.PackageName == "" {
		var dir string
		if dir, err = os.Getwd(); err != nil {
			return fmt.Errorf("get current dir: %w", err)
		}
		config.PackageName = filepath.Base(dir)
	}

	if config.bast, err = bast.Load(config.BastConfig, config.Packages...); err != nil {
		return
	}

	if config.Model.ImportMap == nil {
		config.Model.ImportMap = make(ImportMap)
	}
	config.Model.ImportMap["github.com/vedranvuk/cmdline"] = ""

	for _, s := range config.bast.AllStructs() {

		var tag = strutils.Tag{
			KnownPairKeys:     knownPairKeys,
			TagKey:            config.TagKey,
			ErrorOnUnknownKey: true,
		}

		for _, line := range config.uncommentDocs(s.Doc) {
			if err = tag.Parse(line); err != nil {
				if err != strutils.ErrTagNotFound {
					return
				}
			}
		}

		if len(tag.Values) == 0 {
			continue
		}

		var c = Command{
			Name:                    s.Name,
			VarName:                 tag.First(VarNameKey),
			NoDeclareVar:            tag.Exists(NoDeclareVarKey),
			CmdName:                 tag.First(CmdNameKey),
			HandlerName:             tag.First(HandlerNameKey),
			GenerateHandler:         tag.Exists(GenHandlerKey),
			Help:                    strings.Join(tag.Values[HelpKey], "\n"),
			SourceStructType:        s.Name,
			SourceStructPackageName: filepath.Base(s.GetPackage().Path),
		}

		if tag.Exists(NameKey) {
			if tag.First(NameKey) == "" {
				err = errors.New("invalid name tag, no value")
			}
			c.Name = tag.First(NameKey)
		}

		var (
			optional = tag.Exists(OptionalKey)
			required = tag.Exists(RequiredKey)
		)
		if optional && required {
			err = errors.New("optional and required keys are mutually exclusive")
			return
		}

		config.Model.ImportMap[s.GetPackage().Path] = s.GetPackage().Name

		if err = config.parseStruct(s, "", &c); err != nil {
			return
		}

		config.Model.Commands = append(config.Model.Commands, c)
	}

	if err = config.generateOutput(); err != nil {
		return
	}

	return nil
}

// parseStruct parses a struct definition into a command.
func (self *Config) parseStruct(s *bast.Struct, path string, c *Command) (err error) {

	for _, f := range s.Fields.Values() {
		if err = self.parseField(f, path, c); err != nil {
			return
		}
	}

	generateShortNames(c)

	return nil
}

// parseField parses a struct field into a command option.
func (self *Config) parseField(f *bast.Field, path string, c *Command) (err error) {

	if path != "" {
		path += "."
	}
	if path += f.Name; f.Name == "" {
		path += f.Type
	}

	if s := f.GetFile().Struct(f.Type); s != nil {
		return self.parseStruct(s, path, c)
	}

	if imp := f.ImportSpecBySelectorExpr(f.Type); imp != nil {
		if _, name, valid := strings.Cut(f.Type, "."); valid {
			if s := self.bast.PkgStruct(imp.Path, name); s != nil {
				self.Model.ImportMap[imp.Path] = imp.Name
				return self.parseStruct(s, path, c)
			}
		}
	}

	var tag = strutils.Tag{
		KnownPairKeys:     knownPairKeys,
		TagKey:            self.TagKey,
		ErrorOnUnknownKey: true,
	}

	if tag.Exists(IgnoreKey) {
		return nil
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

	var opt = &Option{
		LongName:  name,
		ShortName: "",
		Help:      self.makeHelp(tag.Values[HelpKey], f.Doc),
		FieldName: f.Name,
		FieldPath: path,
	}
	switch opt.BasicType = self.bast.ResolveBasicType(f.Type); opt.BasicType {
	case "bool":
		opt.Kind = cmdline.Boolean
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"string", "[]string":
		if optional {
			opt.Kind = cmdline.Optional
		}
		if required {
			opt.Kind = cmdline.Required
		}
	case "":
		if f.Type == "time.Time" {
			if optional {
				opt.Kind = cmdline.Optional
			}
			if required {
				opt.Kind = cmdline.Required
			}
			break
		}
		if self.ErrorOnUnsupportedField {
			return errors.New("unsupported field type: " + f.Type)
		}
		log.Printf("Cannot determine basic type for field %s, skipping.\n", f.Type)
		return nil
	default:
		log.Printf("Unknown basic type: %s\n", opt.BasicType)
		return nil
	}

	c.Options = append(c.Options, opt)

	return nil
}

// generateOutput generates output go file with command definitions.
func (self *Config) generateOutput() (err error) {

	const tmplName = "generate.declarative.tmpl"

	var buf []byte
	if buf, err = fs.ReadFile(FS(), tmplName); err != nil {
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
func (self *Config) uncommentDocs(in []string) (out []string) {
	out = make([]string, 0, len(in))
	for _, line := range in {
		out = append(out, strings.TrimSpace(strings.TrimPrefix(line, "//")))
	}
	return
}

// helpFromDoc generates help from tag and doc comment.
//
// It strips comment prefixes from each doc line.
func (self *Config) makeHelp(tag, doc []string) string {
	const col = 80
	var out []string
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
		if strings.HasPrefix(line, self.TagKey+":") {
			continue
		}
		out = append(out, line)
	}
	out = strutils.WrapText(strings.Join(out, " "), col, false)
	return strings.Join(out, "\\n")
}

// generateShortNames generates short Option names.
func generateShortNames(c *Command) {
	// Sequentially go through options, setting shortcmd to lowercase forst
	// letter from longname. Each time check if it is already used and advance
	// to next letter in longname until unique or exhausted.
	for idx, option := range c.Options {
		var name = strings.ToLower(option.LongName)
	GenShort:
		for _, r := range name {
			option.ShortName = string(r)
			for i := 0; i < idx; i++ {
				if c.Options[i].ShortName == option.ShortName {
					continue GenShort
				}
			}
			break GenShort
		}
	}
	// Check all short names for duplicates and unset duplicates.
	for _, option := range c.Options {
		var n = option.ShortName
		for _, other := range c.Options {
			if option == other {
				continue
			}
			if n == other.ShortName {
				other.ShortName = ""
			}
		}
	}
}
