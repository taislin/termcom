package engine

import (
	"math/rand"

	"github.com/civ13/termcom/internal/data"
	"github.com/gdamore/tcell/v3"
)

type PortraitLayer int

const (
	LayerSkin PortraitLayer = iota
	LayerEyes
	LayerHair
	LayerHelmet
	LayerArmour
	LayerDecal
	LayerCount
)

type PortraitSpec struct {
	Width       int
	Height      int
	SkinColor   tcell.Color
	EyeColor    tcell.Color
	HairColor   tcell.Color
	HelmetColor tcell.Color // tcell.ColorDefault = none
	ArmourColor tcell.Color // tcell.ColorDefault = none
	DecalColor  tcell.Color
	Seed        int64
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
	}
	skinColor := skinColors[nameSeed%int64(len(skinColors))]

	eyeColors := []tcell.Color{
		tcell.NewRGBColor(50, 100, 200),
		tcell.NewRGBColor(40, 150, 50),
		tcell.NewRGBColor(100, 60, 30),
	}
	eyeColor := eyeColors[(nameSeed/3)%int64(len(eyeColors))]

	hairColors := []tcell.Color{
		tcell.NewRGBColor(10, 10, 10),
		tcell.NewRGBColor(120, 60, 20),
		tcell.NewRGBColor(230, 200, 50),
		tcell.NewRGBColor(200, 80, 20),
	}
	hairColor := hairColors[(nameSeed/7)%int64(len(hairColors))]

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
		Width:       w,
		Height:      h,
		SkinColor:   skinColor,
		EyeColor:    eyeColor,
		HairColor:   hairColor,
		HelmetColor: helmetColor,
		ArmourColor: armourColor,
		DecalColor:  tcell.NewRGBColor(255, 215, 0),
		Seed:        nameSeed,
	})
}

// GenerateSoldierPortrait generates a procedural soldier portrait with stacked layers.
func GenerateSoldierPortrait(spec PortraitSpec) *PixelImage {
	w, h := spec.Width, spec.Height
	if w <= 0 {
		w = 20
	}
	if h <= 0 {
		h = 40
	}

	rng := rand.New(rand.NewSource(spec.Seed))

	skin := generateSkinLayer(w, h, spec.SkinColor)
	eyes := generateEyeLayer(w, h, spec.EyeColor)
	hair := generateHairLayer(w, h, spec.HairColor, rng.Intn(4))

	res := skin
	res = CompositeImages(res, eyes)
	res = CompositeImages(res, hair)

	if spec.HelmetColor != tcell.ColorDefault {
		helmet := generateHelmetLayer(w, h, spec.HelmetColor)
		res = CompositeImages(res, helmet)
	}

	if spec.ArmourColor != tcell.ColorDefault {
		armour := generateArmourLayer(w, h, spec.ArmourColor)
		res = CompositeImages(res, armour)
	}

	if spec.DecalColor != tcell.ColorDefault {
		decal := generateDecalLayer(w, h, spec.DecalColor, rng.Intn(5)+1)
		res = CompositeImages(res, decal)
	}

	return res
}

func generateSkinLayer(w, h int, color tcell.Color) *PixelImage {
	img := NewPixelImage(w, h)
	cx, cy := w/2, h/3 // Head center
	rx, ry := w/3, h/5 // Head radii

	// 1. Draw head (oval)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			val := (dx*dx)/(float64(rx*rx)) + (dy*dy)/(float64(ry*ry))
			if val <= 1.0 {
				img.Pixels[y][x] = color
			}
		}
	}

	// 2. Draw neck
	neckW := w / 5
	neckH := h / 10
	neckY := cy + ry
	for y := neckY; y < neckY+neckH && y < h; y++ {
		for x := cx - neckW/2; x <= cx+neckW/2; x++ {
			if x >= 0 && x < w {
				img.Pixels[y][x] = color
			}
		}
	}

	// 3. Draw shoulders/torso
	torsoY := neckY + neckH
	for y := torsoY; y < h; y++ {
		// Shoulders slope outwards
		slope := (y - torsoY) * (w / 10)
		left := cx - w/4 - slope
		right := cx + w/4 + slope
		if left < 1 {
			left = 1
		}
		if right >= w-1 {
			right = w - 2
		}
		for x := left; x <= right; x++ {
			img.Pixels[y][x] = color
		}
	}

	// 4. Add rim shading (darken edge pixels)
	shaded := NewPixelImage(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if img.Pixels[y][x] == tcell.ColorDefault {
				continue
			}

			// Check if it's an edge pixel (has any transparent neighbor)
			isEdge := false
			dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
			for _, d := range dirs {
				nx, ny := x+d[0], y+d[1]
				if nx < 0 || nx >= w || ny < 0 || ny >= h || img.Pixels[ny][nx] == tcell.ColorDefault {
					isEdge = true
					break
				}
			}

			if isEdge {
				shaded.Pixels[y][x] = DarkenColor(color, 0.7)
			} else {
				shaded.Pixels[y][x] = color
			}
		}
	}

	return shaded
}

