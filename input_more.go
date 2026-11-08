package kit

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// TextAreaProps are a text area's options.
type TextAreaProps struct {
	Placeholder string
	Disabled    bool
	ReadOnly    bool
	Status      Status
	Variant     Variant
	Width       float32
	// Rows is how many lines it shows; 2 when zero. AutoSize makes it as
	// tall as its text instead, from MinRows (1) up to MaxRows (no limit
	// when zero), past which it scrolls.
	Rows             int
	AutoSize         bool
	MinRows, MaxRows int
	AllowClear       bool
	// MaxLength cuts what is typed to so many characters; ShowCount shows
	// how many there are under the box.
	MaxLength int
	ShowCount bool
	// NoFeedback hides the status icon Ant shows for Status.
	NoFeedback bool
	// AutoFocus gives the area the keyboard focus the frame it appears.
	AutoFocus bool
	// OnPressEnter runs when Enter is pressed (Ant onPressEnter).
	OnPressEnter func()
}

// textLine is the height of a line of 14px text in a text area.
const textLine = 22

// TextArea is Ant's Input.TextArea: it edits *value over several lines
// and reports a change.
func TextArea(c *ui.Context, value *string, props ...TextAreaProps) bool {
	p := first(props)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	var in ui.Element
	slot := compactBegin()
	defer slot.done()
	s := shellSpec{variant: p.Variant, status: p.Status, disabled: p.Disabled, width: p.Width,
		clear: p.AllowClear, maxLength: p.MaxLength, count: p.ShowCount, multiline: true, compact: slot,
		feedback: pick(p.NoFeedback, StatusNone, p.Status)}
	changed := fieldShell(c, t, s, value, func(c *ui.Context) ui.Element {
		in = ui.TextAreaBase(c, value).FontSize(t.ControlFont(SizeInherit)).TextColor(t.Text).
			PlaceholderColor(t.TextPlaceholder).FocusRing(false).Disabled(p.Disabled).ReadOnly(p.ReadOnly).
			Grow(1).MinWidth(0)
		if p.AutoSize {
			in.Lines(max(p.MinRows, 1), pick(p.MaxRows > 0, p.MaxRows, 1<<20))
		} else {
			in.Height(float32(pick(p.Rows > 0, p.Rows, 2)) * textLine)
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

// SearchProps are a search box's options.
type SearchProps struct {
	Placeholder string
	Size        Size
	Disabled    bool
	Width       float32
	AllowClear  bool
	// EnterButton, when set, is the label of a primary button joined to
	// the end of the box ("" leaves a grey search button there).
	EnterButton string
	// Loading spins in the button while a search runs.
	Loading bool
}

// Search is Ant's Input.Search: a box with a search button at its end.
// It reports a search, by Enter or the button.
func Search(c *ui.Context, value *string, props ...SearchProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	searched := false
	var in ui.Element
	h := t.Height(p.Size)
	slot := compactBegin()
	defer slot.done()
	s := shellSpec{size: p.Size, disabled: p.Disabled, width: p.Width, clear: p.AllowClear, compact: slot,
		addonAfterFn: func() {
			r := t.RadiusControl
			if p.EnterButton != "" {
				b := ui.ButtonBase(c).Label(p.EnterButton).Height(h).Padding(0, t.PaddingButton).Gap(t.SpaceSM).
					Radius(0, r, r, 0).AlignItems(ui.Center).FocusRing(false).Disabled(p.Disabled).
					Margin(0, 0, 0, -1).Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
				st := stateOf(b, p.Disabled, rest)
				bg, fg, edge := buttonColors(t, ButtonProps{Type: Primary}, st)
				b.Background(bg).Border(1, edge)
				slot.part(b, r, false, edge)
				b.Children(func() {
					if p.Loading {
						ui.Spinner(c).Size(14, 14).TextColor(fg)
					} else {
						ui.Icon(c, iconSearch).Size(14, 14).TextColor(fg)
					}
					ui.Text(c, p.EnterButton).FontSize(t.ControlFont(p.Size)).TextColor(fg).SingleLine()
				})
				searched = b.Clicked() && !p.Loading
				return
			}
			b := ui.ButtonBase(c).Label(L(c).Search).Size(h, h).Radius(0, r, r, 0).Center().FocusRing(false).
				Margin(0, 0, 0, -1).Disabled(p.Disabled).Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
			st := stateOf(b, p.Disabled, rest)
			bg, fg, edge := buttonColors(t, ButtonProps{}, st)
			b.Background(bg).Border(1, edge)
			slot.part(b, r, st == hover || st == focus || st == active, edge)
			b.Children(func() {
				if p.Loading {
					ui.Spinner(c).Size(14, 14).TextColor(fg)
				} else {
					ui.Icon(c, iconSearch).Size(14, 14).TextColor(pick(st == hover, fg, t.TextTertiary))
				}
			})
			searched = b.Clicked() && !p.Loading
		}}
	fieldShell(c, t, s, value, func(c *ui.Context) ui.Element {
		in = ui.TextInputBase(c, value).FontSize(t.ControlFont(p.Size)).TextColor(t.Text).
			PlaceholderColor(t.TextPlaceholder).FocusRing(false).Disabled(p.Disabled).Grow(1).MinWidth(0).FillHeight()
		if p.Placeholder != "" {
			in.Placeholder(p.Placeholder)
		}
		return in
	})
	if in.Submitted() && !p.Loading {
		searched = true
	}
	return searched
}

// OTPProps are a one-time-code input's options.
type OTPProps struct {
	// Length is how many characters the code has; 6 when zero.
	Length   int
	Size     Size
	Disabled bool
	Status   Status
	// Mask shows each character as a dot, as for a PIN.
	Mask bool
}

// otpApply puts what was typed into box i of a code: its last character
// there, or, pasted, one character a box from i on. It returns the code
// and the box to go to next (i when the box was emptied).
func otpApply(code []string, i int, typed string) ([]string, int) {
	out := append([]string(nil), code...)
	if typed == "" {
		out[i] = ""
		return out, i
	}
	rs := []rune(typed)
	if len(rs) == 2 && out[i] != "" && string(rs[0]) == out[i] {
		rs = rs[1:]
	}
	if len(rs) == 1 {
		out[i] = string(rs[0])
		return out, min(i+1, len(out)-1)
	}
	j := i
	for _, r := range rs {
		if j >= len(out) {
			break
		}
		out[j] = string(r)
		j++
	}
	return out, min(j, len(out)-1)
}

// otpSplit is code as one string per box of n.
func otpSplit(code string, n int) []string {
	out := make([]string, n)
	i := 0
	for _, r := range code {
		if i >= n {
			break
		}
		out[i] = string(r)
		i++
	}
	return out
}

// OTP is Ant's Input.OTP: a row of one-character boxes editing *value.
// Typing moves to the next box, Backspace in an empty box to the one
// before, and a pasted code fills the boxes. It reports a change.
func OTP(c *ui.Context, value *string, props ...OTPProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	n := pick(p.Length > 0, p.Length, 6)
	h := t.Height(p.Size)
	changed := false
	row := ui.Row(c).Gap(t.SpaceSM).AlignSelf(ui.Start)
	want := ui.Local(row, "focus", func() int { return -1 })
	back := ui.Local(row, "back", func() int { return -1 })
	code := otpSplit(*value, n)
	if b := *back; b > 0 {
		code[b-1] = ""
		*value = strings.Join(code, "")
		*want = b - 1
		changed = true
	}
	*back = -1
	row.Children(func() {
		for i := 0; i < n; i++ {
			cell := code[i]
			in := ui.TextInputBase(c.Key(i), &cell).Size(h, h).Padding(0).FontSize(t.ControlFont(p.Size)).TextColor(t.Text).
				Radius(t.RadiusControl).FocusRing(false).Disabled(p.Disabled).Clip().
				Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorText)).
				TextAlign(ui.Center).Label("验证码第 " + strconv.Itoa(i+1) + " 位")
			if p.Mask && cell != "" {
				in.Password()
			}
			if *want == i {
				in.Focus()
			}
			st := stateOf(in, p.Disabled, rest)
			fieldLook(in, t, st, p.Status)
			if takeCommitted(in.Changed(), in.Composing()) {
				next, at := otpApply(code, i, cell)
				code = next
				*value = strings.Join(code, "")
				*want = at
				changed = true
			}
			i := i
			in.OnTextKey(func(k ui.TextKey) bool {
				if k.Key == ui.KeyBackspace && k.Text() == "" && i > 0 {
					*back = i
					return true
				}
				return false
			})
		}
	})
	if *want >= 0 && !changed {
		*want = -1
	}
	return changed
}
