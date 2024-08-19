package config

import (
	"reflect"
	"strconv"
	"strings"

	"encoding/json"
)

// envVariable represents an environment variable.
type EnvVariable[T any] string

// Get returns the value of the environment variable.
func (e EnvVariable[T]) Get() (result T) {
	parse := func(raw string) (result T) {
		if len(raw) == 0 {
			return
		}

		target := reflect.ValueOf(&result).Elem()

		switch target.Kind() {
		case reflect.Bool:
			v, _ := strconv.ParseBool(raw)
			target.Set(reflect.ValueOf(v))
			return target.Interface().(T)

		case reflect.Float32, reflect.Float64:
			v, _ := strconv.ParseFloat(raw, 64)
			target.Set(reflect.ValueOf(v).Convert(target.Type()))
			return target.Interface().(T)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v, _ := strconv.ParseInt(raw, 10, 64)
			target.Set(reflect.ValueOf(v).Convert(target.Type()))
			return target.Interface().(T)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v, _ := strconv.ParseUint(raw, 10, 64)
			target.Set(reflect.ValueOf(v).Convert(target.Type()))
			return target.Interface().(T)

		case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
			_ = json.Unmarshal([]byte(raw), target.Addr().Interface())
			return target.Interface().(T)

		case reflect.String:
			return any(raw).(T)

		default:
			return

		}
	}

	key, fallback, _ := strings.Cut(string(e), ":")
	raw := Getenv(key)
	if len(raw) == 0 && len(fallback) > 0 {
		return parse(fallback)
	}

	return parse(raw)
}
