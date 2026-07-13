# termcom Development Roadmap

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

## Phase 23: Multi-Platform Audio Engine (DONE)
- [x] Implement procedural sound synthesis in `audio_other.go` (Linux/macOS) using
      `oto` to replace terminal BEL beeps
- [x] Implement weapon-specific fire sounds, explosions, and ambient battle winds
- [x] Ensure parity with `audio_windows.go` synthesis logic

## Phase 24: Radar Visualization (DONE)
- [x] Implement toggle for radar coverage overlay on minimap (key `V`)
- [x] Draw regional radar ranges to illustrate coverage expansion from bases

## Phase 25: Docs & Tests (DONE)
- [x] D1 Update manual.md: victory condition (Cydonia), month length, storage weight,
      battlescape key bindings (q/P/H), stray Chinese characters removed
- [x] E1 Tests: storage cap blocks overflow, capture→containment→interrogation flow,
      interceptor save round-trip, psi-Lab training cap

## Phase 26: Psi Abilities (DONE)
- [x] Wire 'P' key in battlescape input to call PsiAttack()
- [x] Add PsiSkill/PsiStr/Panicked fields to battle Unit struct
- [x] Copy psi stats from Soldier/AlienType to battle units
- [x] Psi-Lab training: daily +1 PsiSkill (max 80) for soldiers in base with Psi-Lab
- [x] Alien psi attacks: high-Psi aliens (Psi>40) use psi with 1/3 chance per turn
- [x] Panic effect: successful psi attack zeros TU and panics target (skips next turn)
- [x] Mind Control research: +20 PsiSkill bonus to all soldiers at that base
- [x] Psi formula: success = attackerSkill - defenderPsiStr/3, min 5% chance
- [x] Verify the codebase for errors/mistakes

## Phase 27: Sub-Cell "Pixel" Portrait Engine (URR Half-Block Renderer) (DONE)
- [x] P1 Create `internal/engine/pixel.go`:
  - [x] P1a Define `PixelImage` struct (`Width, Height int`, `Pixels [][]tcell.Color`)
  - [x] P1b Implement `NewPixelImage(w, h int) *PixelImage` (allocates, fills black)
  - [x] P1c Implement `DrawPixelImage(screen *ScreenRaw, x, y int, img *PixelImage)`:
        iterates rows 2-by-2, maps Pixels[row][col] → FG and Pixels[row+1][col] → BG,
        draws '▀' (U+2580); odd-height images pad last BG to tcell.ColorBlack
  - [x] P1d Implement `CompositeImages(base, overlay *PixelImage) *PixelImage`:
        skips overlay pixels equal to tcell.ColorDefault (transparent), returns new image
  - [x] P1e Implement `DarkenColor(c tcell.Color, factor float64) tcell.Color`:
        extracts RGB, multiplies each channel by factor, clamps to [0,255],
        returns via tcell.NewRGBColor; passes through tcell.ColorDefault unchanged
  - [x] P1f Implement `LightenColor(c tcell.Color, factor float64) tcell.Color`:
        blends toward white (255,255,255) by factor beyond 1.0
- [x] P2 Create `internal/engine/portrait.go`:
  - [x] P2a Define `PortraitLayer` int enum (LayerSkin, LayerEyes, LayerHair, LayerHelmet,
        LayerArmour, LayerDecal, LayerCount)
  - [x] P2b Define `PortraitSpec` struct (Width, Height, SkinColor, EyeColor, HairColor,
        HelmetColor, ArmourColor, DecalColor, Seed int64)
  - [x] P2c Implement `generateSkinLayer(spec PortraitSpec) *PixelImage`:
        oval head + rectangular torso fill; edge pixels DarkenColor by 0.7 for rim shading
  - [x] P2d Implement `generateEyeLayer(spec PortraitSpec) *PixelImage`:
        two small rectangles at fixed head positions; pupils as DarkenColor by 0.4 dots
  - [x] P2e Implement `generateHairLayer(spec PortraitSpec) *PixelImage`:
        top-of-head band; style variant chosen from spec.Seed % 4
  - [x] P2f Implement `generateHelmetLayer(spec PortraitSpec) *PixelImage`:
        trapezoidal upper-head region; visor strip DarkenColor by 0.5; returns nil if
        spec.HelmetColor == tcell.ColorDefault
  - [x] P2g Implement `generateArmourLayer(spec PortraitSpec) *PixelImage`:
        torso region overwrite; highlight edges LightenColor by 1.3; returns nil if
        spec.ArmourColor == tcell.ColorDefault
  - [x] P2h Implement `generateDecalLayer(spec PortraitSpec) *PixelImage`:
        1-5 rank pips at fixed torso position, colored spec.DecalColor
  - [x] P2i Implement `GenerateSoldierPortrait(spec PortraitSpec) *PixelImage`:
        calls each generator, composites Skin → Eyes → Hair → Helmet → Armour → Decal
  - [x] P2j Implement `GenerateAlienPortrait(sp data.StyledPortrait, scale int) *PixelImage`:
        upscales each ASCII rune to scale×scale pixels using a rune-density lookup table
        (space=transparent, '.'=25% fill, '@'=100% fill, etc.)
