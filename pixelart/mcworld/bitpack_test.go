package mcworld

import "testing"

func TestBitsForPalette(t *testing.T) {
	cases := map[int]int{1: 0, 2: 4, 15: 4, 16: 4, 17: 5, 256: 8, 257: 9}
	for size, want := range cases {
		if got := bitsForPalette(size); got != want {
			t.Errorf("bitsForPalette(%d) = %d, want %d", size, got, want)
		}
	}
}

func TestPackLongArrayRoundTrip(t *testing.T) {
	indices := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 1, 2, 3}
	bits := bitsForPalette(16)
	packed := packLongArray(indices, bits)

	valuesPerLong := 64 / bits
	wantLongs := (len(indices) + valuesPerLong - 1) / valuesPerLong
	if len(packed) != wantLongs {
		t.Fatalf("expected %d longs, got %d", wantLongs, len(packed))
	}

	mask := int64(1)<<uint(bits) - 1
	for i, want := range indices {
		longIdx := i / valuesPerLong
		bitIdx := uint(i%valuesPerLong) * uint(bits)
		got := (packed[longIdx] >> bitIdx) & mask
		if got != int64(want) {
			t.Errorf("index %d: got %d, want %d", i, got, want)
		}
	}
}
