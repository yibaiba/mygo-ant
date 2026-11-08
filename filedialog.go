package kit

import (
	"strings"
	"sync"

	"github.com/egoist/mygo/ui"
)

// FileDialogOptions are what a file dialog is asked for.
type FileDialogOptions struct {
	Title string
	// Directory picks folders instead of files; Multiple more than one.
	Directory, Multiple bool
	// Accept lists the extensions, without the dot, offered; empty offers
	// any file.
	Accept []string
}

// FileDialog shows the system's file dialog and returns the picked paths,
// none when the user cancels. It may block: the kit calls it off the UI
// goroutine.
type FileDialog func(FileDialogOptions) ([]string, error)

// SaveDialogOptions are what a save dialog is asked for.
type SaveDialogOptions struct {
	Title, DefaultName, DefaultPath string
	Accept                          []string
}

// SaveDialog shows the system's save dialog and returns the path, empty
// when the user cancels. It may block: the kit calls it off the UI
// goroutine.
type SaveDialog func(SaveDialogOptions) (string, error)

var (
	fileDialogMu sync.RWMutex
	fileDialogFn FileDialog
	saveDialogFn SaveDialog
)

// SetFileDialog sets the OS file dialog Upload, UploadList and PathPicker
// open on a click. That is not an Ant explorer clone; Upload is the Ant
// "选择文件" counterpart (dashed drop zone + list). The kit only depends
// on ui, so the app passes the platform's (mygo's Dialog.Open). Nil goes
// back to reporting the click.
func SetFileDialog(open FileDialog) {
	fileDialogMu.Lock()
	fileDialogFn = open
	fileDialogMu.Unlock()
}

func fileDialog(own FileDialog) FileDialog {
	if own != nil {
		return own
	}
	fileDialogMu.RLock()
	defer fileDialogMu.RUnlock()
	return fileDialogFn
}

// SetSaveDialog sets the save dialog PathPicker-style callers open; the
// kit only depends on ui, so the app passes mygo's Dialog.Save. Nil
// clears it.
func SetSaveDialog(save SaveDialog) {
	fileDialogMu.Lock()
	saveDialogFn = save
	fileDialogMu.Unlock()
}

func saveDialog() SaveDialog {
	fileDialogMu.RLock()
	defer fileDialogMu.RUnlock()
	return saveDialogFn
}

// OpenSaveDialog opens the dialog set with SetSaveDialog on a goroutine of
// its own and calls done there with the path, then asks c for a frame.
func OpenSaveDialog(c *ui.Context, opts SaveDialogOptions, done func(path string)) bool {
	save := saveDialog()
	if save == nil {
		return false
	}
	go func() {
		path, err := save(opts)
		if err != nil || path == "" {
			return
		}
		done(path)
		c.Invalidate()
	}()
	return true
}

// SafeSaveFileName is the basename, with
// characters a save dialog cannot take stripped.
func SafeSaveFileName(name, fallback string) string {
	if fallback == "" {
		fallback = "diagnostic.zip"
	}
	base := name
	if i := strings.LastIndexAny(name, `/\`); i >= 0 {
		base = name[i+1:]
	}
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case '<', '>', ':', '"', '|', '?', '*', 0:
			return -1
		}
		return r
	}, strings.TrimSpace(base))
	if cleaned == "" {
		return fallback
	}
	return cleaned
}

// OpenFileDialog opens the dialog set with SetFileDialog on a goroutine of
// its own and calls done there with what the user picks, then asks c for
// a frame; it does not call done on a cancel or an error. It reports
// false, opening nothing, when no dialog is set.
func OpenFileDialog(c *ui.Context, opts FileDialogOptions, done func(paths []string)) bool {
	return openWith(nil, c, opts, done)
}

func openWith(own FileDialog, c *ui.Context, opts FileDialogOptions, done func([]string)) bool {
	open := fileDialog(own)
	if open == nil {
		return false
	}
	go func() {
		paths, err := open(opts)
		if err != nil || len(paths) == 0 {
			return
		}
		done(paths)
		c.Invalidate()
	}()
	return true
}

// picked holds what a control's dialog returned until the control's next
// build takes it.
type picked struct {
	mu    sync.Mutex
	paths []string
}

func (p *picked) put(paths []string) {
	p.mu.Lock()
	p.paths = paths
	p.mu.Unlock()
}

func (p *picked) take() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := p.paths
	p.paths = nil
	return out
}

// pickOnClick opens a control's dialog on a click: its OnPick takes the
// paths on the dialog's goroutine when set, else hold does for the next
// build. It reports whether a dialog opened.
func pickOnClick(c *ui.Context, own FileDialog, opts FileDialogOptions, onPick func([]string), hold *picked) bool {
	return openWith(own, c, opts, func(paths []string) {
		if onPick != nil {
			onPick(paths)
			return
		}
		hold.put(paths)
	})
}
