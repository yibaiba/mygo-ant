package kit

import (
	"errors"

	"github.com/egoist/mygo/ui"
)

// QRLevel is a QR code's error correction level.
type QRLevel uint8

// QR error correction levels, recovering about 15%, 7%, 25% and 30% of
// the code; M, Ant's default, is the zero value.
const (
	QRLevelM QRLevel = iota
	QRLevelL
	QRLevelQ
	QRLevelH
)

// row is the level's row of qrBlocks; bits its format information bits.
func (l QRLevel) row() int  { return [4]int{1, 0, 2, 3}[l] }
func (l QRLevel) bits() int { return [4]int{0, 1, 3, 2}[l] }

// errQRTooLong is the error for text that no QR version holds.
var errQRTooLong = errors.New("kit: text too long for a QR code")

// qrMatrix is an encoded QR code: size×size modules, true for dark.
type qrMatrix struct {
	size     int
	dark     []bool
	function []bool
}

func (m *qrMatrix) at(x, y int) bool { return m.dark[y*m.size+x] }

func (m *qrMatrix) set(x, y int, dark, function bool) {
	m.dark[y*m.size+x] = dark
	if function {
		m.function[y*m.size+x] = true
	}
}

// qrCapacity is the data codewords of a version at a level.
func qrCapacity(version int, level QRLevel) int {
	b := qrBlocks[version-1][level.row()]
	return int(b[1])*int(b[2]) + int(b[3])*int(b[4])
}

// qrEncMode is a QR data encoding. Auto picks numeric, alphanumeric or
// byte, as Ant's QRCode / the ISO default.
type qrEncMode uint8

const (
	qrModeAuto qrEncMode = iota
	qrModeNumeric
	qrModeAlphanumeric
	qrModeByte
)

const qrAlphanumeric = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"

func qrAlphaVal(c byte) int {
	for i := 0; i < len(qrAlphanumeric); i++ {
		if qrAlphanumeric[i] == c {
			return i
		}
	}
	return -1
}

func qrPickMode(text string) qrEncMode {
	if text == "" {
		return qrModeByte
	}
	num, alpha := true, true
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c < '0' || c > '9' {
			num = false
		}
		if qrAlphaVal(c) < 0 {
			alpha = false
		}
	}
	switch {
	case num:
		return qrModeNumeric
	case alpha:
		return qrModeAlphanumeric
	}
	return qrModeByte
}

func qrCountBits(mode qrEncMode, version int) int {
	switch mode {
	case qrModeNumeric:
		return pick(version <= 9, 10, pick(version <= 26, 12, 14))
	case qrModeAlphanumeric:
		return pick(version <= 9, 9, pick(version <= 26, 11, 13))
	}
	return pick(version <= 9, 8, 16)
}

func qrPayloadBits(n int, mode qrEncMode) int {
	switch mode {
	case qrModeNumeric:
		return 10*(n/3) + []int{0, 4, 7}[n%3]
	case qrModeAlphanumeric:
		return 11*(n/2) + 6*(n%2)
	}
	return 8 * n
}

func qrBitLen(n, version int, mode qrEncMode) int {
	return 4 + qrCountBits(mode, version) + qrPayloadBits(n, mode)
}

// qrVersionFor is the smallest version holding n characters in mode.
func qrVersionFor(n int, level QRLevel, mode qrEncMode) (int, error) {
	if mode == qrModeAuto {
		mode = qrModeByte
	}
	for v := 1; v <= 40; v++ {
		if qrBitLen(n, v, mode) <= qrCapacity(v, level)*8 {
			return v, nil
		}
	}
	return 0, errQRTooLong
}

// qrVersion is the smallest version holding n bytes in byte mode.
func qrVersion(n int, level QRLevel) (int, error) {
	return qrVersionFor(n, level, qrModeByte)
}

