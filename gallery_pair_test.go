package kit

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

type galleryPair struct {
	sel   int
	files []string
	path  string
	on    bool
	menu  string
}

func newGalleryPair() *galleryPair {
	return &galleryPair{sel: 1, files: []string{"会议纪要.pdf"}, path: "/opt/safe/package.zip", on: true, menu: "mail"}
}

func pairTag(c *ui.Context, tag string, tone ui.Color, fn func()) {
	t := Of(c)
	ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(6).Padding(8).Border(1, t.BorderSecondary).Radius(8).Clip().Children(func() {
		ui.Box(c).Padding(1, 8).Radius(4).Background(tone).AlignSelf(ui.Start).Children(func() {
			ui.Text(c, tag).FontSize(11).FontWeight(600).TextColor(t.TextOnPrimary)
		})
		fn()
	})
}

func pairRow(c *ui.Context, title string, ant, native func()) {
	t := Of(c)
	ui.Column(c).FillWidth().Gap(6).Children(func() {
		ui.Text(c, title).FontSize(13).FontWeight(600).TextColor(t.TextHeading)
		ui.Row(c).FillWidth().Gap(10).AlignItems(ui.Start).Children(func() {
			pairTag(c, "Ant", t.Primary, ant)
			pairTag(c, "Native", t.Warning, native)
		})
	})
}

