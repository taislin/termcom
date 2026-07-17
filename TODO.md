# TODO — Game Improvements

Scope: Features and fixes for the battlescape tactical combat system.

---

### `internal/battle/battlescape.go`
- [x] **R2 (Low)** `Render` ~550 lines. Extracted `drawSidebar`, `drawTargetInfo`,
  `drawUnitInfo`, `drawBattleLog`, `drawCompactBanner`, `sidebarLayout` methods.
- [x] **U2 (Low)** Magic numbers (reaction 15, OverwatchFlash 30, grenade TU 20,
  damage 40+Str*2, mine 60+rng20, scanner 15). Extracted game-balance constants.

### `internal/data/spritebuilder.go`
- [ ] **R1 (Low)** 4 duplicated layer-stamp switch blocks (head/torso/legs/weapon).
  Extract `stampLayer(ch, dst, y, x, allowWeapon)`.
- [ ] **R2 (Low)** 6 repeated biology post-pass loops (lines 973–1029). Use config
  table.
- [ ] **R3 (Low)** `GenerateAlienPixels` ~270-line monolith. Split assembly/shading/
  biology passes.

### `internal/engine/portrait.go`

### `internal/data/procedural.go`
- [x] **D2 (Low)** `generateLore(name, ...)` `name` param unused. Removed.
- [x] **B1/D1 (Low)** `midSyllIdx` is a no-op wrapper, misleadingly named. Inlined.
- [x] **M2 (Low)** `generateLegCount` Silicon comment says "2 or 4" but code yields
  2/4/6. Fixed code to produce "2 or 4".
- [x] **U1 (Low)** Exported `clamp` lacks doc comment. Added.
- [ ] **U2 (Low)** Magic tuning numbers (speciesCount 5+rng3, maxRank 1+rng4,
  `rng.Intn(3)==0` synthetic, sense rolls) undocumented. Name constants.
- [ ] **R1 (Low)** Five duplicated `rng.Intn(10)` weighted-roll switches (464–616).
  Add `weightedIndex(rng, cumThresholds)`.
- [ ] **R3 (Low)** `generateVariant` 116-line dense fn (882–997); stat formulas
  repetitive. Split / table-drive.

### `internal/battle/ai.go`
- [ ] **B4 (Medium)** Pathfinding (`GetNextPathStep`→`AStar`) ignores map Level
  (path.go) while `findNearest` filters by level (line 936) → bad paths on
  multi-level maps. Pass/compare Level.
- [ ] **M2 (Medium)** AI uses global unseeded `rand` (lines 220, 900) instead of a
  seeded RNG → non-reproducible across save/load. Use a seeded `*rand.Rand`.
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
- [ ] **B2 (Low)** Quit-confirm mouse rects only set in TouchMode (524–546) →
  unclickable on desktop (424–430). Always compute rects or guard handler.
- [ ] **B3 (Low)** `lastState` defaults to `StateMenu` (125,338–343) → first screen
  transition may skip `OnScreenChange`. Init to sentinel -1.
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
- [ ] **D3 (Low)** `FacilityInfo.Size` set everywhere but ignored (BuildFacility
  hardcodes 8-col grid, 250–251). Use or drop.
- [ ] **M1 (Low)** `ChangeInterceptorWeapon` cycle includes "cannon" not sold via
  BuyInterceptor; curIdx -1 silently resets to avalanche (215,232–237).
- [ ] **U1 (Low)** `HireCost=50000`, `interceptorWeaponOrder` undocumented; ammo
  `w.FireRate*4` magic 4.
- [ ] **R1 (Low)** Research-unlock logic dup'd in InterrogateAlien + AdvanceResearch
  (613–630,717–739). Extract `applyUnlocks(topic)`.

### `internal/base/base.go`
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
- [ ] **B3 (Low)** `nextIcon(-1,...)` fallback uses `len(used)` map len for pool
  index (46–73) → can reassign used glyph. Use a counter.
- [ ] **M1 (Low)** Two `DamageType` fields: `Morphology.DamageType` (126) vs
  `AlienType.DamageType` (166) — confusing. Document roles.
