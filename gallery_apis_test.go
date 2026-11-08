package kit

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

type galleryAPIs struct {
	userName, userMail string
	files              []UploadFile
	path               []string
	tree               string
	nodes, cascade     []TreeNode
	color              ui.Color
	from, to           time.Time
	t0, t1             time.Time
	form               Form
}

func newGalleryAPIs() *galleryAPIs {
	day := time.Date(2026, 11, 6, 9, 30, 0, 0, time.Local)
	g := &galleryAPIs{
		userName: "张三", userMail: "zhang@corp.cn",
		files: []UploadFile{{Path: "/data/议程.pdf", Status: UploadUploading, Percent: 55}},
		path:  []string{"cn"},
		tree:  "gd",
		color: ui.Hex("#1677ff"),
		from:  day, to: day.Add(8 * time.Hour),
		t0: day, t1: day.Add(3 * time.Hour),
		cascade: []TreeNode{
			{Key: "cn", Title: "中国", Children: []TreeNode{
				{Key: "gd", Title: "广东", IsLeaf: true},
				{Key: "zj", Title: "浙江"},
			}},
		},
	}
	g.nodes = []TreeNode{
		{Key: "cn", Title: "中国", Children: []TreeNode{
			{Key: "gd", Title: "广东", IsLeaf: true},
			{Key: "zj", Title: "浙江", IsLeaf: true},
		}},
	}
	return g
}

func (g *galleryAPIs) page(c *ui.Context) {
	f := &g.form
	Page(c, Work, func() {
		ui.Row(c).FillWidth().Height(40).Background(Work.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · Ant API 补齐").FontSize(16).FontWeight(600).TextColor(Work.TextOnPrimary)
		})
		ui.Column(c).Fill().Padding(12).Gap(12).Children(func() {
			ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(12).Children(func() {
					SectionTitle(c, "Form 嵌套名 NamePath / user.name")
					Sheet(c, func() {
						f.Begin()
						FormView(c, func() {
							nameErr := f.Check([]string{"user", "name"}, g.userName, Required("请输入姓名"))
							FormItem(c, "用户姓名", func() {
								Input(c, &g.userName, InputProps{Status: ErrorStatus(nameErr)})
							}, FormItemProps{Name: []string{"user", "name"}, Required: true, Extra: `Name: []string{"user","name"}`, Error: nameErr})
							mailErr := f.Check("user.email", g.userMail)
							FormItem(c, "邮箱", func() {
								Input(c, &g.userMail, InputProps{Status: ErrorStatus(mailErr)})
							}, FormItemProps{Name: "user.email"})
						}, FormProps{FormConfig: FormConfig{Layout: FormHorizontal}})
					})
					SectionTitle(c, "Upload Action · 进度由 kit 写回")
					Sheet(c, func() {
						UploadList(c, &g.files, UploadListProps{Action: "https://example.invalid/upload", Multiple: true})
					})
				})
				ui.Column(c).Width(520).Shrink(0).Gap(12).Children(func() {
					SectionTitle(c, "表头分组 Column.Children")
					Sheet(c, func() {
						Table(c, []Column{
							{Title: "会议", Children: []Column{{Title: "名称", Width: 160}, {Title: "会议室", Width: 110}}},
							{Title: "状态", Width: 90},
						}, 3, func(r, col int) Cell {
							row := [][]string{{"季度经营会", "一号", "进行中"}, {"预算评审", "二号", "待开始"}, {"安全培训", "报告厅", "已结束"}}[r]
							return Cell{Text: row[col]}
						})
					})
					SectionTitle(c, "RangePicker ShowTime · TimeRangePicker")
					Sheet(c, func() {
						RangePicker(c, &g.from, &g.to, RangePickerProps{
							Width: 360, ShowTime: true, AllowClear: true,
							DisabledDate:  func(d time.Time) bool { return d.Before(dayOf(g.from)) },
							DisabledHours: func() []int { return []int{0, 1, 2, 3, 4, 5} },
						})
						TimeRangePicker(c, &g.t0, &g.t1, TimePickerProps{
							Width: 280, HideSeconds: true, AllowClear: true,
							DisabledHours: func() []int { return []int{0, 1, 2, 3, 4, 5} },
						})
					})
				})
			})
			ui.Row(c).FillWidth().Height(500).Gap(24).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Width(280).Shrink(0).Children(func() {
					SectionTitle(c, "Cascader LoadData")
					Sheet(c, func() {
						Cascader(c, &g.path, g.cascade, CascaderProps{
							SelectProps: SelectProps{Width: 240, open: true},
							LoadData: func(selected []TreeNode) {
								if len(selected) == 0 {
									return
								}
								for i := range g.cascade {
									fillCascade(&g.cascade[i], selected[len(selected)-1].Key)
								}
							},
						})
					})
				})
				ui.Column(c).Width(320).Shrink(0).Children(func() {
					SectionTitle(c, "ColorPicker HSB")
					Sheet(c, func() {
						ColorPicker(c, &g.color, ColorPickerProps{ShowText: true, Format: FormatHSB, open: true})
					})
				})
				ui.Column(c).Grow(1).Children(func() {
					SectionTitle(c, "TimeRangePicker 面板")
					Sheet(c, func() {
						TimeRangePicker(c, &g.t0, &g.t1, TimePickerProps{
							Width: 260, open: true, HideSeconds: true, AllowClear: true,
							DisabledHours: func() []int { return []int{0, 1, 2, 3, 4, 5} },
						})
					})
				})
			})
		})
	})
}

func fillCascade(n *TreeNode, key string) {
	if n.Key == key && len(n.Children) == 0 && !n.IsLeaf {
		n.Children = []TreeNode{{Key: n.Key + "-a", Title: n.Title + " · 市", IsLeaf: true}}
		return
	}
	for i := range n.Children {
		fillCascade(&n.Children[i], key)
	}
}

func TestGalleryAPIsShot(t *testing.T) {
	g := newGalleryAPIs()
	page := ui.NewTester(g.page, 1440, 980)
	page.SetLocale("zh-CN")
	page.Frame()
	page.Frame()
	page.Frame()
	for _, want := range []string{"用户姓名", "议程.pdf", "会议", "会议室", "HSB", "开始时间", "浙江"} {
		if !page.HasText(want) {
			t.Fatalf("apis gallery missing %q in %v", want, page.Texts())
		}
	}
	shot(t, page, "mygo-kit-apis.png")
}
