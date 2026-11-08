package kit

import (
	"fmt"
	"testing"

	"github.com/egoist/mygo/ui"
)

func hex(c ui.Color) string { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }

func TestAntPaletteMatchesAntBlue(t *testing.T) {
	want := []string{"#e6f4ff", "#bae0ff", "#91caff", "#69b1ff", "#4096ff", "#1677ff", "#0958d9", "#003eb3", "#002c8c", "#001d66"}
	got := AntPalette(ui.Hex("#1677ff"))
	for i, w := range want {
		if hex(got[i]) != w {
			t.Errorf("blue[%d] = %s, want %s", i, hex(got[i]), w)
		}
	}
}

func TestAntPaletteMatchesAntRed(t *testing.T) {
	want := []string{"#fff1f0", "#ffccc7", "#ffa39e", "#ff7875", "#ff4d4f", "#f5222d", "#cf1322", "#a8071a", "#820014", "#5c0011"}
	got := AntPalette(ui.Hex("#f5222d"))
	for i, w := range want {
		if hex(got[i]) != w {
			t.Errorf("red[%d] = %s, want %s", i, hex(got[i]), w)
		}
	}
}

func TestErrorTokensMatchAntV5(t *testing.T) {
	got := map[string]ui.Color{"hover": Work.ErrorHover, "active": Work.ErrorActive, "bg": Work.ErrorBg, "border": Work.ErrorBorder}
	want := map[string]string{"hover": "#ff7875", "active": "#d9363e", "bg": "#fff2f0", "border": "#ffccc7"}
	for k, w := range want {
		if hex(got[k]) != w {
			t.Errorf("error %s = %s, want %s", k, hex(got[k]), w)
		}
	}
}

func TestHomeChromeIsCoolBlueNotLime(t *testing.T) {
	if hex(Home.PageBg) != "#eef6ff" {
		t.Errorf("PageBg = %s, want #eef6ff", hex(Home.PageBg))
	}
	if hex(Home.Primary) != "#2f80ed" {
		t.Errorf("Primary = %s, want #2f80ed", hex(Home.Primary))
	}
	if hex(Home.Text) != "#1f2937" {
		t.Errorf("Text = %s, want #1f2937", hex(Home.Text))
	}
	if hex(Home.Glass.Title) != "#303640" {
		t.Errorf("Glass.Title = %s, want #303640", hex(Home.Glass.Title))
	}
	if hex(Home.Glass.Subtle) != "#9aa3b2" {
		t.Errorf("Glass.Subtle = %s, want #9aa3b2", hex(Home.Glass.Subtle))
	}
	for name, c := range map[string]ui.Color{"text": Home.Text, "title": Home.Glass.Title, "subtle": Home.Glass.Subtle, "page": Home.PageBg} {
		if int(c.G)-int(c.B) > 8 {
			t.Errorf("%s = %s looks lime (G>B); home chrome is cool blue", name, hex(c))
		}
	}
}

func TestAntDarkPaletteMatchesAntDarkBlue(t *testing.T) {
	want := []string{"#111a2c", "#112545", "#15325b", "#15417e", "#1554ad", "#1668dc", "#3c89e8", "#65a9f3", "#8dc5f8", "#b7dcfa"}
	got := AntDarkPalette(ui.Hex("#1677ff"), ui.Hex("#141414"))
	for i, w := range want {
		if hex(got[i]) != w {
			t.Errorf("dark blue[%d] = %s, want %s", i, hex(got[i]), w)
		}
	}
}
