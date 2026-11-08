package kit

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestToggleKey(t *testing.T) {
	cases := []struct {
		active    []string
		key       string
		accordion bool
		want      []string
	}{
		{nil, "a", false, []string{"a"}},
		{[]string{"a"}, "b", false, []string{"a", "b"}},
		{[]string{"a", "b"}, "a", false, []string{"b"}},
		{[]string{"a"}, "b", true, []string{"b"}},
		{[]string{"b"}, "b", true, []string{}},
	}
	for _, tc := range cases {
		in := slices.Clone(tc.active)
		got := toggleKey(in, tc.key, tc.accordion)
		if !slices.Equal(got, tc.want) {
			t.Errorf("toggleKey(%v, %q, %v) = %v, want %v", tc.active, tc.key, tc.accordion, got, tc.want)
		}
		if !slices.Equal(in, tc.active) {
			t.Errorf("toggleKey changed its input %v to %v", tc.active, in)
		}
	}
}

func TestCollapseOpensPanels(t *testing.T) {
	active := []string{"db"}
	changes := 0
	accordion := false
	items := []CollapseItem{
		{Key: "db", Label: "数据库", Body: func() {}},
		{Key: "net", Label: "网络"},
		{Key: "off", Label: "已锁定", Disabled: true},
	}
	var tester *ui.Tester
	tester = ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			items[0].Body = func() { Muted(c, "数据库面板") }
			items[1].Body = func() { Muted(c, "网络面板") }
			if Collapse(c, &active, items, CollapseProps{Accordion: accordion}) {
				changes++
			}
		})
	}, 500, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("数据库面板") || tester.HasText("网络面板") {
		t.Fatal("only the active panel should show")
	}
	if err := tester.Click("网络"); err != nil {
		t.Fatal(err)
	}
	tester.Advance(time.Second)
	if !slices.Equal(active, []string{"db", "net"}) || !tester.HasText("网络面板") || !tester.HasText("数据库面板") {
		t.Fatalf("active = %v after opening a second panel", active)
	}
	_ = tester.Click("已锁定")
	tester.Advance(time.Second)
	if slices.Contains(active, "off") {
		t.Fatal("a disabled panel opened")
	}
	accordion = true
	tester.Frame()
	if err := tester.Click("数据库"); err != nil {
		t.Fatal(err)
	}
	tester.Advance(time.Second)
	if err := tester.Click("数据库"); err != nil {
		t.Fatal(err)
	}
	tester.Advance(time.Second)
	if !slices.Equal(active, []string{"db"}) || tester.HasText("网络面板") {
		t.Fatalf("accordion left %v open", active)
	}
	if changes != 3 {
		t.Fatalf("changes = %d", changes)
	}
}

func TestDropdownChooses(t *testing.T) {
	got := []string{}
	items := []MenuItem{
		{Key: "refresh", Label: "刷新"},
		{Key: "copy", Label: "复制路径", Disabled: true},
		{Divider: true},
		{Key: "clear", Label: "清空日志", Danger: true},
	}
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if k := DropdownButton(c, "更多操作", items); k != "" {
				got = append(got, k)
			}
		})
	}, 500, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if tester.HasText("刷新") {
		t.Fatal("menu open before the click")
	}
	if err := tester.Click("更多操作"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	_ = tester.Click("复制路径")
	tester.Frame()
	if len(got) != 0 || !tester.HasText("清空日志") {
		t.Fatalf("a disabled entry was chosen or closed the menu: %v", got)
	}
	if err := tester.Click("清空日志"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !slices.Equal(got, []string{"clear"}) || tester.HasText("刷新") {
		t.Fatalf("got %v; the menu should close after a choice", got)
	}
	_ = tester.Click("更多操作")
	tester.Frame()
	tester.Key(0, ui.KeyEscape)
	tester.Frame()
	if tester.HasText("刷新") {
		t.Fatal("Escape did not close the menu")
	}
}

func TestDrawerCloses(t *testing.T) {
	open := true
	for _, bottom := range []bool{false, true} {
		open = true
		tester := ui.NewTester(func(c *ui.Context) {
			Page(c, Work, func() {
				Muted(c, "页面")
				Drawer(c, &open, "服务日志", func() { Muted(c, "日志内容") }, DrawerProps{Width: 420, Placement: pick(bottom, DrawerBottom, DrawerRight)})
			})
		}, 800, 600)
		tester.SetLocale("zh-CN")
		tester.Frame()
		tester.Advance(drawerMotion)
		r, ok := tester.Find("日志内容")
		if !ok {
			t.Fatal("drawer body missing")
		}
		if !bottom && r.X < 800-420 {
			t.Fatalf("right drawer body at x=%v, want inside the right 420", r.X)
		}
		if bottom && r.Y < 600-378 {
			t.Fatalf("bottom drawer body at y=%v, want inside the bottom 378", r.Y)
		}
		if err := tester.Click("关闭"); err != nil {
			t.Fatal(err)
		}
		tester.Frame()
		if open {
			t.Fatal("close button left the drawer open")
		}
		if !tester.HasText("日志内容") {
			t.Fatal("drawer vanished without sliding out")
		}
		tester.Advance(drawerMotion + 50*time.Millisecond)
		if tester.HasText("日志内容") {
			t.Fatal("close button left the drawer open")
		}
	}
}

func TestSpinOverBlocksContent(t *testing.T) {
	clicks := 0
	spinning := true
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			SpinOver(c, spinning, func() {
				if Button(c, "保存") {
					clicks++
				}
			}, SpinProps{Tip: "加载中"})
		})
	}, 400, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("加载中") {
		t.Fatal("spin tip missing")
	}
	_ = tester.Click("保存")
	tester.Frame()
	if clicks != 0 {
		t.Fatal("content took a click while spinning")
	}
	spinning = false
	tester.Frame()
	if err := tester.Click("保存"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if clicks != 1 || tester.HasText("加载中") {
		t.Fatalf("clicks = %d after spinning stopped", clicks)
	}
}

func TestResultShowsExtra(t *testing.T) {
	done := false
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Result(c, "安装完成", ResultProps{Status: Success, SubTitle: "可以继续设置安全参数", Extra: func() {
				if Button(c, "返回首页", ButtonProps{Type: Primary}) {
					done = true
				}
			}})
			Skeleton(c, SkeletonProps{Active: true, Avatar: true})
		})
	}, 600, 600)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("安装完成") || !tester.HasText("可以继续设置安全参数") {
		t.Fatal("result text missing")
	}
	if err := tester.Click("返回首页"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !done {
		t.Fatal("extra button not clicked")
	}
}
