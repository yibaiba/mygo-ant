package kit

import (
	"strconv"
	"testing"

	"github.com/egoist/mygo/ui"
)

// galleryNav is what the layout and navigation pages edit.
type galleryLayout struct {
	tok       *Tokens
	collapsed bool
	side      string
	top       string
	topDark   string
	inline    string
	open      []string
	vertical  string
	tab       string
	cardTab   string
	cards     []Option
	leftTab   string
	manyTab   string
	step      int
	navStep   int
	anchor    AnchorState
	title     string
	sizes     []float32
	scroll    ui.ScrollState
	expanded  bool
	forceMenu string
}

func newGalleryLayout() *galleryLayout {
	return &galleryLayout{
		tok: Work, side: "meetings", top: "mail", topDark: "app", inline: "m-2", open: []string{"sub1"},
		vertical: "v-1", forceMenu: "v-sub", tab: "t1", cardTab: "c1", leftTab: "l2", manyTab: "p3", step: 1, navStep: 1,
		cards: []Option{{Value: "c1", Label: "议程"}, {Value: "c2", Label: "参会人"}, {Value: "c3", Label: "资料"}},
		title: "可编辑的会议标题",
	}
}

var galleryNavMenu = []MenuItem{
	{Key: "sub1", Label: "会议管理", Icon: LookupIcon("calendar"), Children: []MenuItem{
		{Key: "g1", Label: "进行中", Group: true, Children: []MenuItem{
			{Key: "m-1", Label: "今日会议"},
			{Key: "m-2", Label: "本周会议"},
		}},
		{Key: "g2", Label: "历史", Group: true, Children: []MenuItem{
			{Key: "m-3", Label: "已归档"},
		}},
	}},
	{Key: "sub2", Label: "资料中心", Icon: LookupIcon("folder"), Children: []MenuItem{
		{Key: "m-4", Label: "共享文件"},
		{Key: "sub3", Label: "模板", Children: []MenuItem{{Key: "m-5", Label: "议程模板"}, {Key: "m-6", Label: "纪要模板"}}},
	}},
	{Key: "m-7", Label: "系统设置", Icon: LookupIcon("setting")},
	{Divider: true},
	{Key: "m-8", Label: "禁用项", Icon: LookupIcon("lock"), Disabled: true},
	{Key: "m-9", Label: "清除数据", Icon: LookupIcon("delete"), Danger: true},
}

var galleryTopMenu = []MenuItem{
	{Key: "mail", Label: "导航一", Icon: LookupIcon("mail")},
	{Key: "app", Label: "导航二", Icon: LookupIcon("appstore")},
	{Key: "off", Label: "禁用导航", Icon: LookupIcon("lock"), Disabled: true},
	{Key: "sub", Label: "子菜单", Icon: LookupIcon("setting"), Children: []MenuItem{
		{Key: "s1", Label: "选项 1"}, {Key: "s2", Label: "选项 2"},
	}},
	{Key: "link", Label: "外部链接", Icon: LookupIcon("link")},
	{Key: "five", Label: "导航五"},
	{Key: "six", Label: "导航六"},
}

// shell is Layout with a dark Header, a collapsible Sider holding an
// inline Menu, Breadcrumb, Grid, Flex and Typography.
func (g *galleryLayout) shell(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		Layout(c, func() {
			LayoutHeader(c, func() {
				ui.Text(c, "Confidential").FontSize(18).FontWeight(600).TextColor(ui.Hex("#FFFFFF")).Width(168)
				ui.Box(c).Grow(1).Children(func() {
					Menu(c, &g.topDark, galleryTopMenu[:4], MenuProps{Mode: MenuHorizontal, Dark: true})
				})
			})
			Layout(c, func() {
				Sider(c, func(collapsed bool) {
					Menu(c, &g.side, []MenuItem{
						{Key: "meetings", Label: "会议", Icon: LookupIcon("calendar")},
						{Key: "people", Label: "人员", Icon: LookupIcon("team")},
						{Key: "files", Label: "文件", Icon: LookupIcon("file"), Children: []MenuItem{
							{Key: "f1", Label: "议程"}, {Key: "f2", Label: "纪要"},
						}},
						{Key: "settings", Label: "设置", Icon: LookupIcon("setting")},
					}, MenuProps{Dark: true, Collapsed: collapsed, Open: &[]string{"files"}})
				}, SiderProps{Collapsible: true, Collapsed: &g.collapsed})
				Layout(c, func() {
					LayoutContent(c, func() {
						ui.Scroll(c).Fill().Children(func() {
							ui.Column(c).FillWidth().Padding(16, 24).Gap(16).Children(func() {
								Breadcrumb(c, []BreadcrumbItem{
									{Key: "home", Icon: LookupIcon("home")},
									{Key: "list", Title: "会议列表", Icon: LookupIcon("calendar")},
									{Key: "room", Title: "会议室", Menu: []MenuItem{{Key: "r1", Label: "一号会议室"}, {Key: "r2", Label: "二号会议室"}}},
									{Key: "detail", Title: "会议详情"},
								})
								ui.Column(c).FillWidth().Padding(24).Gap(16).Background(tok.Surface).Radius(8).Children(func() {
									Title(c, "栅格 Grid 与 Flex", TitleProps{Level: 4})
									gridDemo(c, tok)
									Flex(c, func() {
										Button(c, "主按钮", ButtonProps{Type: Primary})
										Button(c, "次按钮")
										Button(c, "文本", ButtonProps{Type: TextButton})
									}, FlexProps{Gap: Small, Justify: ui.SpaceBetween})
									g.typography(c)
								})
							})
						})
					})
					LayoutFooter(c, func() {
						ui.Text(c, "Demo ©2026 · Ant Design 布局").TextColor(tok.TextTertiary).FontSize(14).Grow(1).TextAlign(ui.Center)
					})
				})
			}, LayoutProps{HasSider: true})
		})
	})
}

