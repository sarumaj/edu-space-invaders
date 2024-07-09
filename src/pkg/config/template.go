package config

import (
	"bytes"
	"fmt"
	"text/template"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// printer is a message printer for English.
var printer = message.NewPrinter(language.English)

type templateString string

// Template represents a message template.
type Template struct {
	Damage int
	Name   string
	Level  int
}

// Render renders the template with the given string.
func (t Template) Execute(str templateString) string {
	out := bytes.NewBuffer(nil)
	parsed, err := template.New(fmt.Sprintf("%p", out)).Funcs(template.FuncMap{
		"color": func(color string, args ...any) string {
			return fmt.Sprintf(`<span style="color: %s;">%s</span>`, color, fmt.Sprint(args...))
		},
		"bold": func(args ...any) string {
			return fmt.Sprintf(`<b>%s</b>`, fmt.Sprint(args...))
		},
		"italic": func(args ...any) string {
			return fmt.Sprintf(`<i>%s</i>`, fmt.Sprint(args...))
		},
		"printf": printer.Sprintf,
		"strike": func(args ...any) string {
			return fmt.Sprintf(`<s>%s</s>`, fmt.Sprint(args...))
		},
		"underline": func(args ...any) string {
			return fmt.Sprintf(`<u>%s</u>`, fmt.Sprint(args...))
		},
	}).Parse(string(str))
	LogError(err)
	LogError(parsed.Execute(out, t))

	return out.String()
}

// Execute returns a template string with the given arguments.
func Execute(str templateString) string { return Template{}.Execute(str) }

// Sprintf returns a template string with the given arguments.
func Sprintf(in string, args ...any) templateString { return templateString(fmt.Sprintf(in, args...)) }
