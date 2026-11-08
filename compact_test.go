package kit

import (
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestCompactCorners(t *testing.T) {
	all := [4]bool{true, true, true, true}
	cases := []struct {
		j, k        int
		first, last bool
		vertical    bool
		want        [4]bool
	}{
		{0, 0, true, true, false, all},
		{0, 0, true, false, false, [4]bool{true, false, false, true}},
		{0, 0, false, false, false, [4]bool{}},
		{0, 0, false, true, false, [4]bool{false, true, true, false}},
		{0, 1, true, true, false, [4]bool{true, false, false, true}},
		{1, 1, true, true, false, [4]bool{false, true, true, false}},
		{0, 0, true, false, true, [4]bool{true, true, false, false}},
		{0, 0, false, true, true, [4]bool{false, false, true, true}},
		{0, 1, true, false, true, [4]bool{true, false, false, false}},
	}
	for _, tc := range cases {
		if got := compactCorners(tc.j, tc.k, tc.first, tc.last, tc.vertical); got != tc.want {
			t.Errorf("corners(%d of %d, first %v last %v vertical %v) = %v, want %v", tc.j, tc.k, tc.first, tc.last, tc.vertical, got, tc.want)
		}
	}
}

func TestCompactClaimsOncePerControl(t *testing.T) {
	if compactBegin() != nil {
		t.Fatal("no claim outside a compact group")
	}
	var items int
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			g := &compactGroup{}
			compactStack = append(compactStack, g)
			ui.Row(c).Children(func() {
				q := ""
				Search(c, &q)
				Button(c, "提交")
				AutoComplete(c, &q, Opts("a"))
			})
			compactStack = compactStack[:0]
			items = len(g.items)
			if len(g.items) == 3 && len(g.items[0].parts) != 2 {
				t.Errorf("search should join its box and button, got %d parts", len(g.items[0].parts))
			}
		})
	}, 600, 120)
	tester.Frame()
	if items != 3 {
		t.Fatalf("items = %d, want 3 (the search button is a part, not an item)", items)
	}
}

func nearColor(a, b color.Color) bool {
	r1, g1, b1, _ := a.RGBA()
	r2, g2, b2, _ := b.RGBA()
	d := func(x, y uint32) bool { return x>>8 <= y>>8+12 && y>>8 <= x>>8+12 }
	return d(r1, r2) && d(g1, g2) && d(b1, b2)
}

func TestSpaceCompactJoinsAndLightsSharedEdge(t *testing.T) {
	vertical := false
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(20).Children(func() {
				SpaceCompact(c, func() {
					Button(c, "左边")
					Button(c, "中间")
					Button(c, "右边")
				}, CompactProps{Vertical: vertical})
			})
		})
	}, 600, 240)
	tester.SetLocale("zh-CN")
	tester.Frame()
	a, _ := tester.Find("左边")
	b, _ := tester.Find("中间")
	if b.X != a.X+a.W-1 || b.Y != a.Y {
		t.Fatalf("second button at %v, first %v: should overlap by 1px", b, a)
	}
	img := tester.Image()
	sx := float32(img.Bounds().Dx()) / 600
	px := func(x, y float32) color.Color { return img.At(int(x*sx), int(y*sx)) }
	edge := px(b.X+0.5, b.Y+b.H/2)
	if !nearColor(edge, color.RGBA{217, 217, 217, 255}) {
		t.Fatalf("shared edge at rest = %v, want the grey border", edge)
	}
	tester.Move(a.X+a.W/2, a.Y+a.H/2)
	tester.Frame()
	tester.Advance(300 * time.Millisecond)
	img = tester.Image()
	if r, _, bl, _ := px(b.X+0.5, b.Y+b.H/2).RGBA(); bl>>8 < r>>8+80 {
		t.Fatalf("shared edge with the first hovered = %v, want the primary hover blue", px(b.X+0.5, b.Y+b.H/2))
	}
	vertical = true
	tester.Move(500, 220)
	tester.Frame()
	tester.Frame()
	a, _ = tester.Find("左边")
	b, _ = tester.Find("中间")
	if b.Y != a.Y+a.H-1 || b.X != a.X {
		t.Fatalf("vertical: second at %v, first %v", b, a)
	}
}

func TestGalleryCompactShots(t *testing.T) {
	for _, dark := range []bool{false, true} {
		tok := pick(dark, WorkDark, Work)
		var url, ip, port, q, city, area, num = "mygo", "172.30.0.9", "36678", "会议", "sh", "", "8"
		var day, at time.Time
		day = time.Date(2026, 11, 6, 0, 0, 0, 0, time.Local)
		var start, end time.Time
		var path []string
		tester := ui.NewTester(func(c *ui.Context) {
			Page(c, tok, func() {
				ui.Column(c).Fill().Padding(24).Gap(16).Children(func() {
					panel(c, tok, "Space.Compact · 按钮", func() {
						ui.Column(c).Gap(12).Children(func() {
							SpaceCompact(c, func() {
								Button(c, "上一页")
								Button(c, "刷新", ButtonProps{force: hover})
								Button(c, "下一页")
							})
							SpaceCompact(c, func() {
								Button(c, "保存", ButtonProps{Type: Primary})
								Button(c, "另存为", ButtonProps{Type: Primary})
								Button(c, "删除", ButtonProps{Danger: true})
							})
							SpaceCompact(c, func() {
								Button(c, "小", ButtonProps{})
								Button(c, "号")
								Button(c, "组")
							}, CompactProps{Size: Small})
						})
					})
					panel(c, tok, "Space.Compact · 输入类", func() {
						ui.Column(c).Gap(12).Width(720).Children(func() {
							SpaceCompact(c, func() {
								Input(c, &ip, InputProps{AddonBefore: "IP", Width: 200})
								Input(c, &port, InputProps{Width: 100, force: hover})
								Button(c, "连接", ButtonProps{Type: Primary})
							})
							SpaceCompact(c, func() {
								Select(c, &city, []Option{{Value: "sh", Label: "上海"}, {Value: "bj", Label: "北京"}}, SelectProps{Width: 100})
								Input(c, &url, InputProps{AddonAfter: ".com", Width: 220})
								InputNumber(c, &num, InputNumberProps{Width: 90})
							}, CompactProps{Block: true})
							SpaceCompact(c, func() {
								DatePicker(c, &day, DatePickerProps{Width: 160})
								TimePicker(c, &at, TimePickerProps{Width: 130, force: focus})
								Search(c, &q, SearchProps{Width: 220})
							})
							SpaceCompact(c, func() {
								AutoComplete(c, &area, Opts("浦东", "徐汇"), AutoCompleteProps{Width: 160, Placeholder: "区域"})
								Cascader(c, &path, []TreeNode{{Key: "sh", Title: "上海"}}, CascaderProps{SelectProps: SelectProps{Width: 160, Placeholder: "地区"}})
								RangePicker(c, &start, &end, RangePickerProps{Width: 260})
							})
						})
					})
					panel(c, tok, "Space.Compact · 竖排", func() {
						SpaceCompact(c, func() {
							Button(c, "置顶")
							Button(c, "上移", ButtonProps{force: hover})
							Button(c, "下移")
							Button(c, "置底")
						}, CompactProps{Vertical: true})
					})
				})
			})
		}, 1100, 760)
		tester.SetLocale("zh-CN")
		tester.Frame()
		tester.Frame()
		shot(t, tester, pick(dark, "mygo-kit-gallery-dark-compact.png", "mygo-kit-gallery-compact.png"))
	}
}
