package kit

import (
	"slices"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
)

// MenuMode is how a Menu lays out, as Ant's mode prop.
type MenuMode uint8

// Menu modes.
const (
	// MenuInline stacks the items, opening submenus in place below
	// their titles, as a sidebar.
	MenuInline MenuMode = iota
	// MenuVertical stacks the items, opening submenus beside them on
	// hover.
	MenuVertical
	// MenuHorizontal lines the items up in a bar, opening submenus below
	// them on hover and folding what does not fit behind "…".
	MenuHorizontal
)

// MenuProps are a menu's options.
type MenuProps struct {
	Mode MenuMode
	// Dark is Ant's dark theme, on the sider's #001529.
	Dark bool
	// Collapsed narrows an inline menu to its icons, 80 wide, its
	// submenus opening beside it, as in a collapsed Sider.
	Collapsed bool
	// Open, when set, holds the keys of the open inline submenus.
	Open *[]string
	// Accordion keeps one top-level submenu open at a time.
	Accordion bool
	// Indent is how far each inline level steps in; 24 when zero.
	Indent float32

	force    visual
	forceKey string
}

// menuLook is how a menu's rows are drawn.
type menuLook struct {
	dark     bool
	dropdown bool
	selected string
	height   float32
	padX     float32
	radius   float32
	force    visual
	forceKey string
}

func (l menuLook) bg(t *Tokens) ui.Color {
	if l.dark {
		return t.Layout.Sider
	}
	return t.Elevated
}

// menuHas reports whether key is it or among its descendants.
func menuHas(it MenuItem, key string) bool {
	if key == "" {
		return false
	}
	if it.Key == key && len(it.Children) == 0 && !it.Group {
		return true
	}
	for _, ch := range it.Children {
		if menuHas(ch, key) {
			return true
		}
	}
	return false
}

// menuPath is the keys of the submenus leading to key, outermost first.
func menuPath(items []MenuItem, key string) []string {
	for _, it := range items {
		if !menuHas(it, key) {
			continue
		}
		if len(it.Children) == 0 {
			return nil
		}
		path := menuPath(it.Children, key)
		if it.Group {
			return path
		}
		return append([]string{it.Key}, path...)
	}
	return nil
}

// menuColors are the background and text of a row in state st.
func menuColors(t *Tokens, l menuLook, it MenuItem, on, within bool, st visual) (ui.Color, ui.Color) {
	white := ui.Hex("#FFFFFF")
	bg, fg := ui.Color{}, pick(l.dark, t.Layout.MenuDarkText, t.Text)
	lit := st == hover || st == focus || st == active
	switch {
	case it.Disabled:
		fg = pick(l.dark, ui.RGBA(255, 255, 255, 0.25), t.TextDisabled)
	case on && it.Danger:
		bg, fg = pick(l.dark, t.Error, t.ErrorBg), pick(l.dark, white, t.Error)
	case on:
		bg, fg = pick(l.dark, t.Primary, t.PrimaryBg), pick(l.dark, white, t.Primary)
	case it.Danger && lit && (l.dropdown || l.dark):
		bg, fg = t.Error, white
	case it.Danger:
		fg = t.Error
		if lit {
			bg = t.ErrorBg
		}
	case lit:
		if l.dark {
			fg = white
		} else {
			bg = pick(l.dropdown, t.FillTertiary, t.FillSecondary)
			if st == active {
				bg = t.Fill
			}
		}
	case within:
		fg = pick(l.dark, white, t.Primary)
	}
	return bg, fg
}

// menuArrow is a submenu title's arrow.
type menuArrow uint8

const (
	menuArrowNone menuArrow = iota
	menuArrowRight
	menuArrowDown
	menuArrowUp
)

