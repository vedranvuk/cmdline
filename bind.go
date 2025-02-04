package cmdline

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

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

	LongKey  = "long"
	ShortKey = "short"
	HelpKey  = "help"
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
		var (
			long  = path + strutils.KebabCase(v.Type().Field(i).Name)
			short = tag.First(LongKey)
			help  = tag.First(HelpKey)
		)
		if tag.Exists(SkipKey) {
			continue
		}
		if tag.Exists(LongKey) {
			long = tag.First(LongKey)
		}
		switch v.Type().Field(i).Type.Kind() {
		case reflect.Bool:
			c.Globals.BooleanVar(long, short, help, v.Field(i).Addr().Interface())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.String:
			if tag.Exists(RequiredKey) {
				c.Globals.RequiredVar(long, short, help, v.Field(i).Addr().Interface())
			} else {
				c.Globals.OptionalVar(long, short, help, v.Field(i).Addr().Interface())
			}
		case reflect.Array, reflect.Slice:
			c.Globals.RepeatedVar(long, short, help, v.Field(i).Addr().Interface())
		case reflect.Struct:
			if err = bindStruct(v.Field(i), c, long); err != nil {
				return
			}
		}
	}
	return
}

// generateOptionShortNames generates short Option names.
//
// The algorithm is trivial; a single pass of generating shortname from
// lowercased longname, starting with first char and advancing to next if
// non-unique in set thus far, until unique or exhausted.
//
// Second pass makes sure all short options are unique in set, and if not, the
// latter option shortname is unset.
//
// If a unique letter was not generated from long name option gets no short
// option name.
func generateOptionShortNames(o Options) {
	// Sequentially go through options, setting shortcmd to lowercase first
	// letter from longname. Each time check if it is already used and advance
	// to next letter until unique or exhausted.
	for idx, option := range o {
		if option.ShortName != "" {
			continue
		}
		var name = strings.ToLower(option.LongName)
	GenShort:
		for _, r := range name {
			option.ShortName = string(r)
			if !strings.ContainsAny(option.ShortName, strutils.AlphaNums) {
				continue GenShort
			}
			for i := 0; i < idx; i++ {
				if o[i].ShortName == option.ShortName {
					continue GenShort
				}
			}
			break GenShort
		}
	}
	// Check all short names for duplicates and unset duplicates.
	for _, option := range o {
		var n = option.ShortName
		for _, other := range o {
			if option == other {
				continue
			}
			if n == other.ShortName {
				other.ShortName = ""
			}
		}
	}
}
