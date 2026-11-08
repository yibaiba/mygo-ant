package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// ui's Animate starts a value at its first target, so an element that
// should grow or fade in as it first appears jumps the value to where it
// comes from in that frame, and then animates it to its place.

// motionFor is d, or 0 when the ConfigProvider set token.motion false.
// OS ReduceMotion is honoured by ui.Animate even when d stays positive.
func motionFor(c *ui.Context, d time.Duration) time.Duration {
	if c != nil && Of(c).MotionOff {
		return 0
	}
	return d
}

// stillMotion is true when kit enter/leave should jump and when custom
// Now() loops (Progress Active, Badge ripple, Spin, Skeleton) should not
// move: token.motion=false or the OS ReduceMotion preference.
func stillMotion(c *ui.Context) bool {
	if c == nil {
		return false
	}
	return Of(c).MotionOff || c.Preferences().ReduceMotion
}

// enterIn returns how far e, an element that appeared in some frame, is
// on its way in, from 0 to 1 over d: 0 in the frame it appears. d of 0
// (motionFor) is there at once.
func enterIn(e ui.Element, key string, d time.Duration) float32 {
	if d <= 0 {
		return 1
	}
	seen := ui.Local(e, key+".seen", func() bool { return false })
	if !*seen {
		*seen = true
		e.Jump(key, 0)
	}
	return e.AnimateTo(key, 1, ui.Animation{Duration: d, Ease: ui.EaseOut})
}

// presence follows an overlay that animates in as open turns true and out
// as it turns false: it reports whether the overlay should still be built
// (while open, and on its way out) and how far in it is, from 0 to 1.
// host is an element built in every frame, open or not. d of 0 skips the
// leave motion.
func presence(host ui.Element, key string, open bool, d time.Duration) (bool, float32) {
	if d <= 0 {
		return open, pick[float32](open, 1, 0)
	}
	seen := ui.Local(host, key+".seen", func() bool { return false })
	if !*seen {
		*seen = true
		if open {
			host.Jump(key, 0)
		}
	}
	v := host.AnimateTo(key, pick[float32](open, 1, 0), ui.Animation{Duration: d, Ease: ui.EaseOut})
	return open || v > 0.01, v
}

// motionHost is an element built in every frame that takes no room, to
// keep the animations of overlays that are not built while closed.
func motionHost(c *ui.Context) ui.Element {
	var h ui.Element
	ui.Overlay(c, func() {
		h = ui.Box(c).Absolute().Left(0).Top(0).Size(0, 0).PassThrough()
	})
	return h
}

// listMotion is how notifications and the rows of animated lists come and
// go: sliding in from the side as they fade in, collapsing as they leave.
func listMotion(fromX float32) ui.ElementTransition {
	return ui.ElementTransition{
		Duration: 200 * time.Millisecond,
		Enter:    &ui.Motion{X: fromX, Opacity: 0},
		Exit:     &ui.Motion{X: fromX, Opacity: 0, Collapse: true},
	}
}

// transitionFor is t, or none when token.motion is false.
func transitionFor(c *ui.Context, t ui.ElementTransition) ui.ElementTransition {
	if motionFor(c, t.Duration) <= 0 {
		return ui.ElementTransition{}
	}
	return t
}