- [ ] **M2 (Low)** Default fallback color `tcell.Color(9)` repeated 3× (459,586,
  448). Name as constant.
- [ ] **D-typo (Low)** Duplicate `'ቿ'` twice in same pool (472). Copy-paste typo.
- [ ] **R1 (Low)** `switch m.BodySubtype` repeats 3-color pick 9× (501–589). Extract
  `pickColor(rng, names...)`.

### `internal/battle/crash.go`
- [ ] **B1 (Medium)** `ExteriorTiles` may contain duplicate coordinates (no dedup check) (131–146)
- [ ] **U1 (Low)** `0.3` — undocumented factor for HP-ratio boost to destroy chance (48)
- [ ] **U2 (Low)** `3` in `rand.Intn(3)` — blood spawn probability denominator (61)
- [ ] **U3 (Low)** `50` in `rand.Intn(100) < damage*50` — 50% per damage point undocumented (188)
- [ ] **U4 (Low)** `radius * 2` for power-core chain damage multiplier (185)
- [ ] **R1 (Low)** `TileCover(TileRubble)` called repeatedly; use cached constant (56–57, 191–192)

### `internal/battle/gas.go`
- [ ] **U1 (Low)** `3` used as max-gas-density sentinel in `Set`/`BlocksLOS`/`CoverPenalty`/`Draw` — extract `const MaxGasDensity = 3` (49,68,79,151)
- [ ] **U2 (Low)** `40`, `20` — cover penalty values for density 3/2 (80,82)
- [ ] **U3 (Low)** `v.Density <= 1` — minimum density threshold for diffusion (100)
- [ ] **U4 (Low)** All RGB color literals in `Draw` (154–171) — extract named palette constants
- [ ] **R1 (Low)** `Draw` has near-duplicate color blocks for GasSmoke vs GasPoison per density level — extract `gasStyle(density, gType)` helper

### `internal/battle/input.go`
- [ ] **B2 (Medium)** `bs.ScrollX`/`ScrollY`/`Phase`/`Camera` accessed inside `handleKey`/`handleMouse` with only `bs.State.mu` held — data race if another goroutine writes them (40,114,177)
- [ ] **B3 (Medium)** `handleMouse` accesses `bs.HoveredUnit`/`bs.Selected` without synchronization on outer `Battlescape` fields (147–187)
- [ ] **U1 (Low)** `3` — camera pan distance hardcoded; extract `const CamPanStep = 3` (66–72, 121–142)
- [ ] **U2 (Low)** `1` — hardcoded column offset for help-bar hit testing (204)

### `internal/battle/modifiers.go`
- [ ] **U1 (Low)** All probability denominators in `RollModifiers` — `3`, `4`, `5`, `6` used directly; name as `const` like `nightOpsChance = 3` (78–111)
- [ ] **U2 (Low)** `5`, `3`, `2` in `AccuracyPenalty`, `SightReduction` (171–197)
- [ ] **U3 (Low)** `5`, `30`, `20` in `FireSpreadChance` (200–207)
- [ ] **U4 (Low)** `1 + rng.Intn(2)` random fog range (154,164)
- [ ] **R1 (Low)** `RollModifiers` has near-identical `rng.Intn(N) == 0` ×7 — extract `func roll(rng *rand.Rand, chance int) bool`

### `internal/battle/path.go`
- [ ] **B1 (Medium)** `passableFor` checks `y >= m.LevelHeight` but never considers calling unit's `Level` — pathfinding ignores vertical level (64)
- [ ] **U1 (Low)** `15` — TU threshold for reaction fire (138)
- [ ] **U2 (Low)** `20` — distance cutoff for reaction fire (143)
- [ ] **U3 (Low)** `2`, `3`, `5` — multipliers in reaction-fire chance formula (149)
- [ ] **U4 (Low)** `1` — minimum reaction-fire chance floor (151)

