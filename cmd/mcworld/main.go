package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Sanjar0126/go-pixel-art/pixelart"
	"github.com/Sanjar0126/go-pixel-art/pixelart/mcworld"
)

func main() {
	inputImage := flag.String("image", "", "Input image to render as a block mosaic (required)")
	iconsDir := flag.String("iconsDir", "./web/minecraft-icons", "Directory of Minecraft block texture PNGs used to build the block palette")
	outDir := flag.String("outDir", "./out/mcworld", "Output world save directory (copy into <minecraft>/saves/ to play)")
	worldName := flag.String("worldName", "Pixel Art World", "World name stored in level.dat")
	groundY := flag.Int("groundY", 64, "Y level (0-255) of the flat mosaic ground layer")
	maxWidth := flag.Int("maxWidth", 256, "Maximum mosaic width in blocks (image is downscaled to fit)")
	maxHeight := flag.Int("maxHeight", 256, "Maximum mosaic height in blocks (image is downscaled to fit)")
	flag.Parse()

	if *inputImage == "" {
		fmt.Fprintln(os.Stderr, "mcworld: -image is required")
		flag.Usage()
		os.Exit(2)
	}

	palette, err := mcworld.BuildPaletteFromIcons(*iconsDir)
	if err != nil {
		log.Fatalf("building block palette from %s: %v", *iconsDir, err)
	}
	if len(palette) == 0 {
		log.Fatalf("no usable block textures found in %s", *iconsDir)
	}
	log.Printf("Loaded %d candidate blocks from %s", len(palette), *iconsDir)

	img, err := pixelart.LoadImage(*inputImage)
	if err != nil {
		log.Fatalf("loading %s: %v", *inputImage, err)
	}
	img = mcworld.DownscaleForWorld(img, *maxWidth, *maxHeight)

	opts := mcworld.Options{GroundY: *groundY, WorldName: *worldName}
	if err := mcworld.GenerateFlatWorld(img, palette, *outDir, opts); err != nil {
		log.Fatalf("generating world: %v", err)
	}

	b := img.Bounds()
	log.Printf("Wrote a %dx%d block mosaic world to %s", b.Dx(), b.Dy(), *outDir)
	log.Println("Copy that folder into your Minecraft 'saves' directory to open it (targets Minecraft 1.16.5; not verified against a real client).")
}
