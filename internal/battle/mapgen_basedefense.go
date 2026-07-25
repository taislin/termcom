package battle

import (
	"math/rand"

	"github.com/taislin/termcom/internal/base"
)

const (
	bdRoomW   = 7
	bdRoomH   = 7
	bdInterior = 5
	bdCorrW   = 2
	bdStride  = bdRoomW + bdCorrW  // 9
	bdBorder  = 2
	bdCols    = 8
)

func GenerateBaseDefenseMap(b *base.Base, rng *rand.Rand) *BattleMap {
	maxRow := 0
	type pos struct{ r, c int }
	grid := make(map[pos]*base.Facility)
	for _, f := range b.Facilities {
		grid[pos{f.Row, f.Col}] = f
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

	baseX := bdBorder
	baseY := bdBorder
	baseW := bdCols*bdStride - bdCorrW
	baseH := numRows*bdStride - bdCorrW

	m.fillRect(baseX, baseY, baseW, baseH, TileFloor)
	m.drawRect(baseX, baseY, baseW, baseH, TileWall)

	for _, f := range b.Facilities {
		roomX := baseX + f.Col*bdStride
		roomY := baseY + f.Row*bdStride

		m.drawRect(roomX, roomY, bdRoomW, bdRoomH, TileWall)
		m.fillRect(roomX+1, roomY+1, bdInterior, bdInterior, TileFloor)

		if f.Row > 0 {
			m.Set(roomX+bdRoomW/2, roomY, TileDoor)
		}
		if f.Row < numRows-1 {
			m.Set(roomX+bdRoomW/2, roomY+bdRoomH-1, TileDoor)
		}
		if f.Col > 0 {
			m.Set(roomX, roomY+bdRoomH/2, TileDoor)
		}
		if f.Col < bdCols-1 {
			m.Set(roomX+bdRoomW-1, roomY+bdRoomH/2, TileDoor)
		}

		placeBDFurniture(m, f.Type, roomX, roomY, rng)
	}

	breaches := addBDBreaches(m, baseX, baseY, baseW, baseH, numRows, rng)

	m.BreachPoints = breaches

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
		m.Set(ix+bdInterior/2, iy+bdInterior/2, TileLocker)
	case base.FacLab:
		m.Set(ix+1, iy+1, TileDesk)
		m.Set(ix+bdInterior-2, iy+1, TileDesk)
		m.Set(ix+bdInterior/2, iy+bdInterior-2, TileComputer)
	case base.FacWorkshop:
		m.Set(ix+1, iy+1, TileMachinery)
		m.Set(ix+bdInterior-2, iy+bdInterior-2, TileStorage)
	case base.FacStorage:
		m.Set(ix+1, iy+1, TileCabinet)
		m.Set(ix+bdInterior-2, iy+1, TileCabinet)
		m.Set(ix+1, iy+bdInterior-2, TileLocker)
		m.Set(ix+bdInterior-2, iy+bdInterior-2, TileLocker)
	case base.FacRadar:
		m.Set(ix+bdInterior/2, iy+bdInterior/2, TileDish)
		m.Set(ix+1, iy+1, TileConsole)
	case base.FacContainment:
		m.Set(ix+1, iy+1, TileUFOWall)
		m.Set(ix+bdInterior-2, iy+1, TileUFOWall)
		m.Set(ix+bdInterior/2, iy+bdInterior-2, TilePod)
		m.Set(ix+bdInterior/2, iy+1, TileAlienTech)
	case base.FacPsiLab:
		m.Set(ix+bdInterior/2, iy+bdInterior/2, TileChair)
		m.Set(ix+1, iy+1, TileConsole)
	case base.FacHangar:
		m.Set(ix+bdInterior/2, iy+bdInterior-1, TileMachinery)
		if rng.Intn(2) == 0 {
			m.Set(ix+1, iy+bdInterior/2, TileForklift)
			m.Set(ix+2, iy+bdInterior/2, TileForkliftRight)
		}
	}
}

func addBDBreaches(m *BattleMap, baseX, baseY, baseW, baseH, numRows int, rng *rand.Rand) [][2]int {
	var pts [][2]int

	numBreaches := 2 + rng.Intn(3)
	for i := 0; i < numBreaches; i++ {
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