### `internal/battle/terrain.go`
- [ ] **B1 (Medium)** `TilePalette` is exported `var` (not `const`) — external code can mutate at runtime causing data races or visual corruption (32)
- [ ] **U1 (Low)** `0.25` — background darkening factor in `RenderTile` (199)
- [ ] **U2 (Low)** `0.08` — ambient occlusion factor per opaque neighbor (221)
- [ ] **U3 (Low)** `0.6` — minimum AO darkening clamp (223)
- [ ] **U4 (Low)** `0.92` — checkerboard dither factor (231)
- [ ] **U5 (Low)** `0.45` — fog-of-war dim factor (236–237)
- [ ] **U6 (Low)** All RGB triples in `TilePalette`, `bloodColor`, `fireColor` — name as constants
- [ ] **R1 (Low)** `isOpaqueTile` switch (182–187) — use `map[TileType]bool` set once at init

### `internal/battle/unit.go`
- [ ] **U1 (Low)** `99` used as infinite-ammo sentinel (`w.AmmoMax < 99`) — extract `const InfAmmoThreshold = 99` (151,154)
- [ ] **U2 (Low)** `3` in `int(dist*3)` — accuracy distance penalty multiplier (160)
- [ ] **U3 (Low)** `10` — minimum accuracy-mod floor (161–162)
- [ ] **U4 (Low)** `5` — minimum hit-chance floor (166–167)
- [ ] **U5 (Low)** `110/100`, `75/100`, `115/100`, `120/100` — crouch/night/marksman/close-combat/steady-aim/overwatch percentage bonuses (170–191)
- [ ] **U6 (Low)** `8` — marksman distance threshold (180)
- [ ] **U7 (Low)** `4` — close-combat distance threshold (183)
- [ ] **U8 (Low)** `1.5` — melee-cover distance threshold raw float (202)
- [ ] **U9 (Low)** `15` — fatal wound chance percent (250)
- [ ] **U10 (Low)** `4` in `dmg / 4` — bleed-rate divisor (252)
- [ ] **U11 (Low)** `5` — max bleed-rate cap (253–254)
- [ ] **U12 (Low)** `7/10` — crouching damage reduction factor (234)
- [ ] **U13 (Low)** `4` in `int(dist) * 4` — TU cost per tile in `MoveTo` duplicates `pathMoveCost` from path.go (299)
- [ ] **R1 (Low)** `FireAt` ~140 lines (135–277) deeply nested conditionals — extract `calcHitChance`, `applyDamage`, `applyFatalWound`, `applyBattleMods`
- [ ] **R2 (Low)** `WeaponDamageType` switch hardcodes weapon ID strings; drive from item definition field instead (281–294)

### `internal/base/equip.go`
- [ ] **M1 (Medium)** Help-bar click zones (317–330) use hardcoded x-coordinate ranges that break if help text changes
- [ ] **M2 (Low)** Magic numbers `20, 24` in `MakeSoldierPortrait(s.Name, s.Armor, 20, 24)` (59)
- [ ] **M3 (Low)** `es.Message` never cleared on navigation — stale messages linger (128)

### `internal/base/manufacture.go`
- [ ] **M1 (Low)** `Selection` indexes both buildable-plans list and queue via `ms.Selection-len(plans)`; if `len(plans)` changes between frames queue indexing drifts (93,194,199)

### `internal/base/planedesigner.go`
- [ ] **M2 (Low)** `CalcPlaneStats` called twice per frame — cache value (116,131)
- [ ] **M3 (Low)** Magic numbers: `45` (col split), `8` (paramY), `1000` (fund display), `20` (click threshold)
- [ ] **R1 (Low)** `bar()` method extracted from screen type to package-level utility (223)

### `internal/base/research.go`
- [ ] **D1 (Low)** Duplicate identical `if entry.status == topicDone` / `else` branches both set `rs.Message = language.String("MSG_CANNOT_RESEARCH")` — dead logic (341–345)
- [ ] **R1 (Low)** Duplicate `StyleGray.Bold(true)` for `topicDone`/`topicLocked` in selected-state styles (114–121)

