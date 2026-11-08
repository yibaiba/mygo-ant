package kit

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

type galleryDesktop struct {
	name     string
	vol      float64
	page     int
	sel      int
	expanded []string
	files    []string
}

func newGalleryDesktop() *galleryDesktop {
	return &galleryDesktop{
		name: "Example Inc.", vol: 40, page: 2, sel: 0,
		expanded: []string{"zj"},
	}
}

func (g *galleryDesktop) pageView(c *ui.Context) {
	Page(c, Work, func() {
		ui.Column(c).Fill().Padding(16, 20).Gap(12).Children(func() {
			ui.Row(c).FillWidth().Height(40).Background(Work.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "kit · 桌面交互").FontSize(16).FontWeight(600).TextColor(Work.TextOnPrimary)
			})
			SectionTitle(c, "光标 / 加载 / 输入 / Tooltip")
			Sheet(c, func() {
				row(c, "按钮", func() {
					Button(c, "保存", ButtonProps{Type: Primary})
					Button(c, "禁用", ButtonProps{Disabled: true})
					Button(c, "连接中", ButtonProps{Type: Primary, Loading: true})
					Tooltip(c, "Ant 100ms 进入 / 150ms 离开", func() { Button(c, "锚点") })
				})
				row(c, "输入", func() {
					Input(c, &g.name, InputProps{Prefix: "名称", Width: 280})
				})
			})
			SectionTitle(c, "滚轮 · Slider / Pagination")
			Sheet(c, func() {
				ui.Box(c).Width(280).Children(func() {
					Slider(c, &g.vol, SliderProps{Step: 10, Tooltip: TooltipAlways})
				})
				Pagination(c, &g.page, 95, 10, PaginationProps{ShowTotal: true})
			})
			SectionTitle(c, "双击 · Table / Tree")
			Sheet(c, func() {
				Table(c, []Column{{Title: "备注", Width: 120}, {Title: "状态", Width: 80}}, 1, func(row, col int) Cell {
					if col == 0 {
						return Cell{Text: "超长会议纪要需要提示才能看完整句"}
					}
					return Cell{Text: "在线"}
				}, TableProps{Selected: &g.sel, OnDoubleClick: func(int) {}})
				Tree(c, regions(), TreeProps{Expanded: &g.expanded})
			})
			SectionTitle(c, "拖放 / 分隔")
			Sheet(c, func() {
				ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Stretch).Children(func() {
					ui.Column(c).Grow(1).MinWidth(0).Children(func() {
						Upload(c, &g.files, UploadProps{Accept: []string{"otf", "ttf"}, Text: "点击或拖拽文件到此区域"})
					})
					ui.Column(c).Grow(1).MinWidth(0).Height(120).Children(func() {
						Splitter(c, []SplitPanel{
							{Render: func() { Muted(c, "左栏") }, Size: 120},
							{Render: func() { Muted(c, "右栏") }},
						})
					})
				})
			})
		})
	})
}

func TestGalleryDesktopShot(t *testing.T) {
	g := newGalleryDesktop()
	tester := ui.NewTester(g.pageView, 1100, 1100)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	for _, want := range []string{"锚点", "连接中", "共 95 条", "超长会议纪要需要提示才能看完整句", "杭州", "点击或拖拽文件到此区域", "拖动调整大小"} {
		if !tester.HasText(want) {
			t.Fatalf("desktop gallery missing %q in %v", want, tester.Texts())
		}
	}
	if r, ok := tester.Find("锚点"); ok {
		tester.Move(r.X+r.W/2, r.Y+r.H/2)
	}
	shot(t, tester, "mygo-kit-desktop.png")
}
