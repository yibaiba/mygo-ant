package kit

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// StatisticProps are a statistic's options.
type StatisticProps struct {
	// Precision fixes that many decimals; 0 shows the value as it is.
	Precision int
	// GroupSeparator splits thousands; "," when empty, none when "-".
	GroupSeparator string
	// Prefix and Suffix sit before and after the value; PrefixIcon is an
	// icon before it.
	Prefix, Suffix string
	PrefixIcon     *ui.SVG
	// Color tints the value; Ant's heading color by default.
	Color ui.Color
	// Text shows instead of the number, for values that are not one.
	Text    string
	Loading bool
	// ValueSize is the value's font size; 24 when zero.
	ValueSize float32
}

// formatStatistic splits v into its grouped integer part and its decimal
// part (with the point), as Ant's Statistic formats a number.
func formatStatistic(v float64, precision int, group string) (whole, frac string) {
	s := strconv.FormatFloat(v, 'f', pick(precision > 0, precision, -1), 64)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	whole, frac, _ = strings.Cut(s, ".")
	if frac != "" {
		frac = "." + frac
	}
	if group == "" {
		group = ","
	}
	if group != "-" {
		var b strings.Builder
		for i, r := range whole {
			if i > 0 && (len(whole)-i)%3 == 0 {
				b.WriteString(group)
			}
			b.WriteRune(r)
		}
		whole = b.String()
	}
	if neg {
		whole = "-" + whole
	}
	return whole, frac
}

// Statistic is Ant's Statistic: a grey title over a large value with its
// prefix and suffix.
func Statistic(c *ui.Context, title string, value float64, props ...StatisticProps) {
	p := first(props)
	t := Of(c)
	text := p.Text
	if text == "" {
		whole, frac := formatStatistic(value, p.Precision, p.GroupSeparator)
		text = whole + frac
	}
	statisticBody(c, t, title, text, p)
}

func statisticBody(c *ui.Context, t *Tokens, title, text string, p StatisticProps) {
	tone := p.Color
	if tone == (ui.Color{}) {
		tone = t.TextHeading
	}
	size := pick(p.ValueSize == 0, 24, p.ValueSize)
	ui.Column(c).Gap(4).Shrink(0).Children(func() {
		if title != "" {
			ui.Text(c, title).FontSize(t.FontSize).TextColor(t.TextTertiary)
		}
		if p.Loading {
			ui.Box(c).Width(120).Children(func() { SkeletonBlock(c, 0, size, SkeletonBlockProps{Active: true}) })
			return
		}
		ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
			if p.PrefixIcon != nil {
				ui.Icon(c, p.PrefixIcon).Size(size*0.75, size*0.75).TextColor(tone)
			}
			if p.Prefix != "" {
				ui.Text(c, p.Prefix).FontSize(size).TextColor(tone).SingleLine()
			}
			ui.Text(c, text).FontSize(size).TextColor(tone).SingleLine().FontFeatures("tnum")
			if p.Suffix != "" {
				ui.Text(c, p.Suffix).FontSize(size * 0.66).TextColor(tone).SingleLine()
			}
		})
	})
}

// CountdownProps are a countdown's options.
type CountdownProps struct {
	StatisticProps
	// Format lays the time left out: D days, H hours, m minutes, s
	// seconds, S fractions (one digit per S), text in [brackets] kept as
	// is; "HH:mm:ss" when empty.
	Format string
	// OnFinish runs once when the deadline passes.
	OnFinish func()
}

// formatCountdown lays d out by format as Ant's Countdown does: each run
// of a unit letter is that unit, zero padded to the run's length, the
// biggest unit of the format taking what the smaller ones leave.
func formatCountdown(d time.Duration, format string) string {
	if d < 0 {
		d = 0
	}
	if format == "" {
		format = "HH:mm:ss"
	}
	units := []struct {
		r    byte
		unit time.Duration
	}{{'Y', 365 * 24 * time.Hour}, {'M', 30 * 24 * time.Hour}, {'D', 24 * time.Hour}, {'H', time.Hour}, {'m', time.Minute}, {'s', time.Second}}
	left := d
	vals := map[byte]int64{}
	for _, u := range units {
		if strings.IndexByte(stripBrackets(format), u.r) < 0 {
			continue
		}
		vals[u.r] = int64(left / u.unit)
		left -= time.Duration(vals[u.r]) * u.unit
	}
	var b strings.Builder
	for i := 0; i < len(format); {
		ch := format[i]
		if ch == '[' {
			end := strings.IndexByte(format[i:], ']')
			if end < 0 {
				b.WriteString(format[i+1:])
				break
			}
			b.WriteString(format[i+1 : i+end])
			i += end + 1
			continue
		}
		j := i
		for j < len(format) && format[j] == ch {
			j++
		}
		n := j - i
		switch {
		case ch == 'S':
			ms := float64(left) / float64(time.Second)
			digits := int64(math.Floor(ms * math.Pow10(n)))
			b.WriteString(pad(digits, n))
		case strings.IndexByte("YMDHms", ch) >= 0:
			b.WriteString(pad(vals[ch], n))
		default:
			b.WriteString(format[i:j])
		}
		i = j
	}
	return b.String()
}

func stripBrackets(s string) string {
	var b strings.Builder
	in := false
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '[':
			in = true
		case s[i] == ']':
			in = false
		case !in:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func pad(v int64, n int) string {
	s := strconv.FormatInt(v, 10)
	for len(s) < n {
		s = "0" + s
	}
	return s
}

// Countdown is Ant's Statistic.Countdown: the time left to deadline,
// ticking each second (or each frame with fractions shown). It reports
// whether the deadline has passed.
func Countdown(c *ui.Context, title string, deadline time.Time, props ...CountdownProps) bool {
	p := first(props)
	t := Of(c)
	now := c.Now()
	left := deadline.Sub(now)
	statisticBody(c, t, title, formatCountdown(left, p.Format), p.StatisticProps)
	host := motionHost(c)
	done := ui.Local(host, "countdown.done", func() bool { return false })
	if left <= 0 {
		if !*done {
			*done = true
			if p.OnFinish != nil {
				p.OnFinish()
			}
		}
		return true
	}
	*done = false
	if strings.Contains(stripBrackets(p.Format), "S") {
		c.After(33 * time.Millisecond)
	} else {
		c.After(left - left.Truncate(time.Second) + time.Millisecond)
	}
	return false
}
