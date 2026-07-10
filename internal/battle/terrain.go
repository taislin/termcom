package battle

type TerrainDef struct {
	Name           string
	TileProbs      map[TileType]int // Probability out of 100
	DefaultTile    TileType
	StructureTypes []string
	ObjectDensity  float64
}

var Biomes = map[string]TerrainDef{
	"crash": {
		Name: "Crash Site",
		TileProbs: map[TileType]int{
			TileTree: 3,
			TileBush: 2,
			TileRock: 2,
			TileFence: 1,
		},
		DefaultTile:    TileGrass,
		ObjectDensity:  0.15,
	},
	"forest": {
		Name: "Forest",
		TileProbs: map[TileType]int{
			TileTree: 15,
			TileBush: 5,
			TileRock: 2,
		},
		DefaultTile:    TileGrass,
		ObjectDensity:  0.25,
	},
	"desert": {
		Name: "Desert",
		TileProbs: map[TileType]int{
			TileRock: 5,
			TileSand: 3,
			TileBush: 2,
		},
		DefaultTile:    TileSand, // Use as base for sand/dust
		ObjectDensity:  0.1,
	},
}
