package kit

import "github.com/egoist/mygo/ui"

// Size picks a control height, as Ant's size prop: Middle 32, Large 40, Small 24.
type Size uint8

// Control sizes. SizeInherit (zero) is unset: a control uses the
// ConfigProvider's size, and a nested provider keeps the outer Size.
// Middle is explicit, so a Middle provider under a Large one can restore
// 32px controls.
const (
	SizeInherit Size = iota
	// Middle is 32 high, the default at the root.
	Middle
	// Large is 40 high, as the settings forms.
	Large
	// Small is 24 high.
	Small
)

// Tokens are the design values of one chrome. Values follow Ant Design v5
// default tokens (ConfigProvider seed colors and control sizes). Hover and
// pressed colors come from AntPalette, as Ant derives them.
type Tokens struct {
	Name string
	Lang string
	Font string
	// Dark marks a dark variant, which themes ui's own widgets dark too.
	Dark bool

	// Primary drives buttons, focus, checked controls (Ant colorPrimary).
	Primary, PrimaryHover, PrimaryActive, PrimaryBg, PrimaryBorder ui.Color
	// Brand is the chrome blue of headers, section bars, steps and the
	// active settings tab; BrandSoft fills the database type chip.
	Brand, BrandSoft ui.Color

	// Error is Ant colorError (danger buttons, invalid fields); Danger is
	// the default status red (stopped services, warnings in copy).
	Error, ErrorHover, ErrorActive, ErrorBg, ErrorBorder ui.Color
	Danger, Success, Warning                             ui.Color

	Text, TextSecondary, TextMuted, TextDisabled, TextOnPrimary ui.Color

	// PageBg fills the page; Surface the controls and sheets; SurfaceSubtle
	// wells and table heads; DisabledBg disabled controls.
	PageBg, Surface, SurfaceSubtle, DisabledBg ui.Color
	// Border outlines inputs and selects, BorderButton default buttons,
	// Split the section title rule, SheetBorder sheets and cards.
	Border, BorderButton, Split, SheetBorder ui.Color

	RadiusButton, RadiusControl, RadiusSheet, RadiusCard, RadiusBanner, RadiusPopup float32

	ControlHeightSM, ControlHeight, ControlHeightLG float32
	PaddingInput, PaddingButton                     float32

	FontSizeSM, FontSize, FontSizeLG, FontSizeTitle, FontSizeHeading float32
	// LabelSize is the form label at Middle, LabelSizeLG at Large.
	LabelSize, LabelSizeLG float32

	SpaceXS, SpaceSM, SpaceMD, SpaceLG, SpaceXL float32

	// SwitchOff is the track of an unchecked switch, SwitchKnob its handle.
	SwitchOff, SwitchKnob ui.Color

	// TextTertiary is Ant colorTextTertiary (description labels, upload
	// hints, close marks); TextHeading colorTextHeading; TextPlaceholder
	// colorTextPlaceholder (#bfbfbf on light).
	TextTertiary, TextHeading, TextPlaceholder ui.Color
	// ControlOutline is Ant's input focus glow: primary at ~0.2 alpha,
	// a 2px spread outside the 1px primary border.
	ControlOutline ui.Color
	// Fill to FillQuaternary are Ant's fills, strongest first: pressed and
	// hovered text buttons, rows and tracks.
	Fill, FillSecondary, FillTertiary, FillQuaternary ui.Color
	// FillAlter fills table heads and description labels; BorderSecondary
	// rules tables and tabs; DividerColor is Divider's hairline.
	FillAlter, BorderSecondary, DividerColor ui.Color
	// Mask dims the window under dialogs and drawers; Elevated is the
	// background of popups, messages, dialogs and drawers.
	Mask, Elevated ui.Color
	// TooltipBg and TooltipText are Ant's dark tooltip.
	TooltipBg, TooltipText ui.Color

	// The success and warning families of alerts and tags, and the text
	// of error tags (Ant's green, gold and red 1 / 3 / 6 / 7).
	SuccessBg, SuccessBorder, SuccessIcon, SuccessText ui.Color
	WarningBg, WarningBorder, WarningIcon, WarningText ui.Color
	ErrorText                                          ui.Color
	// Banner holds the default flat notice colors, by AlertType.
	Banner [4]Pair

	// EmptyBg and EmptyBorder frame Empty; EmptyIcon tints its mark.
	EmptyBg, EmptyBorder, EmptyIcon ui.Color
	// NavBorder rules the settings sidebar and chips; NavHover and
	// NavPressed fill a hovered and a pressed entry.
	NavBorder, NavHover, NavPressed ui.Color
	// PickerTop, PickerBottom and PickerPressed are the default grey
	// gradient file button (PathSetting.scss .select-file).
	PickerTop, PickerBottom, PickerPressed ui.Color

	// Glass is the home page's frosted card and health pill.
	Glass GlassTokens

	// AvatarBg fills an avatar without an image or color (Ant
	// colorTextPlaceholder); BadgeShadow rings a count badge (Ant
	// colorBorderBg).
	AvatarBg, BadgeShadow ui.Color

	// Layout holds Ant's Layout and dark Menu colors.
	Layout LayoutTokens

	// Feedback holds Notification, Progress and Watermark values.
	Feedback FeedbackTokens

	// Size, Disabled and Locale are what a ConfigProvider sets for the
	// controls under it; Middle, enabled and LocaleZhCN outside one.
	Size     Size
	Disabled bool
	Locale   *Locale
	// Form is the layout FormView gives the FormItems under it.
	Form FormConfig
	// MotionOff is Ant token.motion=false: kit enter/leave jumps. OS
	// ReduceMotion is already honoured by ui.Animate.
	MotionOff bool

	theme *ui.Theme
	base  *Tokens
}

