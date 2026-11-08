package kit

import (
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

var (
	galleryUser  = svgIcon(`<circle cx="8" cy="5.5" r="2.8"/><path d="M2.8 14c.6-2.9 2.7-4.4 5.2-4.4s4.6 1.5 5.2 4.4"/>`)
	galleryClock = svgIcon(`<circle cx="8" cy="8" r="6"/><path d="M8 4.5V8l2.4 1.6"/>`)
	galleryEdit  = svgIcon(`<path d="M3 13h2.5l7-7L10 3.5l-7 7z"/><path d="m8.8 4.7 2.5 2.5"/>`)
	galleryGear  = svgIcon(`<circle cx="8" cy="8" r="2.2"/><path d="M8 1.8v1.8M8 12.4v1.8M1.8 8h1.8M12.4 8h1.8M3.6 3.6l1.3 1.3M11.1 11.1l1.3 1.3M3.6 12.4l1.3-1.3M11.1 4.9l1.3-1.3"/>`)
)

// galleryDisplay is what the data display pages edit.
type galleryDisplay struct {
	tok        *Tokens
	tags       []string
	checks     [4]bool
	tab        string
	listPage   int
	slide      int
	day        time.Time
	mode       CalendarMode
	pics       []ui.ImageSource
	deadline   time.Time
	tourOpen   bool
	tourStep   int
	tourTarget ui.Element
	preview    bool
	previewAt  int
}

func newGalleryDisplay(t testing.TB) *galleryDisplay {
	return &galleryDisplay{
		tok: Work, tags: []string{"一号会议室", "保密", "已归档"}, checks: [4]bool{true, false, true, false},
		tab: "agenda", listPage: 1, day: time.Date(2026, 11, 6, 0, 0, 0, 0, time.Local),
		pics: []ui.ImageSource{
			testPicture(t, 320, 200, color.RGBA{22, 119, 255, 255}, color.RGBA{114, 46, 209, 255}),
			testPicture(t, 320, 200, color.RGBA{82, 196, 26, 255}, color.RGBA{19, 194, 194, 255}),
			testPicture(t, 320, 200, color.RGBA{250, 140, 22, 255}, color.RGBA{245, 34, 45, 255}),
		},
	}
}

func (g *galleryDisplay) header(c *ui.Context, title string) {
	ui.Row(c).FillWidth().Height(40).Background(g.tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
		ui.Text(c, title).FontSize(16).FontWeight(600).TextColor(g.tok.TextOnPrimary)
	})
}

