# TODO — Battlescape Improvements

Scope: Features and fixes for the battlescape tactical combat system.
Nothing here is implemented yet unless marked `[x]`.

---

## Burst / Auto Fire

- [x] `data/items.go`: `FireMode` enum, `ModeTU`, `ModeAccuracy`, `ModeRounds` methods
- [x] `battle/unit.go`: `FireMode` field on `Unit`, `FireAt` multi-round loop
- [x] `battle/input.go`: `Tab` cycles fire mode
- [x] `battle/ai.go`: `selectFireMode()` auto-picks mode by distance
- [x] `battle/battlescape.go`: `CycleFireMode()` method, sidebar display
- [x] Language keys: `FIRE_MODE_AIMED/BURST/AUTO`, `SIDE_FIRE_MODE`, `MSG_FIRE_MODE`
- [x] Bump version

## Inventory Management

- [x] `soldier/soldier.go`: `Inventory []string` field, `Encumbrance`, `TUPenalty`, `AddItem`, `RemoveItem`, `HasItem`, `CountItem`
- [x] `battle/unit.go`: `NewSoldierUnit` applies TU penalty from encumbrance
- [x] `save/save.go`: persist/restore `Inventory` in `SoldierSave`
- [x] `base/equip.go`: slot 2 for inventory items (grenades, medikits, scanners, mines, melee)
- [x] `base/equip.go`: encumbrance display, `3` key for inventory slot
- [x] `battle/battlescape.go`: sidebar shows inventory items when selected unit has them
- [x] Language keys: `LABEL_INVENTORY`, `SIDE_INVENTORY`, `SIDE_ENCUMBRANCE`, `MSG_ADDED_INVENTORY`
- [x] Bump version

## Soldier Progression (OXCE+ port)

- [ ] **Phase 1** Per-action XP + `improveStat` + halo growth
  - [ ] `soldier.go`: add transient XP counters and `GainedXP bool`
  - [ ] `soldier.go`: add stat caps (TU 80, HP 60, Acc 120, React 100, etc.)
  - [ ] `soldier.go`: add `AddFiringExp/AddThrowingExp/AddReactionsExp/AddBraveryExp/
        AddPsiSkillExp/AddPsiStrExp/AddMeleeExp` methods
  - [ ] `soldier.go`: add `improveStat(exp int) int` (OXCE formula)
  - [ ] `soldier.go`: add `PostMission()` with stat improvement + halo growth
  - [ ] `battle/unit.go` `FireAt`: award `AddFiringExp()` on hit
  - [ ] `battle/battlescape.go` reaction fire: award `AddReactionsExp()` on hit
  - [ ] `battle/battlescape.go` psi attack: award `AddPsiSkillExp()` on success
  - [ ] Replace kill-based `GainXP` calls with `PostMission()`

- [ ] **Phase 2** Ranks by headcount
  - [ ] `soldier.go`: add `HandlePromotions(roster []*Soldier)` (headcount-based)
  - [ ] Each promotion rolls `RollPerk`/`ApplyPerk`
  - [ ] Call from `finishBattle`
  - [ ] Update tests

- [ ] **Phase 3** Fatal wounds & bleeding
  - [ ] `battle/unit.go`: add `FatalWounds int` + `BleedRate int`
  - [ ] `FireAt`: chance to add fatal wound on damage
  - [ ] Turn loop: apply bleed tick
  - [ ] `finishBattle`: fold into `Wounds` field

- [ ] **Phase 4** Bravery-on-panic (minimal morale)
  - [ ] `battle/unit.go`: add `Morale int` (default 100)
  - [ ] Per-turn morale recovery + panic roll
  - [ ] On avoiding panic → `AddBraveryExp()`

- [ ] **Phase 5** Psi-lab training refinement
  - [ ] `base/facility.go`: replace daily `PsiSkill++` with probability model
  - [ ] Update `facility_test.go`

- [ ] **Phase 6** Wire `HasBattleMod` into combat
  - [ ] `unit.go` `FireAt`: apply Marksman, CloseCombat, Overwatch, SteadyAim
  - [ ] Reaction fire: apply battle modifiers
  - [ ] Death/memorial: `KIA` record in `finishBattle`

## Code Audit (largest files)

### `internal/battle/battlescape.go`
- [ ] **B1 (Critical)** Movement-range BFS charges 8 TU for Tree/Rock/Water but
  `MoveTo`/player movement charges flat 4/tile (lines 240–244 vs 1767–1791).
  Reachable overlay over-promises range. Align cost models.
