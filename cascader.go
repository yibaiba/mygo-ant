package kit

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// CascaderProps are a cascader's options.
type CascaderProps struct {
	SelectProps
	// ChangeOnSelect takes the path to any node clicked, not only to a
	// leaf, as Ant's changeOnSelect.
	ChangeOnSelect bool
	// HoverOpen opens a node's column as the pointer rests on it (Ant's
	// expandTrigger hover) rather than on a click.
	HoverOpen bool
	// Separator joins the labels of the path shown; " / " when empty.
	Separator string
	// LoadData is asked for the children of a node without any that is
	// not IsLeaf, once, as it opens, with the path of options to it
	// (Ant's loadData(selectedOptions)). Mutate the tree in place and
	// the next frame shows the new column.
	LoadData func(selected []TreeNode)
}

// cascaderColumns is the columns a cascader shows with the nodes of
// active open: the options, then the children of each active node.
func cascaderColumns(options []TreeNode, active []string) [][]TreeNode {
	cols := [][]TreeNode{options}
	level := options
	for _, k := range active {
		found := false
		for i := range level {
			if level[i].Key == k {
				found = true
				if len(level[i].Children) > 0 {
					level = level[i].Children
					cols = append(cols, level)
				} else {
					return cols
				}
				break
			}
		}
		if !found {
			break
		}
	}
	return cols
}