### `internal/base/weapondesigner.go`
- [ ] **M1 (Low)** `wd.cost()/1000` displays "0" for designs < 1000 (77)
- [ ] **M2 (Low)** Magic numbers: `45` (col split), `8` (paramY), `1000` (fund display), `3` (param column split), `20` (click threshold)
- [ ] **M3 (Low)** `nextID` set from `len(b.CustomWeapons)` once, never incremented — IDs collide if screen reused (19)
- [ ] **R1 (Low)** `CalcDesignStats` called twice per frame (166,313) — reuse return value
- [ ] **R2 (Low)** `renderWeaponArt` 82 lines (86–161) — extract muzzle/barrel/optics/receiver/grip/stock/magazine helpers

### `internal/data/alien_equipment.go`
- [ ] **M1 (Low)** `GetAlienEquipTier` iterates full slice every call — O(n) for sorted list that could use binary search (40–48)

### `internal/data/plane.go`
- [ ] **M1 (Low)** Undocumented magic numbers: `30` (base hull), `5` (hull/length), `20.0` (thrust/engine), `*2` (wing-mass calc) (93,111,117)
- [ ] **R1 (Low)** `RenderPlanePreview` 100 lines (149–249) with 6 sections — extract per-section helpers
- [ ] **R2 (Low)** `CalcPlaneStats` doesn't clamp `cfg.Wingspan`/`cfg.Fuel` while `RenderPlanePreview` does — inconsistent validation (89 vs 161–166)

### `internal/data/procedural_items.go`
- [ ] **M1 (Low)** All stat ranges (`20+rand(40)`, `55+rand(30)`, etc.) undocumented (104–148)
- [ ] **M2 (Low)** `Strength: 10` hardcoded for all procedural weapons (180)

### `internal/data/research.go`
- [ ] **R1 (Low)** `DisplayName` uses `strings.ToUpper`+`strings.ReplaceAll` every call — cache if hot path (21)

### `internal/data/techgen.go`
- [ ] **M1 (Low)** Magic numbers: `40+rng.Intn(30)` autopsy cost, `60+rng.Intn(50)` study cost, `0.85`/`0.30` cost modifier range, floor `10`, `1+rng.Intn(2)` prereqs (53,71,101,103–104,151)
- [ ] **R1 (Low)** `GenerateTechTree` 123 lines (60–183) with 5 phases — extract helpers
- [ ] **R2 (Low)** `isWeaponTech` uses `switch` on string literals — add `IsWeapon` field to `techDef` (185)
- [ ] **R3 (Low)** `checkTechTreeValidity` DFS + fixpoint pass redundant — DFS alone can detect cycles and dead ends (198)

### `internal/data/weapondesign.go`
- [ ] **M1 (Low)** Clamping magic literals: `1` (dmg/range/ammoMax), `10` (acc), `5` (TU/str), `2.5` str/weight ratio (184–203)
- [ ] **D1 (Low)** `AmmoTypes` entries all have `IsAlien: false` — dead field or future-use marker (75–79)

### `internal/engine/camera.go`
- [ ] **U1 (Low)** Hardcoded `decay: 8.0` — extract named constant (21)

### `internal/engine/config.go`
- [ ] **B1 (Medium)** `Config` is mutable global (`var Config = ...`) with no synchronization — `LoadConfig()` writes via `json.Unmarshal` while goroutines may read (47)
- [ ] **D1 (Low)** `WebsiteURL` exported — check for callers; dead export if none (13)
- [ ] **U1 (Low)** Magic numbers: `ActionDelay: 8`, `SfxVolume: 10`, `TouchButtonSize: 4` undocumented

### `internal/engine/control_menu.go`
- [ ] **B1 (Medium)** `HandleMouse` uses stale `cm.screenW` — never calls `SetScreenSize`; hit-detection inaccurate after resize (200)
- [ ] **D1 (Low)** `ControlMenu.ScrollOff` field set but never read (19)
- [ ] **U1 (Low)** Magic numbers: `btnH=3`, `btnMinW=10`, `cols=3/2/1`, `padX/padY=1`, thresholds `60`/`40` (58–68)
- [ ] **R1 (Low)** Repeated `StringWidth` calls in label truncation loop — O(n²) for CJK (167–173)

