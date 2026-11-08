package kit

import (
	"strconv"
	"time"

	"github.com/egoist/mygo/ui"
)

// pickerField is the trigger of the time and range pickers: a field
// look over a button that toggles *open, with an icon at its end.
func pickerField(c *ui.Context, label string, size Size, width float32, off bool, status Status, force visual, open *bool, slot *compactSlot) ui.Element {
	t := Of(c)
	trig := ui.ButtonBase(c).Label(label).Height(t.Height(size)).Padding(t.SpaceXS, t.PaddingInput).Gap(t.SpaceSM).
		Radius(t.RadiusControl).AlignItems(ui.Center).FocusRing(false).Disabled(off).Clip().
		Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
	if width > 0 {
		trig.Width(width).Shrink(0)
	} else {
		trig.FillWidth()
	}
	takeLabelFocus(trig)
	if trig.Clicked() && !off {
		*open = !*open
	}
	st := stateOf(trig, off, force)
	if *open && st != disabled {
		st = focus
	}
	slot.field(trig, t, st, status, fieldLook(trig, t, st, status))
	return trig
}

// pickerLit reports whether a picker's clear × should show: hovered or
// focused, matching Input and Select.
func pickerLit(trig ui.Element) bool {
	return trig.Hovered() || trig.Focused() || trig.FocusWithin()
}

// TimePickerProps are a time picker's options.
type TimePickerProps struct {
	Size        Size
	Disabled    bool
	Placeholder string
	Status      Status
	Width       float32
	// HideSeconds drops the seconds column, for HH:mm.
	HideSeconds bool
	// Use12Hours shows hours 12, 1 … 11 and a 上午 / 下午 column.
	Use12Hours bool
	// HourStep, MinuteStep and SecondStep space the cells; 1 when zero.
	HourStep, MinuteStep, SecondStep int
	// DisabledHours, DisabledMinutes and DisabledSeconds grey out the
	// values they return, for the hour (and minute) chosen.
	DisabledHours   func() []int
	DisabledMinutes func(hour int) []int
	DisabledSeconds func(hour, minute int) []int
	// AllowClear shows a clear button over the clock while there is a
	// time and the field is hovered or focused, as DatePicker.
	AllowClear bool

	force visual
	open  bool
}

// TimePicker is Ant's TimePicker: a field showing the time of day of *at
// (empty while it is zero) over a panel of hour, minute and second
// columns. A pick changes a draft; 确定 writes it, 此刻 writes the time
// now. AllowClear empties *at. It reports a change. The date of *at is
// kept.
func TimePicker(c *ui.Context, at *time.Time, props ...TimePickerProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	placeholder := pick(p.Placeholder == "", L(c).SelectTime, p.Placeholder)
	layout := pick(p.HideSeconds, "15:04", "15:04:05")
	if p.Use12Hours {
		layout = pick(p.HideSeconds, "03:04 PM", "03:04:05 PM")
	}
	slot := compactBegin()
	defer slot.done()
	holder := ui.Row(c).Shrink(0)
	slot.at(holder)
	if p.Width == 0 {
		holder.FillWidth()
	}
	open := ui.Local(holder, "open", func() bool { return p.open })
	draft := ui.Local(holder, "draft", func() time.Time { return *at })
	shown := ui.Local(holder, "shown", func() bool { return false })
	changed := false
	var trig ui.Element
	holder.Children(func() {
		trig = pickerField(c, placeholder, p.Size, p.Width, p.Disabled, p.Status, p.force, open, slot)
		text, fg := placeholder, t.TextPlaceholder
		if !at.IsZero() {
			text, fg = timeText(*at, layout, p.Use12Hours, L(c)), t.Text
		}
		trig.Children(func() {
			ui.Text(c, text).FontSize(t.ControlFont(p.Size)).TextColor(pick(p.Disabled, t.TextDisabled, fg)).SingleLine().Grow(1)
			if p.AllowClear && !at.IsZero() && pickerLit(trig) && !p.Disabled {
				if clearButton(c, t, true) {
					*at = time.Time{}
					changed = true
					*open = false
				}
			} else {
				clockMark(c, t.TextDisabled)
			}
		})
	})
	if !*open {
		*shown = false
		return changed
	}
	if !*shown {
		*draft = *at
		if draft.IsZero() {
			*draft = c.Now().Truncate(time.Hour * 24)
		}
	}
	reveal := !*shown
	*shown = true
	ui.PopoverBase(c, trig, open, func(panel ui.Element) {
		popupLook(panel.MarginY(4).Padding(0), t)
		panel.Children(func() {
			ui.Row(c).Padding(4, 0).Children(func() {
				applyTimeColumns(c, t, draft, p, reveal)
			})
			ui.Row(c).FillWidth().Padding(t.SpaceXS, t.SpaceSM).AlignItems(ui.Center).Draw(topRule(t.BorderSecondary)).Children(func() {
				if Button(c, L(c).Now, ButtonProps{Type: Link, Size: Small}) {
					now := c.Now()
					*draft = withClock(*draft, now.Hour(), snapStep(now.Minute(), p.MinuteStep), pick(p.HideSeconds, 0, snapStep(now.Second(), p.SecondStep)))
					*at, *open, changed = *draft, false, true
				}
				ui.Spacer(c)
				if Button(c, L(c).OK, ButtonProps{Type: Primary, Size: Small}) {
					changed = !draft.Equal(*at)
					*at, *open = *draft, false
				}
			})
		})
	})
	return changed
}

