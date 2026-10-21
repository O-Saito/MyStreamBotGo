// Package qrcode implements a QR Code encoder (ISO/IEC 18004, Model 2)
// using only the Go standard library. It supports byte mode, error
// correction level M, and versions 1 to 10.
//
// The algorithm follows the public-domain reference described at
// https://www.nayuki.io/page/qr-code-generator-library (MIT licensed),
// re-implemented here from scratch to avoid third-party dependencies.
package qrcode

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
)

const (
	minVersion = 1
	maxVersion = 10
)

// eccCodecPerBlock[v] is the number of error correction codewords per
// block for version v at error correction level M (index 0 unused).
var eccCodecPerBlock = [maxVersion + 1]int{
	0, 10, 16, 26, 18, 24, 16, 18, 22, 22, 26,
}

// numErrorCorrectionBlocks[v] is the number of error correction blocks
// for version v at error correction level M (index 0 unused).
var numErrorCorrectionBlocks = [maxVersion + 1]int{
	0, 1, 1, 1, 2, 2, 4, 4, 4, 5, 5,
}

// EncodePNG renders the given text (UTF-8) as a QR Code and returns it
// as a PNG byte string. The image includes the standard quiet zone.
func EncodePNG(text string) ([]byte, error) {
	c, err := encode(text)
	if err != nil {
		return nil, err
	}
	img := c.render(8, 4)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("png encode: %w", err)
	}
	return buf.Bytes(), nil
}

// EncodeDataURI renders the given text as a QR Code and returns a
// data URI suitable for use as an <img> src.
func EncodeDataURI(text string) (string, error) {
	pngBytes, err := EncodePNG(text)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes), nil
}

// code is a QR Code symbol in progress.
type code struct {
	version    int
	size       int
	mask       int
	modules    [][]bool // [y][x]; true = dark
	isFunction [][]bool // [y][x]; true = function module (not masked)
}

// encode builds a QR Code for the text using the smallest version that fits.
func encode(text string) (*code, error) {
	data := []byte(text)

	// Find the smallest version that fits the byte-mode bit stream.
	var version int
	var usedBits int
	for version = minVersion; version <= maxVersion; version++ {
		countBits := 8
		if len(data) >= 256 {
			countBits = 16
		}
		usedBits = 4 + countBits + len(data)*8
		if usedBits <= getNumDataCodewords(version)*8 {
			break
		}
	}
	if version > maxVersion {
		return nil, fmt.Errorf("qrcode: data too long (%d bytes, max %d)", len(data), getNumDataCodewords(maxVersion)-2)
	}

	// Build the bit stream (byte mode).
	bb := &bitBuffer{}
	bb.appendInt(0x4, 4) // Byte mode indicator
	if len(data) >= 256 {
		bb.appendInt(uint32(len(data)), 16)
	} else {
		bb.appendInt(uint32(len(data)), 8)
	}
	for _, b := range data {
		bb.appendInt(uint32(b), 8)
	}

	// Terminate and pad up to a byte boundary.
	capacity := getNumDataCodewords(version) * 8
	if bb.len() > capacity {
		return nil, fmt.Errorf("qrcode: data too long for version %d", version)
	}
	bb.appendInt(0, minInt(4, capacity-bb.len()))
	bb.appendInt(0, (8-bb.len()%8)%8)

	// Pad with alternating bytes until capacity.
	for pad := 0xEC; bb.len() < capacity; pad ^= 0xEC ^ 0x11 {
		bb.appendInt(uint32(pad), 8)
	}

	// Pack bits into bytes, big-endian.
	dataCodewords := make([]byte, bb.len()/8)
	for i, bit := range bb.bits {
		if bit {
			dataCodewords[i>>3] |= 1 << uint(7-(i&7))
		}
	}

	return newCode(version, dataCodewords)
}

// newCode builds the symbol matrix for the given version and data codewords.
func newCode(version int, dataCodewords []byte) (*code, error) {
	if version < minVersion || version > maxVersion {
		return nil, fmt.Errorf("qrcode: version %d out of range", version)
	}
	if len(dataCodewords) != getNumDataCodewords(version) {
		return nil, fmt.Errorf("qrcode: invalid data codewords length %d", len(dataCodewords))
	}

	c := &code{
		version:    version,
		size:       version*4 + 17,
		modules:    make([][]bool, version*4+17),
		isFunction: make([][]bool, version*4+17),
	}
	for y := range c.modules {
		c.modules[y] = make([]bool, c.size)
		c.isFunction[y] = make([]bool, c.size)
	}

	c.drawFunctionPatterns()
	all := addEccAndInterleave(dataCodewords, version)
	c.drawCodewords(all)

	// Choose the mask with the lowest penalty.
	bestMask := 0
	minPenalty := math.MaxInt
	for m := 0; m < 8; m++ {
		c.applyMask(m)
		c.drawFormatBits(m)
		p := c.penalty()
		if p < minPenalty {
			bestMask = m
			minPenalty = p
		}
		c.applyMask(m) // undo
	}
	c.applyMask(bestMask)
	c.drawFormatBits(bestMask)
	c.mask = bestMask

	return c, nil
}

