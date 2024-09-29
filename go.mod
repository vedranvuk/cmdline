module github.com/vedranvuk/cmdline

go 1.23.1

require (
	github.com/vedranvuk/bast v0.0.0-20240929120422-2c6245b25ba6
	github.com/vedranvuk/strutils v0.0.0-20240929151339-c677c08dd040
)

require (
	github.com/vedranvuk/ds v0.0.0-20240913183506-6b66a044517c // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
)

replace github.com/vedranvuk/strutils => ../strutils

replace github.com/vedranvuk/bast => ../bast
