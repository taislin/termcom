package engine

import (
	"sync"

	"github.com/civ13/termcom/internal/data"
	"github.com/gdamore/tcell/v3"
)

var alienSpriteCache sync.Map

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

type PortraitLayer int

const (
	LayerSkin PortraitLayer = iota
	LayerEyes
	LayerNose
	LayerMouth
	LayerHair
	LayerMarkings
	LayerHelmet
	LayerArmour
	LayerDecal
	LayerCount
)

type PortraitSpec struct {
	Width         int
	Height        int
	SkinColor     tcell.Color
	EyeColor      tcell.Color
	HairColor     tcell.Color
	MarkingsColor tcell.Color // tcell.ColorDefault = none
	HelmetColor   tcell.Color // tcell.ColorDefault = none
	ArmourColor   tcell.Color // tcell.ColorDefault = none
	DecalColor    tcell.Color
	Seed          int64
}

// MakeSoldierPortrait builds a portrait from a soldier's name and armor string.
func MakeSoldierPortrait(name, armor string, w, h int) *PixelImage {
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

	markingsColors := []tcell.Color{
		tcell.ColorDefault,
		tcell.ColorDefault,
		tcell.ColorDefault,
		tcell.NewRGBColor(200, 40, 40),
		tcell.NewRGBColor(40, 120, 200),
		tcell.NewRGBColor(220, 180, 40),
	}
	markingsColor := markingsColors[(nameSeed/11)%int64(len(markingsColors))]

	var armourColor tcell.Color = tcell.ColorDefault
	var helmetColor tcell.Color = tcell.ColorDefault
	if armor != "" && armor != "none" {
		if armor == "personal_armor" {
			armourColor = tcell.NewRGBColor(50, 120, 50)
			helmetColor = tcell.NewRGBColor(50, 120, 50)
		} else if armor == "power_suit" {
			armourColor = tcell.NewRGBColor(120, 120, 120)
			helmetColor = tcell.NewRGBColor(120, 120, 120)
		} else {
			armourColor = tcell.NewRGBColor(80, 80, 150)
			helmetColor = tcell.NewRGBColor(80, 80, 150)
		}
	}

	return GenerateSoldierPortrait(PortraitSpec{
		Width:         w,
		Height:        h,
		SkinColor:     skinColor,
		EyeColor:      eyeColor,
		HairColor:     hairColor,
		MarkingsColor: markingsColor,
		HelmetColor:   helmetColor,
		ArmourColor:   armourColor,
		DecalColor:    tcell.NewRGBColor(255, 215, 0),
		Seed:          nameSeed,
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

	bgColor := tcell.NewRGBColor(20, 20, 28)
	skin := generateSkinLayer(w, h, spec.SkinColor, bgColor)
	eyes := generateEyeLayer(w, h, spec.EyeColor)
	nose := generateNoseLayer(w, h, spec.SkinColor)
	mouth := generateMouthLayer(w, h, spec.SkinColor)
	hair := generateHairLayer(w, h, spec.HairColor, rng.Intn(8))

	res := skin
	res = CompositeImages(res, eyes)
	res = CompositeImages(res, nose)
	res = CompositeImages(res, mouth)
	res = CompositeImages(res, hair)

	if spec.HelmetColor != tcell.ColorDefault {
		helmet := generateHelmetLayer(w, h, spec.HelmetColor)
		res = CompositeImages(res, helmet)
	}

	return res
}

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
	return int((r.seed >> 33) % int64(n))
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
	eyeOff := rx * 5 / 8
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
	dx := float64(x - g.cx)
	dy := float64(y - g.cy)
	return (dx*dx)/(float64(g.rx*g.rx))+(dy*dy)/(float64(g.ry*g.ry)) <= 1.0
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
	// ... (rest of function)

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
					img.Pixels[y][px] = DarkenColor(baseColor, 0.8)
				}
			}
			// Inner ear
			ix := ex + side
			if ix >= 0 && ix < w && y >= 0 && y < h {
				img.Pixels[y][ix] = DarkenColor(baseColor, 0.6)
			}
		}
	}

	return img
}

