package kit

// Breakpoints of the default responsive layout, in window DIPs.
const (
	// Wide shows the settings sidebar and the one-line form rows; below it
	// the sidebar turns into chips and the rows stack.
	WideMin = 960
	// NavTileW and NavTileH are SettingTab.module.scss .tab-item 110×100.
	NavTileW, NavTileH float32 = 110, 100
	// NavTileGap and NavTilePad are the 20px gap and padding around those tiles.
	NavTileGap, NavTilePad float32 = 20, 20
	// SidebarWidth is NavList: one 110px tile plus 20px padding each side.
	SidebarWidth = NavTileW + NavTilePad*2
)

// Wide reports whether a window w wide gets the sidebar layout.
func Wide(w float32) bool { return w >= WideMin }

// HomeCols is the service grid's column count: 4 from 1180 (home layout
// .service-grid), 3 from 960, 2 from 700, else 1.
func HomeCols(w float32) int {
	switch {
	case w < 700:
		return 1
	case w < WideMin:
		return 2
	case w < 1180:
		return 3
	}
	return 4
}

// PagePad is the home page's side padding, clamp(16px, 3vw, 36px) in steps.
func PagePad(w float32) float32 {
	switch {
	case w < 800:
		return 12
	case w < 1100:
		return 20
	}
	return 36
}

// HomeTitleSize is home layout: section-head h1: clamp(24px, 2.6vw, 32px)
// in the same width steps as the service grid.
func HomeTitleSize(w float32) float32 {
	switch {
	case w >= 1180:
		return 32
	case w < 800:
		return 24
	}
	return 28
}

// HomeNavGap is the space between HomeBar TopNav items: 24 from 1180, 20
// from 960, else 16 (home layout: nav gap clamp).
func HomeNavGap(w float32) float32 {
	switch {
	case w >= 1180:
		return 24
	case w >= WideMin:
		return 20
	}
	return 16
}

// CellMin is the Min of a FormRow cell: wide on wide windows, else the
// whole form width so that cells stack one per line.
func CellMin(w, wide float32) float32 {
	if Wide(w) {
		return wide
	}
	return max(w-88, 260)
}

// FormWidth is the width of a form column in a window w wide, beside the
// sidebar when it shows, capped at limit.
func FormWidth(w float32, sidebar bool, limit float32) float32 {
	if sidebar && Wide(w) {
		w -= SidebarWidth
	}
	w -= 48
	return max(min(w, limit), 300)
}
