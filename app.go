package kit

import (
	"strconv"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
)

// MessageConfig is one message, as Ant's message.open config.
type MessageConfig struct {
	// Key names it: opening one with the key of one showing updates it,
	// as a loading message turning into a success.
	Key string
	// Kind is "info", "success", "warning", "error" or "loading".
	Kind    string
	Content string
	// Duration is how long it shows: 3s when zero, until closed when
	// negative.
	Duration time.Duration
	// NoPause keeps its time running while the pointer rests on it.
	NoPause bool
	OnClose func()
}

// MessageAPI holds the window's messages, as Ant's message.useMessage:
// open them from the view or a goroutine, and build Messages with it once
// in the view (App does).
type MessageAPI struct {
	mu   sync.Mutex
	list []*messageItem
	seq  int
	wake func()
	// drainAt is where the toasts drained from ui last went; draining
	// tells that more may wait behind them, older, to go there too.
	drainAt  int
	draining bool
}

type messageItem struct {
	MessageConfig
	left  time.Duration
	last  time.Time
	hover bool
}

// life is how long it shows with def as the default, 0 until closed.
func (m *messageItem) life(def time.Duration) time.Duration {
	switch {
	case m.Duration < 0:
		return 0
	case m.Duration == 0:
		return def
	}
	return m.Duration
}

// Open shows cfg and returns its key; one with the key of a message
// showing replaces it in place, its time anew.
func (m *MessageAPI) Open(cfg MessageConfig) string {
	m.mu.Lock()
	if cfg.Key == "" {
		m.seq++
		cfg.Key = "message-" + strconv.Itoa(m.seq)
	}
	m.put(len(m.list), []MessageConfig{cfg})
	wake := m.wake
	m.mu.Unlock()
	if wake != nil {
		wake()
	}
	return cfg.Key
}

// put inserts cfgs at i, each in place of a message of its key instead
// when one shows. m.mu is held.
func (m *MessageAPI) put(i int, cfgs []MessageConfig) {
	i = min(max(i, 0), len(m.list))
	var fresh []*messageItem
	for _, cfg := range cfgs {
		it := &messageItem{MessageConfig: cfg}
		replaced := false
		for j, have := range m.list {
			if have.Key == cfg.Key {
				m.list[j] = it
				replaced = true
			}
		}
		if !replaced {
			fresh = append(fresh, it)
		}
	}
	m.list = append(m.list[:i], append(fresh, m.list[i:]...)...)
}

// Info, Success, Warning, Error and Loading open a message of that kind;
// a loading one shows until closed.
func (m *MessageAPI) Info(text string) string {
	return m.Open(MessageConfig{Kind: "info", Content: text})
}
func (m *MessageAPI) Success(text string) string {
	return m.Open(MessageConfig{Kind: "success", Content: text})
}
func (m *MessageAPI) Warning(text string) string {
	return m.Open(MessageConfig{Kind: "warning", Content: text})
}
func (m *MessageAPI) Error(text string) string {
	return m.Open(MessageConfig{Kind: "error", Content: text})
}
func (m *MessageAPI) Loading(text string) string {
	return m.Open(MessageConfig{Kind: "loading", Content: text, Duration: -1})
}

// Destroy closes the message with key, or every one when key is empty,
// and runs their OnClose.
func (m *MessageAPI) Destroy(key string) {
	m.mu.Lock()
	var gone []func()
	kept := m.list[:0]
	for _, it := range m.list {
		if key == "" || it.Key == key {
			if it.OnClose != nil {
				gone = append(gone, it.OnClose)
			}
			continue
		}
		kept = append(kept, it)
	}
	clear(m.list[len(kept):])
	m.list = kept
	wake := m.wake
	m.mu.Unlock()
	for _, fn := range gone {
		fn()
	}
	if wake != nil {
		wake()
	}
}

// Keys are the keys of the messages showing, oldest first.
func (m *MessageAPI) Keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.list))
	for i, it := range m.list {
		out[i] = it.Key
	}
	return out
}

// tick spends the time since the last frame on the messages whose time
// runs (not hovered, unless NoPause), and returns the keys whose time is
// over and how long until the next one's is.
func (m *MessageAPI) tick(now time.Time, def time.Duration) (over []string, next time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, it := range m.list {
		life := it.life(def)
		if it.last.IsZero() {
			it.last, it.left = now, life
		}
		if life > 0 && !(it.hover && !it.NoPause) {
			it.left -= now.Sub(it.last)
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
	return over, next
}

// AppState is what Ant's App component provides through App.useApp: the
// message, notification and modal APIs. Keep one per window in the app's
// state; set Wake to the app's redraw so a goroutine's message shows at
// once.
type AppState struct {
	Message      MessageAPI
	Notification Notifications
	Modal        Modals
	Wake         func()
	// Watermark is the window watermark's state; see AppProps.Watermark.
	Watermark WatermarkGuard
}

// AppProps are the app holder's options.
type AppProps struct {
	Messages      MessagesProps
	Notifications NotificationsProps
	// Config wraps the app in a ConfigProvider when set.
	Config *ConfigProps
	// Watermark tiles the whole window, over the pages, their popups and
	// the holders, every frame. It is an option of the root only: once set
	// it stays for the AppState's life, a later frame may change it but a
	// nil, empty or transparent one keeps the last, and nothing a page
	// builds can switch it off or paint over it.
	Watermark *WatermarkProps
	// OnWatermarkBroken is called, at the next frame, when the watermark
	// was not painted over the whole window (see WatermarkGuard.Intact).
	OnWatermarkBroken func(reason string)
}

// App is Ant's App: it builds fn, then the holders of the message,
// notification and modal APIs of s over it, under the same theme, size
// and locale. Build it once at the root of the view.
func App(c *ui.Context, s *AppState, fn func(), props ...AppProps) {
	p := first(props)
	s.Message.mu.Lock()
	s.Message.wake = s.Wake
	s.Message.mu.Unlock()
	p.Messages.API = &s.Message
	s.Notification.mu.Lock()
	if s.Notification.Wake == nil {
		s.Notification.Wake = s.Wake
	}
	s.Notification.mu.Unlock()
	s.Modal.mu.Lock()
	if s.Modal.Wake == nil {
		s.Modal.Wake = s.Wake
	}
	s.Modal.mu.Unlock()
	body := func() {
		fn()
		Messages(c, p.Messages)
		NotificationHolder(c, &s.Notification, p.Notifications)
		ModalHolder(c, &s.Modal)
		rootWatermark(c, &s.Watermark, p.Watermark, p.OnWatermarkBroken)
	}
	if p.Config != nil {
		ConfigProvider(c, *p.Config, body).Fill()
		return
	}
	body()
}
