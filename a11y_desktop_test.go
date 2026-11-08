package kit

import (
	"image"
	"image/color"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestMessagesAnnounceNewCapsule(t *testing.T) {
	var m MessageAPI
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {})
		Messages(c, MessagesProps{API: &m})
	}, 600, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	_ = tester.Announcements()
	m.Info("备份完成")
	tester.Frame()
	if got := tester.Announcements(); !slices.Contains(got, "备份完成") {
		t.Fatalf("Announce missing 备份完成: %v", got)
	}
	tester.Frame()
	if got := tester.Announcements(); len(got) != 0 {
		t.Fatalf("announced the same capsule again: %v", got)
	}
}

func TestFormItemFieldErrorIsFindable(t *testing.T) {
	host := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			FormItem(c, "主机", func() {
				Input(c, &host, InputProps{Status: StatusError})
			}, FormItemProps{Required: true, Error: "请输入 IPv4 地址", Help: "会议室显示名", Extra: "可改", Warning: "将覆盖"})
		})
	}, 480, 280)
	tester.SetLocale("zh-CN")
	tester.Frame()
	for _, want := range []string{"请输入 IPv4 地址", "会议室显示名", "可改", "将覆盖"} {
		if !tester.HasText(want) {
			t.Fatalf("missing %q in %v", want, tester.Texts())
		}
	}
}

func TestWindowButtonsScaleHairlineAndMaximize(t *testing.T) {
	max, name, sel := 0, "", 0
	const W, H = 640, 400
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Padding(12).Gap(12).Children(func() {
				ui.Row(c).FillWidth().Justify(ui.End).Children(func() {
					WindowButtons(c, WindowButtonsProps{OnMaximize: func() { max++ }})
				})
				Input(c, &name, InputProps{Prefix: "输入名称", Width: 220})
				Table(c, []Column{{Title: "IP", Width: 160}, {Title: "状态", Width: 80}}, len(hosts), hostCell, TableProps{Selected: &sel})
			})
		})
	}, W, H)
	tester.SetLocale("zh-CN")
	tester.Frame()
	for _, scale := range []float32{1, 1.25, 1.5, 2} {
		max = 0
		tester.SetScale(scale)
		if _, ok := tester.Find("最大化"); !ok {
			t.Fatalf("scale %v: no maximize", scale)
		}
		if !tester.HasText("192.168.0.1") {
			t.Fatalf("scale %v: table missing: %v", scale, tester.Texts())
		}
		if _, ok := tester.Find("输入名称"); !ok {
			t.Fatalf("scale %v: no input in %v", scale, tester.Texts())
		}
		if err := tester.Click("最大化"); err != nil {
			t.Fatalf("scale %v: %v", scale, err)
		}
		tester.Frame()
		if max != 1 {
			t.Fatalf("scale %v: maximize clicks %d", scale, max)
		}
	}
	tester.SetScale(1.25)
	btn, ok := tester.Find("最大化")
	if !ok {
		t.Fatal("no maximize at 1.25")
	}
	img := tester.Image()
	if img == nil {
		t.Fatal("no frame")
	}
	if !maximizeHairlineSharp(img, float32(img.Bounds().Dx())/W, btn) {
		t.Fatal("maximize box edges are half-pixel blur at scale 1.25")
	}
}

func maximizeHairlineSharp(img *image.RGBA, sx float32, btn ui.Rect) bool {
	ib := img.Bounds()
	cx := int((btn.X + btn.W/2) * sx)
	y0 := int((btn.Y + 4) * sx)
	y1 := int((btn.Y + btn.H - 4) * sx)
	if cx < ib.Min.X || cx >= ib.Max.X || y0 < ib.Min.Y {
		return false
	}
	if y1 > ib.Max.Y {
		y1 = ib.Max.Y
	}
	first := 0
	for y := y0; y < y1; y++ {
		if isWindowGlyph(img.RGBAAt(cx, y)) {
			first = y
			break
		}
	}
	if first == 0 {
		return false
	}
	mids := 0
	for y := first - 1; y >= y0 && y >= ib.Min.Y; y-- {
		if !isWindowMid(img.RGBAAt(cx, y)) {
			break
		}
		mids++
	}
	return mids <= 1
}

func isWindowGlyph(c color.RGBA) bool {
	return abs8(c.R, 0x6B) < 36 && abs8(c.G, 0x72) < 36 && abs8(c.B, 0x80) < 36 && c.A > 180
}

func isWindowMid(c color.RGBA) bool {
	return c.R > 140 && c.R < 200 && c.G > 145 && c.G < 205 && c.B > 150 && c.B < 210 && c.A > 180
}

func TestMentionsComposingFlipsPopupAboveCaret(t *testing.T) {
	value := ""
	opts := []Option{{Value: "alice", Label: "Alice"}, {Value: "bob", Label: "Bob"}}
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(160, 16, 16).Children(func() {
				ui.Box(c).Width(360).Children(func() { Mentions(c, &value, opts, MentionsProps{Rows: 3}) })
			})
		})
	}, 600, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.ClickAt(100, 180)
	tester.Type("@")
	tester.Frame()
	tester.Frame()
	alice, ok := tester.Find("Alice")
	caret, _ := tester.TextCaret()
	if !ok || alice.Y < caret.Y+caret.H-2 {
		t.Fatalf("committed @ should open under the caret: popup %v caret %v texts %v", alice, caret, tester.Texts())
	}
	tester.Compose("zhang", 5)
	tester.Frame()
	alice, ok = tester.Find("Alice")
	caret, _ = tester.TextCaret()
	if !ok {
		t.Fatalf("popup closed while composing: %v", tester.Texts())
	}
	if alice.Y+alice.H > caret.Y+2 {
		t.Fatalf("composing should flip the popup above the caret: popup %v caret %v", alice, caret)
	}
}

