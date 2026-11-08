package kit

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// galleryState is what the gallery's controls edit.
type galleryState struct {
	text, empty, pass, sel, mode, seg, radio, nav string
	on, off, checked, unchecked                   bool
	sel2                                          int
	dialog                                        bool
	tok                                           *Tokens
}

func newGallery() *galleryState {
	return &galleryState{
		text: "Example Inc.", pass: "secret", sel: "eth0", mode: "static", seg: "day",
		radio: "normal", nav: "base", on: true, checked: true, sel2: 1, tok: Work,
	}
}

var galleryNav = []NavItem{{ID: "base", Label: "基础设置"}, {ID: "database", Label: "数据库设置"}, {ID: "security", Label: "安全参数设置"}, {ID: "backup", Label: "密钥备份与恢复"}}

func row(c *ui.Context, label string, fn func()) {
	ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Center).Wrap().GapY(8).Children(func() {
		ui.Text(c, label).FontSize(12).TextColor(Of(c).TextMuted).Width(96).Shrink(0)
		fn()
	})
}

func (g *galleryState) work(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 工作页").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
			NavList(c, &g.nav, galleryNav)
			ui.Column(c).Grow(1).Padding(8, 0).Children(func() {
				SectionTitle(c, "按钮 Button")
				Sheet(c, func() {
					row(c, "类型", func() {
						Button(c, "保存基础设置", ButtonProps{Type: Primary})
						Button(c, "选择")
						Button(c, "文本按钮", ButtonProps{Type: TextButton})
						Button(c, "链接按钮", ButtonProps{Type: Link})
						Button(c, "卸载服务器", ButtonProps{Type: Primary, Danger: true})
						Button(c, "删除", ButtonProps{Danger: true})
					})
					row(c, "悬停 / 按下", func() {
						Button(c, "悬停", ButtonProps{Type: Primary, force: hover})
						Button(c, "按下", ButtonProps{Type: Primary, force: active})
						Button(c, "悬停", ButtonProps{force: hover})
						Button(c, "聚焦", ButtonProps{force: focus})
						Button(c, "文本悬停", ButtonProps{Type: TextButton, force: hover})
					})
					row(c, "禁用 / 加载", func() {
						Button(c, "应用到系统", ButtonProps{Type: Primary, Disabled: true})
						Button(c, "刷新预览", ButtonProps{Disabled: true})
						Button(c, "连接中", ButtonProps{Type: Primary, Loading: true})
						Button(c, "导出诊断包", ButtonProps{Loading: true})
					})
					row(c, "尺寸", func() {
						Button(c, "小", ButtonProps{Size: Small})
						Button(c, "中 32", ButtonProps{Type: Primary})
						Button(c, "大 40", ButtonProps{Type: Primary, Size: Large})
						Button(c, "按 IP 登录", ButtonProps{Type: Primary, Size: Large})
					})
					row(c, "幽灵 / 形状", func() {
						Button(c, "幽灵主色", ButtonProps{Type: Primary, Ghost: true})
						Button(c, "幽灵危险", ButtonProps{Type: Primary, Danger: true, Ghost: true})
						Button(c, "圆角", ButtonProps{Type: Primary, Shape: ShapeRound})
						Button(c, "", ButtonProps{Type: Primary, Shape: ShapeCircle, Icon: LookupIcon("SearchOutlined")})
					})
				})
				SectionTitle(c, "表单 Form")
				Sheet(c, func() {
					FormRow(c, func() {
						FormItem(c, "公司名称", func() { Input(c, &g.text) }, FormItemProps{Min: 220, Required: true})
						FormItem(c, "悬停", func() { Input(c, &g.empty, InputProps{Placeholder: "例如 192.168.1.10", force: hover}) }, FormItemProps{Min: 160})
						FormItem(c, "聚焦", func() { Input(c, &g.empty, InputProps{Placeholder: "聚焦中", force: focus}) }, FormItemProps{Min: 160})
					})
					FormRow(c, func() {
						FormItem(c, "密码", func() { Input(c, &g.pass, InputProps{Password: true}) }, FormItemProps{Min: 160, Help: "留空则保持"})
						FormItem(c, "禁用", func() { Input(c, &g.text, InputProps{Disabled: true}) }, FormItemProps{Min: 160})
						FormItem(c, "出错", func() { Input(c, &g.empty, InputProps{Status: StatusError}) }, FormItemProps{Min: 160, Error: "请输入 IPv4 地址"})
					})
					FormRow(c, func() {
						FormItem(c, "网卡 (40)", func() { Select(c, &g.sel, Opts("eth0", "eth1"), SelectProps{Size: Large}) }, FormItemProps{Min: 200, Size: Large})
						FormItem(c, "悬停", func() { Select(c, &g.sel, Opts("eth0"), SelectProps{force: hover}) }, FormItemProps{Min: 160})
						FormItem(c, "禁用", func() { Select(c, &g.empty, nil, SelectProps{Disabled: true, Placeholder: "请选择"}) }, FormItemProps{Min: 160})
					})
					FormItem(c, "离线安装包", func() {
						PathPicker(c, &g.empty, "选择安装包", PathPickerProps{Size: Large, Placeholder: "请选择 zip"})
					}, FormItemProps{Size: Large})
					FormRow(c, func() {
						FormItem(c, "配置方式", func() {
							RadioButtons(c, &g.mode, []Option{{Value: "static", Label: "静态地址"}, {Value: "dhcp", Label: "自动获取 (DHCP)"}})
						}, FormItemProps{Min: 240})
						FormItem(c, "分段", func() {
							Segmented(c, &g.seg, []Option{{Value: "day", Label: "日"}, {Value: "week", Label: "周"}, {Value: "month", Label: "月"}})
						}, FormItemProps{Min: 160})
						FormItem(c, "账户", func() {
							ui.Row(c).Gap(16).Children(func() {
								Radio(c, &g.radio, "normal", "普通管理员")
								Radio(c, &g.radio, "super", "超级管理员")
								Radio(c, &g.radio, "x", "禁用", ChoiceProps{Disabled: true})
							})
						}, FormItemProps{Min: 300})
					})
					FormRow(c, func() {
						Slot(c, Middle, func() {
							Checkbox(c, &g.checked, "开机自启")
							Checkbox(c, &g.unchecked, "开启静默模式")
							Checkbox(c, &g.unchecked, "悬停", ChoiceProps{force: hover})
							Checkbox(c, &g.checked, "禁用", ChoiceProps{Disabled: true})
						})
						Slot(c, Middle, func() {
							Switch(c, &g.on)
							Switch(c, &g.off)
							Switch(c, &g.on, ChoiceProps{Size: Small})
							Switch(c, &g.on, ChoiceProps{Loading: true})
							Switch(c, &g.off, ChoiceProps{Disabled: true})
							Switch(c, &g.on, ChoiceProps{CheckedChildren: "开", UnCheckedChildren: "关"})
						})
					}, FormRowProps{Gap: 32})
				})
			})
			ui.Column(c).Width(420).Padding(8, 16, 8, 0).Gap(14).Children(func() {
				Card(c, func() {
					Alert(c, "正在管理 172.30.0.9:36678", AlertProps{Banner: true})
					Alert(c, "已断开 172.30.0.9:36678", AlertProps{Banner: true, Type: Error})
					Alert(c, "改完会跟到新地址", AlertProps{Type: Info})
					Alert(c, "初始化已提交", AlertProps{Type: Success})
					Alert(c, "脚本策略", AlertProps{Type: Warning, Description: "PowerShell 执行策略受限，安装脚本可能无法运行。"})
					Alert(c, "连接失败", AlertProps{Type: Error})
				}, CardProps{Title: "提示 Alert / Banner"})
				Card(c, func() {
					ui.Row(c).Gap(8).Wrap().GapY(8).Children(func() {
						Tag(c, "默认")
						Tag(c, "在线", TagProps{Tone: TagSuccess})
						Tag(c, "主节点", TagProps{Tone: TagBlue})
						Tag(c, "待升级", TagProps{Tone: TagWarning})
						Tag(c, "离线", TagProps{Tone: TagError})
					})
					ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
						Tag(c, "mysql", TagProps{Tone: TagBrand, Size: Large})
						Badge(c, "正常", true)
						Badge(c, "已停止", false)
					})
					Progress(c, 62)
					Progress(c, 35, ProgressProps{Error: true})
				}, CardProps{Title: "标签 / 状态 / 进度"})
				Card(c, func() {
					Steps(c, 1, []string{"路径设置", "数据库设置", "网络设置"}, StepsProps{Compact: true})
				}, CardProps{Title: "步骤 Steps"})
				Card(c, func() {
					Table(c, []Column{{Title: "IP", Width: 150}, {Title: "MAC"}}, 3, func(r, col int) Cell {
						ips := []string{"192.168.0.1", "192.168.0.23", "192.168.1.7"}
						if col == 0 {
							return Cell{Text: ips[r]}
						}
						return Cell{Text: "aa:bb:cc:dd:ee:0" + string(rune('1'+r))}
					}, TableProps{Selected: &g.sel2})
					Empty(c, "尚未扫描")
				}, CardProps{Title: "表格 / 空状态", Extra: func() { Button(c, "扫描", ButtonProps{Type: Primary, Size: Small}) }})
				Card(c, func() {
					NavChips(c, &g.nav, galleryNav)
				}, CardProps{Title: "窄屏导航 NavChips"})
			})
		})
	})
	if g.dialog {
		Dialog(c, &g.dialog, "Demo Service日志", func() {
			Muted(c, "2026-11-05 14:40:00 服务已启动")
		}, DialogProps{Width: 420, Footer: func() { Button(c, "关闭") }})
	}
}

