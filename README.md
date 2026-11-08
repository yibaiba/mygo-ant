# mygo-ant

Ant Design v5 外观的 [MyGo](https://github.com/egoist/mygo) 原生组件库。包名 `kit`。只依赖 `mygo/ui` 和标准库，不碰业务状态。

默认用官方 `github.com/egoist/mygo` v0.8.3。LoongArch ABI 1.0（或需要 Linux GTK 的 DRI / 字体补丁）时，在**你的应用**里可选：

```
replace github.com/egoist/mygo => github.com/yibaiba/mygo v0.8.3-loong64.1
```

## 分层

```
应用窗口 (mygo.NewWindow + Content)
        │
        ▼
┌─────────────────────────────────────┐
│  4 壳层  HomeBar / NavList / WindowButtons / App  │
│  3 组合  Form / Table / Upload / Dialog / Menu     │
│  2 控件  Button / Input / Select / Switch / DatePicker │
│  1 主题  Tokens / Page / ConfigProvider / Of(c)    │
│  0 基件  mygo/ui 的 *Base（无外观，kit 负责画）      │
└─────────────────────────────────────┘
```

| 层 | 职责 | 你怎么用 |
| --- | --- | --- |
| 0 `ui` | 命中、焦点、IME、布局 | 不要直接当成品按钮用 |
| 1 主题 | 色、字号、圆角、语言 | `Page(c, kit.Work, fn)`，里面 `kit.Of(c)` |
| 2 控件 | 一个交互 | `if kit.Button(...) { }` / `kit.Input(c, &s)` |
| 3 组合 | 表单、表、弹层 | `FormItem` + `Validate`；`Table` 吃列定义 |
| 4 壳 | 顶栏、侧栏、窗钮 | 可选；没有 Ant 对等物的才放这里 |

约定（每一层都一样）：

- 签名：`Name(c, 必填..., ...NameProps)`，Props 零值即默认。
- 按钮 / 开关 / 输入：返回 `bool`（是否点击或值变了）。
- 容器：收 `func()` 画子节点。
- 带触发器的浮层（`Popconfirm`、`Dropdown`）：收 `func() bool`，里面放一个 kit 按钮并返回它的点击。

## 最小用法

```go
package main

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/yibaiba/mygo-ant"
)

func main() {
	app := mygo.New()
	app.OnReady(func() {
		w := app.NewWindow(mygo.WindowOptions{Title: "Demo", Width: 960, Height: 640})
		name := ""
		w.SetContent(func(c *ui.Context) {
			kit.Page(c, kit.Work, func() {
				kit.SectionTitle(c, "基础设置")
				kit.Sheet(c, func() {
					kit.FormItem(c, "名称", func() {
						kit.Input(c, &name, kit.InputProps{Placeholder: "请输入"})
					}, kit.FormItemProps{Required: true})
					if kit.Button(c, "保存", kit.ButtonProps{Type: kit.Primary}) {
						// 保存
					}
				})
			})
		})
	})
	app.Run()
}
```

`kit.Work` 白底工作页，`kit.Home` 玻璃首页，`kit.WorkDark` 暗色。整页包在 `Page` 里，子组件才能读到令牌。

## 1. 主题

```go
kit.Page(c, kit.Work, func() {
	t := kit.Of(c)          // 当前 Tokens
	_ = t.Primary           // #006adb
	kit.ConfigProvider(c, kit.ConfigProps{
		Size:     kit.Large,           // 控件 40px；嵌套里 SizeInherit 会继承
		Disabled: kit.Bool(false),     // nil 表示继承
		Dark:     kit.Bool(kit.DesktopDark(c)),
		Motion:   kit.Bool(true),
	}, func() {
		// 这里的 Input / Button 用 Large
	})
})
```

- `FollowDesktop` 默认关：不把系统高对比混进 Ant 色。
- 中文 Linux 会尝试注册文泉驿 / Droid Sans Fallback，字体栈仍是系统 UI 优先。

## 2. 控件

```go
var on bool
var n float64
var sel string

if kit.Button(c, "确定", kit.ButtonProps{Type: kit.Primary, Size: kit.Large}) {
	submit()
}
kit.Input(c, &name, kit.InputProps{AllowClear: true, ShowCount: true, MaxLength: 32})
kit.InputNumber(c, &n, kit.InputNumberProps{Min: 0, Max: 100, Step: 1})
kit.Select(c, &sel, []kit.Option{{Value: "a", Label: "甲"}, {Value: "b", Label: "乙"}}, kit.SelectProps{ShowSearch: true})
if kit.Switch(c, &on, kit.ChoiceProps{CheckedChildren: "开", UnCheckedChildren: "关"}) {
	// on 已翻转
}
```

尺寸：`Small` 24 / `Middle` 32（默认）/ `Large` 40。`SizeInherit` 跟 `ConfigProvider`。

## 3. 表单与校验

```go
f := &kit.Form{} // 或嵌在你的 struct 里
f.Begin()
err := f.Check("name", name, kit.Required("请输入名称"))
kit.FormItem(c, "名称", func() {
	kit.Input(c, &name, kit.InputProps{Status: kit.ErrorStatus(err)})
}, kit.FormItemProps{Required: true, Error: err})
if kit.Button(c, "提交", kit.ButtonProps{Type: kit.Primary}) && f.Validate(c) == "" {
	save()
}
```

`FormView` 可设 `Layout`: `FormVertical`（默认）/ `FormHorizontal` / `FormInline`。

## 4. 表、弹层、导航、壳

```go
// 表
kit.Table(c, rows, []kit.Column{
	{Title: "名称", DataIndex: "name"},
	{Title: "状态", DataIndex: "status"},
}, kit.TableProps{RowKey: "id"})

// 弹层：trigger 里放按钮并返回点击
kit.Popconfirm(c, func() bool {
	return kit.Button(c, "删除", kit.ButtonProps{Danger: true})
}, kit.PopconfirmProps{Title: "确认删除？", OK: "删除", Danger: true})

// 侧栏
sel := "base"
kit.NavList(c, &sel, []kit.NavItem{{Key: "base", Label: "基础"}, {Key: "net", Label: "网络"}})

// 无边框窗钮（把 mygo.Window 的 Minimize 等传进来）
kit.WindowButtons(c, kit.WindowButtonsProps{
	OnMinimize: w.Minimize, OnMaximize: w.ToggleMaximize, OnClose: w.Close,
})
```

首页玻璃卡：`HomeBar`、`HomeHeader`、`GlassCard` / `ServiceCard`。这是壳层，不是 Ant 的 Card。

## 5. 要应用注入的能力

kit 不 import `mygo` 主包，文件框和系统通知由应用注入：

```go
kit.SetFileDialog(func(opt kit.FileDialogOptions) ([]string, error) {
	return w.Dialog().Open(...) // 你的 mygo.Window
})
kit.SetSaveDialog(...)
kit.SetMessageBox(...)  // 可选，系统确认框
kit.SetOSNotify(...)    // 可选，系统通知
```

不注入时，Upload 仍然把点击交给 `OnPick` / 调用方。

## 测试

```go
tester := ui.NewTester(func(c *ui.Context) {
	kit.Page(c, kit.Work, func() {
		kit.Button(c, "保存", kit.ButtonProps{Type: kit.Primary})
	})
}, 800, 600)
tester.SetLocale("zh-CN")
tester.Frame()
if !tester.HasText("保存") { t.Fatal("missing") }
```

组件清单与 Ant 对齐情况见 [doc.go](doc.go)。MIT。