// basics is Avatar, Badge, Tag, Statistic, Timeline and QRCode.
func (g *galleryDisplay) basics(c *ui.Context) {
	tok := g.tok
	if g.deadline.IsZero() {
		g.deadline = c.Now().Add(26*time.Hour + 15*time.Minute + 30*time.Second)
	}
	Page(c, tok, func() {
		g.header(c, "组件库 · 数据展示")
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(8, 0).Children(func() {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
				SectionTitle(c, "头像 Avatar")
				Sheet(c, func() {
					row(c, "尺寸与形状", func() {
						Avatar(c, "张", AvatarProps{Size: Large})
						Avatar(c, "李", AvatarProps{Color: ui.Hex("#f56a00"), TextColor: ui.Hex("#fde3cf")})
						Avatar(c, "", AvatarProps{Icon: galleryUser, Size: Small})
						Avatar(c, "", AvatarProps{Icon: galleryUser, Square: true, Size: Large, Color: ui.Hex("#87d068")})
						Avatar(c, "", AvatarProps{Src: g.pics[0], Size: Large})
						Avatar(c, "Administrator", AvatarProps{Size: Large, Color: tok.Primary})
						Avatar(c, "会", AvatarProps{Pixels: 64, Color: ui.Hex("#7265e6")})
					})
					row(c, "头像组 · 最多 3", func() {
						AvatarGroup(c, []AvatarItem{
							{Text: "张", AvatarProps: AvatarProps{Color: ui.Hex("#f56a00")}},
							{Text: "李", AvatarProps: AvatarProps{Color: ui.Hex("#7265e6")}},
							{Text: "王", AvatarProps: AvatarProps{Color: ui.Hex("#ffbf00")}},
							{Text: "赵", AvatarProps: AvatarProps{Color: ui.Hex("#00a2ae")}},
							{Text: "钱"},
						}, AvatarGroupProps{MaxCount: 3, MoreColor: ui.Hex("#fde3cf"), MoreTextColor: ui.Hex("#f56a00")})
					})
				})
				SectionTitle(c, "徽标 Badge · 缎带 Ribbon")
				Sheet(c, func() {
					square := func() { Avatar(c, "", AvatarProps{Square: true, Size: Large, Icon: galleryUser}) }
					row(c, "计数", func() {
						BadgeCount(c, square, BadgeProps{Count: 5})
						BadgeCount(c, square, BadgeProps{Count: 120})
						BadgeCount(c, square, BadgeProps{Count: 1000, Overflow: 999})
						BadgeCount(c, square, BadgeProps{Dot: true, Count: 1})
						BadgeCount(c, square, BadgeProps{Count: 8, Offset: [2]float32{-6, 6}, Preset: "blue"})
						BadgeCount(c, square, BadgeProps{Text: "NEW", Preset: "volcano"})
						BadgeCount(c, nil, BadgeProps{Count: 25, Preset: "green"})
						BadgeCount(c, nil, BadgeProps{Count: 4, Size: Small, Color: tok.TextDisabled})
					})
					row(c, "状态点", func() {
						StatusDot(c, "success", "运行中")
						StatusDot(c, "processing", "同步中")
						StatusDot(c, "error", "已停止")
						StatusDot(c, "warning", "告警")
						StatusDot(c, "default", "未启动")
						StatusDot(c, "", "紫色", StatusDotProps{Preset: "purple"})
					})
					ui.Row(c).FillWidth().Gap(24).Padding(0, 8, 0, 0).Children(func() {
						ui.Column(c).Grow(1).Basis(0).Children(func() {
							Ribbon(c, "推荐", func() {
								Card(c, func() { Muted(c, "缎带挂在卡片右上角，折角压在边缘之下。") }, CardProps{Title: "季度经营分析会"})
							})
						})
						ui.Column(c).Grow(1).Basis(0).Children(func() {
							Ribbon(c, "加急", func() {
								Card(c, func() { Muted(c, "Start 放在左侧，可选任意预设色。") }, CardProps{Title: "保密培训"})
							}, RibbonProps{Preset: "red", Start: true})
						})
					})
				})
				SectionTitle(c, "标签 Tag")
				Sheet(c, func() {
					row(c, "预设色", func() {
						for _, name := range []string{"magenta", "red", "volcano", "orange", "gold", "lime", "green", "cyan", "blue", "geekblue", "purple"} {
							Tag(c.Key(name), name, TagProps{Preset: name})
						}
					})
					row(c, "自定义 · 图标", func() {
						Tag(c, "#f50", TagProps{Color: ui.Hex("#f50")})
						Tag(c, "#108ee9", TagProps{Color: ui.Hex("#108ee9")})
						Tag(c, "定时", TagProps{Icon: galleryClock, Preset: "blue"})
						Tag(c, "无边框", TagProps{Borderless: true, Tone: TagSuccess})
					})
					row(c, "可关闭", func() {
						for i, s := range g.tags {
							if Tag(c.Key(s), s, TagProps{Closable: true}) {
								g.tags = append(g.tags[:i:i], g.tags[i+1:]...)
							}
						}
					})
					row(c, "可选标签", func() {
						for i, s := range []string{"议程", "纪要", "表决", "签到"} {
							CheckableTag(c.Key(s), s, &g.checks[i])
						}
					})
				})
			})
			ui.Column(c).Width(560).Shrink(0).Children(func() {
				SectionTitle(c, "统计数值 Statistic · 倒计时")
				Sheet(c, func() {
					ui.Row(c).FillWidth().Gap(32).Wrap().GapY(16).Children(func() {
						Statistic(c, "今日会议", 112893)
						Statistic(c, "签到率", 93.12, StatisticProps{Precision: 2, Suffix: "%", Color: ui.Hex("#3f8600"), PrefixIcon: galleryClock})
						Statistic(c, "账户余额 (CNY)", 112893.5, StatisticProps{Precision: 2, Prefix: "¥"})
						Statistic(c, "加载中", 0, StatisticProps{Loading: true})
						Countdown(c, "距离开会", g.deadline)
						Countdown(c, "距离截止", g.deadline, CountdownProps{Format: "D 天 H 时 m 分 s 秒"})
					})
				})
				SectionTitle(c, "时间轴 Timeline")
				Sheet(c, func() {
					ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Width(200).Shrink(0).Children(func() {
							Timeline(c, []TimelineItem{
								{Content: "创建会议 09:00"},
								{Content: "上传材料 09:20", Color: "green"},
								{Content: "密级审批未通过", Color: "red"},
								{Content: "自定义节点", Dot: func() { ui.Icon(c, galleryClock).Size(16, 16).TextColor(tok.Primary) }},
							}, TimelineProps{Pending: "记录中…"})
						})
						ui.Column(c).Grow(1).MinWidth(0).Children(func() {
							Timeline(c, []TimelineItem{
								{Label: "2026-11-01", Content: "立项"},
								{Label: "2026-11-03", Content: "评审", Color: "gray"},
								{Label: "2026-11-05", Content: "上线", Color: "purple"},
							}, TimelineProps{Mode: TimelineAlternate})
						})
					})
				})
				SectionTitle(c, "二维码 QRCode")
				Sheet(c, func() {
					ui.Row(c).Gap(16).Wrap().GapY(12).Children(func() {
						QRCode(c, "https://example.com/join?room=1001", QRCodeProps{Size: 112})
						QRCode(c, "https://example.com/join?room=1002", QRCodeProps{Size: 112, Level: QRLevelH, Icon: g.pics[2], IconSize: 28})
						QRCode(c, "https://example.com/join?room=1003", QRCodeProps{Size: 112, Color: tok.Primary, Status: QRExpired})
						QRCode(c, "https://example.com/join?room=1004", QRCodeProps{Size: 112, Status: QRScanned})
					})
				})
			})
		})
	})
}

