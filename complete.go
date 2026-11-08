package kit

import (
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/egoist/mygo/ui"
)

// AutoCompleteProps are an auto complete's options.
type AutoCompleteProps struct {
	Placeholder string
	Size        Size
	Disabled    bool
	Status      Status
	Width       float32
	AllowClear  bool
	// NoFilter shows every option whatever is typed, for options the app
	// computes from the text itself.
	NoFilter bool
	NotFound string
}

// AutoComplete is Ant's AutoComplete: a text input editing *value with a
// popup of the options matching it; choosing one, by click or the arrows
// and Enter, puts its value in the input. It reports a change.
func AutoComplete(c *ui.Context, value *string, options []Option, props ...AutoCompleteProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	options = valueLabels(options)
	changed := false
	slot := compactBegin()
	defer slot.done()
	host := ui.Column(c)
	slot.at(host)
	if p.Width > 0 {
		host.Width(p.Width).Shrink(0).AlignSelf(ui.Start)
	} else {
		host.FillWidth()
	}
	host.Children(func() {
		open := ui.Local(host, "open", func() bool { return false })
		active := ui.Local(host, "active", func() int { return -1 })
		shown := options
		if !p.NoFilter {
			shown = filterOptions(options, *value)
		}
		rows := optionRows(shown)
		var in ui.Element
		s := shellSpec{size: p.Size, status: p.Status, disabled: p.Disabled, clear: p.AllowClear, compact: slot}
		if fieldShell(c, t, s, value, func(c *ui.Context) ui.Element {
			in = ui.TextInputBase(c, value).FontSize(t.ControlFont(p.Size)).TextColor(t.Text).PlaceholderColor(t.TextPlaceholder).
				FocusRing(false).Disabled(p.Disabled).Grow(1).MinWidth(0).FillHeight()
			if p.Placeholder != "" {
				in.Placeholder(p.Placeholder)
			}
			in.OnTextKey(func(k ui.TextKey) bool {
				switch k.Key {
				case ui.KeyDown, ui.KeyUp:
					*open = true
					*active = stepActive(rows, *active, pick(k.Key == ui.KeyDown, 1, -1))
					return true
				case ui.KeyEnter:
					if *open && *active >= 0 && *active < len(rows) && rows[*active].header == "" {
						k.SetText(rows[*active].opt.Value)
						*open = false
						return true
					}
				case ui.KeyEscape:
					if *open {
						*open = false
						return true
					}
				}
				return false
			})
			return in
		}) {
			changed = true
		}
		if in.Changed() {
			changed = true
			*open = true
			*active = -1
		}
		was := ui.Local(host, "focused", func() bool { return false })
		if in.Focused() && !*was && len(options) > 0 {
			*open = true
		}
		*was = in.Focused()
		if len(rows) == 0 && p.NotFound == "" {
			*open = false
		}
		width := host.Bounds().W
		ui.PopoverBase(c, host, open, func(wrap ui.Element) {
			side := imePlacement(PlaceBottom, in.Valid() && in.Composing())
			placePanel(c, wrap, host, panelPlace{Side: side, Align: AlignFirst, Gap: 4}, t.Elevated, func(panel ui.Element) {
				popupLook(panel.Padding(4).Gap(0).Width(max(width, 120)), t).KeepFocus()
				panel.Children(func() {
					if v, ok := optionPanel(c, t, rows, func(v string) bool { return v == *value }, *active, false, len(rows) > virtualOver, p.NotFound, in); ok {
						*value = v
						*open = false
						changed = true
					}
				})
			})
		})
	})
	return changed
}

