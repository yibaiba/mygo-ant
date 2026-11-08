package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Sheet is the white form panel of the work pages, inset from the page
// edges as the settings content (settings layout: content padding 0 24).
func Sheet(c *ui.Context, fn func()) {
	t := Of(c)
	ui.Column(c).FillWidth().MinWidth(0).Padding(0, t.SpaceXL, t.SpaceLG).Children(func() {
		ui.Column(c).FillWidth().MinWidth(0).Clip().Gap(t.SpaceXL).Padding(t.SpaceXL, 24, 24).
			Background(t.Surface).Border(1, t.SheetBorder).Radius(t.RadiusSheet).
			Children(fn)
	})
}

// CardProps are a card's options.
type CardProps struct {
	Title string
	// Extra builds actions at the end of the title row.
	Extra func()
	// Cover builds a picture across the top edge, above the title.
	Cover func()
	// Actions are the bottom bar's items, split by rules.
	Actions []func()
	// Tabs put a tab bar under the title choosing *Tab; fn builds the
	// chosen pane. TabExtra builds content at the bar's end.
	Tabs     []Option
	Tab      *string
	TabExtra func()
	// Loading shows a skeleton instead of what fn builds.
	Loading bool
	// Hoverable lifts the card on hover.
	Hoverable bool
	// Size Small tightens the padding to 12 and the title to 14.
	Size Size
	// Borderless drops the edge (Ant bordered={false}).
	Borderless bool
	// Width fixes the card's width; it fills its column by default.
	Width float32
	// Radius overrides RadiusCard; 0 keeps the token.
	Radius float32
	// Raised is the maintenance card's light shadow (MaintenanceCheck .card).
	Raised bool

	hover bool
}

// Card is the remote page's panel (RemoteConnect .panel): white, a
// #E5EAF1 edge, 10px corners, and a 15px title; with Cover, Tabs or
// Actions it splits into Ant's cover, head, body and action bar.
func Card(c *ui.Context, fn func(), props ...CardProps) ui.Element {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	t := Of(c)
	pad := pick(p.Size == Small, t.SpaceMD, t.SpaceLG)
	titleSize := pick[float32](p.Size == Small, t.FontSize, 15)
	radius := pick(p.Radius == 0, t.RadiusCard, p.Radius)
	card := ui.Column(c).Background(t.Surface).Radius(radius).MinWidth(0).Clip()
	if p.Width > 0 {
		card.Width(p.Width).Shrink(0)
	} else {
		card.FillWidth()
	}
	if !p.Borderless {
		card.Border(1, t.SheetBorder)
	}
	if p.Raised {
		card.Shadow(0, 1, 4, 0, ui.RGBA(15, 40, 80, 0.06))
	}
	if p.Hoverable {
		card.Cursor(ui.CursorPointer)
		if card.Hovered() || p.hover {
			card.Shadow(0, 1, 2, -2, ui.RGBA(0, 0, 0, 0.16)).Shadow(0, 3, 6, 0, ui.RGBA(0, 0, 0, 0.12)).Shadow(0, 5, 12, 4, ui.RGBA(0, 0, 0, 0.09))
		}
	}
	body := func() {
		if p.Loading {
			Skeleton(c, SkeletonProps{Active: true, NoTitle: true, Rows: 3})
			return
		}
		fn()
	}
	head := func() {
		if p.Title == "" && p.Extra == nil {
			return
		}
		ui.Row(c).FillWidth().Gap(t.SpaceSM).AlignItems(ui.Center).Children(func() {
			ui.Text(c, p.Title).FontWeight(600).FontSize(titleSize).TextColor(t.Text).Grow(1).MinWidth(0).SingleLine()
			if p.Extra != nil {
				ui.Row(c).Shrink(0).Gap(t.SpaceSM).AlignItems(ui.Center).Wrap().GapY(4).Children(p.Extra)
			}
		})
	}
	if p.Cover == nil && p.Tabs == nil && p.Actions == nil {
		card.Gap(t.SpaceMD).Padding(pad).Children(func() {
			head()
			body()
		})
		return card
	}
	card.Clip().Children(func() {
		if p.Cover != nil {
			ui.Column(c).FillWidth().Children(p.Cover)
		}
		if p.Title != "" || p.Extra != nil || p.Tabs != nil {
			ui.Column(c).FillWidth().Padding(pad, pad, 0).Gap(t.SpaceSM).Children(func() {
				head()
				if p.Tabs != nil && p.Tab != nil {
					Tabs(c, p.Tab, p.Tabs, TabsProps{Size: pick(p.Size == Small, Small, Middle), Extra: p.TabExtra})
				}
			})
		}
		ui.Column(c).FillWidth().Gap(t.SpaceMD).Padding(pad).Children(body)
		if len(p.Actions) > 0 {
			ui.Row(c).FillWidth().Background(t.Surface).Draw(topRule(t.SheetBorder)).Padding(t.SpaceMD, 0).AlignItems(ui.Center).Children(func() {
				for i, a := range p.Actions {
					cell := ui.Row(c).Basis(0).Grow(1).Justify(ui.Center).AlignItems(ui.Center)
					if i < len(p.Actions)-1 {
						cell.Draw(rightRule(t.SheetBorder))
					}
					cell.Children(a)
				}
			})
		}
	})
	return card
}

