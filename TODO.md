# TODO — Soldier Progression Port (OpenXcom OXCE+ → ycom)

Scope: faithful vanilla X-COM soldier progression + best OXCE+ extras, ported into
`github.com/taislin/termcom`. Grounded in the current codebase state (see "Current
State" below). This file is the implementation plan only — nothing here is code yet.

References are into the OpenXcom OXCE+ source at `_assets/OpenXcom-oxce-plus/src/`
and the ycom Go packages under `internal/`.

---

## Current Codebase State (baseline)

Already present:
- `internal/soldier/soldier.go` — `GainXP(kills) *Perk`: kill-threshold rank bumps
  (+2 HP, +1 TU, +2 Acc, +1 Str/React) + `RollPerk`/`ApplyPerk`. Still **kill-based**.
- `internal/soldier/perks.go` — `Perk`, `StatBonus`, `BattleModifier` (Marksman,
  Overwatch, CloseCombat, SteadyAim, …), `RollPerk`, `ApplyPerk`, `HasBattleMod`.
  **`HasBattleMod` is never called in combat → battle perks are inert.**
- `internal/soldier/soldier.go` — `Fatigue int`; `CanDeploy()` = `Wounds==0 && Fatigue==0`.
- `internal/base/facility.go` — daily `Wounds--` heal (+`AdjacentHealBonus`) and
  `Fatigue--`; daily `PsiSkill++` to 80.
- `internal/battle/battlescape.go` — `finishBattle` maps HP loss → `Wounds = dmg*3`
  (cap 30) and awards `Fatigue += 1+turn/3` (cap 5); reaction fire at ~`:709`;
  psi attack at ~`:2295`.
- `internal/save/save.go` — persists `Rank, Wounds, Fatigue, Perks, PsiSkill`
  (transient XP fields need NO schema change).
- Module path: **`github.com/taislin/termcom`**.

Absent (this plan's targets):
- Per-action XP counters, `improveStat()` buckets, halo growth.
- Headcount-based rank promotion (replacing kill thresholds).
- Fatal wounds / bleeding in-battle.
- Minimal morale (only as bravery-growth driver).
- Probability-based psi-lab training.
- `BattleMod` perks wired into `FireAt`/reaction fire.

---

## Phase 1 — Per-action XP + `improveStat` + halo growth
Ref: `BattleUnit.cpp:3948-4179`, `BattleUnit.cpp:4042-4179`

- [ ] **1.1** `soldier.go`: add transient XP counters
  `ExpFiring, ExpThrowing, ExpReactions, ExpBravery, ExpPsiSkill, ExpPsiStr, ExpMelee int`
  and `GainedXP bool` to `Soldier`.
- [ ] **1.2** `soldier.go`: add package-level stat caps
  (TU 80, HP 60, Acc 120, React 100, Brave 100, Str 70, Psi 100, Melee 120).
- [ ] **1.3** `soldier.go`: add `AddFiringExp / AddThrowingExp / AddReactionsExp /
  AddBraveryExp / AddPsiSkillExp / AddPsiStrExp / AddMeleeExp` methods
  (increment counter + set `GainedXP = true`).
- [ ] **1.4** `soldier.go`: add `improveStat(exp int) int` faithful to OXCE:
  `>10 → RNG(2,6)`, `>5 → RNG(1,4)`, `>2 → RNG(1,3)`, `>0 → RNG(0,1)`.
- [ ] **1.5** `soldier.go`: add `PostMission()`:
  - per stat with `exp>0` and `<cap` → apply `improveStat`;
  - bravery `+10` if `ExpBravery > RNG(0,10)`;
  - **halo** if `GainedXP`: auto Rookie→Squaddie; `TU += RNG(0,(capTU-TU)/10+2)`;
    `HP += RNG(0,(capHP-HP)/10+2)`; `Strength += RNG(0,(capStr-Str)/10+2)`.
- [ ] **1.6** `battle/unit.go` `FireAt`: on hit award `AddFiringExp()`
  (grenade path → `AddThrowingExp`, melee path → `AddMeleeExp`).
- [ ] **1.7** `battle/battlescape.go` reaction fire (~`:709`): on hit award `AddReactionsExp()`.
- [ ] **1.8** `battle/battlescape.go` psi attack (~`:2295`): on success award `AddPsiSkillExp()`.
- [ ] **1.9** Replace kill-based `GainXP(xp)` calls (`battlescape.go:1085`,
  `geoscape.go:1436`) with `u.Soldier.PostMission()` + `Missions++`; keep `Kills += n`.

## Phase 2 — Ranks by headcount (preserve perks)
Ref: `RankCount.cpp:68-87`, `SavedGame.cpp:2348-2390`

- [ ] **2.1** `soldier.go`: add `HandlePromotions(roster []*Soldier)`:
  count per rank; openings scaled to 8-rank ladder
  (e.g. thresholds `[_,_,4,8,14,22,30,40]` for Corporal→Colonel by total roster);
  promote highest-Kills/mission soldiers.
- [ ] **2.2** Each promotion rolls `RollPerk`/`ApplyPerk` (reuse `perks.go`) so the
  existing perk-on-rank behavior is preserved.
- [ ] **2.3** Call `HandlePromotions` from `finishBattle` (post-mission, full roster).
- [ ] **2.4** Remove rank-bumping from `GainXP` (keep it only for kill/record tally,
  or delete and update callers).
- [ ] **2.5** Update tests: `soldier_test.go:53` (`TestGainXP`),
  `battle_integration_test.go:316`, `bench_test.go:132` to new model.

## Phase 3 — Fatal wounds & bleeding
Ref: `BattleUnit.cpp:1555`, `BattleUnit.cpp:2654-2689`

- [ ] **3.1** `battle/unit.go`: add `FatalWounds int` + `BleedRate int` to `Unit`.
- [ ] **3.2** `battle/unit.go` `FireAt`: on damage, small chance to add a fatal
  wound; apply overkill clamp `HP >= -Overkill*MaxHP`.
- [ ] **3.3** `battle/battlescape.go` turn loop: apply bleed tick `HP -= BleedRate`
  each turn; mark unit dead at `HP <= 0`.
- [ ] **3.4** `battle/battlescape.go` `finishBattle`: fold fatal wounds into existing
  `Wounds` (days) field, e.g. `Wounds = min(30, (MaxHP-HP)*3 + FatalWounds*k)`
  so `save.go` / `save_test.go` semantics stay valid.

## Phase 4 — Bravery-on-panic (minimal morale)
Ref: `BattleUnit.cpp:2729-2757`

- [ ] **4.1** `battle/unit.go`: add `Morale int` to `Unit` (default 100).
- [ ] **4.2** `battle/battlescape.go`: per-turn morale recovery + panic roll
  `RNG::percent(100 - 2*morale)`.
- [ ] **4.3** On **avoiding** panic → `u.Soldier.AddBraveryExp()`.
  (Full morale/panic behavior is the separate Battlescape tier-1 task; this only
  provides the bravery-growth trigger.)

## Phase 5 — Psi-lab training refinement
Ref: `Soldier.cpp:1374-1409`

- [ ] **5.1** `base/facility.go` `AdvanceDay`: replace daily `PsiSkill++` with
  probability model: `if PsiSkill>0 && RNG(100) < 8*100/threshold && PsiSkill<cap { PsiSkill++ }`
  (+ optional `PsiStr` improvement behind a flag).
- [ ] **5.2** Update `facility_test.go:660` (`TestPsiLabTraining`) to the
  probabilistic model (currently asserts deterministic +1/day, cap 80).

## Phase 6 — Best OXCE+ extras

- [ ] **6.1** Wire `HasBattleMod` into combat (fixes currently-dead perks):
  `unit.go FireAt` + reaction fire apply Marksman (`range>8` +15% acc),
  CloseCombat (`≤4` +15%), Overwatch (+20% RF acc), SteadyAim (no-move +10%),
  etc. Low effort, makes shipped Perk system functional.
- [ ] **6.2** Death/memorial: add `KIA`/death record on `HP<=0` in `finishBattle`;
  memorial list in game state.
- [ ] **6.3** (Optional stretch) Mana (12th stat) `Mod.h:252`: add `Mana`/`MaxMana`,
  daily training, wound-threshold gating. Flag high-effort (needs weapon/recovery
  wiring).
- [ ] **6.4** (Noted, out of scope) Equipment loadout templates, soldier
  transformations, nationalities/StatStrings — belong to base/inventory tasks.

---

## Wiring & verification

- [ ] Call `HandlePromotions` from `finishBattle` (post-mission, full roster known).
- [ ] Confirm new XP/caps are transient (applied in `PostMission` before save) →
  **no `save.go` schema change required**.
- [ ] Add unit tests: `improveStat` bucket boundaries; `PostMission` halo growth;
  headcount `HandlePromotions`; psi probability; wound/fatigue recovery.
- [ ] `gofmt -w` all touched files.
- [ ] `make test` and `make lint` pass; update `docs/manual.md` for any
  balance/mechanic changes (per AGENTS.md).

## Dependency notes
- Phase 4 requires a minimal `Morale` field on `Unit` — recommended (tiny); it is
  the only OXCE driver for the Bravery stat.
- Phase 6.1 (`HasBattleMod` wiring) is recommended to make the already-shipped Perk
  system actually matter in combat.
- Phase 6.3 (Mana) is the highest-effort item; treat as optional.

---

# Mobile-Friendly Port

Scope: enable playable mobile experience in the browser via touch, collapsible sidebars,
and an expandable on-screen control menu. Everything lives in the Go engine (shared
between terminal + web); mobile-only behaviour activates when `cols < 100` (auto-detected
on web connect). The control menu only appears when the user taps the hamburger `☰` toggle.

---

## Phase 1 — Touch input pipeline
| # | Task | File(s) | Detail |
|---|------|---------|--------|
| 1 | Config: `TouchMode bool`, `TouchButtonSize int` | `internal/engine/config.go` | New JSON fields; default button size 4. `TouchMode` set automatically on web connect when `cols < 100`, also toggleable in Options. |
| 2 | `Game.InjectMouse(ev)` method | `internal/engine/game.go` | Matches existing `InjectKey`/`InjectResize`; posts `*tcell.EventMouse` to `g.keyChan`. |
| 3 | WebSocket message type `"mouse"` | `cmd/webserver/main.go`, `web/server.go` | New `MouseHandler func(x, y int, button string)`; creates `tcell.NewEventMouse` and calls `g.InjectMouse`. |
| 4 | Browser touch → cell coords | `web/index.html` | `touchstart`/`touchend` listeners → `cellX = floor(touchX / (elW / cols))`, `cellY = floor(touchY / (elH / rows))` → send `{"type":"mouse","cols":x,"rows":y,"data":"left"}`. |
| 5 | Smaller default web terminal | `cmd/webserver/main.go` | Initialize at 80×40 on mobile connect instead of 220×50. |

## Phase 2 — Responsive collapsible layouts
| # | Task | File(s) | Detail |
|---|------|---------|--------|
| 6 | `LayoutManager` helper | `internal/engine/layout.go` (new) | Modes: `full`, `sidebar-collapsed`, `mobile`; holds `SidebarVisible bool`, `ToggleSidebar()`; provides viewport/sidebar/control-area dimensions per mode. |
| 7 | **Battlescape**: collapsible sidebar | `internal/battle/battlescape.go` | Collapsed: viewport full-width, unit info → 2-line top banner. Uses `LayoutManager.ViewportSize()`. |
| 8 | **Geoscape**: hide minimap | `internal/geo/geoscape.go` | Table full-width; minimap collapsed → mini status bar showing only ufo/event count. |
| 9 | **Equip/Research/Manufacture**: single-column | `internal/base/equip.go`, `research.go`, `manufacture.go` | Scrollable vertical layout instead of left/right split; selected item shows detail popup. |
| 10 | **Encyclopedia**: stack panels | `internal/engine/encyclopedia.go` | List top, detail bottom (or full-screen detail). |
| 11 | **CustomBattle**: stack panels | `internal/engine/custom_battle.go` | Mission list top, detail bottom. |
| 12 | Help bar: enlarged touch zones | All screens with help bar | In `mobile` mode, each `[hotkey]` segment gets min 3-cell-wide hit area. |

## Phase 3 — Expandable control menu
| # | Task | File(s) | Detail |
|---|------|---------|--------|
| 13 | `ControlMenu` struct + rendering | `internal/engine/control_menu.go` (new) | Button array; each button = `label, actionKey, enabled bool`; renders as white-on-black rectangles with 1-cell gap; grid-aligns at bottom-right; scrollable if many buttons. |
| 14 | Control menu lifecycle | `internal/engine/control_menu.go` | Toggle visible/hidden via hamburger `☰` at `(w-4, 0)`; in `TouchMode` the hamburger is always visible. |
| 15 | Hook into game loop | `internal/engine/game.go` | Render control menu after screen `Render()` but before `Flush()`; touch events consumed by menu first, fall through to screen if missed. |
| 16 | **Battlescape** buttons | `internal/battle/battlescape.go` | `[Select] [Move] [Fire] [Reload] [End Turn] [Grenade] [Medikit] [Crouch] [Cycle] [Cancel]` |
| 17 | **Geoscape** buttons | `internal/geo/geoscape.go` | `[Pause] [1×] [2×] [3×] [4×] [Base] [Launch] [Save] [Load]` |
| 18 | **Base/Equip/Research/Manufacture** buttons | `internal/base/*.go` | `[Facilities] [Soldiers] [Research] [Manufacture] [Transfer] [Hangars] [Back]` |
| 19 | **Menu/Options/Difficulty/etc** buttons | `internal/engine/*.go` | Help bar hotkeys, option toggles, selection confirm. |
| 20 | Auto-show on first touch | `internal/engine/control_menu.go` | In `TouchMode`, first touch on a screen reveals the control menu automatically (dismiss with `☰` or tap outside). |

## Phase 4 — Polish & mobile ergonomics
| # | Task | Detail |
|---|------|--------|
| 21 | Default mobile terminal size | Init web terminal at portrait-friendly 50×120 on narrow connect. |
| 22 | Tap debounce | 250ms window to suppress double-tap; long-press (≥500ms) → right-click. |
| 23 | Touch scrolling | List areas scroll on vertical drag in `mobile` mode. |
| 24 | CSS viewport tuning | Remove `user-scalable=no`; ensure xterm.js scales to mobile viewport. |
| 25 | Document mobile key bindings | Add to `AGENTS.md` and `docs/manual.md`.
