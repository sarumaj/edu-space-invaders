package graphics

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/sarumaj/edu-space-invaders/src/pkg/numeric"
)

var catalogue = namedColor{}

var (
	namedColorRegex = regexp.MustCompile(`^([a-zA-Z]+)$`)                                       // Named color
	hexColorRegex   = regexp.MustCompile(`^#([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})$`) // Hex color
	rgbColorRegex   = regexp.MustCompile(`^rgb\((\d+),\s*(\d+),\s*(\d+)\)$`)                    // RGB color
	rgbaColorRegex  = regexp.MustCompile(`^rgba\((\d+),\s*(\d+),\s*(\d+),\s*(\d+(\.\d+)?)\)$`)  // RGBA color
)

// Color represents a color with red, green, blue, and alpha components.
type Color [4]numeric.Number

// Equal returns true if the color is equal to the other color.
func (c Color) Equal(other Color) bool {
	return numeric.Equal(c[0], other[0], 1e-9) &&
		numeric.Equal(c[1], other[1], 1e-9) &&
		numeric.Equal(c[2], other[2], 1e-9) &&
		numeric.Equal(c[3], other[3], 1e-9)
}

// FormatHex returns the string representation of the color in the format "#RRGGBBAA" or "#RRGGBB".
func (c Color) FormatHex() string {
	if numeric.Equal(c[3], 1, 1e-9) {
		return fmt.Sprintf("#%02x%02x%02x", c[0].Int(), c[1].Int(), c[2].Int())
	}
	return fmt.Sprintf("#%02x%02x%02x%02x", c[0].Int(), c[1].Int(), c[2].Int(), (c[3] * 255).Int())
}

// FormatRGB returns the string representation of the color in the format "rgb(R, G, B)".
func (c Color) FormatRGB() string {
	return fmt.Sprintf("rgb(%d, %d, %d)", c[0].Int(), c[1].Int(), c[2].Int())
}

// FormatRGBA returns the string representation of the color in the format "rgba(R, G, B, A)".
func (c Color) FormatRGBA() string {
	return fmt.Sprintf("rgba(%d, %d, %d, %f)", c[0].Int(), c[1].Int(), c[2].Int(), c[3])
}

// SetAt sets the color value at the given index.
func (c Color) SetAt(index int, n numeric.Number) Color {
	switch index {
	case 0, 1, 2:
		if n < 1 {
			n *= 255
		}
		c[index] = n.Clamp(0, 255)

	case 3:
		if n > 1 {
			n /= 255
		}
		c[index] = n.Clamp(0, 1)

	}

	return c
}

func (c Color) R() numeric.Number { return c[0] }
func (c Color) G() numeric.Number { return c[1] }
func (c Color) B() numeric.Number { return c[2] }
func (c Color) A() numeric.Number { return c[3] }

func (c Color) SetR(r numeric.Number) Color { return c.SetAt(0, r) }
func (c Color) SetG(g numeric.Number) Color { return c.SetAt(1, g) }
func (c Color) SetB(b numeric.Number) Color { return c.SetAt(2, b) }
func (c Color) SetA(a numeric.Number) Color { return c.SetAt(3, a) }

// NamedColor is a color type that can be retrieved by name.
type namedColor struct{}

// ByName returns the named color by name.
func (n namedColor) ByName(name string) (out Color, err error) {
	v := reflect.ValueOf(&n).Elem()
	methodFuncType := reflect.TypeOf(func(namedColor) Color { return Color{} })
	if method, ok := v.Type().MethodByName(name); ok && method.Func.Type() == methodFuncType {
		return v.MethodByName(name).Call(nil)[0].Interface().(Color), nil
	}

	for i := 0; i < v.Type().NumMethod(); i++ {
		if method := v.Type().Method(i); method.Func.Type() == methodFuncType && strings.EqualFold(method.Name, name) {
			return v.Method(i).Call(nil)[0].Interface().(Color), nil
		}
	}

	return out, fmt.Errorf("unknown named color: %s", name)
}

