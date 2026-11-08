package kit

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// TestAffixPinsInTheFrameItCrosses checks every frame built: the affix is
// pinned exactly when the scroll offset of that frame has carried its
// holder (50 down the content) past the top, never a frame late.
func TestAffixPinsInTheFrameItCrosses(t *testing.T) {
	var s ui.ScrollState
	type frame struct {
		pinned bool
		y      float32
	}
	var frames []frame
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			sc := ui.Scroll(c).Fill().TrackScroll(&s)
			sc.Children(func() {
				ui.Column(c).FillWidth().Children(func() {
					ui.Box(c).Height(50)
					pinned := Affix(c, func() { ui.Text(c, "固定条").Height(30) }, AffixProps{Target: sc, Scroll: &s})
					frames = append(frames, frame{pinned, s.Y})
					ui.Box(c).Height(2000)
				})
			})
		})
	}, 400, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	check := func(what string) {
		t.Helper()
		for i, f := range frames {
			if want := f.y > 50; f.pinned != want {
				t.Fatalf("%s: frame %d at scroll %v pinned=%v, want %v", what, i, f.y, f.pinned, want)
			}
		}
		frames = frames[:0]
	}
	check("start")
	for _, dy := range []float32{30, 15, 10, 10, 40, -40, -20, -10, 100, -200} {
		tester.Scroll(200, 150, 0, dy)
		tester.Frame()
		check("scroll by " + itoa(int(dy)))
		r, _ := tester.Find("固定条")
		want := 50 - s.Y
		if s.Y > 50 {
			want = 0
		}
		if r.Y < want-0.5 || r.Y > want+0.5 {
			t.Fatalf("at scroll %v the bar is at %v, want %v", s.Y, r.Y, want)
		}
	}
	tester.Wheel(200, 150, 0, 120)
	for range 12 {
		tester.Advance(16 * time.Millisecond)
		check("wheel glide")
	}
	s.Y = 0
	tester.Frame()
	check("set to 0")
	s.Y = 400
	tester.Frame()
	check("set to 400")
}
