package pixelart

import (
	"image"
	"image/color"
)

// AverageColor computes the average color of an image.
func AverageColor(img image.Image) color.Color {
	b := img.Bounds()
	var r, g, bsum, count uint64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			cr, cg, cb, _ := img.At(x, y).RGBA()
			r += uint64(cr)
			g += uint64(cg)
			bsum += uint64(cb)
			count++
		}
	}
	if count == 0 {
		return color.RGBA{0, 0, 0, 255}
	}
	return color.RGBA{
		uint8((r / count) >> 8),
		uint8((g / count) >> 8),
		uint8((bsum / count) >> 8),
		255,
	}
}
