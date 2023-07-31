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
		if c.GetParentCommand() != nil {
			t.Fatal("GetParentCommand failed")
		}
		if c.GetCommand() != a {
			t.Fatal("GetCommand failed")
		}
		return nil
	}

	b.Handler = func(c Context) error {
		if c.GetParentCommand() != a {
			t.Fatal("GetParentCommand failed")
		}
		if c.GetParentCommand() != b {
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
			&Boolean{
				LongName:    "boolean",
				ShortName:   "b",
				MappedValue: &boolean,
			},
			&Optional{
				LongName:    "optional",
				ShortName:   "o",
				MappedValue: &optional,
			},
			&Required{
				LongName:    "required",
				ShortName:   "r",
				MappedValue: &required,
			},
			&Repeated{
				LongName:    "repeated",
				ShortName:   "p",
				MappedValue: &repeated,
			},
			&Indexed{
				Name:        "indexed",
				MappedValue: &indexed,
			},
			&Variadic{
				Name:        "variadic",
				MappedValue: &variadic,
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
			&Boolean{
				LongName:    "boolean",
				ShortName:   "b",
				MappedValue: &boolean,
			},
			&Optional{
				LongName:    "optional",
				ShortName:   "o",
				MappedValue: &optional,
			},
			&Required{
				LongName:    "required",
				ShortName:   "r",
				MappedValue: &required,
			},
			&Repeated{
				LongName:    "repeated",
				ShortName:   "p",
				MappedValue: &repeated,
			},
			&Indexed{
				Name:        "indexed",
				MappedValue: &indexed,
			},
			&Variadic{
				Name:        "variadic",
				MappedValue: &variadic,
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
			&Optional{
				LongName:    "bool",
				MappedValue: &Bool,
			},
			&Optional{
				LongName:    "int",
				MappedValue: &Int,
			},
			&Optional{
				LongName:    "int8",
				MappedValue: &Int8,
			},
			&Optional{
				LongName:    "int16",
				MappedValue: &Int16,
			},
			&Optional{
				LongName:    "int32",
				MappedValue: &Int32,
			},
			&Optional{
				LongName:    "int64",
				MappedValue: &Int64,
			},
			&Optional{
				LongName:    "uint",
				MappedValue: &Uint,
			},
			&Optional{
				LongName:    "uint8",
				MappedValue: &Uint8,
			},
			&Optional{
				LongName:    "uint16",
				MappedValue: &Uint16,
			},
			&Optional{
				LongName:    "uint32",
				MappedValue: &Uint32,
			},
			&Optional{
				LongName:    "uint64",
				MappedValue: &Uint64,
			},
			&Optional{
				LongName:    "float32",
				MappedValue: &Float32,
			},
			&Optional{
				LongName:    "float64",
				MappedValue: &Float64,
			},
			&Optional{
				LongName:    "string",
				MappedValue: &String,
			},
			&Optional{
				LongName:    "duration",
				MappedValue: &Duration,
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

func (self *Custom) Set(v RawValues) error {
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
			&Required{
				LongName:    "custom",
				MappedValue: v,
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
			&Boolean{
				LongName:  "boolean",
				ShortName: "b",
				Help:      "A Boolean Option.",
			},
			&Optional{
				LongName:  "optional",
				ShortName: "o",
				Help:      "An Optional Option.",
			},
			&Required{
				LongName:  "required",
				ShortName: "r",
				Help:      "A Required Option.",
			},
			&Indexed{
				Name: "indexed",
				Help: "An Indexed Option.",
			},
			&Repeated{
				LongName:  "repeated",
				ShortName: "p",
				Help:      "A Repeated Option.",
			},
			&Variadic{
				Name: "variadic",
				Help: "A Variadic Option.",
			},
		},
		Commands: Commands{
			{
				Name: "one",
				Help: "Command One.",
				Options: Options{
					&Boolean{
						LongName:  "boolean",
						ShortName: "b",
						Help:      "A Boolean Option.",
					},
					&Optional{
						LongName:  "optional",
						ShortName: "o",
						Help:      "An Optional Option.",
					},
					&Required{
						LongName:  "required",
						ShortName: "r",
						Help:      "A Required Option.",
					},
					&Indexed{
						Name: "indexed",
						Help: "An Indexed Option.",
					},
					&Repeated{
						LongName:  "repeated",
						ShortName: "p",
						Help:      "A Repeated Option.",
					},
					&Variadic{
						Name: "variadic",
						Help: "A Variadic Option.",
					},
				},
				SubCommands: Commands{
					{
						Name: "one",
						Help: "Command One.",
						Options: Options{
							&Boolean{
								LongName:  "boolean",
								ShortName: "b",
								Help:      "A Boolean Option.",
							},
							&Optional{
								LongName:  "optional",
								ShortName: "o",
								Help:      "An Optional Option.",
							},
							&Required{
								LongName:  "required",
								ShortName: "r",
								Help:      "A Required Option.",
							},
							&Indexed{
								Name: "indexed",
								Help: "An Indexed Option.",
							},
							&Repeated{
								LongName:  "repeated",
								ShortName: "p",
								Help:      "A Repeated Option.",
							},
							&Variadic{
								Name: "variadic",
								Help: "A Variadic Option.",
							},
						},
					},
					{
						Name: "two",
						Help: "Command Two.",
						Options: Options{
							&Boolean{
								LongName:  "boolean",
								ShortName: "b",
								Help:      "A Boolean Option.",
							},
							&Optional{
								LongName:  "optional",
								ShortName: "o",
								Help:      "An Optional Option.",
							},
							&Required{
								LongName:  "required",
								ShortName: "r",
								Help:      "A Required Option.",
							},
							&Indexed{
								Name: "indexed",
								Help: "An Indexed Option.",
							},
							&Repeated{
								LongName:  "repeated",
								ShortName: "p",
								Help:      "A Repeated Option.",
							},
							&Variadic{
								Name: "variadic",
								Help: "A Variadic Option.",
							},
						},
					},
				},
			},
			{
				Name: "two",
				Help: "Command Two.",
				Options: Options{
					&Boolean{
						LongName:  "boolean",
						ShortName: "b",
						Help:      "A Boolean Option.",
					},
					&Optional{
						LongName:  "optional",
						ShortName: "o",
						Help:      "An Optional Option.",
					},
					&Required{
						LongName:  "required",
						ShortName: "r",
						Help:      "A Required Option.",
					},
					&Indexed{
						Name: "indexed",
						Help: "An Indexed Option.",
					},
					&Repeated{
						LongName:  "repeated",
						ShortName: "p",
						Help:      "A Repeated Option.",
					},
					&Variadic{
						Name: "variadic",
						Help: "A Variadic Option.",
					},
				},
			},
			{
				Name: "three",
				Help: "Command Three.",
				Options: Options{
					&Boolean{
						LongName:  "boolean",
						ShortName: "b",
						Help:      "A Boolean Option.",
					},
					&Optional{
						LongName:  "optional",
						ShortName: "o",
						Help:      "An Optional Option.",
					},
					&Required{
						LongName:  "required",
						ShortName: "r",
						Help:      "A Required Option.",
					},
					&Indexed{
						Name: "indexed",
						Help: "An Indexed Option.",
					},
					&Repeated{
						LongName:  "repeated",
						ShortName: "p",
						Help:      "A Repeated Option.",
					},
					&Variadic{
						Name: "variadic",
						Help: "A Variadic Option.",
					},
				},
			},
		},
	}
	PrintConfig(os.Stdout, config)
}
