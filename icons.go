package kit

import "github.com/egoist/mygo/ui"

// ui.Icon tints a whole SVG one color, so two-tone marks are painted.
var (
	iconClose = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path d="m3.5 3.5 9 9m0-9-9 9" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"/></svg>`))
	iconEmpty = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 40 28" fill="none"><path d="M6 12 11 3h18l5 9v12a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V12Z" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"/><path d="M6 12h9a5 5 0 0 0 10 0h9" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round"/></svg>`))

	iconEye    = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3"><path d="M1.5 8S4 3.5 8 3.5 14.5 8 14.5 8 12 12.5 8 12.5 1.5 8 1.5 8Z" stroke-linejoin="round"/><circle cx="8" cy="8" r="2.2"/></svg>`))
	iconEyeOff = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"><path d="M2.5 6.2C1.9 7.1 1.5 8 1.5 8S4 12.5 8 12.5c.9 0 1.8-.2 2.5-.6M6.2 3.8C6.8 3.6 7.4 3.5 8 3.5c4 0 6.5 4.5 6.5 4.5s-.6 1.1-1.7 2.2"/><path d="m2 2 12 12"/></svg>`))
	iconSearch = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"><circle cx="7" cy="7" r="4.5"/><path d="m10.5 10.5 3.5 3.5"/></svg>`))
)

// statusIcon is Ant's filled status mark: a disc with a white glyph.
func statusIcon(c *ui.Context, k AlertType, size float32, tone ui.Color) {
	ui.Box(c).Size(size, size).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		p.Fill(r, tone, r.W/2)
		white := ui.RGB(255, 255, 255)
		u := r.W / 16
		x, y := r.X, r.Y
		w := 1.6 * u
		switch k {
		case Success:
			var path ui.Path
			path.MoveTo(x+4.8*u, y+8.2*u).LineTo(x+7*u, y+10.4*u).LineTo(x+11.2*u, y+6*u)
			p.StrokePath(&path, w, white)
		case Error:
			var a, b ui.Path
			a.MoveTo(x+5.4*u, y+5.4*u).LineTo(x+10.6*u, y+10.6*u)
			b.MoveTo(x+10.6*u, y+5.4*u).LineTo(x+5.4*u, y+10.6*u)
			p.StrokePath(&a, w, white)
			p.StrokePath(&b, w, white)
		case Warning:
			p.Line(x+8*u, y+4.2*u, x+8*u, y+9.2*u, w, white)
			p.Fill(ui.Rect{X: x + 7.1*u, Y: y + 10.6*u, W: 1.8 * u, H: 1.8 * u}, white, 0.9*u)
		default:
			p.Fill(ui.Rect{X: x + 7.1*u, Y: y + 3.8*u, W: 1.8 * u, H: 1.8 * u}, white, 0.9*u)
			p.Line(x+8*u, y+6.8*u, x+8*u, y+11.8*u, w, white)
		}
	})
}
