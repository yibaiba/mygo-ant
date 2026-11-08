package kit

import (
	"strconv"

	"github.com/egoist/mygo/ui"
)

// NavItem is one entry of a NavList or NavChips.
type NavItem struct {
	ID, Label string
	Icon      *ui.SVG
}

// NavList is the settings sidebar for wide windows: 110×100 tiles, 20px
// gap, the chosen one filled whole-tile brand blue. It sets *selected and
// reports a change.
func NavList(c *ui.Context, selected *string, items []NavItem) bool {
	t := Of(c)
	changed := false
	ui.Column(c).Width(SidebarWidth).FillHeight().Background(t.Surface).
		BorderWidth(0, 1, 0, 0).BorderColor(t.NavBorder).Shrink(0).Clip().Children(func() {
		ui.Scroll(c).Fill().Children(func() {
			ui.Column(c).FillWidth().Padding(NavTilePad).Gap(NavTileGap).Children(func() {
				for _, item := range items {
					if navTile(c, t, item, *selected == item.ID) {
						*selected, changed = item.ID, true
					}
				}
			})
		})
	})
	return changed
}

// NavChips is NavList for narrow windows: a wrapping row of chips under
// the header.
func NavChips(c *ui.Context, selected *string, items []NavItem) bool {
	t := Of(c)
	changed := false
	ui.Row(c).FillWidth().Wrap().Gap(6).Padding(10, t.SpaceMD).Background(t.Surface).
		BorderWidth(0, 0, 1, 0).BorderColor(t.NavBorder).Children(func() {
		for _, item := range items {
			if navEntry(c, t, item, *selected == item.ID) {
				*selected, changed = item.ID, true
			}
		}
	})
	return changed
}

// navTile is one SettingTab .tab-item: 110×100, 24px icon, 16px name,
// #F2F5F7 rest and whole-tile #1E59D7 when chosen.
func navTile(c *ui.Context, t *Tokens, item NavItem, on bool) bool {
	btn := ui.ButtonBase(c).Label(item.Label).Width(NavTileW).Height(NavTileH).
		Radius(8).FocusRing(false).Cursor(ui.CursorPointer).Clip().Shrink(0)
	st := stateOf(btn, false, rest)
	bg, fg, ic := t.PageBg, t.Brand, t.Brand
	switch {
	case on:
		bg, fg, ic = t.Brand, t.TextOnPrimary, t.TextOnPrimary
	case st == hover || st == focus:
		bg = t.NavHover
		fg, ic = t.Brand, t.Brand
	case st == active:
		bg = t.NavPressed
	}
	btn.Background(bg)
	if st == focus && !on {
		focusRing(btn, t)
	}
	btn.Children(func() {
		ui.Column(c).Fill().Gap(8).AlignItems(ui.Center).Justify(ui.Center).Padding(0, 15).Children(func() {
			if item.Icon != nil {
				ui.Icon(c, item.Icon).Size(24, 24).TextColor(ic).Shrink(0)
			}
			ui.Text(c, item.Label).FontSize(t.FontSizeLG).TextColor(fg).MaxLines(2).FillWidth()
		})
	})
	return btn.Clicked()
}

func navEntry(c *ui.Context, t *Tokens, item NavItem, on bool) bool {
	btn := ui.ButtonBase(c).Label(item.Label).Radius(4).FocusRing(false).Cursor(ui.CursorPointer)
	btn.Height(t.ControlHeight).Padding(0, 10).Gap(6)
	st := stateOf(btn, false, rest)
	bg, fg, ic := t.PageBg, t.Brand, t.Brand
	switch {
	case on:
		bg, fg, ic = t.Brand, t.TextOnPrimary, t.TextOnPrimary
	case st == hover || st == focus:
		bg = t.NavHover
		fg, ic = t.Brand, t.Brand
	case st == active:
		bg = t.NavPressed
	}
	btn.Background(bg)
	if st == focus && !on {
		focusRing(btn, t)
	}
	btn.Children(func() {
		if item.Icon != nil {
			ui.Icon(c, item.Icon).Size(16, 16).TextColor(ic)
		}
		ui.Text(c, item.Label).FontSize(t.FontSize).TextColor(fg).SingleLine()
	})
	return btn.Clicked()
}

