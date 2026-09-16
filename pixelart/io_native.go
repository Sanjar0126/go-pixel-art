//go:build !js

package pixelart

import (
	"image"
	"os"
	"path/filepath"
	"strings"

	webpEncoder "github.com/chai2010/webp"
)

// SaveImageAsWebP saves an image as WebP to the given path.
func SaveImageAsWebP(img image.Image, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return webpEncoder.Encode(f, img, &webpEncoder.Options{Lossless: true})
}

// SaveImage chooses WebP for .webp paths and PNG for all other paths.
func SaveImage(img image.Image, path string) error {
	if strings.EqualFold(filepath.Ext(path), ".webp") {
		return SaveImageAsWebP(img, path)
	}
	return SaveImageAsPNG(img, path)
}
