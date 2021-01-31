// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	var value string
	var handler = func(ctx Context) error {
		if !ctx.Parsed("bar") {
			return errors.New("'bar' not parsed")
		}
		if ctx.RawValue("bar") != "baz" {
			return errors.New("value not 'bar'")
		}
		return nil
	}
	var commands = NewCommands(nil).MustAdd("foo", "", handler).
		Parameters().MustAddNamed("bar", "", "", true, &value).
		Parent()
	var err error
	if err = ParseRaw(commands, "foo", "--bar", "baz"); err != nil {
		t.Fatal(err)
	}
}

func TestParseRaw(t *testing.T) {
	var value string
	var handler = func(ctx Context) error {
		if !ctx.Parsed("bar") {
			return errors.New("'bar' not parsed")
		}
		if ctx.RawValue("bar") != "baz" {
			return errors.New("value not 'bar'")
		}
		if len(ctx.Extra()) != 1 {
			return errors.New("invalid number of extra arguments")
		}
		if ctx.Extra()[0] != "bat" {
			return errors.New("invalid extra arguments")
		}
		return nil
	}
	var commands = NewCommands(nil).MustAdd("foo", "", handler).
		Parameters().MustAddNamed("bar", "", "", true, &value).
		Parent()
	var err error
	if err = ParseRaw(commands, "foo", "--bar", "baz", "bat"); err != nil {
		t.Fatal(err)
	}
}
