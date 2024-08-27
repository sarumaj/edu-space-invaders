package numeric

import (
	"reflect"
	"testing"
)

func testNumber[
	O interface {
		Number | []Number
	},
](t *testing.T, method string, n Number, other O, want Number) {
	t.Run(method, func(t *testing.T) {
		callable := reflect.ValueOf(&n).Elem().MethodByName(method)
		if !callable.IsValid() {
			t.Fatalf("Number.%s() not found", method)
		}

		var args []reflect.Value
		switch reflect.ValueOf(other).Kind() {
		case reflect.Struct, reflect.Float64:
			args = append(args, reflect.ValueOf(other))

		case reflect.Array, reflect.Slice:
			for i := 0; i < reflect.ValueOf(other).Len(); i++ {
				args = append(args, reflect.ValueOf(other).Index(i))
			}

		}

		got := callable.Call(args)
		if len(got) == 0 && n != want {
			t.Errorf("Number.%s() = %v, want %v", method, n, want)
		} else if !Equal(got[0].Interface().(Number), want, 1e-9) {
			t.Errorf("Number.%s() = %v, want %v", method, got[0].Interface(), want)
		}
	})
}

func TestMaxMin(t *testing.T) {
	testNumber(t, "Max", 1, []Number{2, 3, 4, 5}, 5)
	testNumber(t, "Min", 1, []Number{2, 3, 4, 5}, 1)
}
