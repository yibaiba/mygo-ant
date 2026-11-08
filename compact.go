package kit

import "github.com/egoist/mygo/ui"

// CompactProps are Space.Compact's options.
type CompactProps struct {
	// Vertical stacks the controls, joined top to bottom.
	Vertical bool
	// Block fills the width of the parent.
	Block bool
	// Size is the size of every control inside (Middle keeps theirs).
	Size Size
}

// compactPart is a bordered element of a control in a compact group: its
// corner radius, and whether its edge is lit (hovered, focused, a status
// colour) and in which colour.
type compactPart struct {
	e    ui.Element
	r    float32
	hot  bool
	edge ui.Color
	// join, when set, is drawn between two such parts side by side, as
	// Ant's line between primary buttons.
	join ui.Color
}

// compactItem is one control of a compact group: the element directly in
// the group and its bordered parts, start to end (an addon, the box, a
// joined button).
type compactItem struct {
	outer ui.Element
	parts []compactPart
}

type compactGroup struct {
	vertical bool
	items    []compactItem
	cur      *compactSlot
}

// compactSlot is a control's claim on its place in the group being built.
type compactSlot struct {
	g     *compactGroup
	outer ui.Element
	parts []compactPart
}

// compactStack is the compact groups being built, innermost last. A
// window is built on one goroutine and ui has no value context, so the
// group a control sits in is found here rather than through c.
var compactStack []*compactGroup

// compactBegin claims a place in the innermost compact group for the
// control being built, or returns nil outside one or inside a control
// that already claimed (a button in a picker's panel). The control calls
// done when built, deferred.
func compactBegin() *compactSlot {
	if len(compactStack) == 0 {
		return nil
	}
	g := compactStack[len(compactStack)-1]
	if g.cur != nil {
		return nil
	}
	g.cur = &compactSlot{g: g}
	return g.cur
}

// at sets the element the control puts in the group, once: the first
// call, by the outermost builder, wins.
func (s *compactSlot) at(outer ui.Element) {
	if s != nil && !s.outer.Valid() {
		s.outer = outer
	}
}

// part adds a bordered element of the control, after those added before.
func (s *compactSlot) part(e ui.Element, r float32, hot bool, edge ui.Color) {
	if s != nil {
		s.parts = append(s.parts, compactPart{e: e, r: r, hot: hot, edge: edge})
	}
}

// field adds the box of a field drawn by fieldLookV in st and status.
func (s *compactSlot) field(e ui.Element, t *Tokens, st visual, status Status, edge ui.Color) {
	s.part(e, t.RadiusControl, st == hover || st == focus || st == active || status != StatusNone, edge)
}

func (s *compactSlot) done() {
	if s == nil {
		return
	}
	if s.outer.Valid() && len(s.parts) > 0 {
		s.g.items = append(s.g.items, compactItem{outer: s.outer, parts: s.parts})
	}
	s.g.cur = nil
}

// compactCorners is which corners of part j of k+1 parts of an item keep
// their radius, top-left, top-right, bottom-right and bottom-left, for an
// item first and last in its group.
func compactCorners(j, k int, first, last, vertical bool) [4]bool {
	s, e := j == 0, j == k
	if vertical {
		return [4]bool{s && first, e && first, e && last, s && last}
	}
	return [4]bool{s && first, e && last, e && last, s && first}
}

// apply joins the group's controls as Ant's Space.Compact does: only the
// outer corners stay round, each control after the first overlaps the
// one before by its 1px border, and where the one before is lit its edge
// is drawn again over the overlap, as Ant raises it with z-index.
func (g *compactGroup) apply() {
	n := len(g.items)
	for i, it := range g.items {
		k := len(it.parts) - 1
		for j, pt := range it.parts {
			cs := compactCorners(j, k, i == 0, i == n-1, g.vertical)
			r := func(on bool) float32 { return pick(on, pt.r, 0) }
			pt.e.Radius(r(cs[0]), r(cs[1]), r(cs[2]), r(cs[3]))
		}
		if i == 0 {
			continue
		}
		if g.vertical {
			it.outer.Margin(-1, 0, 0, 0)
		} else {
			it.outer.Margin(0, 0, 0, -1)
		}
		prev, lit := compactJoin(g.items[i-1], it, g.vertical)
		if !lit {
			continue
		}
		if prev == (ui.Color{}) {
			continue
		}
		vertical := g.vertical
		it.outer.DrawOver(func(pa *ui.Painter, r ui.Rect) {
			if vertical {
				pa.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: 1}, prev, 0)
			} else {
				pa.Fill(ui.Rect{X: r.X, Y: r.Y, W: 1, H: r.H}, prev, 0)
			}
		})
	}
}

// compactJoin is the colour of the edge item b shares with a, before it,
// and whether it must be drawn over b: a's side of the edge is lit and
// b's is not.
func compactJoin(a, b compactItem, vertical bool) (ui.Color, bool) {
	if vertical {
		ae, aHot := anyHot(a.parts)
		_, bHot := anyHot(b.parts)
		return ae, aHot && !bHot
	}
	pa, pb := a.parts[len(a.parts)-1], b.parts[0]
	switch {
	case pa.hot && !pb.hot:
		return pa.edge, true
	case !pa.hot && !pb.hot && pa.join != ui.Color{} && pa.join == pb.join:
		return pa.join, true
	}
	return ui.Color{}, false
}

func anyHot(parts []compactPart) (ui.Color, bool) {
	for _, p := range parts {
		if p.hot {
			return p.edge, true
		}
	}
	return ui.Color{}, false
}

// SpaceCompact is Ant's Space.Compact: the controls fn builds (Button,
// Input with its addons, Search, InputNumber, Select, MultiSelect,
// TreeSelect, Cascader, AutoComplete, DatePicker, TimePicker,
// TimeRangePicker, RangePicker) joined edge to edge in a row, or a column when Vertical,
// sharing borders, with only the outer corners round. A control lit by
// hover or focus shows its whole border.
func SpaceCompact(c *ui.Context, fn func(), props ...CompactProps) ui.Element {
	p := first(props)
	if p.Size == Large || p.Size == Small {
		size := p.Size
		p.Size = SizeInherit
		return ConfigProvider(c, ConfigProps{Size: size, Disabled: Bool(Of(c).Disabled)}, func() { SpaceCompact(c, fn, p) })
	}
	var box ui.Element
	if p.Vertical {
		box = ui.Column(c).AlignItems(ui.Stretch)
	} else {
		box = ui.Row(c).AlignItems(ui.Center)
	}
	if p.Block {
		box.FillWidth()
	} else {
		box.AlignSelf(ui.Start)
	}
	g := &compactGroup{vertical: p.Vertical}
	compactStack = append(compactStack, g)
	box.Children(fn)
	compactStack = compactStack[:len(compactStack)-1]
	g.apply()
	return box
}

// joinLast sets the join colour of the part added last.
func (s *compactSlot) joinLast(c ui.Color) {
	if s != nil && len(s.parts) > 0 {
		s.parts[len(s.parts)-1].join = c
	}
}
