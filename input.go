package kit

import (
	"fmt"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
)

// Status marks a field invalid or doubtful, as Ant's status prop.
type Status uint8

// Field statuses.
const (
	// StatusNone is a field without a mark.
	StatusNone Status = iota
	// StatusError draws a red border, as for a failed validation.
	StatusError
	// StatusWarning draws an orange border.
	StatusWarning
)

// Variant is how a field's box is drawn, as Ant's variant prop.
type Variant uint8

// Field variants.
const (
	// Outlined is a white box with a border, the default.
	Outlined Variant = iota
	// Filled is a grey box without a border until hovered or focused.
	Filled
	// Borderless is the bare text, without a box.
	Borderless
)

// InputProps are a text input's options.
type InputProps struct {
	// Password hides the text, with an eye button at the end showing it.
	Password    bool
	Placeholder string
	Size        Size
	Disabled    bool
	ReadOnly    bool
	Status      Status
	Variant     Variant
	// Width fixes the width; zero fills the parent.
	Width float32

	// Prefix and Suffix are short texts inside the box, before and after
	// the text (as "¥" or "RMB"); PrefixIcon and SuffixIcon icons there.
	Prefix, Suffix         string
	PrefixIcon, SuffixIcon *ui.SVG
	// AddonBefore and AddonAfter are grey labels joined to the box
	// outside it (as "https://" or ".com"); AddonAfterFn builds controls
	// there instead, as Search's button.
	AddonBefore, AddonAfter string
	AddonAfterFn            func()
	// AllowClear shows a clear button while the input holds text.
	AllowClear bool
	// MaxLength cuts what is typed to so many characters; ShowCount shows
	// how many there are (out of MaxLength).
	MaxLength int
	ShowCount bool
	// NoFeedback hides the status icon Ant shows for Status (hasFeedback
	// false). The border still follows Status.
	NoFeedback bool
	// AutoFocus gives the field the keyboard focus the frame it appears
	// (ui.AutoFocus), as Ant autoFocus.
	AutoFocus bool
	// OnPressEnter runs when Enter is pressed in the field (Ant
	// onPressEnter), from ui.TextInputBase.Submitted.
	OnPressEnter func()

	force visual
}

// Input edits *value in an Ant input box and reports a change.
func Input(c *ui.Context, value *string, props ...InputProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	var in ui.Element
	shell := inputChrome(p)
	shell.compact = compactBegin()
	defer shell.compact.done()
	var reveal *bool
	if p.Password {
		shell.suffix = func(c *ui.Context, fg ui.Color) {
			if iconButton(c, pick(*reveal, iconEye, iconEyeOff), pick(*reveal, "隐藏密码", "显示密码"), fg, 14) {
				*reveal = !*reveal
				c.Invalidate()
			}
		}
	}
	changed := fieldShell(c, t, shell, value, func(c *ui.Context) ui.Element {
		in = ui.TextInputBase(c, value).FontSize(t.ControlFont(p.Size)).TextColor(t.Text).
			PlaceholderColor(t.TextPlaceholder).FocusRing(false).
			Disabled(p.Disabled).ReadOnly(p.ReadOnly).Grow(1).MinWidth(0).FillHeight()
		if p.Password {
			reveal = ui.Local(in, "reveal", func() bool { return false })
			if !*reveal {
				in.Password()
			}
		}
		if p.Placeholder != "" {
			in.Placeholder(p.Placeholder)
		}
		if p.AutoFocus {
			in.AutoFocus()
		}
		return in
	})
	if takeCommitted(in.Changed(), in.Composing()) {
		clampRunes(value, p.MaxLength)
		changed = true
	}
	if in.Submitted() && p.OnPressEnter != nil {
		p.OnPressEnter()
	}
	return changed
}

// takeCommitted is whether a text field's bound value should be treated
// as committed this frame. While an IME is composing (Element.Composing),
// MaxLength, OTP auto-advance and Mentions caret layout wait until
// composition ends, so preedit is not clamped, advanced or shaped as text.
func takeCommitted(changed, composing bool) bool {
	return changed && !composing
}