func TestSelectComposingFlipsPopupAboveCaret(t *testing.T) {
	value := ""
	opts := []Option{{Value: "alice", Label: "Alice"}, {Value: "bob", Label: "Bob"}}
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(160, 16, 16).Children(func() {
				Select(c, &value, opts, SelectProps{ShowSearch: true, Placeholder: "选择成员", Width: 240})
			})
		})
	}, 400, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.ClickAt(80, 176)
	tester.Frame()
	alice, ok := tester.Find("Alice")
	caret, _ := tester.TextCaret()
	if !ok || alice.Y < caret.Y+caret.H-2 {
		t.Fatalf("select popup should open under the field: popup %v caret %v texts %v", alice, caret, tester.Texts())
	}
	tester.Compose("ali", 3)
	tester.Frame()
	alice, ok = tester.Find("Alice")
	caret, _ = tester.TextCaret()
	if !ok {
		t.Fatalf("select popup closed while composing: %v", tester.Texts())
	}
	if alice.Y+alice.H > caret.Y+2 {
		t.Fatalf("composing should flip select above the caret: popup %v caret %v", alice, caret)
	}
}

func TestFollowDesktopHighContrastKeepsPrimary(t *testing.T) {
	var got *Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		got = FollowDesktop(c, Work)
	}, 80, 40)
	tester.Frame()
	if got != Work {
		t.Fatalf("ordinary desktop must keep Work, got %p", got)
	}
	tester.SetPreferences(ui.Preferences{HighContrast: true, Accent: ui.Hex("#ff00aa"), TextScale: 1})
	if hex(got.Primary) != hex(Work.Primary) {
		t.Fatalf("HC Primary = %s, want %s", hex(got.Primary), hex(Work.Primary))
	}
	if got.ControlOutline.A != 255 {
		t.Fatalf("HC ControlOutline alpha %d, want opaque", got.ControlOutline.A)
	}
	if hex(got.PageBg) != "#ffffff" || hex(got.Surface) != "#ffffff" || hex(got.Text) != "#000000" || hex(got.Border) != "#000000" {
		t.Fatalf("HC tokens page=%s surface=%s text=%s border=%s", hex(got.PageBg), hex(got.Surface), hex(got.Text), hex(got.Border))
	}
	if got == Work {
		t.Fatal("HC must copy tokens")
	}
}

func TestFollowDesktopHighContrastDarkKeepsPrimary(t *testing.T) {
	var got *Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		got = FollowDesktop(c, WorkDark)
	}, 80, 40)
	tester.SetDark(true)
	tester.Frame()
	if got != WorkDark {
		t.Fatalf("ordinary dark desktop must keep WorkDark, got %p", got)
	}
	tester.SetPreferences(ui.Preferences{HighContrast: true, Accent: ui.Hex("#ff00aa"), TextScale: 1})
	if hex(got.Primary) != hex(WorkDark.Primary) {
		t.Fatalf("dark HC Primary = %s, want WorkDark %s", hex(got.Primary), hex(WorkDark.Primary))
	}
	if hex(Work.Primary) != "#006adb" {
		t.Fatalf("light Work Primary = %s, want #006adb", hex(Work.Primary))
	}
	if got.ControlOutline.A != 255 {
		t.Fatalf("dark HC ControlOutline alpha %d, want opaque", got.ControlOutline.A)
	}
	if hex(got.PageBg) != "#000000" || hex(got.Surface) != "#000000" || hex(got.Text) != "#ffffff" || hex(got.Border) != "#ffffff" {
		t.Fatalf("dark HC tokens page=%s surface=%s text=%s border=%s", hex(got.PageBg), hex(got.Surface), hex(got.Text), hex(got.Border))
	}
	if got == WorkDark {
		t.Fatal("dark HC must copy tokens")
	}
}

func TestConfigProviderDarkFollowDesktopHighContrast(t *testing.T) {
	var got *Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		ConfigProvider(c, ConfigProps{Theme: Work, FollowDesktop: true, Dark: Bool(true)}, func() {
			got = Of(c)
		})
	}, 80, 40)
	tester.SetPreferences(ui.Preferences{HighContrast: true, Accent: ui.Hex("#ff00aa"), TextScale: 1})
	if got == nil || !got.Dark {
		t.Fatal("Dark+FollowDesktop should stay dark")
	}
	if hex(got.PageBg) != "#000000" || hex(got.Surface) != "#000000" || hex(got.Text) != "#ffffff" || hex(got.Border) != "#ffffff" {
		t.Fatalf("HC+Dark page=%s surface=%s text=%s border=%s", hex(got.PageBg), hex(got.Surface), hex(got.Text), hex(got.Border))
	}
	if hex(got.Primary) == "#ff00aa" {
		t.Fatal("OS accent must not mix into Primary")
	}
	if hex(Work.Primary) != "#006adb" {
		t.Fatalf("Work Primary = %s, want #006adb", hex(Work.Primary))
	}
	if hex(got.Primary) != hex(WorkDark.Primary) {
		t.Fatalf("HC+Dark Primary = %s, want WorkDark %s", hex(got.Primary), hex(WorkDark.Primary))
	}
}

func TestVibrancyTrueAfterSetVibrancy(t *testing.T) {
	var on bool
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			on = c.Vibrancy()
			Glass(c, func() { ui.Text(c, "玻璃卡片") })
		})
	}, 240, 120)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if on {
		t.Fatal("Vibrancy default")
	}
	tester.SetVibrancy(true)
	if !on {
		t.Fatal("Vibrancy() false after SetVibrancy")
	}
}
