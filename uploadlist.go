package kit

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/egoist/mygo/ui"
)

// UploadStatus is the state of one file of an upload list.
type UploadStatus uint8

// Upload statuses.
const (
	UploadDone UploadStatus = iota
	UploadUploading
	UploadError
)

// UploadFile is one file of an upload list, as Ant's UploadFile.
type UploadFile struct {
	// UID names it; its Path when empty.
	UID string
	// Name is shown; the base of Path when empty.
	Name   string
	Path   string
	Status UploadStatus
	// Percent is the progress while Uploading.
	Percent int
	// Error is the tooltip of a failed file.
	Error string
	// Thumb is the picture of the picture list types; image files load
	// their own.
	Thumb ui.ImageSource
}

func (f UploadFile) key() string { return pick(f.UID == "", f.Path, f.UID) }

func (f UploadFile) title() string { return pick(f.Name == "", filepath.Base(f.Path), f.Name) }

// UploadListType is Ant's Upload listType.
type UploadListType uint8

// Upload list types.
const (
	ListText UploadListType = iota
	ListPicture
	ListPictureCard
	ListPictureCircle
)

// UploadListProps are an upload list's options.
type UploadListProps struct {
	ListType UploadListType
	// Accept lists the extensions, without the dot, files must have.
	Accept []string
	// Multiple adds files; otherwise a new one replaces the list.
	Multiple bool
	// MaxCount caps the list: the trigger hides at it, and a new file
	// replaces the last when it is 1 (Ant's maxCount).
	MaxCount int
	// Directory takes dropped folders as the files in them, and tells the
	// caller to open a folder dialog on a click.
	Directory bool
	Disabled  bool
	// Dragger makes the trigger of a text or picture list the dashed drop
	// zone, with Hint under its text.
	Dragger bool
	Hint    string
	// Text labels the trigger; the locale's "上传" by default.
	Text string
	// NoPreview leaves the eye off picture cards; NoRemove the delete.
	NoPreview, NoRemove bool
	// Title is the file dialog's, the trigger's text by default; Pick
	// opens it instead of SetFileDialog's.
	Title string
	Pick  FileDialog
	// Action is the URL kit POSTs each added file to as multipart (Ant
	// action). Empty leaves Status and Percent to the caller, as before.
	Action string
	// Method is the Action request method; POST when empty.
	Method string
	// Filename is the multipart field; "file" when empty (Ant name).
	Filename string
	// Data is extra multipart fields; Headers extra request headers.
	Data    map[string]string
	Headers map[string]string
	// CustomRequest, when set, is used instead of Action: kit still
	// marks the file uploading and writes Percent / Status from the
	// callbacks. Run it off the UI goroutine if it blocks.
	CustomRequest func(UploadRequest)
}

// UploadEvent is what happened to an upload list in a frame.
type UploadEvent struct {
	// Pick is a click on the trigger the kit had no file dialog for (see
	// SetFileDialog): open one (a folder one with Directory) and pass
	// what the user picks to AddUploads.
	Pick bool
	// Added are the files dropped on the list or picked in its dialog and
	// added to it, for the caller to start uploading; Removed those the
	// user removed.
	Added, Removed []UploadFile
	// Preview is a file whose name or eye was clicked; image files also
	// open the list's own preview.
	Preview *UploadFile
}

// AddUploads returns files with the files at paths added as p takes them:
// filtered by Accept, folders expanded with Directory (skipped without),
// duplicates dropped, and capped by Multiple and MaxCount. It also
// returns the files it added.
func AddUploads(files []UploadFile, paths []string, p UploadListProps) ([]UploadFile, []UploadFile) {
	var flat []string
	for _, path := range paths {
		if st, err := os.Stat(path); err == nil && st.IsDir() {
			if p.Directory {
				flat = append(flat, filesUnder(path)...)
			}
			continue
		}
		flat = append(flat, path)
	}
	flat = acceptFiles(flat, p.Accept)
	var added []UploadFile
	for _, path := range flat {
		if slices.ContainsFunc(files, func(f UploadFile) bool { return f.key() == path }) ||
			slices.ContainsFunc(added, func(f UploadFile) bool { return f.Path == path }) {
			continue
		}
		added = append(added, UploadFile{Path: path})
	}
	if len(added) == 0 {
		return files, nil
	}
	if !p.Multiple || p.MaxCount == 1 {
		added = added[len(added)-1:]
		return added, added
	}
	out := append(slices.Clone(files), added...)
	if p.MaxCount > 0 && len(out) > p.MaxCount {
		out = out[:p.MaxCount]
		added = added[:max(0, p.MaxCount-len(files))]
	}
	return out, added
}

