package mcworld

import (
	"bytes"
	"compress/zlib"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Sanjar0126/go-pixel-art/pixelart/nbt"
)

func TestWriteRegionFilesRoundTrip(t *testing.T) {
	dir := t.TempDir()

	chunks := map[chunkCoord][]byte{
		{X: 0, Z: 0}:  []byte("chunk-0-0-payload"),
		{X: 1, Z: 0}:  []byte("chunk-1-0-payload"),
		{X: 33, Z: 0}: []byte("chunk-in-next-region"), // region (1,0)
	}

	if err := WriteRegionFiles(dir, chunks, 12345); err != nil {
		t.Fatalf("WriteRegionFiles: %v", err)
	}

	assertChunk := func(regionFile string, localX, localZ int, want []byte) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(dir, regionFile))
		if err != nil {
			t.Fatalf("read %s: %v", regionFile, err)
		}
		if len(data)%sectorSize != 0 {
			t.Fatalf("%s: length %d is not sector-aligned", regionFile, len(data))
		}
		idx := localZ*regionChunkSpan + localX
		loc := data[idx*4 : idx*4+4]
		sectorOffset := int(loc[0])<<16 | int(loc[1])<<8 | int(loc[2])
		sectorCount := int(loc[3])
		if sectorOffset == 0 && sectorCount == 0 {
			t.Fatalf("%s: chunk (%d,%d) not present", regionFile, localX, localZ)
		}
		start := sectorOffset * sectorSize
		end := start + sectorCount*sectorSize
		if end > len(data) {
			t.Fatalf("%s: chunk data out of bounds", regionFile)
		}
		chunkBytes := data[start:end]
		length := int(chunkBytes[0])<<24 | int(chunkBytes[1])<<16 | int(chunkBytes[2])<<8 | int(chunkBytes[3])
		compressionType := chunkBytes[4]
		if compressionType != 2 {
			t.Fatalf("%s: expected zlib compression type 2, got %d", regionFile, compressionType)
		}
		compressed := chunkBytes[5 : 5+length-1]
		zr, err := zlib.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatalf("%s: zlib reader: %v", regionFile, err)
		}
		raw, err := io.ReadAll(zr)
		if err != nil {
			t.Fatalf("%s: zlib read: %v", regionFile, err)
		}
		if !bytes.Equal(raw, want) {
			t.Fatalf("%s: got %q, want %q", regionFile, raw, want)
		}
	}

	assertChunk("r.0.0.mca", 0, 0, []byte("chunk-0-0-payload"))
	assertChunk("r.0.0.mca", 1, 0, []byte("chunk-1-0-payload"))
	assertChunk("r.1.0.mca", 1, 0, []byte("chunk-in-next-region"))
}

func TestBuildChunksProducesDecodableNBT(t *testing.T) {
	blockIDs := []string{
		"minecraft:stone", "minecraft:dirt",
		"minecraft:sand", "minecraft:stone",
	}
	chunks := buildChunks(blockIDs, 2, 2, 64)
	data, ok := chunks[chunkCoord{X: 0, Z: 0}]
	if !ok {
		t.Fatal("missing chunk (0,0)")
	}

	name, root, err := nbt.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode chunk nbt: %v", err)
	}
	if name != "" {
		t.Fatalf("expected empty root name, got %q", name)
	}
	dv, ok := root.Get("DataVersion")
	if !ok || dv != nbt.Int(dataVersion1_16_5) {
		t.Fatalf("unexpected DataVersion: %#v", dv)
	}
	levelVal, ok := root.Get("Level")
	if !ok {
		t.Fatal("missing Level compound")
	}
	level, ok := levelVal.(*nbt.Compound)
	if !ok {
		t.Fatalf("Level is not a compound: %#v", levelVal)
	}
	sectionsVal, ok := level.Get("Sections")
	if !ok {
		t.Fatal("missing Sections")
	}
	sections, ok := sectionsVal.(*nbt.List)
	if !ok || len(sections.Items) == 0 {
		t.Fatalf("unexpected Sections: %#v", sectionsVal)
	}
}
