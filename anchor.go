package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// AnchorState follows the sections of a scrolling view for an Anchor:
// track Scroll on the view (ui.Scroll(c).TrackScroll(&s.Scroll)), pass
// the view to Container and each section to Mark as they are built.
type AnchorState struct {
	Scroll ui.ScrollState

	top   float32
	marks map[string]float32
	// active is the link a click chose, kept while the scroll glides to it.
	clicked   string
	clickedAt time.Time
}

// Container records the scrolling view the sections are in.
func (s *AnchorState) Container(view ui.Element) {
	s.top = view.Bounds().Y
}

// Mark records section as the target of the link key.
func (s *AnchorState) Mark(key string, section ui.Element) {
	if s.marks == nil {
		s.marks = map[string]float32{}
	}
	s.marks[key] = section.Bounds().Y - s.top + s.Scroll.Y
}

// AnchorLink is one link of an Anchor, to the section marked Key.
type AnchorLink struct {
	Key, Title string
	Children   []AnchorLink
}

// AnchorProps are an anchor's options.
type AnchorProps struct {
	// Offset is how far below the view's top a section lands when its
	// link is clicked, and counts as reached (Ant's targetOffset).
	Offset float32
	// Bounds is how near the top a section counts as reached; 5 when
	// zero.
	Bounds float32
	// Horizontal lays the links out in a row with an ink bar under the
	// current one.
	Horizontal bool
}

// anchorKeys flattens the links, depth first.
func anchorKeys(links []AnchorLink) []string {
	var out []string
	for _, l := range links {
		out = append(out, l.Key)
		out = append(out, anchorKeys(l.Children)...)
	}
	return out
}

// anchorActive is the link whose section is the last to have reached the
// view's top: its top at most offset + bounds below the scroll.
func anchorActive(keys []string, marks map[string]float32, scrollY, offset, bounds float32) string {
	best, bestTop := "", float32(-1e9)
	for _, k := range keys {
		top, ok := marks[k]
		if !ok {
			continue
		}
		if top-scrollY <= offset+bounds && top >= bestTop {
			best, bestTop = k, top
		}
	}
	return best
}

// Anchor is Ant's Anchor: links to the sections of the view s follows,
// beside a rail with an ink bar at the one being read; a click scrolls
// its section to the view's top. It returns the clicked key, or "".
func Anchor(c *ui.Context, s *AnchorState, links []AnchorLink, props ...AnchorProps) string {
	p := first(props)
	t := Of(c)
	bounds := pick(p.Bounds == 0, 5, p.Bounds)
	active := anchorActive(anchorKeys(links), s.marks, s.Scroll.Y, p.Offset, bounds)
	if s.clicked != "" {
		if c.Now().Sub(s.clickedAt) < 600*time.Millisecond {
			active = s.clicked
			c.After(600*time.Millisecond - c.Now().Sub(s.clickedAt))
		} else {
			s.clicked = ""
		}
	}
	clicked := ""
	var list ui.Element
	if p.Horizontal {
		list = ui.Row(c).Gap(0).Draw(bottomRule(t.Split))
	} else {
		list = ui.Column(c).Padding(0, 0, 0, 2).Draw(func(pa *ui.Painter, r ui.Rect) {
			pa.Fill(ui.Rect{X: r.X, Y: r.Y, W: 2, H: r.H}, t.Split, 0)
		})
	}
	list.Label("锚点")
	var link func(l AnchorLink, depth int)
	link = func(l AnchorLink, depth int) {
		kc := c.Key(l.Key)
		e := ui.ButtonBase(kc).Label(l.Title).FocusRing(false).AlignItems(ui.Center).Radius(0).
			Cursor(ui.CursorPointer).Shrink(0)
		e.Justify(ui.Start)
		if p.Horizontal {
			e.Padding(8, 16)
		} else {
			e.FillWidth().Padding(4, 0, 4, 16+float32(depth)*12)
		}
		on := l.Key == active
		fg := pick(on, t.Primary, t.Text)
		if !on && (e.Hovered() || e.FocusVisible()) {
			fg = t.PrimaryHover
		}
		if on {
			ink := t.Primary
			e.DrawOver(func(pa *ui.Painter, r ui.Rect) {
				if p.Horizontal {
					pa.Fill(ui.Rect{X: r.X, Y: r.Y + r.H - 2, W: r.W, H: 2}, ink, 0)
				} else {
					pa.Fill(ui.Rect{X: r.X - 2, Y: r.Y, W: 2, H: r.H}, ink, 0)
				}
			})
		}
		if e.FocusVisible() {
			focusRing(e, t)
		}
		e.Children(func() {
			ui.Text(kc, l.Title).FontSize(t.FontSize).TextColor(fg).SingleLine()
		})
		if e.Clicked() {
			clicked = l.Key
		}
		for _, ch := range l.Children {
			link(ch, depth+1)
		}
	}
	list.Children(func() {
		for _, l := range links {
			link(l, 0)
		}
	})
	if clicked != "" {
		if top, ok := s.marks[clicked]; ok {
			s.Scroll.ScrollTo(s.Scroll.X, max(top-p.Offset, 0), ui.Animation{Duration: 450 * time.Millisecond, Ease: ui.EaseOut})
		}
		s.clicked, s.clickedAt = clicked, c.Now()
	}
	return clicked
}

