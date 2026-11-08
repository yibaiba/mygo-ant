package kit

import (
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// selectRowH is the height of an option row in a select's popup.
const selectRowH = 32

// selectMaxH is how tall a select's popup grows before it scrolls.
const selectMaxH = 256

// filterOptions keeps the options whose label or value holds query,
// ignoring case: Ant's default filterOption.
func filterOptions(options []Option, query string) []Option {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return options
	}
	var out []Option
	for _, o := range options {
		if strings.Contains(strings.ToLower(o.Label), q) || strings.Contains(strings.ToLower(o.Value), q) {
			out = append(out, o)
		}
	}
	return out
}

// valueLabels is options with each empty Label set to its Value, as Ant
// shows an AutoComplete option given only a value.
func valueLabels(options []Option) []Option {
	for i := range options {
		if options[i].Label == "" {
			out := append([]Option(nil), options...)
			for j := range out {
				if out[j].Label == "" {
					out[j].Label = out[j].Value
				}
			}
			return out
		}
	}
	return options
}

// hasGroups reports whether any option sits in a group.
func hasGroups(options []Option) bool {
	for _, o := range options {
		if o.Group != "" {
			return true
		}
	}
	return false
}

// optionRow is a row of a select popup: a group header or an option.
type optionRow struct {
	header string
	opt    Option
}

// optionRows lays options out as popup rows, a header before each run of
// options of one group.
func optionRows(options []Option) []optionRow {
	var rows []optionRow
	group := ""
	for _, o := range options {
		if o.Group != group && o.Group != "" {
			rows = append(rows, optionRow{header: o.Group})
		}
		group = o.Group
		rows = append(rows, optionRow{opt: o})
	}
	return rows
}

// stepActive moves the keyboard highlight from i by d among rows, past
// headers and disabled options, wrapping at the ends; -1 when none can be
// highlighted.
func stepActive(rows []optionRow, i, d int) int {
	n := len(rows)
	if n == 0 {
		return -1
	}
	if i < 0 && d < 0 {
		i = n
	}
	for k := 0; k < n; k++ {
		i = ((i+d)%n + n) % n
		if rows[i].header == "" && !rows[i].opt.Disabled {
			return i
		}
	}
	return -1
}

// toggleValue adds v to values, or takes it out when there.
func toggleValue(values []string, v string) []string {
	for i, x := range values {
		if x == v {
			return append(values[:i:i], values[i+1:]...)
		}
	}
	return append(values, v)
}

func hasValue(values []string, v string) bool {
	for _, x := range values {
		if x == v {
			return true
		}
	}
	return false
}