// mentionQuery finds the mention being typed before the caret (a rune
// offset into text): a prefix at the start or after a space, and the
// word after it up to the caret. It returns where the prefix starts, in
// runes, the prefix and the word.
func mentionQuery(text string, caret int, prefixes []string) (int, string, string, bool) {
	rs := []rune(text)
	if caret > len(rs) {
		caret = len(rs)
	}
	for i := caret - 1; i >= 0; i-- {
		if unicode.IsSpace(rs[i]) {
			return 0, "", "", false
		}
		for _, pre := range prefixes {
			pr := []rune(pre)
			if len(pr) == 0 || i+len(pr) > caret || string(rs[i:i+len(pr)]) != pre {
				continue
			}
			if i > 0 && !unicode.IsSpace(rs[i-1]) {
				continue
			}
			return i, pre, string(rs[i+len(pr) : caret]), true
		}
	}
	return 0, "", "", false
}

// insertMention replaces the mention typed from start to caret with
// prefix, value and a space, and returns the text and the caret after it.
func insertMention(text string, start, caret int, prefix, value string) (string, int) {
	rs := []rune(text)
	caret = min(caret, len(rs))
	ins := []rune(prefix + value + " ")
	out := append(append(append([]rune(nil), rs[:start]...), ins...), rs[caret:]...)
	return string(out), start + len(ins)
}

// MentionsProps are a mentions box's options.
type MentionsProps struct {
	// Prefix is what starts a mention; "@" when empty.
	Prefix      []string
	Placeholder string
	Rows        int
	AutoSize    bool
	Disabled    bool
	Status      Status
	NotFound    string
}

// caretRect is the caret at rune offset caret of s, shaped in font and
// wrapped at width (0 wraps only at newlines), from the top-left of the
// text: what ui lays a text input out with, measured again since ui
// gives the caret of an input only as offsets (TextSelection).
func caretRect(s string, caret int, font ui.Font, width float32) ui.Rect {
	rs := []rune(s)
	caret = min(max(caret, 0), len(rs))
	units := 0
	for _, r := range rs[:caret] {
		units += max(utf16.RuneLen(r), 1)
	}
	return ui.ShapeText(s, font, width).Caret(units)
}

// caretInView moves r, a caret in text h high at most, into a view of
// it viewH high scrolled as a text input scrolls to show its caret: by
// as little as it takes.
func caretInView(r ui.Rect, viewH float32) ui.Rect {
	if over := r.Y + r.H - viewH; viewH > 0 && over > 0 {
		r.Y -= over
	}
	return r
}

