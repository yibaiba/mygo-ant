package kit

import (
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/egoist/mygo/ui"
)

// iconRegistry holds the named icons IconNamed draws: the kit's outlined
// set, the generated @ant-design/icons glyphs, and what the app registers.
// It is filled as a package variable, not in init, so that package
// variables of the app (menus, nav items) can look icons up.
var iconRegistry = builtinIcons()

type iconStore struct {
	sync.RWMutex
	byName map[string]*ui.SVG
	twin   map[string]*ui.SVG
	src    map[string]*iconSrc
}

// iconSrc is an unparsed glyph; parse runs once on Lookup.
type iconSrc struct {
	viewBox, primary, fill string
	kit                    bool
	once                   sync.Once
	svg, twinSVG           *ui.SVG
}

func (s *iconSrc) parse() (*ui.SVG, *ui.SVG) {
	s.once.Do(func() {
		if s.kit {
			s.svg = svgIcon(s.primary)
			if s.fill != "" {
				s.twinSVG = svgIcon(s.fill)
			}
			return
		}
		s.svg = antSVG(s.viewBox, s.primary)
		if s.fill != "" {
			s.twinSVG = antSVG(s.viewBox, s.fill)
		}
	})
	return s.svg, s.twinSVG
}

// RegisterIcon adds (or replaces) the icon name, drawn in one color. A
// two-tone icon also gives fill, the part painted in the light second
// tone under it; nil for none.
func RegisterIcon(name string, svg, fill *ui.SVG) {
	iconRegistry.Lock()
	defer iconRegistry.Unlock()
	iconRegistry.byName[name] = svg
	if fill != nil {
		iconRegistry.twin[name] = fill
	} else {
		delete(iconRegistry.twin, name)
	}
}

// LookupIcon returns the icon registered as name, or nil. Names are Ant's
// (SearchOutlined) and the kit's short aliases (search).
func LookupIcon(name string) *ui.SVG {
	svg, _ := iconRegistry.lookupPair(name, IconOutlined)
	return svg
}

