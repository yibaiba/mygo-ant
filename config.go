package kit

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// ConfigProps are what a ConfigProvider sets for the components under it,
// as Ant's ConfigProvider props.
type ConfigProps struct {
	// Theme swaps the tokens; the ones in effect when nil.
	Theme *Tokens
	// Size is the controls' size where they leave SizeInherit. Zero
	// inherits the outer provider (Middle at the root). Middle is an
	// explicit 32px size, so a nested provider can restore it under Large.
	Size Size
	// Disabled, when set, disables (true) or enables (false) every control
	// under it (componentDisabled). Nil inherits the outer provider.
	Disabled *bool
	// Locale is the language of the components' own words; the one in
	// effect when nil.
	Locale *Locale
	// Form is the layout of the FormItems under it; FormView sets it.
	Form FormConfig
	// FollowDesktop mixes the OS HighContrast and TextScale into the
	// tokens. Off by default: Ant ConfigProvider colors stay as they are.
	// It does not switch Dark; pass Dark from DesktopDark for OS
	// light/dark (the product window does this each frame).
	FollowDesktop bool
	// Dark applies Ant's darkAlgorithm to the tokens (theme.darkAlgorithm).
	// Nil inherits the outer provider; Bool(true) darkens a light theme.
	// Bool(false) does not lighten a theme that is already dark.
	Dark *bool
	// Motion, when Bool(false), is Ant token.motion=false: kit enter and
	// leave jump. Nil inherits. OS ReduceMotion is already honoured by
	// ui.Animate even when this stays true.
	Motion *bool
}

// Bool is a *bool for ConfigProps.Disabled so nested providers can inherit
// when they leave it nil.
func Bool(v bool) *bool {
	b := v
	return &b
}

// configKey names one derived set of tokens.
type configKey struct {
	base      *Tokens
	size      Size
	disabled  bool
	locale    *Locale
	form      FormConfig
	motionOff bool
}

var configs = map[configKey]*Tokens{}

