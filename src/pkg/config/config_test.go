package config

import (
	"io"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

func TestConfigLoaded(t *testing.T) {
	v := reflect.ValueOf(&Config).Elem()

	var checkFields func(tb testing.TB, v reflect.Value, parents ...reflect.StructField)
	checkFields = func(tb testing.TB, v reflect.Value, parents ...reflect.StructField) {
		tb.Helper()

		if v.Kind() != reflect.Struct {
			t.Fatal("Config is not a struct")
		}

		reportFieldNok := func(fl reflect.StructField, err error) string {
			var str strings.Builder
			_, _ = str.WriteString("Config not loaded: ")

			for _, parent := range parents {
				_, _ = str.WriteString(parent.Name)
				_, _ = str.WriteString(".")
			}

			_, _ = str.WriteString(fl.Name)

			if err != nil {
				_, _ = str.WriteString(": ")
				_, _ = str.WriteString(err.Error())
			} else {
				_, _ = str.WriteString(": is zero")
			}

			return str.String()
		}

		for i := 0; i < v.NumField(); i++ {
			// Check if the field is zero.
			if v.Field(i).IsZero() {
				tb.Error(reportFieldNok(v.Type().Field(i), nil))
				continue
			}

			// Check if the field is a template string and parse it.
			if v.Field(i).Type() == reflect.TypeOf(TemplateString("")) {
				parsed, err := template.New(v.Field(i).Type().Name()).Funcs(funcsMap).Parse(v.Field(i).String())
				if err != nil {
					tb.Error(reportFieldNok(v.Type().Field(i), err))
					continue
				}

				if err := parsed.Execute(io.Discard, map[string]any{}); err != nil {
					tb.Error(reportFieldNok(v.Type().Field(i), err))
					continue
				}
			}

			// Check if the field is a struct and recursively check its fields.
			if v.Field(i).Kind() == reflect.Struct {
				checkFields(tb, v.Field(i), append(parents, v.Type().Field(i))...)
			}

		}
	}

	checkFields(t, v)
}
