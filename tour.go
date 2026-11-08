package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// TourStep is one step of a tour.
type TourStep struct {
	Title, Description string
	// Cover builds a picture above the title.
	Cover func()
	// Target returns the element the step points at, built earlier in
	// the frame; the step sits in the middle of the window without one.
	Target    func() ui.Element
	Placement Placement
	// NoArrow drops the arrow at the target.
	NoArrow bool
}

// TourProps are a tour's options.
type TourProps struct {
	// NoMask leaves the window undimmed around the target.
	NoMask bool
	// Primary fills the card with the primary color, as Ant's type
	// "primary".
	Primary bool
	// Gap is the room kept around the target in the mask's hole; 6 when
	// zero.
	Gap float32
	// OnFinish runs when the last step's button ends the tour.
	OnFinish func()
}

// tourHole is the mask's hole around target, gap wider each way.
func tourHole(target ui.Rect, gap float32) ui.Rect {
	return ui.Rect{X: target.X - gap, Y: target.Y - gap, W: target.W + gap*2, H: target.H + gap*2}
}

// tourMask is the four bands of a w×h window around hole, which keep the
// window from taking input outside it.
func tourMask(hole ui.Rect, w, h float32) [4]ui.Rect {
	return [4]ui.Rect{
		{X: 0, Y: 0, W: w, H: max(hole.Y, 0)},
		{X: 0, Y: hole.Y + hole.H, W: w, H: max(h-hole.Y-hole.H, 0)},
		{X: 0, Y: hole.Y, W: max(hole.X, 0), H: hole.H},
		{X: hole.X + hole.W, Y: hole.Y, W: max(w-hole.X-hole.W, 0), H: hole.H},
	}
}

// Tour is Ant's Tour while *open: each step's card beside its target
// (with an arrow), the window dimmed around a hole showing the target,
// dots for the steps and the previous / next buttons stepping *current.
// Escape or the close mark closes it. It reports the tour ending on its
// last step.
func Tour(c *ui.Context, open *bool, current *int, steps []TourStep, props ...TourProps) bool {
	p := first(props)
	host := motionHost(c)
	shown, in := presence(host, "tour", *open, motionFor(c, 150*time.Millisecond))
	if !shown || len(steps) == 0 {
		return false
	}
	*current = min(max(*current, 0), len(steps)-1)
	step := steps[*current]
	t := Of(c)
	gap := pick(p.Gap == 0, 6, p.Gap)
	finished := false
	bg, fg, sub := t.Elevated, t.TextHeading, t.Text
	if p.Primary {
		bg, fg, sub = t.Primary, t.TextOnPrimary, t.TextOnPrimary
	}
	ui.Overlay(c, func() {
		ww, wh := c.Size()
		var target ui.Element
		var hole ui.Rect
		if step.Target != nil {
			target = step.Target()
			hole = tourHole(target.Bounds(), gap)
		} else {
			hole = ui.Rect{X: ww / 2, Y: wh / 2}
		}
		if !p.NoMask {
			for i, r := range tourMask(hole, ww, wh) {
				ui.Box(c.Key(i)).Absolute().Left(r.X).Top(r.Y).Size(r.W, r.H).Background(t.Mask.Alpha(in))
			}
			if step.Target != nil {
				ui.Box(c).Absolute().Left(hole.X).Top(hole.Y).Size(hole.W, hole.H).Radius(2).PassThrough().
					Border(1, ui.RGBA(255, 255, 255, 0.4*in))
			}
		}
		anchor := ui.Box(c).Absolute().Left(hole.X).Top(hole.Y).Size(hole.W, hole.H).PassThrough()
		wrap := ui.Box(c).Opacity(in).Focusable().AutoFocus().Label(step.Title)
		if c.Shortcut(0, ui.KeyEscape) {
			*open = false
		}
		card := func(panel ui.Element) {
			panel.Width(min(520, ww-32)).Background(bg).Radius(8)
			shadowSecondary(panel)
			panel.Children(func() {
				if step.Cover != nil {
					ui.Column(c).FillWidth().Padding(16, 16, 0).Children(step.Cover)
				}
				ui.Row(c).FillWidth().Padding(16, 16, 8).AlignItems(ui.Start).Gap(8).Children(func() {
					ui.Text(c, step.Title).FontSize(t.FontSizeLG).FontWeight(600).TextColor(fg).Grow(1)
					x := ui.ButtonBase(c).Label(L(c).Close).Size(22, 22).Radius(4).Center().FocusRing(false).Cursor(ui.CursorPointer)
					if x.Hovered() {
						x.Background(pick(p.Primary, ui.RGBA(255, 255, 255, 0.15), t.FillSecondary))
					}
					x.Children(func() { ui.Icon(c, iconClose).Size(14, 14).TextColor(pick(p.Primary, fg, t.TextTertiary)) })
					if x.Clicked() {
						*open = false
					}
				})
				if step.Description != "" {
					ui.Text(c, step.Description).FontSize(t.FontSize).TextColor(sub).Padding(0, 16)
				}
				ui.Row(c).FillWidth().Padding(12, 16, 16).AlignItems(ui.Center).Children(func() {
					ui.Row(c).Gap(6).Grow(1).Children(func() {
						for i := range steps {
							on := i == *current
							dot := pick(p.Primary, ui.RGBA(255, 255, 255, pick[float32](on, 1, 0.15)), pick(on, t.Primary, t.FillSecondary))
							ui.Box(c.Key(i)).Size(6, 6).Radius(3).Background(dot)
						}
					})
					ui.Row(c).Gap(8).Children(func() {
						if *current > 0 && Button(c, L(c).Prev, ButtonProps{Size: Small, Type: pick(p.Primary, TextButton, Default)}) {
							*current--
						}
						last := *current == len(steps)-1
						if Button(c, pick(last, L(c).TourFinish, L(c).Next), ButtonProps{Size: Small, Type: pick(p.Primary, Default, Primary)}) {
							if last {
								*open = false
								finished = true
								if p.OnFinish != nil {
									p.OnFinish()
								}
							} else {
								*current++
							}
						}
					})
				})
			})
		}
		if step.Target == nil {
			wrap.Absolute().Left(0).Top(0).Size(ww, wh).Center().PassThrough().Children(func() { card(ui.Column(c)) })
			return
		}
		placePanel(c, wrap, anchor, panelPlace{Side: step.Placement, Arrow: !step.NoArrow, Gap: 4}, bg, card)
	})
	return finished
}
