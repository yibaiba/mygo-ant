// Gallery of kit components. From the module root:
//
//	go run ./examples/gallery
package main

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/yibaiba/mygo-ant"
)

func main() {
	mygo.App.SetName("kit-gallery")
	mygo.App.WhenReady(func() {
		d := newDemo()
		minW, minH := kit.WindowMinSize()
		mygo.NewWindow(mygo.WindowOptions{
			Title:     "kit gallery",
			Width:     1280,
			Height:    800,
			MinWidth:  minW,
			MinHeight: minH,
			Content:   ui.View(d.view),
		})
	})
	if err := mygo.App.Run(); err != nil {
		panic(err)
	}
}

type demo struct {
	nav string

	name, pass, search, area string
	on, check                bool
	num                      string
	sel, seg, radio          string
	page                     int
	tab                      string
	dialog, drawer           bool
}

func newDemo() *demo {
	return &demo{
		nav: "button", on: true, check: true, num: "3", sel: "a",
		seg: "day", radio: "ok", page: 2, tab: "one", name: "Demo",
	}
}

var demoNav = []kit.NavItem{
	{ID: "button", Label: "按钮"},
	{ID: "input", Label: "输入"},
	{ID: "choice", Label: "选择"},
	{ID: "form", Label: "表单"},
	{ID: "data", Label: "数据"},
	{ID: "feedback", Label: "反馈"},
	{ID: "nav", Label: "导航"},
	{ID: "overlay", Label: "弹层"},
	{ID: "home", Label: "首页壳"},
}

func (d *demo) view(c *ui.Context) {
	kit.Page(c, kit.Work, func() {
		ui.Row(c).FillWidth().Height(40).Background(kit.Work.Brand).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "kit gallery").FontSize(16).FontWeight(600).TextColor(kit.Work.TextOnPrimary).Margin(0, 0, 0, 16)
			ui.Spacer(c)
			kit.WindowButtons(c, kit.WindowButtonsProps{Light: true, Width: 48})
		})
		ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
			kit.NavList(c, &d.nav, demoNav)
			ui.Column(c).Grow(1).MinWidth(0).Fill().Padding(16, 20).Gap(12).Clip().Children(func() {
				switch d.nav {
				case "input":
					d.pageInput(c)
				case "choice":
					d.pageChoice(c)
				case "form":
					d.pageForm(c)
				case "data":
					d.pageData(c)
				case "feedback":
					d.pageFeedback(c)
				case "nav":
					d.pageNav(c)
				case "overlay":
					d.pageOverlay(c)
				case "home":
					d.pageHome(c)
				default:
					d.pageButton(c)
				}
			})
		})
	})
}

func section(c *ui.Context, title string, fn func()) {
	kit.SectionTitle(c, title)
	kit.Sheet(c, fn)
}

func labeled(c *ui.Context, label string, fn func()) {
	ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Center).Wrap().GapY(8).Children(func() {
		ui.Text(c, label).FontSize(12).TextColor(kit.Of(c).TextMuted).Width(88).Shrink(0)
		fn()
	})
}

func (d *demo) pageButton(c *ui.Context) {
	section(c, "Button", func() {
		labeled(c, "类型", func() {
			kit.Button(c, "Primary", kit.ButtonProps{Type: kit.Primary})
			kit.Button(c, "Default")
			kit.Button(c, "文本", kit.ButtonProps{Type: kit.TextButton})
			kit.Button(c, "链接", kit.ButtonProps{Type: kit.Link})
			kit.Button(c, "危险", kit.ButtonProps{Type: kit.Primary, Danger: true})
			kit.Button(c, "虚线", kit.ButtonProps{Type: kit.Dashed})
		})
		labeled(c, "状态", func() {
			kit.Button(c, "禁用", kit.ButtonProps{Type: kit.Primary, Disabled: true})
			kit.Button(c, "加载", kit.ButtonProps{Type: kit.Primary, Loading: true})
			kit.Button(c, "幽灵", kit.ButtonProps{Type: kit.Primary, Ghost: true})
			kit.Button(c, "圆角", kit.ButtonProps{Type: kit.Primary, Shape: kit.ShapeRound})
		})
		labeled(c, "尺寸", func() {
			kit.Button(c, "小", kit.ButtonProps{Size: kit.Small})
			kit.Button(c, "中", kit.ButtonProps{Type: kit.Primary})
			kit.Button(c, "大", kit.ButtonProps{Type: kit.Primary, Size: kit.Large})
		})
	})
}

func (d *demo) pageInput(c *ui.Context) {
	section(c, "Input / InputNumber / TextArea / Search", func() {
		labeled(c, "输入", func() {
			kit.Input(c, &d.name, kit.InputProps{Placeholder: "名称", AllowClear: true, ShowCount: true, MaxLength: 20})
		})
		labeled(c, "密码", func() {
			kit.Input(c, &d.pass, kit.InputProps{Password: true, Placeholder: "密码"})
		})
		labeled(c, "数字", func() {
			kit.InputNumber(c, &d.num, kit.InputNumberProps{Min: 0, Max: 10, Step: 1})
		})
		labeled(c, "搜索", func() {
			kit.Search(c, &d.search, kit.SearchProps{Placeholder: "搜索"})
		})
		labeled(c, "多行", func() {
			kit.TextArea(c, &d.area, kit.TextAreaProps{Rows: 3, Placeholder: "备注"})
		})
	})
}

