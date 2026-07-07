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

## Phase 3 — Base Management (partial)
- [x] Base with facilities (Lab, Workshop, Quarters, Storage, Radar, Containment)
- [x] Build facilities (deducts funds, construction timer)
- [ ] Sell facilities
- [ ] Hire / dismiss soldiers (PLANNED)
- [ ] Equip soldiers (armour + two weapons)
- [ ] Monthly budget cycle

## Phase 4 — Research & Manufacturing (stub)
- [ ] Research screen (assign scientists, track progress)
- [ ] Manufacture screen (queue items, assign engineers)
- [ ] Transfer screen between bases

## Phase 5 — Battlescape (partial)
- [x] Tile-based tactical map (crash site / terror)
- [x] Time-unit system: move, turn, fire, reload, crouch
- [x] Line-of-sight (Bresenham)
- [x] Cover system
- [x] Weapon firing with accuracy + damage
- [x] Alien AI: patrol, seek, attack, flee
- [x] Win/lose conditions
- [ ] Use base soldier roster instead of hardcoded rookies (PLANNED)
- [ ] Award XP on kills (PLANNED)
- [ ] Loot recovery after battle (PLANNED)
- [ ] Sync soldier state back to base after battle

## Phase 6 — Soldiers (partial)
- [x] Stats: HP, TU, Accuracy, Bravery, Reactions, Strength, Psi
- [x] Rank progression (Rookie → Colonel) with stat bonuses
- [ ] Soldier roster in base screen (PLANNED)
- [ ] Injury / recovery system
- [ ] Equipment screen (PLANNED)

## Phase 7 — Data (done)
- [x] Weapons: Pistol, Rifle, Heavy, Rocket, Laser, Plasma
- [x] Armour: Personal, Light, Medium, Heavy, Power Suit
- [x] Aliens: Sectoid, Floater, Muton, Ethereal (+ stat blocks)
- [x] Items: Medikit, Motion Scanner, Grenades, Alloys, Elerium
- [x] Research tree with prerequisites

## Phase 8 — Wiring the Game Loop
- [ ] Geoscape → Battlescape transition on UFO crash
- [ ] Post-battle: XP gain, loot recovery, soldier sync
- [ ] Monthly budget cycle (salaries, funding)
- [ ] Funding & diplomacy system

## Phase 9 — Polish
- [ ] Save / load game
- [ ] Sound effects (terminal bell)
- [ ] More map variety (UFO interior, Cydonia)
- [ ] Alien missions (base defence, terror, supply)
- [ ] Autoresolve for interceptions
- [ ] Campaign victory condition (destroy alien base)

## Phase 10 — Quality
- [ ] Increase test coverage (currently 28%)
- [ ] Add integration tests for game loop
- [ ] Benchmarks for AI and pathfinding