- [x] P3 Create `internal/engine/pixel_test.go`:
  - [x] P3a TestDrawPixelImageOddHeight: odd height must not panic; last cell BG = black
  - [x] P3b TestCompositeImages: ColorDefault in overlay leaves base pixel unchanged
  - [x] P3c TestDarkenColor: factor=0.0→black, factor=1.0→identity, factor=0.5→half channels
  - [x] P3d TestDarkenColor_Transparent: ColorDefault passes through unchanged
- [x] P4 Integrate portrait rendering into existing screens:
  - [x] P4a `internal/base/equip.go`: replace `StyledPortrait` text drawing with
        `GenerateSoldierPortrait` + `DrawPixelImage` in the right-panel portrait region
  - [x] P4b `internal/engine/encyclopedia.go`: replace `StyledPortrait` text drawing with
        `GenerateAlienPortrait(at.GetPortrait(), 2)` + `DrawPixelImage` in portrait region
- [x] P5 Update `docs/manual.md` to note half-block portrait rendering


## Phase 28: Geometric Terrain Engine (URR ASCII Geometry) (DONE)
- [x] G1 Extend `Tile` struct in `internal/battle/map.go`:
  - [x] G1a Add `Elevation int` field (skipped per user request)
  - [x] G1b Add `BaseColor tcell.Color` field (tcell.ColorDefault = use TilePalette lookup)
  - [x] G1c Add `Rune rune` field (0 = use TileGeomRune contextual logic)
  - [x] G1d Verify `NewBattleMap` and `NewMultiLevelBattleMap` zero-initialize new fields (skipped elevation verification)
- [x] G2 Create `internal/battle/terrain.go`:
  - [x] G2a Define UFO geometry rune constants (◤ ◥ ◣ ◢ ▬ ▐ ⊠) with comments
  - [x] G2b Define human building box-drawing rune constants (╔ ═ ╗ ║ ╚ ╝ ┼) with comments
  - [x] G2c Define `TilePalette map[TileType]tcell.Color` for all tile types with
        curated RGB values (dark earth tones for terrain, blue-grey for UFO, etc.)
  - [x] G2d Implement `ElevationDarken(elevation int) float64` (skipped per user request)
  - [x] G2e Implement `TileBaseColor(t Tile) tcell.Color`:
        returns t.BaseColor if not ColorDefault, else TilePalette[t.Type], else neutral grey
  - [x] G2f Implement `(m *BattleMap) neighbourhood(x, y int) [3][3]TileType`:
        reads 3×3 grid centred on (x,y), clamping OOB to TileGrass; added to map.go
  - [x] G2g Implement `TileGeomRune(t Tile, ctx [3][3]TileType) rune`:
        - Tile.Rune override (non-zero) returned immediately
        - TileUFOWall: check N/S/E/W neighbours for non-UFO → select ◤/◥/◣/◢ or █
        - TileWall: check N/S/E/W neighbours → select ╔/╗/╚/╝/═/║/# accordingly
        - Fallback: tileChars[t.Type]
  - [x] G2h Implement private `bloodColor(bloodType int) tcell.Color` mapping 1→red,
        2→green, 3→purple
  - [x] G2i Implement private `fireColor(frame int) tcell.Color` returning an animated
        orange-yellow flickered via frame%3 step
  - [x] G2j Implement `RenderTile(t Tile, ctx [3][3]TileType, visible, seen bool) (rune, tcell.Style)`:
        full pipeline: TileBaseColor → DarkenColor FG/BG if unseen → blood/fire overlay → TileGeomRune → return style
- [x] G3 Create `internal/battle/terrain_test.go`:
  - [x] G3a TestTileGeomRune_UFOCornerNW: north+west neighbour non-UFO → ◤
  - [x] G3b TestTileGeomRune_UFOSolid: all UFO neighbours → █
  - [x] G3c TestTileGeomRune_BuildingCornerTL: wall with south+east neighbours → ╔
  - [x] G3d TestRenderTile_Unseen: !visible && !seen → blank rune returned
  - [x] G3e TestRenderTile_ElevationDarkens (skipped per user request)
  - [x] G3f TestElevationDarken_Clamped (skipped per user request)
