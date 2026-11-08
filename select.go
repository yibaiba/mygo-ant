package kit

import "github.com/egoist/mygo/ui"

// Option is one choice of a Select, RadioButtons or Segmented.
type Option struct {
	Value, Label string
	// Disabled shows the option greyed out and keeps it from being chosen.
	Disabled bool
	// Group puts the option under a header in a Select's popup (Ant's
	// OptGroup); options of one group go together.
	Group string
}

// Opts makes options whose labels are their values.
func Opts(values ...string) []Option {
	out := make([]Option, len(values))
	for i, v := range values {
		out[i] = Option{Value: v, Label: v}
	}
	return out
}

func labelOf(options []Option, value string) string {
	for _, o := range options {
		if o.Value == value {
			return o.Label
		}
	}
	return value
}

// SelectProps are a select's options.
type SelectProps struct {
	Size        Size
	Disabled    bool
	Placeholder string
	Status      Status
	Width       float32
	// Loading shows a spinner in place of the arrow, as while options load.
	Loading bool
	Variant Variant
	// ShowSearch lets the user type to filter the options.
	ShowSearch bool
	// AllowClear shows a clear button over the arrow while there is a
	// choice.
	AllowClear bool
	// Mode is Multiple or Tags for MultiSelect.
	Mode SelectMode
	// MaxTagCount shows so many chosen values of a MultiSelect and "+ N"
	// for the rest; all of them when zero.
	MaxTagCount int
	// Virtual builds only the options in view; it is on anyway past 64.
	Virtual bool
	// NotFound is what the popup says when no option matches; "暂无数据"
	// when empty.
	NotFound string
	// NoFeedback hides the status icon Ant shows for Status.
	NoFeedback bool
	// AutoFocus gives the trigger the keyboard focus the frame it appears.
	AutoFocus bool

	force visual
	open  bool
}

// SelectMode is how many values a select takes, as Ant's mode.
type SelectMode uint8

// Select modes.
const (
	// Single takes one value, the default.
	Single SelectMode = iota
	// Multiple takes several, chosen from the options.
	Multiple
	// Tags takes several, and values typed that are not options.
	Tags
)

// Select chooses *value among options in an Ant select and reports a change.
func Select(c *ui.Context, value *string, options []Option, props ...SelectProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	if p.ShowSearch || p.AllowClear || p.Virtual || p.open || len(options) > virtualOver || hasGroups(options) {
		vals := []string{}
		if *value != "" {
			vals = []string{*value}
		}
		if !selectCore(c, &vals, options, p, false) {
			return false
		}
		*value = ""
		if len(vals) > 0 {
			*value = vals[0]
		}
		return true
	}
	t := Of(c)
	prev := *value
	slot := compactBegin()
	defer slot.done()
	sel := ui.SelectBase(c, value)
	trig := sel.Trigger.Height(t.Height(p.Size)).Padding(t.SpaceXS, t.PaddingInput).Gap(t.SpaceSM).
		Radius(t.RadiusControl).Justify(ui.SpaceBetween).AlignItems(ui.Center).
		FocusRing(false).Disabled(p.Disabled).Clip().
		Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
	if p.Width > 0 {
		trig.Width(p.Width).Shrink(0)
	} else {
		trig.FillWidth()
	}
	if p.AutoFocus {
		trig.AutoFocus()
	}
	takeLabelFocus(trig)
	open := sel.Open()
	st := stateOf(trig, p.Disabled, p.force)
	if open && st != disabled {
		st = focus
	}
	slot.at(trig)
	slot.field(trig, t, st, p.Status, fieldLookV(trig, t, st, p.Status, p.Variant))
	text, fg := labelOf(options, *value), t.Text
	if text == "" {
		text, fg = p.Placeholder, t.TextPlaceholder
	}
	if st == disabled {
		fg = t.TextDisabled
	}
	trig.Children(func() {
		ui.Text(c, text).FontSize(t.ControlFont(p.Size)).TextColor(fg).SingleLine().Grow(1)
		fieldStatusIcon(c, t, pick(p.NoFeedback, StatusNone, p.Status))
		if p.Loading {
			ui.Spinner(c).Size(12, 12).TextColor(t.TextDisabled)
		} else {
			chevron(c, t.TextDisabled, open)
		}
	})
	width := trig.Bounds().W
	sel.Popup(func(panel ui.Element) {
		panel.Margin(4, 0, 0, 0).Padding(4).Gap(2).Background(t.Elevated).Radius(t.RadiusPopup)
		shadowSecondary(panel)
		if width > 0 {
			panel.MinWidth(width)
		}
		for _, opt := range options {
			item := sel.Item(opt.Value).Height(t.ControlHeight).Padding(0, 12).Radius(4).AlignItems(ui.Center).
				Disabled(opt.Disabled).Cursor(pick(opt.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
			chosen := opt.Value == *value
			fg := pick(opt.Disabled, t.TextDisabled, t.Text)
			switch {
			case opt.Disabled:
			case chosen:
				item.Background(t.PrimaryBg)
			case item.Highlighted():
				item.Background(t.FillTertiary)
			}
			label := opt.Label
			item.Children(func() {
				txt := ui.Text(c, label).FontSize(t.FontSize).TextColor(fg).SingleLine()
				if chosen {
					txt.FontWeight(600)
				}
			})
		}
	})
	if trig.Changed() && optionOff(options, *value) {
		*value = prev
		return false
	}
	return trig.Changed()
}

func optionOff(options []Option, value string) bool {
	i := indexOf(options, value)
	return i >= 0 && options[i].Disabled
}

func chevron(c *ui.Context, color ui.Color, up bool) {
	ui.Box(c).Size(12, 12).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		var path ui.Path
		y0, y1 := r.Y+r.H*0.35, r.Y+r.H*0.68
		if up {
			y0, y1 = y1, y0
		}
		path.MoveTo(r.X+r.W*0.15, y0).LineTo(r.X+r.W*0.5, y1).LineTo(r.X+r.W*0.85, y0)
		p.StrokePath(&path, 1.4, color)
	})
}
