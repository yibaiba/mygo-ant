package kit

import (
	"math"
	"strconv"

	"github.com/egoist/mygo/ui"
)

// SliderMark is a labelled point under a slider's rail.
type SliderMark struct {
	Value float64
	Label string
}

// SliderTooltip says when a slider shows its value over the handle.
type SliderTooltip uint8

// Slider tooltip modes.
const (
	// TooltipAuto shows it while the handle is hovered, focused or dragged.
	TooltipAuto SliderTooltip = iota
	// TooltipAlways shows it all the time.
	TooltipAlways
	// TooltipNever never shows it.
	TooltipNever
)

// SliderProps are a slider's options.
type SliderProps struct {
	// Min and Max bound the value (0 and 100 when both are zero); Step is
	// what it moves by (1 when zero).
	Min, Max, Step float64
	Marks          []SliderMark
	// Dots draws a dot at every mark; MarksOnly lets the value rest only
	// on marks, as Ant's step null.
	Dots, MarksOnly bool
	// Vertical stands it up, Length tall (200 when zero); Length is its
	// width when horizontal, filling the parent when zero.
	Vertical bool
	Length   float32
	Disabled bool
	Tooltip  SliderTooltip
	// Format turns the value into the tooltip text; the number when nil.
	Format func(float64) string
	// Excluded leaves the rail between the start and the handle grey, as
	// Ant's included={false}.
	Excluded bool

	force visual
}

func (p SliderProps) bounds() (lo, hi, step float64) {
	lo, hi, step = p.Min, p.Max, p.Step
	if lo == 0 && hi == 0 {
		hi = 100
	}
	if step <= 0 {
		step = 1
	}
	return
}

// snapSlider is the value a slider takes for v: within lo..hi, on a step
// from lo, or, with marksOnly, on the nearest mark.
func snapSlider(v, lo, hi, step float64, marks []SliderMark, marksOnly bool) float64 {
	v = math.Max(lo, math.Min(hi, v))
	if marksOnly && len(marks) > 0 {
		best := marks[0].Value
		for _, m := range marks[1:] {
			if math.Abs(m.Value-v) < math.Abs(best-v) {
				best = m.Value
			}
		}
		return best
	}
	n := math.Round((v - lo) / step)
	out := lo + n*step
	if out > hi {
		out -= step
	}
	// Steps like 0.1 add up float error; keep the step's decimals.
	return roundTo(out, decimals(step))
}

// decimals is how many digits a step has after the point.
func decimals(step float64) int {
	s := strconv.FormatFloat(step, 'f', -1, 64)
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			return len(s) - i - 1
		}
	}
	return 0
}

func roundTo(v float64, d int) float64 {
	k := math.Pow(10, float64(d))
	return math.Round(v*k) / k
}

// sliderFraction is where v sits between lo and hi, from 0 to 1.
func sliderFraction(v, lo, hi float64) float32 {
	if hi <= lo {
		return 0
	}
	return float32(math.Max(0, math.Min(1, (v-lo)/(hi-lo))))
}

// nearestHandle is which of the values a press at v moves: the nearest,
// the later of two equal ones when v is past them.
func nearestHandle(values []float64, v float64) int {
	best := 0
	for i := 1; i < len(values); i++ {
		if math.Abs(values[i]-v) < math.Abs(values[best]-v) || values[i] == values[best] && v > values[i] {
			best = i
		}
	}
	return best
}

// sliderStep moves a handle at v by n steps (or to the next mark when
// MarksOnly).
func sliderStep(v float64, n int, p SliderProps) float64 {
	lo, hi, step := p.bounds()
	if p.MarksOnly && len(p.Marks) > 0 {
		best := v
		for _, m := range p.Marks {
			if n > 0 && m.Value > v && (best == v || m.Value < best) || n < 0 && m.Value < v && (best == v || m.Value > best) {
				best = m.Value
			}
		}
		return best
	}
	return snapSlider(v+float64(n)*step, lo, hi, step, nil, false)
}

// Slider is Ant's Slider: a rail with a handle, dragged, clicked or moved
// with the arrow keys or the mouse wheel while a handle is focused,
// editing *value. It reports a change.
func Slider(c *ui.Context, value *float64, props ...SliderProps) bool {
	vals := []float64{*value}
	if !sliderCore(c, vals, first(props)) {
		return false
	}
	*value = vals[0]
	return true
}

