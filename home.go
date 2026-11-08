package kit

import "github.com/egoist/mygo/ui"

// HomeHeaderProps are the home page section head's options.
type HomeHeaderProps struct {
	// Extra is typically the HealthPill, after the title.
	Extra func()
	// Actions are optional nav buttons (设置 / 安装 / 连接) after Extra.
	Actions func()
	// TitleSize is the h1 size; 28 when zero. HomeTitleSize(w) matches
	// the default 24 / 28 / 32 steps.
	TitleSize float32
}

// HomeHeader is home layout: section-head: the 服务管理 title, Extra
// (health pill, amber when not ok) and optional Actions. It returns the
// row so messages can sit below it.
func HomeHeader(c *ui.Context, title string, props ...HomeHeaderProps) ui.Element {
	p := first(props)
	t := Of(c)
	size := pick(p.TitleSize == 0, float32(28), p.TitleSize)
	return ui.Row(c).FillWidth().Wrap().GapX(16).GapY(10).Padding(2, 0, 12).AlignItems(ui.Center).Children(func() {
		ui.Text(c, title).FontWeight(600).FontSize(size).TextColor(t.Text).Grow(1).MinWidth(160)
		if p.Extra != nil {
			p.Extra()
		}
		if p.Actions != nil {
			ui.Row(c).Shrink(0).Gap(12).AlignItems(ui.Center).Wrap().GapY(4).Children(p.Actions)
		}
	})
}

// HomeBarProps are the home window top bar's options (home layout: topbar).
type HomeBarProps struct {
	Logo            ui.ImageSource
	Title, Subtitle string
	// Brand, when set, replaces Logo / Title / Subtitle.
	Brand func()
	// Nav builds the TopNav cluster (连接 / 检测 / 启动 / 停止 / 设置 / 安装).
	Nav func()
	// Extra is the trailing cluster, typically window min / max / hide.
	Extra func()
	// Compact stacks brand + Extra above Nav (windows under 960).
	Compact bool
	// Gap between nav items; HomeNavGap(Width) when zero.
	Gap, Width float32
}

// HomeBar is home layout: topbar: app brand, centered nav, Extra at the
// end. Window min / max / close stay in Extra; there is no Ant window
// chrome, and HomeBar is not replaced with a fake browser bar. The Ant
// counterpart of the in-page cluster is LayoutHeader + Button. It returns
// the bar so messages can sit below it.
func HomeBar(c *ui.Context, props HomeBarProps) ui.Element {
	t := Of(c)
	gap := props.Gap
	if gap == 0 {
		gap = HomeNavGap(props.Width)
	}
	brand := props.Brand
	if brand == nil {
		brand = func() {
			ui.Row(c).Gap(10).Shrink(0).AlignItems(ui.Center).Children(func() {
				if props.Logo != nil {
					ui.Image(c, props.Logo).Size(32, 32).Fit(ui.Contain)
				}
				ui.Column(c).Gap(2).Children(func() {
					if props.Title != "" {
						ui.Text(c, props.Title).FontWeight(600).FontSize(16).TextColor(t.Text)
					}
					if props.Subtitle != "" {
						ui.Text(c, props.Subtitle).FontSize(12).TextColor(t.TextMuted)
					}
				})
			})
		}
	}
	nav := func() {
		if props.Nav == nil {
			return
		}
		row := ui.Row(c).Gap(gap).Center().Wrap()
		if props.Compact {
			row.FillWidth()
		} else {
			row.Grow(1)
		}
		row.Children(props.Nav)
	}
	if props.Compact {
		pl, pr := chromePad(c, 12, 12)
		return ui.Column(c).FillWidth().Padding(10, pr, 4, pl).Gap(4).DragWindow().Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
				brand()
				ui.Spacer(c)
				if props.Extra != nil {
					props.Extra()
				}
			})
			nav()
		})
	}
	pl, pr := chromePad(c, 24, 18)
	return ui.Row(c).FillWidth().Height(max(72, TitleBarInsets(c).Height)).Padding(0, pr, 0, pl).DragWindow().Children(func() {
		ui.Row(c).Width(240).Children(brand)
		nav()
		ui.Row(c).Width(240).Justify(ui.End).Children(func() {
			if props.Extra != nil {
				props.Extra()
			}
		})
	})
}

