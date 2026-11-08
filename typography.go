package kit

import (
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
)

// TypoType colors a Typography text, as Ant's type prop.
type TypoType uint8

// Typography types.
const (
	TypoDefault TypoType = iota
	TypoSecondary
	TypoSuccess
	TypoWarning
	TypoDanger
	TypoDisabled
)

// TypoProps are the options of Ant's Typography texts.
type TypoProps struct {
	Type TypoType
	// The text styles, as Ant's strong, italic, underline, delete, code,
	// mark and keyboard props.
	Strong, Italic, Underline, Delete, Code, Mark, Keyboard bool
	// Size overrides the font size.
	Size float32

	// Copyable puts a copy button after the text, copying CopyText (the
	// text when empty) and turning into a check for a few seconds.
	Copyable bool
	CopyText string
	// Editable, when set, is the text, which an edit button after it
	// turns into an input: Enter keeps the edit, Escape drops it.
	Editable *string

	// Ellipsis cuts the text to one line, or to Rows lines, with "…";
	// Expandable then adds "展开" after it, and Collapsible "收起" once
	// it is open. Expanded, when set, holds whether it is.
	Ellipsis    bool
	Rows        int
	Expandable  bool
	Collapsible bool
	Expanded    *bool
	// EllipsisTooltip shows the whole text on hover while it is cut.
	EllipsisTooltip bool
}

func typoColor(t *Tokens, k TypoType, base ui.Color) ui.Color {
	switch k {
	case TypoSecondary:
		return t.TextTertiary
	case TypoSuccess:
		return t.SuccessIcon
	case TypoWarning:
		return t.WarningIcon
	case TypoDanger:
		return t.Error
	case TypoDisabled:
		return t.TextDisabled
	}
	return base
}

// typoSpan is text in the styles of p.
func typoSpan(t *Tokens, text string, p TypoProps, base ui.Color, size float32) ui.Span {
	s := ui.Span{Text: text, Size: size, Color: typoColor(t, p.Type, base),
		Italic: p.Italic, Underline: p.Underline, Strikethrough: p.Delete}
	if p.Strong {
		s.Weight = 600
	}
	if p.Mark {
		s.Background, s.Radius = ui.Hex("#ffe58f"), 0
		if t.Dark {
			s.Background = ui.Hex("#594214")
		}
	}
	if p.Code || p.Keyboard {
		s.Font = "monospace"
		s.Size = size * 0.85
	}
	return s
}

// codePad is the room inside a code or keyboard box on either side of
// its text, as Ant's 0.4em padding: a narrow no-break space, part of the
// span, so the box takes its room in the line and wraps with the text.
const codePad = "\u202f"

// codeEnd closes a code run after its pad: ui measures a text's width
// without trailing spaces but lays it out with them, so a run that ends a
// text on a pad would wrap the pad to a line of its own.
const codeEnd = codePad + "\u2060"

// codeRun is a code or keyboard run of a rich text, in UTF-16 offsets as
// TextLayout takes them.
type codeRun struct {
	start, end int
	kbd        bool
}

// codeSpans pads the code and keyboard runs of spans and returns them
// with where those runs are.
func codeSpans(spans []ui.Span, kinds []TypoProps) ([]ui.Span, []codeRun) {
	var runs []codeRun
	at := 0
	out := make([]ui.Span, len(spans))
	for i, sp := range spans {
		if k := kinds[i]; k.Code || k.Keyboard {
			sp.Text = codePad + sp.Text + codeEnd
			n := utf16Len(sp.Text)
			runs = append(runs, codeRun{start: at, end: at + n, kbd: k.Keyboard && !k.Code})
		}
		out[i] = sp
		at += utf16Len(sp.Text)
	}
	return out, runs
}

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += max(utf16.RuneLen(r), 1)
	}
	return n
}

