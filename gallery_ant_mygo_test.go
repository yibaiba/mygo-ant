package kit

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

type galleryAntMyGo struct {
	host, warn string
	form       Form
	blur       Form
}

func newGalleryAntMyGo() *galleryAntMyGo {
	g := &galleryAntMyGo{host: "", warn: "待确认"}
	g.blur.ValidateTrigger = ValidateBlur
	return g
}

func (g *galleryAntMyGo) page(c *ui.Context) {
	Page(c, Work, func() {
		ui.Row(c).FillWidth().Height(40).Background(Work.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · Ant ↔ MyGo").FontSize(16).FontWeight(600).TextColor(Work.TextOnPrimary)
		})
		ui.Column(c).Fill().Padding(12).Gap(12).Children(func() {
			ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(12).Children(func() {
					SectionTitle(c, "Button Ghost / Shape / AutoFocus")
					Sheet(c, func() {
						ui.Row(c).Gap(8).Wrap().GapY(8).AlignItems(ui.Center).Children(func() {
							Button(c, "幽灵主色", ButtonProps{Type: Primary, Ghost: true})
							Button(c, "幽灵默认", ButtonProps{Ghost: true})
							Button(c, "幽灵危险", ButtonProps{Type: Primary, Danger: true, Ghost: true})
							Button(c, "幽灵虚线", ButtonProps{Type: Dashed, Ghost: true})
							Button(c, "圆角胶囊", ButtonProps{Type: Primary, Shape: ShapeRound})
							Button(c, "", ButtonProps{Type: Primary, Shape: ShapeCircle, Icon: LookupIcon("SearchOutlined")})
							Button(c, "", ButtonProps{Shape: ShapeCircle, Icon: LookupIcon("PlusOutlined")})
						})
					})
					SectionTitle(c, "Input status 反馈图标 · onPressEnter")
					Sheet(c, func() {
						ui.Row(c).Gap(12).Wrap().GapY(8).Children(func() {
							Input(c, &g.host, InputProps{Width: 200, Status: StatusError, Placeholder: "必填出错"})
							Input(c, &g.warn, InputProps{Width: 200, Status: StatusWarning, Placeholder: "警告"})
							Input(c, &g.warn, InputProps{Width: 180, Status: StatusError, NoFeedback: true, Placeholder: "无图标"})
						})
					})
					SectionTitle(c, "Form validateTrigger=onBlur")
					Sheet(c, func() {
						g.blur.Begin()
						err := g.blur.Check("host", g.host, Required("失焦后才显示"))
						FormItem(c, "主机", func() {
							Input(c, &g.host, InputProps{Status: ErrorStatus(err), Placeholder: "点进去再点外面"})
						}, FormItemProps{Name: "host", Form: &g.blur, Required: true, Error: err, Extra: "失焦后才显示"})
					})
				})
				ui.Column(c).Width(520).Shrink(0).Gap(12).Children(func() {
					SectionTitle(c, "Modal 宽度 416 / 520 / 736")
					Sheet(c, func() {
						ui.Column(c).Gap(8).Children(func() {
							widthBar(c, "Small 416", DialogWidthSM)
							widthBar(c, "Default 520", DialogWidth)
							widthBar(c, "Large 736", DialogWidthLG)
						})
					})
					SectionTitle(c, "DarkAlgorithm · token.motion=false")
					ConfigProvider(c, ConfigProps{Dark: Bool(true)}, func() {
						t := Of(c)
						ui.Column(c).FillWidth().Padding(12).Gap(8).Background(t.Elevated).Border(1, t.Border).Radius(8).Children(func() {
							ui.Text(c, "Work 经 darkAlgorithm").FontSize(13).FontWeight(600).TextColor(t.TextHeading)
							ui.Row(c).Gap(8).Children(func() {
								Button(c, "主色", ButtonProps{Type: Primary})
								Button(c, "幽灵", ButtonProps{Type: Primary, Ghost: true})
								Button(c, "危险", ButtonProps{Danger: true})
							})
							in := "暗色输入"
							Input(c, &in, InputProps{Status: StatusError})
						})
					})
					ConfigProvider(c, ConfigProps{Motion: Bool(false)}, func() {
						Muted(c, "Motion: false 进出场跳过（ui.Animate 仍听 OS ReduceMotion）")
					})
				})
			})
			Muted(c, "zIndex / getPopupContainer：ui 只有一层 Overlay，后建的在上；弹层 AttachTo 触发器，不伪造 CSS 容器。")
		})
	})
}

func widthBar(c *ui.Context, label string, w float32) {
	t := Of(c)
	ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(8).Children(func() {
		ui.Text(c, label).FontSize(12).TextColor(t.TextMuted).Width(96).Shrink(0)
		ui.Box(c).Height(28).Width(w*0.5).Background(t.PrimaryBg).Border(1, t.PrimaryBorder).Radius(4).
			Children(func() {
				ui.Row(c).Fill().Padding(0, 8).AlignItems(ui.Center).Children(func() {
					ui.Text(c, pick(w == DialogWidthSM, "confirm", pick(w == DialogWidthLG, "Drawer Large", "Modal"))).
						FontSize(12).TextColor(t.Primary)
				})
			})
	})
}

func TestGalleryAntMyGoShot(t *testing.T) {
	g := newGalleryAntMyGo()
	page := ui.NewTester(g.page, 1440, 780)
	page.SetLocale("zh-CN")
	page.Frame()
	page.Frame()
	for _, want := range []string{"幽灵主色", "圆角胶囊", "失焦后才显示", "Small 416", "darkAlgorithm", "token.motion"} {
		if !page.HasText(want) {
			t.Fatalf("ant-mygo gallery missing %q in %v", want, page.Texts())
		}
	}
	shot(t, page, "mygo-kit-ant-mygo.png")
}
