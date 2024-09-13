package cmdline

import (
	"errors"
	"reflect"
)

// DefaultTagName is the default name of a struct tag read by cmdline package.
const DefaultTagName = "cmdline"

// TagKey is a known cmdline key read from the cmdline tag in a struct field
// being bound to.
type TagKey = string

const (
	// OptionalTag denotes that a field should use the Optional option.
	OptionalTag TagKey = "optional"
	// RequiredTag denotes that a field should use the Required option.
	RequiredTag TagKey = "required"
)

// Bind binds a struct to a new command that binds options to a struct.
//
// It generates an option for each field in input struct recursively.
// Nested struct fields are addressed with a standard dot hierarchy syntax,
// e.g. Parent.Child.
//
// name and help define command name and help.
// v must be a pointer to a struct which will be written from options at
// parse time.
//
// Certain types are bound to options in the following way:
// bool: Boolean
//
// Tags can specify which kind of option is used for a field.
func Bind(config *Config, v any, name, help string) (*Command, error) {

	var s = reflect.ValueOf(v)
	if s.Kind() != reflect.Ptr {
		return nil, errors.New("v must be a pointer to a struct")
	}
	s = reflect.Indirect(s)
	if s.Kind() != reflect.Struct {
		return nil, errors.New("v must be a pointer to a struct")
	}

	for i := 0; i < s.NumField(); i++ {
		var tag, exists = s.Type().Field(i).Tag.Lookup(config.TagName)
		_ = tag
		_ = exists
	}

	return nil, nil
}
