package kit

import (
	"slices"
	"strconv"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSortStateCycles(t *testing.T) {
	s := SortState{Col: -1}
	want := []SortState{{0, SortAscend}, {0, SortDescend}, {0, SortNone}, {0, SortAscend}}
	for i, w := range want {
		s = s.next(0)
		if s != w {
			t.Fatalf("click %d: %+v, want %+v", i+1, s, w)
		}
	}
	s = SortState{Col: 0, Order: SortDescend}.next(2)
	if s != (SortState{2, SortAscend}) {
		t.Fatalf("another column: %+v, want ascending on it", s)
	}
}

func TestSortRowsIsStableAndReverses(t *testing.T) {
	keys := []int{3, 1, 2, 1, 3}
	less := ByNumber(func(r int) float64 { return float64(keys[r]) })
	if got := sortRows(5, less, SortNone); !slices.Equal(got, []int{0, 1, 2, 3, 4}) {
		t.Errorf("none = %v", got)
	}
	if got := sortRows(5, less, SortAscend); !slices.Equal(got, []int{1, 3, 2, 0, 4}) {
		t.Errorf("ascend = %v", got)
	}
	if got := sortRows(5, less, SortDescend); !slices.Equal(got, []int{0, 4, 2, 1, 3}) {
		t.Errorf("descend = %v, want ties in their given order", got)
	}
	if got := sortRows(3, nil, SortAscend); !slices.Equal(got, []int{0, 1, 2}) {
		t.Errorf("nil less = %v", got)
	}
	names := []string{"b", "a", "c"}
	if got := sortRows(3, ByText(func(r int) string { return names[r] }), SortAscend); !slices.Equal(got, []int{1, 0, 2}) {
		t.Errorf("by text = %v", got)
	}
}

func TestPageMath(t *testing.T) {
	for _, tc := range []struct{ total, size, want int }{{0, 10, 1}, {1, 10, 1}, {10, 10, 1}, {11, 10, 2}, {95, 10, 10}, {5, 0, 1}} {
		if got := PageCount(tc.total, tc.size); got != tc.want {
			t.Errorf("PageCount(%d, %d) = %d, want %d", tc.total, tc.size, got, tc.want)
		}
	}
	order := []int{4, 3, 2, 1, 0}
	if got := pageRows(order, 2, 2); !slices.Equal(got, []int{2, 1}) {
		t.Errorf("page 2 = %v", got)
	}
	if got := pageRows(order, 3, 2); !slices.Equal(got, []int{0}) {
		t.Errorf("last page = %v", got)
	}
	if got := pageRows(order, 9, 2); got != nil {
		t.Errorf("past the end = %v", got)
	}
	if got := pageRows(order, 1, 0); len(got) != 5 {
		t.Errorf("unpaged = %v", got)
	}
}

func TestPageItems(t *testing.T) {
	cases := []struct {
		cur, pages int
		want       []int
	}{
		{1, 1, []int{1}},
		{3, 7, []int{1, 2, 3, 4, 5, 6, 7}},
		{1, 10, []int{1, 2, 3, 4, 5, jumpNext, 10}},
		{4, 10, []int{1, 2, 3, 4, 5, jumpNext, 10}},
		{5, 10, []int{1, jumpPrev, 3, 4, 5, 6, 7, jumpNext, 10}},
		{7, 10, []int{1, jumpPrev, 6, 7, 8, 9, 10}},
		{10, 10, []int{1, jumpPrev, 6, 7, 8, 9, 10}},
		{50, 20, []int{1, jumpPrev, 16, 17, 18, 19, 20}},
	}
	for _, tc := range cases {
		if got := pageItems(tc.cur, tc.pages); !slices.Equal(got, tc.want) {
			t.Errorf("pageItems(%d, %d) = %v, want %v", tc.cur, tc.pages, got, tc.want)
		}
	}
	if got := jumpTarget(3, 20, jumpPrev); got != 1 {
		t.Errorf("jump back from 3 = %d", got)
	}
	if got := jumpTarget(18, 20, jumpNext); got != 20 {
		t.Errorf("jump on from 18 = %d", got)
	}
}

var sortHosts = []string{"192.168.0.30", "192.168.0.4", "192.168.0.200", "192.168.0.17", "192.168.0.9"}

func sortTable(c *ui.Context, sel *int, size int) {
	Table(c, []Column{
		{Title: "IP", Width: 160, Sorter: ByText(func(r int) string { return sortHosts[r] })},
		{Title: "序号", Width: 80, Sorter: ByNumber(func(r int) float64 { return float64(r) })},
	}, len(sortHosts), func(r, col int) Cell {
		if col == 0 {
			return Cell{Text: sortHosts[r]}
		}
		return Cell{Text: "#" + strconv.Itoa(r)}
	}, TableProps{Selected: sel, PageSize: size})
}

// shownOrder is the hosts on screen, top to bottom.
func shownOrder(tester *ui.Tester) []string {
	type at struct {
		y    float32
		host string
	}
	var seen []at
	for _, h := range sortHosts {
		if r, ok := tester.Find(h); ok {
			seen = append(seen, at{r.Y, h})
		}
	}
	slices.SortFunc(seen, func(a, b at) int { return int(a.y - b.y) })
	out := make([]string, len(seen))
	for i, s := range seen {
		out[i] = s.host
	}
	return out
}

func TestTableSortsOnHeaderClicks(t *testing.T) {
	sel := 1
	tester := ui.NewTester(func(c *ui.Context) { Page(c, Work, func() { sortTable(c, &sel, 0) }) }, 600, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if got := shownOrder(tester); !slices.Equal(got, sortHosts) {
		t.Fatalf("unsorted = %v", got)
	}
	click := func() {
		if err := tester.Click("IP"); err != nil {
			t.Fatal(err)
		}
		tester.Frame()
	}
	click()
	asc := []string{"192.168.0.17", "192.168.0.200", "192.168.0.30", "192.168.0.4", "192.168.0.9"}
	if got := shownOrder(tester); !slices.Equal(got, asc) {
		t.Fatalf("ascending = %v", got)
	}
	click()
	desc := slices.Clone(asc)
	slices.Reverse(desc)
	if got := shownOrder(tester); !slices.Equal(got, desc) {
		t.Fatalf("descending = %v", got)
	}
	click()
	if got := shownOrder(tester); !slices.Equal(got, sortHosts) {
		t.Fatalf("third click = %v, want unsorted", got)
	}
	click()
	if err := tester.Click("192.168.0.9"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if sel != 4 {
		t.Fatalf("selected %d after a click on a sorted row, want its index 4", sel)
	}
}

func TestTablePages(t *testing.T) {
	sel := -1
	tester := ui.NewTester(func(c *ui.Context) { Page(c, Work, func() { sortTable(c, &sel, 2) }) }, 600, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if got := shownOrder(tester); !slices.Equal(got, sortHosts[:2]) {
		t.Fatalf("page 1 = %v", got)
	}
	if err := tester.Click("下一页"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if got := shownOrder(tester); !slices.Equal(got, sortHosts[2:4]) {
		t.Fatalf("page 2 = %v", got)
	}
	if err := tester.Click("第 3 页"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if got := shownOrder(tester); !slices.Equal(got, sortHosts[4:]) {
		t.Fatalf("page 3 = %v", got)
	}
	_ = tester.Click("下一页")
	tester.Frame()
	if got := shownOrder(tester); !slices.Equal(got, sortHosts[4:]) {
		t.Fatalf("next past the last page moved to %v", got)
	}
	if err := tester.Click("IP"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if got := shownOrder(tester); !slices.Equal(got, []string{"192.168.0.17", "192.168.0.200"}) {
		t.Fatalf("sorting did not go back to page 1: %v", got)
	}
}

func TestPaginationStandalone(t *testing.T) {
	cur, changes := 1, 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			if Pagination(c, &cur, 95, 10, PaginationProps{ShowTotal: true}) {
				changes++
			}
		})
	}, 700, 120)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if !tester.HasText("共 95 条") || !tester.HasText("10") {
		t.Fatal("missing total or last page")
	}
	_ = tester.Click("上一页")
	tester.Frame()
	if cur != 1 || changes != 0 {
		t.Fatal("previous on the first page moved")
	}
	if err := tester.Click("向后 5 页"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if cur != 6 || changes != 1 {
		t.Fatalf("jump on = %d after %d changes", cur, changes)
	}
	if err := tester.Click("第 10 页"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if cur != 10 {
		t.Fatalf("last = %d", cur)
	}
	cur = 40
	tester.Frame()
	if cur != 10 {
		t.Fatalf("out of range page kept: %d", cur)
	}
}

func TestTableScrollsSideways(t *testing.T) {
	cols := []Column{{Title: "名称", Width: 200}, {Title: "地址", Width: 200}, {Title: "指纹", Width: 300}}
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Table(c, cols, 1, func(r, col int) Cell { return Cell{Text: []string{"会议室", "172.30.0.9", "末列指纹"}[col]} })
		})
	}, 400, 200)
	tester.SetLocale("zh-CN")
	tester.Frame()
	name, _ := tester.Find("会议室")
	last, ok := tester.Find("末列指纹")
	if ok && last.X < 400 {
		t.Fatalf("last column at x=%v: the columns were squeezed into 400", last.X)
	}
	tester.Scroll(name.X+10, name.Y, 800, 0)
	tester.Frame()
	last, ok = tester.Find("末列指纹")
	if !ok || last.X+last.W > 400 {
		t.Fatalf("scrolling sideways did not bring the last column in: %v %v", last, ok)
	}
}
