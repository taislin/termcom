//go:build !windows

package audio

import "os"

func bell() { os.Stdout.WriteString("\a") }

func Init()                  {}
func Close()                 {}
func PlayClick()             { bell() }
func PlayShoot()             { bell() }
func PlayExplosion()         { bell() }
func PlayChime()             { bell() }
func PlayAlert()             { bell() }
func PlayHit()               { bell() }
func PlayMiss()              { bell() }
func PlayAlienTurn()         { bell() }
func PlayVictory()           { bell() }
func PlayDefeat()            { bell() }
func PlayGrenade()           { bell() }
func PlaySelect()            { bell() }
func PlayMove()              { bell() }
func PlayReload()            { bell() }
func PlayMedikit()           { bell() }
func PlayLaserFire()         { bell() }
func PlayPlasmaFire()        { bell() }
func PlayMeleeFire()         { bell() }
func PlayBallisticFire()     { bell() }
func PlayMenuNav()           { bell() }
func PlayResearchComplete()  { bell() }
func PlayManufactureComplete() { bell() }
func PlayUFODetected()       { bell() }
func PlayMissionWarning()    { bell() }
func PlayDistantExplosion()  { bell() }
func PlayWind()              { bell() }
