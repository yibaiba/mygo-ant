package kit

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestSnapSlider(t *testing.T) {
	cases := []struct {
		v, lo, hi, step, want float64
	}{
		{13, 0, 100, 10, 10},
		{16, 0, 100, 10, 20},
		{-5, 0, 100, 10, 0},
		{120, 0, 100, 10, 100},
		{0.31, 0, 1, 0.1, 0.3},
		{97, 0, 98, 5, 95},
		{7, 2, 20, 4, 6},
	}
	for _, c := range cases {
		if got := snapSlider(c.v, c.lo, c.hi, c.step, nil, false); got != c.want {
			t.Errorf("snapSlider(%v,%v,%v,%v) = %v, want %v", c.v, c.lo, c.hi, c.step, got, c.want)
		}
	}
	marks := []SliderMark{{Value: 0}, {Value: 26}, {Value: 37}, {Value: 100}}
	if got := snapSlider(33, 0, 100, 1, marks, true); got != 37 {
		t.Fatalf("marks only snap = %v, want 37", got)
	}
	p := SliderProps{Marks: marks, MarksOnly: true}
	if got := sliderStep(26, 1, p); got != 37 {
		t.Fatalf("mark step up = %v", got)
	}
	if got := sliderStep(26, -1, p); got != 0 {
		t.Fatalf("mark step down = %v", got)
	}
	if got := sliderStep(100, 1, p); got != 100 {
		t.Fatalf("mark step past end = %v", got)
	}
	if got := sliderStep(50, 3, SliderProps{Step: 5}); got != 65 {
		t.Fatalf("step = %v", got)
	}
	if got := sliderFraction(25, 0, 200); got != 0.125 {
		t.Fatalf("fraction = %v", got)
	}
	if decimals(0.25) != 2 || decimals(5) != 0 {
		t.Fatal("decimals")
	}
}

func TestNearestHandle(t *testing.T) {
	if got := nearestHandle([]float64{20, 60}, 30); got != 0 {
		t.Fatalf("got %d", got)
	}
	if got := nearestHandle([]float64{20, 60}, 50); got != 1 {
		t.Fatalf("got %d", got)
	}
	// Two handles on one spot: a press past them takes the later one so
	// the range can open up.
	if got := nearestHandle([]float64{40, 40}, 70); got != 1 {
		t.Fatalf("equal, past: got %d", got)
	}
	if got := nearestHandle([]float64{40, 40}, 10); got != 0 {
		t.Fatalf("equal, before: got %d", got)
	}
}

func TestRateHalfAndClear(t *testing.T) {
	const size = 20
	if got := rateAt(-1, size, 5, true); got != 0 {
		t.Fatalf("before = %v", got)
	}
	if got := rateAt(5, size, 5, true); got != 0.5 {
		t.Fatalf("first half = %v", got)
	}
	if got := rateAt(15, size, 5, true); got != 1 {
		t.Fatalf("second half = %v", got)
	}
	if got := rateAt(5, size, 5, false); got != 1 {
		t.Fatalf("whole star = %v", got)
	}
	cell := float32(size + rateGap)
	if got := rateAt(2*cell+3, size, 5, true); got != 2.5 {
		t.Fatalf("third star half = %v", got)
	}
	if got := rateAt(99*cell, size, 5, true); got != 5 {
		t.Fatalf("past end = %v", got)
	}
	if rateClick(3, 3, true) != 0 || rateClick(3, 3, false) != 3 || rateClick(3, 4, true) != 4 {
		t.Fatal("rateClick")
	}
}

func TestOTPApply(t *testing.T) {
	code := otpSplit("12", 4)
	if !reflect.DeepEqual(code, []string{"1", "2", "", ""}) {
		t.Fatalf("split = %q", code)
	}
	code, next := otpApply(code, 2, "7")
	if code[2] != "7" || next != 3 {
		t.Fatalf("type = %q next %d", code, next)
	}
	// Typing over a filled box keeps the new character.
	code, next = otpApply(code, 0, "19")
	if code[0] != "9" || next != 1 {
		t.Fatalf("overtype = %q next %d", code, next)
	}
	code, next = otpApply(code, 1, "4567")
	if !reflect.DeepEqual(code, []string{"9", "4", "5", "6"}) || next != 3 {
		t.Fatalf("paste = %q next %d", code, next)
	}
	code, next = otpApply(code, 3, "")
	if code[3] != "" || next != 3 {
		t.Fatalf("clear = %q next %d", code, next)
	}
	if code, next = otpApply([]string{"", ""}, 1, "9"); code[1] != "9" || next != 1 {
		t.Fatalf("last box = %q next %d", code, next)
	}
}