func gridDemo(c *ui.Context, tok *Tokens) {
	cell := func(text string, strong bool) func() {
		return func() {
			ui.Box(c).FillWidth().Height(32).Center().Radius(2).
				Background(pick(strong, tok.Primary, tok.PrimaryHover)).Children(func() {
				ui.Text(c, text).FontSize(12).TextColor(ui.Hex("#FFFFFF"))
			})
		}
	}
	GridRow(c, func(g *GridCols) {
		for i := 0; i < 4; i++ {
			g.Col(6, cell("col-6", i%2 == 0))
		}
	}, RowProps{Gutter: [2]float32{16, 8}})
	GridRow(c, func(g *GridCols) {
		g.Col(8, cell("col-8", true))
		g.Col(8, cell("col-8 offset-8", false), ColProps{ColSpan: ColSpan{Offset: 8}})
	}, RowProps{Gutter: [2]float32{16, 8}})
	GridRow(c, func(g *GridCols) {
		g.Col(0, cell("xs-24 md-12 xl-6", true), ColProps{XS: &ColSpan{Span: 24}, MD: &ColSpan{Span: 12}, XL: &ColSpan{Span: 6}})
		g.Col(0, cell("xs-24 md-12 xl-18", false), ColProps{XS: &ColSpan{Span: 24}, MD: &ColSpan{Span: 12}, XL: &ColSpan{Span: 18}})
	}, RowProps{Gutter: [2]float32{16, 8}})
	GridRow(c, func(g *GridCols) {
		g.Col(4, cell("order 3", true), ColProps{ColSpan: ColSpan{Order: 3}})
		g.Col(4, cell("order 1", false), ColProps{ColSpan: ColSpan{Order: 1}})
		g.Col(4, cell("order 2", true), ColProps{ColSpan: ColSpan{Order: 2}})
		g.Col(0, cell("flex auto", false), ColProps{Auto: true})
	}, RowProps{Gutter: [2]float32{16, 8}, Justify: ui.SpaceBetween})
}

