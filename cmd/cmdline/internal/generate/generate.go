// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// TODO Escape doc comments for quotes, etc.

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

//go:embed generate.declarative.tmpl generate.chained.tmpl
var resources embed.FS

// FS returns the embedded resources as a file system.
func FS() embed.FS { return resources }

const (
	ChainedTmplName     = "generate.chained.tmpl"
	DeclarativeTmplName = "generate.declarative.tmpl"
)

const (
	// DefaultTagKey is the default key of a tag value parsed by cmdline.
	DefaultTagKey = "cmdline"

	// DefaultOutputFile is the default base name of an output go file that
	// will contain generated code.
	DefaultOutputFile = "cmdline.go"

	// DefaultConfigFileName is the default cmdline config file name.
	DefaultConfigFileName = "cmdline.json"
)

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

	// Template is the filename of the template to use for code generation.
	//
	// If it is a value of [GenChainedTmplName] or [GenDeclarativeTmplName]
	// specified template is used, otherwise it is read from the file specified
	// by the field value.
	//
	// Default: GenChainedTmplName
	Template string `cmdline:"template" json:"template,omitempty"`

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
	c.Template = ChainedTmplName
	c.NoWrite = false
	c.ErrorOnUnsupportedField = false
	c.BastConfig = bast.DefaultConfig()
	c.BastConfig.TypeCheckingErrors = false
	return
}

// Generate generates the go source code containing cmdline.Command definitions.
//
// It skips the structs that have no cmdline tags. Structs that are to be used
// as generate source must have the NameTag at minimum.
//
// The struct has to have at least one cmdline tag to be parsed.
func Generate(config *Config) (err error) {

	// TODO Check name colisions, vars, commands, etc.

	if config.TagKey == "" {
		config.TagKey = DefaultTagKey
	}
	if config.OutputFile == "" {
		config.OutputFile = DefaultOutputFile
	}
	if config.Packages == nil {
		config.Packages = append(config.Packages, "./...")
	}
	if config.Template == "" {
		config.Template = ChainedTmplName
	}
	if config.PackageName == "" {
		var dir string
		if dir, err = os.Getwd(); err != nil {
			return fmt.Errorf("get current dir: %w", err)
		}
		config.PackageName = filepath.Base(dir)
	}
	if config.BastConfig == nil {
		config.BastConfig = bast.DefaultConfig()
		config.BastConfig.TypeCheckingErrors = false
	}

	if config.bast, err = bast.Load(config.BastConfig, config.Packages...); err != nil {
		return
	}

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
			Name:              tag.First(NameKey),
			TargetName:        tag.First(TargetNameKey),
			CommandName:       tag.First(CommandNameKey),
			HandlerName:       tag.First(HandlerNameKey),
			GenTarget:         tag.Exists(GenTargetKey),
			GenHandler:        tag.Exists(GenHandlerKey),
			Help:              config.makeHelp(tag.Values[HelpKey], s.Doc),
			SourceType:        s.Name,
			SourcePackagePath: s.GetPackage().Path,
			SourcePackageName: filepath.Base(s.GetPackage().Path),
		}
		if c.Name == "" {
			c.Name = s.Name
		}
		if c.TargetName == "" {
			c.TargetName = strutils.CamelCase(c.Name) + "Var"
		}
		if c.CommandName == "" {
			c.CommandName = strutils.CamelCase(c.SourceType) + "Cmd"
		}
		if c.HandlerName == "" {
			c.HandlerName = c.CommandName + "Handler"
		}
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

	generateOptionShortNames(c)

	return nil
}