// IconNames lists the registered icons, sorted.
func IconNames() []string {
	iconRegistry.RLock()
	defer iconRegistry.RUnlock()
	seen := make(map[string]struct{}, len(iconRegistry.src)+len(iconRegistry.byName))
	for k := range iconRegistry.byName {
		seen[k] = struct{}{}
	}
	for k := range iconRegistry.src {
		seen[k] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// IconTheme is Ant's icon theme: outlined by default, or filled / two-tone.
type IconTheme uint8

// Icon themes, as @ant-design/icons theme.
const (
	IconOutlined IconTheme = iota
	IconFilled
	IconTwoTone
)

// IconProps are a named icon's options. kit.Icon is the same type, so
// kit.Icon{Name: "SearchOutlined"} is the Ant call.
type IconProps struct {
	// Name is the glyph, as Ant: "SearchOutlined", "HeartFilled",
	// "CheckCircleTwoTone". Short kit names ("search") still work.
	Name string
	// Size is its width and height; 16 when zero.
	Size float32
	// Color tints it (the text color by default); TwoTone, when set,
	// paints its fill part in the light tone of that color, as Ant's
	// twoToneColor.
	Color   ui.Color
	TwoTone ui.Color
	// Theme picks Outlined / Filled / TwoTone when Name has no suffix
	// (Name: "Search", Theme: IconFilled → SearchFilled).
	Theme IconTheme
	// Spin turns it once a second, as a loading mark; Rotate turns it by
	// that many degrees.
	Spin   bool
	Rotate float32
}

// Icon is IconProps, so callers write kit.Icon{Name: "SearchOutlined"}.
type Icon = IconProps

// IconNamed draws the registered icon name, as Ant's icon components; an
// unknown name draws nothing in its place. An empty name uses props.Name.
func IconNamed(c *ui.Context, name string, props ...IconProps) ui.Element {
	p := first(props)
	if name == "" {
		name = p.Name
	}
	t := Of(c)
	size := pick(p.Size == 0, 16, p.Size)
	tone := p.Color
	if tone == (ui.Color{}) {
		tone = t.Text
	}
	svg, fill := iconRegistry.lookupPair(name, p.Theme)
	box := ui.Box(c).Size(size, size).Shrink(0)
	if svg == nil {
		return box
	}
	turn := p.Rotate
	if p.Spin {
		turn += box.Loop("spin", time.Second, ui.Linear) * 360
	}
	two := fill != nil && (p.TwoTone != (ui.Color{}) || p.Theme == IconTwoTone || strings.Contains(strings.ToLower(name), "twotone"))
	primary := tone
	if p.TwoTone != (ui.Color{}) {
		primary = p.TwoTone
	}
	box.Draw(func(pa *ui.Painter, r ui.Rect) {
		draw := func() {
			if two {
				pa.Icon(fill, r, AntPalette(primary)[1])
				pa.Icon(svg, r, primary)
				return
			}
			pa.Icon(svg, r, tone)
		}
		if turn == 0 {
			draw()
			return
		}
		pa.Transform(ui.Rotation(turn), r.X+r.W/2, r.Y+r.H/2, draw)
	})
	return box
}

func (s *iconStore) lookupPair(name string, theme IconTheme) (svg, fill *ui.SVG) {
	s.Lock()
	defer s.Unlock()
	name = s.resolveLocked(name, theme)
	return s.ensureLocked(name)
}

func (s *iconStore) hasLocked(name string) bool {
	if s.byName[name] != nil {
		return true
	}
	_, ok := s.src[name]
	return ok
}

func (s *iconStore) ensureLocked(name string) (*ui.SVG, *ui.SVG) {
	if svg := s.byName[name]; svg != nil {
		return svg, s.twin[name]
	}
	src := s.src[name]
	if src == nil {
		return nil, nil
	}
	svg, fill := src.parse()
	if svg != nil {
		s.byName[name] = svg
	}
	if fill != nil {
		s.twin[name] = fill
	}
	return svg, fill
}

// resolveIconNameLocked maps Ant names, short aliases and Theme onto a
// registered key. The registry lock must be held.
func resolveIconNameLocked(name string, theme IconTheme) string {
	return iconRegistry.resolveLocked(name, theme)
}

func (s *iconStore) resolveLocked(name string, theme IconTheme) string {
	if name == "" {
		return name
	}
	if theme != IconOutlined {
		if cand := applyIconTheme(name, theme); s.hasLocked(cand) {
			return cand
		}
	}
	if s.hasLocked(name) {
		return name
	}
	if cand := applyIconTheme(name, theme); s.hasLocked(cand) {
		return cand
	}
	if theme != IconOutlined {
		if cand := applyIconTheme(name, IconOutlined); s.hasLocked(cand) {
			return cand
		}
	}
	return name
}

func applyIconTheme(name string, theme IconTheme) string {
	base, encoded := splitAntIconName(name)
	if encoded != "" && theme == IconOutlined {
		return name
	}
	p := kebabToPascal(base)
	switch theme {
	case IconFilled:
		return p + "Filled"
	case IconTwoTone:
		return p + "TwoTone"
	default:
		return p + "Outlined"
	}
}

func splitAntIconName(name string) (base, theme string) {
	switch {
	case strings.HasSuffix(name, "TwoTone"):
		return name[:len(name)-7], "twotone"
	case strings.HasSuffix(name, "Outlined"):
		return name[:len(name)-8], "outlined"
	case strings.HasSuffix(name, "Filled"):
		return name[:len(name)-6], "filled"
	case strings.HasSuffix(name, "-twotone"):
		return name[:len(name)-8], "twotone"
	case strings.HasSuffix(name, "-filled"):
		return name[:len(name)-7], "filled"
	}
	return name, ""
}

func kebabToPascal(s string) string {
	parts := strings.Split(s, "-")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	return b.String()
}

func pascalToKebab(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte('-')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func builtinIcons() *iconStore {
	reg := &iconStore{byName: map[string]*ui.SVG{}, twin: map[string]*ui.SVG{}, src: map[string]*iconSrc{}}
	icons := map[string]string{
		"search":             `<circle cx="7" cy="7" r="4.6"/><path d="m10.5 10.5 3.5 3.5"/>`,
		"close":              `<path d="m3.5 3.5 9 9m0-9-9 9"/>`,
		"check":              `<path d="m2.8 8.4 3.4 3.3 7-7.2"/>`,
		"plus":               `<path d="M8 2.5v11M2.5 8h11"/>`,
		"minus-circle":       `<circle cx="8" cy="8" r="6"/><path d="M5.2 8h5.6"/>`,
		"paper-clip":         `<path d="M10.5 4.5 5.2 9.8a1.3 1.3 0 0 0 1.8 1.8l5.6-5.6a2.6 2.6 0 0 0-3.7-3.7L3.3 7.9a3.9 3.9 0 0 0 5.5 5.5l4.7-4.7"/>`,
		"picture":            `<rect x="2" y="3" width="12" height="10" rx="1"/><circle cx="5.6" cy="6.4" r="1.1"/><path d="m2.5 11.5 3.6-3.4 2.4 2.2 2.2-1.9 2.8 2.6"/>`,
		"minus":              `<path d="M2.5 8h11"/>`,
		"edit":               `<path d="M3 13h2.5l7-7L10 3.5l-7 7z"/><path d="m8.8 4.7 2.5 2.5"/>`,
		"delete":             `<path d="M2.5 4h11M6 4V2.6h4V4M4 4l.7 9.4h6.6L12 4M6.6 6.5v4.5M9.4 6.5v4.5"/>`,
		"copy":               `<rect x="5" y="5" width="8.5" height="8.5" rx="1"/><path d="M11 5V3a.5.5 0 0 0-.5-.5h-7a.5.5 0 0 0-.5.5v7a.5.5 0 0 0 .5.5H5"/>`,
		"setting":            `<circle cx="8" cy="8" r="2.2"/><path d="M8 1.8v1.8M8 12.4v1.8M1.8 8h1.8M12.4 8h1.8M3.6 3.6l1.3 1.3M11.1 11.1l1.3 1.3M3.6 12.4l1.3-1.3M11.1 4.9l1.3-1.3"/>`,
		"user":               `<circle cx="8" cy="5.5" r="2.8"/><path d="M2.8 14c.6-2.9 2.7-4.4 5.2-4.4s4.6 1.5 5.2 4.4"/>`,
		"team":               `<circle cx="6" cy="5.5" r="2.3"/><path d="M1.8 13c.5-2.4 2.2-3.6 4.2-3.6s3.7 1.2 4.2 3.6"/><path d="M10.4 3.4a2.3 2.3 0 0 1 0 4.4M12 9.6c1.2.5 2 1.6 2.3 3.4"/>`,
		"home":               `<path d="M2.5 7.5 8 2.8l5.5 4.7M4 6.4V13.5h3V10h2v3.5h3V6.4"/>`,
		"menu-fold":          `<path d="M2.5 3.5h11M7 8h6.5M2.5 12.5h11M5 6 3 8l2 2"/>`,
		"menu-unfold":        `<path d="M2.5 3.5h11M7 8h6.5M2.5 12.5h11M3 6l2 2-2 2"/>`,
		"up":                 `<path d="m3 10.5 5-5 5 5"/>`,
		"down":               `<path d="m3 5.5 5 5 5-5"/>`,
		"left":               `<path d="M10.5 3 5.5 8l5 5"/>`,
		"right":              `<path d="m5.5 3 5 5-5 5"/>`,
		"double-left":        `<path d="M8 3 3 8l5 5M13 3 8 8l5 5"/>`,
		"double-right":       `<path d="m3 3 5 5-5 5M8 3l5 5-5 5"/>`,
		"arrow-up":           `<path d="M8 13.5v-11M3.5 7 8 2.5 12.5 7"/>`,
		"arrow-down":         `<path d="M8 2.5v11M3.5 9 8 13.5 12.5 9"/>`,
		"vertical-align-top": `<path d="M2.5 2.5h11M8 13.5V5.5M4.5 9 8 5.5 11.5 9"/>`,
		"info-circle":        `<circle cx="8" cy="8" r="6"/><path d="M8 7.2v4"/><path d="M8 4.8v.2"/>`,
		"check-circle":       `<circle cx="8" cy="8" r="6"/><path d="m5.3 8.2 1.9 1.9 3.6-3.8"/>`,
		"close-circle":       `<circle cx="8" cy="8" r="6"/><path d="m5.9 5.9 4.2 4.2m0-4.2-4.2 4.2"/>`,
		"exclamation-circle": `<circle cx="8" cy="8" r="6"/><path d="M8 4.8v4"/><path d="M8 11v.2"/>`,
		"question-circle":    `<circle cx="8" cy="8" r="6"/><path d="M6.2 6.3A1.9 1.9 0 0 1 9.8 7c0 1.3-1.8 1.5-1.8 2.7"/><path d="M8 11.4v.2"/>`,
		"loading":            `<path d="M8 2a6 6 0 1 1-6 6"/>`,
		"reload":             `<path d="M13 8a5 5 0 1 1-1.5-3.6"/><path d="M11.8 1.8v2.8H9"/>`,
		"upload":             `<path d="M8 10.5v-8M4.8 5.5 8 2.3l3.2 3.2M2.5 10.5v3h11v-3"/>`,
		"download":           `<path d="M8 2.5v8M4.8 7.3 8 10.5l3.2-3.2M2.5 10.5v3h11v-3"/>`,
		"eye":                `<path d="M1.5 8S4 3.5 8 3.5 14.5 8 14.5 8 12 12.5 8 12.5 1.5 8 1.5 8Z"/><circle cx="8" cy="8" r="2.2"/>`,
		"eye-invisible":      `<path d="M2.5 6.2C1.9 7.1 1.5 8 1.5 8S4 12.5 8 12.5c.9 0 1.8-.2 2.5-.6M6.2 3.8C6.8 3.6 7.4 3.5 8 3.5c4 0 6.5 4.5 6.5 4.5s-.6 1.1-1.7 2.2"/><path d="m2 2 12 12"/>`,
		"calendar":           `<rect x="2.5" y="3.5" width="11" height="10" rx="1"/><path d="M2.5 6.5h11M5.5 2v3M10.5 2v3"/>`,
		"clock-circle":       `<circle cx="8" cy="8" r="6"/><path d="M8 4.5V8l2.4 1.6"/>`,
		"star":               `<path d="m8 2 1.8 3.8 4.1.5-3 2.9.8 4.1L8 11.3l-3.7 2 .8-4.1-3-2.9 4.1-.5z"/>`,
		"heart":              `<path d="M8 13.5S2 10 2 6a3 3 0 0 1 6-1 3 3 0 0 1 6 1c0 4-6 7.5-6 7.5Z"/>`,
		"bell":               `<path d="M4 11V7a4 4 0 0 1 8 0v4l1.2 1.5H2.8zM6.5 14h3"/>`,
		"mail":               `<rect x="2" y="3.5" width="12" height="9" rx="1"/><path d="m2.5 4.5 5.5 4 5.5-4"/>`,
		"lock":               `<rect x="3" y="7" width="10" height="7" rx="1"/><path d="M5 7V5a3 3 0 0 1 6 0v2M8 9.5v2"/>`,
		"unlock":             `<rect x="3" y="7" width="10" height="7" rx="1"/><path d="M5 7V5a3 3 0 0 1 5.8-1M8 9.5v2"/>`,
		"file":               `<path d="M4 1.8h5l3.5 3.5V14.2H4z"/><path d="M9 1.8v3.5h3.5"/>`,
		"folder":             `<path d="M1.8 3.5h4.4l1.5 1.5h6.5v8H1.8z"/>`,
		"appstore":           `<rect x="2.5" y="2.5" width="4.5" height="4.5"/><rect x="9" y="2.5" width="4.5" height="4.5"/><rect x="2.5" y="9" width="4.5" height="4.5"/><rect x="9" y="9" width="4.5" height="4.5"/>`,
		"bars":               `<path d="M2.5 4h11M2.5 8h11M2.5 12h11"/>`,
		"more":               `<path d="M8 3.5v.1M8 8v.1M8 12.5v.1" stroke-width="2.2"/>`,
		"ellipsis":           `<path d="M3.5 8h.1M8 8h.1M12.5 8h.1" stroke-width="2.2"/>`,
		"link":               `<path d="M6.8 9.2a2.5 2.5 0 0 0 3.5 0l2.4-2.4a2.5 2.5 0 0 0-3.5-3.5l-.9.9M9.2 6.8a2.5 2.5 0 0 0-3.5 0L3.3 9.2a2.5 2.5 0 0 0 3.5 3.5l.9-.9"/>`,
		"global":             `<circle cx="8" cy="8" r="6"/><path d="M2 8h12M8 2c-2.2 2.5-2.2 9.5 0 12M8 2c2.2 2.5 2.2 9.5 0 12"/>`,
		"cloud":              `<path d="M4.5 12.5a3 3 0 0 1-.4-6 4 4 0 0 1 7.7-.6 3.3 3.3 0 0 1-.3 6.6z"/>`,
		"database":           `<ellipse cx="8" cy="3.8" rx="5" ry="1.8"/><path d="M3 3.8v8.4c0 1 2.2 1.8 5 1.8s5-.8 5-1.8V3.8M3 8c0 1 2.2 1.8 5 1.8S13 9 13 8"/>`,
		"desktop":            `<rect x="1.8" y="2.5" width="12.4" height="8.5" rx="1"/><path d="M6 14h4M8 11v3"/>`,
		"mobile":             `<rect x="4.5" y="1.8" width="7" height="12.4" rx="1"/><path d="M7.4 12h1.2"/>`,
		"safety":             `<path d="M8 1.8 13 3.6v4.2c0 3-2.2 5.2-5 6.4-2.8-1.2-5-3.4-5-6.4V3.6z"/><path d="m5.6 8 1.8 1.8 3-3.2"/>`,
		"filter":             `<path d="M2 2.8h12L9.4 8.3v5L6.6 12V8.3z"/>`,
		"holder":             `<path d="M6 3.5v.1M10 3.5v.1M6 8v.1M10 8v.1M6 12.5v.1M10 12.5v.1" stroke-width="2"/>`,
		"poweroff":           `<path d="M8 2v5.5M5 3.8a5.2 5.2 0 1 0 6 0"/>`,
		"swap":               `<path d="M2.5 5.5h10l-2.5-2.5M13.5 10.5h-10l2.5 2.5"/>`,
		"enter":              `<path d="M13 3v5.5H3.5M6 6 3.5 8.5 6 11"/>`,
		"scan":               `<path d="M2 5V2h3M11 2h3v3M14 11v3h-3M5 14H2v-3M2 8h12"/>`,
		"qrcode":             `<rect x="2" y="2" width="4.5" height="4.5"/><rect x="9.5" y="2" width="4.5" height="4.5"/><rect x="2" y="9.5" width="4.5" height="4.5"/><path d="M9.5 9.5h1.5V11M14 9.5v2M9.5 14h4.5v-1.5"/>`,
	}
	for name, body := range icons {
		reg.src[name] = &iconSrc{kit: true, primary: body}
	}
	twin := map[string][2]string{
		"check-circle-twotone": {`<circle cx="8" cy="8" r="6"/><path d="m5.3 8.2 1.9 1.9 3.6-3.8"/>`, `<circle cx="8" cy="8" r="6" fill="currentColor" stroke="none"/>`},
		"info-circle-twotone":  {`<circle cx="8" cy="8" r="6"/><path d="M8 7.2v4"/><path d="M8 4.8v.2"/>`, `<circle cx="8" cy="8" r="6" fill="currentColor" stroke="none"/>`},
		"heart-twotone":        {`<path d="M8 13.5S2 10 2 6a3 3 0 0 1 6-1 3 3 0 0 1 6 1c0 4-6 7.5-6 7.5Z"/>`, `<path d="M8 13.5S2 10 2 6a3 3 0 0 1 6-1 3 3 0 0 1 6 1c0 4-6 7.5-6 7.5Z" fill="currentColor" stroke="none"/>`},
		"star-twotone":         {`<path d="m8 2 1.8 3.8 4.1.5-3 2.9.8 4.1L8 11.3l-3.7 2 .8-4.1-3-2.9 4.1-.5z"/>`, `<path d="m8 2 1.8 3.8 4.1.5-3 2.9.8 4.1L8 11.3l-3.7 2 .8-4.1-3-2.9 4.1-.5z" fill="currentColor" stroke="none"/>`},
		"safety-twotone":       {`<path d="M8 1.8 13 3.6v4.2c0 3-2.2 5.2-5 6.4-2.8-1.2-5-3.4-5-6.4V3.6z"/><path d="m5.6 8 1.8 1.8 3-3.2"/>`, `<path d="M8 1.8 13 3.6v4.2c0 3-2.2 5.2-5 6.4-2.8-1.2-5-3.4-5-6.4V3.6z" fill="currentColor" stroke="none"/>`},
	}
	for name, parts := range twin {
		reg.src[name] = &iconSrc{kit: true, primary: parts[0], fill: parts[1]}
	}
	registerAntIcons(reg)
	return reg
}

func antSVG(viewBox, inner string) *ui.SVG {
	if inner == "" {
		return nil
	}
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="` + viewBox + `" fill="currentColor">` + inner + `</svg>`))
}

func registerAntIcons(reg *iconStore) {
	put := func(name string, src *iconSrc) {
		if name == "" || src == nil || src.primary == "" {
			return
		}
		if _, ok := reg.src[name]; ok {
			return
		}
		if _, ok := reg.byName[name]; ok {
			return
		}
		reg.src[name] = src
	}
	for i := range antGlyphs {
		g := &antGlyphs[i]
		src := &iconSrc{viewBox: g.viewBox, primary: g.primary, fill: g.fill}
		put(g.name, src)
		base, theme := splitAntIconName(g.name)
		kebab := pascalToKebab(base)
		switch theme {
		case "outlined":
			put(kebab, src)
			put(base, src)
		case "filled":
			put(kebab+"-filled", src)
		case "twotone":
			put(kebab+"-twotone", src)
		}
	}
}
