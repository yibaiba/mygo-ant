package kit

import (
	"strconv"

	"github.com/egoist/mygo/ui"
)

// StepStatus is the state of one step, as Ant's status prop.
type StepStatus uint8

// Step statuses; StepAuto takes it from the current step.
const (
	StepAuto StepStatus = iota
	StepWait
	StepProcess
	StepFinish
	StepError
)

// StepItem is one step of a StepBar.
type StepItem struct {
	Title, SubTitle, Description string
	// Status overrides the one from the current step.
	Status StepStatus
	// Icon replaces the numbered circle.
	Icon     *ui.SVG
	Disabled bool
}

// StepsType is a StepBar's look.
type StepsType uint8

// Step bar looks.
const (
	// StepsDefault numbers the steps in circles joined by lines.
	StepsDefault StepsType = iota
	// StepsNavigation lines the steps up as tabs, arrows between them
	// and a bar under the current one.
	StepsNavigation
	// StepsInline is a small row of dots with the titles under them, to
	// sit inside a list item.
	StepsInline
)

// StepBarProps are a step bar's options.
type StepBarProps struct {
	Type StepsType
	// Vertical stacks the steps, a line down the side joining them.
	Vertical bool
	// Small shrinks the circles to 24 and the titles to 14.
	Small bool
	// Dot draws dots instead of circles, the titles under them.
	Dot bool
	// LabelVertical puts the titles under the circles.
	LabelVertical bool
	// Status is the current step's: StepProcess when StepAuto.
	Status StepStatus
	// Percent rings the current circle with a progress arc, 0 to 100,
	// when HasPercent.
	Percent    float32
	HasPercent bool
	// Clickable lets a click on a step make it current.
	Clickable bool
	// Initial numbers the first step; its index is still 0.
	Initial int
}

// stepStatus is step i's status with current the current step.
func stepStatus(i, current int, it StepItem, cur StepStatus) StepStatus {
	if it.Status != StepAuto {
		return it.Status
	}
	switch {
	case i < current:
		return StepFinish
	case i == current:
		return pick(cur == StepAuto, StepProcess, cur)
	}
	return StepWait
}

// stepColors are a step's circle fill, its number or icon, its title and
// its description.
func stepColors(t *Tokens, s StepStatus) (bg, mark, title, desc ui.Color) {
	switch s {
	case StepProcess:
		return t.Primary, t.TextOnPrimary, t.TextHeading, t.Text
	case StepFinish:
		return t.PrimaryBg, t.Primary, t.Text, t.TextTertiary
	case StepError:
		return t.ErrorBg, t.Error, t.Error, t.Error
	}
	return t.FillSecondary, t.TextTertiary, t.TextTertiary, t.TextTertiary
}

// stepMark is a step's circle (or dot, or icon).
func stepMark(c *ui.Context, t *Tokens, p StepBarProps, s StepStatus, it StepItem, n int, size float32) {
	bg, mark, _, _ := stepColors(t, s)
	switch {
	case p.Dot || p.Type == StepsInline:
		d := pick[float32](p.Type == StepsInline, 6, pick[float32](s == StepProcess, 10, 8))
		col := pick(s == StepWait, t.Fill, pick(s == StepError, t.Error, t.Primary))
		if p.Type == StepsInline && s == StepWait {
			col = t.Border
		}
		ui.Box(c).Size(10, 10).Center().Shrink(0).Children(func() {
			ui.Box(c).Size(d, d).Radius(d / 2).Background(col)
		})
	case it.Icon != nil:
		ui.Box(c).Size(size, size).Center().Shrink(0).Children(func() {
			ui.Icon(c, it.Icon).Size(size*0.75, size*0.75).TextColor(pick(s == StepWait, t.TextTertiary, pick(s == StepError, t.Error, t.Primary)))
		})
	default:
		circle := ui.Box(c).Size(size, size).Radius(size / 2).Background(bg).Center().Shrink(0)
		if p.HasPercent && s == StepProcess {
			pct, ring, track := p.Percent, t.Primary, t.FillSecondary
			circle.DrawOver(func(pa *ui.Painter, r ui.Rect) {
				cx, cy, rad := r.X+r.W/2, r.Y+r.H/2, r.W/2+4
				pa.StrokePath(arcPath(cx, cy, rad, 0, 360), 2, track)
				if pct > 0 {
					pa.StrokePath(arcPath(cx, cy, rad, 0, min(pct, 100)*3.6), 2, ring)
				}
			})
		}
		circle.Children(func() {
			switch s {
			case StepFinish:
				IconNamed(c, "check", IconProps{Size: size * 0.5, Color: mark})
			case StepError:
				IconNamed(c, "close", IconProps{Size: size * 0.5, Color: mark})
			default:
				ui.Text(c, strconv.Itoa(n)).FontSize(pick[float32](p.Small, 12, 16)).TextColor(mark)
			}
		})
	}
}