### `internal/engine/custom_battle.go`
- [ ] **B1 (Medium)** Description word-wrap uses `len([]rune(line))` instead of `StringWidth` — CJK lines overflow (143,153)
- [ ] **R1 (Low)** Scroll offset calc duplicated in `Render` (98–101) and `HandleMouse` (231–234) — extract helper
- [ ] **U1 (Low)** Magic numbers: `leftW-7` name truncation, `h-6` scroll threshold, positions `2`,`3`,`h-3`

### `internal/engine/debrief.go`
- [ ] **B1 (Low)** `BaseDestroyed` + `Won=true` contradictory — title won't show "BASE LOST" because override only in `else` branch (89–95)
- [ ] **U1 (Low)** `d.FundsEarned/1000` — undocumented divisor; extract `const FundsDisplayK` (132)

### `internal/engine/difficulty.go`
- [ ] **U1 (Low)** `500000` starting funds repeated ×3 — name constant (95,103,121)

### `internal/engine/encyclopedia.go`
- [ ] **B3 (Low)** Description text wraps by byte slice — `desc[:end]` splits multi-byte runes for CJK (181–188)
- [ ] **U1 (Low)** Magic number `3` for tab spacing, list positions `5`, info panel height `4` (137,141,178)

### `internal/engine/filters.go`
- [ ] **B1 (Medium)** `math/rand` used without explicit seeding — `rand.Intn(100)` always same sequence, night-vision noise deterministic (37)
- [ ] **U1 (Low)** Luminance coefficients `0.299`, `0.587`, `0.114` — name as ITU-R BT.601 constant (19)
- [ ] **U2 (Low)** Thresholds `128`, `40` in night vision (47,49) and thermal (86,89)

### `internal/engine/game_over.go`
- [ ] **B1 (Low)** Only `Escape` dismisses the screen; inconsistent with `DebriefScreen` which also accepts Enter/Space (36)

### `internal/engine/help.go`
- [ ] **R1 (Low)** `getPages()` called multiple times in `HandleKey` — cache result (195,203,273,278)
- [ ] **U1 (Low)** Hardcoded page count `5` in `"1"`..`"5"` key handlers — fragile if pages array changes (208–218)

### `internal/engine/language_select.go`
- [ ] **U1 (Low)** Many undocumented magic numbers: phase multipliers `0.3`/`0.2`/`2.0`, RGB glow `128,40,180`+amplitude `127,60,75`, column offsets `w/2-26`/`w/2+3`, `startLangY=13`, row spacing `4`, flag width `6` (64–171)

### `internal/engine/layout.go`
- [ ] **B1 (Medium)** Mobile `BattleViewWidth` clamped to min 10 — 10 columns unusably narrow for tactical view (70)
- [ ] **R1 (Low)** `MinSidebarWidth` duplicates `BattleSidebarWidth` min-width logic — dead code or unify (181,51–55)
- [ ] **U1 (Low)** Magic numbers: `30` min sidebar, `10` min battle view, `60`% geo table, `20` min encyclo list, `5` battle view height offset, `3` sidebar Y spacing

### `internal/engine/menu.go`
- [ ] **M1 (Low)** `menuY = 13` hardcodes title line count (6) + gap (1) + subtitle offset (4) — fragile if title changes (121)
- [ ] **M2 (Low)** Star runes `[3]rune{'.','+','*'}` allocated every render — use package-level var (143)
- [ ] **M3 (Low)** `0.55` spread, `0.15` min-dist, `180.0`/`175.0` brightness — magic numbers (149,153–157)
- [ ] **R1 (Low)** Numeric shortcuts `"1"`–`"6"` nearly identical — extract loop/helper (319–346)

