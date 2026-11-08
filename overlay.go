package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// popupLook is the white floating panel of Ant's popover, select and
// message (boxShadowSecondary).
func popupLook(e ui.Element, t *Tokens) ui.Element {
	return shadowSecondary(e.Background(t.Elevated).Radius(t.RadiusPopup))
}

// tipLook is Ant's dark tooltip bubble.
func tipLook(c *ui.Context, e ui.Element, text string) {
	t := Of(c)
	e.Padding(6, 8).Radius(6).MaxWidth(250).MinHeight(32).Background(t.TooltipBg).
		Shadow(0, 6, 16, 0, ui.RGBA(0, 0, 0, 0.08)).Shadow(0, 3, 6, -4, ui.RGBA(0, 0, 0, 0.12))
	e.Children(func() {
		ui.Text(c, text).FontSize(t.FontSize).TextColor(t.TooltipText)
	})
}

// TooltipProps are a tooltip's options.
type TooltipProps struct {
	// Placement is the side of the anchor it opens on; PlaceTop by default.
	Placement Placement
	// MouseEnterDelay is how long the pointer rests before it shows;
	// 100ms when zero (Ant). MouseLeaveDelay is how long after it leaves
	// before it hides; 150ms when zero. Keyboard focus (FocusVisible)
	// shows it at once. ui.TooltipBase is 600ms with no delay API.
	MouseEnterDelay, MouseLeaveDelay time.Duration
}

// TooltipOn shows Ant's dark tooltip with text, and an arrow pointing at
// anchor, while the pointer rests on anchor or the keyboard focus is in
// it. Escape dismisses it until the pointer or focus leaves.
func TooltipOn(c *ui.Context, anchor ui.Element, text string, props ...TooltipProps) {
	if text == "" {
		return
	}
	p := first(props)
	t := Of(c)
	hold := ui.Local(anchor, "tip", func() tooltipHold { return tooltipHold{} })
	now := c.Now()
	over, keyed := anchor.Hovered(), anchor.FocusVisible()
	if (anchor.Pressed() || anchor.Clicked()) && over {
		hold.dismissed = true
		hold.open = false
	}
	if hold.dismissed && !over && !keyed {
		hold.dismissed = false
	}
	enter := pick(p.MouseEnterDelay == 0, tooltipEnterDelay, p.MouseEnterDelay)
	leave := pick(p.MouseLeaveDelay == 0, tooltipLeaveDelay, p.MouseLeaveDelay)
	switch {
	case hold.dismissed:
		hold.open = false
		hold.enter, hold.leave = time.Time{}, time.Time{}
	case keyed:
		hold.open = true
		hold.enter, hold.leave = now, time.Time{}
	case over:
		hold.leave = time.Time{}
		if hold.enter.IsZero() {
			hold.enter = now
		}
		if !hold.open {
			if wait := enter - now.Sub(hold.enter); wait > 0 {
				c.After(wait)
			} else {
				hold.open = true
			}
		}
	default:
		hold.enter = time.Time{}
		if hold.open {
			if hold.leave.IsZero() {
				hold.leave = now
			}
			if wait := leave - now.Sub(hold.leave); wait > 0 {
				c.After(wait)
			} else {
				hold.open = false
			}
		}
	}
	if !hold.open {
		return
	}
	ui.Overlay(c, func() {
		tip := ui.Box(c).Absolute().PassThrough().Role(ui.RoleTooltip)
		arrowPanel(c, tip, anchor, p.Placement, t.TooltipBg, func(bubble ui.Element) {
			tipLook(c, bubble, text)
		})
		if tip.OverlayShortcut(0, ui.KeyEscape) {
			hold.dismissed = true
			hold.open = false
		}
	})
}

// Tooltip wraps what fn builds, as a kit Button, in an anchor carrying a
// tooltip, since kit controls report clicks rather than return elements.
func Tooltip(c *ui.Context, text string, fn func(), props ...TooltipProps) {
	anchor := ui.Row(c).Shrink(0).AlignSelf(ui.Start).AlignItems(ui.Center).Children(fn)
	TooltipOn(c, anchor, text, props...)
}

// PopconfirmProps are a pop confirm's options.
type PopconfirmProps struct {
	Description string
	// OK and Cancel label the buttons; "确定" and "取消" by default.
	OK, Cancel string
	// Danger makes OK a danger button.
	Danger bool
	// Placement is the side of the trigger it opens on; PlaceTop by
	// default, as Ant.
	Placement Placement

	open bool
}