// getNumDataCodewords returns the number of 8-bit data codewords for a
// version at error correction level M.
func getNumDataCodewords(version int) int {
	return getNumRawDataModules(version)/8 -
		eccCodecPerBlock[version]*numErrorCorrectionBlocks[version]
}

// getNumRawDataModules returns the number of data modules (including
// remainder bits) available in a version.
func getNumRawDataModules(version int) int {
	result := (16*version + 128) * version + 64
	if version >= 2 {
		numAlign := version/7 + 2
		result -= (25*numAlign - 10) * numAlign
		result += 55
		if version >= 7 {
			result -= 36
		}
	}
	return result
}

// drawFunctionPatterns draws and marks all function modules.
func (c *code) drawFunctionPatterns() {
	// Timing patterns.
	for i := 0; i < c.size; i++ {
		c.setFunctionModule(6, i, i%2 == 0)
		c.setFunctionModule(i, 6, i%2 == 0)
	}

	// Finder patterns (three corners; overwrites some timing modules).
	c.drawFinderPattern(3, 3)
	c.drawFinderPattern(c.size-4, 3)
	c.drawFinderPattern(3, c.size-4)

	// Alignment patterns.
	pos := c.alignmentPatternPositions()
	for i := 0; i < len(pos); i++ {
		for j := 0; j < len(pos); j++ {
			if !(i == 0 && j == 0 || i == 0 && j == len(pos)-1 || i == len(pos)-1 && j == 0) {
				c.drawAlignmentPattern(pos[i], pos[j])
			}
		}
	}

	// Dummy format bits + version info reserve those areas before data placement.
	c.drawFormatBits(0)
	c.drawVersion()
}

// drawFormatBits draws both copies of the format bits for the given mask.
func (c *code) drawFormatBits(mask int) {
	data := mask // error correction level M has formatBits == 0
	rem := data
	for i := 0; i < 10; i++ {
		rem = (rem << 1) ^ (rem>>9)*0x537
	}
	bits := (data<<10 | rem) ^ 0x5412

	// First copy (around top-left finder).
	for i := 0; i <= 5; i++ {
		c.setFunctionModule(8, i, getBit(bits, i))
	}
	c.setFunctionModule(8, 7, getBit(bits, 6))
	c.setFunctionModule(8, 8, getBit(bits, 7))
	c.setFunctionModule(7, 8, getBit(bits, 8))
	for i := 9; i < 15; i++ {
		c.setFunctionModule(14-i, 8, getBit(bits, i))
	}

	// Second copy (split into top-right and bottom-left).
	for i := 0; i < 8; i++ {
		c.setFunctionModule(c.size-1-i, 8, getBit(bits, i))
	}
	for i := 8; i < 15; i++ {
		c.setFunctionModule(8, c.size-15+i, getBit(bits, i))
	}
	c.setFunctionModule(8, c.size-8, true) // always dark
}

// drawVersion draws both copies of the version info bits (for version >= 7).
func (c *code) drawVersion() {
	if c.version < 7 {
		return
	}

	rem := c.version
	for i := 0; i < 12; i++ {
		rem = (rem << 1) ^ (rem>>11)*0x1F25
	}
	bits := c.version<<12 | rem

	for i := 0; i < 18; i++ {
		dark := getBit(bits, i)
		a := c.size - 11 + i%3
		b := i / 3
		c.setFunctionModule(a, b, dark)
		c.setFunctionModule(b, a, dark)
	}
}

// drawFinderPattern draws a 9x9 finder pattern including its border
// separator, centered at (x, y). Modules may be out of bounds.
func (c *code) drawFinderPattern(x, y int) {
	for dy := -4; dy <= 4; dy++ {
		for dx := -4; dx <= 4; dx++ {
			dist := maxInt(absInt(dx), absInt(dy))
			xx, yy := x+dx, y+dy
			if 0 <= xx && xx < c.size && 0 <= yy && yy < c.size {
				c.setFunctionModule(xx, yy, dist != 2 && dist != 4)
			}
		}
	}
}

// drawAlignmentPattern draws a 5x5 alignment pattern centered at (x, y).
func (c *code) drawAlignmentPattern(x, y int) {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			c.setFunctionModule(x+dx, y+dy, maxInt(absInt(dx), absInt(dy)) != 1)
		}
	}
}

// setFunctionModule sets a module and marks it as a function module.
func (c *code) setFunctionModule(x, y int, dark bool) {
	c.modules[y][x] = dark
	c.isFunction[y][x] = true
}

