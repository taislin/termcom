package battle

import (
	"math/rand"

	"github.com/taislin/termcom/internal/base"
)

const (
	bdRoomW    = 9
	bdRoomH    = 9
	bdInterior = 7
	bdCorrW    = 3
	bdStride   = bdRoomW + bdCorrW  // 12
	bdBorder   = 3
	bdCols     = 8
)

func GenerateBaseDefenseMap(b *base.Base, rng *rand.Rand) *BattleMap {
	maxRow := 0
	grid := make(map[[2]int]*base.Facility)
	for _, f := range b.Facilities {
		grid[[2]int{f.Row, f.Col}] = f
		if f.Row > maxRow {
			maxRow = f.Row
		}
	}
	numRows := maxRow + 1
	if numRows < 2 {
		numRows = 2
	}

	mapW := bdBorder*2 + bdCols*bdStride - bdCorrW
	mapH := bdBorder*2 + numRows*bdStride - bdCorrW

	m := NewBattleMap(mapW, mapH)

	m.fillRect(0, 0, mapW, mapH, TileGrass)

	scatterExterior(m, rng, mapW, mapH, bdBorder)

	baseX := bdBorder
	baseY := bdBorder
	baseW := bdCols*bdStride - bdCorrW
	baseH := numRows*bdStride - bdCorrW

	m.fillRect(baseX, baseY, baseW, baseH, TileFloor)
	m.drawRect(baseX, baseY, baseW, baseH, TileWall)

	placeCorridorScatter(m, rng, baseX, baseY, baseW, baseH, grid, numRows)

	for _, f := range b.Facilities {
		roomX := baseX + f.Col*bdStride
		roomY := baseY + f.Row*bdStride

		m.drawRect(roomX, roomY, bdRoomW, bdRoomH, TileWall)
		m.fillRect(roomX+1, roomY+1, bdInterior, bdInterior, TileFloor)

		if f.Col > 0 {
			m.Set(roomX, roomY+bdRoomH/2, TileDoor)
		}
		if f.Col < bdCols-1 {
			m.Set(roomX+bdRoomW-1, roomY+bdRoomH/2, TileDoor)
		}
		if f.Row > 0 {
			m.Set(roomX+bdRoomW/2, roomY, TileDoor)
		}
		if f.Row < numRows-1 {
			m.Set(roomX+bdRoomW/2, roomY+bdRoomH-1, TileDoor)
		}

		if f.Col < bdCols-1 {
			m.Set(roomX+bdRoomW, roomY+1+rng.Intn(bdInterior), TileWindow)
			m.Set(roomX+bdRoomW, roomY+1+rng.Intn(bdInterior), TileWindow)
		}
		if f.Row < numRows-1 {
			m.Set(roomX+1+rng.Intn(bdInterior), roomY+bdRoomH, TileWindow)
			m.Set(roomX+1+rng.Intn(bdInterior), roomY+bdRoomH, TileWindow)
		}

		placeBDFurniture(m, f.Type, roomX, roomY, rng)
	}

	m.BreachPoints = addBDBreaches(m, baseX, baseY, baseW, baseH, numRows, rng)

	return m
}

