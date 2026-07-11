# termcom — ASCII X-COM Demake

A roguelike-ified demake of **X-COM: UFO Defense** *(1994, MicroProse)* rendered entirely in coloured ASCII on a terminal. Written in Go with [tcell](https://github.com/gdamore/tcell). It brings the classic alien invasion strategy experience to your terminal. It features a complete implementation of all gameplay loops: the Geoscape (global strategy), Base management, and the tactical Battlescape.

## Features

- **Geoscape** — Real-time world map with time compression, UFO tracking, interceptor launch.
- **Battlescape** — Turn-based tactical combat with Time Units, cover, line-of-sight.
- **Base Management** — Build facilities, hire soldiers, equip squad.
- **Research & Manufacturing** — Unlock alien tech, build plasma rifles and power suits.
- **Alien AI** — Patrol, seek, attack, flee, flank, and retreat behaviors.
- **Procedurally Generated Aliens** — A roster of aliens is generated at the start of each campaign, each with unique abilities, strengths, weaknesses, and weapons.
- **Destructible Terrain** — Grenades destroy walls, trees, and rocks, leaving rubble.
- **Dynamic VFX** — Particle explosions, screen shake, rubble physics, night lighting.

## Requirements

- Go 1.22+
- Terminal with Unicode support (for box-drawing characters)
- Web browser (for browser version)

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

**Note:** The browser version is experimental and may have limited functionality compared to the terminal version.

## Controls

### Geoscape
| Key | Action |
|-----|--------|
| Space | Pause / unpause time |
| 1-4 | Time compression |
| B | Open base |
| L | Launch interceptor |
| A | Autoresolve nearest UFO |
| M | Respond to mission |
| R | Dispatch transport |
| E | Open encyclopedia |
| F5 / F9 | Save / Load |
| Q | Quit |

### Battlescape
| Key | Action |
|-----|--------|
| Arrow keys / hjkl / WASD | Move cursor |
| Space / Enter | Select unit / confirm |
| Q | Cycle soldiers |
| F | Fire weapon |
| R | Reload |
| E / N | End turn |
| G | Grenade |
| M | Medikit |
| C | Crouch |
| V | Toggle vision mode (Normal / Night / Thermal) |
| P | Psi attack |
| Esc | Cancel / deselect |
| ? | Help |

## Project Structure

See `AGENTS.md` for architecture details.

## License

MIT, see `LICENSE` file.
