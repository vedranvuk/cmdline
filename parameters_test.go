// Copyright 2020 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type ParameterTest struct {
	Test             string
	Name             string
	Required         bool
	Indexed          bool
	Value            interface{}
	ExpectedParsed   bool
	ExpectedBaseErr  error
	ExpectedRawValue string
	Arguments
}

var ParameterTests = []*ParameterTest{
	{"Named", "foo", false, false, nil, true, nil, "", NewArguments("--foo")},
	{"Named unspecified", "foo", false, false, nil, false, nil, "", NewArguments("--bar")},
	{"Named ommited", "foo", false, false, nil, false, ErrNoArguments, "", NewArguments()},
	{"Named text argument", "foo", false, false, nil, false, nil, "", NewArguments("foo")},
	{"Named invalid argument short", "foo", false, false, nil, false, ErrInvalidArgument, "", NewArguments("-foo")},
	{"Named invalid argument long", "foo", false, false, nil, false, ErrInvalidArgument, "", NewArguments("---foo")},
	{"Named required", "foo", true, false, nil, true, nil, "bar", NewArguments("--foo", "bar")},
	{"Named required unspecified", "foo", true, false, nil, false, nil, "", NewArguments("--bar", "baz")},
	{"Named required ommited", "foo", true, false, nil, false, ErrNoArguments, "", NewArguments()},
	{"Named required text argument", "foo", true, false, nil, false, nil, "", NewArguments("foo")},
	{"Named required invalid argument short", "foo", true, false, nil, false, ErrInvalidArgument, "", NewArguments("-foo")},
	{"Named required invalid argument long", "foo", true, false, nil, false, ErrInvalidArgument, "", NewArguments("---foo")},
	{"Named value", "foo", false, false, "bar", true, nil, "bar", NewArguments("--foo", "bar")},
	{"Named value unspecified", "foo", false, false, nil, false, nil, "", NewArguments("--bar", "baz")},
	{"Named value ommited", "foo", false, false, nil, false, ErrNoArguments, "", NewArguments()},
	{"Named value text argument", "foo", false, false, nil, false, nil, "", NewArguments("foo")},
	{"Named value invalid argument short", "foo", false, false, nil, false, ErrInvalidArgument, "", NewArguments("-foo")},
	{"Named value invalid argument long", "foo", false, false, nil, false, ErrInvalidArgument, "", NewArguments("---foo")},
	{"Named value invalid value argument short", "foo", false, false, nil, false, ErrInvalidArgument, "", NewArguments("--foo", "-foo")},
	{"Named value invalid value argument long", "foo", false, false, nil, false, ErrInvalidArgument, "", NewArguments("--foo", "---foo")},
	{"Named required value", "foo", true, false, "bar", true, nil, "bar", NewArguments("--foo", "bar")},
	{"Named required value unspecified", "foo", true, false, nil, false, nil, "", NewArguments("--bar", "baz")},
	{"Named required value ommited", "foo", true, false, nil, false, ErrNoArguments, "", NewArguments()},
	{"Named required value text argument", "foo", true, false, nil, false, nil, "", NewArguments("foo")},
	{"Named required value invalid argument short", "foo", true, false, nil, false, ErrInvalidArgument, "", NewArguments("-foo")},
	{"Named required value invalid argument long", "foo", true, false, nil, false, ErrInvalidArgument, "", NewArguments("---foo")},
	{"Named required value invalid argument argument", "foo", true, false, nil, false, ErrValueRequired, "", NewArguments("--foo", "--bar")},
	{"Indexed", "foo", false, true, nil, true, nil, "bar", NewArguments("bar")},
	{"Indexed parameter", "foo", false, true, nil, false, nil, "", NewArguments("--foo")},
	{"Indexed required", "foo", true, true, nil, true, nil, "bar", NewArguments("bar")},
	{"Indexed value", "foo", false, true, "bar", true, nil, "bar", NewArguments("bar")},
	{"Indexed required value", "foo", true, true, "bar", true, nil, "bar", NewArguments("bar")},
}