// filesUnder lists the files in dir and the folders under it, in order.
func filesUnder(dir string) []string {
	var out []string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			out = append(out, path)
		}
		return nil
	})
	return out
}

var imageExts = []string{"png", "jpg", "jpeg", "gif", "webp", "bmp", "svg"}

func isImagePath(path string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	return slices.Contains(imageExts, ext)
}

var (
	thumbMu    sync.Mutex
	thumbCache = map[string]ui.ImageSource{}
)

// uploadThumb is f's picture: its Thumb, or its image file decoded once.
func uploadThumb(f UploadFile) ui.ImageSource {
	if f.Thumb != nil {
		return f.Thumb
	}
	if f.Path == "" || !isImagePath(f.Path) {
		return nil
	}
	thumbMu.Lock()
	defer thumbMu.Unlock()
	if src, ok := thumbCache[f.Path]; ok {
		return src
	}
	var src ui.ImageSource
	if st, err := os.Stat(f.Path); err == nil && st.Size() < 32<<20 {
		if data, err := os.ReadFile(f.Path); err == nil {
			src, _ = ui.DecodeImage(data)
		}
	}
	thumbCache[f.Path] = src
	return src
}

type uploadHold struct {
	open   bool
	cur    int
	picked *picked
}

// UploadList is Ant's Upload with its file list: a button (or with
// Dragger the drop zone) over a text or picture list, or picture cards
// with the trigger as the last card. Files show their status: a spinner
// and progress while uploading, red with the error as tooltip when
// failed. Files dropped on it are added; the caller opens the file dialog
// on Pick. With Action or CustomRequest, kit POSTs the added files and
// writes Status / Percent; otherwise the caller uploads Added.
func UploadList(c *ui.Context, files *[]UploadFile, props ...UploadListProps) UploadEvent {
	p := first(props)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	l := L(c)
	var ev UploadEvent
	full := p.MaxCount > 1 && len(*files) >= p.MaxCount
	cards := p.ListType == ListPictureCard || p.ListType == ListPictureCircle
	box := ui.Column(c).FillWidth().Gap(t.SpaceSM)
	hold := ui.Local(box, "upload", func() uploadHold { return uploadHold{picked: &picked{}} })
	var dropped []string
	if !p.Disabled && !p.Dragger {
		dropped = box.DroppedFiles()
	}
	remove := -1
	box.Children(func() {
		trigger := func() {
			if full {
				return
			}
			label := pick(p.Text == "", l.Upload, p.Text)
			switch {
			case cards:
				ev.Pick = uploadCardTrigger(c, t, label, p) || ev.Pick
			case p.Dragger:
				var got []string
				ev.Pick, got = uploadZone(c, t, pick(p.Text == "", l.UploadDrop, p.Text), p.Hint, nil, p.Disabled, rest, false)
				dropped = append(dropped, got...)
			default:
				ev.Pick = Button(c, label, ButtonProps{Icon: LookupIcon("upload"), Disabled: p.Disabled})
			}
		}
		if cards {
			ui.Row(c).FillWidth().Wrap().Gap(8).Children(func() {
				for i, f := range *files {
					rm, pv := uploadCard(c.Key(f.key()), t, f, p)
					if rm {
						remove = i
					}
					if pv {
						ev.Preview = &(*files)[i]
					}
				}
				trigger()
			})
			return
		}
		trigger()
		if len(*files) == 0 {
			return
		}
		ui.Column(c).FillWidth().Gap(pick[float32](p.ListType == ListPicture, 8, 0)).Children(func() {
			for i, f := range *files {
				rm, pv := uploadRow(c.Key(f.key()), t, f, p)
				if rm {
					remove = i
				}
				if pv {
					ev.Preview = &(*files)[i]
				}
			}
		})
	})
	if ev.Pick && !full {
		opts := FileDialogOptions{Title: pick(p.Title == "", pick(p.Text == "", l.Upload, p.Text), p.Title),
			Directory: p.Directory, Multiple: p.Multiple && p.MaxCount != 1, Accept: p.Accept}
		ev.Pick = !pickOnClick(c, p.Pick, opts, nil, hold.picked)
	}
	jobs := ui.Local(box, "jobs", func() uploadJobs { return uploadJobs{started: map[string]bool{}} })
	jobs.forgetMissing(*files)
	applyUploadPatches(*files, jobs.take())
	dropped = append(dropped, hold.picked.take()...)
	if len(dropped) > 0 && !full {
		*files, ev.Added = AddUploads(*files, dropped, p)
		if kitUploads(p) {
			for i := range ev.Added {
				ev.Added[i].Status = UploadUploading
			}
			for i := range *files {
				for _, add := range ev.Added {
					if (*files)[i].key() == add.key() {
						(*files)[i].Status = UploadUploading
					}
				}
			}
			startUploads(c, *files, ev.Added, p, jobs)
		}
	}
	if remove >= 0 {
		ev.Removed = []UploadFile{(*files)[remove]}
		*files = slices.Delete(slices.Clone(*files), remove, remove+1)
		ev.Preview = nil
	}
	if ev.Preview != nil && isImagePath(ev.Preview.Path) || ev.Preview != nil && ev.Preview.Thumb != nil {
		hold.open = true
		hold.cur = 0
		for _, f := range *files {
			if f.key() == ev.Preview.key() {
				break
			}
			if uploadThumb(f) != nil {
				hold.cur++
			}
		}
	}
	if hold.open {
		var srcs []ui.ImageSource
		for _, f := range *files {
			if src := uploadThumb(f); src != nil {
				srcs = append(srcs, src)
			}
		}
		ImagePreview(c, &hold.open, srcs, &hold.cur)
	}
	return ev
}

