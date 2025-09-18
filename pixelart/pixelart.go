package pixelart

import (
	"image"
	"image/color"

	"golang.org/x/image/draw"
)

// BuildPaletteFromDir creates a palette of colors (stubbed for demo).
// Replace with k-means if needed.
func BuildPaletteFromDir(dir string, paletteSize, resizeW, resizeH int) ([]color.Color, error) {
	return []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 0, 255, 255},
		color.RGBA{255, 255, 0, 255},
	}, nil
}

// ProcessImageToPixelArt converts an image to pixel art using palette colors.
func ProcessImageToPixelArt(img image.Image, palette []color.Color, pixelW, pixelH, scale int) image.Image {
	if pixelH == 0 {
		pixelH = pixelW
	}
	small := image.NewRGBA(image.Rect(0, 0, pixelW, pixelH))
	draw.ApproxBiLinear.Scale(small, small.Bounds(), img, img.Bounds(), draw.Over, nil)

	for y := 0; y < small.Bounds().Dy(); y++ {
		for x := 0; x < small.Bounds().Dx(); x++ {
			c := small.At(x, y)
			nearest := closestColor(c, palette)
			small.Set(x, y, nearest)
		}
	}

	out := image.NewRGBA(image.Rect(0, 0, pixelW*scale, pixelH*scale))
	draw.NearestNeighbor.Scale(out, out.Bounds(), small, small.Bounds(), draw.Over, nil)
	return out
}

func closestColor(c color.Color, palette []color.Color) color.Color {
	r1, g1, b1, _ := c.RGBA()
	best := palette[0]
	bestDist := 1 << 62
	for _, pc := range palette {
		r2, g2, b2, _ := pc.RGBA()
		dr := int(r1>>8) - int(r2>>8)
		dg := int(g1>>8) - int(g2>>8)
		db := int(b1>>8) - int(b2>>8)
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			bestDist = dist
			best = pc
		}
	}
	return best
}
