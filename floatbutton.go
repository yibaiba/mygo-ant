package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// FloatButtonProps are a floating button's options.
type FloatButtonProps struct {
	// Icon is its mark, Ant's file icon when nil and there is no
	// Description.
	Icon *ui.SVG
	// Description is text under the icon, on a Square button.
	Description string
	// Tooltip shows beside it on hover.
	Tooltip string
	// Primary fills it with the primary color; Square rounds it to a
	// square instead of a circle.
	Primary, Square bool
	// Badge, when set, badges it, as Ant's badge prop.
	Badge *BadgeProps
	// Right and Bottom place it from the window's corner; 24 and 48 when
	// zero.
	Right, Bottom float32

	force visual
}

func floatCorner(p FloatButtonProps) (float32, float32) {
	return pick(p.Right == 0, 24, p.Right), pick(p.Bottom == 0, 48, p.Bottom)
}

// FloatButton is Ant's FloatButton, floating over the window at its
// bottom right, labelled label; it reports a click.
func FloatButton(c *ui.Context, label string, props ...FloatButtonProps) bool {
	p := first(props)
	right, bottom := floatCorner(p)
	clicked := false
	ui.Overlay(c, func() {
		ui.Column(c).Absolute().Right(right).Bottom(bottom).Children(func() {
			clicked = floatButtonBody(c, label, p, 0)
		})
	})
	return clicked
}

// floatJoin is where a square button sits in a joined group: 0 alone,
// 1 first, 2 between, 3 last.
type floatJoin uint8

// floatButtonBody is one floating button in place.
func floatButtonBody(c *ui.Context, label string, p FloatButtonProps, join floatJoin) bool {
	t := Of(c)
	clicked := false
	build := func() {
		btn := ui.ButtonBase(c).Label(label).FocusRing(false).Size(40, 40).Column().Center().Gap(2).
			Cursor(ui.CursorPointer)
		st := stateOf(btn, false, p.force)
		bg, fg := t.Elevated, t.Text
		if p.Primary {
			bg, fg = t.Primary, t.TextOnPrimary
			if st == hover || st == focus {
				bg = t.PrimaryHover
			} else if st == active {
				bg = t.PrimaryActive
			}
		} else if st == hover || st == focus || st == active {
			bg = t.FillTertiary.Over(t.Elevated)
		}
		btn.Background(bg)
		switch {
		case !p.Square:
			btn.Radius(20)
		case join == 1:
			btn.Radius(8, 8, 0, 0)
		case join == 2:
			btn.Radius(0)
		case join == 3:
			btn.Radius(0, 0, 8, 8)
		default:
			btn.Radius(8)
		}
		if join == 0 {
			shadowSecondary(btn)
		}
		if p.Description != "" && p.Square {
			btn.Height(pick[float32](p.Icon != nil, 56, 40)).Width(40)
		}
		if st == focus {
			focusRing(btn, t)
		}
		btn.Children(func() {
			switch {
			case p.Icon != nil:
				ui.Icon(c, p.Icon).Size(18, 18).TextColor(fg)
			case p.Description == "":
				IconNamed(c, "file", IconProps{Size: 18, Color: fg})
			}
			if p.Description != "" {
				ui.Text(c, p.Description).FontSize(t.FontSizeSM).TextColor(fg).SingleLine()
			}
		})
		TooltipOn(c, btn, p.Tooltip, TooltipProps{Placement: PlaceLeft})
		clicked = btn.Clicked()
	}
	if p.Badge != nil {
		b := *p.Badge
		if b.Offset == ([2]float32{}) && !p.Square {
			b.Offset = [2]float32{-6, 6}
		}
		BadgeCount(c, build, b)
	} else {
		build()
	}
	return clicked
}

// FloatItem is one button of a FloatButtonGroup.
type FloatItem struct {
	Key, Label string
	FloatButtonProps
}

// FloatGroupProps are a floating button group's options.
type FloatGroupProps struct {
	// Menu folds the buttons behind a trigger button, opened by a click
	// (or the pointer resting on it with TriggerHover); Open, when set,
	// holds whether it is open.
	Menu    bool
	Trigger Trigger
	Open    *bool
	// Icon, Primary and Tooltip are the trigger's.
	Icon    *ui.SVG
	Primary bool
	Tooltip string
	// Square joins square buttons into one bar; otherwise they float 16
	// apart.
	Square        bool
	Right, Bottom float32

	open bool
}

