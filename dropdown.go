package kit

import "github.com/egoist/mygo/ui"

// MenuItem is one entry of a Dropdown menu.
type MenuItem struct {
	Key, Label string
	Icon       *ui.SVG
	// Danger colors the entry as a destructive action.
	Danger   bool
	Disabled bool
	// Divider makes the entry a line between groups; Key and Label are
	// unused.
	Divider bool
	// Children make the entry a submenu holding them, or, with Group, a
	// titled group of them (Ant's ItemGroup).
	Children []MenuItem
	Group    bool
	// Checked is a native-menu check mark (ui.MenuItem.Checked); painted
	// dropdowns ignore it. With Radio it is the chosen radio item.
	Checked bool
	// Radio draws a native-menu radio mark (ui.MenuItem.Radio).
	Radio bool
	// ShortcutMods and ShortcutKey label the accelerator: native menus
	// (ui.MenuItem.Shortcut) and the painted Ant dropdown. The view still
	// handles the key itself.
	ShortcutMods ui.Modifiers
	ShortcutKey  ui.Key
}

// DropdownProps are a dropdown's options.
type DropdownProps struct {
	// End lines the menu up with the trigger's end rather than its start.
	End bool
	// Top opens it above the trigger (below it near the window's top).
	Top bool
	// Trigger is what opens it: a click by default, or TriggerHover,
	// TriggerContextMenu (a right-click, opening at the pointer).
	Trigger Trigger
	// Arrow draws an arrow pointing at the trigger.
	Arrow bool
	// Open, when set, holds whether it is open (Ant's controlled open).
	Open *bool
	// MaxHeight, when set, scrolls a longer menu.
	MaxHeight float32
	// Native opens the system's menu (ui.Element.Menu / ContextMenu)
	// instead of the painted Ant panel. Submenus, checkmarks and
	// shortcuts work; custom icons and danger colours do not, because
	// the OS draws the items.
	Native bool

	open bool
}

// Dropdown is Ant's Dropdown: a painted in-window menu by default.
// Native: true (or AttachMenu / AttachContextMenu) opens the OS menu
// instead. trigger builds the control, as a kit Button, and returns its
// click, which opens a menu of items below it (above it near the
// window's bottom). The arrows move between entries; a click or Enter
// chooses one and closes the menu, as Escape or a click outside does.
// TriggerContextMenu is the painted counterpart of AttachContextMenu.
// It returns the chosen key, or "".
func Dropdown(c *ui.Context, items []MenuItem, trigger func() bool, props ...DropdownProps) string {
	p := first(props)
	if p.Native {
		return nativeDropdown(c, items, trigger, p)
	}
	t := Of(c)
	chosen := ""
	var width float32
	inner := false
	look := menuLook{dropdown: true, height: t.ControlHeight, padX: 12, radius: 4}
	Popup(c, trigger, func(panel ui.Element, close func()) {
		panel.Role(ui.RoleMenu)
		popupLook(panel.Padding(4).Gap(0).MinWidth(max(width, 120)), t).FocusGroup(ui.Vertical)
		panel.Children(func() {
			body := func() {
				if menuItems(c, t, look, items, &chosen, &inner) {
					close()
				}
			}
			if p.MaxHeight > 0 {
				ui.Scroll(c).MaxHeight(p.MaxHeight).FillWidth().Children(func() {
					ui.Column(c).FillWidth().Children(body)
				})
				return
			}
			body()
		})
	}, PopupProps{
		Placement: pick(p.Top, PlaceTop, PlaceBottom),
		Align:     pick(p.End, AlignLast, AlignFirst),
		Arrow:     p.Arrow,
		Trigger:   p.Trigger,
		Open:      p.Open,
		open:      p.open,
		width:     &width,
		keep:      &inner,
	})
	return chosen
}

