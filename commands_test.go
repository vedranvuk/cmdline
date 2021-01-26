// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"testing"
)

type CommandRun struct {
	Arguments
	ExpectedError error
}

type CommandTest struct {
	Test string
	*Commands
	Runs []CommandRun
}

func (test *CommandTest) RunCommandTest() error {
	var err error
	var run CommandRun
	for _, run = range test.Runs {
		test.Commands.Reset()
		if err = test.Commands.Parse(run.Arguments, nil); !errors.Is(err, run.ExpectedError) {
			return err
		}
	}
	return nil
}

var CommandTests = []CommandTest{
	CommandTest{
		"Basic",
		NewCommands(nil).MustAdd("foo", "", nil).
			Parameters().MustAddNamed("foo", "", "", false, nil).
			Parameters().MustAddNamed("bar", "", "", false, nil).
			Parameters().MustAddNamed("baz", "", "", false, nil).
			Parent(),
		[]CommandRun{
			{
				NewArguments("foo"),
				nil,
			},
			{
				NewArguments("foo", "--foo", "--bar", "--baz"),
				nil,
			},
			{
				NewArguments(),
				ErrNoArguments,
			},
			{
				NewArguments("foo", "bar"),
				ErrExtraArguments,
			},
			{
				NewArguments("bar"),
				ErrCommandNotFound,
			},
			{
				NewArguments("-foo"),
				ErrCommandNotFound,
			},
		},
	},
}

func TestCommands(t *testing.T) {
	var test CommandTest
	for _, test = range CommandTests {
		t.Run(test.Test, func(t *testing.T) {
			var err error
			if err = test.RunCommandTest(); err != nil {
				t.Error(err)
			}
		})
	}
}
