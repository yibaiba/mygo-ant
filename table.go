package kit

import (
	"cmp"
	"slices"
	"strconv"

	"github.com/egoist/mygo/ui"
)

// Align places a column's content.
type Align uint8

// Column alignments.
const (
	// AlignStart lines content up on the left.
	AlignStart Align = iota
	// AlignCenter centers it.
	AlignCenter
	// AlignEnd lines it up on the right, as for numbers.
	AlignEnd
)

// Fixed pins a column to an edge of a table that scrolls sideways.
type Fixed uint8

// Fixed sides.
const (
	FixedNone Fixed = iota
	// FixedLeft keeps the column at the left edge, with the fixed-left
	// columns before it.
	FixedLeft
	// FixedRight keeps it at the right edge.
	FixedRight
)

// Column is one column of a Table.
type Column struct {
	Title string
	// Width is the column's share before the rest is handed out; zero makes
	// it a flexible column that takes the space left. With no flexible
	// column the last one grows. A Fixed column is exactly Width wide.
	Width float32
	// Min keeps the column from shrinking below it.
	Min   float32
	Align Align
	// Wrap lets the text break over lines; otherwise it ends in an ellipsis.
	// A virtual body is a measured ui.List, so a wrapping row can grow.
	Wrap bool
	// Sorter compares rows a and b of the column, negative when a comes
	// first ascending; it makes the header a button cycling ascending,
	// descending and unsorted, as Ant's sorter.
	Sorter func(a, b int) int
	// Fixed pins the column to the left or right edge while the table
	// scrolls sideways; it needs a Width. The table moves the fixed-left
	// columns first and the fixed-right ones last.
	Fixed Fixed
	// Filters put a funnel in the header that opens a checklist of them;
	// OnFilter keeps a row when it reports true for any value checked.
	Filters []Option
	// FilterSearch adds a search box over the checklist; FilterSingle
	// makes it a radio list.
	FilterSearch, FilterSingle bool
	OnFilter                   func(row int, value string) bool
	// Editable lets a click edit the cell's text in place; TableProps.OnEdit
	// takes the new text on Enter or when the field loses focus.
	Editable bool
	// Children group leaf columns under this title, as Ant's column
	// children. Only the leaves are cells. Native ui.Table cannot group
	// headers, so a grouped table stays on the painted path.
	Children []Column
}

// SortOrder is how a table column is sorted.
type SortOrder uint8

// Sort orders, in the order a header click cycles through them.
const (
	// SortNone keeps the rows in their own order.
	SortNone SortOrder = iota
	// SortAscend puts the smallest first.
	SortAscend
	// SortDescend puts the largest first.
	SortDescend
)

// SortState is the sorted column of a table and its order.
type SortState struct {
	Col   int
	Order SortOrder
}

// next is the state after a click on column col's header: ascending on
// a new column, then descending, then unsorted, as Ant.
func (s SortState) next(col int) SortState {
	if s.Col != col || s.Order == SortNone {
		return SortState{Col: col, Order: SortAscend}
	}
	if s.Order == SortAscend {
		return SortState{Col: col, Order: SortDescend}
	}
	return SortState{Col: col}
}

// sortRows is the order a table shows rows 0..n-1 in: stable by less,
// reversed for SortDescend, as given for SortNone or a nil less.
func sortRows(n int, less func(a, b int) int, order SortOrder) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return sortSome(out, less, order)
}

// sortSome sorts rows in place as sortRows does and returns them.
func sortSome(rows []int, less func(a, b int) int, order SortOrder) []int {
	if less == nil || order == SortNone {
		return rows
	}
	slices.SortStableFunc(rows, func(a, b int) int {
		if order == SortDescend {
			return less(b, a)
		}
		return less(a, b)
	})
	return rows
}

// ByText is a Sorter comparing the strings key returns for rows.
func ByText(key func(row int) string) func(a, b int) int {
	return func(a, b int) int { return cmp.Compare(key(a), key(b)) }
}

// ByNumber is a Sorter comparing the numbers key returns for rows.
func ByNumber(key func(row int) float64) func(a, b int) int {
	return func(a, b int) int { return cmp.Compare(key(a), key(b)) }
}

// Cell is what a Table shows in one place: Text in Color (zero is the
// body text color), or whatever Render builds.
type Cell struct {
	Text   string
	Color  ui.Color
	Strong bool
	Render func()
}

// TableFilters are the values checked in each filtered column, by
// column index.
type TableFilters map[int][]string

