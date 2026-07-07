//go:build !windows

package audio

import "fmt"

func Init() {}
func Close() {}
func PlayClick() { fmt.Print("\a") }
func PlayShoot() { fmt.Print("\a") }
func PlayExplosion() { fmt.Print("\a") }
func PlayChime() { fmt.Print("\a") }
func PlayAlert() { fmt.Print("\a") }
