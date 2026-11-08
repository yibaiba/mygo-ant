package kit

import (
	"strconv"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
)

// NoticePlacement is the window corner or edge notifications gather at.
type NoticePlacement uint8

// Notification placements, Ant's topRight by default.
const (
	NoticeTopRight NoticePlacement = iota
	NoticeTopLeft
	NoticeBottomRight
	NoticeBottomLeft
	NoticeTop
	NoticeBottom
	noticePlacements
)

func (p NoticePlacement) bottom() bool {
	return p == NoticeBottomRight || p == NoticeBottomLeft || p == NoticeBottom
}

// NoticeKind is a notification's status mark.
type NoticeKind uint8

// Notification kinds: NoticeOpen has no mark, as Ant's notification.open.
const (
	NoticeOpen NoticeKind = iota
	NoticeInfo
	NoticeSuccess
	NoticeWarning
	NoticeError
)

// Notice is one notification, as Ant's notification.open config.
type Notice struct {
	// Key names it: opening one with the key of one showing updates that
	// one and restarts its time. Open makes one up when empty.
	Key         string
	Kind        NoticeKind
	Title       string
	Description string
	// Icon replaces the mark of Kind.
	Icon *ui.SVG
	// Btn builds the buttons at the bottom right; close closes it.
	Btn func(close func())
	// Duration is how long it shows: 4.5s when zero, until closed when
	// negative.
	Duration  time.Duration
	Placement NoticePlacement
	// ShowProgress draws the time left as a bar along the bottom.
	ShowProgress bool
	// NoPause keeps the time running while the pointer rests on it.
	NoPause bool
	// NoClose drops the close button.
	NoClose bool
	OnClick func()
	// OnClose runs as it closes, by time, button or Close.
	OnClose func()
}

const noticeTime = 4500 * time.Millisecond

func (n *Notice) life() time.Duration {
	switch {
	case n.Duration == 0:
		return noticeTime
	case n.Duration < 0:
		return 0
	}
	return n.Duration
}

type noticeItem struct {
	Notice
	left  time.Duration
	last  time.Time
	hover bool
	h     float32
}

// Notifications holds the window's notifications, as the api of Ant's
// notification.useNotification: open them from the view or another
// goroutine, and build NotificationHolder once in the view.
type Notifications struct {
	mu   sync.Mutex
	list []*noticeItem
	seq  int
	// woke asks the view to build again when a goroutine opens one; set it
	// to the app's redraw.
	Wake func()
}

// Open shows n and returns its key.
func (ns *Notifications) Open(n Notice) string {
	ns.mu.Lock()
	if n.Key == "" {
		ns.seq++
		n.Key = "notice-" + strconv.Itoa(ns.seq)
	}
	item := &noticeItem{Notice: n, left: n.life()}
	replaced := false
	for i, have := range ns.list {
		if have.Key == n.Key {
			item.h = have.h
			ns.list[i] = item
			replaced = true
		}
	}
	if !replaced {
		ns.list = append(ns.list, item)
	}
	wake := ns.Wake
	ns.mu.Unlock()
	if wake != nil {
		wake()
	}
	return n.Key
}

// Info, Success, Warning and Error open a notification with a mark.
func (ns *Notifications) Info(title, desc string) string {
	return ns.Open(Notice{Kind: NoticeInfo, Title: title, Description: desc})
}

func (ns *Notifications) Success(title, desc string) string {
	return ns.Open(Notice{Kind: NoticeSuccess, Title: title, Description: desc})
}

func (ns *Notifications) Warning(title, desc string) string {
	return ns.Open(Notice{Kind: NoticeWarning, Title: title, Description: desc})
}

func (ns *Notifications) Error(title, desc string) string {
	return ns.Open(Notice{Kind: NoticeError, Title: title, Description: desc})
}

// Close closes the notification with key, or every one when key is empty
// (Ant's destroy), and runs their OnClose.
func (ns *Notifications) Close(key string) {
	ns.mu.Lock()
	var gone []func()
	kept := ns.list[:0]
	for _, it := range ns.list {
		if key == "" || it.Key == key {
			if it.OnClose != nil {
				gone = append(gone, it.OnClose)
			}
			continue
		}
		kept = append(kept, it)
	}
	clear(ns.list[len(kept):])
	ns.list = kept
	ns.mu.Unlock()
	for _, fn := range gone {
		fn()
	}
}