// GlassCardProps are a home glass card's options (home layout: service-card).
type GlassCardProps struct {
	Icon     ui.ImageSource
	IconFn   func()
	Title    string
	Subtitle string
	// Status is the badge label: 正常 / 已停止 / 已关闭.
	Status   string
	StatusOK bool
	// Extra is the right-side meta, typically the port.
	Extra string
	// On is the switch; nil hides it.
	On *bool
	// Disabled greys the switch (bulk start / stop).
	Disabled bool
	// Actions are the footer row, typically 日志.
	Actions func()
	// Loading shows the service-card skeleton instead of the body.
	Loading bool
}

// GlassCard is home layout: service-card: glass chrome, not an Ant
// Card. The Ant counterpart is Card. A Glass with icon, switch, a
// single-line title, two-line subtitle, status badge, extra and footer
// actions, all clamped to the card. It reports a change of the switch.
func GlassCard(c *ui.Context, props GlassCardProps) bool {
	if props.Loading {
		glassCardSkeleton(c)
		return false
	}
	t := Of(c)
	changed := false
	Glass(c, func() {
		ui.Row(c).FillWidth().Children(func() {
			switch {
			case props.Icon != nil:
				ui.Image(c, props.Icon).Size(64, 64).Fit(ui.Contain)
			case props.IconFn != nil:
				props.IconFn()
			}
			ui.Spacer(c)
			if props.On != nil {
				changed = Switch(c, props.On, ChoiceProps{Disabled: props.Disabled})
			}
		})
		ui.Text(c, props.Title).FontWeight(600).FontSize(16).TextColor(t.Glass.Title).
			SingleLine().MinWidth(0).FillWidth().Margin(16, 0, 4)
		ui.Text(c, props.Subtitle).FontSize(13).TextColor(t.Glass.Subtle).MaxLines(2).MinWidth(0).FillWidth()
		ui.Row(c).FillWidth().AlignItems(ui.Center).Margin(18, 0, 0).Children(func() {
			if props.Status != "" {
				Badge(c, props.Status, props.StatusOK)
			}
			ui.Spacer(c)
			if props.Extra != "" {
				ui.Text(c, props.Extra).FontSize(12).TextColor(t.Glass.Subtle).SingleLine().Shrink(0)
			}
		})
		if props.Actions != nil {
			ui.Row(c).FillWidth().Padding(10, 0, 0).Justify(ui.End).Children(props.Actions)
		}
	})
	return changed
}

// ServiceCard is GlassCard under the home layout: service-card.
func ServiceCard(c *ui.Context, props GlassCardProps) bool {
	return GlassCard(c, props)
}

// glassCardSkeleton stands in for a service card while the list loads.
func glassCardSkeleton(c *ui.Context) {
	Glass(c, func() {
		ui.Row(c).FillWidth().Children(func() {
			SkeletonBlock(c, 64, 64, SkeletonBlockProps{Radius: 14, Active: true})
			ui.Spacer(c)
			SkeletonBlock(c, 44, 22, SkeletonBlockProps{Round: true, Active: true})
		})
		ui.Column(c).FillWidth().Gap(8).Margin(16, 0, 0).Children(func() {
			SkeletonBlock(c, 120, 18, SkeletonBlockProps{Active: true})
			SkeletonBlock(c, 180, 14, SkeletonBlockProps{Active: true})
		})
		ui.Row(c).FillWidth().AlignItems(ui.Center).Margin(18, 0, 0).Children(func() {
			SkeletonBlock(c, 56, 22, SkeletonBlockProps{Round: true, Active: true})
			ui.Spacer(c)
			SkeletonBlock(c, 40, 12, SkeletonBlockProps{Active: true})
		})
		ui.Row(c).FillWidth().Padding(10, 0, 0).Justify(ui.End).Children(func() {
			SkeletonBlock(c, 48, 16, SkeletonBlockProps{Active: true})
		})
	})
}
