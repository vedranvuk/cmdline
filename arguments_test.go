// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import "testing"

type ArgumentTestResult struct {
	Name  string
	Value string
	Kind  Argument
}

type ArgumentTest struct {
	Arguments       Arguments
	ExpectedResults []ArgumentTestResult
}

var UnixArgumentsTests = []ArgumentTest{
	{
		NewUnixArguments("", "--", "=bar", "-f", "--foo", "-foo", "foo", "--foo=bar", "---foo", "--f o o"),
		[]ArgumentTestResult{
			{"", "", InvalidArgument},
			{"", "", InvalidArgument},
			{"", "", InvalidArgument},
			{"f", "", ShortArgument},
			{"foo", "", LongArgument},
			{"foo", "", CombinedArgument},
			{"foo", "", TextArgument},
			{"foo", "bar", AssignmentArgument},
			{"", "", InvalidArgument},
			{"f o o", "", InvalidArgument},
			{"", "", NoArgument},
		},
	},
}

func TestUnixArguments(t *testing.T) {
	var test ArgumentTest
	var result ArgumentTestResult
	for _, test = range UnixArgumentsTests {
		for _, result = range test.ExpectedResults {
			if test.Arguments.Kind() != result.Kind {
				t.Fatalf("TestUnixArguments(%s) failed, Kind mismatch, expected '%s', got '%s'", test.Arguments.Raw(), result.Kind, test.Arguments.Kind())
			}
			if test.Arguments.Name() != result.Name {
				t.Fatalf("TestUnixArguments(%s) failed, Name mismatch, expected '%s', got '%s'", test.Arguments.Raw(), result.Name, test.Arguments.Name())
			}
			if test.Arguments.Value() != result.Value {
				t.Fatalf("TestUnixArguments(%s) failed, Value mismatch, expected '%s', got '%s'", test.Arguments.Raw(), result.Value, test.Arguments.Value())
			}
			test.Arguments.Advance()
		}
	}
}

func BenchmarkArguments(b *testing.B) {
	var tests = make([]Arguments, 0, b.N)
	for i := 0; i < b.N; i++ {
		tests = append(tests, NewUnixArguments("--foo", "-foo", "-f", "foo", "--foo=bar"))
	}
	for i := 0; i < b.N; i++ {
		for tests[i].Advance() {}
	}
}
