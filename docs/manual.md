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

The Geoscape uses a **regional dashboard** layout:

- **Left pane (60%):** Table of all regions with status columns
- **Right pane (40%):** ASCII minimap showing node positions and connections

### Region Table Columns

| Column | Description |
|--------|-------------|
| REGION | City name (j/k to select) |
| THREAT | Visual bar: `█` = threat level, `░` = safe |
| RADAR | `R` if you have radar coverage, `-` otherwise |
| SQD | Number of interceptors stationed |
| STATUS | BASE, clear, ALERT, DANGER, or MISSION |

### Minimap Symbols

| Symbol | Meaning |
|--------|---------|
| ◆ | A base (any base node) |
| ◉ | Currently selected node |
| ○ | Regional hub (green=safe, yellow=threat, red=danger) |
| · | Radar coverage (faint dots around each base) |

### Controls

| Key | Action |
|-----|--------|
| j/k | Navigate region list |
| L | Launch interceptor at selected node |
| A | Autoresolve interception |
| M | Respond to alien mission |
| B | Open base management |
| R | Dispatch transport to crash site |
| C | Cycle to the next base |
| N | Build a new base at the selected node ($500K) |
| T | Open the transfer screen (move soldiers/items between bases) |
| Space | Pause/unpause time |
| 1-4 | Time compression |

### Multiple Bases

You can build additional bases to expand radar coverage and split your forces.
Each base is constructed at a regional node for $500,000 and comes with Living
Quarters, Storage, and a Radar facility. Press `N` while the cursor is on an
empty node to build there.

- **Radar coverage** is drawn as a faint `·` ring around every base on the minimap.
- Press `C` to cycle the *active* base. All base screens (management, research,
  manufacture, equipment) operate on the active base.
- Press `T` to open the **Transfer** screen and move soldiers or items between
  bases. Use `Tab` to choose the destination base, `Space` to move the selected
  soldier, and `Enter` to move one unit of the selected item.

### Base Defense

If an alien mission targets a node that hosts one of your bases, responding with
`M` launches a **Base Defense** battle on that base's map. Each base defends with
its own stationed squad. If you **lose** a base defense battle — or let a base
defense mission expire — the base is **destroyed** and all its personnel are lost.
Losing your last remaining base ends the game.

### Time Controls

| Key | Speed | Description |
|-----|-------|-------------|
| Space | Pause | Toggle pause |
| 1 | 1x | 1 minute per tick |
| 2 | 5x | 5 minutes per tick |
| 3 | 20x | 20 minutes per tick |
| 4 | 60x | 60 minutes per tick |

### UFO Interception

UFOs travel along edges between nodes. You can:

1. **Target a specific UFO** — press `L`, interceptor pursues the nearest UFO
2. **Patrol a node** — move cursor to a node, press `L`, interceptor flies there and engages any UFOs

**Interceptor stats:** 60 HP, 8 missiles, speed 36, damage 15–34 per shot.
**UFO retaliation:** 30% chance per tick, 5–14 damage.

### Alien Missions

Every ~30 minutes (game time), alien missions spawn targeting random nodes.
Mission types are weighted so common raids appear more often than rare,
high-value assaults:

| Mission | Response | Map | Reward on victory |
|---------|----------|-----|-------------------|
| Terror | 24h | Urban (Terror) | Standard loot |
| Supply Raid | 24h | UFO Interior | Bonus alloys/elerium/nav data |
| Abduction | 24h | Rural (Abduction) | Rescue civilians |
| Alien Research | 24h | UFO Interior | Bonus alien tech (power/weapon) |
| Council | 36h | Urban (Council) | **Bonus funding +$100K** and loot |
| Alien Base Assault | 12h | Rocky Alien Base | Major alien tech haul |

When multiple missions are active, move the cursor onto a region and press `M`
to respond to the mission at that node; otherwise `M` responds to the first
available mission.

If the timer expires without response, Alien Activity increases by 10%.

Press `M` to respond and deploy your squad.

### UFO Retrieval

After shooting down a UFO, a crash site marker appears at the destination node.
Press `R` to dispatch a transport to the nearest crash site and recover salvage.

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
| Alien Containment | $100K | 10 days | Live alien capture (capacity: 10 per facility) |
| Psi-Lab | $150K | 14 days | Trains psi skill (+1/day, max 80) |
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

### Procedural Tech Tree

Each playthrough generates a unique tech tree from a seeded random algorithm.
The tree always contains core technologies (Alien Alloys, Elerium-115, weapons,
armour), but prerequisites and costs vary between runs. Alien autopsies are
dynamically generated based on the procedural species spawned in your game.

Key mechanics:
- **Tiered DAG:** Technologies are organized in Tiers 1-5. Higher-tier techs
  always require lower-tier prerequisites, guaranteeing no circular dependencies.
- **Dynamic Autopsies:** Each procedural alien species gets a unique autopsy
  topic. Some weapon techs require a specific autopsy as a biological catalyst.
- **Cost Variance:** Non-tier-1 tech costs are multiplied by a random factor
  (0.85x - 1.15x) each run. Laser Weapons might cost 102 one run and 138 the
  next.
