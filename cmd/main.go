package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sanjar0126/go-pixel-art/pixelart"
)

func main() {
	inputDir := flag.String("inputDir", "./input", "Input images directory")
	outDir := flag.String("outDir", "./out", "Output directory")
	paletteDir := flag.String("paletteDir", "./palette", "Palette directory (for colors or mosaic tiles)")
	paletteSize := flag.Int("paletteSize", 32, "Palette size for pixel mode")
	pixelW := flag.Int("pixelW", 64, "Pixel width")
	scale := flag.Int("scale", 8, "Upscale factor")
	tileSize := flag.Int("tileSize", 16, "Mosaic tile size")
	mosaicMode := flag.Bool("mosaic", false, "Use mosaic mode instead of flat pixel mode")
	flag.Parse()

	os.MkdirAll(*outDir, 0755)

	if *mosaicMode {
		tiles, err := pixelart.LoadMosaicTiles(*paletteDir, *tileSize)
		if err != nil {
			panic(err)
		}
		filepath.Walk(*inputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			img, err := pixelart.LoadImage(path)
			if err != nil {
				return nil
			}
			out, err := pixelart.BuildMosaic(img, tiles, *pixelW, *pixelW, *tileSize)
			if err != nil {
				return err
			}
			outPath := filepath.Join(*outDir, info.Name())
			pixelart.SaveImageAsPNG(out, outPath)
			fmt.Println("Saved:", outPath)
			return nil
		})
	} else {
		palette, err := pixelart.BuildPaletteFromDir(*paletteDir, *paletteSize, 500, 200)
		if err != nil {
			panic(err)
		}
		filepath.Walk(*inputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			img, err := pixelart.LoadImage(path)
			if err != nil {
				return nil
			}
			out := pixelart.ProcessImageToPixelArt(img, palette, *pixelW, 0, *scale)
			outPath := filepath.Join(*outDir, info.Name())
			pixelart.SaveImageAsPNG(out, outPath)
			fmt.Println("Saved:", outPath)
			return nil
		})
	}
}