func generateEyeLayer(w, h int, color tcell.Color) *PixelImage {
	img := NewPixelImage(w, h)
	cx, cy := w/2, h/3
	rx, ry := w/3, h/5

	// Put eyes at about cy - ry/5
	eyeY := cy - ry/4
	eyeOffset := rx / 2 // distance from center line

	// Left eye
	lx := cx - eyeOffset
	// Right eye
	rxOffset := cx + eyeOffset

	// Draw eye sockets (white background if possible, or just the eye color)
	for dy := -1; dy <= 1; dy++ {
		for dx := -2; dx <= 2; dx++ {
			lyy, lxx := eyeY+dy, lx+dx
			ryy, rxx := eyeY+dy, rxOffset+dx

			if lxx >= 0 && lxx < w && lyy >= 0 && lyy < h {
				img.Pixels[lyy][lxx] = tcell.ColorWhite
			}
			if rxx >= 0 && rxx < w && ryy >= 0 && ryy < h {
				img.Pixels[ryy][rxx] = tcell.ColorWhite
			}
		}
	}

	// Draw pupils
	if lx >= 0 && lx < w && eyeY >= 0 && eyeY < h {
		img.Pixels[eyeY][lx] = color
	}
	if rxOffset >= 0 && rxOffset < w && eyeY >= 0 && eyeY < h {
		img.Pixels[eyeY][rxOffset] = color
	}

	return img
}

func generateHairLayer(w, h int, color tcell.Color, style int) *PixelImage {
	img := NewPixelImage(w, h)
	cx, cy := w/2, h/3
	rx, ry := w/3, h/5

	hairTop := cy - ry - 2
	hairBottom := cy - ry/3

	switch style {
	case 0: // Buzzcut / short hair
		for y := hairTop; y <= hairBottom; y++ {
			for x := cx - rx - 1; x <= cx+rx+1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					// Check distance from head center to fit shape
					dx := float64(x - cx)
					dy := float64(y - cy)
					val := (dx*dx)/(float64(rx*rx)) + (dy*dy)/(float64(ry*ry))
					if val <= 1.2 && val >= 0.8 && y < cy {
						img.Pixels[y][x] = color
					}
				}
			}
		}
	case 1: // Long hair / sides
		for y := hairTop; y <= cy+ry; y++ {
			for x := cx - rx - 2; x <= cx+rx+2; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					dx := float64(x - cx)
					dy := float64(y - cy)
					val := (dx*dx)/(float64(rx*rx)) + (dy*dy)/(float64(ry*ry))
					// Hair on top and cascading down the sides
					if (val <= 1.3 && val >= 0.9 && y < cy) || (x < cx-rx+2 || x > cx+rx-2) && y >= cy && y <= cy+ry/2 {
						img.Pixels[y][x] = color
					}
				}
			}
		}
	case 2: // Spiky / mohawk
		for y := hairTop - 3; y <= hairBottom; y++ {
			for x := cx - rx/2; x <= cx+rx/2; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					// Spiky center
					if y < cy-ry+2 {
						if (x-cx)%2 == 0 {
							img.Pixels[y][x] = color
						}
					} else {
						img.Pixels[y][x] = color
					}
				}
			}
		}
	default: // Afroglow / curly round hair
		for y := hairTop - 2; y <= hairBottom; y++ {
			for x := cx - rx - 3; x <= cx+rx+3; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					dx := float64(x - cx)
					dy := float64(y - (cy - ry/2))
					dist := (dx*dx) + (dy*dy)
					rad := float64(rx) * 1.2
					if dist <= rad*rad && y < cy {
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
	cx, cy := w/2, h/3
	rx, ry := w/3, h/5

	// Helmet covers top of head, ears, and forehead down to just above eyes
	helmetBottom := cy - ry/5

	for y := cy - ry - 4; y <= helmetBottom; y++ {
		for x := cx - rx - 2; x <= cx+rx+2; x++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				dx := float64(x - cx)
				dy := float64(y - cy)
				val := (dx*dx)/(float64(rx*rx)) + (dy*dy)/(float64(ry*ry))
				if val <= 1.4 && y < cy {
					img.Pixels[y][x] = color
				}
			}
		}
	}

	// Add visor strip
	visorY := cy - ry/3
	for x := cx - rx + 1; x <= cx+rx-1; x++ {
		if x >= 0 && x < w && visorY >= 0 && visorY < h {
			img.Pixels[visorY][x] = DarkenColor(color, 0.4)
		}
	}

	return img
}

