package kit

import (
	"strconv"

	"github.com/egoist/mygo/ui"
)

// PaginationProps are a pagination's options.
type PaginationProps struct {
	// Size Small is the table's compact pager.
	Size Size
	// ShowTotal puts "共 N 条" before the pages.
	ShowTotal bool
	Disabled  bool
	// ShowQuickJumper puts Ant's "跳至 [n] 页" field, which Enter or a
	// step commits.
	ShowQuickJumper bool
	// ShowSizeChanger puts a page-size Select; PageSize, when set, is
	// updated (10 / 20 / 50 / 100, or PageSizeOptions).
	ShowSizeChanger bool
	PageSizeOptions []int
	PageSize        *int
	// Simple is "上一页 n/m 下一页".
	Simple bool
}

// PageCount is how many pages total items fill at size per page, at
// least one.
func PageCount(total, size int) int {
	if size <= 0 || total <= 0 {
		return 1
	}
	return (total + size - 1) / size
}

// Page item markers in pageItems: the jumps of Ant's ellipsis items.
const (
	jumpPrev = -1
	jumpNext = -2
)

// pageItems lists the page numbers a pager shows for page cur of pages,
// as Ant's: all of them up to seven, else the first, the last and two
// around cur, with jumpPrev and jumpNext ellipses for the gaps.
func pageItems(cur, pages int) []int {
	cur = max(1, min(cur, pages))
	if pages <= 7 {
		out := make([]int, pages)
		for i := range out {
			out[i] = i + 1
		}
		return out
	}
	left, right := max(2, cur-2), min(pages-1, cur+2)
	if cur <= 4 {
		left, right = 2, 5
	}
	if cur >= pages-3 {
		left, right = pages-4, pages-1
	}
	out := []int{1}
	if left > 2 {
		out = append(out, jumpPrev)
	}
	for i := left; i <= right; i++ {
		out = append(out, i)
	}
	if right < pages-1 {
		out = append(out, jumpNext)
	}
	return append(out, pages)
}

// jumpTarget is where an ellipsis goes: five pages back or on, as Ant.
func jumpTarget(cur, pages, item int) int {
	if item == jumpPrev {
		return max(1, cur-5)
	}
	return min(pages, cur+5)
}

// Pagination is Ant's Pagination: previous and next arrows around the
// page numbers, the current one outlined. A wheel over it turns the page.
// *current is 1-based and kept within the pages; it reports a change.
func Pagination(c *ui.Context, current *int, total, pageSize int, props ...PaginationProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	pages := PageCount(total, pageSize)
	if *current < 1 || *current > pages {
		*current = max(1, min(*current, pages))
	}
	side := pick[float32](p.Size == Small, 24, 32)
	changed := false
	goTo := func(n int) {
		if n != *current {
			*current = n
			changed = true
		}
	}
	row := ui.Row(c).Gap(pick[float32](p.Size == Small, 2, 8)).AlignItems(ui.Center).Shrink(0).Wrap().GapY(4)
	jump := ui.Local(row, "jump", func() string { return "" })
	wheel := ui.Local(row, "wheel", func() int { return 0 })
	if *wheel != 0 && !p.Disabled {
		goTo(max(1, min(pages, *current+*wheel)))
	}
	*wheel = 0
	row.Children(func() {
		if p.ShowTotal {
			ui.Text(c, L(c).Total(total)).FontSize(t.FontSize).TextColor(t.Text).Margin(0, 8, 0, 0)
		}
		if p.ShowSizeChanger {
			pageSizeChanger(c, p, current, &changed)
		}
		if pageButton(c, t, side, L(c).PrevPage, "", *current <= 1 || p.Disabled, false, -1) {
			goTo(*current - 1)
		}
		if p.Simple {
			ui.Text(c, strconv.Itoa(*current)+"/"+strconv.Itoa(pages)).FontSize(t.FontSize).TextColor(t.Text).Margin(0, 4)
		} else {
			for _, item := range pageItems(*current, pages) {
				switch item {
				case jumpPrev, jumpNext:
					label := pick(item == jumpPrev, L(c).JumpPrev, L(c).JumpNext)
					if pageButton(c.Key(item), t, side, label, "•••", p.Disabled, false, 0) {
						goTo(jumpTarget(*current, pages, item))
					}
				default:
					n := strconv.Itoa(item)
					if pageButton(c.Key(item), t, side, L(c).Page(n), n, p.Disabled, item == *current, 0) {
						goTo(item)
					}
				}
			}
		}
		if pageButton(c, t, side, L(c).NextPage, "", *current >= pages || p.Disabled, false, 1) {
			goTo(*current + 1)
		}
		if p.ShowQuickJumper {
			pageJumper(c, t, p, pages, jump, goTo)
		}
	})
	if !p.Disabled {
		row.HandleInput(func(ev ui.InputEvent) bool {
			if ev.Kind != ui.InputScroll || ev.DY == 0 {
				return false
			}
			*wheel = -wheelStep(ev.DY)
			return *wheel != 0
		})
	}
	return changed
}

