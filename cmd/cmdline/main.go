package main

import (
	"log"

	"github.com/vedranvuk/cmdline"
)

func main() {
	var config = cmdline.Default()

	config.Commands.Handle("help", "h", cmdline.HelpHandler)
	config.Commands.Handle(
		"generate",
		"Generates commandline classes.",
		func(c cmdline.Context) error {
			var gc = &GenerateConfig{
				Packages:    c.Values("packages"),
				OutputFile:  c.Values("output-file").First(),
				PackageName: c.Values("package-name").First(),
				TagKey:      c.Values("tag-key").First(),
			}
			return gc.Generate()
		},
	).Options.
		Boolean("help-from-tag", "g", "Include help from tag.").
		Boolean("help-from-doc", "d", "Include help from doc comments.").
		Boolean("error-on-unsupported-field", "e", "Throws an error if unsupporrted field was encountered.").
		Boolean("print", "r", "Print output.").
		Boolean("no-wrote", "n", "Do not write output file.").
		Optional("output-file", "o", "Output file name.").
		Optional("tag-key", "t", "Name of the tag key to parse.").
		Required("package-name", "p", "Name of the package output go file belongs to.").
		Variadic("packages", "Packages to parse.")

	if err := config.Parse(nil); err != nil {
		log.Fatal(err)
	}
}
