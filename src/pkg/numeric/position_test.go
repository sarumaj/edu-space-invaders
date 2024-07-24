package numeric

import (
	"reflect"
	"testing"
)

func testPosition[
	O interface {
		Number | []Position | Position | any
	},
	W interface {
		[]Number | Number | Position | bool
	},
](t *testing.T, method string, pos Position, other O, want W) {
	t.Run(method, func(t *testing.T) {
		callable := reflect.ValueOf(&pos).Elem().MethodByName(method)
		if !callable.IsValid() {
			t.Fatalf("Position.%s() not found", method)
		}

		var args []reflect.Value
		switch reflect.ValueOf(other).Kind() {
		case reflect.Struct, reflect.Float64, reflect.Slice:
			args = append(args, reflect.ValueOf(other))

		case reflect.Array:
			array := reflect.ValueOf(other)
			slice := reflect.MakeSlice(reflect.SliceOf(array.Type().Elem()), array.Len(), array.Len())
			_ = reflect.Copy(slice, array)
			args = append(args, slice)

		}

		got := callable.Call(args)
		if len(got) == 0 {
			t.Fatalf("Position.%s() returned no value", method)
		}

		switch want := any(want).(type) {
		case Number:
			if !Equal(got[0].Interface().(Number), want, 1e-9) {
				t.Errorf("Position.%s() = %v, want %v", method, got[0].Interface(), want)
			}

		case []Number:
			if len(got) != len(want) {
				t.Fatalf("Position.%s() returned %d values, want %d", method, len(got), len(want))
			}

			for i, w := range want {
				if !Equal(got[i].Interface().(Number), w, 1e-9) {
					t.Errorf("Position.%s() = %v, want %v", method, got[i].Interface(), w)
				}
			}

		case Position:
			if !Equal(got[0].Interface().(Position), want, 1e-9) {
				t.Errorf("Position.%s() = %v, want %v", method, got[0].Interface(), want)
			}

		case bool:
			if got[0].Interface().(bool) != want {
				t.Errorf("Position.%s() = %v, want %v", method, got[0].Interface(), want)
			}

		}
	})
}

func TestPosition(t *testing.T) {
	testPosition(t, "Add", Position{X: 1, Y: 2}, Position{X: 3, Y: 4}, Position{X: 4, Y: 6})
	testPosition(t, "AddN", Position{X: 1, Y: 2}, Number(3), Position{X: 4, Y: 5})
	testPosition(t, "Angle", Position{X: 1, Y: 2}, any(nil), Number(1.1071487177940904))
	testPosition(t, "AngleTo", Position{X: 1, Y: 2}, Position{X: 3, Y: 4}, Number(0.7853981633974483))
	testPosition(t, "Average", Position{X: 1, Y: 2}, any(nil), Number(1.5))
	testPosition(t, "Cross", Position{X: 1, Y: 2}, Position{X: 3, Y: 4}, Number(-2))
	testPosition(t, "Distance", Position{X: 1, Y: 2}, Position{X: 4, Y: 6}, Number(5))
	testPosition(t, "Div", Position{X: 4, Y: 6}, Number(2), Position{X: 2, Y: 3})
	testPosition(t, "DivX", Position{X: 4, Y: 6}, Position{X: 2, Y: 3}, Position{X: 2, Y: 2})
	testPosition(t, "DivX", Position{X: 4, Y: 6}, Position{X: 0, Y: 3}, Position{X: 0, Y: 2})
	testPosition(t, "Dot", Position{X: 1, Y: 2}, Position{X: 3, Y: 4}, Number(11))
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
	testPosition(t, "Perpendicular", Position{X: 3, Y: 4}, any(nil), Position{X: -4, Y: 3})
	testPosition(t, "Project", Position{X: 1, Y: 2}, [3]Position{{X: 3, Y: 4}, {2, 4}, {-7, 8}}, []Number{9, 11})
	testPosition(t, "Root", Position{X: 4, Y: 4}, any(nil), Number(4))
	testPosition(t, "Sub", Position{X: 4, Y: 6}, Position{X: 1, Y: 2}, Position{X: 3, Y: 4})
	testPosition(t, "SubN", Position{X: 4, Y: 6}, Number(2), Position{X: 2, Y: 4})
}