- [ ] **B4 (Critical)** `PsiAttack`/`Reload`/`UseMedikit` deref
  `bs.Selected.Soldier.Weapon` without nil guard (line 2482/2507). Add nil guard.
- [ ] **M1 (Medium)** `GetMovementRange` cache key `X*10000+Y*100+TU` collides when
  switching selected unit / at X,Y≥100 (lines 197–204). Include unit ID in key.
- [ ] **M4 (Medium)** Unreachable `else` lighting branches inside `if bs.IsNight`
  (lines 2654–2658, 2671–2673). Drop inner checks.
- [ ] **D1 (Low)** `bs.viewW`/`bs.viewH` assigned (line 2577) but never read. Remove.
- [ ] **D2 (Low)** `PlayerFlankCount` declared but never used (line 142). Remove.
- [ ] **D4 (Low)** Commented-out ambient-particle block (lines 777–788). Delete.
- [ ] **D3 (Low)** Overwatch `CombatStatus` set then overwritten by flash reset
  (lines 736–745). Reconcile or drop.
- [ ] **R1 (Low)** Shot-FX duplicated across player reaction, alien reaction, alien
  AI fire (lines 855–884, 1003–1025, 1076–1098). Extract `spawnShotFX(...)`.
- [ ] **R2 (Low)** `Render` ~550 lines. Extract sidebar drawers.
- [ ] **R3 (Low)** Reinforcement spawn dup'd in `checkReinforcements`/
  `spawnReinforcementWave` (lines 1398–1568). Extract `spawnAlienWave`.
- [ ] **U1 (Low)** Exported methods (SelectUnit, LeftClick, Render, HandleKey, …)
  lack doc comments. Add per AGENTS.md.
- [ ] **U2 (Low)** Magic numbers (reaction 15, OverwatchFlash 30, grenade TU 20,
  damage 40+Str*2, mine 60+rng20, scanner 15). Extract constants.

### `internal/geo/geoscape.go`
- [ ] **M1 (Critical)** Defeat (`AlienActivity>=100`) sets `gs.Victory=true`
  (lines 414–419) → victory screen after game-over (line 428). Add `Defeated` flag.
- [ ] **M2 (Critical)** Losing last base sets `gs.Victory=true` (lines 703–705) →
  same win-after-loss bug. Use `Defeated` flag.
- [ ] **B4 (Critical)** Save/load drops UFO `ID`/`X`/`Y`/`TurnsLeft` (lines 1791–1802)
  → `DefendingUFOID`/`findUFOByID` never match, alien-base defender respawns every
  tick (1154–1158). Persist + restore these fields.
- [ ] **B1 (Medium)** `defBase := gs.HasBaseAt(...)` shadowed by
  `defBase := gs.SelectedBase()` (lines 1313/1320) → wrong base in multi-base games.
  Remove shadow.
- [ ] **B3 (Medium)** `Autoresolve` uses squared-distance vs `bestDist=9999.0`
  (~99-unit cutoff) and kills nearest UFO regardless of range (lines 1601–1639).
  Fix distance semantics.
- [ ] **M3 (Medium)** Target-select uses `CursorNode % len(targets)` where CursorNode
  is a city ID (lines 2139, 2652). Add separate target cursor index.
- [ ] **M4 (Medium)** `enterMissionSelectMode` omits `overwatch` perk bonus that
  `AutoresolveMission` includes (lines 1553–1562 vs 1413–1418) → odds mismatch. Add.
- [ ] **M6 (Medium)** `processBattleResult` replaces `defendingBase.Soldiers`
  wholesale with `r.Soldiers` (lines 167–168); if partial, roster wiped. Verify/merge.
- [ ] **M7 (Medium)** `respondedAlienBase` not cleared on loss (lines 259–261) →
  stale base destroyed later. Clear in loss branch.
- [ ] **B5 (Low)** Transport return uses `SelectedBase()` not source base
  (lines 609–612). Store originating base ID.
- [ ] **B6 (Low)** `confirmLaunch` `*CrashSite` derefs `CityByID(t.NodeID)` w/o nil
  check (line 1931). Guard.
- [ ] **B7 (Low)** `Autoresolve` can `rand.Intn(0)` if `alive` empty. Guard len>0.
- [ ] **M5 (Low)** `winChance` clamped twice identically (lines 1424–1446).
- [ ] **D2 (Low)** `moveCursor` `dx` param ignored; left/right arrows dead
  (lines 2745–2749).
