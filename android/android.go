//go:build android

// Package android provides a Go library bridge for the native Android port of termcom.
// It exports types and functions via gomobile bind, wrapping the game engine with
// an androidScreen (tcell.Screen implementation) and a flat-bytecell format for
// the Java Canvas renderer.
//
// Build with:
//
//	gomobile bind -target=android -androidapi 21 -o android/app/libs/termcom.aar ./android/
package android
