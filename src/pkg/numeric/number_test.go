package numeric

import (
	"reflect"
	"testing"
)

func testNumber[
	O interface {
		Number | []Number
	},
	W interface {
		Number | bool
	},
](t *testing.T, method string, n Number, other O, want W) {
	t.Run(method, func(t *testing.T) {
		callable := reflect.ValueOf(&n).Elem().MethodByName(method)
		if !callable.IsValid() {
			t.Fatalf("Number.%s() not found", method)
		}

		var args []reflect.Value
		switch reflect.ValueOf(other).Kind() {
		case reflect.Float64:
			args = append(args, reflect.ValueOf(other))

		case reflect.Array, reflect.Slice:
			for i := 0; i < reflect.ValueOf(other).Len(); i++ {
				args = append(args, reflect.ValueOf(other).Index(i))
			}

		}

		got := callable.Call(args)
		switch want := any(want).(type) {
		case Number:
			if len(got) == 0 && n != want {
				t.Errorf("Number.%s() = %v, want %v", method, n, want)
			} else if !Equal(got[0].Interface().(Number), want, 1e-9) {
				t.Errorf("Number.%s() = %v, want %v", method, got[0].Interface(), want)
			}

		case bool:
			if len(got) == 0 && n != 0 {
				t.Errorf("Number.%s() = %v, want %v", method, n, 0)
			} else if got[0].Interface().(bool) != want {
				t.Errorf("Number.%s() = %v, want %v", method, got[0].Interface(), want)
			}
		}
	})
}

func TestMaxMin(t *testing.T) {
	testNumber(t, "Clamp", 0.5, []Number{-1, 1}, Number(0.5))
	testNumber(t, "Clamp", 30, []Number{-1, 1}, Number(1))
	testNumber(t, "InRange", 2, []Number{-1, 1}, false)
	testNumber(t, "InRange", 0.5, []Number{-1, 1}, true)
	testNumber(t, "Max", 1, []Number{2, 3, 4, 5}, Number(5))
	testNumber(t, "Min", 1, []Number{2, 3, 4, 5}, Number(1))
	testNumber(t, "Pow", 4, Number(0.5), Number(2))
}
