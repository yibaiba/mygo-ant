package kit

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

type galleryGlobal struct {
	tok    *Tokens
	page   int
	date   time.Time
	choice string
	text   string
	ns     Notifications
	modals Modals
}

func panel(c *ui.Context, tok *Tokens, title string, fn func()) {
	ui.Column(c).FillWidth().Padding(16).Gap(12).Background(tok.Surface).Radius(8).Border(1, tok.BorderSecondary).Children(func() {
		Title(c, title, TitleProps{Level: 5})
		fn()
	})
}

func (g *galleryGlobal) global(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 反馈与全局配置").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(16).Gap(16).Children(func() {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(16).Children(func() {
				panel(c, tok, "警告提示 Alert", func() {
					Alert(c, "证书将在 7 天后过期", AlertProps{Type: Warning, Closable: true, Action: func() {
						Button(c, "续期", ButtonProps{Size: Small, Type: Primary})
					}})
					Alert(c, "同步完成", AlertProps{Type: Success, Description: "共同步 128 条会议记录，用时 3.2 秒。", Closable: true})
					Alert(c, "网络不稳定", AlertProps{Type: Error, CloseText: "知道了"})
					Alert(c, "横幅模式带图标", AlertProps{Type: Info, Banner: true, ShowIcon: true})
					Alert(c, "自定义图标", AlertProps{Type: Info, Icon: LookupIcon("bell"), Action: func() {
						ui.Row(c).Gap(8).Children(func() {
							Button(c, "详情", ButtonProps{Size: Small, Type: TextButton})
							Button(c, "忽略", ButtonProps{Size: Small})
						})
					}})
				})
				panel(c, tok, "进度条 Progress", func() {
					ui.Row(c).FillWidth().Gap(24).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(10).Children(func() {
							Progress(c, 30)
							Progress(c, 50, ProgressProps{Status: ProgressActive})
							Progress(c, 70, ProgressProps{Error: true})
							Progress(c, 100)
							Progress(c, 60, ProgressProps{SuccessPercent: 30})
							Progress(c, 80, ProgressProps{From: ui.Hex("#108ee9"), To: ui.Hex("#87d068")})
							Progress(c, 45, ProgressProps{Cap: CapButt, Small: true})
							ui.Row(c).Gap(24).Children(func() {
								Progress(c, 50, ProgressProps{Steps: 5})
								Progress(c, 30, ProgressProps{Steps: 10, Small: true, Color: tok.SuccessIcon})
							})
						})
						ui.Column(c).Gap(16).Children(func() {
							ui.Row(c).Gap(16).Children(func() {
								Progress(c, 75, ProgressProps{Type: ProgressCircle})
								Progress(c, 70, ProgressProps{Type: ProgressCircle, Error: true})
								Progress(c, 100, ProgressProps{Type: ProgressCircle})
							})
							ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
								Progress(c, 90, ProgressProps{Type: ProgressCircle, From: ui.Hex("#108ee9"), To: ui.Hex("#87d068")})
								Progress(c, 50, ProgressProps{Type: ProgressCircle, Steps: 8, StrokeWidth: 12})
								Progress(c, 75, ProgressProps{Type: ProgressDashboard})
								Progress(c, 40, ProgressProps{Type: ProgressDashboard, Small: true, Cap: CapButt, SuccessPercent: 20})
							})
						})
					})
				})
				panel(c, tok, "水印 Watermark", func() {
					ui.Row(c).FillWidth().Gap(16).Children(func() {
						Watermark(c, func() {
							ui.Column(c).FillWidth().Height(150).Padding(16).Gap(8).Children(func() {
								ui.Text(c, "会议纪要（内部）").FontWeight(600).TextColor(tok.Text)
								Paragraph(c, "本次会议讨论了示例系统升级计划，确定了证书轮换时间与数据库迁移窗口。", TypoProps{Type: TypoSecondary})
							})
						}, WatermarkProps{Content: []string{"Confidential", "张三 0315"}, FontSize: 14, Gap: [2]float32{60, 50}}).Border(1, tok.BorderSecondary).Radius(6)
						Watermark(c, func() {
							ui.Column(c).FillWidth().Height(150).Padding(16).Children(func() {
								ui.Text(c, "图标水印").TextColor(tok.Text)
							})
						}, WatermarkProps{Icon: LookupIcon("safety"), Width: 32, Height: 32, Gap: [2]float32{60, 40}}).Border(1, tok.BorderSecondary).Radius(6)
					})
				})
			})
			ui.Column(c).Width(470).Gap(16).Children(func() {
				panel(c, tok, "通知提醒框 Notification", func() {
					Notification(c, Notice{Kind: NoticeSuccess, Title: "备份完成", Description: "数据库已备份到 /data/backup，可在设置中查看。",
						Btn: func(close func()) {
							Button(c, "查看", ButtonProps{Type: Primary, Size: Small})
						}})
					Notification(c, Notice{Title: "新版本可用", Description: "v2.4.0 已发布，重启后生效。", ShowProgress: true, Duration: 10 * time.Second})
					Notification(c, Notice{Kind: NoticeError, Title: "连接失败", Description: "无法连接 172.30.0.9:36678。", NoClose: true})
				})
				panel(c, tok, "全局提示 Message", func() {
					Message(c, "loading", "正在上传证书…")
					Message(c, "success", "保存成功")
				})
				panel(c, tok, "全局化配置 ConfigProvider", func() {
					ConfigProvider(c, ConfigProps{Size: Large, Locale: LocaleEnUS}, func() {
						ui.Column(c).FillWidth().Gap(8).Children(func() {
							Muted(c, "size=large · locale=en-US")
							ui.Row(c).Gap(8).Children(func() {
								Button(c, "Submit", ButtonProps{Type: Primary})
								DatePicker(c, &g.date, DatePickerProps{Width: 180})
							})
							Pagination(c, &g.page, 50, 10, PaginationProps{ShowTotal: true})
						})
					})
					ConfigProvider(c, ConfigProps{Size: Small, Disabled: Bool(true)}, func() {
						ui.Column(c).FillWidth().Gap(8).Children(func() {
							Muted(c, "size=small · componentDisabled · zh-CN")
							ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
								Button(c, "提交", ButtonProps{Type: Primary})
								Input(c, &g.text, InputProps{Placeholder: "已禁用", Width: 140})
								RangePicker(c, new(time.Time), new(time.Time), RangePickerProps{Width: 220})
							})
						})
					})
				})
			})
		})
	})
}

