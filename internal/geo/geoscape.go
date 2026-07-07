package geo

import (
	"fmt"
	"time"

	"github.com/civ13/ycom/internal/engine"
	"github.com/gdamore/tcell/v2"
)

type Geoscape struct {
	Game          *engine.Game
	UFOs          UFOList
	Interceptors  InterceptorList
	BaseX, BaseY  int
	BaseName      string
	Message       string
	MessageTimer  time.Time
	TickCounter   int
}

func NewGeoscape(g *engine.Game) *Geoscape {
	return &Geoscape{
		Game:         g,
		BaseX:        28,
		BaseY:        32,
		BaseName:     "Base 1",
		Message:      "Welcome, Commander. Your mission: defend Earth from alien invasion.",
		MessageTimer: time.Now(),
	}
}

func (gs *Geoscape) Update() {
	gs.TickCounter++

	if gs.TickCounter%600 == 0 && gs.UFOs.Count() < 5 {
		ufo := SpawnUFO()
		gs.UFOs = append(gs.UFOs, ufo)
		gs.Message = fmt.Sprintf("UFO detected! %s at [%d,%d]", ufo.Type.Name, ufo.TileX(), ufo.TileY())
		gs.MessageTimer = time.Now()
	}

	for _, u := range gs.UFOs {
		u.Update()
	}

	for _, i := range gs.Interceptors {
		if i.Launching {
			reached := i.Update()
			if reached {
				gs.dogfight(i)
			}
		}
	}

	if !gs.Game.Paused && gs.Game.TimeSpeed > 0 {
		speedMult := []int{0, 1, 5, 20, 60}
		minutes := speedMult[gs.Game.TimeSpeed]
		gs.Game.GameTime = gs.Game.GameTime.Add(time.Duration(minutes) * time.Minute)
	}
}

func (gs *Geoscape) dogfight(inter *Interceptor) {
	if inter.Target == nil {
		return
	}
	ufo := inter.Target
	damage := inter.FireAt(ufo)
	if damage == -1 {
		gs.Message = fmt.Sprintf("UFO DESTROYED! %s shot down.", ufo.Type.Name)
		gs.MessageTimer = time.Now()
		gs.Game.Funds += int64(ufo.Type.Points * 1000)
		inter.Disengage()
	} else {
		gs.Message = fmt.Sprintf("Hit UFO for %d damage!", damage)
		gs.MessageTimer = time.Now()
	}
}

func (gs *Geoscape) TogglePause() {
	gs.Game.Paused = !gs.Game.Paused
	if gs.Game.Paused {
		gs.Message = "TIME PAUSED"
	} else {
		gs.Message = fmt.Sprintf("TIME RUNNING (Speed %dx)", gs.Game.TimeSpeed)
	}
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) SetSpeed(s int) {
	gs.Game.TimeSpeed = s
	gs.Game.Paused = false
	gs.Message = fmt.Sprintf("Time speed: %dx", s)
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) LaunchInterceptor() {
	var nearest *UFO
	bestDist := 9999.0
	for _, u := range gs.UFOs {
		if !u.Active {
			continue
		}
		dx := u.X - float64(gs.BaseX)
		dy := u.Y - float64(gs.BaseY)
		dist := dx*dx + dy*dy
		if dist < bestDist {
			bestDist = dist
			nearest = u
		}
	}
	if nearest == nil {
		gs.Message = "No UFOs detected on radar."
		gs.MessageTimer = time.Now()
		return
	}

	inter := NewInterceptor(gs.BaseX, gs.BaseY)
	inter.Launch(nearest)
	gs.Interceptors = append(gs.Interceptors, inter)
	gs.Message = fmt.Sprintf("Interceptor launched! Pursuing %s.", nearest.Type.Name)
	gs.MessageTimer = time.Now()
}