// menuRow is one row of a menu: an item, or a submenu's title.
func menuRow(c *ui.Context, t *Tokens, l menuLook, it MenuItem, indent float32, arrow menuArrow, iconOnly bool) ui.Element {
	row := ui.ButtonBase(c).Label(it.Label).Disabled(it.Disabled).FocusRing(false).
		Height(l.height).Radius(l.radius).Gap(pick[float32](l.dropdown, t.SpaceSM, 10)).AlignItems(ui.Center).
		Justify(pick(iconOnly, ui.Center, ui.Start)).Shrink(0).Role(ui.RoleMenuItem).
		Cursor(pick(it.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
	if iconOnly {
		row.FillWidth().Padding(0, 0)
	} else {
		row.FillWidth().Padding(0, l.padX, 0, max(indent, l.padX))
	}
	on := len(it.Children) == 0 && it.Key == l.selected && l.selected != ""
	within := len(it.Children) > 0 && menuHas(it, l.selected)
	st := stateOf(row, it.Disabled, pick(l.forceKey == it.Key, l.force, rest))
	bg, fg := menuColors(t, l, it, on, within, st)
	row.Background(bg)
	if st == focus {
		focusRing(row, t)
	}
	row.Children(func() {
		switch {
		case it.Icon != nil:
			ui.Icon(c, it.Icon).Size(14, 14).TextColor(fg)
		case iconOnly:
			ui.Text(c, string([]rune(it.Label)[:min(1, len([]rune(it.Label)))])).FontSize(t.FontSizeLG).TextColor(fg)
		}
		if iconOnly {
			return
		}
		ui.Text(c, it.Label).FontSize(t.FontSize).TextColor(fg).SingleLine().Grow(1).MinWidth(0)
		if acc := shortcutLabel(it.ShortcutMods, it.ShortcutKey); acc != "" {
			ui.Text(c, acc).FontSize(t.FontSizeSM).TextColor(pick(it.Disabled, t.TextDisabled, t.TextTertiary)).SingleLine().Shrink(0)
		}
		switch arrow {
		case menuArrowRight:
			IconNamed(c, "right", IconProps{Size: 10, Color: fg})
		case menuArrowDown, menuArrowUp:
			chevron(c, fg, arrow == menuArrowUp)
		}
	})
	return row
}

// shortcutLabel is the accelerator painted on an Ant menu row, matching
// the native-menu ShortcutMods / ShortcutKey field (Ctrl+C).
func shortcutLabel(mods ui.Modifiers, key ui.Key) string {
	if key == 0 {
		return ""
	}
	name := ""
	switch {
	case key >= ui.KeyA && key <= ui.KeyZ:
		name = string(rune('A' + (key - ui.KeyA)))
	case key >= ui.Key0 && key <= ui.Key9:
		name = string(rune('0' + (key - ui.Key0)))
	case key >= ui.KeyF1 && key <= ui.KeyF12:
		name = "F" + strconv.Itoa(int(key-ui.KeyF1)+1)
	case key == ui.KeyEnter:
		name = "Enter"
	case key == ui.KeyEscape:
		name = "Esc"
	case key == ui.KeyDelete:
		name = "Del"
	case key == ui.KeyBackspace:
		name = "Backspace"
	case key == ui.KeyTab:
		name = "Tab"
	case key == ui.KeySpace:
		name = "Space"
	default:
		return ""
	}
	var b strings.Builder
	if mods&ui.Ctrl != 0 {
		b.WriteString("Ctrl+")
	}
	if mods&ui.Alt != 0 {
		b.WriteString("Alt+")
	}
	if mods&ui.Shift != 0 {
		b.WriteString("Shift+")
	}
	if mods&ui.Super != 0 {
		b.WriteString("Super+")
	}
	b.WriteString(name)
	return b.String()
}

// menuGroupTitle is an item group's title.
func menuGroupTitle(c *ui.Context, t *Tokens, l menuLook, label string, indent float32) {
	ui.Text(c.Key("group-title"), label).FontSize(t.FontSize).TextColor(pick(l.dark, ui.RGBA(255, 255, 255, 0.45), t.TextTertiary)).
		Padding(8, l.padX, 8, max(indent, l.padX)).SingleLine()
}

func menuDivider(c *ui.Context, t *Tokens, l menuLook) {
	ui.Box(c).FillWidth().Height(1).Margin(pick[float32](l.dropdown, 4, 1), 0).
		Background(pick(l.dark, ui.RGBA(255, 255, 255, 0.12), t.Split))
}

// menuItems builds the entries of a popup menu, opening submenus beside
// them on hover, and reports whether one was chosen, storing its key in
// chosen. open reports whether a submenu's popup is open.
func menuItems(c *ui.Context, t *Tokens, l menuLook, items []MenuItem, chosen *string, open *bool) bool {
	done := false
	for i, it := range items {
		kc := c.Key(i)
		switch {
		case it.Divider:
			menuDivider(kc, t, l)
		case it.Group:
			ui.Column(kc).FillWidth().Children(func() {
				menuGroupTitle(kc, t, l, it.Label, 0)
				if menuItems(kc, t, l, it.Children, chosen, open) {
					done = true
				}
			})
		case len(it.Children) > 0:
			if subMenuPopup(kc, t, l, it, PlaceRight, func() ui.Element {
				return menuRow(kc, t, l, it, 0, menuArrowRight, false)
			}, chosen, open) {
				done = true
			}
		default:
			if menuRow(kc, t, l, it, 0, menuArrowNone, false).Clicked() && !it.Disabled {
				*chosen = it.Key
				done = true
			}
		}
	}
	return done
}

// subMenuPopup is a submenu opening beside (or below) its title on
// hover; it reports a choice in it.
func subMenuPopup(c *ui.Context, t *Tokens, l menuLook, it MenuItem, side Placement, titleFn func() ui.Element, chosen *string, open *bool) bool {
	done := false
	inner := false
	var title ui.Element
	o := Popup(c, func() bool {
		title = titleFn()
		return false
	}, func(panel ui.Element, close func()) {
		panel.Role(ui.RoleMenu)
		popupLook(panel.Padding(4).Gap(0).MinWidth(160), t).FocusGroup(ui.Vertical).Background(l.bg(t))
		panel.Children(func() {
			if menuItems(c, t, l, it.Children, chosen, &inner) {
				done = true
				close()
			}
		})
	}, PopupProps{Placement: side, Align: AlignFirst, Trigger: TriggerHover, Gap: 4, keep: &inner, stretch: true,
		open: l.forceKey == it.Key && l.force == hover})
	if title.Valid() {
		title.Role(ui.RoleMenuItem).Expanded(o)
	}
	if it.Disabled {
		return false
	}
	*open = *open || o
	return done
}

// Menu is Ant's Menu: always painted (inline, vertical or horizontal,
// light or Dark). The OS menu is Dropdown Native / AttachMenu; there is
// no native sidebar. A click chooses an item, storing it in *selected;
// the submenus leading to it show it. It returns the clicked key, or "".
func Menu(c *ui.Context, selected *string, items []MenuItem, props ...MenuProps) string {
	p := first(props)
	t := Of(c)
	l := menuLook{dark: p.Dark, selected: *selected, height: 40, padX: 16, radius: 8, force: p.force, forceKey: p.forceKey}
	chosen := ""
	var box ui.Element
	switch {
	case p.Mode == MenuHorizontal:
		box = ui.Row(c).FillWidth().Height(46).AlignItems(ui.Stretch).Shrink(0)
		if !p.Dark {
			box.Draw(bottomRule(t.BorderSecondary))
		}
	case p.Collapsed:
		box = ui.Column(c).Width(80).Padding(4, 0).Shrink(0)
	default:
		box = ui.Column(c).FillWidth().Padding(4, 0)
		if !p.Dark && p.Mode == MenuInline {
			box.Draw(rightRule(t.BorderSecondary))
		}
	}
	if p.Mode == MenuHorizontal {
		box.Role(ui.RoleMenuBar)
	} else {
		box.Role(ui.RoleMenu)
	}
	box.Background(pick(p.Dark, t.Layout.Sider, t.Surface)).Label("菜单")
	open := p.Open
	if open == nil {
		open = ui.Local(box, "open", func() []string { return menuPath(items, *selected) })
	}
	box.Children(func() {
		var unused bool
		switch {
		case p.Mode == MenuHorizontal:
			horizontalMenu(c, t, l, box, items, &chosen)
		case p.Mode == MenuVertical || p.Collapsed:
			for i, it := range items {
				kc := c.Key(i)
				wrap := ui.Column(kc).FillWidth().Padding(0, 4)
				wrap.Children(func() {
					switch {
					case it.Divider:
						menuDivider(kc, t, l)
					case it.Group && !p.Collapsed:
						menuGroupTitle(kc, t, l, it.Label, 0)
						menuItems(kc, t, l, it.Children, &chosen, &unused)
					case len(it.Children) > 0 || it.Group:
						subMenuPopup(kc, t, l, it, PlaceRight, func() ui.Element {
							row := menuRow(kc, t, l, it, 0, pick(p.Collapsed, menuArrowNone, menuArrowRight), p.Collapsed)
							return row
						}, &chosen, &unused)
					default:
						row := menuRow(kc, t, l, it, 0, menuArrowNone, p.Collapsed)
						if p.Collapsed {
							TooltipOn(kc, row, it.Label, TooltipProps{Placement: PlaceRight})
						}
						if row.Clicked() && !it.Disabled {
							chosen = it.Key
						}
					}
				})
			}
		default:
			indent := pick(p.Indent == 0, 24, p.Indent)
			inlineMenu(c, t, l, items, 1, indent, open, p.Accordion, &chosen)
		}
	})
	if chosen != "" {
		*selected = chosen
	}
	return chosen
}

// inlineMenu builds the rows of an inline menu at level, its open
// submenus in place under their titles.
func inlineMenu(c *ui.Context, t *Tokens, l menuLook, items []MenuItem, level int, indent float32, open *[]string, accordion bool, chosen *string) {
	for i, it := range items {
		kc := c.Key(i)
		pad := indent * float32(level)
		wrap := ui.Column(kc).FillWidth()
		wrap.Children(func() {
			switch {
			case it.Divider:
				menuDivider(kc, t, l)
			case it.Group:
				menuGroupTitle(kc, t, l, it.Label, pad)
				inlineMenu(kc, t, l, it.Children, level, indent, open, accordion, chosen)
			case len(it.Children) > 0:
				on := slices.Contains(*open, it.Key)
				row := menuRow(kc, t, l, it, pad, pick(on, menuArrowUp, menuArrowDown), false).Margin(4, 4).Expanded(on)
				if row.Clicked() && !it.Disabled {
					*open = toggleKey(*open, it.Key, accordion && level == 1)
				}
				if on {
					sub := ui.Column(kc).FillWidth().Background(pick(l.dark, t.Layout.MenuDarkSub, t.Layout.MenuSub))
					sub.Children(func() {
						inlineMenu(kc, t, l, it.Children, level+1, indent, open, accordion, chosen)
					})
				}
			default:
				row := menuRow(kc, t, l, it, pad, menuArrowNone, false).Margin(4, 4)
				if row.Clicked() && !it.Disabled {
					*chosen = it.Key
				}
			}
		})
	}
}

// menuItemWidth is how wide a horizontal menu item is.
func menuItemWidth(c *ui.Context, t *Tokens, it MenuItem) float32 {
	w, _ := c.MeasureText(0, ui.Span{Text: it.Label, Size: t.FontSize})
	w += 40
	if it.Icon != nil {
		w += 24
	}
	if len(it.Children) > 0 {
		w += 18
	}
	return w
}

// menuFit is how many items of widths fit in room, leaving room for an
// overflow item of more wide when not all do.
func menuFit(widths []float32, room, more float32) int {
	var sum float32
	for _, w := range widths {
		sum += w
	}
	if room <= 0 || sum <= room {
		return len(widths)
	}
	sum = 0
	for i, w := range widths {
		if sum+w+more > room {
			return i
		}
		sum += w
	}
	return len(widths)
}

// horizontalMenu builds a horizontal menu's bar.
func horizontalMenu(c *ui.Context, t *Tokens, l menuLook, box ui.Element, items []MenuItem, chosen *string) {
	widths := make([]float32, len(items))
	for i, it := range items {
		widths[i] = menuItemWidth(c, t, it)
	}
	n := menuFit(widths, box.Bounds().W, 56)
	var unused bool
	for i, it := range items[:n] {
		kc := c.Key(i)
		if len(it.Children) > 0 || it.Group {
			subMenuPopup(kc, t, l, it, PlaceBottom, func() ui.Element {
				return barItem(kc, t, l, it)
			}, chosen, &unused)
			continue
		}
		if barItem(kc, t, l, it).Clicked() && !it.Disabled {
			*chosen = it.Key
		}
	}
	if n < len(items) {
		more := MenuItem{Key: "__more", Label: L(c).More, Icon: LookupIcon("ellipsis"), Children: items[n:]}
		subMenuPopup(c.Key("more"), t, l, more, PlaceBottom, func() ui.Element {
			return barItem(c.Key("more"), t, l, more)
		}, chosen, &unused)
	}
}

// barItem is one item of a horizontal menu: the primary color and a
// 2px ink bar under it when chosen or hovered.
func barItem(c *ui.Context, t *Tokens, l menuLook, it MenuItem) ui.Element {
	e := ui.ButtonBase(c).Label(it.Label).Disabled(it.Disabled).FocusRing(false).FillHeight().Padding(0, 20).
		Gap(10).AlignItems(ui.Center).Radius(0).Shrink(0).Role(ui.RoleMenuItem).
		Cursor(pick(it.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
	on := (len(it.Children) == 0 && it.Key == l.selected && l.selected != "") || (len(it.Children) > 0 && menuHas(it, l.selected))
	st := stateOf(e, it.Disabled, pick(l.forceKey == it.Key, l.force, rest))
	lit := st == hover || st == focus || st == active
	fg := pick(l.dark, t.Layout.MenuDarkText, t.Text)
	switch {
	case it.Disabled:
		fg = pick(l.dark, ui.RGBA(255, 255, 255, 0.25), t.TextDisabled)
	case l.dark && on:
		e.Background(t.Primary)
		fg = ui.Hex("#FFFFFF")
	case l.dark && lit:
		fg = ui.Hex("#FFFFFF")
	case it.Danger:
		fg = t.Error
	case on || lit:
		fg = t.Primary
	}
	if !l.dark && (on || lit) && !it.Disabled {
		bar := pick(it.Danger, t.Error, t.Primary)
		e.DrawOver(func(pa *ui.Painter, r ui.Rect) {
			pa.Fill(ui.Rect{X: r.X + 20, Y: r.Y + r.H - 2, W: r.W - 40, H: 2}, bar, 0)
		})
	}
	if st == focus {
		focusRing(e, t)
	}
	e.Children(func() {
		if it.Icon != nil {
			ui.Icon(c, it.Icon).Size(14, 14).TextColor(fg)
		}
		if it.Key != "__more" {
			ui.Text(c, it.Label).FontSize(t.FontSize).TextColor(fg).SingleLine()
		}
	})
	return e
}
