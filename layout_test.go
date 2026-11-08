package kit

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestInlineMenuOpensAndSelects(t *testing.T) {
	selected := ""
	var open []string
	tester := mount(func(c *ui.Context) {
		ui.Box(c).Width(240).Children(func() {
			Menu(c, &selected, []MenuItem{
				{Key: "sub", Label: "会议管理", Children: []MenuItem{{Key: "a", Label: "今日会议"}, {Key: "b", Label: "本周会议"}}},
				{Key: "c", Label: "系统设置"},
				{Key: "d", Label: "禁用项", Disabled: true},
			}, MenuProps{Open: &open})
		})
	}, 400, 400)
	if tester.HasText("今日会议") {
		t.Fatal("closed submenu shows its items")
	}
	if err := tester.Click("会议管理"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !slices.Equal(open, []string{"sub"}) || !tester.HasText("今日会议") {
		t.Fatalf("open = %v, texts %v", open, tester.Texts())
	}
	if err := tester.Click("本周会议"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if selected != "b" {
		t.Fatalf("selected %q", selected)
	}
	if err := tester.Click("禁用项"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if selected != "b" {
		t.Fatalf("a disabled item was chosen: %q", selected)
	}
}

func TestHorizontalMenuSubmenuAndOverflow(t *testing.T) {
	selected := "a"
	items := []MenuItem{
		{Key: "a", Label: "导航一"},
		{Key: "sub", Label: "子菜单", Children: []MenuItem{{Key: "s1", Label: "选项一"}}},
	}
	for i := 0; i < 8; i++ {
		items = append(items, MenuItem{Key: "x" + itoa(i), Label: "很长的导航项 " + itoa(i)})
	}
	tester := mount(func(c *ui.Context) {
		Menu(c, &selected, items, MenuProps{Mode: MenuHorizontal})
	}, 700, 400)
	tester.Frame()
	if !tester.HasText("导航一") || tester.HasText("很长的导航项 7") {
		t.Fatalf("overflow should fold the last items: %v", tester.Texts())
	}
	x, y := center(t, tester, "子菜单")
	tester.Move(x, y)
	tester.Advance(200 * time.Millisecond)
	if err := tester.Click("选项一"); err != nil {
		t.Fatalf("hover submenu: %v (%v)", err, tester.Texts())
	}
	tester.Frame()
	if selected != "s1" {
		t.Fatalf("selected %q", selected)
	}
	x, y = center(t, tester, "更多")
	tester.Move(x, y)
	tester.Advance(200 * time.Millisecond)
	if err := tester.Click("很长的导航项 7"); err != nil {
		t.Fatalf("overflow menu: %v (%v)", err, tester.Texts())
	}
	tester.Frame()
	if selected != "x7" {
		t.Fatalf("selected %q", selected)
	}
}

func TestDropdownSubmenuChooses(t *testing.T) {
	chosen := ""
	tester := mount(func(c *ui.Context) {
		if k := DropdownButton(c, "操作", []MenuItem{
			{Key: "a", Label: "第一项"},
			{Key: "sub", Label: "更多操作", Children: []MenuItem{{Key: "deep", Label: "三级项"}}},
		}); k != "" {
			chosen = k
		}
	}, 500, 400)
	if err := tester.Click("操作"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	x, y := center(t, tester, "更多操作")
	tester.Move(x, y)
	tester.Advance(200 * time.Millisecond)
	if err := tester.Click("三级项"); err != nil {
		t.Fatalf("%v (%v)", err, tester.Texts())
	}
	tester.Advance(300 * time.Millisecond)
	if chosen != "deep" {
		t.Fatalf("chosen %q", chosen)
	}
	if tester.HasText("第一项") {
		t.Fatal("the menu stayed open after a choice")
	}
}

func TestSiderCollapsesAndBreakpoint(t *testing.T) {
	collapsed := false
	var seen []bool
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Layout(c, func() {
				bp := ScreenLG
				Sider(c, func(col bool) {
					seen = append(seen, col)
					ui.Text(c, "侧栏内容")
				}, SiderProps{Collapsible: true, Collapsed: &collapsed, Breakpoint: &bp})
				LayoutContent(c, func() { ui.Text(c, "内容") })
			}, LayoutProps{HasSider: true})
		})
	}, 1200, 600)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("收起侧边栏"); err != nil {
		t.Fatal(err)
	}
	tester.Advance(300 * time.Millisecond)
	if !collapsed || !seen[len(seen)-1] {
		t.Fatal("trigger did not collapse the sider")
	}
	r, _ := tester.Find("展开侧边栏")
	if r.W < 70 || r.W > 90 {
		t.Fatalf("collapsed trigger is %v wide, want 80", r.W)
	}
	if err := tester.Click("展开侧边栏"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if collapsed {
		t.Fatal("trigger did not open it")
	}
	tester.SetSize(800, 600)
	tester.Frame()
	if !collapsed {
		t.Fatal("narrowing below lg should collapse it")
	}
	tester.SetSize(1200, 600)
	tester.Frame()
	if collapsed {
		t.Fatal("widening should open it again")
	}
}

func TestSplitterDragsWithinBounds(t *testing.T) {
	var sizes []float32
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Row(c).Size(602, 200).Children(func() {
				Splitter(c, []SplitPanel{
					{Render: func() { ui.Text(c, "左") }, Min: 100},
					{Render: func() { ui.Text(c, "右") }, Collapsible: true},
				}, SplitterProps{Sizes: &sizes})
			})
		})
	}, 700, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	x, y := center(t, tester, "拖动调整大小")
	tester.Press(x, y)
	tester.Move(x-100, y)
	tester.Move(x-150, y)
	tester.Release(x-150, y)
	tester.Frame()
	if len(sizes) != 2 || sizes[0] < 0.24 || sizes[0] > 0.26 {
		t.Fatalf("sizes after a 150 drag = %v, want about 0.25", sizes)
	}
	x, y = center(t, tester, "拖动调整大小")
	tester.Press(x, y)
	tester.Move(x-100, y)
	tester.Move(x-300, y)
	tester.Release(x-300, y)
	tester.Frame()
	if px := sizes[0] * 600; px < 99 || px > 101 {
		t.Fatalf("left pane %v px, want held at its min 100", px)
	}
	if err := tester.Click("向后折叠"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if sizes[1] != 0 {
		t.Fatalf("fold = %v", sizes)
	}
	if err := tester.Click("向前折叠"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if sizes[1] == 0 {
		t.Fatalf("restore = %v", sizes)
	}
}

func TestFloatButtonsAndBackTop(t *testing.T) {
	var scroll ui.ScrollState
	clicks, chosen := 0, ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Scroll(c).Fill().TrackScroll(&scroll).Children(func() {
				ui.Box(c).Size(200, 3000)
			})
			if FloatButton(c, "帮助", FloatButtonProps{Right: 100}) {
				clicks++
			}
			if k := FloatButtonGroup(c, []FloatItem{{Key: "e", Label: "编辑"}}, FloatGroupProps{Menu: true, Right: 200}); k != "" {
				chosen = k
			}
			BackTop(c, &scroll)
		})
	}, 600, 500)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("帮助"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if clicks != 1 {
		t.Fatalf("clicks %d", clicks)
	}
	if tester.HasText("编辑") {
		if _, ok := tester.Find("编辑"); ok {
			t.Fatal("a closed group shows its buttons")
		}
	}
	if err := tester.Click("展开"); err != nil {
		t.Fatal(err)
	}
	tester.Advance(300 * time.Millisecond)
	if err := tester.Click("编辑"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if chosen != "e" {
		t.Fatalf("chosen %q", chosen)
	}
	if _, ok := tester.Find("回到顶部"); ok {
		t.Fatal("back-top shows at the top")
	}
	scroll.Y = 800
	tester.Frame()
	tester.Advance(300 * time.Millisecond)
	if err := tester.Click("回到顶部"); err != nil {
		t.Fatal(err)
	}
	tester.Advance(time.Second)
	if scroll.Y > 1 {
		t.Fatalf("scroll %v after back-top", scroll.Y)
	}
}

func TestTypographyCopyEditExpand(t *testing.T) {
	title := "原标题"
	long := "这是一段很长的文字，用来测试段落的省略与展开。" + "这是一段很长的文字，用来测试段落的省略与展开。" + "这是一段很长的文字，用来测试段落的省略与展开。"
	tester := mount(func(c *ui.Context) {
		ui.Box(c).Width(300).Children(func() {
			Paragraph(c, "192.168.1.20", TypoProps{Copyable: true})
		})
		TypoText(c, "", TypoProps{Editable: &title})
		ui.Box(c).Width(300).Children(func() {
			Paragraph(c, long, TypoProps{Rows: 2, Expandable: true, Collapsible: true})
		})
	}, 500, 500)
	tester.Frame()
	cp, ok := tester.Find("复制")
	if !ok {
		t.Fatal("missing copy affordance")
	}
	// FillWidth wrap used to park the copy control in a trailing 24px
	// column. Last-line RichText children sit just after the glyphs.
	if cp.X > 180 {
		t.Fatalf("copy not after the last glyphs (trailing column?): %v", cp)
	}
	if err := tester.Click("复制"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if tester.Clipboard() != "192.168.1.20" {
		t.Fatalf("clipboard %q", tester.Clipboard())
	}
	if _, ok := tester.Find("复制成功"); !ok {
		t.Fatal("no copied mark")
	}
	tester.Advance(3500 * time.Millisecond)
	tester.Frame()
	if _, ok := tester.Find("复制成功"); ok {
		t.Fatal("copied mark stayed")
	}
	if err := tester.Click("编辑"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	tester.Key(ui.Ctrl, ui.KeyA)
	tester.Type("新标题")
	tester.Key(0, ui.KeyEnter)
	tester.Frame()
	if title != "新标题" {
		t.Fatalf("title %q", title)
	}
	if tester.HasText(long) {
		t.Fatal("paragraph not cut")
	}
	if err := tester.Click(" 展开"); err != nil {
		t.Fatalf("%v (%v)", err, tester.Texts())
	}
	tester.Frame()
	if !tester.HasText(long) {
		t.Fatal("expand did not show the whole text")
	}
	if err := tester.Click(" 收起"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if tester.HasText(long) {
		t.Fatal("collapse did not cut it again")
	}
}

func TestBreadcrumbAndStepBarClicks(t *testing.T) {
	chosen, step := "", 0
	tester := mount(func(c *ui.Context) {
		if k := Breadcrumb(c, []BreadcrumbItem{
			{Key: "home", Title: "首页"},
			{Key: "room", Title: "会议室", Menu: []MenuItem{{Key: "r2", Label: "二号会议室"}}},
			{Key: "here", Title: "详情"},
		}); k != "" {
			chosen = k
		}
		StepBar(c, &step, []StepItem{{Title: "第一步"}, {Title: "第二步"}, {Title: "第三步", Disabled: true}}, StepBarProps{Clickable: true})
	}, 600, 400)
	if err := tester.Click("首页"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if chosen != "home" {
		t.Fatalf("chosen %q", chosen)
	}
	x, y := center(t, tester, "会议室")
	tester.Move(x, y)
	tester.Advance(200 * time.Millisecond)
	if err := tester.Click("二号会议室"); err != nil {
		t.Fatalf("%v (%v)", err, tester.Texts())
	}
	tester.Frame()
	if chosen != "r2" {
		t.Fatalf("chosen %q", chosen)
	}
	if err := tester.Click("第二步"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if step != 1 {
		t.Fatalf("step %d", step)
	}
	_ = tester.Click("第三步")
	tester.Frame()
	if step != 1 {
		t.Fatal("a disabled step was chosen")
	}
}

func TestAnchorScrollsAndAffixPins(t *testing.T) {
	var s AnchorState
	pinned := false
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Row(c).Fill().Children(func() {
				ui.Column(c).Width(160).Children(func() {
					Anchor(c, &s, []AnchorLink{{Key: "a", Title: "甲"}, {Key: "b", Title: "乙"}, {Key: "c", Title: "丙"}})
				})
				sc := ui.Scroll(c).Grow(1).FillHeight().TrackScroll(&s.Scroll)
				s.Container(sc)
				sc.Children(func() {
					ui.Column(c).FillWidth().Children(func() {
						ui.Box(c).Height(50)
						pinned = Affix(c, func() { ui.Text(c, "固定条").Height(30) }, AffixProps{Target: sc, Scroll: &s.Scroll})
						for _, k := range []string{"a", "b", "c"} {
							sec := ui.Column(c.Key(k)).FillWidth().Height(500)
							s.Mark(k, sec)
						}
					})
				})
			})
		})
	}, 600, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	if pinned {
		t.Fatal("pinned at the top")
	}
	if err := tester.Click("乙"); err != nil {
		t.Fatal(err)
	}
	tester.Advance(time.Second)
	tester.Frame()
	if s.Scroll.Y < 575 || s.Scroll.Y > 585 {
		t.Fatalf("scroll %v, want section b (at 580) at the top", s.Scroll.Y)
	}
	if !pinned {
		t.Fatal("affix not pinned after scrolling past it")
	}
	r, ok := tester.Find("固定条")
	if !ok || r.Y > 2 {
		t.Fatalf("pinned bar at %v", r)
	}
	if got := anchorActive([]string{"a", "b", "c"}, s.marks, s.Scroll.Y, 0, 5); got != "b" {
		t.Fatalf("active %q", got)
	}
	s.Scroll.Y = 0
	tester.Frame()
	tester.Frame()
	if pinned {
		t.Fatal("still pinned back at the top")
	}
}

func TestTabsEditableAndOverflow(t *testing.T) {
	value := "a"
	items := []Option{{Value: "a", Label: "甲"}, {Value: "b", Label: "乙"}}
	added := 0
	tester := mount(func(c *ui.Context) {
		Tabs(c, &value, items, TabsProps{Type: TabsEditableCard,
			OnAdd: func() { added++; items = append(items, Option{Value: "n" + itoa(added), Label: "新" + itoa(added)}) },
			OnRemove: func(v string) {
				items = slices.DeleteFunc(items, func(o Option) bool { return o.Value == v })
			}})
	}, 500, 300)
	if err := tester.Click("添加标签页"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if len(items) != 3 || !tester.HasText("新1") {
		t.Fatalf("items %v", items)
	}
	if err := tester.Click("关闭 乙"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if len(items) != 2 || tester.HasText("乙") {
		t.Fatalf("items %v", items)
	}

	many := "p1"
	var opts []Option
	for i := 1; i <= 9; i++ {
		opts = append(opts, Option{Value: "p" + itoa(i), Label: "页面 " + itoa(i)})
	}
	tester = mount(func(c *ui.Context) {
		ui.Box(c).Width(400).Children(func() { Tabs(c, &many, opts) })
	}, 500, 500)
	tester.Frame()
	if tester.HasText("页面 9") {
		t.Fatal("overflowing tabs not folded")
	}
	if err := tester.Click("更多标签页"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("页面 9"); err != nil {
		t.Fatalf("%v (%v)", err, tester.Texts())
	}
	tester.Frame()
	tester.Frame()
	if many != "p9" {
		t.Fatalf("value %q", many)
	}
	if _, ok := tester.Find("页面 9"); !ok {
		t.Fatal("chosen tab not swapped into the bar")
	}
}