// clampRunes cuts *s to n characters, when n is positive. Call it only
// after takeCommitted; preedit is not in the bound value, but a Changed
// that fires while composing must not cut the field.
func clampRunes(s *string, n int) {
	if n <= 0 || utf8.RuneCountInString(*s) <= n {
		return
	}
	r := []rune(*s)
	*s = string(r[:n])
}

// shellSpec is the chrome around a text field: what the Ant inputs share.
type shellSpec struct {
	size       Size
	variant    Variant
	status     Status
	disabled   bool
	width      float32
	force      visual
	prefixText string
	prefixIcon *ui.SVG
	suffixText string
	suffixIcon *ui.SVG
	// suffix builds controls at the end of the box, in fg.
	suffix                  func(c *ui.Context, fg ui.Color)
	addonBefore, addonAfter string
	addonAfterFn            func()
	clear                   bool
	maxLength               int
	count                   bool
	// feedback is the status icon drawn at the end of the box.
	feedback Status
	// multiline lets the box grow with a text area, the count under it.
	multiline bool
	// compact is the control's place in a SpaceCompact, if any.
	compact *compactSlot
}

func inputChrome(p InputProps) shellSpec {
	return shellSpec{
		size: p.Size, variant: p.Variant, status: p.Status, disabled: p.Disabled, width: p.Width, force: p.force,
		prefixText: p.Prefix, prefixIcon: p.PrefixIcon, suffixText: p.Suffix, suffixIcon: p.SuffixIcon,
		addonBefore: p.AddonBefore, addonAfter: p.AddonAfter, addonAfterFn: p.AddonAfterFn,
		clear: p.AllowClear, maxLength: p.MaxLength, count: p.ShowCount,
		feedback: pick(p.NoFeedback, StatusNone, p.Status),
	}
}

// countText is what ShowCount shows for value.
func countText(value string, maxLength int) string {
	n := utf8.RuneCountInString(value)
	if maxLength > 0 {
		return fmt.Sprintf("%d / %d", n, maxLength)
	}
	return fmt.Sprint(n)
}