func (g *galleryPair) page(c *ui.Context) {
	Page(c, Work, func() {
		ui.Column(c).Fill().Padding(12, 16).Gap(10).Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
				ui.Text(c, "kit · Ant 默认 / Native 可选").FontSize(16).FontWeight(600).TextColor(Work.TextHeading)
				ui.Spacer(c)
				Muted(c, "原生控件全部 opt-in，默认自绘 Ant v5")
			})
			ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(10).Children(func() {
					pairRow(c, "菜单 / 下拉 / 右键", func() {
						ui.Column(c).Gap(8).Height(150).AlignSelf(ui.Start).Children(func() {
							DropdownButton(c, "更多操作", []MenuItem{
								{Key: "copy", Label: "复制", Icon: LookupIcon("CopyOutlined")},
								{Key: "edit", Label: "编辑"},
								{Divider: true},
								{Key: "del", Label: "删除", Danger: true},
							}, DropdownProps{open: true})
						})
						Muted(c, "Dropdown / Menu 自绘")
					}, func() {
						ui.Column(c).Gap(8).AlignSelf(ui.Start).Children(func() {
							ui.Row(c).Shrink(0).AlignSelf(ui.Start).Children(func() {
								_ = DropdownButton(c, "系统菜单", []MenuItem{
									{Key: "copy", Label: "复制", ShortcutKey: ui.KeyC, ShortcutMods: ui.Ctrl},
									{Key: "native", Label: "原生项", Checked: true},
								}, DropdownProps{Native: true})
							})
							Muted(c, "Native / AttachMenu / AttachContextMenu")
						})
					})
					pairRow(c, "表格 Table", func() {
						Table(c, []Column{
							{Title: "IP", Width: 140, Sorter: ByText(func(r int) string { return hosts[r] })},
							{Title: "状态", Width: 72},
						}, len(hosts), hostCell, TableProps{Selected: &g.sel, PageSize: 3, ShowTotal: true})
						Muted(c, "默认自绘 Ant 表")
					}, func() {
						Table(c, []Column{
							{Title: "IP", Width: 140, Sorter: ByText(func(r int) string { return hosts[r] })},
							{Title: "状态", Width: 72},
						}, len(hosts), hostCell, TableProps{Selected: &g.sel, Native: true, PageSize: 3, ShowTotal: true})
						Muted(c, "Table Native 走 ui.Table")
					})
				})
				ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(10).Children(func() {
					pairRow(c, "通知 / 确认", func() {
						t := Of(c)
						ui.Column(c).Gap(8).Children(func() {
							Message(c, "success", "保存成功")
							Alert(c, "备份完成", AlertProps{Type: Info, Description: "已写入 /data/backup"})
							ui.Column(c).Gap(6).Padding(10).Border(1, t.BorderSecondary).Radius(8).Background(t.Elevated).Children(func() {
								ui.Row(c).Gap(8).AlignItems(ui.Start).Children(func() {
									statusIcon(c, Warning, 16, t.WarningIcon)
									ui.Text(c, "删除后不可恢复？").FontSize(t.FontSize)
								})
								ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
									Button(c, "取消", ButtonProps{Size: Small})
									Button(c, "确定", ButtonProps{Type: Primary, Danger: true, Size: Small})
								})
							})
							Muted(c, "Message / Notification / Modal.confirm 窗口内")
						})
					}, func() {
						ui.Column(c).Gap(8).AlignSelf(ui.Start).Children(func() {
							ui.Row(c).Gap(8).AlignSelf(ui.Start).Children(func() {
								Button(c, "NativeConfirm")
								Button(c, "NotifyOS")
							})
							Muted(c, "SetMessageBox / SetOSNotify 注入后才是系统框；应用确认框 仍走窗口内 Modal")
						})
					})
					pairRow(c, "文件", func() {
						Upload(c, &g.files, UploadProps{Text: "选择文件", Hint: "Ant Upload，不是系统资源管理器"})
					}, func() {
						PathPicker(c, &g.path, "浏览", PathPickerProps{Placeholder: "OS 文件框"})
						Muted(c, "PathPicker / SetFileDialog 保持系统对话框")
					})
				})
			})
			ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
					pairRow(c, "跟桌面 FollowDesktop", func() {
						ConfigProvider(c, ConfigProps{}, func() {
							ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
								Button(c, "Ant 令牌", ButtonProps{Type: Primary})
								Switch(c, &g.on)
							})
							Muted(c, "ConfigProvider 默认不叠 OS 高对比")
						})
					}, func() {
						ConfigProvider(c, ConfigProps{FollowDesktop: true}, func() {
							ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
								Button(c, "FollowDesktop", ButtonProps{Type: Primary})
								Switch(c, &g.on)
							})
							Muted(c, "可选：叠 HighContrast / TextScale")
						})
					})
				})
				ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
					pairRow(c, "窗钮 / HomeBar / GlassCard", func() {
						Layout(c, func() {
							LayoutHeader(c, func() {
								ui.Text(c, "Layout.Header").TextColor(ui.Hex("#FFFFFF"))
								ui.Spacer(c)
								Button(c, "操作", ButtonProps{Type: Primary, Size: Small})
							})
							LayoutContent(c, func() {
								Card(c, func() {
									Muted(c, "Ant 对应：Layout + Button + Card")
								}, CardProps{Title: "Card"})
							})
						})
					}, func() {
						ui.Column(c).Gap(8).FillWidth().Children(func() {
							ui.Row(c).FillWidth().Height(40).Background(Work.Brand).AlignItems(ui.Center).Children(func() {
								ui.Text(c, "应用壳").FontSize(13).TextColor(Work.TextOnPrimary).Margin(0, 0, 0, 8)
								ui.Spacer(c)
								WindowButtons(c, WindowButtonsProps{CloseLabel: "隐藏", Light: true, Width: 42, Height: 40})
							})
							Muted(c, "WindowButtons 是窗口控件，没有 Ant 窗钮；HomeBar 不是假浏览器壳")
						})
					})
				})
			})
			Menu(c, &g.menu, []MenuItem{
				{Key: "mail", Label: "导航一", Icon: LookupIcon("MailOutlined")},
				{Key: "app", Label: "导航二", Icon: LookupIcon("AppstoreOutlined")},
			}, MenuProps{Mode: MenuHorizontal})
		})
	})
}

func TestGalleryAntVsNative(t *testing.T) {
	g := newGalleryPair()
	tester := ui.NewTester(g.page, 1440, 1100)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	for _, want := range []string{"Ant", "Native", "更多操作", "系统菜单", "选择文件", "浏览", "NativeConfirm", "NotifyOS", "Layout.Header", "隐藏", "默认自绘 Ant 表"} {
		if !tester.HasText(want) {
			t.Fatalf("ant-vs-native missing %q in %v", want, tester.Texts())
		}
	}
	shot(t, tester, "mygo-kit-ant-vs-native.png")
}

