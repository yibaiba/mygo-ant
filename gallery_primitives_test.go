package kit

import (
	"strconv"
	"testing"

	"github.com/egoist/mygo/ui"
)

// galleryPrimitives is what the shared-primitives page edits.
type galleryPrimitives struct {
	tok                                 *Tokens
	plain, filled, bare, url, pass, cnt string
	price                               string
	drawer                              bool
}

func newGalleryPrimitives() *galleryPrimitives {
	return &galleryPrimitives{tok: Work, plain: "172.30.0.2", filled: "填充样式", url: "example", pass: "123456", cnt: "Demo Meeting", price: "1280", drawer: true}
}

// primitives is the Batch 0 page: placements, triggers, the input shell,
// the painted shapes and a virtual list.
func (g *galleryPrimitives) primitives(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 基础能力").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(8, 0).Children(func() {
			ui.Column(c).Grow(1).Basis(0).Children(func() {
				SectionTitle(c, "弹出层定位 Popover · 四向 + 箭头")
				Sheet(c, func() {
					ui.Column(c).FillWidth().Height(300).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
						ui.Row(c).Gap(110).AlignItems(ui.Center).Children(func() {
							Popover(c, func() { Muted(c, "左侧弹出") }, func() bool { return Button(c, "左") }, PopoverProps{PopupProps: PopupProps{Placement: PlaceLeft, open: true}})
							ui.Column(c).Gap(140).AlignItems(ui.Center).Children(func() {
								Popover(c, func() { Muted(c, "上方居中") }, func() bool { return Button(c, "上") }, PopoverProps{Title: "标题", PopupProps: PopupProps{open: true}})
								Popover(c, func() { Muted(c, "下方左对齐") }, func() bool { return Button(c, "下") }, PopoverProps{PopupProps: PopupProps{Placement: PlaceBottom, Align: AlignFirst, open: true}})
							})
							Popover(c, func() { Muted(c, "右侧弹出") }, func() bool { return Button(c, "右") }, PopoverProps{PopupProps: PopupProps{Placement: PlaceRight, open: true}})
						})
					})
				})
				SectionTitle(c, "输入框外壳 Input")
				Sheet(c, func() {
					row(c, "三种变体", func() {
						Input(c, &g.plain, InputProps{Width: 140, AllowClear: true})
						Input(c, &g.filled, InputProps{Width: 120, Variant: Filled})
						Input(c, &g.bare, InputProps{Width: 100, Variant: Borderless, Placeholder: "无边框"})
					})
					row(c, "前后缀", func() {
						Input(c, &g.price, InputProps{Width: 200, Prefix: "¥", Suffix: "RMB"})
						Input(c, &g.bare, InputProps{Width: 200, PrefixIcon: iconSearch, Placeholder: "搜索服务"})
					})
					row(c, "前后置标签", func() {
						Input(c, &g.url, InputProps{Width: 360, AddonBefore: "https://", AddonAfter: ".local"})
					})
					row(c, "密码 · 计数", func() {
						Input(c, &g.pass, InputProps{Width: 200, Password: true})
						Input(c, &g.cnt, InputProps{Width: 220, ShowCount: true, MaxLength: 20, AllowClear: true})
					})
					row(c, "状态", func() {
						Input(c, &g.plain, InputProps{Width: 120, Status: StatusError, Variant: Filled})
						Input(c, &g.plain, InputProps{Width: 130, Status: StatusWarning, Prefix: "IP"})
						Input(c, &g.plain, InputProps{Width: 130, Disabled: true, AddonAfter: "端口"})
					})
				})
			})
			ui.Column(c).Grow(1).Basis(0).Children(func() {
				SectionTitle(c, "绘制图形 · 圆弧 / 星形 / 缎带 / 平铺文字")
				Sheet(c, func() {
					ui.Row(c).FillWidth().Gap(24).AlignItems(ui.Center).Children(func() {
						ui.Box(c).Size(96, 96).Draw(func(p *ui.Painter, r ui.Rect) {
							cx, cy := r.X+r.W/2, r.Y+r.H/2
							p.StrokePath(arcPath(cx, cy, 42, 0, 360), 6, tok.FillSecondary)
							p.StrokePath(arcPath(cx, cy, 42, 0, 270).Cap(ui.RoundCap), 6, tok.Primary)
							p.Text(cx-14, cy-10, "75%", 14, tok.Text)
						})
						ui.Box(c).Size(96, 96).Draw(func(p *ui.Painter, r ui.Rect) {
							cx, cy := r.X+r.W/2, r.Y+r.H/2
							p.StrokePath(arcPath(cx, cy, 42, -135, 270).Cap(ui.RoundCap), 6, tok.FillSecondary)
							p.StrokePath(arcPath(cx, cy, 42, -135, 270*0.6).Cap(ui.RoundCap), 6, tok.Success)
						})
						ui.Row(c).Gap(6).Children(func() {
							for i := range 5 {
								ui.Box(c.Key(i)).Size(20, 20).Draw(func(p *ui.Painter, r ui.Rect) {
									star := starPath(r.X+r.W/2, r.Y+r.H/2+1, 10, 4)
									p.FillPath(star, pick(i < 3, ui.Hex("#fadb14"), tok.FillSecondary))
								})
							}
						})
					})
					ui.Row(c).FillWidth().Gap(24).Children(func() {
						ui.Column(c).Width(220).Height(110).Padding(16).Border(1, tok.BorderSecondary).Radius(8).
							Draw(func(p *ui.Painter, r ui.Rect) {
								drawRibbon(p, ui.Rect{X: r.X + r.W - 64 + ribbonFold, Y: r.Y + 10, W: 64, H: 22}, tok.Primary, false)
								p.Text(r.X+r.W-52+ribbonFold, r.Y+13, "推荐", 12, tok.TextOnPrimary)
							}).Children(func() {
							ui.Text(c, "缎带 Ribbon").FontWeight(600).TextColor(tok.Text)
							Muted(c, "角标折边用路径绘制")
						})
						ui.Column(c).Grow(1).Height(110).Border(1, tok.BorderSecondary).Radius(8).
							Draw(func(p *ui.Painter, r ui.Rect) {
								drawTiledText(p, r, []string{"Confidential", "内部资料"}, 13, tok.Text.Alpha(0.15), -22, 60, 40)
							})
					})
				})
				SectionTitle(c, "虚拟列表 VirtualList · 10000 行")
				Sheet(c, func() {
					VirtualList(c, 10000, func(c *ui.Context, i int) {
						ui.Row(c).FillWidth().Height(36).Padding(0, 12).AlignItems(ui.Center).Children(func() {
							ui.Text(c, "会议终端 "+strconv.Itoa(i+1)).FontSize(tok.FontSize).TextColor(tok.Text).Grow(1)
							Muted(c, "192.168.1."+strconv.Itoa(i%254+1))
						})
					}, VirtualListProps{Height: 220, Split: true})
				})
				SectionTitle(c, "下拉菜单 · 箭头 + 右对齐")
				Sheet(c, func() {
					ui.Row(c).FillWidth().Justify(ui.End).Children(func() {
						Dropdown(c, galleryMenu, func() bool { return Button(c, "更多操作") }, DropdownProps{End: true, Arrow: true, open: true})
					})
					ui.Box(c).Height(200)
				})
			})
		})
	})
}

