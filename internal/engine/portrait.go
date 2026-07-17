package engine

import (
	"math"
	"sync"

	"github.com/gdamore/tcell/v3"
	"github.com/taislin/termcom/internal/data"
)

var (
	alienSpriteCache sync.Map
	portraitBg       = tcell.NewRGBColor(20, 20, 28)
)



type spriteCacheKey struct {
	seed          int64
	bgR, bgG, bgB int32
	mk            uint64
}

func morphKey(m *data.Morphology) uint64 {
	if m == nil {
		return 0
	}
	h := uint64(m.Arms) | uint64(m.Legs)<<8
	h |= hashStr(m.Eyesight) << 16
	h |= hashStr(m.Hearing) << 24
	h |= hashStr(m.PsionicSense) << 32
	h |= hashStr(m.ThermalSense) << 40
	h |= hashStr(m.ChemicalSense) << 48
	return h
}

func hashStr(s string) uint64 {
	var h uint64
	for i := 0; i < len(s); i++ {
		h = h*31 + uint64(s[i])
	}
	return h
}

type PortraitSpec struct {
	Width     int
	Height    int
	SkinColor tcell.Color
	EyeColor  tcell.Color
	HairColor tcell.Color
	Seed      int64
}

// MakeSoldierPortrait builds a portrait from a soldier's name.
func MakeSoldierPortrait(name string, w, h int) *PixelImage {
	var nameSeed int64
	for _, r := range name {
		nameSeed += int64(r)
	}
	skinColors := []tcell.Color{
		tcell.NewRGBColor(230, 190, 160),
		tcell.NewRGBColor(140, 95, 60),
		tcell.NewRGBColor(240, 220, 200),
		tcell.NewRGBColor(190, 150, 120),
		tcell.NewRGBColor(170, 120, 80),
		tcell.NewRGBColor(210, 170, 140),
	}
	skinColor := skinColors[nameSeed%int64(len(skinColors))]

	eyeColors := []tcell.Color{
		tcell.NewRGBColor(50, 100, 200),
		tcell.NewRGBColor(40, 150, 50),
		tcell.NewRGBColor(100, 60, 30),
		tcell.NewRGBColor(70, 130, 70),
	}
	eyeColor := eyeColors[(nameSeed/3)%int64(len(eyeColors))]

	hairColors := []tcell.Color{
		tcell.NewRGBColor(10, 10, 10),
		tcell.NewRGBColor(120, 60, 20),
		tcell.NewRGBColor(230, 200, 50),
		tcell.NewRGBColor(200, 80, 20),
		tcell.NewRGBColor(160, 140, 100),
		tcell.NewRGBColor(80, 50, 30),
	}
	hairColor := hairColors[(nameSeed/7)%int64(len(hairColors))]

	return GenerateSoldierPortrait(PortraitSpec{
		Width:     w,
		Height:    h,
		SkinColor: skinColor,
		EyeColor:  eyeColor,
		HairColor: hairColor,
		Seed:      nameSeed,
	})
}

// GenerateSoldierPortrait generates a procedural soldier portrait with stacked layers.
func GenerateSoldierPortrait(spec PortraitSpec) *PixelImage {
	w, h := spec.Width, spec.Height
	if w <= 0 {
		w = 16
	}
	if h <= 0 {
		h = 24
	}

	rng := rngFromSeed(spec.Seed)

	bgColor := portraitBg
	skin := generateSkinLayer(w, h, spec.SkinColor, bgColor)
	eyes := generateEyeLayer(w, h, spec.EyeColor, spec.SkinColor)
	nose := generateNoseLayer(w, h, spec.SkinColor)
	mouth := generateMouthLayer(w, h, spec.SkinColor)
	hair := generateHairLayer(w, h, spec.HairColor, rng.Intn(8))

	res := skin
	res = CompositeImages(res, eyes)
	res = CompositeImages(res, nose)
	res = CompositeImages(res, mouth)
	res = CompositeImages(res, hair)

	applyPortraitDithering(res, spec)

	return res
}

