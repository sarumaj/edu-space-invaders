package config

import (
	"fmt"
	"reflect"
	"testing"
)

func testEnvVariable[E any](t *testing.T, key, value string, want E) {
	t.Run(fmt.Sprintf("%s_%s", key, value), func(t *testing.T) {
		Setenv(key, value)
		defer Unsetenv(key)

		if got := envVariable[E](key).Get(); !reflect.DeepEqual(got, want) {
			t.Errorf("envVariable[%[1]T].Get() failed, got: %[2]v, want: %[1]v", got, want)
		}
	})
}

func TestEnvVariable(t *testing.T) {
	testEnvVariable(t, "TEST_BOOL", "true", true)
	testEnvVariable(t, "TEST_BOOL", "false", false)
	testEnvVariable(t, "TEST_FLOAT32", "3.14", float32(3.14))
	testEnvVariable(t, "TEST_FLOAT64", "3.14", 3.14)
	testEnvVariable(t, "TEST_INT", "42", 42)
	testEnvVariable(t, "TEST_INT8", "42", int8(42))
	testEnvVariable(t, "TEST_INT16", "42", int16(42))
	testEnvVariable(t, "TEST_INT32", "42", int32(42))
	testEnvVariable(t, "TEST_INT64", "42", int64(42))
	testEnvVariable(t, "TEST_UINT", "42", uint(42))
	testEnvVariable(t, "TEST_UINT8", "42", uint8(42))
	testEnvVariable(t, "TEST_UINT16", "42", uint16(42))
	testEnvVariable(t, "TEST_UINT32", "42", uint32(42))
	testEnvVariable(t, "TEST_UINT64", "42", uint64(42))
	testEnvVariable(t, "TEST_STRING", "hello", "hello")
}