// leftDrawer is a page under a drawer opened from the left edge.
func (g *galleryPrimitives) leftDrawer(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Column(c).Fill().Padding(24).Gap(12).Children(func() {
			SectionTitle(c, "服务管理")
			Skeleton(c, SkeletonProps{Rows: 4})
		})
		Drawer(c, &g.drawer, "左侧抽屉", func() {
			Muted(c, "从左侧滑入，遮罩随之淡入")
		}, DrawerProps{Placement: DrawerLeft, Footer: func() { Button(c, "确定", ButtonProps{Type: Primary}) }})
	})
}

func TestGalleryPrimitivesShots(t *testing.T) {
	g := newGalleryPrimitives()
	page := ui.NewTester(g.primitives, 1440, 1100)
	page.SetLocale("zh-CN")
	page.Frame()
	page.Frame()
	for _, want := range []string{"左侧弹出", "上方居中", "下方左对齐", "右侧弹出", "https://", "会议终端 1", "清空日志"} {
		if !page.HasText(want) {
			t.Fatalf("primitives gallery missing %q", want)
		}
	}
	shot(t, page, "mygo-kit-gallery-primitives.png")

	drawer := ui.NewTester(g.leftDrawer, 1000, 560)
	drawer.SetLocale("zh-CN")
	drawer.Frame()
	shot(t, drawer, "mygo-kit-gallery-drawer-left.png")
}
