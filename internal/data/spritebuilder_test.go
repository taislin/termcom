package data

import (
	"fmt"
	"testing"
)

func TestTraitMapping(t *testing.T) {
	cases := []struct {
		name      string
		m         *Morphology
		wantSense Sense
		wantManip Manipulators
		wantLoco  Locomotion
	}{
		{"Standard human", &Morphology{Arms: 2, Legs: 2, Eyesight: "normal", Hearing: "normal"}, SenseStandard, ManipBipedal, LocomBipedal},
		{"Floating 0-arm", &Morphology{Arms: 0, Legs: 0, Eyesight: "normal", Hearing: "normal"}, SenseStandard, ManipNone, LocomFloating},
		{"4-arm 4-leg", &Morphology{Arms: 4, Legs: 4, Eyesight: "normal", Hearing: "normal"}, SenseStandard, ManipMultiArmed, LocomArachnid},
		{"Echolocation", &Morphology{Arms: 2, Legs: 2, Eyesight: "poor", Hearing: "echolocation"}, SenseEcholocation, ManipBipedal, LocomBipedal},
		{"Omni-sense", &Morphology{Arms: 2, Legs: 2, PsionicSense: "high", ThermalSense: "high"}, SenseOmni, ManipBipedal, LocomBipedal},
	}
	for _, c := range cases {
		s := SenseFromMorphology(c.m)
		m := ManipulatorsFromMorphology(c.m)
		l := LocomotionFromMorphology(c.m)
		if s != c.wantSense || m != c.wantManip || l != c.wantLoco {
			t.Errorf("%s: got Sense=%d Manip=%d Loco=%d, want Sense=%d Manip=%d Loco=%d",
				c.name, s, m, l, c.wantSense, c.wantManip, c.wantLoco)
		}
	}
}

func TestTraitDrivenPixels(t *testing.T) {
	morphs := []*Morphology{
		{Arms: 0, Legs: 0, Eyesight: "normal", Hearing: "normal"},
		{Arms: 2, Legs: 2, Eyesight: "normal", Hearing: "normal"},
		{Arms: 4, Legs: 4, Eyesight: "normal", Hearing: "normal"},
		{Arms: 2, Legs: 2, Eyesight: "poor", Hearing: "echolocation"},
		{Arms: 2, Legs: 2, PsionicSense: "high", ThermalSense: "high"},
		{Arms: 3, Legs: 6, Eyesight: "normal", Hearing: "normal"},
	}

	for i, m := range morphs {
		ap := GenerateAlienPixels(int64(i*1000), m)
		bodyCount := 0
		weaponCount := 0
		for y := 0; y < 24; y++ {
			for x := 0; x < 20; x++ {
				if ap.Body[y][x] {
					bodyCount++
				}
				if ap.Weapon[y][x] {
					weaponCount++
				}
			}
		}
		fmt.Printf("morph[%d] arms=%d legs=%d -> body=%d weapon=%d\n", i, m.Arms, m.Legs, bodyCount, weaponCount)
		if bodyCount == 0 {
			t.Errorf("morph[%d]: all body pixels empty", i)
		}
	}
}

func TestWeaponOnlyOnBipedalTorso(t *testing.T) {
	mBipedal := &Morphology{Arms: 2, Legs: 2, Eyesight: "normal", Hearing: "normal"}
	ap := GenerateAlienPixels(42, mBipedal)
	weaponRows := 0
	for y := 10; y < 18; y++ {
		for x := 0; x < 20; x++ {
			if ap.Weapon[y][x] {
				weaponRows++
				break
			}
		}
	}
	if weaponRows == 0 {
		t.Error("bipedal torso should have weapon pixels")
	}

	mNone := &Morphology{Arms: 0, Legs: 0, Eyesight: "normal", Hearing: "normal"}
	ap2 := GenerateAlienPixels(42, mNone)
	for y := 0; y < 24; y++ {
		for x := 0; x < 20; x++ {
			if ap2.Weapon[y][x] {
				t.Error("none torso should have no weapon pixels")
				return
			}
		}
	}
}

func TestNilMorphologyFallback(t *testing.T) {
	ap := GenerateAlienPixels(42, nil)
	total := 0
	for y := 0; y < 24; y++ {
		for x := 0; x < 20; x++ {
			if ap.Body[y][x] {
				total++
			}
		}
	}
	if total == 0 {
		t.Error("nil morphology produced empty pixels")
	}
}

func TestConsistency(t *testing.T) {
	m := &Morphology{Arms: 2, Legs: 2, Eyesight: "normal", Hearing: "normal"}
	p1 := GenerateAlienPixels(123, m)
	p2 := GenerateAlienPixels(123, m)
	if p1 != p2 {
		t.Error("same seed + morphology produced different results")
	}
}

func TestAlienNewLayersPopulated(t *testing.T) {
	m := &Morphology{Arms: 2, Legs: 2, Eyesight: "normal", Hearing: "normal"}
	ap := GenerateAlienPixels(42, m)
	interior, belly, texture := 0, 0, 0
	for y := 0; y < 24; y++ {
		for x := 0; x < 20; x++ {
			if ap.Interior[y][x] {
				interior++
			}
			if ap.Belly[y][x] {
				belly++
			}
			if ap.Texture[y][x] {
				texture++
			}
		}
	}
	if interior == 0 {
		t.Error("expected interior pixels for 3D rounding")
	}
	if belly == 0 {
		t.Error("expected belly patch pixels")
	}
	if texture == 0 {
		t.Error("expected texture speckle pixels")
	}
}

func TestWeaponColorBrighter(t *testing.T) {
	bR, bG, bB := int32(100), int32(120), int32(80)
	wR, wG, wB := AlienWeaponColor()
	if wR <= bR || wG <= bG || wB <= bB {
		t.Errorf("weapon color should be brighter: body=(%d,%d,%d) weapon=(%d,%d,%d)", bR, bG, bB, wR, wG, wB)
	}
	if wR > 255 || wG > 255 || wB > 255 {
		t.Errorf("weapon color overflow: (%d,%d,%d)", wR, wG, wB)
	}
}
