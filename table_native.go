package kit

import (
	"strconv"

	"github.com/egoist/mygo/ui"
)

// nativeTableOK reports whether p can use ui.Table: the painted extras
// (expand, tree, checkbox/radio selection, summary, edit, filters, fixed
// and wrapping columns) have no native counterpart in this replica.
func nativeTableOK(columns []Column, p TableProps) bool {
	if p.Expand != nil || p.Parent != nil || p.RowSelection != nil || p.Summary != nil || p.OnEdit != nil {
		return false
	}
	for _, col := range columns {
		if col.Editable || len(col.Filters) > 0 || col.Fixed != FixedNone || col.Wrap || len(col.Children) > 0 {
			return false
		}
	}
	return true
}

func nativeColID(i int) string { return strconv.Itoa(i) }

// nativeTable is Table's ui.Table path: OS-style resize/reorder/sort and
// a virtual body. *Selected is the caller's global row index; ui.Table's
// Selected is the index on this page. The list grows to fill unless given
// an explicit height, so the body is sized to the page.
func nativeTable(c *ui.Context, root ui.Element, t *Tokens, columns []Column, cell func(row, col int) Cell, shown []tableRow, rowH float32, p TableProps, sort *SortState, page *int) bool {
	st := ui.Local(root, "nlist", func() ui.ListState { return ui.ListState{} })
	nsel := ui.Local(root, "nsel", func() int { return -1 })
	nsort := ui.Local(root, "nsort", func() ui.SortOrder { return ui.SortOrder{} })
	rows := make([]int, len(shown))
	for i, tr := range shown {
		rows[i] = tr.row
	}
	*nsel = -1
	if p.Selected != nil {
		for i, r := range rows {
			if r == *p.Selected {
				*nsel = i
				break
			}
		}
		st.Selected = nsel
	} else {
		st.Selected = nil
	}
	if sort != nil && sort.Col >= 0 && sort.Col < len(columns) && sort.Order != SortNone {
		*nsort = ui.SortOrder{Column: nativeColID(sort.Col), Descending: sort.Order == SortDescend}
	} else {
		*nsort = ui.SortOrder{}
	}
	st.Sort = nsort
	st.Key = func(row int) any {
		if row >= 0 && row < len(rows) {
			return rows[row]
		}
		return row
	}
	st.Label = func(row int) string {
		if row < 0 || row >= len(rows) || len(columns) == 0 {
			return ""
		}
		return cell(rows[row], 0).Text
	}
	cols := make([]ui.TableColumn, len(columns))
	for i, col := range columns {
		cols[i] = ui.TableColumn{
			Title:    col.Title,
			ID:       nativeColID(i),
			Width:    col.Width,
			MinWidth: col.Min,
			Align:    justify(col.Align),
			Sortable: col.Sorter != nil,
		}
	}
	n := len(rows)
	bodyRows := max(n, 1)
	if p.PageSize > 0 {
		bodyRows = min(max(n, 1), p.PageSize)
	}
	h := rowH * float32(bodyRows+1)
	if p.MaxHeight > 0 {
		h = min(h, p.MaxHeight)
	}
	h = max(h, rowH*2)
	changed := false
	body := func() {
		tbl := ui.Table(c, st, cols, n, func(row, col int) {
			if row < 0 || row >= len(rows) || col < 0 || col >= len(columns) {
				return
			}
			host := ui.Box(c).FillWidth()
			host.Children(func() {
				cellContent(c, t, cell(rows[row], col), false)
			})
			if p.OnDoubleClick != nil && host.RowDoubleClicked() {
				p.OnDoubleClick(rows[row])
			}
		}).FillWidth().Height(h)
		if tbl.Changed() && p.Selected != nil {
			if *nsel >= 0 && *nsel < len(rows) && *p.Selected != rows[*nsel] {
				*p.Selected = rows[*nsel]
				changed = true
			}
		}
	}
	if p.Loading {
		SpinOver(c, true, body)
	} else {
		body()
	}
	if sort != nil {
		next := SortState{Col: -1}
		if nsort.Column != "" {
			if i, err := strconv.Atoi(nsort.Column); err == nil && i >= 0 && i < len(columns) {
				next.Col = i
				next.Order = pick(nsort.Descending, SortDescend, SortAscend)
			}
		}
		if next.Col != sort.Col || next.Order != sort.Order {
			*sort = next
			if page != nil {
				*page = 1
			}
			changed = true
			c.Invalidate()
		}
	}
	return changed
}
