package geo

// Equirectangular world map. Each char = 1 cell. Land is '#' or '.', water is '·'.
// Map is 180 wide x 90 tall (2-degree resolution).
const mapW = 180
const mapH = 90

// worldMap stores the ASCII map. 0=water, 1=land, 2=city, 3=base, 4=ufo, 5=interceptor
var worldMap [mapH][mapW]int

func init() {
	for y := 0; y < mapH; y++ {
		for x := 0; x < mapW; x++ {
			worldMap[y][x] = 0
		}
	}
	// Rough continent shapes
	// North America
	setLand(8, 18, 20, 14)
	// Central America
	setLand(14, 28, 6, 6)
	// South America
	setLand(12, 34, 12, 20)
	// Europe
	setLand(72, 14, 20, 14)
	// Africa
	setLand(72, 28, 20, 22)
	// Asia
	setLand(92, 10, 50, 24)
	// Middle East
	setLand(82, 22, 12, 10)
	// India
	setLand(100, 24, 14, 12)
	// SE Asia
	setLand(110, 30, 14, 10)
	// Australia
	setLand(118, 42, 16, 10)

	// Major cities (lon, lat → x, y)
	cities = []City{
		{"New York", 292, 58},
		{"Los Angeles", 242, 60},
		{"Chicago", 268, 58},
		{"London", 345, 54},
		{"Paris", 350, 56},
		{"Moscow", 380, 52},
		{"Tokyo", 450, 58},
		{"Beijing", 430, 56},
		{"Sydney", 470, 72},
		{"Cairo", 370, 62},
		{"Rio de Janeiro", 315, 72},
		{"Mexico City", 255, 64},
		{"Buenos Aires", 305, 76},
		{"Lima", 285, 70},
		{"Delhi", 410, 60},
		{"Singapore", 428, 66},
		{"Cape Town", 365, 80},
		{"Lagos", 345, 68},
		{"Berlin", 358, 52},
		{"Rome", 362, 56},
	}
}

type City struct {
	Name string
	X    int // longitude 0-360 → 0-179 on map
	Y    int // latitude 0-180 → 0-89 on map (inverted)
}

var cities []City

func setLand(x, y, w, h int) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			mx := x + dx
			my := y + dy
			if mx >= 0 && mx < mapW && my >= 0 && my < mapH {
				// Irregular edges
				if dy == 0 || dy == h-1 || dx == 0 || dx == w-1 {
					if (mx+my)%3 != 0 {
						worldMap[my][mx] = 1
					}
				} else {
					worldMap[my][mx] = 1
				}
			}
		}
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
