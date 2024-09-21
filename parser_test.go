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
		Arguments: []string{"one"},
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
		Arguments: []string{"one", "two", "three"},
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
		Arguments: []string{"one", "two"},
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

func TestOptionsParsing(t *testing.T) {

	var (
		boolean  = false
		optional = ""
		required = 0
		repeated = []string{}
		indexed  = ""
		variadic = []string{}
	)

	if err := Parse(&Config{
		Arguments:      []string{"-b", "--optional=\"opt\"", "-r=42", "--repeated=1", "-p=2", "idxd", "one", "two", "three"},
		NoIndexedFirst: true,
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
		Arguments:      []string{"-b", "--optional", "opt", "-r", "42", "--repeated", "1", "-p", "2", "idxd", "one", "two", "three"},
		NoIndexedFirst: true,
		NoAssignment:   true,
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
		Arguments: []string{
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
		Arguments: []string{"--custom=42"},
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

func TestPrint(t *testing.T) {
	var config = &Config{
		LongPrefix:  DefaultLongPrefix,
		ShortPrefix: DefaultShortPrefix,
		Globals: Options{
			&Option{
				LongName:  "boolean",
				ShortName: "b",
				Help:      "A Boolean Option.",
				Kind:      Boolean,
			},
			&Option{
				LongName:  "optional",
				ShortName: "o",
				Help:      "An Optional Option.",
				Kind:      Optional,
			},
			&Option{
				LongName:  "required",
				ShortName: "r",
				Help:      "A Required Option.",
				Kind:      Required,
			},
			&Option{
				LongName: "indexed",
				Help:     "An Indexed Option.",
				Kind:     Indexed,
			},
			&Option{
				LongName:  "repeated",
				ShortName: "p",
				Help:      "A Repeated Option.",
				Kind:      Repeated,
			},
			&Option{
				LongName: "variadic",
				Help:     "A Variadic Option.",
				Kind:     Variadic,
			},
		},
		Commands: Commands{
			{
				Name: "one",
				Help: "Command One.",
				Options: Options{
					&Option{
						LongName:  "boolean",
						ShortName: "b",
						Help:      "A Boolean Option.",
						Kind:      Boolean,
					},
					&Option{
						LongName:  "optional",
						ShortName: "o",
						Help:      "An Optional Option.",
						Kind:      Optional,
					},
					&Option{
						LongName:  "required",
						ShortName: "r",
						Help:      "A Required Option.",
						Kind:      Required,
					},
					&Option{
						LongName: "indexed",
						Help:     "An Indexed Option.",
						Kind:     Indexed,
					},
					&Option{
						LongName:  "repeated",
						ShortName: "p",
						Help:      "A Repeated Option.",
						Kind:      Repeated,
					},
					&Option{
						LongName: "variadic",
						Help:     "A Variadic Option.",
						Kind:     Variadic,
					},
				},
				SubCommands: Commands{
					{
						Name: "one",
						Help: "Command One.",
						Options: Options{
							&Option{
								LongName:  "boolean",
								ShortName: "b",
								Help:      "A Boolean Option.",
								Kind:      Boolean,
							},
							&Option{
								LongName:  "optional",
								ShortName: "o",
								Help:      "An Optional Option.",
								Kind:      Optional,
							},
							&Option{
								LongName:  "required",
								ShortName: "r",
								Help:      "A Required Option.",
								Kind:      Required,
							},
							&Option{
								LongName: "indexed",
								Help:     "An Indexed Option.",
								Kind:     Indexed,
							},
							&Option{
								LongName:  "repeated",
								ShortName: "p",
								Help:      "A Repeated Option.",
								Kind:      Repeated,
							},
							&Option{
								LongName: "variadic",
								Help:     "A Variadic Option.",
								Kind:     Variadic,
							},
						},
					},
					{
						Name: "two",
						Help: "Command Two.",
						Options: Options{
							&Option{
								LongName:  "boolean",
								ShortName: "b",
								Help:      "A Boolean Option.",
								Kind:      Boolean,
							},
							&Option{
								LongName:  "optional",
								ShortName: "o",
								Help:      "An Optional Option.",
								Kind:      Optional,
							},
							&Option{
								LongName:  "required",
								ShortName: "r",
								Help:      "A Required Option.",
								Kind:      Required,
							},
							&Option{
								LongName: "indexed",
								Help:     "An Indexed Option.",
								Kind:     Indexed,
							},
							&Option{
								LongName:  "repeated",
								ShortName: "p",
								Help:      "A Repeated Option.",
								Kind:      Repeated,
							},
							&Option{
								LongName: "variadic",
								Help:     "A Variadic Option.",
								Kind:     Variadic,
							},
						},
					},
				},
			},
			{
				Name: "two",
				Help: "Command Two.",
				Options: Options{
					&Option{
						LongName:  "boolean",
						ShortName: "b",
						Help:      "A Boolean Option.",
						Kind:      Boolean,
					},
					&Option{
						LongName:  "optional",
						ShortName: "o",
						Help:      "An Optional Option.",
						Kind:      Optional,
					},
					&Option{
						LongName:  "required",
						ShortName: "r",
						Help:      "A Required Option.",
						Kind:      Required,
					},
					&Option{
						LongName: "indexed",
						Help:     "An Indexed Option.",
						Kind:     Indexed,
					},
					&Option{
						LongName:  "repeated",
						ShortName: "p",
						Help:      "A Repeated Option.",
						Kind:      Repeated,
					},
					&Option{
						LongName: "variadic",
						Help:     "A Variadic Option.",
						Kind:     Variadic,
					},
				},
			},
			{
				Name: "three",
				Help: "Command Three.",
				Options: Options{
					&Option{
						LongName:  "boolean",
						ShortName: "b",
						Help:      "A Boolean Option.",
						Kind:      Boolean,
					},
					&Option{
						LongName:  "optional",
						ShortName: "o",
						Help:      "An Optional Option.",
						Kind:      Optional,
					},
					&Option{
						LongName:  "required",
						ShortName: "r",
						Help:      "A Required Option.",
						Kind:      Required,
					},
					&Option{
						LongName: "indexed",
						Help:     "An Indexed Option.",
						Kind:     Indexed,
					},
					&Option{
						LongName:  "repeated",
						ShortName: "p",
						Help:      "A Repeated Option.",
						Kind:      Repeated,
					},
					&Option{
						LongName: "variadic",
						Help:     "A Variadic Option.",
						Kind:     Variadic,
					},
				},
			},
		},
	}
	PrintConfig(os.Stdout, config)
}
