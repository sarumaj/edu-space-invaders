package graphics

import "testing"

func TestColorByName(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   string
		out  Color
	}{
		{name: "Black", in: "black", out: Color{0, 0, 0, 1}},
		{name: "White", in: "white", out: Color{255, 255, 255, 1}},
		{name: "Red", in: "red", out: Color{255, 0, 0, 1}},
		{name: "Green", in: "green", out: Color{0, 128, 0, 1}},
		{name: "Blue", in: "blue", out: Color{0, 0, 255, 1}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseColor(tt.in)
			if err != nil {
				t.Errorf("ParseColor() error = %v", err)
				return
			}
			if got != tt.out {
				t.Errorf("ParseColor() = %v, want %v", got, tt.out)
			}
		})
	}
}

func TestParseColor(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   string
		out  Color
	}{
		{name: "Hex color", in: "#ff0000", out: Color{255, 0, 0, 1}},
		{name: "RGB color", in: "rgb(255, 0, 0)", out: Color{255, 0, 0, 1}},
		{name: "RGBA color", in: "rgba(255, 0, 0, 1)", out: Color{255, 0, 0, 1}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseColor(tt.in)
			if err != nil {
				t.Errorf("ParseColor() error = %v", err)
				return
			}
			if got != tt.out {
				t.Errorf("ParseColor() = %v, want %v", got, tt.out)
			}
		})
	}
}
