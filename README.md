# ycom — ASCII X-COM Demake

A faithful demake of **X-COM: UFO Defense** (1994, MicroProse) rendered entirely
in coloured ASCII on a terminal. Written in Go with [tcell](https://github.com/gdamore/tcell).


## Features

- **Geoscape** — Real-time world map with time compression, UFO tracking, interceptor launch
- **Battlescape** — Turn-based tactical combat with Time Units, cover, line-of-sight
- **Base Management** — Build facilities, hire soldiers, equip squad
- **Research & Manufacturing** — Unlock alien tech, build plasma rifles and power suits
- **Soldier Progression** — Stats improve with combat experience, ranks from Rookie to Colonel
- **Alien AI** — Patrol, seek, attack, and flee behaviours
- **Multiple Alien Species** — Sectoids, Floaters, Mutons, Ethereals, Chryssalids, Cyberdiscs, and more
- **Destructible Terrain** — Grenades destroy walls, trees, and rocks, leaving rubble
- **Volumetric Gas** — Smoke clouds expand and thin over turns, blocking LOS and providing cover
- **Vision Modes** — Night vision (green phosphor) and thermal overlay (heat signatures)
- **Dynamic VFX** — Particle explosions, screen shake, rubble physics, night lighting
- **Blood & Fire** — Persistent blood decals (faction-colored), animated fire with spread
- **Browser Version** — Play in your web browser (experimental)

## Requirements

- Go 1.22+
- Terminal with Unicode support (for box-drawing characters)
- Web browser (for browser version)

## Build & Run

### Terminal Version

```bash
go run ./cmd/ycom
```

Or build a binary:

```bash
go build -o ycom ./cmd/ycom
./ycom
```

### Browser Version (Experimental)

The browser version allows you to play YCOM in a web browser using xterm.js.

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

MIT

## Browser Version

The browser version uses:
- [xterm.js](https://xtermjs.org/) for terminal rendering in the browser
- [gorilla/websocket](https://github.com/gorilla/websocket) for WebSocket communication
- Go HTTP server for serving the web application

To run the browser version:

```bash
# Start the web server
go run ./cmd/webserver

# Or with custom port
go run ./cmd/webserver :3000
```

Then open http://localhost:8080 in your browser.