func generateArmourLayer(w, h int, color tcell.Color) *PixelImage {
	img := NewPixelImage(w, h)
	cx, cy := w/2, h/3
	_, ry := w/3, h/5

	// Armour sits on shoulders and chest starting at neck bottom
	torsoY := cy + ry + h/10


	for y := torsoY; y < h; y++ {
		slope := (y - torsoY) * (w / 10)
		left := cx - w/4 - slope
		right := cx + w/4 + slope
		if left < 0 {
			left = 0
		}
		if right >= w {
			right = w - 1
		}

		for x := left; x <= right; x++ {
			// Heavy shoulder pads
			if y < torsoY+4 && (x < left+3 || x > right-3) {
				img.Pixels[y][x] = LightenColor(color, 1.2)
			} else {
				img.Pixels[y][x] = color
			}
		}
	}

	// Add highlights
	for y := torsoY; y < h; y++ {
		slope := (y - torsoY) * (w / 10)
		left := cx - w/4 - slope
		right := cx + w/4 + slope
		if left >= 0 && left < w {
			img.Pixels[y][left] = LightenColor(color, 1.4)
		}
		if right >= 0 && right < w {
			img.Pixels[y][right] = LightenColor(color, 1.4)
		}
	}

	return img
}

func generateDecalLayer(w, h int, color tcell.Color, rank int) *PixelImage {
	img := NewPixelImage(w, h)
	cx, cy := w/2, h/3
	_, ry := w/3, h/5

	// Put rank pips on the chest area
	decalY := cy + ry + h/5

	decalX := cx

	// 1 to 5 pips
	for i := 0; i < rank; i++ {
		offset := (i - rank/2) * 2
		px := decalX + offset
		if px >= 0 && px < w && decalY >= 0 && decalY < h {
			img.Pixels[decalY][px] = color
			// Make it 2x2 for visibility
			if decalY+1 < h {
				img.Pixels[decalY+1][px] = color
			}
		}
	}

	return img
}

// GenerateAlienPortrait upscales a StyledPortrait text block to a sub-cell half-block PixelImage.
// If scale is 2, it converts a 7x7 StyledPortrait to 14x14 pixels.
// Let's use rune-density lookup table to fill scale x scale pixels.
func GenerateAlienPortrait(sp data.StyledPortrait, scale int) *PixelImage {
	if len(sp.Lines) == 0 {
		return NewPixelImage(14, 14)
	}

	linesH := len(sp.Lines)
	linesW := len(sp.Lines[0].Content)

	imgW := linesW * scale
	imgH := linesH * scale
	img := NewPixelImage(imgW, imgH)

	// Rune-density tables. A rune maps to a pattern of size scale x scale.
	// We'll support scale = 2 or scale = 4. Let's make it general.
	for rY, line := range sp.Lines {
		runes := []rune(line.Content)
		cVal := tcell.NewRGBColor(line.Color[0], line.Color[1], line.Color[2])

		for rX, rn := range runes {
			// Get standard density pattern
			// 0 = empty, 1 = sparse, 2 = dense, 3 = full fill
			density := 0
			switch rn {
			case ' ', 0:
				density = 0
			case '.', '·', '°', '*':
				density = 1
			case '|', '-', '/', '\\', '+', '¤', '~', '†', 'o':
				density = 2
			default:
				density = 3
			}

			// Fill a scale x scale block
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					pixelY := rY*scale + dy
					pixelX := rX*scale + dx

					if pixelY < imgH && pixelX < imgW {
						if density == 3 {
							img.Pixels[pixelY][pixelX] = cVal
						} else if density == 2 {
							// checkerboard / hatch pattern
							if (dy+dx)%2 == 0 {
								img.Pixels[pixelY][pixelX] = cVal
							}
						} else if density == 1 {
							// single pixel
							if dy == scale/2 && dx == scale/2 {
								img.Pixels[pixelY][pixelX] = cVal
							}
						}
					}
				}
			}
		}
	}

	return img
}
