package kit

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

func svgIcon(body string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round">` + body + `</svg>`))
}

var (
	iconZoomIn      = svgIcon(`<circle cx="7" cy="7" r="4.6"/><path d="m10.5 10.5 3.5 3.5M5 7h4M7 5v4"/>`)
	iconZoomOut     = svgIcon(`<circle cx="7" cy="7" r="4.6"/><path d="m10.5 10.5 3.5 3.5M5 7h4"/>`)
	iconRotateLeft  = svgIcon(`<path d="M3.5 7.5A5 5 0 1 1 5 11.6"/><path d="M1.8 5.6 3.5 7.6l2-1.6"/>`)
	iconRotateRight = svgIcon(`<path d="M12.5 7.5A5 5 0 1 0 11 11.6"/><path d="m14.2 5.6-1.7 2-2-1.6"/>`)
	iconFlipX       = svgIcon(`<path d="M8 2v12"/><path d="M6 4.5 2 11.5h4z"/><path d="m10 4.5 4 7h-4z" fill="currentColor"/>`)
	iconFlipY       = svgIcon(`<path d="M2 8h12"/><path d="M4.5 6 11.5 2v4z"/><path d="m4.5 10 7 4v-4z" fill="currentColor"/>`)
	iconLeft        = svgIcon(`<path d="M10 3 5 8l5 5"/>`)
	iconRight       = svgIcon(`<path d="m6 3 5 5-5 5"/>`)
)

// ImageProps are an image's options.
type ImageProps struct {
	// Width and Height size the image; it shows at its own size when both
	// are zero, scaled to Width (or Height) when one is set.
	Width, Height float32
	Fit           ui.Fit
	Radius        float32
	// NoPreview turns the click-to-preview off.
	NoPreview bool
	// Fallback shows when src is nil, as a broken image.
	Fallback ui.ImageSource
	// Alt labels the image for assistive technology.
	Alt string
	// Group previews these images instead of src alone, cycling from
	// Index.
	Group []ui.ImageSource
	Index int
}

// imageBox sizes an image shown in w×h (either may be zero) by its own
// size iw×ih.
func imageBox(w, h, iw, ih float32) (float32, float32) {
	switch {
	case w > 0 && h > 0:
		return w, h
	case w > 0 && iw > 0:
		return w, w * ih / iw
	case h > 0 && ih > 0:
		return h * iw / ih, h
	}
	return iw, ih
}