func placeBDFurniture(m *BattleMap, ft base.FacilityType, roomX, roomY int, rng *rand.Rand) {
	ix := roomX + 1
	iy := roomY + 1
	switch ft {
	case base.FacLivingQuarters:
		m.Set(ix+1, iy, TileBed)
		m.Set(ix+bdInterior-2, iy, TileBed)
		m.Set(ix+1, iy+bdInterior-1, TileBed)
		m.Set(ix+bdInterior-2, iy+bdInterior-1, TileBed)
		m.Set(ix+1, iy+2, TileBed)
		m.Set(ix+bdInterior-2, iy+2, TileBed)
		m.Set(ix+bdInterior/2, iy+bdInterior/2, TileLocker)
	case base.FacLab:
		m.Set(ix+1, iy+1, TileDesk)
		m.Set(ix+bdInterior-2, iy+1, TileDesk)
		m.Set(ix+1, iy+bdInterior-2, TileComputer)
		m.Set(ix+bdInterior-2, iy+bdInterior-2, TileComputer)
		m.Set(ix+bdInterior/2, iy+2, TileDesk)
	case base.FacWorkshop:
		m.Set(ix+1, iy+1, TileMachinery)
		m.Set(ix+bdInterior-2, iy+1, TileMachinery)
		m.Set(ix+1, iy+bdInterior-2, TileStorage)
		m.Set(ix+bdInterior-2, iy+bdInterior-2, TileStorage)
		m.Set(ix+bdInterior/2, iy+bdInterior/2, TileMachinery)
	case base.FacStorage:
		m.Set(ix+1, iy+1, TileCabinet)
		m.Set(ix+bdInterior-2, iy+1, TileCabinet)
		m.Set(ix+1, iy+bdInterior-2, TileCabinet)
		m.Set(ix+bdInterior-2, iy+bdInterior-2, TileCabinet)
		m.Set(ix+2, iy+2, TileLocker)
		m.Set(ix+bdInterior-3, iy+2, TileLocker)
		m.Set(ix+2, iy+bdInterior-3, TileLocker)
		m.Set(ix+bdInterior-3, iy+bdInterior-3, TileLocker)
	case base.FacRadar:
		m.Set(ix+bdInterior/2, iy+bdInterior/2, TileDish)
		m.Set(ix+1, iy+1, TileConsole)
		m.Set(ix+bdInterior-2, iy+1, TileConsole)
		m.Set(ix+1, iy+bdInterior-2, TileDesk)
	case base.FacContainment:
		m.Set(ix+1, iy+1, TileUFOWall)
		m.Set(ix+1, iy+3, TileUFOWall)
		m.Set(ix+1, iy+5, TileUFOWall)
		m.Set(ix+bdInterior-2, iy+1, TileUFOWall)
		m.Set(ix+bdInterior-2, iy+3, TileUFOWall)
		m.Set(ix+bdInterior-2, iy+5, TileUFOWall)
		m.Set(ix+2, iy+1, TileAlienTech)
		m.Set(ix+2, iy+3, TileAlienTech)
		m.Set(ix+2, iy+5, TileAlienTech)
		m.Set(ix+bdInterior-3, iy+1, TilePod)
		m.Set(ix+bdInterior-3, iy+3, TilePod)
		m.Set(ix+bdInterior-3, iy+5, TilePod)
	case base.FacPsiLab:
		m.Set(ix+bdInterior/2, iy+bdInterior/2, TileChair)
		m.Set(ix+1, iy+1, TileConsole)
		m.Set(ix+bdInterior-2, iy+1, TileConsole)
		m.Set(ix+1, iy+bdInterior-2, TileDesk)
	case base.FacHangar:
		m.fillRect(ix, iy, bdInterior, bdInterior, TilePavement)
		m.Set(ix+2, iy+bdInterior/2, TileForklift)
		m.Set(ix+3, iy+bdInterior/2, TileForkliftRight)
		m.Set(ix+bdInterior-3, iy+bdInterior/2, TileTruck)
		m.Set(ix+bdInterior/2, iy+2, TileMachinery)
		m.Set(ix+bdInterior/2, iy+bdInterior-3, TileCabinet)
	}
}

func scatterExterior(m *BattleMap, rng *rand.Rand, mapW, mapH, border int) {
	for y := 0; y < mapH; y++ {
		for x := 0; x < mapW; x++ {
			if x >= border && x < mapW-border && y >= border && y < mapH-border {
				continue
			}
			if rng.Intn(100) < 8 {
				switch rng.Intn(5) {
				case 0:
					m.Set(x, y, TileBush)
				case 1:
					m.Set(x, y, TileRock)
				case 2:
					m.Set(x, y, TileFence)
				case 3:
					m.Set(x, y, TileTree)
				case 4:
					m.Set(x, y, TileRubble)
				}
			}
		}
	}
}

func placeCorridorScatter(m *BattleMap, rng *rand.Rand, baseX, baseY, baseW, baseH int, grid map[[2]int]*base.Facility, numRows int) {
	corrFloor := make(map[[2]int]bool)
	for y := baseY; y < baseY+baseH; y++ {
		for x := baseX; x < baseX+baseW; x++ {
			if m.At(x, y).Type == TileFloor {
				corrFloor[[2]int{x, y}] = true
			}
		}
	}
	for pt := range corrFloor {
		if rng.Intn(100) < 6 {
			switch rng.Intn(4) {
			case 0:
				m.Set(pt[0], pt[1], TileContainerRed)
			case 1:
				m.Set(pt[0], pt[1], TileContainerBlue)
			case 2:
				m.Set(pt[0], pt[1], TileDebris)
			case 3:
				m.Set(pt[0], pt[1], TileDesk)
			}
		}
	}
}

func addBDBreaches(m *BattleMap, baseX, baseY, baseW, baseH, numRows int, rng *rand.Rand) [][2]int {
	var pts [][2]int

	breachCount := 3 + rng.Intn(3)
	for i := 0; i < breachCount; i++ {
		side := rng.Intn(4)
		var bx, by int
		switch side {
		case 0:
			bx = baseX + bdStride/2 + rng.Intn(max(baseW-bdStride, 1))
			by = baseY - 1
		case 1:
			bx = baseX + bdStride/2 + rng.Intn(max(baseW-bdStride, 1))
			by = baseY + baseH
		case 2:
			bx = baseX - 1
			by = baseY + bdStride/2 + rng.Intn(max(baseH-bdStride, 1))
		case 3:
			bx = baseX + baseW
			by = baseY + bdStride/2 + rng.Intn(max(baseH-bdStride, 1))
		}

		if bx >= 0 && bx < m.Width && by >= 0 && by < m.Height {
			m.Set(bx, by, TileDoorOpen)
			pts = append(pts, [2]int{bx, by})
		}
	}

	return pts
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
