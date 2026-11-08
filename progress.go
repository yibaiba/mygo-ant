package kit

import (
	"strconv"

	"github.com/egoist/mygo/ui"
)

// ProgressType is the shape of a progress indicator.
type ProgressType uint8

// Progress types.
const (
	ProgressLine ProgressType = iota
	ProgressCircle
	// ProgressDashboard is a circle open at the bottom by GapDegree.
	ProgressDashboard
)

// ProgressStatus is Ant's progress status.
type ProgressStatus uint8

// Progress statuses: Normal turns Success at 100.
const (
	ProgressNormal ProgressStatus = iota
	ProgressActive
	ProgressSuccess
	ProgressException
)

// ProgressCap is how a progress stroke ends (strokeLinecap).
type ProgressCap uint8

// Progress caps.
const (
	CapRound ProgressCap = iota
	CapButt
	CapSquare
)

// ProgressProps are a progress indicator's options.
type ProgressProps struct {
	Type   ProgressType
	Status ProgressStatus
	// Error is Status ProgressException, as the install view sets it.
	Error    bool
	HideInfo bool
	// Format writes the info text; "n%" by default.
	Format func(percent int) string
	// SuccessPercent paints the first part green (success.percent).
	SuccessPercent int
	// Steps splits a line into that many blocks, or a circle into arcs.
	Steps int
	Cap   ProgressCap
	// Color is the stroke; From and To, when set, a gradient along it.
	Color, From, To ui.Color
	// Trail is the unfilled stroke; FillSecondary by default.
	Trail ui.Color
	// Small is Ant's size="small"; Width is a circle's diameter (120,
	// small 80) and StrokeWidth its stroke in percent of it (6), or a
	// line's height (8, small 6).
	Small       bool
	Width       float32
	StrokeWidth float32
	// GapDegree is a dashboard's opening, 75 by default.
	GapDegree float32
}

// progressStatus is the status the indicator shows.
func progressStatus(percent int, p ProgressProps) ProgressStatus {
	switch {
	case p.Error:
		return ProgressException
	case p.Status == ProgressNormal && percent >= 100:
		return ProgressSuccess
	}
	return p.Status
}

// progressSegments are the arcs of a circle of n steps through sweep
// degrees from start, gap degrees apart: start and sweep of each.
func progressSegments(start, sweep float32, n int, gap float32) [][2]float32 {
	if n <= 1 {
		return [][2]float32{{start, sweep}}
	}
	each := (sweep - gap*float32(n)) / float32(n)
	out := make([][2]float32, n)
	for i := range out {
		out[i] = [2]float32{start + float32(i)*(each+gap) + gap/2, each}
	}
	return out
}

// Progress is Ant's Progress: a line (in the default brand stroke, as
// InstallProgressView's strokeColor #1E59D7), circle or dashboard, with
// steps, success segment, gradient and status marks.
func Progress(c *ui.Context, percent int, props ...ProgressProps) {
	p := first(props)
	t := Of(c)
	percent = max(0, min(percent, 100))
	status := progressStatus(percent, p)
	stroke := pick(p.Color == (ui.Color{}), t.Feedback.Progress, p.Color)
	switch status {
	case ProgressException:
		stroke = t.Error
	case ProgressSuccess:
		if p.Color == (ui.Color{}) && p.Type != ProgressLine {
			stroke = t.Feedback.ProgressSuccess
		}
	}
	trail := pick(p.Trail == (ui.Color{}), t.FillSecondary, p.Trail)
	if p.Type == ProgressLine {
		progressLine(c, t, percent, status, stroke, trail, p)
		return
	}
	progressCircle(c, t, percent, status, stroke, trail, p)
}

func progressText(percent int, p ProgressProps) string {
	if p.Format != nil {
		return p.Format(percent)
	}
	return strconv.Itoa(percent) + "%"
}

func capRadius(p ProgressProps, h float32) float32 {
	return pick(p.Cap == CapRound, h/2, 0)
}

