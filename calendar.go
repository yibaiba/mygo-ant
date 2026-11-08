package kit

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// CalendarMode is what a calendar shows: a month of days or a year of
// months.
type CalendarMode uint8

// Calendar modes.
const (
	CalendarMonth CalendarMode = iota
	CalendarYear
)

// CalendarProps are a calendar's options.
type CalendarProps struct {
	// Card is Ant's fullscreen={false}: a small bordered panel; the
	// default fills its column with tall cells.
	Card bool
	// Mode, when set, holds the mode the header switches.
	Mode *CalendarMode
	// CellRender builds a day's content under its number;
	// MonthCellRender a month's in the year mode.
	CellRender      func(day time.Time)
	MonthCellRender func(month time.Time)
	DisabledDate    func(day time.Time) bool
	// ValidRange, when both are set, limits the days and the header's
	// years to it.
	ValidRange [2]time.Time
	// HideHeader drops the year / month / mode header.
	HideHeader bool
}

// calendarYears is the years the header offers around year, clamped to
// the valid range.
func calendarYears(year int, valid [2]time.Time) []int {
	lo, hi := year-10, year+9
	if !valid[0].IsZero() && !valid[1].IsZero() {
		lo, hi = valid[0].Year(), valid[1].Year()
	}
	var out []int
	for y := lo; y <= hi; y++ {
		out = append(out, y)
	}
	return out
}

// calendarOff reports whether day is out of the valid range or disabled.
func calendarOff(p CalendarProps, day time.Time) bool {
	if !p.ValidRange[0].IsZero() && day.Before(dayOf(p.ValidRange[0])) {
		return true
	}
	if !p.ValidRange[1].IsZero() && day.After(dayOf(p.ValidRange[1])) {
		return true
	}
	return p.DisabledDate != nil && p.DisabledDate(day)
}

// Calendar is Ant's Calendar: a header choosing the year, month and mode
// over a month of days (or a year of months), choosing *value. It
// reports a pick.
func Calendar(c *ui.Context, value *time.Time, props ...CalendarProps) bool {
	p := first(props)
	t := Of(c)
	if value.IsZero() {
		*value = dayOf(c.Now())
	}
	box := ui.Column(c).Background(t.Surface)
	mode := p.Mode
	if mode == nil {
		mode = ui.Local(box, "mode", func() CalendarMode { return CalendarMonth })
	}
	today := dayOf(c.Now())
	picked := false
	set := func(d time.Time) {
		if !calendarOff(p, d) || *mode == CalendarYear {
			*value = d
			picked = true
		}
	}
	if p.Card {
		box.Width(300).Shrink(0).Border(1, t.Split.Alpha(0.6).Over(t.Surface)).Radius(8)
	} else {
		box.FillWidth()
	}
	box.Children(func() {
		if !p.HideHeader {
			calendarHeader(c, t, value, mode, p)
		}
		switch {
		case *mode == CalendarYear && p.Card:
			ui.Grid(c).Columns(3).FillWidth().Padding(8).Children(func() {
				for m := 1; m <= 12; m++ {
					d := time.Date(value.Year(), time.Month(m), 1, 0, 0, 0, 0, value.Location())
					if gridCell(c.Key(m), t, L(c).MonthTitle(m), value.Month() == d.Month(), false, today.Year() == d.Year() && today.Month() == d.Month()) {
						set(time.Date(d.Year(), d.Month(), min(value.Day(), daysIn(d)), 0, 0, 0, 0, d.Location()))
					}
				}
			})
		case p.Card:
			var hover time.Time
			if d, ok := dayGrid(c, t, *value, *value, &hover, false, func(d time.Time) bool { return calendarOff(p, d) }, today); ok {
				set(d)
			}
		case *mode == CalendarYear:
			ui.Column(c).FillWidth().Children(func() {
				for r := 0; r < 4; r++ {
					ui.Row(c.Key(r)).FillWidth().Children(func() {
						for i := 0; i < 3; i++ {
							m := r*3 + i + 1
							d := time.Date(value.Year(), time.Month(m), 1, 0, 0, 0, 0, value.Location())
							on := value.Month() == d.Month()
							now := today.Year() == d.Year() && today.Month() == d.Month()
							if calendarCell(c.Key(m), t, L(c).MonthTitle(m), d.Format("2006-01"), on, now, false, true, func() {
								if p.MonthCellRender != nil {
									p.MonthCellRender(d)
								}
							}) {
								set(time.Date(d.Year(), d.Month(), min(value.Day(), daysIn(d)), 0, 0, 0, 0, d.Location()))
							}
						}
					})
				}
			})
		default:
			grid := monthGrid(*value)
			ui.Column(c).FillWidth().Children(func() {
				ui.Row(c).FillWidth().Children(func() {
					for _, w := range L(c).Weekdays {
						ui.Row(c.Key(w)).Basis(0).Grow(1).Height(36).Padding(0, 16).AlignItems(ui.Center).Justify(ui.End).Children(func() {
							ui.Text(c, w).FontSize(t.FontSize).TextColor(t.Text)
						})
					}
				})
				for r := 0; r < 6; r++ {
					ui.Row(c.Key(r)).FillWidth().Children(func() {
						for i := 0; i < 7; i++ {
							d := grid[r*7+i]
							if calendarCell(c.Key(i), t, pad(int64(d.Day()), 2), d.Format("2006-01-02"), sameDay(d, *value), sameDay(d, today),
								calendarOff(p, d), d.Month() == value.Month(), func() {
									if p.CellRender != nil {
										p.CellRender(d)
									}
								}) {
								set(d)
							}
						}
					})
				}
			})
		}
	})
	return picked
}