// TimeRangePicker is Ant's TimePicker.RangePicker: a field showing
// *start → *end over two time panels. 确定 writes both. AllowClear
// empties both. DisabledHours / Minutes / Seconds apply to each panel,
// as TimePicker. It reports a change. The dates of *start and *end are
// kept.
func TimeRangePicker(c *ui.Context, start, end *time.Time, props ...TimePickerProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	layout := pick(p.HideSeconds, "15:04", "15:04:05")
	if p.Use12Hours {
		layout = pick(p.HideSeconds, "03:04 PM", "03:04:05 PM")
	}
	hints := [2]string{pick(p.Placeholder == "", L(c).StartTime, p.Placeholder), L(c).EndTime}
	slot := compactBegin()
	defer slot.done()
	holder := ui.Row(c).Shrink(0)
	slot.at(holder)
	if p.Width == 0 {
		holder.FillWidth()
	}
	open := ui.Local(holder, "open", func() bool { return p.open })
	drafts := ui.Local(holder, "drafts", func() [2]time.Time { return [2]time.Time{*start, *end} })
	shown := ui.Local(holder, "shown", func() bool { return false })
	changed := false
	var trig ui.Element
	holder.Children(func() {
		trig = pickerField(c, hints[0]+" → "+hints[1], p.Size, p.Width, p.Disabled, p.Status, p.force, open, slot)
		trig.Children(func() {
			for i, d := range []*time.Time{start, end} {
				text, fg := hints[i], t.TextPlaceholder
				if !d.IsZero() {
					text, fg = timeText(*d, layout, p.Use12Hours, L(c)), t.Text
				}
				ui.Text(c, text).FontSize(t.ControlFont(p.Size)).TextColor(pick(p.Disabled, t.TextDisabled, fg)).SingleLine().Grow(1).Basis(0)
				if i == 0 {
					arrowRight(c, t.TextDisabled)
				}
			}
			if p.AllowClear && (!start.IsZero() || !end.IsZero()) && pickerLit(trig) && !p.Disabled {
				if clearButton(c, t, true) {
					*start, *end = time.Time{}, time.Time{}
					changed = true
					*open = false
				}
			} else {
				clockMark(c, t.TextDisabled)
			}
		})
	})
	if !*open {
		*shown = false
		return changed
	}
	if !*shown {
		*drafts = [2]time.Time{*start, *end}
		now := c.Now().Truncate(time.Hour * 24)
		if drafts[0].IsZero() {
			drafts[0] = now
		}
		if drafts[1].IsZero() {
			drafts[1] = now
		}
	}
	reveal := !*shown
	*shown = true
	ui.PopoverBase(c, trig, open, func(panel ui.Element) {
		popupLook(panel.MarginY(4).Padding(0), t)
		panel.Children(func() {
			ui.Row(c).AlignItems(ui.Stretch).Children(func() {
				for i, label := range []string{L(c).StartTime, L(c).EndTime} {
					i := i
					ui.Column(c.Key(i)).Children(func() {
						ui.Row(c).Height(32).FillWidth().Center().Draw(bottomRule(t.Split)).Children(func() {
							ui.Text(c, label).FontSize(t.FontSize).FontWeight(600).TextColor(t.Text)
						})
						ui.Row(c).Padding(4, 0).Children(func() {
							applyTimeColumns(c, t, &drafts[i], p, reveal)
						})
					})
				}
			})
			ui.Row(c).FillWidth().Padding(t.SpaceXS, t.SpaceSM).AlignItems(ui.Center).Draw(topRule(t.BorderSecondary)).Children(func() {
				ui.Spacer(c)
				if Button(c, L(c).OK, ButtonProps{Type: Primary, Size: Small}) {
					a, b := drafts[0], drafts[1]
					changed = !a.Equal(*start) || !b.Equal(*end)
					*start, *end, *open = a, b, false
				}
			})
		})
	})
	return changed
}

