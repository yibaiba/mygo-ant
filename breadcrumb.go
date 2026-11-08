package kit

import "github.com/egoist/mygo/ui"

// BreadcrumbItem is one level of a Breadcrumb.
type BreadcrumbItem struct {
	Key, Title string
	Icon       *ui.SVG
	// Menu, when set, opens a dropdown of its entries on hover, as Ant's
	// menu prop; choosing one reports its key.
	Menu []MenuItem
	// Separator overrides the one after this item (Ant's type
	// "separator" items).
	Separator string
}

// BreadcrumbProps are a breadcrumb's options.
type BreadcrumbProps struct {
	// Separator goes between items; "/" when empty.
	Separator string
}

// Breadcrumb is Ant's Breadcrumb: the items in a row, the last the
// current page in the text color, those before it links in the tertiary
// color with Separator between. It returns the key of a clicked item or
// a chosen menu entry, or "".
func Breadcrumb(c *ui.Context, items []BreadcrumbItem, props ...BreadcrumbProps) string {
	p := first(props)
	t := Of(c)
	sep := pick(p.Separator == "", "/", p.Separator)
	chosen := ""
	ui.Row(c).Wrap().AlignItems(ui.Center).Label("面包屑").Children(func() {
		for i, it := range items {
			last := i == len(items)-1
			kc := c.Key(i)
			if len(it.Menu) > 0 {
				if k := Dropdown(kc, it.Menu, func() bool {
					return crumb(kc, t, it, last, true)
				}, DropdownProps{Trigger: TriggerHover}); k != "" {
					chosen = k
				}
			} else if crumb(kc, t, it, last, false) && !last {
				chosen = it.Key
			}
			if !last {
				s := pick(it.Separator == "", sep, it.Separator)
				ui.Text(c.Key(i+1000), s).FontSize(t.FontSize).TextColor(t.TextTertiary).Margin(0, 8)
			}
		}
	})
	return chosen
}

// crumb is one breadcrumb item; it reports a click.
func crumb(c *ui.Context, t *Tokens, it BreadcrumbItem, last, menu bool) bool {
	link := !last || menu
	e := ui.ButtonBase(c).Label(it.Title).FocusRing(false).Height(22).Padding(0, 4).Margin(0, -4).
		Radius(4).Gap(4).AlignItems(ui.Center).Cursor(pick(link, ui.CursorPointer, ui.CursorDefault))
	fg := pick(last, t.Text, t.TextTertiary)
	if link && (e.Hovered() || e.FocusVisible()) {
		fg = t.Text
		e.Background(t.FillSecondary)
	}
	if e.FocusVisible() {
		focusRing(e, t)
	}
	e.Children(func() {
		if it.Icon != nil {
			ui.Icon(c, it.Icon).Size(14, 14).TextColor(fg)
		}
		if it.Title != "" {
			ui.Text(c, it.Title).FontSize(t.FontSize).TextColor(fg).SingleLine()
		}
		if menu {
			IconNamed(c, "down", IconProps{Size: 10, Color: fg})
		}
	})
	return e.Clicked()
}