### `internal/engine/openbrowser.go`
- [ ] **B1 (Low)** Returned `error` from `cmd.Start()` ignored by caller in `menu.go:382`
- [ ] **R1 (Low)** Single-purpose file for 3-line function — inline into `menu.go`

### `internal/engine/options.go`
- [ ] **M1 (Low)** Index constants `themeIdx=9`, `speedIdx=10`, `volIdx=11`, `langIdx=12` — fragile if `boolOpts` grows/shrinks (103–107)
- [ ] **M2 (Low)** Magic numbers: `baseX = w/2 - 15`, `startY = h/2 - 10`, hit-test widths `30`/`35`, max delay `20`, max volume `10`, flag offset `+7` (108,109,229,235,340,353)
- [ ] **R1 (Low)** `HandleKey` and `HandleMouse` duplicate same volume/speed/theme/language cycling logic (204–241 vs 364–407) — extract helper
- [ ] **R2 (Low)** `cycleTheme` and `cycleLang` structurally identical — extract generic `cycleSlice` helper (274,304)

### `internal/engine/particles.go`
- [ ] **B1 (Low)** `SpawnRain`/`SpawnSnow`/`SpawnDust`/`SpawnEmbers` use `rand.Intn(w)`/`rand.Intn(h)` — panics if width or height is 0 or negative
- [ ] **M1 (Low)** `Gravity = 9.8` is Earth's gravitational constant in m/s², used as pixel-velocity — misleading name/units (23)
- [ ] **M2 (Low)** Over 40 distinct undocumented numeric literals across spawn functions (144–251): RGB triplets, velocities, life ranges, fade speeds
- [ ] **R1 (Low)** `SpawnRain`/`SpawnSnow`/`SpawnDust`/`SpawnEmbers` differ only in parameters — use single parametric spawn helper

### `internal/engine/pixel.go`
- [ ] **M1 (Low)** `'▀'` (U+2580) appears in 3 places — name `const halfBlockRune`
- [ ] **M2 (Low)** `ColorBlackTcell` fallback theme-dependent — transparent pixels get themed "black", not true transparent, surprising (37–69)
- [ ] **R1 (Low)** `DrawPixelImage` and `DrawPixelImageFramed` duplicate `topColor`/`bottomColor` resolution logic — extract `drawHalfBlockCell`

### `internal/engine/screen.go`
- [ ] **B1 (Low)** `StyleDefault`, `StyleHighlight` etc. are package-level vars mutated by `ApplyTheme` — data race if called concurrently with rendering (220–320)
- [ ] **R1 (Low)** `ApplyTheme` 100 lines with near-identical blocks repeated 5× — extract theme config struct, data-driven map (220–320)

### `internal/engine/slotpicker.go`
- [ ] **B1 (Medium)** `HandleMouse` `y < startY+len(sp.Slots)+1` allows clicking "new slot" area in load mode — unexpected dismissal on accidental click below last slot (141)
- [ ] **M1 (Low)** `10` — max save slots magic number (158)
- [ ] **M2 (Low)** Save mode's `newSlot = len(sp.Slots) + 1` assumes contiguous slot numbering — collision if slots sparse (157)
- [ ] **R1 (Low)** `HandleKey` duplicates up/down logic in both `Key` switch (84–94) and `Str` switch (103–113) — extract `moveSelection(delta)`

### `internal/engine/tutorial.go`
- [ ] **B1 (Low)** `wrapDrawString` counts runes rather than display width for word-wrap — CJK (double-width) chars under-count, overflow box (130)
- [ ] **M1 (Low)** `boxW = 62`, `boxH = 14` magic numbers (43–44)
- [ ] **M2 (Low)** `HandleMouse` advances on ANY left click anywhere, not just within dialog — accidental advancement
- [ ] **R1 (Low)** Progress bar rebuilds string every frame — precompute or cache (67–77)

