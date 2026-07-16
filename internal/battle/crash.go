package battle

import (
	"image"
	"math/rand"

	"github.com/taislin/termcom/internal/data"
)

// CrashResult captures the outcome of a vehicle crash landing.
type CrashResult struct {
	PartsSurvived int
	PartsDestroyed int
	LootItems      []string // loot IDs from surviving lootable parts
	Explosions     []image.Point // positions of parts that exploded on destruction
	InteriorTiles  []image.Point // passable tiles inside the UFO (for alien crew placement)
	ExteriorTiles  []image.Point // passable outdoor tiles adjacent to the UFO (perimeter guards)
}

// StampVehicleOnMap places a vehicle blueprint onto the tactical map.
// crashSeverity ranges from 0.0 (perfect landing) to 1.0 (total destruction).
// Returns a CrashResult describing what survived.
func StampVehicleOnMap(
	bp *data.VehicleBlueprint,
	startX, startY int,
	battleMap *BattleMap,
	crashSeverity float64,
) CrashResult {
	result := CrashResult{}

	if crashSeverity < 0 {
		crashSeverity = 0
	}
	if crashSeverity > 1 {
		crashSeverity = 1
	}

	// First pass: stamp vehicle parts
	for pt, part := range bp.Parts {
		x := startX + pt.X
		y := startY + pt.Y

		if x < 0 || x >= battleMap.Width || y < 0 || y >= battleMap.Height {
			continue
		}

		hpRatio := float64(part.CurrentHP) / float64(part.Def.BattlescapeHP)
		destroyChance := crashSeverity * (1.0 + (1.0-hpRatio)*0.3)
		if rand.Float64() < destroyChance {
			result.PartsDestroyed++
			if part.Def.ExplodesOnDeath {
				result.Explosions = append(result.Explosions, image.Point{X: x, Y: y})
			}
			battleMap.Tiles[y][x] = Tile{
				Type:      TileRubble,
				Cover:     TileCover(TileRubble),
				Level:     0,
				Rune:      '.',
				BaseColor: 0,
			}
			if rand.Intn(3) == 0 {
				battleMap.SpawnBlood(x, y, 1)
			}
		} else {
			result.PartsSurvived++
			tileType := TileObject
			cover := TileCover(TileObject)

			switch part.Def.Category {
			case data.PartHull:
				tileType = TileUFOWall
				cover = TileCover(TileUFOWall)
			case data.PartPowerCore:
				tileType = TilePowerSource
				cover = TileCover(TilePowerSource)
			case data.PartWeapon:
				tileType = TileMachinery
				cover = TileCover(TileMachinery)
			case data.PartCockpit:
				tileType = TileConsole
				cover = TileCover(TileConsole)
			case data.PartEngine:
				tileType = TileMachinery
				cover = TileCover(TileMachinery)
			}

			battleMap.Tiles[y][x] = Tile{
				Type:      tileType,
				Cover:     cover,
				Level:     0,
				Rune:      part.Def.TacticalRune,
				BaseColor: part.Def.Color,
			}

			if lootID := part.LootID(); lootID != "" {
				result.LootItems = append(result.LootItems, lootID)
			}
		}
	}

	// Second pass: fill interior floor tiles (positions inside the bounding box
	// that don't have a part and aren't wall tiles)
	for dy := 0; dy < bp.Height; dy++ {
		for dx := 0; dx < bp.Width; dx++ {
			x := startX + dx
			y := startY + dy
			if x < 0 || x >= battleMap.Width || y < 0 || y >= battleMap.Height {
				continue
			}
			tile := &battleMap.Tiles[y][x]
			// Only fill empty grass/grass tiles inside the bounding box
			if tile.Type == TileGrass || tile.Type == TileTree || tile.Type == TileBush || tile.Type == TileFence || tile.Type == TileRock {
				tile.Type = TileUFOFloor
				tile.Cover = TileCover(TileUFOFloor)
				tile.Rune = '.'
				tile.BaseColor = 0
			}
			// Collect passable interior tiles for crew placement
			if battleMap.Passable(x, y) {
				result.InteriorTiles = append(result.InteriorTiles, image.Point{X: x, Y: y})
			}
		}
	}

	// Collect passable outdoor tiles adjacent to the UFO so a portion of the
	// crew can deploy as perimeter guards instead of all spawning inside.
	interiorSet := make(map[image.Point]bool, len(result.InteriorTiles))
	for _, pt := range result.InteriorTiles {
		interiorSet[pt] = true
	}
	adj := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	for pt := range interiorSet {
		for _, d := range adj {
			ex, ey := pt.X+d[0], pt.Y+d[1]
			if ex < 0 || ex >= battleMap.Width || ey < 0 || ey >= battleMap.Height {
				continue
			}
			ep := image.Point{X: ex, Y: ey}
			if interiorSet[ep] {
				continue
			}
			if battleMap.Passable(ex, ey) {
				result.ExteriorTiles = append(result.ExteriorTiles, ep)
			}
		}
	}

	return result
}

// ExplodeTile triggers an explosion at (x,y) on the battle map.
// It damages all VehiclePart tiles in the blast radius and converts
// destroyed parts to rubble. Used when a soldier shoots an explosive part.
func ExplodeTile(x, y int, battleMap *BattleMap, radius int) {
	type blastTarget struct {
		x, y int
		dist int
	}

	var targets []blastTarget
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			nx, ny := x+dx, y+dy
			if nx < 0 || nx >= battleMap.Width || ny < 0 || ny >= battleMap.Height {
				continue
			}
			dist := abs(dx) + abs(dy)
			if dist > radius {
				continue
			}
			tile := &battleMap.Tiles[ny][nx]
			if tile.Type == TileObject || tile.Type == TileUFOWall ||
				tile.Type == TilePowerSource || tile.Type == TileMachinery ||
				tile.Type == TileConsole || tile.Type == TileStorage {
				targets = append(targets, blastTarget{nx, ny, dist})
			}
		}
	}

	for _, t := range targets {
		tile := &battleMap.Tiles[t.y][t.x]
		damage := radius - t.dist + 1
		if tile.Type == TilePowerSource {
			// Power cores chain-explode
			damage = radius * 2
		}
		// 50% chance per point of damage to destroy
		if rand.Intn(100) < damage*50 {
			tile.Type = TileRubble
			tile.Cover = TileCover(TileRubble)
			tile.Rune = '.'
			tile.BaseColor = 0
		}
	}
}