// CardMetaProps are a card meta's options.
type CardMetaProps struct {
	// Avatar builds the picture at the start.
	Avatar      func()
	Title       string
	Description string
}

// CardMeta is Ant's Card.Meta: an avatar beside a bold title and a grey
// description.
func CardMeta(c *ui.Context, props CardMetaProps) {
	t := Of(c)
	ui.Row(c).FillWidth().Gap(t.SpaceLG).AlignItems(ui.Start).Children(func() {
		if props.Avatar != nil {
			ui.Row(c).Shrink(0).Children(props.Avatar)
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(t.SpaceSM).Children(func() {
			if props.Title != "" {
				ui.Text(c, props.Title).FontSize(t.FontSizeLG).FontWeight(600).TextColor(t.TextHeading).SingleLine()
			}
			if props.Description != "" {
				ui.Text(c, props.Description).FontSize(t.FontSize).TextColor(t.TextTertiary)
			}
		})
	})
}

// CardGridProps are a card grid's options.
type CardGridProps struct {
	// Columns is how many cells share a line; 0 is 3.
	Columns int
	// Hoverable lifts a cell on hover.
	Hoverable bool
}

// CardGrid is Ant's Card.Grid inside a card: n cells in Columns, split
// by hairlines, each built by cell.
func CardGrid(c *ui.Context, n int, cell func(i int), props ...CardGridProps) {
	p := first(props)
	t := Of(c)
	cols := max(p.Columns, 1)
	if p.Columns == 0 {
		cols = 3
	}
	ui.Column(c).FillWidth().Border(1, t.SheetBorder).Radius(4).Clip().Children(func() {
		for r := 0; r*cols < n; r++ {
			line := ui.Row(c).FillWidth().AlignItems(ui.Stretch)
			if (r+1)*cols < n {
				line.DrawOver(bottomRule(t.SheetBorder))
			}
			line.Children(func() {
				for i := r * cols; i < (r+1)*cols; i++ {
					g := ui.Column(c).Basis(0).Grow(1).MinWidth(0).Padding(24)
					if i%cols < cols-1 {
						g.DrawOver(rightRule(t.SheetBorder))
					}
					if i >= n {
						continue
					}
					if p.Hoverable && g.Hovered() {
						g.Background(t.Surface).Shadow(0, 1, 2, -2, ui.RGBA(0, 0, 0, 0.16)).Shadow(0, 3, 6, 0, ui.RGBA(0, 0, 0, 0.12))
					}
					g.Children(func() { cell(i) })
				}
			})
		}
	})
}

// SectionTitle is the settings section heading: a 4×24 brand bar, the
// title, and a rule to the end (settings layout: title).
func SectionTitle(c *ui.Context, title string) {
	t := Of(c)
	ui.Row(c).FillWidth().Height(48).Padding(0, t.SpaceXL).AlignItems(ui.Center).Shrink(0).Children(func() {
		ui.Box(c).Size(4, 24).Radius(4).Background(t.Brand).Shrink(0)
		ui.Text(c, title).FontWeight(600).FontSize(t.FontSizeTitle).TextColor(t.Text).SingleLine().MarginStart(12).
			Role(ui.RoleHeading).Level(2)
		ui.Box(c).Grow(1).Height(1).Background(t.Split).MarginStart(t.SpaceXL)
	})
}

// GlassProps are a glass card's options.
type GlassProps struct {
	// Padding overrides the service card's 20 20 14.
	Padding []float32
	Gap     float32
}

// Glass is the home page's frosted card (home layout: service-card).
func Glass(c *ui.Context, fn func(), props ...GlassProps) ui.Element {
	p := first(props)
	g := Of(c).Glass
	pad := p.Padding
	if len(pad) == 0 {
		pad = []float32{20, 20, 14}
	}
	fill := g.Card
	if c.Vibrancy() {
		fill = fill.Alpha(0.4)
	}
	return ui.Column(c).FillWidth().MinWidth(0).Clip().Padding(pad...).Gap(p.Gap).
		Background(fill).Radius(g.CardRadius).Border(1, g.CardEdge).
		Shadow(0, 10, 30, 0, g.CardShadow).Children(fn)
}

// HealthPillProps are a health pill's options.
type HealthPillProps struct {
	Icon ui.ImageSource
	// Fill stretches the pill across a narrow header.
	Fill bool
}

// HealthPill is the home header's status capsule: icon, title (amber when
// not ok), a sub line and a chevron (home layout: health-pill).
func HealthPill(c *ui.Context, title, sub string, ok bool, props ...HealthPillProps) {
	p := first(props)
	t := Of(c)
	g := t.Glass
	fill := g.Pill
	if c.Vibrancy() {
		fill = fill.Alpha(0.4)
	}
	pill := ui.Row(c).Padding(10, 16).Gap(12).AlignItems(ui.Center).Clip().
		Background(fill).Radius(999).Border(1, g.PillEdge).Shadow(0, 8, 24, 0, g.PillShadow)
	if p.Fill {
		pill.FillWidth().MinWidth(220)
	} else {
		pill.MinWidth(280)
	}
	pill.Children(func() {
		if p.Icon != nil {
			ui.Image(c, p.Icon).Size(22, 22).Fit(ui.Contain).Shrink(0)
		}
		ui.Column(c).Gap(2).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, title).FontWeight(600).FontSize(13).TextColor(pick(ok, t.Glass.PillText, t.Warning)).SingleLine().MinWidth(0)
			ui.Text(c, sub).FontSize(12).TextColor(g.Subtle).SingleLine().MinWidth(0)
		})
		ui.Box(c).Size(8, 8).BorderWidth(1.5, 1.5, 0, 0).BorderColor(g.Subtle).Rotate(45).Shrink(0)
	})
}