// Popconfirm is Ant's Popconfirm: trigger builds the control, as a kit
// Button, and returns its click, which opens a small panel with an arrow
// above it (or on the side of Placement) asking title. It reports true
// when the user confirms; Cancel, Escape or a click outside closes it.
func Popconfirm(c *ui.Context, title string, trigger func() bool, props ...PopconfirmProps) bool {
	p := first(props)
	t := Of(c)
	ok := false
	Popup(c, trigger, func(bubble ui.Element, close func()) {
		popupLook(bubble.Padding(12).Gap(t.SpaceSM).MaxWidth(320), t)
		bubble.Children(func() {
			ui.Row(c).Gap(t.SpaceSM).AlignItems(ui.Start).Children(func() {
				ui.Box(c).Padding(3, 0, 0).Children(func() { statusIcon(c, Warning, 14, t.WarningIcon) })
				ui.Column(c).Gap(t.SpaceXS).Grow(1).Children(func() {
					ui.Text(c, title).FontWeight(600).FontSize(t.FontSize).TextColor(t.Text)
					if p.Description != "" {
						ui.Text(c, p.Description).FontSize(t.FontSize).TextColor(t.Text)
					}
				})
			})
			ui.Row(c).FillWidth().Gap(t.SpaceSM).Justify(ui.End).Padding(t.SpaceXS, 0, 0).Children(func() {
				if Button(c, pick(p.Cancel == "", L(c).Cancel, p.Cancel), ButtonProps{Size: Small}) {
					close()
				}
				if Button(c, pick(p.OK == "", L(c).OK, p.OK), ButtonProps{Type: Primary, Size: Small, Danger: p.Danger}) {
					close()
					ok = true
				}
			})
		})
	}, PopupProps{Placement: p.Placement, Arrow: true, open: p.open})
	return ok
}

func calendarMark(c *ui.Context, color ui.Color) {
	ui.Box(c).Size(14, 14).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		box := ui.Rect{X: r.X + 1, Y: r.Y + 2.5, W: r.W - 2, H: r.H - 3.5}
		p.Stroke(box, color, 1.5, 1.2)
		p.Fill(ui.Rect{X: box.X, Y: box.Y + 3, W: box.W, H: 1.2}, color, 0)
		p.Fill(ui.Rect{X: r.X + 4, Y: r.Y + 1, W: 1.2, H: 3}, color, 0)
		p.Fill(ui.Rect{X: r.X + r.W - 5.2, Y: r.Y + 1, W: 1.2, H: 3}, color, 0)
	})
}

// MessagesProps are the message stack's options, as Ant's message.config.
type MessagesProps struct {
	// Top keeps the messages below a header, where they would cover its
	// controls; Ant's 8 when zero.
	Top float32
	// Below are page titles and navigation the messages must not cover:
	// they go 8 under the lowest one still on screen, as measured in the
	// previous frame, when that is below Top.
	Below []ui.Element
	// BelowPtr, when set, is read when Messages builds, so titles collected
	// in the same frame (kit.App runs the page first) are included.
	BelowPtr *[]ui.Element
	// API holds the messages; when nil the stack keeps its own, which
	// only the toasts of c.AddToast fill.
	API *MessageAPI
	// MaxCount closes the oldest beyond it; no limit when zero, as Ant.
	MaxCount int
	// Duration is how long a message without its own shows; 3s when zero.
	Duration time.Duration
}

// messagesTop is where Messages starts: Top, or under the lowest of the
// rectangles of Below.
func messagesTop(top float32, below []ui.Rect) float32 {
	top = pick(top == 0, 8, top)
	for _, b := range below {
		if b.W > 0 && b.H > 0 && b.Y+b.H > 0 {
			top = max(top, b.Y+b.H+8)
		}
	}
	return top
}

const messageTime = 3 * time.Second

// toastsShown is how many toasts ui hands ToastViewportBase at most
// (toastLimit, ui/toast.go line 16); a full hand means older ones wait.
const toastsShown = 3