// uploadMark is the leading mark of a file row: a spinner while
// uploading, else the clip or the picture.
func uploadMark(c *ui.Context, t *Tokens, f UploadFile, size float32, picture bool) {
	tone := pick(f.Status == UploadError, t.Error, t.TextTertiary)
	switch {
	case f.Status == UploadUploading:
		IconNamed(c, "loading", IconProps{Size: size, Color: t.TextTertiary, Spin: true})
	case picture:
		ui.Icon(c, LookupIcon("picture")).Size(size, size).TextColor(pick(f.Status == UploadError, t.Error, t.Primary)).Shrink(0)
	default:
		ui.Icon(c, LookupIcon("paper-clip")).Size(size, size).TextColor(tone).Shrink(0)
	}
}

// uploadRow is a file of a text or picture list; it reports a click of
// its delete and of its name.
func uploadRow(c *ui.Context, t *Tokens, f UploadFile, p UploadListProps) (removed, preview bool) {
	l := L(c)
	pic := p.ListType == ListPicture
	row := ui.Column(c).FillWidth().Radius(pick[float32](pic, 8, 4)).Gap(4)
	if pic {
		row.Padding(8).Border(1, pick(f.Status == UploadError, t.Error, t.Border))
	} else {
		row.Padding(0, 4)
	}
	hovered := row.Hovered()
	if hovered && !pic {
		row.Background(t.FillSecondary)
	}
	row.Children(func() {
		ui.Row(c).FillWidth().Height(pick[float32](pic, 48, 22)).Gap(8).AlignItems(ui.Center).Children(func() {
			if pic {
				ui.Box(c).Size(48, 48).Shrink(0).Center().Radius(4).Clip().Children(func() {
					if src := uploadThumb(f); src != nil && f.Status != UploadUploading {
						ui.Image(c, src).Size(48, 48).Fit(ui.Cover)
					} else {
						uploadMark(c, t, f, 26, true)
					}
				})
			} else {
				uploadMark(c, t, f, 14, false)
			}
			name := ui.ButtonBase(c).Label(l.PreviewFile + " " + f.title()).FocusRing(false).Grow(1).MinWidth(0).Justify(ui.Start).Cursor(ui.CursorPointer)
			name.Children(func() {
				fg := pick(f.Status == UploadError, t.Error, t.Text)
				if f.Status == UploadDone && name.Hovered() {
					fg = t.Primary
				}
				ui.Text(c, f.title()).FontSize(t.FontSize).TextColor(fg).SingleLine()
			})
			if f.Status == UploadError && f.Error != "" {
				TooltipOn(c, name, f.Error)
			}
			preview = name.Clicked()
			if !p.NoRemove && !p.Disabled && (hovered || pic || f.Status == UploadError) {
				removed = iconButton(c, LookupIcon("delete"), l.Remove+" "+f.title(), pick(f.Status == UploadError, t.Error, t.TextTertiary), 14)
			}
		})
		if f.Status == UploadUploading {
			ui.Box(c).FillWidth().Padding(0, 0, 0, pick[float32](pic, 56, 22)).Children(func() {
				Progress(c, f.Percent, ProgressProps{HideInfo: true, StrokeWidth: 2, Color: t.Primary, Status: ProgressActive})
			})
		}
	})
	return removed, preview
}

