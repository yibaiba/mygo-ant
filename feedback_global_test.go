package kit

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestProgressLogic(t *testing.T) {
	if progressStatus(100, ProgressProps{}) != ProgressSuccess {
		t.Fatal("100% should turn success")
	}
	if progressStatus(100, ProgressProps{Status: ProgressActive}) != ProgressActive {
		t.Fatal("an explicit status should stay")
	}
	if progressStatus(40, ProgressProps{Error: true}) != ProgressException {
		t.Fatal("Error should be the exception status")
	}
	segs := progressSegments(0, 360, 4, 2)
	if len(segs) != 4 || segs[0][1] != 88 || segs[1][0] != 91 {
		t.Fatalf("segments = %v", segs)
	}
	if one := progressSegments(217.5, 285, 0, 2); len(one) != 1 || one[0] != [2]float32{217.5, 285} {
		t.Fatalf("no steps = %v", one)
	}
	if progressText(30, ProgressProps{Format: func(p int) string { return "已完成" }}) != "已完成" {
		t.Fatal("Format ignored")
	}
}

func TestStillMotionFollowsReduceMotion(t *testing.T) {
	var still bool
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			still = stillMotion(c)
			Progress(c, 40, ProgressProps{Status: ProgressActive})
			StatusDot(c, "processing", "运行中")
			Spin(c)
			Skeleton(c, SkeletonProps{Active: true})
		})
	}, 400, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if still {
		t.Fatal("default should animate")
	}
	tester.SetPreferences(ui.Preferences{ReduceMotion: true})
	tester.Frame()
	if !still {
		t.Fatal("OS ReduceMotion should still Progress/Badge/Spin Now() loops")
	}
}

func TestConfiguredTokensAreCachedAndScoped(t *testing.T) {
	a := Configured(Work, ConfigProps{Size: Large, Locale: LocaleEnUS})
	b := Configured(Work, ConfigProps{Size: Large, Locale: LocaleEnUS})
	if a != b || a == Work {
		t.Fatal("same settings should give the same derived tokens")
	}
	if Configured(Work, ConfigProps{}) != Work {
		t.Fatal("no settings should keep the base tokens")
	}
	if Configured(a, ConfigProps{}) != Work {
		t.Fatal("deriving from derived tokens should start from the base")
	}
	if a.Height(SizeInherit) != 40 || a.Height(Small) != 24 || a.Lang != "en-US" {
		t.Fatalf("derived height %v lang %q", a.Height(SizeInherit), a.Lang)
	}
	var inner, outer, popup *Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			outer = Of(c)
			ConfigProvider(c, ConfigProps{Size: Small, Locale: LocaleEnUS, Disabled: Bool(true)}, func() {
				inner = Of(c)
				ui.Overlay(c, func() { popup = Of(c) })
				Pagination(c, new(int), 30, 10, PaginationProps{ShowTotal: true})
			})
		})
	}, 600, 300)
	tester.Frame()
	if outer != Work || inner.Size != Small || !inner.Disabled || L2(inner) != LocaleEnUS || popup != inner {
		t.Fatalf("outer %v inner size %v disabled %v popup same %v", outer.Name, inner.Size, inner.Disabled, popup == inner)
	}
	if !tester.HasText("Total 30 items") {
		t.Fatal("the provider's locale should reach Pagination")
	}
}

// L2 is the locale of tokens, as L reads it from a context.
func L2(t *Tokens) *Locale {
	if t.Locale != nil {
		return t.Locale
	}
	return LocaleZhCN
}

func TestConfigDisablesControls(t *testing.T) {
	clicks := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ConfigProvider(c, ConfigProps{Disabled: Bool(true)}, func() {
				if Button(c, "保存") {
					clicks++
				}
			})
		})
	}, 300, 120)
	tester.Frame()
	_ = tester.Click("保存")
	tester.Frame()
	if clicks != 0 {
		t.Fatal("a button under a disabled provider clicked")
	}
}