func qrDataMode(data []byte, version int, level QRLevel, mode qrEncMode) []byte {
	var bits []bool
	put := func(v, n int) {
		for i := n - 1; i >= 0; i-- {
			bits = append(bits, v>>i&1 == 1)
		}
	}
	switch mode {
	case qrModeNumeric:
		put(0b0001, 4)
		put(len(data), qrCountBits(mode, version))
		for i := 0; i < len(data); {
			left := min(3, len(data)-i)
			v := 0
			for j := 0; j < left; j++ {
				v = v*10 + int(data[i+j]-'0')
			}
			put(v, []int{4, 7, 10}[left-1])
			i += left
		}
	case qrModeAlphanumeric:
		put(0b0010, 4)
		put(len(data), qrCountBits(mode, version))
		for i := 0; i < len(data); {
			if i+1 < len(data) {
				put(45*qrAlphaVal(data[i])+qrAlphaVal(data[i+1]), 11)
				i += 2
			} else {
				put(qrAlphaVal(data[i]), 6)
				i++
			}
		}
	default:
		put(0b0100, 4)
		put(len(data), qrCountBits(qrModeByte, version))
		for _, b := range data {
			put(int(b), 8)
		}
	}
	capBits := qrCapacity(version, level) * 8
	put(0, min(4, capBits-len(bits)))
	for len(bits)%8 != 0 {
		bits = append(bits, false)
	}
	out := make([]byte, 0, capBits/8)
	for i := 0; i < len(bits); i += 8 {
		var b byte
		for j := 0; j < 8; j++ {
			if bits[i+j] {
				b |= 1 << (7 - j)
			}
		}
		out = append(out, b)
	}
	for pad := byte(0xEC); len(out) < capBits/8; pad ^= 0xEC ^ 0x11 {
		out = append(out, pad)
	}
	return out
}

// qrData is the data codewords for text in byte mode, terminated and
// padded to the version's capacity.
func qrData(data []byte, version int, level QRLevel) []byte {
	return qrDataMode(data, version, level, qrModeByte)
}

// gfMul multiplies in GF(256) over the QR polynomial 0x11D.
func gfMul(x, y byte) byte {
	var z byte
	for i := 7; i >= 0; i-- {
		z = z<<1 ^ byte(int(z>>7)*0x1D)
		if y>>i&1 == 1 {
			z ^= x
		}
	}
	return z
}

// rsDivisor is the Reed-Solomon generator of degree n, highest term
// dropped.
func rsDivisor(n int) []byte {
	out := make([]byte, n)
	out[n-1] = 1
	root := byte(1)
	for i := 0; i < n; i++ {
		for j := range out {
			out[j] = gfMul(out[j], root)
			if j+1 < n {
				out[j] ^= out[j+1]
			}
		}
		root = gfMul(root, 2)
	}
	return out
}

// rsRemainder is the error correction codewords of data.
func rsRemainder(data, divisor []byte) []byte {
	out := make([]byte, len(divisor))
	for _, b := range data {
		f := b ^ out[0]
		copy(out, out[1:])
		out[len(out)-1] = 0
		for i, d := range divisor {
			out[i] ^= gfMul(d, f)
		}
	}
	return out
}

// qrCodewords splits data into the version's blocks, adds each block's
// error correction and interleaves them.
func qrCodewords(data []byte, version int, level QRLevel) []byte {
	b := qrBlocks[version-1][level.row()]
	ec := int(b[0])
	div := rsDivisor(ec)
	var blocks, eccs [][]byte
	at := 0
	for g := 0; g < 2; g++ {
		count, size := int(b[1+g*2]), int(b[2+g*2])
		for i := 0; i < count; i++ {
			blk := data[at : at+size]
			at += size
			blocks = append(blocks, blk)
			eccs = append(eccs, rsRemainder(blk, div))
		}
	}
	var out []byte
	for i := 0; ; i++ {
		took := false
		for _, blk := range blocks {
			if i < len(blk) {
				out = append(out, blk[i])
				took = true
			}
		}
		if !took {
			break
		}
	}
	for i := 0; i < ec; i++ {
		for _, e := range eccs {
			out = append(out, e[i])
		}
	}
	return out
}

