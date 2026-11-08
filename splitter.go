package kit

import "github.com/egoist/mygo/ui"

// SplitPanel is one pane of a Splitter.
type SplitPanel struct {
	Render func()
	// Size is its share of the room at first, in DIPs; panels without
	// one share what the others leave equally.
	Size float32
	// Min and Max bound it while dragging, in DIPs; 0 leaves it unbounded.
	Min, Max float32
	// Collapsible puts arrows on its bars that fold it away and back.
	Collapsible bool
	// Fixed keeps its bars from dragging it (Ant's resizable={false}).
	Fixed bool
}

// SplitterProps are a Splitter's options.
type SplitterProps struct {
	// Vertical stacks the panels instead of laying them side by side.
	Vertical bool
	// Lazy drags a preview line and resizes on release.
	Lazy bool
	// Sizes, when set, holds the panels' shares of the room, summing to
	// 1; the splitter writes it as they move.
	Sizes *[]float32
}

// splitHold is what a splitter keeps between frames.
type splitHold struct {
	sizes  []float32 // shares of the room, summing to 1
	before []float32 // a collapsed panel's share before it folded
	lazy   float32   // how far a lazy drag has gone
	bar    int       // the bar being lazily dragged, or -1
	// measured is whether the shares were set from the room, which the
	// first frame does not know yet.
	measured bool
}

// splitShares turns the panels' starting sizes into shares of room.
func splitShares(panels []SplitPanel, room float32) []float32 {
	out := make([]float32, len(panels))
	var used float32
	free := 0
	for i, p := range panels {
		if p.Size > 0 && room > 0 {
			out[i] = min(p.Size/room, 1)
			used += out[i]
		} else {
			free++
		}
	}
	for i, p := range panels {
		if !(p.Size > 0 && room > 0) {
			out[i] = max(1-used, 0) / float32(free)
		}
	}
	return normalize(out)
}

func normalize(s []float32) []float32 {
	var sum float32
	for _, v := range s {
		sum += v
	}
	if sum <= 0 {
		return s
	}
	for i := range s {
		s[i] /= sum
	}
	return s
}

// splitMove moves the bar after panel i by d DIPs of room, keeping both
// panels beside it within their bounds; px are the panels' sizes.
func splitMove(px []float32, panels []SplitPanel, i int, d float32) []float32 {
	out := append([]float32(nil), px...)
	a, b := &out[i], &out[i+1]
	total := *a + *b
	lo, hi := float32(0), total
	if m := panels[i].Min; m > 0 {
		lo = max(lo, m)
	}
	if m := panels[i].Max; m > 0 {
		hi = min(hi, m)
	}
	if m := panels[i+1].Min; m > 0 {
		hi = min(hi, total-m)
	}
	if m := panels[i+1].Max; m > 0 {
		lo = max(lo, total-m)
	}
	na := *a + d
	if hi >= lo {
		na = max(lo, min(hi, na))
	}
	*a, *b = na, total-na
	return out
}

// splitFold folds or restores panels about the bar after panel i:
// toStart folds panel i into i+1 (or brings back i+1 if it was
// folded), otherwise the other way. before keeps folded shares.
func splitFold(s, before []float32, i int, toStart bool) {
	a, b := i, i+1
	if !toStart {
		a, b = b, a
	}
	if s[b] == 0 && before[b] > 0 {
		give := min(before[b], s[a])
		s[a] -= give
		s[b] = give
		before[b] = 0
		return
	}
	before[a] = s[a]
	s[b] += s[a]
	s[a] = 0
}

// Splitter is Ant's Splitter: panels side by side (or stacked) with a
// bar between each two that the pointer drags to share the room between
// them, within their Min and Max, and arrows on the bars of Collapsible
// panels folding them away and back. It reports a resize.
func Splitter(c *ui.Context, panels []SplitPanel, props ...SplitterProps) bool {
	p := first(props)
	t := Of(c)
	box := pick(p.Vertical, ui.Column(c), ui.Row(c)).Grow(1).MinWidth(0).MinHeight(0).AlignItems(ui.Stretch)
	hold := ui.Local(box, "hold", func() splitHold { return splitHold{bar: -1} })
	b := box.Bounds()
	room := pick(p.Vertical, b.H, b.W) - float32(len(panels)-1)*2
	if p.Sizes != nil && len(*p.Sizes) == len(panels) {
		hold.sizes, hold.measured = *p.Sizes, true
	}
	if len(hold.sizes) != len(panels) || !hold.measured && room > 0 {
		hold.sizes = splitShares(panels, room)
		hold.before = make([]float32, len(panels))
		hold.measured = room > 0
		if p.Sizes != nil && hold.measured {
			*p.Sizes = append((*p.Sizes)[:0], hold.sizes...)
		}
	}
	if len(hold.before) != len(panels) {
		hold.before = make([]float32, len(panels))
	}
	changed := false
	resize := func(i int, d float32) {
		if room <= 0 || d == 0 {
			return
		}
		px := make([]float32, len(hold.sizes))
		for k, v := range hold.sizes {
			px[k] = v * room
		}
		px = splitMove(px, panels, i, d)
		for k, v := range px {
			hold.sizes[k] = v / room
		}
		changed = true
	}
	box.Children(func() {
		for i, panel := range panels {
			pane := ui.Column(c.Key(i)).Basis(0).Grow(max(hold.sizes[i], 0.0001)).MinWidth(0).MinHeight(0).Clip()
			if hold.sizes[i] == 0 {
				pane.Grow(0).Width(0).Height(0)
				if !p.Vertical {
					pane.FillHeight()
				} else {
					pane.FillWidth()
				}
			}
			pane.Children(panel.Render)
			if i == len(panels)-1 {
				continue
			}
			splitBar(c.Key(i+500), t, p, panels, hold, i, resize, &changed)
		}
	})
	if changed && p.Sizes != nil {
		*p.Sizes = append((*p.Sizes)[:0], hold.sizes...)
	}
	return changed
}

