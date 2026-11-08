package kit

import (
	"strings"

	"github.com/egoist/mygo/ui"
)

// TreeNode is a node of a Tree, TreeSelect or Cascader.
type TreeNode struct {
	Key, Title string
	Children   []TreeNode
	Disabled   bool
	// DisableCheckbox keeps the node's box from being checked while the
	// node can still be chosen.
	DisableCheckbox bool
	// IsLeaf marks a node without children that LoadData should not be
	// asked about.
	IsLeaf bool
	Icon   *ui.SVG
}

func (n *TreeNode) leaf(async bool) bool {
	return len(n.Children) == 0 && (n.IsLeaf || !async)
}

// treeInfo is what treeIndex knows of a node.
type treeInfo struct {
	node   *TreeNode
	parent string
	depth  int
	order  int
}

// treeIndex maps each key of nodes to its node, parent and depth.
func treeIndex(nodes []TreeNode) map[string]treeInfo {
	idx := map[string]treeInfo{}
	n := 0
	var walk func(ns []TreeNode, parent string, depth int)
	walk = func(ns []TreeNode, parent string, depth int) {
		for i := range ns {
			idx[ns[i].Key] = treeInfo{node: &ns[i], parent: parent, depth: depth, order: n}
			n++
			walk(ns[i].Children, ns[i].Key, depth+1)
		}
	}
	walk(nodes, "", 0)
	return idx
}

// checkable reports whether a node's box takes part in checking.
func (n *TreeNode) checkable() bool { return !n.Disabled && !n.DisableCheckbox }

// conductCheck is the checked keys after a node's box is set to on, as
// Ant's conducted check: the node's checkable descendants follow it, and
// each ancestor is checked exactly when all its checkable children are.
func conductCheck(nodes []TreeNode, checked []string, key string, on bool) []string {
	set := map[string]bool{}
	for _, k := range conducted(nodes, checked) {
		set[k] = true
	}
	idx := treeIndex(nodes)
	info, ok := idx[key]
	if !ok {
		return checked
	}
	var down func(n *TreeNode)
	down = func(n *TreeNode) {
		if n.checkable() {
			set[n.Key] = on
		}
		for i := range n.Children {
			down(&n.Children[i])
		}
	}
	down(info.node)
	settleParents(idx, set, info.parent)
	return keysInOrder(nodes, set)
}

// conducted is checked as a conducted check shows it: with the
// checkable descendants of each checked node, and each node whose
// checkable children are all checked.
func conducted(nodes []TreeNode, checked []string) []string {
	if len(checked) == 0 {
		return checked
	}
	set := map[string]bool{}
	for _, k := range checked {
		set[k] = true
	}
	var down func(n *TreeNode, on bool) bool
	down = func(n *TreeNode, on bool) bool {
		on = on || set[n.Key]
		if on && n.checkable() {
			set[n.Key] = true
		}
		all, any := true, false
		for i := range n.Children {
			ch := &n.Children[i]
			got := down(ch, on)
			if ch.checkable() {
				any = true
				all = all && got
			}
		}
		if any && all && n.checkable() {
			set[n.Key] = true
		}
		return set[n.Key]
	}
	for i := range nodes {
		down(&nodes[i], false)
	}
	return keysInOrder(nodes, set)
}

// settleParents checks or unchecks each ancestor from parent up by
// whether all its checkable children are checked.
func settleParents(idx map[string]treeInfo, set map[string]bool, parent string) {
	for parent != "" {
		pi := idx[parent]
		all, any := true, false
		for i := range pi.node.Children {
			ch := &pi.node.Children[i]
			if !ch.checkable() {
				continue
			}
			any = true
			if !set[ch.Key] {
				all = false
			}
		}
		if any && pi.node.checkable() {
			set[parent] = all
		}
		parent = pi.parent
	}
}

// keysInOrder is the keys set holds, in the order of the tree.
func keysInOrder(nodes []TreeNode, set map[string]bool) []string {
	var out []string
	var walk func(ns []TreeNode)
	walk = func(ns []TreeNode) {
		for i := range ns {
			if set[ns[i].Key] {
				out = append(out, ns[i].Key)
			}
			walk(ns[i].Children)
		}
	}
	walk(nodes)
	return out
}

