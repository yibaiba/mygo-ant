package kit

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestAvatarTextScaleAndSplit(t *testing.T) {
	if s := avatarTextScale(20, 32, 4); s != 1 {
		t.Fatalf("short text scaled %v", s)
	}
	if s := avatarTextScale(48, 32, 4); s != 0.5 {
		t.Fatalf("48 wide in 24 room: %v", s)
	}
	items := make([]AvatarItem, 5)
	shown, rest := splitAvatars(items, 3)
	if len(shown) != 3 || len(rest) != 2 {
		t.Fatalf("split %d/%d", len(shown), len(rest))
	}
	if shown, rest = splitAvatars(items, 0); len(shown) != 5 || rest != nil {
		t.Fatal("no max shows all")
	}
}

func TestBadgeText(t *testing.T) {
	cases := []struct {
		p    BadgeProps
		text string
		show bool
	}{
		{BadgeProps{Count: 5}, "5", true},
		{BadgeProps{Count: 100}, "99+", true},
		{BadgeProps{Count: 1000, Overflow: 999}, "999+", true},
		{BadgeProps{Count: 0}, "", false},
		{BadgeProps{Count: 0, ShowZero: true}, "0", true},
		{BadgeProps{Count: 3, Dot: true}, "", true},
		{BadgeProps{Dot: true}, "", false},
		{BadgeProps{Text: "NEW"}, "NEW", true},
	}
	for _, k := range cases {
		text, show := badgeText(k.p)
		if text != k.text || show != k.show {
			t.Errorf("%+v: %q %v, want %q %v", k.p, text, show, k.text, k.show)
		}
	}
	if r := rippleAt(time.UnixMilli(600)); r != 0.5 {
		t.Fatalf("ripple at 600ms %v", r)
	}
}

func TestFormatStatistic(t *testing.T) {
	cases := []struct {
		v           float64
		prec        int
		group       string
		whole, frac string
	}{
		{112893, 0, "", "112,893", ""},
		{112893.456, 2, "", "112,893", ".46"},
		{-1234567.5, 0, "", "-1,234,567", ".5"},
		{999, 0, "", "999", ""},
		{1234, 0, "-", "1234", ""},
		{1234, 1, " ", "1 234", ".0"},
	}
	for _, k := range cases {
		w, f := formatStatistic(k.v, k.prec, k.group)
		if w != k.whole || f != k.frac {
			t.Errorf("%v: %q%q, want %q%q", k.v, w, f, k.whole, k.frac)
		}
	}
}

func TestFormatCountdown(t *testing.T) {
	d := 2*24*time.Hour + 3*time.Hour + 4*time.Minute + 5*time.Second + 678*time.Millisecond
	cases := map[string]string{
		"":                "51:04:05",
		"HH:mm:ss":        "51:04:05",
		"D 天 H 时 m 分 s 秒": "2 天 3 时 4 分 5 秒",
		"HH:mm:ss:SSS":    "51:04:05:678",
		"[还剩] mm:ss":      "还剩 3064:05",
		"DD[d] HH[h]":     "02d 03h",
	}
	for f, want := range cases {
		if got := formatCountdown(d, f); got != want {
			t.Errorf("%q: %q, want %q", f, got, want)
		}
	}
	if got := formatCountdown(-time.Second, "mm:ss"); got != "00:00" {
		t.Fatalf("past deadline %q", got)
	}
}

func TestTimelineRows(t *testing.T) {
	items := []TimelineItem{{Content: "a"}, {Content: "b"}, {Content: "c"}}
	rows := timelineRows(items, TimelineProps{Mode: TimelineAlternate, Pending: "记录中"})
	if len(rows) != 4 || !rows[3].pending {
		t.Fatalf("rows %+v", rows)
	}
	if !rows[0].contentStart || rows[1].contentStart || !rows[2].contentStart {
		t.Fatal("alternate sides should start with content after the axis")
	}
	if rows[0].tail != 1 || rows[2].tail != 2 || rows[3].tail != 0 {
		t.Fatalf("tails %d %d %d", rows[0].tail, rows[2].tail, rows[3].tail)
	}
	rev := timelineRows(items, TimelineProps{Reverse: true, Pending: "记录中"})
	if !rev[0].pending || rev[0].tail != 2 || rev[1].item.Content != "c" {
		t.Fatalf("reverse %+v", rev[:2])
	}
	right := timelineRows(items, TimelineProps{Mode: TimelineRight})
	if right[0].contentStart {
		t.Fatal("right mode keeps content before the axis")
	}
}

