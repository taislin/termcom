# TODO — Game Improvements
Scope: Features and fixes for the battlescape tactical combat system.
---
### `internal/data/spritebuilder.go`
- [ ] **R1 (Low)** 4 duplicated layer-stamp switch blocks (head/torso/legs/weapon).
  Extract `stampLayer(ch, dst, y, x, allowWeapon)`.
- [ ] **R2 (Low)** 6 repeated biology post-pass loops (lines 973–1029). Use config
  table.
- [ ] **R3 (Low)** `GenerateAlienPixels` ~270-line monolith. Split assembly/shading/
  biology passes.
### `internal/data/procedural.go`
- [ ] **R1 (Low)** Five duplicated `rng.Intn(10)` weighted-roll switches (464–616).
  Add `weightedIndex(rng, cumThresholds)`.
- [ ] **R3 (Low)** `generateVariant` 116-line dense fn (882–997); stat formulas
  repetitive. Split / table-drive.
### `internal/battle/ai.go`
- [~] **U1 (Low)** Magic TU/sight thresholds. Constants begun (`VisualRangeThreshold`,
  etc.); remaining `dist`/`TU` magic numbers still to be named.
### `internal/engine/game.go`
- [ ] **M1 (Low)** `NewGameWeb` re-inlines full Game literal (193–208) instead of
  calling `newGameWithScreen` → drift risk. Consolidate.
- [ ] **M3 (Low)** `setupControlMenu` uses positional button indices (589–593,
  633–640) → brittle. Use named refs/map.
- [ ] **U2 (Low)** Magic numbers: boxW/H 46/7, btnW 16, gap 4, frameSleep 16ms,
  keyChan buf 20, Funds 500000, start date 1999-03-01. Name constants.
- [ ] **R1 (Low)** `setupControlMenu` ~140-line switch with dup closures. Add
  `keyBtn(label,hotkey,key,str)` helper.
- [ ] **R3 (Low)** `RegisterScreen`/`SetScreen`/`OpenEncyclopedia` redundant nil
  guards (282–303). Drop.
### `internal/base/facility.go`
- [ ] **M1 (Low)** `ChangeInterceptorWeapon` cycle includes "cannon" not sold via
  BuyInterceptor; curIdx -1 silently resets to avalanche (215,232–237).
- [ ] **U1 (Low)** `HireCost=50000`, `interceptorWeaponOrder` undocumented; ammo
  `w.FireRate*4` magic 4.
- [ ] **R1 (Low)** Research-unlock logic dup'd in InterrogateAlien + AdvanceResearch
  (613–630,717–739). Extract `applyUnlocks(topic)`.
### `internal/base/base.go`
- [ ] **M1 (Low)** `HandleMouse` vs `HandleKey` hotkey dispatch dup'd (465–551 vs
  349–456); they already differ ('g' opens different designers). Extract
  `dispatchHotkey`.
- [ ] **U1 (Low)** `BaseScreen` fields lack doc; `storesItems` order-dependent on
  render. Comment.
- [ ] **R1 (Low)** 4 preset generators dup'd coordinate lists (312–412). Fold into
  tierConfigs/seed tables.