// halfChecked is the nodes that have some, not all, checked descendants:
// the indeterminate boxes.
func halfChecked(nodes []TreeNode, checked []string) map[string]bool {
	set := map[string]bool{}
	for _, k := range checked {
		set[k] = true
	}
	half := map[string]bool{}
	// count returns how many checkable descendants n has and how many of
	// them are checked.
	var count func(n *TreeNode) (int, int)
	count = func(n *TreeNode) (int, int) {
		total, on := 0, 0
		for i := range n.Children {
			ch := &n.Children[i]
			t, o := count(ch)
			if ch.checkable() {
				t++
				if set[ch.Key] {
					o++
				}
			}
			total += t
			on += o
		}
		if on > 0 && on < total && !set[n.Key] {
			half[n.Key] = true
		}
		return total, on
	}
	for i := range nodes {
		count(&nodes[i])
	}
	return half
}

// leafKeys keeps the keys in checked whose nodes have no children: what
// a TreeSelect shows of a conducted check (Ant's SHOW_CHILD).
func leafKeys(nodes []TreeNode, checked []string) []string {
	idx := treeIndex(nodes)
	var out []string
	for _, k := range checked {
		if in, ok := idx[k]; ok && len(in.node.Children) == 0 {
			out = append(out, k)
		}
	}
	return out
}

// flatNode is a row of a tree as shown: a node, its depth and, for
// showLine, whether each of its ancestors (and it) is the last child.
type flatNode struct {
	node  *TreeNode
	depth int
	last  []bool
}

// flattenTree is the rows of nodes with the keys in open expanded.
func flattenTree(nodes []TreeNode, open map[string]bool) []flatNode {
	var out []flatNode
	var walk func(ns []TreeNode, depth int, last []bool)
	walk = func(ns []TreeNode, depth int, last []bool) {
		for i := range ns {
			l := append(append([]bool(nil), last...), i == len(ns)-1)
			out = append(out, flatNode{node: &ns[i], depth: depth, last: l})
			if open[ns[i].Key] {
				walk(ns[i].Children, depth+1, l)
			}
		}
	}
	walk(nodes, 0, nil)
	return out
}

// filterTree keeps the nodes whose title holds q and their ancestors,
// and reports the keys to open so that the matches show.
func filterTree(nodes []TreeNode, q string) ([]TreeNode, map[string]bool) {
	q = strings.ToLower(strings.TrimSpace(q))
	open := map[string]bool{}
	if q == "" {
		return nodes, open
	}
	var walk func(ns []TreeNode) []TreeNode
	walk = func(ns []TreeNode) []TreeNode {
		var out []TreeNode
		for _, n := range ns {
			kids := walk(n.Children)
			if strings.Contains(strings.ToLower(n.Title), q) || len(kids) > 0 {
				if len(kids) > 0 {
					open[n.Key] = true
				}
				m := n
				if !strings.Contains(strings.ToLower(n.Title), q) {
					m.Children = kids
				}
				out = append(out, m)
			}
		}
		return out
	}
	return walk(nodes), open
}

// allKeys is the key of every node with children, to expand them all.
func allKeys(nodes []TreeNode) []string {
	var out []string
	var walk func(ns []TreeNode)
	walk = func(ns []TreeNode) {
		for _, n := range ns {
			if len(n.Children) > 0 {
				out = append(out, n.Key)
				walk(n.Children)
			}
		}
	}
	walk(nodes)
	return out
}

// DropPosition is where a dragged tree node goes against the node it is
// dropped on.
type DropPosition int8

// Drop positions.
const (
	DropBefore DropPosition = -1
	DropInside DropPosition = 0
	DropAfter  DropPosition = 1
)

// dropPositionAt is where a drop at y within a row h tall goes: the top
// quarter before, the bottom quarter after, inside between.
func dropPositionAt(y, h float32) DropPosition {
	switch {
	case y < h/4:
		return DropBefore
	case y > h*3/4:
		return DropAfter
	}
	return DropInside
}

// MoveTreeNode is nodes with the node drag moved before, into or after
// target, as a Tree's OnDrop asks; nodes is left as it is when target is
// drag or within it.
func MoveTreeNode(nodes []TreeNode, drag, target string, pos DropPosition) []TreeNode {
	if drag == target {
		return nodes
	}
	idx := treeIndex(nodes)
	for p := idx[target].parent; p != ""; p = idx[p].parent {
		if p == drag {
			return nodes
		}
	}
	var moved *TreeNode
	var remove func(ns []TreeNode) []TreeNode
	remove = func(ns []TreeNode) []TreeNode {
		out := make([]TreeNode, 0, len(ns))
		for _, n := range ns {
			if n.Key == drag {
				m := n
				moved = &m
				continue
			}
			n.Children = remove(n.Children)
			out = append(out, n)
		}
		return out
	}
	rest := remove(nodes)
	if moved == nil {
		return nodes
	}
	var insert func(ns []TreeNode) []TreeNode
	insert = func(ns []TreeNode) []TreeNode {
		out := make([]TreeNode, 0, len(ns)+1)
		for _, n := range ns {
			if n.Key == target {
				switch pos {
				case DropBefore:
					out = append(out, *moved, n)
				case DropAfter:
					out = append(out, n, *moved)
				default:
					n.Children = append(append([]TreeNode(nil), n.Children...), *moved)
					out = append(out, n)
				}
				continue
			}
			n.Children = insert(n.Children)
			out = append(out, n)
		}
		return out
	}
	return insert(rest)
}

