package numeric

import (
	"reflect"
	"testing"
)

func testPosition[
	O interface{ Number | Position | any },
	W interface{ Number | Position | bool },
](t *testing.T, method string, pos Position, other O, want W) {
	t.Run(method, func(t *testing.T) {
		callable := reflect.ValueOf(&pos).Elem().MethodByName(method)
		if !callable.IsValid() {
			t.Fatalf("Position.%s() not found", method)
		}

		var args []reflect.Value
		switch reflect.ValueOf(other).Kind() {
		case reflect.Struct, reflect.Float64:
			args = append(args, reflect.ValueOf(other))
		}

		got := callable.Call(args)[0].Interface().(W)
		switch want := any(want).(type) {
		case Number:
			if !Equal(any(got).(Number), want, 1e-9) {
				t.Errorf("Position.%s() = %v, want %v", method, got, want)
			}

		case Position:
			if !Equal(any(got).(Position), want, 1e-9) {
				t.Errorf("Position.%s() = %v, want %v", method, got, want)
			}

		case bool:
			if any(got).(bool) != want {
				t.Errorf("Position.%s() = %v, want %v", method, got, want)
			}

		}
	})
}

func TestPosition(t *testing.T) {
	testPosition(t, "Add", Position{X: 1, Y: 2}, Position{X: 3, Y: 4}, Position{X: 4, Y: 6})
	testPosition(t, "AddN", Position{X: 1, Y: 2}, Number(3), Position{X: 4, Y: 5})
	testPosition(t, "Average", Position{X: 1, Y: 2}, any(nil), Number(1.5))
	testPosition(t, "Distance", Position{X: 1, Y: 2}, Position{X: 4, Y: 6}, Number(5))
	testPosition(t, "Div", Position{X: 4, Y: 6}, Number(2), Position{X: 2, Y: 3})
	testPosition(t, "DivX", Position{X: 4, Y: 6}, Position{X: 2, Y: 3}, Position{X: 2, Y: 2})
	testPosition(t, "DivX", Position{X: 4, Y: 6}, Position{X: 0, Y: 3}, Position{X: 0, Y: 2})
	testPosition(t, "Equal", Position{X: 1, Y: 2}, Position{X: 1, Y: 2}, true)
	testPosition(t, "Greater", Position{X: 4, Y: 6}, Position{X: 1, Y: 2}, true)
	testPosition(t, "GreaterOrEqual", Position{X: 4, Y: 6}, Position{X: 4, Y: 6}, true)
	testPosition(t, "IsZero", Position{X: 0, Y: 0}, any(nil), true)
	testPosition(t, "IsZero", Position{X: 1, Y: 0}, any(nil), false)
	testPosition(t, "Less", Position{X: 1, Y: 2}, Position{X: 4, Y: 6}, true)
	testPosition(t, "LessOrEqual", Position{X: 1, Y: 2}, Position{X: 4, Y: 6}, true)
	testPosition(t, "Magnitude", Position{X: 3, Y: 4}, any(nil), Number(5))
	testPosition(t, "Mul", Position{X: 1, Y: 2}, Number(3), Position{X: 3, Y: 6})
	testPosition(t, "MulX", Position{X: 1, Y: 2}, Position{X: 3, Y: 4}, Position{X: 3, Y: 8})
	testPosition(t, "MulX", Position{X: 1, Y: 2}, Position{X: 0, Y: 4}, Position{X: 0, Y: 8})
	testPosition(t, "Normalize", Position{X: 3, Y: 4}, any(nil), Position{X: 0.6, Y: 0.8})
	testPosition(t, "Root", Position{X: 4, Y: 4}, any(nil), Number(4))
	testPosition(t, "Sub", Position{X: 4, Y: 6}, Position{X: 1, Y: 2}, Position{X: 3, Y: 4})
	testPosition(t, "SubN", Position{X: 4, Y: 6}, Number(2), Position{X: 2, Y: 4})
}
