package kit

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// chip is a value a picker's trigger shows as a tag.
type chip struct{ key, label string }

// chipField is the trigger of TreeSelect and Cascader: a select box
// with the choice as text or chips, a search field when searching, and a
// clear button. It returns the box, the search field (when searching),
// the chip whose × was clicked, and whether the clear button was.
type chipField struct {
	box, in ui.Element
	removed string
	cleared bool
}

func buildChipField(c *ui.Context, t *Tokens, p SelectProps, open *bool, query *string, single string, chips []chip, multiple, has bool, slot *compactSlot) chipField {
	var f chipField
	h := t.Height(p.Size)
	box := ui.Row(c).MinHeight(h).Padding(t.SpaceXS, t.PaddingInput).Gap(t.SpaceSM).AlignItems(ui.Center).
		Radius(t.RadiusControl).Disabled(p.Disabled).Clip().
		Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer)).Focusable()
	if p.Width > 0 {
		box.Width(p.Width).Shrink(0)
	} else {
		box.FillWidth()
	}
	f.box = box
	box.Children(func() {
		ui.Row(c).Grow(1).MinWidth(0).Gap(4).Wrap().AlignItems(ui.Center).Padding(pick[float32](multiple, 3, 0), 0).Children(func() {
			if multiple {
				limit := len(chips)
				if p.MaxTagCount > 0 && limit > p.MaxTagCount {
					limit = p.MaxTagCount
				}
				for _, ch := range chips[:limit] {
					if selectTag(c.Key(ch.key), t, ch.label, !p.Disabled, p.Size == Small) {
						f.removed = ch.key
					}
				}
				if limit < len(chips) {
					selectTag(c, t, "+ "+strconv.Itoa(len(chips)-limit)+" ...", false, p.Size == Small)
				}
			}
			if p.ShowSearch && !p.Disabled {
				in := ui.TextInputBase(c, query).FontSize(t.ControlFont(p.Size)).TextColor(t.Text).FocusRing(false).
					Grow(1).MinWidth(pick[float32](multiple, 24, 0)).Height(pick(multiple, h-8, h-2))
				switch {
				case !multiple && single != "":
					in.Placeholder(single).PlaceholderColor(pick(*open, t.TextPlaceholder, t.Text))
				case multiple && len(chips) > 0:
				default:
					in.Placeholder(p.Placeholder).PlaceholderColor(t.TextPlaceholder)
				}
				if in.Changed() {
					*open = true
				}
				f.in = in
				return
			}
			if !multiple || len(chips) == 0 {
				text, fg := single, pick(p.Disabled, t.TextDisabled, t.Text)
				if text == "" {
					text, fg = p.Placeholder, t.TextPlaceholder
				}
				ui.Text(c, text).FontSize(t.ControlFont(p.Size)).TextColor(fg).SingleLine().Grow(1).MinWidth(0)
			}
		})
		if p.AllowClear && !p.Disabled && has && (box.Hovered() || box.FocusWithin()) {
			f.cleared = clearButton(c, t, true)
		} else {
			chevron(c, t.TextDisabled, false)
		}
	})
	if f.in.Valid() {
		was := ui.Local(box, "focused", func() bool { return false })
		if f.in.Focused() && !*was {
			*open = true
		}
		*was = f.in.Focused()
	} else if box.Clicked() && !p.Disabled && f.removed == "" && !f.cleared {
		*open = !*open
	}
	if p.Disabled {
		*open = false
	}
	st := stateOf(box, p.Disabled, p.force)
	if *open && st != disabled {
		st = focus
	}
	slot.field(box, t, st, p.Status, fieldLookV(box, t, st, p.Status, p.Variant))
	return f
}

// TreeSelectProps are a tree select's options.
type TreeSelectProps struct {
	SelectProps
	// Checkable puts checkboxes in the tree, conducted; TreeMultiSelect
	// shows the checked leaves.
	Checkable bool
	// DefaultExpandAll opens the whole tree at first.
	DefaultExpandAll bool
	ShowLine         bool
	// LoadData is asked for the children of a node without any that is
	// not IsLeaf, once, as it opens (Ant's loadData). Mutate the tree
	// in place; the next frame shows the children.
	LoadData func(key string)
}

