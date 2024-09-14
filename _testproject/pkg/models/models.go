package models

// testTag:"name=options"
// testTag:"help=Defines a set of options."
type Options struct {
	// OutputDirectory is the output directory.
	OutputDirectory string `testTag:"name=outDir,required"`
}
