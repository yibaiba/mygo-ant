package kit

import (
	"math"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// BadgeProps are a count badge's options.
type BadgeProps struct {
	Count int
	// Text, when set, shows instead of the count, as "NEW".
	Text string
	// Overflow caps the number shown, as "99+"; 99 when zero.
	Overflow int
	// Dot shows a 6px dot instead of the number.
	Dot bool
	// ShowZero keeps a zero count showing.
	ShowZero bool
	// Offset moves the badge right and down from the corner, as Ant's
	// [x, y].
	Offset [2]float32
	// Color fills the badge; Preset names one of Ant's preset or status
	// colors instead. Red (colorError) by default.
	Color  ui.Color
	Preset string
	// Size Small is Ant's small badge: 14 high.
	Size Size
}

// badgeText is what a count badge shows: Text, the count capped as
// "99+", or nothing (hidden) for a zero count without ShowZero.
func badgeText(p BadgeProps) (string, bool) {
	if p.Text != "" {
		return p.Text, true
	}
	if p.Count <= 0 && !p.ShowZero && !p.Dot {
		return "", false
	}
	if p.Dot {
		return "", p.Count > 0 || p.ShowZero
	}
	over := pick(p.Overflow == 0, 99, p.Overflow)
	if p.Count > over {
		return strconv.Itoa(over) + "+", true
	}
	return strconv.Itoa(p.Count), true
}

// badgeBox is the size of a count badge showing text: Ant's 20 (14
// small) high pill, a circle for one character.
func badgeBox(c *ui.Context, t *Tokens, text string, small, dot bool) (w, h float32) {
	if dot {
		return 6, 6
	}
	h = pick[float32](small, 14, 20)
	tw, _ := c.MeasureText(0, ui.Span{Text: text, Size: t.FontSizeSM})
	pad := pick[float32](small, 4, 8)
	if len([]rune(text)) <= 1 {
		return max(h, tw), h
	}
	return max(h, tw+pad*2), h
}

// BadgeCount is Ant's Badge with a count or dot: on the top right corner
// of what child builds, or inline when child is nil. A changed count
// rolls in from below.
func BadgeCount(c *ui.Context, child func(), props ...BadgeProps) {
	p := first(props)
	t := Of(c)
	text, show := badgeText(p)
	tone := p.Color
	if tone == (ui.Color{}) {
		tone = toneColor(t, p.Preset, t.Error)
	}
	w, h := badgeBox(c, t, text, p.Size == Small, p.Dot)
	pill := func(e ui.Element) {
		e.Size(w, h).Radius(h/2).Background(tone).Shadow(0, 0, 0, 1, t.BadgeShadow).Clip().
			AlignItems(ui.Center).Justify(ui.Center).Shrink(0)
		if p.Dot {
			return
		}
		e.Children(func() {
			roll := ui.Local(e, "roll", func() int { return p.Count })
			if *roll != p.Count {
				e.Jump("roll", pick[float32](p.Count > *roll, 1, -1))
				*roll = p.Count
			}
			y := e.AnimateTo("roll", 0, ui.Animation{Duration: motionFor(c, 300*time.Millisecond), Ease: ui.EaseOut})
			ui.Text(c, text).FontSize(t.FontSizeSM).TextColor(ui.Hex("#FFFFFF")).SingleLine().Translate(0, y*h)
		})
	}
	if child == nil {
		if show {
			pill(ui.Row(c).AlignSelf(ui.Center))
		}
		return
	}
	ui.Row(c).Shrink(0).AlignSelf(ui.Start).Children(func() {
		child()
		if !show {
			return
		}
		e := ui.Row(c).Absolute().Top(-h/2 + p.Offset[1]).Right(-w/2 - p.Offset[0]).PassThrough()
		in := enterIn(e, "zoom", motionFor(c, 200*time.Millisecond))
		e.Scale(in, in)
		pill(e)
	})
}

// StatusDotProps are a status badge's options.
type StatusDotProps struct {
	// Color, or Preset by name, overrides the status color.
	Color  ui.Color
	Preset string
}

// StatusDot is Ant's status Badge: a 6px dot and text. status is
// "success", "processing" (blue, with a spreading ripple), "error",
// "warning" or "default"; the older Badge is its success / error pair in
// the default colors.
func StatusDot(c *ui.Context, status, text string, props ...StatusDotProps) {
	p := first(props)
	t := Of(c)
	tone := p.Color
	if tone == (ui.Color{}) {
		tone = toneColor(t, pick(p.Preset == "", status, p.Preset), t.TextDisabled)
	}
	ripple := status == "processing"
	ui.Row(c).Gap(8).AlignItems(ui.Center).Shrink(0).Children(func() {
		ui.Box(c).Size(6, 6).Shrink(0).Draw(func(pa *ui.Painter, r ui.Rect) {
			pa.Fill(r, tone, 3)
			if ripple && !stillMotion(c) {
				phase := rippleAt(pa.Now())
				rr := 3 * (0.8 + 1.6*phase)
				cx, cy := r.X+3, r.Y+3
				pa.Stroke(ui.Rect{X: cx - rr, Y: cy - rr, W: rr * 2, H: rr * 2}, tone.Alpha(0.5*(1-phase)), rr, 1)
				pa.AnimationFrame()
			}
		})
		if text != "" {
			ui.Text(c, text).FontSize(t.FontSize).TextColor(t.Text).SingleLine()
		}
	})
}

// rippleAt is how far a processing badge's ripple has spread at now, 0 to
// 1 over Ant's 1.2s loop.
func rippleAt(now time.Time) float32 {
	return float32(math.Mod(float64(now.UnixMilli()), 1200)) / 1200
}

// RibbonProps are a ribbon's options.
type RibbonProps struct {
	// Color, or Preset by name, fills the ribbon; primary by default.
	Color  ui.Color
	Preset string
	// Start hangs it on the start edge instead of the end.
	Start bool
}

// Ribbon is Ant's Badge.Ribbon: a label hanging over the top end edge of
// what fn builds, with its folded corner.
func Ribbon(c *ui.Context, text string, fn func(), props ...RibbonProps) {
	p := first(props)
	t := Of(c)
	tone := p.Color
	if tone == (ui.Color{}) {
		tone = toneColor(t, p.Preset, t.Primary)
	}
	ui.Column(c).FillWidth().Children(func() {
		fn()
		e := ui.Row(c).Absolute().Top(8).Height(22).Padding(0, 8).AlignItems(ui.Center).PassThrough()
		if p.Start {
			e.Left(-ribbonFold)
		} else {
			e.Right(-ribbonFold)
		}
		e.Draw(func(pa *ui.Painter, r ui.Rect) { drawRibbon(pa, r, tone, p.Start) }).Children(func() {
			ui.Text(c, text).FontSize(t.FontSize).TextColor(ui.Hex("#FFFFFF")).SingleLine()
		})
	})
}