func splitBar(c *ui.Context, t *Tokens, p SplitterProps, panels []SplitPanel, hold *splitHold, i int, resize func(int, float32), changed *bool) {
	fixed := panels[i].Fixed || panels[i+1].Fixed
	bar := ui.Box(c).Shrink(0)
	if p.Vertical {
		bar.FillWidth().Height(2)
	} else {
		bar.FillHeight().Width(2)
	}
	bar.Children(func() {
		grip := ui.Box(c).Absolute().Label("拖动调整大小")
		if p.Vertical {
			grip.Left(0).Right(0).Top(-4).Height(10).Cursor(pick(fixed, ui.CursorDefault, ui.CursorResizeNS))
		} else {
			grip.Top(0).Bottom(0).Left(-4).Width(10).Cursor(pick(fixed, ui.CursorDefault, ui.CursorResizeEW))
		}
		line := t.FillSecondary
		if !fixed {
			if grip.Hovered() {
				line = t.Fill
			}
			if dx, dy, ok := grip.Dragged(); ok {
				d := pick(p.Vertical, dy, dx)
				line = t.PrimaryHover
				if p.Lazy {
					hold.bar = i
					hold.lazy += d
				} else {
					resize(i, d)
				}
			}
			if _, _, ended := grip.DragEnded(); ended && p.Lazy && hold.bar == i {
				resize(i, hold.lazy)
				hold.lazy, hold.bar = 0, -1
			}
		}
		grip.Draw(func(pa *ui.Painter, r ui.Rect) {
			if p.Vertical {
				pa.Fill(ui.Rect{X: r.X, Y: r.Y + 4, W: r.W, H: 2}, line, 0)
			} else {
				pa.Fill(ui.Rect{X: r.X + 4, Y: r.Y, W: 2, H: r.H}, line, 0)
			}
		})
		if p.Lazy && hold.bar == i {
			ghost := ui.Box(c).Absolute().Background(t.Primary).PassThrough()
			if p.Vertical {
				ghost.Left(0).Right(0).Top(hold.lazy).Height(2)
			} else {
				ghost.Top(0).Bottom(0).Left(hold.lazy).Width(2)
			}
		}
		s := hold.sizes
		start := (panels[i].Collapsible && s[i] > 0) || (s[i+1] == 0 && hold.before[i+1] > 0)
		end := (panels[i+1].Collapsible && s[i+1] > 0) || (s[i] == 0 && hold.before[i] > 0)
		if start && splitArrow(c.Key("start"), t, p.Vertical, true) {
			splitFold(hold.sizes, hold.before, i, true)
			*changed = true
		}
		if end && splitArrow(c.Key("end"), t, p.Vertical, false) {
			splitFold(hold.sizes, hold.before, i, false)
			*changed = true
		}
	})
}

// splitArrow is the small fold button on a splitter bar, pointing to
// its start or end.
func splitArrow(c *ui.Context, t *Tokens, vertical, toStart bool) bool {
	btn := ui.ButtonBase(c).Label(pick(toStart, "向前折叠", "向后折叠")).FocusRing(false).Absolute().
		Radius(4).Center().Cursor(ui.CursorPointer)
	off := pick[float32](toStart, -22, 6)
	if vertical {
		btn.Size(20, 12).Top(-5).LeftPercent(50).Translate(off, 0)
	} else {
		btn.Size(12, 20).Left(-5).TopPercent(50).Translate(0, off)
	}
	btn.Background(pick(btn.Hovered(), t.Fill, t.FillSecondary))
	name := map[[2]bool]string{{false, true}: "left", {false, false}: "right", {true, true}: "up", {true, false}: "down"}[[2]bool{vertical, toStart}]
	btn.Children(func() { IconNamed(c, name, IconProps{Size: 10, Color: t.TextTertiary}) })
	return btn.Clicked()
}