func sampleTree() []TreeNode {
	return []TreeNode{
		{Key: "a", Title: "Alpha", Children: []TreeNode{
			{Key: "a1", Title: "Apple"},
			{Key: "a2", Title: "Avocado"},
			{Key: "a3", Title: "Locked", Disabled: true},
		}},
		{Key: "b", Title: "Beta", Children: []TreeNode{
			{Key: "b1", Title: "Banana", Children: []TreeNode{
				{Key: "b11", Title: "Plantain"},
				{Key: "b12", Title: "Cavendish"},
			}},
		}},
	}
}

func TestTreeCheckPropagation(t *testing.T) {
	nodes := sampleTree()
	got := conductCheck(nodes, nil, "a", true)
	if !reflect.DeepEqual(got, []string{"a", "a1", "a2"}) {
		t.Fatalf("check parent = %v", got)
	}
	got = conductCheck(nodes, got, "a1", false)
	if !reflect.DeepEqual(got, []string{"a2"}) {
		t.Fatalf("uncheck child = %v", got)
	}
	if half := halfChecked(nodes, got); !half["a"] || half["b"] {
		t.Fatalf("half = %v", half)
	}
	got = conductCheck(nodes, got, "b11", true)
	if half := halfChecked(nodes, got); !half["b"] || !half["b1"] {
		t.Fatalf("deep half = %v", half)
	}
	got = conductCheck(nodes, got, "b12", true)
	if !reflect.DeepEqual(got, []string{"a2", "b", "b1", "b11", "b12"}) {
		t.Fatalf("all children up = %v", got)
	}
	if leaves := leafKeys(nodes, got); !reflect.DeepEqual(leaves, []string{"a2", "b11", "b12"}) {
		t.Fatalf("leaves = %v", leaves)
	}
	if top := parentKeys(nodes, got); !reflect.DeepEqual(top, []string{"a2", "b"}) {
		t.Fatalf("parents = %v", top)
	}
	if got := conductCheck(nodes, nil, "nope", true); got != nil {
		t.Fatalf("unknown key = %v", got)
	}
}

func TestTreeFlattenFilterAndMove(t *testing.T) {
	nodes := sampleTree()
	rows := flattenTree(nodes, map[string]bool{"b": true})
	keys := []string{}
	for _, r := range rows {
		keys = append(keys, r.node.Key)
	}
	if !reflect.DeepEqual(keys, []string{"a", "b", "b1"}) || rows[2].depth != 1 || !rows[2].last[1] {
		t.Fatalf("flatten = %v", keys)
	}
	if !reflect.DeepEqual(allKeys(nodes), []string{"a", "b", "b1"}) {
		t.Fatalf("allKeys = %v", allKeys(nodes))
	}
	kept, open := filterTree(nodes, "plan")
	if len(kept) != 1 || kept[0].Key != "b" || !open["b"] || !open["b1"] || len(kept[0].Children[0].Children) != 1 {
		t.Fatalf("filter = %+v open %v", kept, open)
	}
	if dropPositionAt(2, 28) != DropBefore || dropPositionAt(14, 28) != DropInside || dropPositionAt(26, 28) != DropAfter {
		t.Fatal("dropPositionAt")
	}
	moved := MoveTreeNode(nodes, "a1", "b11", DropAfter)
	if !reflect.DeepEqual(pathTo(moved, "a1"), []string{"b", "b1", "a1"}) {
		t.Fatalf("move after = %v", pathTo(moved, "a1"))
	}
	if kids := moved[1].Children[0].Children; kids[1].Key != "a1" {
		t.Fatalf("order = %+v", kids)
	}
	moved = MoveTreeNode(nodes, "a2", "b", DropInside)
	if !reflect.DeepEqual(pathTo(moved, "a2"), []string{"b", "a2"}) {
		t.Fatalf("move inside = %v", pathTo(moved, "a2"))
	}
	moved = MoveTreeNode(nodes, "b", "a", DropBefore)
	if moved[0].Key != "b" {
		t.Fatalf("move before = %v", moved[0].Key)
	}
	if same := MoveTreeNode(nodes, "b", "b11", DropInside); !reflect.DeepEqual(same, nodes) {
		t.Fatal("moving a node into its own subtree must be refused")
	}
}

