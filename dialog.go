package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Dialog width presets, as Ant Modal's 416 / 520 / 736 (confirm, default,
// large — the last matches Drawer Large).
const (
	DialogWidthSM float32 = 416
	DialogWidth   float32 = 520
	DialogWidthLG float32 = 736
)

// DialogProps are a dialog's options.
type DialogProps struct {
	Width, Height float32
	// Small is width 416 when Width is 0 (Modal.confirm); Large is 736.
	Small, Large bool
	// Footer builds the buttons at the bottom right; nil leaves none, or
	// with OK set Ant's cancel / OK pair.
	Footer func()
	// OK and Cancel label the default footer; Cancel defaults to the
	// locale's when OK is set.
	OK, Cancel string
	// OKDanger makes OK a danger button; Loading spins it (confirmLoading).
	OKDanger, Loading bool
	// Top puts the dialog that far from the window's top instead of in
	// its middle (Ant's style.top, centered={false}).
	Top float32
	// NoMaskClose keeps it open on a click on the mask (maskClosable
	// false); NoKeyboard on Escape (keyboard false); NoClose drops the ×
	// (closable false).
	NoMaskClose, NoKeyboard, NoClose bool
	// AfterClose runs once the leave motion has finished (Ant afterClose).
	AfterClose func()
}

// Dialog is Ant's Modal frame over a dimmed window: a title row with a
// close button, the body fn builds, and an optional footer. Clicking the
// mask, the close button or pressing Escape sets *open to false. With OK
// set it returns 0 for a click on OK, 1 for cancel or dismissal, and -1
// otherwise.
func Dialog(c *ui.Context, open *bool, title string, fn func(), props ...DialogProps) int {
	wasOpen := *open
	p := first(props)
	host := motionHost(c)
	show, v := presence(host, "dialog", wasOpen, motionFor(c, dialogMotion))
	closed := ui.Local(host, "afterClose", func() bool { return false })
	seen := ui.Local(host, "wasShown", func() bool { return false })
	if wasOpen {
		*seen = true
	}
	if !show {
		if p.AfterClose != nil && *seen && !*closed {
			*closed = true
			p.AfterClose()
		}
		return -1
	}
	*closed = false
	shown := true
	t := Of(c)
	answer := -1
	ui.DialogBase(c, &shown, func(backdrop, panel ui.Element) {
		backdrop.Background(t.Mask.Alpha(v))
		if p.NoMaskClose && backdrop.Clicked() || p.NoKeyboard && !backdrop.Clicked() && !shown {
			shown = true
		}
		if p.Top > 0 {
			backdrop.Column().Justify(ui.Start).AlignItems(ui.Center).Padding(p.Top, 0, 0)
		}
		zoomIn(panel, v)
		w := dialogWidth(p.Width, p.Small, p.Large, DialogWidth)
		shadowSecondary(panel.Width(w).Background(t.Elevated).Radius(8).Padding(20, 24).Gap(t.SpaceMD))
		if p.Height > 0 {
			panel.Height(p.Height)
		}
		panel.Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
				ui.Text(c, title).FontWeight(600).FontSize(t.FontSizeLG).TextColor(t.Text).Grow(1)
				if !p.NoClose && closeButton(c, t, 14) {
					answer = 1
				}
			})
			ui.Column(c).FillWidth().Grow(1).Gap(t.SpaceMD).Children(fn)
			switch {
			case p.Footer != nil:
				ui.Row(c).FillWidth().Gap(t.SpaceSM).Justify(ui.End).Children(p.Footer)
			case p.OK != "":
				ui.Row(c).FillWidth().Gap(t.SpaceSM).Justify(ui.End).Children(func() {
					if Button(c, pick(p.Cancel == "", L(c).Cancel, p.Cancel), ButtonProps{MinWidth: 64}) {
						answer = 1
					}
					if Button(c, p.OK, ButtonProps{Type: Primary, MinWidth: 64, Danger: p.OKDanger, Loading: p.Loading, AutoFocus: true}) && !p.Loading {
						answer = 0
					}
				})
			}
		})
	})
	if !wasOpen {
		return -1
	}
	if answer == -1 && !shown {
		answer = 1
	}
	if answer >= 0 {
		*open = false
	}
	return answer
}

// dialogMotion is how long a modal takes to zoom in or out.
const dialogMotion = 200 * time.Millisecond

// dialogWidth is Width, or a Small / Large / fallback preset.
func dialogWidth(width float32, small, large bool, fallback float32) float32 {
	if width > 0 {
		return width
	}
	switch {
	case large:
		return DialogWidthLG
	case small:
		return DialogWidthSM
	}
	return fallback
}

// dialogPresence follows a modal's open state so it zooms out after
// closing: it returns the open flag to hand DialogBase, how far in the
// modal is, and whether to build it at all.
func dialogPresence(c *ui.Context, open bool) (*bool, float32, bool) {
	show, v := presence(motionHost(c), "dialog", open, motionFor(c, dialogMotion))
	shown := true
	return &shown, v, show
}

// zoomIn scales and fades a modal panel in as v goes from 0 to 1, as
// Ant's zoom motion.
func zoomIn(panel ui.Element, v float32) {
	s := 0.2 + 0.8*v
	panel.Scale(s, s).Opacity(v)
}

// Confirm is Ant's Modal.confirm: a warning icon, title, message, and
// cancel / OK buttons. It returns 0 for OK, 1 for cancel (or dismissal),
// and -1 while the user has not answered.
func Confirm(c *ui.Context, open *bool, title, message, ok, cancel string) int {
	wasOpen := *open
	shown, v, live := dialogPresence(c, wasOpen)
	if !live {
		return -1
	}
	answer := modalPanel(c, shown, v, ModalFunc{Kind: ModalConfirm, Title: title, Content: message, OK: ok, Cancel: cancel, MaskClosable: true})
	if !wasOpen {
		return -1
	}
	if answer == -1 && !*shown {
		answer = 1
	}
	if answer >= 0 {
		*open = false
	}
	return answer
}