// Messages draws the window's messages as Ant's message: white pills at
// the top center with the status mark of their kind ("success", "error",
// "warning", "loading", else info), oldest on top, each sliding in and
// out, its time held while the pointer rests on it. They are the kit's
// own, in the overlay, so there is no limit but MaxCount. The toasts of
// c.AddToast join them too. Build it once in the view.
func Messages(c *ui.Context, props ...MessagesProps) {
	p := first(props)
	t := Of(c)
	els := p.Below
	if p.BelowPtr != nil {
		els = *p.BelowPtr
	}
	below := make([]ui.Rect, len(els))
	for i, e := range els {
		below[i] = e.Bounds()
	}
	top := messagesTop(p.Top, below)
	var toasts []ui.Toast
	ui.ToastViewportBase(c, func(_ ui.Element, list []ui.Toast) { toasts = append(toasts, list...) })
	def := pick(p.Duration == 0, messageTime, p.Duration)
	ui.Overlay(c, func() {
		col := ui.Column(c.Key("kit.messages")).Absolute().Left(0).Right(0).Top(top).
			AlignItems(ui.Center).Gap(t.SpaceSM).PassThrough()
		m := p.API
		if m == nil {
			m = *ui.Local(col, "messages", func() *MessageAPI { return &MessageAPI{} })
		}
		m.drain(c, toasts)
		over, next := m.tick(c.Now(), def)
		for _, k := range over {
			m.Destroy(k)
		}
		if p.MaxCount > 0 {
			if keys := m.Keys(); len(keys) > p.MaxCount {
				for _, k := range keys[:len(keys)-p.MaxCount] {
					m.Destroy(k)
				}
			}
		}
		if next > 0 {
			c.After(next)
		}
		m.mu.Lock()
		list := append([]*messageItem(nil), m.list...)
		m.mu.Unlock()
		announced := ui.Local(col, "announced", func() map[string]string { return map[string]string{} })
		live := make(map[string]bool, len(list))
		for _, it := range list {
			live[it.Key] = true
			if (*announced)[it.Key] != it.Content {
				c.Announce(it.Content)
				(*announced)[it.Key] = it.Content
			}
		}
		for k := range *announced {
			if !live[k] {
				delete(*announced, k)
			}
		}
		col.Children(func() {
			for _, it := range list {
				e := ui.Row(c.Key(it.Key)).AlignSelf(ui.Center).Transition(transitionFor(c, messageMotion))
				messageLook(c, e, it.Kind, it.Content)
				hover := e.Hovered()
				m.mu.Lock()
				it.hover = hover
				m.mu.Unlock()
			}
		})
	})
}

// messageMotion is Ant's move-up: in from above, fading, out the same way
// while the ones under it close up.
var messageMotion = ui.ElementTransition{
	Duration: 200 * time.Millisecond,
	Enter:    &ui.Motion{Y: -16, Opacity: 0},
	Exit:     &ui.Motion{Y: -16, Opacity: 0, Collapse: true},
}

// drain moves the toasts of c.AddToast into m, closing them in ui
// without running their OnClose, which m runs when it closes them. ui
// hands over the newest three; older ones come in the next frames and go
// before them.
func (m *MessageAPI) drain(c *ui.Context, toasts []ui.Toast) {
	m.mu.Lock()
	if len(toasts) == 0 {
		m.draining = false
		m.mu.Unlock()
		return
	}
	cfgs := make([]MessageConfig, len(toasts))
	for i, ts := range toasts {
		d := ts.Timeout
		if d < 0 {
			d = -1
		}
		text := ts.Title
		if ts.Description != "" {
			text += " " + ts.Description
		}
		cfgs[i] = MessageConfig{Key: ts.ID, Kind: ts.Type, Content: text, Duration: d, OnClose: ts.OnClose}
	}
	at := len(m.list)
	if m.draining {
		at = m.drainAt
	}
	m.put(at, cfgs)
	m.drainAt, m.draining = at, len(toasts) >= toastsShown
	m.mu.Unlock()
	for _, ts := range toasts {
		// Stripped of its OnClose and its text first, which AddToast
		// would read out again.
		c.AddToast(ui.Toast{ID: ts.ID})
		c.CloseToast(ts.ID)
	}
}

func messageLook(c *ui.Context, e ui.Element, kind, text string) {
	t := Of(c)
	popupLook(e.Row().Gap(t.SpaceSM).Padding(9, 12).AlignItems(ui.Center).MaxWidth(520), t)
	e.Children(func() {
		if kind == "loading" {
			IconNamed(c, "loading", IconProps{Size: 16, Color: t.Primary, Spin: true})
		} else {
			mk := messageKind(kind)
			_, _, tone, _ := alertColors(t, mk)
			statusIcon(c, mk, 16, tone)
		}
		ui.Text(c, text).FontSize(t.FontSize).TextColor(t.Text)
	})
}

// Message is one message pill outside the stack, for the gallery and for
// anything that shows a message in place.
func Message(c *ui.Context, kind string, text string) {
	messageLook(c, ui.Row(c).AlignSelf(ui.Start), kind, text)
}

func messageKind(typ string) AlertType {
	switch typ {
	case "success":
		return Success
	case "error":
		return Error
	case "warning":
		return Warning
	}
	return Info
}
