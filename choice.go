package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// ChoiceProps are the options of checkboxes, radios and switches.
type ChoiceProps struct {
	Disabled bool
	Size     Size
	Loading  bool
	// Indeterminate marks an unchecked checkbox as partly checked, as a
	// "check all" over a list some of which is checked: Ant's filled
	// square, painted, since ui.CheckboxBase has no third state.
	Indeterminate bool
	// CheckedChildren and UnCheckedChildren are Ant Switch's inner
	// labels (开 / 关). Other choice controls ignore them.
	CheckedChildren, UnCheckedChildren string

	force visual
}

// Checkbox toggles *on with an Ant check box and its label, and reports a change.
func Checkbox(c *ui.Context, on *bool, label string, props ...ChoiceProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	mark := choiceMark(p.Size)
	cb := ui.CheckboxBase(c, on).Label(label).Gap(t.SpaceSM).MinHeight(mark + 6).AlignSelf(ui.Start).
		AlignItems(ui.Center).Disabled(p.Disabled).FocusRing(false).
		Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
	st := stateOf(cb, p.Disabled, p.force)
	mixed := p.Indeterminate && !*on
	if mixed {
		cb.Mixed()
	}
	cb.Children(func() {
		box := ui.Box(c).Size(mark, mark).Radius(pick[float32](p.Size == Small, 3, 4)).Shrink(0)
		switch {
		case st == disabled:
			box.Background(t.DisabledBg).Border(1, t.Border)
		case *on:
			box.Background(pick(st == hover || st == active, t.PrimaryHover, t.Primary))
		default:
			box.Background(t.Surface).Border(1, pick(st == rest, t.BorderButton, t.Primary))
		}
		if st == focus {
			focusRing(box, t)
		}
		if mixed {
			dot := pick(st == disabled, t.TextDisabled, t.Primary)
			box.DrawOver(func(pa *ui.Painter, r ui.Rect) {
				s := r.W * 0.5
				pa.Fill(ui.Rect{X: r.X + r.W/2 - s/2, Y: r.Y + r.H/2 - s/2, W: s, H: s}, dot, 0)
			})
		}
		if *on {
			tick := pick(st == disabled, t.TextDisabled, t.TextOnPrimary)
			box.DrawOver(func(pa *ui.Painter, r ui.Rect) {
				var path ui.Path
				path.MoveTo(r.X+r.W*0.24, r.Y+r.H*0.52).
					LineTo(r.X+r.W*0.43, r.Y+r.H*0.70).
					LineTo(r.X+r.W*0.77, r.Y+r.H*0.32)
				pa.StrokePath(&path, pick[float32](p.Size == Large, 2.0, 1.8), tick)
			})
		}
		ui.Text(c, label).FontSize(t.ControlFont(p.Size)).TextColor(pick(st == disabled, t.TextDisabled, t.Text))
	})
	return cb.Changed()
}

// CheckboxGroupProps are a checkbox group's options.
type CheckboxGroupProps struct {
	Disabled bool
	// Vertical stacks the boxes; they wrap in a row otherwise.
	Vertical bool
}

// CheckboxGroup is Ant's Checkbox.Group: a checkbox per option, *values
// the checked ones in the options' order. It reports a change.
func CheckboxGroup(c *ui.Context, values *[]string, options []Option, props ...CheckboxGroupProps) bool {
	p := first(props)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	changed := false
	g := ui.Row(c).Gap(t.SpaceMD).Wrap()
	if p.Vertical {
		g = ui.Column(c).Gap(t.SpaceSM)
	}
	g.Children(func() {
		for _, o := range options {
			on := hasValue(*values, o.Value)
			if Checkbox(c.Key(o.Value), &on, o.Label, ChoiceProps{Disabled: p.Disabled || o.Disabled}) {
				*values = orderedValues(options, toggleValue(append([]string(nil), *values...), o.Value))
				changed = true
			}
		}
	})
	return changed
}

// orderedValues sorts values in the order of options, values that are
// no option last.
func orderedValues(options []Option, values []string) []string {
	out := make([]string, 0, len(values))
	for _, o := range options {
		if hasValue(values, o.Value) {
			out = append(out, o.Value)
		}
	}
	for _, v := range values {
		if indexOf(options, v) < 0 {
			out = append(out, v)
		}
	}
	return out
}

// CheckAllState is how a "check all" box over options shows values: all
// checked, or some (indeterminate). Disabled options do not count.
func CheckAllState(values []string, options []Option) (all, some bool) {
	n, on := 0, 0
	for _, o := range options {
		if o.Disabled {
			continue
		}
		n++
		if hasValue(values, o.Value) {
			on++
		}
	}
	return n > 0 && on == n, on > 0 && on < n
}

// CheckAll is a "check all" checkbox over options: checked when all of
// them are in *values, indeterminate when some are; toggling it checks or
// unchecks every option not disabled. It reports a change.
func CheckAll(c *ui.Context, values *[]string, options []Option, label string, props ...ChoiceProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	all, some := CheckAllState(*values, options)
	p.Indeterminate = some
	on := all
	if !Checkbox(c, &on, label, p) {
		return false
	}
	var next []string
	for _, o := range options {
		if o.Disabled && hasValue(*values, o.Value) || !o.Disabled && on {
			next = append(next, o.Value)
		}
	}
	*values = next
	return true
}

