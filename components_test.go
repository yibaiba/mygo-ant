package kit

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

var tabItems = []Option{
	{Value: "base", Label: "基础设置"},
	{Value: "db", Label: "数据库设置"},
	{Value: "off", Label: "已停用", Disabled: true},
	{Value: "fonts", Label: "字体安装"},
}

func TestTabsSelect(t *testing.T) {
	for _, typ := range []TabsType{TabsLine, TabsCard} {
		tab, changes := "base", 0
		tester := ui.NewTester(func(c *ui.Context) {
			Page(c, Work, func() {
				if Tabs(c, &tab, tabItems, TabsProps{Type: typ}) {
					changes++
				}
				Muted(c, "面板 "+tab)
			})
		}, 600, 200)
		tester.SetLocale("zh-CN")
		tester.Frame()
		if err := tester.Click("数据库设置"); err != nil {
			t.Fatal(err)
		}
		tester.Frame()
		if tab != "db" || changes != 1 || !tester.HasText("面板 db") {
			t.Fatalf("type %d: tab = %q after %d changes", typ, tab, changes)
		}
		tester.Key(0, ui.KeyRight)
		tester.Frame()
		tester.Frame()
		if tab != "fonts" {
			t.Fatalf("type %d: the arrow landed on %q, want the disabled tab skipped", typ, tab)
		}
		tester.Key(0, ui.KeyLeft)
		tester.Frame()
		tester.Frame()
		if tab != "db" {
			t.Fatalf("type %d: back from fonts landed on %q", typ, tab)
		}
		_ = tester.Click("已停用")
		tester.Frame()
		if tab != "db" {
			t.Fatalf("type %d: a disabled tab was chosen", typ)
		}
	}
}

func TestUploadDropsFiltersAndRemoves(t *testing.T) {
	files := []string{"/fonts/a.ttf"}
	picks := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if Upload(c, &files, UploadProps{Accept: []string{"ttf", "otf"}, Multiple: true}) {
				picks++
			}
		})
	}, 500, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	r, ok := tester.Find("点击或拖拽文件到此区域")
	if !ok {
		t.Fatal("no drop zone")
	}
	tester.DropFiles(r.X+r.W/2, r.Y+r.H/2, "/tmp/b.OTF", "/tmp/readme.txt", "/fonts/a.ttf")
	tester.Frame()
	tester.Frame()
	if len(files) != 2 || files[1] != "/tmp/b.OTF" {
		t.Fatalf("files after drop = %v", files)
	}
	if err := tester.Click("移除 a.ttf"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if len(files) != 1 || files[0] != "/tmp/b.OTF" {
		t.Fatalf("files after remove = %v", files)
	}
	if err := tester.Click("点击或拖拽文件到此区域"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if picks != 1 {
		t.Fatalf("picks = %d", picks)
	}
}

func TestMergeFilesSingle(t *testing.T) {
	got := mergeFiles([]string{"a"}, []string{"b", "c"}, false)
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("single upload kept %v", got)
	}
}

func TestPopconfirm(t *testing.T) {
	confirmed := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if Popconfirm(c, "确定移除这个字体？", func() bool {
				return Button(c, "移除", ButtonProps{Danger: true})
			}, PopconfirmProps{OK: "移除字体", Danger: true}) {
				confirmed++
			}
		})
	}, 500, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if tester.HasText("确定移除这个字体？") {
		t.Fatal("open before the click")
	}
	if err := tester.Click("移除"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("取消"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if confirmed != 0 || tester.HasText("确定移除这个字体？") {
		t.Fatal("cancel did not close")
	}
	_ = tester.Click("移除")
	tester.Frame()
	if err := tester.Click("移除字体"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if confirmed != 1 {
		t.Fatalf("confirmed = %d", confirmed)
	}
}

func TestDescRowsFillLines(t *testing.T) {
	items := []DescItem{{Label: "a"}, {Label: "b", Span: 2}, {Label: "c"}, {Label: "d"}}
	rows := descRows(items, 2)
	if len(rows) != 3 {
		t.Fatalf("rows = %d", len(rows))
	}
	if rows[0][0].Span != 2 || rows[1][0].Span != 2 || len(rows[2]) != 2 {
		t.Fatalf("spans = %+v", rows)
	}
}

func TestDatePickerChooses(t *testing.T) {
	var day time.Time
	changed := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if DatePicker(c, &day, DatePickerProps{Width: 200}) {
				changed++
			}
		})
	}, 500, 500)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("请选择日期"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("15"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if changed != 1 || day.Day() != 15 {
		t.Fatalf("day = %v after %d changes", day, changed)
	}
	if !tester.HasText(day.Format("2006-01-02")) {
		t.Fatal("field does not show the day")
	}
}

func TestMessagesShowToasts(t *testing.T) {
	sent := false
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {})
		if !sent {
			c.AddToast(ui.Toast{Title: "初始化已提交", Type: "success"})
			sent = true
		}
		Messages(c)
	}, 500, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	if !tester.HasText("初始化已提交") {
		t.Fatal("message not shown")
	}
}

func TestMessagesTop(t *testing.T) {
	cases := []struct {
		top   float32
		below []ui.Rect
		want  float32
	}{
		{0, nil, 8},
		{48, nil, 48},
		{48, []ui.Rect{{X: 0, Y: 0, W: 720, H: 40}}, 48},
		{48, []ui.Rect{{X: 0, Y: 40, W: 720, H: 90}, {X: 12, Y: 56, W: 100, H: 30}}, 138},
		{48, []ui.Rect{{X: 12, Y: -80, W: 100, H: 30}}, 48},
		{48, []ui.Rect{{}}, 48},
	}
	for i, tc := range cases {
		if got := messagesTop(tc.top, tc.below); got != tc.want {
			t.Errorf("case %d: top = %v, want %v", i, got, tc.want)
		}
	}
}

func TestMessagesStayBelowTitle(t *testing.T) {
	sent := false
	var title ui.Element
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			title = ui.Column(c).FillWidth().Padding(60, 0, 0).Children(func() {
				ui.Text(c, "连接服务器").FontSize(20)
			})
		})
		if !sent {
			c.AddToast(ui.Toast{Title: "网络连接失败", Type: "error"})
			sent = true
		}
		Messages(c, MessagesProps{Top: 48, Below: []ui.Element{title}})
	}, 720, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	head, _ := tester.Find("连接服务器")
	msg, ok := tester.Find("网络连接失败")
	if !ok {
		t.Fatal("message not shown")
	}
	if msg.Y < head.Y+head.H {
		t.Errorf("message at y=%v covers the title ending at %v", msg.Y, head.Y+head.H)
	}
}
