package kit

import (
	"cmp"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// galleryMore is what the data display and data entry pages edit.
type galleryMore struct {
	sel, none                int
	tab, card                string
	port, prefix, bad, empty string
	day, noDay               time.Time
	clock, from, to          time.Time
	fonts, oneFile, noFiles  []string
	nic, role, mode          string
	tok                      *Tokens
	sort                     SortState
	page, bigPage, pageSize  int
	form                     Form
	host, ip, pass           string
}

func newGalleryMore() *galleryMore {
	return &galleryMore{
		sel: 1, none: -1, tab: "base", card: "server",
		port: "8080", prefix: "24", bad: "70000",
		day:   time.Date(2026, 11, 5, 0, 0, 0, 0, time.Local),
		clock: time.Date(2026, 11, 5, 8, 30, 0, 0, time.Local),
		from:  time.Date(2026, 11, 3, 0, 0, 0, 0, time.Local),
		to:    time.Date(2026, 11, 12, 0, 0, 0, 0, time.Local),
		fonts: []string{"/usr/share/fonts/SourceHanSans-Regular.otf", "/usr/share/fonts/仿宋_GB2312.ttf"},
		nic:   "eth0", role: "normal", mode: "static", tok: Work,
		sort: SortState{Col: 3, Order: SortDescend}, page: 2, bigPage: 6, pageSize: 10,
		ip: "172.30.0.300", pass: "会议2026", form: Form{submitted: true},
	}
}

var probeColumns = []Column{
	{Title: "目标", Width: 140},
	{Title: "类型", Width: 100},
	{Title: "结果", Width: 90, Sorter: func(a, b int) int { return cmp.Compare(pick(probeRows[a].ok, 1, 0), pick(probeRows[b].ok, 1, 0)) }},
	{Title: "耗时", Width: 96, Align: AlignEnd, Sorter: ByNumber(func(r int) float64 {
		n, err := strconv.ParseFloat(strings.TrimSuffix(probeRows[r].ms, " ms"), 64)
		return pick(err == nil, n, -1)
	})},
	{Title: "说明", Min: 120},
}

type probeRow struct {
	label, kind string
	ok          bool
	ms, detail  string
}

var probeRows = []probeRow{
	{"管理接口", "HTTP", true, "12 ms", "200 OK"},
	{"消息推送", "WebSocket", false, "-", "连接被拒绝：目标主机 172.30.0.9:36678 没有在监听这个端口"},
	{"文档服务", "TCP", true, "3 ms", "已连接"},
	{"Socket.IO", "Socket.IO", true, "41 ms", "握手完成"},
}

func probeCell(c *ui.Context) func(r, col int) Cell {
	return func(r, col int) Cell {
		row := probeRows[r]
		switch col {
		case 0:
			return Cell{Text: row.label, Strong: true}
		case 1:
			return Cell{Text: row.kind}
		case 2:
			return Cell{Render: func() { Badge(c, pick(row.ok, "正常", "失败"), row.ok) }}
		case 3:
			return Cell{Text: row.ms}
		}
		return Cell{Text: row.detail, Color: pick(row.ok, Of(c).TextSecondary, Of(c).Danger)}
	}
}

// data is the data display page: Table, Descriptions, Tabs, Divider,
// Space, Tooltip and Message.
func (g *galleryMore) data(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 数据展示").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(8, 0).Children(func() {
			ui.Column(c).Grow(1).Basis(0).Children(func() {
				SectionTitle(c, "表格 Table")
				Sheet(c, func() {
					Muted(c, "表头 40、行分隔线、悬停行、选中行、右对齐耗时、说明列超长省略；“结果”“耗时”可排序，当前按耗时降序")
					Table(c, probeColumns, len(probeRows), probeCell(c), TableProps{Selected: &g.sel, forceHover: 3, Sort: &g.sort})
					ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Grow(1).Basis(0).Gap(8).Children(func() {
							Muted(c, "空状态")
							Table(c, []Column{{Title: "IP", Width: 140}, {Title: "MAC"}}, 0, nil, TableProps{Selected: &g.none, Empty: "没有发现直连设备"})
							Muted(c, "加载中")
							Table(c, []Column{{Title: "IP", Width: 140}, {Title: "MAC"}}, 1, func(r, col int) Cell {
								return Cell{Text: pick(col == 0, "192.168.0.1", "aa:bb:cc:dd:ee:01")}
							}, TableProps{Selected: &g.none, Loading: true})
						})
						ui.Column(c).Grow(1).Basis(0).Gap(8).Children(func() {
							Muted(c, "小尺寸、换行、选中且悬停")
							Table(c, []Column{{Title: "名称", Width: 110}, {Title: "说明", Wrap: true}}, 2, func(r, col int) Cell {
								if col == 0 {
									return Cell{Text: []string{"第一会议室", "备用服务器"}[r]}
								}
								return Cell{Text: []string{"指纹 9f2c4e8a11d0b7c3e5a6f48291bd0c7e", "离线"}[r]}
							}, TableProps{Size: Small, Selected: &g.sel, forceHover: 2})
						})
					})
				})
				SectionTitle(c, "分页 Pagination / 横向滚动")
				Sheet(c, func() {
					ui.Row(c).FillWidth().Gap(24).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Width(360).Shrink(0).Gap(8).Children(func() {
							Muted(c, "列宽之和 600 > 360：横向滚动，表格分页 3 条/页")
							Table(c, []Column{{Title: "名称", Width: 140}, {Title: "地址", Width: 140}, {Title: "口令码", Width: 100}, {Title: "指纹", Width: 220}}, 7, func(r, col int) Cell {
								return Cell{Text: []string{"会议室 " + strconv.Itoa(r+1), "172.30.0." + strconv.Itoa(9+r), "48" + strconv.Itoa(10+r), "9f2c4e8a11d0b7c3e5a6f4"}[col]}
							}, TableProps{Size: Small, PageSize: 3, Page: &g.page})
						})
						ui.Column(c).Grow(1).Basis(0).Gap(12).Children(func() {
							Muted(c, "独立使用，共 95 条，第 6 页")
							Pagination(c, &g.bigPage, 95, 10, PaginationProps{ShowTotal: true})
							Muted(c, "跳至 / 条/页")
							Pagination(c, &g.bigPage, 95, g.pageSize, PaginationProps{ShowTotal: true, ShowQuickJumper: true, ShowSizeChanger: true, PageSize: &g.pageSize})
							Muted(c, "小尺寸 / 禁用")
							Pagination(c, &g.page, 30, 10, PaginationProps{Size: Small})
							Pagination(c, &g.page, 30, 10, PaginationProps{Size: Small, Disabled: true})
						})
					})
				})
				SectionTitle(c, "描述列表 Descriptions")
				Sheet(c, func() {
					Descriptions(c, []DescItem{
						{Label: "主题", Value: "CN=example-host, O=Example Inc.", Span: 2},
						{Label: "序列号", Value: "6F2A11", Copyable: true},
						{Label: "算法", Value: "SM2"},
						{Label: "有效期", Value: "2026-01-01 ~ 2028-01-01", Span: 2},
					}, DescriptionsProps{Title: "服务器证书", Columns: 2, Bordered: true})
					Descriptions(c, []DescItem{
						{Label: "平台", Value: "linux/amd64"},
						{Label: "安装目录", Value: "/opt/example"},
						{Label: "数据库", Value: "mysql"},
						{Label: "状态", Render: func() { Badge(c, "正常", true) }},
					}, DescriptionsProps{Columns: 2})
				})
			})
			ui.Column(c).Width(460).Padding(0, 16, 0, 0).Gap(14).Children(func() {
				Card(c, func() {
					Tabs(c, &g.tab, []Option{
						{Value: "base", Label: "基础设置"}, {Value: "db", Label: "数据库"},
						{Value: "off", Label: "已停用", Disabled: true}, {Value: "fonts", Label: "字体"},
					}, TabsProps{force: hover, forceItem: 2})
					Muted(c, "线条式：选中带下划线，悬停“数据库”，“已停用”禁用")
					Tabs(c, &g.card, []Option{{Value: "server", Label: "服务器证书"}, {Value: "tls", Label: "TLS 证书"}},
						TabsProps{Type: TabsCard, Extra: func() { Button(c, "导入", ButtonProps{Size: Small}) }})
					Muted(c, "卡片式，带右侧操作")
				}, CardProps{Title: "标签页 Tabs"})
				Card(c, func() {
					Divider(c, DividerProps{Text: "网关和 DNS"})
					Divider(c, DividerProps{Dashed: true})
					Space(c, func() {
						Button(c, "链接一", ButtonProps{Type: Link})
						Divider(c, DividerProps{Vertical: true})
						Button(c, "链接二", ButtonProps{Type: Link})
						Divider(c, DividerProps{Vertical: true})
						Tag(c, "在线", TagProps{Tone: TagSuccess})
					})
					Space(c, func() {
						Button(c, "保存", ButtonProps{Type: Primary})
						Button(c, "取消")
					}, SpaceProps{Gap: 16})
				}, CardProps{Title: "分隔线 Divider / 间距 Space"})
				Card(c, func() {
					ui.Column(c).FillWidth().Gap(18).Padding(72, 0, 40).AlignItems(ui.Center).Children(func() {
						shownTip(c, "完整指纹 9f2c4e8a11d0b7c3e5a6f48291bd0c7e", PlaceTop, func() {
							Button(c, "9f2c4e8a…91bd0c7e", ButtonProps{Type: TextButton, force: hover})
						})
						ui.Row(c).FillWidth().Justify(ui.SpaceBetween).Padding(0, 72).Children(func() {
							shownTip(c, "左侧", PlaceLeft, func() { Button(c, "左") })
							shownTip(c, "右侧", PlaceRight, func() { Button(c, "右") })
						})
						shownTip(c, "下方提示", PlaceBottom, func() { Button(c, "下") })
					})
				}, CardProps{Title: "文字提示 Tooltip（上 / 左 / 右 / 下，带箭头）"})
				Card(c, func() {
					Message(c, "success", "初始化已提交")
					Message(c, "error", "连接失败：目标主机拒绝连接")
					Message(c, "warning", "PowerShell 执行策略受限")
					Message(c, "info", "正在搜索会场…")
				}, CardProps{Title: "全局提示 Message"})
			})
		})
	})
}