// fieldShell builds the box of a text field as s says around the field
// field builds, which value is the text of, and reports whether the clear
// button emptied it.
func fieldShell(c *ui.Context, t *Tokens, s shellSpec, value *string, field func(c *ui.Context) ui.Element) bool {
	if s.count && s.multiline {
		// The count sits under the box, at its end: hold both in a column
		// as wide as the field.
		col := ui.Column(c).Gap(2)
		s.compact.at(col)
		if s.width > 0 {
			col.Width(s.width).Shrink(0).AlignSelf(ui.Start)
		} else {
			col.FillWidth()
		}
		cleared := false
		col.Children(func() {
			inner := s
			inner.width, inner.count = 0, false
			cleared = fieldShell(c, t, inner, value, field)
			ui.Row(c).FillWidth().Justify(ui.End).Children(func() {
				ui.Text(c, countText(*value, s.maxLength)).FontSize(t.ControlFont(s.size)).TextColor(t.TextTertiary)
			})
		})
		return cleared
	}
	h := t.Height(s.size)
	outer := ui.Row(c).AlignItems(pick(s.multiline, ui.Start, ui.Center))
	s.compact.at(outer)
	if s.width > 0 {
		outer.Width(s.width).Shrink(0)
	} else {
		outer.FillWidth()
	}
	before := s.addonBefore != ""
	after := s.addonAfter != "" || s.addonAfterFn != nil
	cleared := false
	r := t.RadiusControl
	outer.Children(func() {
		if before {
			s.compact.part(addon(c, t, s.addonBefore, nil, h, true), r, false, t.Border)
		}
		// Ant v5: 4px 11px padding inside a 32/40px border-box, clipped so
		// the eye / clear / count stay in the shell.
		box := ui.Row(c).Grow(1).MinWidth(0).Padding(t.SpaceXS, t.PaddingInput).Gap(4).AlignItems(ui.Center).
			Radius(pick[float32](before, 0, r), pick[float32](after, 0, r), pick[float32](after, 0, r), pick[float32](before, 0, r)).
			Clip().Cursor(pick(s.disabled, ui.CursorNotAllowed, ui.CursorText))
		if s.multiline {
			box.MinHeight(h).AlignItems(ui.Start)
		} else {
			box.Height(h)
		}
		var in ui.Element
		box.Children(func() {
			fg := pick(s.disabled, t.TextDisabled, t.TextTertiary)
			if s.prefixIcon != nil {
				ui.Icon(c, s.prefixIcon).Size(14, 14).TextColor(fg).Shrink(0)
			}
			if s.prefixText != "" {
				ui.Text(c, s.prefixText).FontSize(t.ControlFont(s.size)).TextColor(t.Text).SingleLine().Shrink(0)
			}
			in = field(c)
			takeLabelFocus(in)
			lit := box.Hovered() || in.Focused()
			if s.clear && !s.disabled && *value != "" {
				clr := clearButton(c, t, lit)
				if clr {
					*value = ""
					cleared = true
				}
			}
			if s.count && !s.multiline {
				ui.Text(c, countText(*value, s.maxLength)).FontSize(t.ControlFont(s.size)).TextColor(t.TextTertiary).SingleLine().Shrink(0)
			}
			if s.suffix != nil {
				s.suffix(c, fg)
			}
			if s.suffixText != "" {
				ui.Text(c, s.suffixText).FontSize(t.ControlFont(s.size)).TextColor(t.Text).SingleLine().Shrink(0)
			}
			if s.suffixIcon != nil {
				ui.Icon(c, s.suffixIcon).Size(14, 14).TextColor(fg).Shrink(0)
			}
			fieldStatusIcon(c, t, s.feedback)
		})
		st := s.force
		if st == rest {
			switch {
			case s.disabled:
				st = disabled
			case in.FocusVisible():
				st = focus
			case box.Hovered():
				st = hover
			}
		}
		s.compact.field(box, t, st, s.status, fieldLookV(box, t, st, s.status, s.variant))
		if after {
			if s.addonAfterFn != nil {
				ui.Row(c).Shrink(0).Height(h).AlignItems(ui.Stretch).Children(s.addonAfterFn)
			} else {
				s.compact.part(addon(c, t, s.addonAfter, nil, h, false), r, false, t.Border)
			}
		}
	})
	return cleared
}

// addon is a grey label joined to a field's box, on its start when
// atStart.
func addon(c *ui.Context, t *Tokens, label string, icon *ui.SVG, h float32, atStart bool) ui.Element {
	r := t.RadiusControl
	a := ui.Row(c).Shrink(0).Height(h).Padding(0, t.PaddingInput).AlignItems(ui.Center).Gap(4).
		Background(t.FillAlter).BorderColor(t.Border)
	if atStart {
		a.BorderWidth(1, 0, 1, 1).Radius(r, 0, 0, r)
	} else {
		a.BorderWidth(1, 1, 1, 0).Radius(0, r, r, 0)
	}
	a.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).Size(14, 14).TextColor(t.Text)
		}
		ui.Text(c, label).FontSize(t.ControlFont(SizeInherit)).TextColor(t.Text).SingleLine()
	})
	return a
}

// clearButton is the grey round × that empties a field, faint until the
// field is hovered or focused (lit); it reports a click.
func clearButton(c *ui.Context, t *Tokens, lit bool) bool {
	b := ui.ButtonBase(c).Label("清除").Size(14, 14).Shrink(0).FocusRing(false).Cursor(ui.CursorPointer).
		Opacity(pick[float32](lit, 1, 0))
	tone := pick(b.Hovered(), t.TextTertiary, t.TextDisabled)
	surface := t.Surface
	b.Draw(func(p *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		var dot ui.Path
		p.FillPath(dot.Circle(cx, cy, 6), tone)
		var x ui.Path
		x.MoveTo(cx-2.5, cy-2.5).LineTo(cx+2.5, cy+2.5).MoveTo(cx+2.5, cy-2.5).LineTo(cx-2.5, cy+2.5)
		p.StrokePath(x.Cap(ui.RoundCap), 1.2, surface)
	})
	return b.Clicked()
}

