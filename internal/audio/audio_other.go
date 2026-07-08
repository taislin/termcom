//go:build !windows

package audio

import "os"

func Init() {}
func Close() {}
func PlayClick() { os.Stdout.WriteString("\a") }
func PlayShoot() { os.Stdout.WriteString("\a") }
func PlayExplosion() { os.Stdout.WriteString("\a") }
func PlayChime() { os.Stdout.WriteString("\a") }
func PlayAlert() { os.Stdout.WriteString("\a") }
func PlayHit() { os.Stdout.WriteString("\a") }
func PlayMiss() { os.Stdout.WriteString("\a") }
func PlayAlienTurn() { os.Stdout.WriteString("\a") }
func PlayVictory() { os.Stdout.WriteString("\a") }
func PlayDefeat() { os.Stdout.WriteString("\a") }
func PlayGrenade() { os.Stdout.WriteString("\a") }
func PlaySelect() { os.Stdout.WriteString("\a") }
func PlayMove() { os.Stdout.WriteString("\a") }
func PlayReload() { os.Stdout.WriteString("\a") }
func PlayMedikit() { os.Stdout.WriteString("\a") }