func progressLine(c *ui.Context, t *Tokens, percent int, status ProgressStatus, stroke, trail ui.Color, p ProgressProps) {
	h := pick(p.StrokeWidth > 0, p.StrokeWidth, pick[float32](p.Small, 6, 8))
	gradient := p.From != (ui.Color{}) && status != ProgressException
	ui.Row(c).FillWidth().Gap(t.SpaceSM).AlignItems(ui.Center).Children(func() {
		if p.Steps > 0 {
			w := pick[float32](p.Small, 2, 14)
			ui.Row(c).Gap(2).Shrink(0).Children(func() {
				done := int(float32(p.Steps)*float32(percent)/100 + 0.5)
				for i := 0; i < p.Steps; i++ {
					ui.Box(c.Key(i)).Size(w, h).Background(pick(i < done, stroke, trail))
				}
			})
		} else {
			ui.Box(c).Grow(1).Height(h).Radius(capRadius(p, h)).Background(trail).Draw(func(pa *ui.Painter, r ui.Rect) {
				w := r.W * float32(percent) / 100
				if w > 0 {
					bar := ui.Rect{X: r.X, Y: r.Y, W: w, H: r.H}
					if gradient {
						pa.FillGradient(bar, ui.LinearGradient{From: p.From, To: pick(p.To == (ui.Color{}), p.From, p.To), Angle: 90}, capRadius(p, h))
					} else {
						pa.Fill(bar, stroke, capRadius(p, h))
					}
					if status == ProgressActive && !stillMotion(c) {
						// Ant's ant-progress-active: a white sheen sweeping the
						// done part every 2.4s. ReduceMotion hides the band.
						phase := float32(pa.Now().UnixMilli()%2400) / 2400
						pa.Fill(ui.Rect{X: r.X, Y: r.Y, W: w * phase, H: r.H}, ui.RGBA(255, 255, 255, 0.5*(1-phase)), capRadius(p, h))
						pa.AnimationFrame()
					}
				}
				if s := max(0, min(p.SuccessPercent, percent)); s > 0 {
					pa.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W * float32(s) / 100, H: r.H}, t.Feedback.ProgressSuccess, capRadius(p, h))
				}
			})
		}
		if p.HideInfo {
			return
		}
		info := ui.Row(c).MinWidth(pick[float32](p.Small, 32, 40)).Shrink(0).AlignItems(ui.Center)
		info.Children(func() {
			size := pick(p.Small, t.FontSizeSM, t.FontSize)
			switch {
			case status == ProgressSuccess && p.Format == nil:
				statusIcon(c, Success, size, t.Feedback.ProgressSuccess)
			case status == ProgressException && p.Format == nil:
				statusIcon(c, Error, size, t.Error)
			default:
				ui.Text(c, progressText(percent, p)).FontSize(size).TextColor(t.Text).SingleLine()
			}
		})
	})
}

func progressCircle(c *ui.Context, t *Tokens, percent int, status ProgressStatus, stroke, trail ui.Color, p ProgressProps) {
	size := pick(p.Width > 0, p.Width, pick[float32](p.Small, 80, 120))
	sw := pick[float32](p.StrokeWidth > 0, p.StrokeWidth, 6) * size / 100
	start, sweep := float32(0), float32(360)
	if p.Type == ProgressDashboard {
		gap := pick[float32](p.GapDegree > 0, p.GapDegree, 75)
		start, sweep = 180+gap/2, 360-gap
	}
	lineCap := map[ProgressCap]ui.LineCap{CapRound: ui.RoundCap, CapButt: ui.ButtCap, CapSquare: ui.SquareCap}[p.Cap]
	if p.Steps > 1 {
		lineCap = ui.ButtCap
	}
	gradient := p.From != (ui.Color{}) && status != ProgressException
	ui.Column(c).Size(size, size).Shrink(0).Center().Draw(func(pa *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		rad := (size - sw) / 2
		segs := progressSegments(start, sweep, p.Steps, 2)
		done := sweep * float32(percent) / 100
		ok := sweep * float32(max(0, min(p.SuccessPercent, percent))) / 100
		for _, sg := range segs {
			pa.StrokePath(arcPath(cx, cy, rad, sg[0], sg[1]).Cap(lineCap), sw, trail)
		}
		paint := func(upTo float32, color ui.Color, grad bool) {
			if upTo <= 0 {
				return
			}
			for _, sg := range segs {
				from := sg[0] - start
				part := min(sg[1], upTo-from)
				if part <= 0 {
					continue
				}
				path := arcPath(cx, cy, rad, sg[0], part).Cap(lineCap)
				if grad {
					pa.StrokePathConicGradient(path, sw, ui.ConicGradient{Angle: start, Stops: []ui.GradientStop{
						{Color: p.From, At: 0}, {Color: pick(p.To == (ui.Color{}), p.From, p.To), At: sweep / 360},
					}})
				} else {
					pa.StrokePath(path, sw, color)
				}
			}
		}
		paint(done, stroke, gradient)
		paint(ok, t.Feedback.ProgressSuccess, false)
	}).Children(func() {
		if p.HideInfo {
			return
		}
		fs := size*0.15 + 6
		switch {
		case status == ProgressSuccess && p.Format == nil:
			ui.Icon(c, LookupIcon("check")).Size(fs, fs).TextColor(t.Feedback.ProgressSuccess)
		case status == ProgressException && p.Format == nil:
			ui.Icon(c, LookupIcon("close")).Size(fs, fs).TextColor(t.Error)
		default:
			ui.Text(c, progressText(percent, p)).FontSize(fs).TextColor(t.TextHeading).SingleLine()
		}
	})
}
