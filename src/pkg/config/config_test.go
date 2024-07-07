package config

import (
	"reflect"
	"strings"
	"testing"
)

func TestConfigLoaded(t *testing.T) {
	v := reflect.ValueOf(&Config).Elem()

	var checkFields func(tb testing.TB, v reflect.Value, parents ...reflect.StructField)
	checkFields = func(tb testing.TB, v reflect.Value, parents ...reflect.StructField) {
		if v.Kind() != reflect.Struct {
			t.Fatal("Config is not a struct")
		}

		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).IsZero() {
				var str strings.Builder
				_, _ = str.WriteString("Config not loaded: ")

				for _, parent := range parents {
					_, _ = str.WriteString(parent.Name)
					_, _ = str.WriteString(".")
				}

				_, _ = str.WriteString(v.Type().Field(i).Name)
				tb.Error(str.String())

				continue
			}

			if v.Field(i).Kind() == reflect.Struct {
				checkFields(tb, v.Field(i), append(parents, v.Type().Field(i))...)
			}

		}
	}

	checkFields(t, v)
}