func newQRMatrix(version int) *qrMatrix {
	size := version*4 + 17
	m := &qrMatrix{size: size, dark: make([]bool, size*size), function: make([]bool, size*size)}
	for i := 0; i < size; i++ {
		m.set(6, i, i%2 == 0, true)
		m.set(i, 6, i%2 == 0, true)
	}
	finder := func(cx, cy int) {
		for dy := -4; dy <= 4; dy++ {
			for dx := -4; dx <= 4; dx++ {
				x, y := cx+dx, cy+dy
				if x < 0 || y < 0 || x >= size || y >= size {
					continue
				}
				d := max(abs(dx), abs(dy))
				m.set(x, y, d != 2 && d != 4, true)
			}
		}
	}
	finder(3, 3)
	finder(size-4, 3)
	finder(3, size-4)
	al := qrAlign[version-1]
	for i, ax := range al {
		for j, ay := range al {
			if i == 0 && j == 0 || i == 0 && j == len(al)-1 || i == len(al)-1 && j == 0 {
				continue
			}
			for dy := -2; dy <= 2; dy++ {
				for dx := -2; dx <= 2; dx++ {
					m.set(ax+dx, ay+dy, max(abs(dx), abs(dy)) != 1, true)
				}
			}
		}
	}
	m.drawFormat(QRLevelM, 0)
	if version >= 7 {
		rem := version
		for i := 0; i < 12; i++ {
			rem = rem<<1 ^ (rem>>11)*0x1F25
		}
		bits := version<<12 | rem
		for i := 0; i < 18; i++ {
			dark := bits>>i&1 == 1
			a, b := size-11+i%3, i/3
			m.set(a, b, dark, true)
			m.set(b, a, dark, true)
		}
	}
	return m
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// qrFormatBits is the 15 format information bits of a level and mask.
func qrFormatBits(level QRLevel, mask int) int {
	data := level.bits()<<3 | mask
	rem := data
	for i := 0; i < 10; i++ {
		rem = rem<<1 ^ (rem>>9)*0x537
	}
	return (data<<10 | rem) ^ 0x5412
}

// drawFormat draws both copies of the format information, and the dark
// module.
func (m *qrMatrix) drawFormat(level QRLevel, mask int) {
	bits := qrFormatBits(level, mask)
	bit := func(i int) bool { return bits>>i&1 == 1 }
	for i := 0; i <= 5; i++ {
		m.set(8, i, bit(i), true)
	}
	m.set(8, 7, bit(6), true)
	m.set(8, 8, bit(7), true)
	m.set(7, 8, bit(8), true)
	for i := 9; i < 15; i++ {
		m.set(14-i, 8, bit(i), true)
	}
	for i := 0; i < 8; i++ {
		m.set(m.size-1-i, 8, bit(i), true)
	}
	for i := 8; i < 15; i++ {
		m.set(8, m.size-15+i, bit(i), true)
	}
	m.set(8, m.size-8, true, true)
}

// drawCodewords lays the codewords' bits in the two-column zigzag from
// the bottom right, around the function patterns.
func (m *qrMatrix) drawCodewords(data []byte) {
	i := 0
	for right := m.size - 1; right >= 1; right -= 2 {
		if right == 6 {
			right = 5
		}
		for vert := 0; vert < m.size; vert++ {
			for j := 0; j < 2; j++ {
				x := right - j
				y := vert
				if (right+1)&2 == 0 {
					y = m.size - 1 - vert
				}
				if !m.function[y*m.size+x] && i < len(data)*8 {
					m.dark[y*m.size+x] = data[i>>3]>>(7-i&7)&1 == 1
					i++
				}
			}
		}
	}
}

func qrMaskBit(mask, x, y int) bool {
	switch mask {
	case 0:
		return (x+y)%2 == 0
	case 1:
		return y%2 == 0
	case 2:
		return x%3 == 0
	case 3:
		return (x+y)%3 == 0
	case 4:
		return (x/3+y/2)%2 == 0
	case 5:
		return x*y%2+x*y%3 == 0
	case 6:
		return (x*y%2+x*y%3)%2 == 0
	}
	return ((x+y)%2+x*y%3)%2 == 0
}

func (m *qrMatrix) applyMask(mask int) {
	for y := 0; y < m.size; y++ {
		for x := 0; x < m.size; x++ {
			if !m.function[y*m.size+x] && qrMaskBit(mask, x, y) {
				m.dark[y*m.size+x] = !m.dark[y*m.size+x]
			}
		}
	}
}

// penalty scores the matrix by the standard's four rules; the mask with
// the lowest score is used.
func (m *qrMatrix) penalty() int {
	n := m.size
	score := 0
	line := func(get func(i int) bool) {
		run := 1
		for i := 1; i <= n; i++ {
			if i < n && get(i) == get(i-1) {
				run++
				continue
			}
			if run >= 5 {
				score += 3 + run - 5
			}
			run = 1
		}
		pat := []bool{true, false, true, true, true, false, true}
		for i := 0; i+7 <= n; i++ {
			match := true
			for k, v := range pat {
				if get(i+k) != v {
					match = false
					break
				}
			}
			if !match {
				continue
			}
			light := func(from, to int) bool {
				if from < 0 || to > n {
					return false
				}
				for k := from; k < to; k++ {
					if get(k) {
						return false
					}
				}
				return true
			}
			if light(i-4, i) || light(i+7, i+11) {
				score += 40
			}
		}
	}
	for y := 0; y < n; y++ {
		line(func(i int) bool { return m.at(i, y) })
	}
	for x := 0; x < n; x++ {
		line(func(i int) bool { return m.at(x, i) })
	}
	dark := 0
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			c := m.at(x, y)
			if c {
				dark++
			}
			if x < n-1 && y < n-1 && c == m.at(x+1, y) && c == m.at(x, y+1) && c == m.at(x+1, y+1) {
				score += 3
			}
		}
	}
	total := n * n
	k := (abs(dark*20-total*10)+total-1)/total - 1
	return score + k*10
}