// TableProps are a table's options.
type TableProps struct {
	// Selected makes rows clickable and holds the chosen row, -1 for none.
	Selected *int
	// Empty is the text shown when there are no rows; "暂无数据" by default.
	Empty string
	// MaxHeight scrolls the body under a header that stays; zero lets it
	// grow.
	MaxHeight float32
	Size      Size
	// Loading dims the rows under a spinner and takes their clicks away,
	// as while a scan runs.
	Loading bool
	// Sort holds the sorted column of columns with a Sorter; the table
	// keeps its own when nil.
	Sort *SortState
	// PageSize > 0 shows that many rows a page over a Pagination, which
	// hides while one page holds them all; Page holds the 1-based page, or
	// the table keeps its own when nil. With Parent, it pages the top rows.
	PageSize int
	Page     *int
	// ScrollX is the width under which the table scrolls sideways rather
	// than squeeze its columns; zero sums the columns' Width and Min.
	ScrollX float32
	// Bordered rules every cell, as Ant's bordered table.
	Bordered bool

	// RowSelection adds a selection column, as Ant's rowSelection: it holds
	// the rows checked. RowSelectionType "radio" picks one row (Ant
	// rowSelection.type); empty or "checkbox" is the default checkbox
	// column, whose header box checks every row shown.
	// RowDisabled greys out the boxes of the rows it reports.
	RowSelection     *[]int
	RowSelectionType string
	RowDisabled      func(row int) bool
	// Filters holds the columns' filters; the table keeps its own when nil.
	Filters *TableFilters
	// Expand, when set, adds a + / − column; a row it returns a builder
	// for opens to show what that builds under it. Expanded holds the open
	// rows (and the open rows of a tree); the table keeps its own when nil.
	Expand   func(row int) func()
	Expanded *[]int
	// Parent makes the rows a tree, as Ant's tree data: row's parent, -1
	// at the top. The first column indents by depth and opens children.
	Parent func(row int) int
	// Virtual builds only the rows in view of MaxHeight. The painted body
	// is a ui.List: rows take their measured height (wrapping cells and
	// expand panels included) while the header, grouped or not, stays
	// outside and ScrollWith the list so columns stay a table. Tables with
	// more than virtualOver (64) shown rows virtualize on their own: a
	// tree flattens the open outline first, then only that window is built.
	Virtual bool
	// Summary is a last row of totals, as Ant's summary: the cell of each
	// column (bold unless set), or none for a zero Cell.
	Summary func(col int) Cell
	// OnEdit takes the text typed into an Editable column's cell.
	OnEdit func(row, col int, text string)
	// OnDoubleClick is Ant onRow onDoubleClick: a double-click of a row.
	OnDoubleClick func(row int)
	// ShowTotal puts "共 N 条" on the pager, as Pagination's.
	ShowTotal bool
	// Native opts into MyGo's ui.Table (resize, reorder, sort, virtual
	// rows) when the painted extras are unused (no expand, tree, checkbox
	// selection, summary, in-cell edit, filters, fixed or wrapping
	// columns). The default is the painted Ant table. Empty and loading
	// stay on the painted path.
	Native bool

	forceHover int // row index + 1
}

// contentWidth is the width a table's columns need: their Width or Min,
// and 80 for a flexible column without Min, plus padding.
func contentWidth(columns []Column, padX float32) float32 {
	var w float32
	for _, col := range columns {
		w += max(col.Width, col.Min, pick[float32](col.Width == 0 && col.Min == 0, 80, 0)) + 2*padX
	}
	return w
}

// pageRows is the part of order on page (1-based) at size rows a page.
func pageRows(order []int, page, size int) []int {
	if size <= 0 {
		return order
	}
	start := (max(page, 1) - 1) * size
	if start >= len(order) {
		return nil
	}
	return order[start:min(start+size, len(order))]
}

// tableRow is a row as shown: its index, its depth in a tree and
// whether it has children there.
type tableRow struct {
	row, depth int
	kids       bool
}

// colKind is what a column of the laid-out table holds.
type colKind uint8

const (
	dataCol colKind = iota
	selectCol
	expandCol
)

// tcol is a column as laid out: a data column (idx into the columns) or
// the selection or expand column, its width and side.
type tcol struct {
	kind  colKind
	idx   int
	width float32
	fixed Fixed
}

// prefixW is the width of the selection and expand columns (Ant's 32 +
// padding and expandColumnWidth).
const prefixW = 48

// tableColumns lays out the columns: the selection and expand columns
// first, then the fixed-left, the free and the fixed-right columns.
// The prefix columns are fixed left when any column is.
func tableColumns(columns []Column, selection, expand bool) []tcol {
	var left, mid, right []tcol
	anyLeft := false
	for i, col := range columns {
		tc := tcol{kind: dataCol, idx: i, width: col.Width, fixed: col.Fixed}
		if col.Fixed != FixedNone && col.Width == 0 {
			tc.fixed = FixedNone
		}
		switch tc.fixed {
		case FixedLeft:
			anyLeft = true
			left = append(left, tc)
		case FixedRight:
			right = append(right, tc)
		default:
			mid = append(mid, tc)
		}
	}
	var pre []tcol
	side := pick(anyLeft, FixedLeft, FixedNone)
	if selection {
		pre = append(pre, tcol{kind: selectCol, width: prefixW, fixed: side})
	}
	if expand {
		pre = append(pre, tcol{kind: expandCol, width: prefixW, fixed: side})
	}
	return slices.Concat(pre, left, mid, right)
}

// fixedOffsets is where each fixed column sits from its edge: the widths
// of the fixed columns before it on the left, after it on the right.
func fixedOffsets(cols []tcol) []float32 {
	out := make([]float32, len(cols))
	var x float32
	for i, tc := range cols {
		if tc.fixed == FixedLeft {
			out[i] = x
			x += tc.width
		}
	}
	x = 0
	for i := len(cols) - 1; i >= 0; i-- {
		if cols[i].fixed == FixedRight {
			out[i] = x
			x += cols[i].width
		}
	}
	return out
}

