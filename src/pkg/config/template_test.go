package config

import "testing"

func TestTemplate(t *testing.T) {
	type args struct {
		t   Template
		str templateString
	}
	for _, tt := range []struct {
		name string
		args args
		want string
	}{
		{"test#1",
			args{Template{Damage: 10000, Name: "Test", Level: 1}, `Damage: {{ printf "%d" .Damage }}, Name: {{ printf "%q" .Name }}, Level: {{ printf "%d" .Level }}`},
			`Damage: 10,000, Name: "Test", Level: 1`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.t.Execute(tt.args.str); got != tt.want {
				t.Errorf("Template.Render() = %v, want %v", got, tt.want)
			}
		})
	}
}
