// Copyright 2023-2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cmdline

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestContext(t *testing.T) {

	var (
		a = &Command{Name: "a"}
		b = &Command{Name: "b"}
	)

	a.Handler = func(c Context) error {
		if c.ParentCommand() != nil {
			t.Fatal("GetParentCommand failed")
		}
		if c.Command() != a {
			t.Fatal("GetCommand failed")
		}
		return nil
	}

	b.Handler = func(c Context) error {
		if c.ParentCommand() != a {
			t.Fatal("GetParentCommand failed")
		}
		if c.ParentCommand() != b {
			t.Fatal("GetCommand failed")
		}
		return nil
	}

}

func TestRequireSubCommandExecution(t *testing.T) {
	if err := Parse(&Config{
		Args: []string{"one"},
		Commands: Commands{
			{
				Name:                "one",
				RequireSubExecution: true,
				Handler:             NopHandler,
				SubCommands: Commands{
					{
						Name:    "two",
						Handler: NopHandler,
					},
				},
			},
		},
	}); err == nil {
		t.Fatal()
	}
}

func TestLastInChain(t *testing.T) {
	var e = errors.New("e")

	if err := Parse(&Config{
		Args: []string{"one", "two", "three"},
		Commands: Commands{
			{
				Name:    "one",
				Handler: NopHandler,
				SubCommands: Commands{
					{
						Name:    "two",
						Handler: NopHandler,
						SubCommands: Commands{
							{
								Name: "three",
								Handler: func(c Context) error {
									return e
								},
							},
						},
					},
				},
			},
		},
	}); err != e {
		t.Fatal()
	}
}

func TestMidInChain(t *testing.T) {
	var e = errors.New("e")

	if err := Parse(&Config{
		Args: []string{"one", "two"},
		Commands: Commands{
			{
				Name:    "one",
				Handler: NopHandler,
				SubCommands: Commands{
					{
						Name: "two",
						Handler: func(c Context) error {
							return e
						},
						SubCommands: Commands{
							{
								Name:    "three",
								Handler: NopHandler,
							},
						},
					},
				},
			},
		},
	}); err != e {
		t.Fatal()
	}
}

