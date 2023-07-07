package cmdline

import (
	"strconv"
	"time"
)

// MapBool maps value pointing to a bool to an Option under key.
//
// If the Option gets parsed the value will be set to true.
func (self Options) MapBool(key string, value *bool) Options {
	return self.mapOption(key, value)
}

// MapString maps value pointing to a string to an Option under key.
//
// If the Option gets parsed the value will be set to the parsed Option's value.
func (self Options) MapString(key string, value *string) Options {
	return self.mapOption(key, value)
}

// MapInt maps value pointing to an int to an Option under key.
//
// If the Option gets parsed the value will be set to the parsed Option's value.
func (self Options) MapInt(key string, value *int) Options {
	return self.mapOption(key, value)
}

// MapInt64 maps value pointing to an int64 to an Option under key.
//
// If the Option gets parsed the value will be set to the parsed Option's value.
func (self Options) MapInt64(key string, value *int64) Options {
	return self.mapOption(key, value)
}

// MapUint maps value pointing to a uint to an Option under key.
//
// If the Option gets parsed the value will be set to the parsed Option's value.
func (self Options) MapUint(key string, value *uint) Options {
	return self.mapOption(key, value)
}

// MapUint64 maps value pointing to a uint64 to an Option under key.
//
// If the Option gets parsed the value will be set to the parsed Option's value.
func (self Options) MapUint64(key string, value *uint64) Options {
	return self.mapOption(key, value)
}

// MapFloat64 maps value pointing to a float64 to an Option under key.
//
// If the Option gets parsed the value will be set to the parsed Option's value.
func (self Options) MapFloat64(key string, value *float64) Options {
	return self.mapOption(key, value)
}

// MapDuration maps value pointing to a time.Duration to an Option under key.
//
// If the Option gets parsed the value will be set to the parsed Option's value.
func (self Options) MapDuration(key string, value *time.Duration) Options {
	return self.mapOption(key, value)
}

// Value is any type that implements this interface.
type Value interface {
	String() string
	Set(string) error
}

// MapValue maps value pointing to a type that implements Value to an Option
// under key.
//
// If the Option gets parsed the value will be set to the parsed Option's value.
func (self Options) MapValue(key string, value Value) Options { 
	return self.mapOption(key, value) 
}

// mapOption maps a value to an Option under key.
func (self Options) mapOption(key string, value any) Options {

	if value == nil {
		panic("nil pointer given as value for key " + key)
	}

	var opt Option = self.Get(key)
	if opt == nil {
		panic("no option under key " + key)
	}

	if _, valisbool := value.(*bool); valisbool {
		if _, optisbool := opt.(*Boolean); !optisbool {
			panic("boolean options can map only to *bool")
		}
	}
	for i := 0; i < len(self); i++ {
		if self[i].Key() == key {
			switch p := self[i].(type) {
			case *Boolean:
				p.option.value = value
			case *Optional:
				p.option.value = value
			case *Required:
				p.option.value = value
			case *Indexed:
				p.option.value = value
			case *Variadic:
				p.option.value = value
			}
			break
		}
	}

	return self
}

// rawToMapped converts opt.raw to a value mapped to that option, if exists.
// Returns nil if no value mapped and on success. Returns a non nil error on
// failed conversion only.
func (self Options) rawToMapped(opt Option) (err error) {

	var (
		value = opt.Value()
		raw   = opt.Raw()
	)
	if value == nil || raw == "" {
		return nil
	}

	switch p := value.(type) {
	case *bool:
		*p = true
	case *int:
		var v int64
		if v, err = strconv.ParseInt(raw, 10, 0); err == nil {
			*p = int(v)
		}
	case *uint:
		var v uint64
		if v, err = strconv.ParseUint(raw, 10, 0); err == nil {
			*p = uint(v)
		}
	case *int64:
		*p, err = strconv.ParseInt(raw, 10, 64)
	case *uint64:
		*p, err = strconv.ParseUint(raw, 10, 64)
	case *float64:
		*p, err = strconv.ParseFloat(raw, 64)
	case *time.Duration:
		*p, err = time.ParseDuration(raw)
	default:
		if v, ok := p.(Value); ok {
			err = v.Set(raw)
		}
	}

	return
}