// alignmentPatternPositions returns the ascending list of alignment
// pattern positions for the version.
func (c *code) alignmentPatternPositions() []int {
	if c.version == 1 {
		return []int{}
	}
	numAlign := c.version/7 + 2
	step := (c.version*8+numAlign*3+5)/(numAlign*4-4) * 2
	result := []int{6}
	for pos := c.size - 7; len(result) < numAlign; pos -= step {
		result = append(append([]int{}, result[0:1]...), append([]int{pos}, result[1:]...)...)
	}
	return result
}

// addEccAndInterleave splits the data into blocks, appends error
// correction codewords, and interleaves them into a single sequence.
func addEccAndInterleave(data []byte, version int) []byte {
	numBlocks := numErrorCorrectionBlocks[version]
	blockEccLen := eccCodecPerBlock[version]
	rawCodewords := getNumRawDataModules(version) / 8
	numShortBlocks := numBlocks - rawCodewords%numBlocks
	shortBlockLen := rawCodewords / numBlocks

	divisor := rsComputeDivisor(blockEccLen)

	var blocks [][]byte
	k := 0
	for i := 0; i < numBlocks; i++ {
		datLen := shortBlockLen - blockEccLen
		if i >= numShortBlocks {
			datLen++
		}
		dat := data[k : k+datLen]
		k += datLen
		ecc := rsComputeRemainder(dat, divisor)

		block := append([]byte(nil), dat...)
		if i < numShortBlocks {
			block = append(block, 0) // padding byte to equalize block length
		}
		block = append(block, ecc...)
		blocks = append(blocks, block)
	}

	var result []byte
	for i := 0; i < len(blocks[0]); i++ {
		for j, block := range blocks {
			if i != shortBlockLen-blockEccLen || j >= numShortBlocks {
				result = append(result, block[i])
			}
		}
	}
	return result
}

// drawCodewords places the codeword bit stream into the data area using
// the standard zigzag scan, skipping function modules.
func (c *code) drawCodewords(data []byte) {
	i := 0
	for right := c.size - 1; right >= 1; right -= 2 {
		if right == 6 {
			right = 5
		}
		for vert := 0; vert < c.size; vert++ {
			for j := 0; j < 2; j++ {
				x := right - j
				upward := (right+1)&2 == 0
				y := vert
				if upward {
					y = c.size - 1 - vert
				}
				if !c.isFunction[y][x] && i < len(data)*8 {
					c.modules[y][x] = getBit(int(data[i>>3]), 7-(i&7))
					i++
				}
			}
		}
	}
}

// applyMask XORs the codeword modules with the given mask pattern.
// Calling applyMask twice with the same mask undoes the operation.
func (c *code) applyMask(mask int) {
	for y := 0; y < c.size; y++ {
		for x := 0; x < c.size; x++ {
			var invert bool
			switch mask {
			case 0:
				invert = (x+y)%2 == 0
			case 1:
				invert = y%2 == 0
			case 2:
				invert = x%3 == 0
			case 3:
				invert = (x+y)%3 == 0
			case 4:
				invert = (x/3+y/2)%2 == 0
			case 5:
				invert = x*y%2+x*y%3 == 0
			case 6:
				invert = (x*y%2+x*y%3)%2 == 0
			case 7:
				invert = ((x+y)%2+x*y%3)%2 == 0
			}
			if !c.isFunction[y][x] && invert {
				c.modules[y][x] = !c.modules[y][x]
			}
		}
	}
}

// penalty computes the mask penalty score of the current modules.
func (c *code) penalty() int {
	result := 0

	// Rows: adjacent same-color runs + finder-like patterns.
	for y := 0; y < c.size; y++ {
		runColor := false
		runX := 0
		var rh [7]int
		for x := 0; x < c.size; x++ {
			if c.modules[y][x] == runColor {
				runX++
				if runX == 5 {
					result += 3
				} else if runX > 5 {
					result++
				}
			} else {
				addHistory(runX, &rh, c.size)
				if !runColor {
					result += countPatterns(&rh, c.size) * 40
				}
				runColor = c.modules[y][x]
				runX = 1
			}
		}
		result += terminateAndCount(runColor, runX, &rh, c.size) * 40
	}

	// Columns: adjacent same-color runs + finder-like patterns.
	for x := 0; x < c.size; x++ {
		runColor := false
		runY := 0
		var rh [7]int
		for y := 0; y < c.size; y++ {
			if c.modules[y][x] == runColor {
				runY++
				if runY == 5 {
					result += 3
				} else if runY > 5 {
					result++
				}
			} else {
				addHistory(runY, &rh, c.size)
				if !runColor {
					result += countPatterns(&rh, c.size) * 40
				}
				runColor = c.modules[y][x]
				runY = 1
			}
		}
		result += terminateAndCount(runColor, runY, &rh, c.size) * 40
	}

	// 2x2 blocks of the same color.
	for y := 0; y < c.size-1; y++ {
		for x := 0; x < c.size-1; x++ {
			col := c.modules[y][x]
			if col == c.modules[y][x+1] && col == c.modules[y+1][x] && col == c.modules[y+1][x+1] {
				result += 3
			}
		}
	}

	// Balance of dark and light modules.
	dark := 0
	for y := 0; y < c.size; y++ {
		for x := 0; x < c.size; x++ {
			if c.modules[y][x] {
				dark++
			}
		}
	}
	total := c.size * c.size
	k := int(math.Ceil(math.Abs(float64(dark*20-total*10))/float64(total))) - 1
	if k < 0 {
		k = 0
	}
	result += k * 10

	return result
}

