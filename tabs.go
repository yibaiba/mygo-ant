package kit

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// TabsType is the look of a Tabs bar.
type TabsType uint8

const (
	// TabsLine underlines the chosen tab with an ink bar.
	TabsLine TabsType = iota
	// TabsCard puts each tab on a grey card, the chosen one white and open
	// to the content below.
	TabsCard
	// TabsEditableCard is TabsCard with a close mark on each tab and an
	// add button after them, calling OnRemove and OnAdd.
	TabsEditableCard
)

// TabsPosition is the side of the content the bar is on.
type TabsPosition uint8

// Tab bar positions.
const (
	TabsTop TabsPosition = iota
	TabsBottom
	TabsLeft
	TabsRight
)

// TabsProps are a tab bar's options.
type TabsProps struct {
	Type TabsType
	Size Size
	// Extra builds content at the end of the bar.
	Extra func()
	// Centered centers the tabs in the bar.
	Centered bool
	// Position is the side of the content the bar sits on; on the left
	// or right the tabs stack, the ink bar beside them.
	Position TabsPosition
	// OnAdd and OnRemove are called for the add button and a tab's close
	// mark of an editable card bar; HideAdd drops the add button.
	OnAdd    func()
	OnRemove func(value string)
	HideAdd  bool

	force     visual
	forceItem int // item index + 1 that force applies to
}

// tabsHold is what a tab bar keeps between frames.
type tabsHold struct {
	refocus int
	ink     ui.Rect
	inkSeen bool
}

// tabsVisible is which of the tabs, widths wide, show in room: those
// that fit, keeping room for a "more" button of more wide when not all do
// and swapping the chosen tab in for the last that fits when it would be
// left out.
func tabsVisible(widths []float32, room, more float32, chosen int) []int {
	n := menuFit(widths, room, more)
	out := make([]int, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, i)
	}
	if n < len(widths) && chosen >= n && n > 0 {
		out[n-1] = chosen
	}
	if n == 0 && chosen >= 0 && chosen < len(widths) {
		out = append(out, chosen)
	}
	return out
}

