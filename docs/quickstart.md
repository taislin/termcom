# termcom Quick Start

An ASCII X-COM demake for your terminal. Command humanity's defense against alien invasion.

## Run

### Terminal Version

```bash
go run ./cmd/termcom      # or: make run
```

### Browser Version (WASM)

No backend server required — runs natively in the browser.

```bash
# Build and serve
cd cmd/termcom_wasm
GOOS=js GOARCH=wasm go build -o ../../web_wasm/termcom.wasm .
cd ../../web_wasm
python -m http.server 8080
# Open http://localhost:8080
```

Or use the build script: `./scripts/build_wasm.sh`

## Gameplay Loop

1. **Geoscape** -- UFOs fly toward cities. Detect and intercept them.
2. **Intercept** -- Launch fighters (L) or autoresolve (A) to shoot UFOs down.
3. **Battle** -- Deploy to crash sites (R). Enter tactical combat.
4. **Base** -- Research alien tech, manufacture gear, hire/equip soldiers.
5. **Repeat** -- Win 10 battles, then assault Cydonia to save Earth.

Lose if Alien Activity reaches 100%.

## Essential Keys (Geoscape)

| Key | Action |
|-----|--------|
| Space | Pause |
| 1-4 | Time speed |
| L | Launch interceptor |
| A | Autoresolve UFO |
| M | Respond to mission |
| R | Send transport to crash |
| B | Open base |
| F5/F9 | Save/Load |
| Q | Quit |

## Essential Keys (Battlescape)

| Key | Action |
|-----|--------|
| Arrow/WASD | Move cursor |
| Space/Enter | Select/Confirm |
| F | Fire weapon |
| R | Reload |
| Q | Cycle soldier |
| E | End turn |
| C | Crouch |
| Esc | Cancel |

## Quick Strategy

- **Early:** Hire soldiers, research Alien Alloys, build Lab + Workshop
- **Mid:** Custom laser weapons (Weapon Designer) → Personal Armour, expand bases
- **Late:** Custom plasma weapons, Power/Flying Suits, psi training
- Always equip soldiers before battle. Wounded heal 2 HP/day.
- Design custom interceptors in the Plane Designer — Stingray missiles + Light Alloy armour is a solid start.
- Sell excess alien artifacts for cash. Radar facilities boost funding.

## Victory

Win 10 ground battles to unlock the Cydonia final mission. Destroy Cydonia to win.

For the full manual see [manual.md](manual.md).