// TreeSelect is Ant's TreeSelect: a select whose popup is a Tree,
// choosing one node into *value. It reports a change.
func TreeSelect(c *ui.Context, value *string, nodes []TreeNode, props ...TreeSelectProps) bool {
	p := first(props)
	t := Of(c)
	changed := false
	slot := compactBegin()
	defer slot.done()
	anchor := ui.Column(c).AlignSelf(ui.Stretch)
	slot.at(anchor)
	if p.Width > 0 {
		anchor.AlignSelf(ui.Start)
	}
	anchor.Children(func() {
		open := ui.Local(anchor, "open", func() bool { return p.open })
		query := ui.Local(anchor, "query", func() string { return "" })
		idx := treeIndex(nodes)
		label := ""
		if in, ok := idx[*value]; ok {
			label = in.node.Title
		}
		f := buildChipField(c, t, p.SelectProps, open, query, label, nil, false, *value != "", slot)
		if f.cleared {
			*value = ""
			changed = true
		}
		treePopup(c, t, f.box, open, func() {
			sel := *value
			if Tree(c, nodes, TreeProps{Selected: &sel, DefaultExpandAll: p.DefaultExpandAll, ShowLine: p.ShowLine,
				Filter: *query, BlockNode: true, LoadData: p.LoadData}) {
				*value = sel
				*open = false
				*query = ""
				changed = true
			}
		})
	})
	return changed
}

// TreeMultiSelect is Ant's TreeSelect with multiple or treeCheckable:
// several nodes into *values, as chips. With Checkable the check is
// conducted, *values holds every checked key and the chips show the
// checked leaves. It reports a change.
func TreeMultiSelect(c *ui.Context, values *[]string, nodes []TreeNode, props ...TreeSelectProps) bool {
	p := first(props)
	t := Of(c)
	changed := false
	slot := compactBegin()
	defer slot.done()
	anchor := ui.Column(c).AlignSelf(ui.Stretch)
	slot.at(anchor)
	if p.Width > 0 {
		anchor.AlignSelf(ui.Start)
	}
	anchor.Children(func() {
		open := ui.Local(anchor, "open", func() bool { return p.open })
		query := ui.Local(anchor, "query", func() string { return "" })
		idx := treeIndex(nodes)
		shown := *values
		if p.Checkable {
			shown = leafKeys(nodes, conducted(nodes, *values))
		}
		var chips []chip
		for _, k := range shown {
			if in, ok := idx[k]; ok {
				chips = append(chips, chip{key: k, label: in.node.Title})
			}
		}
		f := buildChipField(c, t, p.SelectProps, open, query, "", chips, true, len(*values) > 0, slot)
		switch {
		case f.cleared:
			*values = nil
			changed = true
		case f.removed != "" && p.Checkable:
			*values = conductCheck(nodes, *values, f.removed, false)
			changed = true
		case f.removed != "":
			*values = toggleValue(append([]string(nil), *values...), f.removed)
			changed = true
		}
		treePopup(c, t, f.box, open, func() {
			if p.Checkable {
				if Tree(c, nodes, TreeProps{Checkable: true, Checked: values, DefaultExpandAll: p.DefaultExpandAll,
					ShowLine: p.ShowLine, Filter: *query, BlockNode: true, LoadData: p.LoadData}) {
					changed = true
				}
				return
			}
			if Tree(c, nodes, TreeProps{MultiSelected: values, DefaultExpandAll: p.DefaultExpandAll, ShowLine: p.ShowLine,
				Filter: *query, BlockNode: true, LoadData: p.LoadData}) {
				changed = true
			}
		})
	})
	return changed
}

// treePopup is the popup under a picker's box holding what fn builds.
func treePopup(c *ui.Context, t *Tokens, box ui.Element, open *bool, fn func()) {
	width := box.Bounds().W
	ui.PopoverBase(c, box, open, func(wrap ui.Element) {
		wrap.Opacity(enterIn(wrap, "in", motionFor(c, 120*time.Millisecond)))
		placePanel(c, wrap, box, panelPlace{Side: PlaceBottom, Align: AlignFirst, Gap: 4}, t.Elevated, func(panel ui.Element) {
			popupLook(panel.Padding(4).Gap(0).MinWidth(max(width, 160)), t).KeepFocus()
			panel.Children(func() {
				ui.Scroll(c).FillWidth().MaxHeight(selectMaxH).Children(func() {
					ui.Column(c).FillWidth().Children(fn)
				})
			})
		})
	})
}
