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

// Template represents a message template.
type Template struct {
	Damage int
	Name   string
	Level  int
}

type templateString string

// Render renders the template with the given string.
func (t Template) Execute(str templateString) string {
	out := bytes.NewBuffer(nil)
	parsed, err := template.New(fmt.Sprintf("%p", out)).Funcs(template.FuncMap{
		"printf": printer.Sprintf,
	}).Parse(string(str))
	LogError(err)
	LogError(parsed.Execute(out, t))

	return out.String()
}