func (namedColor) AliceBlue() Color            { return Color{240, 248, 255, 1} }
func (namedColor) AntiqueWhite() Color         { return Color{250, 235, 215, 1} }
func (namedColor) Aqua() Color                 { return Color{0, 255, 255, 1} }
func (namedColor) Aquamarine() Color           { return Color{127, 255, 212, 1} }
func (namedColor) Azure() Color                { return Color{240, 255, 255, 1} }
func (namedColor) Beige() Color                { return Color{245, 245, 220, 1} }
func (namedColor) Bisque() Color               { return Color{255, 228, 196, 1} }
func (namedColor) Black() Color                { return Color{0, 0, 0, 1} }
func (namedColor) BlancheDalmond() Color       { return Color{255, 235, 205, 1} }
func (namedColor) Blue() Color                 { return Color{0, 0, 255, 1} }
func (namedColor) BlueViolet() Color           { return Color{138, 43, 226, 1} }
func (namedColor) Brown() Color                { return Color{165, 42, 42, 1} }
func (namedColor) BurlyWood() Color            { return Color{222, 184, 135, 1} }
func (namedColor) CadetBlue() Color            { return Color{95, 158, 160, 1} }
func (namedColor) Chartreuse() Color           { return Color{127, 255, 0, 1} }
func (namedColor) Chocolate() Color            { return Color{210, 105, 30, 1} }
func (namedColor) Coral() Color                { return Color{255, 127, 80, 1} }
func (namedColor) CornflowerBlue() Color       { return Color{100, 149, 237, 1} }
func (namedColor) CornSilk() Color             { return Color{255, 248, 220, 1} }
func (namedColor) Crimson() Color              { return Color{220, 20, 60, 1} }
func (namedColor) Cyan() Color                 { return Color{0, 255, 255, 1} }
func (namedColor) DarkBlue() Color             { return Color{0, 0, 139, 1} }
func (namedColor) DarkCyan() Color             { return Color{0, 139, 139, 1} }
func (namedColor) DarkGoldenRod() Color        { return Color{184, 134, 11, 1} }
func (namedColor) DarkGray() Color             { return Color{169, 169, 169, 1} }
func (namedColor) DarkGreen() Color            { return Color{0, 100, 0, 1} }
func (namedColor) DarkGrey() Color             { return Color{169, 169, 169, 1} }
func (namedColor) DarkKhaki() Color            { return Color{189, 183, 107, 1} }
func (namedColor) DarkMagenta() Color          { return Color{139, 0, 139, 1} }
func (namedColor) DarkOliveGreen() Color       { return Color{85, 107, 47, 1} }
func (namedColor) DarkOrange() Color           { return Color{255, 140, 0, 1} }
func (namedColor) DarkOrchid() Color           { return Color{153, 50, 204, 1} }
func (namedColor) DarkRed() Color              { return Color{139, 0, 0, 1} }
func (namedColor) DarkSalmon() Color           { return Color{233, 150, 122, 1} }
func (namedColor) DarkSeaGreen() Color         { return Color{143, 188, 143, 1} }
func (namedColor) DarkSlateBlue() Color        { return Color{72, 61, 139, 1} }
func (namedColor) DarkSlateGray() Color        { return Color{47, 79, 79, 1} }
func (namedColor) DarkSlateGrey() Color        { return Color{47, 79, 79, 1} }
func (namedColor) DarkTurquoise() Color        { return Color{0, 206, 209, 1} }
func (namedColor) DarkViolet() Color           { return Color{148, 0, 211, 1} }
func (namedColor) DeepPink() Color             { return Color{255, 20, 147, 1} }
func (namedColor) DeepSkyBlue() Color          { return Color{0, 191, 255, 1} }
func (namedColor) DimGray() Color              { return Color{105, 105, 105, 1} }
func (namedColor) DimGrey() Color              { return Color{105, 105, 105, 1} }
func (namedColor) DodgerBlue() Color           { return Color{30, 144, 255, 1} }
func (namedColor) Firebrick() Color            { return Color{178, 34, 34, 1} }
func (namedColor) FloralWhite() Color          { return Color{255, 250, 240, 1} }
func (namedColor) ForestGreen() Color          { return Color{34, 139, 34, 1} }
func (namedColor) Fuchsia() Color              { return Color{255, 0, 255, 1} }
func (namedColor) Gainsboro() Color            { return Color{220, 220, 220, 1} }
func (namedColor) GhostWhite() Color           { return Color{248, 248, 255, 1} }
func (namedColor) GoldenRod() Color            { return Color{218, 165, 32, 1} }
func (namedColor) Gold() Color                 { return Color{255, 215, 0, 1} }
func (namedColor) Gray() Color                 { return Color{128, 128, 128, 1} }
func (namedColor) Green() Color                { return Color{0, 128, 0, 1} }
func (namedColor) GreenYellow() Color          { return Color{173, 255, 47, 1} }
func (namedColor) Grey() Color                 { return Color{128, 128, 128, 1} }
func (namedColor) Honeydew() Color             { return Color{240, 255, 240, 1} }
func (namedColor) HotPink() Color              { return Color{255, 105, 180, 1} }
func (namedColor) IndianRed() Color            { return Color{205, 92, 92, 1} }
func (namedColor) Indigo() Color               { return Color{75, 0, 130, 1} }
func (namedColor) Ivory() Color                { return Color{255, 255, 240, 1} }
func (namedColor) Khaki() Color                { return Color{240, 230, 140, 1} }
func (namedColor) LavenderBlush() Color        { return Color{255, 240, 245, 1} }
func (namedColor) Lavender() Color             { return Color{230, 230, 250, 1} }
func (namedColor) LawnGreen() Color            { return Color{124, 252, 0, 1} }
func (namedColor) LemonChiffon() Color         { return Color{255, 250, 205, 1} }
func (namedColor) LightBlue() Color            { return Color{173, 216, 230, 1} }
func (namedColor) LightCoral() Color           { return Color{240, 128, 128, 1} }
func (namedColor) LightCyan() Color            { return Color{224, 255, 255, 1} }
func (namedColor) LightGoldenRodYellow() Color { return Color{250, 250, 210, 1} }
func (namedColor) LightGray() Color            { return Color{211, 211, 211, 1} }
func (namedColor) LightGreen() Color           { return Color{144, 238, 144, 1} }
func (namedColor) LightGrey() Color            { return Color{211, 211, 211, 1} }
func (namedColor) LightPink() Color            { return Color{255, 182, 193, 1} }
func (namedColor) LightSalmon() Color          { return Color{255, 160, 122, 1} }
func (namedColor) LightSeaGreen() Color        { return Color{32, 178, 170, 1} }
func (namedColor) LightSkyBlue() Color         { return Color{135, 206, 250, 1} }
func (namedColor) LightSlateGray() Color       { return Color{119, 136, 153, 1} }
func (namedColor) LightSlateGrey() Color       { return Color{119, 136, 153, 1} }
func (namedColor) LightSteelBlue() Color       { return Color{176, 196, 222, 1} }
func (namedColor) LightYellow() Color          { return Color{255, 255, 224, 1} }
func (namedColor) Lime() Color                 { return Color{0, 255, 0, 1} }
func (namedColor) LimeGreen() Color            { return Color{50, 205, 50, 1} }
func (namedColor) Linen() Color                { return Color{250, 240, 230, 1} }
func (namedColor) Magenta() Color              { return Color{255, 0, 255, 1} }
func (namedColor) Maroon() Color               { return Color{128, 0, 0, 1} }
func (namedColor) MediumAquaMarine() Color     { return Color{102, 205, 170, 1} }
func (namedColor) MediumBlue() Color           { return Color{0, 0, 205, 1} }
func (namedColor) MediumOrchid() Color         { return Color{186, 85, 211, 1} }
func (namedColor) MediumPurple() Color         { return Color{147, 112, 219, 1} }
func (namedColor) MediumSeaGreen() Color       { return Color{60, 179, 113, 1} }
func (namedColor) MediumSlateBlue() Color      { return Color{123, 104, 238, 1} }
func (namedColor) MediumSpringGreen() Color    { return Color{0, 250, 154, 1} }
func (namedColor) MediumTurquoise() Color      { return Color{72, 209, 204, 1} }
func (namedColor) MediumVioletRed() Color      { return Color{199, 21, 133, 1} }
func (namedColor) MidnightBlue() Color         { return Color{25, 25, 112, 1} }
func (namedColor) MintCream() Color            { return Color{245, 255, 250, 1} }
func (namedColor) MistyRose() Color            { return Color{255, 228, 225, 1} }
func (namedColor) Moccasin() Color             { return Color{255, 228, 181, 1} }
func (namedColor) NavajoWhite() Color          { return Color{255, 222, 173, 1} }
func (namedColor) Navy() Color                 { return Color{0, 0, 128, 1} }
func (namedColor) OldLace() Color              { return Color{253, 245, 230, 1} }
func (namedColor) Olive() Color                { return Color{128, 128, 0, 1} }
func (namedColor) OliveDrab() Color            { return Color{107, 142, 35, 1} }
func (namedColor) Orange() Color               { return Color{255, 165, 0, 1} }
func (namedColor) OrangeRed() Color            { return Color{255, 69, 0, 1} }
func (namedColor) Orchid() Color               { return Color{218, 112, 214, 1} }
func (namedColor) PaleGoldenRod() Color        { return Color{238, 232, 170, 1} }
func (namedColor) PaleGreen() Color            { return Color{152, 251, 152, 1} }
func (namedColor) PaleTurquoise() Color        { return Color{175, 238, 238, 1} }
func (namedColor) PaleVioletRed() Color        { return Color{219, 112, 147, 1} }
func (namedColor) PapayaWhip() Color           { return Color{255, 239, 213, 1} }
func (namedColor) PeachPuff() Color            { return Color{255, 218, 185, 1} }
func (namedColor) Peru() Color                 { return Color{205, 133, 63, 1} }
func (namedColor) Pink() Color                 { return Color{255, 192, 203, 1} }
func (namedColor) Plum() Color                 { return Color{221, 160, 221, 1} }
func (namedColor) PowderBlue() Color           { return Color{176, 224, 230, 1} }
func (namedColor) Purple() Color               { return Color{128, 0, 128, 1} }
func (namedColor) RebeccaPurple() Color        { return Color{102, 51, 153, 1} }
func (namedColor) Red() Color                  { return Color{255, 0, 0, 1} }
func (namedColor) RosyBrown() Color            { return Color{188, 143, 143, 1} }
func (namedColor) RoyalBlue() Color            { return Color{65, 105, 225, 1} }
func (namedColor) SaddleBrown() Color          { return Color{139, 69, 19, 1} }
func (namedColor) Salmon() Color               { return Color{250, 128, 114, 1} }
func (namedColor) SandyBrown() Color           { return Color{244, 164, 96, 1} }
func (namedColor) SeaGreen() Color             { return Color{46, 139, 87, 1} }
func (namedColor) SeaShell() Color             { return Color{255, 245, 238, 1} }
func (namedColor) Sienna() Color               { return Color{160, 82, 45, 1} }
func (namedColor) Silver() Color               { return Color{192, 192, 192, 1} }
func (namedColor) SkyBlue() Color              { return Color{135, 206, 235, 1} }
func (namedColor) SlateBlue() Color            { return Color{106, 90, 205, 1} }
func (namedColor) SlateGray() Color            { return Color{112, 128, 144, 1} }
func (namedColor) SlateGrey() Color            { return Color{112, 128, 144, 1} }
func (namedColor) Snow() Color                 { return Color{255, 250, 250, 1} }
func (namedColor) SpringGreen() Color          { return Color{0, 255, 127, 1} }
func (namedColor) SteelBlue() Color            { return Color{70, 130, 180, 1} }
func (namedColor) Tan() Color                  { return Color{210, 180, 140, 1} }
func (namedColor) Teal() Color                 { return Color{0, 128, 128, 1} }
func (namedColor) Thistle() Color              { return Color{216, 191, 216, 1} }
func (namedColor) Tomato() Color               { return Color{255, 99, 71, 1} }
func (namedColor) Turquoise() Color            { return Color{64, 224, 208, 1} }
func (namedColor) Violet() Color               { return Color{238, 130, 238, 1} }
func (namedColor) Wheat() Color                { return Color{245, 222, 179, 1} }
func (namedColor) White() Color                { return Color{255, 255, 255, 1} }
func (namedColor) WhiteSmoke() Color           { return Color{245, 245, 245, 1} }
func (namedColor) Yellow() Color               { return Color{255, 255, 0, 1} }
func (namedColor) YellowGreen() Color          { return Color{154, 205, 50, 1} }

