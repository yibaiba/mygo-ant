package kit

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestAntIconLookup(t *testing.T) {
	for _, name := range []string{"SearchOutlined", "HeartFilled", "CheckCircleTwoTone", "AccountBookOutlined"} {
		if LookupIcon(name) == nil {
			t.Fatalf("missing Ant icon %s", name)
		}
	}
	if LookupIcon("search") == nil {
		t.Fatal("short name search must still resolve")
	}
	if LookupIcon("AccountBook") == nil {
		t.Fatal("unsuffixed AccountBook should alias Outlined")
	}
	var nOut, nFill, nTwin int
	for _, g := range antGlyphs {
		if LookupIcon(g.name) == nil {
			t.Fatalf("generated icon missing: %s", g.name)
		}
		switch {
		case strings.HasSuffix(g.name, "TwoTone"):
			nTwin++
		case strings.HasSuffix(g.name, "Filled"):
			nFill++
		case strings.HasSuffix(g.name, "Outlined"):
			nOut++
		default:
			t.Fatalf("unexpected name %s", g.name)
		}
	}
	if nOut != 447 || nFill != 251 || nTwin != 150 {
		t.Fatalf("icon counts Outlined=%d Filled=%d TwoTone=%d", nOut, nFill, nTwin)
	}
	if len(antGlyphs) != 848 {
		t.Fatalf("antGlyphs = %d, want 848", len(antGlyphs))
	}
}

func TestAntIconThemeAndPaint(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			IconNamed(c, "SearchOutlined")
			IconNamed(c, "", Icon{Name: "HeartFilled"})
			IconNamed(c, "Heart", IconProps{Theme: IconFilled})
			IconNamed(c, "check-circle", IconProps{Theme: IconTwoTone, TwoTone: Work.Primary})
			IconNamed(c, "no-such-icon-xyz")
		})
	}, 200, 80)
	tester.Frame()
	if LookupIcon("HeartFilled") == nil {
		t.Fatal("HeartFilled")
	}
	if applyIconTheme("Heart", IconFilled) != "HeartFilled" {
		t.Fatalf("theme apply %q", applyIconTheme("Heart", IconFilled))
	}
	if applyIconTheme("search", IconOutlined) != "SearchOutlined" {
		t.Fatalf("outlined apply %q", applyIconTheme("search", IconOutlined))
	}
}

func TestMenuAndTableDefaultPainted(t *testing.T) {
	if (TableProps{}).Native || (DropdownProps{}).Native || (SplitButtonProps{}).Native {
		t.Fatal("Native must be opt-in; painted Ant is the default")
	}
	if (ConfigProps{}).FollowDesktop {
		t.Fatal("FollowDesktop must be opt-in")
	}
	got := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if k := DropdownButton(c, "更多操作", []MenuItem{
				{Key: "copy", Label: "复制路径"},
				{Key: "clear", Label: "清空日志"},
			}); k != "" {
				got = k
			}
			sel := 0
			Table(c, []Column{
				{Title: "IP", Width: 120},
				{Title: "状态", Width: 80},
			}, len(hosts), hostCell, TableProps{Selected: &sel, PageSize: 8})
		})
	}, 640, 420)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if tester.HasText("复制路径") {
		t.Fatal("painted menu should be closed until click")
	}
	if err := tester.Click("更多操作"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !tester.HasText("复制路径") || !tester.HasText("清空日志") {
		t.Fatalf("painted dropdown items missing: %v", tester.Texts())
	}
	if err := tester.Click("清空日志"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if got != "clear" {
		t.Fatalf("chose %q", got)
	}
	if !tester.HasText("192.168.0.1") {
		t.Fatal("painted table rows missing")
	}
}