// cascaderLabels is the titles of the nodes along path, as far as it
// leads.
func cascaderLabels(options []TreeNode, path []string) []string {
	var out []string
	level := options
	for _, k := range path {
		found := false
		for i := range level {
			if level[i].Key == k {
				out = append(out, level[i].Title)
				level = level[i].Children
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return out
}

// cascadePath is a path a cascader search found.
type cascadePath struct {
	keys, labels []string
	disabled     bool
}

// cascaderSearch is the paths to leaves (to every node with
// changeOnSelect) one of whose labels holds q, ignoring case.
func cascaderSearch(options []TreeNode, q string, changeOnSelect bool) []cascadePath {
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return nil
	}
	var out []cascadePath
	var walk func(ns []TreeNode, keys, labels []string, off, hit bool)
	walk = func(ns []TreeNode, keys, labels []string, off, hit bool) {
		for _, n := range ns {
			k := append(append([]string(nil), keys...), n.Key)
			l := append(append([]string(nil), labels...), n.Title)
			h := hit || strings.Contains(strings.ToLower(n.Title), q)
			d := off || n.Disabled
			if h && (len(n.Children) == 0 || changeOnSelect) {
				out = append(out, cascadePath{keys: k, labels: l, disabled: d})
			}
			walk(n.Children, k, l, d, h)
		}
	}
	walk(options, nil, nil, false, false)
	return out
}

// parentKeys keeps the checked keys whose parent is not checked: a
// conducted check shown by its highest checked nodes (Ant's SHOW_PARENT).
func parentKeys(nodes []TreeNode, checked []string) []string {
	idx := treeIndex(nodes)
	set := map[string]bool{}
	for _, k := range checked {
		set[k] = true
	}
	var out []string
	for _, k := range checked {
		if in, ok := idx[k]; ok && !set[in.parent] {
			out = append(out, k)
		}
	}
	return out
}

// pathNodes is the nodes along path, as Ant's selectedOptions.
func pathNodes(options []TreeNode, path []string) []TreeNode {
	var out []TreeNode
	level := options
	for _, k := range path {
		found := false
		for i := range level {
			if level[i].Key == k {
				out = append(out, level[i])
				level = level[i].Children
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return out
}

// pathTo is the keys from the root down to key.
func pathTo(nodes []TreeNode, key string) []string {
	idx := treeIndex(nodes)
	var out []string
	for k := key; k != ""; k = idx[k].parent {
		if _, ok := idx[k]; !ok {
			return nil
		}
		out = append([]string{k}, out...)
	}
	return out
}

// Cascader is Ant's Cascader: a select whose popup shows columns of
// options, each column the children of the node chosen in the one before,
// choosing a path into *path. It reports a change.
func Cascader(c *ui.Context, path *[]string, options []TreeNode, props ...CascaderProps) bool {
	return cascaderCore(c, path, nil, options, first(props))
}

// CascaderMultiple is Ant's Cascader multiple: checkboxes in the columns,
// conducted, *values holding every checked key and the chips its highest
// checked nodes. Keys must be unique across the tree. It reports a change.
func CascaderMultiple(c *ui.Context, values *[]string, options []TreeNode, props ...CascaderProps) bool {
	return cascaderCore(c, nil, values, options, first(props))
}

func cascaderCore(c *ui.Context, path *[]string, values *[]string, options []TreeNode, p CascaderProps) bool {
	t := Of(c)
	changed := false
	multiple := values != nil
	sep := pick(p.Separator == "", " / ", p.Separator)
	slot := compactBegin()
	defer slot.done()
	anchor := ui.Column(c).AlignSelf(pick(p.Width > 0, ui.Start, ui.Stretch))
	slot.at(anchor)
	anchor.Children(func() {
		open := ui.Local(anchor, "open", func() bool { return p.open })
		query := ui.Local(anchor, "query", func() string { return "" })
		active := ui.Local(anchor, "active", func() []string {
			if multiple {
				return nil
			}
			return append([]string(nil), *path...)
		})
		loading := ui.Local(anchor, "loading", func() map[string]bool { return map[string]bool{} })
		for k := range *loading {
			if in, ok := treeIndex(options)[k]; !ok || len(in.node.Children) > 0 {
				delete(*loading, k)
			}
		}
		var chips []chip
		single := ""
		has := false
		if multiple {
			for _, k := range parentKeys(options, conducted(options, *values)) {
				chips = append(chips, chip{key: k, label: strings.Join(cascaderLabels(options, pathTo(options, k)), sep)})
			}
			has = len(*values) > 0
		} else {
			single = strings.Join(cascaderLabels(options, *path), sep)
			has = len(*path) > 0
		}
		wasOpen := *open
		f := buildChipField(c, t, p.SelectProps, open, query, single, chips, multiple, has, slot)
		if *open && !wasOpen {
			if !multiple {
				*active = append([]string(nil), *path...)
			}
		}
		switch {
		case f.cleared && multiple:
			*values = nil
			changed = true
		case f.cleared:
			*path = nil
			changed = true
		case f.removed != "":
			*values = conductCheck(options, *values, f.removed, false)
			changed = true
		}
		width := f.box.Bounds().W
		ui.PopoverBase(c, f.box, open, func(wrap ui.Element) {
			wrap.Opacity(enterIn(wrap, "in", motionFor(c, 120*time.Millisecond)))
			placePanel(c, wrap, f.box, panelPlace{Side: PlaceBottom, Align: AlignFirst, Gap: 4}, t.Elevated, func(panel ui.Element) {
				popupLook(panel.Padding(4).Gap(0), t).KeepFocus()
				panel.Children(func() {
					if *query != "" {
						found := cascaderSearch(options, *query, p.ChangeOnSelect)
						if len(found) == 0 {
							optionPanel(c, t, nil, nil, -1, false, false, p.NotFound, ui.Element{})
							return
						}
						ui.Column(c).MinWidth(max(width, 160)).Children(func() {
							popupList(c, len(found), selectRowH, selectMaxH, -1, func(c *ui.Context, i int) {
								fp := found[i]
								label := strings.Join(fp.labels, sep)
								b := ui.ButtonBase(c).Label(label).FillWidth().Height(selectRowH).Padding(0, 12).Radius(4).
									AlignItems(ui.Center).Justify(ui.Start).FocusRing(false).Disabled(fp.disabled).
									Cursor(pick(fp.disabled, ui.CursorNotAllowed, ui.CursorPointer))
								if b.Hovered() && !fp.disabled {
									b.Background(t.FillTertiary)
								}
								b.Children(func() {
									treeTitle(c, t, label, strings.TrimSpace(*query), pick(fp.disabled, t.TextDisabled, t.Text))
								})
								if b.Clicked() && !fp.disabled {
									if multiple {
										last := fp.keys[len(fp.keys)-1]
										*values = conductCheck(options, *values, last, !hasValue(*values, last))
									} else {
										*path = fp.keys
										*open = false
									}
									*query = ""
									changed = true
								}
							})
						})
						return
					}
					checked := map[string]bool{}
					var half map[string]bool
					if multiple {
						shown := conducted(options, *values)
						for _, k := range shown {
							checked[k] = true
						}
						half = halfChecked(options, shown)
					}
					ui.Row(c).AlignItems(ui.Stretch).Children(func() {
						for ci, col := range cascaderColumns(options, *active) {
							ci, col := ci, col
							host := ui.Column(c.Key(ci)).MinWidth(111).MaxHeight(180).Shrink(0)
							if ci > 0 {
								host.Draw(func(pa *ui.Painter, r ui.Rect) { pa.Fill(ui.Rect{X: r.X, Y: r.Y, W: 1, H: r.H}, t.Split, 0) })
							}
							st := ui.Local(host, "list", func() ui.ListState { return ui.ListState{} })
							h := min(float32(len(col))*selectRowH, 180)
							host.Children(func() {
								ui.List(c, st, len(col), func(i int) {
									n := col[i]
									on := ci < len(*active) && (*active)[ci] == n.Key
									item := ui.ButtonBase(c.Key(n.Key)).Label(n.Title).FillWidth().Height(selectRowH).Padding(0, 12, 0, pick[float32](multiple, 8, 12)).
										Radius(4).Gap(t.SpaceSM).AlignItems(ui.Center).FocusRing(false).Disabled(n.Disabled).
										Cursor(pick(n.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
									switch {
									case n.Disabled:
									case on:
										item.Background(t.PrimaryBg)
									case item.Hovered():
										item.Background(t.FillTertiary)
									}
									toggled := false
									item.Children(func() {
										if multiple {
											cb := ui.Box(c).Shrink(0)
											cb.Children(func() {
												checkGlyph(c, t, checked[n.Key], half[n.Key], n.Disabled || n.DisableCheckbox, false)
											})
										}
										txt := ui.Text(c, n.Title).FontSize(t.FontSize).TextColor(pick(n.Disabled, t.TextDisabled, t.Text)).SingleLine().Grow(1)
										if on && !multiple {
											txt.FontWeight(600)
										}
										async := p.LoadData != nil
										leaf := n.leaf(async)
										switch {
										case (*loading)[n.Key]:
											ui.Spinner(c).Size(12, 12).TextColor(t.Primary)
										case !leaf:
											ui.Box(c).Size(12, 12).Rotate(-90).Children(func() { chevron(c, t.TextTertiary, false) })
										}
									})
									if n.Disabled {
										return
									}
									async := p.LoadData != nil
									leaf := n.leaf(async)
									go_ := item.Clicked() || p.HoverOpen && item.Hovered() && !leaf && !on
									if !go_ {
										return
									}
									next := append(append([]string(nil), (*active)[:min(ci, len(*active))]...), n.Key)
									*active = next
									if !item.Clicked() {
										return
									}
									if !leaf && len(n.Children) == 0 && p.LoadData != nil && !(*loading)[n.Key] {
										(*loading)[n.Key] = true
										p.LoadData(pathNodes(options, next))
									}
									switch {
									case multiple:
										x, _, _ := item.PointerPosition()
										if leaf || p.ChangeOnSelect || x < 30 {
											toggled = true
										}
									case leaf:
										*path = next
										*open = false
										changed = true
									case p.ChangeOnSelect:
										*path = next
										changed = true
									}
									if toggled && n.checkable() {
										*values = conductCheck(options, *values, n.Key, !checked[n.Key])
										changed = true
									}
								}).MinWidth(111).Height(h).MaxHeight(180)
							})
						}
					})
				})
			})
		})
		if !*open {
			*query = ""
		}
	})
	return changed
}
