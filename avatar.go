package kit

import (
	"strconv"

	"github.com/egoist/mygo/ui"
)

// AvatarProps are an avatar's options. Src shows an image, else Icon, else
// the text.
type AvatarProps struct {
	Src  ui.ImageSource
	Icon *ui.SVG
	// Size is Ant's preset: Middle 32, Large 40, Small 24; Pixels, when
	// set, overrides it.
	Size   Size
	Pixels float32
	// Square rounds the corners instead of making a circle.
	Square bool
	// Color fills the avatar; TextColor tints its text and icon (white).
	Color, TextColor ui.Color
	// Gap is the room kept on each side of the text, which shrinks to fit;
	// 4 when zero.
	Gap float32

	ring bool
}

func avatarSize(t *Tokens, p AvatarProps) float32 {
	if p.Pixels > 0 {
		return p.Pixels
	}
	switch t.SizeOf(p.Size) {
	case Large:
		return 40
	case Small:
		return 24
	}
	return 32
}

// avatarTextScale is how much text w wide shrinks to fit an avatar size
// across with gap on each side, as Ant scales it: never grown.
func avatarTextScale(w, size, gap float32) float32 {
	room := size - gap*2
	if w <= 0 || w <= room {
		return 1
	}
	return max(room, 1) / w
}

// Avatar is Ant's Avatar: a circle (or rounded square) showing an
// image, an icon, or text shrunk to fit.
func Avatar(c *ui.Context, text string, props ...AvatarProps) ui.Element {
	p := first(props)
	t := Of(c)
	size := avatarSize(t, p)
	bg := p.Color
	if bg == (ui.Color{}) {
		bg = t.AvatarBg
	}
	fg := p.TextColor
	if fg == (ui.Color{}) {
		fg = ui.Hex("#FFFFFF")
	}
	radius := size / 2
	if p.Square {
		radius = map[Size]float32{Middle: 6, Large: 8, Small: 4}[t.SizeOf(p.Size)]
	}
	box := ui.Row(c).Size(size, size).Shrink(0).Radius(radius).Clip().AlignItems(ui.Center).Justify(ui.Center)
	if p.Src == nil {
		box.Background(bg)
	}
	if p.ring {
		box.Border(1, t.BadgeShadow)
	}
	box.Children(func() {
		switch {
		case p.Src != nil:
			ui.Image(c, p.Src).Size(size, size).Fit(ui.Cover)
		case p.Icon != nil:
			is := map[Size]float32{Middle: 18, Large: 24, Small: 14}[t.SizeOf(p.Size)]
			if p.Pixels > 0 {
				is = p.Pixels / 2
			}
			ui.Icon(c, p.Icon).Size(is, is).TextColor(fg)
		default:
			gap := pick(p.Gap == 0, 4, p.Gap)
			w, _ := c.MeasureText(0, ui.Span{Text: text, Size: t.FontSize})
			s := avatarTextScale(w, size, gap)
			ui.Text(c, text).FontSize(t.FontSize * s).TextColor(fg).SingleLine()
		}
	})
	return box
}

// AvatarItem is one avatar of a group.
type AvatarItem struct {
	Text string
	AvatarProps
}

// AvatarGroupProps are an avatar group's options.
type AvatarGroupProps struct {
	// MaxCount shows that many and a "+N" avatar for the rest, which
	// lists them in a popover on hover; 0 shows all.
	MaxCount int
	// Size, when set, sizes every avatar of the group.
	Size Size
	// MoreColor fills the "+N" avatar.
	MoreColor, MoreTextColor ui.Color
	Trigger                  Trigger
}

// AvatarGroup is Ant's Avatar.Group: avatars overlapping by 8, each
// ringed in the container color, folding the ones past MaxCount into a
// "+N" avatar.
func AvatarGroup(c *ui.Context, items []AvatarItem, props ...AvatarGroupProps) {
	p := first(props)
	shown, rest := splitAvatars(items, p.MaxCount)
	one := func(it AvatarItem, i int) ui.Element {
		ap := it.AvatarProps
		if p.Size != SizeInherit {
			ap.Size = p.Size
		}
		ap.ring = true
		e := Avatar(c, it.Text, ap)
		if i > 0 {
			e.MarginStart(-8)
		}
		return e
	}
	ui.Row(c).Shrink(0).AlignItems(ui.Center).AlignSelf(ui.Start).Children(func() {
		for i, it := range shown {
			one(it, i)
		}
		if len(rest) == 0 {
			return
		}
		more := AvatarItem{Text: "+" + strconv.Itoa(len(rest)), AvatarProps: AvatarProps{Color: p.MoreColor, TextColor: p.MoreTextColor}}
		trig := p.Trigger
		if trig == TriggerClick {
			trig = TriggerHover
		}
		ui.Row(c).MarginStart(-8).Children(func() {
			Popover(c, func() {
				ui.Row(c).Children(func() {
					for i, it := range rest {
						one(it, i)
					}
				})
			}, func() bool {
				return one(more, 0).Clicked()
			}, PopoverProps{PopupProps: PopupProps{Trigger: trig, Placement: PlaceTop}})
		})
	})
}

// splitAvatars is the avatars a group shows and those it folds.
func splitAvatars(items []AvatarItem, maxCount int) (shown, rest []AvatarItem) {
	if maxCount <= 0 || len(items) <= maxCount {
		return items, nil
	}
	return items[:maxCount], items[maxCount:]
}
