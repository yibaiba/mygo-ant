package kit

import (
	"image/color"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestWorkFieldChromeMatchesAntV5(t *testing.T) {
	if hex(Work.Border) != "#d9d9d9" {
		t.Errorf("Border = %s, want #d9d9d9", hex(Work.Border))
	}
	if Work.RadiusControl != 6 {
		t.Errorf("RadiusControl = %v, want 6", Work.RadiusControl)
	}
	if Work.ControlHeight != 32 || Work.ControlHeightLG != 40 {
		t.Errorf("heights %v / %v, want 32 / 40", Work.ControlHeight, Work.ControlHeightLG)
	}
	if Work.PaddingInput != 11 || Work.SpaceXS != 4 {
		t.Errorf("padding %v / %v, want 11 / 4", Work.PaddingInput, Work.SpaceXS)
	}
	if hex(Work.Primary) != "#006adb" {
		t.Errorf("Primary = %s, want #006adb", hex(Work.Primary))
	}
	if hex(Work.TextPlaceholder) != "#bfbfbf" {
		t.Errorf("TextPlaceholder = %s, want #bfbfbf", hex(Work.TextPlaceholder))
	}
	if hex(Work.DisabledBg) != "#f5f5f5" {
		t.Errorf("DisabledBg = %s, want #f5f5f5", hex(Work.DisabledBg))
	}
	if Work.ControlOutline.A < 40 || Work.ControlOutline.A > 60 {
		t.Errorf("ControlOutline alpha = %d, want ~51 (0.2 of primary)", Work.ControlOutline.A)
	}
	if Work.Height(Large) != 40 || Home.Height(Middle) != 32 {
		t.Errorf("Large/Middle heights %v / %v", Work.Height(Large), Home.Height(Middle))
	}
}

func TestFocusedInputGlowsPrimaryNotFatBox(t *testing.T) {
	v := ""
	const w, h = 280, 100
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Background(Work.Surface).Padding(30).Children(func() {
				Input(c, &v, InputProps{Width: 200, Size: Large, Placeholder: "例如 192.168.1.10", force: focus})
			})
		})
	}, w, h)
	tester.SetLocale("zh-CN")
	tester.Frame()
	img := tester.Image()
	if img == nil {
		t.Fatal("no frame")
	}
	sx := float32(img.Bounds().Dx()) / w
	px := func(x, y float32) color.RGBA { return img.RGBAAt(int(x*sx), int(y*sx)) }
	border := px(30, 50)
	if int(border.B) < int(border.R)+80 {
		t.Fatalf("focus border %v, want primary #006adb", border)
	}
	glow := px(28, 50)
	if glow.R > 250 && glow.G > 250 && glow.B > 250 {
		t.Fatalf("no focus glow outside the box: %v", glow)
	}
	if glow.B <= glow.R {
		t.Fatalf("focus glow %v looks like a fat dark box, want primary @ 0.2", glow)
	}
}

func TestHoverInputBorderIsPrimary(t *testing.T) {
	v := ""
	const w = 280
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Background(Work.Surface).Padding(30).Children(func() {
				Input(c, &v, InputProps{Width: 200, Size: Large, force: hover})
			})
		})
	}, w, 100)
	tester.SetLocale("zh-CN")
	tester.Frame()
	img := tester.Image()
	sx := float32(img.Bounds().Dx()) / w
	border := img.RGBAAt(int(30*sx), int(50*sx))
	want := Work.Primary
	if !nearColor(border, color.RGBA{want.R, want.G, want.B, 255}) {
		t.Fatalf("hover border %v, want primary #006adb", border)
	}
}

func TestRestInputBorderIsAntGrey(t *testing.T) {
	v := "127.0.0.1"
	const w = 280
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Background(Work.Surface).Padding(30).Children(func() {
				Input(c, &v, InputProps{Width: 200, Size: Large})
			})
		})
	}, w, 100)
	tester.SetLocale("zh-CN")
	tester.Frame()
	img := tester.Image()
	sx := float32(img.Bounds().Dx()) / w
	border := img.RGBAAt(int(30*sx), int(50*sx))
	if !nearColor(border, color.RGBA{217, 217, 217, 255}) {
		t.Fatalf("rest border %v, want #d9d9d9", border)
	}
}

func TestLargeFieldChromeFitsSuffixAndDisabledFill(t *testing.T) {
	v, pass, n, q := "root", "secret", "24", ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Background(Work.Surface).Padding(20).Gap(12).Children(func() {
				Input(c, &v, InputProps{Width: 240, Size: Large, Disabled: true})
				Input(c, &pass, InputProps{Width: 240, Size: Large, Password: true})
				InputNumber(c, &n, InputNumberProps{Width: 160, Size: Large, Min: 1, Max: 32})
				Search(c, &q, SearchProps{Width: 280, Size: Large, EnterButton: "搜索"})
			})
		})
	}, 360, 260)
	tester.SetLocale("zh-CN")
	tester.Frame()
	eye, ok := tester.Find("显示密码")
	if !ok {
		t.Fatal("missing password eye")
	}
	if eye.H > 40 {
		t.Fatalf("password eye %v overflows a 40px field", eye)
	}
	up, ok := tester.Find("增加")
	if !ok {
		t.Fatal("missing step up")
	}
	down, ok := tester.Find("减少")
	if !ok {
		t.Fatal("missing step down")
	}
	if h := down.Y + down.H - up.Y; h < 38 || h > 42 {
		t.Fatalf("InputNumber large shell %v, want 40 (up %v down %v)", h, up, down)
	}
	btn, ok := tester.Find("搜索")
	if !ok {
		t.Fatal("missing search button")
	}
	if btn.H < 38 || btn.H > 42 {
		t.Fatalf("search button %v, want 40px", btn)
	}
	img := tester.Image()
	sx := float32(img.Bounds().Dx()) / 360
	fill := img.RGBAAt(int(32*sx), int(22*sx))
	if !nearColor(fill, color.RGBA{245, 245, 245, 255}) {
		t.Fatalf("disabled fill %v, want #f5f5f5", fill)
	}
}

func TestLargeSelectSharesInputHeight(t *testing.T) {
	sel := "MySQL"
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Background(Work.Surface).Padding(20).Children(func() {
				Select(c, &sel, Opts("MySQL", "PostgreSQL"), SelectProps{Size: Large, Width: 240, force: focus})
			})
		})
	}, 320, 100)
	tester.SetLocale("zh-CN")
	tester.Frame()
	img := tester.Image()
	sx := float32(img.Bounds().Dx()) / 320
	border := img.RGBAAt(int(20*sx), int(40*sx))
	if int(border.B) < int(border.R)+80 {
		t.Fatalf("focused select border %v, want primary", border)
	}
	glow := img.RGBAAt(int(18*sx), int(40*sx))
	if glow.B <= glow.R {
		t.Fatalf("select focus glow %v missing", glow)
	}
}
