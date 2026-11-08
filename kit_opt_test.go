package kit

import (
	"strconv"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestFormItemLabelFocusesControl(t *testing.T) {
	name := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			FormView(c, func() {
				FormItem(c, "名称", func() {
					Input(c, &name, InputProps{Placeholder: "请输入名称"})
				}, FormItemProps{Required: true, Help: "会议室显示名", Extra: "可改"})
			})
		})
	}, 480, 240)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("名称") || !tester.HasText("会议室显示名") || !tester.HasText("可改") {
		t.Fatalf("FormItem layout lost: %v", tester.Texts())
	}
	if err := tester.Click("名称"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	active, _ := tester.TextInputActive()
	if !active {
		t.Fatal("clicking the FormItem label should focus the input")
	}
	tester.Type("第二会议室")
	tester.Frame()
	if name != "第二会议室" {
		t.Fatalf("typed into label-focused input: %q", name)
	}
}

func TestPickerClearShowsWhenFocused(t *testing.T) {
	at := day(2026, 11, 7)
	changes := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(24).Children(func() {
				if DatePicker(c, &at, DatePickerProps{Width: 220, AllowClear: true}) {
					changes++
				}
			})
		})
	}, 400, 160)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if tester.HasText("清除") {
		t.Fatal("clear × should wait for hover or focus")
	}
	r, ok := tester.Find("2026-11-07")
	if !ok {
		t.Fatal("no date field")
	}
	tester.ClickAt(r.X+20, r.Y+r.H/2)
	tester.Frame()
	tester.Move(2, 2)
	tester.Frame()
	if err := tester.Click("清除"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !at.IsZero() || changes != 1 {
		t.Fatalf("focused clear left %v after %d", at, changes)
	}
}

func TestSelectShowSearchIsCombobox(t *testing.T) {
	value := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(20).Children(func() {
				Select(c, &value, dbOptions(), SelectProps{ShowSearch: true, Placeholder: "选择数据库", Width: 240})
			})
		})
	}, 500, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.ClickAt(100, 36)
	tester.Frame()
	active, _ := tester.TextInputActive()
	if !active {
		t.Fatal("ShowSearch Select should keep ComboboxBase's text field (RoleComboBox)")
	}
	tester.Type("金")
	tester.Frame()
	if tester.HasText("MySQL") || !tester.HasText("人大金仓") {
		t.Fatalf("combobox did not filter: %v", tester.Texts())
	}
	tester.Key(0, ui.KeyEnter)
	tester.Frame()
	if value != "kingbase" {
		t.Fatalf("combobox Enter chose %q", value)
	}
}

func TestSliderAndSegmentedInheritDisabled(t *testing.T) {
	v := 40.0
	seg := "day"
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ConfigProvider(c, ConfigProps{Disabled: Bool(true)}, func() {
				ui.Column(c).Padding(20).Gap(16).Children(func() {
					Slider(c, &v)
					Segmented(c, &seg, []Option{{Value: "day", Label: "日"}, {Value: "week", Label: "周"}})
				})
			})
		})
	}, 400, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("周"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if seg != "day" {
		t.Fatalf("disabled Segmented changed to %q", seg)
	}
	h, ok := tester.Find("滑块")
	if !ok {
		t.Fatal("no slider")
	}
	tester.ClickAt(h.X+120, h.Y+h.H/2)
	tester.Frame()
	if v != 40 {
		t.Fatalf("disabled Slider moved to %v", v)
	}
}

func TestControlFontLargeIs16(t *testing.T) {
	if Work.FontSize != 14 || Work.FontSizeLG != 16 {
		t.Fatalf("tokens FontSize %v / LG %v, want 14 / 16", Work.FontSize, Work.FontSizeLG)
	}
	if Work.ControlFont(Middle) != 14 || Work.ControlFont(Small) != 14 {
		t.Fatalf("Middle/Small ControlFont %v / %v, want 14", Work.ControlFont(Middle), Work.ControlFont(Small))
	}
	if Work.ControlFont(Large) != 16 {
		t.Fatalf("Large ControlFont %v, want 16", Work.ControlFont(Large))
	}
	large := Configured(Work, ConfigProps{Size: Large})
	if large.ControlFont(SizeInherit) != 16 {
		t.Fatalf("ConfigProvider Large ControlFont %v", large.ControlFont(SizeInherit))
	}
}

