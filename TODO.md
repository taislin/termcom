# TODO — Game Improvements
Scope: Features and fixes for the battlescape tactical combat system.
---
### `internal/data/spritebuilder.go`
- [ ] **R3 (Low)** `GenerateAlienPixels` ~270-line monolith. Split assembly/shading/
  biology passes.
### `internal/battle/ai.go`
- [ ] **U1 (Low)** Magic TU/sight thresholds. Constants begun (`VisualRangeThreshold`,
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
### `internal/base/base.go`
- [ ] **M1 (Low)** `HandleMouse` vs `HandleKey` hotkey dispatch dup'd (465–551 vs
  349–456); they already differ ('g' opens different designers). Extract
  `dispatchHotkey`.
- [ ] **R1 (Low)** 4 preset generators dup'd coordinate lists (312–412). Fold into
  tierConfigs/seed tables.
### `internal/base/planedesigner.go`
- [ ] **M2 (Low)** `CalcPlaneStats` called twice per frame — cache value (116,131)
### `internal/base/weapondesigner.go`
- [ ] **M3 (Low)** `nextID` set from `len(b.CustomWeapons)` once, never incremented — IDs collide if screen reused (19)
### `internal/data/alien_equipment.go`
- [ ] **M1 (Low)** `GetAlienEquipTier` iterates full slice every call — O(n) for sorted list that could use binary search (40–48)
### `internal/data/plane.go`
- [ ] **M1 (Low)** Undocumented magic numbers: `30` (base hull), `5` (hull/length), `20.0` (thrust/engine), `*2` (wing-mass calc) (93,111,117)
### `internal/data/techgen.go`
- [ ] **M1 (Low)** Magic numbers: `40+rng.Intn(30)` autopsy cost, `60+rng.Intn(50)` study cost, `0.85`/`0.30` cost modifier range, floor `10`, `1+rng.Intn(2)` prereqs (53,71,101,103–104,151)
### `internal/engine/custom_battle.go`
- [ ] **U1 (Low)** Magic numbers: `leftW-7` name truncation, `h-6` scroll threshold, positions `2`,`3`,`h-3`
### `internal/engine/layout.go`
- [ ] **U1 (Low)** Magic numbers: `30` min sidebar, `10` min battle view, `60`% geo table, `20` min encyclo list, `5` battle view height offset, `3` sidebar Y spacing
### `internal/engine/menu.go`
- [ ] **M1 (Low)** `menuY = 13` hardcodes title line count (6) + gap (1) + subtitle offset (4) — fragile if title changes (121)
- [ ] **M3 (Low)** `0.55` spread, `0.15` min-dist, `180.0`/`175.0` brightness — magic numbers (149,153–157)
### `internal/engine/options.go`
- [ ] **M2 (Low)** Magic numbers: `baseX = w/2 - 15`, `startY = h/2 - 10`, hit-test widths `30`/`35`, max delay `20`, max volume `10`, flag offset `+7` (108,109,229,235,340,353)
### `internal/engine/particles.go`
- [ ] **M2 (Low)** Over 40 distinct undocumented numeric literals across spawn functions (144–251): RGB triplets, velocities, life ranges, fade speeds
### `internal/engine/pixel.go`
- [ ] **M2 (Low)** `ColorBlackTcell` fallback theme-dependent — transparent pixels get themed "black", not true transparent, surprising (37–69)
### `internal/engine/slotpicker.go`
- [ ] **M2 (Low)** Save mode's `newSlot = len(sp.Slots) + 1` assumes contiguous slot numbering — collision if slots sparse (157)
### `internal/engine/tutorial.go`
- [ ] **M1 (Low)** `boxW = 62`, `boxH = 14` magic numbers (43–44)
- [ ] **M2 (Low)** `HandleMouse` advances on ANY left click anywhere, not just within dialog — accidental advancement
### `internal/engine/vfx.go`
- [ ] **M1 (Low)** Magic numbers: radius `1.5`, falloff `0.3` (bloom), falloff `0.4` (directional), cone dot `0.7`, distortion freq `0.05`/`0.1` and amp `2.0` (165,181,206,252,270)
### `internal/engine/water.go`
- [ ] **M1 (Low)** `waterColors` hard-coded RGB triples — extract `waterPalette` type (11–15)
### `internal/engine/webscreen.go`
- [ ] **M1 (Low)** Event queue buffer size `64` (32), pre-allocation multiplier `20` (147)
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
### `internal/geo/transfer.go`
- [ ] **M1 (Low)** Layout literals `2,1`, `2,2`, `2,4`, `4,5`, `h-2`, `h-3`, `h-1` sprinkled throughout `Render` (41–89)
### `internal/geo/ufo.go`
- [ ] **M1 (Low)** `difficulty * 5` HP bonus clamped at 40 (72–74)
- [ ] **M2 (Low)** `500 + rand.Intn(500)` default `TurnsLeft` (97,142)
- [ ] **M3 (Low)** `0.3` initial progress for `SpawnUFOAtCity` (141)
- [ ] **M4 (Low)** `float64(u.Type.Speed) * 0.002` speed conversion (154)
- [ ] **M5 (Low)** `u.Progress < 0.5` threshold for `CurrentNode` (209)
- [ ] **M6 (Low)** `accuracy := 30`, `5 + rand.Intn(10)` damage in `FireAtInterceptor` (219–220)
### `internal/geo/world.go`
- [ ] **M1 (Low)** City coordinates (50–74) bare literals in `init()`
### `internal/save/save.go`
- [ ] **R2 (Low)** `ToBase` calls `soldier.NewSoldier(ss.Name)` which rolls random stats only to immediately overwrite — use no-init constructor (328)
