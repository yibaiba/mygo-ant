package kit

import (
	"os"
	"runtime"
	"sync"

	"github.com/egoist/mygo/ui"
)

var (
	registryMu sync.RWMutex
	registry   = map[*ui.Theme]*Tokens{}
)

// Theme returns the ui.Theme of the tokens, so that ui's own widgets
// (scroll bars, toasts, spinners) and Of agree with the kit.
func (t *Tokens) Theme() *ui.Theme {
	registryMu.Lock()
	defer registryMu.Unlock()
	if t.theme != nil {
		return t.theme
	}
	th := ui.LightTheme()
	if t.Dark {
		th = ui.DarkTheme()
	}
	th.Background = t.PageBg
	th.Surface = t.Surface
	th.SurfaceHover = t.PrimaryBg
	th.SurfacePressed = t.PrimaryBorder
	th.Border = t.Border
	th.Text = t.Text
	th.TextMuted = t.TextMuted
	th.Accent = t.Primary
	th.AccentHover = t.PrimaryHover
	th.AccentPressed = t.PrimaryActive
	th.AccentText = t.TextOnPrimary
	th.Success = t.Success
	th.Danger = t.Error
	th.Warning = t.Warning
	th.Focus = t.Primary
	th.Radius = t.RadiusControl
	th.Spacing = 4
	th.FontSize = t.FontSize
	th.Font = t.Font
	if t == Home {
		th.Scrollbar = ui.Hex("#7EB3F5") // home layout: service-list thumb
	}
	t.theme = th
	registry[th] = t
	return th
}

// Apply themes the window with t; components built after it read t.
// Ant tokens stay as they are. Mix OS HighContrast / TextScale with
// Apply(c, FollowDesktop(c, t)) or ConfigProps.FollowDesktop.
func Apply(c *ui.Context, t *Tokens) {
	c.SetTheme(t.Theme())
}

// Of returns the tokens of the theme in effect, Work when it is not one
// of the kit's.
func Of(c *ui.Context) *Tokens {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if t, ok := registry[c.Theme()]; ok {
		return t
	}
	return Work
}

// Page is the root column of a chrome: its language, font and page color.
func Page(c *ui.Context, t *Tokens, fn func()) ui.Element {
	Apply(c, t)
	return ui.Column(c).Fill().Lang(t.Lang).Font(t.Font).Background(t.PageBg).Children(fn)
}

var fontFamilies = registerFonts()

// fontFaces are CJK faces registered as fallbacks on Linux. They are not
// the kit's first family; systemUIStack is.
var fontFaces = []struct{ path, name string }{
	{"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc", ""},
	{"/usr/share/fonts/truetype/wqy/wqy-microhei.ttf", ProductFont},
	{"/usr/share/fonts/truetype/droid/DroidSansFallbackFull.ttf", "Droid Sans Fallback"},
}

// ProductFont is the CJK fallback the kit registers on Linux. It is not
// the first family in FontList.
const ProductFont = "WenQuanYi Micro Hei"

func registerFonts() []string {
	var out []string
	for _, face := range fontFaces {
		data, err := os.ReadFile(face.path)
		if err != nil || ui.RegisterFont(data, face.name) != nil {
			continue
		}
		name := face.name
		if name == "" {
			name = ProductFont
		}
		out = appendUnique(out, name)
	}
	return out
}

// systemUIStack is the kit default text family list, without turning
// FollowDesktop on: Windows Segoe UI / 微软雅黑, macOS SF Pro / PingFang,
// Linux system-ui with WenQuanYi only as a later fallback. Page applies
// this on the chrome root; a control that already called Font() keeps it.
func systemUIStack() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"Segoe UI Variable", "Segoe UI", "Microsoft YaHei", "微软雅黑"}
	case "darwin":
		return []string{"SF Pro Text", "SF Pro", "PingFang SC", "PingFang HK"}
	default:
		return []string{"system-ui", "sans-serif"}
	}
}

func appendUnique(list []string, name string) []string {
	for _, have := range list {
		if have == name {
			return list
		}
	}
	return append(list, name)
}

// FontList is the font family list every chrome uses: system UI first,
// then registered CJK fallbacks (WenQuanYi is never first).
func FontList() string {
	names := append([]string(nil), systemUIStack()...)
	for _, name := range fontFamilies {
		names = appendUnique(names, name)
	}
	if len(names) == 0 {
		return ProductFont
	}
	out := names[0]
	for _, name := range names[1:] {
		out += ", " + name
	}
	return out
}
