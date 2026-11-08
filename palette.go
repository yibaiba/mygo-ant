package kit

import (
	"math"

	"github.com/egoist/mygo/ui"
)

// Palette is the ten-step color scale Ant Design derives from one seed
// (@ant-design/colors generate): [0] is the lightest background, [5] the
// seed, [9] the darkest. Ant's tokens read from it: PrimaryBg is [0],
// PrimaryBorder [2], PrimaryHover [4], Primary [5], PrimaryActive [6].
type Palette [10]ui.Color

const (
	hueStep         = 2
	saturationStep  = 0.16
	saturationStep2 = 0.05
	brightnessStep1 = 0.05
	brightnessStep2 = 0.15
	lightCount      = 5
	darkCount       = 4
)

// AntPalette ports @ant-design/colors' generate for a light theme.
func AntPalette(seed ui.Color) Palette {
	h, s, v := toHSV(seed)
	var out Palette
	for i := lightCount; i > 0; i-- {
		out[lightCount-i] = fromHSV(antHue(h, i, true), antSat(h, s, i, true), antVal(v, i, true))
	}
	out[5] = ui.RGB(seed.R, seed.G, seed.B)
	for i := 1; i <= darkCount; i++ {
		out[5+i] = fromHSV(antHue(h, i, false), antSat(h, s, i, false), antVal(v, i, false))
	}
	return out
}

// antDarkMap is @ant-design/colors' darkColorMap: step i of a dark scale
// mixes the background toward light step index by opacity.
var antDarkMap = [10]struct {
	index   int
	opacity float64
}{{7, .15}, {6, .25}, {5, .3}, {5, .45}, {5, .65}, {5, .85}, {4, .9}, {3, .95}, {2, .97}, {1, .98}}

// AntDarkPalette ports generate(seed, {theme: 'dark', backgroundColor: bg}).
func AntDarkPalette(seed, bg ui.Color) Palette {
	light := AntPalette(seed)
	var out Palette
	for i, m := range antDarkMap {
		c := light[m.index]
		mix := func(a, b uint8) uint8 { return uint8(math.Round(float64(a) + (float64(b)-float64(a))*m.opacity)) }
		out[i] = ui.RGB(mix(bg.R, c.R), mix(bg.G, c.G), mix(bg.B, c.B))
	}
	return out
}

func antHue(h float64, i int, light bool) float64 {
	r := math.Round(h)
	var hue float64
	if r >= 60 && r <= 240 {
		if light {
			hue = r - hueStep*float64(i)
		} else {
			hue = r + hueStep*float64(i)
		}
	} else if light {
		hue = r + hueStep*float64(i)
	} else {
		hue = r - hueStep*float64(i)
	}
	if hue < 0 {
		hue += 360
	} else if hue >= 360 {
		hue -= 360
	}
	return hue
}

func antSat(h, s float64, i int, light bool) float64 {
	if h == 0 && s == 0 {
		return s
	}
	var sat float64
	switch {
	case light:
		sat = s - saturationStep*float64(i)
	case i == darkCount:
		sat = s + saturationStep
	default:
		sat = s + saturationStep2*float64(i)
	}
	if sat > 1 {
		sat = 1
	}
	if light && i == lightCount && sat > 0.1 {
		sat = 0.1
	}
	if sat < 0.06 {
		sat = 0.06
	}
	return math.Round(sat*100) / 100
}

func antVal(v float64, i int, light bool) float64 {
	var val float64
	if light {
		val = v + brightnessStep1*float64(i)
	} else {
		val = v - brightnessStep2*float64(i)
	}
	if val > 1 {
		val = 1
	}
	return math.Round(val*100) / 100
}

func toHSV(c ui.Color) (h, s, v float64) {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	mx := math.Max(r, math.Max(g, b))
	mn := math.Min(r, math.Min(g, b))
	d := mx - mn
	v = mx
	if mx != 0 {
		s = d / mx
	}
	if d == 0 {
		return 0, s, v
	}
	switch mx {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	return h * 60, s, v
}

func fromHSV(h, s, v float64) ui.Color {
	h = h / 60
	i := math.Floor(h)
	f := h - i
	p := v * (1 - s)
	q := v * (1 - f*s)
	t := v * (1 - (1-f)*s)
	var r, g, b float64
	switch int(i) % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	default:
		r, g, b = v, p, q
	}
	return ui.RGB(uint8(math.Round(r*255)), uint8(math.Round(g*255)), uint8(math.Round(b*255)))
}
