package kit

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

var hosts = []string{"192.168.0.1", "192.168.0.23", "192.168.1.7"}

func hostCell(r, col int) Cell {
	if col == 0 {
		return Cell{Text: hosts[r]}
	}
	return Cell{Text: "在线", Color: Work.Success}
}

func TestTableSelectsRows(t *testing.T) {
	sel, changes := -1, 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if Table(c, []Column{{Title: "IP", Width: 160}, {Title: "状态"}}, len(hosts), hostCell, TableProps{Selected: &sel}) {
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
	if sel != 1 || changes != 1 {
		t.Fatalf("sel = %d after %d changes", sel, changes)
	}
	tester.Key(0, ui.KeyDown)
	tester.Frame()
	tester.Key(0, ui.KeyEnter)
	tester.Frame()
	if sel != 2 {
		t.Fatalf("arrow + Enter left sel = %d, want 2", sel)
	}
	if err := tester.Click("192.168.1.7"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if changes != 2 {
		t.Fatalf("clicking the chosen row counted as a change (%d)", changes)
	}
}

func TestTabOrderFollowsBuildAndTableIsOneStop(t *testing.T) {
	sel := -1
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Button(c, "上一项")
			Table(c, []Column{{Title: "IP"}, {Title: "状态"}}, len(hosts), hostCell, TableProps{Selected: &sel})
			Button(c, "下一项")
		})
	}, 480, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	steps := []string{"上一项", "第 1 行", "下一项"}
	for _, want := range steps {
		tester.Key(0, ui.KeyTab)
		tester.Frame()
		if !tester.Focused(want) {
			t.Fatalf("Tab did not reach %q", want)
		}
	}
}

func TestTableWithoutSelectionIgnoresClicks(t *testing.T) {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if Table(c, []Column{{Title: "IP"}, {Title: "状态", Align: AlignEnd}}, len(hosts), hostCell) {
				t.Error("a table without Selected reported a choice")
			}
		})
	}, 480, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	_ = tester.Click("192.168.0.1")
	tester.Frame()
}

func TestTableEmpty(t *testing.T) {
	sel := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Table(c, []Column{{Title: "IP"}, {Title: "MAC"}}, 0, hostCell, TableProps{Selected: &sel, Empty: "没有发现直连设备"})
			Table(c, []Column{{Title: "目标"}}, 0, hostCell)
		})
	}, 480, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	for _, want := range []string{"IP", "MAC", "没有发现直连设备", "暂无数据"} {
		if !tester.HasText(want) {
			t.Errorf("empty table missing %q", want)
		}
	}
}

func TestClampNumber(t *testing.T) {
	port := InputNumberProps{Min: 1, Max: 65535}
	cases := []struct {
		in, want string
		ok       bool
	}{
		{"8080", "8080", true},
		{" 0 ", "1", true},
		{"70000", "65535", true},
		{"443.6", "444", true},
		{"", "", true},
		{"abc", "", false},
		{"1e3", "1000", true},
	}
	for _, tc := range cases {
		got, ok := ClampNumber(tc.in, port)
		if got != tc.want || ok != tc.ok {
			t.Errorf("ClampNumber(%q) = %q, %v; want %q, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
	if got, _ := ClampNumber("-3.14159", InputNumberProps{Precision: 2}); got != "-3.14" {
		t.Errorf("unbounded precision 2 = %q", got)
	}
}

func TestStepNumber(t *testing.T) {
	prefix := InputNumberProps{Min: 1, Max: 32}
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{"24", 1, "25"},
		{"24", -1, "23"},
		{"32", 1, "32"},
		{"1", -1, "1"},
		{"", 1, "1"},
		{"x", -1, "1"},
	}
	for _, tc := range cases {
		if got := StepNumber(tc.in, tc.n, prefix); got != tc.want {
			t.Errorf("StepNumber(%q, %d) = %q, want %q", tc.in, tc.n, got, tc.want)
		}
	}
	half := InputNumberProps{Step: 0.5, Precision: 1}
	if got := StepNumber("1", 3, half); got != "2.5" {
		t.Errorf("half steps = %q", got)
	}
	if got := StepNumber("", 1, InputNumberProps{}); got != "0" {
		t.Errorf("unbounded empty start = %q", got)
	}
}

func TestInputNumberStepsAndClamps(t *testing.T) {
	port := "65534"
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			InputNumber(c, &port, InputNumberProps{Min: 1, Max: 65535})
			Button(c, "别处")
		})
	}, 400, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("增加"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if port != "65535" {
		t.Fatalf("增加 → %q", port)
	}
	_ = tester.Click("增加")
	tester.Frame()
	if port != "65535" {
		t.Fatalf("增加 past Max → %q", port)
	}
	if err := tester.Click("减少"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if port != "65534" {
		t.Fatalf("减少 → %q", port)
	}
	tester.Key(0, ui.KeyTab)
	tester.Frame()
	tester.Key(0, ui.KeyUp)
	tester.Frame()
	if port != "65535" {
		t.Fatalf("arrow up → %q", port)
	}
	port = "99999"
	tester.Frame()
	tester.Key(0, ui.KeyTab)
	tester.Frame()
	tester.Frame()
	if !tester.Focused("别处") {
		t.Fatal("Tab from the field stopped on a step button")
	}
	if port != "65535" {
		t.Fatalf("blur did not clamp: %q", port)
	}
}
