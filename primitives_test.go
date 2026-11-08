package kit

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestAttachAnchorsAlign(t *testing.T) {
	cases := []struct {
		side     Placement
		align    PopupAlign
		at, self ui.Anchor
	}{
		{PlaceBottom, AlignFirst, ui.AnchorBottomLeft, ui.AnchorTopLeft},
		{PlaceBottom, AlignLast, ui.AnchorBottomRight, ui.AnchorTopRight},
		{PlaceTop, AlignMiddle, ui.AnchorTop, ui.AnchorBottom},
		{PlaceRight, AlignFirst, ui.AnchorTopRight, ui.AnchorTopLeft},
		{PlaceLeft, AlignLast, ui.AnchorBottomLeft, ui.AnchorBottomRight},
	}
	for _, tc := range cases {
		at, self := attachAnchors(tc.side, tc.align)
		if at != tc.at || self != tc.self {
			t.Errorf("side %d align %d: got %v/%v", tc.side, tc.align, at, self)
		}
	}
}

func TestVirtualRange(t *testing.T) {
	cases := []struct {
		off, view, row float32
		n, over        int
		from, to       int
	}{
		{0, 100, 20, 1000, 0, 0, 6},
		{0, 100, 20, 1000, 2, 0, 8},
		{400, 100, 20, 1000, 2, 18, 28},
		{19990, 100, 20, 1000, 2, 997, 1000},
		{0, 100, 20, 3, 4, 0, 3},
		{-50, 100, 20, 10, 0, 0, 6},
		{0, 100, 0, 10, 0, 0, 0},
		{0, 100, 20, 0, 0, 0, 0},
	}
	for _, tc := range cases {
		from, to := virtualRange(tc.off, tc.view, tc.row, tc.n, tc.over)
		if from != tc.from || to != tc.to {
			t.Errorf("virtualRange(%v, %v, %v, %d, %d) = %d, %d; want %d, %d", tc.off, tc.view, tc.row, tc.n, tc.over, from, to, tc.from, tc.to)
		}
	}
}

func TestFixedWindowBuildsOnlyVisibleRows(t *testing.T) {
	built := 0
	tester := ui.NewTester(func(c *ui.Context) {
		built = 0
		Page(c, Work, func() {
			fixedWindow(c, 5000, 32, 256, func(c *ui.Context, i int) {
				built++
				ui.Text(c, "选项 "+strconv.Itoa(i))
			}, -1)
		})
	}, 400, 400)
	tester.Frame()
	if built > 20 {
		t.Fatalf("built %d rows of 5000 for a 256 view", built)
	}
	if !tester.HasText("选项 0") || tester.HasText("选项 100") {
		t.Fatal("wrong rows in view")
	}
}

func TestVirtualListBuildsOnlyVisibleRows(t *testing.T) {
	built := 0
	tester := ui.NewTester(func(c *ui.Context) {
		built = 0
		Page(c, Work, func() {
			VirtualList(c, 10000, func(c *ui.Context, i int) {
				built++
				ui.Text(c, "行 "+strconv.Itoa(i)).Height(24)
			}, VirtualListProps{Height: 240, Split: true})
		})
	}, 400, 400)
	tester.Frame()
	if built == 0 || built > 60 {
		t.Fatalf("built %d rows of 10000", built)
	}
}

func near(a, b float32) bool { return math.Abs(float64(a-b)) < 0.5 }

func TestArcPath(t *testing.T) {
	b := arcPath(50, 50, 40, 0, 90).Bounds()
	if !near(b.X, 50) || !near(b.Y, 10) || !near(b.X+b.W, 90) || !near(b.Y+b.H, 50) {
		t.Errorf("quarter arc from twelve o'clock bounds %+v, want x 50..90, y 10..50", b)
	}
	full := arcPath(50, 50, 40, 0, 360).Bounds()
	if !near(full.W, 80) || !near(full.H, 80) {
		t.Errorf("full circle bounds %+v", full)
	}
	back := arcPath(50, 50, 40, 0, -90).Bounds()
	if !near(back.X, 10) || !near(back.X+back.W, 50) {
		t.Errorf("counterclockwise arc bounds %+v, want x 10..50", back)
	}
}

func TestStarPath(t *testing.T) {
	b := starPath(10, 10, 10, 4).Bounds()
	if !near(b.Y, 0) || b.W > 20.5 || b.H > 20 {
		t.Errorf("star bounds %+v", b)
	}
}

func TestTileOriginsCover(t *testing.T) {
	r := ui.Rect{X: 0, Y: 0, W: 300, H: 200}
	tiles := tileOrigins(r, 80, 30, 40, 40)
	if len(tiles) == 0 {
		t.Fatal("no tiles")
	}
	for y := float32(5); y < r.H; y += 37 {
		for x := float32(5); x < r.W; x += 41 {
			hit := false
			for _, o := range tiles {
				// Every point lies in some tile's step cell.
				if x >= o.X && x < o.X+120 && y >= o.Y && y < o.Y+70 {
					hit = true
					break
				}
			}
			if !hit {
				t.Fatalf("(%v, %v) is in no tile cell", x, y)
			}
		}
	}
	if tileOrigins(r, 0, 0, 0, 0) != nil {
		t.Error("zero step should give no tiles")
	}
}