// applyPortraitDithering adds dramatic shading, texture noise, and depth
// to soldier portraits. Effects accumulate: each reads the current pixel
// so edge shading, noise, and cheek shadows all stack.
func applyPortraitDithering(img *PixelImage, spec PortraitSpec) {
	w, h := img.Width, img.Height
	g := computeFaceGeom(w, h)
	rng := rngFromSeed(spec.Seed + 9999)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.Pixels[y][x]
			if c == tcell.ColorDefault || c == portraitBg {
				continue
			}
			isSkin := isSkinTone(c) && inHead(x, y, g)
			isHair := isHairColor(c, spec.HairColor)
			if !isSkin && !isHair {
				continue
			}

			// --- SKIN EFFECTS ---
			if isSkin {
				noise := rng.Intn(100)
				if noise < 25 {
					img.Pixels[y][x] = DarkenColor(c, 0.72)
				} else if noise < 35 {
					img.Pixels[y][x] = LightenColor(c, 1.15)
				}
			}

			if isSkin {
				nx := float64(x-g.cx) / float64(g.rx)
				ny := float64(y-g.cy) / float64(g.ry)
				dist := nx*nx + ny*ny
				if dist > 0.7 {
					f := 1.0 - (dist-0.7)*0.6
					if f < 0.55 {
						f = 0.55
					}
					img.Pixels[y][x] = DarkenColor(c, f)
				}
			}

			if isSkin {
				dy := float64(y-g.cy) / float64(g.ry)
				if dy > 0.35 {
					f := 1.0 - (dy-0.35)*0.6
					if f < 0.4 {
						f = 0.4
					}
					img.Pixels[y][x] = DarkenColor(c, f)
				}
			}

			if isSkin && x == g.cx && y >= g.cy-g.ry/3 && y <= g.noseTipY {
				img.Pixels[y][x] = LightenColor(c, 1.2)
			}

			if isSkin && (x+y)%2 == 0 {
				img.Pixels[y][x] = DarkenColor(c, 0.82)
			}

			if isSkin {
				wrinkleY := g.cy - g.ry/3
				if y == wrinkleY && x >= g.cx-2 && x <= g.cx+2 {
					img.Pixels[y][x] = DarkenColor(c, 0.65)
				}
				wrinkleY2 := g.cy - g.ry/5
				if y == wrinkleY2 && x >= g.cx-1 && x <= g.cx+1 {
					img.Pixels[y][x] = DarkenColor(c, 0.7)
				}
			}

			if isSkin && y >= g.mouthY+2 && y <= g.mouthY+5 {
				dx := x - g.cx
				if dx >= -3 && dx <= 3 {
					img.Pixels[y][x] = DarkenColor(c, 0.65)
				}
			}

			if isSkin && y >= g.eyeY-4 && y < g.eyeY-1 {
				img.Pixels[y][x] = DarkenColor(c, 0.78)
			}

			// --- HAIR EFFECTS ---
			if isHair {
				if (x+y)%3 == 0 {
					img.Pixels[y][x] = DarkenColor(c, 0.7)
				} else if (x+y)%4 == 0 {
					img.Pixels[y][x] = LightenColor(c, 1.12)
				}
			}
		}
	}

	// Border darkening pass — darkens ALL non-background pixels near head edge
	bg := portraitBg
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.Pixels[y][x]
			if c == tcell.ColorDefault || c == bg {
				continue
			}
			nx := float64(x-g.cx) / float64(g.rx)
			ny := float64(y-g.cy) / float64(g.ry)
			dist := nx*nx + ny*ny
			if dist > 0.65 {
				f := 1.0 - (dist-0.65)*1.4
				if f < 0.35 {
					f = 0.35
				}
				img.Pixels[y][x] = DarkenColor(c, f)
			}
		}
	}
}

// isSkinTone checks if a color is a warm skin-like tone (R > G > B tendency).
func isSkinTone(c tcell.Color) bool {
	r, g, b := c.RGB()
	// Skin tones have R > B and R > 100 generally
	return r > 100 && r > b && r >= g-20
}

// colorClose reports whether two colors differ by less than tol in total RGB distance.
func colorClose(c, ref tcell.Color, tol int) bool {
	cr, cg, cb := c.RGB()
	rr, rg, rb := ref.RGB()
	dr := cr - rr
	if dr < 0 {
		dr = -dr
	}
	dg := cg - rg
	if dg < 0 {
		dg = -dg
	}
	db := cb - rb
	if db < 0 {
		db = -db
	}
	return int(dr+dg+db) < tol
}

