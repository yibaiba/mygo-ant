package kit

import (
	"fmt"
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// PickerMode is what a DatePicker picks, as Ant's picker prop.
type PickerMode uint8

// Picker modes.
const (
	PickDate PickerMode = iota
	PickWeek
	PickMonth
	PickQuarter
	PickYear
)

// DatePreset is a named date a picker offers beside its panel.
type DatePreset struct {
	Label string
	Value time.Time
}

// dayOf is t at midnight, in its location.
func dayOf(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// weekStart is the Monday of t's week, as zh-CN weeks start.
func weekStart(t time.Time) time.Time {
	d := dayOf(t)
	return d.AddDate(0, 0, -((int(d.Weekday()) + 6) % 7))
}

// quarterOf is t's quarter, 1 to 4.
func quarterOf(t time.Time) int { return (int(t.Month())-1)/3 + 1 }

// decadeOf is the first year of t's decade.
func decadeOf(year int) int { return year - ((year%10)+10)%10 }

// pickerText is t as a picker of mode shows it.
func pickerText(t time.Time, mode PickerMode, format string) string {
	switch mode {
	case PickWeek:
		y, w := t.ISOWeek()
		return fmt.Sprintf("%d-%02d周", y, w)
	case PickMonth:
		return t.Format("2006-01")
	case PickQuarter:
		return fmt.Sprintf("%d-Q%d", t.Year(), quarterOf(t))
	case PickYear:
		return strconv.Itoa(t.Year())
	}
	return t.Format(format)
}

// normalizePick is what a pick of t stands for in mode: the day, the
// Monday of its week, the first of its month or quarter, or New Year.
func normalizePick(t time.Time, mode PickerMode) time.Time {
	switch mode {
	case PickWeek:
		return weekStart(t)
	case PickMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	case PickQuarter:
		return time.Date(t.Year(), time.Month((quarterOf(t)-1)*3+1), 1, 0, 0, 0, 0, t.Location())
	case PickYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	}
	return dayOf(t)
}

// DatePickerProps are a date picker's options.
type DatePickerProps struct {
	Size        Size
	Disabled    bool
	Placeholder string
	Status      Status
	Width       float32
	// Format lays out the date; "2006-01-02" by default ("2006-01-02
	// 15:04:05" with ShowTime).
	Format string
	// Picker picks a week, month, quarter or year instead of a day.
	Picker  PickerMode
	Presets []DatePreset
	// ShowTime adds hour, minute and second columns beside the days and
	// an OK button.
	ShowTime bool
	// DisabledDate greys out the days (months, years) it reports.
	DisabledDate func(time.Time) bool
	AllowClear   bool

	force visual
	open  bool
}

// panelView is which grid a date panel shows.
type panelView uint8

const (
	viewDays panelView = iota
	viewMonths
	viewQuarters
	viewYears
)

func viewFor(mode PickerMode) panelView {
	switch mode {
	case PickMonth:
		return viewMonths
	case PickQuarter:
		return viewQuarters
	case PickYear:
		return viewYears
	}
	return viewDays
}

// DatePicker is Ant's DatePicker: a field showing *date (empty while it
// is zero) over a panel of days, weeks, months, quarters or years, with
// presets beside it and, with ShowTime, time columns. The header's year
// and month open the year and month grids. It reports a change.
func DatePicker(c *ui.Context, date *time.Time, props ...DatePickerProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	format := p.Format
	if format == "" {
		format = pick(p.ShowTime, "2006-01-02 15:04:05", "2006-01-02")
	}
	placeholder := p.Placeholder
	if placeholder == "" {
		l := L(c)
		placeholder = [...]string{l.SelectDate, l.SelectWeek, l.SelectMonth, l.SelectQuarter, l.SelectYear}[p.Picker]
	}
	slot := compactBegin()
	defer slot.done()
	holder := ui.Row(c).Shrink(0)
	slot.at(holder)
	if p.Width == 0 {
		holder.FillWidth()
	}
	open := ui.Local(holder, "open", func() bool { return p.open })
	view := ui.Local(holder, "view", func() panelView { return viewFor(p.Picker) })
	page := ui.Local(holder, "page", func() time.Time { return time.Time{} })
	draft := ui.Local(holder, "draft", func() time.Time { return time.Time{} })
	hoverDay := ui.Local(holder, "hover", func() time.Time { return time.Time{} })
	changed := false
	var trig ui.Element
	holder.Children(func() {
		trig = pickerField(c, placeholder, p.Size, p.Width, p.Disabled, p.Status, p.force, open, slot)
		text, fg := placeholder, t.TextPlaceholder
		if !date.IsZero() {
			text, fg = pickerText(*date, p.Picker, format), t.Text
		}
		trig.Children(func() {
			ui.Text(c, text).FontSize(t.ControlFont(p.Size)).TextColor(pick(p.Disabled, t.TextDisabled, fg)).SingleLine().Grow(1)
			if p.AllowClear && !date.IsZero() && pickerLit(trig) && !p.Disabled {
				if clearButton(c, t, true) {
					*date = time.Time{}
					changed = true
					*open = false
				}
			} else {
				calendarMark(c, t.TextDisabled)
			}
		})
	})
	if !*open {
		*page, *draft = time.Time{}, time.Time{}
		*view = viewFor(p.Picker)
		return changed
	}
	reveal := page.IsZero()
	if reveal {
		base := *date
		if base.IsZero() {
			base = c.Now()
		}
		*page = time.Date(base.Year(), base.Month(), 1, 0, 0, 0, 0, base.Location())
		*draft = *date
		if draft.IsZero() && p.ShowTime {
			*draft = dayOf(c.Now())
		}
	}
	commit := func(v time.Time) {
		if !v.Equal(*date) {
			*date = v
			changed = true
		}
		*open = false
	}
	ui.PopoverBase(c, trig, open, func(wrap ui.Element) {
		wrap.Opacity(enterIn(wrap, "in", motionFor(c, 120*time.Millisecond)))
		placePanel(c, wrap, trig, panelPlace{Side: PlaceBottom, Align: AlignFirst, Gap: 4}, t.Elevated, func(panel ui.Element) {
			popupLook(panel.Padding(0).Gap(0), t)
			panel.Children(func() {
				ui.Row(c).AlignItems(ui.Stretch).Children(func() {
					if len(p.Presets) > 0 {
						ui.Column(c).Width(120).Padding(8).Gap(4).Draw(rightRule(t.Split)).Children(func() {
							for i, pr := range p.Presets {
								if presetButton(c.Key(i), t, pr.Label) {
									commit(normalizePick(pr.Value, p.Picker))
								}
							}
						})
					}
					ui.Column(c).Children(func() {
						ui.Row(c).AlignItems(ui.Stretch).Children(func() {
							picked, ok := datePanel(c, t, view, page, *draft, hoverDay, p.Picker, p.DisabledDate, c.Now())
							if ok {
								if p.ShowTime {
									h, m, s := draft.Clock()
									*draft = withClock(picked, h, m, s)
								} else {
									commit(normalizePick(picked, p.Picker))
								}
							}
							if p.ShowTime && *view == viewDays {
								ui.Column(c).Draw(func(pa *ui.Painter, r ui.Rect) { pa.Fill(ui.Rect{X: r.X, Y: r.Y, W: 1, H: r.H}, t.Split, 0) }).Children(func() {
									ui.Row(c).Height(40).FillWidth().Center().Draw(bottomRule(t.Split)).Children(func() {
										ui.Text(c, draft.Format("15:04:05")).FontWeight(600).FontSize(t.FontSize).TextColor(t.Text)
									})
									ui.Row(c).Padding(4, 0).Children(func() {
										h, m, s := draft.Clock()
										if n, ok := timeColumn(c, t, L(c).Hour, hourCells(1), h, reveal, nil); ok {
											*draft = withClock(*draft, n, m, s)
										}
										if n, ok := timeColumn(c, t, L(c).Minute, stepCells(60, 1), m, reveal, nil); ok {
											*draft = withClock(*draft, h, n, s)
										}
										if n, ok := timeColumn(c, t, L(c).Second, stepCells(60, 1), s, reveal, nil); ok {
											*draft = withClock(*draft, h, m, n)
										}
									})
								})
							}
						})
						if p.ShowTime {
							ui.Row(c).FillWidth().Padding(t.SpaceXS, t.SpaceSM).AlignItems(ui.Center).Draw(topRule(t.BorderSecondary)).Children(func() {
								if Button(c, L(c).Now, ButtonProps{Type: Link, Size: Small}) {
									commit(c.Now().Truncate(time.Second))
								}
								ui.Spacer(c)
								if Button(c, L(c).OK, ButtonProps{Type: Primary, Size: Small, Disabled: draft.IsZero()}) {
									commit(*draft)
								}
							})
						} else if p.Picker == PickDate {
							ui.Row(c).FillWidth().Height(40).Center().Draw(topRule(t.BorderSecondary)).Children(func() {
								if Button(c, L(c).Today, ButtonProps{Type: Link, Size: Small}) {
									commit(dayOf(c.Now()))
								}
							})
						}
					})
				})
			})
		})
	})
	return changed
}

// presetButton is an entry of a picker's preset list.
func presetButton(c *ui.Context, t *Tokens, label string) bool {
	b := ui.ButtonBase(c).Label(label).FillWidth().Height(24).Padding(0, 8).Radius(4).AlignItems(ui.Center).
		Justify(ui.Start).FocusRing(false).Cursor(ui.CursorPointer)
	if b.Hovered() {
		b.Background(t.PrimaryBg)
	}
	b.Children(func() {
		ui.Text(c, label).FontSize(t.FontSize).TextColor(pick(b.Hovered(), t.Primary, t.Text)).SingleLine()
	})
	return b.Clicked()
}

// datePanel is the grid of a DatePicker: days (or weeks), months,
// quarters or the years of a decade, with a header moving page. The
// header's year and month switch to the year and month grids, whose
// picks come back to the mode's own grid. It reports a pick in mode.
func datePanel(c *ui.Context, t *Tokens, view *panelView, page *time.Time, cur time.Time, hover *time.Time, mode PickerMode, off func(time.Time) bool, today time.Time) (time.Time, bool) {
	var picked time.Time
	ok := false
	own := viewFor(mode)
	ui.Column(c).Width(280).Children(func() {
		ui.Row(c).FillWidth().Height(40).Padding(0, 8).AlignItems(ui.Center).Draw(bottomRule(t.Split)).Children(func() {
			jump := map[panelView][2]int{viewDays: {1, 0}, viewMonths: {1, 0}, viewQuarters: {1, 0}, viewYears: {10, 0}}[*view]
			if navArrow(c, t, L(c).PrevYear, -1) {
				*page = page.AddDate(-jump[0], 0, 0)
			}
			if *view == viewDays && navArrow(c.Key("m"), t, L(c).PrevMonth, -1) {
				*page = page.AddDate(0, -1, 0)
			}
			ui.Row(c).Grow(1).Justify(ui.Center).Gap(4).Children(func() {
				switch *view {
				case viewYears:
					d := decadeOf(page.Year())
					ui.Text(c, fmt.Sprintf("%d-%d", d, d+9)).FontWeight(600).FontSize(t.FontSize).TextColor(t.Text)
				default:
					if headerLink(c, t, L(c).YearTitle(page.Year())) {
						*view = viewYears
					}
					if *view == viewDays && headerLink(c.Key("mo"), t, L(c).MonthTitle(int(page.Month()))) {
						*view = viewMonths
					}
				}
			})
			if *view == viewDays && navArrow(c.Key("m+"), t, L(c).NextMonth, 1) {
				*page = page.AddDate(0, 1, 0)
			}
			if navArrow(c.Key("y+"), t, L(c).NextYear, 1) {
				*page = page.AddDate(jump[0], 0, 0)
			}
		})
		disabled := func(d time.Time) bool { return off != nil && off(d) }
		switch *view {
		case viewDays:
			if d, k := dayGrid(c, t, *page, cur, hover, mode == PickWeek, disabled, today); k {
				picked, ok = d, true
			}
		case viewMonths:
			ui.Grid(c).Columns(3).FillWidth().Padding(8).Children(func() {
				for m := 1; m <= 12; m++ {
					d := time.Date(page.Year(), time.Month(m), 1, 0, 0, 0, 0, page.Location())
					on := !cur.IsZero() && cur.Year() == d.Year() && cur.Month() == d.Month()
					if gridCell(c.Key(m), t, L(c).MonthTitle(m), on, disabled(d), today.Year() == d.Year() && today.Month() == d.Month()) {
						if own == viewMonths {
							picked, ok = d, true
						} else {
							*page = d
							*view = own
						}
					}
				}
			})
		case viewQuarters:
			ui.Row(c).FillWidth().Padding(8).Children(func() {
				for q := 1; q <= 4; q++ {
					d := time.Date(page.Year(), time.Month((q-1)*3+1), 1, 0, 0, 0, 0, page.Location())
					on := !cur.IsZero() && cur.Year() == d.Year() && quarterOf(cur) == q
					if gridCell(c.Key(q), t, "Q"+strconv.Itoa(q), on, disabled(d), today.Year() == d.Year() && quarterOf(today) == q) {
						picked, ok = d, true
					}
				}
			})
		case viewYears:
			ui.Grid(c).Columns(3).FillWidth().Padding(8).Children(func() {
				d0 := decadeOf(page.Year())
				for y := d0 - 1; y <= d0+10; y++ {
					d := time.Date(y, 1, 1, 0, 0, 0, 0, page.Location())
					on := !cur.IsZero() && cur.Year() == y
					cell := gridCell(c.Key(y), t, strconv.Itoa(y), on, disabled(d), today.Year() == y)
					if !cell {
						continue
					}
					if own == viewYears {
						picked, ok = d, true
						continue
					}
					*page = time.Date(y, page.Month(), 1, 0, 0, 0, 0, page.Location())
					*view = pick(own == viewQuarters, viewQuarters, viewMonths)
					if own == viewDays {
						*view = viewMonths
					}
				}
			})
		}
	})
	return picked, ok
}

// headerLink is a year or month in a panel's header, opening its grid.
func headerLink(c *ui.Context, t *Tokens, label string) bool {
	b := ui.ButtonBase(c).Label(label).FocusRing(false).Cursor(ui.CursorPointer)
	b.Children(func() {
		ui.Text(c, label).FontWeight(600).FontSize(t.FontSize).TextColor(pick(b.Hovered(), t.Primary, t.Text))
	})
	return b.Clicked()
}

// gridCell is a month, quarter or year of a panel grid.
func gridCell(c *ui.Context, t *Tokens, label string, on, off, now bool) bool {
	lane := ui.Box(c).Height(56).Center().Grow(1)
	clicked := false
	lane.Children(func() {
		b := ui.ButtonBase(c).Label(label).Height(24).MinWidth(60).Padding(0, 8).Radius(4).Center().FocusRing(false).
			Disabled(off).Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
		fg := pick(off, t.TextDisabled, t.Text)
		switch {
		case on:
			b.Background(t.Primary)
			fg = t.TextOnPrimary
		case off:
			b.Background(t.DisabledBg)
		case b.Hovered():
			b.Background(t.FillTertiary)
		}
		if now && !on {
			b.Border(1, t.Primary)
		}
		b.Children(func() { ui.Text(c, label).FontSize(t.FontSize).TextColor(fg) })
		clicked = b.Clicked() && !off
	})
	return clicked
}

// dayGrid is a month of days from Monday, cur chosen and today outlined;
// in week mode a row is a week, chosen and hovered whole. hover gets the
// day under the pointer. It reports a chosen day.
func dayGrid(c *ui.Context, t *Tokens, month, cur time.Time, hover *time.Time, week bool, off func(time.Time) bool, today time.Time) (time.Time, bool) {
	var chosen time.Time
	ok := false
	grid := monthGrid(month)
	hovered := time.Time{}
	ui.Column(c).FillWidth().Padding(8, 12).Children(func() {
		ui.Row(c).FillWidth().Children(func() {
			if week {
				ui.Box(c).Width(32).Height(32)
			}
			for _, w := range L(c).Weekdays {
				ui.Box(c.Key(w)).Height(32).Grow(1).Basis(0).Center().Children(func() {
					ui.Text(c, w).FontSize(t.FontSize).TextColor(t.Text)
				})
			}
		})
		days := ui.Column(c).FillWidth().FocusGroup(ui.Horizontal | ui.Vertical)
		days.Children(func() {
			for r := 0; r < 6; r++ {
				row := ui.Row(c.Key(r)).FillWidth().Radius(4)
				mon := grid[r*7]
				inWeek := func(d time.Time) bool { return !d.IsZero() && sameDay(weekStart(d), mon) }
				if week {
					switch {
					case inWeek(cur):
						row.Background(t.Primary)
					case inWeek(*hover):
						row.Background(t.FillTertiary)
					}
				}
				row.Children(func() {
					if week {
						_, wn := mon.ISOWeek()
						ui.Box(c).Width(32).Height(32).Center().Children(func() {
							ui.Text(c, strconv.Itoa(wn)).FontSize(t.FontSizeSM).TextColor(pick(inWeek(cur), t.TextOnPrimary, t.TextTertiary))
						})
					}
					for i := 0; i < 7; i++ {
						d := grid[r*7+i]
						inMonth := d.Month() == month.Month()
						dis := off(d)
						lane := ui.Box(c.Key(i)).Height(32).Grow(1).Basis(0).Center()
						lane.Children(func() {
							cell := ui.ButtonBase(c).Label(d.Format("2006-01-02")).Size(24, 24).Radius(4).Center().
								FocusRing(false).Disabled(dis).Cursor(pick(dis, ui.CursorNotAllowed, ui.CursorPointer))
							fg := pick(inMonth && !dis, t.Text, t.TextDisabled)
							on := !week && !cur.IsZero() && sameDay(d, cur)
							switch {
							case week && inWeek(cur):
								fg = pick(inMonth, t.TextOnPrimary, t.TextOnPrimary.Alpha(0.6))
							case on:
								cell.Background(t.Primary)
								fg = t.TextOnPrimary
							case dis:
								lane.Background(t.DisabledBg)
							case cell.Hovered() && !week:
								cell.Background(t.FillTertiary)
							}
							if inMonth && sameDay(d, today) && !on && !week {
								cell.Border(1, t.Primary)
							}
							if cell.Hovered() {
								hovered = d
							}
							cell.Children(func() {
								ui.Text(c, strconv.Itoa(d.Day())).FontSize(t.FontSize).TextColor(fg)
							})
							if cell.Clicked() && !dis {
								chosen, ok = d, true
							}
						})
					}
				})
			}
		})
	})
	*hover = hovered
	return chosen, ok
}

// TimeCell is a value a time column offers.
type timeCell struct {
	value int
	label string
}

// stepCells is 0 to n-1 by step, two digits each.
func stepCells(n, step int) []timeCell {
	step = max(step, 1)
	var out []timeCell
	for i := 0; i < n; i += step {
		out = append(out, timeCell{i, fmt.Sprintf("%02d", i)})
	}
	return out
}

// hourCells is the hours by step, 00 to 23.
func hourCells(step int) []timeCell { return stepCells(24, step) }

// hours12 is the hours of a 12-hour clock in its order, 12 first, as
// their 24-hour values in the morning (pm adds 12).
func hours12(step int, pm bool) []timeCell {
	step = max(step, 1)
	var out []timeCell
	for i := 0; i < 12; i += step {
		label := i
		if label == 0 {
			label = 12
		}
		out = append(out, timeCell{i + pick(pm, 12, 0), fmt.Sprintf("%02d", label)})
	}
	return out
}

// snapStep is v down to a multiple of step.
func snapStep(v, step int) int {
	if step <= 1 {
		return v
	}
	return v - v%step
}
