package kit

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestBreakpoints(t *testing.T) {
	cols := map[float32]int{600: 1, 699: 1, 700: 2, 959: 2, 960: 3, 1179: 3, 1180: 4, 1920: 4}
	for w, want := range cols {
		if got := HomeCols(w); got != want {
			t.Errorf("HomeCols(%v) = %d, want %d", w, got, want)
		}
	}
	if Wide(959) || !Wide(960) {
		t.Error("sidebar breakpoint is not 960")
	}
	if got := CellMin(1280, 160); got != 160 {
		t.Errorf("wide CellMin = %v", got)
	}
	if got := CellMin(720, 160); got != 632 {
		t.Errorf("narrow CellMin = %v, want the whole row", got)
	}
	if got := CellMin(300, 160); got != 260 {
		t.Errorf("tiny CellMin = %v", got)
	}
	if got := FormWidth(1280, true, 700); got != 700 {
		t.Errorf("FormWidth capped = %v", got)
	}
	if got := FormWidth(900, true, 860); got != 852 {
		t.Errorf("FormWidth 900 = %v", got)
	}
	if got := FormWidth(1000, true, 900); got != 802 {
		t.Errorf("FormWidth beside sidebar = %v", got)
	}
	if got := FormWidth(320, false, 700); got != 300 {
		t.Errorf("FormWidth floor = %v", got)
	}
	if PagePad(720) != 12 || PagePad(900) != 20 || PagePad(1280) != 36 {
		t.Error("PagePad steps changed")
	}
	if HomeTitleSize(720) != 24 || HomeTitleSize(900) != 28 || HomeTitleSize(1180) != 32 {
		t.Error("HomeTitleSize steps changed")
	}
	if HomeNavGap(720) != 16 || HomeNavGap(960) != 20 || HomeNavGap(1180) != 24 {
		t.Error("HomeNavGap steps changed")
	}
	if SidebarWidth != 150 || NavTileW != 110 || NavTileH != 100 || NavTileGap != 20 {
		t.Errorf("settings tiles %v×%v gap %v sidebar %v, want 110×100 gap 20 sidebar 150", NavTileW, NavTileH, NavTileGap, SidebarWidth)
	}
}

func TestMinButtonWidthFitsCJK(t *testing.T) {
	if got := MinButtonWidth("搜索", Middle); got != 88 {
		t.Errorf("short label = %v, want the 88 floor", got)
	}
	cjk := MinButtonWidth("生成服务器证书请求", Middle)
	if cjk < 9*14+30 {
		t.Errorf("9 CJK glyphs need at least %v, got %v", 9*14+30, cjk)
	}
	if MinButtonWidth("按 IP 登录", Middle) <= MinButtonWidth("登录", Middle) {
		t.Error("longer label is not wider")
	}
	if got := MinButtonWidth("小", Small); got >= 88 {
		t.Errorf("small buttons have no floor, got %v", got)
	}
}

func TestOfFollowsAppliedTokens(t *testing.T) {
	var got []*Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		got = got[:0]
		Apply(c, Home)
		got = append(got, Of(c))
		Apply(c, Work)
		got = append(got, Of(c))
		ui.Themed(c, Home.Theme()).Children(func() {
			got = append(got, Of(c))
		})
	}, 200, 100)
	tester.Frame()
	if len(got) != 3 || got[0] != Home || got[1] != Work || got[2] != Home {
		t.Fatalf("Of = %v", got)
	}
}

func TestRadioButtonsAndSegmentedWriteValues(t *testing.T) {
	mode, seg := "static", "day"
	changes := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if RadioButtons(c, &mode, []Option{{Value: "static", Label: "静态地址"}, {Value: "dhcp", Label: "自动获取 (DHCP)"}}) {
				changes++
			}
			Segmented(c, &seg, []Option{{Value: "day", Label: "日"}, {Value: "week", Label: "周"}})
		})
	}, 480, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("自动获取 (DHCP)"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if mode != "dhcp" || changes != 1 {
		t.Fatalf("mode = %q after %d changes", mode, changes)
	}
	if err := tester.Click("周"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if seg != "week" {
		t.Fatalf("seg = %q", seg)
	}
}

func TestDisabledButtonDoesNotClick(t *testing.T) {
	clicks := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if Button(c, "应用到系统", ButtonProps{Type: Primary, Disabled: true}) {
				clicks++
			}
			if Button(c, "保存", ButtonProps{Type: Primary}) {
				clicks += 10
			}
		})
	}, 300, 120)
	tester.SetLocale("zh-CN")
	tester.Frame()
	_ = tester.Click("应用到系统")
	tester.Frame()
	if err := tester.Click("保存"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if clicks != 10 {
		t.Fatalf("clicks = %d", clicks)
	}
}

func TestConfirmAnswers(t *testing.T) {
	open, answer := true, -1
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {})
		if n := Confirm(c, &open, "停止服务", "是否停止所有服务", "确定", "取消"); n >= 0 {
			answer = n
		}
	}, 600, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("确定"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if answer != 0 || open {
		t.Fatalf("answer = %d open = %v", answer, open)
	}
}
