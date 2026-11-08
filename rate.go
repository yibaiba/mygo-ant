package kit

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// RateProps are a rating's options.
type RateProps struct {
	// Count is how many stars; 5 when zero.
	Count int
	// AllowHalf lets the value end in a half star.
	AllowHalf bool
	// NoClear keeps a click on the current value from clearing it (Ant's
	// allowClear={false}).
	NoClear  bool
	Disabled bool
	// Character is drawn in place of the star, as "好" or "A".
	Character string
	// Size is a star's; 20 when zero.
	Size float32
	// Tooltips name the values 1, 2, … shown over the star hovered.
	Tooltips []string
}

// rateGap is the space between stars.
const rateGap = 8

// rateYellow is Ant Rate's star color (colorFillContent → #fadb14).
var rateYellow = ui.Hex("#fadb14")

// rateAt is the value a pointer x along stars size wide with rateGap
// between them picks: the star under it, or half of it when half and the
// pointer is on its first half. 0 before the first star.
func rateAt(x, size float32, count int, half bool) float64 {
	if x < 0 || size <= 0 {
		return 0
	}
	cell := size + rateGap
	i := int(x / cell)
	if i >= count {
		return float64(count)
	}
	in := x - float32(i)*cell
	if half && in < size/2 {
		return float64(i) + 0.5
	}
	return float64(i + 1)
}

// rateClick is the value after a click picking v over current: v, or 0
// when it is the current value and clearing is allowed.
func rateClick(current, v float64, clear bool) float64 {
	if clear && current == v {
		return 0
	}
	return v
}

// Rate is Ant's Rate: a row of stars editing *value, by click, hover
// preview and the arrow keys. It reports a change.
func Rate(c *ui.Context, value *float64, props ...RateProps) bool {
	p := first(props)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	n := pick(p.Count > 0, p.Count, 5)
	size := pick(p.Size > 0, p.Size, 20)
	step := pick(p.AllowHalf, 0.5, 1.0)
	changed := false
	w := float32(n)*(size+rateGap) - rateGap
	row := ui.Box(c).Size(w, size).Shrink(0).Focusable().FocusRing(false).Disabled(p.Disabled).
		Cursor(pick(p.Disabled, ui.CursorDefault, ui.CursorPointer)).Label("评分").Range(0, float64(n), *value)
	shown := *value
	x, _, over := row.PointerPosition()
	hoverV := 0.0
	if over && !p.Disabled {
		hoverV = rateAt(x, size, n, p.AllowHalf)
		shown = hoverV
	}
	if row.Clicked() && !p.Disabled && over {
		*value = rateClick(*value, hoverV, !p.NoClear)
		changed = true
	}
	if !p.Disabled {
		switch {
		case row.Shortcut(0, ui.KeyRight):
			*value = math.Min(float64(n), *value+step)
			changed = true
		case row.Shortcut(0, ui.KeyLeft):
			*value = math.Max(0, *value-step)
			changed = true
		}
	}
	if row.FocusVisible() {
		focusRing(row, t)
		row.Radius(4)
	}
	off := t.FillSecondary
	char := p.Character
	row.Draw(func(pa *ui.Painter, r ui.Rect) {
		for i := 0; i < n; i++ {
			cell := ui.Rect{X: r.X + float32(i)*(size+rateGap), Y: r.Y, W: size, H: size}
			fill := float32(math.Max(0, math.Min(1, shown-float64(i))))
			draw := func(tone ui.Color) {
				if char != "" {
					cw, ch := pa.MeasureText(0, ui.Span{Text: char, Size: size * 0.9})
					pa.Text(cell.X+(size-cw)/2, cell.Y+(size-ch)/2, char, size*0.9, tone)
					return
				}
				pa.FillPath(starPath(cell.X+size/2, cell.Y+size/2+size*0.04, size/2, size/2*0.45), tone)
			}
			draw(off)
			if fill > 0 {
				pa.Clip(ui.Rect{X: cell.X, Y: cell.Y, W: size * fill, H: size}, 0, func() { draw(rateYellow) })
			}
		}
	})
	if over && len(p.Tooltips) > 0 && !p.Disabled {
		i := int(math.Ceil(hoverV)) - 1
		if i >= 0 && i < len(p.Tooltips) {
			label := p.Tooltips[i]
			ui.Overlay(c, func() {
				tip := ui.Box(c).PassThrough()
				target := ui.Box(c).Absolute().PassThrough().Size(size, size).
					Left(row.Bounds().X + float32(i)*(size+rateGap)).Top(row.Bounds().Y)
				placePanel(c, tip, target, panelPlace{Side: PlaceTop, Arrow: true, Gap: 4}, t.TooltipBg, func(b ui.Element) {
					b.Background(t.TooltipBg).Radius(6).Padding(6, 8).Children(func() {
						ui.Text(c, label).FontSize(t.FontSize).TextColor(t.TooltipText).SingleLine()
					})
				})
			})
		}
	}
	return changed
}