- [ ] **R1 (Low)** Squad-power/win-chance dup'd in two functions. Extract helper.
- [ ] **R2 (Low)** `resumeRealtime()` pattern dup'd 3×. Extract.
- [ ] **R4 (Low)** `cityName(id)` helper would dedupe ~5 call sites.
- [ ] **U2 (Low)** Magic numbers: speedMult, ufoSpawnRate, spawnRate, tick consts
  7200/3600/2400/1800/600, radarRange, bestDist. Extract constants.
- [ ] **U3 (Low)** Day/night `nightBoundary` uses worldW inconsistently
  (lines 2289–2296). Document or fix.

### `internal/battle/map.go`
- [ ] **M1 (Medium)** `SpreadFire` sets `Type=TILEFloor, Fire=3` but doesn't reset
  `Cover`/`Rune` (lines 170–177) → burning tree still has Cover 60. Set
  `tile.Cover = TileCover(TileFloor)` on ignite.
- [ ] **M2 (Medium)** `TileRubble` not in `Passable` allow-list (lines 341–350) →
  after `DestroyWall`, units can't path through rubble. Add or document.
- [ ] **M3 (Low/Medium)** `hasLOS` treats endpoint tile opacity as blocking (line 516)
  → no LOS when target stands on Rock/cover. Exclude endpoint.
- [ ] **D1 (Low)** `GenerateProcedural`, `Biomes`, `Biome` (lines 764–838) have no
  non-test callers — dead exported API. Wire in or remove.
- [ ] **D2 (Low)** `CmdClearArea` duplicates `CmdFillRect` (lines 694–713). Drop.
- [ ] **R1 (Low)** `generateCorridor`/`corridorLevel` near-identical (~80 lines).
  Unify with level-param helper.
- [ ] **R2 (Low)** `DrawRect`/`drawRectLevel`, `fillRect`/`fillRectLevel` dup pairs.
  Share helpers.
- [ ] **R3 (Low)** Door placement 4-case switch repeated 4×. Extract `placeDoor`.
- [ ] **U1 (Low)** Magic numbers: `3.14159`→`math.Pi` (1201), `15` debris (880),
  `80` crashSeverity (868), literal `3` fire (175). Name them.

### `internal/data/spritebuilder.go`
- [ ] **B1 (Medium)** Weapon mask drawn at `x` with no centering offset while torso
  is drawn at `x+torsoOffset` (lines 862–886 vs 835–860) → weapon misaligned for
  non-centered torsos (Asymmetric/Bladed/Mechanical/Crystalline). Fix: draw weapon
  at `x+torsoOffset`.
- [ ] **B2 (Medium)** Legs negative-offset clamp `if legsOffset<0 {=0}` (lines
  888–891) defeats `centerOffset` for legsFloating/Serpentine/Crab → legs left-
  shifted vs centered head/torso. Remove clamp (centerOffset keeps in bounds).
- [ ] **B3 (Low)** `Morphology.BodyType` ("organic"/"synthetic") never consulted by
  builder; `BodySubtype` is sole driver. Document or wire in.
- [ ] **D1 (Low)** Unreachable `break` in torso loop (`ty>=18`, line 838). Remove.
- [ ] **D2 (Low)** Unreachable `break` in weapon loop (`ty>=18`, line 867). Remove.
- [ ] **D3 (Low)** Unreachable `break` in head loop (`y>=10`, line 784). Remove.
- [ ] **U1 (Low)** Grid dims 24/20 bare literals everywhere. Add `SpriteW/SpriteH`.
- [ ] **U2 (Low)** Exported ids (Sense, Manipulators, Locomotion, EyeStyle, Tagged*,
  AlienPixels, SpriteRegistry, AlienColorFromSeed, AlienWeaponColor) lack doc
  comments. Add per AGENTS.md.
- [ ] **M1 (Low)** 0-leg Silicon/Crystalline/BioSynthetic fall to LocomSlither
  (lines 736–753). Decide if synthetic should float.
- [ ] **M3 (Low)** Eye mask doesn't carve `Mouth` layer (lines 809–833) → mouth
  pixel remains under eye. Carve Mouth too.
- [ ] **R1 (Low)** 4 duplicated layer-stamp switch blocks (head/torso/legs/weapon).
  Extract `stampLayer(ch, dst, y, x, allowWeapon)`.
- [ ] **R2 (Low)** 6 repeated biology post-pass loops (lines 973–1029). Use config
  table.