### `internal/engine/vfx.go`
- [ ] **B1 (Medium)** `FrameBuffer.Resize` (29–42) copies cells linearly — corrupts buffer when aspect ratio changes (e.g. 100×50→80×60 reinterprets row data)
- [ ] **M1 (Low)** Magic numbers: radius `1.5`, falloff `0.3` (bloom), falloff `0.4` (directional), cone dot `0.7`, distortion freq `0.05`/`0.1` and amp `2.0` (165,181,206,252,270)
- [ ] **R1 (Low)** `ApplyLightSource` runs two identical nested loops — merge into single loop with flag/counter (116–162)
- [ ] **R2 (Low)** `ApplyLightSource` and `ApplyDirectionalLight` share iteration pattern — extract `forEachCellInRadius`

### `internal/engine/water.go`
- [ ] **M1 (Low)** `waterColors` hard-coded RGB triples — extract `waterPalette` type (11–15)
- [ ] **M2 (Low)** Magic numbers: wave freq `0.5`, color index scale `3`, wave threshold `0.3`, random `≈` chance `5`, FG offsets `40`/`60`/`40` (18,20,29,31,35)

### `internal/engine/webscreen.go`
- [ ] **B1 (Low)** `sgrCode` emits `\x1b[0;1;...` — some legacy terminals treat params after `0` differently; minor portability issue
- [ ] **M1 (Low)** Event queue buffer size `64` (32), pre-allocation multiplier `20` (147)
- [ ] **R1 (Low)** Force-mode block (152–184) and differential-mode block (186–225) both track `prevFg`/`prevBg`/`prevAttr` for SGR — unify

### `internal/geo/interceptor.go`
- [ ] **M1 (Low)** Magic numbers: `Speed: 36`, `HP: 60`, `MaxHP: 60`, `PilotSkill: 50` (42–44,51)
- [ ] **M2 (Low)** `w.FireRate * 4` ammo calc — name constant (47)
- [ ] **M3 (Low)** `i.Range * 3` fuel range multiplier repeated (107,118)
- [ ] **M4 (Low)** `0.3`/`0.5`/`0.7` range fractions in combat modes (134,137,144)
- [ ] **M5 (Low)** `i.MaxHP/3` breakoff threshold (140)
- [ ] **M6 (Low)** `float64(i.Speed) * 0.015` speed conversion duplicated (193,246)
- [ ] **M7 (Low)** `30` max trail length, `1.5` arrival threshold (232,242)
- [ ] **M8 (Low)** `10`/`-10` accuracy mode modifiers, `10`/`100` clamp bounds (303–314)
- [ ] **M9 (Low)** `i.Weapon.Damage/3+1` damage variance (327)
- [ ] **M10 (Low)** `10` critical hit %, `3/2` crit multiplier (330–331)
- [ ] **M11 (Low)** `0.7` effective range ratio threshold, `1.5` falloff multiplier (295–297)
- [ ] **R1 (Low)** `moveTo` and `moveToWithTarget` share ~80% — extract `moveStep` helper (193–214 vs 246–264)

### `internal/geo/transfer.go`
- [ ] **M1 (Low)** Layout literals `2,1`, `2,2`, `2,4`, `4,5`, `h-2`, `h-3`, `h-1` sprinkled throughout `Render` (41–89)
- [ ] **R1 (Low)** Soldiers tab render (49–64) and Items tab render (65–83) share identical list-draw — extract `drawList` helper
- [ ] **R2 (Low)** `sortedStoreItems` reimplements insertion sort — use `sort.Strings` (246)

### `internal/geo/ufo.go`
- [ ] **M1 (Low)** `difficulty * 5` HP bonus clamped at 40 (72–74)
- [ ] **M2 (Low)** `500 + rand.Intn(500)` default `TurnsLeft` (97,142)
- [ ] **M3 (Low)** `0.3` initial progress for `SpawnUFOAtCity` (141)
- [ ] **M4 (Low)** `float64(u.Type.Speed) * 0.002` speed conversion (154)
- [ ] **M5 (Low)** `u.Progress < 0.5` threshold for `CurrentNode` (209)
- [ ] **M6 (Low)** `accuracy := 30`, `5 + rand.Intn(10)` damage in `FireAtInterceptor` (219–220)
- [ ] **R1 (Low)** Difficulty-weighted type selection duplicated in `SpawnUFOOnCities` (62–70) and `SpawnUFOAtCity` (108–115) — extract `pickUFOType`

