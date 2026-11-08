package kit

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// galleryEntry is what the data-entry pages edit.
type galleryEntry struct {
	tok *Tokens

	note, short, query, query2, otp, pin, mention string
	fruits                                        []string
	rate, rate2, rate3                            float64
	vol, vert, tip                                float64
	band                                          [2]float64
	db, tags                                      []string
	one                                           string
	target                                        []string
	checked, expanded                             []string
	treeSel                                       string

	path, multiPath, treeVals []string
	treeOne, mail             string
	color, color2             ui.Color

	day, week, quarter, year, stamp, clock time.Time
	from, to                               time.Time
}

func newGalleryEntry() *galleryEntry {
	day := time.Date(2026, 11, 6, 14, 30, 0, 0, time.Local)
	return &galleryEntry{
		tok:  Work,
		note: "会议纪要需在会后 24 小时内归档，涉密材料按规定销毁。", short: "自动高度",
		otp: "2580", pin: "1234", mention: "@张三 请审阅本次议程",
		fruits: []string{"pdf", "ofd"}, rate: 3.5, rate2: 4, rate3: 6,
		vol: 40, vert: 30, tip: 60, band: [2]float64{20, 70},
		db: []string{"mysql", "pg", "dm", "kingbase"}, tags: []string{"会议室A", "终端"}, one: "dm",
		target: []string{"t3", "t5"}, checked: []string{"hz", "xh", "bj"}, expanded: []string{"zj", "hz", "js"}, treeSel: "nj",
		path: []string{"zj", "hz", "xh"}, multiPath: []string{"hz", "nj"}, treeVals: []string{"xh", "bj"}, treeOne: "hz", mail: "a",
		color: ui.Hex("#1677ff"), color2: ui.RGBA(82, 196, 26, 0.6),
		day: day, week: day, quarter: day, year: day, stamp: day, clock: day,
		from: day.AddDate(0, 0, -3), to: day.AddDate(0, 0, 4),
	}
}

var entryFormats = []Option{{Value: "pdf", Label: "PDF"}, {Value: "ofd", Label: "OFD"}, {Value: "docx", Label: "Word"}, {Value: "xlsx", Label: "Excel", Disabled: true}}

var entryPeople = []Option{{Value: "张三", Label: "张三 · 主持人"}, {Value: "李四", Label: "李四 · 记录员"}, {Value: "王五", Label: "王五 · 参会"}}

var entryTerminals = func() []TransferItem {
	var out []TransferItem
	for i, n := range []string{"一号会议室", "二号会议室", "主席台", "签到屏", "投影终端", "备用终端"} {
		out = append(out, TransferItem{Key: "t" + string(rune('0'+i)), Title: n, Description: "终端", Disabled: i == 4})
	}
	return out
}()