func (g *galleryLayout) typography(c *ui.Context) {
	tok := g.tok
	Title(c, "排版 Typography", TitleProps{Level: 4})
	ui.Column(c).FillWidth().Children(func() {
		for lvl := 1; lvl <= 5; lvl++ {
			Title(c.Key(lvl), "h"+strconv.Itoa(lvl)+" 标题", TitleProps{Level: lvl}).Margin(0)
		}
	})
	ui.Row(c).Gap(12).Wrap().GapY(8).AlignItems(ui.Center).Children(func() {
		TypoText(c, "默认")
		TypoText(c, "次要", TypoProps{Type: TypoSecondary})
		TypoText(c, "成功", TypoProps{Type: TypoSuccess})
		TypoText(c, "警告", TypoProps{Type: TypoWarning})
		TypoText(c, "危险", TypoProps{Type: TypoDanger})
		TypoText(c, "禁用", TypoProps{Type: TypoDisabled})
		TypoText(c, "标记", TypoProps{Mark: true})
		TypoText(c, "go test ./...", TypoProps{Code: true})
		TypoText(c, "Ctrl", TypoProps{Keyboard: true})
		TypoText(c, "下划线", TypoProps{Underline: true})
		TypoText(c, "删除线", TypoProps{Delete: true})
		TypoText(c, "加粗", TypoProps{Strong: true})
		TypoText(c, "斜体 Italic", TypoProps{Italic: true})
		TextLink(c, "链接 Link", "")
		TextLink(c, "危险链接", "", TypoProps{Type: TypoDanger})
	})
	ui.Row(c).Gap(24).Wrap().AlignItems(ui.Center).Children(func() {
		TypoText(c, "192.168.1.20:8443", TypoProps{Copyable: true})
		TypoText(c, "", TypoProps{Editable: &g.title})
	})
	ui.Box(c).Width(560).Children(func() {
		Paragraph(c, "Demo App把议程、资料、表决和纪要放在同一个受控的桌面应用中：资料在会前加密分发，会中按权限查看，会后统一回收；"+
			"管理员可以在设置页配置数据库、证书与网络，安装向导引导首次部署。这一段落演示两行省略与展开。",
			TypoProps{Rows: 2, Expandable: true, Collapsible: true, Expanded: &g.expanded})
	})
	_ = tok
}

// nav is Menu in every mode, Dropdown submenus, Tabs, StepBar and Anchor.
func (g *galleryLayout) nav(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 导航").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(8, 0).Children(func() {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
				SectionTitle(c, "菜单 Menu")
				Sheet(c, func() {
					Menu(c, &g.top, galleryTopMenu, MenuProps{Mode: MenuHorizontal, force: hover, forceKey: "app"})
					ui.Row(c).FillWidth().Gap(16).Wrap().GapY(16).AlignItems(ui.Start).Children(func() {
						ui.Box(c).Width(220).Children(func() {
							Menu(c, &g.inline, galleryNavMenu, MenuProps{Open: &g.open})
						})
						ui.Box(c).Width(180).Children(func() {
							Menu(c, &g.vertical, []MenuItem{
								{Key: "v-1", Label: "导航一", Icon: LookupIcon("mail")},
								{Key: "v-sub", Label: "导航二", Icon: LookupIcon("appstore"), Children: []MenuItem{
									{Key: "v-2", Label: "选项 5"}, {Key: "v-3", Label: "选项 6"},
								}},
								{Key: "v-4", Label: "导航三", Icon: LookupIcon("setting")},
							}, MenuProps{Mode: MenuVertical, force: hover, forceKey: g.forceMenu})
						})
						ui.Column(c).Gap(8).Children(func() {
							Menu(c, &g.inline, galleryNavMenu[:3], MenuProps{Dark: true, Collapsed: true})
						})
						ui.Box(c).Width(200).Children(func() {
							Menu(c, &g.inline, galleryNavMenu, MenuProps{Dark: true, Open: &g.open})
						})
					})
				})
				SectionTitle(c, "标签页 Tabs")
				Sheet(c, func() {
					Tabs(c, &g.tab, []Option{{Value: "t1", Label: "基础设置"}, {Value: "t2", Label: "数据库"}, {Value: "t3", Label: "证书"}},
						TabsProps{Centered: true, Extra: func() { Button(c, "额外操作", ButtonProps{Size: Small}) }})
					Tabs(c, &g.cardTab, g.cards, TabsProps{Type: TabsEditableCard,
						OnAdd: func() {
							v := "c" + strconv.Itoa(len(g.cards)+1)
							g.cards = append(g.cards, Option{Value: v, Label: "新标签 " + strconv.Itoa(len(g.cards)+1)})
						},
						OnRemove: func(v string) {
							for i, o := range g.cards {
								if o.Value == v {
									g.cards = append(g.cards[:i:i], g.cards[i+1:]...)
									break
								}
							}
						}})
					ui.Box(c).Width(520).Children(func() {
						var many []Option
						for i := 1; i <= 12; i++ {
							many = append(many, Option{Value: "p" + strconv.Itoa(i), Label: "页面 " + strconv.Itoa(i)})
						}
						Tabs(c, &g.manyTab, many)
					})
					ui.Row(c).Height(150).AlignItems(ui.Stretch).Children(func() {
						Tabs(c, &g.leftTab, []Option{{Value: "l1", Label: "左侧一"}, {Value: "l2", Label: "左侧二"}, {Value: "l3", Label: "左侧三"}},
							TabsProps{Position: TabsLeft})
						ui.Box(c).Padding(16).Children(func() { ui.Text(c, "左侧标签的内容").TextColor(tok.Text) })
					})
				})
			})
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
				SectionTitle(c, "步骤条 Steps")
				Sheet(c, func() {
					items := []StepItem{
						{Title: "已完成", Description: "数据库已配置"},
						{Title: "进行中", SubTitle: "00:00:08", Description: "正在签发证书"},
						{Title: "待处理", Description: "启动服务"},
					}
					StepBar(c, &g.step, items, StepBarProps{Clickable: true})
					StepBar(c, &g.step, items, StepBarProps{Small: true, Status: StepError})
					StepBar(c, &g.step, items, StepBarProps{Dot: true})
					StepBar(c, &g.step, items, StepBarProps{LabelVertical: true, HasPercent: true, Percent: 60})
					StepBar(c, &g.navStep, items, StepBarProps{Type: StepsNavigation})
					ui.Row(c).FillWidth().Gap(24).Children(func() {
						ui.Box(c).Width(260).Children(func() {
							StepBar(c, &g.step, []StepItem{
								{Title: "登录", Icon: LookupIcon("user")},
								{Title: "校验", Icon: LookupIcon("safety")},
								{Title: "完成", Icon: LookupIcon("check-circle")},
							}, StepBarProps{Vertical: true})
						})
						ui.Box(c).Grow(1).Children(func() {
							StepBar(c, &g.step, items, StepBarProps{Type: StepsInline})
						})
					})
				})
				SectionTitle(c, "面包屑与下拉 Breadcrumb · Dropdown")
				Sheet(c, func() {
					Breadcrumb(c, []BreadcrumbItem{
						{Key: "a", Title: "首页"}, {Key: "b", Title: "设置"}, {Key: "c", Title: "证书管理"},
					}, BreadcrumbProps{Separator: ">"})
					ui.Row(c).Gap(16).Height(220).AlignItems(ui.Start).Children(func() {
						DropdownButton(c, "级联菜单", []MenuItem{
							{Key: "a", Label: "第一项"},
							{Key: "grp", Label: "分组", Group: true, Children: []MenuItem{{Key: "g1", Label: "分组项"}}},
							{Key: "sub", Label: "子菜单", Children: []MenuItem{{Key: "s1", Label: "三级项 1"}, {Key: "s2", Label: "三级项 2"}}},
							{Divider: true},
							{Key: "d", Label: "删除", Danger: true},
						}, DropdownProps{open: true})
					})
				})
			})
		})
	})
}