- **Reachability:** Every Tier 2+ tech always has at least one prerequisite from
  the tier directly below, guaranteeing the tree is fully reachable from Tier 1.

### Core Technologies (always present)

| Tier | Topic | Base Cost | Unlocks |
|------|-------|-----------|---------|
| 1 | Alien Alloys | 60 | Alloys item |
| 1 | Elerium-115 | 80 | Elerium item |
| 1 | UFO Navigation | 100 | Alien lore |
| 1 | UFO Power Source | 120 | Alien lore |
| 1 | Alien Communications | 90 | Alien lore |
| 1 | [Species] Autopsy | 40-70 | Alien lore |
| 2 | Laser Weapons | 120 | Laser Pistol, Laser Rifle |
| 2 | Personal Armour | 80 | Personal Armour |
| 3 | Plasma Weapons | 200 | Plasma Pistol, Plasma Rifle |
| 3 | Light Suit | 150 | Light Suit |
| 3 | UFO Propulsion | 110 | Alien lore |
| 4 | Heavy Plasma | 250 | Heavy Plasma |
| 4 | Medium Suit | 200 | Medium Suit |
| 4 | Mind Control | 150 | Alien lore |
| 5 | Heavy Suit | 280 | Heavy Suit |
| 5 | Power Suit | 400 | Power Suit |
| 5 | Flying Suit | 500 | Flying Suit |

### Recommended Early Research