// Mentions is Ant's Mentions: a text area in which typing a prefix (@)
// opens a popup of the options matching the word after it, under the
// caret; choosing one writes it in. It reports a change of *value.
func Mentions(c *ui.Context, value *string, options []Option, props ...MentionsProps) bool {
	p := first(props)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	options = valueLabels(options)
	prefixes := p.Prefix
	if len(prefixes) == 0 {
		prefixes = []string{"@"}
	}
	changed := false
	host := ui.Column(c).FillWidth()
	host.Children(func() {
		active := ui.Local(host, "active", func() int { return -1 })
		pending := ui.Local(host, "pending", func() string { return "" })
		dismissed := ui.Local(host, "dismissed", func() string { return "\x00" })
		lastCaret := ui.Local(host, "caret", func() int { return 0 })
		heldOK := ui.Local(host, "ok", func() bool { return false })
		heldRows := ui.Local(host, "rows", func() []optionRow { return nil })
		moveCaret := ui.Local(host, "move", func() int { return -1 })
		if v := *pending; v != "" {
			*pending = ""
			if st, pr, _, found := mentionQuery(*value, *lastCaret, prefixes); found {
				text, at := insertMention(*value, st, *lastCaret, pr, v)
				*value = text
				*moveCaret = at
				changed = true
			}
		}
		var in ui.Element
		var rows []optionRow
		q, ok := "", false
		caret := 0
		s := shellSpec{status: p.Status, disabled: p.Disabled, multiline: true}
		fieldShell(c, t, s, value, func(c *ui.Context) ui.Element {
			in = ui.TextAreaBase(c, value).FontSize(t.ControlFont(SizeInherit)).TextColor(t.Text).PlaceholderColor(t.TextPlaceholder).
				FocusRing(false).Disabled(p.Disabled).Grow(1).MinWidth(0)
			if p.AutoSize {
				in.Lines(1, 1<<20)
			} else {
				in.Height(float32(pick(p.Rows > 0, p.Rows, 1)) * textLine)
			}
			if p.Placeholder != "" {
				in.Placeholder(p.Placeholder)
			}
			if *moveCaret >= 0 {
				in.SetTextSelection(*moveCaret, *moveCaret)
				*moveCaret = -1
			}
			if !in.Composing() {
				_, caret = in.TextSelection()
				*lastCaret = caret
				_, _, q, ok = mentionQuery(*value, caret, prefixes)
				if ok {
					rows = optionRows(filterOptions(options, q))
				}
				*heldOK = ok
				*heldRows = rows
			} else {
				ok, rows, caret = *heldOK, *heldRows, *lastCaret
			}
			in.OnTextKey(func(k ui.TextKey) bool {
				if !ok || *dismissed == *value {
					return false
				}
				switch k.Key {
				case ui.KeyDown, ui.KeyUp:
					*active = stepActive(rows, *active, pick(k.Key == ui.KeyDown, 1, -1))
					return true
				case ui.KeyEnter:
					if *active >= 0 && *active < len(rows) && rows[*active].header == "" {
						*pending = rows[*active].opt.Value
						return true
					}
				case ui.KeyEscape:
					*dismissed = *value
					return true
				}
				return false
			})
			return in
		})
		if takeCommitted(in.Changed(), in.Composing()) {
			changed = true
			*active = stepActive(rows, -1, 1)
		}
		composing := in.Valid() && in.Composing()
		open := ok && in.Focused() && *dismissed != *value
		if !open {
			*active = -1
		}
		width := host.Bounds().W
		anchor := host
		if ib, hb := in.Bounds(), host.Bounds(); open && ib.W > 0 {
			font := ui.Font{Family: t.Font, Size: t.FontSize, Lang: t.Lang}
			at := caretInView(caretRect(*value, caret, font, ib.W), ib.H)
			if composing {
				at.H += imeBand
			}
			anchor = ui.Box(c).Absolute().Left(ib.X-hb.X+at.X).Top(ib.Y-hb.Y+at.Y).Size(1, at.H).PassThrough()
		}
		ui.PopoverBase(c, anchor, &open, func(wrap ui.Element) {
			side := imePlacement(PlaceBottom, composing)
			placePanel(c, wrap, anchor, panelPlace{Side: side, Align: AlignFirst, Gap: 4}, t.Elevated, func(panel ui.Element) {
				popupLook(panel.Padding(4).Gap(0).Width(min(max(width/2, 160), 280)), t).KeepFocus()
				panel.Children(func() {
					if v, picked := optionPanel(c, t, rows, func(string) bool { return false }, *active, false, false, p.NotFound, ui.Element{}); picked {
						*pending = v
						c.Invalidate()
					}
				})
			})
		})
		if !open {
			// A press outside closed it; keep it closed for this text.
			if ok && in.Focused() && *dismissed != *value {
				*dismissed = *value
			}
		}
	})
	return changed
}

// mentionWords is the mentions in text after any of prefixes, for apps
// reading who was mentioned.
func mentionWords(text string, prefixes []string) []string {
	var out []string
	for _, f := range strings.Fields(text) {
		for _, pre := range prefixes {
			if strings.HasPrefix(f, pre) && len(f) > len(pre) {
				out = append(out, strings.TrimPrefix(f, pre))
			}
		}
	}
	return out
}

// MentionsOf is the names mentioned in text with prefix (@ when empty).
func MentionsOf(text string, prefix ...string) []string {
	if len(prefix) == 0 {
		prefix = []string{"@"}
	}
	return mentionWords(text, prefix)
}
