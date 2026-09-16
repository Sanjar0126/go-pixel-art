# go-pixel-art
Make pixel art from image set

# PixelArt & Mosaic Generator (Go)

This project is a Go library and CLI tool that can:

- Convert an image into **pixel art** using the closest matching colors from a palette.
- Convert an image into a **mosaic**, replacing each pixel block with the closest matching image from a set of tiles.

Supports **variable tile sizes** and can be used both as a **CLI tool** and as a **Go library**.

---

## Features

- Pixel art mode (flat color blocks).
- Mosaic mode (tile images replace pixel blocks).
- Variable tile size support (e.g. 8x8, 16x16, 32x32).
- Easy to use as a Go package or via command line.

---

## Installation

Clone the repository and build the native binary with Make:

```bash
git clone https://github.com/Sanjar0126/go-pixel-art.git
cd go-pixel-art
make build
```

The binary is written to `dist/pixelart-<os>-<arch>`.

### Cross-Platform Builds

Build all supported release targets:

```bash
make build-all
```

This creates binaries for Linux (`amd64`, `arm64`), macOS (`amd64`, `arm64`),
and Windows (`amd64`) in `dist/`. Build one target with a named Make target:

```bash
make build-linux-amd64
make build-darwin-arm64
make build-windows-amd64
```

You can also choose any Go-supported target directly:

```bash
make build GOOS=freebsd GOARCH=amd64
```

Run the test suite with:

```bash
make test
```

---

## CLI Usage

The CLI processes every image in the input directory and writes the results to
the output directory. By default, it uses pixel art mode:

```bash
./pixelart -inputDir ./input -outDir ./out -paletteDir ./palette
```

To generate a mosaic instead, add `-mosaic` and provide a directory of tile
images through `-paletteDir`:

```bash
./pixelart -inputDir ./input -outDir ./out -paletteDir ./tiles -mosaic
```

### Flags

| Flag           | Description                                        | Default    |
| -------------- | -------------------------------------------------- | ---------- |
| `-inputDir`    | Directory containing input images                  | `./input`  |
| `-outDir`      | Directory where generated images are written       | `./out`    |
| `-paletteDir`  | Directory containing palette or mosaic tile images | `./palette`|
| `-paletteSize` | Number of colors to build for pixel art mode       | `32`       |
| `-pixelW`      | Width of the pixel grid                            | `64`       |
| `-scale`       | Upscale factor for pixel art mode                  | `8`        |
| `-tileSize`    | Width and height of each mosaic tile               | `16`       |
| `-mosaic`      | Use mosaic mode instead of flat pixel art mode     | `false`    |

---

## Library Usage

You can also import and use this as a Go package:
```go
package main

import (
    "image"
    "os"
    "github.com/Sanjar0126/go-pixel-art"
)

func main() {
    // Open input image
    f, _ := os.Open("input.jpg")
    defer f.Close()
    img, _, _ := image.Decode(f)

    // Example: Pixel Art Mode
    palette := []color.Color{color.Black, color.White, color.RGBA{255,0,0,255}}
    pix, _ := pixelart.ProcessImageToPixelArt(img, palette, 64, 64, 8)
    pixelart.SaveImage("pixel_output.png", pix)

    // Example: Mosaic Mode
    tiles, _ := pixelart.LoadMosaicTiles("./palette", 16)
    mosaic, _ := pixelart.BuildMosaic(img, tiles, 64, 64, 8)
    pixelart.SaveImage("mosaic_output.png", mosaic)
}
```

---

## Palette Notes
* Pixel art mode: palette can be defined manually as a list of colors.
* Mosaic mode: put your tile images inside a folder and pass it to -palette ./tiles.
---