Alien Alloys and Elerium-115 are always available at Tier 1 and should be
researched first. After that, check which autopsies your scientists can
perform — some weapon techs require a specific alien autopsy as a catalyst.

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
Cover from objects: × (100 - cover%) / 100
Minimum: 1
```

#### Object Cover

Shots passing through tiles with cover values have their damage reduced.
The highest cover value along the line of fire (excluding shooter and target)
is applied as damage reduction.

| Object | Cover % | Tile Symbol |
|--------|---------|-------------|
| Wall / UFO Wall | 80% | # / █ |
| Rock | 70% | ∩ |
| Tree | 60% | ♣ |
| UFO Furniture | 50% | ░ ⚙ ◈ ⌁ ▤ ⊕ |
| Bush | 40% | † |
| Heavy Smoke | 40% | ▓ |
| Fence | 30% | ║ |
| Medium Smoke | 20% | ▒ |
| Rubble | 20% | ▒ |

**Strategy:** Position soldiers behind walls (80% reduction) for maximum protection.
Trees (60%) are decent野外 cover. Fences (30%) provide minimal protection.

#### Damage Types & Resistance

Each alien species has a primary damage type and unique resistance/weakness
spread. These are discovered as you encounter and autopsy aliens.

| Damage Type | Weapons |
|-------------|---------|
| Plasma | Plasma Pistol, Plasma Rifle, Heavy Plasma, Alien Grenade |
| Laser | Laser Pistol, Laser Rifle |
| Explosive | Rocket Launcher |
| Melee | Chryssalid Claw, Reaper Claw, Stun Rod |
| Kinetic | Pistol, Rifle, Heavy Cannon, Auto Cannon |
| Psionic | Ethereal attacks |

Alien resistance values: positive = damage reduced, negative = damage increased.

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
- **Destruction:** Destroys walls, trees, rocks, fences within blast radius (radius 2),
  converting them to rubble with reduced cover
- **Smoke:** Creates a smoke cloud (density 3 center, density 2 adjacent) that
  expands and thins each turn, blocking LOS at heavy density

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

### Night/Day Missions

Battles occur in real-time. The game hour determines whether it's night:
- **Night:** Before 6:00 or after 18:00
- **Day:** 6:00 to 18:00

Night penalties:
- **Accuracy:** 75% of daytime accuracy
- **Sight range:** Reduced from 20 to 10 tiles
- **Visual effects:** Soldiers emit a warm glow, aliens emit a faint blue glow

### Psi Combat

Requires a **Psi-Amplifier** weapon and the **Psi-Lab** facility.

| Action | TU Cost |
|--------|---------|
| Psi attack | 20 |

**Formula:** `success chance = attacker.PsiSkill - (target.PsiStr / 3)` (min 5%)

Success: Target is **panicked** — loses all TU and skips their next turn. Psi resistance
varies by alien species (Ethereals are highly resistant, Chryssalids have none).

**Training:** Soldiers in a base with a Psi-Lab gain +1 PsiSkill per day (max 80).

**Mind Control research:** Completing this topic grants +20 PsiSkill to all soldiers
at that base.

### Visual Effects

The Battlescape includes dynamic visual effects:
- **Explosions:** Grenade detonations and weapon impacts spawn particle bursts
- **Screen shake:** Camera shakes on explosions (intensity scales with damage)
- **Smoke particles:** Grenade impacts produce lingering smoke particles
- **Night lighting:** Units emit subtle radial glow in dark missions
- **Volumetric gas:** Grenades create expanding smoke clouds (density 3→2→1→dissipate)
  that block LOS at heavy density and provide cover penalties. Diffuses each turn.
- **Destructible terrain:** Grenades destroy walls/trees/rocks in their blast radius,
  converting them to rubble. Rubble particles fly in parabolic arcs on destruction.
- **Vision modes:** Press `V` to cycle Normal → Night Vision → Thermal → Normal.
  - Night Vision: green phosphor overlay with static noise
  - Thermal: living entities glow hot (red/orange/yellow), terrain is cold (dark blue)
- **Blood splatter:** Damage leaves persistent blood decals on floor tiles.
  Humans bleed dark red; Mutons/Chryssalids bleed neon green; others bleed purple.
- **Animated fire:** Plasma and explosive weapons ignite flammable tiles.
  Fire cycles between `^`/`w`/`*` in yellow/orange/red, spreads 20% chance per turn
  to adjacent grass/trees/bushes, and consumes the tile to ash after 3 turns.

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

### Procedural Species

Each game session generates 5–7 unique alien species from a seed. Every species
has 2–5 rank variants (Soldier → Navigator → Commander → Elite → Overlord).

Species traits are determined at generation:
- **Primary damage type** — the damage type the species deals
- **Resistance spread** — each species has unique resistances and weaknesses
- **Weapon preference** — lower ranks use pistols, higher ranks use heavy weapons

This means every playthrough features different alien threats. One run may have
a psionic-heavy species resistant to plasma, while another has melee predators
weak to explosives.

### Knowledge Levels

As you encounter aliens, your knowledge increases:

| Level | Trigger | Effect |
|-------|---------|--------|
| 0 — Unknown | Never seen | Name appears as "???" in encyclopedia |
| 1 — Sighted | Alien visible in FOV | Name and icon revealed |
| 2 — Killed | Alien killed in combat | Stats and resistances revealed |
| 3 — Autopsied | Research completed | Full lore and detailed weaknesses |

### Alien AI Behavior

- **Patrol:** Wander until a human is within 15 tiles, then switch to Attack.
- **Attack:** Fire at target. High-aggression aliens (>5) charge if distance > 3 tiles.
- **Search:** Move toward last-seen position for 5 turns, then resume Patrol.
- **Flee:** Triggered when HP < 25% AND Bravery < 50. Runs for 3 turns.
- **Adaptive:** Across missions the aliens study your habits (stored in `Game.Tactics`).
  If you snipe from long range they rush to close distance; if you lean on grenades
  they spread out to avoid clusters; if you flank often they post more suppressors
  to pin you; if they are losing badly they retreat sooner, and if they dominate
  they fight on aggressively.

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
active missions, procedural species seed, and alien knowledge levels.

The species seed ensures the same alien species are regenerated when loading a save.

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
| R | Dispatch transport to crash site |
| E | Open encyclopedia |
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
| v | Toggle vision mode (Normal / Night / Thermal) |
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

- **Use object cover** — position behind walls (80% reduction) or trees (60%).
- **Crouch before firing** — +10% accuracy and 30% damage reduction.
- **Reload early** — don't wait until empty.
- **Grenades** bypass cover — useful against enemies behind walls.
- **Learn alien weaknesses** — each species has unique resistances. Use the right weapon.
- **Medikits** — keep one on a dedicated medic. 25 TU is expensive, but saves lives.
- **Don't overextend** — advance cautiously; aliens get reaction shots.

### Base Building

- **Radar** facilities are worth it: +$50K monthly funding each.
- **Storage** is essential — you'll fill up fast with corpses and alloys.
- **Alien Containment** is needed for live captures (research bonuses).
  Each facility holds 10 captured aliens. Build multiple for larger rosters.
- Build **Hangars** to field multiple interceptors.
- **Psi-Lab** trains psi skill (+1/day, max 80). Build it early if you want psi
  capabilities. Mind Control research grants +20 PsiSkill to all soldiers.

### Alien Capture & Interrogation

Use the **Stun Rod** (melee weapon, $2K) to stun aliens instead of killing them.
When an alien's stun points exceed their HP, they fall unconscious and can be
collected after the mission — provided you have an Alien Containment facility
with available capacity.

Captured aliens are listed in the Research screen. Press **I** to interrogate
a captured alien:
- If the matching autopsy topic is active, interrogation completes it instantly.
- If the autopsy is not yet started, interrogation auto-completes it and
  grants all associated unlocks.
- If the autopsy is already done, interrogation grants a 25% progress bonus
  to your current active research topic.
- Interrogation requires at least one Laboratory.

### Research Priorities

1. Alien Alloys → Laser Weapons (always Tier 1 → Tier 2)
2. Personal Armour (always Tier 2)
3. Autopsies of encountered species (Tier 1 — unlocks alien lore and may gate weapon techs)
4. Elerium-115 → Plasma Weapons (check autopsy requirement)
5. Mid-game: Medium Suit, Heavy Plasma, UFO Propulsion
6. Late-game: Power Suit, Flying Suit, Mind Control

### Economy

- Sell alien corpses and excess loot for quick cash.
- Monthly salary costs add up — balance soldiers vs. income.
- $50K reward per battle win helps offset expenses.
- Early game is tight on funds — manufacture and sell items for profit.