// StepsProps are a steps bar's options.
type StepsProps struct {
	// Compact shortens the connectors for narrow windows.
	Compact bool
}

// Steps is the install guide's step bar (InstallationGuide .steps-container):
// 32px numbered circles, the current one filled with the brand blue, done
// ones ringed with it, and 4px #E1E4E6 connectors.
func Steps(c *ui.Context, current int, labels []string, props ...StepsProps) {
	p := first(props)
	t := Of(c)
	ui.Row(c).FillWidth().Padding(t.SpaceXS, 27).Wrap().GapY(t.SpaceSM).AlignItems(ui.Center).Children(func() {
		for i, label := range labels {
			on, done := i == current, i < current
			ui.Row(c).Gap(t.SpaceSM).AlignItems(ui.Center).Shrink(0).Children(func() {
				bg, fg, edge := t.Surface, t.Text, t.BorderButton
				switch {
				case on:
					bg, fg, edge = t.Brand, t.TextOnPrimary, t.Brand
				case done:
					fg, edge = t.Brand, t.Brand
				}
				ui.Box(c).Size(32, 32).Radius(16).Background(bg).Border(1, edge).Center().Shrink(0).Children(func() {
					ui.Text(c, strconv.Itoa(i+1)).FontWeight(600).FontSize(16).TextColor(fg)
				})
				ui.Text(c, label).FontWeight(600).FontSize(t.FontSizeLG).
					TextColor(pick(on || done, t.Text, t.TextSecondary)).SingleLine()
			})
			if i < len(labels)-1 {
				line := ui.Box(c).Height(4).Radius(2).Background(t.Border).Margin(0, t.SpaceMD)
				if p.Compact {
					line.Width(16)
				} else {
					line.Grow(1).MinWidth(24)
				}
			}
		}
	})
}

// TopNavProps are a home header tab's options.
type TopNavProps struct {
	Icon     ui.ImageSource
	Current  bool
	Disabled bool
	// Large is the minimal-mode button (home layout: minimalist-mode .nav-item).
	Large bool
}

// TopNav is a home header tab (home layout: nav-item): icon and label, an
// underline when current, faded when disabled. It reports a click.
func TopNav(c *ui.Context, label string, props ...TopNavProps) bool {
	p := first(props)
	t := Of(c)
	btn := ui.ButtonBase(c).Label(label).Disabled(p.Disabled).FocusRing(false).Radius(0).Gap(t.SpaceSM).
		Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
	size := float32(15)
	if p.Large {
		btn.Height(48).Padding(0, 18).Radius(8)
		size = 16
	} else {
		btn.Height(40).Padding(0, 2, 8)
	}
	st := stateOf(btn, p.Disabled, rest)
	fg := t.TextSecondary
	switch {
	case p.Current:
		fg = t.Primary
	case st == hover || st == active || st == focus:
		fg = t.Text
	}
	if p.Large && (st == hover || st == focus) {
		btn.Background(t.PrimaryBg)
	}
	if p.Disabled {
		btn.Opacity(0.4)
	}
	if st == focus {
		btn.Radius(4)
		focusRing(btn, t)
	}
	if p.Current {
		line := t.Primary
		btn.DrawOver(func(pa *ui.Painter, r ui.Rect) {
			pa.Fill(ui.Rect{X: r.X, Y: r.Y + r.H - 3, W: r.W, H: 3}, line, 2)
		})
	}
	btn.Children(func() {
		if p.Icon != nil {
			ui.Image(c, p.Icon).Size(22, 22).Fit(ui.Contain)
		}
		ui.Text(c, label).FontSize(size).TextColor(fg).SingleLine()
	})
	return btn.Clicked() && !p.Disabled
}
