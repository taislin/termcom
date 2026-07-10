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
- [ ] AI adapts to player tactics across missions

## Phase 19: Geoscape Mission Variety
- [ ] Expand mission types beyond crash sites and terror missions
- [ ] Alien base assault missions with unique maps
- [ ] Supply raid missions (intercept alien transports)
- [ ] Council missions with special objectives and bonus rewards
