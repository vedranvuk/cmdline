module github.com/vedranvuk/cmdline

go 1.23.1

require (
	github.com/vedranvuk/bast v0.0.0-20241007112749-76ec14b809f7
	github.com/vedranvuk/strutils v0.0.0-20250315164838-9084a2535a47
)

require (
	github.com/vedranvuk/ds v0.0.0-20250101185545-07c58dce50fc // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/tools v0.31.0 // indirect
)

replace github.com/vedranvuk/strutils => ../strutils

replace github.com/vedranvuk/bast => ../bast
