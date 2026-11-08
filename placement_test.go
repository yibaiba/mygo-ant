package kit

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestFlipPlacement(t *testing.T) {
	const winW, winH = 800, 600
	mid := ui.Rect{X: 380, Y: 280, W: 40, H: 32}
	cases := []struct {
		name   string
		want   Placement
		anchor ui.Rect
		w, h   float32
		got    Placement
	}{
		{"room above", PlaceTop, mid, 200, 60, PlaceTop},
		{"top edge goes below", PlaceTop, ui.Rect{X: 380, Y: 10, W: 40, H: 32}, 200, 60, PlaceBottom},
		{"bottom edge goes above", PlaceBottom, ui.Rect{X: 380, Y: 560, W: 40, H: 32}, 200, 60, PlaceTop},
		{"left edge goes right", PlaceLeft, ui.Rect{X: 4, Y: 280, W: 40, H: 32}, 120, 40, PlaceRight},
		{"right edge goes left", PlaceRight, ui.Rect{X: 740, Y: 280, W: 40, H: 32}, 120, 40, PlaceLeft},
		{"neither fits keeps the roomier", PlaceTop, ui.Rect{X: 380, Y: 100, W: 40, H: 32}, 200, 700, PlaceBottom},
		{"no worse side keeps want", PlaceBottom, ui.Rect{X: 380, Y: 10, W: 40, H: 580}, 200, 60, PlaceBottom},
		{"unmeasured keeps want", PlaceTop, ui.Rect{X: 380, Y: 0, W: 40, H: 32}, 0, 0, PlaceTop},
	}
	for _, tc := range cases {
		if got := flipPlacement(tc.want, tc.anchor, tc.w, tc.h, winW, winH); got != tc.got {
			t.Errorf("%s: got %d, want %d", tc.name, got, tc.got)
		}
	}
}

func TestArrowShiftPointsAtAnchor(t *testing.T) {
	panel := ui.Rect{X: 100, Y: 100, W: 200, H: 60}
	if d := arrowShift(PlaceTop, ui.Rect{X: 190, Y: 170, W: 20, H: 20}, panel); d != 0 {
		t.Errorf("centered anchor shift = %v", d)
	}
	if d := arrowShift(PlaceTop, ui.Rect{X: 110, Y: 170, W: 20, H: 20}, panel); d != -80 {
		t.Errorf("left anchor shift = %v, want -80", d)
	}
	if d := arrowShift(PlaceBottom, ui.Rect{X: 0, Y: 0, W: 10, H: 10}, panel); d != -80 {
		t.Errorf("far anchor shift = %v, want clamped to -80", d)
	}
	if d := arrowShift(PlaceRight, ui.Rect{X: 0, Y: 400, W: 10, H: 10}, panel); d != 10 {
		t.Errorf("side shift = %v, want clamped to 10", d)
	}
}

func hoverTip(t *testing.T, tester *ui.Tester, label string) (anchor, tip ui.Rect) {
	t.Helper()
	tester.Frame()
	a, ok := tester.Find(label)
	if !ok {
		t.Fatalf("no %q", label)
	}
	tester.Move(a.X+a.W/2, a.Y+a.H/2)
	tester.Advance(800 * time.Millisecond)
	tester.Frame()
	tester.Frame()
	tip, ok = tester.Find("提示内容")
	if !ok {
		t.Fatal("tooltip did not show")
	}
	return a, tip
}

func TestTooltipFlipsAtWindowEdges(t *testing.T) {
	for _, at := range []string{"top", "bottom"} {
		tester := ui.NewTester(func(c *ui.Context) {
			Page(c, Work, func() {
				col := ui.Column(c).Fill().Padding(4).AlignItems(ui.Center)
				col.Justify(pick(at == "top", ui.Start, ui.End))
				col.Children(func() {
					Tooltip(c, "提示内容", func() { Button(c, "锚点") })
				})
			})
		}, 400, 300)
		tester.SetLocale("zh-CN")
		a, tip := hoverTip(t, tester, "锚点")
		switch at {
		case "top":
			if tip.Y < a.Y+a.H {
				t.Errorf("at the top edge the tip is at y=%v, above the anchor's bottom %v", tip.Y, a.Y+a.H)
			}
		default:
			if tip.Y+tip.H > a.Y {
				t.Errorf("with room above the tip ends at y=%v, below the anchor's top %v", tip.Y+tip.H, a.Y)
			}
		}
	}
}

func TestTooltipPlacementRight(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Padding(40).Children(func() {
				Tooltip(c, "提示内容", func() { Button(c, "锚点") }, TooltipProps{Placement: PlaceRight})
			})
		})
	}, 400, 300)
	tester.SetLocale("zh-CN")
	a, tip := hoverTip(t, tester, "锚点")
	if tip.X < a.X+a.W {
		t.Errorf("right tip starts at x=%v, inside the anchor ending at %v", tip.X, a.X+a.W)
	}
}

func TestPopconfirmOpensAbove(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Padding(200, 40).Children(func() {
				Popconfirm(c, "确定移除？", func() bool { return Button(c, "移除") })
			})
		})
	}, 500, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	a, _ := tester.Find("移除")
	if err := tester.Click("移除"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	tester.Frame()
	title, ok := tester.Find("确定移除？")
	if !ok {
		t.Fatal("did not open")
	}
	if title.Y > a.Y {
		t.Errorf("title at y=%v, not above the trigger at %v", title.Y, a.Y)
	}
}
