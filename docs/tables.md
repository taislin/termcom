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

### Hardcoded Weapons

| Key | Name | Damage Type | Notable Traits |
|-----|------|-------------|----------------|
| pistol | Pistol | kinetic | Starting sidearm |
| rifle | Rifle | kinetic | Standard issue |
| heavy | Heavy Cannon | explosive | High damage, slow |
| auto | Auto Cannon | explosive | 3-round burst |
| rocket | Rocket Launcher | explosive | Devastating, 1 ammo |
| laser_pistol | Laser Pistol | laser | Unlimited ammo |
| laser_rifle | Laser Rifle | laser | Unlimited ammo |
| plasma_rifle | Plasma Rifle | plasma | Alien tech, high damage |
| plasma_pistol | Plasma Pistol | plasma | Alien sidearm |
| heavy_plasma | Heavy Plasma | plasma | Best weapon in game |
| chryssalid_claw | Chryssalid Claw | melee | Alien melee |
| reaper_claw | Reaper Claw | melee | Boss melee |
| stun_rod | Stun Rod | melee | Non-lethal |
| medi_kit | Medi-Kit | special | Heals 10 HP |

### Procedural Weapons

Each playthrough generates 2-3 unique weapons based on the procedural species'
damage types. Generated in `internal/data/procedural_items.go`.

| Property | Range | Description |
|----------|-------|-------------|
| Damage | 20-60 | Varies by damage type |
| Accuracy | 55-85% | Base accuracy |
| TU | 15-30 | Time units to fire |
| Range | 10-30 | Tiles |
| Ammo | 6-45 | Depends on burst mode |
| Burst | 1 or 3 | 33% chance of burst weapon |

Weapon names combine damage-type prefixes (Plasma, Laser, Rail, Psi, etc.) with
weapon suffixes (Pistol, Rifle, Carbine, Blaster, Cannon, Emitter).

## Armor

Armor stats are defined in `internal/data/items.go`.

### Hardcoded Armor

| Key | Name | Undersuit | TU Modifier |
|-----|------|-----------|-------------|
| none | None | 0 | 0 |
| personal | Personal Armour | 10 | 0 |
| light | Light Suit | 20 | -5 |
| medium | Medium Suit | 30 | -10 |
| heavy | Heavy Suit | 40 | -15 |
| power_suit | Power Suit | 50 | -10 |
| flight_suit | Flying Suit | 45 | -5 |

### Procedural Armor

Each playthrough generates 1-2 unique armor pieces based on the procedural species'
damage types. Generated in `internal/data/procedural_items.go`.

| Property | Range | Description |
|----------|-------|-------------|
| Undersuit | 15-45 | Base armor value |
| Health | 0-15 | Bonus HP |
| TU Modifier | -5 to -15% | Movement penalty |
| Value | $20K-$60K | Sell value |

Armor names combine damage-type prefixes (Plasma-Shielded, Reflective, Ballistic,
Psi-Shielded, etc.) with armor suffixes (Vest, Suit, Plating, Armour, Guard).

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

## Mission Modifiers

| Modifier | Description | Effect |
|----------|-------------|--------|
| Night Ops | Forced night battle | -25% accuracy, +20% loot |
| Reinforcements | Extra alien wave | +2 aliens on turn 4 |
| Time Limit | Turn limit | Defeat if >15 turns with aliens alive |
| VIP Rescue | Protect civilian | +$50K bonus if VIP survives |
| Booby Trapped | More explosives | Extra grenades/mines on map |
| Heavy Fog | Reduced visibility | -40% sight range |
| Alien Ambush | Pre-positioned aliens | Aliens start in overwatch |
| Low Visibility | Poor conditions | -10% accuracy for all |
| High Ground | Elevated terrain | Accuracy bonus from height |

## Weather Effects

| Weather | Accuracy | Sight | Fire Spread |
|---------|----------|-------|-------------|
| Clear | 0% | 0 | 20% |
| Rain | -5% | -2 | 5% |
| Wind | 0% | 0 | 30% |
| Snow | -3% | -2 | 20% |
| Fog | -5 to -10% | -3 to -6 | 20% |
| Storm | -5% | -2 | 30% |
| Cold | -3% | 0 | 20% |

## Soldier Perks

| Perk | Stat Bonus | Battle Effect |
|------|------------|---------------|
| Lightning Reflexes | +10 Reactions | — |
| Marksman | — | +15% accuracy at range > 8 |
| Grenadier | — | +2 grenade splash radius |
| Field Medic | — | Medikit heals 15 HP |
| Iron Will | +10 PsiSkill | +20 Psi Strength |
| Steady Aim | — | +10% accuracy when not moving |
| Close Combat Specialist | — | +15% accuracy at range ≤ 4 |
| Overwatch Expert | — | +20% reaction fire accuracy |
| Demolitions | — | +50% grenade damage |
| Scavenger | — | +25% loot from battles |
| Tough | +5 MaxHP | — |
| Quick Learner | — | +50% XP from battles |

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