// applyTimeColumns is the hour / minute / second (and 12-hour) columns
// of a time panel; it writes a pick into *draft.
func applyTimeColumns(c *ui.Context, t *Tokens, draft *time.Time, p TimePickerProps, reveal bool) {
	h, m, s := draft.Clock()
	offH := intSet(p.DisabledHours)
	hours := hourCells(p.HourStep)
	if p.Use12Hours {
		hours = hours12(p.HourStep, h >= 12)
	}
	if n, ok := timeColumn(c, t, L(c).Hour, hours, h, reveal, offH); ok {
		*draft = withClock(*draft, n, m, s)
	}
	var offM map[int]bool
	if p.DisabledMinutes != nil {
		offM = intSet(func() []int { return p.DisabledMinutes(h) })
	}
	if n, ok := timeColumn(c, t, L(c).Minute, stepCells(60, p.MinuteStep), m, reveal, offM); ok {
		*draft = withClock(*draft, h, n, s)
	}
	if !p.HideSeconds {
		var offS map[int]bool
		if p.DisabledSeconds != nil {
			offS = intSet(func() []int { return p.DisabledSeconds(h, m) })
		}
		if n, ok := timeColumn(c, t, L(c).Second, stepCells(60, p.SecondStep), s, reveal, offS); ok {
			*draft = withClock(*draft, h, m, n)
		}
	}
	if p.Use12Hours {
		half := []timeCell{{0, L(c).AM}, {12, L(c).PM}}
		if n, ok := timeColumn(c, t, "午", half, pick(h >= 12, 12, 0), false, nil); ok && n != pick(h >= 12, 12, 0) {
			*draft = withClock(*draft, h%12+n, m, s)
		}
	}
}

// withClock is day at h:m:s, in its location.
func withClock(day time.Time, h, m, s int) time.Time {
	y, mo, d := day.Date()
	return time.Date(y, mo, d, h, m, s, 0, day.Location())
}

// timeRows is how many cells a time column shows.
const timeRows = 8

// timeColumn is one scrolling column of a time panel, its cells by
// value, the one at cur highlighted and those in off greyed out; when
// reveal is set it is scrolled so cur is at the top, as Ant's: the cell
// timeRows-1 below it is brought into view, past the end onto blank
// rows. It reports a chosen value.
func timeColumn(c *ui.Context, t *Tokens, unit string, cells []timeCell, cur int, reveal bool, off map[int]bool) (int, bool) {
	chosen, ok := 0, false
	at := 0
	for i, cell := range cells {
		if cell.value == cur {
			at = i
		}
	}
	n := len(cells) + timeRows - 1
	host := ui.Column(c.Key(unit)).Width(56).Height(28 * timeRows).Clip().Draw(rightRule(t.BorderSecondary))
	st := ui.Local(host, "list", func() ui.ListState { return ui.ListState{} })
	if reveal {
		st.ScrollTo(at, ui.Start)
	}
	host.Children(func() {
		ui.List(c, st, n, func(i int) {
			if i >= len(cells) {
				ui.Box(c).FillWidth().Height(28)
				return
			}
			tc := cells[i]
			dis := off[tc.value]
			cell := ui.ButtonBase(c).Label(tc.label + " " + unit).FillWidth().Height(28).Radius(4).Center().
				FocusRing(false).Disabled(dis).Cursor(pick(dis, ui.CursorNotAllowed, ui.CursorPointer))
			on := tc.value == cur
			switch {
			case on:
				cell.Background(t.PrimaryBg)
			case dis:
			case cell.Hovered():
				cell.Background(t.FillTertiary)
			}
			if cell.FocusVisible() {
				focusRing(cell, t)
			}
			cell.Children(func() {
				txt := ui.Text(c, tc.label).FontSize(t.FontSize).TextColor(pick(dis, t.TextDisabled, t.Text))
				if on {
					txt.FontWeight(600)
				}
			})
			if cell.Clicked() && !dis {
				chosen, ok = tc.value, true
			}
		}).FillWidth().Height(28 * timeRows).Scrollbars(false)
	})
	return chosen, ok
}