func (g *galleryState) home(c *ui.Context) {
	Page(c, Home, func() {
		ui.Column(c).Fill().Children(func() {
			HomeBar(c, HomeBarProps{
				Title:    "Demo App",
				Subtitle: "高效 · 稳定 · 安全",
				Width:    1280,
				Nav: func() {
					TopNav(c, "连接", TopNavProps{Current: true})
					TopNav(c, "检测")
					TopNav(c, "启动", TopNavProps{Disabled: true})
					TopNav(c, "设置")
					TopNav(c, "安装")
				},
			})
			ui.Column(c).FillWidth().Padding(8, 36).Gap(16).Children(func() {
				HomeHeader(c, "服务管理", HomeHeaderProps{
					TitleSize: 28,
					Extra: func() {
						HealthPill(c, "服务安全运行中", "共 7 个服务模块", true)
					},
				})
				HealthPill(c, "1 个服务未运行", "共 7 个服务模块", false)
				ui.Grid(c).Columns(3).GapX(20).GapY(22).FillWidth().Children(func() {
					for i, name := range []string{"Demo Service", "缓存服务", "数据服务"} {
						ok := i != 2
						on := ok
						status := "正常"
						if !ok {
							status = "已停止"
						}
						GlassCard(c.Key(i), GlassCardProps{
							IconFn: func() {
								ui.Box(c).Size(64, 64).Radius(16).Background(ui.Hex("#D6E8FF"))
							},
							Title:    name,
							Subtitle: "会议文档与议程",
							Status:   status,
							StatusOK: ok,
							Extra:    "6379",
							On:       &on,
							Actions:  func() { Button(c, "日志", ButtonProps{Type: Link, Size: Small}) },
						})
					}
				})
				ui.Row(c).Gap(12).Children(func() {
					Button(c, "启动", ButtonProps{Type: Primary})
					Button(c, "取消")
					TopNav(c, "极简按钮", TopNavProps{Large: true})
				})
			})
		})
	})
}