func daysIn(month time.Time) int {
	return time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, month.Location()).Day()
}

// calendarHeader is the year and month selects and the month / year
// switch at the top end of a calendar.
func calendarHeader(c *ui.Context, t *Tokens, value *time.Time, mode *CalendarMode, p CalendarProps) {
	size := pick(p.Card, Small, Middle)
	ui.Row(c).FillWidth().Gap(8).Padding(pick[float32](p.Card, 8, 12), pick[float32](p.Card, 8, 0)).Justify(ui.End).AlignItems(ui.Center).Children(func() {
		var years []Option
		for _, y := range calendarYears(value.Year(), p.ValidRange) {
			years = append(years, Option{Value: strconv.Itoa(y), Label: L(c).YearTitle(y)})
		}
		year := strconv.Itoa(value.Year())
		if Select(c, &year, years, SelectProps{Size: size, Width: 96}) {
			y, _ := strconv.Atoi(year)
			*value = time.Date(y, value.Month(), min(value.Day(), daysIn(time.Date(y, value.Month(), 1, 0, 0, 0, 0, value.Location()))), 0, 0, 0, 0, value.Location())
		}
		if *mode == CalendarMonth {
			var months []Option
			for m := 1; m <= 12; m++ {
				months = append(months, Option{Value: strconv.Itoa(m), Label: L(c).MonthTitle(m)})
			}
			month := strconv.Itoa(int(value.Month()))
			if Select(c.Key("month"), &month, months, SelectProps{Size: size, Width: 80}) {
				m, _ := strconv.Atoi(month)
				first := time.Date(value.Year(), time.Month(m), 1, 0, 0, 0, 0, value.Location())
				*value = time.Date(value.Year(), time.Month(m), min(value.Day(), daysIn(first)), 0, 0, 0, 0, value.Location())
			}
		}
		m := pick(*mode == CalendarYear, "year", "month")
		if RadioButtons(c, &m, []Option{{Value: "month", Label: "月"}, {Value: "year", Label: "年"}}, RadioButtonsProps{Size: size, Outline: true}) {
			*mode = pick(m == "year", CalendarYear, CalendarMonth)
		}
	})
}

// calendarCell is a day (or month) of a full calendar: its number at the
// top end over a rule (primary for today) and what content builds; it
// reports a click.
func calendarCell(c *ui.Context, t *Tokens, num, label string, on, now, off, inMonth bool, content func()) bool {
	cell := ui.Column(c).Basis(0).Grow(1).MinWidth(0).Margin(0, 4).Padding(4, 8, 0).Height(116).
		Label(label).Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer)).Disabled(off)
	rule := pick(now, t.Primary, t.Split.Alpha(0.6).Over(t.Surface))
	cell.Draw(func(pa *ui.Painter, r ui.Rect) { pa.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: 2}, rule, 0) })
	switch {
	case off:
		cell.Background(t.DisabledBg)
	case on:
		cell.Background(t.PrimaryBg)
	case cell.Hovered():
		cell.Background(t.FillTertiary)
	}
	fg := t.Text
	switch {
	case off || !inMonth:
		fg = t.TextDisabled
	case on:
		fg = t.Primary
	}
	cell.Children(func() {
		ui.Text(c, num).FontSize(t.FontSize).TextColor(fg).TextAlign(ui.End).FillWidth().LineHeight(24.0 / 14)
		ui.Scroll(c).FillWidth().Grow(1).MinHeight(0).Scrollbars(false).Children(content)
	})
	return cell.Clicked() && !off
}
