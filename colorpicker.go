package kit

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// ColorPreset is a titled row of swatches in a ColorPicker's panel.
type ColorPreset struct {
	Label  string
	Colors []ui.Color
}

// ColorFormat is how a ColorPicker writes its color.
type ColorFormat uint8

// Color formats.
const (
	FormatHex ColorFormat = iota
	FormatRGB
	FormatHSB
)

// ColorPickerProps are a color picker's options.
type ColorPickerProps struct {
	Size     Size
	Disabled bool
	// ShowText writes the color beside the swatch.
	ShowText bool
	Format   ColorFormat
	Presets  []ColorPreset
	// AllowClear adds a "no color" swatch that makes the color transparent.
	AllowClear bool

	open bool
}

// colorHex is c as #RRGGBB, or #RRGGBBAA when not opaque.
func colorHex(c ui.Color) string {
	if c.A == 255 {
		return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
	}
	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
}

// colorRGB is c as Ant writes it in rgb format.
func colorRGB(c ui.Color) string {
	if c.A == 255 {
		return fmt.Sprintf("rgb(%d,%d,%d)", c.R, c.G, c.B)
	}
	return fmt.Sprintf("rgba(%d,%d,%d,%s)", c.R, c.G, c.B, strconv.FormatFloat(float64(c.A)/255, 'f', 2, 64))
}

// parseHex reads #RGB, #RRGGBB or #RRGGBBAA, the # optional.
func parseHex(s string) (ui.Color, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 && len(s) != 8 {
		return ui.Color{}, false
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return ui.Color{}, false
	}
	if len(s) == 6 {
		return ui.RGB(uint8(v>>16), uint8(v>>8), uint8(v)), true
	}
	return ui.RGBA(uint8(v>>24), uint8(v>>16), uint8(v>>8), float32(uint8(v))/255), true
}

// colorHSB is c as Ant writes it in hsb format (brightness = HSV value).
func colorHSB(c ui.Color) string {
	h, s, v := toHSV(c)
	hs, vs := int(h+0.5), int(v*100+0.5)
	ss := int(s*100 + 0.5)
	if c.A == 255 {
		return fmt.Sprintf("hsb(%d,%d%%,%d%%)", hs, ss, vs)
	}
	return fmt.Sprintf("hsba(%d,%d%%,%d%%,%s)", hs, ss, vs, strconv.FormatFloat(float64(c.A)/255, 'f', 2, 64))
}

func colorText(c ui.Color, format ColorFormat) string {
	switch format {
	case FormatRGB:
		return colorRGB(c)
	case FormatHSB:
		return colorHSB(c)
	}
	return colorHex(c)
}

func formatTag(format ColorFormat) string {
	switch format {
	case FormatRGB:
		return "rgb"
	case FormatHSB:
		return "hsb"
	}
	return "hex"
}

func formatOf(tag string) ColorFormat {
	switch tag {
	case "rgb":
		return FormatRGB
	case "hsb":
		return FormatHSB
	}
	return FormatHex
}