// filterRows keeps the rows of order that every filtered column keeps.
func filterRows(order []int, columns []Column, filters TableFilters) []int {
	active := false
	for i, vals := range filters {
		if len(vals) > 0 && i < len(columns) && columns[i].OnFilter != nil {
			active = true
		}
	}
	if !active {
		return order
	}
	var out []int
	for _, r := range order {
		keep := true
		for i, vals := range filters {
			if len(vals) == 0 || i >= len(columns) || columns[i].OnFilter == nil {
				continue
			}
			any := false
			for _, v := range vals {
				if columns[i].OnFilter(r, v) {
					any = true
					break
				}
			}
			if !any {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, r)
		}
	}
	return out
}

// tableTree splits rows into the top ones and the children of each, in
// the order of rows, by parent. A row whose parent is not among rows is
// at the top.
func tableTree(rows []int, parent func(int) int) (roots []int, kids map[int][]int) {
	in := map[int]bool{}
	for _, r := range rows {
		in[r] = true
	}
	kids = map[int][]int{}
	for _, r := range rows {
		if p := parent(r); p >= 0 && in[p] && p != r {
			kids[p] = append(kids[p], r)
		} else {
			roots = append(roots, r)
		}
	}
	return roots, kids
}

// flattenRows is the rows shown of a tree: each of roots, and under an
// open one its children, deeper by one.
func flattenRows(roots []int, kids map[int][]int, open map[int]bool) []tableRow {
	var out []tableRow
	var walk func(rs []int, depth int)
	walk = func(rs []int, depth int) {
		for _, r := range rs {
			ch := kids[r]
			out = append(out, tableRow{row: r, depth: depth, kids: len(ch) > 0})
			if len(ch) > 0 && open[r] {
				walk(ch, depth+1)
			}
		}
	}
	walk(roots, 0)
	return out
}

// selectionState is how the header box over shown rows looks with
// checked rows: all of the enabled ones, or some.
func selectionState(checked, shown []int, off func(int) bool) (all, some bool) {
	set := map[int]bool{}
	for _, r := range checked {
		set[r] = true
	}
	n, on := 0, 0
	for _, r := range shown {
		if off != nil && off(r) {
			continue
		}
		n++
		if set[r] {
			on++
		}
	}
	return n > 0 && on == n, on > 0 && on < n
}

// selectShown checks (or unchecks) every enabled row of shown in
// checked, keeping the rows not shown and the disabled ones as they are.
func selectShown(checked, shown []int, off func(int) bool, on bool) []int {
	set := map[int]bool{}
	for _, r := range checked {
		set[r] = true
	}
	for _, r := range shown {
		if off == nil || !off(r) {
			set[r] = on
		}
	}
	out := make([]int, 0, len(set))
	for _, r := range checked {
		if set[r] {
			out = append(out, r)
			delete(set, r)
		}
	}
	for _, r := range shown {
		if set[r] {
			out = append(out, r)
			delete(set, r)
		}
	}
	return out
}

func toggleRow(rows []int, r int) []int {
	if i := slices.Index(rows, r); i >= 0 {
		return slices.Delete(slices.Clone(rows), i, i+1)
	}
	return append(slices.Clone(rows), r)
}

// tableEdit is the cell being edited and its text.
type tableEdit struct {
	row, col int
	draft    string
	started  bool
}

