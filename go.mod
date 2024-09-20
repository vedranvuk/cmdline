module github.com/vedranvuk/cmdline

go 1.22.5

toolchain go1.23.1

require (
	github.com/vedranvuk/bast v0.0.0-20240918170510-be4384daa935
	github.com/vedranvuk/strutils v0.0.0-20240920091046-104abbd7218f
)

require (
	github.com/vedranvuk/ds v0.0.0-20240913183506-6b66a044517c // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
)

replace github.com/vedranvuk/strutils => ../strutils

replace github.com/vedranvuk/bast => ../bast