// isHairColor checks if a pixel color matches the given hair color closely.
func isHairColor(c, hair tcell.Color) bool { return colorClose(c, hair, 80) }

func rngFromSeed(seed int64) *rng {
	return &rng{seed: seed}
}

type rng struct {
	seed int64
}

func (r *rng) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	r.seed = r.seed*6364136223846793005 + 1442695040888963407
	return int((uint64(r.seed) >> 33) % uint64(n))
}

// faceGeom returns head geometry proportions scaled to w,h.
type faceGeom struct {
	cx, cy   int // head center
	rx, ry   int // head radii
	eyeY     int // eye row
	eyeOff   int // horizontal distance from center to eye
	noseTipY int
	mouthY   int
	earTop   int
	earBot   int
	neckY    int
	torsoY   int
}

func computeFaceGeom(w, h int) faceGeom {
	// Head proportions: cy at 45% of height, radii at 42% of width/height.
	// These model a front-facing human head on a 16x24 or 20x24 canvas.
	cx := w / 2
	cy := h * 45 / 100
	rx := w * 42 / 100
	ry := h * 42 / 100
	if rx < 3 {
		rx = 3
	}
	if ry < 3 {
		ry = 3
	}

	eyeY := cy - ry/6
	eyeOff := rx * 5 / 8 // eyes at 5/8 of head radius from center
	if eyeOff < 1 {
		eyeOff = 1
	}
	noseTipY := cy + ry/5
	mouthY := cy + ry*5/10 // Mouth lower
	earTop := eyeY - ry/6
	earBot := noseTipY + ry/8
	neckY := cy + ry + 1
	torsoY := neckY

	return faceGeom{
		cx: cx, cy: cy, rx: rx, ry: ry,
		eyeY: eyeY, eyeOff: eyeOff,
		noseTipY: noseTipY, mouthY: mouthY,
		earTop: earTop, earBot: earBot,
		neckY: neckY, torsoY: torsoY,
	}
}

func inHead(x, y int, g faceGeom) bool {
	return inEllipse(x, y, g.cx, g.cy, g.rx, g.ry)
}

// inEllipse reports whether (x,y) is inside the ellipse centered at (cx,cy)
// with radii rx, ry (all integers, converted to float internally).
func inEllipse(x, y, cx, cy, rx, ry int) bool {
	dx := float64(x - cx)
	dy := float64(y - cy)
	return inEllipseF(dx, dy, float64(rx), float64(ry))
}

// inEllipseF is the float64 version of inEllipse.
func inEllipseF(dx, dy, rx, ry float64) bool {
	return (dx*dx)/(rx*rx)+(dy*dy)/(ry*ry) <= 1.0
}

func generateSkinLayer(w, h int, baseColor tcell.Color, bgColor tcell.Color) *PixelImage {
	img := NewPixelImage(w, h)
	g := computeFaceGeom(w, h)

	// Fill background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Pixels[y][x] = bgColor
		}
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !inHead(x, y, g) {
				continue
			}
			img.Pixels[y][x] = baseColor
		}
	}

	// Ears
	earW := g.rx / 3
	if earW < 1 {
		earW = 1
	}
	for y := g.earTop; y <= g.earBot; y++ {
		for side := -1; side <= 1; side += 2 {
			ex := g.cx + side*(g.rx+1)
			for dx := 0; dx < earW; dx++ {
				px := ex + side*dx
				if px >= 0 && px < w && y >= 0 && y < h {
					factor := 0.88 - float64(dx)*0.12
					if factor < 0.55 {
						factor = 0.55
					}
					img.Pixels[y][px] = DarkenColor(baseColor, factor)
				}
			}
			// Inner ear
			ix := ex + side
			if ix >= 0 && ix < w && y >= 0 && y < h {
				img.Pixels[y][ix] = DarkenColor(baseColor, 0.95)
			}
		}
	}

	return img
}

