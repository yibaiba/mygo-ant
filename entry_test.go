package kit

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// mount is a zh-CN tester with fn in a padded work page.
func mount(fn func(c *ui.Context), w, h int) *ui.Tester {
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(20).Gap(16).Children(func() { fn(c) })
		})
	}, w, h)
	tester.SetLocale("zh-CN")
	tester.Frame()
	return tester
}

func center(t *testing.T, tester *ui.Tester, label string) (float32, float32) {
	t.Helper()
	r, ok := tester.Find(label)
	if !ok {
		t.Fatalf("no %q in %v", label, tester.Texts())
	}
	return r.X + r.W/2, r.Y + r.H/2
}

func TestTextAreaCountAndMax(t *testing.T) {
	value := ""
	tester := mount(func(c *ui.Context) {
		TextArea(c, &value, TextAreaProps{Rows: 3, ShowCount: true, MaxLength: 5, Width: 300})
	}, 400, 300)
	if !tester.HasText("0 / 5") {
		t.Fatalf("no count: %v", tester.Texts())
	}
	tester.ClickAt(100, 40)
	tester.Type("一二三四五六七")
	tester.Frame()
	if value != "一二三四五" {
		t.Fatalf("maxLength kept %q", value)
	}
	if !tester.HasText("5 / 5") {
		t.Fatalf("count = %v", tester.Texts())
	}
}

