# TODO

## Phase 1 — Core Engine (done)
- [x] Project scaffolding (go mod, tcell, directory layout)
- [x] Terminal renderer abstraction (`engine/screen.go`)
- [x] Game state machine with screen stack (`engine/game.go`)
- [x] Main loop with tick-based time
- [x] Mouse support across all screens
- [x] Arrow key navigation

## Phase 2 — Geoscape (done)
- [x] Equirectangular world map rendered as ASCII
- [x] Real-time clock with pause / time-compression
- [x] City & base placement on map
- [x] UFO spawning and movement along random paths
- [x] Interceptor launch and vector-based pursuit
- [x] Dogfight resolution (simplified)
- [x] Autoresolve interceptions
- [x] Alien missions (Terror, Supply, Base Assault)
- [x] Victory condition (alien activity 100%)

## Phase 3 — Base Management (done)
- [x] Base with facilities (Lab, Workshop, Quarters, Storage, Radar, Containment)
- [x] Build facilities (deducts funds, construction timer)
- [x] Sell facilities (50% refund)
- [x] Hire / dismiss soldiers
- [x] Stores system (track inventory)
- [x] Equip soldiers (weapon + armor from stores)
- [x] Monthly budget cycle (salaries, government funding)

## Phase 4 — Research & Manufacturing (done)
- [x] Research screen (assign scientists, track progress)
- [x] Manufacture screen (queue items, assign engineers)
- [x] Transfer screen (stores display)

## Phase 5 — Battlescape (done)
- [x] Tile-based tactical map (crash site / terror)
- [x] Time-unit system: move, turn, fire, reload, crouch
- [x] Line-of-sight (Bresenham)
- [x] Cover system
- [x] Weapon firing with accuracy + damage
- [x] Alien AI: patrol, seek, attack, flee
- [x] Win/lose conditions
- [x] Use base soldier roster
- [x] Award XP on kills
- [x] Loot recovery after battle (goes to stores)
- [x] Sync soldier state back to base after battle

## Phase 6 — Soldiers (done)
- [x] Stats: HP, TU, Accuracy, Bravery, Reactions, Strength, Psi
- [x] Rank progression (Rookie → Colonel) with stat bonuses
- [x] Soldier roster in base screen
- [x] Injury / recovery system (wounds heal over time, +2 HP/day)
- [x] Wounded soldiers shown in red on roster

## Phase 7 — Data (done)
- [x] Weapons: Pistol, Rifle, Heavy, Rocket, Laser, Plasma
- [x] Armour: Personal, Light, Medium, Heavy, Power Suit
- [x] Aliens: Sectoid, Floater, Muton, Ethereal (+ stat blocks)
- [x] Items: Medikit, Motion Scanner, Grenades, Alloys, Elerium
- [x] Research tree with prerequisites

## Phase 8 — Wiring the Game Loop (done)
- [x] Geoscape → Battlescape transition on UFO crash
- [x] Post-battle: XP gain, loot recovery, soldier sync
- [x] Monthly budget cycle (salaries, funding)
- [x] Research progresses over time
- [x] Manufacturing progresses over time
- [x] Injury recovery progresses each day
- [x] Alien missions spawn and countdown

## Phase 9 — Polish
- [ ] Save / load game
- [ ] Sound effects (terminal bell)
- [ ] More map variety (UFO interior, Cydonia)

## Phase 10 — Quality
- [ ] Increase test coverage (currently ~45%)
- [ ] Add integration tests for game loop
- [ ] Benchmarks for AI and pathfinding
