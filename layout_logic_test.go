package kit

import (
	"slices"
	"testing"
	"unicode/utf8"
)

func TestScreenOfAndSpans(t *testing.T) {
	for _, tc := range []struct {
		w    float32
		want Screen
	}{{0, ScreenXS}, {575, ScreenXS}, {576, ScreenSM}, {768, ScreenMD}, {991, ScreenMD}, {992, ScreenLG}, {1200, ScreenXL}, {1600, ScreenXXL}} {
		if got := ScreenOf(tc.w); got != tc.want {
			t.Errorf("ScreenOf(%v) = %v, want %v", tc.w, got, tc.want)
		}
	}
	p := ColProps{ColSpan: ColSpan{Span: 24}, MD: &ColSpan{Span: 12}, XL: &ColSpan{Span: 6, Offset: 2}}
	if got := p.spanAt(ScreenSM).Span; got != 24 {
		t.Errorf("sm span = %d, want 24", got)
	}
	if got := p.spanAt(ScreenLG).Span; got != 12 {
		t.Errorf("lg span = %d, want md's 12", got)
	}
	if got := p.spanAt(ScreenXXL); got.Span != 6 || got.Offset != 2 {
		t.Errorf("xxl = %+v, want xl's", got)
	}
	r := RowProps{Gutter: [2]float32{8, 8}, GutterBy: map[Screen][2]float32{ScreenMD: {16, 16}, ScreenXL: {32, 24}}}
	if g := r.gutterAt(ScreenSM); g != [2]float32{8, 8} {
		t.Errorf("sm gutter = %v", g)
	}
	if g := r.gutterAt(ScreenLG); g != [2]float32{16, 16} {
		t.Errorf("lg gutter = %v", g)
	}
	if g := r.gutterAt(ScreenXXL); g != [2]float32{32, 24} {
		t.Errorf("xxl gutter = %v", g)
	}
	cols := []gridCol{
		{props: ColProps{ColSpan: ColSpan{Span: 6, Order: 2}}, at: 0},
		{props: ColProps{ColSpan: ColSpan{Span: 6, Order: 1}}, at: 1},
		{props: ColProps{XS: &ColSpan{Span: 0}, MD: &ColSpan{Span: 6}}, at: 2},
		{props: ColProps{Auto: true}, at: 3},
	}
	var order []int
	for _, col := range gridPlan(cols, ScreenSM) {
		order = append(order, col.at)
	}
	if !slices.Equal(order, []int{3, 1, 0}) {
		t.Errorf("sm plan = %v, want the hidden column dropped and the rest by order", order)
	}
	order = order[:0]
	for _, col := range gridPlan(cols, ScreenMD) {
		order = append(order, col.at)
	}
	if !slices.Equal(order, []int{2, 3, 1, 0}) {
		t.Errorf("md plan = %v", order)
	}
	bp := ScreenLG
	if !siderBroken(&bp, ScreenMD) || siderBroken(&bp, ScreenLG) || siderBroken(nil, ScreenXS) {
		t.Error("siderBroken should hold only below the breakpoint")
	}
}

func TestSplitterMath(t *testing.T) {
	panels := []SplitPanel{{Min: 100}, {}, {Size: 200, Max: 300}}
	s := splitShares(panels, 1000)
	if s[2] != 0.2 || s[0] != 0.4 || s[1] != 0.4 {
		t.Fatalf("shares = %v, want 0.4 0.4 0.2", s)
	}
	px := []float32{400, 400, 200}
	if got := splitMove(px, panels, 0, -350); got[0] != 100 || got[1] != 700 {
		t.Errorf("move past min = %v, want first held at 100", got)
	}
	if got := splitMove(px, panels, 1, -300); got[1] != 300 || got[2] != 300 {
		t.Errorf("move past max = %v, want last held at 300", got)
	}
	if got := splitMove(px, panels, 0, 50); got[0] != 450 || got[1] != 350 || got[2] != 200 {
		t.Errorf("move = %v", got)
	}
	sizes, before := []float32{0.3, 0.5, 0.2}, make([]float32, 3)
	splitFold(sizes, before, 0, true)
	if sizes[0] != 0 || sizes[1] != 0.8 || before[0] != 0.3 {
		t.Fatalf("fold start = %v %v", sizes, before)
	}
	splitFold(sizes, before, 0, false)
	if sizes[0] != 0.3 || sizes[1] != 0.5 || before[0] != 0 {
		t.Errorf("restore = %v %v", sizes, before)
	}
}

