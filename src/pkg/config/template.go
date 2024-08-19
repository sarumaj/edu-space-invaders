package config

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// printer is a message printer for English.
var printer = message.NewPrinter(language.English)

// funcsMap contains the template functions.
var funcsMap = template.FuncMap{
	"bold": func(args ...any) string { return fmt.Sprintf(`<b>%s</b>`, fmt.Sprint(args...)) },
	"char": func(n any) string {
		switch num := n.(type) {
		case string:
			return string(map[string]rune{
				"hash":      '#',
				"semicolon": ';',
			}[num])
		case int:
			return string(rune(num))
		default:
			return ""
		}
	},
	"color": func(color string, args ...any) string {
		return fmt.Sprintf(`<span style="color: %s;">%s</span>`, color, fmt.Sprint(args...))
	},
	"config": func() config { return Config },
	"float":  convert[float64],
	"greet": func() string {
		switch now := time.Now(); {
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
	"inc": func(n any) int {
		num, ok := n.(int)
		if !ok {
			return 0
		}
		return num + 1
	},
	"int":           convert[int],
	"isTouchDevice": IsTouchDevice,
	"italic":        func(args ...any) string { return fmt.Sprintf(`<i>%s</i>`, fmt.Sprint(args...)) },
	"print":         printer.Sprint,
	"printf":        printer.Sprintf,
	"strike":        func(args ...any) string { return fmt.Sprintf(`<s>%s</s>`, fmt.Sprint(args...)) },
	"timestamp":     func() string { return fmt.Sprintf("[%s]", time.Now().Format("15:04:05.000")) },
	"underline":     func(args ...any) string { return fmt.Sprintf(`<u>%s</u>`, fmt.Sprint(args...)) },
}

// convert converts the given value to the type T.
// If the conversion fails, it returns the zero value of T.
func convert[T any](n any) T {
	r, ok := n.(T)
	if !ok {
		LogError(fmt.Errorf("cannot convert %T to %T", n, r))
	}

	return r
}

// TemplateString represents a template string.
// It is used to render templates.
type TemplateString string

// Sanitize removes all newlines, replaces multiple spaces with a single space,
// adds a space after periods followed by uppercase letters and after commas followed by alphanumeric characters,
// while preserving the content within {{ ... }} blocks.
func (t TemplateString) Sanitize() TemplateString {
	// Regular expression to match {{ ... }} blocks
	blockPattern := regexp.MustCompile(`\{\{.*?\}\}`)

	// Split input into segments by the {{ ... }} blocks
	segments := blockPattern.Split(string(t), -1)
	matches := blockPattern.FindAllString(string(t), -1)

	// Regular expressions to replace newlines, multiple spaces, periods, and commas
	replacements := []struct {
		regex  *regexp.Regexp
		repl   string
		global bool
	}{
		{regexp.MustCompile(`\r?\n`), "", true},               // Replace newlines with a space.
		{regexp.MustCompile(`\s+`), " ", true},                // Replace multiple spaces with a single space.
		{regexp.MustCompile(`\.([A-Z])`), ". $1", false},      // Add a space after a period.
		{regexp.MustCompile(`,([a-zA-Z0-9]+)`), ", $1", true}, // Add a space after a comma.
	}

	// Process each segment that is outside of {{ ... }} blocks
	var result strings.Builder
	for i := range segments {
		for _, rep := range replacements {
			if rep.global {
				continue
			}
			segments[i] = rep.regex.ReplaceAllString(segments[i], rep.repl)
		}
		_, _ = result.WriteString(segments[i])

		// If there was a match (i.e., a {{ ... }} block), append it back to the result
		if i < len(matches) {
			_, _ = result.WriteString(matches[i])
		}
	}

	out := strings.TrimSpace(result.String())
	for _, rep := range replacements {
		if !rep.global {
			continue
		}
		out = rep.regex.ReplaceAllString(out, rep.repl)
	}

	// Return the sanitized string, trimmed of any leading or trailing whitespace
	return TemplateString(out)
}

// Template represents a message template.
type Template map[string]any

// execute renders the template with the given string.
func (t Template) execute(str TemplateString) string {
	out := bytes.NewBuffer(nil)
	parsed, err := template.New(fmt.Sprintf("%p", out)).Funcs(funcsMap).Parse(string(str))
	LogError(err)
	LogError(parsed.Execute(out, t))

	return out.String()
}

// Execute returns a template string with the given arguments.
func Execute[String interface{ string | TemplateString }](str String, t ...Template) string {
	if len(t) > 0 {
		return t[0].execute(TemplateString(str))
	}

	return Template{}.execute(TemplateString(str))
}

// Sprintf returns a template string with the given arguments.
func Sprintf[String interface{ string | TemplateString }](in String, args ...any) String {
	return String(fmt.Sprintf(string(in), args...))
}