func generateEyeLayer(w, h int, irisColor tcell.Color) *PixelImage {
	img := NewPixelImage(w, h)
	g := computeFaceGeom(w, h)

	white := tcell.NewRGBColor(240, 240, 240)
	black := tcell.NewRGBColor(15, 15, 15)
	irisHighlight := LightenColor(irisColor, 1.4)
	browColor := DarkenColor(irisColor, 0.3)
	if browColor == tcell.ColorDefault {
		browColor = tcell.NewRGBColor(60, 50, 40)
	}
	eyeShadow := DarkenColor(irisColor, 0.12)

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

	// Mouth line (dark crease) — curves up slightly at corners
	for dx := -mouthW; dx <= mouthW; dx++ {
		px := g.cx + dx
		lineY := my
		if dx == -mouthW || dx == mouthW {
			lineY = my - 1 // corners lift slightly
		}
		if px >= 0 && px < w && lineY >= 0 && lineY < h {
			img.Pixels[lineY][px] = darkLine
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

	hairTop := g.cy - g.ry - 2
	if hairTop < 0 {
		hairTop = 0
	}
	darkHair := DarkenColor(color, 0.7)
	lightHair := LightenColor(color, 1.15)

	switch style {
	case 0: // Buzzcut
		for y := hairTop; y <= g.cy-g.ry/3; y++ {
			for x := g.cx - g.rx - 1; x <= g.cx+g.rx+1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					if inHead(x, y, g) && y < g.cy {
						img.Pixels[y][x] = color
					}
				}
			}
		}

	case 1: // Long hair / sides
		for y := hairTop; y <= g.cy+g.ry/2; y++ {
			for x := g.cx - g.rx - 2; x <= g.cx+g.rx+2; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					dx := float64(x - g.cx)
					dy := float64(y - g.cy)
					val := (dx*dx)/(float64(g.rx*g.rx)) + (dy*dy)/(float64(g.ry*g.ry))
					onTop := val <= 1.3 && val >= 0.85 && y < g.cy
					onSide := (x < g.cx-g.rx+2 || x > g.cx+g.rx-2) && y >= g.cy && y <= g.cy+g.ry/2
					if onTop || onSide {
						img.Pixels[y][x] = color
						if onSide && (x == g.cx-g.rx-1 || x == g.cx+g.rx+1 || y == g.cy+g.ry/2) {
							img.Pixels[y][x] = darkHair
						}
					}
				}
			}
		}

	case 2: // Spiky / mohawk
		for y := hairTop - 3; y <= g.cy-g.ry/3; y++ {
			for x := g.cx - g.rx/2; x <= g.cx+g.rx/2; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					if y < g.cy-g.ry+2 {
						if (x-g.cx)%2 == 0 {
							img.Pixels[y][x] = color
						}
					} else if inHead(x, y, g) && y < g.cy {
						img.Pixels[y][x] = color
					}
				}
			}
		}

	case 3: // Afro
		afroR := float64(g.rx) * 1.4
		afroCY := float64(g.cy - g.ry/2)
		for y := hairTop - 2; y <= g.cy-g.ry/3; y++ {
			for x := g.cx - int(afroR) - 1; x <= g.cx+int(afroR)+1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					dx := float64(x - g.cx)
					dy := float64(y) - afroCY
					if dx*dx+dy*dy <= afroR*afroR && float64(y) < afroCY+afroR*0.5 {
						img.Pixels[y][x] = color
						if dx*dx+dy*dy > afroR*afroR*0.7 {
							img.Pixels[y][x] = darkHair
						}
					}
				}
			}
		}

	case 4: // Parted / slicked
		for y := hairTop; y <= g.cy-g.ry/4; y++ {
			for x := g.cx - g.rx - 1; x <= g.cx+g.rx+1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					if inHead(x, y, g) && y < g.cy {
						img.Pixels[y][x] = color
					}
				}
			}
		}
		// Part line
		partX := g.cx + g.rx/3
		for y := hairTop; y <= g.cy-g.ry/2; y++ {
			if partX >= 0 && partX < w && y >= 0 && y < h {
				img.Pixels[y][partX] = darkHair
			}
		}

	case 5: // Curly — rounded tufts
		afroR := float64(g.rx) * 1.2
		afroCY := float64(g.cy - g.ry/2)
		for y := hairTop - 1; y <= g.cy-g.ry/3; y++ {
			for x := g.cx - int(afroR) - 1; x <= g.cx+int(afroR)+1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					dx := float64(x - g.cx)
					dy := float64(y) - afroCY
					dist := dx*dx + dy*dy
					if dist <= afroR*afroR && float64(y) < afroCY+afroR*0.4 {
						// Create curl pattern with alternating light/dark
						curl := (x+y)%3 == 0
						if curl {
							img.Pixels[y][x] = lightHair
						} else if dist > afroR*afroR*0.6 {
							img.Pixels[y][x] = darkHair
						} else {
							img.Pixels[y][x] = color
						}
					}
				}
			}
		}

	case 6: // Ponytail — hair on top + tail down one side
		// Top part
		for y := hairTop; y <= g.cy-g.ry/3; y++ {
			for x := g.cx - g.rx - 1; x <= g.cx+g.rx+1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					if inHead(x, y, g) && y < g.cy {
						img.Pixels[y][x] = color
					}
				}
			}
		}
		// Tail — hangs down from right side
		tailX := g.cx + g.rx + 1
		for y := g.cy - g.ry/2; y <= g.cy+g.ry; y++ {
			if tailX >= 0 && tailX < w && y >= 0 && y < h {
				img.Pixels[y][tailX] = color
				if tailX+1 < w {
					img.Pixels[y][tailX+1] = darkHair
				}
			}
		}

	case 7: // Shaved sides / undercut — short on top, skin on sides
		// Short hair on top only
		for y := hairTop; y <= g.cy-g.ry/2; y++ {
			for x := g.cx - g.rx/2; x <= g.cx+g.rx/2; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					if inHead(x, y, g) && y < g.cy {
						img.Pixels[y][x] = color
					}
				}
			}
		}
		// Fade gradient on sides
		for y := g.cy - g.ry/2; y <= g.cy-g.ry/4; y++ {
			for side := -1; side <= 1; side += 2 {
				fadeX := g.cx + side*(g.rx/2+1)
				if fadeX >= 0 && fadeX < w && y >= 0 && y < h {
					img.Pixels[y][fadeX] = darkHair
				}
			}
		}

	default: // Short cropped
		for y := hairTop; y <= g.cy-g.ry/3; y++ {
			for x := g.cx - g.rx; x <= g.cx+g.rx; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					if inHead(x, y, g) && y < g.cy {
						img.Pixels[y][x] = color
					}
				}
			}
		}
	}

	return img
}

