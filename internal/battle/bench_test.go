package battle

import (
	"math/rand"
	"testing"

	"github.com/taislin/termcom/internal/data"
	"github.com/taislin/termcom/internal/soldier"
)

func BenchmarkAIUpdate(b *testing.B) {
	m, _ := GenerateCrashSite(30, 24, 42)
	alien := &Unit{X: 10, Y: 10, TU: 40, MaxTU: 40, HP: 20, MaxHP: 20, Accuracy: 60, Alive: true, Faction: 1}
	ai := NewAlienAI(alien)
	human := &Unit{X: 12, Y: 10, TU: 50, HP: 20, MaxHP: 20, Armour: 0, Alive: true, Faction: 0}
	humans := UnitList{human}
	allUnits := UnitList{alien, human}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alien.TU = alien.MaxTU
		ai.Update(allUnits, m, humans, nil, nil)
	}
}

func BenchmarkAIPatrol(b *testing.B) {
	m, _ := GenerateCrashSite(30, 24, 42)
	alien := &Unit{X: 10, Y: 10, TU: 40, MaxTU: 40, HP: 20, MaxHP: 20, Alive: true, Faction: 1}
	ai := NewAlienAI(alien)
	ai.PatrolX = 15
	ai.PatrolY = 15

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alien.TU = alien.MaxTU
		ai.patrolTarget(m)
	}
}

func BenchmarkLOS(b *testing.B) {
	m, _ := GenerateCrashSite(30, 24, 42)
	s := &Unit{X: 5, Y: 5, Alive: true, Faction: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.CanSee(20, 20, m)
	}
}

func BenchmarkLOSClose(b *testing.B) {
	m, _ := GenerateCrashSite(30, 24, 42)
	s := &Unit{X: 5, Y: 5, Alive: true, Faction: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.CanSee(6, 6, m)
	}
}

func BenchmarkFireAt(b *testing.B) {
	attacker := &Unit{
		X: 5, Y: 5, TU: 50, MaxTU: 50,
		Accuracy: 70, Weapon: "rifle",
		Alive: true, Faction: 0,
	}
	defender := &Unit{
		X: 6, Y: 5, HP: 50, MaxHP: 50,
		Armour: 0, Alive: true, Faction: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attacker.TU = 50
		defender.HP = 50
		defender.Alive = true
		_, _, _, _ = attacker.FireAt(defender, nil, nil)
	}
}

func BenchmarkMovement(b *testing.B) {
	m, _ := GenerateCrashSite(30, 24, 42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u := &Unit{X: 10, Y: 10, TU: 50, MaxTU: 50, Alive: true, Faction: 0}
		u.MoveTo(15, 15, m)
	}
}

func BenchmarkMapGeneration(b *testing.B) {
	b.Run("CrashSite", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateCrashSite(30, 24, 42)
		}
	})
	b.Run("TerrorSite", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateTerrorSite(30, 24, 42)
		}
	})
	b.Run("UFOInterior", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateUFOInterior(30, 24, 42)
		}
	})
	b.Run("UFOInteriorWFC", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateUFOInteriorWFC(30, 24, rand.New(rand.NewSource(1)))
		}
	})
	b.Run("Cydonia", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateCydonia(30, 24)
		}
	})
}

func BenchmarkSoldierCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		soldier.NewSoldier("Bench")
	}
}

func BenchmarkAlienCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewAlienUnit(data.GetAlienByRank(0))
	}
}

func BenchmarkRankProgression(b *testing.B) {
	s := soldier.NewSoldier("Bench")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.ExpFiring = 5
		s.GainedXP = true
		s.PostMission()
	}
}