- [ ] **R3 (Low)** `GenerateAlienPixels` ~270-line monolith. Split assembly/shading/
  biology passes.

### `internal/engine/portrait.go`
- [ ] **B1 (Medium)** `ArmourColor` set (lines 115–124) but no `generateArmourLayer`
  composited → armour has no visual effect; `isArm`/armour-dither branch dead
  (lines 193, 281–287). Add armour layer or remove the field.
- [ ] **B2 (Low)** `MarkingsColor`/`DecalColor` and `LayerMarkings`/`LayerArmour`/
  `LayerDecal`/`LayerCount` enum (lines 43–56, 64, 67) never used. Remove or impl.
- [ ] **B3 (Low)** Magic bg color `tcell.NewRGBColor(20,20,28)` duplicated at lines
  153, 187, 292. Extract `portraitBg` package var.
- [ ] **B4 (Low)** `rng.Intn` can return negative for negative seed (lines 385–391)
  → panic on index. Use `uint64` shift or abs result.
- [ ] **M1 (Low)** `browColor==ColorDefault` guard unreachable (input always RGB,
  lines 502–505). Remove guard + fallback.
- [ ] **D3 (Low)** `isHairColor`/`isHelmetColor`/`isArmorColor` are near-identical
  (lines 320–375). Factor `colorClose(c, ref, tol)`.
- [ ] **U2 (Low)** Undocumented face-proportion fractions (45%, 42%, 5/8, 5/10…,
  lines 409–428). Name constants or comment the model.
- [ ] **U3 (Low)** Hardcoded `NewPixelImage(20,24)` (line 1126) disconnected from
  `AlienPixels` `[24][20]`. Derive from `len(ap.Body)`.
- [ ] **R1 (Low)** 8 hair-style switch cases share scaffolding (lines 835–1058).
  Table-drive with `{k,d}` per style.
- [ ] **R2 (Low)** Redundant `sqrt64` wrapper (lines 13–14, 807). Inline `math.Sqrt`.
- [ ] **R3 (Low)** Ellipse-test `(dx*dx)/(RX*RX)+(dy*dy)/(RY*RY)` dup'd 4×
  (443,907,916,1079). Add `inEllipse(dx,dy,rx,ry)`.
- [ ] **Note** Helmets intentionally removed earlier (helmetColor forced ColorDefault
  in MakeSoldierPortrait); keep helmet code dormant or delete.

### `internal/data/procedural.go`
- [ ] **B4 (Medium)** `primaryDMG := rng.Intn(6)` (line 308) assumes exactly 6 DMG
  types; breaks silently if a `DMG_*` is added to aliens.go. Use `DMG_PSIONIC+1`
  or `numDamageTypes` const.
- [ ] **D2 (Low)** `generateLore(name, ...)` `name` param unused (line 1017). Remove
  or use it.
- [ ] **B1/D1 (Low)** `midSyllIdx` is a no-op wrapper (lines 342, 363–365),
  misleadingly named. Inline as `rng.Intn(len(p))`.
- [ ] **M2 (Low)** `generateLegCount` Silicon comment says "2 or 4" but code yields
  2/4/6 (line 558). Fix comment or `rng.Intn(2)`.
- [ ] **U1 (Low)** Exported `clamp` lacks doc comment (line 1046). Add.
- [ ] **U2 (Low)** Magic tuning numbers (speciesCount 5+rng3, maxRank 1+rng4,
  `rng.Intn(3)==0` synthetic, sense rolls) undocumented. Name constants.
- [ ] **R1 (Low)** Five duplicated `rng.Intn(10)` weighted-roll switches (464–616).
  Add `weightedIndex(rng, cumThresholds)`.
- [ ] **R3 (Low)** `generateVariant` 116-line dense fn (882–997); stat formulas
  repetitive. Split / table-drive.

### `internal/battle/ai.go`
- [ ] **B1 (Critical)** Grenade branch manually `ai.Unit.TU -= 18` (line 234) but
  `executeAlienAction` doesn't deduct for grenades → double/early TU loss. Remove
  the manual deduction; let execute own TU costs.
- [ ] **B2 (Critical)** Action list planned against full TU at turn start, executed
  sequentially with per-action TU costs; no re-check → moves+fires can exceed
  MaxTU (lines 183–426). Emit ≤1 expensive action or re-verify per action.
- [ ] **B3 (Medium)** `canFireAt` floor `TU>=15` (lines 512–518) ≠ actual weapon
  `w.TU` (rocket=28) → emits unaffordable fire that silently no-ops. Use `TU>=w.TU`.
