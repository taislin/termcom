# Game Tables

## Alien Units

Alien stats are defined in `internal/data/aliens.go`. See that file for the full
stat blocks. The table below lists species, icons, and general role.

| Name | Icon | Rank Tiers | Role |
|------|------|-----------|------|
| Sectoid | Ω | Grunt / Navigator / Commander | Psi-focused, low HP |
| Floater | ∞ | Grunt / Navigator / Commander | Balanced, mid-range |
| Chryssalid | ψ | Grunt / Queen | Melee only, high aggression |
| Hyperworm | ≈ | Grunt only | Fast, low HP, pack tactics |
| Silacoid | ▓ | Grunt only | Heavy armor, slow, tank |
| Muton | Σ | Grunt / Navigator / Commander | Heavy hitter, high HP |
| Cyberdisc | ◎ | Commander only | Flying, heavy plasma |
| Celatid | ◇ | Grunt only | Ranged acid attack |
| Ethereal | Ψ | Grunt / Navigator / Commander | Psi master, high stats |
| Reaper | ♠ | Commander only | Boss, very high HP |
| Sectopod | ⊞ | Boss only | Endgame, massive HP + armor |

## Human Soldiers

Soldier stat ranges are defined in `internal/soldier/soldier.go`. Stats scale
with rank and combat experience.

| Rank | Color | Notes |
|------|-------|-------|
| Rookie | White | Starting stat ranges |
| Soldier | Green | First promotion |
| Sergeant | Cyan | |
| Captain | Blue | |
| Major | Magenta | |
| Colonel | Red | |
| Commander | Yellow | Highest rank |

## Weapons

Weapon stats are defined in `internal/data/items.go`. See that file for full
stat blocks. The table below lists key, name, damage type, and notable traits.

| Key | Name | Damage Type | Notable Traits |
|-----|------|-------------|----------------|
| pistol | Pistol | conventional | Starting sidearm |
| rifle | Rifle | conventional | Standard issue |
| heavy | Heavy Cannon | explosive | High damage, slow |
| auto | Auto Cannon | explosive | 3-round burst |
| rocket | Rocket Launcher | explosive | Devastating, 1 ammo |
| laser_pistol | Laser Pistol | energy | Unlimited ammo |
| laser_rifle | Laser Rifle | energy | Unlimited ammo |
| plasma_rifle | Plasma Rifle | plasma | Alien tech, high damage |
| plasma_pistol | Plasma Pistol | plasma | Alien sidearm |
| heavy_plasma | Heavy Plasma | plasma | Best weapon in game |
| chryssalid_claw | Chryssalid Claw | melee | Alien melee |
| reaper_claw | Reaper Claw | melee | Boss melee |
| stun_rod | Stun Rod | melee | Non-lethal |
| medi_kit | Medi-Kit | special | Heals 10 HP |

## Armor

Armor stats are defined in `internal/data/items.go`.

| Key | Name | Health Bonus | TU Modifier |
|-----|------|-------------|-------------|
| none | None | 0 | 0 |
| personal | Personal Armour | +10 | 0 |
| light | Light Suit | +20 | -5 |
| medium | Medium Suit | +30 | -10 |
| heavy | Heavy Suit | +40 | -15 |
| power_suit | Power Suit | +50 | -10 |
| flight_suit | Flying Suit | +45 | -5 |

## Tile Types

Tile definitions are in `internal/battle/map.go`. The `tileChars` map defines the
rendering character for each type, and `TileCover()` defines cover values.

Key tile categories:
- **Floors:** Floor `.`, UFO Floor `≡`, Pavement `░`, Sand `·`
- **Walls:** Wall `#`, UFO Wall `█`, Fence `║`
- **Nature:** Grass `·`, Tree `♣`, Rock `∩`, Water `≈`, Snow `∗`, Marsh `≋`, Bush `†`
- **Doors:** Door `+`
- **Objects:** Object `■`, Rubble `▒`, Stairs `▓`
- **UFO furniture:** Console `░`, Machinery `⚙`, Pod `◈`, Power Source `⌁`, Storage `▤`, Alien Tech `⊕`

## Soldier Ranks

Ranks are defined in `internal/soldier/soldier.go`.

| Rank | Name | Color |
|------|------|-------|
| 0 | Rookie | White |
| 1 | Soldier | Green |
| 2 | Sergeant | Cyan |
| 3 | Captain | Blue |
| 4 | Major | Magenta |
| 5 | Colonel | Red |
| 6 | Commander | Yellow |