func generateEyeLayer(w, h int, irisColor tcell.Color, skinColor tcell.Color) *PixelImage {
	img := NewPixelImage(w, h)
	g := computeFaceGeom(w, h)

	white := tcell.NewRGBColor(240, 240, 240)
	black := tcell.NewRGBColor(15, 15, 15)
	irisHighlight := LightenColor(irisColor, 1.4)
	browColor := DarkenColor(irisColor, 0.3)
	eyeShadow := DarkenColor(skinColor, 0.8)

	eyeW := g.rx / 3
	if eyeW < 2 {
		eyeW = 2
	}

	for side := -1; side <= 1; side += 2 {
		ex := g.cx + side*g.eyeOff
		ey := g.eyeY

		// Eyebrow — arched, wider for larger faces
		browY := ey - 3
		if browY >= 0 && browY < h {
			for dx := -(eyeW + 1); dx <= eyeW+1; dx++ {
				px := ex + dx
				if px >= 0 && px < w {
					thickness := 1
					arch := 0
					edge := dx == -(eyeW+1) || dx == eyeW+1
					if edge {
						arch = 1
						thickness = 0
					}
					for t := 0; t <= thickness; t++ {
						by := browY - arch - t
						if by >= 0 {
							img.Pixels[by][px] = browColor
						}
					}
				}
			}
		}

		// Eye socket shadow (subtle)
		for dy := -2; dy <= 2; dy++ {
			for dx := -(eyeW + 1); dx <= eyeW+1; dx++ {
				px, py := ex+dx, ey+dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Pixels[py][px] = eyeShadow
				}
			}
		}

		// Sclera (eye white) — almond shape, wider
		// Top row: narrower
		for dx := -(eyeW - 1); dx <= eyeW-1; dx++ {
			px := ex + dx
			if px >= 0 && px < w && ey-1 >= 0 && ey-1 < h {
				img.Pixels[ey-1][px] = white
			}
		}
		// Middle row: full width
		for dx := -eyeW; dx <= eyeW; dx++ {
			px := ex + dx
			if px >= 0 && px < w && ey >= 0 && ey < h {
				img.Pixels[ey][px] = white
			}
		}
		// Bottom row: narrower
		for dx := -(eyeW - 1); dx <= eyeW-1; dx++ {
			px := ex + dx
			if px >= 0 && px < w && ey+1 >= 0 && ey+1 < h {
				img.Pixels[ey+1][px] = white
			}
		}

		// Iris — 2x2 block centered
		for dy := 0; dy <= 1; dy++ {
			for dx := -1; dx <= 0; dx++ {
				px, py := ex+dx, ey+dy
				if px >= 0 && px < w && py >= 0 && py < h {
					img.Pixels[py][px] = irisColor
				}
			}
		}

		// Pupil — center of iris
		if ex >= 0 && ex < w && ey >= 0 && ey < h {
			img.Pixels[ey][ex] = black
		}

		// Iris highlight — small bright dot
		highX := ex - 1
		highY := ey - 1
		if highX >= 0 && highX < w && highY >= 0 && highY < h {
			img.Pixels[highY][highX] = irisHighlight
		}

		// Upper eyelid crease
		lidY := ey - 2
		if lidY >= 0 && lidY < h {
			for dx := -eyeW; dx <= eyeW; dx++ {
				px := ex + dx
				if px >= 0 && px < w {
					img.Pixels[lidY][px] = DarkenColor(irisColor, 0.35)
				}
			}
		}

		// Lower lash line
		lashY := ey + 2
		if lashY >= 0 && lashY < h {
			for dx := -(eyeW - 1); dx <= eyeW-1; dx++ {
				px := ex + dx
				if px >= 0 && px < w {
					img.Pixels[lashY][px] = DarkenColor(irisColor, 0.2)
				}
			}
		}
	}

	return img
}