func TestConfigProviderInheritsSizeAndDisabled(t *testing.T) {
	var outer, inherited, middle *Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ConfigProvider(c, ConfigProps{Size: Large, Disabled: Bool(true)}, func() {
				outer = Of(c)
				ConfigProvider(c, ConfigProps{}, func() {
					inherited = Of(c)
					Button(c, "禁用大号")
				})
				ConfigProvider(c, ConfigProps{Size: Middle, Disabled: Bool(false)}, func() {
					middle = Of(c)
					Button(c, "中号可用")
				})
			})
		})
	}, 400, 200)
	tester.Frame()
	if outer == nil || outer.Size != Large || !outer.Disabled {
		t.Fatalf("outer %+v", outer)
	}
	if inherited != outer {
		t.Fatal("a nested provider that leaves Size and Disabled unset should keep the outer tokens")
	}
	if middle.Size != Middle || middle.Disabled || middle.Height(SizeInherit) != 32 {
		t.Fatalf("explicit Middle and Disabled=false should override, size %v disabled %v height %v", middle.Size, middle.Disabled, middle.Height(SizeInherit))
	}
	if inherited.Size != Large || inherited.Height(SizeInherit) != 40 {
		t.Fatalf("inherited size %v height %v", inherited.Size, inherited.Height(SizeInherit))
	}
}

func TestNotificationsTimeStackAndClose(t *testing.T) {
	var ns Notifications
	closed := []string{}
	var tester *ui.Tester
	tester = ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			NotificationHolder(c, &ns)
		})
	}, 900, 700)
	tester.SetLocale("zh-CN")
	tester.Frame()
	ns.Open(Notice{Key: "a", Kind: NoticeSuccess, Title: "备份完成", Description: "已写入", Duration: 2 * time.Second, OnClose: func() { closed = append(closed, "a") }})
	ns.Open(Notice{Key: "b", Title: "常驻提醒", Duration: -1})
	tester.Frame()
	if !tester.HasText("备份完成") || !tester.HasText("常驻提醒") {
		t.Fatal("opened notifications should show")
	}
	tester.Advance(2500 * time.Millisecond)
	tester.Frame()
	if slices.Contains(ns.Keys(), "a") || !slices.Equal(closed, []string{"a"}) {
		t.Fatalf("keys %v closed %v after its duration", ns.Keys(), closed)
	}
	ns.Open(Notice{Key: "b", Title: "常驻提醒（已更新）", Duration: -1})
	tester.Frame()
	if len(ns.Keys()) != 1 || !tester.HasText("常驻提醒（已更新）") {
		t.Fatal("opening with a key should update in place")
	}
	if err := tester.Click("关闭"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if len(ns.Keys()) != 0 {
		t.Fatal("the close button should close it")
	}
	for i := range 5 {
		ns.Open(Notice{Title: "第 " + itoa(i+1) + " 条", Duration: -1})
	}
	tester.Move(5, 690)
	tester.Frame()
	tester.Advance(time.Second)
	if !tester.HasText("第 5 条") || tester.HasText("第 1 条") {
		t.Fatal("a folded stack should show only the newest", tester.Texts())
	}
	ns.Close("")
	tester.Frame()
	if len(ns.Keys()) != 0 {
		t.Fatal("Close(\"\") should close all")
	}
}

func TestNotificationPausesOnHover(t *testing.T) {
	var ns Notifications
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() { NotificationHolder(c, &ns) })
	}, 900, 500)
	tester.Frame()
	ns.Open(Notice{Key: "p", Title: "悬停暂停", Duration: time.Second, ShowProgress: true})
	tester.Frame()
	tester.Advance(300 * time.Millisecond)
	r, ok := tester.Find("悬停暂停")
	if !ok {
		t.Fatal("notification not found")
	}
	tester.Move(r.X+r.W/2, r.Y+r.H/2)
	tester.Advance(3 * time.Second)
	if !slices.Contains(ns.Keys(), "p") {
		t.Fatal("time should stop while the pointer rests on it")
	}
	tester.Move(5, 480)
	tester.Advance(3 * time.Second)
	if slices.Contains(ns.Keys(), "p") {
		t.Fatal("time should run again once the pointer leaves")
	}
}

