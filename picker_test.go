package kit

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.Local) }

func TestMonthGridStartsOnMonday(t *testing.T) {
	g := monthGrid(day(2026, 11, 17))
	if !sameDay(g[0], day(2026, 10, 26)) {
		t.Errorf("Nov 2026 page starts %v, want Mon Oct 26", g[0])
	}
	if !sameDay(g[6], day(2026, 11, 1)) {
		t.Errorf("Nov 1 (a Sunday) is at %v", g[6])
	}
	if !sameDay(g[41], day(2026, 12, 6)) {
		t.Errorf("page ends %v", g[41])
	}
	if g := monthGrid(day(2026, 6, 1)); !sameDay(g[0], day(2026, 6, 1)) {
		t.Errorf("a month starting on Monday leads with %v", g[0])
	}
}

func TestRangeHelpers(t *testing.T) {
	a, b := orderRange(day(2026, 11, 10), day(2026, 11, 3))
	if !sameDay(a, day(2026, 11, 3)) || !sameDay(b, day(2026, 11, 10)) {
		t.Errorf("orderRange = %v, %v", a, b)
	}
	lo, hi := day(2026, 11, 3), day(2026, 11, 10).Add(15*time.Hour)
	for _, tc := range []struct {
		d    time.Time
		want bool
	}{
		{day(2026, 11, 3).Add(23 * time.Hour), true},
		{day(2026, 11, 10), true},
		{day(2026, 11, 2), false},
		{day(2026, 11, 11), false},
	} {
		if got := inRange(tc.d, lo, hi); got != tc.want {
			t.Errorf("inRange(%v) = %v", tc.d, got)
		}
	}
	if inRange(day(2026, 11, 5), time.Time{}, hi) {
		t.Error("an open range shaded a day")
	}
	if got := withClock(day(2026, 11, 5), 8, 30, 15); got.Hour() != 8 || got.Minute() != 30 || got.Second() != 15 || got.Day() != 5 {
		t.Errorf("withClock = %v", got)
	}
}