func generateNoseLayer(w, h int, skinColor tcell.Color) *PixelImage {
	img := NewPixelImage(w, h)
	g := computeFaceGeom(w, h)

	highlight := LightenColor(skinColor, 1.12)
	shadow := DarkenColor(skinColor, 0.82)
	deepShadow := DarkenColor(skinColor, 0.7)

	tipW := g.rx / 3
	if tipW < 1 {
		tipW = 1
	}

	// Nose bridge — highlight down center
	for y := g.eyeY + 1; y <= g.noseTipY; y++ {
		if y >= 0 && y < h && g.cx >= 0 && g.cx < w {
			img.Pixels[y][g.cx] = highlight
		}
	}

	// Nose bridge shadow — both sides, tapering
	for y := g.eyeY + 2; y <= g.noseTipY; y++ {
		progress := float64(y-g.eyeY) / float64(g.noseTipY-g.eyeY)
		sideOff := tipW + int(progress*float64(tipW))
		for side := -1; side <= 1; side += 2 {
			sx := g.cx + side*sideOff
			if sx >= 0 && sx < w && y >= 0 && y < h {
				img.Pixels[y][sx] = shadow
			}
		}
	}

	// Nose tip — rounded, wider
	for dx := -tipW; dx <= tipW; dx++ {
		px := g.cx + dx
		y := g.noseTipY
		if px >= 0 && px < w && y >= 0 && y < h {
			img.Pixels[y][px] = shadow
		}
	}
	// Tip highlight
	if g.cx >= 0 && g.cx < w && g.noseTipY >= 0 && g.noseTipY < h {
		img.Pixels[g.noseTipY][g.cx] = highlight
	}

	// Nostrils — two dark dots
	for side := -1; side <= 1; side += 2 {
		nx := g.cx + side*tipW
		ny := g.noseTipY + 1
		if nx >= 0 && nx < w && ny >= 0 && ny < h {
			img.Pixels[ny][nx] = deepShadow
		}
	}

	// Nose base shadow
	baseY := g.noseTipY + 1
	for dx := -tipW + 1; dx <= tipW-1; dx++ {
		px := g.cx + dx
		if px >= 0 && px < w && baseY >= 0 && baseY < h {
			img.Pixels[baseY][px] = DarkenColor(skinColor, 0.75)
		}
	}

	return img
}

func generateMouthLayer(w, h int, skinColor tcell.Color) *PixelImage {
	img := NewPixelImage(w, h)
	g := computeFaceGeom(w, h)

	lipColor := DarkenColor(skinColor, 0.78)
	darkLine := DarkenColor(skinColor, 0.55)
	lowerLip := LightenColor(skinColor, 1.05)
	upperLipDark := DarkenColor(skinColor, 0.72)

	mouthW := g.rx * 3 / 5
	if mouthW < 2 {
		mouthW = 2
	}

	my := g.mouthY

	// Mouth line (dark crease) — flat neutral expression
	for dx := -mouthW; dx <= mouthW; dx++ {
		px := g.cx + dx
		if px >= 0 && px < w && my >= 0 && my < h {
			img.Pixels[my][px] = darkLine
		}
	}

	// Upper lip — cupid's bow shape
	for dx := -mouthW + 1; dx <= mouthW-1; dx++ {
		px := g.cx + dx
		py := my - 1
		if px >= 0 && px < w && py >= 0 && py < h {
			// Cupid's bow: dip in center, peaks at sides
			absDx := dx
			if absDx < 0 {
				absDx = -absDx
			}
			bowOffset := 0
			if absDx <= mouthW/3 {
				bowOffset = 1 // center part dips down
			}
			if bowOffset == 0 {
				img.Pixels[py][px] = lipColor
			} else {
				img.Pixels[py][px] = upperLipDark
			}
		}
	}
	// Upper lip peak (philtrum columns)
	for side := -1; side <= 1; side += 2 {
		peakX := g.cx + side*(mouthW/3)
		py := my - 2
		if peakX >= 0 && peakX < w && py >= 0 && py < h {
			img.Pixels[py][peakX] = lipColor
		}
	}

	// Lower lip — fuller, rounded
	lipW := mouthW
	for dx := -lipW; dx <= lipW; dx++ {
		px := g.cx + dx
		py := my + 1
		if px >= 0 && px < w && py >= 0 && py < h {
			img.Pixels[py][px] = lowerLip
		}
	}
	// Lower lip bottom edge — shadow under lip
	for dx := -lipW + 1; dx <= lipW-1; dx++ {
		px := g.cx + dx
		py := my + 2
		if px >= 0 && px < w && py >= 0 && py < h {
			img.Pixels[py][px] = DarkenColor(skinColor, 0.85)
		}
	}
	// Lower lip highlight — center
	if mouthW > 1 {
		px := g.cx
		py := my + 1
		if px >= 0 && px < w && py >= 0 && py < h {
			img.Pixels[py][px] = LightenColor(skinColor, 1.1)
		}
	}

	return img
}

