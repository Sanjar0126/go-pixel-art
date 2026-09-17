package mcworld

import "github.com/Sanjar0126/go-pixel-art/pixelart/nbt"

const dataVersion1_16_5 = 2586

// sectionBlocks holds the block id index (into a section's local palette)
// for every one of the 4096 positions in a 16x16x16 section, addressed as
// index = ((y*16)+z)*16+x.
type sectionBlocks struct {
	y       int8
	indices []int
	palette []string // block ids, e.g. "minecraft:stone"
}

func newUniformSection(y int8, blockID string) *sectionBlocks {
	return &sectionBlocks{y: y, palette: []string{blockID}}
}

// buildSectionTag encodes one chunk section as an NBT compound.
func buildSectionTag(s *sectionBlocks) *nbt.Compound {
	c := nbt.NewCompound()
	c.Put("Y", nbt.Byte(s.y))

	paletteList := &nbt.List{Elem: nbt.TagCompound}
	for _, id := range s.palette {
		entry := nbt.NewCompound()
		entry.Put("Name", nbt.String(id))
		paletteList.Items = append(paletteList.Items, entry)
	}
	c.Put("Palette", paletteList)

	if len(s.palette) > 1 {
		bits := bitsForPalette(len(s.palette))
		packed := packLongArray(s.indices, bits)
		c.Put("BlockStates", nbt.LongArray(packed))
	}
	return c
}

// buildChunkNBT encodes a full chunk (16x16 columns) as its root NBT
// document payload, following the pre-1.18 "Level" wrapped chunk format
// (Minecraft 1.16.5, DataVersion 2586).
func buildChunkNBT(chunkX, chunkZ int32, sections []*sectionBlocks, biomeID int32) *nbt.Compound {
	root := nbt.NewCompound()
	root.Put("DataVersion", nbt.Int(dataVersion1_16_5))

	level := nbt.NewCompound()
	level.Put("xPos", nbt.Int(chunkX))
	level.Put("zPos", nbt.Int(chunkZ))
	level.Put("Status", nbt.String("full"))
	level.Put("LastUpdate", nbt.Long(0))
	level.Put("InhabitedTime", nbt.Long(0))
	level.Put("isLightOn", nbt.Byte(0))

	sectionList := &nbt.List{Elem: nbt.TagCompound}
	for _, s := range sections {
		sectionList.Items = append(sectionList.Items, buildSectionTag(s))
	}
	level.Put("Sections", sectionList)

	biomes := make(nbt.IntArray, 1024)
	for i := range biomes {
		biomes[i] = biomeID
	}
	level.Put("Biomes", biomes)

	level.Put("Entities", &nbt.List{Elem: nbt.TagEnd})
	level.Put("TileEntities", &nbt.List{Elem: nbt.TagEnd})

	root.Put("Level", level)
	return root
}