// optionPanel builds the rows of a select's popup, rows tall at most
// selectMaxH and only those in view when virtual, with the row at active
// highlighted and kept in view. It returns the value of the option
// clicked.
func optionPanel(c *ui.Context, t *Tokens, rows []optionRow, chosen func(string) bool, active int, multiple, virtual bool, notFound string, desc ui.Element) (string, bool) {
	_ = virtual
	picked, ok := "", false
	if len(rows) == 0 {
		ui.Column(c).FillWidth().Padding(12).AlignItems(ui.Center).Gap(4).Children(func() {
			ui.Icon(c, iconEmpty).Size(32, 22).TextColor(t.EmptyIcon)
			ui.Text(c, pick(notFound == "", L(c).Empty, notFound)).FontSize(t.FontSize).TextColor(t.TextDisabled)
		})
		return "", false
	}
	row := func(c *ui.Context, i int) {
		r := rows[i]
		if r.header != "" {
			ui.Row(c).FillWidth().Height(selectRowH).Padding(0, 12).AlignItems(ui.Center).Children(func() {
				ui.Text(c, r.header).FontSize(t.FontSizeSM).TextColor(t.TextTertiary).SingleLine()
			})
			return
		}
		o := r.opt
		on := chosen(o.Value)
		item := ui.ButtonBase(c).Label(o.Label).FillWidth().Height(selectRowH).Padding(0, pick[float32](hasHeaderBefore(rows, i), 24, 12)).
			Radius(4).AlignItems(ui.Center).Justify(ui.Start).Gap(t.SpaceSM).FocusRing(false).
			Disabled(o.Disabled).Cursor(pick(o.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
		switch {
		case o.Disabled:
		case on:
			item.Background(t.PrimaryBg)
		case i == active || item.Hovered():
			item.Background(t.FillTertiary)
		}
		if i == active {
			if desc.Valid() {
				desc.ActiveDescendant(item)
			}
		}
		item.Children(func() {
			txt := ui.Text(c, o.Label).FontSize(t.FontSize).TextColor(pick(o.Disabled, t.TextDisabled, t.Text)).SingleLine().Grow(1).MinWidth(0)
			if on {
				txt.FontWeight(600)
				if multiple {
					checkMark(c, t.Primary)
				}
			}
		})
		if item.Clicked() && !o.Disabled {
			picked, ok = o.Value, true
		}
	}
	popupList(c, len(rows), selectRowH, selectMaxH, active, row)
	return picked, ok
}

// comboEmpty is a dummy ComboboxBase option so an empty popup still
// paints (ComboboxBase hides a popup with no Item).
const comboEmpty = "\x00kit-empty"

// comboPanel is optionPanel on ComboboxBase: RoleComboBox, ActiveDescendant
// and arrow/Enter live on the input. It returns a chosen value.
func comboPanel(c *ui.Context, t *Tokens, cb ui.ComboboxParts, rows []optionRow, chosen func(string) bool, multiple, virtual, composing bool, notFound string, width float32) (string, bool) {
	_ = virtual
	picked, ok := "", false
	cb.Popup(func(panel ui.Element) {
		if composing {
			panel.AttachTo(cb.Input, ui.AnchorTopStart, ui.AnchorBottomStart)
			panel.Column().MarginY(4)
		}
		popupLook(panel.Padding(4).Gap(0).MinWidth(max(width, 120)), t)
		if width > 0 {
			panel.Width(width)
		}
		if len(rows) == 0 {
			cb.Item(comboEmpty).FillWidth().Height(selectRowH).Disabled(true).Children(func() {
				ui.Column(c).FillWidth().Padding(12).AlignItems(ui.Center).Gap(4).Children(func() {
					ui.Icon(c, iconEmpty).Size(32, 22).TextColor(t.EmptyIcon)
					ui.Text(c, pick(notFound == "", L(c).Empty, notFound)).FontSize(t.FontSize).TextColor(t.TextDisabled)
				})
			})
			return
		}
		row := func(c *ui.Context, i int) {
			r := rows[i]
			if r.header != "" {
				ui.Row(c).FillWidth().Height(selectRowH).Padding(0, 12).AlignItems(ui.Center).Children(func() {
					ui.Text(c, r.header).FontSize(t.FontSizeSM).TextColor(t.TextTertiary).SingleLine()
				})
				return
			}
			o := r.opt
			on := chosen(o.Value)
			item := cb.Item(o.Value).FillWidth().Height(selectRowH).Padding(0, pick[float32](hasHeaderBefore(rows, i), 24, 12)).
				Radius(4).AlignItems(ui.Center).Justify(ui.Start).Gap(t.SpaceSM).FocusRing(false).
				Disabled(o.Disabled).Cursor(pick(o.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
			switch {
			case o.Disabled:
			case on:
				item.Background(t.PrimaryBg)
			case item.Highlighted() || item.Hovered():
				item.Background(t.FillTertiary)
			}
			item.Children(func() {
				txt := ui.Text(c, o.Label).FontSize(t.FontSize).TextColor(pick(o.Disabled, t.TextDisabled, t.Text)).SingleLine().Grow(1).MinWidth(0)
				if on {
					txt.FontWeight(600)
					if multiple {
						checkMark(c, t.Primary)
					}
				}
			})
			if item.Clicked() && !o.Disabled {
				picked, ok = o.Value, true
			}
		}
		popupList(c, len(rows), selectRowH, selectMaxH, -1, row)
	})
	if v, yes := cb.Chosen(); yes && v != comboEmpty {
		return v, true
	}
	return picked, ok
}

func hasHeaderBefore(rows []optionRow, i int) bool {
	for j := i - 1; j >= 0; j-- {
		if rows[j].header != "" {
			return true
		}
	}
	return false
}

// checkMark is the tick beside a chosen option of a multiple select.
func checkMark(c *ui.Context, tone ui.Color) {
	ui.Box(c).Size(14, 14).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		var path ui.Path
		path.MoveTo(r.X+r.W*0.18, r.Y+r.H*0.52).LineTo(r.X+r.W*0.42, r.Y+r.H*0.74).LineTo(r.X+r.W*0.84, r.Y+r.H*0.28)
		p.StrokePath(path.Cap(ui.RoundCap), 1.6, tone)
	})
}

// selectTag is a chip of a multiple select's trigger, with a × that
// reports a click when closable.
func selectTag(c *ui.Context, t *Tokens, label string, closable, small bool) bool {
	removed := false
	h := pick[float32](small, 16, 24)
	ui.Row(c).Height(h).Padding(0, 4, 0, 8).Gap(4).Radius(4).AlignItems(ui.Center).Shrink(0).
		Background(t.FillSecondary).Children(func() {
		ui.Text(c, label).FontSize(pick(small, t.FontSizeSM, t.FontSize)).TextColor(t.Text).SingleLine()
		if closable {
			x := ui.ButtonBase(c).Label(L(c).Remove+" "+label).Size(12, 12).FocusRing(false).Cursor(ui.CursorPointer)
			x.Children(func() {
				ui.Icon(c, iconClose).Size(10, 10).TextColor(pick(x.Hovered(), t.Text, t.TextTertiary))
			})
			removed = x.Clicked()
		}
	})
	return removed
}

// keyPick is a key a select's search field took, for the next frame.
type keyPick struct {
	value    string
	set, pop bool
}

// selectCore is a select that is not just a native one: search, clear,
// groups, a virtual list, or several values. values holds the choice
// (one value at most unless multiple).
func selectCore(c *ui.Context, values *[]string, options []Option, p SelectProps, multiple bool) bool {
	t := Of(c)
	changed := false
	h := t.Height(p.Size)
	slot := compactBegin()
	defer slot.done()
	box := ui.Row(c).MinHeight(h).Padding(t.SpaceXS, t.PaddingInput).Gap(t.SpaceSM).AlignItems(ui.Center).
		Radius(t.RadiusControl).Disabled(p.Disabled).Clip().Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
	slot.at(box)
	if p.Width > 0 {
		box.Width(p.Width).Shrink(0)
	} else {
		box.FillWidth()
	}
	if p.AutoFocus {
		box.AutoFocus()
	}
	open := ui.Local(box, "open", func() bool { return p.open })
	query := ui.Local(box, "query", func() string { return "" })
	active := ui.Local(box, "active", func() int { return -1 })
	rowsFor := func(q string) []optionRow {
		return optionRows(shownOptions(p, options, *values, q))
	}
	rows := rowsFor(*query)
	if *active >= len(rows) {
		*active = -1
	}
	choose := func(v string) {
		if multiple {
			*values = toggleValue(append([]string(nil), *values...), v)
		} else {
			*values = []string{v}
			*open = false
		}
		*query = ""
		changed = true
	}
	// OnTextKey runs after the frame is built, so the keys it takes are
	// done in the next.
	pending := ui.Local(box, "pending", func() keyPick { return keyPick{} })
	switch kp := *pending; {
	case kp.set:
		choose(kp.value)
	case kp.pop && len(*values) > 0:
		*values = append([]string(nil), (*values)[:len(*values)-1]...)
		changed = true
	}
	*pending = keyPick{}
	labelOfValue := func(v string) string { return labelOf(options, v) }
	var in ui.Element
	var cb ui.ComboboxParts
	searching := p.ShowSearch || p.Mode == Tags
	clicked := false
	if !searching {
		box.Role(ui.RoleComboBox)
	}
	box.Children(func() {
		ui.Row(c).Grow(1).MinWidth(0).Gap(4).Wrap().AlignItems(ui.Center).Padding(pick[float32](multiple, 3, 0), 0).Children(func() {
			if multiple {
				n := len(*values)
				limit := n
				if p.MaxTagCount > 0 && n > p.MaxTagCount {
					limit = p.MaxTagCount
				}
				for i := 0; i < limit; i++ {
					v := (*values)[i]
					if selectTag(c.Key(v), t, labelOfValue(v), !p.Disabled, p.Size == Small) {
						*values = toggleValue(append([]string(nil), *values...), v)
						changed = true
					}
				}
				if limit < n {
					selectTag(c, t, "+ "+strconv.Itoa(n-limit)+" ...", false, p.Size == Small)
				}
			}
			single := ""
			if !multiple && len(*values) > 0 {
				single = labelOfValue((*values)[0])
			}
			if searching && !p.Disabled {
				ui.Box(c).Key("combo").Grow(1).MinWidth(pick[float32](multiple, 24, 0)).Children(func() {
					cb = ui.ComboboxBase(c, query)
					in = cb.Input
					in.FontSize(t.ControlFont(p.Size)).TextColor(t.Text).FocusRing(false).
						Grow(1).MinWidth(pick[float32](multiple, 24, 0)).Height(pick(multiple, h-8, h-2))
					ph := p.Placeholder
					switch {
					case !multiple && single != "":
						ph = single
						in.PlaceholderColor(pick(*open, t.TextPlaceholder, t.Text))
					case multiple && len(*values) > 0:
						ph = ""
					default:
						in.PlaceholderColor(t.TextPlaceholder)
					}
					in.Placeholder(ph)
					if in.Changed() {
						*open = true
						*active = stepActive(rowsFor(*query), -1, 1)
					}
					in.OnTextKey(func(k ui.TextKey) bool {
						if k.Key == ui.KeyBackspace && multiple && k.Text() == "" {
							*pending = keyPick{pop: true}
							return true
						}
						return false
					})
				})
			} else if !multiple || len(*values) == 0 {
				text, fg := single, pick(p.Disabled, t.TextDisabled, t.Text)
				if text == "" {
					text, fg = p.Placeholder, t.TextPlaceholder
				}
				ui.Text(c, text).FontSize(t.ControlFont(p.Size)).TextColor(fg).SingleLine().Grow(1).MinWidth(0)
			}
		})
		lit := box.Hovered() || box.FocusWithin()
		fieldStatusIcon(c, t, pick(p.NoFeedback, StatusNone, p.Status))
		if p.AllowClear && !p.Disabled && len(*values) > 0 && lit {
			if clearButton(c, t, true) {
				*values = nil
				*query = ""
				changed = true
			}
		} else if p.Loading {
			ui.Spinner(c).Size(12, 12).TextColor(t.TextDisabled)
		} else if searching && *open {
			ui.Icon(c, iconSearch).Size(12, 12).TextColor(t.TextDisabled).Shrink(0)
		} else {
			chevron(c, t.TextDisabled, false)
		}
	})
	if searching {
		takeLabelFocus(in)
	} else {
		takeLabelFocus(box)
	}
	if !searching {
		box.Focusable()
		if box.Shortcut(0, ui.KeyDown) || box.Shortcut(0, ui.KeyUp) {
			down := !box.Shortcut(0, ui.KeyUp)
			*open = true
			*active = stepActive(rows, *active, pick(down, 1, -1))
		}
		if *open && *active >= 0 && box.Shortcut(0, ui.KeyEnter) && rows[*active].header == "" {
			choose(rows[*active].opt.Value)
		} else if box.Clicked() && !p.Disabled {
			clicked = true
		}
	} else if box.Clicked() && !p.Disabled {
		clicked = true
		if in.Valid() {
			in.Focus()
			cb.SetOpen(true)
		}
	}
	if clicked && !searching {
		*open = !*open
	}
	if searching && in.Valid() {
		if p.open {
			cb.SetOpen(true)
		}
		was := ui.Local(box, "focused", func() bool { return false })
		if in.Focused() && !*was && !p.Disabled {
			cb.SetOpen(true)
		}
		*was = in.Focused()
		*open = cb.Open()
	}
	st := stateOf(box, p.Disabled, p.force)
	if (*open || (in.Valid() && in.FocusVisible())) && st != disabled {
		st = focus
	}
	slot.field(box, t, st, p.Status, fieldLookV(box, t, st, p.Status, p.Variant))
	width := box.Bounds().W
	if p.Disabled {
		*open = false
		if searching && in.Valid() {
			cb.SetOpen(false)
		}
	}
	composing := in.Valid() && in.Composing()
	if searching && in.Valid() && composing {
		shown := cb.Open()
		ui.PopoverBase(c, box, &shown, func(wrap ui.Element) {
			placePanel(c, wrap, box, panelPlace{Side: PlaceTop, Align: AlignFirst, Gap: 4}, t.Elevated, func(panel ui.Element) {
				popupLook(panel.Padding(4).Gap(0).MinWidth(max(width, 120)), t).KeepFocus()
				if width > 0 {
					panel.Width(width)
				}
				panel.Children(func() {
					chosen := func(v string) bool { return hasValue(*values, v) }
					if v, ok := optionPanel(c, t, rows, chosen, -1, multiple, p.Virtual || len(rows) > virtualOver, p.NotFound, in); ok {
						choose(v)
					}
				})
			})
		})
		*open = shown
	} else if searching && in.Valid() {
		if v, ok := comboPanel(c, t, cb, rows, func(v string) bool { return hasValue(*values, v) }, multiple, p.Virtual || len(rows) > virtualOver, false, p.NotFound, width); ok {
			choose(v)
			if multiple {
				cb.SetOpen(true)
			}
		}
		*open = cb.Open()
	} else {
		ui.PopoverBase(c, box, open, func(wrap ui.Element) {
			wrap.Opacity(enterIn(wrap, "in", motionFor(c, 120*time.Millisecond)))
			side := imePlacement(PlaceBottom, composing)
			placePanel(c, wrap, box, panelPlace{Side: side, Align: AlignFirst, Gap: 4}, t.Elevated, func(panel ui.Element) {
				popupLook(panel.Padding(4).Gap(0).MinWidth(max(width, 120)), t).KeepFocus()
				if width > 0 {
					panel.Width(width)
				}
				panel.Children(func() {
					chosen := func(v string) bool { return hasValue(*values, v) }
					if v, ok := optionPanel(c, t, rows, chosen, *active, multiple, p.Virtual || len(rows) > virtualOver, p.NotFound, box); ok {
						choose(v)
					}
				})
			})
		})
	}
	if !*open {
		*active = -1
		if !in.Valid() || !in.Focused() {
			*query = ""
		}
	}
	return changed
}

// shownOptions is the options a select shows for query: those matching
// it when searching, and in Tags mode the values typed before and query
// itself, first, when no option is called that.
func shownOptions(p SelectProps, options []Option, values []string, query string) []Option {
	if !p.ShowSearch && p.Mode != Tags {
		return options
	}
	shown := filterOptions(options, query)
	if p.Mode != Tags {
		return shown
	}
	shown = append([]Option(nil), shown...)
	for _, v := range values {
		if indexOf(options, v) < 0 && (query == "" || strings.Contains(strings.ToLower(v), strings.ToLower(query))) {
			shown = append(shown, Option{Value: v, Label: v})
		}
	}
	if q := strings.TrimSpace(query); q != "" && indexOf(shown, q) < 0 {
		shown = append([]Option{{Value: q, Label: q}}, shown...)
	}
	return shown
}

// MultiSelect chooses *values among options in an Ant select of mode
// multiple (Tags lets the user type new values too) and reports a change.
// The chosen values show as chips, MaxTagCount of them at most.
func MultiSelect(c *ui.Context, values *[]string, options []Option, props ...SelectProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	if p.Mode == Single {
		p.Mode = Multiple
	}
	return selectCore(c, values, options, p, true)
}
