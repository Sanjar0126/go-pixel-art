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

The binary is written to `build/pixelart-<os>-<arch>`.

### Cross-Platform Builds

Build all supported release targets:

```bash
make build-all
```

This creates binaries for Linux (`amd64`, `arm64`), macOS (`amd64`, `arm64`),
and Windows (`amd64`) in `build/`. Build one target with a named Make target:

`make build-all` attempts every target even when one build fails. It returns a
nonzero status at the end if any target failed.

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

### Browser Build

The repository also includes a browser version powered by Go WebAssembly. It
accepts PNG, JPEG, GIF, and WebP uploads, then produces a downloadable PNG.
The browser UI exposes pixel width, scale, tile size, palette size, and mosaic
mode. In mosaic mode, upload the palette tile images with the `Palette tiles`
control, or place PNG tiles in `web/palette/`; `make build-web` generates a
manifest so the bundled palette is loaded automatically. Browser code cannot
read an arbitrary local `paletteDir` path directly.
Build it with:

```bash
make build-web
```

This writes `app.wasm` and `wasm_exec.js` into `web/`. Serve that directory
over HTTP, because browsers do not load WebAssembly correctly from `file://`:

```bash
cd web
python3 -m http.server 8000
```

Open <http://localhost:8000>. The contents of `web/` can be deployed directly
to GitHub Pages.

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

`-paletteDir` is read from the local filesystem. In mosaic mode, every valid
image in that directory becomes a tile. `-pixelW` is the number of grid cells,
while `-tileSize` is the size of each output tile. For example, a 512x512
output made from 16x16 tiles uses 32 grid cells:

```bash
./pixelart -inputDir input -outDir output -pixelW 32 -tileSize 16 \
    -mosaic -paletteDir minecraft-palette
```

Using `-pixelW 512 -tileSize 16` creates an 8192x8192 output. In pixel mode,
`-paletteDir` is currently reserved for the palette builder; the current demo
palette uses four fixed colors.

### Flags

| Flag           | Description                                        | Default    |
| -------------- | -------------------------------------------------- | ---------- |
| `-inputDir`    | Directory containing input images                  | `./input`  |
| `-outDir`      | Directory where generated images are written       | `./out`    |
| `-paletteDir`  | Directory containing palette or mosaic tile images | `./palette`|
| `-paletteSize` | Number of colors to build for pixel art mode       | `32`       |
| `-pixelW`      | Pixel-grid width; `0` preserves the source dimensions | `0`      |
| `-scale`       | Upscale factor for pixel art mode                  | `8`        |
| `-tileSize`    | Width and height of each mosaic tile               | `16`       |
| `-mosaic`      | Use mosaic mode instead of flat pixel art mode     | `false`    |

---

## Minecraft World Generator

`cmd/mcworld` turns an image into a real Minecraft Java Edition world save: it
matches each pixel to the nearest-color block texture (from `web/minecraft-icons`
by default) and places one block per pixel on a flat ground layer.

```bash
make build-mcworld
./build/mcworld-linux-amd64 -image ./input/photo.png -outDir ./out/mcworld
```

Copy the resulting `out/mcworld` folder into your Minecraft `saves/` directory
to open it. The world targets Minecraft **1.16.5** (DataVersion 2586) and has
not been verified by loading it in a real client — treat it as best-effort and
report load issues.

| Flag         | Description                                                    | Default                 |
| ------------ | ---------------------------------------------------------------| ------------------------|
| `-image`     | Input image to render as a block mosaic (required)             | —                        |
| `-iconsDir`  | Directory of block texture PNGs used to build the palette      | `./web/minecraft-icons`  |
| `-outDir`    | Output world save directory                                    | `./out/mcworld`          |
| `-worldName` | World name stored in `level.dat`                                | `Pixel Art World`        |
| `-groundY`   | Y level (0-255) of the flat mosaic ground layer                | `64`                     |
| `-maxWidth`  | Maximum mosaic width in blocks (image is downscaled to fit)     | `256`                    |
| `-maxHeight` | Maximum mosaic height in blocks (image is downscaled to fit)    | `256`                    |

Block texture filtering (which files in `-iconsDir` count as placeable
blocks vs. icons/GUI art) is heuristic — see `nonBlockKeywords` in
`pixelart/mcworld/blocks.go` if you need to tune it.

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
