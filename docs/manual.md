# ycom — ASCII X-COM Demake Manual

## Table of Contents

1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [Geoscape](#geoscape)
4. [Base Management](#base-management)
5. [Research](#research)
6. [Manufacturing](#manufacturing)
7. [Equipping Soldiers](#equipping-soldiers)
8. [Battlescape](#battlescape)
9. [Weapons & Equipment](#weapons--equipment)
10. [Armour](#armour)
11. [Aliens](#aliens)
12. [Soldier Ranks & Stats](#soldier-ranks--stats)
13. [Save/Load](#saveload)
14. [Key Bindings Reference](#key-bindings-reference)
15. [Tips & Strategy](#tips--strategy)

---

## Overview

**ycom** is a faithful ASCII demake of X-COM: UFO Defense (1994), rendered entirely
in a terminal. You command the X-COM organization — an international effort to
combat an alien invasion.

**Your mission:** Research alien technology, manufacture advanced weapons and armour,
and lead squads into tactical combat to eliminate the alien threat.

**Victory condition:** Win 10 battles.

**Defeat condition:** Alien Activity reaches 100%.

**Starting resources:**
- $500,000
- 10 scientists, 10 engineers
- Starting base with Living Quarters, Laboratory, Workshop, Storage, and Radar
- Several Rifles and Pistols

---

## Getting Started

Run the game:
```bash
go run ./cmd/ycom
# or
make run
```

You begin on the **Geoscape** — the world map view. Time advances as you watch UFOs
appear and move across the globe. Your radar detects them as they come into range.

When a UFO is detected, you can:
- **Launch interceptor** (`L`) to shoot it down
- **Autoresolve** (`A`) for an automatic engagement
- **Respond to missions** (`M`) when alien terror or supply missions appear

After shooting down a UFO or engaging an alien mission, you deploy your soldiers for
**tactical combat** on the Battlescape.

---

## Geoscape

The Geoscape is the main strategic view. You see a world map with your base and
detected UFOs.

### Time Controls

| Key | Speed | Description |
|-----|-------|-------------|
| Space | Pause | Toggle pause |
| 1 | 1x | 1 minute per tick |
| 2 | 5x | 5 minutes per tick |
| 3 | 20x | 20 minutes per tick |
| 4 | 60x | 60 minutes per tick |

### UFO Interception

1. Wait for radar to detect a UFO (appears on map)
2. Press `L` to launch your interceptor at the nearest UFO
3. The interceptor pursues and fires Avalanche missiles
4. If the UFO is destroyed, you earn salvage value
5. UFOs fire back — your interceptor can be destroyed

**Interceptor stats:** 60 HP, 8 missiles, speed 36, damage 15–34 per shot.
**UFO retaliation:** 30% chance per tick, 5–14 damage.

### Alien Missions

Every ~30 minutes (game time), alien missions spawn targeting major cities:
London, Tokyo, New York, Moscow, Sydney, Paris, Berlin.

- **Terror missions** and **Supply missions**: 5 minutes to respond
- **Alien Base missions**: 3 minutes to respond

If the timer expires without response, Alien Activity increases by 10%.

Press `M` to respond and deploy your squad.

### Monthly Budget

Each month (3 game days):
- **Income:** $200,000 base + $50,000 per Radar facility
- **Expenses:** $2,000 per soldier, scientist, and engineer

Keep your budget positive — running out of funds means you can't hire or manufacture.

---

## Base Management

Press `B` from the Geoscape to open your base.

### Tabs

| Key | Tab |
|-----|-----|
| 1 | Facilities |
| 2 | Soldiers |
| 3 | Research |
| 4 | Manufacture |
| 5 | Transfer |

Navigate with `j`/`k` or arrow keys. Use `Enter` to select.

### Facilities

| Facility | Cost | Build Time | Effect |
|----------|------|------------|--------|
| Living Quarters | $50K | 5 days | +8 soldier capacity |
| Laboratory | $75K | 7 days | Enables research |
| Workshop | $60K | 7 days | Enables manufacturing |
| Storage | $40K | 3 days | +50 item storage |
| Radar | $80K | 5 days | +$50K monthly funding |
| Alien Containment | $100K | 10 days | Live alien capture |
| Psi-Lab | $150K | 14 days | Psychic training |
| Hangar | $120K | 8 days | Houses interceptors |

**Controls:**
- `B` — Build selected facility
- `S` — Sell selected facility (50% refund)

### Soldiers

Hire soldiers at $50,000 each. Capacity limited by Living Quarters.

- `H` — Hire soldier
- `E` — Open equip screen
- `D` — Dismiss soldier

---

## Research

From the Research tab, assign scientists to research topics. Research progresses
automatically as time runs.

### Research Tree

| Topic | Man-Days | Prerequisites | Unlocks |
|-------|----------|---------------|---------|
| Alien Alloys | 60 | — | Aluminium Alloys item |
| Elerium-115 | 80 | — | Elerium item |
| Sectoid Autopsy | 40 | — | Alien lore |
| Floater Autopsy | 50 | — | Alien lore |
| Muton Autopsy | 60 | — | Alien lore |
| Ethereal Autopsy | 80 | Sectoid + Floater autopsy | Alien lore |
| Laser Weapons | 120 | Alien Alloys | Laser Pistol, Laser Rifle |
| Plasma Weapons | 200 | Elerium + Sectoid Autopsy | Plasma Pistol, Plasma Rifle |
| Heavy Plasma | 250 | Plasma Weapons + Muton Autopsy | Plasma Rifle |
| Personal Armour | 80 | Alien Alloys | Personal Armour |
| Light Suit | 150 | Personal Armour + Alien Alloys | Light Suit |
| Medium Suit | 200 | Light Suit | Medium Suit |
| Heavy Suit | 280 | Medium Suit | Heavy Suit |
| Power Suit | 400 | Heavy Suit + Elerium | Power Suit |
| Flying Suit | 500 | Power Suit | Flying Suit |
| Mind Control | 150 | Ethereal Autopsy | Alien lore |
| UFO Navigation | 100 | — | Alien lore |
| UFO Power Source | 120 | — | Alien lore |
| Alien Communications | 90 | — | Alien lore |

**Recommended early research:** Alien Alloys → Laser Weapons → Personal Armour → Sectoid Autopsy → Elerium → Plasma Weapons.

---

## Manufacturing

From the Manufacture tab, assign engineers to produce items.

### Manufacturable Items

| Item | Time | Materials |
|------|------|-----------|
| Pistol | 3 days | 1 alloy |
| Rifle | 5 days | 2 alloys |
| Heavy Cannon | 7 days | 3 alloys |
| Auto Cannon | 6 days | 3 alloys |
| Rocket Launcher | 8 days | 4 alloys, 1 elerium |
| Stun Rod | 2 days | 1 alloy |
| Personal Armour | 6 days | 2 alloys |
| Light Suit | 10 days | 4 alloys, 1 elerium |
| Medium Suit | 14 days | 6 alloys, 2 elerium |
| Heavy Suit | 18 days | 8 alloys, 3 elerium |
| Medi-Kit | 3 days | 1 alloy |

Manufacturing time scales with quantity: `5 + count × 2` days. More engineers
speed up production.

**Note:** Plasma weapons and Power/Flying Suits cannot be manufactured — they must
be researched and recovered from alien corpses.

---

## Equipping Soldiers

From the Soldiers tab, press `E` to open the equip screen.

### Controls

| Key | Action |
|-----|--------|
| j/k | Select soldier |
| Tab | Cycle available items |
| 1 | Select weapon slot |
| 2 | Select armour slot |
| Space | Equip selected item |
| Esc | Back |

Each soldier has a weapon slot and an armour slot. You can only equip items you
have manufactured or recovered.

---

## Battlescape

The Battlescape is the tactical combat layer. You control a squad of soldiers
against alien forces.

### Turn Structure

1. **Player Turn** — You spend Time Units (TU) to move, shoot, reload, etc.
2. **Alien Turn** — Aliens act using their TU pool and AI.
3. Repeat until one side is eliminated.

### Victory & Defeat

- **Victory:** All aliens eliminated. Soldiers earn XP, you recover loot, and
  receive $50,000.
- **Defeat:** All soldiers killed. Soldiers are lost.

### Time Units (TU)

Every action costs TU from your soldier's pool. TU restore fully at the start of
each player turn.

| Action | TU Cost |
|--------|---------|
| Move (per tile) | 4 (+4 if crouching) |
| Fire weapon | 15 |
| Reload | 8 |
| Crouch | 4 |
| Throw grenade | 20 |
| Use medikit | 25 |

TU pool varies by rank: 45–55 at Rookie, increasing with promotions.

### Combat

#### Hit Chance

```
distance = tiles to target
accMod = 100 - (distance × 3), minimum 10
hitChance = (attacker.Accuracy × accMod) / 100
Crouch bonus: × 110 / 100
```

#### Damage

```
damage = weapon.Damage + random(0 to weapon.Damage / 3)
damage -= target.Armour value
Crouching: × 70 / 100
Minimum: 1
```

### Line of Sight

Uses Bresenham's line algorithm. Blocked by opaque tiles:
- Walls, Trees, Rocks, UFO Walls

Passable (transparent) tiles:
- Floors, Doors, Grass, UFO Floors

### Grenades

- Range: 6 tiles (Euclidean)
- Base damage: `40 + Strength × 2`
- Splash: enemies within distance² ≤ 4 of impact
- Splash damage: `base - (distance² × 5)`, minimum 5

### Medikit

Heals 10 HP per use, costs 25 TU. Must target a friendly unit.

### Out-of-Battle Healing

- Wounded soldiers heal 2 HP per day
- Wounds decrement by 1 per day
- Max wound time: 30 days

### Crouching

- Costs 4 TU to crouch, free to stand
- +10% accuracy bonus
- 30% damage reduction when hit

### Map Types

| Map | Description |
|-----|-------------|
| Crash Site | Outdoor terrain with UFO wreckage |
| Terror Site | Urban map with buildings |
| Supply (UFO Interior) | Inside a UFO, multiple rooms |
| Alien Base | Rocky terrain with alien structure |

---

## Weapons & Equipment

### Weapons

| Weapon | Code | DMG | ACC | TU | Range | Ammo | Notes |
|--------|------|-----|-----|-----|-------|------|-------|
| Pistol | PIS | 15 | 65% | 15 | 8 | 12 | Starting weapon |
| Rifle | RIF | 22 | 70% | 20 | 20 | 20 | Standard issue |
| Heavy Cannon | HVC | 35 | 55% | 25 | 15 | 6 | High damage |
| Auto Cannon | AUC | 20 | 60% | 25 | 18 | 18 | 3-round burst |
| Rocket Launcher | RKT | 80 | 45% | 30 | 30 | 1 | Devastating |
| Laser Pistol | LSP | 28 | 75% | 12 | 12 | ∞ | Energy weapon |
| Laser Rifle | LSR | 40 | 80% | 18 | 25 | ∞ | Energy weapon |
| Plasma Rifle | PLR | 55 | 75% | 22 | 28 | ∞ | Alien tech |
| Plasma Pistol | PLP | 30 | 70% | 14 | 10 | ∞ | Alien tech |
| Stun Rod | STR | 10 | 90% | 20 | 1 | ∞ | Melee only |

### Ammo & Energy Weapons

- **Ammo-based** weapons (Pistol, Rifle, Heavy Cannon, Auto Cannon, Rocket) need
  reloading — press `R` in combat.
- **Energy weapons** (Laser, Plasma) have unlimited ammo — no reload needed.

### Items

| Item | Code | Weight | Value |
|------|------|--------|-------|
| Aluminium Alloys | ALY | 2 | $8,000 |
| Elerium-115 | ELR | 3 | $12,000 |
| Alien Corpse | ALC | 10 | $2,000 |
| Sectoid Corpse | SEC | 10 | $3,000 |
| Floater Corpse | FLT | 10 | $4,000 |
| Muton Corpse | MUT | 15 | $6,000 |
| Ethereal Corpse | ETH | 10 | $8,000 |
| Alien Grenade | AGR | 1 | $4,000 |
| Medi-Kit | MDK | 2 | $6,000 |
| Motion Scanner | MSC | 3 | $5,000 |
| Psi-Amplifier | PSI | 2 | $30,000 |

**Storage:** Weapons weigh 5 per unit, armour weighs 8. Capacity = 50 per Storage facility.

---

## Armour

| Armour | Code | Defence | TU Penalty | Cost |
|--------|------|---------|------------|------|
| None | --- | 0 | 0% | — |
| Personal Armour | PSA | 10 | 0% | $15,000 |
| Light Suit | LIS | 20 | -5% | $35,000 |
| Medium Suit | MDS | 30 | -10% | $55,000 |
| Heavy Suit | HVS | 40 | -15% | $75,000 |
| Power Suit | PWS | 50 | -10% | $100,000 |
| Flying Suit | FLS | 45 | -5% | $140,000 |

Higher armour reduces damage but imposes a TU penalty, reducing actions per turn.

---

## Aliens

### Alien Types

| Alien | HP | TU | ACC | BRA | Armour | Weapon |
|-------|----|----|-----|-----|--------|--------|
| Sectoid | 10 | 50 | 55% | 40 | 5 | Plasma Pistol |
| Sectoid Leader | 12 | 55 | 60% | 50 | 8 | Plasma Rifle |
| Floater | 15 | 55 | 60% | 50 | 10 | Plasma Rifle |
| Floater Leader | 18 | 60 | 65% | 60 | 12 | Plasma Rifle |
| Muton | 25 | 55 | 55% | 70 | 18 | Plasma Rifle |
| Muton Leader | 28 | 60 | 60% | 80 | 20 | Plasma Rifle |
| Ethereal | 18 | 65 | 70% | 100 | 12 | Plasma Rifle |
| Ethereal Leader | 22 | 70 | 75% | 100 | 15 | Plasma Rifle |

### Alien AI Behavior

- **Patrol:** Wander until a human is within 15 tiles, then switch to Attack.
- **Attack:** Fire at target. High-aggression aliens (>5) charge if distance > 3 tiles.
- **Search:** Move toward last-seen position for 5 turns, then resume Patrol.
- **Flee:** Triggered when HP < 25% AND Bravery < 50. Runs for 3 turns.

Aliens spawn in groups: 3 lowest-rank + 2 rank-1 aliens (if available).

---

## Soldier Ranks & Stats

### Ranks

| Rank | Kills Required |
|------|----------------|
| Rookie | 0 |
| Squaddie | 10 |
| Corporal | 25 |
| Sergeant | 50 |
| Lieutenant | 80 |
| Captain | 120 |
| Major | 170 |
| Colonel | 230 |

### Per Rank-Up Bonuses

+2 HP, +1 MaxTU, +2 Accuracy, +1 Strength, +1 Reactions

### XP Earning

`(alien_kills × 5) + 10` (if battle won), applied to all surviving deployed soldiers.

### New Soldier Stat Ranges

| Stat | Min | Max |
|------|-----|-----|
| HP | 20 | 25 |
| TU | 45 | 55 |
| Accuracy | 40 | 60 |
| Bravery | 30 | 70 |
| Reactions | 30 | 50 |
| Strength | 10 | 20 |
| Psi Strength | 0 | 39 |

### Combat Stats

| Stat | Description |
|------|-------------|
| HP | Health points — reach 0 and the soldier dies |
| TU | Time Units — pool for all actions each turn |
| Accuracy | Base hit chance % (modified by distance, crouching) |
| Bravery | Determines if alien flees when low HP |
| Reactions | Used for reaction shots |
| Strength | Affects melee and grenade damage |
| Psi Skill | Psychic ability |

---

## Save/Load

| Key | Action |
|-----|--------|
| F5 | Save game to `xcom_save.json` |
| F9 | Load game from `xcom_save.json` |

Saves include: game time, funds, pausing, speed, alien activity, base state, UFOs,
and active missions.

---

## Key Bindings Reference

### Geoscape

| Key | Action |
|-----|--------|
| Arrow keys | Move camera |
| Space | Pause/unpause |
| 1–4 | Time speed |
| B | Open base |
| L | Launch interceptor |
| A | Autoresolve nearest UFO |
| M | Respond to mission |
| F5 | Save |
| F9 | Load |
| Q | Quit |
| ? | Help |

### Base Management

| Key | Action |
|-----|--------|
| 1–5 | Switch tabs |
| j/k | Navigate items |
| B | Build facility |
| S | Sell facility |
| H | Hire soldier |
| E | Equip screen |
| D | Dismiss soldier |
| Esc | Back to geoscape |

### Battlescape

| Key | Action |
|-----|--------|
| hjkl / Arrows | Move cursor |
| Space / Enter | Select / Confirm |
| s | Cycle soldiers |
| f | Fire weapon |
| r | Reload |
| e / n | End turn |
| c | Crouch |
| g | Throw grenade |
| m | Use medikit |
| ? | Help |

---

## Tips & Strategy

### Early Game

1. Research **Alien Alloys** first — unlocks Laser Weapons and Armour.
2. Build a second **Radar** to increase detection range and monthly funding.
3. Hire 2–4 extra soldiers to fill your squads.
4. Manufacture **Laser Rifles** as soon as researched — they outclass ballistics
   and never need reloading.

### Combat Tips

- **Crouch before firing** — +10% accuracy and 30% damage reduction.
- **Reload early** — don't wait until empty.
- **Use cover** — stand behind walls and trees to block line of sight.
- **Grenades** are powerful against grouped aliens, especially high-STR soldiers.
- **Medikits** — keep one on a dedicated medic. 25 TU is expensive, but saves lives.
- **Don't overextend** — advance cautiously; aliens get reaction shots.

### Base Building

- **Radar** facilities are worth it: +$50K monthly funding each.
- **Storage** is essential — you'll fill up fast with corpses and alloys.
- **Alien Containment** is needed for live captures (research bonuses).
- Build **Hangars** to field multiple interceptors.

### Research Priorities

1. Alien Alloys → Laser Weapons
2. Personal Armour → Light Suit
3. Sectoid/Floater Autopsy (for lore and prerequisites)
4. Elerium → Plasma Weapons
5. Mid-game: Medium/Heavy Suit, Heavy Plasma
6. Late-game: Power Suit, Flying Suit, Mind Control

### Economy

- Sell alien corpses and excess loot for quick cash.
- Monthly salary costs add up — balance soldiers vs. income.
- $50K reward per battle win helps offset expenses.
- Early game is tight on funds — manufacture and sell items for profit.
