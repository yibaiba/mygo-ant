package kit

import (
	"image"
	"image/color"
	"sync"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestFollowDesktopKeepsDefaultPointer(t *testing.T) {
	var got *Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		got = FollowDesktop(c, Home)
	}, 80, 40)
	tester.Frame()
	if got != Home {
		t.Fatalf("ordinary desktop must keep Home, got %p", got)
	}
}

func TestDesktopDarkFollowsTesterSetDark(t *testing.T) {
	var dark bool
	tester := ui.NewTester(func(c *ui.Context) {
		dark = DesktopDark(c)
	}, 40, 40)
	tester.Frame()
	if dark {
		t.Fatal("tester starts light")
	}
	tester.SetDark(true)
	if !dark {
		t.Fatal("SetDark should flip DesktopDark")
	}
	tester.SetDark(false)
	if dark {
		t.Fatal("SetDark(false) should restore light")
	}
}

func TestPageDoesNotFollowDesktopDark(t *testing.T) {
	var tok *Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Home, func() {
			tok = Of(c)
		})
	}, 40, 40)
	tester.SetDark(true)
	if tok != Home || tok.Dark {
		t.Fatal("Page stays the tokens passed in; Dark is opt-in")
	}
}

func TestConfigProviderDarkFollowsDesktopDark(t *testing.T) {
	var tok *Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		ConfigProvider(c, ConfigProps{Theme: Home, FollowDesktop: true, Dark: Bool(DesktopDark(c))}, func() {
			tok = Of(c)
			HomeBar(c, HomeBarProps{Title: "Demo App"})
			HealthPill(c, "服务安全运行中", "共 1 个", true)
			GlassCard(c, GlassCardProps{Title: "缓存服务", Subtitle: "Redis", Status: "正常", StatusOK: true})
		})
	}, 480, 280)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if tok == nil || tok.Dark || hex(tok.Glass.Title) != hex(Home.Glass.Title) {
		t.Fatalf("light glass want Home, got dark=%v title=%s", tok != nil && tok.Dark, hex(tok.Glass.Title))
	}
	tester.SetDark(true)
	if tok == nil || !tok.Dark {
		t.Fatal("OS dark should apply darkAlgorithm")
	}
	if hex(tok.Glass.Title) != hex(WorkDark.Glass.Title) || hex(tok.Glass.Subtle) != hex(WorkDark.Glass.Subtle) {
		t.Fatalf("dark glass title=%s subtle=%s, want WorkDark %s %s", hex(tok.Glass.Title), hex(tok.Glass.Subtle), hex(WorkDark.Glass.Title), hex(WorkDark.Glass.Subtle))
	}
	if tok.Glass.Card.R == 255 && tok.Glass.Card.A > 100 {
		t.Fatalf("dark glass card still light white: %+v", tok.Glass.Card)
	}
	if hex(tok.Primary) == "#ff00aa" {
		t.Fatal("OS accent must not mix into Primary")
	}
}

func TestMessageBoxAndNotifyInjection(t *testing.T) {
	var mu sync.Mutex
	var box MessageBoxOptions
	var note OSNotifyOptions
	var result MessageBoxResult
	SetMessageBox(func(o MessageBoxOptions) (MessageBoxResult, error) {
		mu.Lock()
		box = o
		mu.Unlock()
		return MessageBoxResult{Button: 0}, nil
	})
	SetOSNotify(func(o OSNotifyOptions) error {
		mu.Lock()
		note = o
		mu.Unlock()
		return nil
	})
	t.Cleanup(func() {
		SetMessageBox(nil)
		SetOSNotify(nil)
	})
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			OpenMessageBox(c, MessageBoxOptions{Title: "停止服务", Message: "是否停止", Kind: MessageQuestion, Buttons: []string{"确定", "取消"}}, func(r MessageBoxResult) {
				mu.Lock()
				result = r
				mu.Unlock()
			})
		})
	}, 200, 80)
	tester.Frame()
	if !NotifyOS(OSNotifyOptions{Title: "备份完成", Body: "已写入 /data"}) {
		t.Fatal("NotifyOS reported no injector")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		ok := box.Title == "停止服务" && result.Button == 0 && note.Title == "备份完成"
		mu.Unlock()
		if ok {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	t.Fatalf("box=%+v result=%+v note=%+v", box, result, note)
}

func TestNativeConfirmFallsBackWithoutInjector(t *testing.T) {
	SetMessageBox(nil)
	var called bool
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if NativeConfirm(c, "删除", "不可撤销", true, func(bool) { called = true }) {
				t.Error("NativeConfirm opened without a dialog")
			}
		})
	}, 120, 40)
	tester.Frame()
	if called {
		t.Fatal("done ran with no dialog")
	}
}

