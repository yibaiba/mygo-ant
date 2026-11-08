package kit

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
)

// ResultProps are a result page's options.
type ResultProps struct {
	// Status picks the mark: Info, Success, Warning or Error.
	Status   AlertType
	SubTitle string
	// Extra builds the buttons under the subtitle, centered.
	Extra func()
	// Body builds a grey block under the buttons, for details.
	Body func()
}

// Result is Ant's Result: a large status mark, a title and a subtitle
// centered over the buttons Extra builds, for the end of a long task such
// as an install or an upgrade.
func Result(c *ui.Context, title string, props ...ResultProps) {
	p := first(props)
	t := Of(c)
	_, _, tone, _ := alertColors(t, p.Status)
	ui.Column(c).FillWidth().Padding(48, 32).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Margin(0, 0, 24).Children(func() { statusIcon(c, p.Status, 72, tone) })
		ui.Text(c, title).FontSize(24).FontWeight(500).TextColor(t.TextHeading).TextAlign(ui.Center)
		if p.SubTitle != "" {
			ui.Text(c, p.SubTitle).FontSize(t.FontSize).TextColor(t.TextTertiary).TextAlign(ui.Center).Margin(t.SpaceSM, 0, 0)
		}
		if p.Extra != nil {
			ui.Row(c).Gap(t.SpaceSM).Wrap().GapY(t.SpaceSM).Justify(ui.Center).Margin(24, 0, 0).Children(p.Extra)
		}
		if p.Body != nil {
			ui.Column(c).FillWidth().Gap(t.SpaceSM).Margin(24, 0, 0).Padding(24, 40).
				Background(t.FillAlter).Radius(t.RadiusSheet).Children(p.Body)
		}
	})
}

// SpinProps are a spin indicator's options.
type SpinProps struct {
	// Size is the indicator's: 14, 20 or 32 across for Small, Middle and
	// Large.
	Size Size
	// Tip is a line under the indicator, as "加载中".
	Tip string
}

func spinSize(s Size) float32 {
	return map[Size]float32{Small: 14, Middle: 20, Large: 32}[s]
}

// Spin is Ant's Spin indicator on its own: four dots turning in the
// primary color, with Tip under them.
func Spin(c *ui.Context, props ...SpinProps) {
	p := first(props)
	t := Of(c)
	p.Size = t.SizeOf(p.Size)
	ui.Column(c).Gap(t.SpaceSM).AlignItems(ui.Center).Label(pick(p.Tip == "", L(c).Loading, p.Tip)).
		Role(ui.RoleProgress).Children(func() {
		spinDots(c, spinSize(p.Size), t.Primary)
		if p.Tip != "" {
			ui.Text(c, p.Tip).FontSize(t.FontSize).TextColor(t.Primary)
		}
	})
}

// SpinOver is Ant's Spin around content: while spinning, what fn builds
// dims and takes no input, and the indicator with Tip sits over its middle.
func SpinOver(c *ui.Context, spinning bool, fn func(), props ...SpinProps) {
	ui.Column(c).FillWidth().Children(func() {
		body := ui.Column(c).FillWidth()
		if spinning {
			body.Opacity(0.5).Disabled(true)
		}
		body.Children(fn)
		if spinning {
			ui.Column(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Center().Children(func() {
				Spin(c, props...)
			})
		}
	})
}

// spinDots draws Ant's dot indicator: four dots on the corners of a
// square turning once every 1.2 s, each fading between 0.3 and 1.
func spinDots(c *ui.Context, size float32, tone ui.Color) {
	still := stillMotion(c)
	ui.Box(c).Size(size, size).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		ms := float64(p.Now().UnixMilli())
		turn := math.Pi / 4
		if !still {
			p.AnimationFrameRate(30)
			turn = math.Mod(ms, 1200)/1200*2*math.Pi + math.Pi/4
		}
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		dot := (r.W - 2) / 2
		reach := float64(r.W/2-dot/2) * math.Sqrt2
		for i := range 4 {
			a := turn + float64(i)*math.Pi/2
			x := cx + float32(reach*math.Cos(a))
			y := cy + float32(reach*math.Sin(a))
			alpha := float32(1)
			if !still {
				phase := math.Mod(ms/1000+0.4*float64(i), 2)
				alpha = 0.3 + 0.7*float32(min(phase, 2-phase))
			}
			p.Fill(ui.Rect{X: x - dot/2, Y: y - dot/2, W: dot, H: dot}, tone.Alpha(alpha), dot/2)
		}
	})
}

// SkeletonProps are a skeleton's options.
type SkeletonProps struct {
	// Rows is the number of paragraph lines; 0 means 3.
	Rows int
	// Avatar adds a round placeholder at the start.
	Avatar bool
	// NoTitle leaves out the short title bar.
	NoTitle bool
	// Active runs a light band across the bars, as while loading.
	Active bool
}

// Skeleton is Ant's Skeleton: grey bars standing in for a title and a
// paragraph, the last line shorter, until the content arrives.
func Skeleton(c *ui.Context, props ...SkeletonProps) {
	p := first(props)
	t := Of(c)
	rows := pick(p.Rows == 0, 3, p.Rows)
	block := SkeletonBlockProps{Active: p.Active}
	ui.Row(c).FillWidth().Gap(t.SpaceLG).AlignItems(ui.Start).Children(func() {
		if p.Avatar {
			SkeletonBlock(c, 32, 32, SkeletonBlockProps{Active: p.Active, Round: true})
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(t.SpaceLG).Children(func() {
			if !p.NoTitle {
				ui.Box(c).WidthPercent(38).Margin(0, 0, 8).Children(func() { SkeletonBlock(c, 0, 16, block) })
			}
			for i := range rows {
				line := ui.Box(c).FillWidth()
				if i == rows-1 && rows > 1 {
					line.WidthPercent(61)
				}
				line.Children(func() { SkeletonBlock(c, 0, 16, block) })
			}
		})
	})
}

// SkeletonBlockProps are a skeleton block's options.
type SkeletonBlockProps struct {
	Active bool
	// Round makes the block a circle, or a pill when wider than tall.
	Round bool
	// Radius rounds the corners; 4 when zero.
	Radius float32
}

// SkeletonBlock is one grey placeholder w×h, filling its parent's width
// when w is zero, for building skeletons shaped like a card.
func SkeletonBlock(c *ui.Context, w, h float32, props ...SkeletonBlockProps) {
	p := first(props)
	t := Of(c)
	radius := pick(p.Radius == 0, 4, p.Radius)
	if p.Round {
		radius = h / 2
	}
	e := ui.Box(c).Height(h).Shrink(0).Radius(radius).Background(t.FillSecondary)
	if w > 0 {
		e.Width(w)
	} else {
		e.FillWidth()
	}
	if !p.Active || stillMotion(c) {
		return
	}
	base, hi := t.FillSecondary, t.Fill
	e.DrawOver(func(pa *ui.Painter, r ui.Rect) {
		pa.AnimationFrameRate(30)
		const period = 1400 * time.Millisecond
		at := -0.3 + 1.6*float32(pa.Now().UnixNano()%int64(period))/float32(period)
		clamp := func(v float32) float32 { return max(0, min(1, v)) }
		pa.FillGradient(r, ui.LinearGradient{Angle: 90, Stops: []ui.GradientStop{
			{Color: base, At: clamp(at - 0.3)},
			{Color: hi, At: clamp(at)},
			{Color: base, At: clamp(at + 0.3)},
		}}, radius)
	})
}
