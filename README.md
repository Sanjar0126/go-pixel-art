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

Clone the repository and build:

```bash
git clone https://github.com/yourusername/pixelart.git
cd pixelart/cmd/pixelart
go build -o pixelart
```

Now you can run the CLI with ./pixelart.

---

## Cli Usage

```bash
./pixelart -in input.jpg -out output.png -width 64 -height 64 -tileSize 16 -palette ./palette -mosaic
```
Flags
| Flag        | Description                                                                          | Default                 |
| -------- | ------------------------------------------------------------------------------------ | ----------------------- |
| `-in`       | Input image file path                                                                | required                |
| `-out`      | Output image file path                                                               | required                |
| `-width`    | Grid width in blocks                                                                 | 64                      |
| `-height`   | Grid height in blocks                                                                | 64                      |
| `-tileSize` | Tile size (for mosaic mode, e.g. 16 = 16x16 per block)                          | 16                      |
| `-palette`  | Path to palette directory (images for mosaic mode, or JSON of colors for pixel mode) | required in mosaic mode |
| `-mosaic`   | Use mosaic mode (otherwise defaults to pixel art mode)                               | false                   |

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
    pix, _ := pixelart.Pixelate(img, palette, 64, 64)
    pixelart.SaveImage("pixel_output.png", pix)

    // Example: Mosaic Mode
    tiles, _ := pixelart.LoadMosaicTiles("./palette", 16)
    mosaic, _ := pixelart.BuildMosaic(img, tiles, 64, 64, 16)
    pixelart.SaveImage("mosaic_output.png", mosaic)
}
```

---

## Palette Notes
* Pixel art mode: palette can be defined manually as a list of colors.
* Mosaic mode: put your tile images inside a folder and pass it to -palette ./tiles.
---