- [x] G4 Integrate `RenderTile` into battlescape draw loop in `internal/battle/battlescape.go`:
  - [x] G4a Locate existing per-tile rune+style inline block in the render loop
  - [x] G4b Replace with `ctx := bs.Map.neighbourhood(mapX, mapY)` + `RenderTile(tile, ctx, visible, seen)`
  - [x] G4c Verify blood/fire rendering parity with previous inline code
- [x] G5 Optionally populate Elevation on existing map generators (skipped per user request)
- [x] G6 Update `docs/manual.md` to note geometric terrain rendering

## Phase 29: Interceptor Combat Visuals (DONE)
- [x] Add minimap combat animation during dogfights (interceptor/UFO icon flashes, explosion effects, pulsing engagement icons)
- [x] Show damage numbers and HP bars during air combat (minimap overlay panel with green/yellow/red bars)
- [x] Add visual distinction between interceptors traveling vs engaging (patrol `>` vs intercept `►`)
- [x] Update `docs/manual.md` with interceptor combat visual details

## Phase 30: Alien Equipment Escalation (DONE)
- [x] Define alien tech tiers (early: plasma pistol/rifle, mid: heavy plasma, late: alien cannon/laser)
- [x] Scale alien weapon/armor loadouts with game month (tiers at month 0/3/6/9)
- [x] Ensure loot tables reflect escalated equipment (tier-based stat bonuses applied alongside existing scaling)
- [x] Update `docs/manual.md` with alien equipment escalation details

## Phase 31: Base Facility Adjacency Bonuses (DONE)
- [x] Design adjacency bonus system (Lab+Lab → +10% research, Workshop+Workshop → +10% manufacture, Living Quarters+Living Quarters → +1 HP/day)
- [x] Implement adjacency check on base grid layout (orthogonal adjacency in Row/Col grid)
- [x] Display adjacency bonuses in base management UI (counts shown below facility list)
- [x] Update `docs/manual.md` with adjacency bonus details

## Phase 32: Battlescape Visual Polish (Visual Improvements — Tier 1)

NOTE — verified already implemented (do NOT duplicate):
- Hit feedback: muzzle flash (`engine.SpawnMuzzleFlash` at battlescape.go:736/:1594)
  and blood spray (`bs.SpawnBloodSplatter` at :742/:765/:875/:939/:1601/:1936)
  are already wired. See Phase 15.
- Battlescape HUD portrait: selected soldier's half-block portrait is already drawn
  via `MakeSoldierPortrait` + `ctx.DrawPixelImageFramed` at battlescape.go:2455-2458.

Planned work (not yet implemented):
- [x] Enable the Lighting option (fix dead code): uncomment & gate the directional
      flashlight cone (`engine.ApplyDirectionalLight`) for the selected unit in
      `internal/battle/battlescape.go` (~line 2267); correct the `isVisible` closure
      to use `bs.Map.Opaque(x+bs.ScrollX, y+bs.ScrollY-1)`. (Phase 8 marked directional
      lighting done, but the call is currently commented out — this is a fix.)
- [x] Floating combat text in battlescape: add a `FloatingText` slice on Battlescape;
      spawn rising/fading damage numbers, "MISS", and heal values above hit targets;
      update/draw in `Update()`/`Render()`; spawn at the damage sites
      (battlescape.go:761 and ~:1610). (Interceptor combat already has damage numbers
      per Phase 29; this adds battlescape-level floating text.)
- [x] Unit health/TU pips + selection shadow: for the selected/hovered unit draw a dim
      selection shadow under the sprite and a 3-cell HP pip bar (green→yellow→red by
      HP ratio) on the tile above; keep TU in the sidebar.
- [x] Scene-transition fade: add a `transition` alpha field to `Game`; set to 1.0 on
      `PushState`/`PopState`/`SetState` and ease to 0 each frame; draw a full-screen
      black overlay via `engine.DrawTransparentRect` in `Run()` so state changes fade
      from black instead of cutting abruptly. Skip the overlay while `quitConfirm`.

## Phase 33: Additional Visual Polish (Tier 2/3 remainder)
- [ ] Geoscape day/night terminator: draw a sweeping day/night boundary line across
      the globe minimap that advances with GameTime; tint the night hemisphere darker.
- [ ] Geoscape UFO/interceptor markers: pulsing radar-blip animation for active UFOs
      and persistent trail lines tracing interceptor flight paths on the minimap.
- [ ] Extra color themes: add amber and green CRT-phosphor palettes plus a "paper"
      palette, selectable in the Options menu (extend `ApplyTheme` / theme state).
- [ ] Tile edge shading: apply light ambient-occlusion where wall meets floor and add
      subtle per-tile dither for depth in `RenderTile` (internal/battle/terrain.go).
- [ ] Battlescape HUD bars: replace plain HP/TU text in the sidebar with clearer
      graphical HP and TU bars (color-coded, proportionate to current/max).