func (gs *Geoscape) Render(ctx *engine.ScreenCtx) {
	w, h := ctx.Size()
	mw, mh := MapSize()

	offsetX := 0
	if mw > w-2 {
		offsetX = (mw - w + 2) / 2
	}
	offsetY := 0
	if mh > h-6 {
		offsetY = (mh - h + 6) / 2
	}

	for y := 0; y < h-6 && y < mh; y++ {
		for x := 0; x < w-2 && x < mw; x++ {
			mx := x + offsetX
			my := y + offsetY
			if mx >= mw || my >= mh {
				continue
			}
			tile := GetTile(mx, my)
			ch := '·'
			style := engine.StyleBlue
			switch tile {
			case 1:
				ch = '.'
				style = engine.StyleGreen
			case 2:
				ch = '○'
				style = engine.StyleYellow
			case 3:
				ch = '▲'
				style = engine.StyleCyan
			case 4:
				ch = '?'
				style = engine.StyleRed
			case 5:
				ch = '▸'
				style = engine.StyleCyan
			}
			ctx.SetCell(x+1, y+1, ch, style)
		}
	}

	for _, c := range cities {
		sx := c.X - offsetX + 1
		sy := c.Y - offsetY + 1
		if sx > 0 && sx < w-1 && sy > 0 && sy < h-6 {
			ctx.SetCell(sx, sy, '●', engine.StyleYellow)
			if len(c.Name) < 6 {
				ctx.DrawString(sx+1, sy, c.Name[:5], engine.StyleGray)
			}
		}
	}

	bsx := gs.BaseX - offsetX + 1
	bsy := gs.BaseY - offsetY + 1
	if bsx > 0 && bsx < w-1 && bsy > 0 && bsy < h-6 {
		ctx.SetCell(bsx, bsy, '▲', engine.StyleCyanBold)
	}

	for _, u := range gs.UFOs {
		if !u.Active {
			continue
		}
		ux := int(u.X) - offsetX + 1
		uy := int(u.Y) - offsetY + 1
		if ux > 0 && ux < w-1 && uy > 0 && uy < h-6 {
			ctx.SetCell(ux, uy, '?', engine.StyleRedBold)
		}
	}

	for _, i := range gs.Interceptors {
		if i.HP <= 0 {
			continue
		}
		ix := int(i.X) - offsetX + 1
		iy := int(i.Y) - offsetY + 1
		if ix > 0 && ix < w-1 && iy > 0 && iy < h-6 {
			ctx.SetCell(ix, iy, '▸', engine.StyleCyanBold)
		}
	}

	ctx.DrawPanel(0, h-4, w, 3, "GEOSCAPE", engine.StyleDefault)
	fundsStr := fmt.Sprintf("Funds: $%dK", gs.Game.Funds/1000)
	timeStr := fmt.Sprintf("Time: %s", gs.Game.GameTime.Format("02/01/2006 15:04"))
	pauseStr := "RUNNING"
	if gs.Game.Paused {
		pauseStr = "PAUSED"
	}
	ctx.DrawString(2, h-3, fundsStr, engine.StyleGreen)
	ctx.DrawString(w/3, h-3, timeStr, engine.StyleDefault)
	ctx.DrawString(w*2/3, h-3, pauseStr, engine.StyleYellow)

	if time.Since(gs.MessageTimer) < 4*time.Second && gs.Message != "" {
		ctx.DrawString(2, h-2, gs.Message, engine.StyleDefault)
	}

	ctx.DrawPanel(0, h-1, w, 1, "", engine.StyleGray)
	ctx.DrawString(1, h-1, "[B]ase  [L]aunch  Space=Pause  1-4=Speed  Q=Quit", engine.StyleGray)
}

func (gs *Geoscape) HandleKey(e *tcell.EventKey) {
	switch e.Key() {
	case tcell.KeyUp:
		gs.BaseY--
		if gs.BaseY < 0 {
			gs.BaseY = 0
		}
	case tcell.KeyDown:
		gs.BaseY++
		if gs.BaseY >= 90 {
			gs.BaseY = 89
		}
	case tcell.KeyLeft:
		gs.BaseX--
		if gs.BaseX < 0 {
			gs.BaseX = 0
		}
	case tcell.KeyRight:
		gs.BaseX++
		if gs.BaseX >= 180 {
			gs.BaseX = 179
		}
	case tcell.KeyRune:
		switch e.Rune() {
		case 'b', 'B':
			gs.Game.PushState(engine.StateBase)
		case 'l', 'L':
			gs.LaunchInterceptor()
		case ' ':
			gs.TogglePause()
		case '1':
			gs.SetSpeed(1)
		case '2':
			gs.SetSpeed(2)
		case '3':
			gs.SetSpeed(3)
		case '4':
			gs.SetSpeed(4)
		case 'q', 'Q':
			gs.Game.Quit()
		}
	}
}

func (gs *Geoscape) HandleMouse(e *tcell.EventMouse) {
	buttons := e.Buttons()
	if buttons == 0 {
		return
	}
	x, y := e.Position()
	w, h := gs.Game.ScreenSize()

	// Click on status bar buttons
	if y >= h-4 && y <= h-2 {
		switch {
		case x >= 1 && x <= 8:
			gs.TogglePause()
		case x >= 10 && x <= 20:
			gs.LaunchInterceptor()
		}
		return
	}

	// Click on map to set base position
	if y > 0 && y < h-4 && x > 0 && x < w-1 {
		mw, mh := MapSize()
		offsetX := 0
		if mw > w-2 {
			offsetX = (mw - w + 2) / 2
		}
		offsetY := 0
		if mh > h-6 {
			offsetY = (mh - h + 6) / 2
		}
		mx := x - 1 + offsetX
		my := y - 1 + offsetY
		if mx >= 0 && mx < mw && my >= 0 && my < mh {
			gs.BaseX = mx
			gs.BaseY = my
			gs.Message = fmt.Sprintf("Base moved to [%d,%d]", mx, my)
			gs.MessageTimer = time.Now()
		}
	}
}