### `internal/battle/unit.go`
- [ ] **R1 (Low)** `FireAt` nested conditionals — extract helpers (deferred).
- [ ] **R2 (Low)** `WeaponDamageType` switch hardcoded strings — drive from item field (deferred).
### `internal/base/planedesigner.go`
- [ ] **M2 (Low)** `CalcPlaneStats` called twice per frame — cache value (116,131)
- [ ] **R1 (Low)** `bar()` method extracted from screen type to package-level utility (223)
### `internal/base/weapondesigner.go`
- [ ] **M3 (Low)** `nextID` set from `len(b.CustomWeapons)` once, never incremented — IDs collide if screen reused (19)
- [ ] **R1 (Low)** `CalcDesignStats` called twice per frame (166,313) — reuse return value
- [ ] **R2 (Low)** `renderWeaponArt` 82 lines (86–161) — extract muzzle/barrel/optics/receiver/grip/stock/magazine helpers
### `internal/data/alien_equipment.go`
- [ ] **M1 (Low)** `GetAlienEquipTier` iterates full slice every call — O(n) for sorted list that could use binary search (40–48)
### `internal/data/plane.go`
- [ ] **M1 (Low)** Undocumented magic numbers: `30` (base hull), `5` (hull/length), `20.0` (thrust/engine), `*2` (wing-mass calc) (93,111,117)
- [ ] **R1 (Low)** `RenderPlanePreview` 100 lines (149–249) with 6 sections — extract per-section helpers
- [ ] **R2 (Low)** `CalcPlaneStats` doesn't clamp `cfg.Wingspan`/`cfg.Fuel` while `RenderPlanePreview` does — inconsistent validation (89 vs 161–166)
### `internal/data/research.go`
- [ ] **R1 (Low)** `DisplayName` uses `strings.ToUpper`+`strings.ReplaceAll` every call — cache if hot path (21)
### `internal/data/techgen.go`
- [ ] **M1 (Low)** Magic numbers: `40+rng.Intn(30)` autopsy cost, `60+rng.Intn(50)` study cost, `0.85`/`0.30` cost modifier range, floor `10`, `1+rng.Intn(2)` prereqs (53,71,101,103–104,151)
- [ ] **R1 (Low)** `GenerateTechTree` 123 lines (60–183) with 5 phases — extract helpers
- [ ] **R2 (Low)** `isWeaponTech` uses `switch` on string literals — add `IsWeapon` field to `techDef` (185)
- [ ] **R3 (Low)** `checkTechTreeValidity` DFS + fixpoint pass redundant — DFS alone can detect cycles and dead ends (198)
### `internal/engine/control_menu.go`
- [ ] **U1 (Low)** Magic numbers: `btnH=3`, `btnMinW=10`, `cols=3/2/1`, `padX/padY=1`, thresholds `60`/`40` (58–68)
- [ ] **R1 (Low)** Repeated `StringWidth` calls in label truncation loop — O(n²) for CJK (167–173)
### `internal/engine/custom_battle.go`
- [ ] **R1 (Low)** Scroll offset calc duplicated in `Render` (98–101) and `HandleMouse` (231–234) — extract helper
- [ ] **U1 (Low)** Magic numbers: `leftW-7` name truncation, `h-6` scroll threshold, positions `2`,`3`,`h-3`
### `internal/engine/debrief.go`
- [ ] **B1 (Low)** `BaseDestroyed` + `Won=true` contradictory — title won't show "BASE LOST" because override only in `else` branch (89–95)
### `internal/engine/language_select.go`
- [ ] **U1 (Low)** Many undocumented magic numbers: phase multipliers `0.3`/`0.2`/`2.0`, RGB glow `128,40,180`+amplitude `127,60,75`, column offsets `w/2-26`/`w/2+3`, `startLangY=13`, row spacing `4`, flag width `6` (64–171)
### `internal/engine/layout.go`
- [ ] **R1 (Low)** `MinSidebarWidth` duplicates `BattleSidebarWidth` min-width logic — dead code or unify (181,51–55)
- [ ] **U1 (Low)** Magic numbers: `30` min sidebar, `10` min battle view, `60`% geo table, `20` min encyclo list, `5` battle view height offset, `3` sidebar Y spacing
### `internal/engine/menu.go`
- [ ] **M1 (Low)** `menuY = 13` hardcodes title line count (6) + gap (1) + subtitle offset (4) — fragile if title changes (121)
- [ ] **M3 (Low)** `0.55` spread, `0.15` min-dist, `180.0`/`175.0` brightness — magic numbers (149,153–157)
- [ ] **R1 (Low)** Numeric shortcuts `"1"`–`"6"` nearly identical — extract loop/helper (319–346)
### `internal/engine/openbrowser.go`
- [ ] **R1 (Low)** Single-purpose file for 3-line function — inline into `menu.go`
### `internal/engine/options.go`
- [ ] **M2 (Low)** Magic numbers: `baseX = w/2 - 15`, `startY = h/2 - 10`, hit-test widths `30`/`35`, max delay `20`, max volume `10`, flag offset `+7` (108,109,229,235,340,353)
- [ ] **R1 (Low)** `HandleKey` and `HandleMouse` duplicate same volume/speed/theme/language cycling logic (204–241 vs 364–407) — extract helper
- [ ] **R2 (Low)** `cycleTheme` and `cycleLang` structurally identical — extract generic `cycleSlice` helper (274,304)
### `internal/engine/particles.go`
- [ ] **M2 (Low)** Over 40 distinct undocumented numeric literals across spawn functions (144–251): RGB triplets, velocities, life ranges, fade speeds
- [ ] **R1 (Low)** `SpawnRain`/`SpawnSnow`/`SpawnDust`/`SpawnEmbers` differ only in parameters — use single parametric spawn helper
### `internal/engine/pixel.go`
- [ ] **M2 (Low)** `ColorBlackTcell` fallback theme-dependent — transparent pixels get themed "black", not true transparent, surprising (37–69)
- [ ] **R1 (Low)** `DrawPixelImage` and `DrawPixelImageFramed` duplicate `topColor`/`bottomColor` resolution logic — extract `drawHalfBlockCell`
### `internal/engine/screen.go`
- [ ] **B1 (Low)** `StyleDefault`, `StyleHighlight` etc. are package-level vars mutated by `ApplyTheme` — data race if called concurrently with rendering (220–320)
- [ ] **R1 (Low)** `ApplyTheme` 100 lines with near-identical blocks repeated 5× — extract theme config struct, data-driven map (220–320)
### `internal/engine/slotpicker.go`
- [ ] **M2 (Low)** Save mode's `newSlot = len(sp.Slots) + 1` assumes contiguous slot numbering — collision if slots sparse (157)
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
### `internal/geo/world.go`
- [ ] **M1 (Low)** City coordinates (50–74) bare literals in `init()`
### `internal/save/save.go`
- [ ] **R1 (Low)** `FromBase` (230–300) and `ToBase` (302–380) ~70 lines each — extract soldier/facility/job mapping helpers
- [ ] **R2 (Low)** `ToBase` calls `soldier.NewSoldier(ss.Name)` which rolls random stats only to immediately overwrite — use no-init constructor (328)
### `internal/soldier/perks.go`
- [ ] **R1 (Low)** `HasBattleMod` is O(n*m) — build `map[string]BattleModifier` lookup from `AllPerks` once
- [ ] **R2 (Low)** `PerkNames` and `FormatPerks` have identical iteration — `FormatPerks` should call `PerkNames` then join (186,198)
### `internal/soldier/soldier.go`
- [ ] **R1 (Low)** `PostMission` 57 lines — TU/HP/Strength blocks repeated code (234–250)