// Configured returns base with the provider's size, disabled state and
// locale, the same *Tokens for the same settings, with a ui.Theme of its
// own so that Of finds it while ui.Themed scopes it.
func Configured(base *Tokens, p ConfigProps) *Tokens {
	if base.base != nil {
		base = base.base
	}
	if p.Dark != nil && *p.Dark {
		base = DarkAlgorithm(base)
	}
	if p.Locale == nil {
		p.Locale = base.Locale
	}
	size := p.Size
	if size == SizeInherit {
		size = Middle
	}
	disabled := p.Disabled != nil && *p.Disabled
	motionOff := p.Motion != nil && !*p.Motion
	k := configKey{base, size, disabled, p.Locale, p.Form, motionOff}
	baseSize := base.Size
	if baseSize == SizeInherit {
		baseSize = Middle
	}
	if k.size == baseSize && !k.disabled && k.locale == base.Locale && k.form == (FormConfig{}) && !k.motionOff {
		return base
	}
	registryMu.Lock()
	t, ok := configs[k]
	registryMu.Unlock()
	if ok {
		return t
	}
	th := *base.Theme()
	cp := *base
	cp.Size, cp.Disabled, cp.Locale, cp.Form, cp.MotionOff = size, disabled, p.Locale, p.Form, motionOff
	cp.base, cp.theme = base, &th
	if p.Locale != nil && p.Locale.Lang != "" {
		cp.Lang = p.Locale.Lang
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if t, ok := configs[k]; ok {
		return t
	}
	configs[k] = &cp
	registry[&th] = &cp
	return &cp
}

// ConfigProvider is Ant's ConfigProvider: what fn builds, overlays and
// popups included, reads its theme, size, disabled state and locale.
// Providers nest; an inner one keeps Size, Disabled, Locale, Form, Dark
// and Motion it leaves unset from the outer one. Pass Size: Middle to
// restore 32px under a Large parent; pass Disabled: Bool(false) to
// re-enable. Dark: Bool(true) is theme.darkAlgorithm; Motion: Bool(false)
// is token.motion=false. FollowDesktop is off unless set: Ant tokens
// stay the default.
func ConfigProvider(c *ui.Context, props ConfigProps, fn func()) ui.Element {
	outer := Of(c)
	base := props.Theme
	if base == nil {
		base = outer
	}
	if props.Size == SizeInherit {
		props.Size = outer.Size
	}
	if props.Disabled == nil {
		props.Disabled = Bool(outer.Disabled)
	}
	if props.Locale == nil {
		props.Locale = outer.Locale
	}
	if props.Form == (FormConfig{}) {
		props.Form = outer.Form
	}
	if props.Dark == nil {
		props.Dark = Bool(outer.Dark)
	}
	if props.Motion == nil {
		props.Motion = Bool(!outer.MotionOff)
	}
	t := Configured(base, props)
	if props.FollowDesktop {
		t = FollowDesktop(c, t)
	}
	// Themed fills itself with the theme's page color; a provider has no
	// look of its own.
	return ui.Themed(c, t.Theme()).Background(ui.Color{}).FillWidth().Lang(t.Lang).Children(fn)
}

// Locale is the language of the words the components show themselves:
// buttons, placeholders, empty states and accessibility labels.
type Locale struct {
	Lang string

	OK, Cancel, Close, Confirm, Reset, Search  string
	Copy, Copied, Edit, Expand, Collapse, More string
	Loading, Empty, NoMatch                    string
	Prev, Next, Finish                         string

	// Picker words.
	SelectDate, SelectWeek, SelectMonth, SelectQuarter, SelectYear, SelectTime string
	StartDate, EndDate, StartTime, EndTime, Now, Today                         string
	PrevYear, NextYear, PrevMonth, NextMonth                                   string
	Hour, Minute, Second, AM, PM                                               string
	// Weekdays are Monday to Sunday, as the calendars start the week.
	Weekdays [7]string
	// YearTitle, MonthTitle and YearMonth name a calendar's header and
	// its cells; Week a week as the week picker shows it.
	YearTitle  func(y int) string
	MonthTitle func(m int) string
	YearMonth  func(y, m int) string
	Week       func(y, w int) string

	// Pagination.
	Total              func(n int) string
	PrevPage, NextPage string
	Page               func(n string) string
	JumpPrev, JumpNext string
	JumpTo, PageUnit   string
	PerPage            string

	// Window chrome.
	Minimize, Maximize string

	// Table, Transfer and filters.
	Filter, FilterSearch, FilterEmpty string
	ExpandRow, CollapseRow            string
	Items                             func(n int) string
	SearchHere, MoveRight, MoveLeft   string

	// Upload.
	UploadDrop, UploadRelease, Remove string
	Upload, Uploading, PreviewFile    string

	// Form.
	Optional, Add string

	// Tour, Image and QRCode.
	TourFinish, Preview          string
	QRExpired, QRRefresh, QRScan string
	QRTooLong                    string

	// Typography, Tabs, FloatButton.
	BackTop, AddTab, MoreTabs string
}

// LocaleZhCN is the products's language, Ant's zh_CN.
var LocaleZhCN = &Locale{
	Lang: "zh-CN",
	OK:   "确定", Cancel: "取消", Close: "关闭", Confirm: "确定", Reset: "重置", Search: "搜索",
	Copy: "复制", Copied: "复制成功", Edit: "编辑", Expand: "展开", Collapse: "收起", More: "更多",
	Loading: "加载中", Empty: "暂无数据", NoMatch: "无匹配结果",
	Prev: "上一步", Next: "下一步", Finish: "完成",

	SelectDate: "请选择日期", SelectWeek: "请选择周", SelectMonth: "请选择月份", SelectQuarter: "请选择季度", SelectYear: "请选择年份", SelectTime: "请选择时间",
	StartDate: "开始日期", EndDate: "结束日期", StartTime: "开始时间", EndTime: "结束时间", Now: "此刻", Today: "今天",
	PrevYear: "上一年", NextYear: "下一年", PrevMonth: "上个月", NextMonth: "下个月",
	Hour: "时", Minute: "分", Second: "秒", AM: "上午", PM: "下午",
	Weekdays:   [7]string{"一", "二", "三", "四", "五", "六", "日"},
	YearTitle:  func(y int) string { return strconv.Itoa(y) + "年" },
	MonthTitle: func(m int) string { return strconv.Itoa(m) + "月" },
	YearMonth:  func(y, m int) string { return strconv.Itoa(y) + "年" + strconv.Itoa(m) + "月" },
	Week:       func(y, w int) string { return strconv.Itoa(y) + "-" + pad2(w) + "周" },

	Total:    func(n int) string { return "共 " + strconv.Itoa(n) + " 条" },
	PrevPage: "上一页", NextPage: "下一页",
	Page:     func(n string) string { return "第 " + n + " 页" },
	JumpPrev: "向前 5 页", JumpNext: "向后 5 页",
	JumpTo: "跳至", PageUnit: "页", PerPage: "条/页",
	Minimize: "最小化", Maximize: "最大化",

	Filter: "筛选", FilterSearch: "在筛选项中搜索", FilterEmpty: "暂无筛选项",
	ExpandRow: "展开行", CollapseRow: "收起行",
	Items:      func(n int) string { return strconv.Itoa(n) + " 项" },
	SearchHere: "请输入搜索内容", MoveRight: "移到右侧", MoveLeft: "移到左侧",

	UploadDrop: "点击或拖拽文件到此区域", UploadRelease: "松开即可添加", Remove: "移除",
	Upload: "上传", Uploading: "上传中", PreviewFile: "预览文件",
	Optional: "（可选）", Add: "添加",

	TourFinish: "结束导览", Preview: "预览",
	QRExpired: "二维码过期", QRRefresh: "点击刷新", QRScan: "已扫描", QRTooLong: "内容过长",

	BackTop: "回到顶部", AddTab: "添加标签页", MoreTabs: "更多标签页",
}

// LocaleEnUS is Ant's en_US.
var LocaleEnUS = &Locale{
	Lang: "en-US",
	OK:   "OK", Cancel: "Cancel", Close: "Close", Confirm: "OK", Reset: "Reset", Search: "Search",
	Copy: "Copy", Copied: "Copied", Edit: "Edit", Expand: "Expand", Collapse: "Collapse", More: "More",
	Loading: "Loading", Empty: "No data", NoMatch: "Not found",
	Prev: "Previous", Next: "Next", Finish: "Finish",

	SelectDate: "Select date", SelectWeek: "Select week", SelectMonth: "Select month", SelectQuarter: "Select quarter", SelectYear: "Select year", SelectTime: "Select time",
	StartDate: "Start date", EndDate: "End date", StartTime: "Start time", EndTime: "End time", Now: "Now", Today: "Today",
	PrevYear: "Last year", NextYear: "Next year", PrevMonth: "Previous month", NextMonth: "Next month",
	Hour: "h", Minute: "m", Second: "s", AM: "AM", PM: "PM",
	Weekdays:   [7]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"},
	YearTitle:  func(y int) string { return strconv.Itoa(y) },
	MonthTitle: func(m int) string { return time.Month(m).String()[:3] },
	YearMonth:  func(y, m int) string { return time.Month(m).String()[:3] + " " + strconv.Itoa(y) },
	Week:       func(y, w int) string { return strconv.Itoa(y) + "-" + pad2(w) + "th" },

	Total:    func(n int) string { return "Total " + strconv.Itoa(n) + " items" },
	PrevPage: "Previous Page", NextPage: "Next Page",
	Page:     func(n string) string { return "Page " + n },
	JumpPrev: "Previous 5 Pages", JumpNext: "Next 5 Pages",
	JumpTo: "Go to", PageUnit: "", PerPage: "/ page",
	Minimize: "Minimize", Maximize: "Maximize",

	Filter: "Filter", FilterSearch: "Search in filters", FilterEmpty: "No filters",
	ExpandRow: "Expand row", CollapseRow: "Collapse row",
	Items: func(n int) string {
		return strconv.Itoa(n) + pick(n == 1, " item", " items")
	},
	SearchHere: "Search here", MoveRight: "Move right", MoveLeft: "Move left",

	UploadDrop: "Click or drag file to this area", UploadRelease: "Release to add", Remove: "Remove",
	Upload: "Upload", Uploading: "Uploading", PreviewFile: "Preview file",
	Optional: "(optional)", Add: "Add",

	TourFinish: "Finish", Preview: "Preview",
	QRExpired: "QR code expired", QRRefresh: "Refresh", QRScan: "Scanned", QRTooLong: "Content too long",

	BackTop: "Back to top", AddTab: "Add tab", MoreTabs: "More tabs",
}

func pad2(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// L returns the locale in effect: the ConfigProvider's, else LocaleZhCN.
func L(c *ui.Context) *Locale {
	if l := Of(c).Locale; l != nil {
		return l
	}
	return LocaleZhCN
}
