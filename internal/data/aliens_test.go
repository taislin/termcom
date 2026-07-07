package data

import "testing"

func TestAlienTypesPopulated(t *testing.T) {
	if len(AlienTypes) == 0 {
		t.Fatal("no alien types defined")
	}
}

func TestAlienStatsPositive(t *testing.T) {
	for _, at := range AlienTypes {
		if at.HP <= 0 {
			t.Errorf("%s: invalid HP %d", at.Name, at.HP)
		}
		if at.TU <= 0 {
			t.Errorf("%s: invalid TU %d", at.Name, at.TU)
		}
		if at.Weapon == "" {
			t.Errorf("%s: no weapon assigned", at.Name)
		}
		if _, ok := Weapons[at.Weapon]; !ok {
			t.Errorf("%s: unknown weapon %s", at.Name, at.Weapon)
		}
	}
}

func TestGetAlienByName(t *testing.T) {
	s := GetAlienByName("Sectoid")
	if s == nil {
		t.Fatal("GetAlienByName(Sectoid) returned nil")
	}
	if s.Name != "Sectoid" {
		t.Errorf("expected Sectoid, got %s", s.Name)
	}
	if GetAlienByName("Nonexistent") != nil {
		t.Error("expected nil for nonexistent alien")
	}
}

func TestGetAlienByRank(t *testing.T) {
	a := GetAlienByRank(0)
	if a == nil {
		t.Fatal("GetAlienByRank(0) returned nil")
	}
	if a.Rank < 0 {
		t.Errorf("expected rank >= 0, got %d", a.Rank)
	}

	a5 := GetAlienByRank(100)
	if a5 != nil {
		t.Error("expected nil for impossible rank")
	}
}