func generateHelmetLayer(w, h int, color tcell.Color) *PixelImage {
	img := NewPixelImage(w, h)
	g := computeFaceGeom(w, h)

	dark := DarkenColor(color, 0.7)
	light := LightenColor(color, 1.2)

	helmetBottom := g.eyeY - 2

	for y := g.cy - g.ry - 3; y <= helmetBottom; y++ {
		for x := g.cx - g.rx - 2; x <= g.cx+g.rx+2; x++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				dx := float64(x - g.cx)
				dy := float64(y - g.cy)
				val := (dx*dx)/(float64(g.rx*g.rx)) + (dy*dy)/(float64(g.ry*g.ry))
				if val <= 1.5 && y < g.cy {
					img.Pixels[y][x] = color
				}
			}
		}
	}

	// Rim
	for x := g.cx - g.rx - 1; x <= g.cx+g.rx+1; x++ {
		if x >= 0 && x < w && helmetBottom >= 0 && helmetBottom < h {
			img.Pixels[helmetBottom][x] = dark
		}
		if x >= 0 && x < w && helmetBottom-1 >= 0 && helmetBottom-1 < h {
			img.Pixels[helmetBottom-1][x] = light
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

	img := NewPixelImage(20, 24)
	for y := 0; y < 24; y++ {
		for x := 0; x < 20; x++ {
			switch {
			case ap.Weapon[y][x]:
				img.Pixels[y][x] = weaponColor
			case ap.Eyes[y][x]:
				img.Pixels[y][x] = tcell.NewRGBColor(255, 255, 255)
			case ap.Highlight[y][x]:
				img.Pixels[y][x] = lightColor
			case ap.Shadow[y][x]:
				img.Pixels[y][x] = darkColor
			case ap.Accent[y][x]:
				img.Pixels[y][x] = accentColor
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
