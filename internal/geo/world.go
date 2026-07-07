package geo

// Equirectangular world map. Each char = 1 cell. Land is 1, water is 0.
// Map is 180 wide x 90 tall (2-degree resolution).
// x = longitude/2 (0..360 -> 0..179)
// y = (90 - latitude)/2 (90N..90S -> 0..89)
const mapW = 180
const mapH = 90

var worldMap [mapH][mapW]int

type City struct {
	Name string
	X    int
	Y    int
}

var cities []City

// lonLat converts standard degrees (positive=north, positive=east) to map coords.
func lonLat(lon, lat float64) (int, int) {
	x := int(lon / 2.0)
	y := int((90.0 - lat) / 2.0)
	if x < 0 {
		x += mapW
	}
	if x >= mapW {
		x -= mapW
	}
	if y < 0 {
		y = 0
	}
	if y >= mapH {
		y = mapH - 1
	}
	return x, y
}

// polygonFill fills a polygon defined by lon/lat vertices onto the map.
// Uses scanline fill with edge intersection sorting.
func polygonFill(vertices [][2]float64) {
	// Convert vertices to map coordinates
	type point struct{ x, y float64 }
	pts := make([]point, len(vertices))
	for i, v := range vertices {
		x, y := lonLat(v[0], v[1])
		pts[i] = point{float64(x), float64(y)}
	}

	n := len(pts)
	if n < 3 {
		return
	}

	// Find Y bounds
	minY := int(pts[0].y)
	maxY := int(pts[0].y)
	for _, p := range pts {
		if int(p.y) < minY {
			minY = int(p.y)
		}
		if int(p.y) > maxY {
			maxY = int(p.y)
		}
	}
	if minY < 0 {
		minY = 0
	}
	if maxY >= mapH {
		maxY = mapH - 1
	}

	// Scanline fill
	for row := minY; row <= maxY; row++ {
		// Find intersections of polygon edges with this row
		var intersections []float64
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			yi := pts[i].y
			yj := pts[j].y
			xi := pts[i].x
			xj := pts[j].x

			// Check if edge crosses this row
			if (yi <= float64(row) && yj > float64(row)) || (yj <= float64(row) && yi > float64(row)) {
				// Calculate intersection x
				t := (float64(row) - yi) / (yj - yi)
				ix := xi + t*(xj-xi)
				intersections = append(intersections, ix)
			}
		}

		// Sort intersections
		for i := 0; i < len(intersections); i++ {
			for j := i + 1; j < len(intersections); j++ {
				if intersections[i] > intersections[j] {
					intersections[i], intersections[j] = intersections[j], intersections[i]
				}
			}
		}

		// Fill between pairs
		for i := 0; i+1 < len(intersections); i += 2 {
			x1 := int(intersections[i])
			x2 := int(intersections[i+1])
			if x1 < 0 {
				x1 = 0
			}
			if x2 >= mapW {
				x2 = mapW - 1
			}
			for x := x1; x <= x2; x++ {
				worldMap[row][x] = 1
			}
		}
	}
}

// fillCircle fills a rough circle of land (for island-like shapes)
func fillCircle(cx, cy, r int) {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				x := cx + dx
				y := cy + dy
				if x >= 0 && x < mapW && y >= 0 && y < mapH {
					worldMap[y][x] = 1
				}
			}
		}
	}
}

