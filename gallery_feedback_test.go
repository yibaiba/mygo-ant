package kit

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// galleryFeedback is what the feedback page edits.
type galleryFeedback struct {
	tok          *Tokens
	panels       []string
	ghost        []string
	drawer       bool
	drawerBottom bool
}

func newGalleryFeedback() *galleryFeedback {
	return &galleryFeedback{tok: Work, panels: []string{"db", "net"}, ghost: []string{"why"}, drawer: true}
}

var galleryMenu = []MenuItem{
	{Key: "refresh", Label: "刷新日志"},
	{Key: "copy", Label: "复制日志路径"},
	{Key: "export", Label: "导出诊断包", Disabled: true},
	{Divider: true},
	{Key: "clear", Label: "清空日志", Danger: true},
}

// feedback is the feedback page: Result, Spin, Skeleton, Collapse and
// Dropdown.
func (g *galleryFeedback) feedback(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 反馈与折叠").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(8, 0).Children(func() {
			ui.Column(c).Grow(1).Basis(0).Children(func() {
				SectionTitle(c, "结果 Result")
				Sheet(c, func() {
					ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Stretch).Children(func() {
						ui.Column(c).Grow(1).Basis(0).Border(1, tok.BorderSecondary).Radius(8).Children(func() {
							Result(c, "服务器安装完成", ResultProps{Status: Success, SubTitle: "路径、数据库和网络均已设置，请继续设置安全参数。",
								Extra: func() {
									Button(c, "跳转设置", ButtonProps{Type: Primary})
									Button(c, "返回首页")
								}})
						})
						ui.Column(c).Grow(1).Basis(0).Border(1, tok.BorderSecondary).Radius(8).Children(func() {
							Result(c, "升级失败", ResultProps{Status: Error, SubTitle: "升级包校验失败：签名与当前服务器不匹配",
								Extra: func() {
									Button(c, "重试", ButtonProps{Type: Primary})
									Button(c, "重新选择")
								}})
						})
					})
					ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Stretch).Children(func() {
						ui.Column(c).Grow(1).Basis(0).Border(1, tok.BorderSecondary).Radius(8).Children(func() {
							Result(c, "卸载任务已提交", ResultProps{Status: Info, SubTitle: "完成后服务器会自动重启"})
						})
						ui.Column(c).Grow(1).Basis(0).Border(1, tok.BorderSecondary).Radius(8).Children(func() {
							Result(c, "证书即将过期", ResultProps{Status: Warning, Body: func() {
								Muted(c, "TLS 证书将在 2026-12-01 失效")
								Muted(c, "请在安全参数设置中导入新证书")
							}})
						})
					})
				})
				SectionTitle(c, "加载中 Spin")
				Sheet(c, func() {
					row(c, "尺寸", func() {
						Spin(c, SpinProps{Size: Small})
						Spin(c)
						Spin(c, SpinProps{Size: Large})
						Spin(c, SpinProps{Tip: "加载中"})
					})
					row(c, "包住内容", func() {
						ui.Column(c).Grow(1).Children(func() {
							SpinOver(c, true, func() {
								Descriptions(c, []DescItem{
									{Label: "数据库类型", Value: "mysql"}, {Label: "主机", Value: "127.0.0.1"},
									{Label: "端口", Value: "3306"}, {Label: "库名", Value: "example"},
								}, DescriptionsProps{Bordered: true, Size: Small, Columns: 2})
							}, SpinProps{Tip: "正在连接数据库"})
						})
					})
				})
			})
			ui.Column(c).Grow(1).Basis(0).Children(func() {
				SectionTitle(c, "骨架屏 Skeleton")
				Sheet(c, func() {
					Skeleton(c, SkeletonProps{Avatar: true, Active: true})
					Divider(c)
					Muted(c, "按服务卡片拼出的骨架（首页加载服务时）")
					ui.Row(c).FillWidth().Gap(16).Children(func() {
						for range 2 {
							ui.Column(c).Grow(1).Basis(0).Padding(18).Gap(10).Border(1, tok.BorderSecondary).Radius(12).Children(func() {
								ui.Row(c).FillWidth().Children(func() {
									SkeletonBlock(c, 64, 64, SkeletonBlockProps{Radius: 12, Active: true})
									ui.Spacer(c)
									SkeletonBlock(c, 44, 22, SkeletonBlockProps{Round: true, Active: true})
								})
								SkeletonBlock(c, 120, 18, SkeletonBlockProps{Active: true})
								SkeletonBlock(c, 180, 14, SkeletonBlockProps{Active: true})
							})
						}
					})
				})
				SectionTitle(c, "折叠面板 Collapse")
				Sheet(c, func() {
					Collapse(c, &g.panels, []CollapseItem{
						{Key: "db", Label: "数据库连接", Body: func() { Muted(c, "mysql · 127.0.0.1:3306 · example") }},
						{Key: "net", Label: "网络", Extra: func() { Tag(c, "静态 IP", TagProps{Tone: TagBlue}) }, Body: func() {
							Muted(c, "172.30.0.2/24 · 网关 172.30.0.1")
						}},
						{Key: "cert", Label: "证书", Body: func() { Muted(c, "未导入") }},
						{Key: "off", Label: "高级（远程连接时不可用）", Disabled: true},
					})
					Muted(c, "幽灵 + 手风琴 + 小尺寸")
					Collapse(c, &g.ghost, []CollapseItem{
						{Key: "why", Label: "为什么检测不到服务器？", Body: func() { Muted(c, "确认两台机器在同一网段，并允许 36678 端口。") }},
						{Key: "pass", Label: "忘记连接密码怎么办？", Body: func() { Muted(c, "在服务器本机的设置页重新生成连接卡。") }},
					}, CollapseProps{Ghost: true, Accordion: true, Size: Small})
				})
				SectionTitle(c, "下拉菜单 Dropdown")
				Sheet(c, func() {
					row(c, "展开", func() {
						DropdownButton(c, "更多操作", galleryMenu, DropdownProps{open: true})
					})
					ui.Box(c).Height(170)
				})
			})
		})
	})
}

