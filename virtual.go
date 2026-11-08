package kit

import "github.com/egoist/mygo/ui"

// virtualOver is the row/option count past which Select, Table, List and
// Tree build only what is in view, whether or not Virtual is set. A table
// virtualizes the rows it shows: a tree is flattened first (open parents
// and their children), grouped headers still paint in full and only the
// body is a ui.List, which measures wrapping rows and expand panels.
const virtualOver = 64

// virtualBodyH is the body window (DIPs) used when auto-virtualizing a
// table, list or tree that did not set a height, matching Ant scroll.y.
const virtualBodyH float32 = 400

func autoVirtual(n int) bool { return n > virtualOver }

// VirtualListProps are a virtual list's options.
type VirtualListProps struct {
	// Height is the list's; it fills its container's height when zero.
	Height float32
	// State, when set, is the ui.ListState the list keeps its scroll and
	// choice in, for Key, Selected, Selection, FollowEnd and the rest.
	State *ui.ListState
	// Split draws a line between rows, as Ant's List does.
	Split bool
}

// popupList is a ui.List of n rows rowH tall, at most maxH, for select,
// cascader and time-column popups. It does not mix spacer virtualization
// (fixedWindow) with ui.List. ensure, when non-negative, is scrolled into
// view.
func popupList(c *ui.Context, n int, rowH, maxH float32, ensure int, row func(c *ui.Context, i int)) {
	host := ui.Column(c).FillWidth()
	st := ui.Local(host, "plist", func() ui.ListState { return ui.ListState{} })
	if ensure >= 0 {
		st.ScrollIntoView(ensure)
	}
	h := float32(n) * rowH
	if maxH > 0 && h > maxH {
		h = maxH
	}
	if h < rowH {
		h = rowH
	}
	host.Children(func() {
		ui.List(c, st, n, func(i int) {
			ui.Column(c.Key(i)).FillWidth().Height(rowH).Shrink(0).Children(func() { row(c.Key(i), i) })
		}).FillWidth().Height(h)
	})
}

// VirtualList is a list of n rows of which only those in view are built,
// as Ant's virtual lists: ui.List with the kit's look. row builds row i
// in its own keyed context.
func VirtualList(c *ui.Context, n int, row func(c *ui.Context, i int), props ...VirtualListProps) ui.Element {
	p := first(props)
	t := Of(c)
	l := ui.List(c, p.State, n, func(i int) {
		r := ui.Column(c.Key(i)).FillWidth()
		if p.Split && i < n-1 {
			r.Draw(bottomRule(t.Split))
		}
		r.Children(func() { row(c.Key(i), i) })
	}).FillWidth()
	if p.Height > 0 {
		l.Height(p.Height)
	} else {
		l.Grow(1).MinHeight(0)
	}
	return l
}

// virtualRange is the rows [from, to) of n rows rowH tall to build in a
// view viewH tall scrolled offset down, with overscan rows more on each
// side.
func virtualRange(offset, viewH, rowH float32, n, overscan int) (from, to int) {
	if n <= 0 || rowH <= 0 {
		return 0, 0
	}
	offset = max(offset, 0)
	from = int(offset/rowH) - overscan
	to = int((offset+max(viewH, 0))/rowH) + 1 + overscan
	return max(from, 0), min(max(to, 0), n)
}

// fixedWindow builds, in a scroll view at most maxH tall, the rows of n
// rows rowH tall in view (and a few more), standing in empty space for
// the others, so that thousands of options cost no more than a screenful.
// Unlike VirtualList it sizes itself to its rows, for popups. It scrolls
// row ensure into view when it changes (none when negative).
func fixedWindow(c *ui.Context, n int, rowH, maxH float32, row func(c *ui.Context, i int), ensure int) ui.Element {
	sc := ui.Scroll(c).FillWidth().MaxHeight(maxH)
	st := ui.Local(sc, "scroll", func() ui.ScrollState { return ui.ScrollState{} })
	last := ui.Local(sc, "ensure", func() int { return -1 })
	if ensure >= 0 && ensure != *last {
		st.Y = ensureVisible(st.Y, maxH, rowH, ensure)
	}
	*last = ensure
	sc.TrackScroll(st)
	from, to := virtualRange(st.Y, min(maxH, float32(n)*rowH), rowH, n, 4)
	sc.Children(func() {
		col := ui.Column(c).FillWidth().Height(float32(n) * rowH)
		col.Children(func() {
			if from > 0 {
				ui.Box(c.Key("above")).FillWidth().Height(float32(from) * rowH).Shrink(0)
			}
			for i := from; i < to; i++ {
				ui.Column(c.Key(i)).FillWidth().Height(rowH).Shrink(0).Children(func() { row(c.Key(i), i) })
			}
		})
	})
	return sc
}

// ensureVisible is the scroll offset of a view viewH tall over rows rowH
// tall nearest to offset that shows row i whole.
func ensureVisible(offset, viewH, rowH float32, i int) float32 {
	top := float32(i) * rowH
	switch {
	case top < offset:
		return top
	case top+rowH > offset+viewH:
		return top + rowH - viewH
	}
	return offset
}
