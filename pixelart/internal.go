package pixelart

import (
	"bufio"
	"image"
	"image/color"
	"image/png"
	"os"
)

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func rgbToVec(c color.Color) ColorVec {
	r, g, b, _ := c.RGBA()
	return ColorVec{float64(r >> 8), float64(g >> 8), float64(b >> 8)}
}

func dist2(a, b ColorVec) float64 {
	dx := a[0] - b[0]
	dy := a[1] - b[1]
	dz := a[2] - b[2]
	return dx*dx + dy*dy + dz*dz
}

// --- Palette building (kmeans etc.) ---
// (same functions: kmeans, loadImage, sampleColorsFromImage, buildPaletteFromDir, nearestPaletteColor)

// --- Processing ---
// (same function: processImageToPixelArt)

// --- Saving ---
func saveImageAsPNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	err = png.Encode(w, img)
	if err != nil {
		return err
	}
	return w.Flush()
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
