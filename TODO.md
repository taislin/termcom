# YCOM Development Roadmap

## Phase 1-9: Core Systems (DONE)
- [x] Technical debt, dogfights, alien tactics, hangar management
- [x] Mission variety, polish & QoL
- [x] Research & manufacturing systems
- [x] Campaign & endgame (Alien Activity, monthly funding, Cydonia, game over)
- [x] Modern VFX & options menu (bloom, distortion, directional lighting)

## Phase 10: Procedural Tech Tree (DONE)
- [x] Tiered DAG tech tree generator with cost variance
- [x] Dynamic autopsy injection from procedural alien species
- [x] Save/load species seed for deterministic tree regeneration
- [x] Code review: bug fixes (nil checks, serialization, DAG validation)
- [x] Dead code removal and test coverage improvements

## Phase 11: Save System Enhancements
- [x] Distinguish Continue vs Load Game in menu (Continue = load last save silently)
- [x] Multiple save slots with file picker UI
- [x] Auto-save on Geoscape monthly report

## Phase 12: Battle Loot & Economy Balance
- [x] Balance generateUFOLoot() per species and UFO type
- [x] Tune manufacturing costs and material requirements for varied tech tree
- [x] Adjust monthly funding to account for randomized research costs

## Phase 13: Research Screen Improvements (DONE)
- [x] Show checkmarks on researched topics
- [x] Grey out completed topics in the list
- [x] Display tier info so players can plan their research path
- [x] Show prerequisite tree graph in research UI

## Phase 14: Alien Progression Scaling (DONE)
- [x] Scale alien stats with game time (harder aliens as months pass)
- [x] Increase UFO frequency and strength over the campaign
- [x] Gate elite alien species behind certain mission thresholds

## Phase 15: Expanded Battle VFX
- [x] Muzzle flash particles on weapon fire
- [x] Explosion debris particles for grenades and heavy weapons
- [x] Ambient particles per mission type (rain, snow, dust, embers)
- [x] Blood splatter on hit

## Phase 16: Sound Effects (if feasible with our tone system - some are done already so check first)
- [x] Weapon fire sounds (laser, plasma, ballistic, melee)
- [x] UI navigation sounds (menu select, research complete, manufacturing done)
- [x] Ambient battle sounds (wind, distant explosions)
- [x] Geoscape alert sounds (UFO detected, mission warning)

## Phase 17: Multi-Base Support
- [x] Build and manage multiple bases on the Geoscape
- [x] Transfer soldiers and items between bases
- [x] Regional radar coverage per base
- [x] Base defense missions when aliens attack a base

## Phase 18: Battlescape AI Polish
- [x] Enhance alien AI in internal/battle/ai.go for more challenging tactical combat
- [x] Smarter flanking and cover usage
- [x] Coordinated squad behavior (suppression, focused fire, retreat)
- [x] AI adapts to player tactics across missions

## Phase 19: Geoscape Mission Variety
- [x] Expand mission types beyond crash sites and terror missions
- [x] Alien base assault missions with unique maps
- [x] Supply raid missions (intercept alien transports)
- [x] Council missions with special objectives and bonus rewards

## Phase 20: Campaign Completion & Save Integrity (BLOCKERS + MAJORS)
- [x] A1 Victory flow: set gs.Victory on winning the Cydonia final mission; guard
      triggerCydonia() so it fires only once (no infinite re-trigger)
- [x] A2 Verify defeat paths (AlienActivity>=100, last base destroyed) reach GameOver
- [x] A3 Fix interception node/crash-site bug: Interceptor.Update must use the real
      UFO list (gs.UFOs), not a throwaway &UFOList{} (interceptor.go:161)
- [x] A4 Save/load interceptor roster: round-trip Hangars in BaseSave/FromBase/ToBase
- [x] A5 Enforce storage capacity (MaxStorage/UsedStorage) on AddLoot/Equip/
      Manufacture/Transfer so bases cannot hoard unlimited loot

## Phase 21: Alien Capture (capture-only scope) (DONE)
- [x] B1 Stun mechanic: Stun Rod stuns instead of kills at low HP (Stunned flag)
- [x] B2 Live-alien storage: stunned aliens added to base if Alien Containment
      exists (capacity-gated); otherwise lost
- [x] B3 Interrogation -> research bonus: consume a captured alien to auto-complete
      an autopsy or grant large progress to an active topic
- [x] B4 Leave Psi-Lab cosmetic (no psi training); document in manual

## Phase 22: Economy / Balance / Polish (MINOR) (DONE)
- [x] C1 Healing pacing: advance healing daily (or boost +2 HP/day) so wounds recover
      in reasonable time; optionally gate wounded (HP<MaxHP) from deployment
- [x] C2 Difficulty selection at new game (Beginner..Superhuman) affecting UFO spawn
      rate and alien stat scaling
- [x] C3 Nil-safety hardening: guard SelectedBase() usages in geoscape Update

## Phase 23: Multi-Platform Audio Engine
- [ ] Implement procedural sound synthesis in `audio_other.go` (Linux/macOS) using
      `beep` or `oto` to replace terminal BEL beeps
- [ ] Implement weapon-specific fire sounds, explosions, and ambient battle winds
- [ ] Ensure parity with `audio_windows.go` synthesis logic

## Phase 24: Radar Visualization
- [ ] Implement toggle for radar coverage overlay on minimap (e.g., key `V`)
- [ ] Draw regional radar ranges to illustrate coverage expansion from bases

## Phase 25: Docs & Tests
- [ ] D1 Update manual.md: capture/containment real, Psi-Lab cosmetic, corrected
      healing rate and final-mission steps; fix other mismatched claims
- [ ] E1 Tests: victory reachable & Cydonia fires once; interceptor save round-trip;
      interception node path engages UFO; storage cap blocks overflow; capture ->
      containment -> interrogation flow
