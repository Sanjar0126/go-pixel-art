package mcworld

import (
	"compress/gzip"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/Sanjar0126/go-pixel-art/pixelart/nbt"
	"golang.org/x/image/draw"
)

const (
	chunkSize    = 16
	worldMinY    = 0
	worldMaxY    = 255
	plainsBiome  = 1 // legacy numeric biome id for "minecraft:plains"
	airBlockID   = "minecraft:air"
	bedrockBlock = "minecraft:bedrock"
)

// Options configures a flat-mosaic world build.
type Options struct {
	// GroundY is the Y level (0-255) where the mosaic surface layer sits.
	GroundY int
	// WorldName is stored in level.dat.
	WorldName string
}

// GenerateFlatWorld renders src onto a flat, one-block-per-pixel mosaic and
// writes a full Minecraft Java world save (level.dat + region/*.mca) to
// outDir, matching each pixel to the nearest color in palette.
func GenerateFlatWorld(src image.Image, palette []BlockColor, outDir string, opts Options) error {
	if len(palette) == 0 {
		return fmt.Errorf("mcworld: palette must not be empty")
	}
	if opts.GroundY < worldMinY || opts.GroundY > worldMaxY {
		return fmt.Errorf("mcworld: groundY must be between %d and %d", worldMinY, worldMaxY)
	}
	if opts.WorldName == "" {
		opts.WorldName = "Pixel Art World"
	}

	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	blockIDs := make([]string, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := src.At(src.Bounds().Min.X+x, src.Bounds().Min.Y+y)
			blockIDs[y*w+x] = closestBlock(c, palette).ID
		}
	}

	chunks := buildChunks(blockIDs, w, h, opts.GroundY)

	regionDir := filepath.Join(outDir, "region")
	timestamp := uint32(time.Now().Unix())
	if err := WriteRegionFiles(regionDir, chunks, timestamp); err != nil {
		return err
	}

	return writeLevelDat(outDir, opts)
}

// DownscaleForWorld resizes src so it fits within maxWidth/maxHeight,
// preserving aspect ratio, using nearest-neighbor sampling so pixel
// boundaries stay crisp when later mapped one-to-one onto blocks.
func DownscaleForWorld(src image.Image, maxWidth, maxHeight int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxWidth && h <= maxHeight {
		return src
	}
	scale := float64(maxWidth) / float64(w)
	if hs := float64(maxHeight) / float64(h); hs < scale {
		scale = hs
	}
	nw := int(float64(w) * scale)
	nh := int(float64(h) * scale)
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.NearestNeighbor.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	return dst
}

// buildChunks lays out blockIDs (row-major, w wide, h tall) starting at
// world (0, groundY, 0), one block per (x, z) column, on top of a bedrock
// floor, and returns the raw (uncompressed) NBT payload for every chunk that
// contains at least one column of the image.
func buildChunks(blockIDs []string, w, h, groundY int) map[chunkCoord][]byte {
	chunks := make(map[chunkCoord][]byte)

	minCX, maxCX := 0, (w-1)/chunkSize
	minCZ, maxCZ := 0, (h-1)/chunkSize

	for cz := minCZ; cz <= maxCZ; cz++ {
		for cx := minCX; cx <= maxCX; cx++ {
			sections := buildChunkSections(blockIDs, w, h, cx, cz, groundY)
			payload := buildChunkNBT(int32(cx), int32(cz), sections, plainsBiome)
			data, err := nbt.EncodeBytes("", payload)
			if err != nil {
				// Encoding is purely in-memory and deterministic; a failure
				// here indicates a programming error, not bad input.
				panic(fmt.Sprintf("mcworld: encode chunk (%d,%d): %v", cx, cz, err))
			}
			chunks[chunkCoord{X: int32(cx), Z: int32(cz)}] = data
		}
	}
	return chunks
}

func buildChunkSections(blockIDs []string, w, h, cx, cz, groundY int) []*sectionBlocks {
	floorY := groundY - 1
	groundSectionY := groundY / chunkSize
	floorSectionY := floorY / chunkSize

	var sections []*sectionBlocks
	if floorSectionY >= 0 && floorSectionY != groundSectionY {
		sections = append(sections, newUniformSection(int8(floorSectionY), bedrockBlock))
	}

	mosaic := buildMosaicSection(blockIDs, w, h, cx, cz, groundY, floorY, floorSectionY == groundSectionY)
	sections = append(sections, mosaic)
	return sections
}

// buildMosaicSection builds the section containing the mosaic surface layer
// (local Y = groundY % 16), optionally also filling local Y for the
// bedrock floor when it shares the same section as the mosaic layer.
func buildMosaicSection(blockIDs []string, w, h, cx, cz, groundY, floorY int, includeFloor bool) *sectionBlocks {
	sectionY := int8(groundY / chunkSize)
	localGroundY := groundY % chunkSize

	palette := []string{airBlockID}
	paletteIndex := map[string]int{airBlockID: 0}
	indices := make([]int, chunkSize*chunkSize*chunkSize)

	indexOf := func(id string) int {
		if idx, ok := paletteIndex[id]; ok {
			return idx
		}
		idx := len(palette)
		palette = append(palette, id)
		paletteIndex[id] = idx
		return idx
	}

	if includeFloor {
		localFloorY := floorY % chunkSize
		floorIdx := indexOf(bedrockBlock)
		fillLayer(indices, localFloorY, floorIdx)
	}

	baseWX := cx * chunkSize
	baseWZ := cz * chunkSize
	for lz := 0; lz < chunkSize; lz++ {
		wz := baseWZ + lz
		if wz < 0 || wz >= h {
			continue
		}
		for lx := 0; lx < chunkSize; lx++ {
			wx := baseWX + lx
			if wx < 0 || wx >= w {
				continue
			}
			id := blockIDs[wz*w+wx]
			idx := indexOf(id)
			pos := (localGroundY*chunkSize+lz)*chunkSize + lx
			indices[pos] = idx
		}
	}

	if len(palette) == 1 {
		return &sectionBlocks{y: sectionY, palette: palette}
	}
	return &sectionBlocks{y: sectionY, indices: indices, palette: palette}
}