// codeBoxes paints Ant's code and keyboard boxes under the runs of txt, a
// RichText of spans in font: ui spans take a background but no border,
// so the text is shaped again at the width txt has this frame and each
// line of a run gets a box with a 1px border (2px under a key), as Ant's
// inline code and kbd elements wrap.
func codeBoxes(txt ui.Element, t *Tokens, spans []ui.Span, runs []codeRun, font ui.Font) {
	if len(runs) == 0 {
		return
	}
	type shaped struct {
		w     float32
		rects [][]ui.Rect
	}
	var cache shaped
	edge := ui.RGBA(100, 100, 100, 0.2)
	txt.Draw(func(pa *ui.Painter, r ui.Rect) {
		if cache.rects == nil || cache.w != r.W {
			l := ui.ShapeRichText(spans, font, r.W)
			cache = shaped{w: r.W, rects: make([][]ui.Rect, len(runs))}
			for i, run := range runs {
				cache.rects[i] = l.SelectionRects(ui.TextInputRange{Start: run.start, End: run.end})
			}
		}
		for i, run := range runs {
			for _, b := range cache.rects[i] {
				box := ui.Rect{X: r.X + b.X, Y: r.Y + b.Y + 1, W: b.W, H: b.H - 2}
				if box.Y >= r.Y+r.H || box.W <= 0 {
					continue
				}
				pa.Fill(box, ui.RGBA(150, 150, 150, pick[float32](run.kbd, 0.06, 0.1)), 3)
				pa.Stroke(box, edge, 3, 1)
				if run.kbd {
					pa.Fill(ui.Rect{X: box.X + 2, Y: box.Y + box.H - 1.5, W: box.W - 4, H: 1}, edge, 0)
				}
			}
		}
	})
}

// typoFont is the font a typography text of size is shaped in.
func typoFont(t *Tokens, size float32) ui.Font {
	return ui.Font{Family: t.Font, Size: size, Lang: t.Lang}
}

