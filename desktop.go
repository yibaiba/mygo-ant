package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Ant tooltip delays (mouseEnterDelay / mouseLeaveDelay). ui.TooltipBase
// is 600ms with no delay API, so kit paints its own.
const (
	tooltipEnterDelay = 100 * time.Millisecond
	tooltipLeaveDelay = 150 * time.Millisecond
)

// WindowMinSize is a typical desktop minimum in DIPs for
// mygo.WindowOptions.MinWidth / MinHeight. Kit does not import mygo; apps
// pass these into WindowOptions. Apps set remembered bounds with
// WindowOptions.StateKey.
func WindowMinSize() (width, height int) { return 960, 600 }

// tooltipHold is how long a kit tooltip has been hovered or left.
type tooltipHold struct {
	enter, leave time.Time
	open         bool
	dismissed    bool
}

// wheelStep is +1 for a wheel notch up (raise a slider / previous page)
// and -1 for a notch down. Zero when the wheel did not move vertically.
func wheelStep(dy float32) int {
	switch {
	case dy < 0:
		return 1
	case dy > 0:
		return -1
	}
	return 0
}

// ellipsisTip shows the full text when e's last-frame box is narrower
// than the text, as Ant ellipsis + tooltip on table and list cells.
func ellipsisTip(c *ui.Context, e ui.Element, text string, size float32) {
	if text == "" || !e.Valid() {
		return
	}
	b := e.Bounds()
	if b.W <= 0 {
		return
	}
	w, _ := c.MeasureText(0, ui.Span{Text: text, Size: size})
	if w > b.W+1 {
		TooltipOn(c, e, text)
	}
}
