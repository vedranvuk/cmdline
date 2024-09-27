// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/vedranvuk/bast"
	"github.com/vedranvuk/cmdline"
	"github.com/vedranvuk/cmdline/cmd/cmdline/internal/generate"
)

const version = "0.0.0-dev"

func main() {
	var config = cmdline.DefaultOS()
	config.PrintInDefinedOrder = true
	config.Commands.Handle("help", "Show help.", handleHelp).Options.
		Variadic("topic", "Show help for a specific topic.")
	config.Commands.Handle("help-topics", "Show help topics.", handleHelpTopics)
	config.Commands.Handle("version", "Show version.", handleVersion)
	config.Commands.Handle("generate", "Generates go code that parses arguments to structs.", handleGenerate).Options.
		Required("package-name", "p", "Name of the package output go file belongs to.").
		Optional("output-file", "o", "Output file name.").
		Optional("tag-key", "k", "Name of the tag key to parse from docs and struct tags.").
		Optional("build-dir", "b", "Specify build directory.").
		// Optional("template", "t", "Template to use for code generation.").
		Boolean("help-from-tag", "g", "Include help from tag.").
		Boolean("help-from-doc", "d", "Include help from doc comments.").
		Boolean("error-on-unsupported-field", "e", "Throws an error if unsupporrted field was encountered.").
		Boolean("print", "r", "Print output.").
		Boolean("no-write", "n", "Do not write output file.").
		Variadic("packages", "Packages to parse.")
	config.Commands.Handle("dump", "Dumps the codegen template.", handleDump).Options.
		Variadic("dir", "Output directory.")
	config.Commands.Handle("makecfg", "Write a default configuration file.", handleMakeCfg).Options.
		Optional("output-dir", "o", "Output directory.").
		Boolean("force", "f", "Force overwrite if file already exists.")

	if err := config.Parse(nil); err != nil && err != cmdline.ErrNoArgs {
		log.Fatal(err)
	}
}

func handleVersion(c cmdline.Context) error {
	fmt.Fprintf(c.Config().GetOutput(), "%s %s\n", filepath.Base(os.Args[0]), version)
	return nil
}

func handleHelp(c cmdline.Context) error {
	if c.Parsed("topic") {
		if text, exists := help[c.Values("topic").First()]; exists {
			fmt.Fprint(c.Config().GetOutput(), text)
			return nil
		}
		return errors.New("help topic not found: " + c.Values("topic").First())
	}
	return cmdline.HelpHandler(c)
}

func handleHelpTopics(c cmdline.Context) error {
	var tw = tabwriter.NewWriter(c.Config().GetOutput(), 2, 2, 2, 32, 0)
	fmt.Fprintf(tw, "%s\t%s\n", "generate", "Show help on generate command.")
	fmt.Fprintf(tw, "%s\t%s\n", "tags", "Show help on cmdline tags.")
	tw.Flush()
	return nil
}

func handleGenerate(c cmdline.Context) error {

	var config = &generate.Config{
		Packages:                c.Values("packages"),
		OutputFile:              c.Values("output-file").First(),
		PackageName:             c.Values("package-name").First(),
		TagKey:                  c.Values("tag-key").First(),
		Template:                c.Values("template").First(),
		NoWrite:                 c.Parsed("no-write"),
		Print:                   c.Parsed("print"),
		ErrorOnUnsupportedField: c.Parsed("error-on-unsupported-field"),
		HelpFromTag:             c.Parsed("help-from-tag"),
		HelpFromDocs:            c.Parsed("help-from-doc"),
		BastConfig:              bast.DefaultConfig(),
	}
	config.BastConfig.Dir = c.Values("build-dir").First()

	return generate.Generate(config)
}

func handleMakeCfg(c cmdline.Context) (err error) {

	var (
		cfg = generate.Default()
		buf []byte
		dir = c.Values("output-dir").First()
	)

	if buf, err = json.MarshalIndent(cfg, "", "\t"); err != nil {
		return
	}

	if dir == "" {
		dir = "."
	}

	var fn = filepath.Join(dir, generate.DefaultConfigFileName)
	if _, err = os.Stat(fn); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return
		}
		if !c.Parsed("force") {
			return errors.New("target config file already exists: " + fn)
		}
	}

	return os.WriteFile(fn, buf, os.ModePerm)

}

func handleDump(c cmdline.Context) error {
	return nil
}

var help = map[string]string{

	"generate": `Generate command

Generate command generates go source that parses command line arguments into go
structs.
`,

	"tags": `cmdline tags
	 
"include"

A placeholder key for when a command line command is to be generated for the 
struct but there is no need to specify any other properties for the generated 
code. By default, structs that have no cmdline tags are skipped.


"name"

Used on a source struct and specifies the name of the command that represents 
the struct being bound to. It is also the name of the generated variable of a 
struct manipulated by the command options. It takes a single value in the 
key=value format that defines the command name. E.g.: name=MyStruct.


"cmdName"

Specifies the name for the generated command.


"varName" 

Names the generated command variable name.If undpecified name is generated from 
the command name such that thecommand name is appended with "Var" suffix, e.g. 
"CommandVar".


"noDeclareVar"
 
NoDeclareVarKey specifies that the variable for the command should not be 
declared. This is useful if the variable is already declared in some other file
in the package.

"handlerName"

HandlerNameKey specifies the name for the command handler. If not specified 
defaults to name of generated command immediatelly followed with "Handler."


"genHandler"

GenHandlerKey if specified generates the handler stub for the command.


"help" 

HelpKey is used on a source struct and specifies the help text to be set with 
the command that will represent the struct. It takes a single value in the 
key=value format that defines the command help. E.g.: help=This is a help text
for ca command representing a bound struct. Help text cannnot span multiple 
lines, it is a one-line shown to user when cmdline config help is requested.


"ignore" 

Read from struct fields and specifies that the tagged fieldshould be excluded 
from generated command options.It takes no values.


"optional"

OptionalKey is read from struct fields and specifies that the tagged field 
should use the Optional option. It takes no values.


"required" 

Eead from struct fields and specifies that the taggedfield should use the 
Required option.It takes no values.
`,
}