func TestListPageAndCalendarHelpers(t *testing.T) {
	if f, to := listPage(23, 3, 10); f != 20 || to != 23 {
		t.Fatalf("page 3: %d..%d", f, to)
	}
	if f, to := listPage(5, 1, 0); f != 0 || to != 5 {
		t.Fatal("unpaged")
	}
	ys := calendarYears(2026, [2]time.Time{})
	if ys[0] != 2016 || ys[len(ys)-1] != 2035 {
		t.Fatalf("years %v", ys)
	}
	lo, hi := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)
	p := CalendarProps{ValidRange: [2]time.Time{lo, hi}, DisabledDate: func(d time.Time) bool { return d.Weekday() == time.Sunday }}
	if ys := calendarYears(2026, p.ValidRange); len(ys) != 2 {
		t.Fatalf("valid years %v", ys)
	}
	if !calendarOff(p, lo.AddDate(0, 0, -1)) || calendarOff(p, lo.AddDate(0, 0, 1)) || !calendarOff(p, hi.AddDate(0, 0, 1)) {
		t.Fatal("valid range")
	}
	if !calendarOff(p, time.Date(2026, 3, 8, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("sunday disabled")
	}
	if daysIn(time.Date(2028, 2, 1, 0, 0, 0, 0, time.UTC)) != 29 {
		t.Fatal("leap february")
	}
}

func TestCarouselRel(t *testing.T) {
	cases := []struct {
		i, n int
		pos  float32
		want float32
	}{
		{0, 3, 0, 0}, {1, 3, 0, 1}, {2, 3, 0, -1},
		{0, 3, 2.5, 0.5}, {2, 3, 2.5, -0.5},
		{0, 2, 0.5, -0.5}, {1, 2, 0.5, 0.5},
		{2, 3, -0.5, -0.5}, {0, 3, -0.5, 0.5},
	}
	for _, k := range cases {
		if got := carouselRel(k.i, k.n, k.pos); got != k.want {
			t.Errorf("slide %d of %d at %v: %v, want %v", k.i, k.n, k.pos, got, k.want)
		}
	}
	if wrapIndex(-1, 3) != 2 || wrapIndex(4, 3) != 1 {
		t.Fatal("wrapIndex")
	}
}

func TestImageViewZoom(t *testing.T) {
	if w, h := imageBox(200, 0, 400, 300); w != 200 || h != 150 {
		t.Fatalf("width only %v×%v", w, h)
	}
	if w, h := imageBox(0, 0, 40, 30); w != 40 || h != 30 {
		t.Fatal("natural size")
	}
	v := imageView{scale: 1}
	v.zoomAt(2, 100, 50)
	if v.scale != 2 || v.x != -100 || v.y != -50 {
		t.Fatalf("zoom at point: %+v", v)
	}
	v.zoomAt(0.1, 0, 0)
	if v.scale != 1 || v.x != 0 || v.y != 0 {
		t.Fatalf("zoom out recenters at 1: %+v", v)
	}
	v.zoomAt(100, 0, 0)
	if v.scale != 50 {
		t.Fatalf("max 50: %v", v.scale)
	}
	if wheelRatio(-1, false) != 1.5 || wheelRatio(1, false) != 1/1.5 || wheelRatio(-20, true) != 1.1 {
		t.Fatal("wheel ratios")
	}
}

func TestReedSolomonVector(t *testing.T) {
	// HELLO WORLD at 1-M in alphanumeric mode (thonky.com's worked example).
	data := []byte{32, 91, 11, 120, 209, 114, 220, 77, 67, 64, 236, 17, 236, 17, 236, 17}
	want := []byte{196, 35, 39, 119, 235, 215, 231, 226, 93, 23}
	got := rsRemainder(data, rsDivisor(10))
	if string(got) != string(want) {
		t.Fatalf("ecc %v, want %v", got, want)
	}
}

func TestQRFormatAndVersion(t *testing.T) {
	if got := qrFormatBits(QRLevelL, 4); got != 0b110011000101111 {
		t.Fatalf("L/4 format %015b", got)
	}
	if got := qrFormatBits(QRLevelM, 0); got != 0b101010000010010 {
		t.Fatalf("M/0 format %015b", got)
	}
	for _, k := range []struct {
		n, v int
		l    QRLevel
	}{{14, 1, QRLevelM}, {15, 2, QRLevelM}, {17, 1, QRLevelL}, {2953, 40, QRLevelL}} {
		if v, err := qrVersion(k.n, k.l); err != nil || v != k.v {
			t.Errorf("%d bytes: version %d %v, want %d", k.n, v, err, k.v)
		}
	}
	if _, err := qrVersion(2954, QRLevelL); err == nil {
		t.Fatal("2954 bytes fit no version")
	}
}

