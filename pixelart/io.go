package pixelart

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	webpEncoder "github.com/chai2010/webp"
	_ "golang.org/x/image/webp"
)

// LoadImage loads an image from a file.
func LoadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

// SaveImageAsPNG saves an image as PNG to given path.
func SaveImageAsPNG(img image.Image, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

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
