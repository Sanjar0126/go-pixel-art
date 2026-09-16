//go:build js

package pixelart

import "image"

// SaveImage writes PNG output in browser builds.
func SaveImage(img image.Image, path string) error {
	return SaveImageAsPNG(img, path)
}