func TestCascaderPaths(t *testing.T) {
	opts := []TreeNode{
		{Key: "zj", Title: "浙江", Children: []TreeNode{
			{Key: "hz", Title: "杭州", Children: []TreeNode{{Key: "xh", Title: "西湖"}}},
			{Key: "nb", Title: "宁波", Disabled: true, Children: []TreeNode{{Key: "jb", Title: "江北"}}},
		}},
		{Key: "js", Title: "江苏", Children: []TreeNode{{Key: "nj", Title: "南京"}}},
	}
	cols := cascaderColumns(opts, []string{"zj", "hz"})
	if len(cols) != 3 || cols[2][0].Key != "xh" {
		t.Fatalf("columns = %v", cols)
	}
	if cols := cascaderColumns(opts, []string{"zj", "hz", "xh"}); len(cols) != 3 {
		t.Fatalf("leaf adds no column: %d", len(cols))
	}
	if cols := cascaderColumns(opts, []string{"bogus"}); len(cols) != 1 {
		t.Fatalf("unknown key: %d", len(cols))
	}
	if l := cascaderLabels(opts, []string{"zj", "hz", "xh"}); !reflect.DeepEqual(l, []string{"浙江", "杭州", "西湖"}) {
		t.Fatalf("labels = %v", l)
	}
	found := cascaderSearch(opts, "江", false)
	if len(found) != 3 {
		t.Fatalf("search leaves = %+v", found)
	}
	if !found[1].disabled || found[1].keys[2] != "jb" {
		t.Fatalf("disabled branch = %+v", found[1])
	}
	if got := cascaderSearch(opts, "江", true); len(got) != 7 {
		t.Fatalf("changeOnSelect search = %d", len(got))
	}
	if cascaderSearch(opts, "  ", false) != nil {
		t.Fatal("empty search")
	}
	if p := pathTo(opts, "xh"); !reflect.DeepEqual(p, []string{"zj", "hz", "xh"}) {
		t.Fatalf("pathTo = %v", p)
	}
	if pathTo(opts, "nope") != nil {
		t.Fatal("pathTo unknown")
	}
	sel := pathNodes(opts, []string{"zj", "hz", "xh"})
	if len(sel) != 3 || sel[0].Title != "浙江" || sel[2].Key != "xh" {
		t.Fatalf("pathNodes = %+v", sel)
	}
}

func TestMentionQueryInsert(t *testing.T) {
	pre := []string{"@", "#"}
	start, p, word, ok := mentionQuery("hi @ali", 7, pre)
	if !ok || start != 3 || p != "@" || word != "ali" {
		t.Fatalf("query = %d %q %q %v", start, p, word, ok)
	}
	if _, _, _, ok := mentionQuery("mail a@b", 8, pre); ok {
		t.Fatal("@ inside a word is no mention")
	}
	if _, _, _, ok := mentionQuery("@ali done", 9, pre); ok {
		t.Fatal("space after the word ends the mention")
	}
	if _, p, w, ok := mentionQuery("#", 1, pre); !ok || p != "#" || w != "" {
		t.Fatal("bare prefix")
	}
	text, caret := insertMention("hi @ali!", 3, 7, "@", "alice")
	if text != "hi @alice !" || caret != 10 {
		t.Fatalf("insert = %q %d", text, caret)
	}
	if got := MentionsOf("cc @bob and @carol, #x"); !reflect.DeepEqual(got, []string{"bob", "carol,"}) && !reflect.DeepEqual(got, []string{"bob", "carol"}) {
		t.Fatalf("MentionsOf = %v", got)
	}
}