func TestTypoCut(t *testing.T) {
	// Ten runes a line, 20 high a line.
	height := func(s string) float32 {
		n := utf8.RuneCountInString(s)
		return float32((n+9)/10) * 20
	}
	if got := typoCut("短文本", "…", 40, height); got != 3 {
		t.Errorf("fitting text cut to %d", got)
	}
	text := "一二三四五六七八九十一二三四五六七八九十一二三四五"
	got := typoCut(text, "… 展开", 40, height)
	if got != 16 {
		t.Errorf("cut = %d, want 16 runes leaving room for the tail", got)
	}
}

func TestMenuPathsAndFit(t *testing.T) {
	items := []MenuItem{
		{Key: "a", Children: []MenuItem{
			{Key: "g", Group: true, Children: []MenuItem{{Key: "a1"}}},
			{Key: "b", Children: []MenuItem{{Key: "b1"}}},
		}},
		{Key: "c"},
	}
	if got := menuPath(items, "b1"); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("path to b1 = %v", got)
	}
	if got := menuPath(items, "a1"); !slices.Equal(got, []string{"a"}) {
		t.Errorf("path to a1 = %v, want the group skipped", got)
	}
	if got := menuPath(items, "c"); got != nil {
		t.Errorf("path to a top item = %v", got)
	}
	if !menuHas(items[0], "b1") || menuHas(items[0], "c") || menuHas(items[0], "") {
		t.Error("menuHas")
	}
	widths := []float32{100, 100, 100, 100}
	if n := menuFit(widths, 400, 50); n != 4 {
		t.Errorf("all fit = %d", n)
	}
	if n := menuFit(widths, 399, 50); n != 3 {
		t.Errorf("overflow fit = %d, want 3 leaving room for more", n)
	}
	if got := tabsVisible(widths, 349, 50, 3); !slices.Equal(got, []int{0, 3}) {
		t.Errorf("visible = %v, want the chosen tab swapped in", got)
	}
	if got := tabsVisible(widths, 0, 50, 1); len(got) != 4 {
		t.Errorf("unmeasured bar shows %v", got)
	}
}

func TestAnchorAffixStepsLogic(t *testing.T) {
	marks := map[string]float32{"a": 0, "b": 300, "c": 600}
	keys := []string{"a", "b", "c"}
	for _, tc := range []struct {
		y, off float32
		want   string
	}{{0, 0, "a"}, {299, 0, "b"}, {290, 0, "a"}, {250, 50, "b"}, {1000, 0, "c"}} {
		if got := anchorActive(keys, marks, tc.y, tc.off, 5); got != tc.want {
			t.Errorf("active at %v (offset %v) = %q, want %q", tc.y, tc.off, got, tc.want)
		}
	}
	if anchorActive(keys, map[string]float32{"b": 300}, 0, 0, 5) != "" {
		t.Error("nothing reached should be no link")
	}
	if !affixStuck(10, 30, 20, 500, AffixProps{}) || affixStuck(30, 30, 20, 500, AffixProps{}) {
		t.Error("top affix")
	}
	if !affixStuck(480, 30, 0, 500, AffixProps{Bottom: true}) || affixStuck(400, 30, 0, 500, AffixProps{OffsetBottom: 10}) {
		t.Error("bottom affix")
	}
	for i, want := range []StepStatus{StepFinish, StepError, StepWait} {
		if got := stepStatus(i, 1, StepItem{}, StepError); got != want {
			t.Errorf("step %d = %v, want %v", i, got, want)
		}
	}
	if stepStatus(2, 1, StepItem{Status: StepFinish}, StepAuto) != StepFinish {
		t.Error("an item's own status wins")
	}
	if stepStatus(1, 1, StepItem{}, StepAuto) != StepProcess {
		t.Error("current defaults to process")
	}
}

func TestIconRegistry(t *testing.T) {
	if LookupIcon("search") == nil || LookupIcon("check-circle-twotone") == nil {
		t.Fatal("built-in icons missing")
	}
	RegisterIcon("kit-test-icon", svgIcon(`<path d="M2 2h12"/>`), nil)
	if LookupIcon("kit-test-icon") == nil || !slices.Contains(IconNames(), "kit-test-icon") {
		t.Error("registered icon not found")
	}
	if !slices.IsSorted(IconNames()) {
		t.Error("IconNames not sorted")
	}
}
