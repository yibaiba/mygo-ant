package kit

import (
	"sync"

	"github.com/egoist/mygo/ui"
)

// PixelRatio is the view's zoom (Context.Zoom), 1 when unset. Lengths in
// the kit are DIPs; MyGo already paints 1px borders on the device grid.
// Window-button 1.5 DIP hairlines also round onto this grid (snapDIP)
// before Painter.Stroke/Fill snap them to whole device pixels.
func PixelRatio(c *ui.Context) float32 {
	z := c.Zoom()
	if z <= 0 {
		return 1
	}
	return z
}

// Vibrancy reports whether the window shows a system material behind
// transparent fills (mygo.WindowOptions.Vibrancy).
func Vibrancy(c *ui.Context) bool { return c.Vibrancy() }

// TitleBarInsets is the room the system window controls take in a window
// with a hidden title bar (ui.Context.TitleBar).
func TitleBarInsets(c *ui.Context) ui.TitleBar { return c.TitleBar() }

// Desktop is the OS settings the default theme follows (accent, reduced
// motion, high contrast, text scale).
func Desktop(c *ui.Context) ui.Preferences { return c.Preferences() }

// DesktopDark is the window's OS light/dark appearance, from the default
// theme (the host's IsDark; tester.SetDark in tests). Call it at the
// start of a view, before Apply, Page or ConfigProvider replace the
// theme with kit tokens. It does not call mygo.Theme.IsDark, which
// panics on the main goroutine before App.Run and therefore inside
// ui.NewTester.
func DesktopDark(c *ui.Context) bool {
	th := c.Theme()
	return th != nil && th.Dark
}

type followKey struct {
	src      *Tokens
	contrast bool
	dark     bool
	scale    int
}

var followed = map[followKey]*Tokens{}

// FollowDesktop returns t adjusted for the desktop's HighContrast and
// TextScale. It is opt-in: Apply and ConfigProvider keep Ant tokens
// unless ConfigProps.FollowDesktop is set, or the caller passes the
// result of FollowDesktop to Apply. ReduceMotion is already honoured by
// ui.Animate. The product primary stays Ant's; OS accent is not mixed in.
// High contrast is black/white PageBg, Surface, Text and Border with an
// opaque ControlOutline, not a Mix of Text into Border. The same *Tokens
// is returned when the desktop is ordinary, so Of still finds Work / Home.
func FollowDesktop(c *ui.Context, t *Tokens) *Tokens {
	if t == nil {
		t = Work
	}
	p := c.Preferences()
	scale := p.TextScale
	if scale <= 0 {
		scale = 1
	}
	if !p.HighContrast && scale == 1 {
		return t
	}
	k := followKey{src: t, contrast: p.HighContrast, dark: t.Dark, scale: int(scale*100 + 0.5)}
	registryMu.Lock()
	got, ok := followed[k]
	registryMu.Unlock()
	if ok {
		return got
	}
	cp := *t
	cp.theme = nil
	if p.HighContrast {
		black, white := ui.Hex("#000000"), ui.Hex("#FFFFFF")
		page, ink := white, black
		if t.Dark {
			page, ink = black, white
		}
		cp.PageBg, cp.Surface, cp.Elevated, cp.SurfaceSubtle, cp.DisabledBg, cp.FillAlter = page, page, page, page, page, page
		cp.Text, cp.TextHeading, cp.TextSecondary, cp.TextMuted, cp.TextTertiary = ink, ink, ink, ink, ink
		cp.TextDisabled, cp.TextPlaceholder = ink, ink
		cp.Border, cp.BorderButton, cp.SheetBorder, cp.Split = ink, ink, ink, ink
		cp.BorderSecondary, cp.DividerColor = ink, ink
		cp.ControlOutline = cp.Primary
	}
	if scale != 1 {
		cp.FontSize *= scale
		cp.FontSizeSM *= scale
		cp.FontSizeLG *= scale
		cp.FontSizeTitle *= scale
		cp.FontSizeHeading *= scale
		cp.LabelSize *= scale
		cp.LabelSizeLG *= scale
		cp.ControlHeightSM *= scale
		cp.ControlHeight *= scale
		cp.ControlHeightLG *= scale
	}
	_ = cp.Theme()
	registryMu.Lock()
	defer registryMu.Unlock()
	if got, ok := followed[k]; ok {
		return got
	}
	followed[k] = &cp
	return &cp
}

