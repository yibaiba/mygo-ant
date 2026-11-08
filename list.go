package kit

import "github.com/egoist/mygo/ui"

// ListProps are a list's options.
type ListProps struct {
	// Header and Footer build the bands above and below the items;
	// LoadMore builds what sits under them, as a "加载更多" button.
	Header, Footer, LoadMore func()
	Bordered                 bool
	// Size Small and Large tighten and loosen the rows, as Ant's.
	Size Size
	// NoSplit drops the lines between items.
	NoSplit bool
	// Grid lays the items out as cards in that many columns, with
	// GridGutter between them (16 when zero).
	Grid       int
	GridGutter float32
	// Loading dims the items under a spinner.
	Loading bool
	// Virtual builds only the rows in view of a list Height tall. Lists
	// with more than virtualOver (64) items virtualize on their own
	// unless they are a grid or paged.
	Virtual bool
	Height  float32
	// Page, when set, pages the items PageSize (10) at a time with a
	// pagination under them.
	Page     *int
	PageSize int
	// Empty is the text shown without items; "暂无数据" when empty.
	Empty string
}

func listPad(p ListProps) (y, x float32) {
	y = map[Size]float32{Small: 8, Middle: 12, Large: 16}[p.Size]
	if p.Bordered {
		x = pick[float32](p.Size == Small, 16, 24)
	}
	return y, x
}

// listPage is the items [from, to) of page (1-based) of n items size at a
// time.
func listPage(n, page, size int) (from, to int) {
	if size <= 0 {
		return 0, n
	}
	from = min(max(page-1, 0)*size, n)
	return from, min(from+size, n)
}

// List is Ant's List: item builds item i (with ListItem, often holding a
// ListItemMeta) in a row of its own, split by hairlines, between an
// optional header and footer, as a grid of cards, virtual, or paged.
func List(c *ui.Context, n int, item func(i int), props ...ListProps) {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	t := Of(c)
	py, px := listPad(p)
	box := ui.Column(c).FillWidth()
	if p.Bordered {
		box.Border(1, t.Border).Radius(8)
	}
	band := func(fn func()) {
		ui.Column(c).FillWidth().Padding(py, px).DrawOver(bottomRule(t.Split.Alpha(0.6).Over(t.Surface))).Children(fn)
	}
	split := t.Split.Alpha(0.6).Over(t.Surface)
	from, to := 0, n
	if p.Page != nil {
		size := pick(p.PageSize == 0, 10, p.PageSize)
		from, to = listPage(n, *p.Page, size)
	}
	if !p.Virtual && p.Grid == 0 && p.Page == nil && autoVirtual(n) {
		p.Virtual = true
		if p.Height <= 0 {
			p.Height = virtualBodyH
		}
	}
	rows := func() {
		switch {
		case n == 0:
			ui.Column(c).FillWidth().Padding(py, px).Children(func() { emptyBody(c, pick(p.Empty == "", L(c).Empty, p.Empty)) })
		case p.Grid > 0:
			gap := pick(p.GridGutter == 0, 16, p.GridGutter)
			ui.Column(c).FillWidth().Gap(gap).Children(func() {
				for r := from; r < to; r += p.Grid {
					ui.Row(c).FillWidth().Gap(gap).AlignItems(ui.Stretch).Children(func() {
						for i := r; i < r+p.Grid; i++ {
							ui.Column(c).Basis(0).Grow(1).MinWidth(0).Children(func() {
								if i < to {
									item(i)
								}
							})
						}
					})
				}
			})
		case p.Virtual:
			VirtualList(c, n, func(kc *ui.Context, i int) {
				ui.Column(kc).FillWidth().Padding(py, px).Children(func() { item(i) })
			}, VirtualListProps{Height: p.Height, Split: !p.NoSplit})
		default:
			for i := from; i < to; i++ {
				r := ui.Column(c).FillWidth().Padding(py, px)
				if !p.NoSplit && (i < to-1 || p.Footer != nil || p.LoadMore != nil) {
					r.DrawOver(bottomRule(split))
				}
				r.Children(func() { item(i) })
			}
		}
	}
	box.Children(func() {
		if p.Header != nil {
			band(p.Header)
		}
		SpinOver(c, p.Loading, func() {
			ui.Column(c).FillWidth().Children(rows)
		})
		if p.LoadMore != nil {
			ui.Row(c).FillWidth().Justify(ui.Center).Padding(py, px).Children(p.LoadMore)
		}
		if p.Footer != nil {
			ui.Column(c).FillWidth().Padding(py, px).Children(p.Footer)
		}
		if p.Page != nil && n > 0 {
			ui.Row(c).FillWidth().Justify(ui.End).Padding(py, px).Children(func() {
				Pagination(c, p.Page, n, pick(p.PageSize == 0, 10, p.PageSize))
			})
		}
	})
}