- [ ] **B6 (Medium)** `executeAlienAction` `"psi"`/`"melee"` cases deduct NO TU
  (battlescape.go 919–950, 886–913) → free psi/melee. Deduct there; gate emission.
- [ ] **B5 (Medium)** Operator-precedence bug at line 263: `A && B && C || D`
  makes inner `else if longRange && dist>4` always-true → cover-seeking unreachable
  under longRange. Parenthesize + restructure.
- [ ] **B4 (Medium)** Pathfinding (`GetNextPathStep`→`AStar`) ignores map Level
  (path.go) while `findNearest` filters by level (line 936) → bad paths on
  multi-level maps. Pass/compare Level.
- [ ] **M2 (Medium)** AI uses global unseeded `rand` (lines 220, 900) instead of a
  seeded RNG → non-reproducible across save/load. Use a seeded `*rand.Rand`.
- [ ] **D1 (Low)** `AIFlee` constant never assigned (line 20). Remove or implement.
- [ ] **D2 (Low)** `SquadPlan.SecondaryTarget` never written by planSquadActions →
  selectTarget secondary branch dead. Populate or drop.
- [ ] **M1 (Low)** `patrolTarget` uses `m.Height-1` single-level but `LevelHeight-1`
  multi-level (lines 911–917). Confirm LevelHeight==Height for single-level.
- [ ] **R1 (Low)** `disperseFrom`/`moveTowardTargetCover` dup candidate construction
  (846–862 vs 791–807). Extract `orthogonalCandidates`.
- [ ] **R2 (Low)** Move-action append pattern copy-pasted ~8× (181–390). Extract
  `appendMove(...)` helper.
- [ ] **R3 (Low)** `Update` ~300 lines, 7-case switch. Split per-state handlers.
- [ ] **R4 (Low)** Faction literals 0/1/2 repeated (lines 532,777,970). Name
  `FactionHuman/Alien/Civilian`.
- [ ] **U1 (Low)** Magic TU/sight thresholds (20,18,16,14,15,3, dist<12, <=8, …)
  undocumented. Name constants.

### `internal/engine/game.go`
- [ ] **B1 (Medium)** `GameOver` calls `PushState(StateGameOver)` (lines 130–133)
  pushing the dead battle onto the stack → Esc could resurrect a finished battle.
  Use `SetState` or clear `stateStack`.
- [ ] **B2 (Low)** Quit-confirm mouse rects only set in TouchMode (524–546) →
  unclickable on desktop (424–430). Always compute rects or guard handler.
- [ ] **B3 (Low)** `lastState` defaults to `StateMenu` (125,338–343) → first screen
  transition may skip `OnScreenChange`. Init to sentinel -1.
- [ ] **D1 (Low)** `GetHardcodedAliens` has no callers (273–280). Remove or wire in.
- [ ] **M1 (Low)** `NewGameWeb` re-inlines full Game literal (193–208) instead of
  calling `newGameWithScreen` → drift risk. Consolidate.
- [ ] **M2 (Low)** `GetAlienTypes` fallback returns `&data.AlienTypes[i]` (262–271)
  → mutating caller corrupts shared global. Copy values.
- [ ] **M3 (Low)** `setupControlMenu` uses positional button indices (589–593,
  633–640) → brittle. Use named refs/map.
- [ ] **U2 (Low)** Magic numbers: boxW/H 46/7, btnW 16, gap 4, frameSleep 16ms,
  keyChan buf 20, Funds 500000, start date 1999-03-01. Name constants.
- [ ] **R1 (Low)** `setupControlMenu` ~140-line switch with dup closures. Add
  `keyBtn(label,hotkey,key,str)` helper.
- [ ] **R3 (Low)** `RegisterScreen`/`SetScreen`/`OpenEncyclopedia` redundant nil
  guards (282–303). Drop.

### `internal/base/facility.go`
- [ ] **B1 (Critical)** Psi-training chance always 100%, not 8%: `rand.Intn(100) <
  8*100/50` (line 392) → integer math = `16<100` always true. Fix: `< 8`.
- [ ] **B2 (Medium)** `BuyInterceptor` ignores per-weapon Cost, hardcodes 100000
  (167,192). Use `InterceptorWeapons[weaponKey].Cost`.
- [ ] **D1 (Low)** `Base.MaxStorage` set but never read (109,131). Remove or use.
- [ ] **D2 (Low)** `ManufactureItem` struct has no callers (399–405). Delete.
- [ ] **D3 (Low)** `FacilityInfo.Size` set everywhere but ignored (BuildFacility
  hardcodes 8-col grid, 250–251). Use or drop.
