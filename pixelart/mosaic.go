package pixelart

import (
	"errors"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/image/draw"
)

type MosaicTile struct {
	Img   image.Image
	Color color.Color
}

// LoadMosaicTiles loads all images as mosaic tiles of given size.
func LoadMosaicTiles(dir string, tileSize int) ([]MosaicTile, error) {
	var tiles []MosaicTile
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		img, err := LoadImage(path)
		if err != nil {
			return nil
		}
		dst := image.NewRGBA(image.Rect(0, 0, tileSize, tileSize))
		draw.ApproxBiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
		tiles = append(tiles, MosaicTile{Img: dst, Color: AverageColor(dst)})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(tiles) == 0 {
		return nil, errors.New("no valid tiles found")
	}
	return tiles, nil
}

// BuildMosaic builds a mosaic image from src using tiles.
func BuildMosaic(src image.Image, tiles []MosaicTile, gridW, gridH, tileSize int) (image.Image, error) {
	if len(tiles) == 0 {
		return nil, errors.New("tiles required")
	}

	small := image.NewRGBA(image.Rect(0, 0, gridW, gridH))
	draw.ApproxBiLinear.Scale(small, small.Bounds(), src, src.Bounds(), draw.Over, nil)

	out := image.NewRGBA(image.Rect(0, 0, gridW*tileSize, gridH*tileSize))

	total := gridW * gridH
	bar := progressbar.Default(int64(total))

	for y := 0; y < gridH; y++ {
		for x := 0; x < gridW; x++ {
			c := small.At(x, y)
			tile := closestTile(c, tiles)
			r := image.Rect(x*tileSize, y*tileSize, (x+1)*tileSize, (y+1)*tileSize)
			draw.Draw(out, r, tile.Img, image.Point{}, draw.Over)

			bar.Add(1)
		}
	}
	return out, nil
}

func closestTile(c color.Color, tiles []MosaicTile) MosaicTile {
	r1, g1, b1, _ := c.RGBA()
	best := tiles[0]
	bestDist := 1 << 62
	for _, t := range tiles {
		r2, g2, b2, _ := t.Color.RGBA()
		dr := int(r1>>8) - int(r2>>8)
		dg := int(g1>>8) - int(g2>>8)
		db := int(b1>>8) - int(b2>>8)
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			bestDist = dist
			best = t
		}
	}
	return best
}
