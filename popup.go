package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Trigger is what opens a popup, as Ant's trigger prop.
type Trigger uint8

// Popup triggers.
const (
	// TriggerClick opens and closes it on a click of the trigger.
	TriggerClick Trigger = iota
	// TriggerHover opens it while the pointer rests on the trigger or the
	// popup, closing a moment after it leaves both.
	TriggerHover
	// TriggerFocus opens it while the keyboard focus is in the trigger.
	TriggerFocus
	// TriggerContextMenu opens it at the pointer on a right-click.
	TriggerContextMenu
)

// PopupProps place a popup and say what opens it; Popover, Popconfirm and
// Dropdown take them.
type PopupProps struct {
	// Placement is the side of the trigger it opens on; Align lines it up
	// along that side. Near a window edge it flips to the other side.
	Placement Placement
	Align     PopupAlign
	// Arrow draws a small arrow pointing at the trigger.
	Arrow bool
	// Gap keeps the popup apart from the trigger; 4 when zero (2 with an
	// arrow).
	Gap     float32
	Trigger Trigger
	// Open, when set, holds whether it is open, for an app that opens or
	// closes it itself (Ant's controlled open); the popup still writes it
	// as the trigger asks.
	Open *bool
	// Tone colors the arrow; the popup background by default.
	Tone ui.Color

	open bool
	// width, when set, gets the trigger's width, for a menu at least as
	// wide as its control.
	width *float32
	// keep, when set, is whether something built in the panel, as a
	// submenu's popup, is open; a hover popup stays open while it is.
	keep *bool
	// stretch fills the trigger's line across, as a menu bar's items.
	stretch bool
}

// popupHold is what a popup keeps between frames.
type popupHold struct {
	lastHover  time.Time
	panelHover bool
	panelFocus bool
	at         ui.Point
}

// hoverGrace is how long a hover popup stays after the pointer leaves.
const hoverGrace = 150 * time.Millisecond

// Popup is the shared floating panel of the kit's popovers: trigger
// builds the control and returns its click, and fn styles the panel and
// builds its content while it is open; its close closes it. It returns whether the popup is
// open; Escape or a press outside closes it.
//
// Ant getPopupContainer defaults to document.body; here the panel is
// ui.Overlay (the window layer) attached to the trigger with AttachTo.
// ui has one overlay root and no CSS z-index, so a custom container is
// not faked: later overlays paint on top.
func Popup(c *ui.Context, trigger func() bool, fn func(panel ui.Element, close func()), props ...PopupProps) bool {
	p := first(props)
	t := Of(c)
	anchor := ui.Row(c).Shrink(0).AlignSelf(pick(p.stretch, ui.Stretch, ui.Start))
	open := p.Open
	if open == nil {
		open = ui.Local(anchor, "open", func() bool { return p.open })
	}
	hold := ui.Local(anchor, "hold", func() popupHold { return popupHold{} })
	target := anchor
	anchor.Children(func() {
		clicked := trigger()
		if p.Trigger == TriggerClick && clicked {
			*open = !*open
		}
		if p.Trigger == TriggerContextMenu {
			if x, y, over := anchor.PointerPosition(); over && anchor.RightClicked() {
				hold.at = ui.Point{X: x, Y: y}
				*open = true
			}
			target = ui.Box(c).Absolute().Left(hold.at.X).Top(hold.at.Y).Size(1, 1).PassThrough()
		}
	})
	switch p.Trigger {
	case TriggerHover:
		now := c.Now()
		if anchor.Hovered() || hold.panelHover || p.open {
			hold.lastHover = now
			*open = true
		} else if *open {
			if now.Sub(hold.lastHover) >= hoverGrace {
				*open = false
			} else {
				c.After(hoverGrace - now.Sub(hold.lastHover))
			}
		}
	case TriggerFocus:
		*open = anchor.FocusWithin() || hold.panelFocus
	}
	tone := p.Tone
	if tone == (ui.Color{}) {
		tone = t.Elevated
	}
	gap := p.Gap
	if gap == 0 {
		gap = pick[float32](p.Arrow, 2, 4)
	}
	place := panelPlace{Side: p.Placement, Align: p.Align, Arrow: p.Arrow, Gap: gap}
	if p.Trigger == TriggerContextMenu {
		place = panelPlace{Side: PlaceBottom, Align: AlignFirst}
	}
	if p.width != nil {
		*p.width = anchor.Bounds().W
	}
	hold.panelHover, hold.panelFocus = false, false
	ui.PopoverBase(c, target, open, func(wrap ui.Element) {
		hold.panelHover = wrap.Hovered()
		hold.panelFocus = wrap.FocusWithin()
		fade := enterIn(wrap, "in", motionFor(c, 120*time.Millisecond))
		wrap.Opacity(fade)
		placePanel(c, wrap, target, place, tone, func(panel ui.Element) {
			fn(panel, func() { *open, hold.panelHover = false, false })
		})
	})
	if p.keep != nil && *p.keep && *open {
		hold.panelHover = true
	}
	return *open
}

// PopoverProps are a popover's options.
type PopoverProps struct {
	PopupProps
	// Title heads the card, bold, above the content.
	Title string
}

// Popover is Ant's Popover: a white card with an arrow beside trigger
// holding Title and what content builds, opened as Trigger says (a click
// by default; Ant's default hover is TriggerHover). It returns whether it
// is open.
func Popover(c *ui.Context, content func(), trigger func() bool, props ...PopoverProps) bool {
	p := first(props)
	t := Of(c)
	pp := p.PopupProps
	pp.Arrow = true
	return Popup(c, trigger, func(panel ui.Element, _ func()) {
		popupLook(panel.Padding(12).Gap(t.SpaceSM).MaxWidth(360), t)
		panel.Children(func() {
			if p.Title != "" {
				ui.Text(c, p.Title).FontWeight(600).FontSize(t.FontSize).TextColor(t.TextHeading)
			}
			content()
		})
	}, pp)
}