// uploadCardSize is the side of a picture card (Ant 102 inside an 8 gap).
const uploadCardSize = 102

// uploadCard is a file of a picture-card or picture-circle list: its
// picture with the eye and delete over it on hover, "上传中" and a bar
// while uploading, red when failed. It reports a click of delete and eye.
func uploadCard(c *ui.Context, t *Tokens, f UploadFile, p UploadListProps) (removed, preview bool) {
	l := L(c)
	round := p.ListType == ListPictureCircle
	radius := pick[float32](round, uploadCardSize/2, 8)
	card := ui.Column(c).Size(uploadCardSize, uploadCardSize).Shrink(0).Padding(8).Radius(radius).
		Border(1, pick(f.Status == UploadError, t.Error, t.Border)).Center()
	hovered := card.Hovered()
	card.Children(func() {
		inner := ui.Column(c).Fill().Radius(pick[float32](round, (uploadCardSize-16)/2, 4)).Clip().Center().Gap(6)
		inner.Children(func() {
			src := uploadThumb(f)
			switch {
			case f.Status == UploadUploading:
				ui.Text(c, l.Uploading).FontSize(t.FontSize).TextColor(t.TextTertiary)
				ui.Box(c).Width(66).Children(func() {
					Progress(c, f.Percent, ProgressProps{HideInfo: true, StrokeWidth: 2, Color: t.Primary})
				})
			case src != nil:
				ui.Image(c, src).Fill().Fit(ui.Cover)
			default:
				uploadMark(c, t, f, 26, true)
				ui.Text(c, f.title()).FontSize(t.FontSizeSM).TextColor(pick(f.Status == UploadError, t.Error, t.TextTertiary)).SingleLine().MaxWidth(70)
			}
			if !hovered || f.Status == UploadUploading {
				return
			}
			ui.Row(c).Absolute().Left(0).Top(0).Right(0).Bottom(0).Background(t.Mask).Center().Gap(4).Children(func() {
				if !p.NoPreview {
					preview = iconButton(c, LookupIcon("eye"), l.Preview+" "+f.title(), ui.RGBA(255, 255, 255, 0.85), 16)
				}
				if !p.NoRemove && !p.Disabled {
					removed = iconButton(c, LookupIcon("delete"), l.Remove+" "+f.title(), ui.RGBA(255, 255, 255, 0.85), 16)
				}
			})
		})
		if f.Status == UploadError && f.Error != "" {
			TooltipOn(c, inner, f.Error)
		}
	})
	return removed, preview
}

// uploadCardTrigger is the dashed "+ 上传" card ending a picture-card list.
func uploadCardTrigger(c *ui.Context, t *Tokens, label string, p UploadListProps) bool {
	round := p.ListType == ListPictureCircle
	b := ui.ButtonBase(c).Label(label).Disabled(p.Disabled).FocusRing(false).Size(uploadCardSize, uploadCardSize).Shrink(0).
		Column().Center().Gap(8).Radius(pick[float32](round, uploadCardSize/2, 8)).
		Border(1, t.Border).BorderStyle(ui.BorderDashed).Background(t.FillQuaternary).
		Cursor(pick(p.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
	st := stateOf(b, p.Disabled, rest)
	switch st {
	case hover, active:
		b.BorderColor(t.Primary)
	case focus:
		focusRing(b, t)
	case disabled:
		b.Background(t.DisabledBg)
	}
	fg := pick(st == disabled, t.TextDisabled, t.Text)
	b.Children(func() {
		ui.Icon(c, LookupIcon("plus")).Size(16, 16).TextColor(fg)
		ui.Text(c, label).FontSize(t.FontSize).TextColor(fg)
	})
	return b.Clicked() && !p.Disabled
}