func TestCountAndClamp(t *testing.T) {
	s := "服务器地址"
	clampRunes(&s, 3)
	if s != "服务器" {
		t.Errorf("clamp = %q", s)
	}
	if got := countText(s, 10); got != "3 / 10" {
		t.Errorf("count = %q", got)
	}
	if got := countText("abc", 0); got != "3" {
		t.Errorf("count = %q", got)
	}
}

func TestPopupHoverGrace(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Padding(120, 40).Children(func() {
				Popover(c, func() { Muted(c, "浮层内容") }, func() bool { return Button(c, "悬停我") }, PopoverProps{PopupProps: PopupProps{Trigger: TriggerHover}})
			})
		})
	}, 500, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	a, _ := tester.Find("悬停我")
	tester.Move(a.X+a.W/2, a.Y+a.H/2)
	tester.Frame()
	if !tester.HasText("浮层内容") {
		t.Fatal("hover did not open")
	}
	tester.Move(490, 390)
	tester.Advance(hoverGrace / 2)
	if !tester.HasText("浮层内容") {
		t.Fatal("closed before the grace period")
	}
	tester.Advance(hoverGrace)
	tester.Frame()
	if tester.HasText("浮层内容") {
		t.Fatal("still open after leaving")
	}
}

func TestPopupControlledAndClick(t *testing.T) {
	open := false
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Padding(120, 40).Children(func() {
				Popover(c, func() { Muted(c, "浮层内容") }, func() bool { return Button(c, "点我") }, PopoverProps{Title: "标题", PopupProps: PopupProps{Open: &open, Placement: PlaceBottom}})
			})
		})
	}, 500, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if tester.HasText("浮层内容") {
		t.Fatal("open while closed")
	}
	open = true
	tester.Frame()
	if !tester.HasText("标题") {
		t.Fatal("did not follow Open")
	}
	a, _ := tester.Find("点我")
	b, _ := tester.Find("标题")
	if b.Y < a.Y+a.H {
		t.Errorf("bottom popover at y=%v, above the trigger's bottom %v", b.Y, a.Y+a.H)
	}
	if err := tester.Click("点我"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if open || tester.HasText("浮层内容") {
		t.Fatal("trigger click did not close it")
	}
}

func TestDropdownContextMenuOpensAtPointer(t *testing.T) {
	got := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Padding(40).Children(func() {
				if k := Dropdown(c, []MenuItem{{Key: "copy", Label: "复制"}, {Key: "del", Label: "删除", Danger: true}}, func() bool {
					ui.Box(c).Size(300, 200).Background(Work.FillTertiary).Children(func() { Muted(c, "右键区域") })
					return false
				}, DropdownProps{Trigger: TriggerContextMenu}); k != "" {
					got = k
				}
			})
		})
	}, 500, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.RightClickAt(200, 150)
	tester.Frame()
	tester.Frame()
	r, ok := tester.Find("复制")
	if !ok {
		t.Fatal("right-click did not open the menu")
	}
	if r.Y < 150 || r.Y > 190 || r.X < 190 || r.X > 240 {
		t.Errorf("menu at %+v, not at the pointer (200, 150)", r)
	}
	if err := tester.Click("删除"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if got != "del" {
		t.Fatalf("chose %q", got)
	}
}

func TestInputChrome(t *testing.T) {
	value := "secret"
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Input(c, &value, InputProps{Password: true, AllowClear: true, ShowCount: true, MaxLength: 8, AddonBefore: "密钥"})
		})
	}, 500, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("密钥") || !tester.HasText("6 / 8") {
		t.Fatalf("chrome missing: %v", tester.Texts())
	}
	if err := tester.Click("显示密码"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if _, ok := tester.Find("隐藏密码"); !ok {
		t.Fatal("eye did not reveal the password")
	}
	if err := tester.Click("清除"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if value != "" || !tester.HasText("0 / 8") {
		t.Fatalf("clear left %q", value)
	}
	k, _ := tester.Find("密钥")
	tester.ClickAt(k.X+k.W+60, k.Y+k.H/2)
	tester.Type("abcdefghijk")
	tester.Frame()
	if value != "abcdefgh" {
		t.Fatalf("MaxLength left %q", value)
	}
}

func TestDialogZoomsOut(t *testing.T) {
	open := true
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {})
		Dialog(c, &open, "弹窗标题", func() { Muted(c, "弹窗内容") })
	}, 800, 600)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Advance(dialogMotion)
	if err := tester.Click("关闭"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if open {
		t.Fatal("close did not close")
	}
	if !tester.HasText("弹窗内容") {
		t.Fatal("vanished without zooming out")
	}
	tester.Advance(dialogMotion + 50*time.Millisecond)
	if tester.HasText("弹窗内容") {
		t.Fatal("still shown after the motion")
	}
}
