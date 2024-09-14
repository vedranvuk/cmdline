package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vedranvuk/cmdline"
)

const helpText = `cmdline generates commands that bind their options to a struct
and modify target truct values via command line options.`

func main() {
	var config = &cmdline.Config{

		Arguments: os.Args[1:],

		Commands: cmdline.Commands{

			&cmdline.Command{
				Name: "help",
				Help: "Show help",
				Handler: func(c cmdline.Context) error {
					fmt.Print(helpText)
					return nil
				},
			},

			&cmdline.Command{
				Name: "generate",
				Help: "Generates commands from structs.",
				Options: cmdline.Options{
					&cmdline.Optional{
						LongName:  "tagName",
						ShortName: "t",
						Help:      "Specify name of tag to be interpreted as cmdline tag.",
					},
					&cmdline.Required{
						LongName:  "outputFile",
						ShortName: "o",
						Help:      "Filename of the output go file to contain generated commands.",
					},
					&cmdline.Repeated{
						LongName:  "packages",
						ShortName: "p",
						Help:      "Packages to parse",
					},
				},
				Handler: func(c cmdline.Context) error {
					var gc = &GenerateConfig{}
					c.SetIfParsed("tagName", &gc.TagName)
					c.SetIfParsed("outputFile", &gc.OutputFile)
					return nil
				},
			},
		},
	}

	if err := config.Parse(context.Background()); err != nil {
		if err == cmdline.ErrNoArgs {
			usage(config)
			os.Exit(0)
		}
		log.Fatal(err)
	}
}

func usage(config *cmdline.Config) {

}
