package kit

import (
	"strconv"

	"github.com/egoist/mygo/ui"
)

// FormLayout is Ant's Form layout.
type FormLayout uint8

// Form layouts.
const (
	// FormVertical puts labels above the controls: the default forms.
	FormVertical FormLayout = iota
	// FormHorizontal puts them in a right-aligned column before the
	// controls, LabelCol and WrapperCol of 24 wide.
	FormHorizontal
	// FormInline lines items up in a wrapping row.
	FormInline
)

// RequiredMark is how labels mark required items (Ant requiredMark).
type RequiredMark uint8

// Required marks.
const (
	// MarkRequired puts a red * before required labels.
	MarkRequired RequiredMark = iota
	// MarkOptional puts "（可选）" after the labels of the others.
	MarkOptional
	// MarkHidden marks neither.
	MarkHidden
)

// FormConfig is the layout a FormView gives its FormItems.
type FormConfig struct {
	Layout FormLayout
	// LabelCol and WrapperCol split a horizontal item's 24 columns; 6 and
	// 18 when zero. LabelWidth fixes the label column instead.
	LabelCol, WrapperCol int
	LabelWidth           float32
	// LabelLeft aligns horizontal labels left (labelAlign left).
	LabelLeft bool
	// NoColon drops the colon after horizontal and inline labels.
	NoColon      bool
	RequiredMark RequiredMark
	// LabelWrap lets long labels wrap instead of cutting them short.
	LabelWrap bool
}

// FormProps are a form view's options.
type FormProps struct {
	FormConfig
	// Size and Disabled reach every control in it, as Ant Form's size and
	// disabled.
	Size     Size
	Disabled bool
	// Gap is the room between items; 16 when zero (8 across inline).
	Gap float32
}

// FormView is Ant's Form container: it lays the FormItems fn builds out
// as Layout says and passes its size and disabled state to the controls
// in them, through the same scope as ConfigProvider.
func FormView(c *ui.Context, fn func(), props ...FormProps) ui.Element {
	p := first(props)
	outer := Of(c)
	size := p.Size
	if size == SizeInherit {
		size = outer.Size
	}
	t := Configured(outer, ConfigProps{
		Size:     size,
		Disabled: Bool(p.Disabled || outer.Disabled),
		Locale:   outer.Locale,
		Form:     p.FormConfig,
		Dark:     Bool(outer.Dark),
		Motion:   Bool(!outer.MotionOff),
	})
	gap := pick(p.Gap == 0, t.SpaceLG, p.Gap)
	box := ui.Themed(c, t.Theme()).Background(ui.Color{}).FillWidth().MinWidth(0)
	if p.Layout == FormInline {
		box.Row().Wrap().GapX(gap).GapY(t.SpaceSM).AlignItems(ui.Start)
	} else {
		box.Gap(gap)
	}
	return box.Children(fn)
}

// FormItemProps are a form item's options.
type FormItemProps struct {
	Size Size
	// Help shows under the control in muted text, Warning in orange or
	// Error in red instead; Extra stays under them.
	Help, Warning, Error, Extra string
	Required                    bool
	// Tooltip adds a question mark after the label with it.
	Tooltip string
	// Min makes the item a cell of a FormRow: it grows from Min wide and
	// wraps to the next line below it. Max caps it. Zero fills the width.
	Min, Max float32
	// LabelCol and WrapperCol override the FormView's for this item;
	// NoColon drops its colon.
	LabelCol, WrapperCol int
	NoColon              bool
	// Name is the field path, as Ant Form.Item name: a string ("user"
	// or "user.name") or []string{"user","name"}. FormItem only lays
	// out; pass the same name to Form.Check.
	Name any
	// Form, with Name, is the validator Watch reports focus to, so
	// ValidateBlur can wait until the control is left.
	Form *Form
}