// parseField parses a struct field into a command option.
func (self *Config) parseField(f *bast.Field, path string, c *Command) (err error) {

	if path != "" {
		path += "."
	}
	if f.Unnamed {
		path += typeNameFromSelector(f.Type)
	} else {
		path += f.Name
	}

	if s := f.GetPackage().Struct(f.Type); s != nil {
		return self.parseStruct(s, path, c)
	}

	if imp := f.GetFile().ImportSpecFromSelector(f.Type); imp != nil {
		if s := self.bast.PkgStruct(imp.Path, typeNameFromSelector(f.Type)); s != nil {
			c.AddImport(imp.Path)
			return self.parseStruct(s, path, c)
		}
	}

	var tag = strutils.Tag{
		KnownPairKeys:     knownPairKeys,
		TagKey:            self.TagKey,
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

	if tag.Exists(IgnoreKey) {
		return nil
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
		LongName:        tag.First(NameKey),
		ShortName:       tag.First(ShortNameKey),
		Help:            self.makeHelp(tag.Values[HelpKey], f.Doc),
		SourceFieldName: f.Name,
		SourceFieldPath: path,
	}
	if opt.LongName == "" {
		opt.LongName = f.Name
	}
	switch opt.SourceBasicType = self.bast.ResolveBasicType(f.Type); opt.SourceBasicType {
	case "bool":
		opt.Kind = cmdline.Boolean
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"string":
		if optional {
			opt.Kind = cmdline.Optional
		}
		if required {
			opt.Kind = cmdline.Required
		}
	case "[]string":
		opt.Kind = cmdline.Variadic
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
		log.Printf("cmdline: cannot determine basic type for field %s, skipping.\n", f.Type)
		return nil
	default:
		log.Printf("cmdline: unknown basic type: %s\n", opt.SourceBasicType)
		return nil
	}

	c.Options = append(c.Options, opt)

	return nil
}

// typeNameFromSelector returns the type name without the package prefix from
// a selector expression. If not a selector expression returns input as is.
func typeNameFromSelector(selectorExpr string) string {
	if _, name, selector := strings.Cut(selectorExpr, "."); selector {
		return name
	}
	return selectorExpr
}

// generateOutput generates output go file with command definitions.
func (self *Config) generateOutput() (err error) {

	var (
		t  *template.Template
		tt string
		bb = bytes.NewBuffer(nil)
		m  = parse.ParseComments | parse.SkipFuncCheck
		s  []byte
	)

	if tt, err = loadTemplate(self.Template); err != nil {
		return
	}
	if t, err = parseTemplateWithMode("cmdline", tt, m); err != nil {
		return fmt.Errorf("parse output template: %w", err)
	}
	if err = t.Execute(bb, self); err != nil {
		return fmt.Errorf("execute output template: %w", err)
	}
	if s, err = format.Source(bb.Bytes()); err != nil {
		return fmt.Errorf("format output: %w", err)
	}
	if self.Print {
		if _, err = fmt.Print(string(s)); err != nil {
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
		if _, err = file.Write(s); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
	}

	return nil
}

// loadTemplate loads the template text depending on filename.
// If filename is one of built-in template names it is loaded.
// If filename is any other filename it is loaded from disk.
func loadTemplate(filename string) (text string, err error) {
	var buf []byte
	switch filename {
	case ChainedTmplName, DeclarativeTmplName:
		buf, err = fs.ReadFile(FS(), filename)
	default:
		buf, err = os.ReadFile(filename)
	}
	if err != nil {
		return "", fmt.Errorf("load template: %w", err)
	}
	return string(buf), nil
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

// makeHelp extracts help from tag and doc depending on config and returns a
// formatted string to be set as [Command] or [Option] help.
//
// It strips comment prefixes from each doc line.
//
// Right now it wraps the result to 80 columns.
func (self *Config) makeHelp(tag, doc []string) string {
	// TODO Get terminal width, calc help width, include other columns in calc,
	// basically redo printing/tabwriting.
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
	if self.HelpFromTag {
		for _, line := range tag {
			if line == "" {
				continue
			}
			out = append(out, line)
		}
	}
	if lt > 0 && ld > 0 {
		out = append(out, "")
	}
	if self.HelpFromDocs {
		for _, line := range doc {
			line = strings.TrimSpace(strings.TrimPrefix(line, "//"))
			if strings.HasPrefix(line, "go:") {
				continue
			}
			if strings.HasPrefix(line, self.TagKey+":") {
				continue
			}
			if line == "" {
				continue
			}
			out = append(out, line)
		}
	}
	out = strutils.WrapText(strings.Join(out, " "), col, false)
	return strings.Join(out, "\\n")
}

// generateOptionShortNames generates short Option names.
//
// The algorithm is trivial; a single pass of generating shortname from
// lowercased longname, starting with first char and advancing to next if
// non-unique in set thus far, until unique or exhausted.
//
// Second pass makes sure all short options are unique in set, and if not, the
// latter option shortname is unset.
//
// If a unique letter was not generated from long name option gets no short
// option name.
func generateOptionShortNames(c *Command) {
	// Sequentially go through options, setting shortcmd to lowercase first
	// letter from longname. Each time check if it is already used and advance
	// to next letter until unique or exhausted.
	for idx, option := range c.Options {
		if option.ShortName != "" {
			continue
		}
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
