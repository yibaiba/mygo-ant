package kit

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
)

var iconFile = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none"><path d="M10.5 4.5 5.2 9.8a1.3 1.3 0 0 0 1.8 1.8l5.6-5.6a2.6 2.6 0 0 0-3.7-3.7L3.3 7.9a3.9 3.9 0 0 0 5.5 5.5l4.7-4.7" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></svg>`))

// UploadProps are an upload area's options.
type UploadProps struct {
	// Text is the zone's main line, Hint the muted one under it.
	Text, Hint string
	// Accept lists the extensions, without the dot, that dropped files
	// must have; empty takes any.
	Accept []string
	// Multiple adds files to the list; otherwise a new one replaces it.
	Multiple bool
	Disabled bool
	// Title is the file dialog's, the zone's text by default; Pick opens
	// it instead of SetFileDialog's.
	Title string
	Pick  FileDialog
	// OnPick takes the picked files on the dialog's goroutine; without it
	// the next frame adds them to *files as dropped ones.
	OnPick func(paths []string)

	force    visual
	dragOver bool
}

// Upload is Ant's Upload.Dragger over a list of paths: a dashed zone that
// takes files dragged onto it or picked in the OS file dialog a click
// opens, then the files with remove buttons. It is the Ant counterpart of
// PathPicker / SetFileDialog; kit does not draw a fake Explorer. It
// returns true for a click the kit had no dialog for (see SetFileDialog),
// for the caller to open its own.
func Upload(c *ui.Context, files *[]string, props ...UploadProps) bool {
	p := first(props)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	text := pick(p.Text == "", L(c).UploadDrop, p.Text)
	clicked := false
	box := ui.Column(c).FillWidth().Gap(t.SpaceSM)
	hold := ui.Local(box, "pick", func() *picked { return &picked{} })
	box.Children(func() {
		var dropped []string
		clicked, dropped = uploadZone(c, t, text, p.Hint, p.Accept, p.Disabled, p.force, p.dragOver)
		if clicked {
			opts := FileDialogOptions{Title: pick(p.Title == "", text, p.Title), Multiple: p.Multiple, Accept: p.Accept}
			clicked = !pickOnClick(c, p.Pick, opts, p.OnPick, *hold)
		}
		dropped = append(dropped, acceptFiles((*hold).take(), p.Accept)...)
		if len(dropped) > 0 {
			*files = mergeFiles(*files, dropped, p.Multiple)
		}
		remove := -1
		for i, path := range *files {
			if fileRow(c.Key(path), t, path, p.Disabled) {
				remove = i
			}
		}
		if remove >= 0 {
			*files = slices.Delete(slices.Clone(*files), remove, remove+1)
		}
	})
	return clicked
}

// uploadZone is the dashed drop zone of Upload.Dragger: it reports a
// click and the files dropped on it that accept takes.
func uploadZone(c *ui.Context, t *Tokens, text, hint string, accept []string, off bool, force visual, dragOver bool) (bool, []string) {
	zone := ui.ButtonBase(c).Label(text).Disabled(off).FocusRing(false).FillWidth().
		Column().Center().Gap(t.SpaceXS).Padding(t.SpaceLG).Radius(t.RadiusPopup).
		Border(1, t.Border).BorderStyle(ui.BorderDashed).
		Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
	var dropped []string
	if !off {
		dropped = acceptFiles(zone.DroppedFiles(), accept)
	}
	st := stateOf(zone, off, force)
	over := (zone.FileDragOver() || dragOver) && !off
	bg := t.FillAlter
	switch {
	case st == disabled:
		bg = t.DisabledBg
	case over:
		bg = t.PrimaryBg
		zone.BorderColor(t.Primary)
	case st == hover || st == active || st == focus:
		zone.BorderColor(t.PrimaryHover)
	}
	zone.Background(bg)
	if st == focus {
		focusRing(zone, t)
	}
	zone.Children(func() {
		ui.Icon(c, iconEmpty).Size(48, 34).Margin(0, 0, t.SpaceSM).
			TextColor(pick(st == disabled, t.TextDisabled, t.Primary))
		ui.Text(c, pick(over, L(c).UploadRelease, text)).FontSize(t.FontSizeLG).
			TextColor(pick(st == disabled, t.TextDisabled, t.Text))
		if hint != "" {
			ui.Text(c, hint).FontSize(t.FontSize).TextColor(t.TextTertiary).TextAlign(ui.Center)
		}
	})
	return zone.Clicked() && !off, dropped
}

func fileRow(c *ui.Context, t *Tokens, path string, off bool) bool {
	removed := false
	row := ui.Row(c).FillWidth().Height(28).Padding(0, t.SpaceXS).Gap(t.SpaceSM).Radius(4).AlignItems(ui.Center)
	if row.Hovered() {
		row.Background(t.FillTertiary)
	}
	row.Children(func() {
		ui.Icon(c, iconFile).Size(14, 14).Shrink(0).TextColor(t.TextTertiary)
		name := ui.Text(c, filepath.Base(path)).FontSize(t.FontSize).TextColor(t.Text).SingleLine().Grow(1).MinWidth(0)
		TooltipOn(c, name, path)
		x := ui.ButtonBase(c).Label(L(c).Remove+" "+filepath.Base(path)).Disabled(off).Size(22, 22).Radius(4).
			Center().FocusRing(false).Shrink(0).Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
		fg := t.TextTertiary
		switch stateOf(x, off, rest) {
		case hover, active:
			x.Background(t.FillSecondary)
			fg = t.Error
		case focus:
			focusRing(x, t)
		case disabled:
			fg = t.TextDisabled
		}
		x.Children(func() { ui.Icon(c, iconClose).Size(12, 12).TextColor(fg) })
		removed = x.Clicked() && !off
	})
	return removed
}

func acceptFiles(paths, accept []string) []string {
	if len(accept) == 0 {
		return paths
	}
	var out []string
	for _, path := range paths {
		ext := strings.TrimPrefix(filepath.Ext(path), ".")
		if slices.ContainsFunc(accept, func(a string) bool { return strings.EqualFold(a, ext) }) {
			out = append(out, path)
		}
	}
	return out
}

func mergeFiles(have, add []string, multiple bool) []string {
	if !multiple {
		return []string{add[0]}
	}
	out := slices.Clone(have)
	for _, path := range add {
		if !slices.Contains(out, path) {
			out = append(out, path)
		}
	}
	return out
}