// FormItem is an Ant form item: label, the control fn builds, then help,
// warning or error text and Extra; vertical unless a FormView lays it out
// otherwise. Error and Help go through Field.Error / Field.Description so
// assistive technology finds them; they are not painted again as Text.
// Extra and Warning stay painted under the control.
func FormItem(c *ui.Context, label string, fn func(), props ...FormItemProps) {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	t := Of(c)
	cfg := t.Form
	colon := cfg.Layout != FormVertical && !cfg.NoColon && !p.NoColon
	control := func() {
		if p.Form == nil || NameKey(p.Name) == "" {
			fn()
			return
		}
		wrap := ui.Column(c).FillWidth().MinWidth(0)
		wrap.Children(fn)
		p.Form.Watch(p.Name, wrap.FocusWithin())
	}
	var hit bool
	named := func() {
		needField := label != "" || p.Error != "" || p.Help != ""
		if needField {
			field := bindField(c, t, label, control, hit)
			if p.Error != "" {
				field.Error(p.Error)
			}
			if p.Help != "" {
				field.Description(p.Help)
			}
		} else {
			control()
		}
		if p.Warning != "" {
			ui.Text(c, p.Warning).FontSize(t.FontSize).TextColor(t.WarningIcon)
		}
		if p.Extra != "" {
			ui.Text(c, p.Extra).FontSize(t.FontSize).TextColor(t.TextTertiary)
		}
	}
	switch cfg.Layout {
	case FormHorizontal:
		lc := pick(p.LabelCol == 0, cfg.LabelCol, p.LabelCol)
		wc := pick(p.WrapperCol == 0, cfg.WrapperCol, p.WrapperCol)
		lc, wc = pick(lc == 0, 6, lc), pick(wc == 0, 18, wc)
		row := ui.Row(c).Gap(t.SpaceSM).AlignItems(ui.Start)
		if p.Min > 0 {
			row.MinWidth(p.Min).Grow(1).Basis(p.Min)
		} else {
			row.FillWidth().MinWidth(0)
		}
		if p.Max > 0 {
			row.MaxWidth(p.Max)
		}
		row.Children(func() {
			cell := ui.Row(c).MinHeight(t.Height(p.Size)).AlignItems(ui.Center).Justify(pick(cfg.LabelLeft, ui.Start, ui.End)).MinWidth(0).Shrink(0).Clip()
			if cfg.LabelWidth > 0 {
				cell.Width(cfg.LabelWidth).Shrink(0)
			} else {
				cell.Basis(0).Grow(float32(lc))
			}
			cell.Children(func() {
				if label != "" {
					hit = formLabel(c, t, label, p, cfg, colon)
				}
			})
			ui.Column(c).Basis(0).Grow(float32(wc)).MinWidth(0).Gap(4).Children(named)
		})
	case FormInline:
		ui.Row(c).Shrink(0).Gap(t.SpaceSM).AlignItems(ui.Start).Children(func() {
			if label != "" {
				ui.Row(c).MinHeight(t.Height(p.Size)).AlignItems(ui.Center).Shrink(0).Children(func() {
					hit = formLabel(c, t, label, p, cfg, colon)
				})
			}
			ui.Column(c).Gap(4).Children(named)
		})
	default:
		col := ui.Column(c).Gap(t.SpaceSM)
		if p.Min > 0 {
			col.MinWidth(p.Min).Grow(1).Basis(p.Min)
		} else {
			col.FillWidth().MinWidth(0)
		}
		if p.Max > 0 {
			col.MaxWidth(p.Max)
		}
		col.Children(func() {
			if label != "" {
				hit = formLabel(c, t, label, p, cfg, colon)
			}
			named()
		})
	}
}

// labelFocus is how many FormItem label clicks this frame still need a
// control to take the keyboard focus (UI is single-threaded).
var labelFocus int

// takeLabelFocus gives e the keyboard focus when a FormItem label was
// clicked this frame, so typing goes into the control Field names.
func takeLabelFocus(e ui.Element) {
	if labelFocus > 0 && e.Valid() {
		labelFocus--
		e.Focus()
	}
}

// bindField wraps fn in a MyGo Field so the label names the control for
// assistive technology and a click on Field's label focuses it. The
// Field's own visual label is collapsed (theme FontSize ~0) so FormItem's
// layout — label width, required *, extra/help — stays as it was. When
// focus is set (the painted label was clicked), the first control that
// calls takeLabelFocus receives the keyboard focus.
func bindField(c *ui.Context, t *Tokens, label string, fn func(), focus bool) ui.Element {
	if focus {
		labelFocus++
	}
	th := new(ui.Theme)
	*th = *t.Theme()
	th.FontSize = 0.01
	var field ui.Element
	ui.Themed(c, th).Background(ui.Color{}).FillWidth().MinWidth(0).Gap(0).Children(func() {
		field = ui.Field(c, label, func() {
			ui.Themed(c, t.Theme()).Background(ui.Color{}).FillWidth().MinWidth(0).Children(fn)
		}).Gap(0).FillWidth().MinWidth(0)
	})
	return field
}

