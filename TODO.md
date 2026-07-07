# TODO

## Phase 1 — Core Engine (done)
- [x] Project scaffolding (go mod, tcell, directory layout)
- [x] Terminal renderer abstraction (`engine/screen.go`)
- [x] Game state machine with screen stack (`engine/game.go`)
- [x] Main loop with tick-based time

## Phase 2 — Geoscape (done)
- [x] Equirectangular world map rendered as ASCII
- [x] Real-time clock with pause / time-compression
- [x] City & base placement on map
- [x] UFO spawning and movement along random paths
- [x] Interceptor launch and vector-based pursuit
- [x] Dogfight resolution (simplified)
- [x] Radar blips & detection range

## Phase 3 — Base Management (done)
- [x] Base with facilities (Lab, Workshop, Quarters, Storage, Radar, Containment)
- [x] Build / sell facilities
- [x] Hire / dismiss soldiers
- [x] Equip soldiers (armour + two weapons)
- [x] Monthly budget cycle

## Phase 4 — Research & Manufacturing (done)
- [x] Research tree (alien tech → plasma, laser, psi)
- [x] Assign scientists, track progress
- [x] Manufacture items with engineers + materials
- [x] Manufacture screen with queue

## Phase 5 — Battlescape (done)
- [x] Tile-based tactical map (crash site / terror)
- [x] Soldiers spawn with loadout
- [x] Time-unit system: move, turn, fire, reload, crouch
- [x] Line-of-sight (Bresenham)
- [x] Cover system (high / low / full)
- [x] Weapon firing with accuracy + damage
- [x] Alien AI: patrol → seek → attack → flee
- [x] Win/lose conditions (kill all aliens or squad wiped)
- [x] After-action: XP gain, loot recovery

## Phase 6 — Soldiers (done)
- [x] Stats: HP, TU, Accuracy, Bravery, Reactions, Strength, Psi
- [x] Rank progression (Rookie → Colonel) with stat bonuses
- [x] Injury / recovery system
- [x] Psi strength & psi skill

## Phase 7 — Data (done)
- [x] Weapons: Pistol, Rifle, Heavy, Rocket, Laser, Plasma
- [x] Armour: Personal, Light, Medium, Heavy, Power Suit
- [x] Aliens: Sectoid, Floater, Muton, Ethereal (+ stat blocks)
- [x] Items: Medikit, Motion Scanner, Grenades, Alloys, Elerium
- [x] Research tree with prerequisites

## Phase 8 — Polish
- [ ] Save / load game
- [ ] Sound effects (terminal bell)
- [ ] More map variety (UFO interior, Cydonia)
- [ ] Alien missions (base defence, terror, supply)
- [ ] Funding & diplomacy system
- [ ] Transfer screen between bases
- [ ] Autoresolve for interceptions
- [ ] Campaign victory condition (destroy alien base)