// countPatterns counts finder-like patterns in the run history.
func countPatterns(rh *[7]int, size int) int {
	n := rh[1]
	if n <= 0 {
		return 0
	}
	core := rh[2] == n && rh[3] == n*3 && rh[4] == n && rh[5] == n
	r := 0
	if core && rh[0] >= n*4 && rh[6] >= n {
		r++
	}
	if core && rh[6] >= n*4 && rh[0] >= n {
		r++
	}
	_ = size
	return r
}

// terminateAndCount finalizes the run history at end of a line.
func terminateAndCount(runColor bool, runLen int, rh *[7]int, size int) int {
	if runColor {
		addHistory(runLen, rh, size)
		runLen = 0
	}
	runLen += size
	addHistory(runLen, rh, size)
	return countPatterns(rh, size)
}

// addHistory pushes a run length to the front of the history, dropping
// the last value.
func addHistory(runLen int, rh *[7]int, size int) {
	if rh[0] == 0 {
		runLen += size
	}
	copy(rh[1:], rh[:6])
	rh[0] = runLen
}

// --- Galois field GF(2^8/0x11D) ---

var gfExp, gfLog [256]byte

func init() {
	x := 1
	for i := 0; i < 255; i++ {
		gfExp[i] = byte(x)
		gfLog[x] = byte(i)
		x <<= 1
		if x&0x100 != 0 {
			x ^= 0x11D
		}
	}
}

func gfMul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[(int(gfLog[a])+int(gfLog[b]))%255]
}

// rsComputeDivisor returns the Reed-Solomon generator polynomial for the
// given degree, stored highest-to-lowest and excluding the leading term.
func rsComputeDivisor(degree int) []byte {
	result := make([]byte, degree)
	result[degree-1] = 1
	root := byte(1)
	for i := 0; i < degree; i++ {
		for j := 0; j < len(result); j++ {
			result[j] = gfMul(result[j], root)
			if j+1 < len(result) {
				result[j] ^= result[j+1]
			}
		}
		root = gfMul(root, 0x02)
	}
	return result
}

// rsComputeRemainder returns the error correction codewords for the given
// data, computed as the remainder of the polynomial division.
func rsComputeRemainder(data, divisor []byte) []byte {
	result := make([]byte, len(divisor))
	for _, b := range data {
		factor := b ^ result[0]
		for i := 0; i+1 < len(result); i++ {
			result[i] = result[i+1]
		}
		result[len(result)-1] = 0
		for i, coef := range divisor {
			result[i] ^= gfMul(coef, factor)
		}
	}
	return result
}

// --- Bit buffer ---

type bitBuffer struct {
	bits []bool
}

func (b *bitBuffer) len() int { return len(b.bits) }

func (b *bitBuffer) appendInt(val uint32, n int) {
	for i := n - 1; i >= 0; i-- {
		b.bits = append(b.bits, (val>>uint(i))&1 == 1)
	}
}

// --- Helpers ---

func getBit(x, i int) bool {
	return (x>>uint(i))&1 != 0
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func absInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// render produces an image of the symbol with the given scale and quiet
// zone (in modules).
func (c *code) render(scale, border int) image.Image {
	dim := (c.size + border*2) * scale
	img := image.NewGray(image.Rect(0, 0, dim, dim))
	for y := 0; y < dim; y++ {
		for x := 0; x < dim; x++ {
			img.SetGray(x, y, color.Gray{Y: 255})
		}
	}
	for y := 0; y < c.size; y++ {
		for x := 0; x < c.size; x++ {
			if !c.modules[y][x] {
				continue
			}
			for dy := 0; dy < scale; dy++ {
				for dx := 0; dx < scale; dx++ {
					px := (x+border)*scale + dx
					py := (y+border)*scale + dy
					img.SetGray(px, py, color.Gray{Y: 0})
				}
			}
		}
	}
	return img
}