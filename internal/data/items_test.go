package data

import "testing"

func TestWeaponsExist(t *testing.T) {
	required := []string{"pistol", "rifle", "heavy", "auto", "rocket", "laser_pistol", "laser_rifle", "plasma_rifle", "plasma_pistol", "stun_rod", "medi_kit"}
	for _, name := range required {
		if _, ok := RuleItems[name]; !ok {
			t.Errorf("missing weapon: %s", name)
		}
	}
}

func TestWeaponStatsPositive(t *testing.T) {
	for name, w := range RuleItems {
		// Only check if it's a weapon (BattleType != BT_CORPSE for example)
		if w.BattleType == BT_FIREARM || w.BattleType == BT_MELEE {
			if w.Damage < 0 {
				t.Errorf("%s: negative damage %d", name, w.Damage)
			}
			if w.Accuracy < 0 || w.Accuracy > 100 {
				// Accuracy for medi-kit is 0, which is fine
				if w.BattleType != BT_MEDIKIT {
					t.Errorf("%s: accuracy out of range %d", name, w.Accuracy)
				}
			}
			if w.TU <= 0 {
				t.Errorf("%s: invalid TU %d", name, w.TU)
			}
			if w.Range <= 0 {
				t.Errorf("%s: invalid range %d", name, w.Range)
			}
		}
	}
}

func TestArmorsExist(t *testing.T) {
	required := []string{"none", "personal", "light", "medium", "heavy", "power_suit", "flight_suit"}
	for _, name := range required {
		if _, ok := Armors[name]; !ok {
			t.Errorf("missing armor: %s", name)
		}
	}
}

func TestArmorValuesNonNegative(t *testing.T) {
	for name, a := range Armors {
		if a.Undersuit < 0 {
			t.Errorf("%s: negative undersuit %d", name, a.Undersuit)
		}
	}
}

func TestItemsExist(t *testing.T) {
	if len(Items) == 0 {
		t.Error("no items defined")
	}
	for name, item := range Items {
		if item.Name == "" {
			t.Errorf("%s: empty name", name)
		}
		if item.Weight < 0 {
			t.Errorf("%s: negative weight", name)
		}
	}
}