func TestPaginationJumperAndSizeChanger(t *testing.T) {
	cur, size := 6, 10
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Pagination(c, &cur, 95, size, PaginationProps{
				ShowTotal: true, ShowQuickJumper: true, ShowSizeChanger: true, PageSize: &size,
			})
		})
	}, 900, 80)
	tester.SetLocale("zh-CN")
	tester.Frame()
	for _, want := range []string{"共 95 条", "跳至", "页", "条/页", "6"} {
		if !tester.HasText(want) {
			t.Fatalf("missing %q in %v", want, tester.Texts())
		}
	}
}

func TestEmptyExtraAndSwitchChildren(t *testing.T) {
	on := true
	clicks := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Empty(c, "尚未扫描", EmptyProps{
				Description: "点扫描开始",
				Extra: func() {
					if Button(c, "扫描", ButtonProps{Type: Primary}) {
						clicks++
					}
				},
			})
			Switch(c, &on, ChoiceProps{CheckedChildren: "开", UnCheckedChildren: "关"})
		})
	}, 400, 280)
	tester.SetLocale("zh-CN")
	tester.Frame()
	for _, want := range []string{"尚未扫描", "点扫描开始", "扫描", "开"} {
		if !tester.HasText(want) {
			t.Fatalf("missing %q in %v", want, tester.Texts())
		}
	}
	if err := tester.Click("扫描"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if clicks != 1 {
		t.Fatalf("clicks = %d", clicks)
	}
	if err := tester.Click("开"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if on {
		t.Fatal("switch with children did not toggle")
	}
	if !tester.HasText("关") {
		t.Fatal("unchecked children missing")
	}
}

func TestSwitchTrackTokens(t *testing.T) {
	w, h, innerMin, innerMax, font := switchTrack(false, 12)
	if w != 44 || h != 22 || innerMin != 9 || innerMax != 24 || font != 12 {
		t.Fatalf("default track %v×%v inner %v/%v font %v", w, h, innerMin, innerMax, font)
	}
	w, h, innerMin, innerMax, font = switchTrack(true, 12)
	if w != 28 || h != 16 || innerMin != 6 || innerMax != 18 || font != 12 {
		t.Fatalf("small track %v×%v inner %v/%v font %v", w, h, innerMin, innerMax, font)
	}
}

func TestSwitchInnerTextMatchesAntTrack(t *testing.T) {
	on, off, plain := true, false, true
	const W, H = 320, 80
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Row(c).Fill().Background(Work.Surface).Padding(24).Gap(20).AlignItems(ui.Center).Children(func() {
				Switch(c, &on, ChoiceProps{CheckedChildren: "开", UnCheckedChildren: "关"})
				Switch(c, &off, ChoiceProps{CheckedChildren: "开", UnCheckedChildren: "关"})
				Switch(c, &plain)
			})
		})
	}, W, H)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Advance(250 * time.Millisecond)
	tester.Frame()
	kai, ok := tester.Find("开")
	if !ok {
		t.Fatal("missing 开")
	}
	guan, ok := tester.Find("关")
	if !ok {
		t.Fatal("missing 关")
	}
	if kai.W > 20 {
		t.Fatalf("开 glyph %v wide, want a 12px inner label not a segment", kai)
	}
	img := tester.Image()
	if img == nil {
		t.Fatal("no frame")
	}
	sx := float32(img.Bounds().Dx()) / W
	around := func(r ui.Rect, pad float32) ui.Rect {
		return ui.Rect{X: r.X - pad, Y: r.Y - pad, W: r.W + pad*2, H: r.H + pad*2}
	}
	track := colorBlob(img, sx, around(kai, 40), isSwitchTrack)
	if track.H < 20 || track.H > 24 {
		t.Fatalf("开 track height %v, want Ant 22", track)
	}
	if track.W < 44 || track.W > 50 {
		t.Fatalf("开 track width %v, want Ant ~44–48 not a fat 52px capsule", track)
	}
	handle := fillInsideWall(img, sx, ui.Rect{X: track.X + track.W*0.4, Y: track.Y, W: track.W * 0.6, H: track.H}, isSwitchTrack, isSwitchKnob)
	if handle.W < 16 || handle.W > 20 || handle.H < 16 || handle.H > 20 {
		t.Fatalf("handle %v inside track %v, want Ant 18×18", handle, track)
	}
	if handle.X+2 < kai.X+kai.W {
		t.Fatalf("handle %+v should sit to the right of 开 %+v", handle, kai)
	}
	offTrack := colorBlob(img, sx, around(guan, 40), isSwitchOff)
	if offTrack.W < 44 || offTrack.W > 50 {
		t.Fatalf("关 track width %v", offTrack)
	}
	if guan.X < offTrack.X+offTrack.W/2 {
		t.Fatalf("关 %+v should sit on the right of its track %+v", guan, offTrack)
	}
	plainTrack := colorBlob(img, sx, ui.Rect{X: offTrack.X + offTrack.W + 8, Y: offTrack.Y - 4, W: 60, H: offTrack.H + 8}, isSwitchTrack)
	if plainTrack.W < 42 || plainTrack.W > 46 || plainTrack.H < 20 || plainTrack.H > 24 {
		t.Fatalf("default switch %v, want Ant 44×22", plainTrack)
	}
}