// extras is Splitter, Anchor, Affix and the floating buttons.
func (g *galleryLayout) extras(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 分隔面板 / 锚点 / 悬浮按钮").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Stretch).Padding(16).Gap(16).Children(func() {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(16).Children(func() {
				ui.Column(c).FillWidth().Height(260).Background(tok.Surface).Border(1, tok.BorderSecondary).Radius(8).Clip().Children(func() {
					pane := func(text string) func() {
						return func() {
							ui.Box(c).Fill().Center().Children(func() { ui.Text(c, text).TextColor(tok.TextTertiary) })
						}
					}
					Splitter(c, []SplitPanel{
						{Render: pane("左侧 · 最小 80"), Min: 80, Collapsible: true},
						{Render: func() {
							Splitter(c, []SplitPanel{
								{Render: pane("上方")},
								{Render: pane("下方 · 可折叠"), Collapsible: true},
							}, SplitterProps{Vertical: true})
						}},
						{Render: pane("右侧 · 固定"), Fixed: true, Size: 160},
					}, SplitterProps{Sizes: &g.sizes})
				})
				ui.Row(c).FillWidth().Height(420).Gap(16).AlignItems(ui.Start).Children(func() {
					ui.Column(c).Width(180).Padding(8).Background(tok.Surface).Radius(8).Children(func() {
						Anchor(c, &g.anchor, []AnchorLink{
							{Key: "a1", Title: "基础设置"},
							{Key: "a2", Title: "数据库", Children: []AnchorLink{{Key: "a2-1", Title: "连接池"}}},
							{Key: "a3", Title: "证书"},
							{Key: "a4", Title: "网络"},
						})
					})
					sc := ui.Scroll(c).Grow(1).FillHeight().Background(tok.Surface).Radius(8).TrackScroll(&g.anchor.Scroll)
					g.anchor.Container(sc)
					sc.Children(func() {
						ui.Column(c).FillWidth().Padding(16).Gap(16).Children(func() {
							Affix(c, func() {
								ui.Row(c).FillWidth().Height(36).Padding(0, 12).AlignItems(ui.Center).Background(tok.PrimaryBg).Radius(6).Children(func() {
									ui.Text(c, "固钉 Affix：滚动时停在顶部").TextColor(tok.Primary)
								})
							}, AffixProps{Target: sc, Scroll: &g.anchor.Scroll})
							for i, k := range []string{"a1", "a2", "a2-1", "a3", "a4"} {
								sec := ui.Column(c.Key(k)).FillWidth().Height(220).Padding(12).Radius(6).
									Background(pick(i%2 == 0, tok.FillAlter, tok.Surface)).Border(1, tok.BorderSecondary)
								sec.Children(func() { Title(c.Key(k), "分区 "+k, TitleProps{Level: 5}) })
								g.anchor.Mark(k, sec)
							}
						})
					})
				})
			})
			ui.Column(c).Width(360).Gap(16).Children(func() {
				ui.Column(c).FillWidth().Padding(16).Gap(12).Background(tok.Surface).Radius(8).Children(func() {
					Title(c, "悬浮按钮 FloatButton", TitleProps{Level: 5})
					ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
						floatButtonBody(c.Key(1), "默认", FloatButtonProps{}, 0)
						floatButtonBody(c.Key(2), "主要", FloatButtonProps{Primary: true, Icon: LookupIcon("question-circle")}, 0)
						floatButtonBody(c.Key(3), "方形", FloatButtonProps{Square: true, Icon: LookupIcon("file"), Description: "文档"}, 0)
						floatButtonBody(c.Key(4), "徽标", FloatButtonProps{Icon: LookupIcon("bell"), Badge: &BadgeProps{Count: 5}}, 0)
						floatButtonBody(c.Key(5), "悬停", FloatButtonProps{Icon: LookupIcon("setting"), force: hover}, 0)
					})
				})
			})
		})
		FloatButtonGroup(c, []FloatItem{
			{Key: "q", Label: "帮助", FloatButtonProps: FloatButtonProps{Icon: LookupIcon("question-circle")}},
			{Key: "s", Label: "设置", FloatButtonProps: FloatButtonProps{Icon: LookupIcon("setting")}},
			{Key: "t", Label: "顶部", FloatButtonProps: FloatButtonProps{Icon: LookupIcon("vertical-align-top")}},
		}, FloatGroupProps{Square: true, Right: 120})
		FloatButtonGroup(c, []FloatItem{
			{Key: "c", Label: "评论", FloatButtonProps: FloatButtonProps{Icon: LookupIcon("mail")}},
			{Key: "e", Label: "编辑", FloatButtonProps: FloatButtonProps{Icon: LookupIcon("edit")}},
		}, FloatGroupProps{Menu: true, Primary: true, Icon: LookupIcon("plus"), open: true, Right: 200})
		BackTop(c, &g.anchor.Scroll, BackTopProps{VisibilityHeight: 100, Bottom: 104})
		FloatButton(c, "客服", FloatButtonProps{Icon: LookupIcon("question-circle"), Primary: true, Tooltip: "联系客服"})
	})
}

