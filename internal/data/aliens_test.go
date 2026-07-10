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
		if _, ok := RuleItems[at.Weapon]; !ok {
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

func TestAlienIconPoolsValid(t *testing.T) {
	for dmg, pool := range AlienIconsByDamage {
		if len(pool) == 0 {
			t.Errorf("damage type %d has an empty icon pool", dmg)
		}
		for _, r := range pool {
			if r < 0 || r > 0xFFFF {
				t.Errorf("damage type %d has non-BMP icon %U", dmg, r)
			}
			if r >= 0x1F300 {
				t.Errorf("damage type %d has emoji-range icon %U", dmg, r)
			}
		}
	}
}

func TestAlienIconsUnique(t *testing.T) {
	seen := map[rune]bool{}
	for _, at := range AlienTypes {
		if at.Icon == 0 {
			t.Errorf("%s: empty icon", at.Name)
		}
		if seen[at.Icon] {
			t.Errorf("duplicate icon %U across aliens", at.Icon)
		}
		seen[at.Icon] = true
	}
}

func TestProceduralAlienIconsUnique(t *testing.T) {
	for _, seed := range []int64{1, 42, 1337, 99999} {
		_, types := GenerateSpecies(seed)
		seen := map[rune]bool{}
		for _, at := range types {
			if seen[at.Icon] {
				t.Errorf("seed %d: duplicate procedural icon %U", seed, at.Icon)
			}
			seen[at.Icon] = true
			if at.Icon < 0 || at.Icon > 0xFFFF {
				t.Errorf("seed %d: non-BMP icon %U", seed, at.Icon)
			}
		}
		if len(types) == 0 {
			t.Errorf("seed %d: no procedural alien types generated", seed)
		}
	}
}
