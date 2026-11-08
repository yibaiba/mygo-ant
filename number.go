package kit

import (
	"math"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// InputNumberProps are a number input's options.
type InputNumberProps struct {
	// Min and Max bound the value when Max > Min.
	Min, Max float64
	// Step is what the arrows add; zero means 1.
	Step float64
	// Precision is the number of decimals kept.
	Precision   int
	Size        Size
	Disabled    bool
	Placeholder string
	Status      Status
	Width       float32
	// AutoFocus gives the field the keyboard focus the frame it appears.
	AutoFocus bool

	force visual
}

func (p InputNumberProps) bounded() bool { return p.Max > p.Min }

func (p InputNumberProps) format(v float64) string {
	return strconv.FormatFloat(v, 'f', max(p.Precision, 0), 64)
}

func (p InputNumberProps) clamp(v float64) float64 {
	if p.bounded() {
		v = math.Min(math.Max(v, p.Min), p.Max)
	}
	scale := math.Pow(10, float64(max(p.Precision, 0)))
	return math.Round(v*scale) / scale
}

// ClampNumber is what an InputNumber keeps of s when it loses the focus:
// "" stays empty, a number is rounded to Precision and pulled within
// Min..Max, and anything else is rejected (ok false).
func ClampNumber(s string, p InputNumberProps) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", true
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return "", false
	}
	return p.format(p.clamp(v)), true
}

// StepNumber adds n steps to s within Min..Max. An empty or invalid s
// starts from Min when bounded, else 0.
func StepNumber(s string, n int, p InputNumberProps) string {
	step := p.Step
	if step == 0 {
		step = 1
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		v = 0
		if p.bounded() {
			v = p.Min
		}
		return p.format(p.clamp(v))
	}
	return p.format(p.clamp(v + float64(n)*step))
}

// InputNumber is Ant's InputNumber over a text value: up and down
// buttons at the end, the arrow keys step, and leaving the field pulls the
// value within Min..Max (or back to the last valid number). It reports a
// change.
func InputNumber(c *ui.Context, value *string, props ...InputNumberProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	h := t.Height(p.Size)
	changed := false
	box := ui.Row(c).Height(h).Radius(t.RadiusControl).AlignItems(ui.Stretch).Clip().Disabled(p.Disabled).
		Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorText))
	slot := compactBegin()
	defer slot.done()
	slot.at(box)
	if p.Width > 0 {
		box.Width(p.Width).Shrink(0)
	} else {
		box.FillWidth()
	}
	last := ui.Local(box, "last", func() string { return *value })
	wasFocused := ui.Local(box, "focused", func() bool { return false })
	v, err := strconv.ParseFloat(strings.TrimSpace(*value), 64)
	atMin := err == nil && p.bounded() && v <= p.Min
	atMax := err == nil && p.bounded() && v >= p.Max
	var in ui.Element
	box.Children(func() {
		in = ui.TextInputBase(c, value).Grow(1).Basis(0).MinWidth(0).Padding(0, t.PaddingInput).
			FontSize(t.ControlFont(p.Size)).TextColor(pick(p.Disabled, t.TextDisabled, t.Text)).
			PlaceholderColor(t.TextPlaceholder).FocusRing(false).Disabled(p.Disabled)
		takeLabelFocus(in)
		if p.Placeholder != "" {
			in.Placeholder(p.Placeholder)
		}
		if p.AutoFocus {
			in.AutoFocus()
		}
		if in.Changed() {
			changed = true
		}
		step := func(n int) {
			*value = StepNumber(*value, n, p)
			changed = true
		}
		if !p.Disabled {
			switch {
			case in.Shortcut(0, ui.KeyUp):
				step(1)
			case in.Shortcut(0, ui.KeyDown):
				step(-1)
			}
		}
		focused := in.Focused()
		if *wasFocused && !focused {
			if s, ok := ClampNumber(*value, p); ok {
				changed = changed || s != *value
				*value = s
			} else {
				*value = *last
				changed = true
			}
		}
		*wasFocused = focused
		if s, ok := ClampNumber(*value, p); ok && s == strings.TrimSpace(*value) {
			*last = *value
		}
		// KeepFocus leaves the focus in the field, so a click on a step
		// button is not a blur that clamps mid-edit.
		ui.Column(c).Width(22).Shrink(0).KeepFocus().Draw(func(pa *ui.Painter, r ui.Rect) {
			pa.Fill(ui.Rect{X: r.X, Y: r.Y, W: 1, H: r.H}, t.Border, 0)
			pa.Fill(ui.Rect{X: r.X, Y: r.Y + r.H/2 - 0.5, W: r.W, H: 1}, t.Border, 0)
		}).Children(func() {
			if stepButton(c, t, "增加", true, p.Disabled || atMax, p.force) {
				step(1)
			}
			if stepButton(c, t, "减少", false, p.Disabled || atMin, p.force) {
				step(-1)
			}
		})
	})
	st := p.force
	if st == rest {
		switch {
		case p.Disabled:
			st = disabled
		case in.Valid() && in.FocusVisible():
			st = focus
		case box.Hovered():
			st = hover
		}
	}
	slot.field(box, t, st, p.Status, fieldLook(box, t, st, p.Status))
	return changed
}

func stepButton(c *ui.Context, t *Tokens, label string, up, off bool, force visual) bool {
	// A plain box, not a button: Tab goes from the field to the next
	// control, as Ant's handlers are outside the tab order.
	b := ui.Box(c).Label(label).Role(ui.RoleButton).Disabled(off).Grow(1).Basis(0).
		FillWidth().Center().Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
	fg := t.TextMuted
	switch stateOf(b, off, pick(force == hover && up, hover, rest)) {
	case hover:
		fg = t.PrimaryHover
		b.Background(t.FillQuaternary)
	case active:
		fg = t.PrimaryActive
		b.Background(t.FillTertiary)
	case disabled:
		fg = t.TextDisabled
	}
	b.Children(func() {
		ui.Box(c).Size(10, 10).Draw(func(pa *ui.Painter, r ui.Rect) {
			var path ui.Path
			y0, y1 := r.Y+r.H*0.68, r.Y+r.H*0.32
			if !up {
				y0, y1 = y1, y0
			}
			path.MoveTo(r.X+r.W*0.15, y0).LineTo(r.X+r.W*0.5, y1).LineTo(r.X+r.W*0.85, y0)
			pa.StrokePath(&path, 1.3, fg)
		})
	})
	return b.Clicked() && !off
}
