// Package mcworld turns an image into a Minecraft Java Edition world save
// (level.dat plus Anvil region files) by placing one block per pixel on a
// flat layer, using a block palette derived from bundled icon textures.
//
// The Anvil/NBT structures written here target Minecraft 1.16.5
// (DataVersion 2586). This has not been validated by loading the result in
// a real Minecraft client; treat it as best-effort and report load issues.
package mcworld

import "math/bits"

// bitsForPalette returns the number of bits needed per block-state index for
// an indirect palette of the given size, with vanilla's minimum of 4 bits.
func bitsForPalette(paletteSize int) int {
	if paletteSize <= 1 {
		return 0
	}
	b := bits.Len(uint(paletteSize - 1))
	if b < 4 {
		b = 4
	}
	return b
}

// packLongArray packs indices into 64-bit words using the post-1.16 scheme,
// where each entry is fully contained within a single long (no entry ever
// spans two longs); unused high bits in the last-used portion of each long
// are zero.
func packLongArray(indices []int, bitsPerEntry int) []int64 {
	if bitsPerEntry <= 0 {
		return nil
	}
	valuesPerLong := 64 / bitsPerEntry
	numLongs := (len(indices) + valuesPerLong - 1) / valuesPerLong
	out := make([]int64, numLongs)
	mask := uint64(1)<<uint(bitsPerEntry) - 1
	for i, v := range indices {
		longIndex := i / valuesPerLong
		bitIndex := uint(i%valuesPerLong) * uint(bitsPerEntry)
		out[longIndex] |= int64((uint64(v) & mask) << bitIndex)
	}
	return out
}