// typoCut is how many runes of text fit maxH high with tail after
// them, height measuring a text; all of them when the text fits.
func typoCut(text, tail string, maxH float32, height func(string) float32) int {
	n := utf8.RuneCountInString(text)
	if height(text) <= maxH {
		return n
	}
	runes := []rune(text)
	lo, hi := 0, n
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if height(string(runes[:mid])+tail) <= maxH {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo
}

// typoHold is what a Typography text keeps between frames.
type typoHold struct {
	copiedAt time.Time
	editing  bool
	started  bool
	draft    string
	expanded bool
	// The last cut, for the text, width and rows it was measured at.
	cutFor   string
	cutWidth float32
	cut      int
}

// typography builds a text in Ant's Typography styles.
func typography(c *ui.Context, text string, p TypoProps, base ui.Color, size float32, weight int, multi bool) ui.Element {
	t := Of(c)
	if p.Size > 0 {
		size = p.Size
	}
	if p.Editable != nil {
		text = *p.Editable
	}
	wrap := ui.Row(c).AlignItems(ui.End).Gap(4).MinWidth(0)
	if multi {
		wrap.FillWidth()
	} else {
		wrap.Shrink(0)
	}
	hold := ui.Local(wrap, "typo", func() typoHold { return typoHold{} })
	expanded := p.Expanded
	if expanded == nil {
		expanded = &hold.expanded
	}
	wrap.Children(func() {
		if hold.editing {
			typoEdit(c, t, p.Editable, hold, size)
			return
		}
		span := typoSpan(t, text, p, base, size)
		if weight > 0 && span.Weight == 0 {
			span.Weight = weight
		}
		rows := p.Rows
		if p.Ellipsis && rows == 0 {
			rows = 1
		}
		var txt ui.Element
		cutShown := false
		showExpand := false
		showCollapse := false
		switch {
		case rows > 0 && !*expanded && (p.Expandable || !multi && rows == 1):
			// Copy/edit are now last-line texts, so they share the wrap
			// width instead of sitting in a 24px sibling column.
			width := wrap.Bounds().W
			tail := "…" + pick(p.Expandable, " "+L(c).Expand, "") + typoTrailTail(c, p)
			if width > 0 && (hold.cutFor != text || hold.cutWidth != width) {
				_, line := c.MeasureText(0, ui.Span{Text: "Ag", Size: size})
				hold.cut = typoCut(text, tail, line*float32(rows)+0.5, func(s string) float32 {
					_, h := c.MeasureText(width, ui.Span{Text: s, Size: size, Weight: span.Weight})
					return h
				})
				hold.cutFor, hold.cutWidth = text, width
			}
			shown := text
			if width > 0 && hold.cut < utf8.RuneCountInString(text) {
				shown = string([]rune(text)[:hold.cut]) + "…"
				cutShown = true
			}
			span.Text = shown
			spans, runs := codeSpans([]ui.Span{span}, []TypoProps{p})
			txt = ui.RichText(c, spans...).Grow(1).MinWidth(0)
			codeBoxes(txt, t, spans, runs, typoFont(t, size))
			showExpand = cutShown && p.Expandable
		case rows > 0 && !*expanded:
			spans, runs := codeSpans([]ui.Span{span}, []TypoProps{p})
			txt = ui.RichText(c, spans...).MaxLines(rows).Ellipsis("…").Grow(1).MinWidth(0)
			codeBoxes(txt, t, spans, runs, typoFont(t, size))
		default:
			spans, runs := codeSpans([]ui.Span{span}, []TypoProps{p})
			txt = ui.RichText(c, spans...)
			codeBoxes(txt, t, spans, runs, typoFont(t, size))
			if multi {
				txt.Grow(1).MinWidth(0)
			}
			showCollapse = *expanded && p.Collapsible
		}
		if cutShown && p.EllipsisTooltip {
			TooltipOn(c, txt, text)
		}
		if showExpand || showCollapse || p.Copyable || p.Editable != nil {
			txt.Children(func() {
				if showExpand {
					link := ui.Text(c, " "+L(c).Expand).TextColor(t.Primary).Cursor(ui.CursorPointer)
					if link.Clicked() {
						*expanded = true
					}
				}
				if showCollapse {
					link := ui.Text(c, " "+L(c).Collapse).TextColor(t.Primary).Cursor(ui.CursorPointer)
					if link.Clicked() {
						*expanded = false
					}
				}
				typoTrail(c, t, text, p, hold)
			})
		}
	})
	return wrap
}

func typoTrailTail(c *ui.Context, p TypoProps) string {
	s := ""
	if p.Editable != nil {
		s += " " + L(c).Edit
	}
	if p.Copyable {
		s += " " + L(c).Copy
	}
	return s
}

// typoTrail is the copy and edit text after the last glyphs of a
// Typography run. ui.RichText only continues a paragraph with text
// elements (mygo ui/richtext.go:80), so an icon ButtonBase cannot join
// the line; clickable locale words are as close as that allows.
func typoTrail(c *ui.Context, t *Tokens, text string, p TypoProps, hold *typoHold) {
	if p.Editable != nil {
		ui.Text(c, " ")
		link := ui.Text(c, L(c).Edit).TextColor(t.Primary).Cursor(ui.CursorPointer)
		if link.Clicked() {
			hold.editing, hold.started, hold.draft = true, false, text
		}
		TooltipOn(c, link, L(c).Edit)
	}
	if p.Copyable {
		copied := !hold.copiedAt.IsZero() && c.Now().Sub(hold.copiedAt) < 3*time.Second
		ui.Text(c, " ")
		link := ui.Text(c, pick(copied, L(c).Copied, L(c).Copy)).
			TextColor(pick(copied, t.SuccessIcon, t.Primary)).Cursor(ui.CursorPointer)
		if link.Clicked() {
			c.WriteClipboard(pick(p.CopyText == "", text, p.CopyText))
			hold.copiedAt = c.Now()
			c.After(3 * time.Second)
		}
		TooltipOn(c, link, pick(copied, L(c).Copied, L(c).Copy))
	}
}

// typoEdit is the input an editable text turns into.
func typoEdit(c *ui.Context, t *Tokens, value *string, hold *typoHold, size float32) {
	in := ui.TextInputBase(c, &hold.draft).Height(t.ControlHeight).Padding(0, 11).FontSize(size).
		TextColor(t.Text).Radius(t.RadiusControl).FocusRing(false).Grow(1).MinWidth(0).Label("编辑文本")
	fieldLook(in, t, focus, StatusNone)
	cancel := false
	in.OnTextKey(func(k ui.TextKey) bool {
		if k.Key == ui.KeyEscape {
			cancel = true
			return true
		}
		return false
	})
	IconNamed(c, "enter", IconProps{Size: 14, Color: t.TextTertiary}).Margin(0, 0, 9, -26)
	if !hold.started {
		in.Focus()
		hold.started = true
		return
	}
	switch {
	case cancel:
		hold.editing = false
	case in.Submitted() || !in.Focused():
		*value = hold.draft
		hold.editing = false
	}
}

// TitleProps are a Typography title's options.
type TitleProps struct {
	// Level is 1 to 5, as h1 to h5; 1 when zero.
	Level int
	TypoProps
}

// titleSizes are Ant's heading sizes, h1 to h5.
var titleSizes = [5]float32{38, 30, 24, 20, 16}

// Title is Ant's Typography.Title: a heading of Level 1 to 5, 600
// weight in the heading color.
func Title(c *ui.Context, text string, props ...TitleProps) ui.Element {
	p := first(props)
	t := Of(c)
	level := max(1, min(5, pick(p.Level == 0, 1, p.Level)))
	e := typography(c, text, p.TypoProps, t.TextHeading, titleSizes[level-1], 600, true)
	return e.Margin(0, 0, titleSizes[level-1]*0.5, 0)
}

// TypoText is Ant's Typography.Text: inline text in the type's color and
// the styles TypoProps turns on.
func TypoText(c *ui.Context, text string, props ...TypoProps) ui.Element {
	t := Of(c)
	return typography(c, text, first(props), t.Text, t.FontSize, 0, false)
}

// Paragraph is Ant's Typography.Paragraph: a wrapping block of text,
// cut to Rows lines with Ellipsis, followed by a 1em gap.
func Paragraph(c *ui.Context, text string, props ...TypoProps) ui.Element {
	t := Of(c)
	return typography(c, text, first(props), t.Text, t.FontSize, 0, true).Margin(0, 0, t.FontSize, 0)
}

// TextLink is Ant's Typography.Link: text in the link color, opening url
// in the browser when set; it reports a click.
func TextLink(c *ui.Context, text, url string, props ...TypoProps) bool {
	p := first(props)
	t := Of(c)
	var e ui.Element
	if url != "" {
		e = ui.Link(c, text, url)
	} else {
		e = ui.Text(c, text).Focusable().Cursor(ui.CursorPointer)
	}
	off := p.Type == TypoDisabled
	fg := t.Primary
	switch {
	case off:
		fg = t.TextDisabled
		e.Cursor(ui.CursorNotAllowed)
	case p.Type == TypoDanger:
		fg = pick(e.Hovered(), t.ErrorHover, t.Error)
	case e.Pressed():
		fg = t.PrimaryActive
	case e.Hovered():
		fg = t.PrimaryHover
	}
	e.TextColor(fg).FontSize(pick(p.Size > 0, p.Size, t.FontSize))
	if p.Strong {
		e.FontWeight(600)
	}
	if p.Underline {
		e.Underline()
	}
	if p.Italic {
		e.Italic()
	}
	if p.Delete {
		e.Strikethrough()
	}
	return e.Clicked() && !off
}

// TypoRun is a run of a RichParagraph: text in the styles of its props
// (Type, Strong, Italic, Underline, Delete, Code, Mark, Keyboard; the
// others are the paragraph's).
type TypoRun struct {
	Text  string
	Props TypoProps
}

// RichParagraph is Ant's Typography.Paragraph holding Text children of
// their own styles, as prose with inline code and keys: one wrapping
// block, each code or keyboard run boxed on every line it wraps to,
// followed by a 1em gap.
func RichParagraph(c *ui.Context, runs ...TypoRun) ui.Element {
	t := Of(c)
	spans := make([]ui.Span, len(runs))
	kinds := make([]TypoProps, len(runs))
	for i, r := range runs {
		spans[i] = typoSpan(t, r.Text, r.Props, t.Text, t.FontSize)
		kinds[i] = r.Props
	}
	spans, codes := codeSpans(spans, kinds)
	txt := ui.RichText(c, spans...).FillWidth().MinWidth(0).Margin(0, 0, t.FontSize, 0)
	codeBoxes(txt, t, spans, codes, typoFont(t, t.FontSize))
	return txt
}