// Switch toggles *on with an Ant switch and reports a change.
func Switch(c *ui.Context, on *bool, props ...ChoiceProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	w, h, innerMin, innerMax, font := switchTrack(p.Size == Small, t.FontSizeSM)
	if p.CheckedChildren != "" || p.UnCheckedChildren != "" {
		onW, _ := c.MeasureText(0, ui.Span{Text: p.CheckedChildren, Size: font})
		offW, _ := c.MeasureText(0, ui.Span{Text: p.UnCheckedChildren, Size: font})
		w = max(w, max(onW, offW)+innerMin+innerMax)
	}
	off := p.Disabled || p.Loading
	sw := ui.SwitchBase(c, on).Size(w, h).Radius(h / 2).Shrink(0).Clip().Disabled(off).FocusRing(false).
		Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
	st := stateOf(sw, off, p.force)
	pos := sw.Animate("knob", pick[float32](*on, 1, 0), 200*time.Millisecond)
	track := t.SwitchOff
	onTone := t.Primary
	if st == hover || st == active {
		onTone = t.PrimaryHover
		if t.SwitchOff.A < 255 {
			track = t.SwitchOff.Alpha(1.8)
		}
	}
	sw.Background(track.Mix(onTone, pos))
	if off {
		sw.Opacity(0.65)
	}
	if st == focus {
		focusRing(sw, t)
	}
	label := pick(*on, p.CheckedChildren, p.UnCheckedChildren)
	if label != "" {
		checked := *on
		sw.Children(func() {
			// 0.8 Tester.Click(label) fails if the named node has ignoreInput.
			row := ui.Row(c).Fill().AlignItems(ui.Center)
			if checked {
				row.Padding(0, innerMax, 0, innerMin).Justify(ui.Start)
			} else {
				row.Padding(0, innerMin, 0, innerMax).Justify(ui.End)
			}
			row.Children(func() {
				ui.Text(c, label).FontSize(font).TextColor(t.TextOnPrimary).SingleLine()
			})
		})
	}
	spin, knobTone := p.Loading, t.SwitchKnob
	// Handle after the inner label, as Ant's .ant-switch-handle sits on top
	// of .ant-switch-inner: a 18×18 (small 12×12) knob, not a second segment.
	sw.DrawOver(func(pa *ui.Painter, r ui.Rect) {
		in := float32(2)
		d := r.H - 2*in
		knob := ui.Rect{X: r.X + in + pos*(r.W-r.H), Y: r.Y + in, W: d, H: d}
		pa.Shadow(knob, d/2, 0, 2, 4, 0, ui.RGBA(0, 35, 11, 0.2))
		pa.Fill(knob, knobTone, d/2)
		if spin {
			pa.Fill(ui.Rect{X: knob.X + d*0.3, Y: knob.Y + d*0.3, W: d * 0.4, H: d * 0.4}, onTone, d*0.2)
		}
	})
	return sw.Changed()
}

// switchTrack is Ant v5 Switch: track 44×22 (small 28×16), innerMinMargin 9
// (small 6), innerMaxMargin 24 (small 18). Inner text uses fontSizeSM.
func switchTrack(small bool, font float32) (w, h, innerMin, innerMax, size float32) {
	w, h, innerMin, innerMax, size = 44, 22, 9, 24, font
	if small {
		w, h, innerMin, innerMax = 28, 16, 6, 18
	}
	if size == 0 {
		size = 12
	}
	return
}

// Radio selects value into *selected with an Ant radio and reports a change.
func Radio(c *ui.Context, selected *string, value, label string, props ...ChoiceProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	mark := choiceMark(p.Size)
	rb := ui.RadioBase(c, selected, value).Label(label).Gap(t.SpaceSM).MinHeight(mark + 6).
		AlignItems(ui.Center).Disabled(p.Disabled).FocusRing(false).
		Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
	st := stateOf(rb, p.Disabled, p.force)
	on := *selected == value
	rb.Children(func() {
		dot := ui.Box(c).Size(mark, mark).Radius(mark / 2).Shrink(0)
		switch {
		case st == disabled:
			dot.Background(t.DisabledBg).Border(1, t.Border)
		case on:
			dot.Background(pick(st == hover, t.PrimaryHover, t.Primary))
		default:
			dot.Background(t.Surface).Border(1, pick(st == rest, t.BorderButton, t.Primary))
		}
		if st == focus {
			focusRing(dot, t)
		}
		if on {
			inner := pick(st == disabled, t.TextDisabled, t.TextOnPrimary)
			dot.DrawOver(func(pa *ui.Painter, r ui.Rect) {
				s := r.W * 0.375
				pa.Fill(ui.Rect{X: r.X + r.W/2 - s/2, Y: r.Y + r.H/2 - s/2, W: s, H: s}, inner, s/2)
			})
		}
		ui.Text(c, label).FontSize(t.ControlFont(p.Size)).TextColor(pick(st == disabled, t.TextDisabled, t.Text))
	})
	return rb.Changed()
}