## Map Sizes

All tactical maps are 50×50 tiles. Map generation is in `internal/battle/map.go`.

## Mission Types

Mission types are defined in `internal/battle/battlescape.go` (`NewBattlescape`).

| Type | Name | Description |
|------|------|-------------|
| crash | Crash Site | Outdoor terrain with UFO wreckage |
| terror | Terror Site | Urban map with buildings |
| supply | Supply (UFO Interior) | Inside a UFO, multiple rooms |
| alien_base | Alien Base | Rocky terrain with alien structure |
| forest | Forest | Dense trees |
| desert | Desert | Rocks and sand |
| polar | Polar | Snow and ice |

## Key Bindings

### Geoscape

| Key | Action |
|-----|--------|
| Space | Pause/unpause time |
| 1-4 | Time compression |
| B | Open base |
| L | Launch interceptor |
| A | Autoresolve nearest UFO |
| M | Respond to mission |
| R | Dispatch transport to crash site |
| E | Open encyclopedia |
| F5 | Save |
| F9 | Load |
| Q | Quit |

### Battlescape

| Key | Action |
|-----|--------|
| Arrow keys / hjkl / WASD | Move cursor / move unit |
| Space / Enter | Select/confirm |
| Q | Cycle soldiers |
| F | Fire weapon |
| R | Reload |
| E / N | End turn |
| G | Grenade |
| M | Medikit |
| C | Crouch |
| V | Toggle vision mode (Normal / Night / Thermal) |
| P | Psi attack |
| Esc | Cancel |
| ? | Help |

### Base Management

| Key | Action |
|-----|--------|
| 1-5 | Switch tabs |
| j/k | Navigate |
| B | Build facility |
| S | Sell facility |
| H | Hire soldier |
| E | Equip soldier |
| D | Dismiss soldier |
| R | Research |
| M | Manufacture |
| Esc | Back to geoscape |

## Visual Effects

| Effect | Trigger | Description |
|--------|---------|-------------|
| Explosion particles | Grenade / weapon hit | Burst of `*` / `+` chars with color fade |
| Screen shake | Explosions | Camera shake scaled by damage |
| Smoke particles | Grenade impact | `~` / `:` chars drifting upward |
| Night lighting | Night missions | Radial glow around units |
| Volumetric gas | Grenade (new) | Expanding `▓`/`▒`/`░` clouds, blocks LOS at density 3 |
| Rubble particles | Wall destruction (new) | `.` `*` `,` `'` in parabolic arcs |
| Night vision filter | `V` key (new) | Green phosphor overlay with static noise |
| Thermal filter | `V` key (new) | Entities glow hot, terrain is cold blue |
| Blood splatter | Damage (new) | `,` `%` `:` runes in red/green/purple on floor tiles |
| Animated fire | Plasma/explosive (new) | Cycles `^` `w` `*` in yellow/orange/red |

## Blood Splatter

| Source | Color | Rune |
|--------|-------|------|
| Human damage | Dark red (140,10,10) | `,` `%` `:` |
| Muton/Chryssalid/Hyperworm | Neon green (20,180,20) | `,` `%` `:` |
| Other aliens | Purple (160,30,200) | `,` `%` `:` |

Splatter appears on the hit tile and 1 adjacent floor tile (33% chance).

## Fire Mechanics

| Property | Value |
|----------|-------|
| Animation | Cycles `^` → `w` → `*` every 4 frames |
| Colors | Yellow → Orange → Red background |
| Spread chance | 20% per turn to adjacent flammable tile |
| Flammable tiles | Grass, Tree, Bush, Fence, Door |
| Duration | 3 turns, then tile becomes Floor |
| Ignition sources | Plasma weapons, explosives, grenades |

## Gas Mechanics

| Density | Char | LOS Block | Cover Penalty |
|---------|------|-----------|---------------|
| 3 (Heavy) | ▓ | Yes | 40% |
| 2 (Medium) | ▒ | No | 20% |
| 1 (Thin) | ░ | No | 0% |

Gas spreads to adjacent tiles and reduces density by 1 each turn until dissipation.

## Destructible Terrain

| Tile | Converts To | Cover Before | Cover After |
|------|-------------|-------------|-------------|
| Wall / UFO Wall | Rubble | 80% | 20% |
| Tree | Rubble | 60% | 20% |
| Rock | Rubble | 70% | 20% |
| Fence | Rubble | 30% | 20% |

Grenades destroy terrain within blast radius (2 tiles).