// rangeTimeProps is the time-column options for one end of a range:
// DisabledTime when set, otherwise the shared DisabledHours / Minutes /
// Seconds.
func rangeTimeProps(p RangePickerProps, date time.Time, end bool) TimePickerProps {
	if p.DisabledTime != nil {
		return p.DisabledTime(date, end)
	}
	return TimePickerProps{
		DisabledHours:   p.DisabledHours,
		DisabledMinutes: p.DisabledMinutes,
		DisabledSeconds: p.DisabledSeconds,
	}
}

// intSet is what fn returns as a set; empty for nil.
func intSet(fn func() []int) map[int]bool {
	out := map[int]bool{}
	if fn == nil {
		return out
	}
	for _, v := range fn() {
		out[v] = true
	}
	return out
}

// timeText is at laid out, with 上午 / 下午 for a 12-hour clock.
func timeText(at time.Time, layout string, h12 bool, l *Locale) string {
	s := at.Format(layout)
	if !h12 {
		return s
	}
	if at.Hour() < 12 {
		return l.AM + " " + s[:len(s)-3]
	}
	return l.PM + " " + s[:len(s)-3]
}

// RangePickerProps are a range picker's options.
type RangePickerProps struct {
	Size     Size
	Disabled bool
	// Placeholders are the start and end hints; 开始日期 and 结束日期 by
	// default.
	Placeholders [2]string
	Status       Status
	Width        float32
	// Format lays out the dates; "2006-01-02" by default ("2006-01-02
	// 15:04:05" with ShowTime).
	Format string
	// ShowTime adds hour, minute and second columns for the start and
	// end, and an OK button: picking two days does not close until OK.
	ShowTime bool
	// DisabledDate greys out the days it reports, as DatePicker.
	DisabledDate func(time.Time) bool
	// AllowClear shows a clear button over the calendar while there is a
	// range and the field is hovered or focused, as DatePicker.
	AllowClear bool
	// DisabledHours, DisabledMinutes and DisabledSeconds grey out times
	// when ShowTime is on, as TimePicker. DisabledTime, when set, is Ant's
	// disabledTime and is used instead for each end (end is true for the
	// finish).
	DisabledHours   func() []int
	DisabledMinutes func(hour int) []int
	DisabledSeconds func(hour, minute int) []int
	DisabledTime    func(date time.Time, end bool) TimePickerProps
	// Presets are named ranges offered beside the months.
	Presets []RangePreset

	force visual
	open  bool
}

// RangePreset is a named range a RangePicker offers.
type RangePreset struct {
	Label      string
	Start, End time.Time
}