// Image is Ant's Image: a picture that darkens under the pointer with
// "预览" and opens the preview on a click.
func Image(c *ui.Context, src ui.ImageSource, props ...ImageProps) {
	p := first(props)
	t := Of(c)
	if src == nil {
		src = p.Fallback
	}
	iw, ih := ui.ImageSize(src)
	w, h := imageBox(p.Width, p.Height, iw, ih)
	box := ui.Column(c).Size(w, h).Shrink(0).Radius(p.Radius).Clip().Label(pick(p.Alt == "", "图片", p.Alt))
	if src == nil {
		box.Background(t.FillTertiary)
	}
	open := ui.Local(box, "preview", func() bool { return false })
	cur := ui.Local(box, "index", func() int { return p.Index })
	previewable := !p.NoPreview && src != nil
	if previewable {
		box.Cursor(ui.CursorPointer)
		if box.Clicked() {
			*open = true
			*cur = p.Index
		}
	}
	box.Children(func() {
		if src != nil {
			ui.Image(c, src).Size(w, h).Fit(p.Fit)
		}
		if previewable {
			mask := ui.Column(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Center().PassThrough()
			fade := mask.AnimateTo("hover", pick[float32](box.Hovered(), 1, 0), ui.Animation{Duration: 200 * time.Millisecond})
			mask.Background(ui.RGBA(0, 0, 0, 0.5*fade)).Children(func() {
				if fade > 0.01 {
					ui.Row(c).Gap(4).AlignItems(ui.Center).Opacity(fade).Children(func() {
						ui.Icon(c, iconEye).Size(14, 14).TextColor(ui.Hex("#FFFFFF"))
						ui.Text(c, L(c).Preview).FontSize(t.FontSize).TextColor(ui.Hex("#FFFFFF")).SingleLine()
					})
				}
			})
		}
	})
	if previewable {
		group := p.Group
		if len(group) == 0 {
			group = []ui.ImageSource{src}
		}
		ImagePreview(c, open, group, cur)
	}
}

// ImageGroup is Ant's Image.PreviewGroup over thumbnails: each image of
// srcs shown at props' size, the preview stepping through all of them.
func ImageGroup(c *ui.Context, srcs []ui.ImageSource, props ...ImageProps) {
	p := first(props)
	ui.Row(c).Gap(8).Wrap().GapY(8).Children(func() {
		for i, s := range srcs {
			q := p
			q.Group, q.Index = srcs, i
			Image(c.Key(i), s, q)
		}
	})
}

// imageView is how the preview shows its image.
type imageView struct {
	scale, rotate float32
	flipX, flipY  bool
	x, y          float32
}

func (v *imageView) reset() { *v = imageView{scale: 1} }

// zoomAt scales the view by ratio about the point (px, py) from the
// view's center, keeping that point under the pointer, within Ant's 1 to
// 50.
func (v *imageView) zoomAt(ratio, px, py float32) {
	next := min(max(v.scale*ratio, 1), 50)
	if next == v.scale {
		return
	}
	k := next / v.scale
	v.x = px - (px-v.x)*k
	v.y = py - (py-v.y)*k
	v.scale = next
	if next == 1 {
		v.x, v.y = 0, 0
	}
}

// wheelRatio is how far a wheel notch of dy zooms, as Ant's 1.5 step
// for a notch, smaller for a touchpad's pixels.
func wheelRatio(dy float32, precise bool) float32 {
	if precise {
		return 1 - dy/200
	}
	if dy < 0 {
		return 1.5
	}
	return 1 / 1.5
}

// ImagePreview is Ant's image preview over the window while *open: the
// image at *current of srcs, zoomed by the wheel, a pinch or the
// toolbar, panned by dragging, rotated and flipped, with arrows and a
// counter for several. Escape or a click outside the image closes it.
func ImagePreview(c *ui.Context, open *bool, srcs []ui.ImageSource, current *int) {
	host := motionHost(c)
	view := ui.Local(host, "view", func() imageView { return imageView{scale: 1} })
	shown, in := presence(host, "preview", *open, motionFor(c, 200*time.Millisecond))
	if !shown || len(srcs) == 0 {
		return
	}
	*current = wrapIndex(*current, len(srcs))
	t := Of(c)
	white := ui.RGBA(255, 255, 255, 0.85)
	step := func(d int) {
		*current = wrapIndex(*current+d, len(srcs))
		view.reset()
	}
	ui.Overlay(c, func() {
		ww, wh := c.Size()
		layer := ui.Column(c).Absolute().Top(0).Left(0).Size(ww, wh).Background(ui.RGBA(0, 0, 0, 0.45*in)).
			Focusable().AutoFocus().Label("图片预览").Opacity(in)
		layer.HandleInput(func(ev ui.InputEvent) bool {
			switch ev.Kind {
			case ui.InputScroll:
				if ev.DY == 0 {
					return false
				}
				view.zoomAt(wheelRatio(ev.DY, ev.Precise), ev.X-ww/2, ev.Y-wh/2)
				return true
			case ui.InputZoom:
				view.zoomAt(ev.Scale, ev.X-ww/2, ev.Y-wh/2)
				return true
			}
			return false
		})
		if c.Shortcut(0, ui.KeyEscape) {
			*open = false
		}
		if len(srcs) > 1 && c.Shortcut(0, ui.KeyLeft) {
			step(-1)
		}
		if len(srcs) > 1 && c.Shortcut(0, ui.KeyRight) {
			step(1)
		}
		if dx, dy, ok := layer.Dragged(); ok {
			view.x += dx
			view.y += dy
		}
		if layer.DoubleClicked() {
			if view.scale > 1 {
				view.reset()
			} else {
				view.zoomAt(2, 0, 0)
			}
		}
		var img ui.Element
		layer.Children(func() {
			src := srcs[*current]
			iw, ih := ui.ImageSize(src)
			fit := min(1, min((ww-80)/max(iw, 1), (wh-160)/max(ih, 1)))
			w, h := iw*fit, ih*fit
			img = ui.Image(c, src).Size(w, h).Absolute().Left((ww - w) / 2).Top((wh - h) / 2).PassThrough()
			dragging := layer.Dragging()
			anim := ui.Animation{Duration: 300 * time.Millisecond, Ease: ui.EaseOut}
			s := img.AnimateTo("scale", view.scale, anim)
			r := img.AnimateTo("rotate", view.rotate, anim)
			var x, y float32
			if dragging {
				img.Jump("x", view.x)
				img.Jump("y", view.y)
				x, y = view.x, view.y
			} else {
				x, y = img.AnimateTo("x", view.x, anim), img.AnimateTo("y", view.y, anim)
			}
			sx, sy := s, s
			if view.flipX {
				sx = -s
			}
			if view.flipY {
				sy = -s
			}
			img.Transform(ui.Rotation(r).Then(ui.Scaling(sx, sy)).Then(ui.Translation(x, y)))
			round := func(e ui.Element, size float32) ui.Element {
				alpha := pick[float32](e.Hovered(), 0.2, 0.1)
				return e.Size(size, size).Radius(size / 2).Background(ui.RGBA(0, 0, 0, alpha)).Center().
					FocusRing(false).Cursor(ui.CursorPointer)
			}
			if b := round(ui.ButtonBase(c).Label("关闭预览").Absolute().Top(32).Right(32), 40); b.Clicked() {
				*open = false
			} else {
				b.Children(func() { ui.Icon(c, iconClose).Size(18, 18).TextColor(white) })
			}
			if len(srcs) > 1 {
				ui.Text(c, strconv.Itoa(*current+1)+" / "+strconv.Itoa(len(srcs))).Absolute().Top(40).Left(0).Right(0).
					TextAlign(ui.Center).FontSize(t.FontSize).TextColor(white).PassThrough()
				for _, d := range []int{-1, 1} {
					b := round(ui.ButtonBase(c.Key("nav"+strconv.Itoa(d))).Label(pick(d < 0, "上一张", "下一张")).Absolute().Top(wh/2-20), 40)
					if d < 0 {
						b.Left(12)
					} else {
						b.Right(12)
					}
					b.Children(func() { ui.Icon(c, pick(d < 0, iconLeft, iconRight)).Size(18, 18).TextColor(white) })
					if b.Clicked() {
						step(d)
					}
				}
			}
			ui.Row(c).Absolute().Bottom(32).Left(0).Right(0).Justify(ui.Center).PassThrough().Children(func() {
				ui.Row(c).Padding(0, 24).Height(44).Gap(12).Radius(100).Background(ui.RGBA(0, 0, 0, 0.1)).AlignItems(ui.Center).Children(func() {
					tool := func(label string, icon *ui.SVG, off bool, fn func()) {
						b := ui.ButtonBase(c.Key(label)).Label(label).Size(28, 28).Center().FocusRing(false).Disabled(off).
							Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
						tone := white
						switch {
						case off:
							tone = ui.RGBA(255, 255, 255, 0.25)
						case b.Hovered():
							tone = ui.RGBA(255, 255, 255, 1)
						}
						b.Children(func() { ui.Icon(c, icon).Size(18, 18).TextColor(tone) })
						if b.Clicked() && !off {
							fn()
						}
					}
					tool("上下翻转", iconFlipY, false, func() { view.flipY = !view.flipY })
					tool("左右翻转", iconFlipX, false, func() { view.flipX = !view.flipX })
					tool("向左旋转", iconRotateLeft, false, func() { view.rotate -= 90 })
					tool("向右旋转", iconRotateRight, false, func() { view.rotate += 90 })
					tool("缩小", iconZoomOut, view.scale <= 1, func() { view.zoomAt(1/1.5, 0, 0) })
					tool("放大", iconZoomIn, view.scale >= 50, func() { view.zoomAt(1.5, 0, 0) })
				})
			})
		})
		if layer.Clicked() {
			px, py, _ := layer.PointerPosition()
			if !img.Bounds().Contains(px, py) {
				*open = false
			}
		}
	})
	if !*open {
		view.reset()
	}
}