// FeedbackTokens are Ant v5's Notification, Progress and Watermark
// component tokens.
type FeedbackTokens struct {
	// NoticeWidth is a notification's width, NoticeMargin its distance
	// from the window's edges and the gap between two (Ant width 384,
	// marginEdge 24, notificationMarginBottom 16).
	NoticeWidth, NoticeMargin, NoticeGap float32
	// Progress fills a bar's done part by default, ProgressSuccess a
	// finished one (colorSuccess); the trail is FillSecondary.
	Progress, ProgressSuccess ui.Color
	// WatermarkText is the watermark's default color (rgba 0 0 0 .15),
	// WatermarkSize its font size and WatermarkGap the gap between tiles.
	WatermarkText               ui.Color
	WatermarkSize, WatermarkGap float32
}

// LayoutTokens are Ant v5's Layout and Menu component tokens.
type LayoutTokens struct {
	// Body fills a Layout and its Footer (colorBgLayout); Header and
	// Sider are the dark headerBg and siderBg; Trigger the collapse bar
	// under a dark Sider.
	Body, Header, Sider, Trigger ui.Color
	// MenuDarkSub fills an open inline submenu of a dark menu
	// (darkSubMenuItemBg); MenuDarkText is its resting text
	// (darkItemColor); MenuSub the light one's (subMenuItemBg).
	MenuDarkSub, MenuDarkText, MenuSub ui.Color
	HeaderHeight, TriggerHeight        float32
}

// Pair is a background and the text on it.
type Pair struct{ Bg, Fg ui.Color }

// GlassTokens holds the home chrome's translucent surfaces (home layout.module.scss).
type GlassTokens struct {
	Card, CardEdge, CardShadow ui.Color
	Pill, PillEdge, PillShadow ui.Color
	CardRadius                 float32
	// Title and Subtle are card text, PillText the health pill's title.
	Title, Subtle, PillText ui.Color
}

var (
	antPrimary = ui.Hex("#006adb") // main.tsx ConfigProvider colorPrimary
	antError   = ui.Hex("#ff4d4f") // Ant v5 colorError
	brandBlue  = ui.Hex("#1E59D7") // app header, settings layout: split-select, steps
	homeBlue   = ui.Hex("#2f80ed") // home layout: nav-item.is-current, switches
)

// Work is the settings / install / remote chrome: blue 40px header over a
// #F2F5F7 page, Ant controls with the default overrides.
var Work = newTokens("work", antPrimary, func(t *Tokens) {
	t.PageBg = ui.Hex("#F2F5F7")
	t.Text = ui.Hex("#262626")
	t.TextSecondary = ui.Hex("#666666")
	t.TextMuted = ui.Hex("#8A94A6")
	// Field chrome stays Ant v5 (#d9d9d9 / 6px / 40px large) even though
	// settings layout.scss used a lighter #e1e4e6 and 4px radius.
	t.RadiusButton = 2 // settings layout: ant-btn radius
})

// Home is the service dashboard chrome: #eef6ff behind glass cards.
var Home = newTokens("home", homeBlue, func(t *Tokens) {
	t.PageBg = ui.Hex("#EEF6FF")
	t.Text = ui.Hex("#1F2937")
	t.TextSecondary = ui.Hex("#4B5563")
	t.TextMuted = ui.Hex("#8A94A6")
	t.SwitchOff = ui.Hex("#D1D5DB") // .service-card .ant-switch
	t.RadiusButton = 6
	t.RadiusControl = 6
})