// RangePicker is Ant's RangePicker: a field showing *start → *end over
// two months side by side. The first click picks the start, the second
// the end (the two swap when the end is earlier), which writes both and
// closes; the days between are shaded. DisabledDate greys out days as
// DatePicker; AllowClear empties both. ShowTime's columns take
// DisabledHours / Minutes / Seconds, or DisabledTime per end. It reports
// a change.
func RangePicker(c *ui.Context, start, end *time.Time, props ...RangePickerProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	format := p.Format
	if format == "" {
		format = pick(p.ShowTime, "2006-01-02 15:04:05", "2006-01-02")
	}
	hints := [2]string{pick(p.Placeholders[0] == "", L(c).StartDate, p.Placeholders[0]), pick(p.Placeholders[1] == "", L(c).EndDate, p.Placeholders[1])}
	slot := compactBegin()
	defer slot.done()
	holder := ui.Row(c).Shrink(0)
	slot.at(holder)
	if p.Width == 0 {
		holder.FillWidth()
	}
	open := ui.Local(holder, "open", func() bool { return p.open })
	pending := ui.Local(holder, "pending", func() time.Time { return time.Time{} })
	month := ui.Local(holder, "month", func() time.Time { return time.Time{} })
	drafts := ui.Local(holder, "drafts", func() [2]time.Time { return [2]time.Time{*start, *end} })
	picked := ui.Local(holder, "picked", func() bool { return false })
	changed := false
	var trig ui.Element
	holder.Children(func() {
		trig = pickerField(c, hints[0]+" → "+hints[1], p.Size, p.Width, p.Disabled, p.Status, p.force, open, slot)
		trig.Children(func() {
			for i, d := range []*time.Time{start, end} {
				text, fg := hints[i], t.TextPlaceholder
				if !d.IsZero() {
					text, fg = d.Format(format), t.Text
				}
				ui.Text(c, text).FontSize(t.ControlFont(p.Size)).TextColor(pick(p.Disabled, t.TextDisabled, fg)).SingleLine().Grow(1).Basis(0)
				if i == 0 {
					arrowRight(c, t.TextDisabled)
				}
			}
			if p.AllowClear && (!start.IsZero() || !end.IsZero()) && pickerLit(trig) && !p.Disabled {
				if clearButton(c, t, true) {
					*start, *end = time.Time{}, time.Time{}
					changed = true
					*open = false
				}
			} else {
				calendarMark(c, t.TextDisabled)
			}
		})
	})
	if !*open {
		*pending, *month, *picked = time.Time{}, time.Time{}, false
		return changed
	}
	if month.IsZero() {
		base := *start
		if base.IsZero() {
			base = c.Now()
		}
		*month = time.Date(base.Year(), base.Month(), 1, 0, 0, 0, 0, base.Location())
		*drafts = [2]time.Time{*start, *end}
		*picked = !start.IsZero() && !end.IsZero()
	}
	hover := ui.Local(holder, "hover", func() time.Time { return time.Time{} })
	lo, hi := drafts[0], drafts[1]
	if !p.ShowTime || !*picked {
		lo, hi = *start, *end
	}
	preview := false
	if !pending.IsZero() {
		lo, hi = *pending, *pending
		if !hover.IsZero() {
			lo, hi = orderRange(*pending, *hover)
			preview = true
		}
	}
	hovered := time.Time{}
	ui.PopoverBase(c, trig, open, func(panel ui.Element) {
		popupLook(panel.MarginY(4).Padding(t.SpaceSM, t.SpaceMD), t)
		panel.Children(func() {
			ui.Row(c).Gap(24).AlignItems(ui.Stretch).Children(func() {
				if len(p.Presets) > 0 {
					ui.Column(c).Width(100).Gap(4).Padding(8, 8, 8, 0).Draw(rightRule(t.Split)).Children(func() {
						for i, pr := range p.Presets {
							if presetButton(c.Key(i), t, pr.Label) {
								a, b := orderRange(dayOf(pr.Start), dayOf(pr.End))
								changed = !a.Equal(*start) || !b.Equal(*end)
								*start, *end, *pending, *open = a, b, time.Time{}, false
							}
						}
					})
				}
				for i := 0; i < 2; i++ {
					m := month.AddDate(0, i, 0)
					day, ok := monthPanel(c.Key(i), t, m, lo, hi, c.Now(), i == 0, i == 1, month, preview, p.DisabledDate, &hovered)
					if !ok {
						continue
					}
					if pending.IsZero() {
						*pending = day
						*picked = false
						continue
					}
					a, b := orderRange(*pending, day)
					ah, am, as := 0, 0, 0
					bh, bm, bs := 0, 0, 0
					if p.ShowTime {
						if !drafts[0].IsZero() {
							ah, am, as = drafts[0].Clock()
						} else if !start.IsZero() {
							ah, am, as = start.Clock()
						}
						if !drafts[1].IsZero() {
							bh, bm, bs = drafts[1].Clock()
						} else if !end.IsZero() {
							bh, bm, bs = end.Clock()
						}
					}
					a, b = withClock(a, ah, am, as), withClock(b, bh, bm, bs)
					*drafts = [2]time.Time{a, b}
					*pending, *picked = time.Time{}, true
					if !p.ShowTime {
						changed = !a.Equal(*start) || !b.Equal(*end)
						*start, *end, *open = a, b, false
					}
				}
			})
			if p.ShowTime {
				ui.Row(c).Gap(24).AlignItems(ui.Stretch).Margin(8, 0, 0).Draw(topRule(t.BorderSecondary)).Children(func() {
					for i, label := range []string{L(c).StartTime, L(c).EndTime} {
						i := i
						ui.Column(c.Key(i)).Children(func() {
							ui.Row(c).Height(32).FillWidth().Center().Children(func() {
								ui.Text(c, label).FontSize(t.FontSize).TextColor(t.TextTertiary)
							})
							ui.Row(c).Padding(4, 0).Children(func() {
								d := drafts[i]
								if d.IsZero() {
									d = pick(i == 1, *end, *start)
								}
								applyTimeColumns(c, t, &drafts[i], rangeTimeProps(p, d, i == 1), false)
							})
						})
					}
				})
				ui.Row(c).FillWidth().Padding(t.SpaceXS, t.SpaceSM).AlignItems(ui.Center).Draw(topRule(t.BorderSecondary)).Children(func() {
					ui.Spacer(c)
					if Button(c, L(c).OK, ButtonProps{Type: Primary, Size: Small, Disabled: !*picked && (start.IsZero() || end.IsZero())}) {
						a, b := drafts[0], drafts[1]
						if a.IsZero() {
							a = *start
						}
						if b.IsZero() {
							b = *end
						}
						changed = !a.Equal(*start) || !b.Equal(*end)
						*start, *end, *open = a, b, false
					}
				})
			}
		})
	})
	*hover = hovered
	return changed
}

