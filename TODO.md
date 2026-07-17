# TODO — Game Improvements

Scope: Features and fixes for the battlescape tactical combat system.

---

### `internal/battle/battlescape.go`

### `internal/data/spritebuilder.go`
- [ ] **R1 (Low)** 4 duplicated layer-stamp switch blocks (head/torso/legs/weapon).
  Extract `stampLayer(ch, dst, y, x, allowWeapon)`.
- [ ] **R2 (Low)** 6 repeated biology post-pass loops (lines 973–1029). Use config
  table.
- [ ] **R3 (Low)** `GenerateAlienPixels` ~270-line monolith. Split assembly/shading/
  biology passes.

### `internal/engine/portrait.go`

### `internal/data/procedural.go`
- [ ] **U2 (Low)** Magic tuning numbers (speciesCount 5+rng3, maxRank 1+rng4,
  `rng.Intn(3)==0` synthetic, sense rolls) undocumented. Name constants.
- [ ] **R1 (Low)** Five duplicated `rng.Intn(10)` weighted-roll switches (464–616).
  Add `weightedIndex(rng, cumThresholds)`.
- [ ] **R3 (Low)** `generateVariant` 116-line dense fn (882–997); stat formulas
  repetitive. Split / table-drive.

### `internal/battle/ai.go`
- [x] **R1 (Low)** `disperseFrom`/`moveTowardTargetCover` dup candidate construction.
  Extracted `getOrthogonalCandidates`.
- [x] **R2 (Low)** Move-action append pattern copy-pasted ~8×. Extracted
  `appendMove(...)` helper.
- [x] **R3 (Low)** `Update` ~300 lines, 7-case switch. Split into `handlePatrol`/
  `handleSearch`/`handleRetreat`/`handleSuppress`/`handleFlank`.
- [x] **R4 (Low)** Faction literals 0/1/2 repeated. Named `FactionHuman/Alien/Civilian`.
- [~] **U1 (Low)** Magic TU/sight thresholds. Constants begun (`VisualRangeThreshold`,
  etc.); remaining `dist`/`TU` magic numbers still to be named.

### `internal/engine/game.go`
- [ ] **B2 (Low)** Quit-confirm mouse rects only set in TouchMode (524–546) →
  unclickable on desktop (424–430). Always compute rects or guard handler.
- [x] **B3 (Low)** `lastState` defaults to `StateMenu` (125,338–343) → first screen
  transition may skip `OnScreenChange`. Init to -1.
- [ ] **M1 (Low)** `NewGameWeb` re-inlines full Game literal (193–208) instead of
  calling `newGameWithScreen` → drift risk. Consolidate.
- [x] **M2 (Low)** `GetAlienTypes` fallback returns `&data.AlienTypes[i]` (262–271)
  → mutating caller corrupts shared global. Now copies values.
- [ ] **M3 (Low)** `setupControlMenu` uses positional button indices (589–593,
  633–640) → brittle. Use named refs/map.
- [ ] **U2 (Low)** Magic numbers: boxW/H 46/7, btnW 16, gap 4, frameSleep 16ms,
  keyChan buf 20, Funds 500000, start date 1999-03-01. Name constants.
- [ ] **R1 (Low)** `setupControlMenu` ~140-line switch with dup closures. Add
  `keyBtn(label,hotkey,key,str)` helper.
- [ ] **R3 (Low)** `RegisterScreen`/`SetScreen`/`OpenEncyclopedia` redundant nil
  guards (282–303). Drop.

### `internal/base/facility.go`
- [x] **D3 (Low)** `FacilityInfo.Size` set everywhere but ignored — removed field.
  hardcodes 8-col grid, 250–251). Use or drop.
- [ ] **M1 (Low)** `ChangeInterceptorWeapon` cycle includes "cannon" not sold via
  BuyInterceptor; curIdx -1 silently resets to avalanche (215,232–237).
- [ ] **U1 (Low)** `HireCost=50000`, `interceptorWeaponOrder` undocumented; ammo
  `w.FireRate*4` magic 4.
- [ ] **R1 (Low)** Research-unlock logic dup'd in InterrogateAlien + AdvanceResearch
  (613–630,717–739). Extract `applyUnlocks(topic)`.

### `internal/base/base.go`
- [x] **B2 (Low/Med)** Magic tab count `6` / wrap `>6` (362,381) but valid max idx 5
  → selection 6 momentarily reachable. Replaced with `numTabs` constant.
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
- [x] **B1 (Medium)** `"alien_grenade"` collides between RuleItems weapon (487) and Items loot (590) — already renamed to `alien_grenade_item` in Items.
- [ ] **B2 (Low)** ShortName collisions MSC/PRM/PSI across RuleItems vs Items
  (504–507). Document or avoid cross-index by ShortName.
- [x] **D1 (Low)** `Weapons` map is write-only (AmmoCur never read, 49–51,663–668).
  Documented as reserved for future runtime tracking.