// ColorPicker is Ant's ColorPicker: a swatch that opens a panel with a
// saturation and brightness square, hue and alpha bars, the color in hex,
// rgb or hsb and its alpha in percent, which the user edits, and preset
// swatches. It reports a change of *color.
func ColorPicker(c *ui.Context, color *ui.Color, props ...ColorPickerProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	changed := false
	h := t.Height(p.Size)
	set := func(col ui.Color) {
		if col != *color {
			*color = col
			changed = true
		}
	}
	Popup(c, func() bool {
		b := ui.ButtonBase(c).Label("选择颜色").Height(h).Padding(3).Gap(t.SpaceSM).Radius(t.RadiusControl).
			AlignItems(ui.Center).FocusRing(false).Disabled(p.Disabled).Value(colorHex(*color)).
			Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
		if p.ShowText {
			b.Padding(3, 8, 3, 3)
		}
		fieldLook(b, t, stateOf(b, p.Disabled, rest), StatusNone)
		col := *color
		b.Children(func() {
			colorSwatch(c, col, h-8, t)
			if p.ShowText {
				ui.Text(c, colorText(col, p.Format)).FontSize(t.FontSize).TextColor(t.Text).SingleLine()
			}
		})
		return b.Clicked() && !p.Disabled
	}, func(panel ui.Element, _ func()) {
		popupLook(panel.Padding(12).Gap(t.SpaceSM).Width(280), t)
		panel.Children(func() {
			// The hue is kept apart from the color: a grey or black color
			// has none, and the square must not lose it while dragged there.
			hsv := ui.Local(panel, "hsv", func() [3]float64 { h, s, v := toHSV(*color); return [3]float64{h, s, v} })
			if hsva(hsv[0], hsv[1], hsv[2], color.A) != *color {
				h, s, v := toHSV(*color)
				if s == 0 || v == 0 {
					h = hsv[0]
				}
				*hsv = [3]float64{h, s, v}
			}
			apply := func() { set(hsva(hsv[0], hsv[1], hsv[2], color.A)) }
			if sx, vy, ok := colorSquare(c, t, hsv[0], hsv[1], hsv[2]); ok {
				hsv[1], hsv[2] = sx, vy
				apply()
			}
			ui.Row(c).FillWidth().Gap(t.SpaceSM).AlignItems(ui.Center).Children(func() {
				ui.Column(c).Grow(1).Gap(t.SpaceSM).Children(func() {
					if f, ok := colorBar(c.Key("hue"), t, "色相", hsv[0]/360, hueGradient(), ui.Color{}); ok {
						hsv[0] = f * 360
						apply()
					}
					opaque := *color
					opaque.A = 255
					if f, ok := colorBar(c.Key("alpha"), t, "透明度", float64(color.A)/255, ui.LinearGradient{Angle: 90, From: opaque.Alpha(0), To: opaque}, opaque); ok {
						col := *color
						col.A = uint8(f*255 + 0.5)
						set(col)
					}
				})
				colorSwatch(c, *color, 28, t)
			})
			format := ui.Local(panel, "format", func() string { return formatTag(p.Format) })
			ui.Row(c).FillWidth().Gap(t.SpaceXS).AlignItems(ui.Center).Children(func() {
				Segmented(c, format, []Option{{Value: "hex", Label: "HEX"}, {Value: "hsb", Label: "HSB"}, {Value: "rgb", Label: "RGB"}}, SegmentedProps{Size: Small})
				shown := colorText(*color, formatOf(*format))
				if *format == "hex" {
					shown = colorHex(opaqueOf(*color))
				}
				if col, ok := colorTextInput(c.Key("text"), t, "颜色值", shown, parseColorText, 0); ok {
					if *format == "hex" && col.A == 255 {
						col.A = color.A
					}
					set(col)
				}
				if *format == "hex" {
					pct := strconv.Itoa(int(float64(color.A)/255*100+0.5)) + "%"
					if col, ok := colorTextInput(c.Key("pct"), t, "透明度百分比", pct, func(s string) (ui.Color, bool) {
						n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(s), "%"))
						if err != nil || n < 0 || n > 100 {
							return ui.Color{}, false
						}
						out := *color
						out.A = uint8(float64(n)/100*255 + 0.5)
						return out, true
					}, 52); ok {
						set(col)
					}
				}
			})
			swatches := p.Presets
			if p.AllowClear {
				swatches = append([]ColorPreset{{Label: "", Colors: []ui.Color{ui.Transparent}}}, swatches...)
			}
			for i, pr := range swatches {
				ui.Column(c.Key(i)).FillWidth().Gap(6).Children(func() {
					if pr.Label != "" {
						ui.Text(c, pr.Label).FontSize(t.FontSizeSM).TextColor(t.TextTertiary)
					}
					ui.Row(c).FillWidth().Gap(6).Wrap().Children(func() {
						for j, sw := range pr.Colors {
							b := ui.ButtonBase(c.Key(j)).Label(colorHex(sw)).Size(20, 20).Radius(4).FocusRing(false).Cursor(ui.CursorPointer)
							if sw == *color {
								b.Shadow(0, 0, 0, 2, t.PrimaryBorder)
							}
							b.Children(func() { colorSwatch(c, sw, 20, t) })
							if b.Clicked() {
								set(sw)
							}
						}
					})
				})
			}
		})
	}, PopupProps{Placement: PlaceBottom, Align: AlignFirst, open: p.open})
	return changed
}

func opaqueOf(c ui.Color) ui.Color {
	c.A = 255
	return c
}

// colorTextInput is a small field showing shown that, while it has the
// focus, keeps what the user types and reports the color parse reads from
// it. width 0 grows it.
func colorTextInput(c *ui.Context, t *Tokens, label, shown string, parse func(string) (ui.Color, bool), width float32) (ui.Color, bool) {
	host := ui.Row(c)
	if width > 0 {
		host.Width(width).Shrink(0)
	} else {
		host.Grow(1).MinWidth(0)
	}
	var out ui.Color
	ok := false
	host.Children(func() {
		draft := ui.Local(host, "draft", func() string { return "" })
		editing := ui.Local(host, "editing", func() bool { return false })
		if !*editing {
			*draft = shown
		}
		in := ui.TextInputBase(c, draft).Height(t.ControlHeightSM).Padding(0, 7).FontSize(t.FontSizeSM).
			TextColor(t.Text).Radius(t.RadiusControl).FocusRing(false).FillWidth().Label(label)
		fieldLook(in, t, pick(in.FocusVisible(), focus, rest), StatusNone)
		*editing = in.Focused()
		if in.Changed() {
			out, ok = parse(*draft)
		}
	})
	return out, ok
}

