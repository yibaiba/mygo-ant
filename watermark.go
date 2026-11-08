package kit

import (
	"strings"
	"sync"

	"github.com/egoist/mygo/ui"
)

// WatermarkProps are a watermark's options, as Ant's Watermark.
type WatermarkProps struct {
	// Content is the text, one line per string.
	Content []string
	// Image tiles a picture instead, Width × Height (120 × 64); Icon an
	// SVG tinted Color.
	Image         ui.ImageSource
	Icon          *ui.SVG
	Width, Height float32
	// FontSize and Color style the text; the Feedback tokens' by default.
	FontSize float32
	Color    ui.Color
	// Rotate turns each tile, -22 degrees by default; Upright keeps them
	// level.
	Rotate  float32
	Upright bool
	// Gap is the room between tiles across and down, 100 × 100.
	Gap [2]float32
}

// Watermark is Ant's Watermark: what fn builds with text or a picture
// tiled over it in staggered, turned rows. The tiles are painted over the
// content and take no pointer events.
func Watermark(c *ui.Context, fn func(), props ...WatermarkProps) ui.Element {
	p := first(props)
	t := Of(c)
	return ui.Column(c).FillWidth().Clip().Children(fn).DrawOver(func(pa *ui.Painter, r ui.Rect) {
		paintWatermark(pa, r, t, p)
	})
}

// paintWatermark tiles p's text or picture over r in the tokens t.
func paintWatermark(pa *ui.Painter, r ui.Rect, t *Tokens, p WatermarkProps) {
	f := t.Feedback
	size := pick(p.FontSize == 0, f.WatermarkSize, p.FontSize)
	color := watermarkColor(t, p)
	deg := pick[float32](p.Rotate == 0, -22, p.Rotate)
	if p.Upright {
		deg = 0
	}
	gx := pick(p.Gap[0] == 0, f.WatermarkGap, p.Gap[0])
	gy := pick(p.Gap[1] == 0, f.WatermarkGap, p.Gap[1])
	if p.Image != nil || p.Icon != nil {
		w := pick[float32](p.Width == 0, 120, p.Width)
		h := pick[float32](p.Height == 0, 64, p.Height)
		drawTiledImage(pa, r, w, h, deg, gx, gy, func(box ui.Rect) {
			if p.Icon != nil {
				pa.Icon(p.Icon, box, color)
			} else {
				pa.Image(p.Image, box, ui.Contain)
			}
		})
		return
	}
	if len(p.Content) > 0 {
		drawTiledText(pa, r, p.Content, size, color, deg, gx, gy)
	}
}

func watermarkColor(t *Tokens, p WatermarkProps) ui.Color {
	return pick(p.Color == (ui.Color{}), t.Feedback.WatermarkText, p.Color)
}

// watermarkShows reports whether p paints anything, and why not.
func watermarkShows(t *Tokens, p WatermarkProps) string {
	switch {
	case p.Image == nil && p.Icon == nil && strings.TrimSpace(strings.Join(p.Content, "")) == "":
		return "no content"
	case p.Image == nil && watermarkColor(t, p).A < 5:
		return "transparent color"
	case p.FontSize < 0 || p.Width < 0 || p.Height < 0 || p.Gap[0] < 0 || p.Gap[1] < 0:
		return "negative size"
	}
	return ""
}

// WatermarkGuard is the state of the window watermark App draws with
// AppProps.Watermark. The app reads it through AppState.Watermark.
type WatermarkGuard struct {
	mu      sync.Mutex
	props   *WatermarkProps
	tok     *Tokens
	built   uint64
	painted uint64
	area    ui.Rect
	window  ui.Rect
	broken  string
	onBreak func(reason string)
}

// Intact reports whether the watermark was painted over the whole window
// in the last frame App built, with something visible to paint; reason
// says what is wrong when it is not. Ask between frames (a timer, a
// goroutine): during a build the frame is not painted yet.
func (g *WatermarkGuard) Intact() (ok bool, reason string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	r := g.check()
	return r == "", r
}

func (g *WatermarkGuard) check() string {
	switch {
	case g.props == nil:
		return "not set"
	case g.built == 0:
		return "not built"
	case g.painted != g.built:
		return "not painted this frame"
	case g.area.X > g.window.X || g.area.Y > g.window.Y ||
		g.area.X+g.area.W < g.window.X+g.window.W || g.area.Y+g.area.H < g.window.Y+g.window.H:
		return "does not cover the window"
	}
	return watermarkShows(g.tok, *g.props)
}

// framePasses is how many times ui builds a frame at most before it paints
// it once (runtime.go, runFrame's pass loop): fewer builds than that since
// the last paint may all be one frame's.
const framePasses = 3

// rootWatermark builds the window watermark last in App's overlay layer,
// so it paints over the page, its popups and the kit's holders. It is a
// zero-size box, which ui never hits, painting the window from its
// DrawOver, which ui does not clip; it takes no pointer events.
func rootWatermark(c *ui.Context, g *WatermarkGuard, set *WatermarkProps, onBreak func(string)) {
	g.mu.Lock()
	if set != nil && watermarkShows(Of(c), *set) == "" {
		cp := *set
		g.props = &cp
	}
	if onBreak != nil {
		g.onBreak = onBreak
	}
	var report func(string)
	if g.props != nil && g.built > 0 {
		r := g.check()
		if r == "not painted this frame" && g.built-g.painted < framePasses {
			r = ""
		}
		if r != "" && r != g.broken && g.onBreak != nil {
			report = g.onBreak
		}
		g.broken = r
	}
	if g.props == nil {
		g.mu.Unlock()
		return
	}
	g.built++
	seq := g.built
	p := *g.props
	t := Of(c)
	g.tok = t
	w, h := c.Size()
	win := ui.Rect{W: w, H: h}
	g.window = win
	g.mu.Unlock()
	if report != nil {
		report(g.broken)
	}
	ui.Overlay(c, func() {
		ui.Box(c.Key("kit.watermark")).Absolute().Left(0).Top(0).Width(0).Height(0).DrawOver(func(pa *ui.Painter, _ ui.Rect) {
			paintWatermark(pa, win, t, p)
			g.mu.Lock()
			if seq == g.built {
				g.painted, g.area = seq, win
			}
			g.mu.Unlock()
		})
	})
}

// drawTiledImage tiles w × h boxes that draw paints over r, as
// drawTiledText tiles text.
func drawTiledImage(pa *ui.Painter, r ui.Rect, w, h, deg, gx, gy float32, draw func(box ui.Rect)) {
	pa.Clip(r, 0, func() {
		for _, o := range tileOrigins(r, w, h, gx, gy) {
			pa.Transform(ui.Rotation(deg), o.X+w/2, o.Y+h/2, func() {
				draw(ui.Rect{X: o.X, Y: o.Y, W: w, H: h})
			})
		}
	})
}
