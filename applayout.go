package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// LayoutProps are a Layout's options.
type LayoutProps struct {
	// HasSider lays the children out in a row, a Sider beside the rest,
	// as Ant's hasSider; otherwise they stack.
	HasSider bool
}

// Layout is Ant's Layout: a box on colorBgLayout holding a LayoutHeader,
// Sider, LayoutContent and LayoutFooter, or Layouts nested in them. It
// fills what it is placed in.
func Layout(c *ui.Context, fn func(), props ...LayoutProps) ui.Element {
	p := first(props)
	t := Of(c)
	e := pick(p.HasSider, ui.Row(c), ui.Column(c)).Grow(1).MinHeight(0).MinWidth(0).
		AlignItems(ui.Stretch).Background(t.Layout.Body)
	return e.Children(fn)
}

// HeaderProps are a LayoutHeader's options.
type HeaderProps struct {
	// Light makes it white (with a bottom rule) instead of Ant's dark
	// #001529 header.
	Light bool
}

// LayoutHeader is Ant's Layout.Header: a 64 high bar, padded 50 at the
// sides, dark unless Light, its children in a centered row.
func LayoutHeader(c *ui.Context, fn func(), props ...HeaderProps) ui.Element {
	p := first(props)
	t := Of(c)
	e := ui.Row(c).FillWidth().Height(t.Layout.HeaderHeight).Padding(0, 50).Shrink(0).
		AlignItems(ui.Center).Gap(t.SpaceLG)
	if p.Light {
		e.Background(t.Surface).Draw(bottomRule(t.BorderSecondary))
	} else {
		e.Background(t.Layout.Header).TextColor(ui.Hex("#FFFFFF"))
	}
	return e.Children(fn)
}

// LayoutContent is Ant's Layout.Content: what is left of the layout,
// down or across.
func LayoutContent(c *ui.Context, fn func()) ui.Element {
	return ui.Column(c).Grow(1).Basis(0).MinHeight(0).MinWidth(0).Children(fn)
}

// LayoutFooter is Ant's Layout.Footer: padded 24 by 50 on the layout's
// background, in the body text.
func LayoutFooter(c *ui.Context, fn func()) ui.Element {
	t := Of(c)
	return ui.Row(c).FillWidth().Padding(24, 50).Shrink(0).AlignItems(ui.Center).
		Background(t.Layout.Body).TextColor(t.Text).Children(fn)
}

// SiderProps are a Sider's options.
type SiderProps struct {
	// Width is its width open, 200 when zero; CollapsedWidth collapsed,
	// 80 when zero. ZeroWidth collapses it to nothing, with a tab beside
	// it that opens it again (Ant's collapsedWidth={0}).
	Width, CollapsedWidth float32
	ZeroWidth             bool
	// Collapsed holds whether it is collapsed; Collapsible draws the
	// trigger bar under it that toggles that, unless NoTrigger.
	Collapsed   *bool
	Collapsible bool
	NoTrigger   bool
	// Light makes it white with a right rule instead of Ant's dark sider.
	Light bool
	// ReverseArrow points the trigger's arrow the other way, for a sider
	// on the right.
	ReverseArrow bool
	// Breakpoint, when set, collapses it as the window narrows below that
	// breakpoint and opens it as it widens again (Ant's breakpoint and
	// onBreakpoint).
	Breakpoint *Screen
}

// siderBroken reports whether a window at screen s is below the
// breakpoint bp.
func siderBroken(bp *Screen, s Screen) bool { return bp != nil && s < *bp }

// Sider is Ant's Layout.Sider: a column, Width wide, that fn fills
// (told whether it is collapsed, as a Menu wants), sliding to
// CollapsedWidth as it collapses. It reports a toggle by its trigger or
// breakpoint.
func Sider(c *ui.Context, fn func(collapsed bool), props ...SiderProps) bool {
	p := first(props)
	t := Of(c)
	width := pick(p.Width == 0, 200, p.Width)
	thin := pick(p.CollapsedWidth == 0, 80, p.CollapsedWidth)
	if p.ZeroWidth {
		thin = 0
	}
	e := ui.Column(c).FillHeight().Shrink(0)
	collapsed := p.Collapsed
	if collapsed == nil {
		collapsed = ui.Local(e, "collapsed", func() bool { return false })
	}
	changed := false
	if p.Breakpoint != nil {
		broken := siderBroken(p.Breakpoint, CurrentScreen(c))
		was := ui.Local(e, "broken", func() bool { return broken })
		if *was != broken {
			*was, *collapsed, changed = broken, broken, true
		}
	}
	w := e.AnimateTo("w", pick(*collapsed, thin, width), ui.Animation{Duration: motionFor(c, 200*time.Millisecond), Ease: ui.EaseOut})
	bg := pick(p.Light, t.Surface, t.Layout.Sider)
	fg := pick(p.Light, t.Text, ui.Hex("#FFFFFF"))
	e.Width(w).Background(bg)
	if p.Light {
		e.Draw(rightRule(t.BorderSecondary))
	}
	trigger := p.Collapsible && !p.NoTrigger
	e.Children(func() {
		ui.Column(c).Grow(1).Basis(0).MinHeight(0).ClipX().Children(func() { fn(*collapsed) })
		if !trigger {
			return
		}
		left := *collapsed == p.ReverseArrow
		if p.ZeroWidth {
			tab := ui.ButtonBase(c).Label(pick(*collapsed, "展开侧边栏", "收起侧边栏")).FocusRing(false).
				Absolute().Top(t.Layout.HeaderHeight).Right(-40).Size(40, 40).Center().
				Radius(0, 6, 6, 0).Background(bg).Cursor(ui.CursorPointer)
			tab.Children(func() { IconNamed(c, pick(*collapsed, "bars", "left"), IconProps{Color: fg}) })
			if tab.Clicked() {
				*collapsed, changed = !*collapsed, true
			}
			return
		}
		bar := ui.ButtonBase(c).Label(pick(*collapsed, "展开侧边栏", "收起侧边栏")).FocusRing(false).
			FillWidth().Height(t.Layout.TriggerHeight).Shrink(0).Center().Radius(0).Cursor(ui.CursorPointer).
			Background(pick(p.Light, t.Surface, t.Layout.Trigger))
		if p.Light {
			bar.Draw(topRule(t.BorderSecondary))
		}
		bar.Children(func() { IconNamed(c, pick(left, "left", "right"), IconProps{Color: fg}) })
		if bar.Clicked() {
			*collapsed, changed = !*collapsed, true
		}
	})
	return changed
}