// RadioButtonsProps are the options of a button-style radio group.
type RadioButtonsProps struct {
	Size     Size
	Disabled bool
	// Outline marks the choice with a primary border instead of a fill
	// (Ant's buttonStyle "outline"; the product uses "solid").
	Outline bool
}

// RadioButtons is Ant's Radio.Group with optionType "button": joined
// buttons, the chosen one filled. It reports a change.
func RadioButtons(c *ui.Context, value *string, options []Option, props ...RadioButtonsProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	idx := indexOf(options, *value)
	seg := ui.SegmentedBase(c, &idx, len(options))
	h := t.Height(p.Size)
	seg.Track.Height(h).AlignSelf(ui.Start).Shrink(0).Disabled(p.Disabled)
	seg.Track.Children(func() {
		for i, opt := range options {
			off := p.Disabled || opt.Disabled
			part := seg.Segment(i).Height(h).Padding(0, t.PaddingButton).MinWidth(MinButtonWidth(opt.Label, Small)).
				Center().FocusRing(false).Shrink(0).Disabled(opt.Disabled).
				Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
			radiusL, radiusR := float32(0), float32(0)
			if i == 0 {
				radiusL = t.RadiusButton
			}
			if i == len(options)-1 {
				radiusR = t.RadiusButton
			}
			part.Radius(radiusL, radiusR, radiusR, radiusL)
			if i > 0 {
				part.Margin(0, 0, 0, -1)
			}
			on := i == idx
			st := stateOf(part, off, rest)
			bg, fg, edge := t.Surface, t.Text, t.BorderButton
			switch {
			case st == disabled:
				bg, fg = t.DisabledBg, t.TextDisabled
				if on {
					bg = t.Fill
				}
			case on && !p.Outline:
				tone := pick(st == hover, t.PrimaryHover, pick(st == active, t.PrimaryActive, t.Primary))
				bg, fg, edge = tone, t.TextOnPrimary, tone
			case on:
				fg, edge = t.Primary, t.Primary
			case st == hover || st == active:
				fg = t.PrimaryHover
			}
			part.Background(bg).Border(1, edge)
			if st == focus {
				focusRing(part, t)
			}
			label := opt.Label
			part.Children(func() {
				ui.Text(c, label).FontSize(t.ControlFont(p.Size)).TextColor(fg).SingleLine()
			})
		}
	})
	return commitOption(seg.Track, idx, options, value)
}

// SegmentedProps are a segmented control's options, the same as
// RadioButtons'.
type SegmentedProps = RadioButtonsProps

// Segmented is Ant's Segmented: a grey track whose chosen item is a raised
// white chip. It reports a change.
func Segmented(c *ui.Context, value *string, options []Option, props ...SegmentedProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	idx := indexOf(options, *value)
	seg := ui.SegmentedBase(c, &idx, len(options))
	h := t.Height(p.Size)
	seg.Track.Height(h).Padding(2).Radius(t.RadiusButton + 2).Background(t.FillTertiary).
		AlignSelf(ui.Start).Shrink(0).Disabled(p.Disabled)
	seg.Track.Children(func() {
		for i, opt := range options {
			off := p.Disabled || opt.Disabled
			part := seg.Segment(i).Height(h-4).Padding(0, 11).Radius(t.RadiusButton).Center().Shrink(0).
				MinWidth(MinButtonWidth(opt.Label, Small)).FocusRing(false).Disabled(opt.Disabled).
				Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
			on := i == idx
			st := stateOf(part, off, rest)
			fg := t.TextSecondary
			switch {
			case st == disabled:
				fg = t.TextDisabled
			case on:
				fg = t.Text
				part.Background(t.Elevated).
					Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.03)).
					Shadow(0, 1, 6, -1, ui.RGBA(0, 0, 0, 0.02)).
					Shadow(0, 2, 4, 0, ui.RGBA(0, 0, 0, 0.02))
			case st == hover:
				fg = t.Text
				part.Background(t.FillSecondary)
			case st == active:
				part.Background(t.Fill)
			}
			if st == focus {
				focusRing(part, t)
			}
			label := opt.Label
			part.Children(func() {
				ui.Text(c, label).FontSize(t.ControlFont(p.Size)).TextColor(fg).SingleLine()
			})
		}
	})
	return commitOption(seg.Track, idx, options, value)
}

// choiceMark is the checkbox / radio glyph: 14 Small, 16 Middle, 18 Large
// (Ant's large checkbox is slightly bigger than 16).
func choiceMark(s Size) float32 {
	switch s {
	case Small:
		return 14
	case Large:
		return 18
	}
	return 16
}

func indexOf(options []Option, value string) int {
	for i, o := range options {
		if o.Value == value {
			return i
		}
	}
	return -1
}

// commitOption writes the option a segmented track moved to, unless it
// is disabled.
func commitOption(track ui.Element, idx int, options []Option, value *string) bool {
	if !track.Changed() || idx < 0 || idx >= len(options) || options[idx].Disabled {
		return false
	}
	*value = options[idx].Value
	return true
}