// AlertType is Ant's alert type.
type AlertType uint8

// Alert types, also the marks of Result and messages.
const (
	// Info is blue, with an "i".
	Info AlertType = iota
	// Success is green, with a tick.
	Success
	// Warning is orange, with an exclamation mark.
	Warning
	// Error is red, with a cross.
	Error
)

// AlertProps are an alert's options.
type AlertProps struct {
	Type        AlertType
	Description string
	// Banner is the default flat strip (RemoteConnect .current /
	// .currentOff): no border, no icon unless ShowIcon.
	Banner   bool
	ShowIcon bool
	// NoIcon drops the mark of a bordered alert; Icon replaces it.
	NoIcon bool
	Icon   *ui.SVG
	// Closable adds the close button, or CloseText a text one; the alert
	// then collapses away, and stays closed while *Closed (a state of
	// its own when nil) is true.
	Closable  bool
	CloseText string
	Closed    *bool
	// Action builds controls at the end, as Ant's action.
	Action func()
}

// alertMotion is how an alert leaves when closed.
var alertMotion = ui.ElementTransition{Duration: 200 * time.Millisecond, Exit: &ui.Motion{Opacity: 0, Collapse: true}}

// Alert is Ant's Alert with its icon, or with Banner the default strip.
// It reports a click of its close button.
func Alert(c *ui.Context, message string, props ...AlertProps) bool {
	p := first(props)
	t := Of(c)
	closable := p.Closable || p.CloseText != ""
	if !closable {
		alertBox(c, t, message, p, nil)
		return false
	}
	host := ui.Column(c).FillWidth()
	closed := p.Closed
	if closed == nil {
		closed = ui.Local(host, "closed", func() bool { return false })
	}
	clicked := false
	host.Children(func() {
		if !*closed {
			alertBox(c.Key("alert"), t, message, p, &clicked).Transition(transitionFor(c, alertMotion))
		}
	})
	if clicked {
		*closed = true
	}
	return clicked
}

