package kit

import (
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// TransferItem is an item of a Transfer.
type TransferItem struct {
	Key, Title, Description string
	Disabled                bool
}

// TransferProps are a transfer's options.
type TransferProps struct {
	// Titles head the source and target lists.
	Titles [2]string
	// ShowSearch puts a search box over each list.
	ShowSearch bool
	// OneWay has items go only to the target, where a × takes each back.
	OneWay bool
	// Width and Height are each list's; 200 × 200 when zero.
	Width, Height float32
	Disabled      bool
}

// transferSplit is the items not in target, in data's order, and those
// in it, in target's order.
func transferSplit(data []TransferItem, target []string) (left, right []TransferItem) {
	in := map[string]bool{}
	for _, k := range target {
		in[k] = true
	}
	byKey := map[string]TransferItem{}
	for _, d := range data {
		byKey[d.Key] = d
		if !in[d.Key] {
			left = append(left, d)
		}
	}
	for _, k := range target {
		if d, ok := byKey[k]; ok {
			right = append(right, d)
		}
	}
	return
}

// transferMove is target with keys added at its end (toRight) or taken
// out.
func transferMove(target, keys []string, toRight bool) []string {
	out := append([]string(nil), target...)
	for _, k := range keys {
		has := hasValue(out, k)
		switch {
		case toRight && !has:
			out = append(out, k)
		case !toRight && has:
			out = toggleValue(out, k)
		}
	}
	return out
}

// filterItems keeps the items whose title or description holds q.
func filterItems(items []TransferItem, q string) []TransferItem {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return items
	}
	var out []TransferItem
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Title), q) || strings.Contains(strings.ToLower(it.Description), q) {
			out = append(out, it)
		}
	}
	return out
}

// Transfer is Ant's Transfer: two lists, the items not in *target and
// those in it, with checkboxes, a check-all with its count, search, and
// buttons moving the checked items across. It reports a change.
func Transfer(c *ui.Context, data []TransferItem, target *[]string, props ...TransferProps) bool {
	p := first(props)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	changed := false
	w := pick(p.Width > 0, p.Width, 200)
	h := pick(p.Height > 0, p.Height, 200)
	row := ui.Row(c).Gap(t.SpaceSM).AlignItems(ui.Center).AlignSelf(ui.Start)
	checks := ui.Local(row, "checks", func() [2][]string { return [2][]string{} })
	left, right := transferSplit(data, *target)
	for side, items := range [2][]TransferItem{left, right} {
		var keep []string
		for _, k := range checks[side] {
			for _, it := range items {
				if it.Key == k && !it.Disabled {
					keep = append(keep, k)
				}
			}
		}
		checks[side] = keep
	}
	row.Children(func() {
		transferList(c.Key("l"), t, left, &checks[0], p.Titles[0], p, w, h, false, nil)
		ui.Column(c).Gap(4).Children(func() {
			if transferButton(c, t, ">", L(c).MoveRight, len(checks[0]) == 0 || p.Disabled) {
				*target = transferMove(*target, checks[0], true)
				checks[0] = nil
				changed = true
			}
			if !p.OneWay && transferButton(c, t, "<", L(c).MoveLeft, len(checks[1]) == 0 || p.Disabled) {
				*target = transferMove(*target, checks[1], false)
				checks[1] = nil
				changed = true
			}
		})
		var removed []string
		transferList(c.Key("r"), t, right, &checks[1], p.Titles[1], p, w, h, p.OneWay, &removed)
		if len(removed) > 0 {
			*target = transferMove(*target, removed, false)
			changed = true
		}
	})
	return changed
}

