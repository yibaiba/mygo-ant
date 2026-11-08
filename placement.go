package kit

import "github.com/egoist/mygo/ui"

// Placement is the side of its anchor a tooltip or pop confirm opens on,
// as Ant's placement. Near a window edge it goes to the other side.
type Placement uint8

// Placements.
const (
	// PlaceTop opens above the anchor, the default.
	PlaceTop Placement = iota
	// PlaceBottom opens below it.
	PlaceBottom
	// PlaceLeft opens on its left.
	PlaceLeft
	// PlaceRight opens on its right.
	PlaceRight
)

func (p Placement) opposite() Placement { return p ^ 1 }

func (p Placement) vertical() bool { return p <= PlaceBottom }

// PopupAlign lines a popup up along the side of its anchor it opens on,
// the second half of Ant's placements (topLeft, bottomRight, leftTop, …).
type PopupAlign uint8

// Popup alignments.
const (
	// AlignMiddle centers the popup on the anchor, the default.
	AlignMiddle PopupAlign = iota
	// AlignFirst lines up their left edges (top edges on the left and
	// right sides): Ant's topLeft, bottomLeft, leftTop, rightTop.
	AlignFirst
	// AlignLast lines up their right (bottom) edges.
	AlignLast
)

// imeBand is the IME candidate strip (~40 DIP below the caret) that a
// PlaceBottom panel would cover while composing.
const imeBand float32 = 40

// imePlacement prefers PlaceTop while composing so a PlaceBottom panel
// does not sit on the IME candidate strip. flipPlacement still runs.
func imePlacement(want Placement, composing bool) Placement {
	if composing && want == PlaceBottom {
		return PlaceTop
	}
	return want
}

// flipPlacement keeps want when a w×h panel fits between anchor and that
// window edge, else takes the opposite side if it has more room, as Ant's
// autoAdjustOverflow. A panel not yet measured (zero size) keeps want.
func flipPlacement(want Placement, anchor ui.Rect, w, h, winW, winH float32) Placement {
	room := func(p Placement) float32 {
		switch p {
		case PlaceTop:
			return anchor.Y
		case PlaceBottom:
			return winH - anchor.Y - anchor.H
		case PlaceLeft:
			return anchor.X
		}
		return winW - anchor.X - anchor.W
	}
	need := pick(want.vertical(), h, w)
	if need <= 0 || room(want) >= need {
		return want
	}
	if o := want.opposite(); room(o) > room(want) {
		return o
	}
	return want
}

// arrowShift is how far an arrow moves from the middle of a panel's side
// to point at the anchor's middle, kept inset from the panel's corners.
func arrowShift(side Placement, anchor, panel ui.Rect) float32 {
	if panel.W <= 0 || panel.H <= 0 {
		return 0
	}
	const inset = 12 + arrowWidth/2
	var d, half float32
	if side.vertical() {
		d, half = anchor.X+anchor.W/2-(panel.X+panel.W/2), panel.W/2
	} else {
		d, half = anchor.Y+anchor.H/2-(panel.Y+panel.H/2), panel.H/2
	}
	lim := max(half-inset, 0)
	return max(-lim, min(d, lim))
}

const (
	arrowWidth = 16
	arrowDepth = 8
)

// attachAnchors are the anchor points that put a panel on side of an
// anchor, lined up as align says: the anchor's point, then the panel's.
func attachAnchors(side Placement, align PopupAlign) (at, self ui.Anchor) {
	type pair struct{ at, self ui.Anchor }
	table := map[Placement][3]pair{
		PlaceTop:    {{ui.AnchorTop, ui.AnchorBottom}, {ui.AnchorTopLeft, ui.AnchorBottomLeft}, {ui.AnchorTopRight, ui.AnchorBottomRight}},
		PlaceBottom: {{ui.AnchorBottom, ui.AnchorTop}, {ui.AnchorBottomLeft, ui.AnchorTopLeft}, {ui.AnchorBottomRight, ui.AnchorTopRight}},
		PlaceLeft:   {{ui.AnchorLeft, ui.AnchorRight}, {ui.AnchorTopLeft, ui.AnchorTopRight}, {ui.AnchorBottomLeft, ui.AnchorBottomRight}},
		PlaceRight:  {{ui.AnchorRight, ui.AnchorLeft}, {ui.AnchorTopRight, ui.AnchorTopLeft}, {ui.AnchorBottomRight, ui.AnchorBottomLeft}},
	}
	pr := table[side][align]
	return pr.at, pr.self
}

// arrowPanel lays out wrap, a floating element attached to anchor, as a
// bubble with an arrow pointing at anchor on the side want, or the other
// side where want has no room. bubble styles the bubble and builds its
// content. It returns the side taken.
func arrowPanel(c *ui.Context, wrap, anchor ui.Element, want Placement, tone ui.Color, bubble func(ui.Element)) Placement {
	return placePanel(c, wrap, anchor, panelPlace{Side: want, Arrow: true, Gap: 2}, tone, bubble)
}

// panelPlace is where placePanel puts a floating panel.
type panelPlace struct {
	Side  Placement
	Align PopupAlign
	Arrow bool
	// Gap keeps the panel apart from the anchor.
	Gap float32
}

// placePanel lays out wrap beside anchor as pl says, flipping to the other
// side where there is no room, with an arrow of tone pointing at the
// anchor when pl.Arrow. bubble styles the panel and builds its content.
// It returns the side taken.
func placePanel(c *ui.Context, wrap, anchor ui.Element, pl panelPlace, tone ui.Color, bubble func(ui.Element)) Placement {
	winW, winH := c.Size()
	prev, at := wrap.Bounds(), anchor.Bounds()
	side := flipPlacement(pl.Side, at, prev.W, prev.H, winW, winH)
	ap, sp := attachAnchors(side, pl.Align)
	wrap.AttachTo(anchor, ap, sp)
	if side.vertical() {
		wrap.Column().MarginY(pl.Gap)
	} else {
		wrap.Row().MarginX(pl.Gap)
	}
	if !pl.Arrow {
		wrap.AlignItems(ui.Stretch)
		wrap.Children(func() { bubble(ui.Column(c).Grow(1)) })
		return side
	}
	wrap.AlignItems(ui.Center)
	shift := arrowShift(side, at, prev)
	arrow := func() {
		a := ui.Box(c).Shrink(0)
		if side.vertical() {
			a.Size(arrowWidth, arrowDepth).Left(shift)
		} else {
			a.Size(arrowDepth, arrowWidth).Top(shift)
		}
		a.Draw(func(p *ui.Painter, r ui.Rect) {
			var path ui.Path
			switch side {
			case PlaceTop:
				path.MoveTo(r.X, r.Y).LineTo(r.X+r.W, r.Y).LineTo(r.X+r.W/2, r.Y+r.H)
			case PlaceBottom:
				path.MoveTo(r.X, r.Y+r.H).LineTo(r.X+r.W, r.Y+r.H).LineTo(r.X+r.W/2, r.Y)
			case PlaceLeft:
				path.MoveTo(r.X, r.Y).LineTo(r.X, r.Y+r.H).LineTo(r.X+r.W, r.Y+r.H/2)
			default:
				path.MoveTo(r.X+r.W, r.Y).LineTo(r.X+r.W, r.Y+r.H).LineTo(r.X, r.Y+r.H/2)
			}
			p.FillPath(path.Close(), tone)
		})
	}
	wrap.Children(func() {
		if side == PlaceBottom || side == PlaceRight {
			arrow()
		}
		bubble(ui.Column(c))
		if side == PlaceTop || side == PlaceLeft {
			arrow()
		}
	})
	return side
}
