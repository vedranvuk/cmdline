module github.com/vedranvuk/cmdline

go 1.23.1

require (
	github.com/vedranvuk/bast v0.0.0-20240930101845-7935a6b6fb3e
	github.com/vedranvuk/strutils v0.0.0-20241001110106-58c3aa315164
)

require (
	github.com/vedranvuk/ds v0.0.0-20240930133828-52d795fc7747 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
)

replace github.com/vedranvuk/strutils => ../strutils

replace github.com/vedranvuk/bast => ../bast
