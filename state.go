package kit

import "github.com/egoist/mygo/ui"

// visual is the look a control takes from its interaction state.
type visual uint8

const (
	rest visual = iota
	hover
	active
	focus
	disabled
)

// stateOf reads the pointer and focus state of e. A non-zero force pins
// the look, which the gallery uses to show every state in one frame.
func stateOf(e ui.Element, off bool, force visual) visual {
	switch {
	case off:
		return disabled
	case force != rest:
		return force
	case e.Pressed():
		return active
	case e.FocusVisible():
		return focus
	case e.Hovered():
		return hover
	}
	return rest
}

// focusRing is Ant's control outline: a 2px primary-alpha glow outside the box.
func focusRing(e ui.Element, t *Tokens) {
	e.Shadow(0, 0, 0, 2, t.ControlOutline)
}

// shadowSecondary is Ant's boxShadowSecondary, the lift of popups and dialogs.
func shadowSecondary(e ui.Element) ui.Element {
	return e.Shadow(0, 6, 16, 0, ui.RGBA(0, 0, 0, 0.08)).
		Shadow(0, 3, 6, -4, ui.RGBA(0, 0, 0, 0.12)).
		Shadow(0, 9, 28, 8, ui.RGBA(0, 0, 0, 0.05))
}

func pick[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