func (d *demo) pageChoice(c *ui.Context) {
	opts := []kit.Option{{Value: "a", Label: "甲"}, {Value: "b", Label: "乙"}, {Value: "c", Label: "丙"}}
	section(c, "Select / Switch / Checkbox / Radio / Segmented", func() {
		labeled(c, "下拉", func() {
			kit.Select(c, &d.sel, opts, kit.SelectProps{ShowSearch: true, AllowClear: true})
		})
		labeled(c, "开关", func() {
			kit.Switch(c, &d.on, kit.ChoiceProps{CheckedChildren: "开", UnCheckedChildren: "关"})
			kit.Switch(c, &d.on)
		})
		labeled(c, "复选", func() {
			kit.Checkbox(c, &d.check, "同意条款")
		})
		labeled(c, "单选", func() {
			kit.RadioButtons(c, &d.radio, []kit.Option{{Value: "ok", Label: "正常"}, {Value: "warn", Label: "告警"}})
		})
		labeled(c, "分段", func() {
			kit.Segmented(c, &d.seg, []kit.Option{{Value: "day", Label: "日"}, {Value: "week", Label: "周"}, {Value: "month", Label: "月"}})
		})
	})
}

func (d *demo) pageForm(c *ui.Context) {
	f := &kit.Form{}
	f.Begin()
	err := f.Check("name", d.name, kit.Required("请输入名称"))
	section(c, "Form / FormItem", func() {
		kit.FormItem(c, "名称", func() {
			kit.Input(c, &d.name, kit.InputProps{Status: kit.ErrorStatus(err)})
		}, kit.FormItemProps{Required: true, Error: err})
		kit.FormItem(c, "开关", func() {
			kit.Switch(c, &d.on)
		})
		if kit.Button(c, "校验", kit.ButtonProps{Type: kit.Primary}) {
			_ = f.Validate(c)
		}
	})
}

func (d *demo) pageData(c *ui.Context) {
	section(c, "Tag / Badge / Empty / Pagination / Table", func() {
		labeled(c, "标签", func() {
			kit.Tag(c, "默认")
			kit.Tag(c, "成功", kit.TagProps{Tone: kit.TagSuccess})
			kit.Tag(c, "处理中", kit.TagProps{Tone: kit.TagBlue})
		})
		labeled(c, "徽标", func() {
			kit.BadgeCount(c, nil, kit.BadgeProps{Count: 12})
			kit.StatusDot(c, "success", "在线")
		})
		kit.Empty(c, "暂无数据")
		kit.Pagination(c, &d.page, 86, 10)
		rows := []struct{ name, st string }{{"Alpha", "运行"}, {"Beta", "停止"}}
		kit.Table(c, []kit.Column{
			{Title: "名称", Width: 160},
			{Title: "状态", Width: 120},
		}, len(rows), func(r, col int) kit.Cell {
			if col == 0 {
				return kit.Cell{Text: rows[r].name}
			}
			return kit.Cell{Text: rows[r].st}
		})
	})
}

func (d *demo) pageFeedback(c *ui.Context) {
	section(c, "Alert / Progress / Spin / Result", func() {
		kit.Alert(c, "提示", kit.AlertProps{Type: kit.Info, Description: "这是一条说明。"})
		kit.Alert(c, "成功", kit.AlertProps{Type: kit.Success})
		kit.Alert(c, "警告", kit.AlertProps{Type: kit.Warning})
		kit.Progress(c, 64, kit.ProgressProps{})
		kit.Spin(c, kit.SpinProps{Tip: "加载中"})
	})
}

func (d *demo) pageNav(c *ui.Context) {
	section(c, "Tabs / Steps", func() {
		kit.Tabs(c, &d.tab, []kit.Option{{Value: "one", Label: "其一"}, {Value: "two", Label: "其二"}})
		kit.Steps(c, 1, []string{"填写", "确认", "完成"})
	})
}

func (d *demo) pageOverlay(c *ui.Context) {
	section(c, "Dialog / Drawer / Popconfirm / Tooltip", func() {
		labeled(c, "对话框", func() {
			if kit.Button(c, "打开 Dialog") {
				d.dialog = true
			}
			if kit.Button(c, "打开 Drawer") {
				d.drawer = true
			}
			kit.Popconfirm(c, "确认删除？", func() bool {
				return kit.Button(c, "删除", kit.ButtonProps{Danger: true})
			}, kit.PopconfirmProps{OK: "删除", Danger: true})
		})
		kit.Dialog(c, &d.dialog, "示例对话框", func() {
			kit.Muted(c, "Dialog 内容。")
		})
		kit.Drawer(c, &d.drawer, "示例抽屉", func() {
			kit.Muted(c, "Drawer 内容。")
		})
	})
}

func (d *demo) pageHome(c *ui.Context) {
	kit.HomeHeader(c, "服务管理")
	ui.Row(c).FillWidth().Gap(16).Wrap().Children(func() {
		kit.ServiceCard(c, kit.GlassCardProps{
			Title: "示例服务", Subtitle: "HTTP API", Status: "正常", StatusOK: true, Extra: "8080", On: &d.on,
		})
		kit.ServiceCard(c, kit.GlassCardProps{
			Title: "缓存服务", Subtitle: "Redis", Status: "已停止", StatusOK: false, Extra: "6379",
		})
	})
}