// TreeProps are a tree's options.
type TreeProps struct {
	// Checkable puts a checkbox before each node; Checked holds the
	// checked keys, conducted through the tree.
	Checkable bool
	Checked   *[]string
	// Selected, when set, lets a click choose a node; MultiSelected lets
	// clicks choose several, each click adding or taking out a node.
	Selected      *string
	MultiSelected *[]string
	// Expanded holds the open nodes; the tree keeps them when nil.
	Expanded         *[]string
	DefaultExpandAll bool
	// ShowLine draws the lines between parents and children.
	ShowLine bool
	// Virtual builds only the rows in view, Height tall (it fills its
	// parent's height when zero). Trees with more than virtualOver (64)
	// flattened rows virtualize on their own.
	Virtual bool
	Height  float32
	// LoadData is asked for the children of a node without any that is
	// not IsLeaf, once, as it opens; the node spins until they come.
	LoadData func(key string)
	// OnDrop, set, lets the user drag nodes onto others: it gets the node
	// dragged, the node dropped on, and where (MoveTreeNode applies it).
	OnDrop func(drag, target string, pos DropPosition)
	// Filter keeps the nodes whose title holds it, and their ancestors,
	// opened, with the match in the primary color.
	Filter string
	// BlockNode makes each row fill the width.
	BlockNode bool
	Disabled  bool
}

// treeRowH is the height of a tree row.
const treeRowH = 28

// treeIndent is how far each level steps in.
const treeIndent = 24

// treeDrag is the value a tree row drags.
type treeDrag struct{ key string }

