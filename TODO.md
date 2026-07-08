# TODO

## Phase 1 — Core Engine (done)
- [x] Project scaffolding (go mod, tcell, directory layout)
- [x] Terminal renderer abstraction (`engine/screen.go`)
- [x] Game state machine with screen stack (`engine/game.go`)
- [x] Main loop with tick-based time
- [x] Mouse support across all screens
- [x] Arrow key navigation

## Phase 2 — Geoscape (done)
- [x] Network graph with 19 regional hub nodes
- [x] Real-time clock with pause / time-compression
- [x] City & base placement on map
- [x] UFO movement along edges with progress
- [x] Interceptor launch and node/UFO targeting
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
- [x] Tile-based tactical map (crash site / terror / UFO interior / Cydonia)
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

## Phase 9 — Polish (done)
- [x] Save / load game (F5/F9, JSON format)
- [x] Sound effects (terminal bell on events)
- [x] Map variety: Crash Site, Terror, UFO Interior, Cydonia

## Phase 10 — Quality (done)
- [x] Integration tests: battle sim, victory/defeat, map gen, LOS, movement
- [x] Benchmarks: AI update, patrol, LOS, fire, movement, map gen, creation
- [x] All 110+ tests pass
- [x] vet + staticcheck clean

## Phase 11 — Polish & Depth
- [x] UFO retrieval (crash sites + transport)
- [x] In-game encyclopedia (discovered via research/autopsy)
- [x] Roguelike pivot (procedural alien species generation per run)
  - [x] Damage type system (Plasma/Laser/Explosive/Melee/Kinetic/Psionic)
  - [x] Alien resistance/weakness per damage type
  - [x] Procedural name generation from syllable pools
  - [x] Species with 2-5 rank variants per run
  - [x] KnowledgeLevel tracking (unknown → sighted → killed → autopsied)
  - [x] Save/load persists species seed + knowledge state
- [x] Object-specific cover system (walls 80%, trees 60%, fences 30%, affecting shot damage)
- [ ] Multi-level maps (stairs/elevators for UFO interiors)
- [ ] Psi combat (use Psi stat in battlescape)
- [ ] Night/day missions (lighting system affecting accuracy and LOS)