// Copy writes text to the system clipboard.
func Copy(c *ui.Context, text string) {
	if text != "" {
		c.WriteClipboard(text)
	}
}

// Paste returns the text on the system clipboard.
func Paste(c *ui.Context) string { return c.ReadClipboard() }

// MessageBoxKind is the icon of a native message dialog, as
// mygo.MessageType.
type MessageBoxKind uint8

// Native message kinds.
const (
	MessageNone MessageBoxKind = iota
	MessageInfo
	MessageWarning
	MessageError
	MessageQuestion
)

func (k MessageBoxKind) String() string {
	switch k {
	case MessageInfo:
		return "info"
	case MessageWarning:
		return "warning"
	case MessageError:
		return "error"
	case MessageQuestion:
		return "question"
	}
	return "none"
}

// MessageBoxOptions are what a native message dialog is asked for
// (mygo.Dialog.Message).
type MessageBoxOptions struct {
	Title, Message, Detail string
	Kind                   MessageBoxKind
	Buttons                []string
	Default, Cancel        int
	Destructive            []int
	CheckboxLabel          string
	CheckboxChecked        bool
}

// MessageBoxResult is which button was pressed, and the checkbox.
type MessageBoxResult struct {
	Button          int
	CheckboxChecked bool
}

// MessageBox shows a system message dialog and may block; the kit calls
// it off the UI goroutine.
type MessageBox func(MessageBoxOptions) (MessageBoxResult, error)

// OSNotifyOptions are a desktop notification (mygo.Notification).
type OSNotifyOptions struct {
	Title, Body, Subtitle string
	Silent                bool
}

// OSNotify shows a desktop notification; it must not block the UI.
type OSNotify func(OSNotifyOptions) error

var (
	nativeMu     sync.RWMutex
	messageBoxFn MessageBox
	osNotifyFn   OSNotify
)

// SetMessageBox injects the platform message dialog, as SetFileDialog
// does for files. The app passes mygo.Dialog.Message. Nil clears it.
func SetMessageBox(fn MessageBox) {
	nativeMu.Lock()
	messageBoxFn = fn
	nativeMu.Unlock()
}

func messageBox() MessageBox {
	nativeMu.RLock()
	defer nativeMu.RUnlock()
	return messageBoxFn
}

// SetOSNotify injects desktop notifications (mygo.NewNotification). Nil
// clears it. In-window Notifications stay the Ant replica; this is the
// OS tray banner for when the window is hidden.
func SetOSNotify(fn OSNotify) {
	nativeMu.Lock()
	osNotifyFn = fn
	nativeMu.Unlock()
}

func osNotify() OSNotify {
	nativeMu.RLock()
	defer nativeMu.RUnlock()
	return osNotifyFn
}

// OpenMessageBox opens the dialog set with SetMessageBox on a goroutine
// of its own and calls done there, then asks c for a frame. It reports
// false, opening nothing, when no dialog is set.
func OpenMessageBox(c *ui.Context, opts MessageBoxOptions, done func(MessageBoxResult)) bool {
	fn := messageBox()
	if fn == nil {
		return false
	}
	go func() {
		res, err := fn(opts)
		if err != nil {
			return
		}
		done(res)
		if c != nil {
			c.Invalidate()
		}
	}()
	return true
}

// NotifyOS shows an OS notification when SetOSNotify was called. It
// reports false when none is set.
func NotifyOS(opts OSNotifyOptions) bool {
	fn := osNotify()
	if fn == nil {
		return false
	}
	go func() { _ = fn(opts) }()
	return true
}

// NativeConfirm asks with the injected message box when one is set
// (question, OK / Cancel). danger marks OK destructive. done runs on the
// dialog's goroutine. It reports false when no dialog is set, so the
// caller can fall back to Confirm / Modal.confirm.
func NativeConfirm(c *ui.Context, title, message string, danger bool, done func(ok bool)) bool {
	loc := LocaleZhCN
	if c != nil {
		loc = L(c)
	}
	opts := MessageBoxOptions{
		Title:   title,
		Message: message,
		Kind:    MessageQuestion,
		Buttons: []string{loc.OK, loc.Cancel},
		Default: 0,
		Cancel:  1,
	}
	if danger {
		opts.Destructive = []int{0}
	}
	return OpenMessageBox(c, opts, func(res MessageBoxResult) {
		if done != nil {
			done(res.Button == 0)
		}
	})
}