// Keys are the keys of the notifications showing, oldest first.
func (ns *Notifications) Keys() []string {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	out := make([]string, len(ns.list))
	for i, it := range ns.list {
		out[i] = it.Key
	}
	return out
}

// tick spends the time between the last frame and now on the notices
// whose time runs, and returns the keys whose time is over and how long
// until the next one's is.
func (ns *Notifications) tick(now time.Time) (over []string, next time.Duration, running bool) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	for _, it := range ns.list {
		life := it.life()
		if it.last.IsZero() {
			it.last = now
		}
		if life > 0 && !(it.hover && !it.NoPause) {
			it.left -= now.Sub(it.last)
			running = running || it.ShowProgress
		}
		it.last = now
		if life <= 0 {
			continue
		}
		if it.left <= 0 {
			over = append(over, it.Key)
		} else if next == 0 || it.left < next {
			next = it.left
		}
	}
	return over, next, running
}

// NotificationsProps are the notification holder's options.
type NotificationsProps struct {
	// Top and Bottom are the distances of the top and bottom placements
	// from the window's edges; 24 when zero.
	Top, Bottom float32
	// NoStack lists every notification; by default more than Threshold
	// (3) in one place fold into a stack that opens under the pointer, as
	// Ant's stack.
	NoStack   bool
	Threshold int
	// MaxCount closes the oldest beyond it; no limit when zero.
	MaxCount int
}

// NotificationHolder builds the notifications of ns at their placements;
// build it once in the view, as Ant's contextHolder.
func NotificationHolder(c *ui.Context, ns *Notifications, props ...NotificationsProps) {
	p := first(props)
	t := Of(c)
	over, next, running := ns.tick(c.Now())
	for _, k := range over {
		ns.Close(k)
	}
	if p.MaxCount > 0 {
		if keys := ns.Keys(); len(keys) > p.MaxCount {
			for _, k := range keys[:len(keys)-p.MaxCount] {
				ns.Close(k)
			}
		}
	}
	if next > 0 {
		c.After(next)
	}
	if running {
		c.AnimationFrame()
	}
	ns.mu.Lock()
	groups := make([][]*noticeItem, noticePlacements)
	for _, it := range ns.list {
		groups[it.Placement] = append(groups[it.Placement], it)
	}
	ns.mu.Unlock()
	threshold := pick(p.Threshold == 0, 3, p.Threshold)
	f := t.Feedback
	ui.Overlay(c, func() {
		for place := range noticePlacements {
			list := groups[place]
			at := place
			col := ui.Column(c.Key("notices-" + strconv.Itoa(int(place)))).Absolute().Gap(f.NoticeGap).Width(f.NoticeWidth).PassThrough()
			top := pick(p.Top == 0, f.NoticeMargin, p.Top)
			bottom := pick(p.Bottom == 0, f.NoticeMargin, p.Bottom)
			winW, _ := c.Size()
			switch at {
			case NoticeTopRight:
				col.Top(top).Right(f.NoticeMargin)
			case NoticeTopLeft:
				col.Top(top).Left(f.NoticeMargin)
			case NoticeBottomRight:
				col.Bottom(bottom).Right(f.NoticeMargin)
			case NoticeBottomLeft:
				col.Bottom(bottom).Left(f.NoticeMargin)
			case NoticeTop:
				col.Top(top).Left((winW - f.NoticeWidth) / 2)
			case NoticeBottom:
				col.Bottom(bottom).Left((winW - f.NoticeWidth) / 2)
			}
			stacked := !p.NoStack && len(list) > threshold
			if stacked && !col.Hovered() {
				col.Children(func() { noticeStack(c, t, ns, list, at) })
				continue
			}
			from := pick[float32](at == NoticeTopLeft || at == NoticeBottomLeft, -f.NoticeWidth, f.NoticeWidth)
			if at == NoticeTop || at == NoticeBottom {
				from = 0
			}
			col.Children(func() {
				// Newest nearest the edge it gathers at.
				for i := range list {
					it := list[len(list)-1-i]
					if at.bottom() {
						it = list[i]
					}
					e := noticeCard(c.Key(it.Key), t, ns, it).Transition(transitionFor(c, listMotion(from)))
					it.h = e.Bounds().H
				}
			})
		}
	})
}