- [ ] **U1 (Low)** `BT_*` constants undocumented (11–23). Add doc.
- [ ] **R1 (Low)** 3 near-identical `DisplayName*` methods (634–661). Extract
  `localizedName(langKey, fallback)`.

### `internal/data/aliens.go`

### `internal/battle/crash.go`

### `internal/battle/gas.go`
- [x] **U1 (Low)** `3` used as max-gas-density sentinel — `MaxGasDensity` const.
- [x] **U2 (Low)** `40`, `20` cover penalty values — `GasCoverDensity3/2`.
- [x] **U3 (Low)** `v.Density <= 1` diffusion threshold — `MinDiffuseDensity`.
- [x] **U4 (Low)** RGB color literals in `Draw` — named `gasSmokeFg/Bg`, `gasPoisonFg/Bg`, `gasRune`.
- [x] **R1 (Low)** `Draw` dup color blocks — extracted `gasStyle(density, gType)`.

### `internal/battle/input.go`
- [x] **U1 (Low)** `3` camera pan — `CamPanStep` const (66–72, 121–142).
- [x] **U2 (Low)** `1` help-bar column offset — `helpBarCol` const (204).

### `internal/battle/modifiers.go`
- [x] **U1 (Low)** Probability denominators — named `nightOpsChance` etc.
- [x] **U2 (Low)** `5`,`3`,`2` accuracy/sight — named `rainAccPenalty` etc.
- [x] **U3 (Low)** `5`,`30`,`20` fire spread — `fireSpreadRain/Wind/Base`.
- [x] **U4 (Low)** `1 + rng.Intn(2)` fog range — `fogRangeMin`/`fogRangeSpan`.
- [x] **R1 (Low)** `rng.Intn(N)==0` ×7 — extracted `roll(rng, chance)`.

### `internal/battle/path.go`
- [x] **U1 (Low)** `15` TU threshold — `MinReactionTU` (battlescape.go).
- [x] **U2 (Low)** `20` distance cutoff — `SightRange`.
- [x] **U3 (Low)** `2`,`3`,`5` multipliers — `ReactionMult/AccDiv/DistPen`.
- [x] **U4 (Low)** `1` min chance — `ReactionMinChance`.

### `internal/battle/terrain.go`
- [x] **U1 (Low)** `0.25` bg darkening — `bgDarkenFactor`.
- [x] **U2 (Low)** `0.08` AO factor — `aoPerNeighbor`.
- [x] **U3 (Low)** `0.6` AO clamp — `aoMinFactor`.
- [x] **U4 (Low)** `0.92` dither — `ditherFactor`.
- [x] **U5 (Low)** `0.45` fog dim — `fogOfWarDim`.
- [x] **U6 (Low)** RGB triples — `tilePalette` (existing), `bloodPalette`, `firePalette`.
- [x] **R1 (Low)** `isOpaqueTile` switch — `opaqueTiles` map at init.

### `internal/battle/unit.go`
- [x] **U1 (Low)** `99` infinite-ammo sentinel — `InfAmmoThreshold`.
- [x] **U2 (Low)** `3` dist accuracy penalty — `distAccPenalty`.
- [x] **U3 (Low)** `10` min accuracy-mod — `minAccMod`.
- [x] **U4 (Low)** `5` min hit-chance — `minHitChance`.
- [x] **U5 (Low)** crouch/night/marksman/close/steady/overwatch bonuses — named consts.
- [x] **U6 (Low)** `8` marksman range — `marksmanDist`.
- [x] **U7 (Low)** `4` close-combat range — `closeCombatDist`.
- [x] **U8 (Low)** `1.5` melee cover dist — `meleeCoverDist`.
- [x] **U9 (Low)** `15` fatal wound chance — `fatalWoundChance`.
- [x] **U10 (Low)** `4` bleed divisor — `bleedDivisor`.
- [x] **U11 (Low)** `5` max bleed — `maxBleedRate`.
- [x] **U12 (Low)** `7/10` crouch dmg reduce — `crouchDmgReduce`.
- [x] **U13 (Low)** `4` TU/tile — `moveTUCostPerTile`.
- [ ] **R1 (Low)** `FireAt` nested conditionals — extract helpers (deferred).
- [ ] **R2 (Low)** `WeaponDamageType` switch hardcoded strings — drive from item field (deferred).

### `internal/base/equip.go`
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

- [x] **U1 (Low)** Hardcoded `decay: 8.0` — extracted `shakeDecay` constant.

### `internal/engine/config.go`
- [x] **D1 (Low)** `WebsiteURL` exported — has callers in menu.go, kept.
- [ ] **U1 (Low)** Magic numbers: `ActionDelay: 8`, `SfxVolume: 10`, `TouchButtonSize: 4` undocumented

### `internal/engine/control_menu.go`
- [ ] **D1 (Low)** `ControlMenu.ScrollOff` field set but never read (19)
- [ ] **U1 (Low)** Magic numbers: `btnH=3`, `btnMinW=10`, `cols=3/2/1`, `padX/padY=1`, thresholds `60`/`40` (58–68)
- [ ] **R1 (Low)** Repeated `StringWidth` calls in label truncation loop — O(n²) for CJK (167–173)

