package kit

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestHomeHeaderShowsTitleExtraActions(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Home, func() {
			ui.Column(c).Fill().Padding(16).Children(func() {
				HomeHeader(c, "服务管理", HomeHeaderProps{
					TitleSize: 28,
					Extra:     func() { HealthPill(c, "1 个服务未运行", "共 7 个服务模块", false) },
					Actions: func() {
						TopNav(c, "设置")
						TopNav(c, "安装")
						TopNav(c, "连接")
					},
				})
			})
		})
	}, 960, 160)
	tester.SetLocale("zh-CN")
	tester.Frame()
	for _, want := range []string{"服务管理", "1 个服务未运行", "设置", "安装", "连接"} {
		if !tester.HasText(want) {
			t.Fatalf("missing %q in %v", want, tester.Texts())
		}
	}
	title, ok := tester.Find("服务管理")
	if !ok {
		t.Fatal("missing title")
	}
	pill, ok := tester.Find("1 个服务未运行")
	if !ok {
		t.Fatal("missing health")
	}
	if pill.X+2 < title.X+title.W {
		t.Fatalf("health should sit after the title: title=%v pill=%v", title, pill)
	}
}

func TestGlassCardStatusSwitchAndClamp(t *testing.T) {
	on := true
	changed := 0
	title := "这是一个非常非常非常非常非常非常长的Demo Service Name用来检查省略"
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Home, func() {
			ui.Column(c).Fill().Padding(12).Children(func() {
				if GlassCard(c, GlassCardProps{
					IconFn: func() {
						ui.Box(c).Size(64, 64).Radius(16).Background(ui.Hex("#D6E8FF"))
					},
					Title:    title,
					Subtitle: "会议文档与议程会写得很长所以要限制在两行以内并且不能撑破卡片",
					Status:   "正常",
					StatusOK: true,
					Extra:    "6379",
					On:       &on,
					Actions:  func() { Button(c, "日志", ButtonProps{Type: Link, Size: Small}) },
				}) {
					changed++
					on = true
				}
			})
		})
	}, 320, 280)
	tester.SetLocale("zh-CN")
	tester.Frame()
	for _, want := range []string{title, "正常", "6379", "日志"} {
		if !tester.HasText(want) {
			t.Fatalf("missing %q in %v", want, tester.Texts())
		}
	}
	name, ok := tester.Find(title)
	if !ok {
		t.Fatal("missing card title")
	}
	if name.X+name.W > 320-8 {
		t.Fatalf("title overflowed the card: %v", name)
	}
	log, ok := tester.Find("日志")
	if !ok {
		t.Fatal("missing 日志")
	}
	// Switch sits on the top row, right-aligned with the footer action.
	tester.ClickAt(log.X+log.W/2, name.Y-40)
	tester.Frame()
	if changed != 1 {
		t.Fatalf("switch changes = %d title=%v log=%v", changed, name, log)
	}
}

func TestGlassCardLoadingHidesBody(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Home, func() {
			GlassCard(c, GlassCardProps{Loading: true, Title: "缓存服务", Status: "正常"})
		})
	}, 320, 240)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if tester.HasText("缓存服务") || tester.HasText("正常") {
		t.Fatalf("loading card leaked body text: %v", tester.Texts())
	}
}

func TestHomeBarBrandAndNav(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Home, func() {
			HomeBar(c, HomeBarProps{
				Title:    "Demo App",
				Subtitle: "高效 · 稳定 · 安全",
				Width:    1280,
				Nav: func() {
					TopNav(c, "设置")
					TopNav(c, "安装")
					TopNav(c, "连接")
				},
			})
		})
	}, 1280, 120)
	tester.SetLocale("zh-CN")
	tester.Frame()
	for _, want := range []string{"Demo App", "高效 · 稳定 · 安全", "设置", "安装", "连接"} {
		if !tester.HasText(want) {
			t.Fatalf("missing %q in %v", want, tester.Texts())
		}
	}
}

func TestGlassCardDarkUsesWorkDarkGlass(t *testing.T) {
	var tok *Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		ConfigProvider(c, ConfigProps{Theme: Home, Dark: Bool(true)}, func() {
			tok = Of(c)
			GlassCard(c, GlassCardProps{Title: "缓存服务", Subtitle: "Redis", Status: "正常", StatusOK: true})
			HealthPill(c, "服务安全运行中", "共 1 个", true)
			HomeBar(c, HomeBarProps{Title: "Demo App", Subtitle: "高效 · 稳定 · 安全"})
		})
	}, 480, 320)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if tok == nil || !tok.Dark {
		t.Fatal("Dark ConfigProvider should darken Home")
	}
	if hex(tok.Glass.Title) != hex(WorkDark.Glass.Title) {
		t.Fatalf("Glass.Title = %s, want WorkDark %s", hex(tok.Glass.Title), hex(WorkDark.Glass.Title))
	}
	for _, want := range []string{"缓存服务", "服务安全运行中", "Demo App"} {
		if !tester.HasText(want) {
			t.Fatalf("missing %q in %v", want, tester.Texts())
		}
	}
}

func TestServiceCardAlias(t *testing.T) {
	on := false
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Home, func() {
			ServiceCard(c, GlassCardProps{Title: "缓存服务", Status: "已停止", Extra: "6379", On: &on})
		})
	}, 360, 240)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("缓存服务") || !tester.HasText("已停止") {
		t.Fatalf("ServiceCard missing body: %v", tester.Texts())
	}
}
