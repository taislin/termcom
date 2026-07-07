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
- **Multiple Alien Species** — Sectoids, Floaters, Mutons, Ethereals

## Requirements

- Go 1.22+
- Terminal with Unicode support (for box-drawing characters)

## Build & Run

```bash
go run ./cmd/ycom
```

Or build a binary:

```bash
go build -o ycom ./cmd/ycom
./ycom
```

## Controls

### Geoscape
| Key | Action |
|-----|--------|
| Space | Pause / unpause time |
| 1-4 | Time compression |
| B | Open base |
| L | Launch interceptor |
| Esc | Quit |

### Battlescape
| Key | Action |
|-----|--------|
| Arrow keys / hjkl | Move cursor |
| Space | Select unit / confirm |
| F | Fire weapon |
| R | Reload |
| E | End turn |
| Esc | Cancel / deselect |
| ? | Help |

## Project Structure

See `AGENTS.md` for architecture details.

## License

MIT