func TestQRMatrixMatchesReference(t *testing.T) {
	// HELLO WORLD in byte mode at M with mask 2, from the qrcode reference
	// encoder (python-qrcode 8.2, mask_pattern=2).
	want := []string{
		"#######.....#.#######",
		"#.....#..#.#..#.....#",
		"#.###.#.###.#.#.###.#",
		"#.###.#.#.#.#.#.###.#",
		"#.###.#.#.#.#.#.###.#",
		"#.....#.##.#..#.....#",
		"#######.#.#.#.#######",
		"........#.#..........",
		"#.#####...##..#####..",
		".##....#.#######.##..",
		"#.##..##....###..###.",
		".###.#..######..###..",
		"#...#.##.##.##....#.#",
		"........###.#....#...",
		"#######..#.#..#...##.",
		"#.....#.###..#.#.####",
		"#.###.#.#..#...#..#.#",
		"#.###.#.#...######...",
		"#.###.#.##..#..#..#..",
		"#.....#...#.##..###..",
		"#######.#.###...#.##.",
	}
	m, err := encodeQRMode("HELLO WORLD", QRLevelM, 2, qrModeByte)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for y := 0; y < m.size; y++ {
		var b strings.Builder
		for x := 0; x < m.size; x++ {
			b.WriteByte(pick[byte](m.at(x, y), '#', '.'))
		}
		got = append(got, b.String())
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("matrix\n%s", strings.Join(got, "\n"))
	}
	big, err := encodeQR(strings.Repeat("会议", 200), QRLevelH, -1)
	if err != nil || big.size != 17+4*qrMustVersion(t, 1200, QRLevelH) {
		t.Fatalf("1200 bytes at H: %v", err)
	}
}

func qrMustVersion(t *testing.T, n int, l QRLevel) int {
	v, err := qrVersion(n, l)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestQRPicksNumericAndAlphanumeric(t *testing.T) {
	if qrPickMode("1234567890") != qrModeNumeric {
		t.Fatal("digits should be numeric")
	}
	if qrPickMode("HELLO WORLD") != qrModeAlphanumeric {
		t.Fatal("HELLO WORLD should be alphanumeric")
	}
	if qrPickMode("https://example.com") != qrModeByte {
		t.Fatal("lowercase URLs stay byte mode")
	}
	digits := strings.Repeat("1", 100)
	numV, err := qrVersionFor(len(digits), QRLevelL, qrModeNumeric)
	if err != nil {
		t.Fatal(err)
	}
	byteV, err := qrVersionFor(len(digits), QRLevelL, qrModeByte)
	if err != nil {
		t.Fatal(err)
	}
	if numV >= byteV {
		t.Fatalf("numeric version %d should beat byte %d for 100 digits", numV, byteV)
	}
	m, err := encodeQR(digits, QRLevelL, -1)
	if err != nil {
		t.Fatal(err)
	}
	want := 17 + 4*numV
	if m.size != want {
		t.Fatalf("auto encode size %d, want %d (numeric v%d)", m.size, want, numV)
	}
}

func TestTourMaskAndTagColors(t *testing.T) {
	hole := tourHole(ui.Rect{X: 100, Y: 50, W: 80, H: 30}, 6)
	if hole != (ui.Rect{X: 94, Y: 44, W: 92, H: 42}) {
		t.Fatalf("hole %+v", hole)
	}
	bands := tourMask(hole, 400, 300)
	var area float32
	for _, b := range bands {
		area += b.W * b.H
	}
	if area+hole.W*hole.H != 400*300 {
		t.Fatalf("mask leaves %v uncovered", 400*300-area-hole.W*hole.H)
	}
	bg, edge, fg := tagColors(Work, TagProps{Preset: "magenta"})
	pal := AntPalette(ui.Hex("#eb2f96"))
	if bg != pal[0] || edge != pal[2] || fg != pal[6] {
		t.Fatal("preset tag shades 1/3/7")
	}
	if _, _, fg := tagColors(Work, TagProps{Color: ui.Hex("#108ee9")}); fg != ui.Hex("#FFFFFF") {
		t.Fatal("custom color tags are white on the color")
	}
	if toneColor(Work, "processing", ui.Color{}) != Work.Primary || toneColor(Work, "nope", Work.Error) != Work.Error {
		t.Fatal("toneColor")
	}
}
