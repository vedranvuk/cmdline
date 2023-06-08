package cmdline

import (
	"fmt"
	"testing"
)

func TestOptions(t *testing.T) {
	opts := new(Options)
	opts.Boolean("verbose", "v", "Be verbose.")
	opts.Optional("force", "f", "", "Force it.")
	opts.Optional("directory", "d", "string", "Enter directory name.")
	opts.Indexed("name", "string", "Input name")
	opts.Variadic("files", "Specify file names.")

	args := []string{"-v", "--force", "-d=/home/yourname", "myname", "arg1", "arg2", "arg3"}
	fmt.Printf("%#v\n", Parse(args, nil, opts))
	for _, v := range opts.options {
		fmt.Printf("%#v\n", v)
	}

	for _, o := range opts.options {
		o.parsed = false
	}
	args = []string{}
	fmt.Printf("%#v\n", Parse(args, nil, opts))
	for _, v := range opts.options {
		fmt.Printf("%#v\n", v)
	}


}