func iconCell(c *ui.Context, name string, props ...IconProps) {
	p := first(props)
	if p.Size == 0 {
		p.Size = 22
	}
	if p.Name == "" {
		p.Name = name
	}
	ui.Column(c).Width(108).Gap(4).AlignItems(ui.Center).Children(func() {
		IconNamed(c, name, p)
		ui.Text(c, name).FontSize(10).TextColor(Of(c).TextTertiary).SingleLine()
	})
}

func TestGalleryIcons(t *testing.T) {
	var nOut, nFill, nTwin int
	for _, g := range antGlyphs {
		switch {
		case strings.HasSuffix(g.name, "TwoTone"):
			nTwin++
		case strings.HasSuffix(g.name, "Filled"):
			nFill++
		default:
			nOut++
		}
	}
	outlined := []string{
		"SearchOutlined", "HomeOutlined", "SettingOutlined", "UserOutlined", "BellOutlined",
		"MailOutlined", "CalendarOutlined", "FolderOutlined", "FileOutlined", "EditOutlined",
		"DeleteOutlined", "PlusOutlined", "CheckOutlined", "CloseOutlined", "StarOutlined",
		"HeartOutlined", "LockOutlined", "CloudOutlined", "DownloadOutlined", "UploadOutlined",
		"FilterOutlined", "ReloadOutlined", "QuestionCircleOutlined", "InfoCircleOutlined",
		"CheckCircleOutlined", "CloseCircleOutlined", "ExclamationCircleOutlined", "AppstoreOutlined",
		"BarsOutlined", "LinkOutlined", "GlobalOutlined", "DatabaseOutlined", "DesktopOutlined",
		"MobileOutlined", "SafetyOutlined", "ScanOutlined", "QrcodeOutlined", "TeamOutlined",
		"PoweroffOutlined", "SwapOutlined",
	}
	filled := []string{"HeartFilled", "StarFilled", "CheckCircleFilled", "CloseCircleFilled", "InfoCircleFilled", "WarningFilled", "HomeFilled", "BellFilled"}
	twotone := []string{"HeartTwoTone", "StarTwoTone", "CheckCircleTwoTone", "InfoCircleTwoTone", "WarningTwoTone", "ExclamationCircleTwoTone"}
	short := []string{"search", "home", "setting", "user", "check", "close"}
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Padding(16, 20).Gap(12).Children(func() {
				ui.Text(c, "kit · @ant-design/icons").FontSize(16).FontWeight(600).TextColor(Work.TextHeading)
				Muted(c, "848 个 Ant 图标（"+itoa(nOut)+" Outlined / "+itoa(nFill)+" Filled / "+itoa(nTwin)+" TwoTone）· MIT · kit.Icon{Name: \"SearchOutlined\"}")
				SectionTitle(c, "短名别名（原有 ~65）")
				ui.Row(c).Gap(8).Wrap().GapY(8).Children(func() {
					for _, name := range short {
						iconCell(c, name)
					}
				})
				SectionTitle(c, "Outlined")
				ui.Row(c).Gap(8).Wrap().GapY(10).Children(func() {
					for _, name := range outlined {
						iconCell(c, name)
					}
				})
				SectionTitle(c, "Filled")
				ui.Row(c).Gap(8).Wrap().GapY(10).Children(func() {
					for _, name := range filled {
						iconCell(c, name)
					}
				})
				SectionTitle(c, "TwoTone")
				ui.Row(c).Gap(8).Wrap().GapY(10).Children(func() {
					for _, name := range twotone {
						iconCell(c, name, IconProps{Theme: IconTwoTone, TwoTone: Work.Primary, Size: 22})
					}
				})
			})
		})
	}, 1440, 980)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	for _, want := range []string{"SearchOutlined", "HeartFilled", "CheckCircleTwoTone", "search", "848"} {
		if !tester.HasText(want) {
			t.Fatalf("icons gallery missing %q", want)
		}
	}
	shot(t, tester, "mygo-kit-icons.png")
}