// FloatButtonGroup is Ant's FloatButton.Group: items stacked over the
// window's bottom right, joined into a bar when Square, or folded behind
// a trigger with Menu. It returns the clicked item's key, or "".
func FloatButtonGroup(c *ui.Context, items []FloatItem, props ...FloatGroupProps) string {
	p := first(props)
	t := Of(c)
	right, bottom := floatCorner(FloatButtonProps{Right: p.Right, Bottom: p.Bottom})
	chosen := ""
	ui.Overlay(c, func() {
		col := ui.Column(c).Absolute().Right(right).Bottom(bottom).AlignItems(ui.Center)
		open := p.Open
		if open == nil {
			open = ui.Local(col, "open", func() bool { return p.open })
		}
		hovered := col.Hovered()
		if p.Menu && p.Trigger == TriggerHover {
			*open = hovered
		}
		col.Children(func() {
			buttons := func() {
				for i, it := range items {
					it.Square = it.Square || p.Square
					join := floatJoin(0)
					if p.Square && !p.Menu {
						join = pick(i == 0, floatJoin(1), pick(i == len(items)-1, floatJoin(3), floatJoin(2)))
					}
					if floatButtonBody(c.Key(it.Key), it.Label, it.FloatButtonProps, join) {
						chosen = it.Key
					}
					if join == 1 || join == 2 {
						ui.Box(c.Key(i)).Size(40, 1).Background(t.BorderSecondary)
					}
				}
			}
			stack := ui.Column(c).AlignItems(ui.Center).Gap(pick[float32](p.Square && !p.Menu, 0, 16))
			if p.Square && !p.Menu {
				shadowSecondary(stack.Radius(8).Background(t.Elevated))
			}
			if !p.Menu {
				stack.Children(buttons)
				return
			}
			show, v := presence(col, "menu", *open, motionFor(c, 200*time.Millisecond))
			if show {
				stack.MarginY(0).Padding(0, 0, 16, 0).Opacity(v).Translate(0, (1-v)*16).Children(buttons)
			}
			icon := p.Icon
			if *open {
				icon = LookupIcon("close")
			}
			if floatButtonBody(c.Key("trigger"), pick(*open, L(c).Collapse, L(c).Expand), FloatButtonProps{
				Icon: icon, Primary: p.Primary, Tooltip: p.Tooltip, Square: p.Square,
			}, 0) && p.Trigger != TriggerHover {
				*open = !*open
			}
		})
	})
	return chosen
}

// BackTopProps are a back-to-top button's options.
type BackTopProps struct {
	// VisibilityHeight is how far the content scrolls before it shows;
	// 400 when zero.
	VisibilityHeight float32
	// Duration is how long the scroll back takes; 450ms when zero.
	Duration      time.Duration
	Tooltip       string
	Right, Bottom float32
}

// BackTop is Ant's FloatButton.BackTop: a floating button that shows
// once scroll is past VisibilityHeight and scrolls it back to the top.
// It reports a click.
func BackTop(c *ui.Context, scroll *ui.ScrollState, props ...BackTopProps) bool {
	p := first(props)
	vh := pick(p.VisibilityHeight == 0, 400, p.VisibilityHeight)
	d := pick(p.Duration == 0, 450*time.Millisecond, p.Duration)
	right, bottom := floatCorner(FloatButtonProps{Right: p.Right, Bottom: p.Bottom})
	clicked := false
	host := motionHost(c)
	show, v := presence(host, "backtop", scroll.Y >= vh, motionFor(c, 200*time.Millisecond))
	if !show {
		return false
	}
	ui.Overlay(c, func() {
		ui.Column(c).Absolute().Right(right).Bottom(bottom).Opacity(v).Children(func() {
			clicked = floatButtonBody(c, L(c).BackTop, FloatButtonProps{Icon: LookupIcon("vertical-align-top"), Tooltip: p.Tooltip}, 0)
		})
	})
	if clicked {
		scroll.ScrollTo(scroll.X, 0, ui.Animation{Duration: d, Ease: ui.EaseOut})
	}
	return clicked
}
