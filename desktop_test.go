package kit

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestWindowMinSize(t *testing.T) {
	w, h := WindowMinSize()
	if w != 960 || h != 600 {
		t.Fatalf("WindowMinSize = %d×%d", w, h)
	}
}

func TestWheelStep(t *testing.T) {
	if wheelStep(-1) != 1 || wheelStep(12) != -1 || wheelStep(0) != 0 {
		t.Fatalf("wheelStep up/down/zero")
	}
}

func TestTooltipAntDelayAndEscape(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(24).Children(func() {
				Tooltip(c, "提示内容", func() { Button(c, "锚点") })
			})
		})
	}, 400, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	a, ok := tester.Find("锚点")
	if !ok {
		t.Fatal("no anchor")
	}
	tester.Move(a.X+a.W/2, a.Y+a.H/2)
	tester.Advance(50 * time.Millisecond)
	if tester.HasText("提示内容") {
		t.Fatal("tooltip showed before 100ms")
	}
	tester.Advance(80 * time.Millisecond)
	tester.Frame()
	if !tester.HasText("提示内容") {
		t.Fatalf("tooltip missing after 130ms: %v", tester.Texts())
	}
	tester.Key(0, ui.KeyEscape)
	tester.Frame()
	if tester.HasText("提示内容") {
		t.Fatal("Escape left the tooltip open")
	}
}

func TestSliderWheelWhenFocused(t *testing.T) {
	v := 40.0
	tester := mount(func(c *ui.Context) {
		ui.Box(c).Width(300).Children(func() {
			Slider(c, &v, SliderProps{Step: 10, Tooltip: TooltipNever})
		})
	}, 400, 160)
	x, y := center(t, tester, "滑块")
	tester.ClickAt(x, y)
	tester.Wheel(x, y, 0, -1)
	if v != 50 {
		t.Fatalf("wheel up while focused = %v, want 50", v)
	}
	tester.Wheel(x, y, 0, 1)
	if v != 40 {
		t.Fatalf("wheel down = %v, want 40", v)
	}
}

func TestPaginationWheel(t *testing.T) {
	cur := 1
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(16).Children(func() {
				Pagination(c, &cur, 95, 10)
			})
		})
	}, 700, 80)
	tester.SetLocale("zh-CN")
	tester.Frame()
	r, ok := tester.Find("下一页")
	if !ok {
		t.Fatal("no next")
	}
	tester.Wheel(r.X-12, r.Y+r.H/2, 0, 40)
	if cur != 2 {
		t.Fatalf("wheel down = page %d, want 2", cur)
	}
	tester.Wheel(r.X-12, r.Y+r.H/2, 0, -40)
	if cur != 1 {
		t.Fatalf("wheel up = page %d, want 1", cur)
	}
}

func TestTableOnDoubleClick(t *testing.T) {
	sel, opened := -1, -1
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Table(c, []Column{{Title: "IP", Width: 160}, {Title: "状态"}}, len(hosts), hostCell, TableProps{
				Selected:      &sel,
				OnDoubleClick: func(row int) { opened = row },
			})
		})
	}, 480, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	r, ok := tester.Find("192.168.0.23")
	if !ok {
		t.Fatal("no row")
	}
	tester.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	tester.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	if sel != 1 || opened != 1 {
		t.Fatalf("double click: sel=%d opened=%d", sel, opened)
	}
}

func TestTreeDoubleClickExpands(t *testing.T) {
	tester := mount(func(c *ui.Context) {
		Tree(c, regions())
	}, 400, 300)
	if tester.HasText("杭州") {
		t.Fatal("collapsed tree shows children")
	}
	r, ok := tester.Find("浙江")
	if !ok {
		t.Fatal("no 浙江")
	}
	tester.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	tester.ClickAt(r.X+r.W/2, r.Y+r.H/2)
	if !tester.HasText("杭州") {
		t.Fatalf("double click did not expand: %v", tester.Texts())
	}
}

func TestInputEngineEditMenu(t *testing.T) {
	value := "hello world"
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(20).Width(280).Children(func() {
				Input(c, &value, InputProps{Prefix: "名称"})
			})
		})
	}, 400, 120)
	tester.Frame()
	r, ok := tester.Find("名称")
	if !ok {
		t.Fatal("no field")
	}
	tester.RightClickAt(r.X+r.W+40, r.Y+r.H/2)
	menu := tester.Menu()
	for _, label := range []string{"Cut", "Copy", "Paste", "Select All"} {
		if !slices.Contains(menu, label) {
			t.Fatalf("engine edit menu %q has no %s", menu, label)
		}
	}
}

func TestDesktopCursors(t *testing.T) {
	name := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(16).Gap(12).Children(func() {
				Button(c, "保存", ButtonProps{Type: Primary})
				Button(c, "禁用", ButtonProps{Disabled: true})
				Button(c, "连接中", ButtonProps{Type: Primary, Loading: true})
				Input(c, &name, InputProps{Prefix: "IP", Width: 220})
				Splitter(c, []SplitPanel{
					{Render: func() { Muted(c, "左栏") }, Size: 160},
					{Render: func() { Muted(c, "右栏") }},
				})
			})
		})
	}, 480, 360)
	tester.SetLocale("zh-CN")
	tester.Frame()
	move := func(label string) {
		t.Helper()
		r, ok := tester.Find(label)
		if !ok {
			t.Fatalf("no %q in %v", label, tester.Texts())
		}
		tester.Move(r.X+r.W/2, r.Y+r.H/2)
	}
	move("保存")
	if tester.Cursor() != ui.CursorPointer {
		t.Fatalf("button cursor %v", tester.Cursor())
	}
	move("禁用")
	if tester.Cursor() != ui.CursorNotAllowed {
		t.Fatalf("disabled cursor %v", tester.Cursor())
	}
	move("连接中")
	if tester.Cursor() != ui.CursorNotAllowed {
		t.Fatalf("loading cursor %v", tester.Cursor())
	}
	move("IP")
	if tester.Cursor() != ui.CursorText {
		t.Fatalf("input cursor %v", tester.Cursor())
	}
	move("拖动调整大小")
	if tester.Cursor() != ui.CursorResizeEW {
		t.Fatalf("splitter cursor %v", tester.Cursor())
	}
}

func TestButtonLoadingDoesNotSubmit(t *testing.T) {
	hits := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if Button(c, "保存", ButtonProps{Type: Primary, Loading: true}) {
				hits++
			}
		})
	}, 300, 80)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("保存"); err != nil {
		t.Fatal(err)
	}
	if hits != 0 {
		t.Fatalf("loading button submitted %d times", hits)
	}
}

func TestTableTruncatedCellTooltip(t *testing.T) {
	const full = "超长会议纪要需要在单元格里用提示看完"
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(16).Width(220).Children(func() {
				Table(c, []Column{{Title: "备注", Width: 72}}, 1, func(row, col int) Cell {
					return Cell{Text: full}
				})
			})
		})
	}, 280, 160)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	r, ok := tester.Find(full)
	if !ok {
		t.Fatal("no cell")
	}
	tester.Move(r.X+8, r.Y+r.H/2)
	tester.Advance(150 * time.Millisecond)
	tester.Frame()
	if !tester.HasText(full) {
		t.Fatalf("truncated tooltip missing: %v", tester.Texts())
	}
}