func TestTransferSplitMove(t *testing.T) {
	data := []TransferItem{{Key: "1", Title: "one"}, {Key: "2", Title: "two", Description: "pair"}, {Key: "3", Title: "three"}}
	left, right := transferSplit(data, []string{"3", "1", "ghost"})
	if len(left) != 1 || left[0].Key != "2" || len(right) != 2 || right[0].Key != "3" {
		t.Fatalf("split = %v / %v", left, right)
	}
	target := transferMove([]string{"3"}, []string{"1", "3"}, true)
	if !reflect.DeepEqual(target, []string{"3", "1"}) {
		t.Fatalf("to right = %v", target)
	}
	target = transferMove(target, []string{"3"}, false)
	if !reflect.DeepEqual(target, []string{"1"}) {
		t.Fatalf("to left = %v", target)
	}
	if got := filterItems(data, "PAIR"); len(got) != 1 || got[0].Key != "2" {
		t.Fatalf("filter = %v", got)
	}
}

func TestCheckAllState(t *testing.T) {
	opts := []Option{{Value: "a"}, {Value: "b"}, {Value: "c", Disabled: true}}
	if all, some := CheckAllState(nil, opts); all || some {
		t.Fatal("none")
	}
	if all, some := CheckAllState([]string{"a"}, opts); all || !some {
		t.Fatal("some")
	}
	if all, some := CheckAllState([]string{"b", "a"}, opts); !all || some {
		t.Fatal("all enabled")
	}
	if got := orderedValues(opts, []string{"x", "b", "a"}); !reflect.DeepEqual(got, []string{"a", "b", "x"}) {
		t.Fatalf("ordered = %v", got)
	}
}

func TestColorText(t *testing.T) {
	c := ui.RGB(0x16, 0x77, 0xff)
	if colorHex(c) != "#1677FF" || colorRGB(c) != "rgb(22,119,255)" {
		t.Fatalf("format = %s %s", colorHex(c), colorRGB(c))
	}
	for _, s := range []string{"#1677ff", "1677FF", "rgb(22, 119, 255)"} {
		got, ok := parseColorText(s)
		if !ok || got.R != 0x16 || got.G != 0x77 || got.B != 0xff || got.A != 255 {
			t.Errorf("parse %q = %+v %v", s, got, ok)
		}
	}
	if got, ok := parseHex("#f00"); !ok || got.R != 255 || got.G != 0 {
		t.Fatalf("short hex = %+v", got)
	}
	if got, ok := parseColorText("rgba(0,0,0,0.5)"); !ok || got.A < 126 || got.A > 128 {
		t.Fatalf("rgba = %+v %v", got, ok)
	}
	for _, s := range []string{"", "#12", "rgb(1,2)", "rgb(300,0,0)", "#zzzzzz"} {
		if _, ok := parseColorText(s); ok {
			t.Errorf("%q should not parse", s)
		}
	}
}

func TestColorHSB(t *testing.T) {
	c := ui.RGB(0x16, 0x77, 0xff)
	got := colorHSB(c)
	if !strings.HasPrefix(got, "hsb(") || !strings.Contains(got, "%") {
		t.Fatalf("colorHSB = %q", got)
	}
	back, ok := parseColorText(got)
	if !ok || abs8(back.R, c.R) > 2 || abs8(back.G, c.G) > 2 || abs8(back.B, c.B) > 2 {
		t.Fatalf("parse HSB %q = %+v want %+v", got, back, c)
	}
	if colorText(c, FormatHSB) != got || formatTag(FormatHSB) != "hsb" || formatOf("hsb") != FormatHSB {
		t.Fatal("format helpers")
	}
	semi := c
	semi.A = 128
	if _, ok := parseColorText(colorHSB(semi)); !ok {
		t.Fatal("hsba")
	}
	if got, ok := parseColorText("hsb(0, 100%, 100%)"); !ok || got.R != 255 || got.G != 0 {
		t.Fatalf("red hsb = %+v", got)
	}
	if _, ok := parseColorText("hsb(400,0,0)"); ok {
		t.Fatal("hue 400")
	}
}