// entry is the inputs, choices, sliders and lists of Batch 1.
func (g *galleryEntry) entry(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 数据录入").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(8, 0).Children(func() {
			ui.Column(c).Grow(1).Basis(0).Children(func() {
				SectionTitle(c, "文本域 · 搜索 · 验证码 · 提及")
				Sheet(c, func() {
					row(c, "TextArea", func() {
						TextArea(c, &g.note, TextAreaProps{Rows: 3, ShowCount: true, MaxLength: 100, Width: 300})
						TextArea(c, &g.short, TextAreaProps{AutoSize: true, AllowClear: true, Width: 160})
					})
					row(c, "Search", func() {
						Search(c, &g.query, SearchProps{Placeholder: "搜索会议", Width: 200})
						Search(c, &g.query2, SearchProps{Placeholder: "搜索终端", EnterButton: "搜索", Loading: true, Width: 240})
					})
					row(c, "OTP", func() {
						OTP(c, &g.otp, OTPProps{Length: 6})
						OTP(c, &g.pin, OTPProps{Length: 4, Mask: true, Size: Small})
					})
					row(c, "Mentions", func() {
						ui.Box(c).Width(320).Children(func() { Mentions(c, &g.mention, entryPeople, MentionsProps{Rows: 2}) })
					})
				})
				SectionTitle(c, "多选框 · 评分 · 滑动条")
				Sheet(c, func() {
					row(c, "Checkbox", func() {
						CheckAll(c, &g.fruits, entryFormats, "全部格式")
						CheckboxGroup(c, &g.fruits, entryFormats)
					})
					row(c, "Rate", func() {
						Rate(c, &g.rate, RateProps{AllowHalf: true})
						Rate(c, &g.rate2, RateProps{Character: "好", Size: 18})
						Rate(c, &g.rate3, RateProps{Count: 10, Size: 14})
						Muted(c, "3.5 星")
					})
					row(c, "Slider", func() {
						ui.Column(c).Width(380).Gap(28).Children(func() {
							Slider(c, &g.vol, SliderProps{Marks: []SliderMark{{0, "0°C"}, {26, "26°C"}, {37, "37°C"}, {100, "100°C"}}, Dots: true})
							RangeSlider(c, &g.band, SliderProps{Step: 10, Tooltip: TooltipAlways})
							Slider(c, &g.tip, SliderProps{Disabled: true})
						})
						ui.Row(c).Gap(24).Children(func() {
							Slider(c, &g.vert, SliderProps{Vertical: true, Length: 140, Marks: []SliderMark{{0, "低"}, {50, "中"}, {100, "高"}}})
							RangeSlider(c, &g.band, SliderProps{Vertical: true, Length: 140})
						})
					})
				})
				SectionTitle(c, "穿梭框 Transfer")
				Sheet(c, func() {
					Transfer(c, entryTerminals, &g.target, TransferProps{Titles: [2]string{"可用终端", "已分配"}, ShowSearch: true, Height: 260})
				})
			})
			ui.Column(c).Grow(1).Basis(0).Children(func() {
				SectionTitle(c, "选择器 Select · 多选 / 标签 / 分组搜索")
				Sheet(c, func() {
					row(c, "多选", func() {
						MultiSelect(c, &g.db, dbGalleryOptions, SelectProps{Width: 300, MaxTagCount: 2, AllowClear: true})
					})
					row(c, "标签", func() {
						MultiSelect(c, &g.tags, nil, SelectProps{Width: 300, Mode: Tags, Placeholder: "输入后回车"})
					})
					row(c, "分组搜索", func() {
						ui.Column(c).Height(300).Children(func() {
							Select(c, &g.one, dbGalleryOptions, SelectProps{Width: 240, ShowSearch: true, AllowClear: true, open: true})
						})
					})
				})
				SectionTitle(c, "树 Tree · 可勾选 / 连接线")
				Sheet(c, func() {
					ui.Row(c).FillWidth().Gap(24).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Grow(1).Basis(0).Children(func() {
							Tree(c, galleryRegions, TreeProps{Checkable: true, Checked: &g.checked, Expanded: &g.expanded})
						})
						ui.Column(c).Grow(1).Basis(0).Children(func() {
							Tree(c, galleryRegions, TreeProps{ShowLine: true, Selected: &g.treeSel, DefaultExpandAll: true, Filter: "江"})
						})
					})
				})
			})
		})
	})
}

var dbGalleryOptions = []Option{
	{Value: "mysql", Label: "MySQL", Group: "开源"},
	{Value: "pg", Label: "PostgreSQL", Group: "开源"},
	{Value: "sqlite", Label: "SQLite", Group: "开源"},
	{Value: "dm", Label: "达梦", Group: "国产"},
	{Value: "kingbase", Label: "人大金仓", Group: "国产"},
	{Value: "oscar", Label: "神通", Group: "国产", Disabled: true},
}

var galleryRegions = []TreeNode{
	{Key: "zj", Title: "浙江", Children: []TreeNode{
		{Key: "hz", Title: "杭州", Children: []TreeNode{{Key: "xh", Title: "西湖"}, {Key: "bj", Title: "滨江"}}},
		{Key: "nb", Title: "宁波", Children: []TreeNode{{Key: "jb", Title: "江北"}, {Key: "yz", Title: "鄞州", Disabled: true}}},
	}},
	{Key: "js", Title: "江苏", Children: []TreeNode{{Key: "nj", Title: "南京"}, {Key: "sz", Title: "苏州"}}},
}

