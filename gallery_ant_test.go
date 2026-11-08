package kit

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

type galleryAnt struct {
	on, off  bool
	sel      int
	page     int
	pageSize int
	tab      string
}

func newGalleryAnt() *galleryAnt {
	return &galleryAnt{on: true, sel: 1, page: 2, pageSize: 10, tab: "base"}
}

func switchSample(c *ui.Context, caption string, fn func()) {
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, caption).FontSize(12).TextColor(Of(c).TextMuted).Shrink(0)
		fn()
	})
}

func (g *galleryAnt) pageView(c *ui.Context) {
	Page(c, Work, func() {
		ui.Column(c).Fill().Padding(20, 24).Gap(16).Children(func() {
			ui.Row(c).FillWidth().Height(40).Background(Work.Brand).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "kit · Ant + MyGo 原生").FontSize(16).FontWeight(600).TextColor(Work.TextOnPrimary).Margin(0, 0, 0, 16)
				ui.Spacer(c)
				WindowButtons(c, WindowButtonsProps{CloseLabel: "隐藏", Light: true, Width: 48})
			})
			SectionTitle(c, "Switch / Empty / SplitButton / Pagination")
			Sheet(c, func() {
				row(c, "开关", func() {
					switchSample(c, "带文字", func() {
						Switch(c, &g.on, ChoiceProps{CheckedChildren: "开", UnCheckedChildren: "关"})
						Switch(c, &g.off, ChoiceProps{CheckedChildren: "开", UnCheckedChildren: "关"})
					})
					switchSample(c, "默认", func() {
						Switch(c, &g.on)
					})
				})
				row(c, "下拉", func() {
					SplitButton(c, "选择 zip", []MenuItem{
						{Key: "dir", Label: "选择目录（旧版）"},
						{Divider: true},
						{Key: "clear", Label: "清除"},
					})
					DropdownButton(c, "更多操作", []MenuItem{
						{Key: "copy", Label: "复制", ShortcutKey: ui.KeyC, ShortcutMods: ui.Ctrl},
						{Key: "native", Label: "原生表", Checked: true},
					})
				})
				Empty(c, "尚未扫描", EmptyProps{
					Description: "点扫描开始查找同一网段的会场",
					Extra:       func() { Button(c, "扫描", ButtonProps{Type: Primary}) },
				})
				Pagination(c, &g.page, 95, g.pageSize, PaginationProps{
					ShowTotal: true, ShowQuickJumper: true, ShowSizeChanger: true, PageSize: &g.pageSize,
				})
			})
			SectionTitle(c, "Descriptions · Copyable")
			Sheet(c, func() {
				Descriptions(c, []DescItem{
					{Label: "版本", Value: "1.0.54", Copyable: true},
					{Label: "IP", Value: "172.30.0.2", Copyable: true},
					{Label: "MAC", Value: "aa:bb:cc:dd:ee:ff", Copyable: true},
				}, DescriptionsProps{Columns: 3})
			})
			SectionTitle(c, "Native Table")
			Sheet(c, func() {
				Table(c, []Column{
					{Title: "IP", Width: 160, Sorter: ByText(func(r int) string { return hosts[r] })},
					{Title: "状态", Width: 80},
				}, len(hosts), hostCell, TableProps{Selected: &g.sel, Native: true, PageSize: 8, ShowTotal: true})
			})
		})
	})
}

func TestGalleryAntShots(t *testing.T) {
	g := newGalleryAnt()
	tester := ui.NewTester(g.pageView, 1280, 900)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	for _, want := range []string{"带文字", "默认", "开", "关", "选择 zip", "尚未扫描", "跳至", "条/页", "复制", "192.168.0.1", "隐藏"} {
		if !tester.HasText(want) {
			t.Fatalf("ant gallery missing %q in %v", want, tester.Texts())
		}
	}
	shot(t, tester, "mygo-kit-ant-native.png")
	shot(t, tester, "mygo-kit-switch.png")

	home := ui.NewTester(func(c *ui.Context) {
		Page(c, Home, func() {
			HomeBar(c, HomeBarProps{
				Title:    "Demo App",
				Subtitle: "高效 · 稳定 · 安全",
				Width:    1280,
				Nav: func() {
					TopNav(c, "连接")
					TopNav(c, "设置")
				},
				Extra: func() { WindowButtons(c, WindowButtonsProps{CloseLabel: "隐藏"}) },
			})
			ui.Column(c).Padding(24).Children(func() {
				Empty(c, "暂无服务数据", EmptyProps{Simple: true, Description: "正在连接本机服务..."})
			})
		})
	}, 1280, 240)
	home.SetLocale("zh-CN")
	home.Frame()
	shot(t, home, "mygo-kit-ant-home-empty.png")
}
