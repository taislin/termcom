package soldier

import (
	"testing"

	"github.com/taislin/termcom/internal/data"
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

func TestPostMission(t *testing.T) {
	s := NewSoldier("Test")
	s.Accuracy = 40
	s.ExpFiring = 15
	s.GainedXP = true
	s.Rank = Rookie
	s.TU = 45
	s.HP = 22
	s.Strength = 15
	s.PostMission()
	if s.Accuracy < 42 || s.Accuracy > 46 {
		t.Errorf("firing XP should raise Accuracy by 2..6, got %d", s.Accuracy)
	}
	if s.Rank != Squaddie {
		t.Errorf("halo growth should promote Rookie to Squaddie, got %v", s.Rank)
	}
	if s.TU < 45 || s.TU > 49 {
		t.Errorf("halo TU growth should stay within cap, got %d", s.TU)
	}
	if s.HP < 22 || s.HP > 60 {
		t.Errorf("halo HP growth should stay within cap, got %d", s.HP)
	}
	if s.Strength < 15 || s.Strength > 70 {
		t.Errorf("halo Strength growth should stay within cap, got %d", s.Strength)
	}
	if s.ExpFiring != 0 || s.GainedXP {
		t.Error("PostMission should reset transient XP counters")
	}
}

func TestImproveStat(t *testing.T) {
	if v := improveStat(11); v < 2 || v > 6 {
		t.Errorf("exp>10 should yield 2..6, got %d", v)
	}
	if v := improveStat(6); v < 1 || v > 4 {
		t.Errorf("exp>5 should yield 1..4, got %d", v)
	}
	if v := improveStat(3); v < 1 || v > 3 {
		t.Errorf("exp>2 should yield 1..3, got %d", v)
	}
	if v := improveStat(1); v < 0 || v > 1 {
		t.Errorf("exp>0 should yield 0..1, got %d", v)
	}
	if v := improveStat(0); v != 0 {
		t.Errorf("exp=0 should yield 0, got %d", v)
	}
}

func TestHandlePromotions(t *testing.T) {
	roster := make([]*Soldier, 6)
	for i := range roster {
		roster[i] = NewSoldier("S")
		roster[i].Rank = Rookie
	}
	roster[0].Rank = Squaddie
	roster[0].Kills = 10
	roster[1].Rank = Squaddie
	roster[1].Kills = 5
	HandlePromotions(roster)
	if roster[0].Rank != Corporal {
		t.Errorf("top-kill soldier should be Corporal, got %v", roster[0].Rank)
	}
	if roster[1].Rank != Corporal {
		t.Errorf("second soldier should be Corporal, got %v", roster[1].Rank)
	}
	if roster[2].Rank != Rookie {
		t.Errorf("low-rank soldier should stay Rookie, got %v", roster[2].Rank)
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
	s1.WeaponAmmo = data.RuleItems["rifle"].AmmoMax
	s2.Armor = "none"

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