- [ ] **M1 (Low)** `ChangeInterceptorWeapon` cycle includes "cannon" not sold via
  BuyInterceptor; curIdx -1 silently resets to avalanche (215,232–237).
- [ ] **U1 (Low)** `HireCost=50000`, `interceptorWeaponOrder` undocumented; ammo
  `w.FireRate*4` magic 4.
- [ ] **R1 (Low)** Research-unlock logic dup'd in InterrogateAlien + AdvanceResearch
  (613–630,717–739). Extract `applyUnlocks(topic)`.

### `internal/base/base.go`
- [ ] **B1 (Medium)** `statusKey := "INTERCEPTOR_STATUS_"+ToUpper(hg.Status)` (331)
  re-looks-up a *localized* status string → fails in non-English. Store raw key.
  Same anti-pattern at facility.go 174/181/208/222.
- [ ] **B2 (Low/Med)** Magic tab count `6` / wrap `>6` (362,381) but valid max idx 5
  → selection 6 momentarily reachable. Use `len(tabs)`.
- [ ] **M1 (Low)** `HandleMouse` vs `HandleKey` hotkey dispatch dup'd (465–551 vs
  349–456); they already differ ('g' opens different designers). Extract
  `dispatchHotkey`.
- [ ] **U1 (Low)** `BaseScreen` fields lack doc; `storesItems` order-dependent on
  render. Comment.

### `internal/data/vehicle.go`
- [ ] **B1 (Low/Med)** `simpleRand.intn` overflows int64 and masks high bits →
  biased, poor distribution (579–587). Use `math/rand.NewSource(seed).Intn(n)`.
- [ ] **B2 (Low)** `seed 0` silently remapped to 42 (573–576) → breaks determinism
  if 0 is a legit seed. Comment or handle.
- [ ] **M1 (Low)** `UFOTier` comment says Interceptor 5x7 but tierConfigs has
  7x5 (420–424,467). Fix comment/dims.
- [ ] **B3 (Low)** Generated UFO can end with 0 engines/weapons if edge slots
  pre-filled (443–444,521–547). Assert `GetTotalFirepower()>0 && HasEngine()`.
- [ ] **R1 (Low)** 4 preset generators dup'd coordinate lists (312–412). Fold into
  tierConfigs/seed tables.
- [ ] **U1 (Low)** `simpleRand` undocumented custom RNG — explain why not math/rand.

### `internal/data/items.go`
- [ ] **B1 (Medium)** `"alien_grenade"` collides between RuleItems weapon (400) and
  Items loot (503). Rename loot → `alien_grenade_item`.
- [ ] **B2 (Low)** ShortName collisions MSC/PRM/PSI across RuleItems vs Items
  (504–507). Document or avoid cross-index by ShortName.
- [ ] **D1 (Low)** `Weapons` map is write-only (AmmoCur never read, 49–51,663–668).
  Remove or document.
- [ ] **U1 (Low)** `BT_*` constants undocumented (11–23). Add doc.
- [ ] **R1 (Low)** 3 near-identical `DisplayName*` methods (634–661). Extract
  `localizedName(langKey, fallback)`.

### `internal/data/aliens.go`
- [ ] **B1 (Medium)** `GetAlienByRank(minRank)` returns *lowest* rank ≥ min, not
  highest (382–392) → difficulty scaling softer than intended. Rename or invert.
- [ ] **B2 (Medium)** Procedural morphology icon pools (468–493) include glyphs
  (⬢ etc.) also used by hardcoded aliens; `UsedHardcodedIcons` doesn't seed them
  → collision on map (77,608). Seed or exclude.
- [ ] **B3 (Low)** `nextIcon(-1,...)` fallback uses `len(used)` map len for pool
  index (46–73) → can reassign used glyph. Use a counter.
- [ ] **M1 (Low)** Two `DamageType` fields: `Morphology.DamageType` (126) vs
  `AlienType.DamageType` (166) — confusing. Document roles.
- [ ] **M2 (Low)** Default fallback color `tcell.Color(9)` repeated 3× (459,586,
  448). Name as constant.
- [ ] **D-typo (Low)** Duplicate `'ቿ'` twice in same pool (472). Copy-paste typo.
- [ ] **R1 (Low)** `switch m.BodySubtype` repeats 3-color pick 9× (501–589). Extract
  `pickColor(rng, names...)`.

