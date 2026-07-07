package soldier

import (
	"testing"

	"github.com/civ13/ycom/internal/data"
)

func TestNewSoldier(t *testing.T) {
	s := NewSoldier("Test")
	if s.Name != "Test" {
		t.Errorf("expected name Test, got %s", s.Name)
	}
	if s.Rank != Rookie {
		t.Errorf("expected Rookie rank, got %v", s.Rank)
	}
	if s.HP < 20 || s.HP > 25 {
		t.Errorf("HP out of range: %d", s.HP)
	}
	if s.TU < 45 || s.TU > 55 {
		t.Errorf("TU out of range: %d", s.TU)
	}
	if s.Weapon != "rifle" {
		t.Errorf("expected rifle, got %s", s.Weapon)
	}
	if s.Armor != "none" {
		t.Errorf("expected none armor, got %s", s.Armor)
	}
}

func TestRankString(t *testing.T) {
	tests := []struct {
		rank Rank
		want string
	}{
		{Rookie, "Rookie"},
		{Squaddie, "Squaddie"},
		{Corporal, "Corporal"},
		{Sergeant, "Sergeant"},
		{Lieutenant, "Lieutenant"},
		{Captain, "Captain"},
		{Major, "Major"},
		{Colonel, "Colonel"},
		{Rank(100), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.rank.String(); got != tt.want {
			t.Errorf("Rank(%d).String() = %q, want %q", tt.rank, got, tt.want)
		}
	}
}

func TestGainXP(t *testing.T) {
	s := NewSoldier("Test")
	if s.Rank != Rookie {
		t.Fatal("should start as Rookie")
	}
	s.GainXP(30)
	if s.Rank <= Rookie {
		t.Error("should have promoted after 30 kills")
	}
	if s.HP <= 20 {
		t.Error("HP should have increased")
	}
}

func TestSquadAlive(t *testing.T) {
	s1 := NewSoldier("A")
	s2 := NewSoldier("B")
	s1.HP = 0
	sq := Squad{s1, s2}
	alive := sq.Alive()
	if len(alive) != 1 {
		t.Errorf("expected 1 alive, got %d", len(alive))
	}
	if alive[0].Name != "B" {
		t.Errorf("expected B alive, got %s", alive[0].Name)
	}
}

func TestSquadAllDead(t *testing.T) {
	s1 := NewSoldier("A")
	s2 := NewSoldier("B")
	s1.HP = 0
	s2.HP = 0
	sq := Squad{s1, s2}
	if !sq.AllDead() {
		t.Error("expected AllDead to be true")
	}
}

func TestFireWeapon(t *testing.T) {
	s1 := NewSoldier("Attacker")
	s2 := NewSoldier("Target")
	s1.Weapon = "rifle"
	s2.Armor = "none"
	w := data.Weapons["rifle"]
	w.AmmoCur = w.AmmoMax
	data.Weapons["rifle"] = w

	s2.HP = 100
	_, _ = s1.FireWeapon(s2)
}

func TestRandomName(t *testing.T) {
	name := RandomName()
	if name == "" {
		t.Error("RandomName returned empty string")
	}
}

func TestFormatSoldier(t *testing.T) {
	s := NewSoldier("Test")
	str := FormatSoldier(s)
	if str == "" {
		t.Error("FormatSoldier returned empty string")
	}
}