// Tree is Ant's Tree: nodes in an outline, opened by their arrows or a
// double-click of the title, with checkboxes, selection, lines, lazily
// loaded children, drag and drop and a virtual list. It reports a change
// of the check or the selection.
func Tree(c *ui.Context, nodes []TreeNode, props ...TreeProps) bool {
	p := first(props)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	changed := false
	host := ui.Column(c).FillWidth().Role(ui.RoleTree)
	expanded := p.Expanded
	if expanded == nil {
		expanded = ui.Local(host, "expanded", func() []string {
			if p.DefaultExpandAll {
				return allKeys(nodes)
			}
			return nil
		})
	}
	loading := ui.Local(host, "loading", func() map[string]bool { return map[string]bool{} })
	shown := nodes
	open := map[string]bool{}
	for _, k := range *expanded {
		open[k] = true
	}
	if p.Filter != "" {
		var extra map[string]bool
		shown, extra = filterTree(nodes, p.Filter)
		for k := range extra {
			open[k] = true
		}
	}
	for k := range *loading {
		if in, ok := treeIndex(nodes)[k]; !ok || len(in.node.Children) > 0 {
			delete(*loading, k)
		}
	}
	rows := flattenTree(shown, open)
	virt := p.Virtual || autoVirtual(len(rows))
	if virt {
		if p.Height > 0 {
			host.Height(p.Height)
		} else if p.Virtual {
			host.Grow(1).MinHeight(0)
		} else {
			host.Height(virtualBodyH)
		}
	}
	var checked []string
	if p.Checked != nil {
		checked = conducted(nodes, *p.Checked)
	}
	half := halfChecked(nodes, checked)
	isChecked := map[string]bool{}
	for _, k := range checked {
		isChecked[k] = true
	}
	row := func(c *ui.Context, i int) {
		fn := rows[i]
		n := fn.node
		async := p.LoadData != nil
		leaf := n.leaf(async)
		r := ui.Row(c).Height(treeRowH).AlignItems(ui.Center).Gap(4).
			Role(ui.RoleTreeItem).Level(fn.depth + 1).Focusable().FocusRing(false)
		if !leaf {
			r.Expanded(open[n.Key])
		}
		if p.BlockNode || virt {
			r.FillWidth()
		}
		if p.OnDrop != nil && !p.Disabled && !n.Disabled {
			r.Drag(treeDrag{n.Key})
			if d, ok := ui.Drop[treeDrag](r); ok {
				_, y, _ := r.PointerPosition()
				p.OnDrop(d.key, n.Key, dropPositionAt(y, treeRowH))
			}
			if _, ok := ui.DragOver[treeDrag](r); ok {
				_, y, _ := r.PointerPosition()
				pos := dropPositionAt(y, treeRowH)
				indent := float32(fn.depth*treeIndent + 24)
				tone := t.Primary
				r.DrawOver(func(pa *ui.Painter, b ui.Rect) {
					switch pos {
					case DropBefore:
						pa.Fill(ui.Rect{X: b.X + indent, Y: b.Y, W: b.W - indent, H: 2}, tone, 1)
					case DropAfter:
						pa.Fill(ui.Rect{X: b.X + indent, Y: b.Y + b.H - 2, W: b.W - indent, H: 2}, tone, 1)
					default:
						pa.Stroke(ui.Rect{X: b.X + indent, Y: b.Y + 1, W: b.W - indent, H: b.H - 2}, tone, 4, 1)
					}
				})
			}
		}
		lineTone := t.Border
		r.Children(func() {
			if fn.depth > 0 {
				ui.Box(c).Size(float32(fn.depth*treeIndent), treeRowH).Shrink(0).Draw(func(pa *ui.Painter, b ui.Rect) {
					if !p.ShowLine {
						return
					}
					for d := 0; d < fn.depth; d++ {
						if fn.last[d] {
							continue
						}
						x := b.X + float32(d*treeIndent) + treeIndent/2
						pa.Line(x, b.Y, x, b.Y+b.H, 1, lineTone)
					}
				})
			}
			sw := ui.Box(c).Label(pick(open[n.Key], L(c).Collapse, L(c).Expand)+" "+n.Title).Size(treeIndent, treeRowH).Shrink(0).
				Center().Disabled(leaf)
			isLast := fn.last[len(fn.last)-1]
			sw.Draw(func(pa *ui.Painter, b ui.Rect) {
				if !p.ShowLine || fn.depth == 0 && !leaf {
					return
				}
				x := b.X + b.W/2
				if fn.depth > 0 || leaf {
					pa.Line(x, b.Y, x, pick(isLast, b.Y+b.H/2, b.Y+b.H), 1, lineTone)
				}
				if leaf {
					pa.Line(x, b.Y+b.H/2, b.X+b.W, b.Y+b.H/2, 1, lineTone)
				}
			})
			sw.Children(func() {
				switch {
				case (*loading)[n.Key]:
					ui.Spinner(c).Size(12, 12).TextColor(t.Primary)
				case leaf:
				case p.ShowLine:
					ui.Box(c).Size(12, 12).Background(t.Surface).Border(1, t.TextTertiary).Radius(2).Draw(func(pa *ui.Painter, b ui.Rect) {
						cx, cy := b.X+b.W/2, b.Y+b.H/2
						pa.Line(cx-3, cy, cx+3, cy, 1, t.TextTertiary)
						if !open[n.Key] {
							pa.Line(cx, cy-3, cx, cy+3, 1, t.TextTertiary)
						}
					})
				default:
					ui.Box(c).Size(12, 12).Rotate(pick[float32](open[n.Key], 0, -90)).Children(func() { chevron(c, t.TextTertiary, false) })
				}
			})
			toggleOpen := func() {
				if leaf {
					return
				}
				*expanded = toggleValue(append([]string(nil), *expanded...), n.Key)
				if !open[n.Key] && len(n.Children) == 0 && p.LoadData != nil {
					(*loading)[n.Key] = true
					p.LoadData(n.Key)
				}
			}
			if sw.Clicked() {
				toggleOpen()
			}
			if p.Checkable {
				on := isChecked[n.Key]
				box := ui.CheckboxBase(c, &on).Label(n.Title).FocusRing(false).Disabled(p.Disabled || !n.checkable()).
					Cursor(pick(p.Disabled || !n.checkable(), ui.CursorNotAllowed, ui.CursorPointer))
				if half[n.Key] {
					box.Mixed()
				}
				box.Children(func() { checkGlyph(c, t, on, half[n.Key], p.Disabled || !n.checkable(), box.Hovered()) })
				if box.Changed() && p.Checked != nil {
					*p.Checked = conductCheck(nodes, *p.Checked, n.Key, on)
					changed = true
				}
			}
			title := ui.Row(c).Height(24).Padding(0, 4).Radius(4).Gap(4).AlignItems(ui.Center).
				Disabled(n.Disabled || p.Disabled).Cursor(pick(n.Disabled || p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
			if p.BlockNode || virt {
				title.Grow(1).Justify(ui.Start)
			}
			sel := p.Selected != nil && *p.Selected == n.Key || p.MultiSelected != nil && hasValue(*p.MultiSelected, n.Key)
			switch {
			case sel:
				title.Background(t.PrimaryBg)
			case title.Hovered() && !n.Disabled:
				title.Background(t.FillTertiary)
			}
			if r.FocusVisible() {
				focusRing(title, t)
			}
			title.Children(func() {
				if n.Icon != nil {
					ui.Icon(c, n.Icon).Size(14, 14).TextColor(t.TextSecondary)
				}
				treeTitle(c, t, n.Title, p.Filter, pick(n.Disabled || p.Disabled, t.TextDisabled, t.Text))
			})
			dbl := title.DoubleClicked() && !leaf && !n.Disabled && !p.Disabled
			if dbl {
				toggleOpen()
			}
			if !dbl && !n.Disabled && !p.Disabled && !sw.Clicked() && (title.Clicked() || r.Clicked()) {
				switch {
				case p.MultiSelected != nil:
					*p.MultiSelected = keysInOrder(nodes, toggleSet(*p.MultiSelected, n.Key))
					changed = true
				case p.Selected != nil:
					*p.Selected = n.Key
					changed = true
				case p.Checkable && p.Checked != nil && n.checkable():
					*p.Checked = conductCheck(nodes, *p.Checked, n.Key, !isChecked[n.Key])
					changed = true
				}
			}
			if !p.Disabled && !leaf {
				switch {
				case r.Shortcut(0, ui.KeyRight) && !open[n.Key]:
					toggleOpen()
				case r.Shortcut(0, ui.KeyLeft) && open[n.Key]:
					toggleOpen()
				}
			}
		})
	}
	host.FocusGroup(ui.Vertical)
	host.Children(func() {
		if len(rows) == 0 {
			Empty(c, L(c).Empty)
			return
		}
		if virt {
			ui.List(c, nil, len(rows), func(i int) { row(c.Key(rows[i].node.Key), i) }).FillWidth().Grow(1).MinHeight(0)
			return
		}
		for i := range rows {
			row(c.Key(rows[i].node.Key), i)
		}
	})
	return changed
}

// treeTitle is a node's title with the part matching filter in the
// primary color.
func treeTitle(c *ui.Context, t *Tokens, title, filter string, fg ui.Color) {
	i := -1
	if filter != "" {
		i = strings.Index(strings.ToLower(title), strings.ToLower(filter))
	}
	if i < 0 {
		ui.Text(c, title).FontSize(t.FontSize).TextColor(fg).SingleLine()
		return
	}
	j := i + len(filter)
	ui.RichText(c,
		ui.Span{Text: title[:i], Size: t.FontSize, Color: fg},
		ui.Span{Text: title[i:j], Size: t.FontSize, Color: t.Error},
		ui.Span{Text: title[j:], Size: t.FontSize, Color: fg},
	).SingleLine().Label(title)
}

// checkGlyph is the box of a checkbox built on ui.CheckboxBase elsewhere:
// checked, indeterminate (a painted square) or empty.
func checkGlyph(c *ui.Context, t *Tokens, on, mixed, off, hovered bool) {
	box := ui.Box(c).Size(16, 16).Radius(4).Shrink(0)
	switch {
	case off:
		box.Background(t.DisabledBg).Border(1, t.Border)
	case on:
		box.Background(pick(hovered, t.PrimaryHover, t.Primary))
	default:
		box.Background(t.Surface).Border(1, pick(hovered, t.Primary, t.BorderButton))
	}
	mark, dot := pick(off, t.TextDisabled, t.TextOnPrimary), pick(off, t.TextDisabled, t.Primary)
	box.DrawOver(func(pa *ui.Painter, r ui.Rect) {
		switch {
		case on:
			var path ui.Path
			path.MoveTo(r.X+r.W*0.24, r.Y+r.H*0.52).LineTo(r.X+r.W*0.43, r.Y+r.H*0.70).LineTo(r.X+r.W*0.77, r.Y+r.H*0.32)
			pa.StrokePath(&path, 1.8, mark)
		case mixed:
			pa.Fill(ui.Rect{X: r.X + r.W/2 - 4, Y: r.Y + r.H/2 - 4, W: 8, H: 8}, dot, 0)
		}
	})
}

// toggleSet is keys as a set with key added or taken out.
func toggleSet(keys []string, key string) map[string]bool {
	set := map[string]bool{}
	for _, k := range keys {
		set[k] = true
	}
	set[key] = !set[key]
	return set
}