// encodeQR encodes text at level in the tightest of numeric,
// alphanumeric and byte mode, in the smallest version, with the best
// mask, or a given mask when mask is 0 to 7.
func encodeQR(text string, level QRLevel, mask int) (*qrMatrix, error) {
	return encodeQRMode(text, level, mask, qrModeAuto)
}

// encodeQRMode encodes text in mode (Auto picks). Tests that pin a
// reference matrix pass qrModeByte so a mode change cannot move modules.
func encodeQRMode(text string, level QRLevel, mask int, mode qrEncMode) (*qrMatrix, error) {
	if mode == qrModeAuto {
		mode = qrPickMode(text)
	}
	data := []byte(text)
	version, err := qrVersionFor(len(data), level, mode)
	if err != nil {
		return nil, err
	}
	words := qrCodewords(qrDataMode(data, version, level, mode), version, level)
	best, bestScore := -1, 0
	var out *qrMatrix
	for k := 0; k < 8; k++ {
		if mask >= 0 && mask < 8 && k != mask {
			continue
		}
		m := newQRMatrix(version)
		m.drawCodewords(words)
		m.applyMask(k)
		m.drawFormat(level, k)
		s := m.penalty()
		if best < 0 || s < bestScore {
			best, bestScore, out = k, s, m
		}
	}
	return out, nil
}

// QRStatus is a QR code's state.
type QRStatus uint8

// QR code states.
const (
	QRActive QRStatus = iota
	// QRExpired covers the code with "二维码过期" and a refresh link.
	QRExpired
	// QRLoading covers it with a spinner.
	QRLoading
	// QRScanned covers it with "已扫描".
	QRScanned
)