func isSwitchTrack(c color.RGBA) bool {
	return int(c.B) > int(c.R)+80 && c.B > 140 && c.A > 200
}

func isSwitchOff(c color.RGBA) bool {
	return c.R > 140 && c.R < 210 && c.G > 140 && c.G < 210 && c.B > 140 && c.B < 210 &&
		abs8(c.R, c.G) < 20 && abs8(c.G, c.B) < 20 && c.A > 200
}

func isSwitchKnob(c color.RGBA) bool {
	return c.R > 240 && c.G > 240 && c.B > 240 && c.A > 200
}

func abs8(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

func fillInsideWall(img *image.RGBA, sx float32, bounds ui.Rect, wall, fill func(color.RGBA) bool) ui.Rect {
	ib := img.Bounds()
	x0, y0 := int(bounds.X*sx), int(bounds.Y*sx)
	x1, y1 := int((bounds.X+bounds.W)*sx), int((bounds.Y+bounds.H)*sx)
	if x0 < ib.Min.X {
		x0 = ib.Min.X
	}
	if y0 < ib.Min.Y {
		y0 = ib.Min.Y
	}
	if x1 > ib.Max.X {
		x1 = ib.Max.X
	}
	if y1 > ib.Max.Y {
		y1 = ib.Max.Y
	}
	minX, minY, maxX, maxY := x1, y1, x0-1, y0-1
	for y := y0; y < y1; y++ {
		left, right := -1, -1
		for x := x0; x < x1; x++ {
			if !wall(img.RGBAAt(x, y)) {
				continue
			}
			if left < 0 {
				left = x
			}
			right = x
		}
		if left < 0 {
			continue
		}
		for x := left; x <= right; x++ {
			if !fill(img.RGBAAt(x, y)) {
				continue
			}
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	if maxX < minX {
		return ui.Rect{}
	}
	return ui.Rect{
		X: float32(minX) / sx,
		Y: float32(minY) / sx,
		W: float32(maxX-minX+1) / sx,
		H: float32(maxY-minY+1) / sx,
	}
}

func colorBlob(img *image.RGBA, sx float32, bounds ui.Rect, hit func(color.RGBA) bool) ui.Rect {
	ib := img.Bounds()
	x0, y0 := int(bounds.X*sx), int(bounds.Y*sx)
	x1, y1 := int((bounds.X+bounds.W)*sx), int((bounds.Y+bounds.H)*sx)
	if x0 < ib.Min.X {
		x0 = ib.Min.X
	}
	if y0 < ib.Min.Y {
		y0 = ib.Min.Y
	}
	if x1 > ib.Max.X {
		x1 = ib.Max.X
	}
	if y1 > ib.Max.Y {
		y1 = ib.Max.Y
	}
	minX, minY, maxX, maxY := x1, y1, x0-1, y0-1
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if !hit(img.RGBAAt(x, y)) {
				continue
			}
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x > maxX {
				maxX = x
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	if maxX < minX {
		return ui.Rect{}
	}
	return ui.Rect{
		X: float32(minX) / sx,
		Y: float32(minY) / sx,
		W: float32(maxX-minX+1) / sx,
		H: float32(maxY-minY+1) / sx,
	}
}

func TestDescItemCopyable(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Descriptions(c, []DescItem{
				{Label: "IP", Value: "172.30.0.2", Copyable: true},
				{Label: "状态", Value: "在线"},
			})
		})
	}, 480, 120)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("复制"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if tester.Clipboard() != "172.30.0.2" {
		t.Fatalf("clipboard %q", tester.Clipboard())
	}
}

func TestSplitButtonAndWindowButtons(t *testing.T) {
	key := ""
	clicked := 0
	min := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			k, hit := SplitButton(c, "选择 zip", []MenuItem{{Key: "dir", Label: "选择目录（旧版）"}})
			if hit {
				clicked++
			}
			if k != "" {
				key = k
			}
			WindowButtons(c, WindowButtonsProps{
				OnMinimize: func() { min++ },
				CloseLabel: "隐藏",
			})
		})
	}, 500, 80)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("选择 zip"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if clicked != 1 {
		t.Fatalf("main click = %d", clicked)
	}
	if err := tester.Click("更多"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("选择目录（旧版）"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if key != "dir" {
		t.Fatalf("menu key %q", key)
	}
	if err := tester.Click("最小化"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if min != 1 {
		t.Fatalf("minimize = %d", min)
	}
	if !tester.HasText("隐藏") {
		t.Fatal("close label missing")
	}
}

func TestNativeTableFallsBackWhenEmpty(t *testing.T) {
	sel := -1
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Table(c, []Column{{Title: "IP"}, {Title: "MAC"}}, 0, hostCell, TableProps{Selected: &sel, Native: true, Empty: "没有发现直连设备"})
		})
	}, 480, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	for _, want := range []string{"IP", "MAC", "没有发现直连设备"} {
		if !tester.HasText(want) {
			t.Errorf("empty native table missing %q", want)
		}
	}
}