// pickers is the select-like popups: cascader, tree select, auto
// complete and color picker, all open.
func (g *galleryEntry) pickers(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 级联 / 树选择 / 自动完成 / 颜色").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Column(c).Fill().Padding(8, 0).Children(func() {
			SectionTitle(c, "级联选择 Cascader · 单选展开 / 多选")
			Sheet(c, func() {
				ui.Row(c).FillWidth().Height(260).Gap(40).AlignItems(ui.Start).Children(func() {
					Cascader(c, &g.path, galleryRegions, CascaderProps{SelectProps: SelectProps{Width: 260, open: true}})
					ui.Box(c).Width(320)
					CascaderMultiple(c, &g.multiPath, galleryRegions, CascaderProps{SelectProps: SelectProps{Width: 320, open: true}})
				})
			})
			SectionTitle(c, "树选择 TreeSelect · 单选 / 可勾选")
			Sheet(c, func() {
				ui.Row(c).FillWidth().Height(300).Gap(40).AlignItems(ui.Start).Children(func() {
					TreeSelect(c, &g.treeOne, galleryRegions, TreeSelectProps{SelectProps: SelectProps{Width: 260, open: true}, DefaultExpandAll: true})
					TreeMultiSelect(c, &g.treeVals, galleryRegions, TreeSelectProps{SelectProps: SelectProps{Width: 320, open: true}, Checkable: true, DefaultExpandAll: true, ShowLine: true})
				})
			})
			SectionTitle(c, "自动完成 · 颜色选择器")
			Sheet(c, func() {
				ui.Row(c).FillWidth().Height(440).Gap(40).AlignItems(ui.Start).Children(func() {
					AutoComplete(c, &g.mail, []Option{{Value: "admin@corp.cn"}, {Value: "audit@corp.cn"}, {Value: "ops@corp.cn"}}, AutoCompleteProps{Width: 260})
					ColorPicker(c, &g.color, ColorPickerProps{ShowText: true, open: true, AllowClear: true, Presets: []ColorPreset{
						{Label: "推荐", Colors: []ui.Color{ui.Hex("#1677ff"), ui.Hex("#52c41a"), ui.Hex("#faad14"), ui.Hex("#ff4d4f"), ui.Hex("#722ed1"), ui.Hex("#13c2c2")}},
					}})
					ui.Box(c).Width(200)
					ColorPicker(c, &g.color2, ColorPickerProps{ShowText: true, Format: FormatRGB})
				})
			})
		})
	})
}

// dates is the date and time panels open.
func (g *galleryEntry) dates(c *ui.Context) {
	tok := g.tok
	today := g.day
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 日期与时间").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Column(c).Fill().Padding(8, 0).Children(func() {
			SectionTitle(c, "日期 · 预设 + 时间 / 周")
			Sheet(c, func() {
				ui.Row(c).FillWidth().Height(400).Gap(24).AlignItems(ui.Start).Children(func() {
					DatePicker(c, &g.stamp, DatePickerProps{Width: 220, ShowTime: true, open: true, Presets: []DatePreset{
						{Label: "昨天", Value: today.AddDate(0, 0, -1)}, {Label: "上周", Value: today.AddDate(0, 0, -7)}, {Label: "上月", Value: today.AddDate(0, -1, 0)},
					}})
					ui.Box(c).Width(500)
					DatePicker(c, &g.week, DatePickerProps{Width: 200, Picker: PickWeek, open: true, DisabledDate: func(d time.Time) bool { return d.After(today.AddDate(0, 0, 10)) }})
				})
			})
			SectionTitle(c, "月 / 季度 / 年 · 时间 12 小时制")
			Sheet(c, func() {
				ui.Row(c).FillWidth().Height(330).Gap(24).AlignItems(ui.Start).Children(func() {
					DatePicker(c, &g.day, DatePickerProps{Width: 180, Picker: PickMonth, open: true})
					ui.Box(c).Width(110)
					DatePicker(c, &g.quarter, DatePickerProps{Width: 180, Picker: PickQuarter, open: true})
					ui.Box(c).Width(110)
					DatePicker(c, &g.year, DatePickerProps{Width: 180, Picker: PickYear, open: true})
					ui.Box(c).Width(110)
					TimePicker(c, &g.clock, TimePickerProps{Width: 180, Use12Hours: true, MinuteStep: 15, HideSeconds: true, open: true,
						DisabledHours: func() []int { return []int{0, 1, 2, 3, 4, 5} }})
				})
			})
			SectionTitle(c, "范围 RangePicker · 预设 / 禁用过去 / 清除")
			Sheet(c, func() {
				ui.Column(c).FillWidth().Height(360).Children(func() {
					RangePicker(c, &g.from, &g.to, RangePickerProps{
						Width: 320, open: true, AllowClear: true,
						DisabledDate: func(d time.Time) bool { return d.Before(dayOf(today)) },
						Presets: []RangePreset{
							{Label: "最近 7 天", Start: today.AddDate(0, 0, -7), End: today},
							{Label: "最近 30 天", Start: today.AddDate(0, 0, -30), End: today},
							{Label: "本月", Start: time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.Local), End: today},
						},
					})
				})
			})
		})
	})
}

