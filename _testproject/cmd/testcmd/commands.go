// Code generated by github.com/vedranvuk/cmdline DO NOT EDIT.

package main

import (
	"os"

	"github.com/vedranvuk/cmdline"
	models "github.com/vedranvuk/cmdline/_testproject/pkg/models"
)

var (
	OptionsVar = new(models.Options)
	ConfigVar  = new(models.Config)
)

var (
	OptionsCmd = &cmdline.Command{
		Name: "Options",
		Help: "Defines a set of options.",
		Options: []cmdline.Option{
			&cmdline.Optional{
				LongName:    "OutputDirectory",
				ShortName:   "",
				Help:        "[Output directory.  OutputDirectory is the output directory.  This is a multiline comment.]",
				MappedValue: &OptionsVar.OutputDirectory,
			},
		},
		Handler: cmdline.NopHandler,
	}
	ConfigCmd = &cmdline.Command{
		Name: "Config",
		Help: "",
		Options: []cmdline.Option{
			&cmdline.Optional{
				LongName:    "Name",
				ShortName:   "",
				Help:        "[]",
				MappedValue: &ConfigVar.Name,
			},
			&cmdline.Optional{
				LongName:    "Age",
				ShortName:   "",
				Help:        "[]",
				MappedValue: &ConfigVar.Age,
			},
			&cmdline.Boolean{
				LongName:    "Subscribed",
				ShortName:   "",
				Help:        "[]",
				MappedValue: &ConfigVar.Subscribed,
			},
		},
		Handler: cmdline.NopHandler,
	}
)

// parseCmdLine parses the command line into defined commands.
func cmdlineConfig() (*cmdline.Config, error) {

	var config = &cmdline.Config{
		Arguments: os.Args[1:],
		Commands: cmdline.Commands{
			OptionsCmd,
			ConfigCmd,
		},
	}

	config.Commands.Register(
		&cmdline.Command{
			Name: "help",
			Help: "Shows help.",
			Handler: func(c cmdline.Context) error {
				cmdline.PrintConfig(os.Stdout, config)
				return nil
			},
		},
	)

	if err := config.Parse(context.Background()); err != nil {
		log.Fatal(err)
	}

	return config, nil
}
