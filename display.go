package kit

import "github.com/egoist/mygo/ui"

// SpaceProps are a Space's options.
type SpaceProps struct {
	// Gap is the space between items; zero is Ant's default 8.
	Gap      float32
	Vertical bool
	// Wrap lets a horizontal Space break onto new lines.
	Wrap bool
	// Align places the items across the Space: Center when horizontal and
	// Start when vertical unless set.
	Align ui.Align
}

// Space lays out what fn builds with even gaps, as Ant's Space.
func Space(c *ui.Context, fn func(), props ...SpaceProps) {
	p := first(props)
	gap := pick(p.Gap == 0, 8, p.Gap)
	var e ui.Element
	if p.Vertical {
		e = ui.Column(c).Gap(gap).AlignItems(p.Align)
	} else {
		e = ui.Row(c).Gap(gap).AlignItems(pick(p.Align == ui.Start, ui.Center, p.Align))
		if p.Wrap {
			e.Wrap().GapY(gap).FillWidth().MinWidth(0)
		}
	}
	e.Children(fn)
}

// DividerProps are a divider's options.
type DividerProps struct {
	// Text sits in the line, near its start.
	Text   string
	Dashed bool
	// Vertical is a short upright rule between inline items.
	Vertical bool
}

// Divider is Ant's Divider: a hairline across its column, optionally
// dashed or with a caption.
func Divider(c *ui.Context, props ...DividerProps) {
	p := first(props)
	t := Of(c)
	tone := t.DividerColor
	line := func(e ui.Element) ui.Element {
		if p.Dashed {
			return e.Height(1).BorderWidth(1, 0, 0, 0).BorderColor(t.DividerColor.Alpha(2)).BorderStyle(ui.BorderDashed)
		}
		return e.Height(1).Background(tone)
	}
	switch {
	case p.Vertical:
		ui.Box(c).Width(1).Height(t.FontSize*1.1).Shrink(0).Background(tone).Margin(0, t.SpaceSM)
	case p.Text == "":
		line(ui.Box(c).FillWidth().Shrink(0)).Margin(t.SpaceXS, 0)
	default:
		ui.Row(c).FillWidth().Gap(t.SpaceLG).AlignItems(ui.Center).Margin(t.SpaceXS, 0).Children(func() {
			line(ui.Box(c).Width(40).Shrink(0))
			ui.Text(c, p.Text).FontSize(t.FontSizeLG).FontWeight(500).TextColor(t.Text).SingleLine()
			line(ui.Box(c).Grow(1))
		})
	}
}

// DescItem is one label / value pair of Descriptions. Render, when set,
// builds the value instead of Value.
type DescItem struct {
	Label, Value string
	// Span is how many columns the item takes; 0 means 1.
	Span   int
	Render func()
	// Copyable puts Ant Typography's copy control after Value (the word
	// 「复制」: ui.RichText cannot sit an icon on the last glyph).
	Copyable bool
}

// DescriptionsProps are a description list's options.
type DescriptionsProps struct {
	Title string
	// Columns is how many items share a line; 0 means 1.
	Columns  int
	Bordered bool
	Size     Size
	// LabelWidth fixes the label column of a bordered list; 0 is 120.
	LabelWidth float32
}

// Descriptions is Ant's Descriptions: label / value pairs in Columns,
// either as "label：value" lines or, Bordered, as a table with grey
// label cells. Values are selectable, for copying serials and paths.
func Descriptions(c *ui.Context, items []DescItem, props ...DescriptionsProps) {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	t := Of(c)
	cols := max(p.Columns, 1)
	rows := descRows(items, cols)
	labelW := pick(p.LabelWidth == 0, 120, p.LabelWidth)
	padY := map[Size]float32{Small: 8, Middle: 12, Large: 16}[p.Size]
	padX := pick[float32](p.Size == Small, 16, 24)
	value := func(it DescItem) {
		if it.Render != nil {
			it.Render()
			return
		}
		text := pick(it.Value == "", "-", it.Value)
		if it.Copyable && it.Value != "" {
			TypoText(c, text, TypoProps{Copyable: true, Size: t.FontSize})
			return
		}
		ui.Text(c, text).FontSize(t.FontSize).TextColor(t.Text).
			Selectable().MaxLines(2).MinWidth(0)
	}
	ui.Column(c).FillWidth().MinWidth(0).Gap(t.SpaceLG).Clip().Children(func() {
		if p.Title != "" {
			ui.Text(c, p.Title).FontSize(t.FontSizeLG).FontWeight(600).TextColor(t.Text).SingleLine().MinWidth(0)
		}
		if !p.Bordered {
			ui.Column(c).FillWidth().MinWidth(0).Gap(t.SpaceLG).Children(func() {
				for _, row := range rows {
					ui.Row(c).FillWidth().Gap(t.SpaceLG).AlignItems(ui.Start).Children(func() {
						for _, it := range row {
							ui.Row(c).Basis(0).Grow(float32(span(it))).MinWidth(0).Gap(t.SpaceXS).AlignItems(ui.Start).Clip().Children(func() {
								ui.Text(c, it.Label+"：").FontSize(t.FontSize).TextColor(t.TextTertiary).Shrink(0).SingleLine()
								ui.Column(c).Grow(1).MinWidth(0).Clip().Children(func() { value(it) })
							})
						}
					})
				}
			})
			return
		}
		ui.Column(c).FillWidth().MinWidth(0).Border(1, t.BorderSecondary).Radius(8).Clip().Children(func() {
			for r, row := range rows {
				line := ui.Row(c).FillWidth().AlignItems(ui.Stretch)
				if r < len(rows)-1 {
					line.DrawOver(bottomRule(t.BorderSecondary))
				}
				line.Children(func() {
					for i, it := range row {
						last := i == len(row)-1
						ui.Row(c).Basis(0).Grow(float32(span(it))).MinWidth(0).AlignItems(ui.Stretch).Clip().Children(func() {
							ui.Row(c).Width(labelW).Shrink(0).Padding(padY, padX).Background(t.FillAlter).
								Draw(rightRule(t.BorderSecondary)).Clip().Children(func() {
								ui.Text(c, it.Label).FontSize(t.FontSize).TextColor(t.TextHeading).SingleLine().MinWidth(0)
							})
							cell := ui.Column(c).Grow(1).MinWidth(0).Padding(padY, padX).Justify(ui.Center).Clip()
							if !last {
								cell.Draw(rightRule(t.BorderSecondary))
							}
							cell.Children(func() { value(it) })
						})
					}
				})
			}
		})
	})
}

func span(it DescItem) int { return max(it.Span, 1) }

// descRows packs items into lines of cols columns, stretching the last
// item of a short line to fill it, as Ant does.
func descRows(items []DescItem, cols int) [][]DescItem {
	var rows [][]DescItem
	var cur []DescItem
	used := 0
	for _, it := range items {
		s := min(span(it), cols)
		if used+s > cols && len(cur) > 0 {
			cur[len(cur)-1].Span = span(cur[len(cur)-1]) + cols - used
			rows = append(rows, cur)
			cur, used = nil, 0
		}
		it.Span = s
		cur = append(cur, it)
		used += s
	}
	if len(cur) > 0 {
		cur[len(cur)-1].Span = span(cur[len(cur)-1]) + cols - used
		rows = append(rows, cur)
	}
	return rows
}

func rightRule(split ui.Color) func(*ui.Painter, ui.Rect) {
	return func(p *ui.Painter, r ui.Rect) {
		p.Fill(ui.Rect{X: r.X + r.W - 1, Y: r.Y, W: 1, H: r.H}, split, 0)
	}
}
