package mcworld

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"path/filepath"
)

const (
	sectorSize      = 4096
	regionChunkSpan = 32 // chunks per region file edge
)

// chunkCoord identifies a chunk by absolute chunk coordinates.
type chunkCoord struct{ X, Z int32 }

// regionCoord identifies a region file by region coordinates.
type regionCoord struct{ X, Z int32 }

func chunkToRegion(c chunkCoord) regionCoord {
	return regionCoord{X: floorDiv(c.X, regionChunkSpan), Z: floorDiv(c.Z, regionChunkSpan)}
}

func floorDiv(a, b int32) int32 {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

// WriteRegionFiles groups the given raw (uncompressed) chunk NBT payloads by
// the region file they belong to and writes standard Anvil .mca files into
// dir (typically "<world>/region").
func WriteRegionFiles(dir string, chunks map[chunkCoord][]byte, timestamp uint32) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	byRegion := make(map[regionCoord]map[chunkCoord][]byte)
	for coord, data := range chunks {
		rc := chunkToRegion(coord)
		if byRegion[rc] == nil {
			byRegion[rc] = make(map[chunkCoord][]byte)
		}
		byRegion[rc][coord] = data
	}
	for rc, chunkSet := range byRegion {
		path := filepath.Join(dir, fmt.Sprintf("r.%d.%d.mca", rc.X, rc.Z))
		if err := writeRegionFile(path, rc, chunkSet, timestamp); err != nil {
			return err
		}
	}
	return nil
}

func writeRegionFile(path string, rc regionCoord, chunkSet map[chunkCoord][]byte, timestamp uint32) error {
	type entry struct {
		localX, localZ int
		compressed     []byte
	}
	var entries []entry
	for coord, raw := range chunkSet {
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)
		if _, err := zw.Write(raw); err != nil {
			return err
		}
		if err := zw.Close(); err != nil {
			return err
		}
		localX := int(coord.X - rc.X*regionChunkSpan)
		localZ := int(coord.Z - rc.Z*regionChunkSpan)
		entries = append(entries, entry{localX: localX, localZ: localZ, compressed: buf.Bytes()})
	}

	locations := make([]byte, sectorSize)
	timestamps := make([]byte, sectorSize)
	var body bytes.Buffer
	nextSector := 2 // sectors 0 and 1 are the header tables

	for _, e := range entries {
		payloadLen := len(e.compressed) + 1 // + compression type byte
		totalLen := payloadLen + 4          // + length field itself
		sectorCount := (totalLen + sectorSize - 1) / sectorSize

		header := make([]byte, 5)
		putUint32(header[0:4], uint32(payloadLen))
		header[4] = 2 // zlib compression
		body.Write(header)
		body.Write(e.compressed)
		if pad := sectorCount*sectorSize - totalLen; pad > 0 {
			body.Write(make([]byte, pad))
		}

		idx := e.localZ*regionChunkSpan + e.localX
		putLocationEntry(locations[idx*4:idx*4+4], nextSector, sectorCount)
		putUint32(timestamps[idx*4:idx*4+4], timestamp)

		nextSector += sectorCount
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(locations); err != nil {
		return err
	}
	if _, err := f.Write(timestamps); err != nil {
		return err
	}
	_, err = f.Write(body.Bytes())
	return err
}

func putUint32(b []byte, v uint32) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

func putLocationEntry(b []byte, sectorOffset, sectorCount int) {
	b[0] = byte(sectorOffset >> 16)
	b[1] = byte(sectorOffset >> 8)
	b[2] = byte(sectorOffset)
	b[3] = byte(sectorCount)
}
