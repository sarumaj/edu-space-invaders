package config

import (
	"reflect"
	"strconv"
)

// envVariable represents an environment variable.
type envVariable[T any] string

func (e envVariable[T]) Get() (result T) {
	raw := Getenv(string(e))
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

	case reflect.String:
		return any(raw).(T)

	default:
		return

	}
}