func alertBox(c *ui.Context, t *Tokens, message string, p AlertProps, clicked *bool) ui.Element {
	bg, edge, icon, fg := alertColors(t, p.Type)
	if p.Banner {
		bg, fg = t.Banner[p.Type].Bg, t.Banner[p.Type].Fg
	}
	box := ui.Row(c).FillWidth().Gap(t.SpaceSM).Padding(t.SpaceSM+1, t.SpaceMD).Background(bg).Radius(t.RadiusBanner)
	if p.Banner {
		box.Padding(t.SpaceMD, t.SpaceLG)
	} else {
		box.Border(1, edge)
	}
	if p.Description != "" {
		box.Padding(t.SpaceMD+8, 24, t.SpaceMD+8, t.SpaceLG)
	}
	box.AlignItems(pick(p.Description == "", ui.Center, ui.Start)).Children(func() {
		if p.Banner && p.ShowIcon || !p.Banner && !p.NoIcon {
			sz := pick[float32](p.Description == "", 14, 24)
			if p.Icon != nil {
				ui.Icon(c, p.Icon).Size(sz, sz).TextColor(icon).Shrink(0)
			} else {
				statusIcon(c, p.Type, sz, icon)
			}
		}
		ui.Column(c).Grow(1).MinWidth(0).Gap(4).Children(func() {
			title := ui.Text(c, message).FontSize(pick(p.Description == "", t.FontSize, t.FontSizeLG)).TextColor(fg)
			if p.Banner {
				title.FontSize(13)
			}
			if p.Description != "" {
				ui.Text(c, p.Description).FontSize(t.FontSize).TextColor(t.Text)
			}
		})
		if p.Action != nil {
			ui.Row(c).Shrink(0).Gap(t.SpaceSM).AlignItems(ui.Center).Children(p.Action)
		}
		if clicked == nil {
			return
		}
		if p.CloseText != "" {
			x := ui.ButtonBase(c).Label(p.CloseText).FocusRing(false).Cursor(ui.CursorPointer).Shrink(0)
			fg := pick(x.Hovered(), t.TextHeading, t.TextTertiary)
			x.Children(func() { ui.Text(c, p.CloseText).FontSize(t.FontSizeSM).TextColor(fg) })
			*clicked = x.Clicked()
			return
		}
		*clicked = iconButton(c, iconClose, L(c).Close, t.TextTertiary, 12)
	})
	return box
}

