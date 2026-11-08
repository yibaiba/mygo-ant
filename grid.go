package kit

import (
	"sort"

	"github.com/egoist/mygo/ui"
)

// Screen is one of Ant's responsive breakpoints, by window width.
type Screen uint8

// Ant's breakpoints: xs below 576, sm from 576, md 768, lg 992, xl 1200,
// xxl 1600.
const (
	ScreenXS Screen = iota
	ScreenSM
	ScreenMD
	ScreenLG
	ScreenXL
	ScreenXXL
)

// ScreenOf is the breakpoint of a window w wide.
func ScreenOf(w float32) Screen {
	switch {
	case w >= 1600:
		return ScreenXXL
	case w >= 1200:
		return ScreenXL
	case w >= 992:
		return ScreenLG
	case w >= 768:
		return ScreenMD
	case w >= 576:
		return ScreenSM
	}
	return ScreenXS
}

// CurrentScreen is the breakpoint of the window being built.
func CurrentScreen(c *ui.Context) Screen {
	w, _ := c.Size()
	return ScreenOf(w)
}

// FlexProps are a flex box's options.
type FlexProps struct {
	Vertical bool
	Wrap     bool
	Justify  ui.Align
	Align    ui.Align
	// Gap is Small 8, Middle 16 or Large 24 apart; Pixels overrides it.
	Gap    Size
	Pixels float32
	// NoGap keeps the items touching.
	NoGap bool
}

// Flex is Ant's Flex: a row (or column) of what fn builds with Ant's
// gaps, justification and alignment.
func Flex(c *ui.Context, fn func(), props ...FlexProps) ui.Element {
	p := first(props)
	gap := map[Size]float32{Small: 8, Middle: 16, Large: 24}[p.Gap]
	if p.Pixels > 0 {
		gap = p.Pixels
	}
	if p.NoGap {
		gap = 0
	}
	e := pick(p.Vertical, ui.Column(c), ui.Row(c))
	e.Gap(gap).Justify(p.Justify).AlignItems(p.Align)
	if p.Wrap {
		e.Wrap().GapY(gap)
	}
	return e.Children(fn)
}

// ColSpan is how a column lays out at one breakpoint, in 24ths of the
// row: Span wide (0 hides it), Offset after the column before it, Order
// among the row's columns, Push and Pull shifting it right or left.
type ColSpan struct {
	Span, Offset, Order, Push, Pull int
}

// ColProps are a column's spans: the default, and those of each
// breakpoint up from it, as Ant's xs … xxl, the widest set one at or
// below the window's breakpoint applying.
type ColProps struct {
	ColSpan
	XS, SM, MD, LG, XL, XXL *ColSpan
	// Flex sizes it by its share of what is left instead of a span, as
	// Ant's flex={n}; Auto sizes it by its content (flex="auto").
	Flex float32
	Auto bool
}

// spanAt is the layout of a column at screen s.
func (p ColProps) spanAt(s Screen) ColSpan {
	out := p.ColSpan
	for i, sp := range []*ColSpan{p.XS, p.SM, p.MD, p.LG, p.XL, p.XXL} {
		if sp != nil && Screen(i) <= s {
			out = *sp
		}
	}
	return out
}

// RowProps are a grid row's options.
type RowProps struct {
	// Gutter is the room between columns and between wrapped lines.
	Gutter [2]float32
	// GutterBy, when set, overrides Gutter at each breakpoint up from it.
	GutterBy map[Screen][2]float32
	Justify  ui.Align
	Align    ui.Align
	// NoWrap keeps the columns on one line.
	NoWrap bool
}

func (p RowProps) gutterAt(s Screen) [2]float32 {
	g := p.Gutter
	best := -1
	for k, v := range p.GutterBy {
		if k <= s && int(k) > best {
			best, g = int(k), v
		}
	}
	return g
}

// GridCols collects the columns of a GridRow.
type GridCols struct {
	cols []gridCol
}

type gridCol struct {
	props ColProps
	fn    func()
	at    int
}

// Col adds a column span 24ths wide (ColProps for the rest) holding
// what fn builds.
func (g *GridCols) Col(span int, fn func(), props ...ColProps) {
	p := first(props)
	if p.Span == 0 && span != 0 {
		p.Span = span
	}
	g.cols = append(g.cols, gridCol{props: p, fn: fn, at: len(g.cols)})
}

// gridPlan is the columns of a row in their order at screen s, hidden
// ones dropped.
func gridPlan(cols []gridCol, s Screen) []gridCol {
	var out []gridCol
	for _, col := range cols {
		sp := col.props.spanAt(s)
		if sp.Span == 0 && col.props.Flex == 0 && !col.props.Auto {
			continue
		}
		out = append(out, col)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].props.spanAt(s).Order < out[j].props.spanAt(s).Order
	})
	return out
}

// GridRow is Ant's Row of 24-column Cols: fn adds the columns through
// GridCols.Col; each is its span's share of the row less the gutter,
// wrapping onto new lines, laid out at the window's breakpoint.
func GridRow(c *ui.Context, fn func(g *GridCols), props ...RowProps) {
	p := first(props)
	s := CurrentScreen(c)
	g := &GridCols{}
	fn(g)
	gut := p.gutterAt(s)
	row := ui.Row(c).FillWidth().Justify(p.Justify).AlignItems(p.Align)
	if gut[0] > 0 {
		row.MarginX(-gut[0] / 2)
	}
	if !p.NoWrap {
		row.Wrap().GapY(gut[1])
	}
	row.Children(func() {
		width := row.Bounds().W
		for _, col := range gridPlan(g.cols, s) {
			sp := col.props.spanAt(s)
			if sp.Offset > 0 {
				ui.Box(c.Key(col.at)).WidthPercent(float32(sp.Offset) * 100 / 24).Shrink(0)
			}
			e := ui.Column(c.Key(col.at + 1000)).PaddingX(gut[0] / 2).MinWidth(0)
			switch {
			case col.props.Auto:
				e.Grow(1).Basis(0)
			case col.props.Flex > 0:
				e.Grow(col.props.Flex).Basis(0)
			default:
				e.WidthPercent(float32(sp.Span) * 100 / 24).Shrink(0)
			}
			if shift := sp.Push - sp.Pull; shift != 0 && width > 0 {
				e.Translate(width*float32(shift)/24, 0)
			}
			e.Children(col.fn)
		}
	})
}