// cards is Card, List, Carousel, Image and the card Calendar.
func (g *galleryDisplay) cards(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		g.header(c, "组件库 · 卡片与列表")
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(8, 0).Children(func() {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
				SectionTitle(c, "卡片 Card")
				ui.Row(c).FillWidth().Gap(16).Padding(0, 20).AlignItems(ui.Start).Children(func() {
					Card(c, func() {
						CardMeta(c, CardMetaProps{Avatar: func() { Avatar(c, "", AvatarProps{Src: g.pics[1]}) }, Title: "会议材料", Description: "共 12 份文件，3 份待审"})
					}, CardProps{Width: 240, Hoverable: true, hover: true,
						Cover: func() { ui.Image(c, g.pics[0]).Size(240, 120).Fit(ui.Cover) },
						Actions: []func(){
							func() { ui.Icon(c, galleryGear).Size(16, 16).TextColor(tok.TextTertiary) },
							func() { ui.Icon(c, galleryEdit).Size(16, 16).TextColor(tok.TextTertiary) },
							func() { Button(c, "更多", ButtonProps{Type: Link, Size: Small}) },
						}})
					ui.Column(c).Grow(1).MinWidth(0).Gap(16).Children(func() {
						Card(c, func() {
							switch g.tab {
							case "agenda":
								Muted(c, "1. 审议季度报告  2. 表决预算方案")
							default:
								Muted(c, "参会 24 人，缺席 1 人")
							}
						}, CardProps{Title: "季度经营分析会", Tabs: []Option{{Value: "agenda", Label: "议程"}, {Value: "people", Label: "人员"}}, Tab: &g.tab,
							TabExtra: func() { Button(c, "导出", ButtonProps{Type: Link, Size: Small}) }})
						Card(c, func() {}, CardProps{Title: "加载中", Loading: true, Size: Small})
					})
				})
				ui.Column(c).FillWidth().Padding(16, 20, 0).Children(func() {
					Card(c, func() {
						CardGrid(c, 6, func(i int) { Muted(c, "终端 "+itoa(i+1)) }, CardGridProps{Columns: 3, Hoverable: true})
					}, CardProps{Title: "网格卡片 Card.Grid"})
				})
				SectionTitle(c, "列表 List")
				Sheet(c, func() {
					List(c, 12, func(i int) {
						ListItem(c, func() {
							ListItemMeta(c, ListMetaProps{
								Avatar: func() {
									Avatar(c, string([]rune("张李王赵钱孙周吴郑冯陈褚")[i]), AvatarProps{Color: tok.Primary})
								},
								Title:       "会议终端 " + itoa(i+1),
								Description: "10.0.0." + itoa(i+11) + " · 在线",
							})
						}, ListItemProps{Actions: []func(){
							func() { Button(c.Key(i), "编辑", ButtonProps{Type: Link, Size: Small}) },
							func() { Button(c.Key(i), "移除", ButtonProps{Type: Link, Size: Small, Danger: true}) },
						}})
					}, ListProps{Bordered: true, Page: &g.listPage, PageSize: 3,
						Header: func() { ui.Text(c, "终端列表").FontWeight(600).TextColor(tok.Text) },
						Footer: func() { Muted(c, "共 12 台") }})
					List(c, 4, func(i int) {
						Card(c, func() { Muted(c, "席位 "+itoa(i*8+8)) }, CardProps{Title: "会议室 " + itoa(i+1), Size: Small})
					}, ListProps{Grid: 4})
				})
			})
			ui.Column(c).Width(420).Shrink(0).Children(func() {
				SectionTitle(c, "走马灯 Carousel")
				Sheet(c, func() {
					Carousel(c, 3, func(i int) {
						ui.Column(c).Fill().Center().Background([]ui.Color{tok.Primary, ui.Hex("#13c2c2"), ui.Hex("#722ed1")}[i]).Children(func() {
							ui.Text(c, "第 "+itoa(i+1)+" 屏").FontSize(18).TextColor(tok.TextOnPrimary)
						})
					}, CarouselProps{Arrows: true, Current: &g.slide, Height: 140})
					Carousel(c, 3, func(i int) {
						ui.Image(c, g.pics[i]).Fill().Fit(ui.Cover)
					}, CarouselProps{Fade: true, DotPosition: "right", Height: 120, Current: &g.slide})
				})
				SectionTitle(c, "图片 Image")
				Sheet(c, func() {
					ImageGroup(c, g.pics, ImageProps{Width: 112, Height: 72, Radius: 4})
				})
				SectionTitle(c, "日历 Calendar · 卡片")
				Sheet(c, func() {
					Calendar(c, &g.day, CalendarProps{Card: true, Mode: &g.mode})
				})
			})
		})
	})
}

