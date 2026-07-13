package geo

// Equirectangular world map. Each char = 1 cell. Land is 1, water is 0.
// Map is 180 wide x 90 tall (2-degree resolution).
// x = longitude/2 (0..360 -> 0..179)
// y = (90 - latitude)/2 (90N..90S -> 0..89)
const mapW = 180
const mapH = 90

var worldMap [mapH][mapW]int

type City struct {
	ID               int
	Name             string
	X, Y             int // screen coordinates (0-179, 0-89)
	Region           string
	Threat           int // 0-100, alien activity level
	HasRadar         bool
	InterceptorCount int
	MissionHere      bool
}

var cities []*City

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

	cities = []*City{
		// North America
		{ID: 0, Name: "New York", X: 48, Y: 31, Region: "NA East"},
		{ID: 1, Name: "Los Angeles", X: 28, Y: 34, Region: "NA West"},
		{ID: 2, Name: "Chicago", X: 41, Y: 30, Region: "NA Central"},
		{ID: 3, Name: "Mexico City", X: 35, Y: 41, Region: "Central Am"},
		// South America
		{ID: 4, Name: "Bogota", X: 48, Y: 48, Region: "SA North"},
		{ID: 5, Name: "Brasilia", X: 61, Y: 54, Region: "SA East"},
		{ID: 6, Name: "Buenos Aires", X: 56, Y: 68, Region: "SA South"},
		// Europe
		{ID: 7, Name: "London", X: 85, Y: 25, Region: "Europe W"},
		{ID: 8, Name: "Paris", X: 86, Y: 27, Region: "Europe W"},
		{ID: 9, Name: "Berlin", X: 92, Y: 25, Region: "Europe C"},
		{ID: 10, Name: "Moscow", X: 104, Y: 23, Region: "Europe E"},
		// Africa
		{ID: 11, Name: "Cairo", X: 102, Y: 36, Region: "Africa N"},
		{ID: 12, Name: "Lagos", X: 87, Y: 48, Region: "Africa W"},
		{ID: 13, Name: "Nairobi", X: 103, Y: 50, Region: "Africa E"},
		// Asia
		{ID: 14, Name: "Delhi", X: 124, Y: 37, Region: "South Asia"},
		{ID: 15, Name: "Beijing", X: 143, Y: 31, Region: "East Asia"},
		{ID: 16, Name: "Tokyo", X: 154, Y: 33, Region: "East Asia"},
		{ID: 17, Name: "Singapore", X: 137, Y: 50, Region: "SE Asia"},
		// Australasia
		{ID: 18, Name: "Sydney", X: 159, Y: 69, Region: "Oceania"},
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
	return t == 1
}

func GetCities() []*City {
	return cities
}

func MapSize() (int, int) {
	return mapW, mapH
}
