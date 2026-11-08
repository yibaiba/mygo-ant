package kit

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// chromePad is the extra left / right padding a bar needs so its content
// clears the system window controls (traffic lights, caption buttons).
func chromePad(c *ui.Context, left, right float32) (float32, float32) {
	bar := c.TitleBar()
	return max(left, bar.Left), max(right, bar.Right)
}

// WindowButtonsProps are the window min / max / close cluster of a
// hidden-title-bar chrome. Callbacks are what the app does (mygo.Window
// Minimize, ToggleMaximize, Hide / Close).
type WindowButtonsProps struct {
	OnMinimize, OnMaximize, OnClose func()
	// CloseLabel is the close button's name; the locale's Close when empty
	// (the product passes "隐藏" when the button hides to the tray).
	CloseLabel string
	// Light paints white glyphs on a brand-coloured bar.
	Light bool
	// Width of each button; 42 when zero, 48 on a 40px brand bar is typical.
	Width float32
	// Height of the cluster; 40 when zero.
	Height float32
}

// WindowButtons is product window chrome: min, max and close with Ant
// tooltips, using the app's window callbacks. Ant has no window-control
// widget; LayoutHeader + Button is the in-page counterpart, not a fake
// browser caption bar. Leave WindowButtons as the only window controls.
func WindowButtons(c *ui.Context, props WindowButtonsProps) {
	p := props
	w := pick(p.Width == 0, float32(42), p.Width)
	h := pick(p.Height == 0, float32(40), p.Height)
	closeLabel := pick(p.CloseLabel == "", L(c).Close, p.CloseLabel)
	icon := ui.Hex("#6B7280")
	switch {
	case p.Light:
		icon = ui.Hex("#FFFFFF")
	case Of(c).Dark:
		icon = Of(c).TextMuted
	}
	if windowBtn(c, L(c).Minimize, w, h, p.Light, false, func() {
		windowMinGlyph(c, icon)
	}) && p.OnMinimize != nil {
		p.OnMinimize()
	}
	if windowBtn(c, L(c).Maximize, w, h, p.Light, true, func() {
		windowMaxGlyph(c, icon)
	}) && p.OnMaximize != nil {
		p.OnMaximize()
	}
	if windowBtn(c, closeLabel, w, h, p.Light, false, func() {
		ui.Text(c, "×").FontSize(16).TextColor(icon)
	}) && p.OnClose != nil {
		p.OnClose()
	}
}

// snapDIP is dip rounded onto the zoom's pixel grid (PixelRatio). Painter
// Stroke/Fill then snap that length onto the host's device pixels, so a
// 1.5 DIP window-button hairline is a whole number of device pixels
// rather than a half-pixel blend. Other 1px kit borders stay as they are.
func snapDIP(c *ui.Context, dip float32) float32 {
	s := PixelRatio(c)
	if s <= 0 {
		s = 1
	}
	px := math.Round(float64(dip) * float64(s))
	if px < 1 {
		px = 1
	}
	return float32(px) / s
}

func windowMinGlyph(c *ui.Context, icon ui.Color) {
	ui.Box(c).Width(snapDIP(c, 12)).Height(snapDIP(c, 2)).Draw(func(p *ui.Painter, r ui.Rect) {
		h := float32(1.5)
		p.Fill(ui.Rect{X: r.X, Y: r.Y + (r.H-h)/2, W: r.W, H: h}, icon, 0)
	})
}

func windowMaxGlyph(c *ui.Context, icon ui.Color) {
	s := snapDIP(c, 10)
	ui.Box(c).Size(s, s).Draw(func(p *ui.Painter, r ui.Rect) {
		p.Stroke(r, icon, 0, 1.5)
	})
}

// windowBtn is one caption button: a white veil on a brand bar, a grey
// one on a light bar, with a tooltip. maximize marks the Win11 snap-
// layouts hit target (Element.MaximizeButton); only the restore/maximize
// control is marked.
func windowBtn(c *ui.Context, label string, width, height float32, light, maximize bool, icon func()) bool {
	btn := ui.ButtonBase(c).Label(label).Width(width).Height(height).Center().FocusRing(false)
	if maximize {
		btn.MaximizeButton()
	}
	if btn.Hovered() {
		switch {
		case light:
			btn.Background(ui.RGBA(255, 255, 255, 0.1))
		case Of(c).Dark:
			btn.Radius(6).Background(ui.RGBA(255, 255, 255, 0.08))
		default:
			btn.Radius(6).Background(ui.RGBA(15, 23, 42, 0.06))
		}
	}
	if btn.FocusVisible() {
		ring := Of(c).PrimaryBorder
		if light {
			ring = ui.RGBA(255, 255, 255, 0.7)
		}
		btn.Radius(6).InsetShadow(0, 0, 0, 2, ring)
	}
	TooltipOn(c, btn, label)
	btn.Children(icon)
	return btn.Clicked()
}
