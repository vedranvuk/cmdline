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
	var verbose bool
	var config = cmdline.DefaultOS()
	config.PrintInDefinedOrder = true
	config.Globals.BooleanVar("verbose", "v", "Enable verbose output.", &verbose)

	config.Commands.Register(cmdline.HelpCommand(topicMap))
	config.Commands.Handle("generate", "Generates go code that parses arguments to structs.", handleGenerate).
		SetDoc(generateDoc).
		Options.
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
	config.Commands.Handle("version", "Show version.", handleVersion)

	if err := config.Parse(nil); err != nil && err != cmdline.ErrNoArgs {
		log.Fatal(err)
	}
}

func handleVersion(c cmdline.Context) error {
	fmt.Fprintf(c.Config().GetOutput(), "%s %s\n", filepath.Base(os.Args[0]), version)
	return nil
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

// topicMap maps topic names to topic contents.
var topicMap = cmdline.TopicMap{
	"tags": tagsDoc,
}

const (
	// generateDoc is the "generate" command doc.
	generateDoc = `Generate generates go source containing command line interface that
maps structs to commands and their fields to options.`
	// tagsDoc is the "tags" topic doc.
	tagsDoc = `The following tags are read from source structs and their fields and configure
the generated cmdline interface.

Any struct in packages parsed by cmdline that contains a cmdline tag in its doc 
comments is included as a generated cmdline command.


"include"

A placeholder key read from source struct docs tag that can be used for when a 
command is to be generated from the tagged struct but there is no need to 
specify any other options for generated code.

By default, structs that have no cmdline tags are skipped so a struct can
be tagged with this key to be included and use all default options. It takes no 
value.


"ignore"

Read from struct fields and specifies that the tagged field should be excluded 
when generating command options from fields. It takes no value.


"help"

Used on a source struct and specifies the help text to be set with the command 
that will represent the struct.

It takes a single value in the key=value format that defines the command help. 
E.g.: help=This is a help text for ca command representing a bound struct.

Help text cannnot span multiple lines, it is a one-line shown to user when 
cmdline config help is requested.


"name"

Used on a source struct and specifies the name of the command that represents 
the struct being bound to. This name is the name by which source struct commands
and their field options are addressed from the command line.

It takes a single value in the key=value format that defines the command
name. E.g.: name=MyStruct.


"shortName"

Used on source struct field specifies the generated option short name.
If a duplicate is set it will be unset automatically during generation.


"commandName"

Specifies the name of the generated command.


"targetName"

Names variable of the output struct command options will write to. This may name
a variable declared in some other package file that the generated command 
options can adress and write from arguments or name the variable that will be 
generated in the output file so some other file in the package can address it.

If unspecified, name is generated from the command name such that the
command name is appended with "Var" suffix, e.g. "CommandVar".


"handlerName"

Specifies the name for the command handler. If not specified defaults to name of
generated command immediatelly followed with "Handler."


"genTarget"

Specifies that the variable for the command should be declared. This is useful 
if the variable is already declared in some other file in the package.
Generated commands will still address the target variable defined by targetName.
It takes no values.


"genHandler"

If specified on a source struct will generate the command handler stub. It takes
no values.


"optional"

Read from struct fields and specifies that the tagged field should use the 
Optional option. It takes no values.


"required"

Read from struct fields and specifies that the tagged field should use the 
Required option. It takes no values.
`
)
