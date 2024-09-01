package enemy

import "testing"

func TestNext(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   EnemyType
		out  EnemyType
	}{
		{name: "Normal", in: Normal, out: Berserker},
		{name: "Tank", in: Tank, out: Tank},
		{name: "Freezer", in: Freezer, out: Berserker},
		{name: "Cloaked", in: Cloaked, out: Berserker},
		{name: "Berserker", in: Berserker, out: Annihilator},
		{name: "Annihilator", in: Annihilator, out: Juggernaut},
		{name: "Juggernaut", in: Juggernaut, out: Dreadnought},
		{name: "Dreadnought", in: Dreadnought, out: Behemoth},
		{name: "Behemoth", in: Behemoth, out: Colossus},
		{name: "Colossus", in: Colossus, out: Leviathan},
		{name: "Leviathan", in: Leviathan, out: Bulwark},
		{name: "Bulwark", in: Bulwark, out: Overlord},
		{name: "Overlord", in: Overlord, out: Overlord},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.Next()
			if got != tt.out {
				t.Errorf("Next() = %v, want %v", got, tt.out)
			}
		})
	}
}