// orderRange returns a and b earliest first.
func orderRange(a, b time.Time) (time.Time, time.Time) {
	if b.Before(a) {
		return b, a
	}
	return a, b
}

// monthGrid is the 42 days of a month page starting on Monday, as zh-CN
// calendars: the tail of the month before, the month, the head of the
// next.
func monthGrid(month time.Time) [42]time.Time {
	first := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	lead := (int(first.Weekday()) + 6) % 7
	var out [42]time.Time
	for i := range out {
		out[i] = first.AddDate(0, 0, i-lead)
	}
	return out
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// inRange reports whether day falls within lo..hi, by date.
func inRange(day, lo, hi time.Time) bool {
	if lo.IsZero() || hi.IsZero() {
		return false
	}
	d := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	l := time.Date(lo.Year(), lo.Month(), lo.Day(), 0, 0, 0, 0, day.Location())
	h := time.Date(hi.Year(), hi.Month(), hi.Day(), 0, 0, 0, 0, day.Location())
	return !d.Before(l) && !d.After(h)
}

// monthPanel is one month of a range picker with lo..hi shaded and today
// outlined; prev and next put the month arrows on it, moving *nav. It
// reports a chosen day.
// With preview, lo..hi is the range the pointer would pick, dashed. hover
// gets the day under the pointer.
func monthPanel(c *ui.Context, t *Tokens, month, lo, hi, today time.Time, prev, next bool, nav *time.Time, preview bool, off func(time.Time) bool, hover *time.Time) (time.Time, bool) {
	var chosen time.Time
	ok := false
	ui.Column(c).Width(252).Gap(4).Children(func() {
		ui.Row(c).FillWidth().Height(40).AlignItems(ui.Center).Draw(bottomRule(t.BorderSecondary)).Children(func() {
			if prev {
				if navArrow(c, t, L(c).PrevMonth, -1) {
					*nav = nav.AddDate(0, -1, 0)
				}
			} else {
				ui.Box(c).Size(28, 28)
			}
			ui.Text(c, L(c).YearMonth(month.Year(), int(month.Month()))).FontWeight(600).
				FontSize(t.FontSize).TextColor(t.Text).Grow(1).TextAlign(ui.Center)
			if next {
				if navArrow(c, t, L(c).NextMonth, 1) {
					*nav = nav.AddDate(0, 1, 0)
				}
			} else {
				ui.Box(c).Size(28, 28)
			}
		})
		ui.Grid(c).Columns(7).FillWidth().FocusGroup(ui.Horizontal | ui.Vertical).Children(func() {
			for _, w := range L(c).Weekdays {
				ui.Box(c).Height(32).Center().Children(func() {
					ui.Text(c, w).FontSize(t.FontSize).TextColor(t.Text)
				})
			}
			for i, day := range monthGrid(month) {
				inMonth := day.Month() == month.Month()
				edge := inMonth && (sameDay(day, lo) || sameDay(day, hi)) && !lo.IsZero()
				band := inMonth && inRange(day, lo, hi)
				lane := ui.Box(c.Key(i)).Height(32).Center()
				if band && !sameDay(lo, hi) {
					if preview {
						dash := t.PrimaryBorder
						first, last := sameDay(day, lo), sameDay(day, hi)
						lane.Draw(func(pa *ui.Painter, r ui.Rect) {
							var path ui.Path
							path.MoveTo(r.X, r.Y+4).LineTo(r.X+r.W, r.Y+4).MoveTo(r.X, r.Y+r.H-4).LineTo(r.X+r.W, r.Y+r.H-4)
							if first {
								path.MoveTo(r.X+2, r.Y+4).LineTo(r.X+2, r.Y+r.H-4)
							}
							if last {
								path.MoveTo(r.X+r.W-2, r.Y+4).LineTo(r.X+r.W-2, r.Y+r.H-4)
							}
							pa.StrokePath(path.Dash(0, 3, 3), 1, dash)
						})
					} else {
						lane.Background(t.PrimaryBg)
					}
				}
				d := day
				dis := off != nil && off(d)
				lane.Children(func() {
					cell := ui.ButtonBase(c).Label(d.Format("2006-01-02")).Size(24, 24).Radius(4).Center().
						FocusRing(false).Disabled(dis).Cursor(pick(dis, ui.CursorNotAllowed, ui.CursorPointer))
					fg := pick(inMonth && !dis, t.Text, t.TextDisabled)
					switch {
					case edge:
						cell.Background(t.Primary)
						fg = t.TextOnPrimary
					case dis && !band:
						lane.Background(t.DisabledBg)
					case cell.Hovered() && !band:
						cell.Background(t.FillTertiary)
					}
					if inMonth && sameDay(d, today) && !edge {
						cell.Border(1, t.Primary)
					}
					if cell.Hovered() && inMonth {
						*hover = d
					}
					if cell.FocusVisible() {
						focusRing(cell, t)
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
	})
	return chosen, ok
}

func navArrow(c *ui.Context, t *Tokens, label string, dir int) bool {
	b := ui.ButtonBase(c).Label(label).Size(28, 28).Radius(4).Center().FocusRing(false).Cursor(ui.CursorPointer)
	fg := t.TextTertiary
	if b.Hovered() {
		fg = t.Text
	}
	if b.FocusVisible() {
		focusRing(b, t)
	}
	b.Children(func() {
		ui.Box(c).Size(10, 10).Draw(func(pa *ui.Painter, r ui.Rect) {
			var path ui.Path
			x0, x1 := r.X+r.W*0.68, r.X+r.W*0.32
			if dir > 0 {
				x0, x1 = x1, x0
			}
			path.MoveTo(x0, r.Y+r.H*0.15).LineTo(x1, r.Y+r.H*0.5).LineTo(x0, r.Y+r.H*0.85)
			pa.StrokePath(&path, 1.4, fg)
		})
	})
	return b.Clicked()
}

func clockMark(c *ui.Context, color ui.Color) {
	ui.Box(c).Size(14, 14).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		p.Stroke(ui.Rect{X: r.X + 1, Y: r.Y + 1, W: r.W - 2, H: r.H - 2}, color, (r.W-2)/2, 1.3)
		var hands ui.Path
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		hands.MoveTo(cx, cy-3.5).LineTo(cx, cy).LineTo(cx+2.8, cy+1.6)
		p.StrokePath(&hands, 1.3, color)
	})
}

func arrowRight(c *ui.Context, color ui.Color) {
	ui.Box(c).Size(16, 14).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) {
		var path ui.Path
		cy := r.Y + r.H/2
		path.MoveTo(r.X+2, cy).LineTo(r.X+r.W-2, cy).MoveTo(r.X+r.W-6, cy-3.5).LineTo(r.X+r.W-2, cy).LineTo(r.X+r.W-6, cy+3.5)
		p.StrokePath(&path, 1.2, color)
	})
}

// topRule draws a 1px split line along the top edge.
func topRule(split ui.Color) func(*ui.Painter, ui.Rect) {
	return func(p *ui.Painter, r ui.Rect) {
		p.Fill(ui.Rect{X: r.X, Y: r.Y, W: r.W, H: 1}, split, 0)
	}
}