func newTokens(name string, primary ui.Color, tweak func(*Tokens)) *Tokens {
	p := AntPalette(primary)
	e := AntPalette(antError)
	t := &Tokens{
		Name:   name,
		Lang:   "zh-CN",
		Font:   FontList(),
		Size:   Middle,
		Locale: LocaleZhCN,

		Primary: p[5], PrimaryHover: p[4], PrimaryActive: p[6], PrimaryBg: p[0], PrimaryBorder: p[2],
		Brand: brandBlue, BrandSoft: ui.Hex("#C8D6F1"),

		Error: e[5], ErrorHover: e[4], ErrorActive: e[6], ErrorBg: e[0], ErrorBorder: e[2],
		Danger:  ui.Hex("#E03A3A"),
		Success: ui.Hex("#16A34A"),
		Warning: ui.Hex("#D97706"),

		Text:          ui.Hex("#262626"),
		TextSecondary: ui.Hex("#666666"),
		TextMuted:     ui.Hex("#8A94A6"),
		TextDisabled:  ui.RGBA(0, 0, 0, 0.25),
		TextOnPrimary: ui.Hex("#FFFFFF"),

		Surface:       ui.Hex("#FFFFFF"),
		SurfaceSubtle: ui.Hex("#FAFAFA"),
		DisabledBg:    ui.Hex("#F5F5F5"),
		Border:        ui.Hex("#D9D9D9"),
		BorderButton:  ui.Hex("#D9D9D9"),
		Split:         ui.Hex("#D9D9D9"),
		SheetBorder:   ui.Hex("#E5EAF1"), // RemoteConnect .panel

		RadiusButton: 6, RadiusControl: 6, RadiusSheet: 2, RadiusCard: 10, RadiusBanner: 8, RadiusPopup: 8,

		ControlHeightSM: 24, ControlHeight: 32, ControlHeightLG: 40,
		PaddingInput: 11, PaddingButton: 15,

		FontSizeSM: 12, FontSize: 14, FontSizeLG: 16, FontSizeTitle: 18, FontSizeHeading: 20,
		LabelSize: 14, LabelSizeLG: 16,

		SpaceXS: 4, SpaceSM: 8, SpaceMD: 12, SpaceLG: 16, SpaceXL: 20,

		SwitchOff: ui.RGBA(0, 0, 0, 0.25), SwitchKnob: ui.Hex("#FFFFFF"),

		TextTertiary: ui.RGBA(0, 0, 0, 0.45), TextHeading: ui.RGBA(0, 0, 0, 0.88),
		TextPlaceholder: ui.Hex("#bfbfbf"), // Ant colorTextPlaceholder on white
		Fill:            ui.RGBA(0, 0, 0, 0.15), FillSecondary: ui.RGBA(0, 0, 0, 0.06),
		FillTertiary: ui.RGBA(0, 0, 0, 0.04), FillQuaternary: ui.RGBA(0, 0, 0, 0.02),
		FillAlter: ui.Hex("#FAFAFA"), BorderSecondary: ui.Hex("#F0F0F0"), DividerColor: ui.RGBA(5, 5, 5, 0.06),
		Mask: ui.RGBA(0, 0, 0, 0.45), Elevated: ui.Hex("#FFFFFF"),
		TooltipBg: ui.RGBA(0, 0, 0, 0.85), TooltipText: ui.Hex("#FFFFFF"),

		SuccessBg: ui.Hex("#f6ffed"), SuccessBorder: ui.Hex("#b7eb8f"), SuccessIcon: ui.Hex("#52c41a"), SuccessText: ui.Hex("#389e0d"),
		WarningBg: ui.Hex("#fffbe6"), WarningBorder: ui.Hex("#ffe58f"), WarningIcon: ui.Hex("#faad14"), WarningText: ui.Hex("#d48806"),
		ErrorText: ui.Hex("#cf1322"),
		// RemoteConnect .banner and its states.
		Banner: [4]Pair{
			Info:    {ui.Hex("#e8f1fc"), ui.Hex("#1d4ed8")},
			Success: {ui.Hex("#f0fdf4"), ui.Hex("#15803d")},
			Warning: {ui.Hex("#fffbeb"), ui.Hex("#b45309")},
			Error:   {ui.Hex("#fef2f2"), ui.Hex("#b91c1c")},
		},

		EmptyBg: ui.Hex("#F8FAFC"), EmptyBorder: ui.Hex("#EEF2F6"), EmptyIcon: ui.RGBA(0, 0, 0, 0.18),
		// SettingTab.module.scss rule, hover and pressed tab fills.
		NavBorder: ui.Hex("#E8E8E8"), NavHover: ui.Hex("#E6ECF7"), NavPressed: ui.Hex("#DCE5F5"),
		PickerTop: ui.Hex("#FFFFFF"), PickerBottom: ui.Hex("#E3E3E3"), PickerPressed: ui.Hex("#D6D6D6"),

		Glass: GlassTokens{
			Card: ui.RGBA(255, 255, 255, 0.52), CardEdge: ui.RGBA(255, 255, 255, 0.72),
			CardShadow: ui.RGBA(80, 130, 200, 0.08), CardRadius: 20,
			Pill: ui.RGBA(255, 255, 255, 0.55), PillEdge: ui.RGBA(255, 255, 255, 0.7),
			PillShadow: ui.RGBA(47, 100, 180, 0.08),
			Title:      ui.Hex("#303640"), Subtle: ui.Hex("#9AA3B2"), PillText: ui.Hex("#374151"),
		},
		AvatarBg: ui.RGBA(0, 0, 0, 0.25), BadgeShadow: ui.Hex("#FFFFFF"),
		Layout: LayoutTokens{
			Body: ui.Hex("#f5f5f5"), Header: ui.Hex("#001529"), Sider: ui.Hex("#001529"), Trigger: ui.Hex("#002140"),
			MenuDarkSub: ui.Hex("#000c17"), MenuDarkText: ui.RGBA(255, 255, 255, 0.65), MenuSub: ui.RGBA(0, 0, 0, 0.02),
			HeaderHeight: 64, TriggerHeight: 48,
		},
		Feedback: FeedbackTokens{
			NoticeWidth: 384, NoticeMargin: 24, NoticeGap: 16,
			Progress: brandBlue, ProgressSuccess: ui.Hex("#52c41a"),
			WatermarkText: ui.RGBA(0, 0, 0, 0.15), WatermarkSize: 16, WatermarkGap: 100,
		},
	}
	tweak(t)
	t.ControlOutline = t.Primary.Alpha(0.2)
	return t
}