func alertColors(t *Tokens, k AlertType) (bg, edge, icon, fg ui.Color) {
	switch k {
	case Success:
		return t.SuccessBg, t.SuccessBorder, t.SuccessIcon, t.Text
	case Warning:
		return t.WarningBg, t.WarningBorder, t.WarningIcon, t.Text
	case Error:
		return t.ErrorBg, t.ErrorBorder, t.Error, t.Text
	}
	return t.PrimaryBg, t.PrimaryBorder, t.Primary, t.Text
}

// EmptyProps are Ant Empty's options.
type EmptyProps struct {
	// Description is a second line under the title.
	Description string
	// Image replaces the inbox mark.
	Image func()
	// Extra builds actions under the copy (a button to retry, say).
	Extra func()
	// Simple drops the framed well (Ant's PRESENTED_IMAGE_SIMPLE).
	Simple bool
}

// Empty is a quiet well with an inbox mark and a line of text, for lists
// with nothing in them yet. Props are optional; Empty(c, text) is the
// original framed well.
func Empty(c *ui.Context, text string, props ...EmptyProps) {
	p := first(props)
	t := Of(c)
	box := emptyBody(c, text, p)
	if !p.Simple {
		box.Background(t.EmptyBg).Border(1, t.EmptyBorder).Radius(t.RadiusSheet)
	}
}

// emptyBody is Empty without its frame, for inside a table.
func emptyBody(c *ui.Context, text string, props ...EmptyProps) ui.Element {
	p := first(props)
	t := Of(c)
	if text == "" {
		text = L(c).Empty
	}
	return ui.Column(c).FillWidth().MinHeight(pick(p.Simple, float32(72), 96)).Gap(t.SpaceSM).Padding(t.SpaceLG).Center().Children(func() {
		switch {
		case p.Image != nil:
			p.Image()
		default:
			ui.Icon(c, iconEmpty).Size(40, 28).TextColor(t.EmptyIcon)
		}
		ui.Text(c, text).FontSize(t.FontSize).TextColor(t.TextDisabled)
		if p.Description != "" {
			ui.Text(c, p.Description).FontSize(t.FontSizeSM).TextColor(t.TextTertiary)
		}
		if p.Extra != nil {
			ui.Row(c).Gap(t.SpaceSM).Wrap().GapY(t.SpaceSM).Justify(ui.Center).Margin(t.SpaceXS, 0, 0).Children(p.Extra)
		}
	})
}

// TagTone is a tag's color preset.
type TagTone uint8

// Tag tones.
const (
	// TagDefault is the grey tag.
	TagDefault TagTone = iota
	// TagBlue is the primary-tinted tag.
	TagBlue
	// TagSuccess is green.
	TagSuccess
	// TagWarning is orange.
	TagWarning
	// TagError is red.
	TagError
	// TagBrand is the database type chip (settings layout: database-type).
	TagBrand
)

// TagProps are a tag's options.
type TagProps struct {
	Tone TagTone
	Size Size
	// Preset colors it with one of Ant's preset palettes by name ("blue",
	// "magenta", …): shade 1 fill, 3 border, 7 text.
	Preset string
	// Color, when set, fills it solid with white text, as Ant's custom
	// color tags.
	Color ui.Color
	// Icon leads the label.
	Icon *ui.SVG
	// Closable adds the close mark; Tag reports its click, and the app
	// drops the tag.
	Closable bool
	// Borderless drops the border (Ant bordered={false}).
	Borderless bool
}

// tagColors is the fill, border and text of a tag.
func tagColors(t *Tokens, p TagProps) (bg, edge, fg ui.Color) {
	if p.Color != (ui.Color{}) {
		return p.Color, p.Color, ui.Hex("#FFFFFF")
	}
	if pal, ok := t.Preset(p.Preset); ok {
		return pal[0], pal[2], pal[6]
	}
	switch p.Tone {
	case TagBlue:
		return t.PrimaryBg, t.PrimaryBorder, t.Primary
	case TagSuccess:
		return t.SuccessBg, t.SuccessBorder, t.SuccessText
	case TagWarning:
		return t.WarningBg, t.WarningBorder, t.WarningText
	case TagError:
		return t.ErrorBg, t.ErrorBorder, t.ErrorText
	case TagBrand:
		return t.BrandSoft, t.BrandSoft, t.Brand
	}
	return t.FillAlter, t.BorderButton, t.Text
}

