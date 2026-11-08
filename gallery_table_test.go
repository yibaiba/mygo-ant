package kit

import (
	"strconv"
	"testing"

	"github.com/egoist/mygo/ui"
)

type meetingRow struct {
	name, room, owner, state string
	seats                    int
	parent                   int
}

var meetingRows = []meetingRow{
	{"季度经营分析会", "一号会议室", "张三", "进行中", 24, -1},
	{"财务预算评审", "二号会议室", "李四", "待开始", 12, -1},
	{"安全保密培训", "报告厅", "王五", "已结束", 120, -1},
	{"项目周例会", "三号会议室", "赵六", "待开始", 8, -1},
	{"议题一：营收", "一号会议室", "张三", "进行中", 24, 0},
	{"议题二：成本", "一号会议室", "钱七", "待开始", 24, 0},
	{"成本明细", "一号会议室", "钱七", "待开始", 24, 5},
}

func meetingCell(r, col int) Cell {
	m := meetingRows[r]
	switch col {
	case 0:
		return Cell{Text: m.name}
	case 1:
		return Cell{Text: m.room}
	case 2:
		return Cell{Text: m.owner}
	case 3:
		tone := map[string]ui.Color{"进行中": Work.Success, "待开始": Work.Primary, "已结束": Work.TextTertiary}[m.state]
		return Cell{Text: m.state, Color: tone}
	case 4:
		return Cell{Text: strconv.Itoa(m.seats)}
	}
	return Cell{Text: "详情"}
}

func meetingColumns(fixed bool) []Column {
	cols := []Column{
		{Title: "会议", Width: 160, Sorter: ByText(func(r int) string { return meetingRows[r].name })},
		{Title: "会议室", Width: 140, Filters: []Option{{Value: "一号会议室", Label: "一号会议室"}, {Value: "二号会议室", Label: "二号会议室"}, {Value: "报告厅", Label: "报告厅"}},
			FilterSearch: true, OnFilter: FilterEquals(func(r int) string { return meetingRows[r].room })},
		{Title: "负责人", Width: 120, Editable: true},
		{Title: "状态", Width: 120, Filters: []Option{{Value: "进行中", Label: "进行中"}, {Value: "待开始", Label: "待开始"}, {Value: "已结束", Label: "已结束"}},
			OnFilter: FilterEquals(func(r int) string { return meetingRows[r].state })},
		{Title: "席位", Width: 100, Align: AlignEnd, Sorter: ByNumber(func(r int) float64 { return float64(meetingRows[r].seats) })},
		{Title: "操作", Width: 90},
	}
	if fixed {
		cols[0].Fixed = FixedLeft
		cols[5].Fixed = FixedRight
	}
	return cols
}

type galleryTable struct {
	tok      *Tokens
	checked  []int
	expanded []int
	tree     []int
	filters  TableFilters
	owners   []string
	sel      int
}

func newGalleryTable() *galleryTable {
	g := &galleryTable{tok: Work, checked: []int{1}, expanded: []int{0}, tree: []int{0}, filters: TableFilters{3: {"进行中", "待开始"}}, sel: -1}
	for _, m := range meetingRows {
		g.owners = append(g.owners, m.owner)
	}
	return g
}

func (g *galleryTable) page(c *ui.Context) {
	tok := g.tok
	cell := func(r, col int) Cell {
		if col == 2 {
			return Cell{Text: g.owners[r]}
		}
		return meetingCell(r, col)
	}
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 表格").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(8, 0).Children(func() {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Children(func() {
				SectionTitle(c, "多选 · 筛选 · 排序 · 可编辑 · 汇总")
				Sheet(c, func() {
					Table(c, meetingColumns(false)[:5], 4, cell, TableProps{RowSelection: &g.checked, Filters: &g.filters,
						RowDisabled: func(r int) bool { return r == 2 },
						OnEdit:      func(r, _ int, s string) { g.owners[r] = s },
						Summary: func(col int) Cell {
							switch col {
							case 0:
								return Cell{Text: "合计"}
							case 4:
								return Cell{Text: "164"}
							}
							return Cell{}
						}})
				})
				SectionTitle(c, "可展开行 · 带边框")
				Sheet(c, func() {
					Table(c, meetingColumns(false)[:4], 3, cell, TableProps{Bordered: true, Expanded: &g.expanded,
						Expand: func(r int) func() {
							if r == 2 {
								return nil
							}
							return func() {
								Muted(c, meetingRows[r].name+" · 议程已发布，参会人 "+strconv.Itoa(meetingRows[r].seats)+" 人")
							}
						}})
				})
				SectionTitle(c, "树形数据")
				Sheet(c, func() {
					Table(c, meetingColumns(false)[:5], len(meetingRows), cell, TableProps{Expanded: &g.tree,
						Parent: func(r int) int { return meetingRows[r].parent }})
				})
			})
			ui.Column(c).Width(520).Shrink(0).Children(func() {
				SectionTitle(c, "固定列 · 横向滚动")
				Sheet(c, func() {
					Table(c, meetingColumns(true), 4, cell, TableProps{Selected: &g.sel})
				})
				SectionTitle(c, "虚拟滚动 · 10000 行 · 固定表头")
				Sheet(c, func() {
					Table(c, []Column{{Title: "#", Width: 80}, {Title: "终端"}, {Title: "地址", Width: 140}}, 10000, func(r, col int) Cell {
						switch col {
						case 0:
							return Cell{Text: strconv.Itoa(r + 1)}
						case 1:
							return Cell{Text: "会议终端 " + strconv.Itoa(r+1)}
						}
						return Cell{Text: "10.0." + strconv.Itoa(r/254) + "." + strconv.Itoa(r%254+1)}
					}, TableProps{Virtual: true, MaxHeight: 240, Size: Small})
				})
				SectionTitle(c, "筛选菜单")
				Sheet(c, func() {
					ui.Column(c).FillWidth().Height(260).Children(func() {
						filterCol := meetingColumns(false)[1]
						filterCol.Title = "地点"
						Table(c, []Column{filterCol, {Title: "说明"}}, 0, meetingCell, TableProps{Filters: &g.filters})
					})
				})
			})
		})
	})
}

func TestGalleryTableShots(t *testing.T) {
	g := newGalleryTable()
	page := ui.NewTester(g.page, 1440, 1240)
	page.SetLocale("zh-CN")
	page.Frame()
	page.Frame()
	for _, want := range []string{"合计", "季度经营分析会 · 议程已发布，参会人 24 人", "议题一：营收", "会议终端 1"} {
		if !page.HasText(want) {
			t.Fatalf("table gallery missing %q", want)
		}
	}
	r, _ := page.Find("详情")
	page.Scroll(r.X-150, r.Y, 200, 0)
	page.Frame()
	if err := page.Click("筛选 地点"); err != nil {
		t.Fatal(err)
	}
	page.Frame()
	shot(t, page, "mygo-kit-gallery-table.png")
}
