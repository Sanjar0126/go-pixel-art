package pixelart

import (
	"image"
	"image/color"
)

type ColorVec [3]float64

func (c ColorVec) toNRGBA() color.NRGBA {
	return color.NRGBA{R: uint8(clamp(c[0], 0, 255)), G: uint8(clamp(c[1], 0, 255)), B: uint8(clamp(c[2], 0, 255)), A: 255}
}

// Exported API
func BuildPaletteFromDir(dir string, paletteSize int, samplePerImage int, maxImages int) ([]ColorVec, error) {
	return buildPaletteFromDir(dir, paletteSize, samplePerImage, maxImages)
}

func ProcessImageToPixelArt(img image.Image, palette []ColorVec, pixelW, pixelH, scale int) image.Image {
	return processImageToPixelArt(img, palette, pixelW, pixelH, scale)
}

func SaveImageAsPNG(img image.Image, path string) error {
	return saveImageAsPNG(img, path)
}