// DropdownButton is a default button labelled label with a down arrow
// that opens a Dropdown of items; it returns the chosen key, or "".
func DropdownButton(c *ui.Context, label string, items []MenuItem, props ...DropdownProps) string {
	p := first(props)
	if p.Native {
		return nativeDropdownButton(c, label, items, ButtonProps{}, p)
	}
	t := Of(c)
	return Dropdown(c, items, func() bool {
		btn := ui.ButtonBase(c).Label(label).FocusRing(false).Height(t.ControlHeight).Padding(0, t.PaddingButton).
			Gap(t.SpaceSM).Radius(t.RadiusButton).AlignItems(ui.Center).Cursor(ui.CursorPointer)
		st := stateOf(btn, false, rest)
		bg, fg, edge := buttonColors(t, ButtonProps{}, st)
		btn.Background(bg).Border(1, edge)
		if st == focus {
			focusRing(btn, t)
		}
		btn.Children(func() {
			ui.Text(c, label).FontSize(t.ControlFont(SizeInherit)).TextColor(fg).SingleLine()
			chevron(c, fg, false)
		})
		return btn.Clicked()
	}, props...)
}

// SplitButtonProps are Ant Dropdown.Button's: the left half clicks, the
// arrow opens items.
type SplitButtonProps struct {
	Type     ButtonType
	Danger   bool
	Size     Size
	Loading  bool
	Disabled bool
	Native   bool
}

// SplitButton is Ant's Dropdown.Button: the main part labelled label
// reports clicked, the trailing arrow opens items and returns the chosen
// key.
func SplitButton(c *ui.Context, label string, items []MenuItem, props ...SplitButtonProps) (key string, clicked bool) {
	p := first(props)
	p.Size = Of(c).SizeOf(p.Size)
	p.Disabled = p.Disabled || Of(c).Disabled
	t := Of(c)
	slot := compactBegin()
	defer slot.done()
	off := p.Disabled || p.Loading
	h := t.Height(p.Size)
	bp := ButtonProps{Type: p.Type, Danger: p.Danger, Size: p.Size, Loading: p.Loading, Disabled: p.Disabled}
	row := ui.Row(c).Shrink(0).AlignItems(ui.Stretch).Height(h)
	slot.at(row)
	r := t.RadiusButton
	row.Children(func() {
		left := ui.ButtonBase(c).Label(label).Disabled(off).FocusRing(false).
			Height(h).Padding(0, t.PaddingButton).Gap(t.SpaceSM).Shrink(0).
			Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
		st := stateOf(left, p.Disabled, rest)
		if p.Loading {
			st = rest
		}
		bg, fg, edge := buttonColors(t, bp, st)
		left.Background(bg).Border(1, edge).Radius(r, 0, 0, r)
		slot.part(left, r, st == hover || st == focus || st == active || p.Danger && st == rest && p.Type == Default, edge)
		if st == focus {
			focusRing(left, t)
		}
		if p.Loading {
			left.Opacity(0.65)
		}
		left.Children(func() {
			if p.Loading {
				ui.Spinner(c).Size(t.ControlFont(p.Size), t.ControlFont(p.Size)).TextColor(fg)
			}
			ui.Text(c, label).FontSize(t.ControlFont(p.Size)).TextColor(fg).SingleLine()
		})
		if left.Clicked() && !off {
			clicked = true
		}
		caret := func() ui.Element {
			b := ui.ButtonBase(c).Label(L(c).More).Disabled(off).FocusRing(false).
				Height(h).Width(h).Center().Margin(0, 0, 0, -1).
				Cursor(pick(off, ui.CursorNotAllowed, ui.CursorPointer))
			cst := stateOf(b, p.Disabled, rest)
			cbg, cfg, cedge := buttonColors(t, bp, cst)
			b.Background(cbg).Border(1, cedge).Radius(0, r, r, 0)
			slot.part(b, r, cst == hover || cst == focus || cst == active, cedge)
			if cst == focus {
				focusRing(b, t)
			}
			b.Children(func() { chevron(c, cfg, false) })
			return b
		}
		if p.Native {
			key = AttachMenu(caret(), items)
			return
		}
		key = Dropdown(c, items, func() bool { return caret().Clicked() && !off })
	})
	return key, clicked
}