// drawerPage is a log viewer page under an open drawer.
func (g *galleryFeedback) drawerPage(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Column(c).Fill().Padding(24).Gap(12).Children(func() {
			SectionTitle(c, "服务管理")
			Skeleton(c, SkeletonProps{Rows: 4})
		})
		Drawer(c, &g.drawer, "Demo Service · 日志", func() {
			for _, line := range []string{
				"2026-11-06 09:12:03 INFO  启动Demo Service，端口 96000",
				"2026-11-06 09:12:04 INFO  连接数据库 mysql://127.0.0.1:3306/example",
				"2026-11-06 09:12:05 WARN  TLS 证书将在 25 天后过期",
				"2026-11-06 09:12:06 INFO  服务已就绪",
			} {
				ui.Text(c, line).FontSize(12).TextColor(tok.Text).Selectable()
			}
		}, DrawerProps{Width: 520, Placement: pick(g.drawerBottom, DrawerBottom, DrawerRight), Extra: func() {
			Dropdown(c, galleryMenu, func() bool { return Button(c, "更多", ButtonProps{Size: Small}) }, DropdownProps{End: true})
		}, Footer: func() {
			Button(c, "关闭")
			Button(c, "刷新", ButtonProps{Type: Primary})
		}})
	})
}

func TestGalleryFeedbackShots(t *testing.T) {
	g := newGalleryFeedback()
	page := ui.NewTester(g.feedback, 1440, 1320)
	page.SetLocale("zh-CN")
	page.Frame()
	page.Frame()
	for _, want := range []string{"服务器安装完成", "正在连接数据库", "数据库连接", "清空日志", "为什么检测不到服务器？"} {
		if !page.HasText(want) {
			t.Fatalf("feedback gallery missing %q", want)
		}
	}
	shot(t, page, "mygo-kit-gallery-feedback.png")

	drawer := ui.NewTester(g.drawerPage, 1280, 720)
	drawer.SetLocale("zh-CN")
	drawer.Frame()
	if !drawer.HasText("服务已就绪") {
		t.Fatal("drawer body missing")
	}
	shot(t, drawer, "mygo-kit-gallery-drawer.png")

	dark := newGalleryFeedback()
	dark.tok = WorkDark
	darkPage := ui.NewTester(dark.feedback, 1440, 1320)
	darkPage.SetLocale("zh-CN")
	darkPage.Frame()
	darkPage.Frame()
	shot(t, darkPage, "mygo-kit-gallery-dark-feedback.png")
}
