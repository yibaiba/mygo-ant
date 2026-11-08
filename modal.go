package kit

import (
	"strconv"
	"sync"

	"github.com/egoist/mygo/ui"
)

// ModalKind is which of Ant's Modal methods a ModalFunc is.
type ModalKind uint8

// Modal methods: Confirm asks with cancel and OK; the others tell, with OK
// alone.
const (
	ModalConfirm ModalKind = iota
	ModalInfo
	ModalSuccess
	ModalWarning
	ModalError
)

// ModalFunc is the config of Ant's Modal.confirm / info / success /
// warning / error.
type ModalFunc struct {
	Kind    ModalKind
	Title   string
	Content string
	// Body builds content under Content.
	Body func()
	// OK and Cancel label the buttons; the locale's when empty.
	OK, Cancel string
	OKDanger   bool
	// Icon replaces the kind's mark.
	Icon  *ui.SVG
	Width float32
	// Small is width 416 when Width is 0; Large is 736. The methods
	// default to 416 (Ant Modal.confirm).
	Small, Large bool
	// MaskClosable closes it on a click on the mask; Closable adds the ×.
	MaskClosable, Closable bool
	OnOK, OnCancel         func()
	// AfterClose runs once the leave motion has finished.
	AfterClose func()
}

func (f ModalFunc) mark(t *Tokens) (AlertType, ui.Color) {
	switch f.Kind {
	case ModalInfo:
		return Info, t.Primary
	case ModalSuccess:
		return Success, t.SuccessIcon
	case ModalError:
		return Error, t.Error
	}
	return Warning, t.WarningIcon
}

// modalPanel builds the confirm look of f in DialogBase: mark, title,
// content and buttons. It returns 0 for OK, 1 for cancel or the ×, and -1.
func modalPanel(c *ui.Context, shown *bool, v float32, f ModalFunc) int {
	t := Of(c)
	l := L(c)
	answer := -1
	ui.DialogBase(c, shown, func(backdrop, panel ui.Element) {
		backdrop.Background(t.Mask.Alpha(v))
		if !f.MaskClosable && backdrop.Clicked() {
			*shown = true
		}
		zoomIn(panel, v)
		shadowSecondary(panel.Width(dialogWidth(f.Width, f.Small, f.Large, DialogWidthSM)).Background(t.Elevated).Radius(8).Padding(20, 24).Gap(t.SpaceMD))
		panel.Children(func() {
			ui.Row(c).FillWidth().Gap(t.SpaceMD).AlignItems(ui.Start).Children(func() {
				kind, tone := f.mark(t)
				if f.Icon != nil {
					ui.Icon(c, f.Icon).Size(22, 22).TextColor(tone).Shrink(0)
				} else {
					statusIcon(c, kind, 22, tone)
				}
				ui.Column(c).Grow(1).MinWidth(0).Gap(t.SpaceSM).Children(func() {
					ui.Text(c, f.Title).FontWeight(600).FontSize(t.FontSizeLG).TextColor(t.TextHeading).MarginEnd(pick[float32](f.Closable, 24, 0))
					if f.Content != "" {
						ui.Text(c, f.Content).FontSize(t.FontSize).TextColor(t.Text)
					}
					if f.Body != nil {
						f.Body()
					}
				})
			})
			if f.Closable {
				ui.Row(c).Absolute().Top(16).Right(16).Children(func() {
					if closeButton(c, t, 14) {
						answer = 1
					}
				})
			}
			ui.Row(c).FillWidth().Gap(t.SpaceSM).Justify(ui.End).Padding(t.SpaceSM, 0, 0).Children(func() {
				if f.Kind == ModalConfirm && Button(c, pick(f.Cancel == "", l.Cancel, f.Cancel), ButtonProps{MinWidth: 64}) {
					answer = 1
				}
				if Button(c, pick(f.OK == "", l.OK, f.OK), ButtonProps{Type: Primary, MinWidth: 64, Danger: f.OKDanger, AutoFocus: true}) {
					answer = 0
				}
			})
		})
	})
	return answer
}

type modalItem struct {
	key  string
	f    ModalFunc
	open bool
}

// Modals holds the modals opened with Ant's Modal methods, as the api of
// Modal.useModal; build ModalHolder once in the view. Each has its own
// zoom, and later ones stack over earlier ones.
type Modals struct {
	mu   sync.Mutex
	list []*modalItem
	seq  int
	// Wake asks the view to build again when a goroutine opens one.
	Wake func()
}

// Open shows f and returns its key.
func (m *Modals) Open(f ModalFunc) string {
	m.mu.Lock()
	m.seq++
	key := "modal-" + strconv.Itoa(m.seq)
	m.list = append(m.list, &modalItem{key: key, f: f, open: true})
	wake := m.Wake
	m.mu.Unlock()
	if wake != nil {
		wake()
	}
	return key
}

// Confirm, Info, Success, Warning and Error open f as that method.
func (m *Modals) Confirm(f ModalFunc) string { f.Kind = ModalConfirm; return m.Open(f) }
func (m *Modals) Info(f ModalFunc) string    { f.Kind = ModalInfo; return m.Open(f) }
func (m *Modals) Success(f ModalFunc) string { f.Kind = ModalSuccess; return m.Open(f) }
func (m *Modals) Warning(f ModalFunc) string { f.Kind = ModalWarning; return m.Open(f) }
func (m *Modals) Error(f ModalFunc) string   { f.Kind = ModalError; return m.Open(f) }

// Update replaces the config of the modal with key (instance.update).
func (m *Modals) Update(key string, f ModalFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, it := range m.list {
		if it.key == key {
			f.Kind = it.f.Kind
			it.f = f
		}
	}
}

// Close closes the modal with key, or all of them when key is empty
// (Modal.destroyAll), without running OnOK or OnCancel.
func (m *Modals) Close(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, it := range m.list {
		if key == "" || it.key == key {
			it.open = false
		}
	}
}

// Len reports how many modals are open.
func (m *Modals) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, it := range m.list {
		if it.open {
			n++
		}
	}
	return n
}

// ModalHolder builds the modals of m, as Modal.useModal's contextHolder.
func ModalHolder(c *ui.Context, m *Modals) {
	m.mu.Lock()
	list := append([]*modalItem(nil), m.list...)
	m.mu.Unlock()
	var gone []*modalItem
	for _, it := range list {
		var host ui.Element
		ui.Overlay(c, func() {
			host = ui.Box(c.Key(it.key)).Absolute().Left(0).Top(0).Size(0, 0).PassThrough()
		})
		live, v := presence(host, "modal", it.open, motionFor(c, dialogMotion))
		if !live {
			gone = append(gone, it)
			if it.f.AfterClose != nil {
				it.f.AfterClose()
			}
			continue
		}
		shown := true
		answer := modalPanel(c, &shown, v, it.f)
		if !it.open {
			continue
		}
		if answer == -1 && !shown {
			answer = 1
		}
		switch answer {
		case 0:
			it.open = false
			if it.f.OnOK != nil {
				it.f.OnOK()
			}
		case 1:
			it.open = false
			if it.f.OnCancel != nil {
				it.f.OnCancel()
			}
		}
	}
	if len(gone) > 0 {
		m.mu.Lock()
		kept := m.list[:0]
		for _, it := range m.list {
			if indexOfModal(gone, it) < 0 {
				kept = append(kept, it)
			}
		}
		clear(m.list[len(kept):])
		m.list = kept
		m.mu.Unlock()
	}
}

func indexOfModal(list []*modalItem, it *modalItem) int {
	for i, have := range list {
		if have == it {
			return i
		}
	}
	return -1
}
