package main

import "embed"

//go:embed commands.tmpl
var resources embed.FS

// FS returns the embedded resources as a file system.
func FS() embed.FS { return resources }

