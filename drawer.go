package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// DrawerPlacement is the window edge a drawer slides in from.
type DrawerPlacement uint8

// Drawer placements.
const (
	// DrawerRight is the default.
	DrawerRight DrawerPlacement = iota
	DrawerLeft
	DrawerTop
	DrawerBottom
)

func (d DrawerPlacement) across() bool { return d >= DrawerTop }

// DrawerProps are a drawer's options.
type DrawerProps struct {
	Placement DrawerPlacement
	// Width is a left or right drawer's, Height a top or bottom one's;
	// 378 when zero (736 when Large), as Ant's, and never more than the
	// window's.
	Width, Height float32
	Large         bool
	// Pushed moves the drawer 180 back toward its edge, as Ant's push, for
	// a drawer with another drawer open over it.
	Pushed bool
	// Extra builds controls at the end of the header.
	Extra func()
	// Footer builds the buttons at the bottom; nil leaves none.
	Footer func()
}

// drawerMotion is how long a drawer takes to slide in or out.
const drawerMotion = 240 * time.Millisecond

// Drawer is Ant's Drawer: a panel sliding in from a window edge over a
// dimmed window, with a close button and title, the scrolling body fn
// builds and an optional footer. Clicking the mask, the close button or
// pressing Escape sets *open to false; it then slides out.
func Drawer(c *ui.Context, open *bool, title string, fn func(), props ...DrawerProps) {
	host := motionHost(c)
	show, v := presence(host, "drawer", *open, motionFor(c, drawerMotion))
	if !show {
		return
	}
	p := first(props)
	t := Of(c)
	winW, winH := c.Size()
	size := pick[float32](p.Large, 736, 378)
	if p.Placement.across() && p.Height > 0 {
		size = p.Height
	} else if !p.Placement.across() && p.Width > 0 {
		size = p.Width
	}
	size = min(size, pick(p.Placement.across(), winH, winW))
	push := host.AnimateTo("push", pick[float32](p.Pushed, 180, 0), ui.Animation{Duration: motionFor(c, drawerMotion), Ease: ui.EaseOut})
	off := (1-v)*size + push
	shown := true
	ui.DialogBase(c, &shown, func(backdrop, panel ui.Element) {
		backdrop.Background(t.Mask.Alpha(v)).AlignItems(ui.Stretch)
		panel.Column().Background(t.Elevated)
		switch p.Placement {
		case DrawerLeft:
			backdrop.Row().Justify(ui.Start)
			panel.Width(size).FillHeight().Translate(-off, 0).
				Shadow(6, 0, 16, 0, ui.RGBA(0, 0, 0, 0.08)).Shadow(3, 0, 6, -4, ui.RGBA(0, 0, 0, 0.12))
		case DrawerTop:
			backdrop.Column().Justify(ui.Start)
			panel.FillWidth().Height(size).Translate(0, -off).
				Shadow(0, 6, 16, 0, ui.RGBA(0, 0, 0, 0.08)).Shadow(0, 3, 6, -4, ui.RGBA(0, 0, 0, 0.12))
		case DrawerBottom:
			backdrop.Column().Justify(ui.End)
			panel.FillWidth().Height(size).Translate(0, off).
				Shadow(0, -6, 16, 0, ui.RGBA(0, 0, 0, 0.08)).Shadow(0, -3, 6, -4, ui.RGBA(0, 0, 0, 0.12))
		default:
			backdrop.Row().Justify(ui.End)
			panel.Width(size).FillHeight().Translate(off, 0).
				Shadow(-6, 0, 16, 0, ui.RGBA(0, 0, 0, 0.08)).Shadow(-3, 0, 6, -4, ui.RGBA(0, 0, 0, 0.12))
		}
		panel.Children(func() {
			ui.Row(c).FillWidth().Shrink(0).Padding(16, 24).Gap(t.SpaceMD).AlignItems(ui.Center).
				Draw(bottomRule(t.BorderSecondary)).Children(func() {
				if closeButton(c, t, 16) {
					*open = false
				}
				ui.Text(c, title).FontWeight(600).FontSize(t.FontSizeLG).TextColor(t.Text).Grow(1).MinWidth(0).SingleLine()
				if p.Extra != nil {
					ui.Row(c).Shrink(0).Gap(t.SpaceSM).AlignItems(ui.Center).Children(p.Extra)
				}
			})
			ui.Scroll(c).FillWidth().Grow(1).MinHeight(0).Children(func() {
				ui.Column(c).FillWidth().Padding(24).Gap(t.SpaceMD).Children(fn)
			})
			if p.Footer != nil {
				ui.Row(c).FillWidth().Shrink(0).Padding(8, 16).Gap(t.SpaceSM).Justify(ui.End).
					Draw(topRule(t.BorderSecondary)).Children(p.Footer)
			}
		})
	})
	if !shown {
		*open = false
	}
}

// closeButton is the small × of a dialog or drawer header, icon px wide;
// it reports a click.
func closeButton(c *ui.Context, t *Tokens, icon float32) bool {
	x := ui.ButtonBase(c).Label(L(c).Close).Size(22, 22).Radius(4).Center().FocusRing(false).Cursor(ui.CursorPointer)
	if x.Hovered() {
		x.Background(t.FillSecondary)
	}
	if x.FocusVisible() {
		focusRing(x, t)
	}
	x.Children(func() {
		ui.Icon(c, iconClose).Size(icon, icon).TextColor(t.TextTertiary)
	})
	return x.Clicked()
}
