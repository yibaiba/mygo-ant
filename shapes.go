package kit

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// arcPoint is the point at angle deg on a circle about (cx, cy) of radius
// r, the angle going clockwise from twelve o'clock, as Ant's progress
// circles count it.
func arcPoint(cx, cy, r, deg float32) (float32, float32) {
	a := float64(deg-90) * math.Pi / 180
	return cx + r*float32(math.Cos(a)), cy + r*float32(math.Sin(a))
}

// arcPath is an open arc of a circle about (cx, cy) of radius r from
// angle start through sweep degrees clockwise (counterclockwise when
// negative), in cubic pieces of at most 90 degrees.
func arcPath(cx, cy, r, start, sweep float32) *ui.Path {
	var p ui.Path
	x, y := arcPoint(cx, cy, r, start)
	p.MoveTo(x, y)
	if sweep == 0 || r <= 0 {
		return &p
	}
	n := int(math.Ceil(math.Abs(float64(sweep)) / 90))
	step := sweep / float32(n)
	k := float32(4.0/3.0*math.Tan(float64(step)*math.Pi/720)) * r
	a := start
	for i := 0; i < n; i++ {
		b := a + step
		x0, y0 := arcPoint(cx, cy, r, a)
		x1, y1 := arcPoint(cx, cy, r, b)
		// The tangents at a and b, of length k, clockwise.
		ta := float64(a) * math.Pi / 180
		tb := float64(b) * math.Pi / 180
		p.CubeTo(
			x0+k*float32(math.Cos(ta)), y0+k*float32(math.Sin(ta)),
			x1-k*float32(math.Cos(tb)), y1-k*float32(math.Sin(tb)),
			x1, y1)
		a = b
	}
	return &p
}

// starPath is a five-pointed star about (cx, cy), its points outer from
// the middle and the notches between them inner, one point straight up.
func starPath(cx, cy, outer, inner float32) *ui.Path {
	var p ui.Path
	for i := 0; i < 10; i++ {
		r := pick(i%2 == 0, outer, inner)
		x, y := arcPoint(cx, cy, r, float32(i)*36)
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	return p.Close()
}

// ribbonFold is how far a Badge.Ribbon hangs past its card's edge, the
// size of the folded corner under it.
const ribbonFold = 8

// drawRibbon paints a ribbon label filling r, sticking out ribbonFold past
// the card's edge on its end (start when atStart), with the fold below
// that end in a darker shade.
func drawRibbon(p *ui.Painter, r ui.Rect, tone ui.Color, atStart bool) {
	p.Fill(r, tone, 4)
	var fold ui.Path
	if atStart {
		p.Fill(ui.Rect{X: r.X, Y: r.Y, W: 4, H: r.H}, tone, 0)
		fold.MoveTo(r.X, r.Y+r.H).LineTo(r.X+ribbonFold, r.Y+r.H).LineTo(r.X+ribbonFold, r.Y+r.H+ribbonFold)
	} else {
		p.Fill(ui.Rect{X: r.X + r.W - 4, Y: r.Y, W: 4, H: r.H}, tone, 0)
		fold.MoveTo(r.X+r.W, r.Y+r.H).LineTo(r.X+r.W-ribbonFold, r.Y+r.H).LineTo(r.X+r.W-ribbonFold, r.Y+r.H+ribbonFold)
	}
	p.FillPath(fold.Close(), tone.Mix(ui.RGB(0, 0, 0), 0.35))
}

// tileOrigins are the top-left corners of the tiles, w×h with gaps gx and
// gy between them, that cover r with the tiles of every other row offset
// by half a step, as Ant's Watermark staggers them.
func tileOrigins(r ui.Rect, w, h, gx, gy float32) []ui.Point {
	sx, sy := w+gx, h+gy
	if sx <= 0 || sy <= 0 {
		return nil
	}
	var out []ui.Point
	row := 0
	for y := r.Y - sy/2; y < r.Y+r.H; y += sy {
		shift := pick(row%2 == 1, sx/2, 0)
		for x := r.X - sx + shift; x < r.X+r.W; x += sx {
			out = append(out, ui.Point{X: x, Y: y})
		}
		row++
	}
	return out
}

// drawTiledText paints lines of text, each line size px, rotated deg
// degrees, tiled over r in staggered rows with gaps gx and gy, clipped to
// r: a watermark.
func drawTiledText(p *ui.Painter, r ui.Rect, lines []string, size float32, c ui.Color, deg, gx, gy float32) {
	var w float32
	for _, l := range lines {
		lw, _ := p.MeasureText(0, ui.Span{Text: l, Size: size})
		w = max(w, lw)
	}
	lh := size * 1.4
	h := lh * float32(len(lines))
	p.Clip(r, 0, func() {
		for _, o := range tileOrigins(r, w, h, gx, gy) {
			cx, cy := o.X+w/2, o.Y+h/2
			p.Transform(ui.Rotation(deg), cx, cy, func() {
				for i, l := range lines {
					lw, _ := p.MeasureText(0, ui.Span{Text: l, Size: size})
					p.Text(cx-lw/2, o.Y+float32(i)*lh, l, size, c)
				}
			})
		}
	})
}