func RunParameterTest(test *ParameterTest) error {
	var value interface{}
	if test.Value != nil {
		value = reflect.New(reflect.ValueOf(test.Value).Type()).Interface()
	}
	var param = NewParameter(test.Name, "", test.Required, test.Indexed, value)
	var err error
	if err = param.Parse(test.Arguments); !errors.Is(err, test.ExpectedBaseErr) {
		return err
	}
	if param.Parsed() != test.ExpectedParsed {
		return fmt.Errorf(
			"%s failed: want parsed '%t', got '%t'",
			test.Test,
			test.ExpectedParsed,
			param.Parsed(),
		)
	}
	if !param.Parsed() {
		return nil
	}
	if test.Value != nil {
		var a, b string
		a = reflect.ValueOf(test.Value).String()
		b = reflect.Indirect(reflect.ValueOf(param.Value())).String()
		if a != b {
			return fmt.Errorf(
				"%s failed: want value '%s', got '%s'",
				test.Test,
				a,
				b,
			)
		}
	}
	if param.Required() || param.Value() != nil {
		if param.RawValue() != test.ExpectedRawValue {
			return fmt.Errorf(
				"%s failed: want rawvalue '%s', got '%s'",
				test.Test,
				test.ExpectedRawValue,
				param.RawValue(),
			)
		}
	}
	return nil
}

func TestParameter(t *testing.T) {
	var test *ParameterTest
	for _, test = range ParameterTests {
		t.Run(test.Test, func(t *testing.T) {
			var err error
			if err = RunParameterTest(test); err != nil {
				t.Error(err)
			}
		})
	}
}

type ParameterBenchmark struct {
	*Parameter
	Arguments
	}

func BenchmarkParameter(b *testing.B) {
	var params = make([]ParameterBenchmark, 0, b.N)
	for i := 0; i < b.N; i++ {
		params = append(params, ParameterBenchmark{
			NewParameter("foo", "", false, false, nil), NewArguments("--foo"),
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params[i].Parse(params[i].Arguments)
	}
}

type ParameterDefinition struct {
	Name     string
	Short    string
	Required bool
	Indexed  bool
	Value    interface{}
}

type ExpectedParsed map[string]bool

type ParametersTestRun struct {
	Arguments
	ExpectedResult error
	ExpectedParsed
}

type ParametersTest struct {
	Test        string
	Definitions []ParameterDefinition
	Runs        []ParametersTestRun
}

var ParametersTests = []ParametersTest{
	{
		"Single named parameter",
		[]ParameterDefinition{
			{"foo", "", false, false, nil},
		},
		[]ParametersTestRun{
			{
				NewArguments("--foo"),
				nil,
				ExpectedParsed{
					"foo": true,
				},
			},
			{
				NewArguments("foo"),
				nil,
				ExpectedParsed{},
			},
			{
				NewArguments(),
				ErrNoArguments,
				ExpectedParsed{},
			},
			{
				NewArguments("---foo"),
				ErrInvalidArgument,
				ExpectedParsed{},
			},
			{
				NewArguments("--bar"),
				ErrParameterNotFound,
				ExpectedParsed{},
			},
			{
				NewArguments("-abc"),
				ErrParameterNotFound,
				ExpectedParsed{},
			},
		},
	},
	{
		"Multiple named parameters",
		[]ParameterDefinition{
			{"foo", "", false, false, nil},
			{"bar", "", false, false, nil},
			{"baz", "", false, false, nil},
		},
		[]ParametersTestRun{
			{
				NewArguments("--foo", "--bar", "--baz"),
				nil,
				ExpectedParsed{
					"foo": true,
					"bar": true,
					"baz": true,
				},
			},
			{
				NewArguments("--foo"),
				nil,
				ExpectedParsed{
					"foo": true,
					"bar": false,
					"baz": false,
				},
			},
		},
	},
}

func RunParametersTest(test *ParametersTest) error {
	var value interface{}
	var params = NewParameters(nil)
	var paramdef ParameterDefinition
	var err error
	for _, paramdef = range test.Definitions {
		if paramdef.Indexed {
			err = params.AddIndexed(paramdef.Name, "", paramdef.Required, value)
		} else {
			err = params.AddNamed(paramdef.Name, paramdef.Short, "", paramdef.Required, value)
		}
		if err != nil {
			return err
		}
	}
	var run ParametersTestRun
	for _, run = range test.Runs {
		params.Reset()
		if err = params.Parse(run.Arguments); err == nil {
			if run.ExpectedResult != nil {
				return errors.New("unexpected nil result")
			}
		}
		if !errors.Is(err, run.ExpectedResult) {
			return err
		}
		var name string
		var parsed bool
		var param *Parameter
		for name, parsed = range run.ExpectedParsed {
			if param = params.MustGet(name); param.Parsed() != parsed {
				return fmt.Errorf(
					"%s failed: want parsed '%t' for param '%s', got '%t'",
					test.Test,
					parsed,
					name,
					param.Parsed(),
				)
			}
		}

	}
	return nil
}

func TestParameters(t *testing.T) {
	var test ParametersTest
	for _, test = range ParametersTests {
		t.Run(test.Test, func(t *testing.T) {
			var err error
			if err = RunParametersTest(&test); err != nil {
				t.Error(err)
			}
		})
	}
}
