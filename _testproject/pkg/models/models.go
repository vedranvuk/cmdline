package models

import "time"

// Options is a demo struct.
// cmdline:"name=options"
// cmdline:"help=Defines a set of options."
type Options struct {
	// OutputDirectory is the output directory.
	//
	// This is a multiline comment.
	//cmdline:"help=Output directory."
	OutputDirectory string `cmdline:"name=outDir,required"`
}

// cmdline:"name=config"
type Config struct {
	// Name is the name.
	Name string `cmdline:"optional"`
	// Age is the age.
	Age int `cmdline:"required"`
	// Subscribed is usually true.
	Subscribed bool
	// Sub doc comments are ignored.
	Sub
}

// cmdline:"ignore"
type Sub struct {
	// DOB is the darte of birth.
	DOB time.Time
	// EMail is the email address.
	EMail string
}