// Table draws Ant's table: a FillAlter header with faint column rules,
// BorderSecondary row dividers, hover and selected rows, and an empty
// state. With Selected set, a click or Enter chooses a row and the arrows
// move between rows; it reports a new choice. OnDoubleClick is a
// double-click of a row (Ant onRow). Truncated cells show the full text
// in a tooltip. Selected, cell and the other row arguments take row
// indexes as given, whatever the sort, the filters and the page. Columns
// with a Sorter sort on a header click, Filters filter, PageSize pages
// the rows, and the table scrolls sideways when narrower than its columns
// need, with Fixed columns staying put. TableProps add checkbox
// selection, expandable rows, tree rows, a virtual body, a summary row
// and editable cells.
func Table(c *ui.Context, columns []Column, rows int, cell func(row, col int) Cell, props ...TableProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	t := Of(c)
	groups := columns
	grouped := groupedColumns(columns)
	columns = columnLeaves(columns)
	radio := p.RowSelection != nil && p.RowSelectionType == "radio"
	rowH := t.Height(p.Size) + 8
	padX := pick[float32](p.Size == Small, 8, 16)
	cols := tableColumns(columns, p.RowSelection != nil, p.Expand != nil && p.Parent == nil)
	radioKey := ""
	if radio && len(*p.RowSelection) == 1 {
		radioKey = strconv.Itoa((*p.RowSelection)[0])
	}
	offsets := fixedOffsets(cols)
	flexible, grower, firstData := false, -1, -1
	for k, tc := range cols {
		if tc.kind == dataCol && firstData < 0 {
			firstData = k
		}
		if tc.kind == dataCol && tc.fixed == FixedNone {
			flexible = flexible || tc.width == 0
			grower = k
		}
	}
	place := func(e ui.Element, k int) ui.Element {
		tc := cols[k]
		e.Padding(0, pick[float32](tc.kind == dataCol, padX, 0)).AlignItems(ui.Center)
		if tc.kind != dataCol {
			return e.Width(tc.width).Shrink(0).Justify(ui.Center)
		}
		col := columns[tc.idx]
		e.MinWidth(col.Min).Justify(justify(col.Align))
		switch {
		case tc.fixed != FixedNone:
			e.Width(tc.width).Shrink(0)
		case col.Width == 0:
			e.Basis(0).Grow(1)
		case !flexible && k == grower:
			e.Basis(col.Width).Grow(1)
		default:
			e.Basis(col.Width).Shrink(1)
		}
		return e
	}
	changed := false
	root := ui.Column(c).FillWidth().Gap(t.SpaceLG)
	sort := p.Sort
	if sort == nil {
		sort = ui.Local(root, "sort", func() SortState { return SortState{Col: -1} })
	}
	page := p.Page
	if page == nil {
		page = ui.Local(root, "page", func() int { return 1 })
	}
	filters := p.Filters
	if filters == nil {
		filters = ui.Local(root, "filters", func() TableFilters { return TableFilters{} })
	}
	if *filters == nil {
		*filters = TableFilters{}
	}
	expanded := p.Expanded
	if expanded == nil {
		expanded = ui.Local(root, "expanded", func() []int { return nil })
	}
	edit := ui.Local(root, "edit", func() tableEdit { return tableEdit{row: -1} })
	open := map[int]bool{}
	for _, r := range *expanded {
		open[r] = true
	}
	var less func(a, b int) int
	if sort.Col >= 0 && sort.Col < len(columns) {
		less = columns[sort.Col].Sorter
	}
	all := make([]int, rows)
	for i := range all {
		all[i] = i
	}
	kept := filterRows(all, columns, *filters)
	var shown []tableRow
	total := len(kept)
	if p.Parent != nil {
		roots, kids := tableTree(kept, p.Parent)
		sortSome(roots, less, sort.Order)
		for k := range kids {
			sortSome(kids[k], less, sort.Order)
		}
		total = len(roots)
		if p.PageSize > 0 {
			*page = max(1, min(*page, PageCount(total, p.PageSize)))
		}
		shown = flattenRows(pageRows(roots, *page, p.PageSize), kids, open)
	} else {
		order := sortSome(kept, less, sort.Order)
		if p.PageSize > 0 {
			*page = max(1, min(*page, PageCount(total, p.PageSize)))
		}
		for _, r := range pageRows(order, *page, p.PageSize) {
			shown = append(shown, tableRow{row: r})
		}
	}
	shownRows := make([]int, len(shown))
	for i, r := range shown {
		shownRows[i] = r.row
	}
	sortedCol := pick(less != nil && sort.Order != SortNone, sort.Col, -1)
	minW := p.ScrollX
	if minW == 0 {
		minW = contentWidth(columns, padX)
		for _, tc := range cols {
			if tc.kind != dataCol {
				minW += tc.width
			}
		}
	}
	// One scroll state holds where the body is scrolled sideways. A
	// virtual body is a ui.List; its header follows through ScrollWith.
	// Otherwise a header in a scroller of its own (under MaxHeight)
	// follows the body's X.
	body := ui.Local(root, "body", func() ui.ScrollState { return ui.ScrollState{} })
	head := ui.Local(root, "head", func() ui.ScrollState { return ui.ScrollState{} })
	lastX := ui.Local(root, "x", func() float32 { return 0 })
	// shown is already the open outline of a tree (or the page of
	// rows). Auto-virtual that list even with expand, tree, wrap or
	// grouped headers: the body is a measured ui.List.
	if !p.Virtual && autoVirtual(len(shown)) {
		p.Virtual = true
	}
	virtualFill := p.Virtual && p.MaxHeight <= 0
	if virtualFill {
		// Ant virtual tables need scroll.y; without a height Virtual was a
		// no-op (split is MaxHeight > 0). virtualBodyH is Ant's usual body
		// window; Grow fills a shorter parent, and the last-frame body
		// height is the virtual viewport.
		p.MaxHeight = virtualBodyH
	}
	split := p.MaxHeight > 0
	useList := p.Virtual && split && len(shown) > 0
	if split && !useList {
		if head.X != *lastX {
			body.X = head.X
		} else {
			head.X = body.X
		}
		*lastX = body.X
	}
	view := ui.Local(root, "view", func() float32 { return 0 })
	bodyH := ui.Local(root, "bodyH", func() float32 { return 0 })
	sx, maxX := body.X, body.MaxX
	width := max(minW, *view)

	// fixedCell builds cell k of a row: in its place in the row, or for a
	// fixed column, as a box above the row at its edge, on bg.
	fixedCell := func(c *ui.Context, k int, bg ui.Color, build func(e ui.Element)) {
		tc := cols[k]
		if tc.fixed == FixedNone {
			build(place(ui.Row(c.Key(k)), k))
			return
		}
		ui.Box(c.Key(k)).Width(tc.width).Shrink(0)
		e := place(ui.Row(c.Key(-1-k)), k).Absolute().Top(0).Bottom(0).Background(bg)
		edge := false
		if tc.fixed == FixedLeft {
			e.Left(sx + offsets[k])
			edge = sx > 0 && (k+1 >= len(cols) || cols[k+1].fixed != FixedLeft)
		} else {
			e.Right(max(maxX-sx, 0) + offsets[k])
			edge = sx < maxX-0.5 && (k == 0 || cols[k-1].fixed != FixedRight)
		}
		if edge {
			left := tc.fixed == FixedLeft
			e.DrawOver(func(pa *ui.Painter, r ui.Rect) {
				shade := ui.RGBA(0, 0, 0, 0.08)
				if left {
					pa.FillGradient(ui.Rect{X: r.X + r.W, Y: r.Y, W: 10, H: r.H}, ui.LinearGradient{Angle: 90, From: shade, To: shade.Alpha(0)}, 0)
				} else {
					pa.FillGradient(ui.Rect{X: r.X - 10, Y: r.Y, W: 10, H: r.H}, ui.LinearGradient{Angle: 90, From: shade.Alpha(0), To: shade}, 0)
				}
			})
		}
		build(e)
	}

	header := func(c *ui.Context) {
		bg := t.FillAlter.Over(t.Surface)
		ui.Row(c).Width(width).Shrink(0).MinHeight(rowH).Background(t.FillAlter).AlignItems(ui.Stretch).
			Role(ui.RoleRow).Draw(bottomRule(t.BorderSecondary)).Children(func() {
			if grouped {
				for k, tc := range cols {
					if tc.kind == dataCol {
						continue
					}
					rule := k < len(cols)-1 || p.Bordered
					fixedCell(c, k, bg, func(h ui.Element) {
						h.Role(ui.RoleColumnHeader)
						if rule {
							h.DrawOver(rightRule(t.BorderSecondary))
						}
						h.Children(func() {
							if tc.kind == selectCol && !radio {
								all, some := selectionState(*p.RowSelection, shownRows, p.RowDisabled)
								on := all
								ui.Row(c).Children(func() {
									if Checkbox(c, &on, "", ChoiceProps{Indeterminate: some, Disabled: len(shownRows) == 0}) {
										*p.RowSelection = selectShown(*p.RowSelection, shownRows, p.RowDisabled, on)
										changed = true
									}
								})
							}
						})
					})
				}
				if tableGroupHeaders(c, t, groups, columns, cols, sortedCol, sort, filters, padX, rowH, p.Bordered, place) {
					*page = 1
				}
				return
			}
			for k, tc := range cols {
				rule := k < len(cols)-1 || p.Bordered
				fixedCell(c, k, bg, func(h ui.Element) {
					h.Role(ui.RoleColumnHeader)
					if rule {
						if p.Bordered {
							h.DrawOver(rightRule(t.BorderSecondary))
						} else if tc.kind == dataCol {
							h.DrawOver(headRule(t.BorderSecondary))
						}
					}
					h.Children(func() {
						switch tc.kind {
						case selectCol:
							if radio {
								return
							}
							all, some := selectionState(*p.RowSelection, shownRows, p.RowDisabled)
							on := all
							ui.Row(c).Children(func() {
								if Checkbox(c, &on, "", ChoiceProps{Indeterminate: some, Disabled: len(shownRows) == 0}) {
									*p.RowSelection = selectShown(*p.RowSelection, shownRows, p.RowDisabled, on)
									changed = true
								}
							})
						case dataCol:
							col := columns[tc.idx]
							if tableHeadCell(c, t, tc.idx, col, sortedCol == tc.idx, sort, filters, padX) {
								*page = 1
							}
						}
					})
				})
			}
		})
	}

	rowCells := func(c *ui.Context, tr tableRow, bg ui.Color, chosen, hovered bool) {
		r := tr.row
		for k, tc := range cols {
			fixedCell(c, k, bg, func(box ui.Element) {
				box.Role(ui.RoleCell)
				box.Padding(t.SpaceSM, pick[float32](tc.kind == dataCol, padX, 0))
				if p.Bordered && k < len(cols)-1 {
					box.DrawOver(rightRule(t.BorderSecondary))
				}
				switch tc.kind {
				case selectCol:
					off := p.RowDisabled != nil && p.RowDisabled(r)
					box.Children(func() {
						ui.Row(c).Children(func() {
							if radio {
								if Radio(c, &radioKey, strconv.Itoa(r), "", ChoiceProps{Disabled: off}) {
									*p.RowSelection = []int{r}
									if p.Selected != nil {
										*p.Selected = r
									}
									changed = true
								}
								return
							}
							on := slices.Contains(*p.RowSelection, r)
							if Checkbox(c, &on, "", ChoiceProps{Disabled: off}) {
								*p.RowSelection = toggleRow(*p.RowSelection, r)
								changed = true
							}
						})
					})
					return
				case expandCol:
					fn := p.Expand(r)
					box.Children(func() {
						if fn != nil && expandIcon(c, t, open[r]) {
							*expanded = toggleRow(*expanded, r)
							changed = true
						}
					})
					return
				}
				i := tc.idx
				col := columns[i]
				if i == sortedCol && !chosen && !hovered {
					box.Background(pick(tc.fixed == FixedNone, t.FillAlter, t.FillAlter.Over(t.Surface)))
				}
				first := k == firstData
				box.Children(func() {
					if p.Parent != nil && first {
						ui.Box(c).Width(float32(tr.depth) * 15).Shrink(0)
						if tr.kids {
							if expandIcon(c, t, open[r]) {
								*expanded = toggleRow(*expanded, r)
								changed = true
							}
						} else {
							ui.Box(c).Width(17).Shrink(0)
						}
						ui.Box(c).Width(8).Shrink(0)
					}
					if col.Editable && p.OnEdit != nil {
						if tableEditCell(c, t, edit, r, i, cell(r, i), col.Align, p.OnEdit) {
							changed = true
						}
						return
					}
					cellContent(c, t, cell(r, i), col.Wrap)
				})
			})
		}
	}

	selHover := t.PrimaryBg.Mix(t.PrimaryBorder, 0.3)
	bodyRow := func(c *ui.Context, k int, tr tableRow) {
		r := tr.row
		var line ui.Element
		clickable := p.Selected != nil || radio || p.OnDoubleClick != nil
		if clickable {
			sel := p.Selected != nil && *p.Selected == r
			line = ui.ButtonBase(c.Key(r)).Label("第 " + strconv.Itoa(r+1) + " 行").FocusRing(false).
				Cursor(ui.CursorPointer).Selected(sel)
		} else {
			line = ui.Row(c.Key(r))
		}
		if useList {
			line.Row().FillWidth().Shrink(0).AlignItems(ui.Stretch).MinHeight(rowH)
		} else {
			line.Row().Width(width).Shrink(0).AlignItems(ui.Stretch).MinHeight(rowH)
		}
		line.Role(ui.RoleRow)
		if k < len(shown)-1 || p.Summary != nil || open[r] && p.Expand != nil {
			line.Draw(bottomRule(t.BorderSecondary))
		}
		chosen := p.Selected != nil && *p.Selected == r || p.RowSelection != nil && slices.Contains(*p.RowSelection, r)
		hovered := line.Hovered() || p.forceHover == r+1
		bg := ui.Color{}
		switch {
		case chosen && hovered:
			bg = selHover
		case chosen:
			bg = t.PrimaryBg
		case hovered:
			bg = t.FillAlter
		}
		if bg != (ui.Color{}) {
			line.Background(bg)
		}
		if line.FocusVisible() {
			line.InsetShadow(0, 0, 0, 2, t.PrimaryBorder)
		}
		line.Children(func() { rowCells(c, tr, bg.Over(t.Surface), chosen, hovered) })
		if clickable && line.Clicked() {
			if p.Selected != nil && *p.Selected != r {
				*p.Selected = r
				changed = true
			}
			if radio {
				if len(*p.RowSelection) != 1 || (*p.RowSelection)[0] != r {
					*p.RowSelection = []int{r}
					changed = true
				}
			}
		}
		if p.OnDoubleClick != nil && line.DoubleClicked() {
			p.OnDoubleClick(r)
		}
		if p.Expand != nil && p.Parent == nil && open[r] {
			if fn := p.Expand(r); fn != nil {
				vw := *view
				panel := ui.Column(c.Key(-1 - r)).Shrink(0).Background(t.FillAlter).Draw(bottomRule(t.BorderSecondary))
				if useList {
					panel.FillWidth()
				} else {
					panel.Width(width)
				}
				panel.Children(func() {
					ui.Column(c).Width(max(vw, 0)).Translate(sx, 0).Padding(t.SpaceLG).Children(fn)
				})
			}
		}
	}

	summaryRow := func(c *ui.Context) {
		if p.Summary == nil {
			return
		}
		row := ui.Row(c).Shrink(0).MinHeight(rowH).AlignItems(ui.Stretch).Background(t.FillAlter)
		if useList {
			row.FillWidth()
		} else {
			row.Width(width)
		}
		row.Children(func() {
			for k, tc := range cols {
				fixedCell(c, k, t.FillAlter.Over(t.Surface), func(box ui.Element) {
					box.Padding(t.SpaceSM, pick[float32](tc.kind == dataCol, padX, 0))
					if tc.kind != dataCol {
						return
					}
					v := p.Summary(tc.idx)
					if v.Render == nil && v.Text == "" {
						return
					}
					if v.Render == nil && v.Color == (ui.Color{}) {
						v.Strong = true
					}
					box.Children(func() { cellContent(c, t, v, false) })
				})
			}
		})
	}

	rowsBody := func(c *ui.Context) {
		col := ui.Column(c).Width(width).Shrink(0)
		if p.Selected != nil || radio {
			col.FocusGroup(ui.Vertical)
		}
		if p.Loading {
			col.Opacity(0.5).Disabled(true)
		}
		col.Children(func() {
			if p.Loading {
				ui.Row(c).FillWidth().Height(rowH).Center().Gap(t.SpaceSM).Draw(bottomRule(t.BorderSecondary)).Children(func() {
					ui.Spinner(c).Size(16, 16).TextColor(t.Primary)
					ui.Text(c, L(c).Loading).FontSize(t.FontSize).TextColor(t.Primary)
				})
			}
			if len(shown) == 0 {
				if !p.Loading {
					ui.Column(c).Width(max(*view, 0)).Translate(sx, 0).Children(func() {
						emptyBody(c, pick(p.Empty == "", L(c).Empty, p.Empty))
					})
				}
				return
			}
			for k, tr := range shown {
				bodyRow(c, k, tr)
			}
			summaryRow(c)
		})
	}

	root.Children(func() {
		if p.Native && !grouped && nativeTableOK(columns, p) && len(shown) > 0 {
			if nativeTable(c, root, t, columns, cell, shown, rowH, p, sort, page) {
				changed = true
			}
			tablePager(c, page, total, p)
			return
		}
		frame := ui.Column(c).FillWidth().Background(t.Surface).Border(1, t.BorderSecondary).Radius(8).Clip().Role(ui.RoleTable)
		frame.Children(func() {
			if useList {
				hs := ui.Column(c).FillWidth().Shrink(0).Clip().AlignItems(ui.Start)
				hs.Children(func() { header(c) })
				n := len(shown)
				if p.Summary != nil {
					n++
				}
				st := ui.Local(root, "vlist", func() ui.ListState { return ui.ListState{} })
				st.Key = func(i int) any {
					if i < len(shown) {
						return shown[i].row
					}
					return "summary"
				}
				list := ui.List(c, st, n, func(i int) {
					if i < len(shown) {
						bodyRow(c, i, shown[i])
						return
					}
					summaryRow(c)
				}).FillWidth().MaxHeight(p.MaxHeight).RowWidth(width).TrackScroll(body)
				if virtualFill {
					list.Grow(1).MinHeight(0)
				}
				if p.Selected != nil || radio {
					list.FocusGroup(ui.Vertical)
				}
				if p.Loading {
					list.Opacity(0.5).Disabled(true)
				}
				hs.ScrollWith(list)
				if w := list.Bounds().W; w != *view {
					*view = w
					c.Invalidate()
				}
			} else if split {
				hs := ui.ScrollHorizontal(c).FillWidth().Scrollbars(false).TrackScroll(head)
				hs.Children(func() { header(c) })
				bs := ui.ScrollBoth(c).FillWidth().MaxHeight(p.MaxHeight).TrackScroll(body)
				if virtualFill {
					bs.Grow(1).MinHeight(0)
				}
				bs.Children(func() { rowsBody(c) })
				if w := bs.Bounds().W; w != *view {
					*view = w
					c.Invalidate()
				}
				if h := bs.Bounds().H; h > 0 && h != *bodyH {
					*bodyH = h
					c.Invalidate()
				}
			} else {
				sc := ui.ScrollHorizontal(c).FillWidth().TrackScroll(body)
				sc.Children(func() {
					ui.Column(c).Width(width).Shrink(0).Children(func() {
						header(c)
						rowsBody(c)
					})
				})
				if w := sc.Bounds().W; w != *view {
					*view = w
					c.Invalidate()
				}
			}
		})
		tablePager(c, page, total, p)
	})
	return changed
}