func TestIconLookupDoesNotRequireEagerParse(t *testing.T) {
	fresh := &iconSrc{kit: true, primary: `<path d="M2 2h12v12H2z"/>`}
	if fresh.svg != nil {
		t.Fatal("iconSrc must stay unparsed until Lookup")
	}
	a, _ := fresh.parse()
	b, _ := fresh.parse()
	if a == nil || a != b {
		t.Fatal("parse should run once and cache")
	}
	iconRegistry.RLock()
	nSrc, nBy := len(iconRegistry.src), len(iconRegistry.byName)
	iconRegistry.RUnlock()
	if nSrc < 848 {
		t.Fatalf("lazy store src=%d, want the 848 Ant glyphs plus aliases", nSrc)
	}
	if nBy > nSrc {
		t.Fatalf("byName %d grew past src %d", nBy, nSrc)
	}
	if LookupIcon("SearchOutlined") == nil || LookupIcon("search") == nil || LookupIcon("AccountBook") == nil {
		t.Fatal("Lookup must still resolve Ant names and short aliases")
	}
}

func TestAutoVirtualThreshold(t *testing.T) {
	if virtualOver != 64 {
		t.Fatalf("virtualOver = %d, want 64 to match Select", virtualOver)
	}
	if autoVirtual(64) || autoVirtual(0) {
		t.Fatal("64 rows and under stay painted")
	}
	if !autoVirtual(65) {
		t.Fatal("65 rows should auto-virtualize")
	}
	builtList := 0
	tester := ui.NewTester(func(c *ui.Context) {
		builtList = 0
		Page(c, Work, func() {
			List(c, 80, func(i int) {
				builtList++
				ui.Text(c, "项"+strconv.Itoa(i))
			})
		})
	}, 400, 500)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if builtList >= 80 || builtList == 0 {
		t.Fatalf("auto-virtual List built %d of 80", builtList)
	}
	if !tester.HasText("项0") {
		t.Fatal("virtual List dropped the visible first row")
	}

	builtTable := 0
	small := ui.NewTester(func(c *ui.Context) {
		builtTable = 0
		Page(c, Work, func() {
			Table(c, []Column{{Title: "#"}}, 20, func(r, _ int) Cell {
				builtTable++
				return Cell{Text: "行" + strconv.Itoa(r)}
			})
		})
	}, 400, 900)
	small.SetLocale("zh-CN")
	small.Frame()
	if builtTable != 20 {
		t.Fatalf("20-row table should stay fully painted, built %d", builtTable)
	}
	for i := 0; i < 20; i++ {
		if !small.HasText("行" + strconv.Itoa(i)) {
			t.Fatalf("20-row table missing 行%d", i)
		}
	}

	builtBig := 0
	big := ui.NewTester(func(c *ui.Context) {
		builtBig = 0
		Page(c, Work, func() {
			Table(c, []Column{{Title: "#"}}, 80, func(r, _ int) Cell {
				builtBig++
				return Cell{Text: "行" + strconv.Itoa(r)}
			})
		})
	}, 400, 500)
	big.SetLocale("zh-CN")
	big.Frame()
	if builtBig >= 80 || builtBig == 0 {
		t.Fatalf("auto-virtual table built %d of 80", builtBig)
	}
	if !big.HasText("行0") {
		t.Fatal("virtual table dropped the visible first row")
	}
}

func TestKitOptShot(t *testing.T) {
	name, sel, seg := "第二会议室", "eth0", "day"
	on := true
	day := time.Date(2026, 11, 7, 0, 0, 0, 0, time.Local)
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Padding(20).Gap(16).Children(func() {
				ui.Text(c, "kit 优化").FontSize(18).FontWeight(600).TextColor(Work.Text)
				FormView(c, func() {
					FormItem(c, "名称", func() { Input(c, &name) }, FormItemProps{Required: true, Extra: "点标签聚焦"})
					FormItem(c, "网卡", func() {
						Select(c, &sel, Opts("eth0", "eth1"), SelectProps{ShowSearch: true, Size: Large})
					})
				}, FormProps{Size: Large})
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					Button(c, "大号 16", ButtonProps{Type: Primary, Size: Large})
					Button(c, "中号 14", ButtonProps{Type: Primary})
					Checkbox(c, &on, "大号勾选", ChoiceProps{Size: Large})
					Segmented(c, &seg, []Option{{Value: "day", Label: "日"}, {Value: "week", Label: "周"}})
				})
				DatePicker(c, &day, DatePickerProps{Width: 200, AllowClear: true})
				Glass(c, func() {
					ui.Text(c, "GlassCard").FontWeight(600)
					HealthPill(c, "服务安全运行中", "共 7 个服务模块", true)
				})
			})
		})
	}, 900, 640)
	tester.SetLocale("zh-CN")
	tester.Frame()
	shot(t, tester, "mygo-kit-opt.png")
}