func TestGalleryLayoutShots(t *testing.T) {
	g := newGalleryLayout()
	shell := ui.NewTester(g.shell, 1440, 900)
	shell.SetLocale("zh-CN")
	shell.Frame()
	shell.Frame()
	for _, want := range []string{"会议详情", "col-8 offset-8", "h1 标题", "go test ./...", "Demo ©2026 · Ant Design 布局"} {
		if !shell.HasText(want) {
			t.Fatalf("layout gallery missing %q", want)
		}
	}
	shot(t, shell, "mygo-kit-gallery-layout.png")
	g.collapsed = true
	shell.Frame()
	shot(t, shell, "mygo-kit-gallery-layout-collapsed.png")

	n := newGalleryLayout()
	nav := ui.NewTester(n.nav, 1440, 1500)
	nav.SetLocale("zh-CN")
	nav.Frame()
	nav.Frame()
	shot(t, nav, "mygo-kit-gallery-nav.png")

	d := newGalleryLayout()
	d.tok = WorkDark
	darkNav := ui.NewTester(d.nav, 1440, 1500)
	darkNav.SetLocale("zh-CN")
	darkNav.Frame()
	darkNav.Frame()
	shot(t, darkNav, "mygo-kit-gallery-dark-nav.png")

	e := newGalleryLayout()
	extras := ui.NewTester(e.extras, 1440, 820)
	extras.SetLocale("zh-CN")
	extras.Frame()
	extras.Frame()
	e.anchor.Scroll.Y = 300
	extras.Frame()
	extras.Frame()
	shot(t, extras, "mygo-kit-gallery-layout-extras.png")
}