// Sizes of the ColorPicker panel, as Ant's: a 234 × 160 square and 8px
// bars with 16px handles.
const (
	colorSquareH = 160
	colorBarH    = 8
	colorHandle  = 16
)

// colorSquare is the saturation (across) and brightness (up) square of
// hue: it reports the s and v a press or drag on it picks.
func colorSquare(c *ui.Context, t *Tokens, hue, s, v float64) (float64, float64, bool) {
	sq := ui.Box(c).FillWidth().Height(colorSquareH).Radius(6).Clip().Cursor(ui.CursorCrosshair).Label("饱和度与亮度")
	sq.Children(func() {
		ui.Box(c).Fill().Cache().Draw(func(p *ui.Painter, r ui.Rect) {
			p.Fill(r, fromHSV(hue, 1, 1), 0)
			p.FillGradient(r, ui.LinearGradient{Angle: 90, From: ui.RGB(255, 255, 255), To: ui.RGBA(255, 255, 255, 0)}, 0)
			p.FillGradient(r, ui.LinearGradient{Angle: 180, From: ui.RGBA(0, 0, 0, 0), To: ui.RGB(0, 0, 0)}, 0)
		})
	})
	sq.DrawOver(func(p *ui.Painter, r ui.Rect) {
		x, y := r.X+float32(s)*r.W, r.Y+float32(1-v)*r.H
		ring := ui.Rect{X: x - colorHandle/2, Y: y - colorHandle/2, W: colorHandle, H: colorHandle}
		p.Shadow(ring, colorHandle/2, 0, 0, 2, 0, ui.RGBA(0, 0, 0, 0.3))
		p.Stroke(ring, ui.RGB(255, 255, 255), colorHandle/2, 2)
	})
	fx, fy, ok := padPointer(sq)
	return float64(fx), 1 - float64(fy), ok
}

// colorBar is a hue or alpha bar painted with g (over a checkerboard when
// base is set) and a handle at f, 0 to 1. It reports the fraction a
// press or drag picks.
func colorBar(c *ui.Context, t *Tokens, label string, f float64, g ui.LinearGradient, base ui.Color) (float64, bool) {
	bar := ui.Box(c).FillWidth().Height(colorHandle).Cursor(ui.CursorPointer).Label(label).Range(0, 1, f)
	bar.Children(func() {
		ui.Box(c).Absolute().Left(0).Top((colorHandle - colorBarH) / 2).Right(0).Height(colorBarH).
			Radius(colorBarH / 2).Clip().Cache().Draw(func(p *ui.Painter, r ui.Rect) {
			if base != (ui.Color{}) {
				checker(p, r, 4)
			}
			p.FillGradient(r, g, 0)
		})
	})
	bar.DrawOver(func(p *ui.Painter, r ui.Rect) {
		rail := ui.Rect{X: r.X, Y: r.Y + (r.H-colorBarH)/2, W: r.W, H: colorBarH}
		x := rail.X + float32(f)*(rail.W-colorHandle) + colorHandle/2
		knob := ui.Rect{X: x - colorHandle/2, Y: r.Y, W: colorHandle, H: colorHandle}
		p.Shadow(knob, colorHandle/2, 0, 0, 2, 0, ui.RGBA(0, 0, 0, 0.3))
		p.Fill(knob, ui.RGB(255, 255, 255), colorHandle/2)
		inner := ui.Rect{X: knob.X + 3, Y: knob.Y + 3, W: colorHandle - 6, H: colorHandle - 6}
		if base != (ui.Color{}) {
			p.Fill(inner, base.Alpha(float32(f)), (colorHandle-6)/2)
		} else {
			p.Fill(inner, fromHSV(f*360, 1, 1), (colorHandle-6)/2)
		}
	})
	fx, _, ok := padPointer(bar)
	return float64(fx), ok
}

// padPointer follows a press on e and the drag after it: where the
// pointer is across and down e, each 0 to 1.
func padPointer(e ui.Element) (float32, float32, bool) {
	at := ui.Local(e, "at", func() [2]float32 { return [2]float32{-1, -1} })
	if !e.Pressed() {
		at[0] = -1
		return 0, 0, false
	}
	if at[0] < 0 {
		x, y, _ := e.PointerPosition()
		*at = [2]float32{x, y}
	} else if dx, dy, ok := e.Dragged(); ok {
		at[0] += dx
		at[1] += dy
	} else {
		return 0, 0, false
	}
	b := e.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return 0, 0, false
	}
	clamp := func(v float32) float32 { return max(0, min(1, v)) }
	return clamp(at[0] / b.W), clamp(at[1] / b.H), true
}