// calendar is the full calendar with events in its cells.
func (g *galleryDisplay) calendar(c *ui.Context) {
	tok := g.tok
	events := map[int][]struct {
		text, preset string
	}{
		3:  {{"预算评审", "blue"}},
		6:  {{"季度经营分析会", "green"}, {"保密培训", "orange"}},
		12: {{"系统升级", "red"}},
		18: {{"项目周例会", "blue"}, {"议题征集", "purple"}, {"纪要归档", "cyan"}},
	}
	Page(c, tok, func() {
		g.header(c, "组件库 · 日历")
		Sheet(c, func() {
			Calendar(c, &g.day, CalendarProps{Mode: &g.mode, CellRender: func(d time.Time) {
				if d.Month() != g.day.Month() {
					return
				}
				ui.Column(c).Gap(4).Children(func() {
					for _, e := range events[d.Day()] {
						StatusDot(c, "", e.text, StatusDotProps{Preset: e.preset})
					}
				})
			}, MonthCellRender: func(m time.Time) {
				if m.Month() == 11 {
					Statistic(c, "", 42, StatisticProps{Suffix: "场", ValueSize: 16})
				}
			}})
		})
	})
}

// preview is the image preview and a tour, each over a small page.
func (g *galleryDisplay) overlays(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		g.header(c, "组件库 · 图片预览")
		Sheet(c, func() {
			ImageGroup(c, g.pics, ImageProps{Width: 160, Height: 100})
		})
		ImagePreview(c, &g.preview, g.pics, &g.previewAt)
	})
}

