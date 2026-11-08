package kit

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestFilterOptions(t *testing.T) {
	opts := []Option{{Value: "mysql", Label: "MySQL"}, {Value: "dm", Label: "达梦"}, {Value: "pg", Label: "PostgreSQL"}}
	if got := filterOptions(opts, "sql"); len(got) != 2 {
		t.Errorf("sql matched %v", got)
	}
	if got := filterOptions(opts, "达"); len(got) != 1 || got[0].Value != "dm" {
		t.Errorf("达 matched %v", got)
	}
	if got := filterOptions(opts, "  "); len(got) != 3 {
		t.Errorf("blank query filtered %v", got)
	}
}

func TestStepActiveSkipsHeadersAndDisabled(t *testing.T) {
	rows := optionRows([]Option{
		{Value: "a", Label: "A", Group: "一"},
		{Value: "b", Label: "B", Group: "一", Disabled: true},
		{Value: "c", Label: "C", Group: "二"},
	})
	// 0 header 一, 1 a, 2 b (off), 3 header 二, 4 c
	if len(rows) != 5 {
		t.Fatalf("rows %d", len(rows))
	}
	if i := stepActive(rows, -1, 1); i != 1 {
		t.Errorf("first = %d", i)
	}
	if i := stepActive(rows, 1, 1); i != 4 {
		t.Errorf("after a = %d, want c", i)
	}
	if i := stepActive(rows, 4, 1); i != 1 {
		t.Errorf("wrap = %d", i)
	}
	if i := stepActive(rows, -1, -1); i != 4 {
		t.Errorf("up from none = %d", i)
	}
	if i := stepActive(nil, -1, 1); i != -1 {
		t.Errorf("empty = %d", i)
	}
}

func TestToggleValueAndEnsureVisible(t *testing.T) {
	v := toggleValue([]string{"a", "b"}, "a")
	if !reflect.DeepEqual(v, []string{"b"}) {
		t.Errorf("remove = %v", v)
	}
	if v = toggleValue(v, "c"); !reflect.DeepEqual(v, []string{"b", "c"}) {
		t.Errorf("add = %v", v)
	}
	if y := ensureVisible(0, 100, 20, 10); y != 120 {
		t.Errorf("down = %v", y)
	}
	if y := ensureVisible(200, 100, 20, 3); y != 60 {
		t.Errorf("up = %v", y)
	}
	if y := ensureVisible(40, 100, 20, 3); y != 40 {
		t.Errorf("in view = %v", y)
	}
}

func dbOptions() []Option {
	return []Option{
		{Value: "mysql", Label: "MySQL", Group: "开源"},
		{Value: "pg", Label: "PostgreSQL", Group: "开源"},
		{Value: "dm", Label: "达梦", Group: "国产"},
		{Value: "kingbase", Label: "人大金仓", Group: "国产"},
	}
}

func TestSelectSearchAndKeys(t *testing.T) {
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
	if !tester.HasText("开源") || !tester.HasText("人大金仓") {
		t.Fatalf("popup missing groups: %v", tester.Texts())
	}
	tester.Type("金")
	tester.Frame()
	if tester.HasText("MySQL") || !tester.HasText("人大金仓") {
		t.Fatalf("search did not filter: %v", tester.Texts())
	}
	tester.Key(0, ui.KeyEnter)
	tester.Frame()
	if value != "kingbase" {
		t.Fatalf("Enter chose %q", value)
	}
	if tester.HasText("国产") {
		t.Fatal("popup stayed open")
	}
}

func TestSelectNotFoundAndClear(t *testing.T) {
	value := "dm"
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(20).Children(func() {
				Select(c, &value, dbOptions(), SelectProps{ShowSearch: true, AllowClear: true, NotFound: "没有匹配的数据库", Width: 240})
			})
		})
	}, 500, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.ClickAt(100, 36)
	tester.Frame()
	tester.Type("oracle")
	tester.Frame()
	if !tester.HasText("没有匹配的数据库") {
		t.Fatalf("no notFound: %v", tester.Texts())
	}
	tester.Key(0, ui.KeyEscape)
	tester.Frame()
	if err := tester.Click("清除"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if value != "" {
		t.Fatalf("clear left %q", value)
	}
}

func TestMultiSelectTagsAndMax(t *testing.T) {
	values := []string{"mysql", "pg", "dm"}
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(20).Children(func() {
				MultiSelect(c, &values, dbOptions(), SelectProps{Mode: Tags, MaxTagCount: 2, Width: 320})
			})
		})
	}, 500, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("+ 1 ...") || tester.HasText("达梦") {
		t.Fatalf("maxTagCount: %v", tester.Texts())
	}
	if err := tester.Click("移除 MySQL"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !reflect.DeepEqual(values, []string{"pg", "dm"}) {
		t.Fatalf("remove left %v", values)
	}
	r, _ := tester.Find("达梦")
	tester.ClickAt(r.X+r.W+40, r.Y+r.H/2)
	tester.Frame()
	tester.Type("Oracle")
	tester.Frame()
	tester.Key(0, ui.KeyEnter)
	tester.Frame()
	if !reflect.DeepEqual(values, []string{"pg", "dm", "Oracle"}) {
		t.Fatalf("tags left %v", values)
	}
	if err := tester.Click("人大金仓"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !hasValue(values, "kingbase") {
		t.Fatalf("click did not add: %v", values)
	}
}

func TestSelectVirtualBuildsFewRows(t *testing.T) {
	var opts []Option
	for i := range 5000 {
		opts = append(opts, Option{Value: strconv.Itoa(i), Label: "终端 " + strconv.Itoa(i)})
	}
	value := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(20).Children(func() {
				Select(c, &value, opts, SelectProps{Placeholder: "选择终端", Width: 240})
			})
		})
	}, 500, 500)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("选择终端"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if n := len(tester.Texts()); n > 60 {
		t.Fatalf("%d texts for 5000 options", n)
	}
	for range 3 {
		tester.Key(0, ui.KeyDown)
	}
	tester.Key(0, ui.KeyEnter)
	tester.Frame()
	if value != "2" {
		t.Fatalf("keys chose %q", value)
	}
}
