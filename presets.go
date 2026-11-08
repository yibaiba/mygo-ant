package kit

import "github.com/egoist/mygo/ui"

// antPresets are Ant v5's preset colors (presetPrimaryColors), the sixth
// shade of each palette; Tag, Badge, Ribbon and Timeline take their names.
var antPresets = map[string]ui.Color{
	"blue":     ui.Hex("#1677ff"),
	"purple":   ui.Hex("#722ed1"),
	"cyan":     ui.Hex("#13c2c2"),
	"green":    ui.Hex("#52c41a"),
	"magenta":  ui.Hex("#eb2f96"),
	"pink":     ui.Hex("#eb2f96"),
	"red":      ui.Hex("#f5222d"),
	"orange":   ui.Hex("#fa8c16"),
	"yellow":   ui.Hex("#fadb14"),
	"volcano":  ui.Hex("#fa541c"),
	"geekblue": ui.Hex("#2f54eb"),
	"lime":     ui.Hex("#a0d911"),
	"gold":     ui.Hex("#faad14"),
}

// Preset returns the palette of Ant's preset color name (blue, purple,
// cyan, green, magenta, pink, red, orange, yellow, volcano, geekblue,
// lime or gold), mixed toward the container under a dark chrome as Ant's
// dark algorithm does; ok is false for another name.
func (t *Tokens) Preset(name string) (Palette, bool) {
	seed, ok := antPresets[name]
	if !ok {
		return Palette{}, false
	}
	if t.Dark {
		return AntDarkPalette(seed, ui.Hex("#141414")), true
	}
	return AntPalette(seed), true
}

// toneColor resolves a component's color: a preset name gives its sixth
// shade, a status name ("success", "processing", "error", "warning",
// "default") the status color, and anything else fallback.
func toneColor(t *Tokens, name string, fallback ui.Color) ui.Color {
	switch name {
	case "success":
		return t.SuccessIcon
	case "processing":
		return t.Primary
	case "error":
		return t.Error
	case "warning":
		return t.WarningIcon
	case "default":
		return t.TextDisabled
	}
	if p, ok := t.Preset(name); ok {
		return p[5]
	}
	return fallback
}
