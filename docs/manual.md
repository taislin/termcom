# termcom — ASCII X-COM Roguelike-ified Demake Manual (v0.51.18)

## Table of Contents

1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [Tutorial / Onboarding](#tutorial--onboarding)
4. [Geoscape](#geoscape)
5. [Base Management](#base-management)
6. [Research & Manufacturing](#research--manufacturing)
7. [Equipping Soldiers](#equipping-soldiers)
8. [Battlescape](#battlescape)
9. [Weapons & Equipment](#weapons--equipment)
10. [Armour](#armour)
11. [Aliens](#aliens)
12. [Soldier Ranks & Progression](#soldier-ranks--progression)
13. [Save/Load](#saveload)
14. [Key Bindings Reference](#key-bindings-reference)
15. [Tips & Strategy](#tips--strategy)

---

## Overview

**termcom** is a roguelike-ified ASCII demake of X-COM: UFO Defense (1994), rendered
entirely in a terminal. You command X-COM — an international task force defending
Earth from an alien invasion.

**Your goal:** Research alien technology, manufacture weapons and armour,
and lead squads into tactical combat to eliminate the alien threat.

**Victory:** Win enough battles to trigger the Cydonia final mission, then win it.

**Defeat:** Alien Activity reaches 100% — the invasion overwhelms Earth.

**Difficulty:** Choose a level before starting. Higher difficulties make aliens
tougher, UFOs more frequent, and starting funds tighter.

- **Beginner** — Weaker aliens, slower UFOs, more starting funds
- **Experienced** — Standard
- **Veteran** — Tougher aliens, faster UFOs, less cash
- **Genius** — Much harder across the board
- **Superhuman** — Maximum alien threat

**Language:** 8 languages available — switch in the Options screen.
English, Chinese, Spanish, French, Russian, Portuguese, Japanese, Korean.

**Options:** Press `?` on any screen to open help, or navigate to the Options
screen to adjust bloom, lighting, sound, autosave, screen shake, mouse support,
grid lines, confirm dialogs, theme, resolution speed, volume, and language.

---

## Tutorial / Onboarding

On your first playthrough (no save files detected), a step-by-step Commander
**Briefing** appears automatically after you select your difficulty. It covers:

1. **Welcome** — Introduction to X-COM
2. **Geoscape & Time** — Pause (`Space`), speed (`1`–`4`)
3. **UFO Detection** — Radar and UFO markers
4. **Interceptor Launch** — Press `L` to engage
5. **Mission Response** — Press `M` to deploy
6. **Base Management** — Press `B` to manage your base
7. **Battlescape** — Time Units, movement, and combat
8. **Done** — You're ready

**Controls:** Enter advances, S skips, Esc dismisses.

**Replay:** Open the Options screen and select "Replay Tutorial" at any time.

---

## Getting Started

You begin on the **Geoscape** — the world map. Time advances automatically.
UFOs appear on radar as they come into range.

**Starting resources:**
- $500,000 (modified by difficulty)
- 10 scientists, 10 engineers
- A base with Living Quarters, Laboratory, Workshop, Storage, and Radar
- Several rifles and pistols

**When a UFO is detected:**

| Action | Key | What happens |
|--------|-----|-------------|
| Launch interceptor | `L` | Send a fighter to shoot it down |
| Dispatch transport | `R` | Send troops to investigate a crash site |
| Autoresolve | `A` | Quick automatic interception result |
| Respond to mission | `M` | Deploy to alien terror/supply missions |

---

## Geoscape

The Geoscape shows a **regional dashboard** with threat levels for each region:

- **Left pane:** list of regions with threat bars and radar status
- **Right pane:** ASCII minimap showing bases, UFOs, interceptors, and routes

### Minimap Symbols

| Symbol | Meaning |
|--------|---------|
| ◆ | Your base |
| ◉ | Currently selected node |
| ○ | Regional hub (green=safe, yellow=threat, red=danger) |
| · | Radar coverage ring |
| ! | UFO (red, bold) |
| > | Interceptor patrolling |
| ► | Interceptor engaging a UFO |
| ✕ | Destroyed interceptor or UFO |
| * | Crash site (yellow=unlooted, gray=looted) |
| ≈ | Transport en route (green) |

### Time Controls

| Key | Speed | Use |
|-----|-------|-----|
| Space | Pause | Stop time to plan |
| 1 | 1x | Slow advance |
| 2 | 5x | Normal patrol speed |
| 3 | 20x | Fast-forward |
| 4 | 60x | Maximum speed |

### Monthly Budget

- **Income:** $200,000 base + $50,000 per Radar facility
- **Expenses:** $2,000 per soldier, scientist, and engineer

### Multiple Bases

Press `N` on an empty node to build a new base ($500K). Each base has its own
facilities, soldiers, and stores. Press `C` to cycle the active base.
Press `T` to open the Transfer screen and move soldiers or items between bases.

### Mission Response

When a mission appears (terror, supply raid, abduction, etc.), press `M` to respond:

| Option | Result |
|--------|--------|
| **Deploy squad** | Full tactical combat — best rewards, highest risk |
| **Auto-resolve** | Quick outcome — reduced XP, no corpses, but safe |
| **Ignore** | Skip it — alien activity rises |

Auto-resolve gives about half the XP of a real fight, no alien corpses,
and a small chance of casualties on loss.

### Base Defense

If a mission targets a node with your base, responding launches a **Base Defense**
battle. Losing a base defense destroys the base and its personnel.
Losing your last base ends the game.

### UFO Interception

Press `L` to launch an interceptor at the nearest UFO. The interceptor pursues
the target across the geoscape. When it reaches the UFO, a dedicated **Dogfight
Screen** opens with turn-based combat:

- **[F] Fire** — fires your interceptor's weapon, consuming 1 ammo. The UFO
  fires back immediately. Repeat until one side is destroyed or you break off.
- **[M] Mode** — cycle between Attack (+10% accuracy, closer range), Cautious
  (balanced), and Breakoff (auto-disengage if heavily damaged).
- **[B] / [Esc] Break Off** — recall the interceptor to base.

Each hit depletes a portion of the target's HP; ammo is tracked per interceptor
and **automatically rearmed** when the craft returns to base. The dogfight
resolves when the UFO is destroyed (creating a crash site for tactical battle),
the interceptor is destroyed, or the player breaks off.

### Alien Missions

Missions appear every ~30 game-minutes with a timer of 12–36 hours:

| Mission | Timer | What to expect |
|---------|-------|----------------|
| Terror | 24h | Urban map, many civilians in danger |
| Supply Raid | 24h | UFO interior, bonus alloys/elerium |
| Abduction | 24h | Rural map, rescue civilians |
| Alien Research | 24h | UFO interior, bonus alien tech |
| Council | 36h | Urban map, bonus $100K funding |
| Alien Base Assault | 12h | Rocky alien base, major tech haul |

Letting a mission expire increases Alien Activity by 10%.

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
| 6 | Hangars |

### Facilities

| Facility | Cost | Build Time | Effect |
|----------|------|------------|--------|
| Living Quarters | $50K | 5 days | +8 soldier capacity |
| Laboratory | $75K | 7 days | Enables research |
| Workshop | $60K | 7 days | Enables manufacturing |
| Storage | $40K | 3 days | +50 item storage |
| Radar | $80K | 5 days | +$50K monthly funding |
| Alien Containment | $100K | 10 days | Hold up to 10 live aliens |
| Psi-Lab | $150K | 14 days | Train psi ability |
| Hangar | $120K | 8 days | Houses one interceptor |

**Adjacency bonuses:** Placing same-type facilities next to each other helps:
- Adjacent Labs: faster research (up to +30%)
- Adjacent Workshops: faster manufacturing (up to +30%)
- Adjacent Living Quarters: faster soldier healing (up to +3 HP/day)

**Controls:** `B` to build, `S` to sell (50% refund).

### Soldiers Tab

Hire soldiers at $50K each.

| Key | Action |
|-----|--------|
| H | Hire soldier |
| E | Open equip screen |
| G | Open weapon designer |
| D | Dismiss soldier |

### Hangars Tab

Each Hangar holds one interceptor. Manage your air force here.

| Key | Action |
|-----|--------|
| B | Buy interceptor |
| W | Equip weapon |
| G | Open weapon designer |
| D | Open plane designer |

---

## Research & Manufacturing

### Research

From the Research tab, assign scientists to topics. Research progresses
automatically as game time passes.

The tech tree is **procedural** — each playthrough generates a unique tree
from a seeded algorithm. Core technologies (Laser Weapons, Personal Armour,
Plasma Weapons) are always present, but prerequisites and costs vary.

**Priorities:**
- **Alien Alloys** and **Elerium-115** should be researched first
- Autopsies of alien species unlock lore and may gate weapon techs
- Interrogate captured aliens (`I` key) to complete research faster

### Manufacturing

From the Manufacture tab, assign engineers to produce items.
This produces **stock weapons and armour** — for more powerful custom gear,
use the Weapon Designer and Plane Designer.

**Manufacturable items** (build multiple at once):

| Item | Time | Materials |
|------|------|-----------|
| Pistol, Rifle, Heavy Cannon, Auto Cannon | 3–8 days | Alloys |
| Rocket Launcher, Stun Rod | 2–8 days | Alloys + Elerium |
| Personal Armour, Light/Medium/Heavy/Power/Flying Suit | 6–18 days | Alloys + Elerium |
| Medi-Kit | 3 days | Alloys |

More engineers = faster production. Energy weapons (Laser, Plasma) cannot be
manufactured — they must be researched, designed, and built through the
Weapon Designer, or recovered from aliens.

---

## Equipping Soldiers

From the Soldiers tab, press `E` to open the equip screen.

### Controls

| Key | Action |
|-----|--------|
| ↑/↓ | Select soldier |
| Tab | Cycle available items |
| 1 | Weapon slot |
| 2 | Armour slot |
| 3 | Inventory slot |
| Space | Equip selected item |
| G | Open weapon designer |
| A | Auto-equip all soldiers |
| Esc | Back |

### Slots

- **Slot 1 (Weapon):** Main weapon — custom-designed or a stock rifle/pistol
- **Slot 2 (Armour):** Body armour — personal, light suit, etc.
- **Slot 3 (Inventory):** Extra items — grenades, medikits, scanners,
  proximity mines, psi-amps, melee weapons

### Encumbrance

Every item has weight. The total weight of weapon + armour + inventory
is your **encumbrance**. Higher encumbrance reduces your Time Units in battle
(roughly 1 TU penalty per 5 weight units). Keep your soldiers lightly loaded
for maximum mobility.

### Auto-Equip

Press `A` to automatically equip every soldier with the best available weapon
and armour from stores. Existing gear is returned to stores — a fast way to
regear your squad after researching new tech.

---

## Battlescape

The Battlescape is turn-based tactical combat. You control a squad of soldiers
against alien forces on a 50×50 grid map.

### Turn Structure

1. **Player Turn** — Move and act with each soldier using Time Units (TU)
2. **Alien Turn** — Aliens act using their own TU pools
3. Repeat until one side is eliminated

### Time Units (TU)

Every action costs TU. TU restore fully at the start of each player turn.

| Action | Approx. TU Cost |
|--------|-----------------|
| Move (per tile) | 4 |
| Crouch | 4 |
| Fire weapon | Varies by weapon (aimed=base, burst=1.5×, auto=2×) |
| Reload | 8 |
| Throw grenade | 20 |
| Use medikit | 25 |
| Psi attack | 20 |

A soldier's TU pool starts at 45–55 and can grow through experience (max ~80).

### Fire Modes

Weapons can have multiple fire modes. Press **Tab** to cycle, and check the
mode shown in the sidebar.

| Mode | Cost | Accuracy | Rounds | When to use |
|------|------|----------|--------|-------------|
| **Aimed** | Base TU | Best | 1 shot | Long range, high-value targets |
| **Burst** | 1.5× TU | -10% | 3 shots | Mid range, supression |
| **Auto** | 2× TU | -20% | All remaining ammo | Close range, emergencies |

Not all weapons support all modes. Rifles and laser rifles support burst;
only a few weapons support auto fire.

### Combat

Combat factors:
- **Accuracy** depends on soldier skill, distance to target, cover, and fire mode
  being used
- **Crouching** gives an accuracy bonus and reduces incoming damage
- **Cover** — walls block 80% of damage, trees 60%, bushes 40%,
  fences 30%. Position your soldiers behind solid cover
- **Bypass cover** with grenades — they explode in an area and ignore
  cover damage reduction

### Line of Sight

Soldiers can only see in a straight line. Walls, trees, and rocks block LOS.
Floors, doors, and grass do not.

### Doors

Doors (`+`) block movement and line of sight when closed. To open a door,
walk through it — your soldier will automatically swing it open as they pass.
Shooting at a closed door will also force it open. An open door is shown
as `▏` (Left One Eighth Block), is fully passable, and provides no cover.

### Objects & Cover

Shots passing through objects have damage reduced by the object's cover value.
The highest cover along the line of fire is applied.

| Object | Cover % |
|--------|---------|
| Wall / UFO Wall | 80% |
| Rock | 70% |
| Tree | 60% |
| UFO Furniture | 50% |
| Bush | 40% |
| Heavy Smoke | 40% |
| Fence | 30% |
| Rubble | 20% |

### Grenades

- Range: ~6 tiles
- Damage: based on strength, with area splash
- Destroys walls, trees, rocks, and fences within blast radius
- Creates smoke clouds that block LOS at high density

### Medikit

Heals 10 HP per use (15 HP with Field Medic perk), costs 25 TU.

### Night Missions

Night (before 6:00 or after 18:00):
- Lower accuracy (roughly 75% of daytime)
- Reduced sight range (from 20 to ~10 tiles)
- Soldiers glow warm, aliens glow faint blue

### Vision Modes

Press `V` to cycle: **Normal → Night Vision → Thermal → Normal**
- Night Vision: green phosphor overlay with static
- Thermal: living entities glow hot, terrain is cold blue

### Psi Combat

Requires a Psi-Amplifier weapon and a Psi-Lab facility. Success depends on
your soldier's psi skill vs the target's psi strength. A successful psi attack
panics the target — they lose their turn.

Soldiers in a base with a Psi-Lab may gain psi skill over time (up to ~80).
Mind Control research grants a significant psi boost to all soldiers.

### Mission Modifiers

Random modifiers that change each battle:

| Modifier | What happens |
|----------|--------------|
| Night Ops | Forced night battle, bonus loot |
| Reinforcements | Extra aliens arrive on turn 4 |
| Time Limit | 15 turns to eliminate all aliens |
| VIP Rescue | Protect a VIP, bonus cash if they survive |
| Booby Trapped | More grenades and mines on the map |
| Heavy Fog | Sight range reduced by 40% |
| Alien Ambush | Aliens start in overwatch positions |
| Low Visibility | Reduced accuracy for all units |
| High Ground | Elevated positions give accuracy bonus |

### Weather

Weather affects combat based on mission location:

| Weather | Effect |
|---------|--------|
| Rain | Lower accuracy, fire spreads slower |
| Wind | Fire spreads faster, grenades may drift |
| Snow | Movement costs more in deep snow |
| Fog | Lower accuracy, reduced sight |
| Storm | Rain + wind combined |
| Cold | Slight accuracy penalty |

### After-Action Report

After each battle, you see:
- Outcome (Victory / Defeat)
- Aliens killed and soldiers lost
- Loot recovered and prisoners captured
- Funds earned
- Per-soldier stat gains or "KIA" marker

Press **Enter**, **Space**, or **Esc** to dismiss.

---

## Weapons & Equipment

### Custom Weapon Designer

Press `G` from the Base, Soldiers, or Equip screen to open the **Weapon Designer**.
This is the primary way to create weapons for your soldiers. Pick a base template
and customize each component:

| Component | Options | What it affects |
|-----------|---------|-----------------|
| **Base** | Pistol / Rifle | Starting damage, range, accuracy, TU cost |
| **Barrel** | Short / Standard / Long / Extended | Range, accuracy, TU cost, weight |
| **Optics** | None / Iron Sights / Scope / Advanced Optics | Accuracy, TU cost, weight |
| **Fire Mode** | Semi-Auto / Full-Auto | Full-auto mode (fires faster, less accurate) |
| **Ammo** | Standard / AP / Incendiary / Explosive | Damage mod, TU cost, weight |
| **Stock** | None / Light / Heavy | Accuracy, TU cost, weight |

Each component affects the weapon's damage, accuracy, TU cost, range, and weight.
The preview panel shows the assembled weapon as colored ASCII art and displays
its final stats. Designs are saved as custom items available in the Equip screen.

**Tip:** Start with a Rifle base for most purposes. Long barrels and scopes
improve ranged accuracy. Explosive ammo hits hard but costs extra TU.

### Stock Weapons

These base items are available from the start and can be manufactured:

| Type | Damage | Ammo | Notes |
|------|--------|------|-------|
| Pistol | Light | Ballistic | Needs reloading, low weight |
| Rifle | Medium | Ballistic | Standard issue, supports burst |
| Heavy Cannon | High | Ballistic | Slow, heavy, hits hard |
| Auto Cannon | Medium | Ballistic | Full-auto option |
| Rocket Launcher | Very High | Explosive | Area damage |

### Energy Weapons

Researched later — never need reloading:

| Type | Damage | Notes |
|------|--------|-------|
| Laser Pistol | Light | Early energy weapon |
| Laser Rifle | Medium | Supports burst fire, never reloads |
| Plasma Pistol | Medium | Alien weapon, never reloads |
| Plasma Rifle | High | Alien weapon, never reloads |
| Heavy Plasma | Very High | Top-tier alien weapon |

### Ammo & Reloading

- **Ballistic weapons** need reloading — press `R` in combat
- **Energy weapons** (Laser, Plasma) never need reloading
- **Consumables** (grenades, medikits) are used from your inventory

### Fire Modes

See [Battlescape → Fire Modes](#fire-modes) for details.

### Inventory Items

Soldiers can carry extra items in their inventory slot:
- **Grenades** — thrown explosive with area damage
- **Medikits** — heal yourself or an adjacent ally
- **Motion Scanners** — detect nearby enemies
- **Proximity Mines** — place on the ground, detonates when an enemy walks over it
- **Psi-Amps** — enable psi attacks (requires psi skill)
- **Melee weapons** — Stun Rod for non-lethal takedowns

Each inventory item adds weight and increases encumbrance, reducing your
available TU in battle. Pack wisely.

### Procedural Items

Each playthrough generates unique weapons and armour based on the alien species
encountered. These have randomized names and stats — every game is different.

**Procedural weapons:** 2–3 weapons with damage types matching the alien species.
**Procedural armour:** 1–2 armour pieces with protection matching alien damage types.

These items are automatically added to your stores at game start.

---

## Armour

| Armour | Defence | TU Penalty | Notes |
|--------|---------|------------|-------|
| None | 0 | None | Default |
| Personal Armour | 10 | None | Early game standard |
| Light Suit | 20 | -5% TU | Good mid-game option |
| Medium Suit | 30 | -10% TU | Strong protection |
| Heavy Suit | 40 | -15% TU | Max defence, heavy penalty |
| Power Suit | 50 | -10% TU | Endgame armour |
| Flying Suit | 45 | -5% TU | Near-endgame, lighter than Power |

Higher defence reduces damage taken, but heavier suits cost Time Units.

---

## Aliens

### Procedural Species

Each game generates 5–7 unique alien species from a seed. Every species has
2–5 rank variants (Soldier → Navigator → Commander → Elite → Overlord).

Species differ in:
- **Damage type** — the kind of damage they deal
- **Resistances & weaknesses** — some are weak to plasma, others to explosives
- **Weapon preference** — lower ranks use pistols, higher ranks use heavy weapons
- **Morphology** — physical body plan affecting stats and resistances

This means **every playthrough features different alien threats**. One run may
have a psionic-heavy species, another may have melee predators weak to explosives.

### Morphology

Morphology determines an alien's physical form. Key factors:

**Limbs:**
- Arms (0–6): Fewer arms = worse accuracy, more arms = better stability or dual-wield
- Legs (0–8): More legs = faster but larger target; zero legs = floating, harder to hit

**Body types and their resistances:**
- **Carbon Flesh:** +Kinetic resistance, -Explosive weakness
- **Silicon Based:** +Laser/+Plasma resistance, -Explosive weakness, reflective
- **Gaseous:** Immune to kinetic, weak to plasma, can phase through walls
- **Crystalline:** Good all-round resistance, very weak to explosives, shatters on death
- **Amorphous:** +Psi resistance, regenerates HP each turn
- **Mechanical:** Immune to psi, +Plasma resistance, -Laser weakness, self-destructs
- **Bio-Synthetic:** Balanced resistances, heals adjacent aliens
- **Nanotech:** +Kinetic resistance, can revive on death

**Senses:**
- **Eyesight:** Affects accuracy — multi-spectrum ignores smoke/darkness
- **Hearing:** Echolocation detects units through smoke at close range
- **Thermal Sense:** Detects living units regardless of cover at close range
- **Psionic Sense:** Boosts psi, detects mind-controlled humans
- **Chemical Sense:** Bonus accuracy vs wounded targets

### Knowledge Levels

As you encounter aliens, intel improves:

| Level | What you learn |
|-------|----------------|
| Unknown | Name appears as "???" |
| Sighted | Name and icon revealed |
| Killed | Stats and resistances revealed |
| Autopsied | Full lore and detailed weaknesses |

### Alien AI

Aliens patrol until they spot a human, then attack. Behaviors include:
- **Search** — move toward last known position for a few turns
- **Flee** — run away when badly hurt and low on bravery
- **Adapt** — aliens study your tactics across missions.
  Snipe from range? They'll rush you. Use grenades? They'll spread out.
  Flank often? They'll post suppressors.

### Equipment Escalation

Aliens get better equipment as the campaign progresses:
- **Early months:** Plasma pistols, basic armour
- **Mid campaign:** Plasma rifles, heavy plasma, alien cannons
- **Late campaign:** Top-tier alien weapons and armour

### Alien Capture

Use a **Stun Rod** (melee, $2K to manufacture) to knock aliens unconscious.
If the stun damage exceeds their HP, they fall unconscious and can be collected
after the mission — provided you have Alien Containment with free capacity.

Captured aliens can be interrogated from the Research screen (`I` key):
- Interrogation can complete an active autopsy instantly
- Or grant a progress bonus to current research
- Requires at least one Laboratory

---

## Soldier Ranks & Progression

### Ranks

Ranks unlock as your total roster grows:

| Rank | Unlocks when roster reaches |
|------|----------------------------|
| Rookie | Always available |
| Squaddie | Always available |
| Corporal | 4 soldiers |
| Sergeant | 8 soldiers |
| Lieutenant | 14 soldiers |
| Captain | 22 soldiers |
| Major | 30 soldiers |
| Colonel | 40 soldiers |

### Stat Growth

Soldiers improve through **per-action experience** during battle:
- **Firing** → improves Accuracy
- **Reactions** → improves Reactions
- **Melee** → improves Strength
- **Bravery** → improves Bravery (from resisting panic)
- **Psi skill** → improves Psi Skill and Psi Strength

After each mission, accumulated XP is converted to stat gains. Soldiers who
gained XP also get general "halo" growth toward their HP, TU, and Strength
caps. Caps are roughly: TU 80, HP 60, Accuracy 120, Reactions 100,
Bravery 100, Strength 70, Psi 100.

### Fatigue & Wounds

- **Wounded soldiers** cannot deploy until healed (2 HP/day recovery)
- **Fatigue:** Battles cause 1–5 days of fatigue
- Healing facilities and Living Quarters speed recovery

### Fatal Wounds

In battle, hits may cause fatal wounds and bleeding. Bleeding drains HP each
turn — get a medikit on them fast. Surviving wounds become recovery days
after the mission.

### Morale

Soldiers recover morale each turn. Low morale can trigger panic (skip turn).
Resisting panic builds bravery XP.

### Perks

Each rank-up grants a random perk:

| Perk | Effect |
|------|--------|
| Lightning Reflexes | +10 Reactions |
| Marksman | +Accuracy at long range |
| Grenadier | Larger grenade splash |
| Field Medic | Medikit heals more |
| Iron Will | +Psi Skill and +Psi Strength |
| Steady Aim | +Accuracy when stationary |
| Close Combat Specialist | +Accuracy at close range |
| Overwatch Expert | +Reaction fire accuracy |
| Demolitions | +Grenade damage |
| Scavenger | +Loot from battles |
| Tough | +5 Max HP |
| Quick Learner | +XP gain |

### Memorial

Soldiers killed in action are recorded in the in-game Memorial.
You can view it to honour the fallen.

---

## Save/Load

| Key | Action |
|-----|--------|
| F5 | Open save slot picker |
| F9 | Open load slot picker |

Saves include: game time, funds, pause state, alien activity, base state,
UFOs, active missions, procedural species seed, and alien knowledge levels.
The seed ensures the same alien species regenerate on reload.

**Autosave:** If enabled in Options, the game auto-saves periodically.

---

## Key Bindings Reference

### Geoscape

| Key | Action |
|-----|--------|
| Arrow keys | Move camera |
| j/k | Navigate region list |
| Space | Pause/unpause |
| 1–4 | Time speed |
| B | Open base |
| L | Launch interceptor |
| A | Autoresolve nearest UFO |
| M | Respond to mission |
| R | Dispatch transport to crash site |
| C | Cycle to next base |
| N | Build new base ($500K) |
| T | Open transfer screen |
| E | Open encyclopedia |
| V | Toggle radar overlay |
| F5 | Save |
| F9 | Load |
| Q | Quit |
| ? | Help |

### Base Management

| Key | Action |
|-----|--------|
| 1–6 | Switch tabs |
| j/k | Navigate items |
| B | Build facility |
| S | Sell facility |
| H | Hire soldier |
| E | Open equip screen |
| G | Open weapon designer |
| D | Dismiss soldier / Plane designer (Hangars) |
| Esc | Back to geoscape |

### Equip Screen

| Key | Action |
|-----|--------|
| ↑/↓ | Select soldier |
| Tab | Cycle available items |
| 1 | Weapon slot |
| 2 | Armour slot |
| 3 | Inventory slot |
| Space | Equip selected item |
| G | Open weapon designer |
| A | Auto-equip all soldiers |
| Esc | Back |

### Weapon Designer

Press `G` from Base, Soldiers, or Equip screen.

| Parameter | Options | Effect |
|-----------|---------|--------|
| Barrel | Short / Standard / Long / Extended | Range, accuracy, weight, TU cost |
| Optics | None / Iron Sights / Scope / Advanced | Accuracy, weight, TU cost |
| Fire Mode | Semi-Auto / Full-Auto | Full-auto mode |
| Ammo | Standard / AP / Incendiary / Explosive | Damage, weight, TU cost |
| Stock | None / Light / Heavy | Accuracy, weight, TU cost |

### Plane Designer (Custom Interceptors)

All interceptors are designed and built through the **Plane Designer**.
Press `D` from the Hangars tab to open it. Configure your aircraft:

| Parameter | Range | What it affects |
|-----------|-------|-----------------|
| **Length** | Short (3) → Long (7) | Hull points, mass, speed |
| **Wingspan** | Short (1) → Wide (4) | Maneuverability, mass |
| **Engines** | 1–3 | Speed, fuel capacity, mass |
| **Fuel** | 20–100 | Operational range |
| **Weapon** | Cannon / Stingray / Avalanche / Plasma | Firepower, weight, cost |
| **Armor** | None / Light Alloy / Heavy Alloy / Alien Plating | Hull bonus, damage reduction, mass |

The designer computes derived stats (speed, firepower, hull, mass-to-thrust ratio)
from your configuration and displays a colored ASCII preview. Heavier designs
are tougher but slower — balance durability against interception speed.

**Plane weapons:**

| Weapon | Damage | Accuracy | Range | Fire Rate | Cost |
|--------|--------|----------|-------|-----------|------|
| Cannon | 15 | 85% | 25 | 3 shots | $5K |
| Stingray | 25 | 70% | 45 | 2 shots | $8K |
| Avalanche | 40 | 55% | 60 | 1 shot | $12K |
| Plasma | 60 | 50% | 50 | 1 shot | $20K |

**Plane armour:**

| Armour | Hull Bonus | Damage Reduction | Cost |
|--------|------------|------------------|------|
| None | 0 | 0% | Free |
| Light Alloy | +10 | 10% | $8K |
| Heavy Alloy | +25 | 25% | $18K |
| Alien Plating | +40 | 40% | $35K |

### Battlescape

| Key | Action |
|-----|--------|
| Arrow keys / WASD / hjkl | Move cursor |
| Space / Enter | Select unit / confirm |
| q | Cycle soldiers |
| f | Fire weapon |
| Tab | Cycle fire mode |
| r | Reload |
| e / n | End turn |
| c | Crouch |
| g | Throw grenade |
| m | Move mode |
| h | Use medikit |
| p | Psi attack |
| y | Motion scanner |
| t | Place proximity mine |
| v | Cycle vision mode |
| o | Options |
| ? | Help |
| Esc | Cancel / deselect |

### Mobile Touch Controls

On browser with narrow screen (cols < 100) or when `touch_mode` is enabled:

| Gesture | Action |
|---------|--------|
| Tap | Select, move, fire |
| Long press (500ms) | Cancel |
| Vertical drag | Scroll |

A `[=]` button opens a touch-friendly on-screen control menu.

---

## Tips & Strategy

### Early Game

1. Research **Alien Alloys** first — it unlocks Laser Weapons and Armour.
2. Build a second **Radar** — more detection, more monthly funding.
3. Hire 2–4 extra soldiers to fill your squads.
4. Use the **Weapon Designer** to create custom rifles — you can build
   better guns than the stock models with the right components.
5. Don't ignore autopsies — some weapon techs require them.
6. Custom-design your interceptors in the **Plane Designer** — a balanced
   design (medium length/wingspan, 2 engines, Stingray missiles) outperforms
   default interceptors.

### Combat

- **Use cover** — walls (80%) > rocks (70%) > trees (60%) > bushes (40%)
- **Crouch** before firing for better accuracy and damage reduction
- **Grenades** bypass cover and destroy walls — perfect for entrenched enemies
- **Learn alien resistances** — check the encyclopedia after first kills
- **Don't overextend** — aliens get reaction shots when you move in their LOS
- **Keep a medic** — one soldier with a medikit can save lives
- **Manage encumbrance** — don't overload soldiers with heavy gear

### Economy

- Sell excess alien corpses and loot for cash

- Monthly salaries add up — balance your roster against income
- Council missions pay $100K bonus — prioritize them
- Manufacture items to sell for profit in the early game

### Research Path

Alloys → Laser Weapons → Personal Armour → Autopsies → Elerium → Plasma Weapons

Mid game: Medium Suit, Heavy Plasma.
Late game: Power/Flying Suit, Mind Control.

### Base Building

- Radars pay for themselves (+$50K/month each)
- Build Storage early — you'll fill up fast
- Alien Containment is needed for live captures and interrogation bonuses
- Adjacent facilities boost each other — plan your base layout
- Build a Psi-Lab if you want psi capabilities