// Tag is Ant's Tag; Large makes it a control-high chip. It reports a
// click of its close mark.
func Tag(c *ui.Context, label string, props ...TagProps) bool {
	p := first(props)
	t := Of(c)
	bg, edge, fg := tagColors(t, p)
	tag := ui.Row(c).Background(bg).Radius(4).Gap(4).AlignItems(ui.Center).Shrink(0).AlignSelf(ui.Start).Clip()
	if !p.Borderless {
		tag.Border(1, edge)
	}
	size := t.FontSizeSM
	switch {
	case p.Tone == TagBrand:
		// settings layout: database-type: 123×40 chip.
		tag.Width(123).Height(t.ControlHeightLG).Padding(0).Justify(ui.Center)
		size = t.FontSize
	case p.Size == Large:
		tag.Height(t.ControlHeightLG).MinWidth(96).Padding(0, t.SpaceLG).Justify(ui.Center)
		size = t.FontSize
	default:
		tag.Height(22).Padding(0, 7)
	}
	closed := false
	tag.Children(func() {
		if p.Icon != nil {
			ui.Icon(c, p.Icon).Size(12, 12).TextColor(fg).Shrink(0)
		}
		ui.Text(c, label).FontSize(size).TextColor(fg).SingleLine().MinWidth(0)
		if p.Closable {
			x := ui.Box(c).Size(12, 12).Shrink(0).Cursor(ui.CursorPointer).Label("关闭标签 " + label).Focusable()
			mark := pick(p.Color != (ui.Color{}), fg, t.TextTertiary)
			if x.Hovered() {
				mark = pick(p.Color != (ui.Color{}), fg, t.TextHeading)
			}
			x.Children(func() { ui.Icon(c, iconClose).Size(10, 10).TextColor(mark).Margin(1) })
			closed = x.Clicked()
		}
	})
	return closed
}

// CheckableTag is Ant's Tag.CheckableTag: a borderless tag that toggles
// *checked, filled primary while checked. It reports a change.
func CheckableTag(c *ui.Context, label string, checked *bool) bool {
	t := Of(c)
	tag := ui.Row(c).Height(22).Padding(0, 7).Radius(4).AlignItems(ui.Center).Shrink(0).AlignSelf(ui.Start).
		Cursor(ui.CursorPointer).Focusable().Label(label).Checked(*checked)
	v := stateOf(tag, false, rest)
	bg, fg := ui.Color{}, t.Text
	switch {
	case *checked && v == active:
		bg, fg = t.PrimaryActive, t.TextOnPrimary
	case *checked && v == hover:
		bg, fg = t.PrimaryHover, t.TextOnPrimary
	case *checked:
		bg, fg = t.Primary, t.TextOnPrimary
	case v == active:
		bg = t.Fill
	case v == hover:
		fg = t.Primary
	}
	tag.Background(bg)
	if v == focus {
		focusRing(tag, t)
	}
	tag.Children(func() {
		ui.Text(c, label).FontSize(t.FontSizeSM).TextColor(fg).SingleLine()
	})
	if tag.Clicked() {
		*checked = !*checked
		return true
	}
	return false
}

// Badge is Ant's Badge status: a dot and its label, green when ok.
func Badge(c *ui.Context, label string, ok bool) {
	t := Of(c)
	tone := pick(ok, t.Success, t.Danger)
	ui.Row(c).Gap(6).AlignItems(ui.Center).Shrink(0).Children(func() {
		ui.Box(c).Size(8, 8).Radius(4).Background(tone).Shrink(0)
		ui.Text(c, label).FontSize(13).TextColor(tone).SingleLine()
	})
}