// entry is the data entry page: InputNumber, DatePicker, Upload,
// Popconfirm, Select states, disabled options.
func (g *galleryMore) entry(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 数据录入").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(8, 0).Children(func() {
			ui.Column(c).Grow(1).Basis(0).Children(func() {
				SectionTitle(c, "数字输入框 InputNumber")
				Sheet(c, func() {
					FormRow(c, func() {
						FormItem(c, "端口 (1–65535)", func() { InputNumber(c, &g.port, InputNumberProps{Min: 1, Max: 65535}) }, FormItemProps{Min: 160})
						FormItem(c, "前缀长度 · 悬停", func() { InputNumber(c, &g.prefix, InputNumberProps{Min: 1, Max: 32, force: hover}) }, FormItemProps{Min: 160})
						FormItem(c, "聚焦", func() { InputNumber(c, &g.port, InputNumberProps{Min: 1, Max: 65535, force: focus}) }, FormItemProps{Min: 160})
					})
					FormRow(c, func() {
						FormItem(c, "大尺寸", func() { InputNumber(c, &g.prefix, InputNumberProps{Min: 1, Max: 32, Size: Large}) }, FormItemProps{Min: 160, Size: Large})
						FormItem(c, "禁用", func() { InputNumber(c, &g.port, InputNumberProps{Disabled: true}) }, FormItemProps{Min: 160})
						FormItem(c, "出错", func() { InputNumber(c, &g.bad, InputNumberProps{Status: StatusError}) }, FormItemProps{Min: 160, Error: "端口超出范围"})
					})
				})
				SectionTitle(c, "上传 Upload")
				Sheet(c, func() {
					ui.Row(c).FillWidth().Gap(16).Wrap().GapY(12).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Grow(1).Basis(240).MinWidth(240).Gap(8).Children(func() {
							Upload(c, &g.fonts, UploadProps{Hint: "支持 ttf / otf / ttc，可多选", Multiple: true})
						})
						ui.Column(c).Grow(1).Basis(240).MinWidth(240).Gap(8).Children(func() {
							Upload(c, &g.noFiles, UploadProps{Text: "松开即可添加", Hint: "拖入时", dragOver: true})
						})
					})
					Upload(c, &g.oneFile, UploadProps{Hint: "禁用", Disabled: true})
				})
				SectionTitle(c, "时间 TimePicker / 日期范围 RangePicker")
				Sheet(c, func() {
					ui.Row(c).FillWidth().Gap(24).AlignItems(ui.Start).Justify(ui.SpaceBetween).Children(func() {
						ui.Column(c).Width(300).Children(func() {
							FormItem(c, "日志范围", func() {
								RangePicker(c, &g.from, &g.to, RangePickerProps{Width: 300, open: true})
							})
						})
						ui.Column(c).Width(168).Children(func() {
							FormItem(c, "定时启动", func() { TimePicker(c, &g.clock, TimePickerProps{Width: 168, open: true}) })
						})
					})
					ui.Box(c).Height(300)
				})
			})
			ui.Column(c).Width(460).Padding(0, 16, 0, 0).Gap(14).Children(func() {
				Card(c, func() {
					FormItem(c, "日期", func() { DatePicker(c, &g.day, DatePickerProps{Width: 200, open: true}) })
					ui.Box(c).Height(300)
					ui.Row(c).Gap(12).Children(func() {
						DatePicker(c, &g.noDay, DatePickerProps{Width: 160})
						DatePicker(c, &g.noDay, DatePickerProps{Width: 160, Disabled: true})
					})
				}, CardProps{Title: "日期选择 DatePicker"})
				Card(c, func() {
					ui.Box(c).Height(104)
					Popconfirm(c, "确定卸载服务器？", func() bool {
						return Button(c, "卸载服务器", ButtonProps{Type: Primary, Danger: true})
					}, PopconfirmProps{Description: "将卸载本机本机应用。", OK: "卸载", Danger: true, open: true})
				}, CardProps{Title: "气泡确认 Popconfirm"})
				Card(c, func() {
					FormRow(c, func() {
						FormItem(c, "加载中", func() { Select(c, &g.nic, Opts("eth0"), SelectProps{Loading: true}) }, FormItemProps{Min: 160})
						FormItem(c, "账户（含禁用项）", func() {
							Select(c, &g.role, []Option{{Value: "normal", Label: "普通管理员"}, {Value: "super", Label: "超级管理员", Disabled: true}})
						}, FormItemProps{Min: 160})
					})
					RadioButtons(c, &g.mode, []Option{{Value: "static", Label: "静态地址"}, {Value: "dhcp", Label: "DHCP"}, {Value: "pppoe", Label: "拨号", Disabled: true}})
					Segmented(c, &g.mode, []Option{{Value: "static", Label: "静态"}, {Value: "dhcp", Label: "自动"}, {Value: "pppoe", Label: "拨号", Disabled: true}})
				}, CardProps{Title: "禁用选项 / 加载"})
				Card(c, func() {
					f := &g.form
					f.Begin()
					item := func(label, name string, value *string, rules ...Rule) {
						err := f.Check(name, *value, rules...)
						FormItem(c, label, func() { Input(c, value, InputProps{Status: ErrorStatus(err)}) }, FormItemProps{Required: true, Error: err})
					}
					item("数据库地址", "host", &g.host, Required("请输入数据库地址"))
					item("服务器 IP", "ip", &g.ip, Required("请输入服务器 IP"), IPv4("请输入正确的 IPv4 地址"))
					item("备份密码", "pass", &g.pass, Required("请输入密码"), func(v string) string {
						return pick(strings.ContainsFunc(v, func(r rune) bool { return r >= 0x4e00 && r <= 0x9fff }), "密码不能包含中文", "")
					})
				}, CardProps{Title: "表单校验 Form（提交后）"})
			})
		})
	})
}

// shownTip is Tooltip held open, for the gallery.
func shownTip(c *ui.Context, text string, side Placement, fn func()) {
	t := Of(c)
	anchor := ui.Row(c).Shrink(0).Children(fn)
	ui.Overlay(c, func() {
		wrap := ui.Box(c).PassThrough()
		arrowPanel(c, wrap, anchor, side, t.TooltipBg, func(bubble ui.Element) { tipLook(c, bubble, text) })
	})
}