// columnLeaves is the data columns of a possibly grouped header: each
// node without Children, in order.
func columnLeaves(cols []Column) []Column {
	var out []Column
	var walk func([]Column)
	walk = func(cs []Column) {
		for _, c := range cs {
			if len(c.Children) == 0 {
				out = append(out, c)
				continue
			}
			walk(c.Children)
		}
	}
	walk(cols)
	return out
}

func groupedColumns(cols []Column) bool {
	for _, c := range cols {
		if len(c.Children) > 0 {
			return true
		}
	}
	return false
}

func leafCount(col Column) int {
	if len(col.Children) == 0 {
		return 1
	}
	n := 0
	for _, ch := range col.Children {
		n += leafCount(ch)
	}
	return n
}

func headerDepth(cols []Column) int {
	d := 1
	for _, c := range cols {
		if len(c.Children) > 0 {
			d = max(d, 1+headerDepth(c.Children))
		}
	}
	return d
}

func groupFlex(col Column) (basis, grow float32) {
	if len(col.Children) == 0 {
		if col.Width == 0 {
			return 0, 1
		}
		return col.Width, 0
	}
	for _, ch := range col.Children {
		b, g := groupFlex(ch)
		basis += b
		grow += g
	}
	return basis, grow
}

// tableGroupHeaders paints grouped header cells (colSpan via nested
// columns, rowSpan via stretch). It reports a sort or filter change.
func tableGroupHeaders(c *ui.Context, t *Tokens, groups, leaves []Column, cols []tcol, sortedCol int, sort *SortState, filters *TableFilters, padX, rowH float32, bordered bool, place func(ui.Element, int) ui.Element) bool {
	changed := false
	first := 0
	for first < len(cols) && cols[first].kind != dataCol {
		first++
	}
	maxd := headerDepth(groups)
	var paint func(cs []Column, tci, depth int) int
	paint = func(cs []Column, tci, depth int) int {
		ui.Row(c).Grow(1).MinWidth(0).AlignItems(ui.Stretch).Children(func() {
			for _, g := range cs {
				n := leafCount(g)
				basis, grow := groupFlex(g)
				box := ui.Column(c.Key(tci)).MinWidth(0).AlignItems(ui.Stretch).AlignSelf(ui.Stretch)
				if grow > 0 {
					box.Basis(basis).Grow(grow)
				} else {
					box.Width(basis + float32(n)*2*padX)
				}
				if bordered {
					box.DrawOver(rightRule(t.BorderSecondary))
				} else {
					box.DrawOver(headRule(t.BorderSecondary))
				}
				if len(g.Children) == 0 {
					box.Justify(ui.Center).Role(ui.RoleColumnHeader)
					box.Children(func() {
						h := place(ui.Row(c), tci)
						h.Role(ui.RoleColumnHeader)
						h.AlignSelf(ui.Stretch).MinHeight(rowH * float32(maxd-depth+1))
						h.Children(func() {
							if tableHeadCell(c, t, cols[tci].idx, leaves[cols[tci].idx], sortedCol == cols[tci].idx, sort, filters, padX) {
								changed = true
							}
						})
					})
					tci++
					continue
				}
				start := tci
				box.Role(ui.RoleColumnHeader)
				box.Children(func() {
					ui.Row(c).FillWidth().Height(rowH).Center().Role(ui.RoleRow).Draw(bottomRule(t.BorderSecondary)).Children(func() {
						ui.Text(c, g.Title).FontSize(t.FontSize).FontWeight(600).TextColor(t.Text).SingleLine()
					})
					tci = paint(g.Children, start, depth+1)
				})
			}
		})
		return tci
	}
	paint(groups, first, 1)
	return changed
}

