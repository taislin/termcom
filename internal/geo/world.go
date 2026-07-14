package geo

import "github.com/taislin/termcom/internal/language"

// Equirectangular world map. Each char = 1 cell. Land is 1, water is 0.
// Map is 180 wide x 90 tall (2-degree resolution).
// x = longitude/2 (0..360 -> 0..179)
// y = (90 - latitude)/2 (90N..90S -> 0..89)
const mapW = 180
const mapH = 90

var worldMap [mapH][mapW]int

type City struct {
	ID               int
	NameKey          string
	RegionKey        string
	X, Y             int // screen coordinates (0-179, 0-89)
	Threat           int // 0-100, alien activity level
	HasRadar         bool
	InterceptorCount int
	MissionHere      bool
}

func (c *City) LangName() string  { return language.String(c.NameKey) }
func (c *City) LangRegion() string { return language.String(c.RegionKey) }

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
		{ID: 0, NameKey: "CITY_NEW_YORK", X: 48, Y: 33, RegionKey: "REGION_NA_EAST"},
		{ID: 1, NameKey: "CITY_LOS_ANGELES", X: 28, Y: 36, RegionKey: "REGION_NA_WEST"},
		{ID: 2, NameKey: "CITY_CHICAGO", X: 41, Y: 32, RegionKey: "REGION_NA_CENTRAL"},
		{ID: 3, NameKey: "CITY_MEXICO_CITY", X: 35, Y: 43, RegionKey: "REGION_CENTRAL_AM"},
		// South America
		{ID: 4, NameKey: "CITY_BOGOTA", X: 48, Y: 50, RegionKey: "REGION_SA_NORTH"},
		{ID: 5, NameKey: "CITY_BRASILIA", X: 61, Y: 56, RegionKey: "REGION_SA_EAST"},
		{ID: 6, NameKey: "CITY_BUENOS_AIRES", X: 56, Y: 70, RegionKey: "REGION_SA_SOUTH"},
		// Europe
		{ID: 7, NameKey: "CITY_LONDON", X: 85, Y: 27, RegionKey: "REGION_EUROPE_W"},
		{ID: 8, NameKey: "CITY_PARIS", X: 86, Y: 29, RegionKey: "REGION_EUROPE_W"},
		{ID: 9, NameKey: "CITY_BERLIN", X: 92, Y: 27, RegionKey: "REGION_EUROPE_C"},
		{ID: 10, NameKey: "CITY_MOSCOW", X: 104, Y: 25, RegionKey: "REGION_EUROPE_E"},
		// Africa
		{ID: 11, NameKey: "CITY_CAIRO", X: 102, Y: 38, RegionKey: "REGION_AFRICA_N"},
		{ID: 12, NameKey: "CITY_LAGOS", X: 87, Y: 50, RegionKey: "REGION_AFRICA_W"},
		{ID: 13, NameKey: "CITY_NAIROBI", X: 103, Y: 52, RegionKey: "REGION_AFRICA_E"},
		// Asia
		{ID: 14, NameKey: "CITY_DELHI", X: 124, Y: 39, RegionKey: "REGION_SOUTH_ASIA"},
		{ID: 15, NameKey: "CITY_BEIJING", X: 143, Y: 33, RegionKey: "REGION_EAST_ASIA"},
		{ID: 16, NameKey: "CITY_TOKYO", X: 154, Y: 33, RegionKey: "REGION_EAST_ASIA"},
		{ID: 17, NameKey: "CITY_SINGAPORE", X: 137, Y: 52, RegionKey: "REGION_SE_ASIA"},
		// Australasia
		{ID: 18, NameKey: "CITY_SYDNEY", X: 159, Y: 69, RegionKey: "REGION_OCEANIA"},
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
