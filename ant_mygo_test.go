package kit

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestDialogWidthPresets(t *testing.T) {
	if dialogWidth(0, false, false, DialogWidth) != 520 || dialogWidth(0, true, false, DialogWidth) != 416 {
		t.Fatal("small / default")
	}
	if dialogWidth(0, false, true, DialogWidth) != 736 || dialogWidth(800, true, true, DialogWidth) != 800 {
		t.Fatal("large / Width wins")
	}
}

func TestDarkAlgorithmCachesAndConfigProvider(t *testing.T) {
	if DarkAlgorithm(Work) != WorkDark {
		t.Fatal("Work dark is cached as WorkDark")
	}
	if DarkAlgorithm(WorkDark) != WorkDark {
		t.Fatal("already dark")
	}
	homeDark := DarkAlgorithm(Home)
	if homeDark == Home || !homeDark.Dark || homeDark == WorkDark {
		t.Fatal("Home should get its own dark map")
	}
	if DarkAlgorithm(Home) != homeDark {
		t.Fatal("Home dark should cache")
	}
	var inner *Tokens
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ConfigProvider(c, ConfigProps{Dark: Bool(true)}, func() {
				inner = Of(c)
			})
		})
	}, 80, 40)
	tester.Frame()
	if inner == nil || !inner.Dark {
		t.Fatalf("ConfigProvider Dark should apply darkAlgorithm, got %+v", inner)
	}
	if hex(homeDark.Glass.Title) != hex(WorkDark.Glass.Title) || hex(homeDark.Glass.Subtle) != hex(WorkDark.Glass.Subtle) {
		t.Fatalf("Home dark glass title=%s subtle=%s, want WorkDark", hex(homeDark.Glass.Title), hex(homeDark.Glass.Subtle))
	}
	if homeDark.Glass.Card.R == 255 && homeDark.Glass.Card.A > 100 {
		t.Fatalf("Home dark glass still light: %+v", homeDark.Glass.Card)
	}
	if hex(Work.Glass.Title) != "#303640" {
		t.Fatalf("light Home/Work glass title = %s", hex(Work.Glass.Title))
	}
}

func TestMotionOffSkipsLeave(t *testing.T) {
	open, after := true, 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ConfigProvider(c, ConfigProps{Motion: Bool(false)}, func() {
				Dialog(c, &open, "宽对话框", func() {
					Muted(c, "内容")
				}, DialogProps{Small: true, AfterClose: func() { after++ }})
			})
		})
	}, 640, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("宽对话框") {
		t.Fatal("open dialog")
	}
	open = false
	tester.Frame()
	if after != 1 {
		t.Fatalf("AfterClose = %d, want 1 when motion is off", after)
	}
	if tester.HasText("宽对话框") {
		t.Fatal("leave motion should be skipped")
	}
}

func TestInputAutoFocusAndPressEnter(t *testing.T) {
	v, enters := "", 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Input(c, &v, InputProps{AutoFocus: true, OnPressEnter: func() { enters++ }})
			Input(c, new(string), InputProps{Placeholder: "第二个"})
		})
	}, 400, 120)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Type("hello")
	tester.Frame()
	if v != "hello" {
		t.Fatalf("AutoFocus typing = %q", v)
	}
	tester.Key(0, ui.KeyEnter)
	tester.Frame()
	if enters != 1 {
		t.Fatalf("OnPressEnter = %d", enters)
	}
}

func TestFormValidateTriggerBlurAndSubmit(t *testing.T) {
	var blur Form
	blur.ValidateTrigger = ValidateBlur
	host := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			blur.Begin()
			err := blur.Check("host", host, Required("请输入地址"))
			FormItem(c, "主机", func() {
				Input(c, &host, InputProps{AutoFocus: true, Status: ErrorStatus(err)})
			}, FormItemProps{Name: "host", Form: &blur, Error: err})
			Button(c, "其它", ButtonProps{})
		})
	}, 400, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	if tester.HasText("请输入地址") {
		t.Fatal("onBlur must not show before leave")
	}
	if err := tester.Click("其它"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	tester.Frame()
	if !tester.HasText("请输入地址") {
		t.Fatal("onBlur should show after leave")
	}

	var sub Form
	sub.ValidateTrigger = ValidateSubmit
	val := ""
	sub.Begin()
	if e := sub.Check("x", val, Required("必填")); e != "" {
		t.Fatalf("onSubmit before Validate: %q", e)
	}
	val = "a"
	sub.Begin()
	if e := sub.Check("x", val, Required("必填")); e != "" {
		t.Fatalf("edit still hidden: %q", e)
	}
	val = ""
	sub.Begin()
	sub.Check("x", val, Required("必填"))
	if e, _ := sub.validate(); e != "必填" {
		t.Fatalf("Validate = %q", e)
	}
	sub.Begin()
	if e := sub.Check("x", val, Required("必填")); e != "必填" {
		t.Fatalf("after submit: %q", e)
	}
}

func TestGhostButtonHasNoPrimaryFill(t *testing.T) {
	const w, h = 280, 80
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Background(Work.Surface).Padding(20).Children(func() {
				Button(c, "幽灵主色", ButtonProps{Type: Primary, Ghost: true})
			})
		})
	}, w, h)
	tester.SetLocale("zh-CN")
	tester.Frame()
	r, ok := tester.Find("幽灵主色")
	if !ok {
		t.Fatal("no ghost button")
	}
	img := tester.Image()
	if img == nil {
		t.Fatal("no frame")
	}
	sx := float32(img.Bounds().Dx()) / w
	px := img.RGBAAt(int((r.X+6)*sx), int((r.Y+r.H/2)*sx))
	if int(px.B) > int(px.R)+40 && px.R < 180 {
		t.Fatalf("ghost fill looks primary: %v at %v", px, r)
	}
}

func TestCircleButtonIsSquare(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Button(c, "搜", ButtonProps{Type: Primary, Shape: ShapeCircle})
		})
	}, 120, 80)
	tester.SetLocale("zh-CN")
	tester.Frame()
	r, ok := tester.Find("搜")
	if !ok {
		t.Fatal("no circle button")
	}
	if r.W != r.H {
		t.Fatalf("circle %vx%v, want square", r.W, r.H)
	}
}