func tablePager(c *ui.Context, page *int, total int, p TableProps) {
	if p.PageSize <= 0 || total <= p.PageSize {
		return
	}
	ui.Row(c).FillWidth().Justify(ui.End).Children(func() {
		Pagination(c, page, total, p.PageSize, PaginationProps{Size: Small, ShowTotal: p.ShowTotal})
	})
}

// cellContent is a cell's text, or what its Render builds. wrap lets the
// text break; maxLines caps it (0 is no cap).
func cellContent(c *ui.Context, t *Tokens, v Cell, wrap bool, maxLines ...int) {
	if v.Render != nil {
		v.Render()
		return
	}
	fg := v.Color
	if fg == (ui.Color{}) {
		fg = t.Text
	}
	txt := ui.Text(c, v.Text).FontSize(t.FontSize).TextColor(fg)
	if v.Strong {
		txt.FontWeight(600)
	}
	if !wrap {
		txt.SingleLine().Ellipsis("…").Grow(1).MinWidth(0)
		ellipsisTip(c, txt, v.Text, t.FontSize)
		return
	}
	if n := first(maxLines); n > 0 {
		txt.MaxLines(n).Ellipsis("…").Grow(1).MinWidth(0)
	}
}

// expandIcon is Ant's 17px square + / − button of an expandable or tree
// row; it reports a click.
func expandIcon(c *ui.Context, t *Tokens, open bool) bool {
	b := ui.ButtonBase(c).Label(pick(open, L(c).CollapseRow, L(c).ExpandRow)).Size(17, 17).Shrink(0).Radius(4).
		Border(1, t.Border).Background(t.Surface).FocusRing(false).Cursor(ui.CursorPointer)
	if b.Hovered() {
		b.Border(1, t.PrimaryHover)
	}
	if b.FocusVisible() {
		focusRing(b, t)
	}
	tone := pick(b.Hovered(), t.PrimaryHover, t.Text)
	b.Draw(func(pa *ui.Painter, r ui.Rect) {
		cx, cy := r.X+r.W/2, r.Y+r.H/2
		pa.Fill(ui.Rect{X: cx - 4.5, Y: cy - 0.5, W: 9, H: 1}, tone, 0)
		if !open {
			pa.Fill(ui.Rect{X: cx - 0.5, Y: cy - 4.5, W: 1, H: 9}, tone, 0)
		}
	})
	return b.Clicked()
}

