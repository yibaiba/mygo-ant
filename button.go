package kit

import "github.com/egoist/mygo/ui"

// ButtonType is Ant's button type.
type ButtonType uint8

// Button types.
const (
	// Default is the white bordered button.
	Default ButtonType = iota
	// Primary is the filled button of a form's main action.
	Primary
	// TextButton has no border or fill until hovered.
	TextButton
	// Link looks like a link, underlined on hover.
	Link
	// Dashed is the default button with a dashed edge, as Form.List's add.
	Dashed
)

// ButtonShape is Ant's button shape.
type ButtonShape uint8

// Button shapes.
const (
	// ShapeDefault is the control-radius rectangle.
	ShapeDefault ButtonShape = iota
	// ShapeRound is a stadium: radius is half the height.
	ShapeRound
	// ShapeCircle is a square as tall as the control, fully round, as
	// icon-only buttons.
	ShapeCircle
)

// ButtonProps are a button's options; the zero value is a Middle default button.
type ButtonProps struct {
	Type     ButtonType
	Danger   bool
	Size     Size
	Loading  bool
	Disabled bool
	// Ghost keeps the fill empty and uses the type colour for the edge
	// and the label, as Ant ghost, for a button on a coloured ground.
	Ghost bool
	// Shape is Default, Round or Circle.
	Shape ButtonShape
	// Block fills the width of the parent.
	Block bool
	Icon  *ui.SVG
	// MinWidth overrides MinButtonWidth.
	MinWidth float32
	// AutoFocus gives the button the keyboard focus the frame it appears
	// (ui.AutoFocus), as Ant autoFocus.
	AutoFocus bool

	force visual
}

// Button draws an Ant button and reports a click. Its width never drops
// below what its label needs, so CJK labels never clip.
func Button(c *ui.Context, label string, props ...ButtonProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	slot := compactBegin()
	defer slot.done()
	off := p.Disabled || p.Loading
	h := t.Height(p.Size)
	padX := t.PaddingButton
	if p.Size == Small {
		padX = 7
	}
	bare := p.Type == TextButton || p.Type == Link
	if p.Type == Link {
		padX = 0
	}
	if p.Shape == ShapeCircle {
		padX = 0
	}
	radius := t.RadiusButton
	if p.Shape == ShapeRound || p.Shape == ShapeCircle {
		radius = h / 2
	}
	btn := ui.ButtonBase(c).Label(label).Disabled(off).FocusRing(false).
		Height(h).Padding(0, padX).Gap(t.SpaceSM).Radius(radius).Shrink(0).
		Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
	if p.Shape == ShapeCircle {
		btn.Size(h, h).Center()
	}
	if p.AutoFocus {
		btn.AutoFocus()
	}
	if p.Block && p.Shape != ShapeCircle {
		btn.FillWidth()
	} else {
		btn.AlignSelf(ui.Start)
	}
	min := p.MinWidth
	if min == 0 && !bare && p.Shape != ShapeCircle {
		min = MinButtonWidth(label, p.Size)
	}
	if min > 0 {
		btn.MinWidth(min)
	}
	st := stateOf(btn, p.Disabled, p.force)
	if p.Loading {
		st = rest
	}
	bg, fg, edge := buttonColors(t, p, st)
	btn.Background(bg)
	slot.at(btn)
	slot.part(btn, radius, st == hover || st == focus || st == active || p.Danger && st == rest && p.Type == Default, edge)
	if p.Type == Primary && !off && !p.Ghost {
		slot.joinLast(t.PrimaryHover)
	}
	if edge != ui.Transparent {
		btn.Border(1, edge)
	}
	if p.Type == Dashed {
		btn.BorderStyle(ui.BorderDashed)
	}
	if st == focus {
		focusRing(btn, t)
	}
	if p.Loading {
		btn.Opacity(0.65)
	}
	size := t.ControlFont(p.Size)
	btn.Children(func() {
		if p.Loading {
			ui.Spinner(c).Size(size, size).TextColor(fg)
		} else if p.Icon != nil {
			ui.Icon(c, p.Icon).Size(size, size).TextColor(fg)
		}
		if label != "" {
			txt := ui.Text(c, label).FontSize(size).TextColor(fg).SingleLine()
			if p.Type == Link && st == hover {
				txt.Underline()
			}
		}
	})
	return btn.Clicked() && !off
}

func buttonColors(t *Tokens, p ButtonProps, st visual) (bg, fg, edge ui.Color) {
	base, hov, act := t.Primary, t.PrimaryHover, t.PrimaryActive
	if p.Danger {
		base, hov, act = t.Error, t.ErrorHover, t.ErrorActive
	}
	tone := base
	switch st {
	case hover, focus:
		tone = hov
	case active:
		tone = act
	}
	if st == disabled {
		switch p.Type {
		case TextButton, Link:
			return ui.Transparent, t.TextDisabled, ui.Transparent
		}
		return t.DisabledBg, t.TextDisabled, t.BorderButton
	}
	if p.Ghost && p.Type != TextButton && p.Type != Link {
		fg, edge = t.Text, t.BorderButton
		if p.Type == Primary || p.Danger {
			fg, edge = tone, tone
		} else {
			switch st {
			case hover, focus:
				fg, edge = hov, hov
			case active:
				fg, edge = act, act
			}
		}
		return ui.Transparent, fg, edge
	}
	switch p.Type {
	case Primary:
		return tone, t.TextOnPrimary, tone
	case TextButton:
		fg = t.Text
		if p.Danger {
			fg = tone
		}
		switch st {
		case hover, focus:
			return t.FillSecondary, fg, ui.Transparent
		case active:
			return t.Fill, fg, ui.Transparent
		}
		return ui.Transparent, fg, ui.Transparent
	case Link:
		return ui.Transparent, tone, ui.Transparent
	}
	if st == rest {
		if p.Danger {
			return t.Surface, base, base
		}
		return t.Surface, t.Text, t.BorderButton
	}
	return t.Surface, tone, tone
}

// MinButtonWidth is the narrowest a button may be for its label: CJK
// glyphs a full em, Latin about half, the button's padding, and Ant's 88px
// floor for the default action buttons (RemoteConnect .searchBtn).
func MinButtonWidth(label string, size Size) float32 {
	em := float32(15)
	pad := float32(32)
	floor := float32(88)
	if size == Small {
		pad, floor = 16, 0
	}
	w := pad
	for _, r := range label {
		if r > 0x2e7f {
			w += em
		} else {
			w += em * 0.55
		}
	}
	return max(w, floor)
}

func first[T any](props []T) T {
	var zero T
	if len(props) == 0 {
		return zero
	}
	return props[0]
}