// AttachMenu makes e a system menu button filled with items
// (ui.Element.Menu). The painted Ant counterpart is Dropdown /
// DropdownButton. It returns the chosen key in the frame after a choice.
func AttachMenu(e ui.Element, items []MenuItem) string {
	chosen := ""
	e.Menu(func(m *ui.Menu) { fillNativeMenu(m, items, &chosen) })
	return chosen
}

// AttachContextMenu gives e a system context menu filled with items
// (ui.Element.ContextMenu). The painted Ant counterpart is Dropdown
// with TriggerContextMenu.
func AttachContextMenu(e ui.Element, items []MenuItem) string {
	chosen := ""
	e.ContextMenu(func(m *ui.Menu) { fillNativeMenu(m, items, &chosen) })
	return chosen
}

// AttachPopoverMenu draws items with MyGo's in-window menu (theme look,
// keyboard, not the OS menu).
func AttachPopoverMenu(e ui.Element, items []MenuItem) string {
	chosen := ""
	e.PopoverMenu(func(m *ui.Menu) { fillNativeMenu(m, items, &chosen) })
	return chosen
}

func nativeDropdown(c *ui.Context, items []MenuItem, trigger func() bool, p DropdownProps) string {
	chosen := ""
	host := ui.Row(c).Shrink(0).AlignSelf(ui.Start).AlignItems(ui.Center)
	fill := func(m *ui.Menu) { fillNativeMenu(m, items, &chosen) }
	if p.Trigger == TriggerContextMenu {
		host.ContextMenu(fill)
	} else {
		host.Menu(fill)
	}
	host.Children(func() {
		// The host is the menu button; a clickable trigger inside would
		// take the press. Paint the trigger, let the host receive it.
		ui.Row(c).IgnoreInput(true).Children(func() {
			if trigger != nil {
				trigger()
			}
		})
	})
	return chosen
}

func nativeDropdownButton(c *ui.Context, label string, items []MenuItem, bp ButtonProps, _ DropdownProps) string {
	t := Of(c)
	bp.Size = t.SizeOf(bp.Size)
	bp.Disabled = bp.Disabled || t.Disabled
	h := t.Height(bp.Size)
	btn := ui.ButtonBase(c).Label(label).Disabled(bp.Disabled).FocusRing(false).
		Height(h).Padding(0, t.PaddingButton).Gap(t.SpaceSM).Radius(t.RadiusButton).
		AlignItems(ui.Center).Cursor(pick(bp.Disabled, ui.CursorNotAllowed, ui.CursorPointer))
	st := stateOf(btn, bp.Disabled, rest)
	bg, fg, edge := buttonColors(t, bp, st)
	btn.Background(bg).Border(1, edge)
	if st == focus {
		focusRing(btn, t)
	}
	btn.Children(func() {
		ui.Text(c, label).FontSize(t.ControlFont(bp.Size)).TextColor(fg).SingleLine()
		chevron(c, fg, false)
	})
	return AttachMenu(btn, items)
}

func fillNativeMenu(m *ui.Menu, items []MenuItem, chosen *string) {
	for _, it := range items {
		switch {
		case it.Divider:
			m.Separator()
		case it.Group:
			if it.Label != "" && len(it.Children) > 0 {
				m.Submenu(it.Label, func(sub *ui.Menu) { fillNativeMenu(sub, it.Children, chosen) })
				continue
			}
			fillNativeMenu(m, it.Children, chosen)
		case len(it.Children) > 0:
			m.Submenu(it.Label, func(sub *ui.Menu) { fillNativeMenu(sub, it.Children, chosen) }).Disabled(it.Disabled)
		default:
			item := m.Item(pick(it.Label == "", it.Key, it.Label)).Disabled(it.Disabled)
			switch {
			case it.Radio:
				item.Radio(it.Checked)
			case it.Checked:
				item.Checked(true)
			}
			if it.ShortcutKey != 0 {
				item.Shortcut(it.ShortcutMods, it.ShortcutKey)
			}
			if item.Chosen() {
				*chosen = pick(it.Key == "", it.Label, it.Key)
			}
		}
	}
}