// tableEditCell is an editable cell: its text, ringed on hover, and a
// field over it after a click; Enter or leaving the field hands the text
// to onEdit, Escape drops it.
func tableEditCell(c *ui.Context, t *Tokens, edit *tableEdit, row, col int, v Cell, align Align, onEdit func(row, col int, text string)) bool {
	if edit.row == row && edit.col == col {
		in := ui.TextInputBase(c, &edit.draft).Height(t.ControlHeight).Padding(0, 11).FontSize(t.FontSize).
			TextColor(t.Text).Radius(t.RadiusControl).FocusRing(false).Grow(1).MinWidth(0).Label("编辑单元格")
		fieldLook(in, t, focus, StatusNone)
		cancel := false
		in.OnTextKey(func(k ui.TextKey) bool {
			if k.Key == ui.KeyEscape {
				cancel = true
				return true
			}
			return false
		})
		if !edit.started {
			in.Focus()
			edit.started = true
			return false
		}
		if cancel {
			*edit = tableEdit{row: -1}
			return false
		}
		if in.Submitted() || !in.Focused() {
			text := edit.draft
			*edit = tableEdit{row: -1}
			if text != v.Text {
				onEdit(row, col, text)
				return true
			}
		}
		return false
	}
	b := ui.ButtonBase(c).Label(v.Text).Grow(1).MinWidth(0).MinHeight(24).Padding(0, 4).Radius(t.RadiusControl).
		AlignItems(ui.Center).Justify(justify(align)).FocusRing(false).Cursor(ui.CursorText).Margin(0, -5)
	if b.Hovered() {
		b.Border(1, t.Border)
	}
	if b.FocusVisible() {
		focusRing(b, t)
	}
	b.Children(func() { cellContent(c, t, v, false) })
	if b.Clicked() {
		*edit = tableEdit{row: row, col: col, draft: v.Text}
		c.Invalidate()
	}
	return false
}

// tableHeadCell fills a header cell: the title, a sort button when the
// column sorts and a filter funnel when it filters. It reports a change
// of the sort or the filters (which the caller pages back for).
func tableHeadCell(c *ui.Context, t *Tokens, i int, col Column, sorted bool, sort *SortState, filters *TableFilters, pad float32) bool {
	changed := false
	if col.Sorter != nil {
		if sortHeader(c, t, col, sorted, sort.Order, pad) {
			*sort = sort.next(i)
			changed = true
		}
	} else {
		ui.Text(c, col.Title).FontSize(t.FontSize).FontWeight(600).TextColor(t.Text).SingleLine().Grow(1).Shrink(1)
	}
	if len(col.Filters) > 0 {
		vals := (*filters)[i]
		if next, ok := filterMenu(c, t, col, vals); ok {
			(*filters)[i] = next
			changed = true
		}
	}
	return changed
}