// transferButton is a small square arrow button between the lists.
func transferButton(c *ui.Context, t *Tokens, glyph, label string, off bool) bool {
	b := ui.ButtonBase(c).Label(label).Size(24, 24).Radius(t.RadiusButton).Center().FocusRing(false).Disabled(off).
		Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
	st := stateOf(b, off, rest)
	bg, fg, edge := buttonColors(t, ButtonProps{Type: Primary, Size: Small}, st)
	if off {
		bg, fg, edge = t.DisabledBg, t.TextDisabled, t.Border
	}
	b.Background(bg).Border(1, edge)
	b.Children(func() {
		ui.Box(c).Size(12, 12).Rotate(pick[float32](glyph == ">", -90, 90)).Children(func() { chevron(c, fg, false) })
	})
	return b.Clicked() && !off
}

// transferList is one side of a Transfer.
func transferList(c *ui.Context, t *Tokens, items []TransferItem, checked *[]string, title string, p TransferProps, w, h float32, removable bool, removed *[]string) {
	box := ui.Column(c).Width(w).Height(h).Border(1, t.Border).Radius(t.RadiusControl).Shrink(0)
	query := ui.Local(box, "query", func() string { return "" })
	shown := filterItems(items, *query)
	opts := make([]Option, len(shown))
	for i, it := range shown {
		opts[i] = Option{Value: it.Key, Label: it.Title, Disabled: it.Disabled}
	}
	box.Children(func() {
		ui.Row(c).FillWidth().Height(40).Padding(0, 12).Gap(t.SpaceSM).AlignItems(ui.Center).Shrink(0).
			Draw(bottomRule(t.Split)).Children(func() {
			if !removable {
				ui.Row(c).Shrink(0).Children(func() { CheckAll(c, checked, opts, "", ChoiceProps{Disabled: p.Disabled}) })
			}
			count := L(c).Items(len(items))
			if len(*checked) > 0 {
				count = strconv.Itoa(len(*checked)) + "/" + count
			}
			ui.Text(c, count).FontSize(t.FontSize).TextColor(t.Text).Grow(1).SingleLine()
			ui.Text(c, title).FontSize(t.FontSize).TextColor(t.Text).SingleLine()
		})
		if p.ShowSearch {
			ui.Column(c).FillWidth().Padding(8, 8, 4).Shrink(0).Children(func() {
				Input(c, query, InputProps{Size: Small, Placeholder: L(c).SearchHere, PrefixIcon: iconSearch, AllowClear: true})
			})
		}
		if len(shown) == 0 {
			ui.Column(c).FillWidth().Grow(1).Center().Children(func() { Empty(c, L(c).Empty) })
			return
		}
		VirtualList(c, len(shown), func(c *ui.Context, i int) {
			it := shown[i]
			r := ui.ButtonBase(c).Label(it.Title).FillWidth().Height(32).Padding(0, 12).Gap(t.SpaceSM).AlignItems(ui.Center).
				Justify(ui.Start).FocusRing(false).Disabled(it.Disabled || p.Disabled).
				Cursor(pick(it.Disabled || p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
			on := hasValue(*checked, it.Key)
			if r.Hovered() && !it.Disabled {
				r.Background(t.FillTertiary)
			}
			r.Children(func() {
				if !removable {
					checkGlyph(c, t, on, false, it.Disabled || p.Disabled, r.Hovered())
				}
				ui.Text(c, it.Title).FontSize(t.FontSize).TextColor(pick(it.Disabled, t.TextDisabled, t.Text)).SingleLine().Grow(1).MinWidth(0)
				if it.Description != "" {
					ui.Text(c, it.Description).FontSize(t.FontSizeSM).TextColor(t.TextTertiary).SingleLine()
				}
				if removable && !p.Disabled {
					x := ui.ButtonBase(c).Label(L(c).Remove+" "+it.Title).Size(16, 16).Center().FocusRing(false).Cursor(ui.CursorPointer)
					x.Children(func() { ui.Icon(c, iconClose).Size(12, 12).TextColor(pick(x.Hovered(), t.Error, t.TextTertiary)) })
					if x.Clicked() {
						*removed = append(*removed, it.Key)
					}
				}
			})
			if r.Clicked() && !it.Disabled && !p.Disabled && !removable {
				*checked = orderedValues(opts, toggleValue(append([]string(nil), *checked...), it.Key))
			}
		})
	})
}