func TestModalMethods(t *testing.T) {
	var m Modals
	oks, cancels := 0, 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() { ModalHolder(c, &m) })
	}, 800, 600)
	tester.SetLocale("zh-CN")
	tester.Frame()
	m.Success(ModalFunc{Title: "导入成功", Content: "共 12 条", OnOK: func() { oks++ }})
	tester.Frame()
	tester.Advance(time.Second)
	if !tester.HasText("导入成功") || tester.HasText("取消") {
		t.Fatal("success should show with OK alone", tester.Texts())
	}
	_ = tester.Click("确定")
	tester.Advance(time.Second)
	if oks != 1 || m.Len() != 0 || tester.HasText("导入成功") {
		t.Fatalf("oks %d len %d after OK", oks, m.Len())
	}
	key := m.Confirm(ModalFunc{Title: "删除证书？", OKDanger: true, OnCancel: func() { cancels++ }})
	tester.Frame()
	tester.Advance(time.Second)
	m.Update(key, ModalFunc{Title: "删除 2 个证书？", OnCancel: func() { cancels++ }})
	tester.Frame()
	if !tester.HasText("删除 2 个证书？") {
		t.Fatal("Update should change the open modal")
	}
	_ = tester.Click("取消")
	tester.Advance(time.Second)
	if cancels != 1 || m.Len() != 0 {
		t.Fatalf("cancels %d len %d", cancels, m.Len())
	}
	m.Info(ModalFunc{Title: "提示"})
	m.Error(ModalFunc{Title: "失败"})
	tester.Frame()
	tester.Advance(time.Second)
	if m.Len() != 2 {
		t.Fatal("modals should stack")
	}
	m.Close("")
	tester.Frame()
	tester.Advance(time.Second)
	if m.Len() != 0 || tester.HasText("失败") {
		t.Fatal("Close(\"\") should close them all")
	}
}

func TestDialogFooterAndMaskOptions(t *testing.T) {
	open := true
	answer := -1
	noMask := true
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if n := Dialog(c, &open, "编辑", func() { ui.Text(c, "内容") }, DialogProps{OK: "保存", NoMaskClose: noMask, Top: 80}); n >= 0 {
				answer = n
			}
		})
	}, 800, 600)
	tester.SetLocale("zh-CN")
	tester.Advance(time.Second)
	tester.ClickAt(10, 590)
	tester.Advance(time.Second)
	if !open {
		t.Fatal("NoMaskClose should keep it open on a mask click")
	}
	if r, _ := tester.Find("编辑"); r.Y > 140 {
		t.Fatalf("Top 80 should put it near the top, title at %v", r.Y)
	}
	_ = tester.Click("保存")
	tester.Advance(time.Second)
	if open || answer != 0 {
		t.Fatalf("open %v answer %d after OK", open, answer)
	}
}

func TestAlertCloses(t *testing.T) {
	var closedCount int
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if Alert(c, "可关闭的提示", AlertProps{Closable: true, Action: func() { Button(c, "撤销", ButtonProps{Size: Small}) }}) {
				closedCount++
			}
			Alert(c, "文字关闭", AlertProps{CloseText: "知道了", Type: Warning})
		})
	}, 600, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("撤销") {
		t.Fatal("action should show")
	}
	if err := tester.Click("关闭"); err != nil {
		t.Fatal(err)
	}
	tester.Advance(time.Second)
	if closedCount != 1 || tester.HasText("可关闭的提示") {
		t.Fatal("the alert should close and leave")
	}
	_ = tester.Click("知道了")
	tester.Advance(time.Second)
	if tester.HasText("文字关闭") {
		t.Fatal("close text should close it")
	}
}

func TestAppHolderShowsMessages(t *testing.T) {
	var s AppState
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			App(c, &s, func() {
				if Button(c, "保存") {
					s.Message.Success("保存成功")
					s.Notification.Info("同步", "后台同步已开始")
				}
			}, AppProps{Config: &ConfigProps{Locale: LocaleEnUS}})
		})
	}, 900, 600)
	tester.Frame()
	_ = tester.Click("保存")
	tester.Advance(300 * time.Millisecond)
	if !tester.HasText("保存成功") || !tester.HasText("后台同步已开始") {
		t.Fatal("the app holder should show messages and notifications")
	}
	key := s.Message.Loading("上传中")
	tester.Frame()
	tester.Advance(5 * time.Second)
	if !tester.HasText("上传中") {
		t.Fatal("a loading message stays until closed")
	}
	s.Message.Destroy(key)
	tester.Frame()
	tester.Advance(time.Second)
	if tester.HasText("上传中") {
		t.Fatal("Destroy should close it")
	}
}

func TestWatermarkKeepsContentClickable(t *testing.T) {
	clicks := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Watermark(c, func() {
				if Button(c, "提交") {
					clicks++
				}
			}, WatermarkProps{Content: []string{"内部资料"}})
		})
	}, 400, 200)
	tester.Frame()
	_ = tester.Click("提交")
	tester.Frame()
	if clicks != 1 {
		t.Fatal("content under the watermark should take clicks")
	}
}