// Tabs is Ant's Tabs bar choosing *value among items; build the chosen
// pane after it. A click or the arrows (with the focus in the bar) choose;
// disabled items are skipped. Tabs that do not fit fold into a "…"
// dropdown, and the ink bar slides to the chosen tab. It reports a
// change.
func Tabs(c *ui.Context, value *string, items []Option, props ...TabsProps) bool {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	t := Of(c)
	idx := indexOf(items, *value)
	tabs := ui.TabsBase(c, &idx, len(items))
	fontSize := pick(p.Size == Large, t.FontSizeLG, t.FontSize)
	padY := map[Size]float32{Small: 8, Middle: 12, Large: 16}[p.Size]
	card := p.Type != TabsLine
	side := p.Position == TabsLeft || p.Position == TabsRight
	list := tabs.List
	hold := ui.Local(list, "tabs", func() tabsHold { return tabsHold{refocus: -1} })
	rule := t.BorderSecondary
	switch p.Position {
	case TabsBottom:
		list.FillWidth().AlignItems(ui.Start).Draw(topRule(rule))
	case TabsLeft:
		list.Column().AlignItems(ui.Stretch).Draw(rightRule(rule))
	case TabsRight:
		list.Column().AlignItems(ui.Stretch).Draw(func(pa *ui.Painter, r ui.Rect) {
			pa.Fill(ui.Rect{X: r.X, Y: r.Y, W: 1, H: r.H}, rule, 0)
		})
	default:
		list.FillWidth().AlignItems(ui.End).Draw(bottomRule(rule))
	}
	list.Gap(pick[float32](card, 2, pick[float32](side, 0, 32)))
	closable := p.Type == TabsEditableCard
	visible := make([]int, len(items))
	for i := range items {
		visible[i] = i
	}
	if !side {
		widths := make([]float32, len(items))
		for i, it := range items {
			w, _ := c.MeasureText(0, ui.Span{Text: it.Label, Size: fontSize})
			widths[i] = w + pick[float32](card, 32+2, 32) + pick[float32](closable, 22, 0)
		}
		room := list.Bounds().W - pick[float32](closable && !p.HideAdd, 44, 0)
		visible = tabsVisible(widths, room, 48, idx)
	}
	var inkTo ui.Rect
	inkOn := false
	list.Children(func() {
		if p.Centered && !side {
			ui.Spacer(c)
		}
		for _, i := range visible {
			item := items[i]
			tab := tabs.Tab(i).Disabled(item.Disabled).FocusRing(false).AlignItems(ui.Center).Gap(6).
				Cursor(pick(item.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
			if i == hold.refocus {
				tab.Focus()
				hold.refocus = -1
			}
			on := i == idx
			st := stateOf(tab, item.Disabled, pick(p.forceItem == i+1, p.force, rest))
			fg := t.Text
			switch {
			case st == disabled:
				fg = t.TextDisabled
			case st == active:
				fg = t.PrimaryActive
			case on:
				fg = t.Primary
			case st == hover || st == focus:
				fg = t.PrimaryHover
			}
			switch {
			case card:
				tab.Padding(padY-4, t.SpaceLG).Border(1, t.BorderSecondary).Background(pick(on, t.Surface, t.FillAlter))
				switch p.Position {
				case TabsBottom:
					tab.Radius(0, 0, 8, 8)
				case TabsLeft:
					tab.Radius(8, 0, 0, 8)
				case TabsRight:
					tab.Radius(0, 8, 8, 0)
				default:
					tab.Radius(8, 8, 0, 0)
				}
				if on {
					pos, bg := p.Position, t.Surface
					tab.DrawOver(func(pa *ui.Painter, r ui.Rect) {
						switch pos {
						case TabsBottom:
							pa.Fill(ui.Rect{X: r.X + 1, Y: r.Y, W: r.W - 2, H: 1}, bg, 0)
						case TabsLeft:
							pa.Fill(ui.Rect{X: r.X + r.W - 1, Y: r.Y + 1, W: 1, H: r.H - 2}, bg, 0)
						case TabsRight:
							pa.Fill(ui.Rect{X: r.X, Y: r.Y + 1, W: 1, H: r.H - 2}, bg, 0)
						default:
							pa.Fill(ui.Rect{X: r.X + 1, Y: r.Y + r.H - 1, W: r.W - 2, H: 1}, bg, 0)
						}
					})
				}
			case side:
				tab.Padding(8, 24)
			default:
				tab.Padding(padY, 0)
			}
			if on && !card {
				inkTo, inkOn = tab.Bounds(), true
			}
			if st == focus {
				focusRing(tab, t)
			}
			label := item.Label
			tab.Children(func() {
				txt := ui.Text(c, label).FontSize(fontSize).TextColor(fg).SingleLine()
				if on && !card {
					txt.FontWeight(500)
				}
				if closable && !item.Disabled {
					if iconButton(c.Key(i), LookupIcon("close"), L(c).Close+" "+label, t.TextTertiary, 12) && p.OnRemove != nil {
						p.OnRemove(item.Value)
					}
				}
			})
		}
		if len(visible) < len(items) {
			var rest []MenuItem
			shown := map[int]bool{}
			for _, i := range visible {
				shown[i] = true
			}
			for i, it := range items {
				if !shown[i] {
					rest = append(rest, MenuItem{Key: it.Value, Label: it.Label, Disabled: it.Disabled})
				}
			}
			ui.Row(c).AlignSelf(ui.Center).Padding(0, 8).Children(func() {
				if k := Dropdown(c, rest, func() bool {
					return iconButton(c, LookupIcon("ellipsis"), L(c).MoreTabs, t.Text, 16)
				}, DropdownProps{End: true, MaxHeight: 200}); k != "" {
					*value = k
					c.Invalidate()
				}
			})
		}
		if closable && !p.HideAdd {
			add := ui.ButtonBase(c).Label(L(c).AddTab).FocusRing(false).Size(40, 40).Center().Radius(8, 8, 0, 0).
				Border(1, t.BorderSecondary).Background(t.FillAlter).Cursor(ui.CursorPointer).Shrink(0)
			add.Children(func() { IconNamed(c, "plus", IconProps{Size: 14, Color: pick(add.Hovered(), t.PrimaryHover, t.Text)}) })
			if add.Clicked() && p.OnAdd != nil {
				p.OnAdd()
			}
		}
		if p.Centered && !side {
			ui.Spacer(c)
		}
		if p.Extra != nil {
			if !side && !p.Centered {
				ui.Spacer(c)
			}
			ui.Row(c).AlignSelf(ui.Center).Gap(t.SpaceSM).Children(p.Extra)
		}
	})
	if inkOn {
		lb := list.Bounds()
		to := ui.Rect{X: inkTo.X - lb.X, Y: inkTo.Y - lb.Y, W: inkTo.W, H: inkTo.H}
		anim := ui.Animation{Duration: motionFor(c, 300*time.Millisecond), Ease: ui.EaseOut}
		if !hold.inkSeen && to.W > 0 {
			hold.inkSeen = true
			list.Jump("inkx", to.X)
			list.Jump("inky", to.Y)
			list.Jump("inkw", to.W)
			list.Jump("inkh", to.H)
		}
		ink := ui.Rect{X: list.AnimateTo("inkx", to.X, anim), Y: list.AnimateTo("inky", to.Y, anim),
			W: list.AnimateTo("inkw", to.W, anim), H: list.AnimateTo("inkh", to.H, anim)}
		pos, color := p.Position, t.Primary
		if to.W > 0 {
			list.DrawOver(func(pa *ui.Painter, r ui.Rect) {
				switch pos {
				case TabsBottom:
					pa.Fill(ui.Rect{X: r.X + ink.X, Y: r.Y, W: ink.W, H: 2}, color, 0)
				case TabsLeft:
					pa.Fill(ui.Rect{X: r.X + r.W - 2, Y: r.Y + ink.Y, W: 2, H: ink.H}, color, 0)
				case TabsRight:
					pa.Fill(ui.Rect{X: r.X, Y: r.Y + ink.Y, W: 2, H: ink.H}, color, 0)
				default:
					pa.Fill(ui.Rect{X: r.X + ink.X, Y: r.Y + r.H - 2, W: ink.W, H: 2}, color, 0)
				}
			})
		}
	}
	if !tabs.List.Changed() || idx < 0 || idx >= len(items) {
		return false
	}
	if items[idx].Disabled {
		idx = skipDisabled(items, indexOf(items, *value), idx)
		if idx < 0 {
			return false
		}
		// The keys put the focus on the disabled tab; move it on with the choice.
		hold.refocus = idx
		c.Invalidate()
	}
	*value = items[idx].Value
	return true
}

// skipDisabled moves on from the disabled item to the next enabled one
// the way the keys went from prev, or -1 when there is none.
func skipDisabled(items []Option, prev, to int) int {
	dir := 1
	if prev >= 0 && to < prev {
		dir = -1
	}
	for i := to; i >= 0 && i < len(items); i += dir {
		if !items[i].Disabled {
			return pick(i == prev, -1, i)
		}
	}
	return -1
}
