package qrcode

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

// bchRemainderLong computes the remainder of poly divided by the format
// generator polynomial 0x537 (degree 10) using plain GF(2) long division,
// independently from the encoder's shift-register approach.
func bchRemainderLong(poly int) int {
	for deg := 14; deg >= 10; deg-- {
		if (poly>>uint(deg))&1 == 1 {
			poly ^= 0x537 << uint(deg-10)
		}
	}
	return poly & 0x3FF
}

// readFormatBitsCopy reconstructs the 15-bit format value from the given
// copy. copy 0 = around the top-left finder, copy 1 = split copies.
func readFormatBits(c *code, copy int) int {
	val := 0
	if copy == 0 {
		for i := 0; i <= 5; i++ {
			if c.modules[i][8] {
				val |= 1 << uint(i)
			}
		}
		if c.modules[7][8] {
			val |= 1 << 6
		}
		if c.modules[8][8] {
			val |= 1 << 7
		}
		if c.modules[8][7] {
			val |= 1 << 8
		}
		for i := 9; i < 15; i++ {
			if c.modules[8][14-i] {
				val |= 1 << uint(i)
			}
		}
	} else {
		for i := 0; i < 8; i++ {
			if c.modules[8][c.size-1-i] {
				val |= 1 << uint(i)
			}
		}
		for i := 8; i < 15; i++ {
			if c.modules[c.size-15+i][8] {
				val |= 1 << uint(i)
			}
		}
	}
	return val
}

func TestFormatBitsBCHValid(t *testing.T) {
	for _, text := range []string{"A", "http://192.168.1.155:1700", strings.Repeat("x", 200)} {
		c, err := encode(text)
		if err != nil {
			t.Fatalf("encode(%q): %v", text, err)
		}
		f0 := readFormatBits(c, 0)
		f1 := readFormatBits(c, 1)
		if f0 != f1 {
			t.Errorf("%q: format copies differ: %#x vs %#x", text, f0, f1)
		}
		masked := f0 ^ 0x5412
		if bchRemainderLong(masked) != 0 {
			t.Errorf("%q: format bits %#x fail BCH check (remainder %#x)", text, f0, bchRemainderLong(masked))
		}
		data := masked >> 10
		if data>>3 != 0 { // ECC level bits must be M (0)
			t.Errorf("%q: unexpected ECC bits in format %#x", text, f0)
		}
		if mask := data & 7; mask != c.mask {
			t.Errorf("%q: format mask %d != chosen mask %d", text, mask, c.mask)
		}
	}
}

func TestFormatBitsKnownValues(t *testing.T) {
	// Independent long-division check against known correct values.
	if got := bchRemainderLong(0x400); got != 0x137 {
		t.Errorf("bchRemainderLong(1<<10) = %#x, want 0x137", got)
	}
	// (ECC M=00, mask 0) -> 0x5412; (M, mask 1) -> 0x5125.
	want0 := 0x5412
	want1 := 0x5125
	f0 := computeFormat(0)
	if f0 != want0 {
		t.Errorf("computeFormat(0) = %#x, want %#x", f0, want0)
	}
	f1 := computeFormat(1)
	if f1 != want1 {
		t.Errorf("computeFormat(1) = %#x, want %#x (indep rem %#x)", f1, want1, bchRemainderLong(0x400))
	}
}

func computeFormat(mask int) int {
	data := mask
	rem := data
	for i := 0; i < 10; i++ {
		rem = (rem << 1) ^ (rem>>9)*0x537
	}
	return (data<<10 | rem) ^ 0x5412
}

func TestReedSolomonDivisibility(t *testing.T) {
	for ver := minVersion; ver <= maxVersion; ver++ {
		blockEccLen := eccCodecPerBlock[ver]
		divisor := rsComputeDivisor(blockEccLen)
		if len(divisor) != blockEccLen {
			t.Fatalf("v%d: divisor length %d != %d", ver, len(divisor), blockEccLen)
		}
		// A full block (data + EC) must divide evenly -> remainder 0.
		dataLen := 10 + ver
		data := make([]byte, dataLen)
		for i := range data {
			data[i] = byte(0x25 * (i + 1))
		}
		ecc := rsComputeRemainder(data, divisor)
		full := append(append([]byte{}, data...), ecc...)
		if rem := rsComputeRemainder(full, divisor); !bytes.Equal(rem, make([]byte, blockEccLen)) {
			t.Errorf("v%d: full block not divisible, remainder = %v", ver, rem)
		}
	}
}

