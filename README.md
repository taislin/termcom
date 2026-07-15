# TERMCOM: An ASCII X-COM Demake

A roguelike-ified demake of **X-COM: UFO Defense** *(1994, MicroProse)* rendered entirely in coloured ASCII on a terminal. Written in Go with [tcell](https://github.com/gdamore/tcell). It brings the classic alien invasion strategy experience to your terminal. It features a complete implementation of all gameplay loops: the Geoscape (global strategy), Base management, and the tactical Battlescape.

**Download the latest version [here](https://github.com/taislin/termcom/releases/).**

## Features

- **Geoscape** - Real-time world map with time compression, UFO tracking, interceptor launch.
- **Battlescape** - Turn-based tactical combat with Time Units, cover, line-of-sight.
- **Base Management** - Build facilities, hire soldiers, equip squad.
- **Research & Manufacturing** - Unlock alien tech, build plasma rifles and power suits.
- **Alien AI** - Patrol, seek, attack, flee, flank, and retreat behaviors.
- **Procedurally Generated Aliens** - A roster of aliens is generated at the start of each campaign, each with unique abilities, strengths, weaknesses, and weapons.
- **Destructible Terrain** - Grenades destroy walls, trees, and rocks, leaving rubble.
- **Dynamic VFX** - Particle explosions, screen shake, rubble physics, night lighting.

## Requirements
If can download the game binaries from [here](https://github.com/taislin/termcom/releases/). If you want to download the source code, you will need the following:

- Go 1.25+
- Terminal with Unicode support (for box-drawing characters)

## Build & Run

### Terminal Version

```bash
go run ./cmd/termcom
```

Or build a binary:

```bash
go build -o termcom ./cmd/termcom
./termcom
```

### Browser Version (Experimental)

The browser version allows you to play termcom in a web browser using xterm.js.

1. Start the web server:

```bash
go run ./cmd/webserver
```

2. Open your browser and navigate to:

```
http://localhost:8080
```

The browser version supports:
- Full keyboard input via xterm.js
- WebSocket-based real-time communication
- Responsive terminal resizing
- All game features (Geoscape, Battlescape, Base Management)
- **Mobile touch play** — tap to click, long-press for right-click, drag to scroll, on-screen control menu with context-sensitive buttons

**Note:** The browser version is experimental and may have limited functionality compared to the terminal version.

### Android Native (Experimental)

The Android port compiles the Go game core into a native `.aar` library via [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile), rendered as a character grid on a `SurfaceView` with full touch input.

**Prerequisites:**

- Go 1.25+
- Android SDK + NDK (API 21+)
- `gomobile`:
  ```bash
  go install golang.org/x/mobile/cmd/gomobile@latest
  gomobile init
  ```

**Build:**

```bash
# 1. Build the Go .aar library
make android-aar

# 2. Open android/ in Android Studio, sync Gradle, and run on device
```

The `.aar` is written to `android/app/libs/termcom.aar`. Open `android/` as an existing project in Android Studio, then build and deploy to a device or emulator running Android 5.0+.

**Controls:**
- Tap to click / select / move
- Long-press (500ms) for right-click / cancel
- Drag to scroll
- Hardware keyboard (DPAD, Enter, Escape, F-keys) supported

## Project Structure

See the [AGENTS file](AGENTS.md) for architecture details.

## License

MIT, see [LICENSE](LICENSE) file.

***AI Usage Disclaimer***: AI was used in this project to generate and update the translations to French, Spanish, Russian, Korean, Chinese and Japanese. No audio or images were generated via AI (the game does not use any, anyway - its all terminal based!)