// ListItemProps are a list item's options.
type ListItemProps struct {
	// Actions are links at the end (under the content when Vertical),
	// split by short rules.
	Actions []func()
	// Extra builds content at the far end, as a picture.
	Extra func()
	// Vertical stacks the content over its actions, Ant's vertical
	// itemLayout.
	Vertical bool
}

// ListItem lays out a list row: what fn builds, its actions and extra.
func ListItem(c *ui.Context, fn func(), props ...ListItemProps) {
	p := first(props)
	t := Of(c)
	actions := func() {
		ui.Row(c).Shrink(0).AlignItems(ui.Center).Children(func() {
			for i, a := range p.Actions {
				cell := ui.Row(c).Padding(0, 8).AlignItems(ui.Center)
				if i == 0 {
					cell.PaddingStart(0)
				}
				if i < len(p.Actions)-1 {
					cell.Draw(func(pa *ui.Painter, r ui.Rect) {
						pa.Fill(ui.Rect{X: r.X + r.W - 1, Y: r.Y + r.H/2 - 7, W: 1, H: 14}, t.Split.Alpha(0.6).Over(t.Surface), 0)
					})
				}
				cell.Children(a)
			}
		})
	}
	ui.Row(c).FillWidth().Gap(t.SpaceXL).AlignItems(pick(p.Vertical, ui.Start, ui.Center)).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(t.SpaceLG).Children(func() {
			fn()
			if p.Vertical && len(p.Actions) > 0 {
				actions()
			}
		})
		if !p.Vertical && len(p.Actions) > 0 {
			ui.Row(c).MarginStart(28).Children(actions)
		}
		if p.Extra != nil {
			ui.Row(c).Shrink(0).Children(p.Extra)
		}
	})
}

// ListMetaProps are a list item meta's options.
type ListMetaProps struct {
	Avatar      func()
	Title       string
	Description string
	// TitleRender builds the title instead of Title, as a link.
	TitleRender func()
}

// ListItemMeta is Ant's List.Item.Meta: an avatar beside a title and a
// grey description. A title that does not fit shows the rest in a tooltip.
func ListItemMeta(c *ui.Context, props ListMetaProps) {
	t := Of(c)
	ui.Row(c).FillWidth().Gap(t.SpaceLG).AlignItems(ui.Start).Children(func() {
		if props.Avatar != nil {
			ui.Row(c).Shrink(0).Children(props.Avatar)
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			switch {
			case props.TitleRender != nil:
				props.TitleRender()
			case props.Title != "":
				title := ui.Text(c, props.Title).FontSize(t.FontSize).TextColor(t.Text).FontWeight(500).
					SingleLine().Ellipsis("…").FillWidth()
				ellipsisTip(c, title, props.Title, t.FontSize)
			}
			if props.Description != "" {
				ui.Text(c, props.Description).FontSize(t.FontSize).TextColor(t.TextTertiary)
			}
		})
	})
}
