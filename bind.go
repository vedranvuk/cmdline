package cmdline

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/vedranvuk/strutils"
)

// CmdlineTag is the struct field tag key recognized by cmdline package.
const CmdlineTag = "cmdline"

// TagKey is a key read from a cmdline tag read from struct fields.
type TagKey string

const (
	// SkipKey defines that the marked field should be skipped when binding.
	SkipKey = "skip"
	// RequiredKey defines that the option should be defined as required.
	RequiredKey = "required"
)

// Bind binds a cmdline config to a target struct.
//
// For each field in the target struct a global option is defined and bound to
// the source field value. Fields of embedded structs are recursively processed
// and named by a path. Only fields of type supported by [Option] are supported.
func Bind(config *Config, target any) (out *Config, err error) {
	var v = reflect.ValueOf(target)
	if !v.IsValid() || v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return config, errors.New("invalid target, must be a pointer to a struct")
	}
	return config, bindStruct(v.Elem(), config, "")
}

func bindStruct(v reflect.Value, c *Config, path string) (err error) {
	var tag = strutils.Tag{
		TagKey:    CmdlineTag,
		Separator: ";",
	}
	if path != "" {
		path += "."
	}
	for i := 0; i < v.NumField(); i++ {
		tag.Clear()
		if err = tag.Parse(string(v.Type().Field(i).Tag)); err != nil {
			if err != strutils.ErrTagNotFound {
				return fmt.Errorf("parse cmdline tag: %w", err)
			}
			err = nil
		}
		var name = path + strutils.KebabCase(v.Type().Field(i).Name)
		if tag.Exists(SkipKey) {
			continue
		}
		switch v.Type().Field(i).Type.Kind() {
		case reflect.Bool:
			c.Globals.BooleanVar(name, "", "", v.Field(i).Addr().Interface())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.String:
			if tag.Exists(RequiredKey) {
				c.Globals.RequiredVar(name, "", "", v.Field(i).Addr().Interface())
			} else {
				c.Globals.OptionalVar(name, "", "", v.Field(i).Addr().Interface())
			}
		case reflect.Array, reflect.Slice:
			c.Globals.RepeatedVar(name, "", "", v.Field(i).Addr().Interface())
		case reflect.Struct:
			if err = bindStruct(v.Field(i), c, name); err != nil {
				return
			}
		}
	}
	return
}