func init() {
	for y := 0; y < mapH; y++ {
		for x := 0; x < mapW; x++ {
			worldMap[y][x] = 0
		}
	}

	// NORTH AMERICA
	// Main continent (lon, lat vertices - positive=east, positive=north)
	polygonFill([][2]float64{
		{-170, 65}, // Alaska NW
		{-140, 70}, // Alaska N
		{-65, 72},  // Northern Canada
		{-55, 50},  // Newfoundland
		{-65, 44},  // Nova Scotia
		{-75, 35},  // Carolinas
		{-80, 25},  // Florida
		{-90, 28},  // Gulf coast
		{-97, 26},  // Texas
		{-105, 20}, // Mexico W
		{-85, 10},  // Central America
		{-78, 8},   // Panama
		{-82, 15},  // Honduras
		{-105, 22}, // Mexico mid
		{-117, 32}, // Baja
		{-122, 37}, // San Francisco
		{-125, 48}, // Vancouver
		{-140, 60}, // Alaska S
		{-170, 65}, // close
	})

	// Greenland
	polygonFill([][2]float64{
		{-55, 60},
		{-20, 65},
		{-18, 76},
		{-40, 83},
		{-60, 82},
		{-70, 76},
		{-55, 60},
	})

	// SOUTH AMERICA
	polygonFill([][2]float64{
		{-80, 12},  // Venezuela/Colombia N
		{-60, 10},  // Guyana
		{-35, -5},  // Brazil E
		{-38, -15}, // Brazil SE
		{-48, -28}, // S Brazil
		{-55, -34}, // Buenos Aires
		{-65, -42}, // Patagonia
		{-70, -52}, // Tierra del Fuego
		{-75, -45}, // Chile S
		{-72, -30}, // Chile mid
		{-80, -5},  // Ecuador
		{-78, 2},   // Colombia
		{-80, 12},  // close
	})

	// EUROPE
	polygonFill([][2]float64{
		{-10, 36},  // Portugal
		{0, 43},    // Spain N
		{-5, 48},   // Brittany
		{2, 51},    // Belgium
		{5, 54},    // Denmark
		{12, 55},   // Germany N
		{20, 55},   // Poland
		{30, 60},   // Baltics
		{40, 68},   // Finland
		{30, 71},   // Norway N
		{15, 65},   // Norway W
		{5, 62},    // Norway SW
		{-5, 58},   // Scotland
		{-10, 52},  // Ireland
		{-6, 50},   // Cornwall
		{-2, 43},   // Bay of Biscay
		{-10, 36},  // close
	})

	// UK/Ireland
	polygonFill([][2]float64{
		{-10, 50},
		{-6, 50},
		{1, 51},
		{2, 53},
		{0, 58},
		{-2, 58},
		{-5, 56},
		{-6, 54},
		{-10, 52},
		{-10, 50},
	})

	// Scandinavia
	polygonFill([][2]float64{
		{5, 58},
		{12, 56},
		{18, 56},
		{24, 60},
		{30, 65},
		{28, 71},
		{18, 70},
		{12, 65},
		{5, 62},
		{5, 58},
	})

	// AFRICA
	polygonFill([][2]float64{
		{-17, 15},  // Senegal
		{-15, 28},  // Western Sahara
		{-5, 36},   // Morocco
		{10, 37},   // Tunisia
		{32, 32},   // Libya E
		{40, 12},   // Horn of Africa
		{51, 11},   // Somalia tip
		{50, 2},    // Kenya coast
		{40, -3},   // Tanzania
		{35, -10},  // Mozambique N
		{33, -26},  // Durban
		{18, -35},  // Cape Town
		{12, -17},  // Angola
		{9, -5},    // Gabon
		{5, 4},     // Nigeria
		{-5, 5},    // Ghana
		{-8, 4},    // Ivory Coast
		{-15, 11},  // Guinea
		{-17, 15},  // close
	})

	// MADAGASCAR
	polygonFill([][2]float64{
		{43, -12},
		{50, -16},
		{47, -25},
		{43, -25},
		{43, -12},
	})

	// ASIA - main landmass
	polygonFill([][2]float64{
		{26, 42},   // Turkey
		{35, 37},   // Syria
		{45, 30},   // Iraq/Arabia
		{55, 25},   // Oman
		{60, 25},   // Pakistan
		{68, 24},   // India W
		{78, 8},    // India S tip
		{88, 22},   // Bangladesh
		{92, 18},   // Myanmar
		{100, 2},   // Malaysia
		{105, -8},  // Indonesia
		{115, -8},  // Borneo
		{120, -5},  // Sulawesi
		{125, 0},   // Moluccas
		{130, -5},  // W New Guinea
		{141, -5},  // E New Guinea
		{145, -10}, // PNG
		{150, -15}, // Coral Sea
		{148, -20}, // Queensland
		{150, -25}, // Sydney area
		{155, -28}, // Brisbane
		{153, -28}, // close
	})

	// RUSSIA/SIBERIA
	polygonFill([][2]float64{
		{28, 45},   // Black Sea
		{40, 42},   // Caucasus
		{52, 42},   // Kazakhstan
		{68, 40},   // Uzbekistan
		{80, 42},   // Kyrgyzstan
		{90, 50},   // Mongolia
		{120, 52},  // Russia E
		{135, 55},  // Sea of Okhotsk
		{170, 65},  // Kamchatka
		{180, 68},  // Bering Strait
		{170, 72},  // Arctic Russia
		{100, 72},  // Siberia N
		{50, 70},   // W Siberia
		{30, 62},   // Urals
		{28, 45},   // close
	})

	// JAPAN
	polygonFill([][2]float64{
		{130, 31},
		{131, 34},
		{134, 35},
		{137, 37},
		{140, 40},
		{142, 43},
		{145, 44},
		{145, 42},
		{141, 38},
		{138, 34},
		{135, 33},
		{130, 31},
	})

	// KOREA
	polygonFill([][2]float64{
		{125, 34},
		{126, 38},
		{129, 42},
		{130, 38},
		{128, 35},
		{125, 34},
	})

	// AUSTRALIA
	polygonFill([][2]float64{
		{114, -14},  // Darwin
		{130, -12},  // N Territory
		{142, -10},  // Cape York
		{150, -15},  // Queensland
		{153, -28},  // Brisbane
		{150, -38},  // Melbourne
		{147, -44},  // Tasmania
		{137, -35},  // SA
		{115, -35},  // Perth
		{113, -25},  // Shark Bay
		{114, -14},  // close
	})

	// NEW ZEALAND
	polygonFill([][2]float64{
		{166, -35},
		{175, -37},
		{178, -42},
		{174, -46},
		{168, -46},
		{166, -44},
		{172, -40},
		{166, -35},
	})

	// INDONESIA (major islands)
	// Sumatra
	polygonFill([][2]float64{
		{95, 6},
		{106, -2},
		{105, -6},
		{95, 2},
		{95, 6},
	})
	// Borneo
	polygonFill([][2]float64{
		{109, 7},
		{119, 7},
		{118, -1},
		{110, -4},
		{109, 2},
		{109, 7},
	})
	// Sulawesi
	polygonFill([][2]float64{
		{119, 2},
		{125, -1},
		{122, -5},
		{120, -3},
		{119, 2},
	})

	// PHILIPPINES
	polygonFill([][2]float64{
		{117, 7},
		{122, 19},
		{127, 12},
		{125, 6},
		{117, 7},
	})

	// Cities from OpenXcom regions.rul (standard coords)
	cities = []City{
		// North America
		{"New York", 143, 25},
		{"Los Angeles", 121, 28},
		{"Chicago", 136, 24},
		{"Montreal", 142, 22},
		{"Vancouver", 118, 20},
		{"Dallas", 132, 29},
		{"Mexico City", 130, 35},
		{"Havana", 141, 34},
		{"Washington", 141, 26},

		// South America
		{"Bogota", 143, 42},
		{"Lima", 142, 48},
		{"Brasilia", 156, 48},
		{"Rio de Janeiro", 158, 52},
		{"Buenos Aires", 151, 62},
		{"Santiago", 145, 62},
		{"Caracas", 147, 40},

		// Europe
		{"London", 0, 19},
		{"Paris", 1, 21},
		{"Berlin", 7, 19},
		{"Rome", 6, 24},
		{"Madrid", 179, 25},
		{"Moscow", 19, 17},
		{"Budapest", 10, 21},

		// Africa
		{"Cairo", 16, 30},
		{"Lagos", 2, 42},
		{"Cape Town", 9, 63},
		{"Nairobi", 18, 44},
		{"Casablanca", 176, 28},
		{"Pretoria", 14, 57},
		{"Kinshasa", 8, 47},

		// Asia
		{"Baghdad", 22, 28},
		{"Tehran", 26, 27},
		{"Karachi", 34, 33},
		{"Delhi", 39, 31},
		{"Bombay", 37, 36},
		{"Calcutta", 44, 34},
		{"Beijing", 58, 25},
		{"Shanghai", 61, 30},
		{"Hong Kong", 57, 34},
		{"Tokyo", 70, 27},
		{"Seoul", 64, 26},
		{"Bangkok", 50, 38},
		{"Singapore", 52, 44},
		{"Jakarta", 53, 47},
		{"Manila", 61, 37},
		{"Novosibirsk", 41, 18},

		// Australasia
		{"Sydney", 74, 63},
		{"Canberra", 75, 63},
		{"Melbourne", 73, 64},
		{"Perth", 58, 61},
		{"Wellington", 88, 66},
	}
}

func GetTile(x, y int) int {
	if x < 0 || x >= mapW || y < 0 || y >= mapH {
		return 0
	}
	return worldMap[y][x]
}

func SetTile(x, y, v int) {
	if x >= 0 && x < mapW && y >= 0 && y < mapH {
		worldMap[y][x] = v
	}
}

func IsLand(x, y int) bool {
	t := GetTile(x, y)
	return t == 1 || t == 2 || t == 3
}

func GetCities() []City {
	return cities
}

func MapSize() (int, int) {
	return mapW, mapH
}