func TestDatePanels(t *testing.T) {
	at := time.Date(2026, 11, 6, 15, 4, 5, 0, time.UTC) // a Friday
	if ws := weekStart(at); ws.Weekday() != time.Monday || ws.Day() != 2 {
		t.Fatalf("weekStart = %v", ws)
	}
	if quarterOf(at) != 4 || quarterOf(time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)) != 1 {
		t.Fatal("quarterOf")
	}
	if decadeOf(2026) != 2020 || decadeOf(2030) != 2030 {
		t.Fatal("decadeOf")
	}
	texts := map[PickerMode]string{
		PickDate:    "2026-11-06",
		PickWeek:    "2026-45周",
		PickMonth:   "2026-11",
		PickQuarter: "2026-Q4",
		PickYear:    "2026",
	}
	for mode, want := range texts {
		if got := pickerText(at, mode, "2006-01-02"); got != want {
			t.Errorf("pickerText(%d) = %q, want %q", mode, got, want)
		}
	}
	norm := map[PickerMode]string{
		PickDate:    "2026-11-06",
		PickWeek:    "2026-11-02",
		PickMonth:   "2026-11-01",
		PickQuarter: "2026-10-01",
		PickYear:    "2026-01-01",
	}
	for mode, want := range norm {
		got := normalizePick(at, mode)
		if got.Format("2006-01-02") != want || got.Hour() != 0 {
			t.Errorf("normalizePick(%d) = %v, want %s", mode, got, want)
		}
	}
	if cells := stepCells(60, 15); len(cells) != 4 || cells[3].value != 45 || cells[3].label != "45" {
		t.Fatalf("stepCells = %v", cells)
	}
	if cells := hourCells(6); len(cells) != 4 || cells[1].label != "06" {
		t.Fatalf("hourCells = %v", cells)
	}
	pm := hours12(1, true)
	if len(pm) != 12 || pm[0].label != "12" || pm[0].value != 12 || pm[11].value != 23 {
		t.Fatalf("hours12 pm = %v", pm)
	}
	if am := hours12(3, false); len(am) != 4 || am[0].value != 0 || am[1].label != "03" {
		t.Fatalf("hours12 am step = %v", am)
	}
	if snapStep(47, 15) != 45 || snapStep(47, 1) != 47 {
		t.Fatal("snapStep")
	}
	if got := timeText(at, "03:04:05 PM", true, LocaleZhCN); got != "下午 03:04:05" {
		t.Fatalf("12h = %q", got)
	}
	if got := timeText(at, "15:04", false, LocaleZhCN); got != "15:04" {
		t.Fatalf("24h = %q", got)
	}
	if s := intSet(func() []int { return []int{1, 3} }); !s[1] || !s[3] || s[2] || len(intSet(nil)) != 0 {
		t.Fatal("intSet")
	}
}

func TestConductedShowsParentsAndChildren(t *testing.T) {
	nodes := sampleTree()
	if got := conducted(nodes, []string{"b11", "b12"}); !reflect.DeepEqual(got, []string{"b", "b1", "b11", "b12"}) {
		t.Fatalf("children up = %v", got)
	}
	if got := conducted(nodes, []string{"a"}); !reflect.DeepEqual(got, []string{"a", "a1", "a2"}) {
		t.Fatalf("parent down = %v", got)
	}
	// Checking a2 against an unconducted [a1] still checks the parent.
	if got := conductCheck(nodes, []string{"a1"}, "a2", true); !reflect.DeepEqual(got, []string{"a", "a1", "a2"}) {
		t.Fatalf("conductCheck from raw = %v", got)
	}
}

func TestHSVRoundTrip(t *testing.T) {
	for _, hex := range []string{"#1677ff", "#52c41a", "#faad14", "#ff4d4f", "#000000", "#ffffff", "#808080"} {
		c, _ := parseHex(hex)
		h, s, v := toHSV(c)
		if got := hsva(h, s, v, 255); got != c {
			t.Errorf("%s -> %v %v %v -> %s", hex, h, s, v, colorHex(got))
		}
	}
	if got := hsva(0, 1, 1, 128); got.R != 255 || got.G != 0 || got.A != 128 {
		t.Fatalf("red half = %+v", got)
	}
}