func (g *galleryDisplay) tour(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		g.header(c, "组件库 · 漫游式引导 Tour")
		Sheet(c, func() {
			ui.Row(c).Gap(12).Children(func() {
				Button(c, "上传材料")
				g.tourTarget = ui.Row(c).Children(func() { Button(c, "保存设置", ButtonProps{Type: Primary}) })
				Button(c, "更多操作")
			})
			Muted(c, "引导卡片指向“保存设置”，遮罩只在目标四周，目标仍可点击。")
		})
		Tour(c, &g.tourOpen, &g.tourStep, []TourStep{
			{Title: "上传材料", Description: "先上传会议材料"},
			{Title: "保存设置", Description: "修改完成后点击这里保存，配置会立即下发到终端。", Placement: PlaceBottom,
				Target: func() ui.Element { return g.tourTarget },
				Cover:  func() { ui.Image(c, g.pics[0]).Size(488, 120).Fit(ui.Cover).Radius(6) }},
			{Title: "完成", Description: "引导结束"},
		})
	})
}

func TestGalleryDisplayShots(t *testing.T) {
	g := newGalleryDisplay(t)
	basics := ui.NewTester(g.basics, 1440, 1180)
	basics.SetLocale("zh-CN")
	basics.Frame()
	basics.Frame()
	for _, want := range []string{"99+", "999+", "+2", "112,893", "93.12", "记录中…", "二维码过期", "已扫描"} {
		if !basics.HasText(want) {
			t.Fatalf("display gallery missing %q", want)
		}
	}
	shot(t, basics, "mygo-kit-gallery-display.png")

	dark := newGalleryDisplay(t)
	dark.tok = WorkDark
	darkPage := ui.NewTester(dark.basics, 1440, 1180)
	darkPage.SetLocale("zh-CN")
	darkPage.Frame()
	darkPage.Frame()
	shot(t, darkPage, "mygo-kit-gallery-dark-display.png")

	cards := ui.NewTester(g.cards, 1440, 1260)
	cards.SetLocale("zh-CN")
	cards.Frame()
	cards.Frame()
	if err := cards.Click("下一张"); err != nil {
		t.Fatal(err)
	}
	cards.Frame()
	shot(t, cards, "mygo-kit-gallery-display-cards.png")

	cal := ui.NewTester(g.calendar, 1280, 900)
	cal.SetLocale("zh-CN")
	cal.Frame()
	cal.Frame()
	shot(t, cal, "mygo-kit-gallery-display-calendar.png")

	g.preview, g.previewAt = true, 1
	pv := ui.NewTester(g.overlays, 1100, 760)
	pv.SetLocale("zh-CN")
	pv.Frame()
	pv.Frame()
	pv.Advance(300 * time.Millisecond)
	pv.Wheel(550, 380, 0, -1)
	pv.Frame()
	if err := pv.Click("向右旋转"); err != nil {
		t.Fatal(err)
	}
	pv.Frame()
	shot(t, pv, "mygo-kit-gallery-display-preview.png")

	g.tourOpen, g.tourStep = true, 1
	tour := ui.NewTester(g.tour, 1100, 640)
	tour.SetLocale("zh-CN")
	tour.Frame()
	tour.Frame()
	tour.Frame()
	shot(t, tour, "mygo-kit-gallery-display-tour.png")
}