// Catalogue returns the named color catalogue.
func Catalogue() namedColor { return catalogue }

// ParseColor parses a color string in the format "#RRGGBB", "rgb(R, G, B)", "rgba(R, G, B, A)",
// or a named HTML color like "Indigo".
func ParseColor(in string) (out Color, err error) {
	if match := hexColorRegex.FindStringSubmatch(in); match != nil {
		// Helper function to convert hex to decimal
		hexToDec := func(hexStr string) (int, error) {
			dec, err := hex.DecodeString(hexStr)
			if err != nil || len(dec) == 0 {
				return 0, fmt.Errorf("invalid hex color: %s", hexStr)
			}
			return int(dec[0]), nil
		}

		r, err := hexToDec(match[1])
		if err != nil {
			return out, err
		}
		g, err := hexToDec(match[2])
		if err != nil {
			return out, err
		}
		b, err := hexToDec(match[3])
		if err != nil {
			return out, err
		}

		return Color{
			numeric.Number(r),
			numeric.Number(g),
			numeric.Number(b),
			1,
		}, nil
	}

	if match := rgbColorRegex.FindStringSubmatch(in); match != nil {
		r, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return out, err
		}
		g, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			return out, err
		}
		b, err := strconv.ParseFloat(match[3], 64)
		if err != nil {
			return out, err
		}

		return Color{
			numeric.Number(r),
			numeric.Number(g),
			numeric.Number(b),
			1,
		}, nil
	}

	if match := rgbaColorRegex.FindStringSubmatch(in); match != nil {
		r, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return out, err
		}
		g, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			return out, err
		}
		b, err := strconv.ParseFloat(match[3], 64)
		if err != nil {
			return out, err
		}
		a, err := strconv.ParseFloat(match[4], 64)
		if err != nil {
			return out, err
		}

		return Color{
			numeric.Number(r),
			numeric.Number(g),
			numeric.Number(b),
			numeric.Number(a),
		}, nil
	}

	if match := namedColorRegex.FindStringSubmatch(in); match != nil {
		return catalogue.ByName(match[1])
	}

	return out, fmt.Errorf("invalid color format: %s", in)
}