func TestNativeTableSelectsRows(t *testing.T) {
	sel, changes := -1, 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if Table(c, []Column{{Title: "IP", Width: 160}, {Title: "状态"}}, len(hosts), hostCell, TableProps{Selected: &sel, Native: true}) {
				changes++
			}
		})
	}, 480, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("192.168.0.23"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if sel != 1 {
		t.Fatalf("native sel = %d after %d changes", sel, changes)
	}
}

func TestNativeTableOKRejectsExtras(t *testing.T) {
	if nativeTableOK([]Column{{Title: "A", Wrap: true}}, TableProps{}) {
		t.Fatal("wrap should stay painted")
	}
	if nativeTableOK([]Column{{Title: "A"}}, TableProps{Expand: func(int) func() { return nil }}) {
		t.Fatal("expand should stay painted")
	}
	sel := []int{}
	if nativeTableOK([]Column{{Title: "A"}}, TableProps{RowSelection: &sel}) {
		t.Fatal("row selection should stay painted")
	}
	if !nativeTableOK([]Column{{Title: "A", Sorter: ByText(func(int) string { return "" })}}, TableProps{Selected: new(int), PageSize: 8}) {
		t.Fatal("plain sortable table should be native-ok")
	}
}

func TestPixelRatioAndDesktop(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if PixelRatio(c) < 1 {
				t.Errorf("pixel ratio %v", PixelRatio(c))
			}
			_ = Vibrancy(c)
			_ = TitleBarInsets(c)
			_ = Desktop(c)
			Copy(c, "kit")
			if Paste(c) != "kit" {
				t.Errorf("clipboard %q", Paste(c))
			}
		})
	}, 80, 40)
	tester.Frame()
}

func TestAttachMenuCompiles(t *testing.T) {
	chosen := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			btn := ui.ButtonBase(c).Label("菜单").Padding(8)
			btn.Children(func() { ui.Text(c, "菜单") })
			if k := AttachMenu(btn, []MenuItem{
				{Key: "a", Label: "复制", ShortcutKey: ui.KeyC, ShortcutMods: ui.Ctrl},
				{Divider: true},
				{Key: "b", Label: "原生", Checked: true},
				{Key: "c", Label: "单选", Radio: true, Checked: true},
			}); k != "" {
				chosen = k
			}
			_ = AttachContextMenu(ui.Box(c).Size(20, 20), []MenuItem{{Key: "x", Label: "X"}})
			_ = AttachPopoverMenu(ui.Box(c).Size(20, 20), []MenuItem{{Key: "y", Label: "Y"}})
			_ = Dropdown(c, []MenuItem{{Key: "z", Label: "Z"}}, func() bool { return Button(c, "原生下拉") }, DropdownProps{Native: true})
		})
	}, 240, 80)
	tester.SetLocale("zh-CN")
	tester.Frame()
	_ = chosen
}