func TestGalleryShots(t *testing.T) {
	g := newGallery()
	work := ui.NewTester(g.work, 1440, 1300)
	work.SetLocale("zh-CN")
	work.Frame()
	for _, want := range []string{"保存基础设置", "自动获取 (DHCP)", "请输入 IPv4 地址", "尚未扫描", "路径设置"} {
		if !work.HasText(want) {
			t.Fatalf("gallery missing %q", want)
		}
	}
	shot(t, work, "mygo-kit-gallery.png")

	home := ui.NewTester(g.home, 1280, 640)
	home.SetLocale("zh-CN")
	home.Frame()
	for _, want := range []string{"服务管理", "Demo Service", "已停止", "服务安全运行中", "1 个服务未运行"} {
		if !home.HasText(want) {
			t.Fatalf("home gallery missing %q", want)
		}
	}
	shot(t, home, "mygo-kit-gallery-home.png")

	g.dialog = true
	modal := ui.NewTester(g.work, 1440, 1300)
	modal.SetLocale("zh-CN")
	modal.Frame()
	open := true
	confirm := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {})
		Confirm(c, &open, "停止服务", "是否停止所有服务", "确定", "取消")
	}, 640, 360)
	confirm.SetLocale("zh-CN")
	confirm.Frame()
	shot(t, confirm, "mygo-kit-gallery-confirm.png")

	more := newGalleryMore()
	data := ui.NewTester(more.data, 1440, 1340)
	data.SetLocale("zh-CN")
	data.Frame()
	data.Frame()
	for _, want := range []string{"没有发现直连设备", "服务器证书", "6F2A11", "网关和 DNS", "初始化已提交"} {
		if !data.HasText(want) {
			t.Fatalf("data gallery missing %q", want)
		}
	}
	shot(t, data, "mygo-kit-gallery-data.png")

	entry := ui.NewTester(more.entry, 1440, 1400)
	entry.SetLocale("zh-CN")
	entry.Frame()
	entry.Frame()
	for _, want := range []string{"确定卸载服务器？", "点击或拖拽文件到此区域", "SourceHanSans-Regular.otf", "2026-11-05", "请输入数据库地址", "密码不能包含中文"} {
		if !entry.HasText(want) {
			t.Fatalf("entry gallery missing %q", want)
		}
	}
	shot(t, entry, "mygo-kit-gallery-entry.png")

	dark := newGallery()
	dark.tok = WorkDark
	darkWork := ui.NewTester(dark.work, 1440, 1300)
	darkWork.SetLocale("zh-CN")
	darkWork.Frame()
	shot(t, darkWork, "mygo-kit-gallery-dark.png")

	darkMore := newGalleryMore()
	darkMore.tok = WorkDark
	darkData := ui.NewTester(darkMore.data, 1440, 1340)
	darkData.SetLocale("zh-CN")
	darkData.Frame()
	darkData.Frame()
	shot(t, darkData, "mygo-kit-gallery-dark-data.png")
}

func shot(t *testing.T, tester *ui.Tester, name string) {
	t.Helper()
	tester.Advance(600 * time.Millisecond)
	img := tester.Image()
	if img == nil {
		t.Fatal("no frame")
	}
	dir := "/opt/cursor/artifacts"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