// QRCodeProps are a QR code's options.
type QRCodeProps struct {
	// Size is the code's width and height; 160 when zero.
	Size  float32
	Level QRLevel
	// Color is the dark modules' (the text color by default), BgColor the
	// light ones' (transparent).
	Color, BgColor ui.Color
	// Icon sits in the middle, IconSize (40) across, the modules under it
	// cleared.
	Icon     ui.ImageSource
	IconSize float32
	// Borderless drops the frame.
	Borderless bool
	Status     QRStatus
}

// qrCache keeps a QR code's matrix while its text and level stay.
type qrCache struct {
	key   string
	level QRLevel
	m     *qrMatrix
	err   error
}

// QRCode is Ant's QRCode: value encoded in numeric, alphanumeric or
// byte mode (UTF-8) at Level,
// drawn in a framed square, with an optional icon and a status cover. It
// reports a click of the expired cover's refresh link.
func QRCode(c *ui.Context, value string, props ...QRCodeProps) bool {
	p := first(props)
	t := Of(c)
	size := pick(p.Size == 0, 160, p.Size)
	fg := p.Color
	if fg == (ui.Color{}) {
		fg = t.Text
	}
	frame := ui.Column(c).Shrink(0).AlignSelf(ui.Start).Background(p.BgColor)
	cache := ui.Local(frame, "qr", func() qrCache { return qrCache{} })
	if cache.m == nil && cache.err == nil || cache.key != value || cache.level != p.Level {
		cache.key, cache.level = value, p.Level
		cache.m, cache.err = encodeQR(value, p.Level, -1)
	}
	iconSize := pick(p.IconSize == 0, 40, p.IconSize)
	refresh := false
	if !p.Borderless {
		frame.Padding(12).Border(1, t.Split.Alpha(0.6).Over(t.Surface)).Radius(8)
	}
	frame.Label("二维码").Children(func() {
		ui.Box(c).Size(size, size).Children(func() {
			ui.Box(c).Size(size, size).Cache().Draw(func(pa *ui.Painter, r ui.Rect) {
				m := cache.m
				if m == nil {
					return
				}
				cell := r.W / float32(m.size)
				hole := func(x, y int) bool {
					if p.Icon == nil {
						return false
					}
					cx, cy := (float32(x)+0.5)*cell, (float32(y)+0.5)*cell
					half := iconSize/2 + cell/2
					return abs32(cx-r.W/2) < half && abs32(cy-r.H/2) < half
				}
				for y := 0; y < m.size; y++ {
					for x := 0; x < m.size; x++ {
						if m.at(x, y) && !hole(x, y) {
							pa.Fill(ui.Rect{X: r.X + float32(x)*cell, Y: r.Y + float32(y)*cell, W: cell + 0.3, H: cell + 0.3}, fg, 0)
						}
					}
				}
			})
			if p.Icon != nil {
				ui.Image(c, p.Icon).Size(iconSize, iconSize).Fit(ui.Contain).Absolute().Left((size - iconSize) / 2).Top((size - iconSize) / 2)
			}
			if cache.err != nil {
				qrCover(c, t, func() { ui.Text(c, L(c).QRTooLong).FontSize(t.FontSize).TextColor(t.Error) })
				return
			}
			switch p.Status {
			case QRExpired:
				qrCover(c, t, func() {
					ui.Text(c, L(c).QRExpired).FontSize(t.FontSize).TextColor(t.TextTertiary)
					refresh = Button(c, L(c).QRRefresh, ButtonProps{Type: Link})
				})
			case QRLoading:
				qrCover(c, t, func() { Spin(c) })
			case QRScanned:
				qrCover(c, t, func() {
					statusIcon(c, Success, 20, t.SuccessIcon)
					ui.Text(c, L(c).QRScan).FontSize(t.FontSize).TextColor(t.Text)
				})
			}
		})
	})
	return refresh
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func qrCover(c *ui.Context, t *Tokens, fn func()) {
	ui.Column(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Background(t.Surface.Alpha(0.96)).
		Center().Gap(4).Children(fn)
}