// iconButton is a bare icon button labelled label for assistive
// technology, as the eye of a password; it reports a click.
func iconButton(c *ui.Context, icon *ui.SVG, label string, fg ui.Color, size float32) bool {
	hit := size + 8 // 22px for a 14px glyph, inside a 32/40px field
	b := ui.ButtonBase(c).Label(label).Size(hit, hit).Shrink(0).FocusRing(false).Cursor(ui.CursorPointer).Center()
	b.Children(func() {
		ui.Icon(c, icon).Size(size, size).TextColor(pick(b.Hovered(), fg.Mix(ui.RGB(0, 0, 0), 0.3), fg))
	})
	return b.Clicked()
}

// fieldStatusIcon is Ant Input's hasFeedback mark: a filled status
// circle after the suffix when Status is error or warning.
func fieldStatusIcon(c *ui.Context, t *Tokens, status Status) {
	switch status {
	case StatusError:
		IconNamed(c, "CloseCircleFilled", IconProps{Size: 14, Color: t.Error})
	case StatusWarning:
		IconNamed(c, "ExclamationCircleFilled", IconProps{Size: 14, Color: t.WarningIcon})
	}
}

// outlineOf is the 2px focus glow: primary (or status) at ~0.2 alpha, as
// Ant's controlOutline — not a solid PrimaryBg slab and not ui's 2px
// black ring.
func outlineOf(t *Tokens, status Status) ui.Color {
	switch status {
	case StatusError:
		return t.Error.Alpha(0.2)
	case StatusWarning:
		return t.Warning.Alpha(0.2)
	}
	return t.ControlOutline
}

// fieldLook is the outlined box of inputs and selects in each state.
func fieldLook(e ui.Element, t *Tokens, st visual, status Status) ui.Color {
	return fieldLookV(e, t, st, status, Outlined)
}

// fieldLookV is the box of inputs and selects in each state and variant;
// it returns the colour of the box's edge.
func fieldLookV(e ui.Element, t *Tokens, st visual, status Status, v Variant) ui.Color {
	edge, hov, ring := t.Border, t.Primary, outlineOf(t, status)
	switch status {
	case StatusError:
		edge, hov = t.Error, t.ErrorHover
	case StatusWarning:
		edge, hov = t.Warning, t.Warning
	}
	bg := t.Surface
	switch v {
	case Filled:
		bg = pick(status == StatusError, t.ErrorBg, pick(status == StatusWarning, t.WarningBg, t.FillTertiary))
		if st == hover {
			bg = pick(status == StatusNone, t.FillSecondary, bg)
		}
		edge = ui.Transparent
		if st == focus || st == active {
			bg, edge = t.Surface, pick(status == StatusNone, t.Primary, hov)
			e.Shadow(0, 0, 0, 2, ring)
		}
		if st == disabled {
			bg, edge = t.DisabledBg, t.Border
			e.TextColor(t.TextDisabled)
			e.Cursor(ui.CursorNotAllowed)
		}
		e.Background(bg).Border(1, edge)
		return edge
	case Borderless:
		if st == disabled {
			e.TextColor(t.TextDisabled)
			e.Cursor(ui.CursorNotAllowed)
		}
		e.Background(ui.Transparent).Border(1, ui.Transparent)
		return ui.Transparent
	}
	switch st {
	case hover:
		edge = hov
	case focus, active:
		edge = pick(status == StatusNone, t.Primary, edge)
		e.Shadow(0, 0, 0, 2, ring)
	case disabled:
		bg, edge = t.DisabledBg, t.Border
		e.TextColor(t.TextDisabled)
		e.Cursor(ui.CursorNotAllowed)
	}
	e.Background(bg).Border(1, edge)
	return edge
}
