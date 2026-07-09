# YCOM Development Roadmap

## Completed
- [x] Fix tcell deprecation warnings (`staticcheck`) — replaced `tcell.Color*` with `color.XTerm*`
- [x] Multi-level maps (UFO interiors: 2 levels with stairs connections)
- [x] Procedural alien portraits (1:2 ratio, full body, per damage type/rank)
- [x] Night/day missions (lighting, accuracy penalty, sight reduction)
- [x] Psi combat (psi_amp item, P key, TU cost, skill vs defense)
- [x] Encyclopedia with alien portraits and stats

## Phase 1: Technical Debt & Polish
- [ ] Refactor geoscape region table column offsets to be dynamic
- [ ] Profile render loop performance

## Phase 2: Enhanced Dogfights
- [ ] Add interceptor weapon variety (Cannon, Stingray, Avalanche)
- [ ] Implement combat maneuvers (Attack, Cautious, Break-off)

## Phase 3: Alien Tactics
- [ ] Implement reaction fire (human and alien)
- [ ] Enhance Alien AI (flanking, reinforcements, retreat behaviors)