var darkAlgos = map[*Tokens]*Tokens{}

// WorkDark is Work under Ant's dark algorithm: #000 page, #141414
// containers, #1f1f1f popups, white-alpha text and fills, dark glass,
// and the primary and error scales mixed toward the dark background as
// Ant generates them. DarkAlgorithm(t) is the same mapping for any
// tokens (ConfigProps.Dark). The product window sets Dark from
// DesktopDark each frame; gallery and Page stay opt-in.
var WorkDark = DarkAlgorithm(Work)

// DarkAlgorithm is Ant theme.darkAlgorithm: the dark map of light's
// primary and error scales, with Ant's dark page / text / fill tokens.
// The same *Tokens is returned for the same light set; a dark set is
// returned as it is.
func DarkAlgorithm(light *Tokens) *Tokens {
	if light == nil {
		light = Work
	}
	if light.Dark {
		return light
	}
	if light.base != nil {
		light = light.base
		if light.Dark {
			return light
		}
	}
	registryMu.Lock()
	if darkAlgos == nil {
		darkAlgos = map[*Tokens]*Tokens{}
	}
	if t, ok := darkAlgos[light]; ok {
		registryMu.Unlock()
		return t
	}
	registryMu.Unlock()
	name := light.Name + "-dark"
	if light == Work {
		name = "work-dark"
	}
	t := newDarkTokens(name, light.Primary, light)
	registryMu.Lock()
	defer registryMu.Unlock()
	if have, ok := darkAlgos[light]; ok {
		return have
	}
	darkAlgos[light] = t
	return t
}