// hueGradient runs through the hues left to right.
func hueGradient() ui.LinearGradient {
	g := ui.LinearGradient{Angle: 90}
	for i := 0; i <= 6; i++ {
		g.Stops = append(g.Stops, ui.GradientStop{Color: fromHSV(float64(i)*60, 1, 1), At: float32(i) / 6})
	}
	return g
}

// hsva is fromHSV with alpha a.
func hsva(h, s, v float64, a uint8) ui.Color {
	c := fromHSV(h, s, v)
	c.A = a
	return c
}

// checker paints a grey checkerboard of q-sized squares over r, as Ant
// shows under transparent colors.
func checker(p *ui.Painter, r ui.Rect, q float32) {
	p.Fill(r, ui.RGB(255, 255, 255), 0)
	for y := float32(0); y < r.H; y += q {
		for x := float32(0); x < r.W; x += q {
			if int(x/q+y/q)%2 == 0 {
				p.Fill(ui.Rect{X: r.X + x, Y: r.Y + y, W: min(q, r.W-x), H: min(q, r.H-y)}, ui.RGB(230, 230, 230), 0)
			}
		}
	}
}

// parseColorText reads a color typed in hex, rgb()/rgba() or hsb()/hsba().
func parseColorText(s string) (ui.Color, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	if strings.HasPrefix(s, "hsb") {
		return parseHSB(s)
	}
	if strings.HasPrefix(s, "rgb") {
		open, end := strings.IndexByte(s, '('), strings.IndexByte(s, ')')
		if open < 0 || end < open {
			return ui.Color{}, false
		}
		parts := strings.Split(s[open+1:end], ",")
		if len(parts) < 3 || len(parts) > 4 {
			return ui.Color{}, false
		}
		var ch [3]uint8
		for i := 0; i < 3; i++ {
			v, err := strconv.Atoi(strings.TrimSpace(parts[i]))
			if err != nil || v < 0 || v > 255 {
				return ui.Color{}, false
			}
			ch[i] = uint8(v)
		}
		a := 1.0
		if len(parts) == 4 {
			f, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
			if err != nil || f < 0 || f > 1 {
				return ui.Color{}, false
			}
			a = f
		}
		return ui.RGBA(ch[0], ch[1], ch[2], float32(a)), true
	}
	return parseHex(s)
}

func parseHSB(s string) (ui.Color, bool) {
	open, end := strings.IndexByte(s, '('), strings.IndexByte(s, ')')
	if open < 0 || end < open {
		return ui.Color{}, false
	}
	parts := strings.Split(s[open+1:end], ",")
	if len(parts) < 3 || len(parts) > 4 {
		return ui.Color{}, false
	}
	h, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil || h < 0 || h > 360 {
		return ui.Color{}, false
	}
	sat, ok := parsePercentUnit(parts[1])
	if !ok {
		return ui.Color{}, false
	}
	bri, ok := parsePercentUnit(parts[2])
	if !ok {
		return ui.Color{}, false
	}
	a := 1.0
	if len(parts) == 4 {
		f, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil || f < 0 || f > 1 {
			return ui.Color{}, false
		}
		a = f
	}
	return hsva(h, sat, bri, uint8(a*255+0.5)), true
}

// parsePercentUnit reads 0–1, 0–100, or a trailing %.
func parsePercentUnit(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	pct := strings.HasSuffix(s, "%")
	s = strings.TrimSuffix(s, "%")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v < 0 {
		return 0, false
	}
	if pct || v > 1 {
		if v > 100 {
			return 0, false
		}
		return v / 100, true
	}
	return v, true
}

// colorSwatch is a square of col over a checkerboard, as transparent
// colors show, with a red slash for none.
func colorSwatch(c *ui.Context, col ui.Color, size float32, t *Tokens) {
	edge := t.Split
	ui.Box(c).Size(size, size).Shrink(0).Radius(4).Draw(func(p *ui.Painter, r ui.Rect) {
		p.Clip(r, 4, func() {
			if col.A < 255 {
				p.Fill(r, ui.RGB(255, 255, 255), 0)
				q := r.W / 4
				for y := 0; y < 4; y++ {
					for x := 0; x < 4; x++ {
						if (x+y)%2 == 0 {
							p.Fill(ui.Rect{X: r.X + float32(x)*q, Y: r.Y + float32(y)*q, W: q, H: q}, ui.RGB(230, 230, 230), 0)
						}
					}
				}
			}
			p.Fill(r, col, 0)
			if col.A == 0 {
				p.Line(r.X+r.W, r.Y, r.X, r.Y+r.H, 2, ui.Hex("#ff4d4f"))
			}
		})
		p.Stroke(r, edge, 4, 1)
	})
}
