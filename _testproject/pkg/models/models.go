package models

// Options is a demo struct.
// cmdline:"name=options"
// cmdline:"help=Defines a set of options."
type Options struct {
	// OutputDirectory is the output directory.
	//
	// This is a multiline comment.
	//cmdline:"help=Output directory."
	OutputDirectory string `testTag:"name=outDir,required"`
}

type Config struct {
	Name       string
	Age        int
	Subscribed bool
}