func TestSearchEnterAndButton(t *testing.T) {
	value, searches := "", 0
	tester := mount(func(c *ui.Context) {
		if Search(c, &value, SearchProps{Placeholder: "搜索服务", EnterButton: "搜索", Width: 300}) {
			searches++
		}
	}, 400, 200)
	tester.ClickAt(80, 36)
	tester.Type("nginx")
	tester.Key(0, ui.KeyEnter)
	tester.Frame()
	if value != "nginx" || searches != 1 {
		t.Fatalf("enter: %q %d", value, searches)
	}
	if err := tester.Click("搜索"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if searches != 2 {
		t.Fatalf("button searches = %d", searches)
	}
}

func TestOTPTyping(t *testing.T) {
	code := ""
	tester := mount(func(c *ui.Context) {
		OTP(c, &code, OTPProps{Length: 4})
	}, 400, 200)
	x, y := center(t, tester, "验证码第 1 位")
	tester.ClickAt(x, y)
	tester.Type("1")
	tester.Frame()
	tester.Type("2")
	tester.Frame()
	if code != "12" {
		t.Fatalf("typed %q", code)
	}
	if !tester.Focused("验证码第 3 位") {
		t.Fatal("focus did not move on")
	}
	// Backspace in the empty third box clears the second and goes there.
	tester.Key(0, ui.KeyBackspace)
	tester.Frame()
	tester.Frame()
	if code != "1" || !tester.Focused("验证码第 2 位") {
		t.Fatalf("backspace from an empty box left %q", code)
	}
}

func TestCheckboxGroupAndCheckAll(t *testing.T) {
	opts := []Option{{Value: "a", Label: "苹果"}, {Value: "b", Label: "香蕉"}, {Value: "c", Label: "樱桃"}}
	values := []string{"a"}
	tester := mount(func(c *ui.Context) {
		CheckAll(c, &values, opts, "全选")
		CheckboxGroup(c, &values, opts)
	}, 400, 200)
	if err := tester.Click("樱桃"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !reflect.DeepEqual(values, []string{"a", "c"}) {
		t.Fatalf("group = %v", values)
	}
	if err := tester.Click("全选"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if len(values) != 3 {
		t.Fatalf("check all = %v", values)
	}
	if err := tester.Click("全选"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if len(values) != 0 {
		t.Fatalf("uncheck all = %v", values)
	}
}

func TestSliderKeysAndPress(t *testing.T) {
	v := 30.0
	rng := [2]float64{20, 60}
	tester := mount(func(c *ui.Context) {
		ui.Box(c).Width(300).Children(func() { Slider(c, &v, SliderProps{Step: 10, Tooltip: TooltipNever}) })
		ui.Box(c).Width(300).Children(func() { RangeSlider(c, &rng, SliderProps{Tooltip: TooltipNever}) })
	}, 400, 200)
	x, y := center(t, tester, "滑块")
	tester.ClickAt(x, y)
	tester.Key(0, ui.KeyRight)
	tester.Frame()
	if v != 40 {
		t.Fatalf("right = %v", v)
	}
	at40, _ := center(t, tester, "滑块")
	tester.Key(0, ui.KeyEnd)
	tester.Frame()
	if v != 100 {
		t.Fatalf("end = %v", v)
	}
	tester.ClickAt(at40, y)
	tester.Frame()
	if v != 40 {
		t.Fatalf("press on the track at 40 gave %v", v)
	}
	ex, ey := center(t, tester, "结束值")
	tester.ClickAt(ex, ey)
	tester.Key(0, ui.KeyLeft)
	tester.Frame()
	if rng != [2]float64{20, 59} {
		t.Fatalf("range = %v", rng)
	}
}

func TestRateClickAndClear(t *testing.T) {
	v := 0.0
	tester := mount(func(c *ui.Context) {
		Rate(c, &v, RateProps{AllowHalf: true})
	}, 400, 200)
	r, ok := tester.Find("评分")
	if !ok {
		t.Fatal("no rate")
	}
	tester.ClickAt(r.X+2*(20+rateGap)+15, r.Y+r.H/2)
	tester.Frame()
	if v != 3 {
		t.Fatalf("third star = %v", v)
	}
	tester.ClickAt(r.X+2*(20+rateGap)+4, r.Y+r.H/2)
	tester.Frame()
	if v != 2.5 {
		t.Fatalf("half = %v", v)
	}
	tester.ClickAt(r.X+2*(20+rateGap)+4, r.Y+r.H/2)
	tester.Frame()
	if v != 0 {
		t.Fatalf("clear = %v", v)
	}
}

func regions() []TreeNode {
	return []TreeNode{
		{Key: "zj", Title: "浙江", Children: []TreeNode{
			{Key: "hz", Title: "杭州", Children: []TreeNode{{Key: "xh", Title: "西湖"}, {Key: "bj", Title: "滨江"}}},
		}},
		{Key: "js", Title: "江苏", Children: []TreeNode{{Key: "nj", Title: "南京"}}},
	}
}

func TestCascaderPick(t *testing.T) {
	var path []string
	tester := mount(func(c *ui.Context) {
		Cascader(c, &path, regions(), CascaderProps{SelectProps: SelectProps{Width: 240}})
	}, 600, 400)
	tester.ClickAt(100, 36)
	tester.Frame()
	for _, step := range []string{"浙江", "杭州", "滨江"} {
		if err := tester.Click(step); err != nil {
			t.Fatal(err)
		}
		tester.Frame()
	}
	if !reflect.DeepEqual(path, []string{"zj", "hz", "bj"}) {
		t.Fatalf("path = %v", path)
	}
	if !tester.HasText("浙江 / 杭州 / 滨江") {
		t.Fatalf("shown = %v", tester.Texts())
	}
}

func TestCascaderSearchPick(t *testing.T) {
	var path []string
	tester := mount(func(c *ui.Context) {
		Cascader(c, &path, regions(), CascaderProps{SelectProps: SelectProps{Width: 240, ShowSearch: true}})
	}, 600, 400)
	tester.ClickAt(100, 36)
	tester.Frame()
	tester.Type("南")
	tester.Frame()
	if err := tester.Click("江苏 / 南京"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !reflect.DeepEqual(path, []string{"js", "nj"}) {
		t.Fatalf("path = %v", path)
	}
}

func TestCascaderLoadData(t *testing.T) {
	nodes := []TreeNode{{Key: "zj", Title: "浙江"}}
	var path []string
	loads := 0
	tester := mount(func(c *ui.Context) {
		Cascader(c, &path, nodes, CascaderProps{
			SelectProps: SelectProps{Width: 240},
			LoadData: func(selected []TreeNode) {
				loads++
				if len(selected) == 1 && selected[0].Key == "zj" {
					nodes[0].Children = []TreeNode{{Key: "hz", Title: "杭州", IsLeaf: true}}
				}
			},
		})
	}, 600, 400)
	tester.ClickAt(100, 36)
	tester.Frame()
	if err := tester.Click("浙江"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if loads != 1 {
		t.Fatalf("LoadData calls = %d", loads)
	}
	if !tester.HasText("杭州") {
		t.Fatalf("children missing: %v", tester.Texts())
	}
	if err := tester.Click("杭州"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !reflect.DeepEqual(path, []string{"zj", "hz"}) {
		t.Fatalf("path = %v", path)
	}
}

func TestTreeSelectLoadData(t *testing.T) {
	nodes := []TreeNode{{Key: "zj", Title: "浙江"}}
	var value string
	loads := 0
	tester := mount(func(c *ui.Context) {
		TreeSelect(c, &value, nodes, TreeSelectProps{
			SelectProps: SelectProps{Width: 240},
			LoadData: func(key string) {
				loads++
				if key == "zj" {
					nodes[0].Children = []TreeNode{{Key: "hz", Title: "杭州", IsLeaf: true}}
				}
			},
		})
	}, 500, 400)
	tester.ClickAt(100, 36)
	tester.Frame()
	if err := tester.Click("展开 浙江"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if loads != 1 || !tester.HasText("杭州") {
		t.Fatalf("loads=%d texts=%v", loads, tester.Texts())
	}
	if err := tester.Click("杭州"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if value != "hz" {
		t.Fatalf("value = %q", value)
	}
}

func TestTreeMultiSelectCheckable(t *testing.T) {
	var values []string
	tester := mount(func(c *ui.Context) {
		TreeMultiSelect(c, &values, regions(), TreeSelectProps{SelectProps: SelectProps{Width: 300}, Checkable: true, DefaultExpandAll: true})
	}, 600, 500)
	tester.ClickAt(100, 36)
	tester.Frame()
	x, y := center(t, tester, "杭州")
	tester.ClickAt(x, y)
	tester.Frame()
	if !reflect.DeepEqual(values, []string{"zj", "hz", "xh", "bj"}) {
		t.Fatalf("checked = %v", values)
	}
	if !tester.HasText("移除 西湖") || !tester.HasText("移除 滨江") {
		t.Fatalf("chips = %v", tester.Texts())
	}
	if err := tester.Click("移除 西湖"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !reflect.DeepEqual(values, []string{"bj"}) {
		t.Fatalf("after remove = %v", values)
	}
}

func TestTreeExpandSelectCheck(t *testing.T) {
	var selected string
	var checked []string
	tester := mount(func(c *ui.Context) {
		Tree(c, regions(), TreeProps{Checkable: true, Checked: &checked, Selected: &selected})
	}, 400, 400)
	if tester.HasText("杭州") {
		t.Fatal("collapsed tree shows children")
	}
	if err := tester.Click("展开 浙江"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !tester.HasText("杭州") {
		t.Fatalf("expand: %v", tester.Texts())
	}
	// The checkbox comes first and is 16 wide; the title button follows.
	r, _ := tester.Find("江苏")
	tester.ClickAt(r.X+r.W+20, r.Y+r.H/2)
	tester.Frame()
	if selected != "js" {
		t.Fatalf("selected %q", selected)
	}
	x, y := center(t, tester, "杭州")
	tester.ClickAt(x, y)
	tester.Frame()
	if !reflect.DeepEqual(checked, []string{"zj", "hz", "xh", "bj"}) {
		t.Fatalf("checked = %v", checked)
	}
}

func TestAutoCompleteFilterAndEnter(t *testing.T) {
	value := ""
	opts := []Option{{Value: "admin@corp.cn"}, {Value: "audit@corp.cn"}, {Value: "ops@corp.cn"}}
	tester := mount(func(c *ui.Context) {
		AutoComplete(c, &value, opts, AutoCompleteProps{Width: 260, Placeholder: "邮箱"})
	}, 500, 400)
	tester.ClickAt(80, 36)
	tester.Type("a")
	tester.Frame()
	if !tester.HasText("audit@corp.cn") || tester.HasText("ops@corp.cn") {
		t.Fatalf("filter: %v", tester.Texts())
	}
	tester.Key(0, ui.KeyDown)
	tester.Frame()
	tester.Key(0, ui.KeyDown)
	tester.Frame()
	tester.Key(0, ui.KeyEnter)
	tester.Frame()
	tester.Frame()
	if value != "audit@corp.cn" {
		t.Fatalf("enter chose %q", value)
	}
}

func TestMentionsPick(t *testing.T) {
	value := ""
	opts := []Option{{Value: "alice", Label: "Alice"}, {Value: "bob", Label: "Bob"}}
	tester := mount(func(c *ui.Context) {
		Mentions(c, &value, opts, MentionsProps{Rows: 2})
	}, 500, 400)
	tester.ClickAt(100, 40)
	tester.Type("hi @b")
	tester.Frame()
	if !tester.HasText("Bob") || tester.HasText("Alice") {
		t.Fatalf("popup: %v", tester.Texts())
	}
	if err := tester.Click("Bob"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	tester.Frame()
	if value != "hi @bob " {
		t.Fatalf("value %q", value)
	}
	if got := MentionsOf(value); !reflect.DeepEqual(got, []string{"bob"}) {
		t.Fatalf("mentions %v", got)
	}
}

func TestTransferMoves(t *testing.T) {
	data := []TransferItem{{Key: "1", Title: "服务一"}, {Key: "2", Title: "服务二"}, {Key: "3", Title: "服务三"}}
	var target []string
	tester := mount(func(c *ui.Context) {
		Transfer(c, data, &target, TransferProps{Titles: [2]string{"可选", "已选"}})
	}, 700, 400)
	for _, s := range []string{"服务一", "服务三"} {
		if err := tester.Click(s); err != nil {
			t.Fatal(err)
		}
		tester.Frame()
	}
	if err := tester.Click("移到右侧"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !reflect.DeepEqual(target, []string{"1", "3"}) {
		t.Fatalf("target = %v", target)
	}
	if !tester.HasText("2 项") {
		t.Fatalf("count: %v", tester.Texts())
	}
}

func TestDatePickerMonthAndWeek(t *testing.T) {
	var month, week time.Time
	tester := mount(func(c *ui.Context) {
		DatePicker(c, &month, DatePickerProps{Picker: PickMonth, Width: 200})
		DatePicker(c, &week, DatePickerProps{Picker: PickWeek, Width: 200})
	}, 600, 700)
	tester.ClickAt(80, 36)
	tester.Frame()
	if err := tester.Click("3月"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if month.Month() != time.March || month.Day() != 1 {
		t.Fatalf("month = %v", month)
	}
	tester.Advance(300 * time.Millisecond)
	tester.ClickAt(80, 84)
	tester.Frame()
	now := time.Now()
	day := time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, time.Local)
	if err := tester.Click(day.Format("2006-01-02")); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if week.Weekday() != time.Monday || !strings.HasSuffix(pickerText(week, PickWeek, ""), "周") || day.Sub(week) >= 7*24*time.Hour || day.Before(week) {
		t.Fatalf("week = %v for %v", week, day)
	}
}

func TestColorPickerSquareAndAlpha(t *testing.T) {
	col := ui.Hex("#1677ff")
	tester := mount(func(c *ui.Context) {
		ColorPicker(c, &col, ColorPickerProps{open: true})
	}, 500, 500)
	tester.Frame()
	r, ok := tester.Find("饱和度与亮度")
	if !ok {
		t.Fatalf("no square: %v", tester.Texts())
	}
	tester.ClickAt(r.X+1, r.Y+1)
	tester.Frame()
	if col.R < 250 || col.G < 250 || col.B < 250 {
		t.Fatalf("top left of the square = %s", colorHex(col))
	}
	tester.ClickAt(r.X+r.W/2, r.Y+r.H-1)
	tester.Frame()
	if col.R > 3 || col.G > 3 || col.B > 3 {
		t.Fatalf("bottom of the square = %s", colorHex(col))
	}
	a, _ := tester.Find("透明度")
	tester.ClickAt(a.X+a.W/2, a.Y+a.H/2)
	tester.Frame()
	if col.A < 118 || col.A > 138 {
		t.Fatalf("alpha mid = %d", col.A)
	}
}