func generateHairLayer(w, h int, color tcell.Color, style int) *PixelImage {
	img := NewPixelImage(w, h)
	g := computeFaceGeom(w, h)

	darkHair := DarkenColor(color, 0.65)
	midHair := DarkenColor(color, 0.82)
	lightHair := LightenColor(color, 1.18)

	// strandColor returns a per-pixel strand-varied shade to break up flat fills.
	strandColor := func(x, y int) tcell.Color {
		v := (x*3 + y*2) % 7
		switch v {
		case 0:
			return darkHair
		case 1, 2:
			return lightHair
		case 3:
			return midHair
		default:
			return color
		}
	}

	// headTopY returns the topmost Y of the head ellipse at column x, or -1 if outside.
	headTopY := func(x int) int {
		dx := float64(x - g.cx)
		if dx < 0 {
			dx = -dx
		}
		rx := float64(g.rx)
		if dx > rx {
			return -1
		}
		t := 1.0 - (dx/rx)*(dx/rx)
		if t < 0 {
			t = 0
		}
		return int(float64(g.cy)-float64(g.ry)*math.Sqrt(t) + 0.5)
	}

	// setHairCol paints one vertical column of hair from topY downward, depth pixels.
	setHairCol := func(x, topY, depth int) {
		for dy := 0; dy < depth; dy++ {
			py := topY + dy
			if py < 0 || py >= h || x < 0 || x >= w {
				continue
			}
			img.Pixels[py][x] = strandColor(x, py)
		}
	}

	// sideHair paints a vertical strip on the sides of the head (for longer styles).
	sideHair := func(thickness int) {
		for side := -1; side <= 1; side += 2 {
			for y := g.cy - g.ry/5; y <= g.cy+g.ry/4; y++ {
				for d := 0; d < thickness; d++ {
					sx := g.cx + side*(g.rx-d)
					if sx >= 0 && sx < w && y >= 0 && y < h {
						img.Pixels[y][sx] = strandColor(sx, y)
					}
				}
			}
		}
	}

	switch style {
	case 0: // Buzzcut — razor-thin, 1-2px cap, heavy temple recession
		for x := 0; x < w; x++ {
			ty := headTopY(x)
			if ty < 0 {
				continue
			}
			dx := x - g.cx
			if dx < 0 {
				dx = -dx
			}
			templeF := 1.0 - float64(dx)/float64(g.rx+1)
			depth := int(2.0*templeF + 0.5)
			if depth < 1 {
				depth = 1
			}
			setHairCol(x, ty, depth)
		}

	case 1: // Medium — natural crown shape, 3-5px, side wisps
		for x := 0; x < w; x++ {
			ty := headTopY(x)
			if ty < 0 {
				continue
			}
			dx := x - g.cx
			if dx < 0 {
				dx = -dx
			}
			templeF := 1.0 - float64(dx)/float64(g.rx+1)*0.7
			depth := int(5.0*templeF + 0.5)
			if depth < 1 {
				depth = 1
			}
			setHairCol(x, ty, depth)
		}
		sideHair(2)

	case 2: // Spiky — alternating taller spikes on crown only
		for x := 0; x < w; x++ {
			ty := headTopY(x)
			if ty < 0 {
				continue
			}
			dx := x - g.cx
			if dx < 0 {
				dx = -dx
			}
			if dx > g.rx*3/5 {
				continue
			}
			spikeExtra := 0
			if x%2 == 0 {
				spikeExtra = 4 - dx/2
				if spikeExtra < 0 {
					spikeExtra = 0
				}
			}
			setHairCol(x, ty-spikeExtra, 3+spikeExtra)
		}

	case 3: // Afro — rounded puff extending beyond head ellipse
		afroRX := float64(g.rx) * 1.35
		afroRY := float64(g.ry) * 0.85
		afroCY := float64(g.cy) - float64(g.ry)*0.15
		for y := int(afroCY-afroRY) - 1; y <= g.cy; y++ {
			for x := g.cx - int(afroRX) - 1; x <= g.cx+int(afroRX)+1; x++ {
				if x < 0 || x >= w || y < 0 || y >= h {
					continue
				}
				dx := float64(x - g.cx)
				dy := float64(y) - afroCY
				if inEllipseF(dx, dy, afroRX, afroRY) {
					rx2, ry2 := float64(g.rx-4), float64(g.ry-4)
					if rx2 < 1 {
						rx2 = 1
					}
					if ry2 < 1 {
						ry2 = 1
					}
					insideHead := inEllipseF(dx, float64(y-g.cy), float64(g.rx), float64(g.ry))
					insideInner := inEllipseF(dx, float64(y-g.cy), rx2, ry2)
					if !insideHead || !insideInner {
						img.Pixels[y][x] = strandColor(x, y)
					}
				}
			}
		}

	case 4: // Side part — full coverage with visible dark part groove
		for x := 0; x < w; x++ {
			ty := headTopY(x)
			if ty < 0 {
				continue
			}
			dx := x - g.cx
			if dx < 0 {
				dx = -dx
			}
			templeF := 1.0 - float64(dx)/float64(g.rx+1)*0.5
			depth := int(6.0*templeF + 0.5)
			if depth < 1 {
				depth = 1
			}
			setHairCol(x, ty, depth)
		}
		partX := g.cx + g.rx/4
		for dy := 0; dy < 5; dy++ {
			ty := headTopY(partX)
			if ty >= 0 {
				py := ty + dy
				if py >= 0 && py < h && partX >= 0 && partX < w {
					img.Pixels[py][partX] = darkHair
				}
			}
		}

	case 5: // Curly — medium depth with curl texture alternation
		for x := 0; x < w; x++ {
			ty := headTopY(x)
			if ty < 0 {
				continue
			}
			dx := x - g.cx
			if dx < 0 {
				dx = -dx
			}
			templeF := 1.0 - float64(dx)/float64(g.rx+1)*0.6
			depth := int(5.0*templeF + 0.5)
			if depth < 1 {
				depth = 1
			}
			for dy := 0; dy < depth; dy++ {
				py := ty + dy
				if py < 0 || py >= h || x < 0 || x >= w {
					continue
				}
				switch (x + py) % 3 {
				case 0:
					img.Pixels[py][x] = lightHair
				case 1:
					img.Pixels[py][x] = darkHair
				default:
					img.Pixels[py][x] = color
				}
			}
		}
		sideHair(2)

	case 6: // Ponytail — top hair + narrow strip falling on right
		for x := 0; x < w; x++ {
			ty := headTopY(x)
			if ty < 0 {
				continue
			}
			dx := x - g.cx
			if dx < 0 {
				dx = -dx
			}
			templeF := 1.0 - float64(dx)/float64(g.rx+1)*0.55
			depth := int(5.0*templeF + 0.5)
			if depth < 1 {
				depth = 1
			}
			setHairCol(x, ty, depth)
		}
		tailX := g.cx + g.rx
		for y := g.cy - g.ry/3; y <= g.cy+g.ry; y++ {
			for d := 0; d < 2; d++ {
				tx := tailX + d
				if tx >= 0 && tx < w && y >= 0 && y < h {
					img.Pixels[y][tx] = strandColor(tx, y)
				}
			}
		}

	case 7: // Undercut — crown-only, stark shaved sides
		crownW := g.rx * 3 / 5
		for x := g.cx - crownW; x <= g.cx+crownW; x++ {
			ty := headTopY(x)
			if ty < 0 {
				continue
			}
			dx := x - g.cx
			if dx < 0 {
				dx = -dx
			}
			templeF := 1.0 - float64(dx)/float64(crownW+1)*0.3
			depth := int(4.0*templeF + 0.5)
			if depth < 1 {
				depth = 1
			}
			setHairCol(x, ty, depth)
		}
		for side := -1; side <= 1; side += 2 {
			ex := g.cx + side*(crownW+1)
			ty := headTopY(ex)
			if ty >= 0 && ex >= 0 && ex < w {
				for dy := 0; dy < 3; dy++ {
					py := ty + dy
					if py >= 0 && py < h {
						img.Pixels[py][ex] = darkHair
					}
				}
			}
		}

	default: // Short cropped — thin curved shell, heavy recession
		for x := 0; x < w; x++ {
			ty := headTopY(x)
			if ty < 0 {
				continue
			}
			dx := x - g.cx
			if dx < 0 {
				dx = -dx
			}
			templeF := 1.0 - float64(dx)/float64(g.rx+1)*0.8
			depth := int(3.0*templeF + 0.5)
			if depth < 1 {
				depth = 1
			}
			setHairCol(x, ty, depth)
		}
	}

	return img
}