// rangeDisable is RangePicker DisabledDate (past days grey) and AllowClear.
func (g *galleryEntry) rangeDisable(c *ui.Context) {
	tok := g.tok
	today := dayOf(g.day)
	start, end := today, today.AddDate(0, 0, 4)
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · RangePicker 禁用日期与清除").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Column(c).Fill().Padding(12).Gap(12).Children(func() {
			SectionTitle(c, "RangePicker DisabledDate（禁用过去）· AllowClear")
			Sheet(c, func() {
				ui.Column(c).FillWidth().Height(360).Children(func() {
					RangePicker(c, &start, &end, RangePickerProps{
						Width: 360, open: true, AllowClear: true,
						DisabledDate: func(d time.Time) bool { return d.Before(today) },
					})
				})
			})
		})
	})
}

func TestGalleryEntryShots(t *testing.T) {
	g := newGalleryEntry()
	page := ui.NewTester(g.entry, 1440, 1320)
	page.SetLocale("zh-CN")
	page.Frame()
	page.Frame()
	for _, want := range []string{"全部格式", "可用终端", "人大金仓", "西湖", "0°C"} {
		if !page.HasText(want) {
			t.Fatalf("entry gallery missing %q", want)
		}
	}
	shot(t, page, "mygo-kit-gallery-entry-input.png")

	pick := ui.NewTester(g.pickers, 1440, 1360)
	pick.SetLocale("zh-CN")
	pick.Frame()
	pick.Frame()
	shot(t, pick, "mygo-kit-gallery-pickers.png")

	ac := ui.NewTester(func(c *ui.Context) {
		Page(c, g.tok, func() {
			ui.Column(c).Padding(20).Height(220).Children(func() {
				AutoComplete(c, &g.mail, []Option{{Value: "admin@corp.cn"}, {Value: "audit@corp.cn"}, {Value: "ops@corp.cn"}}, AutoCompleteProps{Width: 260, AllowClear: true})
			})
		})
	}, 320, 220)
	ac.SetLocale("zh-CN")
	ac.Frame()
	ac.ClickAt(100, 36)
	ac.Frame()
	ac.Key(0, ui.KeyDown)
	ac.Frame()
	shot(t, ac, "mygo-kit-gallery-autocomplete.png")

	dates := ui.NewTester(g.dates, 1440, 1480)
	dates.SetLocale("zh-CN")
	dates.Frame()
	dates.Frame()
	shot(t, dates, "mygo-kit-gallery-date-modes.png")

	rp := ui.NewTester(g.rangeDisable, 900, 520)
	rp.SetLocale("zh-CN")
	rp.Frame()
	rp.Frame()
	if r, ok := rp.Find("开始日期 → 结束日期"); ok {
		rp.Move(r.X+r.W-12, r.Y+r.H/2)
		rp.Frame()
	}
	for _, want := range []string{"禁用过去", "AllowClear", "2026年11月", "2026年12月"} {
		if !rp.HasText(want) {
			t.Fatalf("rangepicker gallery missing %q in %v", want, rp.Texts())
		}
	}
	shot(t, rp, "mygo-kit-rangepicker.png")

	dark := newGalleryEntry()
	dark.tok = WorkDark
	darkPage := ui.NewTester(dark.entry, 1440, 1320)
	darkPage.SetLocale("zh-CN")
	darkPage.SetDark(true)
	darkPage.Frame()
	darkPage.Frame()
	shot(t, darkPage, "mygo-kit-gallery-dark-entry-input.png")
}
