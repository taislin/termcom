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
