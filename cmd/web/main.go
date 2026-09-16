//go:build js && wasm

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"syscall/js"

	"github.com/Sanjar0126/go-pixel-art/pixelart"
	"golang.org/x/image/draw"
)

var palette = []color.Color{
	color.RGBA{255, 0, 0, 255},
	color.RGBA{0, 255, 0, 255},
	color.RGBA{0, 0, 255, 255},
	color.RGBA{255, 255, 0, 255},
}

func processImage(_ js.Value, args []js.Value) interface{} {
	if len(args) < 7 {
		return result(js.Undefined(), "image data and processing options are required")
	}

	input := make([]byte, args[0].Get("byteLength").Int())
	js.CopyBytesToGo(input, args[0])
	pixelW := args[1].Int()
	scale := args[2].Int()
	tileSize := args[3].Int()
	mosaic := args[4].Bool()
	paletteFiles := args[5]
	if pixelW < 0 || scale < 1 || tileSize < 1 || args[6].Int() < 1 {
		return result(js.Undefined(), "pixel width must be zero or positive; other sizes must be positive")
	}

	img, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return result(js.Undefined(), err.Error())
	}
	var out image.Image
	if mosaic {
		tiles, err := loadTiles(paletteFiles, tileSize)
		if err != nil {
			return result(js.Undefined(), err.Error())
		}
		gridW, gridH := pixelW, pixelW
		if pixelW == 0 {
			gridW, gridH = img.Bounds().Dx(), img.Bounds().Dy()
		}
		out, err = pixelart.BuildMosaic(img, tiles, gridW, gridH, tileSize)
		if err != nil {
			return result(js.Undefined(), err.Error())
		}
	} else {
		pixelWidth, pixelHeight := pixelW, 0
		if pixelW == 0 {
			pixelWidth, pixelHeight = img.Bounds().Dx(), img.Bounds().Dy()
		}
		out = pixelart.ProcessImageToPixelArt(img, palette, pixelWidth, pixelHeight, scale)
	}

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, out); err != nil {
		return result(js.Undefined(), err.Error())
	}
	data := js.Global().Get("Uint8Array").New(encoded.Len())
	js.CopyBytesToJS(data, encoded.Bytes())
	return result(data, "")
}

func loadTiles(files js.Value, tileSize int) ([]pixelart.MosaicTile, error) {
	tiles := make([]pixelart.MosaicTile, 0, files.Length())
	for i := 0; i < files.Length(); i++ {
		data := files.Index(i)
		input := make([]byte, data.Get("byteLength").Int())
		js.CopyBytesToGo(input, data)
		img, _, err := image.Decode(bytes.NewReader(input))
		if err != nil {
			continue
		}
		dst := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))
		draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
		tiles = append(tiles, pixelart.MosaicTile{Img: dst, Color: pixelart.AverageColor(dst)})
	}
	if len(tiles) == 0 {
		return nil, fmt.Errorf("no valid palette tiles were uploaded")
	}
	return tiles, nil
}

func result(data js.Value, errorMessage string) map[string]interface{} {
	return map[string]interface{}{"data": data, "error": errorMessage}
}

func main() {
	js.Global().Set("processImage", js.FuncOf(processImage))
	select {}
}
