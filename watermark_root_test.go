package kit

import (
	"image"
	"testing"

	"github.com/egoist/mygo/ui"
)

// coveredDiff counts the pixels of r that differ between two frames.
func coveredDiff(a, b image.Image, r image.Rectangle) int {
	n := 0
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				n++
			}
		}
	}
	return n
}

// rootMarked builds an App whose page covers its right half with an
// opaque overlay of its own and has a button on the left.
func rootMarked(s *AppState, mark **WatermarkProps, clicks *int, broken *[]string) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			App(c, s, func() {
				ui.Column(c).Padding(20).Children(func() {
					if Button(c, "提交") {
						*clicks++
					}
				})
				ui.Overlay(c, func() {
					ui.Box(c).Absolute().Left(200).Top(0).Width(200).Height(200).Background(ui.Hex("#ffffff"))
				})
				Watermark(c, func() {}, WatermarkProps{})
			}, AppProps{Watermark: *mark, OnWatermarkBroken: func(r string) { *broken = append(*broken, r) }})
		})
	}, 400, 200)
}

func TestRootWatermarkPaintsOverPageOverlaysAndKeepsClicks(t *testing.T) {
	var s AppState
	var mark *WatermarkProps
	clicks := 0
	var broken []string
	tester := rootMarked(&s, &mark, &clicks, &broken)
	tester.Frame()
	plain := tester.Image()
	if ok, why := s.Watermark.Intact(); ok || why != "not set" {
		t.Fatalf("no watermark yet: Intact = %v, %q", ok, why)
	}
	mark = &WatermarkProps{Content: []string{"内部资料", "张三"}, Gap: [2]float32{30, 30}}
	tester.Frame()
	tester.Frame()
	marked := tester.Image()
	if ok, why := s.Watermark.Intact(); !ok {
		t.Fatalf("watermark should be intact: %s", why)
	}
	if n := coveredDiff(plain, marked, image.Rect(200, 0, 400, 200)); n < 200 {
		t.Fatalf("only %d pixels of the page's own opaque overlay changed: the watermark is under it", n)
	}
	if err := tester.Click("提交"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if clicks != 1 {
		t.Fatal("the page under the window watermark should take clicks")
	}
	if len(broken) != 0 {
		t.Fatalf("reported broken: %v", broken)
	}
}

func TestRootWatermarkCannotBeSwitchedOff(t *testing.T) {
	var s AppState
	mark := &WatermarkProps{Content: []string{"内部资料"}, Gap: [2]float32{30, 30}}
	clicks := 0
	var broken []string
	tester := rootMarked(&s, &mark, &clicks, &broken)
	tester.Frame()
	tester.Frame()
	before := tester.Image()
	for _, off := range []*WatermarkProps{nil, {}, {Content: []string{"  "}}, {Content: []string{"x"}, Color: ui.RGBA(0, 0, 0, 0.01)}} {
		mark = off
		tester.Frame()
		tester.Frame()
		if ok, why := s.Watermark.Intact(); !ok {
			t.Fatalf("%+v switched it off: %s", off, why)
		}
		if n := coveredDiff(before, tester.Image(), image.Rect(0, 0, 400, 200)); n != 0 {
			t.Fatalf("%+v changed %d pixels: the last watermark should stay", off, n)
		}
	}
	mark = &WatermarkProps{Content: []string{"李四"}, Gap: [2]float32{30, 30}}
	tester.Frame()
	tester.Frame()
	if coveredDiff(before, tester.Image(), image.Rect(0, 0, 400, 200)) == 0 {
		t.Fatal("a new visible watermark should replace the old one")
	}
}

func TestRootWatermarkReportsAMissedPaint(t *testing.T) {
	var s AppState
	mark := &WatermarkProps{Content: []string{"内部资料"}}
	clicks := 0
	var broken []string
	tester := rootMarked(&s, &mark, &clicks, &broken)
	tester.Frame()
	tester.Frame()
	s.Watermark.mu.Lock()
	s.Watermark.painted--
	s.Watermark.mu.Unlock()
	if ok, why := s.Watermark.Intact(); ok || why != "not painted this frame" {
		t.Fatalf("Intact = %v, %q after a lost paint", ok, why)
	}
	tester.Frame()
	if len(broken) != 0 {
		t.Fatalf("one build behind may be a pass of the same frame, reported %v", broken)
	}
	s.Watermark.mu.Lock()
	s.Watermark.painted -= framePasses
	s.Watermark.mu.Unlock()
	tester.Frame()
	if len(broken) != 1 || broken[0] != "not painted this frame" {
		t.Fatalf("OnWatermarkBroken got %v", broken)
	}
	if ok, _ := s.Watermark.Intact(); !ok {
		t.Fatal("the next frame should paint it again")
	}
	s.Watermark.mu.Lock()
	s.Watermark.area.W = 100
	s.Watermark.mu.Unlock()
	if _, why := s.Watermark.Intact(); why != "does not cover the window" {
		t.Fatalf("a part-window paint reads %q", why)
	}
}

func TestGalleryRootWatermarkShot(t *testing.T) {
	var s AppState
	open := true
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			App(c, &s, func() {
				ui.Column(c).Padding(24).Gap(12).Children(func() {
					Title(c, "会议资料", TitleProps{Level: 4})
					Paragraph(c, "窗口水印画在根层最上面：页面、它的弹窗和消息都在水印之下，水印不拦截点击。")
					Button(c, "提交", ButtonProps{Type: Primary})
				})
				Dialog(c, &open, "确认分发", func() {
					ui.Text(c, "资料将分发给 12 位参会人。")
				})
			}, AppProps{Watermark: &WatermarkProps{Content: []string{"Confidential", "张三 0315"}, FontSize: 14, Gap: [2]float32{60, 50}}})
		})
	}, 640, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	shot(t, tester, "mygo-kit-gallery-watermark-root.png")
}
