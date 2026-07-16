# TODO — Battlescape Improvements

Scope: Features and fixes for the battlescape tactical combat system.
Nothing here is implemented yet unless marked `[x]`.

---

## Burst / Auto Fire

Add fire modes to ranged weapons: Aimed (current), Burst (3-round, -10% acc), Auto
(full clip, -20% acc, costs more TU). Faithful to original X-COM.

- [ ] `data/items.go`: add `FireModes []FireMode` to `RuleItem` (Aimed/Burst/Auto)
  with per-mode TU/accuracy modifiers. Default `[Aimed]` for weapons without burst.
- [ ] `battle/unit.go` `FireAt`: accept `fireMode FireMode` param; apply mode
  accuracy modifier and TU cost before roll.
- [ ] `battle/battlescape.go` `FireWeapon()`: cycle modes on `[F]` press; show
  current mode in sidebar.
- [ ] `battle/battlescape.go` `AlienAction`: add `FireMode` to struct; AI picks
  mode based on distance and target count (auto vs clustered, burst vs single).
- [ ] `battle/input.go`: `F` key cycles fire mode when cursor is in fire range.
- [ ] Language keys: `MSG_FIRE_MODE_AIMED`, `MSG_FIRE_MODE_BURST`, `MSG_FIRE_MODE_AUTO`
  in all 8 locales.
- [ ] Update `docs/manual.md`: Key bindings, Weapon stats table.

## Inventory Management

Soldiers carry items in slots (Primary, Secondary, Belt, Pockets). Weight affects
TU. Grenades, medikits, ammo clips as separate equippable items.

- [ ] `data/items.go`: add `Slot EquipSlot` field (EquipPrimary/Secondary/Belt/Pocket);
  `Weight` already exists.
- [ ] `soldier/soldier.go`: add `Inventory []Item` to `Soldier`; `Encumbrance() int`
  returns total weight; TU penalty = `floor(Weight / 5)`.
- [ ] `base/equip.go`: rework equip screen to show slot-based inventory with
  weight/TU penalty display.
- [ ] `battle/unit.go`: apply encumbrance TU penalty at battle start.
- [ ] Battle UI: show grenade/medikit count in sidebar; `[G]` opens grenade selector
  when multiple grenades carried.
- [ ] `save/save.go`: persist inventory list (add `Inventory` to `SoldierData`).
- [ ] Language keys for inventory UI in all 8 locales.
- [ ] Update `docs/manual.md`: Equipment section.

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
- [ ] **B1** Movement-range BFS charges 8 TU but `MoveTo` charges flat 4/tile.
  Align cost models.
- [ ] **B4** `PsiAttack`/`Reload`/`UseMedikit` deref `bs.Selected.Soldier.Weapon`
  without nil guard. Add guard.
- [ ] **M1** `GetMovementRange` cache key collides at same X,Y,TU. Include unit ID.
- [ ] **D2** `PlayerFlankCount` declared but never used. Remove.
- [ ] **R1** Shot-FX duplicated across player/alien reaction fire. Extract helper.

### `internal/geo/geoscape.go`
- [ ] **M1** Defeat sets `gs.Victory=true` → shows victory after game-over.
  Add `Defeated` flag.
- [ ] **M2** Losing last base sets `gs.Victory=true` → same win-after-loss bug.
- [ ] **B4** Save/load drops UFO ID/X/Y/TurnsLeft → defender respawns every tick.
  Persist these fields.
- [ ] **B1** `defBase` shadowed by `SelectedBase()` → wrong base for mission response.
  Remove shadow.