### `internal/geo/vehicle.go`
- [ ] **D1 (Low)** Empty file (only `package geo`) — remove or fill

### `internal/geo/world.go`
- [ ] **M1 (Low)** City coordinates (50–74) bare literals in `init()`

### `internal/save/save.go`
- [ ] **M1 (Low)** `0644` file perm (139), `10` slot limit (214), `Funds/1000` format (227)
- [ ] **R1 (Low)** `FromBase` (230–300) and `ToBase` (302–380) ~70 lines each — extract soldier/facility/job mapping helpers
- [ ] **R2 (Low)** `ToBase` calls `soldier.NewSoldier(ss.Name)` which rolls random stats only to immediately overwrite — use no-init constructor (328)

### `internal/soldier/perks.go`
- [ ] **M1 (Low)** `"quick_learner"` perk `StatBonuses: StatBonus{}` (empty) — bug or should have real bonuses (105)
- [ ] **R1 (Low)** `HasBattleMod` is O(n*m) — build `map[string]BattleModifier` lookup from `AllPerks` once
- [ ] **R2 (Low)** `PerkNames` and `FormatPerks` have identical iteration — `FormatPerks` should call `PerkNames` then join (186,198)

### `internal/soldier/soldier.go`
- [ ] **B1 (Medium)** `HandlePromotions` RNG seeded with `int64(total)*131 + 7` — same roster size → same promotion sequence; not per-soldier (281)
- [ ] **M1 (Low)** Stat ranges in `NewSoldier`: `20+rand.Intn(6)`, `45+rand.Intn(11)`, `40+rand.Intn(21)`, etc. — undocumented (110–124)
- [ ] **M2 (Low)** `improveStat` thresholds `10, 5, 2` and gains `2+rand.Intn(5)`, `1+rand.Intn(4)` — undocumented (192–202)
- [ ] **M3 (Low)** Bravery gain `10` with `rand.Intn(11)` threshold (224–225)
- [ ] **M4 (Low)** TU/HP/Str post-mission formula `(StatCaps.X-s.X)/10 + 2` (235–249)
- [ ] **R1 (Low)** `PostMission` 57 lines — TU/HP/Strength blocks repeated code (234–250)

### `internal/audio/audio_other.go`
- [ ] **B1 (Medium)** `mixerStream.Read` holds `mu.Lock` while iterating backwards and deleting from `m.buffers` via slice re-slicing — `b--` after deletion skips moved element (38–45)
- [ ] **M1 (Low)** `40` ms buffer size (80), `32767` int16 max (57)

### `internal/audio/audio_windows.go`
- [ ] **B1 (Medium)** `Close()` calls `midiOutClose.Call(handle)` but doesn't set `audioDisabled=true` — goroutines in `playNote` may write to closed handle (50–55,78–82)
- [ ] **B2 (Medium)** Error check compares `err.Error()` string to check success — fragile, relies on English error text (36)
- [ ] **M1 (Low)** MIDI note numbers (70,65,60...), velocities (100,80,60...), durations (30ms,50ms...), channels (0,9), status bytes (0x90,0x80) — dozens of undocumented literals (88–210)

### `internal/audio/pcm_synth.go`
- [ ] **M1 (Low)** `sampleRate = 44100`, `440.0` ref freq, `69` MIDI A4, `12.0` semitone ratio (11–14)
- [ ] **M2 (Low)** Mix ratios: `noise()*0.7/square*0.3`, `noise()*0.5+square*0.5`, etc. — dozens of undocumented mixing weights (52–210)
- [ ] **M3 (Low)** Frequency sweep endpoints: laser 2000→800 Hz, plasma 150→60 Hz, grenade 120→30 Hz (162,174,197)
- [ ] **R1 (Low)** `pad`+`append` pattern repeated ~6× — extract `concatWithPad(a, b, padDur)` helper (89–145)

