# TERMCOM: An ASCII X-COM Demake

A roguelike-ified demake of **X-COM: UFO Defense** *(1994, MicroProse)* rendered entirely in coloured ASCII on a terminal. Written in Go with [tcell](https://github.com/gdamore/tcell). It brings the classic alien invasion strategy experience to your terminal. It features a complete implementation of all gameplay loops: the Geoscape (global strategy), Base management, and the tactical Battlescape.

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/) [![GitHub License](https://img.shields.io/github/license/taislin/termcom)](https://github.com/taislin/termcom/blob/master/LICENSE) [![Release](https://img.shields.io/github/v/release/taislin/termcom)](https://github.com/taislin/termcom/releases/latest)

[![Download Here](https://img.shields.io/badge/Download%20Here-8A2BE2)](https://github.com/taislin/termcom/releases/latest)

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
> [!TIP]
> The below is for building from the source code. If you just want to play, you can download the game binaries from [here](https://github.com/taislin/termcom/releases/latest).

- Go 1.25+
- Terminal with Unicode support (for box-drawing characters)

### Terminal Font Troubleshooting

**Termcom** makes heavy use of extended Unicode characters (Runes, Geometrics, and Ethiopic symbols) to render aliens and tactical maps. Most devices should support the characters we use, but if you launch the game natively in your terminal and see overlapping characters, weird spacing, or empty boxes (▯) instead of aliens, your OS is missing the required fallback fonts.

#### 🐧 Linux
To fix this on **Ubuntu/Debian**, install the Noto Fonts and Unifont fallback packages:
```bash
sudo apt update
sudo apt install fonts-noto fonts-unifont
```
For **Arch Linux** users:
```bash
sudo pacman -S noto-fonts unifont
```
For **Fedora** users:
```bash
sudo dnf install google-noto-sans-fonts unifont-fonts
```

#### 🍏 macOS
The default `Terminal.app` on macOS can sometimes struggle with grid alignment or incorrectly render symbols as double-width emojis. 

For the best experience, we highly recommend using **[iTerm2](https://iterm2.com/)**. If you are missing characters, you can install GNU Unifont via Homebrew:
```bash
brew install --cask font-gnu-unifont
```
* **To fix alignment in iTerm2:** Go to **Settings > Profiles > Text**, check the box for *"Use a different font for non-ASCII text"*, and set that secondary font to `Unifont`.

#### 🪟 Windows
Do not use the legacy command prompt (`cmd.exe`) or the old blue PowerShell window, as their Unicode and color support is extremely limited.

1. Use **[Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701)** (Included by default in Windows 11, available in the Microsoft Store for Windows 10).
2. If you see empty boxes `▯`, your system's default font lacks the required symbols. 
3. Download and install a robust, high-compatibility font like **[GNU Unifont](http://unifoundry.com/unifont/)** or **[Noto Sans Mono](https://fonts.google.com/noto/specimen/Noto+Sans+Mono)**.
4. Open Windows Terminal settings (`Ctrl + ,`), go to **Profiles > Defaults > Appearance**, and change the **Font face** to your newly installed font.

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

> [!CAUTION]
> The browser version is experimental and may have limited functionality compared to the terminal version.
 
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

### Android Native (Experimental)

> [!CAUTION]
> The Android version is experimental and may have limited functionality compared to the terminal version.

The Android port compiles the Go game core into a native `.aar` library via [gomobile](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile), rendered as a character grid on a `SurfaceView` with full touch input and audio.

**Prerequisites:**

- Go 1.25+
- Android SDK + NDK (API 21+)
- Gradle 8.2 (for local APK builds)
- `gomobile`:
  ```bash
  go install golang.org/x/mobile/cmd/gomobile@latest
  gomobile init
  ```

**Build the game library:**

```bash
make android-aar
```

This writes `android/app/libs/termcom.aar`.

**Build an APK (CI / GitHub Actions):**

A `.github/workflows/android-release.yml` workflow automatically builds a signed
release APK (or debug APK) on push to `mobile`/`main`/`master` and on `v*` tags.
Debug APKs are produced from any push; tag a release (`v*`) to publish a signed
APK as a GitHub Release. Set these repository secrets for signed releases:
`ANDROID_KEYSTORE_BASE64`, `ANDROID_KEYSTORE_PASSWORD`, `ANDROID_KEY_ALIAS`,
`ANDROID_KEY_PASSWORD`.

**Build an APK locally:**

```bash
make android-aar                                  # step 1: Go .aar
cd android && gradle assembleDebug               # step 2: APK → app/build/outputs/apk/debug/
# or open android/ in Android Studio and Run
```

Install with `adb install android/app/build/outputs/apk/debug/app-debug.apk`.

**Controls:**
- Tap to click / select / move
- Long-press (500ms) for right-click / cancel; vibration on long-press
- Drag to scroll
- Hardware keyboard (DPAD, Enter, Escape, F-keys) supported

## Project Structure

See the [AGENTS file](AGENTS.md) for architecture details.

## License

MIT, see [LICENSE](LICENSE) file.

> [!NOTE]
> ***AI Usage Disclaimer***: AI was used in this project to generate and update the translations to French, Spanish, Russian, Korean, Chinese and Japanese. No audio or images were generated via AI (the game does not use any, anyway - its all terminal based!)
