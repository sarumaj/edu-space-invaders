package config

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// printer is a message printer for English.
var printer = message.NewPrinter(language.English)

var funcsMap = template.FuncMap{
	"color": func(color string, args ...any) string {
		return fmt.Sprintf(`<span style="color: %s;">%s</span>`, color, fmt.Sprint(args...))
	},
	"bold": func(args ...any) string {
		return fmt.Sprintf(`<b>%s</b>`, fmt.Sprint(args...))
	},
	"greet": func() string {
		now := time.Now()
		switch {
		case now.Hour() >= 6 && now.Hour() < 12:
			return "Good morning"
		case now.Hour() >= 12 && now.Hour() < 18:
			return "Good afternoon"
		case now.Hour() >= 18 && now.Hour() < 24:
			return "Good evening"
		default:
			return "Good night"
		}
	},
	"italic": func(args ...any) string {
		return fmt.Sprintf(`<i>%s</i>`, fmt.Sprint(args...))
	},
	"inc": func(n int) int {
		return n + 1
	},
	"printf": printer.Sprintf,
	"strike": func(args ...any) string {
		return fmt.Sprintf(`<s>%s</s>`, fmt.Sprint(args...))
	},
	"underline": func(args ...any) string {
		return fmt.Sprintf(`<u>%s</u>`, fmt.Sprint(args...))
	},
}

type templateString string

// Template represents a message template.
type Template map[string]any

// execute renders the template with the given string.
func (t Template) execute(str templateString) string {
	out := bytes.NewBuffer(nil)
	parsed, err := template.New(fmt.Sprintf("%p", out)).Funcs(funcsMap).Parse(string(str))
	LogError(err)
	LogError(parsed.Execute(out, t))

	return out.String()
}

// Execute returns a template string with the given arguments.
func Execute[String interface{ string | templateString }](str String, t ...Template) string {
	if len(t) > 0 {
		return t[0].execute(templateString(str))
	}

	return Template{}.execute(templateString(str))
}

// Sprintf returns a template string with the given arguments.
func Sprintf[String interface{ string | templateString }](in String, args ...any) String {
	return String(fmt.Sprintf(string(in), args...))
}