// formLabel is an item's label with its required or optional mark, its
// tooltip and colon. It reports a click so FormItem can focus the control.
func formLabel(c *ui.Context, t *Tokens, label string, p FormItemProps, cfg FormConfig, colon bool) bool {
	size := t.Label(p.Size)
	hit := false
	ui.Row(c).Gap(4).AlignItems(ui.Center).MinWidth(0).Children(func() {
		if p.Required && cfg.RequiredMark == MarkRequired {
			ui.Text(c, "*").FontSize(size).TextColor(t.Error).Shrink(0)
		}
		txt := ui.Text(c, label).FontSize(size).TextColor(t.Text).MinWidth(0)
		if !cfg.LabelWrap {
			txt.SingleLine()
		}
		hit = txt.Clicked()
		if !p.Required && cfg.RequiredMark == MarkOptional {
			ui.Text(c, L(c).Optional).FontSize(size).TextColor(t.TextTertiary).SingleLine().Shrink(0)
		}
		if p.Tooltip != "" {
			Tooltip(c, p.Tooltip, func() {
				ui.Icon(c, LookupIcon("question-circle")).Size(14, 14).TextColor(t.TextTertiary)
			})
		}
		if colon {
			ui.Text(c, ":").FontSize(size).TextColor(t.Text).Shrink(0)
		}
	})
	return hit
}

// FormListProps are a form list's options.
type FormListProps struct {
	// Add labels the dashed add button; "添加" by default.
	Add string
	// Min keeps that many rows: their removes do nothing. Max hides the
	// add button at that many. NoAdd hides it always.
	Min, Max int
	NoAdd    bool
}

// FormList is Ant's Form.List over *items: row builds item i, keyed so
// that its controls keep their state as rows come and go, and gets the
// remove for it (FormListRemove draws Ant's minus mark); the dashed
// button under the rows appends fresh(). It reports a change.
func FormList[T any](c *ui.Context, items *[]T, fresh func() T, row func(i int, item *T, remove func()), props ...FormListProps) bool {
	p := first(props)
	t := Of(c)
	box := ui.Column(c).FillWidth().Gap(t.SpaceSM)
	keys := ui.Local(box, "keys", func() formListKeys { return formListKeys{} })
	keys.sync(len(*items))
	removeAt, add := -1, false
	box.Children(func() {
		for i := range *items {
			ui.Column(c.Key(keys.ids[i])).FillWidth().Children(func() {
				row(i, &(*items)[i], func() {
					if len(*items) > p.Min {
						removeAt = i
					}
				})
			})
		}
		if !p.NoAdd && (p.Max == 0 || len(*items) < p.Max) {
			add = Button(c, pick(p.Add == "", L(c).Add, p.Add), ButtonProps{Type: Dashed, Block: true, Icon: LookupIcon("plus")})
		}
	})
	switch {
	case removeAt >= 0:
		*items = append((*items)[:removeAt:removeAt], (*items)[removeAt+1:]...)
		keys.ids = append(keys.ids[:removeAt:removeAt], keys.ids[removeAt+1:]...)
		return true
	case add:
		*items = append(*items, fresh())
		keys.sync(len(*items))
		return true
	}
	return false
}

// formListKeys are the stable keys of a form list's rows.
type formListKeys struct {
	ids  []string
	next int
}

func (k *formListKeys) sync(n int) {
	for len(k.ids) < n {
		k.next++
		k.ids = append(k.ids, "row-"+strconv.Itoa(k.next))
	}
	k.ids = k.ids[:n]
}

// FormListRemove is the minus mark at the end of a Form.List row; it
// reports a click.
func FormListRemove(c *ui.Context, label string) bool {
	t := Of(c)
	if t.Disabled {
		ui.Icon(c, LookupIcon("minus-circle")).Size(16, 16).TextColor(t.TextDisabled).Shrink(0)
		return false
	}
	return iconButton(c, LookupIcon("minus-circle"), pick(label == "", L(c).Remove, label), t.TextTertiary, 16)
}

// FormRowProps are a form row's options.
type FormRowProps struct {
	Gap float32
	// Center aligns cells on their middles instead of their bottoms.
	Center bool
}

// FormRow lays FormItems with Min side by side, wrapping them onto new
// lines when the row is narrower than their Mins, bottoms aligned so that
// buttons sit level with the inputs beside them.
func FormRow(c *ui.Context, fn func(), props ...FormRowProps) {
	p := first(props)
	t := Of(c)
	gap := p.Gap
	if gap == 0 {
		gap = t.SpaceLG
	}
	ui.Row(c).FillWidth().MinWidth(0).Wrap().GapX(gap).GapY(t.SpaceMD).
		AlignItems(pick(p.Center, ui.Center, ui.End)).Children(fn)
}

