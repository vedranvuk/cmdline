package cmdline

import (
	"fmt"
	"os"
	"testing"
)

func TestOptions(t *testing.T) {

	var getOptions = func() *OptionSet {
		opts := new(OptionSet)
		opts.Boolean("verbose", "v", "Be verbose.", "Shows extra debug output.")
		opts.Optional("force", "f", "Force it", "Force something several times.", "int")
		opts.Optional("directory", "d", "dir name", "Enter directory name.", "string")
		opts.Indexed("name", "Input name", "Input the name that is used for naming.", "string")
		opts.Variadic("files", "Specify file names.", "A list of filenames to use.", "strings")
		return opts
	}

	opts := getOptions()

	args := []string{"-v", "--force", "-d=/home/yourname", "myname", "arg1", "arg2", "arg3"}

	err := Parse(&Config{
		Args:    args,
		Globals: opts,
	})
	fmt.Printf("%#v\n", err)

	for _, v := range opts.options {
		fmt.Printf("%#v\n", v)
	}

	opts = getOptions()
	args = []string{}
	err = Parse(&Config{
		Args:    os.Args[1:],
		Globals: opts,
	})
	fmt.Printf("%#v\n", err)
	
	for _, v := range opts.options {
		fmt.Printf("%#v\n", v)
	}

}
