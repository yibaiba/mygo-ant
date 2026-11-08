package kit

import (
	"math"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"
)

// CollapseItem is one panel of a Collapse.
type CollapseItem struct {
	Key, Label string
	// Extra builds controls at the end of the header, as a tag or a link.
	Extra func()
	// Body builds the panel's content.
	Body     func()
	Disabled bool
}

// CollapseProps are a collapse's options.
type CollapseProps struct {
	// Accordion keeps at most one panel open.
	Accordion bool
	// Ghost drops the border and the grey header background.
	Ghost bool
	Size  Size
}

// Collapse is Ant's Collapse: headers with a turning arrow, each opening
// its panel below it; a click, Enter or Space toggles it. *active holds
// the keys of the open panels. It reports a change.
func Collapse(c *ui.Context, active *[]string, items []CollapseItem, props ...CollapseProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	t := Of(c)
	padY := map[Size]float32{Small: 8, Middle: 12, Large: 16}[p.Size]
	padX := map[Size]float32{Small: 12, Middle: 16, Large: 24}[p.Size]
	changed := false
	box := ui.Column(c).FillWidth()
	if !p.Ghost {
		box.Background(t.FillAlter).Border(1, t.Border).Radius(t.RadiusSheet).Clip()
	}
	box.Children(func() {
		for i, it := range items {
			open := slices.Contains(*active, it.Key)
			last := i == len(items)-1
			item := ui.Column(c.Key(it.Key)).FillWidth()
			if !last && !p.Ghost {
				item.DrawOver(bottomRule(t.Border))
			}
			item.Children(func() {
				head := ui.ButtonBase(c).Label(it.Label).Role(ui.RoleDisclosure).Expanded(open).Disabled(it.Disabled).FocusRing(false).
					FillWidth().Padding(padY, padX).Gap(t.SpaceMD).AlignItems(ui.Center).
					Cursor(pick(it.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
				if head.FocusVisible() {
					head.InsetShadow(0, 0, 0, 2, t.PrimaryBorder)
				}
				turn := head.Animate("open", pick[float32](open, 1, 0), 200*time.Millisecond)
				fg := pick(it.Disabled, t.TextDisabled, t.TextHeading)
				head.Children(func() {
					expandArrow(c, fg, turn)
					ui.Text(c, it.Label).FontSize(t.FontSize).TextColor(fg).Grow(1).MinWidth(0)
					if it.Extra != nil {
						ui.Row(c).Shrink(0).AlignItems(ui.Center).Children(it.Extra)
					}
				})
				if head.Clicked() && !it.Disabled {
					*active = toggleKey(*active, it.Key, p.Accordion)
					changed = true
				}
				if turn == 0 && !open {
					return
				}
				clip := ui.Column(c).FillWidth().Shrink(0)
				var inner ui.Element
				clip.Children(func() {
					inner = ui.Column(c).FillWidth().Shrink(0).Gap(t.SpaceSM).Padding(pick(p.Ghost, 4, padX), padX, padX).
						Background(pick(p.Ghost, ui.Transparent, t.Surface))
					if !p.Ghost {
						inner.Draw(topRule(t.Border))
					}
					if it.Body != nil {
						inner.Children(it.Body)
					}
				})
				if turn < 1 {
					clip.ClipY().Height(turn * inner.Bounds().H)
				}
			})
		}
	})
	return changed
}

// toggleKey opens key, closing the others when accordion, or closes it.
func toggleKey(active []string, key string, accordion bool) []string {
	if i := slices.Index(active, key); i >= 0 {
		return slices.Delete(slices.Clone(active), i, i+1)
	}
	if accordion {
		return []string{key}
	}
	return append(slices.Clone(active), key)
}

// expandArrow is the header arrow: pointing right when turn is 0 and
// down when it is 1.
func expandArrow(c *ui.Context, color ui.Color, turn float32) {
	ui.Box(c).Size(12, 12).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		a := float64(turn) * math.Pi / 2
		sin, cos := float32(math.Sin(a)), float32(math.Cos(a))
		at := func(x, y float32) (float32, float32) {
			x, y = (x-0.5)*r.W, (y-0.5)*r.H
			return cx + x*cos - y*sin, cy + x*sin + y*cos
		}
		var path ui.Path
		path.MoveTo(at(0.35, 0.15)).LineTo(at(0.68, 0.5)).LineTo(at(0.35, 0.85))
		p.StrokePath(&path, 1.4, color)
	})
}