// StepBar is Ant's Steps: items numbered from Initial, those before
// *current finished, the current one in process (or Status) and the rest
// waiting, joined by lines that turn the primary color behind a finished
// step. With Clickable a click makes a step current, reported by its
// return. (Steps is the install guide's own bar.)
func StepBar(c *ui.Context, current *int, items []StepItem, props ...StepBarProps) bool {
	p := first(props)
	t := Of(c)
	size := pick[float32](p.Small, 24, 32)
	titleSize := pick(p.Small, t.FontSize, t.FontSizeLG)
	below := (p.LabelVertical || p.Dot || p.Type == StepsInline) && !p.Vertical
	changed := false
	box := pick(p.Vertical, ui.Column(c), ui.Row(c)).FillWidth().AlignItems(ui.Start).Label("步骤条")
	if p.Type == StepsNavigation {
		box.Draw(bottomRule(t.BorderSecondary))
	}
	box.Children(func() {
		for i, it := range items {
			kc := c.Key(i)
			s := stepStatus(i, *current, it, p.Status)
			_, _, titleFg, descFg := stepColors(t, s)
			last := i == len(items)-1
			tail := pick(s == StepFinish, t.Primary, t.Split)
			clickable := p.Clickable && !it.Disabled
			var step ui.Element
			if clickable {
				step = ui.ButtonBase(kc).Label(pick(it.Title == "", strconv.Itoa(i+1), it.Title)).FocusRing(false).Radius(0).
					Cursor(pick(i != *current, ui.CursorPointer, ui.CursorDefault))
				if i != *current && (step.Hovered() || step.FocusVisible()) {
					titleFg, descFg = t.Primary, t.Primary
				}
			} else {
				step = ui.Row(kc)
			}
			title := func() {
				ui.Row(kc).Gap(8).AlignItems(ui.Center).Shrink(0).Children(func() {
					ui.Text(kc, it.Title).FontSize(pick(p.Type == StepsInline, t.FontSizeSM, titleSize)).
						TextColor(titleFg).SingleLine()
					if it.SubTitle != "" && p.Type != StepsInline {
						ui.Text(kc, it.SubTitle).FontSize(t.FontSize).TextColor(t.TextTertiary).SingleLine()
					}
				})
			}
			desc := func() {
				if it.Description != "" && p.Type != StepsInline {
					ui.Text(kc, it.Description).FontSize(t.FontSize).TextColor(descFg).MaxWidth(pick[float32](p.Vertical, 1e4, 140))
				}
			}
			n := i + 1 + p.Initial
			switch {
			case p.Type == StepsNavigation:
				step.Grow(1).Basis(0).Padding(12, 16).Gap(8).AlignItems(ui.Start)
				if i == *current {
					ink := t.Primary
					step.DrawOver(func(pa *ui.Painter, r ui.Rect) {
						pa.Fill(ui.Rect{X: r.X, Y: r.Y + r.H - 3, W: r.W, H: 3}, ink, 0)
					})
				}
				step.Children(func() {
					stepMark(kc, t, p, s, it, n, size)
					ui.Column(kc).Gap(4).Grow(1).MinWidth(0).Children(func() { title(); desc() })
					if !last {
						IconNamed(kc, "right", IconProps{Size: 14, Color: t.TextTertiary}).AlignSelf(ui.Center)
					}
				})
			case p.Vertical:
				step.FillWidth().Gap(16).AlignItems(ui.Stretch)
				step.Children(func() {
					ui.Column(kc).AlignItems(ui.Center).Shrink(0).Width(size).Children(func() {
						stepMark(kc, t, p, s, it, n, size)
						if !last {
							ui.Box(kc).Width(1).Grow(1).MinHeight(24).Margin(4, 0).Background(tail)
						}
					})
					ui.Column(kc).Gap(4).Grow(1).MinWidth(0).Padding(pick[float32](p.Dot, 0, (size-titleSize*1.4)/2), 0, pick[float32](last, 0, 24), 0).
						Children(func() { title(); desc() })
				})
			case below:
				step.Grow(1).Basis(0).Column().AlignItems(ui.Center).Gap(8)
				prevTail := t.Split
				if i > 0 && stepStatus(i-1, *current, items[i-1], p.Status) == StepFinish {
					prevTail = t.Primary
				}
				step.Children(func() {
					ui.Row(kc).FillWidth().AlignItems(ui.Center).Children(func() {
						ui.Box(kc).Grow(1).Height(1).Background(pick(i == 0, ui.Color{}, prevTail))
						stepMark(kc, t, p, s, it, n, size)
						ui.Box(kc).Grow(1).Height(1).Background(pick(last, ui.Color{}, tail))
					})
					ui.Column(kc).AlignItems(ui.Center).Gap(4).Padding(0, 8).Children(func() {
						title()
						if it.Description != "" && p.Type != StepsInline {
							ui.Text(kc, it.Description).FontSize(t.FontSize).TextColor(descFg).MaxWidth(140).TextAlign(ui.Center)
						}
					})
				})
			default:
				step.Gap(8).AlignItems(ui.Start).Padding(0, pick[float32](last, 0, 16), 0, 0)
				if !last {
					step.Grow(1).Basis(0)
				}
				step.Children(func() {
					stepMark(kc, t, p, s, it, n, size)
					ui.Column(kc).Gap(4).Grow(1).MinWidth(0).Children(func() {
						ui.Row(kc).FillWidth().Height(size).AlignItems(ui.Center).Gap(16).Children(func() {
							title()
							if !last {
								ui.Box(kc).Grow(1).MinWidth(16).Height(1).Background(tail)
							}
						})
						desc()
					})
				})
			}
			if clickable && step.Clicked() && i != *current {
				*current = i
				changed = true
			}
		}
	})
	return changed
}