### `internal/engine/custom_battle.go`
- [ ] **R1 (Low)** Scroll offset calc duplicated in `Render` (98–101) and `HandleMouse` (231–234) — extract helper
- [ ] **U1 (Low)** Magic numbers: `leftW-7` name truncation, `h-6` scroll threshold, positions `2`,`3`,`h-3`

### `internal/engine/debrief.go`
- [ ] **B1 (Low)** `BaseDestroyed` + `Won=true` contradictory — title won't show "BASE LOST" because override only in `else` branch (89–95)
- [ ] **U1 (Low)** `d.FundsEarned/1000` — undocumented divisor; extract `const FundsDisplayK` (132)
### `internal/engine/difficulty.go`

- [x] **U1 (Low)** `500000` starting funds repeated ×3 — extracted `startingFunds` constant.

### `internal/engine/encyclopedia.go`
- [ ] **B3 (Low)** Description text wraps by byte slice — `desc[:end]` splits multi-byte runes for CJK (181–188)
- [ ] **U1 (Low)** Magic number `3` for tab spacing, list positions `5`, info panel height `4` (137,141,178)

### `internal/engine/filters.go`
- [ ] **U1 (Low)** Luminance coefficients `0.299`, `0.587`, `0.114` — name as ITU-R BT.601 constant (19)
- [ ] **U2 (Low)** Thresholds `128`, `40` in night vision (47,49) and thermal (86,89)

### `internal/engine/game_over.go`
- [x] **B1 (Low)** Only `Escape` dismisses the screen; now also accepts Enter/Space.

### `internal/engine/help.go`
- [ ] **R1 (Low)** `getPages()` called multiple times in `HandleKey` — cache result (195,203,273,278)
- [ ] **U1 (Low)** Hardcoded page count `5` in `"1"`..`"5"` key handlers — fragile if pages array changes (208–218)

### `internal/engine/language_select.go`
- [ ] **U1 (Low)** Many undocumented magic numbers: phase multipliers `0.3`/`0.2`/`2.0`, RGB glow `128,40,180`+amplitude `127,60,75`, column offsets `w/2-26`/`w/2+3`, `startLangY=13`, row spacing `4`, flag width `6` (64–171)

### `internal/engine/layout.go`
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
- [ ] **M1 (Low)** `10` — max save slots magic number (158)
- [ ] **M2 (Low)** Save mode's `newSlot = len(sp.Slots) + 1` assumes contiguous slot numbering — collision if slots sparse (157)
- [ ] **R1 (Low)** `HandleKey` duplicates up/down logic in both `Key` switch (84–94) and `Str` switch (103–113) — extract `moveSelection(delta)`

### `internal/engine/tutorial.go`
- [ ] **B1 (Low)** `wrapDrawString` counts runes rather than display width for word-wrap — CJK (double-width) chars under-count, overflow box (130)
- [ ] **M1 (Low)** `boxW = 62`, `boxH = 14` magic numbers (43–44)
- [ ] **M2 (Low)** `HandleMouse` advances on ANY left click anywhere, not just within dialog — accidental advancement
- [ ] **R1 (Low)** Progress bar rebuilds string every frame — precompute or cache (67–77)

### `internal/engine/vfx.go`
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
- [x] **D1 (Low)** Empty file (only `package geo`) — removed.

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
- [ ] **M1 (Low)** Stat ranges in `NewSoldier`: `20+rand.Intn(6)`, `45+rand.Intn(11)`, `40+rand.Intn(21)`, etc. — undocumented (110–124)
- [ ] **M2 (Low)** `improveStat` thresholds `10, 5, 2` and gains `2+rand.Intn(5)`, `1+rand.Intn(4)` — undocumented (192–202)
- [ ] **M3 (Low)** Bravery gain `10` with `rand.Intn(11)` threshold (224–225)
- [ ] **M4 (Low)** TU/HP/Str post-mission formula `(StatCaps.X-s.X)/10 + 2` (235–249)
- [ ] **R1 (Low)** `PostMission` 57 lines — TU/HP/Strength blocks repeated code (234–250)

### `internal/audio/audio_other.go`
- [x] **M1 (Low)** `40` ms buffer size (80), `32767` int16 max (57) — named `otoBufferMS`/`int16Max`.

### `internal/audio/audio_windows.go`
- [x] **M1 (Low)** MIDI literals named: `midiNoteOn`/`midiNoteOff`, `midiPercCh`, `midiMinVol`, `midiMapperID`.

### `internal/audio/pcm_synth.go`
- [x] **M1 (Low)** `sampleRate`/`refFreqA4`/`midiA4`/`semitoneRatio` named constants.
- [x] **M2 (Low)** Mix ratios documented inline (noise/square weights per effect).
- [x] **M3 (Low)** Sweep endpoints named: laser/plasma/grenade `SweepStart`/`SweepEnd`.
- [x] **R1 (Low)** `concatWithPad` helper replaces repeated pad+append pattern.