func newDarkTokens(name string, primary ui.Color, light *Tokens) *Tokens {
	t := *light
	t.Name, t.Dark, t.theme = name, true, nil
	bg := ui.Hex("#141414")
	p, e := AntDarkPalette(primary, bg), AntDarkPalette(antError, bg)
	t.Primary, t.PrimaryHover, t.PrimaryActive, t.PrimaryBg, t.PrimaryBorder = p[5], p[6], p[4], p[0], p[2]
	t.Error, t.ErrorHover, t.ErrorActive, t.ErrorBg, t.ErrorBorder = e[5], e[6], e[4], e[0], e[2]
	t.Brand, t.BrandSoft = p[5], p[1]
	white := func(a float32) ui.Color { return ui.RGBA(255, 255, 255, a) }
	t.Text, t.TextSecondary, t.TextMuted, t.TextDisabled, t.TextOnPrimary = white(0.85), white(0.65), white(0.45), white(0.25), ui.Hex("#FFFFFF")
	t.TextTertiary, t.TextHeading, t.TextPlaceholder = white(0.45), white(0.85), white(0.25)
	t.ControlOutline = t.Primary.Alpha(0.2)
	t.PageBg, t.Surface, t.SurfaceSubtle, t.DisabledBg = ui.Hex("#000000"), bg, white(0.04), white(0.08)
	t.Border, t.BorderButton, t.Split, t.SheetBorder = ui.Hex("#424242"), ui.Hex("#424242"), ui.Hex("#303030"), ui.Hex("#303030")
	t.Fill, t.FillSecondary, t.FillTertiary, t.FillQuaternary = white(0.18), white(0.12), white(0.08), white(0.04)
	t.FillAlter, t.BorderSecondary, t.DividerColor = ui.Hex("#1d1d1d"), ui.Hex("#303030"), ui.RGBA(253, 253, 253, 0.12)
	t.Mask, t.Elevated = ui.RGBA(0, 0, 0, 0.45), ui.Hex("#1f1f1f")
	t.TooltipBg, t.TooltipText = ui.Hex("#424242"), white(0.85)
	t.SuccessBg, t.SuccessBorder, t.SuccessIcon, t.SuccessText = ui.Hex("#162312"), ui.Hex("#274916"), ui.Hex("#49aa19"), ui.Hex("#49aa19")
	t.WarningBg, t.WarningBorder, t.WarningIcon, t.WarningText = ui.Hex("#2b2111"), ui.Hex("#594214"), ui.Hex("#d89614"), ui.Hex("#d89614")
	t.ErrorText = e[6]
	t.Banner = [4]Pair{
		Info:    {p[0], p[6]},
		Success: {t.SuccessBg, t.SuccessText},
		Warning: {t.WarningBg, t.WarningText},
		Error:   {t.ErrorBg, t.ErrorText},
	}
	t.EmptyBg, t.EmptyBorder, t.EmptyIcon = white(0.02), ui.Hex("#303030"), white(0.25)
	t.NavBorder, t.NavHover, t.NavPressed = ui.Hex("#303030"), white(0.08), white(0.12)
	t.PickerTop, t.PickerBottom, t.PickerPressed = ui.Hex("#2a2a2a"), ui.Hex("#1f1f1f"), ui.Hex("#141414")
	t.SwitchOff, t.SwitchKnob = white(0.25), ui.Hex("#FFFFFF")
	t.AvatarBg, t.BadgeShadow = white(0.25), bg
	t.Layout.Body, t.Layout.MenuSub = ui.Hex("#000000"), white(0.04)
	t.Feedback.Progress, t.Feedback.WatermarkText = p[5], white(0.15)
	t.Glass = GlassTokens{
		Card: white(0.08), CardEdge: white(0.16),
		CardShadow: ui.RGBA(0, 0, 0, 0.45), CardRadius: t.Glass.CardRadius,
		Pill: white(0.08), PillEdge: white(0.16),
		PillShadow: ui.RGBA(0, 0, 0, 0.45),
		Title:      t.Text, Subtle: t.TextMuted, PillText: t.Text,
	}
	return &t
}

// SizeOf is s when it is set, or the ConfigProvider's size, or Middle.
func (t *Tokens) SizeOf(s Size) Size {
	if s != SizeInherit {
		return s
	}
	if t.Size != SizeInherit {
		return t.Size
	}
	return Middle
}

// Height returns the control height of a size.
func (t *Tokens) Height(s Size) float32 {
	switch t.SizeOf(s) {
	case Large:
		return t.ControlHeightLG
	case Small:
		return t.ControlHeightSM
	}
	return t.ControlHeight
}

// Label returns the form label font size of a size.
func (t *Tokens) Label(s Size) float32 {
	if t.SizeOf(s) == Large {
		return t.LabelSizeLG
	}
	return t.LabelSize
}

// ControlFont is the text size of a control: FontSizeLG (16) at Large,
// FontSize (14) at Middle and Small. Form labels use Label instead.
func (t *Tokens) ControlFont(s Size) float32 {
	if t.SizeOf(s) == Large {
		return t.FontSizeLG
	}
	return t.FontSize
}