func TestOptionsParsingAssign(t *testing.T) {

	var (
		boolean  = false
		optional = ""
		required = 0
		repeated = []string{}
		indexed  = ""
		variadic = []string{}
	)

	if err := Parse(&Config{
		Args:          []string{"-b", "--optional=\"opt\"", "-r=42", "--repeated=1", "-p=2", "idxd", "one", "two", "three"},
		IndexedFirst:  false,
		UseAssignment: true,
		Globals: Options{
			&Option{
				LongName:  "boolean",
				ShortName: "b",
				Var:       &boolean,
				Kind:      Boolean,
			},
			&Option{
				LongName:  "optional",
				ShortName: "o",
				Var:       &optional,
				Kind:      Optional,
			},
			&Option{
				LongName:  "required",
				ShortName: "r",
				Var:       &required,
				Kind:      Required,
			},
			&Option{
				LongName:  "repeated",
				ShortName: "p",
				Var:       &repeated,
				Kind:      Repeated,
			},
			&Option{
				LongName: "indexed",
				Var:      &indexed,
				Kind:     Indexed,
			},
			&Option{
				LongName: "variadic",
				Var:      &variadic,
				Kind:     Variadic,
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if boolean != true {
		t.Fatal("boolean")
	}
	if optional != "opt" {
		t.Fatal("optional")
	}
	if required != 42 {
		t.Fatal("required")
	}
	if strings.Join(repeated, " ") != "1 2" {
		t.Fatal("repeated")
	}
	if indexed != "idxd" {
		t.Fatal("indexed")
	}
	if strings.Join(variadic, " ") != "one two three" {
		t.Fatal("variadic")
	}
}

func TestOptionsParsingNoAssign(t *testing.T) {

	var (
		boolean  = false
		optional = ""
		required = 0
		repeated = []string{}
		indexed  = ""
		variadic = []string{}
	)

	if err := Parse(&Config{
		Args:         []string{"-b", "--optional", "opt", "-r", "42", "--repeated", "1", "-p", "2", "idxd", "one", "two", "three"},
		IndexedFirst: false,
		Globals: Options{
			&Option{
				LongName:  "boolean",
				ShortName: "b",
				Var:       &boolean,
				Kind:      Boolean,
			},
			&Option{
				LongName:  "optional",
				ShortName: "o",
				Var:       &optional,
				Kind:      Optional,
			},
			&Option{
				LongName:  "required",
				ShortName: "r",
				Var:       &required,
				Kind:      Required,
			},
			&Option{
				LongName:  "repeated",
				ShortName: "p",
				Var:       &repeated,
				Kind:      Repeated,
			},
			&Option{
				LongName: "indexed",
				Var:      &indexed,
				Kind:     Indexed,
			},
			&Option{
				LongName: "variadic",
				Var:      &variadic,
				Kind:     Variadic,
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if boolean != true {
		t.Fatal("boolean")
	}
	if optional != "opt" {
		t.Fatal("optional")
	}
	if required != 42 {
		t.Fatal("required")
	}
	if strings.Join(repeated, " ") != "1 2" {
		t.Fatal("repeated")
	}
	if indexed != "idxd" {
		t.Fatal("indexed")
	}
	if strings.Join(variadic, " ") != "one two three" {
		t.Fatal("variadic")
	}
}

func TestMappedValues(t *testing.T) {

	var (
		Bool     bool          = false
		Int      int           = 0
		Int8     int8          = 0
		Int16    int16         = 0
		Int32    int32         = 0
		Int64    int64         = 0
		Uint     uint          = 0
		Uint8    uint8         = 0
		Uint16   uint16        = 0
		Uint32   uint32        = 0
		Uint64   uint64        = 0
		Float32  float32       = 0.0
		Float64  float64       = 0.0
		String   string        = ""
		Duration time.Duration = 0
	)

	if err := Parse(&Config{
		UseAssignment: true,
		Args: []string{
			"--bool=true",
			"--int=2",
			"--int8=4",
			"--int16=8",
			"--int32=16",
			"--int64=32",
			"--uint=64",
			"--uint8=128",
			"--uint16=256",
			"--uint32=512",
			"--uint64=1024",
			"--float32=3.14",
			"--float64=1.16",
			"--string=string",
			"--duration=60s",
		},
		Globals: Options{
			&Option{
				LongName: "bool",
				Var:      &Bool,
				Kind:     Optional,
			},
			&Option{
				LongName: "int",
				Var:      &Int,
				Kind:     Optional,
			},
			&Option{
				LongName: "int8",
				Var:      &Int8,
				Kind:     Optional,
			},
			&Option{
				LongName: "int16",
				Var:      &Int16,
				Kind:     Optional,
			},
			&Option{
				LongName: "int32",
				Var:      &Int32,
				Kind:     Optional,
			},
			&Option{
				LongName: "int64",
				Var:      &Int64,
				Kind:     Optional,
			},
			&Option{
				LongName: "uint",
				Var:      &Uint,
				Kind:     Optional,
			},
			&Option{
				LongName: "uint8",
				Var:      &Uint8,
				Kind:     Optional,
			},
			&Option{
				LongName: "uint16",
				Var:      &Uint16,
				Kind:     Optional,
			},
			&Option{
				LongName: "uint32",
				Var:      &Uint32,
				Kind:     Optional,
			},
			&Option{
				LongName: "uint64",
				Var:      &Uint64,
				Kind:     Optional,
			},
			&Option{
				LongName: "float32",
				Var:      &Float32,
				Kind:     Optional,
			},
			&Option{
				LongName: "float64",
				Var:      &Float64,
				Kind:     Optional,
			},
			&Option{
				LongName: "string",
				Var:      &String,
				Kind:     Optional,
			},
			&Option{
				LongName: "duration",
				Var:      &Duration,
				Kind:     Optional,
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if !Bool {
		t.Fatal("bool")
	}
	if Int != 2 {
		t.Fatal("int")
	}
	if Int8 != 4 {
		t.Fatal("int8")
	}
	if Int16 != 8 {
		t.Fatal("int16")
	}
	if Int32 != 16 {
		t.Fatal("int32")
	}
	if Int64 != 32 {
		t.Fatal("int64")
	}
	if Uint != 64 {
		t.Fatal("uint")
	}
	if Uint8 != 128 {
		t.Fatal("uint8")
	}
	if Uint16 != 256 {
		t.Fatal("uint16")
	}
	if Uint32 != 512 {
		t.Fatal("uint32")
	}
	if Uint64 != 1024 {
		t.Fatal("uint64")
	}
	if Float32 != 3.14 {
		t.Fatal("float32")
	}
	if Float64 != 1.16 {
		t.Fatal("float64")
	}
	if String != "string" {
		t.Fatal("string")
	}
	if Duration != 60*time.Second {
		t.Fatal("duration")
	}
}

type Custom struct{}

func (self Custom) String() string {
	return "42"
}

func (self *Custom) Set(v Values) error {
	if v.First() != "42" {
		return errors.New("fail")
	}
	return nil
}

func TestCustomMappedType(t *testing.T) {

	var v = new(Custom)

	if err := Parse(&Config{
		UseAssignment: true,
		Args:          []string{"--custom=42"},
		Globals: Options{
			&Option{
				LongName: "custom",
				Var:      v,
				Kind:     Required,
			},
		},
	}); err != nil {
		t.Fatal()
	}
}

func TestHelpCommand(t *testing.T) {
	var config = getPrettyPrintDemoConfig()
	config.Commands.Register(HelpCommand(nil))
	config.Args = []string{"help"}
	if err := config.Parse(nil); err != nil {
		t.Fatal(err)
	}
}

func TestParser2(t *testing.T) {
	var config = Default()
	config.UseAssignment = true
	config.Args = []string{
		"generate",
		"-p=penis",
		"penis",
	}

	config.Commands.Register(HelpCommand(nil))
	config.Commands.Handle(
		"generate",
		"Generates commandline classes.",
		func(c Context) error {
			return nil
		},
	).Options.
		Boolean("help-from-tag", "g", "Include help from tag.").
		Boolean("help-from-doc", "d", "Include help from doc comments.").
		Boolean("error-on-unsupported-field", "e", "Throws an error if unsupporrted field was encountered.").
		Boolean("print", "r", "Print output.").
		Boolean("no-wrote", "n", "Do not write output file.").
		Optional("output-file", "o", "Output file name.").
		Optional("tag-key", "t", "Name of the tag key to parse.").
		Required("package-name", "p", "Name of the package output go file belongs to.").
		Variadic("packages", "Packages to parse.")

	if err := config.Parse(nil); err != nil {
		t.Fatal(err)
	}
}

func TestCombinedBooleans(t *testing.T) {
	var (
		config  = Default("-abc")
		a, b, c bool
	)
	config.Globals.
		BooleanVar("A", "a", "", &a).
		BooleanVar("B", "b", "", &b).
		BooleanVar("C", "c", "", &c)
	if err := config.Parse(nil); err != nil {
		t.Fatal(err)
	}
	if !a || !b || !c {
		t.Fatal("Combined booleans failed")
	}
}

func getPrettyPrintDemoConfig() (out *Config) {

	out = Default()
	out.Globals.Boolean("verbose", "v", "Be verbose.")

	var cmd *Command
	cmd = out.Commands.Handle("one", "Command one.", NopHandler)
	cmd.Options.
		Boolean("boolean", "b", "A Boolean option.").
		Optional("optional", "o", "An Optional Option.").
		Required("required", "r", "A Required Option.").
		Repeated("repeated", "p", "A Repeated Option.").
		Indexed("indexed", "An Indexed Option.")

	cmd.SubCommands.Handle("subone", "Subcommand One.", NopHandler).Options.
		Boolean("boolean", "b", "A Boolean option.").
		Optional("optional", "o", "An Optional Option.").
		Required("required", "r", "A Required Option.").
		Repeated("repeated", "p", "A Repeated Option.").
		Indexed("indexed", "An Indexed Option.").
		Variadic("variadic", "A Variadic Option")

	cmd.SubCommands.Handle("subtwo", "Subcommand Two.", NopHandler).Options.
		Boolean("boolean", "b", "A Boolean option.").
		Optional("optional", "o", "An Optional Option.").
		Required("required", "r", "A Required Option.").
		Repeated("repeated", "p", "A Repeated Option.").
		Indexed("indexed", "An Indexed Option.").
		Variadic("variadic", "A Variadic Option")

	out.Commands.Handle("two", "Command two.", NopHandler).Options.
		Boolean("boolean", "b", "A Boolean option.").
		Optional("optional", "o", "An Optional Option.").
		Required("required", "r", "A Required Option.").
		Repeated("repeated", "p", "A Repeated Option.").
		Indexed("indexed", "An Indexed Option.").
		Variadic("variadic", "A Variadic Option")

	out.Commands.Handle("three", "Command three.", NopHandler).Options.
		Boolean("boolean", "b", "A Boolean option.").
		Optional("optional", "o", "An Optional Option.").
		Required("required", "r", "A Required Option.").
		Repeated("repeated", "p", "A Repeated Option.").
		Indexed("indexed", "An Indexed Option.").
		Variadic("variadic", "A Variadic Option")

	return
}

type BindData struct {
	Name   string
	Age    int
	SkipMe bool `cmdline:"skip"`
	Sub    BindSub
}

type BindSub struct {
	Nickname string
}

func TestBind(t *testing.T) {
	var data = BindData{
		Name: "foo",
		Age:  69,
		Sub:  BindSub{Nickname: "baz"},
	}
	var config, err = Bind(Default(), &data)
	if err != nil {
		t.Fatal(err)

	}
	config.UseAssignment = true
	config.Args = []string{
		"--name=bar",
		"--age=42",
		"--sub.nickname=bat",
	}
	if testing.Verbose() {
		PrintConfig(os.Stdout, config)
	}

	if err = config.Parse(nil); err != nil {
		t.Fatal(err)
	}
}