// AffixProps are an affix's options.
type AffixProps struct {
	// OffsetTop pins it that far below the top of Target (the window when
	// unset) once scrolling would take it higher; OffsetBottom, when set
	// instead, that far above the bottom once it would go lower.
	OffsetTop, OffsetBottom float32
	Target                  ui.Element
	// Bottom says OffsetBottom is meant, even when it is 0.
	Bottom bool
	// Scroll is the ScrollState tracked by the scroll container the
	// affix is in (TrackScroll): with it the affix pins and lets go in
	// the very frame the scroll crosses OffsetTop, from where the
	// container shows its content this frame, rather than a frame late
	// from where it was laid out last.
	Scroll *ui.ScrollState
}

// affixHold is what an affix keeps between frames.
type affixHold struct {
	w, h  float32
	stuck bool
	// scrollY is Scroll.Y as the frame whose bounds the holder has was
	// built.
	scrollY float32
}

// affixY is where the holder of an affix is this frame: at y in the last
// frame, built at scroll offset was, moved by how much the content
// scrolled since.
func affixY(y, was, now float32) float32 { return y - (now - was) }

// affixStuck reports whether a box at y, h high, inside a view from top
// to bottom, should be pinned.
func affixStuck(y, h, top, bottom float32, p AffixProps) bool {
	if p.Bottom || p.OffsetBottom > 0 {
		return y+h > bottom-p.OffsetBottom
	}
	return y < top+p.OffsetTop
}

// Affix is Ant's Affix: what fn builds stays in place until scrolling
// would carry it past OffsetTop (or OffsetBottom) of the view, then
// stays pinned there, floating over the content with a placeholder
// keeping its room, as CSS's sticky would. Give it Scroll, the state of
// the container it scrolls in, for it to pin in the frame the scroll
// crosses the offset. It reports whether it is pinned.
func Affix(c *ui.Context, fn func(), props ...AffixProps) bool {
	p := first(props)
	holder := ui.Column(c).FillWidth()
	hold := ui.Local(holder, "affix", func() affixHold { return affixHold{} })
	r := holder.Bounds()
	top, bottom := float32(0), float32(0)
	if p.Target.Valid() {
		vb := p.Target.Bounds()
		top, bottom = vb.Y, vb.Y+vb.H
	} else {
		_, bottom = c.Size()
	}
	y := r.Y
	if p.Scroll != nil {
		y = affixY(r.Y, hold.scrollY, p.Scroll.Y)
		hold.scrollY = p.Scroll.Y
	}
	stuck := r.W > 0 && affixStuck(y, hold.h, top, bottom, p)
	if stuck != hold.stuck {
		hold.stuck = stuck
		c.Invalidate()
	}
	if !stuck {
		holder.Children(func() {
			inner := ui.Column(c).FillWidth().Children(fn)
			b := inner.Bounds()
			hold.w, hold.h = b.W, b.H
		})
		return false
	}
	holder.Height(hold.h)
	y = top + p.OffsetTop
	if p.Bottom || p.OffsetBottom > 0 {
		y = bottom - p.OffsetBottom - hold.h
	}
	ui.Overlay(c, func() {
		ui.Column(c).Absolute().Left(r.X).Top(y).Width(r.W).Children(fn)
	})
	return true
}
