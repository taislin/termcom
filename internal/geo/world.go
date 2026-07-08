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



func init() {
	for y := 0; y < mapH; y++ {
		for x := 0; x < mapW; x++ {
			worldMap[y][x] = 0
		}
	}

	for y, line := range worldLines {
		if y >= mapH {
			break
		}
		for x := 0; x < mapW && x < len(line); x++ {
			if line[x] == '#' {
				worldMap[y][x] = 1
			}
		}
	}

	cities = []City{
		// North America
		{"New York", 48, 31},
		{"Los Angeles", 26, 34},
		{"Chicago", 41, 30},
		{"Montreal", 47, 28},
		{"Vancouver", 25, 26},
		{"Dallas", 37, 35},
		{"Mexico City", 35, 41},
		{"Havana", 46, 40},
		{"Washington", 46, 32},

		// South America
		{"Bogota", 48, 48},
		{"Lima", 47, 54},
		{"Brasilia", 61, 54},
		{"Rio de Janeiro", 63, 58},
		{"Buenos Aires", 56, 68},
		{"Santiago", 50, 68},
		{"Caracas", 52, 46},

		// Europe
		{"London", 85, 25},
		{"Paris", 86, 27},
		{"Berlin", 92, 25},
		{"Rome", 91, 30},
		{"Madrid", 84, 31},
		{"Moscow", 104, 23},
		{"Budapest", 95, 27},

		// Africa
		{"Cairo", 101, 36},
		{"Lagos", 87, 48},
		{"Cape Town", 94, 69},
		{"Nairobi", 103, 50},
		{"Casablanca", 81, 34},
		{"Pretoria", 99, 63},
		{"Kinshasa", 93, 53},

		// Asia
		{"Baghdad", 107, 34},
		{"Tehran", 111, 33},
		{"Karachi", 119, 39},
		{"Delhi", 124, 37},
		{"Bombay", 122, 42},
		{"Calcutta", 129, 40},
		{"Beijing", 143, 31},
		{"Shanghai", 146, 36},
		{"Hong Kong", 142, 40},
		{"Tokyo", 155, 33},
		{"Seoul", 149, 32},
		{"Bangkok", 135, 44},
		{"Singapore", 137, 50},
		{"Jakarta", 138, 53},
		{"Manila", 146, 43},
		{"Novosibirsk", 26, 24},

		// Australasia
		{"Sydney", 159, 69},
		{"Canberra", 160, 69},
		{"Melbourne", 158, 70},
		{"Perth", 143, 67},
		{"Wellington", 173, 72},
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