func TestTimePickerDraftsAndCommits(t *testing.T) {
	at := day(2026, 11, 5).Add(9*time.Hour + 15*time.Minute)
	changes := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if TimePicker(c, &at, TimePickerProps{Width: 160}) {
				changes++
			}
		})
	}, 500, 420)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("09:15:00") {
		t.Fatal("field does not show the time")
	}
	if err := tester.Click("请选择时间"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	tester.Frame()
	hour, _ := tester.Find("09 时")
	minute, _ := tester.Find("15 分")
	if hour.H == 0 || minute.H == 0 || hour.Y != minute.Y {
		t.Fatalf("current values not scrolled to the top together: %v %v", hour, minute)
	}
	for _, label := range []string{"12 时", "20 分"} {
		if err := tester.Click(label); err != nil {
			t.Fatal(err)
		}
		tester.Frame()
	}
	if at.Hour() != 9 || changes != 0 {
		t.Fatal("a pick wrote before 确定")
	}
	if err := tester.Click("确定"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if at.Hour() != 12 || at.Minute() != 20 || at.Day() != 5 || changes != 1 {
		t.Fatalf("committed %v after %d changes", at, changes)
	}
	if tester.HasText("此刻") {
		t.Fatal("panel stayed open")
	}
}

func TestRangePickerPicksTwoDays(t *testing.T) {
	start, end := day(2026, 11, 2), day(2026, 11, 4)
	changes := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if RangePicker(c, &start, &end, RangePickerProps{Width: 280}) {
				changes++
			}
		})
	}, 700, 460)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("开始日期 → 结束日期"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !tester.HasText("2026年11月") || !tester.HasText("2026年12月") {
		t.Fatal("panel does not show November and December")
	}
	if err := tester.Click("2026-11-20"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if changes != 0 || !sameDay(start, day(2026, 11, 2)) {
		t.Fatal("the first click wrote the range")
	}
	if err := tester.Click("2026-11-09"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !sameDay(start, day(2026, 11, 9)) || !sameDay(end, day(2026, 11, 20)) || changes != 1 {
		t.Fatalf("range = %v → %v after %d changes", start, end, changes)
	}
	if !tester.HasText("2026-11-09") || tester.HasText("2026年12月") {
		t.Fatal("field not updated or panel open")
	}
}

func TestRangePickerShowTimeWaitsForOK(t *testing.T) {
	start, end := day(2026, 11, 2), day(2026, 11, 4)
	changes := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if RangePicker(c, &start, &end, RangePickerProps{Width: 360, ShowTime: true}) {
				changes++
			}
		})
	}, 800, 720)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("开始日期 → 结束日期"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("2026-11-20"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("2026-11-09"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if changes != 0 {
		t.Fatal("ShowTime wrote the range before 确定")
	}
	if !tester.HasText("开始时间") || !tester.HasText("结束时间") || !tester.HasText("确定") {
		t.Fatalf("time columns missing: %v", tester.Texts())
	}
	if err := tester.Click("确定"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !sameDay(start, day(2026, 11, 9)) || !sameDay(end, day(2026, 11, 20)) || changes != 1 {
		t.Fatalf("range = %v → %v after %d changes", start, end, changes)
	}
	if tester.HasText("2026年12月") {
		t.Fatal("panel stayed open")
	}
}

func TestRangePickerDisabledDateRejects(t *testing.T) {
	start, end := day(2026, 11, 2), day(2026, 11, 4)
	changes := 0
	off := day(2026, 11, 10)
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if RangePicker(c, &start, &end, RangePickerProps{
				Width:        280,
				DisabledDate: func(d time.Time) bool { return sameDay(d, off) },
			}) {
				changes++
			}
		})
	}, 700, 460)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if err := tester.Click("开始日期 → 结束日期"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("2026-11-10"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if changes != 0 || !sameDay(start, day(2026, 11, 2)) {
		t.Fatal("a disabled day wrote the range")
	}
	if err := tester.Click("2026-11-20"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if changes != 0 || !sameDay(start, day(2026, 11, 2)) {
		t.Fatal("the first valid click wrote the range")
	}
	if err := tester.Click("2026-11-09"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !sameDay(start, day(2026, 11, 9)) || !sameDay(end, day(2026, 11, 20)) || changes != 1 {
		t.Fatalf("range = %v → %v after %d changes", start, end, changes)
	}
}

func TestRangePickerAllowClear(t *testing.T) {
	start, end := day(2026, 11, 2), day(2026, 11, 4)
	changes := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if RangePicker(c, &start, &end, RangePickerProps{Width: 280, AllowClear: true}) {
				changes++
			}
		})
	}, 500, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	r, ok := tester.Find("开始日期 → 结束日期")
	if !ok {
		t.Fatal("no field")
	}
	tester.Move(r.X+r.W-10, r.Y+r.H/2)
	tester.Frame()
	if err := tester.Click("清除"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !start.IsZero() || !end.IsZero() || changes != 1 {
		t.Fatalf("clear left %v → %v after %d", start, end, changes)
	}
}

func TestRangeTimeProps(t *testing.T) {
	hours := func() []int { return []int{1} }
	p := RangePickerProps{DisabledHours: hours}
	if rangeTimeProps(p, time.Time{}, false).DisabledHours == nil {
		t.Fatal("shared hours not forwarded")
	}
	p.DisabledTime = func(date time.Time, end bool) TimePickerProps {
		if !end {
			return TimePickerProps{}
		}
		return TimePickerProps{DisabledHours: func() []int { return []int{23} }}
	}
	if rangeTimeProps(p, time.Time{}, false).DisabledHours != nil {
		t.Fatal("DisabledTime start should replace shared hours")
	}
	if rangeTimeProps(p, time.Time{}, true).DisabledHours == nil {
		t.Fatal("DisabledTime end hours missing")
	}
}

func TestTimeRangePickerAllowClear(t *testing.T) {
	start := day(2026, 11, 5).Add(9 * time.Hour)
	end := day(2026, 11, 5).Add(18 * time.Hour)
	changes := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if TimeRangePicker(c, &start, &end, TimePickerProps{Width: 280, AllowClear: true}) {
				changes++
			}
		})
	}, 500, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	r, ok := tester.Find("开始时间 → 结束时间")
	if !ok {
		t.Fatal("no field")
	}
	tester.Move(r.X+r.W-10, r.Y+r.H/2)
	tester.Frame()
	if err := tester.Click("清除"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !start.IsZero() || !end.IsZero() || changes != 1 {
		t.Fatalf("clear left %v → %v after %d", start, end, changes)
	}
}

func TestTimeRangePickerCommitsBoth(t *testing.T) {
	start := day(2026, 11, 5).Add(9 * time.Hour)
	end := day(2026, 11, 5).Add(18*time.Hour + 30*time.Minute)
	changes := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if TimeRangePicker(c, &start, &end, TimePickerProps{Width: 280}) {
				changes++
			}
		})
	}, 700, 420)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("09:00:00") || !tester.HasText("18:30:00") {
		t.Fatal("field does not show both times")
	}
	if err := tester.Click("开始时间 → 结束时间"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	tester.Frame()
	if !tester.HasText("开始时间") || !tester.HasText("结束时间") {
		t.Fatalf("panels missing: %v", tester.Texts())
	}
	if err := tester.Click("12 时"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if start.Hour() != 9 || changes != 0 {
		t.Fatal("a pick wrote before 确定")
	}
	if err := tester.Click("确定"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if start.Hour() != 12 || end.Hour() != 18 || end.Minute() != 30 || changes != 1 {
		t.Fatalf("committed %v → %v after %d", start, end, changes)
	}
}