// RangeSlider is Ant's Slider range: two handles editing the ends of
// *value. It reports a change.
func RangeSlider(c *ui.Context, value *[2]float64, props ...SliderProps) bool {
	vals := []float64{value[0], value[1]}
	if !sliderCore(c, vals, first(props)) {
		return false
	}
	if vals[0] > vals[1] {
		vals[0], vals[1] = vals[1], vals[0]
	}
	*value = [2]float64{vals[0], vals[1]}
	return true
}

const (
	sliderThick  = 4
	sliderHandle = 14
	sliderPad    = 5
)

func sliderCore(c *ui.Context, vals []float64, p SliderProps) bool {
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	lo, hi, step := p.bounds()
	changed := false
	vert := p.Vertical
	length := p.Length
	if vert && length == 0 {
		length = 200
	}
	cross := float32(sliderHandle + 2*sliderPad)
	wrap := ui.Column(c).AlignSelf(ui.Start)
	if vert {
		wrap = ui.Row(c).AlignSelf(ui.Start)
	}
	if !vert && length == 0 {
		wrap.FillWidth().AlignSelf(ui.Stretch)
	}
	wrap.Children(func() {
		rail := ui.Box(c).Shrink(0).Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer)).Disabled(p.Disabled)
		if vert {
			rail.Size(cross, length)
		} else if length > 0 {
			rail.Size(length, cross)
		} else {
			rail.FillWidth().Height(cross)
		}
		b := rail.Bounds()
		span := pick(vert, b.H, b.W) - sliderHandle
		drag := ui.Local(rail, "drag", func() int { return -1 })
		at := ui.Local(rail, "at", func() float32 { return 0 })
		wheel := ui.Local(rail, "wheel", func() int { return 0 })
		kb := ui.Local(rail, "kb", func() int { return -1 })
		valueAt := func(pos float32) float64 {
			if span <= 0 {
				return vals[0]
			}
			f := float64(max(0, min(1, (pos-sliderHandle/2)/span)))
			if vert {
				f = 1 - f
			}
			return snapSlider(lo+f*(hi-lo), lo, hi, step, p.Marks, p.MarksOnly)
		}
		set := func(i int, v float64) {
			if vals[i] != v {
				vals[i] = v
				changed = true
			}
		}
		if *wheel != 0 && !p.Disabled && *kb >= 0 && *kb < len(vals) {
			set(*kb, sliderStep(vals[*kb], *wheel, p))
		}
		*wheel = 0
		*kb = -1
		if rail.Pressed() && !p.Disabled {
			if *drag < 0 {
				x, y, _ := rail.PointerPosition()
				*at = pick(vert, y, x)
				*drag = nearestHandle(vals, valueAt(*at))
			} else if dx, dy, ok := rail.Dragged(); ok {
				*at += pick(vert, dy, dx)
			}
			set(*drag, valueAt(*at))
		} else {
			*drag = -1
		}
		lit := rail.Hovered() || *drag >= 0
		railTone := pick(lit, t.Fill, t.FillTertiary)
		trackTone := pick(p.Disabled, t.TextDisabled, pick(lit, t.PrimaryHover, t.PrimaryBorder))
		if p.Disabled {
			railTone = t.FillTertiary
		}
		marks, fr := p.Marks, func(v float64) float32 { return sliderFraction(v, lo, hi) }
		from, to := fr(lo), fr(vals[0])
		if len(vals) == 2 {
			from, to = fr(min(vals[0], vals[1])), fr(max(vals[0], vals[1]))
		}
		dot := t.Surface
		rail.Draw(func(pa *ui.Painter, r ui.Rect) {
			line := func(a, z float32) ui.Rect {
				s := span
				if vert {
					return ui.Rect{X: r.X + r.W/2 - sliderThick/2, Y: r.Y + sliderHandle/2 + (1-z)*s, W: sliderThick, H: (z - a) * s}
				}
				return ui.Rect{X: r.X + sliderHandle/2 + a*s, Y: r.Y + r.H/2 - sliderThick/2, W: (z - a) * s, H: sliderThick}
			}
			pa.Fill(line(0, 1), railTone, 2)
			if !p.Excluded {
				pa.Fill(line(from, to), trackTone, 2)
			}
			if p.Dots {
				for _, m := range marks {
					f := fr(m.Value)
					cx, cy := r.X+sliderHandle/2+f*span, r.Y+r.H/2
					if vert {
						cx, cy = r.X+r.W/2, r.Y+sliderHandle/2+(1-f)*span
					}
					in := f >= from && f <= to && !p.Excluded
					var d ui.Path
					d.Circle(cx, cy, 4)
					pa.FillPath(&d, dot)
					pa.StrokePath(&d, 2, pick(in, trackTone, railTone))
				}
			}
		})
		rail.Children(func() {
			for i, v := range vals {
				f := fr(v)
				h := ui.Box(c.Key(i)).Absolute().Size(sliderHandle, sliderHandle).Radius(sliderHandle/2).
					Focusable().FocusRing(false).Label(pick(len(vals) == 2, pick(i == 0, "起始值", "结束值"), "滑块")).
					Range(lo, hi, v)
				if vert {
					h.Left(sliderPad).Top((1 - f) * span)
				} else {
					h.Left(f * span).Top(sliderPad)
				}
				hl := h.Hovered() || h.FocusVisible() || *drag == i
				edge := pick(p.Disabled, t.TextDisabled, pick(hl, t.Primary, t.PrimaryBorder))
				h.Background(t.Surface).Border(2, edge)
				if h.Focused() {
					*kb = i
				}
				if hl && !p.Disabled {
					h.Shadow(0, 0, 0, 4, t.PrimaryBg)
				}
				if !p.Disabled {
					switch {
					case h.Shortcut(0, ui.KeyRight) || h.Shortcut(0, ui.KeyUp):
						set(i, sliderStep(v, 1, p))
					case h.Shortcut(0, ui.KeyLeft) || h.Shortcut(0, ui.KeyDown):
						set(i, sliderStep(v, -1, p))
					case h.Shortcut(0, ui.KeyHome):
						set(i, lo)
					case h.Shortcut(0, ui.KeyEnd):
						set(i, hi)
					}
				}
				show := p.Tooltip == TooltipAlways || p.Tooltip == TooltipAuto && hl
				if show && p.Tooltip != TooltipNever {
					label := sliderLabel(p, vals[i])
					ui.Overlay(c.Key(i), func() {
						tip := ui.Box(c.Key(i)).PassThrough()
						place := PlaceTop
						if vert {
							place = PlaceRight
						}
						placePanel(c.Key(i), tip, h, panelPlace{Side: place, Arrow: true, Gap: 4}, t.TooltipBg, func(bubble ui.Element) {
							bubble.Background(t.TooltipBg).Radius(6).Padding(6, 8).Children(func() {
								ui.Text(c, label).FontSize(t.FontSize).TextColor(t.TooltipText).SingleLine()
							})
						})
					})
				}
			}
		})
		if *kb >= 0 && !p.Disabled {
			rail.HandleInput(func(ev ui.InputEvent) bool {
				if ev.Kind != ui.InputScroll || ev.DY == 0 {
					return false
				}
				*wheel = wheelStep(ev.DY)
				return *wheel != 0
			})
		}
		if len(marks) > 0 {
			sliderMarks(c, t, marks, vert, span, length, fr, from, to, p)
		}
	})
	return changed
}

// sliderLabel is the tooltip text of a slider at v.
func sliderLabel(p SliderProps, v float64) string {
	if p.Format != nil {
		return p.Format(v)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// sliderMarks lays the mark labels along a slider's rail.
func sliderMarks(c *ui.Context, t *Tokens, marks []SliderMark, vert bool, span, length float32, fr func(float64) float32, from, to float32, p SliderProps) {
	box := ui.Box(c).Shrink(0)
	if vert {
		box.Size(60, length)
	} else {
		box.FillWidth().Height(22)
	}
	box.Children(func() {
		for i, m := range marks {
			f := fr(m.Value)
			in := f >= from && f <= to && !p.Excluded
			w, _ := c.MeasureText(0, ui.Span{Text: m.Label, Size: t.FontSize})
			l := ui.Text(c.Key(i), m.Label).Absolute().FontSize(t.FontSize).SingleLine().
				TextColor(pick(in, t.Text, t.TextTertiary))
			if vert {
				l.Left(4).Top(sliderHandle/2 + (1-f)*span - 11)
			} else {
				l.Top(2).Left(sliderHandle/2 + f*span - w/2)
			}
		}
	})
}