// GenerateAlienPixelsImage converts an AlienPixels grid to a PixelImage,
// rendering body and weapon layers in distinct colors.
func GenerateAlienPixelsImage(ap data.AlienPixels, fgColor, bgColor tcell.Color) *PixelImage {
	bR, bG, bB := fgColor.RGB()
	wR, wG, wB := data.AlienWeaponColor()
	weaponColor := tcell.NewRGBColor(wR, wG, wB)

	lightR := clampColor(bR + 45)
	lightG := clampColor(bG + 45)
	lightB := clampColor(bB + 45)
	darkR := clampColor(bR - 55)
	darkG := clampColor(bG - 55)
	darkB := clampColor(bB - 55)
	accentR := clampColor(bR + 70)
	accentG := clampColor(bG + 30)
	accentB := clampColor(bB + 70)

	lightColor := tcell.NewRGBColor(lightR, lightG, lightB)
	darkColor := tcell.NewRGBColor(darkR, darkG, darkB)
	accentColor := tcell.NewRGBColor(accentR, accentG, accentB)
	interiorColor := tcell.NewRGBColor(clampColor(bR+20), clampColor(bG+25), clampColor(bB+15))
	bellyColor := tcell.NewRGBColor(clampColor(bR+40), clampColor(bG+30), clampColor(bB+50))
	textureColor := tcell.NewRGBColor(clampColor(bR-12), clampColor(bG+18), clampColor(bB-8))
	mouthColor := tcell.NewRGBColor(clampColor(bR-40), clampColor(bG-35), clampColor(bB-30))
	pupilColor := tcell.NewRGBColor(60, 60, 70)

	img := NewPixelImage(data.SpriteW, data.SpriteH)
	for y := 0; y < data.SpriteH; y++ {
		for x := 0; x < data.SpriteW; x++ {
			switch {
			case ap.Weapon[y][x] && ap.Highlight[y][x]:
				img.Pixels[y][x] = lightColor
			case ap.Weapon[y][x] && ap.Shadow[y][x]:
				img.Pixels[y][x] = darkColor
			case ap.Weapon[y][x] && ap.Accent[y][x]:
				img.Pixels[y][x] = accentColor
			case ap.Weapon[y][x]:
				img.Pixels[y][x] = weaponColor
			case ap.Eyes[y][x] && ap.Accent[y][x]:
				img.Pixels[y][x] = pupilColor
			case ap.Eyes[y][x] && ap.Shadow[y][x]:
				img.Pixels[y][x] = darkColor
			case ap.Eyes[y][x] && ap.Highlight[y][x]:
				img.Pixels[y][x] = lightColor
			case ap.Eyes[y][x]:
				img.Pixels[y][x] = tcell.NewRGBColor(255, 255, 255)
			case ap.Mouth[y][x]:
				img.Pixels[y][x] = mouthColor
			case ap.Highlight[y][x]:
				img.Pixels[y][x] = lightColor
			case ap.Shadow[y][x]:
				img.Pixels[y][x] = darkColor
			case ap.Accent[y][x]:
				img.Pixels[y][x] = accentColor
			case ap.Interior[y][x]:
				img.Pixels[y][x] = interiorColor
			case ap.Belly[y][x]:
				img.Pixels[y][x] = bellyColor
			case ap.Texture[y][x]:
				img.Pixels[y][x] = textureColor
			case ap.Body[y][x]:
				img.Pixels[y][x] = fgColor
			default:
				img.Pixels[y][x] = bgColor
			}
		}
	}
	return img
}

func clampColor(v int32) int32 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// GenerateAlienSpriteFromSeed creates a PixelImage from a seeded sprite assembly,
// using the alien's morphology to select trait-matched visual templates.
// Results are cached per (seed, bgColor, morphology) since the sprite never changes.
func GenerateAlienSpriteFromSeed(seed int64, m *data.Morphology, bgColor tcell.Color) *PixelImage {
	br, bg, bb := bgColor.RGB()
	key := spriteCacheKey{seed: seed, bgR: br, bgG: bg, bgB: bb, mk: morphKey(m)}
	if v, ok := alienSpriteCache.Load(key); ok {
		return v.(*PixelImage)
	}
	ap := data.GenerateAlienPixels(seed, m)
	r, g, b := data.AlienColorFromSeed(seed)
	fgColor := tcell.NewRGBColor(r, g, b)
	img := GenerateAlienPixelsImage(ap, fgColor, bgColor)
	alienSpriteCache.Store(key, img)
	return img
}