// noticeStack folds the notices of one place: the newest in full, and up
// to two older ones as edges peeking out under it, narrower each.
func noticeStack(c *ui.Context, t *Tokens, ns *Notifications, list []*noticeItem, at NoticePlacement) {
	newest := list[len(list)-1]
	peek := min(len(list)-1, 2)
	host := ui.Column(c).FillWidth().Padding(pick[float32](at.bottom(), float32(peek)*8, 0), 0, pick[float32](at.bottom(), 0, float32(peek)*8), 0)
	// The edges are painted under the children: positioned elements would
	// paint over the newest card.
	h := max(newest.h, 40)
	host.Draw(func(pa *ui.Painter, r ui.Rect) {
		for i := peek; i >= 1; i-- {
			inset, shift := float32(i)*12, float32(i)*8
			y := r.Y + shift
			if at.bottom() {
				y = r.Y + r.H - h - shift
			}
			edge := ui.Rect{X: r.X + inset, Y: y, W: r.W - 2*inset, H: h}
			pa.Shadow(edge, 8, 0, 6, 16, 0, ui.RGBA(0, 0, 0, 0.08))
			pa.Shadow(edge, 8, 0, 3, 6, -4, ui.RGBA(0, 0, 0, 0.12))
			pa.Fill(edge, t.Elevated, 8)
		}
	})
	host.Children(func() {
		e := noticeCard(c.Key(newest.Key), t, ns, newest)
		newest.h = e.Bounds().H
	})
}

// noticeColor is the mark color of a kind.
func noticeMark(t *Tokens, k NoticeKind) (AlertType, ui.Color) {
	kind := AlertType(max(int(k)-1, 0))
	_, _, icon, _ := alertColors(t, kind)
	return kind, icon
}

// noticeCard is one notification: mark, title, description, buttons and
// the time-left bar, with its close button at the top right.
func noticeCard(c *ui.Context, t *Tokens, ns *Notifications, it *noticeItem) ui.Element {
	card := ui.Column(c).FillWidth().Padding(20, 24).Gap(8).Role(ui.RoleStatus)
	popupLook(card.Radius(8), t)
	if it.OnClick != nil {
		card.Cursor(ui.CursorPointer)
		if card.Clicked() {
			it.OnClick()
		}
	}
	if ns != nil {
		it.hover = card.Hovered() || card.FocusWithin()
	}
	life := it.life()
	card.Children(func() {
		ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Start).Children(func() {
			switch {
			case it.Icon != nil:
				_, tone := noticeMark(t, it.Kind)
				ui.Icon(c, it.Icon).Size(24, 24).TextColor(pick(it.Kind == NoticeOpen, t.Primary, tone)).Shrink(0)
			case it.Kind != NoticeOpen:
				kind, tone := noticeMark(t, it.Kind)
				statusIcon(c, kind, 24, tone)
			}
			ui.Column(c).Grow(1).MinWidth(0).Gap(8).Children(func() {
				ui.Text(c, it.Title).FontSize(t.FontSizeLG).TextColor(t.TextHeading).MarginEnd(pick[float32](it.NoClose, 0, 24))
				if it.Description != "" {
					ui.Text(c, it.Description).FontSize(t.FontSize).TextColor(t.Text)
				}
			})
		})
		if it.Btn != nil {
			ui.Row(c).FillWidth().Gap(t.SpaceSM).Justify(ui.End).Margin(8, 0, 0).Children(func() {
				it.Btn(func() {
					if ns != nil {
						ns.Close(it.Key)
					}
				})
			})
		}
		if !it.NoClose {
			ui.Row(c).Absolute().Top(18).Right(20).Children(func() {
				if closeButton(c, t, 14) && ns != nil {
					ns.Close(it.Key)
				}
			})
		}
		if it.ShowProgress && life > 0 {
			frac := max(0, min(float32(it.left)/float32(life), 1))
			ui.Box(c).Absolute().Left(0).Right(0).Bottom(0).Height(2).PassThrough().Draw(func(pa *ui.Painter, r ui.Rect) {
				inset := ui.Rect{X: r.X + 8, Y: r.Y, W: (r.W - 16) * frac, H: r.H}
				if inset.W > 0 {
					pa.Fill(inset, t.Primary, 1)
				}
			})
		}
	})
	return card
}

// Notification is one notification card in place, for galleries and
// anything that shows one inside a page.
func Notification(c *ui.Context, n Notice) {
	it := &noticeItem{Notice: n, left: n.life()}
	ui.Column(c).Width(Of(c).Feedback.NoticeWidth).Shrink(0).Children(func() {
		noticeCard(c, Of(c), nil, it)
	})
}
