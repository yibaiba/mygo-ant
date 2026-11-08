package kit

import "github.com/egoist/mygo/ui"

// TimelineMode is which side of the axis a timeline's content takes.
type TimelineMode uint8

// Timeline modes.
const (
	// TimelineLeft keeps the axis on the left, content to its right.
	TimelineLeft TimelineMode = iota
	// TimelineRight keeps the axis on the right.
	TimelineRight
	// TimelineAlternate centers the axis, the content alternating sides
	// from the right.
	TimelineAlternate
)

// TimelineItem is one event of a timeline.
type TimelineItem struct {
	// Label is the text on the other side of the axis, as a time; any
	// label centers the axis.
	Label   string
	Content string
	// Render builds the content instead of Content.
	Render func()
	// Color names the dot's color: "blue" (the default), "red", "green",
	// "gray", or an Ant preset; ColorValue sets it directly.
	Color      string
	ColorValue ui.Color
	// Dot builds a custom dot, as an icon.
	Dot func()
}

// TimelineProps are a timeline's options.
type TimelineProps struct {
	Mode TimelineMode
	// Pending adds a last item with a spinning dot and this text, joined
	// by a dotted tail; PendingDot builds its dot instead.
	Pending    string
	PendingDot func()
	// Reverse lists the items newest first.
	Reverse bool
}

func timelineColor(t *Tokens, it TimelineItem) ui.Color {
	if it.ColorValue != (ui.Color{}) {
		return it.ColorValue
	}
	switch it.Color {
	case "", "blue":
		return t.Primary
	case "red":
		return t.Error
	case "green":
		return t.SuccessIcon
	case "gray":
		return t.TextDisabled
	}
	return toneColor(t, it.Color, t.Primary)
}

// timelineRow is a laid out timeline item: its side and tail.
type timelineRow struct {
	item    TimelineItem
	pending bool
	// contentStart puts the content after the axis.
	contentStart bool
	// tail is the line down to the next item: 0 none, 1 solid, 2 dotted.
	tail uint8
}

// timelineRows orders items as props say and gives each its side and
// tail.
func timelineRows(items []TimelineItem, p TimelineProps) []timelineRow {
	var rows []timelineRow
	for _, it := range items {
		rows = append(rows, timelineRow{item: it})
	}
	pending := p.Pending != "" || p.PendingDot != nil
	if pending {
		rows = append(rows, timelineRow{item: TimelineItem{Content: p.Pending, Dot: p.PendingDot}, pending: true})
	}
	if p.Reverse {
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}
	for i := range rows {
		switch p.Mode {
		case TimelineRight:
			rows[i].contentStart = false
		case TimelineAlternate:
			rows[i].contentStart = i%2 == 0
		default:
			rows[i].contentStart = true
		}
		if i < len(rows)-1 {
			rows[i].tail = 1
			if rows[i+1].pending || rows[i].pending {
				rows[i].tail = 2
			}
		}
	}
	return rows
}

// Timeline is Ant's Timeline: dots on an axis, each with its content
// (and a label across the axis), colored by status, with a pending tail.
func Timeline(c *ui.Context, items []TimelineItem, props ...TimelineProps) {
	p := first(props)
	t := Of(c)
	labeled := false
	for _, it := range items {
		labeled = labeled || it.Label != ""
	}
	centered := p.Mode == TimelineAlternate || labeled
	rows := timelineRows(items, p)
	ui.Column(c).FillWidth().Children(func() {
		for _, r := range rows {
			text := func(s string, end bool) {
				e := ui.Text(c, s).FontSize(t.FontSize).TextColor(t.Text).LineHeight(22.0 / 14)
				if end {
					e.TextAlign(ui.End)
				}
			}
			content := func(end bool) {
				ui.Column(c).Basis(0).Grow(1).MinWidth(0).Padding(0, 0, 20).AlignItems(pick(end, ui.End, ui.Start)).Children(func() {
					if r.item.Render != nil {
						r.item.Render()
					} else {
						text(r.item.Content, end)
					}
				})
			}
			label := func(end bool) {
				ui.Column(c).Basis(0).Grow(1).MinWidth(0).AlignItems(pick(end, ui.End, ui.Start)).Children(func() {
					if r.item.Label != "" {
						text(r.item.Label, end)
					}
				})
			}
			ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Stretch).Children(func() {
				switch {
				case centered && r.contentStart:
					label(true)
					timelineAxis(c, t, r)
					content(false)
				case centered:
					content(true)
					timelineAxis(c, t, r)
					label(false)
				case r.contentStart:
					timelineAxis(c, t, r)
					content(false)
				default:
					content(true)
					timelineAxis(c, t, r)
				}
			})
		}
	})
}

// timelineAxis is an item's dot, lined up with its first line of text,
// and the tail down to the next.
func timelineAxis(c *ui.Context, t *Tokens, r timelineRow) {
	tone := timelineColor(t, r.item)
	custom := r.item.Dot != nil || r.pending
	ui.Column(c).Width(10).Shrink(0).AlignItems(ui.Center).Draw(func(p *ui.Painter, b ui.Rect) {
		x := b.X + b.W/2 - 1
		top := b.Y + 16
		if custom {
			top = b.Y + 22
		}
		switch r.tail {
		case 1:
			p.Fill(ui.Rect{X: x, Y: top, W: 2, H: b.Y + b.H - top}, t.Split.Alpha(0.6).Over(t.Surface), 0)
		case 2:
			for y := top; y < b.Y+b.H; y += 6 {
				p.Fill(ui.Rect{X: x, Y: y, W: 2, H: min(3, b.Y+b.H-y)}, t.Split.Over(t.Surface), 0)
			}
		}
	}).Children(func() {
		switch {
		case r.item.Dot != nil:
			ui.Row(c).Height(22).AlignItems(ui.Center).Background(t.Surface).Children(r.item.Dot)
		case r.pending:
			ui.Row(c).Height(22).AlignItems(ui.Center).Children(func() { spinDots(c, 14, t.Primary) })
		default:
			ui.Box(c).Size(10, 10).MarginY(6).Radius(5).Background(t.Surface).Border(3, tone)
		}
	})
}