func TestMatrixStructure(t *testing.T) {
	c, err := encode("hello")
	if err != nil {
		t.Fatal(err)
	}
	if c.size != c.version*4+17 {
		t.Errorf("size %d != version*4+17", c.size)
	}
	// Finders: centers at (3,3), (size-4,3), (3,size-4) are dark.
	finder := func(x, y int) bool { return c.modules[y][x] }
	if !finder(3, 3) || !finder(c.size-4, 3) || !finder(3, c.size-4) {
		t.Error("finder pattern centers not dark")
	}
	// Timing row: at y=6, x and x+1 alternate.
	if c.modules[6][6] != c.modules[6][8] || c.modules[6][6] == c.modules[6][7] {
		t.Error("timing pattern does not alternate")
	}
	// Dark module at (8, size-8).
	if !c.modules[c.size-8][8] {
		t.Error("dark module missing")
	}
	// Versions inside supported range.
	if c.version < 1 || c.version > maxVersion {
		t.Errorf("version out of range: %d", c.version)
	}
}

func TestSmallAndLargeInputs(t *testing.T) {
	for _, n := range []int{1, 14, 15, 26, 42, 62, 84, 106, 122, 152, 180, 213} {
		if _, err := encode(strings.Repeat("a", n)); err != nil {
			t.Errorf("encode %d bytes: %v", n, err)
		}
	}
	if _, err := encode(strings.Repeat("a", 220)); err == nil {
		t.Error("expected error for data that does not fit")
	}
}

func TestAlignmentPositions(t *testing.T) {
	cases := []struct {
		ver int
		want []int
	}{
		{2, []int{6, 18}},
		{7, []int{6, 22, 38}},
		{10, []int{6, 28, 50}},
	}
	for _, tc := range cases {
		c := &code{version: tc.ver, size: tc.ver*4 + 17}
		got := c.alignmentPatternPositions()
		if len(got) != len(tc.want) {
			t.Errorf("v%d: got %v want %v", tc.ver, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("v%d: got %v want %v", tc.ver, got, tc.want)
				break
			}
		}
	}
}

func TestDataCodewordCounts(t *testing.T) {
	want := map[int]int{1: 16, 2: 28, 3: 44, 4: 64, 5: 86, 6: 108, 7: 124, 8: 154, 9: 182, 10: 216}
	for ver, expected := range want {
		if got := getNumDataCodewords(ver); got != expected {
			t.Errorf("v%d: data codewords = %d, want %d", ver, got, expected)
		}
	}
}

func TestInterleaveLength(t *testing.T) {
	for ver := minVersion; ver <= maxVersion; ver++ {
		data := make([]byte, getNumDataCodewords(ver))
		all := addEccAndInterleave(data, ver)
		want := getNumRawDataModules(ver) / 8
		if len(all) != want {
			t.Errorf("v%d: interleaved length %d, want %d", ver, len(all), want)
		}
	}
}

func TestEncodePNG(t *testing.T) {
	pngBytes, err := EncodePNG("http://localhost:1700")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(pngBytes, []byte{0x89, 'P', 'N', 'G'}) {
		t.Error("output is not a PNG")
	}
	// Data URI variant embeds the same PNG.
	uri, err := EncodeDataURI("http://localhost:1700")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(uri, "data:image/png;base64,") {
		t.Errorf("unexpected data uri prefix: %s", uri[:30])
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(uri, "data:image/png;base64,"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(decoded, []byte{0x89, 'P', 'N', 'G'}) {
		t.Error("data uri does not decode to PNG")
	}
}

func TestEachMaskRendered(t *testing.T) {
	// Every input must produce a valid BCH-correct format (covers all mask
	// selection paths) — extend the general inputs a bit.
	for i := 0; i < 20; i++ {
		text := strings.Repeat("data", i)
		c, err := encode(text)
		if err != nil {
			t.Fatalf("encode %q: %v", text, err)
		}
		f := readFormatBits(c, 0)
		if bchRemainderLong(f^0x5412) != 0 {
			t.Errorf("input %q: mask %d produced invalid format %#x", text, c.mask, f)
		}
	}
}