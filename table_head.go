package kit

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// sortHeader is the title of a sortable column: a button over its cell
// (pad is the cell's side padding it reaches into) with Ant's stacked
// carets, the one of order lit while the column is sorted.
func sortHeader(c *ui.Context, t *Tokens, col Column, sorted bool, order SortOrder, pad float32) bool {
	h := ui.ButtonBase(c).Label(col.Title).Row().Gap(t.SpaceXS).AlignItems(ui.Center).AlignSelf(ui.Stretch).
		Grow(1).Shrink(1).MinWidth(0).Margin(0, 0, 0, -pad).Padding(0, 0, 0, pad).FocusRing(false).Cursor(ui.CursorPointer).
		Justify(justify(col.Align))
	if len(col.Filters) == 0 {
		h.Margin(0, -pad).Padding(0, pad)
	}
	switch {
	case h.Pressed():
		h.Background(t.Fill.Over(t.FillAlter))
	case h.Hovered():
		h.Background(t.BorderSecondary)
	}
	if h.FocusVisible() {
		h.InsetShadow(0, 0, 0, 2, t.PrimaryBorder)
	}
	up, down := t.TextDisabled, t.TextDisabled
	if sorted {
		up = pick(order == SortAscend, t.Primary, up)
		down = pick(order == SortDescend, t.Primary, down)
	}
	h.Children(func() {
		ui.Text(c, col.Title).FontSize(t.FontSize).FontWeight(600).TextColor(t.Text).SingleLine().Shrink(1)
		ui.Box(c).Size(10, 14).Shrink(0).Draw(func(pa *ui.Painter, r ui.Rect) {
			var a, b ui.Path
			cx := r.X + r.W/2
			a.MoveTo(cx-4, r.Y+6).LineTo(cx+4, r.Y+6).LineTo(cx, r.Y+1.5).Close()
			b.MoveTo(cx-4, r.Y+8).LineTo(cx+4, r.Y+8).LineTo(cx, r.Y+12.5).Close()
			pa.FillPath(&a, up)
			pa.FillPath(&b, down)
		})
	})
	return h.Clicked()
}

// filterMenu is a column's funnel and the dropdown it opens: a search
// box (FilterSearch), the filters as a checklist (or radio list), 重置 and
// 确定. It reports the values checked when one of those is pressed.
func filterMenu(c *ui.Context, t *Tokens, col Column, vals []string) ([]string, bool) {
	var out []string
	ok := false
	Popup(c, func() bool {
		b := ui.ButtonBase(c).Label(L(c).Filter+" "+col.Title).Size(24, 24).Radius(4).Center().FocusRing(false).
			Cursor(ui.CursorPointer).Margin(0, -4, 0, 4)
		if b.Hovered() || b.Pressed() {
			b.Background(t.Fill)
		}
		if b.FocusVisible() {
			focusRing(b, t)
		}
		tone := pick(len(vals) > 0, t.Primary, pick(b.Hovered(), t.TextSecondary, t.TextTertiary))
		b.Draw(func(pa *ui.Painter, r ui.Rect) {
			x, y := r.X+r.W/2-6, r.Y+r.H/2-6
			var f ui.Path
			f.MoveTo(x+0.5, y+1.5).LineTo(x+11.5, y+1.5).LineTo(x+7.2, y+6.6).LineTo(x+7.2, y+10.6).
				LineTo(x+4.8, y+11.6).LineTo(x+4.8, y+6.6).Close()
			pa.FillPath(&f, tone)
		})
		return b.Clicked()
	}, func(panel ui.Element, close func()) {
		popupLook(panel.Padding(0).Gap(0).MinWidth(140), t)
		draft := ui.Local(panel, "draft", func() []string { return append([]string(nil), vals...) })
		query := ui.Local(panel, "query", func() string { return "" })
		panel.Children(func() {
			if col.FilterSearch {
				ui.Column(c).FillWidth().Padding(8, 8, 4).Children(func() {
					Input(c, query, InputProps{Size: Small, PrefixIcon: iconSearch, Placeholder: L(c).FilterSearch, AllowClear: true})
				})
			}
			shown := filterOptions(col.Filters, *query)
			ui.Column(c).FillWidth().Padding(4).Children(func() {
				if len(shown) == 0 {
					ui.Row(c).FillWidth().Height(32).Padding(0, 8).AlignItems(ui.Center).Children(func() {
						ui.Text(c, L(c).FilterEmpty).FontSize(t.FontSize).TextColor(t.TextDisabled)
					})
				}
				for i, o := range shown {
					item := ui.Row(c.Key(i)).FillWidth().Height(32).Padding(0, 8).Radius(4).AlignItems(ui.Center)
					if item.Hovered() {
						item.Background(t.FillTertiary)
					}
					item.Children(func() {
						if col.FilterSingle {
							cur := ""
							if len(*draft) > 0 {
								cur = (*draft)[0]
							}
							if Radio(c, &cur, o.Value, o.Label, ChoiceProps{Disabled: o.Disabled}) {
								*draft = []string{cur}
							}
							return
						}
						on := hasValue(*draft, o.Value)
						if Checkbox(c, &on, o.Label, ChoiceProps{Disabled: o.Disabled}) {
							*draft = toggleValue(append([]string(nil), *draft...), o.Value)
						}
					})
				}
			})
			ui.Row(c).FillWidth().Padding(7, 8).Justify(ui.SpaceBetween).AlignItems(ui.Center).Draw(topRule(t.BorderSecondary)).Children(func() {
				if Button(c, L(c).Reset, ButtonProps{Type: Link, Size: Small, Disabled: len(*draft) == 0 && len(vals) == 0}) {
					out, ok = nil, true
					close()
				}
				if Button(c, L(c).OK, ButtonProps{Type: Primary, Size: Small}) {
					out, ok = orderedValues(col.Filters, *draft), true
					close()
				}
			})
		})
	}, PopupProps{Placement: PlaceBottom, Align: AlignLast})
	return out, ok
}

// FilterEquals is an OnFilter keeping the rows whose key is the value.
func FilterEquals(key func(row int) string) func(row int, value string) bool {
	return func(row int, value string) bool { return key(row) == value }
}

// FilterContains is an OnFilter keeping the rows whose key holds the
// value, ignoring case.
func FilterContains(key func(row int) string) func(row int, value string) bool {
	return func(row int, value string) bool {
		return strings.Contains(strings.ToLower(key(row)), strings.ToLower(value))
	}
}

func justify(a Align) ui.Align {
	switch a {
	case AlignCenter:
		return ui.Center
	case AlignEnd:
		return ui.End
	}
	return ui.Start
}

// bottomRule draws a 1px split line along the bottom edge.
func bottomRule(split ui.Color) func(*ui.Painter, ui.Rect) {
	return func(p *ui.Painter, r ui.Rect) {
		p.Fill(ui.Rect{X: r.X, Y: r.Y + r.H - 1, W: r.W, H: 1}, split, 0)
	}
}

// headRule is the short divider between header cells (Ant's
// headerSplitColor, 1.6em tall).
func headRule(split ui.Color) func(*ui.Painter, ui.Rect) {
	return func(p *ui.Painter, r ui.Rect) {
		h := min(r.H*0.55, 22)
		p.Fill(ui.Rect{X: r.X + r.W - 1, Y: r.Y + (r.H-h)/2, W: 1, H: h}, split, 0)
	}
}