func pageSizeOptions(p PaginationProps) []int {
	if len(p.PageSizeOptions) > 0 {
		return p.PageSizeOptions
	}
	return []int{10, 20, 50, 100}
}

func pageSizeChanger(c *ui.Context, p PaginationProps, current *int, changed *bool) {
	if p.PageSize == nil {
		return
	}
	opts := pageSizeOptions(p)
	cur := strconv.Itoa(max(*p.PageSize, 1))
	labels := make([]Option, len(opts))
	for i, n := range opts {
		s := strconv.Itoa(n)
		labels[i] = Option{Value: s, Label: s + " " + L(c).PerPage}
	}
	if Select(c, &cur, labels, SelectProps{Size: p.Size, Disabled: p.Disabled, Width: pick[float32](p.Size == Small, 96, 110)}) {
		n, err := strconv.Atoi(cur)
		if err == nil && n > 0 && n != *p.PageSize {
			*p.PageSize = n
			if current != nil {
				*current = 1
			}
			*changed = true
		}
	}
}

func pageJumper(c *ui.Context, t *Tokens, p PaginationProps, pages int, jump *string, goTo func(int)) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Shrink(0).Children(func() {
		ui.Text(c, L(c).JumpTo).FontSize(t.FontSize).TextColor(t.Text)
		if InputNumber(c, jump, InputNumberProps{
			Min: 1, Max: float64(pages), Precision: 0, Size: p.Size, Width: 56, Disabled: p.Disabled,
		}) {
			n, err := strconv.Atoi(*jump)
			if err == nil {
				goTo(n)
			}
		}
		if L(c).PageUnit != "" {
			ui.Text(c, L(c).PageUnit).FontSize(t.FontSize).TextColor(t.Text)
		}
	})
}

// pageButton is one pager item: text, or an arrow pointing dir (-1 back,
// 1 on) when text is empty.
func pageButton(c *ui.Context, t *Tokens, side float32, label, text string, off, on bool, dir int) bool {
	b := ui.ButtonBase(c).Label(label).Height(side).MinWidth(side).Padding(0, 6).Radius(t.RadiusButton).Center().
		Shrink(0).FocusRing(false).Disabled(off).Selected(on).Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
	st := stateOf(b, off, rest)
	fg := t.Text
	// Selected paints ui's accent; every state sets its own background.
	b.Background(ui.Transparent)
	switch {
	case st == disabled:
		fg = t.TextDisabled
		if on {
			b.Background(t.Fill).Border(1, t.BorderButton)
		}
	case on:
		fg = t.Primary
		b.Background(t.Surface).Border(1, t.Primary)
		if st == hover {
			fg = t.PrimaryHover
			b.Border(1, t.PrimaryHover)
		}
	case st == hover:
		b.Background(t.FillSecondary)
	case st == active:
		b.Background(t.Fill)
	}
	if st == focus {
		focusRing(b, t)
	}
	b.Children(func() {
		if text == "" {
			ui.Box(c).Size(10, 10).Draw(func(pa *ui.Painter, r ui.Rect) {
				var path ui.Path
				x0, x1 := r.X+r.W*0.68, r.X+r.W*0.32
				if dir > 0 {
					x0, x1 = x1, x0
				}
				path.MoveTo(x0, r.Y+r.H*0.15).LineTo(x1, r.Y+r.H*0.5).LineTo(x0, r.Y+r.H*0.85)
				pa.StrokePath(&path, 1.4, fg)
			})
			return
		}
		txt := ui.Text(c, text).FontSize(t.FontSize).TextColor(fg).SingleLine()
		if on {
			txt.FontWeight(600)
		}
	})
	return b.Clicked() && !off
}
