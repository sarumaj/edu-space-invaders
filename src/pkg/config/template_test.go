package config

import "testing"

func TestTemplate(t *testing.T) {
	type args struct {
		t   Template
		str TemplateString
	}
	for _, tt := range []struct {
		name string
		args args
		want string
	}{
		{"test#1",
			args{
				Template{"Damage": 10000, "Name": "Test", "Level": 1},
				`<p>Damage: {{ printf "%d" .Damage | color "red" }}, {{ "Name" | italic }}: {{ printf "%q" .Name | bold }}, Level: {{ printf "%d" .Level }}</p>`,
			},
			`<p>Damage: <span style="color: red;">10,000</span>, <i>Name</i>: <b>"Test"</b>, Level: 1</p>`},
		{"test#2",
			args{
				Template{"Damage": 100},
				`Your damage is
				{{ if gt (int .Damage) 1000 -}}
				pretty big
				{{- else -}}
				small
				{{- end }}
				comparing to the average damage.`,
			}, `Your damage is small comparing to the average damage.`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.t.execute(tt.args.str.Sanitize()); got != tt.want {
				t.Errorf("Template.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}
