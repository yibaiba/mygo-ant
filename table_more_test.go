package kit

import (
	"reflect"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestTableColumnsOrderAndOffsets(t *testing.T) {
	cols := []Column{
		{Title: "a", Width: 100},
		{Title: "b", Width: 80, Fixed: FixedRight},
		{Title: "c", Width: 120, Fixed: FixedLeft},
		{Title: "d", Fixed: FixedLeft}, // no width: not fixed
		{Title: "e", Width: 60, Fixed: FixedRight},
	}
	tc := tableColumns(cols, true, true)
	var order []int
	for _, c := range tc {
		order = append(order, int(c.kind)*10+c.idx)
	}
	// select, expand, c, a, d, b, e
	if !reflect.DeepEqual(order, []int{10, 20, 2, 0, 3, 1, 4}) {
		t.Fatalf("order = %v", order)
	}
	if tc[0].fixed != FixedLeft || tc[1].fixed != FixedLeft || tc[4].fixed != FixedNone {
		t.Fatalf("sides = %+v", tc)
	}
	off := fixedOffsets(tc)
	if !reflect.DeepEqual(off, []float32{0, prefixW, 2 * prefixW, 0, 0, 60, 0}) {
		t.Fatalf("offsets = %v", off)
	}
	if tc := tableColumns(cols[:1], false, false); len(tc) != 1 || tc[0].fixed != FixedNone {
		t.Fatalf("plain = %+v", tc)
	}
}

func TestFilterRowsTreeAndSelection(t *testing.T) {
	rooms := []string{"A", "B", "A", "C"}
	states := []string{"on", "on", "off", "on"}
	cols := []Column{
		{OnFilter: FilterEquals(func(r int) string { return rooms[r] })},
		{OnFilter: FilterContains(func(r int) string { return states[r] })},
	}
	all := []int{0, 1, 2, 3}
	if got := filterRows(all, cols, TableFilters{}); !reflect.DeepEqual(got, all) {
		t.Fatalf("none = %v", got)
	}
	if got := filterRows(all, cols, TableFilters{0: {"A", "C"}}); !reflect.DeepEqual(got, []int{0, 2, 3}) {
		t.Fatalf("A|C = %v", got)
	}
	if got := filterRows(all, cols, TableFilters{0: {"A", "C"}, 1: {"ON"}}); !reflect.DeepEqual(got, []int{0, 3}) {
		t.Fatalf("A|C & on = %v", got)
	}

	parent := []int{-1, 0, 0, 1, -1, 9}
	rows := []int{0, 1, 2, 3, 4, 5}
	roots, kids := tableTree(rows, func(r int) int { return parent[r] })
	if !reflect.DeepEqual(roots, []int{0, 4, 5}) || !reflect.DeepEqual(kids[0], []int{1, 2}) {
		t.Fatalf("tree = %v %v", roots, kids)
	}
	flat := flattenRows(roots, kids, map[int]bool{0: true})
	if len(flat) != 5 || flat[1].row != 1 || flat[1].depth != 1 || !flat[1].kids || flat[3].row != 4 {
		t.Fatalf("flat = %+v", flat)
	}
	flat = flattenRows(roots, kids, map[int]bool{0: true, 1: true})
	if flat[2].row != 3 || flat[2].depth != 2 {
		t.Fatalf("deep = %+v", flat)
	}

	off := func(r int) bool { return r == 2 }
	if all, some := selectionState([]int{0}, []int{0, 1, 2}, off); all || !some {
		t.Fatal("some")
	}
	if all, some := selectionState([]int{0, 1}, []int{0, 1, 2}, off); !all || some {
		t.Fatal("all enabled")
	}
	got := selectShown([]int{7, 2}, []int{0, 1, 2}, off, true)
	if !reflect.DeepEqual(got, []int{7, 2, 0, 1}) {
		t.Fatalf("select shown = %v", got)
	}
	got = selectShown(got, []int{0, 1, 2}, off, false)
	if !reflect.DeepEqual(got, []int{7, 2}) {
		t.Fatalf("unselect shown keeps others and disabled: %v", got)
	}
}

func TestTableRadioRowSelection(t *testing.T) {
	var keys []int
	sel := -1
	cols := meetingColumns(false)[:3]
	tester := mount(func(c *ui.Context) {
		Table(c, cols, 4, meetingCell, TableProps{
			Selected: &sel, RowSelection: &keys, RowSelectionType: "radio",
		})
	}, 900, 400)
	r, _ := tester.Find("季度经营分析会")
	tester.ClickAt(r.X-40, r.Y+r.H/2)
	tester.Frame()
	if !reflect.DeepEqual(keys, []int{0}) || sel != 0 {
		t.Fatalf("radio = %v sel = %d", keys, sel)
	}
	r, _ = tester.Find("财务预算评审")
	tester.ClickAt(r.X-40, r.Y+r.H/2)
	tester.Frame()
	if !reflect.DeepEqual(keys, []int{1}) || sel != 1 {
		t.Fatalf("radio replaced = %v sel = %d", keys, sel)
	}
	h, _ := tester.Find("会议")
	tester.ClickAt(h.X-prefixW/2, h.Y+h.H/2)
	tester.Frame()
	if !reflect.DeepEqual(keys, []int{1}) {
		t.Fatalf("radio header must not select all: %v", keys)
	}
	if err := tester.Click("安全保密培训"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !reflect.DeepEqual(keys, []int{2}) || sel != 2 {
		t.Fatalf("row click radio = %v sel = %d", keys, sel)
	}
}

func TestTableRowSelectionAndFilterMenu(t *testing.T) {
	var checked []int
	var filters TableFilters
	cols := meetingColumns(false)[:4]
	cols[3].Filters = []Option{{Value: "进行中", Label: "只看进行中"}, {Value: "已结束", Label: "只看已结束"}}
	tester := mount(func(c *ui.Context) {
		Table(c, cols, 4, meetingCell, TableProps{RowSelection: &checked, Filters: &filters,
			RowDisabled: func(r int) bool { return r == 2 }})
	}, 900, 600)
	r, _ := tester.Find("项目周例会")
	tester.ClickAt(r.X-40, r.Y+r.H/2)
	tester.Frame()
	if !reflect.DeepEqual(checked, []int{3}) {
		t.Fatalf("row box = %v", checked)
	}
	h, _ := tester.Find("会议")
	tester.ClickAt(h.X-prefixW/2, h.Y+h.H/2)
	tester.Frame()
	if !reflect.DeepEqual(checked, []int{3, 0, 1}) {
		t.Fatalf("header box = %v", checked)
	}
	if err := tester.Click("筛选 状态"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("只看已结束"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("确定"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !reflect.DeepEqual(filters[3], []string{"已结束"}) {
		t.Fatalf("filters = %v", filters)
	}
	if tester.HasText("项目周例会") || !tester.HasText("安全保密培训") {
		t.Fatalf("rows after filter: %v", tester.Texts())
	}
	if err := tester.Click("筛选 状态"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("重置"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if len(filters[3]) != 0 || !tester.HasText("项目周例会") {
		t.Fatalf("reset: %v", filters)
	}
}

func TestTableExpandTreeAndEdit(t *testing.T) {
	owners := []string{"张三", "李四", "王五", "赵六", "张三", "钱七", "钱七"}
	var open []int
	tester := mount(func(c *ui.Context) {
		cols := meetingColumns(false)[:3]
		Table(c, cols, len(meetingRows), func(r, col int) Cell {
			if col == 2 {
				return Cell{Text: owners[r]}
			}
			return meetingCell(r, col)
		}, TableProps{Expanded: &open, Parent: func(r int) int { return meetingRows[r].parent },
			OnEdit: func(r, _ int, s string) { owners[r] = s }})
	}, 900, 600)
	if tester.HasText("议题一：营收") {
		t.Fatal("children shown while closed")
	}
	if err := tester.Click("展开行"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !tester.HasText("议题一：营收") || !reflect.DeepEqual(open, []int{0}) {
		t.Fatalf("open: %v %v", open, tester.Texts())
	}
	if err := tester.Click("李四"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	tester.Frame()
	if !tester.Focused("编辑单元格") {
		t.Fatal("edit field not focused")
	}
	tester.Key(0, ui.KeyBackspace)
	tester.Key(0, ui.KeyBackspace)
	tester.Type("孙八")
	tester.Key(0, ui.KeyEnter)
	tester.Frame()
	if owners[1] != "孙八" {
		t.Fatalf("edited = %q", owners[1])
	}
}

func TestTableExpandRowContent(t *testing.T) {
	tester := mount(func(c *ui.Context) {
		Table(c, meetingColumns(false)[:2], 2, meetingCell, TableProps{Expand: func(r int) func() {
			return func() { Muted(c, "详情 "+meetingRows[r].name) }
		}})
	}, 700, 400)
	if err := tester.Click("展开行"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !tester.HasText("详情 季度经营分析会") || tester.HasText("详情 财务预算评审") {
		t.Fatalf("expanded: %v", tester.Texts())
	}
	if err := tester.Click("收起行"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if tester.HasText("详情 季度经营分析会") {
		t.Fatal("did not collapse")
	}
}

func TestTableVirtualWithoutMaxHeightBuildsFewRows(t *testing.T) {
	tester := mount(func(c *ui.Context) {
		Table(c, []Column{{Title: "#"}}, 5000, func(r, _ int) Cell { return Cell{Text: "行" + itoa(r)} }, TableProps{Virtual: true})
	}, 400, 400)
	tester.Frame()
	n := 0
	for _, s := range tester.Texts() {
		if len(s) > 3 && s[:3] == "行" {
			n++
		}
	}
	if n == 0 || n > 40 {
		t.Fatalf("Virtual without MaxHeight built %d rows of 5000", n)
	}
}

func TestTableVirtualBuildsFewRows(t *testing.T) {
	tester := mount(func(c *ui.Context) {
		Table(c, []Column{{Title: "#"}}, 5000, func(r, _ int) Cell { return Cell{Text: "行" + itoa(r)} }, TableProps{Virtual: true, MaxHeight: 200})
	}, 400, 400)
	tester.Frame()
	n := 0
	for _, s := range tester.Texts() {
		if len(s) > 3 && s[:3] == "行" {
			n++
		}
	}
	if n == 0 || n > 20 {
		t.Fatalf("built %d rows", n)
	}
	r, _ := tester.Find("行0")
	tester.Scroll(r.X, r.Y, 0, 40*1000)
	tester.Frame()
	tester.Frame()
	if tester.HasText("行0") || !tester.HasText("行1000") {
		t.Fatalf("after scroll: %v", tester.Texts()[:5])
	}
	if !tester.HasText("#") {
		t.Fatal("header scrolled away")
	}
}

func TestTableExpandedTreeVirtualBuildsWindow(t *testing.T) {
	built := 0
	open := []int{0}
	parent := func(r int) int {
		if r == 0 {
			return -1
		}
		return 0
	}
	tester := mount(func(c *ui.Context) {
		built = 0
		Table(c, []Column{{Title: "#"}}, 101, func(r, _ int) Cell {
			built++
			return Cell{Text: "行" + itoa(r)}
		}, TableProps{Expanded: &open, Parent: parent})
	}, 400, 500)
	tester.Frame()
	if built == 0 || built >= 101 {
		t.Fatalf("expanded tree built %d of 101 widgets", built)
	}
	seen := map[string]bool{}
	for _, s := range tester.Texts() {
		seen[s] = true
	}
	if !seen["行0"] || !seen["行1"] {
		t.Fatalf("window dropped parent/child: %v", tester.Texts())
	}
	if seen["行100"] {
		t.Fatal("built the 100th child, not a window")
	}
}

func TestTableHeaderGroupsVirtualizeBody(t *testing.T) {
	built := 0
	cols := []Column{
		{Title: "姓名", Children: []Column{{Title: "姓"}, {Title: "名"}}},
	}
	tester := mount(func(c *ui.Context) {
		built = 0
		Table(c, cols, 80, func(r, col int) Cell {
			built++
			return Cell{Text: "r" + itoa(r) + "c" + itoa(col)}
		})
	}, 400, 500)
	tester.Frame()
	if built == 0 || built >= 160 {
		t.Fatalf("grouped header table built %d of 160 cells", built)
	}
	for _, want := range []string{"姓名", "姓", "名", "r0c0"} {
		if !tester.HasText(want) {
			t.Fatalf("missing %q in %v", want, tester.Texts())
		}
	}
	seen := map[string]bool{}
	for _, s := range tester.Texts() {
		seen[s] = true
	}
	if seen["r79c0"] {
		t.Fatal("grouped header table painted every body row")
	}
}

func TestTableWrapVirtualBuildsWindow(t *testing.T) {
	built := 0
	long := "换行文字换行文字换行文字换行文字换行文字换行文字换行文字换行文字"
	tester := mount(func(c *ui.Context) {
		built = 0
		Table(c, []Column{{Title: "说明", Wrap: true}}, 80, func(r, _ int) Cell {
			built++
			return Cell{Text: "行" + itoa(r) + " " + long}
		})
	}, 400, 500)
	tester.Frame()
	if built == 0 || built >= 80 {
		t.Fatalf("wrap table built %d of 80", built)
	}
	if !tester.HasText("行0") {
		t.Fatal("wrap virtual dropped the first row")
	}
}

func TestTableWrapVirtualRowGrowsPastTwoLines(t *testing.T) {
	built := 0
	long := "换行文字换行文字换行文字换行文字换行文字换行文字换行文字换行文字换行文字换行文字"
	label0 := "行0 " + long
	twoLine := Work.Height(Middle) + 8 + Work.FontSize + 4
	tester := mount(func(c *ui.Context) {
		built = 0
		Table(c, []Column{{Title: "说明", Wrap: true, Width: 120, Min: 120}, {Title: "垫", Width: 400}}, 80, func(r, col int) Cell {
			if col != 0 {
				return Cell{Text: "x" + itoa(r)}
			}
			built++
			return Cell{Text: "行" + itoa(r) + " " + long}
		}, TableProps{Virtual: true, MaxHeight: 240, ScrollX: 600})
	}, 400, 500)
	tester.Frame()
	if built == 0 || built >= 80 {
		t.Fatalf("wrap table built %d of 80", built)
	}
	r0, ok := tester.Find(label0)
	if !ok {
		t.Fatalf("missing first wrap row in %v", tester.Texts())
	}
	if r0.H <= twoLine {
		t.Fatalf("wrap row height %v, want taller than 2-line cap %v", r0.H, twoLine)
	}
	r1, ok := tester.Find("行1 " + long)
	if !ok {
		t.Fatal("missing second wrap row")
	}
	if r1.Y-r0.Y <= twoLine {
		t.Fatalf("row stride %v, want taller than 2-line cap %v", r1.Y-r0.Y, twoLine)
	}
}

func TestTableExpandVirtualKeepsPanelInWindow(t *testing.T) {
	built := 0
	tester := mount(func(c *ui.Context) {
		built = 0
		Table(c, []Column{{Title: "#"}}, 80, func(r, _ int) Cell {
			built++
			return Cell{Text: "行" + itoa(r)}
		}, TableProps{Expand: func(r int) func() {
			return func() { ui.Text(c, "详情"+itoa(r)) }
		}})
	}, 400, 500)
	tester.Frame()
	if built == 0 || built >= 80 {
		t.Fatalf("expand table built %d of 80", built)
	}
	if err := tester.Click("展开行"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !tester.HasText("详情0") {
		t.Fatalf("virtual window dropped expand panel: %v", tester.Texts())
	}
	if built >= 80 {
		t.Fatalf("expand after open built %d of 80", built)
	}
}

func TestTableExpandVirtualPanelInScroll(t *testing.T) {
	built := 0
	const panelH float32 = 160
	tester := mount(func(c *ui.Context) {
		built = 0
		Table(c, []Column{{Title: "#"}}, 80, func(r, _ int) Cell {
			built++
			return Cell{Text: "行" + itoa(r)}
		}, TableProps{Virtual: true, MaxHeight: 220, Expand: func(r int) func() {
			return func() {
				ui.Box(c).Height(panelH)
				ui.Text(c, "详情"+itoa(r))
			}
		}})
	}, 400, 500)
	tester.Frame()
	if err := tester.Click("展开行"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	tester.Frame()
	if !tester.HasText("详情0") {
		t.Fatalf("expand panel missing: %v", tester.Texts())
	}
	r0, ok0 := tester.Find("行0")
	r1, ok1 := tester.Find("行1")
	if !ok0 || !ok1 {
		t.Fatalf("rows after expand: %v", tester.Texts())
	}
	rowH := Work.Height(Middle) + 8
	if r1.Y-r0.Y < rowH+panelH-8 {
		t.Fatalf("expand panel not in row height: stride %v, rowH %v panel %v", r1.Y-r0.Y, rowH, panelH)
	}
	if built == 0 || built >= 80 {
		t.Fatalf("expand still virtual, built %d of 80", built)
	}
}

func TestTableVirtualColumnAlign(t *testing.T) {
	tester := mount(func(c *ui.Context) {
		Table(c, []Column{
			{Title: "姓名", Width: 160},
			{Title: "职务", Width: 220},
			{Title: "备注", Width: 300},
		}, 80, func(r, col int) Cell {
			switch col {
			case 0:
				return Cell{Text: "张" + itoa(r)}
			case 1:
				return Cell{Text: "工" + itoa(r)}
			}
			return Cell{Text: "注" + itoa(r)}
		}, TableProps{Virtual: true, MaxHeight: 240, ScrollX: 760})
	}, 400, 500)
	tester.Frame()
	tester.Frame()
	name, ok := tester.Find("姓名")
	if !ok {
		t.Fatal("missing header")
	}
	cell, ok := tester.Find("张0")
	if !ok {
		t.Fatalf("missing body cell in %v", tester.Texts())
	}
	if abs32(name.X-cell.X) > 1 {
		t.Fatalf("header/body column 0 out of sync: header X %v cell X %v", name.X, cell.X)
	}
	job, _ := tester.Find("职务")
	jobCell, ok := tester.Find("工0")
	if !ok {
		t.Fatal("missing job cell")
	}
	if abs32(job.X-jobCell.X) > 1 {
		t.Fatalf("header/body column 1 out of sync: header X %v cell X %v", job.X, jobCell.X)
	}
	note, ok := tester.Find("备注")
	noteCell, noteOK := tester.Find("注0")
	if !ok || !noteOK {
		t.Fatal("missing note column")
	}
	if abs32(note.X-noteCell.X) > 1 {
		t.Fatalf("header/body column 2 out of sync: header X %v cell X %v", note.X, noteCell.X)
	}
	tester.Scroll(cell.X+40, cell.Y+8, 280, 0)
	tester.Frame()
	tester.Frame()
	note2, ok := tester.Find("备注")
	noteCell2, noteOK := tester.Find("注0")
	if !ok || !noteOK || note2.W < 8 || noteCell2.W < 8 {
		t.Fatalf("note column missing after sideways scroll: header %v cell %v texts %v", note2, noteCell2, tester.Texts())
	}
	if abs32(note2.X-noteCell2.X) > 1 {
		t.Fatalf("after sideways scroll, header/body out of sync: header X %v cell X %v", note2.X, noteCell2.X)
	}
	if note.X-note2.X < 200 {
		t.Fatalf("note column did not scroll with the list: %v → %v", note.X, note2.X)
	}
}

func TestTableFixedColumnStaysAtEdge(t *testing.T) {
	tester := mount(func(c *ui.Context) {
		Table(c, meetingColumns(true), 2, meetingCell)
	}, 500, 300)
	first, _ := tester.Find("季度经营分析会")
	act, _ := tester.Find("详情")
	tester.Scroll(first.X+200, first.Y, 300, 0)
	tester.Frame()
	tester.Frame()
	first2, _ := tester.Find("季度经营分析会")
	act2, _ := tester.Find("详情")
	if first2.X != first.X || act2.X != act.X {
		t.Fatalf("fixed cells moved: %v→%v, %v→%v", first.X, first2.X, act.X, act2.X)
	}
	room, _ := tester.Find("一号会议室")
	if room.X > first.X+200 {
		t.Fatalf("free column did not scroll: %v", room.X)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}

func TestColumnLeavesAndGroups(t *testing.T) {
	cols := []Column{
		{Title: "姓名", Children: []Column{{Title: "姓", Width: 80}, {Title: "名", Width: 90}}},
		{Title: "年龄", Width: 70},
	}
	leaves := columnLeaves(cols)
	if len(leaves) != 3 || leaves[0].Title != "姓" || leaves[2].Title != "年龄" {
		t.Fatalf("leaves = %+v", leaves)
	}
	if !groupedColumns(cols) || groupedColumns(leaves) || leafCount(cols[0]) != 2 || headerDepth(cols) != 2 {
		t.Fatal("groupedColumns / leafCount / headerDepth")
	}
	b, g := groupFlex(cols[0])
	if b != 170 || g != 0 {
		t.Fatalf("groupFlex = %v %v", b, g)
	}
	if nativeTableOK(cols, TableProps{}) {
		t.Fatal("grouped columns must not use Native ui.Table")
	}
}

func TestTableHeaderGroups(t *testing.T) {
	cols := []Column{
		{Title: "姓名", Children: []Column{{Title: "姓", Width: 100}, {Title: "名", Width: 100}}},
		{Title: "状态", Width: 80},
	}
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			Table(c, cols, 2, func(r, col int) Cell {
				switch col {
				case 0:
					return Cell{Text: []string{"张", "李"}[r]}
				case 1:
					return Cell{Text: []string{"三", "四"}[r]}
				}
				return Cell{Text: "在线"}
			}, TableProps{Native: true})
		})
	}, 520, 280)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	for _, want := range []string{"姓名", "姓", "名", "状态", "张", "三"} {
		if !tester.HasText(want) {
			t.Fatalf("missing %q in %v", want, tester.Texts())
		}
	}
}
