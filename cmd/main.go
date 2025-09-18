package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Sanjar0126/go-pixel-art/pixelart"
)

func main() {
	paletteDir := flag.String("paletteDir", "./palette_images", "directory with images used to build the palette")
	inputDir := flag.String("inputDir", "./input", "directory with images to convert")
	outDir := flag.String("outDir", "./out", "output directory")
	paletteSize := flag.Int("paletteSize", 32, "number of colors in palette (k)")
	samplePerImage := flag.Int("samplePerImage", 500, "how many color samples to take from each palette image")
	maxPaletteImages := flag.Int("maxPaletteImages", 200, "max number of images to read from paletteDir")
	pixelW := flag.Int("pixelW", 64, "pixel-art width in pixels")
	pixelH := flag.Int("pixelH", 0, "pixel-art height in pixels (0 = preserve aspect ratio)")
	scale := flag.Int("scale", 8, "upscale factor to make pixels visible")
	workers := flag.Int("workers", 4, "concurrent workers for processing images")
	flag.Parse()

	fmt.Println("Building palette from", *paletteDir)
	palette, err := pixelart.BuildPaletteFromDir(*paletteDir, *paletteSize, *samplePerImage, *maxPaletteImages)
	if err != nil {
		log.Fatalf("failed to build palette: %v", err)
	}
	fmt.Printf("Palette built: %d colors\n", len(palette))

	if err := pixelart.EnsureDir(*outDir); err != nil {
		log.Fatalf("failed to create outDir: %v", err)
	}

	inputPaths := []string{}
	filepath.Walk(*inputDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp" {
				inputPaths = append(inputPaths, path)
			}
		}
		return nil
	})

	tasks := make(chan string, len(inputPaths))
	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for p := range tasks {
				fmt.Printf("worker %d processing %s\n", id, p)
				img, err := pixelart.LoadImage(p)
				if err != nil {
					log.Printf("failed load %s: %v", p, err)
					continue
				}
				outImg := pixelart.ProcessImageToPixelArt(img, palette, *pixelW, *pixelH, *scale)
				rel, _ := filepath.Rel(*inputDir, p)
				outPath := filepath.Join(*outDir, strings.TrimSuffix(rel, filepath.Ext(rel))+"_pixel.png")
				pixelart.EnsureDir(filepath.Dir(outPath))
				if err := pixelart.SaveImageAsPNG(outImg, outPath); err != nil {
					log.Printf("save error for %s: %v", outPath, err)
				}
			}
		}(i)
	}
	for _, p := range inputPaths {
		tasks <- p
	}
	close(tasks)
	wg.Wait()

	fmt.Println("Done. Output in", *outDir)
}