// Slot is a control-high row centering what fn builds, so a button, tag,
// check box or hint lines up with the inputs of its FormRow. Buttons and
// tags hold AlignSelf(Start) so they do not stretch across columns, which
// in a row would lift them to the top: put them in a Slot there.
func Slot(c *ui.Context, size Size, fn func()) {
	t := Of(c)
	ui.Row(c).Height(t.Height(size)).Gap(t.SpaceSM).AlignItems(ui.Center).Shrink(0).Children(fn)
}

// PathPickerProps are a path picker's options.
type PathPickerProps struct {
	Size        Size
	Placeholder string
	Disabled    bool
	Status      Status
	// Dialog is what the button's file dialog asks for, the button's text
	// its title by default; Pick opens it instead of SetFileDialog's.
	Dialog FileDialogOptions
	Pick   FileDialog
	// OnPick takes the picked path on the dialog's goroutine; without it
	// the next frame writes it to *value.
	OnPick func(path string)
	// Primary paints the picker as a primary button, as BackupSetting 浏览.
	Primary bool
}

// PathPicker is an input with a picker button after it, the default
// file field (PathSetting.scss .select-file): it edits *value and opens
// the OS file dialog on a click of the button. That is not a fake
// Explorer; Upload is the Ant "选择文件" counterpart. It reports a click
// the kit had no dialog for (see SetFileDialog), or the frame a picked
// path is written to *value.
func PathPicker(c *ui.Context, value *string, button string, props ...PathPickerProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	out := false
	host := ui.Column(c).FillWidth().MinWidth(0)
	hold := ui.Local(host, "pick", func() *picked { return &picked{} })
	if got := (*hold).take(); len(got) > 0 {
		*value, out = got[0], true
	}
	host.Children(func() {
		SpaceCompact(c, func() {
			Input(c, value, InputProps{Size: p.Size, Placeholder: p.Placeholder, Disabled: p.Disabled, Status: p.Status})
			clicked := false
			if p.Primary {
				clicked = Button(c, button, ButtonProps{Type: Primary, Size: p.Size, Disabled: p.Disabled})
			} else {
				clicked = pickerButton(c, t, button, p)
			}
			if !clicked {
				return
			}
			opts := p.Dialog
			opts.Title = pick(opts.Title == "", button, opts.Title)
			opts.Multiple = false
			var onPick func([]string)
			if p.OnPick != nil {
				onPick = func(paths []string) { p.OnPick(paths[0]) }
			}
			if !pickOnClick(c, p.Pick, opts, onPick, *hold) {
				out = true
			}
		}, CompactProps{Block: true})
	})
	return out
}

// pickerButton is the default grey gradient "选择" button.
func pickerButton(c *ui.Context, t *Tokens, label string, p PathPickerProps) bool {
	slot := compactBegin()
	defer slot.done()
	btn := ui.ButtonBase(c).Label(label).Disabled(p.Disabled).FocusRing(false).
		Height(t.Height(p.Size)).Padding(0, t.PaddingButton).Radius(t.RadiusControl).Shrink(0).
		MinWidth(MinButtonWidth(label, p.Size)).Cursor(ui.CursorPointer)
	slot.at(btn)
	st := stateOf(btn, p.Disabled, rest)
	top, bottom, edge, fg := t.PickerTop, t.PickerBottom, t.BorderButton, t.Text
	switch st {
	case hover, focus:
		edge, fg = t.PrimaryHover, t.PrimaryHover
	case active:
		bottom, edge, fg = t.PickerPressed, t.PrimaryActive, t.PrimaryActive
	case disabled:
		top, bottom, fg = t.DisabledBg, t.DisabledBg, t.TextDisabled
	}
	btn.LinearGradient(ui.LinearGradient{From: top, To: bottom, Angle: 180, Start: 0.25, End: 1}).Border(1, edge)
	slot.part(btn, t.RadiusControl, st == hover || st == focus || st == active, edge)
	if st == focus {
		focusRing(btn, t)
	}
	btn.Children(func() {
		ui.Text(c, label).FontSize(t.FontSize).TextColor(fg).SingleLine()
	})
	return btn.Clicked() && !p.Disabled
}

// Muted is secondary copy: hints under sections, meta lines.
func Muted(c *ui.Context, text string) ui.Element {
	t := Of(c)
	return ui.Text(c, text).FontSize(13).TextColor(t.TextMuted)
}