func fillLayer(indices []int, localY, idx int) {
	for z := 0; z < chunkSize; z++ {
		for x := 0; x < chunkSize; x++ {
			indices[(localY*chunkSize+z)*chunkSize+x] = idx
		}
	}
}

func writeLevelDat(outDir string, opts Options) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	data := nbt.NewCompound()
	data.Put("DataVersion", nbt.Int(dataVersion1_16_5))
	data.Put("LevelName", nbt.String(opts.WorldName))
	data.Put("GameType", nbt.Int(1)) // creative, so the mosaic can be viewed freely
	data.Put("Difficulty", nbt.Byte(0))
	data.Put("DifficultyLocked", nbt.Byte(0))
	data.Put("hardcore", nbt.Byte(0))
	data.Put("allowCommands", nbt.Byte(1))
	data.Put("initialized", nbt.Byte(1))
	data.Put("SpawnX", nbt.Int(0))
	data.Put("SpawnY", nbt.Int(int32(opts.GroundY+1)))
	data.Put("SpawnZ", nbt.Int(0))
	data.Put("Time", nbt.Long(0))
	data.Put("DayTime", nbt.Long(6000))
	data.Put("LastPlayed", nbt.Long(time.Now().UnixMilli()))

	version := nbt.NewCompound()
	version.Put("Id", nbt.Int(dataVersion1_16_5))
	version.Put("Name", nbt.String("1.16.5"))
	version.Put("Snapshot", nbt.Byte(0))
	data.Put("Version", version)

	flatLayers := &nbt.List{Elem: nbt.TagCompound}
	airLayer := nbt.NewCompound()
	airLayer.Put("block", nbt.String(airBlockID))
	airLayer.Put("height", nbt.Int(1))
	flatLayers.Items = append(flatLayers.Items, airLayer)

	flatSettings := nbt.NewCompound()
	flatSettings.Put("layers", flatLayers)
	flatSettings.Put("biome", nbt.String("minecraft:plains"))

	flatGenerator := nbt.NewCompound()
	flatGenerator.Put("type", nbt.String("minecraft:flat"))
	flatGenerator.Put("settings", flatSettings)

	overworld := nbt.NewCompound()
	overworld.Put("type", nbt.String("minecraft:overworld"))
	overworld.Put("generator", flatGenerator)

	netherGenerator := nbt.NewCompound()
	netherGenerator.Put("type", nbt.String("minecraft:noise"))
	netherGenerator.Put("settings", nbt.String("minecraft:nether"))
	netherBiomeSource := nbt.NewCompound()
	netherBiomeSource.Put("type", nbt.String("minecraft:multi_noise"))
	netherBiomeSource.Put("seed", nbt.Long(0))
	netherBiomeSource.Put("preset", nbt.String("minecraft:nether"))
	netherGenerator.Put("biome_source", netherBiomeSource)
	nether := nbt.NewCompound()
	nether.Put("type", nbt.String("minecraft:the_nether"))
	nether.Put("generator", netherGenerator)

	endGenerator := nbt.NewCompound()
	endGenerator.Put("type", nbt.String("minecraft:noise"))
	endGenerator.Put("settings", nbt.String("minecraft:end"))
	endBiomeSource := nbt.NewCompound()
	endBiomeSource.Put("type", nbt.String("minecraft:the_end"))
	endBiomeSource.Put("seed", nbt.Long(0))
	endGenerator.Put("biome_source", endBiomeSource)
	theEnd := nbt.NewCompound()
	theEnd.Put("type", nbt.String("minecraft:the_end"))
	theEnd.Put("generator", endGenerator)

	dimensions := nbt.NewCompound()
	dimensions.Put("minecraft:overworld", overworld)
	dimensions.Put("minecraft:the_nether", nether)
	dimensions.Put("minecraft:the_end", theEnd)

	genSettings := nbt.NewCompound()
	genSettings.Put("seed", nbt.Long(0))
	genSettings.Put("generate_features", nbt.Byte(0))
	genSettings.Put("bonus_chest", nbt.Byte(0))
	genSettings.Put("dimensions", dimensions)
	data.Put("WorldGenSettings", genSettings)

	dataPacksEnabled := &nbt.List{Elem: nbt.TagString, Items: []nbt.Value{nbt.String("vanilla")}}
	dataPacks := nbt.NewCompound()
	dataPacks.Put("Enabled", dataPacksEnabled)
	dataPacks.Put("Disabled", &nbt.List{Elem: nbt.TagEnd})
	data.Put("DataPacks", dataPacks)

	root := nbt.NewCompound()
	root.Put("Data", data)

	f, err := os.Create(filepath.Join(outDir, "level.dat"))
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	if err := nbt.Encode(gw, "", root); err != nil {
		gw.Close()
		return err
	}
	return gw.Close()
}
