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

	var (
		config = cmdline.DefaultOS()
	)
	config.PrintInDefinedOrder = true

	// Help command and subcommands.
	var help = config.Commands.Handle("help", "Show help.", func(c cmdline.Context) error {
		if c.Parsed("topic") {
			if text, exists := help[c.Values("topic").First()]; exists {
				fmt.Fprint(c.Config().GetOutput(), text)
				return nil
			}
			return errors.New("help topic not found: " + c.Values("topic").First())
		}
		return cmdline.HelpHandler(c)
	})
	help.Options.Variadic("topic", "Show help for a specific topic.")

	// Help topics command.
	config.Commands.Handle("help-topics", "Show help topics.", func(c cmdline.Context) error {
		var tw = tabwriter.NewWriter(c.Config().GetOutput(), 2, 2, 2, 32, 0)
		fmt.Fprintf(tw, "%s\t%s\n", "generate", "Show help on generate command.")
		tw.Flush()
		return nil
	})

	// Version command.
	config.Commands.Handle("version", "Show version.", func(c cmdline.Context) error {
		fmt.Fprintf(c.Config().GetOutput(), "%s %s\n", filepath.Base(os.Args[0]), version)
		return nil
	})

	// Generate command.
	config.Commands.Handle(
		"generate",
		"Generates go code that parses arguments to structs.",
		func(c cmdline.Context) error {
			var config = &generate.Config{
				Packages:                c.Values("packages"),
				OutputFile:              c.Values("output-file").First(),
				PackageName:             c.Values("package-name").First(),
				TagKey:                  c.Values("tag-key").First(),
				NoWrite:                 c.Parsed("no-write"),
				Print:                   c.Parsed("print"),
				ErrorOnUnsupportedField: c.Parsed("error-on-unsupported-field"),
				HelpFromTag:             c.Parsed("help-from-tag"),
				HelpFromDocs:            c.Parsed("help-from-doc"),
				BastConfig:              bast.DefaultConfig(),
			}
			config.BastConfig.Dir = c.Values("build-dir").First()
			return generate.Generate(config)
		},
	).Options.
		Required("package-name", "p", "Name of the package output go file belongs to.").
		Optional("output-file", "o", "Output file name.").
		Optional("tag-key", "t", "Name of the tag key to parse from docs and struct tags.").
		Optional("build-dir", "b", "Specify build directory.").
		Boolean("help-from-tag", "g", "Include help from tag.").
		Boolean("help-from-doc", "d", "Include help from doc comments.").
		Boolean("error-on-unsupported-field", "e", "Throws an error if unsupporrted field was encountered.").
		Boolean("print", "r", "Print output.").
		Boolean("no-write", "n", "Do not write output file.").
		Variadic("packages", "Packages to parse.")

	// Make config command.
	config.Commands.Handle(
		"makecfg",
		"Write a default configuration file.",
		func(c cmdline.Context) (err error) {
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
		},
	).Options.
		Optional("output-dir", "o", "Output directory.").
		Boolean("force", "f", "Force overwrite if file already exists.")

	// Parse.
	if err := config.Parse(nil); err != nil && err != cmdline.ErrNoArgs {
		log.Fatal(err)
	}
}

var help = map[string]string{

	"generate": ` Generate command

Generate command generates go source that parses command line arguments into go
structs.
`,
}