func (g *galleryGlobal) live(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 通知位置与堆叠").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Column(c).Fill().Padding(24).Children(func() {
			Muted(c, "右上：逐条展开；左上：超过 3 条折叠为堆叠；右下：带按钮与倒计时进度；底部居中。")
		})
		NotificationHolder(c, &g.ns, NotificationsProps{Top: 64})
		ModalHolder(c, &g.modals)
	})
}

func TestGalleryGlobalShots(t *testing.T) {
	g := &galleryGlobal{tok: Work, page: 2}
	tester := ui.NewTester(g.global, 1440, 1180)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	for _, want := range []string{"证书将在 7 天后过期", "Total 50 items", "75%", "备份完成", "Select date"} {
		if !tester.HasText(want) {
			t.Fatalf("global gallery missing %q", want)
		}
	}
	shot(t, tester, "mygo-kit-gallery-feedback-global.png")

	d := &galleryGlobal{tok: WorkDark, page: 2}
	dark := ui.NewTester(d.global, 1440, 1180)
	dark.SetLocale("zh-CN")
	dark.Frame()
	dark.Frame()
	shot(t, dark, "mygo-kit-gallery-dark-feedback-global.png")

	l := &galleryGlobal{tok: Work}
	live := ui.NewTester(l.live, 1440, 900)
	live.SetLocale("zh-CN")
	live.Frame()
	l.ns.Open(Notice{Kind: NoticeSuccess, Title: "备份完成", Description: "数据库已备份。", Duration: -1})
	l.ns.Open(Notice{Kind: NoticeInfo, Title: "同步中", Description: "正在同步会议资料。", Duration: -1})
	for i := range 5 {
		l.ns.Open(Notice{Kind: NoticeWarning, Title: "证书提醒 " + itoa(i+1), Description: "证书即将过期。", Placement: NoticeTopLeft, Duration: -1})
	}
	l.ns.Open(Notice{Kind: NoticeError, Title: "上传失败", Description: "网络中断，稍后重试。", Placement: NoticeBottomRight, ShowProgress: true, Duration: 20 * time.Second,
		Btn: func(close func()) {}})
	l.ns.Open(Notice{Title: "底部居中通知", Placement: NoticeBottom, Duration: -1})
	live.Move(700, 450)
	live.Frame()
	live.Advance(4 * time.Second)
	shot(t, live, "mygo-kit-gallery-feedback-notify.png")

	l.ns.Close("")
	l.modals.Info(ModalFunc{Title: "提示", Content: "这是一条信息。"})
	l.modals.Confirm(ModalFunc{Title: "删除证书？", Content: "删除后客户端将无法连接，此操作不可撤销。", OK: "删除", OKDanger: true, Closable: true})
	live.Frame()
	live.Advance(time.Second)
	shot(t, live, "mygo-kit-gallery-feedback-modal.png")
}
