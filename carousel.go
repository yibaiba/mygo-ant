package kit

import (
	"math"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// CarouselProps are a carousel's options.
type CarouselProps struct {
	// Autoplay moves to the next slide every Interval (3s), pausing while
	// the pointer rests on the carousel.
	Autoplay bool
	Interval time.Duration
	// Arrows adds previous and next arrows on the sides.
	Arrows bool
	// NoDots hides the dots; DotPosition is "bottom" (the default), "top",
	// "left" or "right", the last two sliding vertically.
	NoDots      bool
	DotPosition string
	// Fade cross-fades the slides instead of sliding them.
	Fade bool
	// Current, when set, holds the slide shown.
	Current *int
	// Height is the carousel's; 160 when zero.
	Height float32
	// Speed is how long a change takes; 500ms when zero.
	Speed time.Duration
}

// carouselHold is what a carousel keeps between frames: its step counts
// changes without wrapping, so the last slide moves on to the first.
type carouselHold struct {
	step  int
	since time.Time
	drag  float32
}

// carouselRel is where slide i of n sits, in slide widths from the view,
// with the carousel at pos: in [-1, n-1), the slides ahead to the right.
func carouselRel(i, n int, pos float32) float32 {
	rel := float32(i) - pos
	fn := float32(n)
	rel -= fn * float32(math.Floor(float64((rel+1)/fn)))
	return rel
}

func wrapIndex(i, n int) int {
	if n == 0 {
		return 0
	}
	return ((i % n) + n) % n
}

// Carousel is Ant's Carousel: n slides built by slide, one at a time,
// sliding (or fading) between them with dots, arrows, swipes and
// autoplay. It reports a change.
func Carousel(c *ui.Context, n int, slide func(i int), props ...CarouselProps) bool {
	p := first(props)
	if n == 0 {
		return false
	}
	h := pick(p.Height == 0, 160, p.Height)
	speed := pick(p.Speed == 0, 500*time.Millisecond, p.Speed)
	interval := pick(p.Interval == 0, 3*time.Second, p.Interval)
	vertical := p.DotPosition == "left" || p.DotPosition == "right"
	box := ui.Column(c).FillWidth().Height(h).Clip().Label("轮播")
	hold := ui.Local(box, "hold", func() carouselHold { return carouselHold{since: c.Now()} })
	if p.Current != nil && wrapIndex(hold.step, n) != wrapIndex(*p.Current, n) {
		hold.step += wrapIndex(*p.Current, n) - wrapIndex(hold.step, n)
	}
	changed := false
	goTo := func(step int) {
		if step != hold.step {
			hold.step = step
			hold.since = c.Now()
			changed = true
			if p.Current != nil {
				*p.Current = wrapIndex(step, n)
			}
		}
	}
	if p.Autoplay && n > 1 {
		if box.Hovered() {
			hold.since = c.Now()
		} else if left := interval - c.Now().Sub(hold.since); left <= 0 {
			goTo(hold.step + 1)
		} else {
			c.After(left)
		}
	}
	size := box.Bounds().W
	if vertical {
		size = box.Bounds().H
	}
	if dx, dy, ok := box.Dragged(); ok && n > 1 && !p.Fade {
		hold.drag += pick(vertical, dy, dx)
	}
	if _, _, ended := box.DragEnded(); ended {
		switch {
		case hold.drag < -40:
			goTo(hold.step + 1)
		case hold.drag > 40:
			goTo(hold.step - 1)
		}
		hold.drag = 0
	}
	pos := box.AnimateTo("pos", float32(hold.step), ui.Animation{Duration: speed, Ease: ui.EaseInOut})
	if size > 0 {
		pos -= hold.drag / size
	}
	cur := wrapIndex(hold.step, n)
	box.Children(func() {
		for i := 0; i < n; i++ {
			s := ui.Column(c.Key(i)).Absolute().Top(0).Left(0).WidthPercent(100).HeightPercent(100)
			if p.Fade {
				op := s.AnimateTo("fade", pick[float32](i == cur, 1, 0), ui.Animation{Duration: speed, Ease: ui.EaseInOut})
				s.Opacity(op)
				if i != cur {
					s.PassThrough()
				}
			} else {
				rel := carouselRel(i, n, pos)
				if vertical {
					s.TopPercent(100 * rel)
				} else {
					s.LeftPercent(100 * rel)
				}
				if rel <= -1 || rel >= 1 {
					s.Invisible()
				}
			}
			s.Children(func() { slide(i) })
		}
		if p.Arrows && n > 1 {
			for _, dir := range []int{-1, 1} {
				a := ui.ButtonBase(c.Key("arrow"+strconv.Itoa(dir))).Label(pick(dir < 0, "上一张", "下一张")).Size(24, 24).
					Absolute().FocusRing(false).Cursor(ui.CursorPointer)
				if vertical {
					a.LeftPercent(50).Translate(-12, 0)
					if dir < 0 {
						a.Top(4)
					} else {
						a.Bottom(4)
					}
				} else {
					a.TopPercent(50).Translate(0, -12)
					if dir < 0 {
						a.Left(8)
					} else {
						a.Right(8)
					}
				}
				alpha := pick[float32](a.Hovered(), 1, 0.6)
				a.Draw(func(pa *ui.Painter, r ui.Rect) {
					var path ui.Path
					cx, cy := r.X+r.W/2, r.Y+r.H/2
					switch {
					case vertical && dir < 0:
						path.MoveTo(cx-6, cy+3).LineTo(cx, cy-3).LineTo(cx+6, cy+3)
					case vertical:
						path.MoveTo(cx-6, cy-3).LineTo(cx, cy+3).LineTo(cx+6, cy-3)
					case dir < 0:
						path.MoveTo(cx+3, cy-6).LineTo(cx-3, cy).LineTo(cx+3, cy+6)
					default:
						path.MoveTo(cx-3, cy-6).LineTo(cx+3, cy).LineTo(cx-3, cy+6)
					}
					pa.StrokePath(&path, 2, ui.RGBA(255, 255, 255, alpha))
				})
				if a.Clicked() {
					goTo(hold.step + dir)
				}
			}
		}
		if !p.NoDots && n > 1 {
			dots := pick(vertical, ui.Column(c), ui.Row(c)).Absolute().Gap(8).AlignItems(ui.Center)
			switch p.DotPosition {
			case "top":
				dots.Top(12).Left(0).Right(0).Justify(ui.Center)
			case "left":
				dots.Left(12).Top(0).Bottom(0).Justify(ui.Center)
			case "right":
				dots.Right(12).Top(0).Bottom(0).Justify(ui.Center)
			default:
				dots.Bottom(12).Left(0).Right(0).Justify(ui.Center)
			}
			dots.Children(func() {
				for i := 0; i < n; i++ {
					d := ui.ButtonBase(c.Key(i)).Label("第 " + strconv.Itoa(i+1) + " 张").FocusRing(false).Cursor(ui.CursorPointer)
					on := i == cur
					long := d.AnimateTo("w", pick[float32](on, 24, 16), ui.Animation{Duration: 300 * time.Millisecond, Ease: ui.EaseOut})
					if vertical {
						d.Size(3, long)
					} else {
						d.Size(long, 3)
					}
					alpha := pick[float32](on, 1, pick[float32](d.Hovered(), 0.75, 0.3))
					d.Radius(1).Background(ui.RGBA(255, 255, 255, alpha))
					if d.Clicked() && !on {
						goTo(hold.step + i - cur)
					}
				}
			})
		}
	})
	return changed
}
