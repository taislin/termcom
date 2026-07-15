package engine

const mobileWidthThreshold = 100

type LayoutMode int

const (
	LayoutFull LayoutMode = iota
	LayoutMobile
)

type LayoutManager struct {
	Mode              LayoutMode
	BattleSidebarOpen bool
	GeoMinimapOpen    bool
}

var Layout = &LayoutManager{
	Mode:              LayoutFull,
	BattleSidebarOpen: true,
	GeoMinimapOpen:    true,
}

func (lm *LayoutManager) UpdateMode(w, h int) {
	if Config.TouchMode && w < mobileWidthThreshold {
		lm.Mode = LayoutMobile
	} else {
		lm.Mode = LayoutFull
	}
}

func (lm *LayoutManager) IsMobile() bool {
	return lm.Mode == LayoutMobile
}

func (lm *LayoutManager) ToggleBattleSidebar() {
	lm.BattleSidebarOpen = !lm.BattleSidebarOpen
}

func (lm *LayoutManager) ToggleGeoMinimap() {
	lm.GeoMinimapOpen = !lm.GeoMinimapOpen
}

func (lm *LayoutManager) BattleSidebarWidth(w int) int {
	if lm.IsMobile() {
		return w - 2
	}
	if !lm.BattleSidebarOpen {
		return 0
	}
	sw := w / 3
	if sw < 30 {
		sw = 30
	}
	return sw
}

// BattleSidebarY returns the screen row where the sidebar begins. On mobile the
// sidebar is stacked underneath the (top) battle view; on desktop it sits beside
// the view at the top.
func (lm *LayoutManager) BattleSidebarY(h int) int {
	if lm.IsMobile() {
		return lm.BattleViewHeight(h) + 3
	}
	return 1
}

func (lm *LayoutManager) BattleViewWidth(w int) int {
	if lm.IsMobile() {
		vw := 10
		if vw > w-2 {
			vw = w - 2
		}
		return vw
	}
	sw := lm.BattleSidebarWidth(w)
	if sw == 0 {
		return w - 2
	}
	vw := w - sw - 2
	if vw < 10 {
		vw = 10
	}
	return vw
}

func (lm *LayoutManager) BattleViewHeight(h int) int {
	if lm.IsMobile() {
		return 10
	}
	return h - 5
}

func (lm *LayoutManager) BattleSidebarX(w int) int {
	vw := lm.BattleViewWidth(w)
	return vw + 2
}

func (lm *LayoutManager) GeoTableWidth(w int) int {
	if lm.IsMobile() {
		return w - 2
	}
	if !lm.GeoMinimapOpen {
		return w - 2
	}
	tw := w * 60 / 100
	if tw < 30 {
		tw = 30
	}
	return tw
}

func (lm *LayoutManager) GeoMapWidth(w int) int {
	if lm.IsMobile() {
		return w - 2
	}
	if !lm.GeoMinimapOpen {
		return 0
	}
	tw := lm.GeoTableWidth(w)
	return w - tw - 2
}

func (lm *LayoutManager) GeoMapX(w int) int {
	if lm.IsMobile() {
		return 1
	}
	tw := lm.GeoTableWidth(w)
	return tw + 2
}

func (lm *LayoutManager) EquipSplitX(w int) int {
	if lm.IsMobile() {
		return 0
	}
	return w / 2
}

func (lm *LayoutManager) EncyclopediaListWidth(w int) int {
	if lm.IsMobile() {
		return w - 2
	}
	lw := w / 3
	if lw < 20 {
		lw = 20
	}
	return lw
}

func (lm *LayoutManager) EncyclopediaInfoX(w int) int {
	if lm.IsMobile() {
		return 0
	}
	lw := lm.EncyclopediaListWidth(w)
	return lw + 3
}

func (lm *LayoutManager) EncyclopediaInfoWidth(w int) int {
	if lm.IsMobile() {
		return w - 2
	}
	ix := lm.EncyclopediaInfoX(w)
	return w - ix - 2
}

func (lm *LayoutManager) CustomBattleLeftWidth(w int) int {
	if lm.IsMobile() {
		return w - 2
	}
	return w/2 - 1
}

func (lm *LayoutManager) CustomBattleRightX(w int) int {
	if lm.IsMobile() {
		return 0
	}
	lw := lm.CustomBattleLeftWidth(w)
	return lw + 2
}

func (lm *LayoutManager) MinSidebarWidth(w int) int {
	sw := w / 3
	if sw < 30 {
		sw = 30
	}
	return sw
}
