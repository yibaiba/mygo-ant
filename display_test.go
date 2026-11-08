package kit

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// testPicture is a w×h picture: a gradient from a to b with a light
// band, to tell its rotations apart.
func testPicture(t testing.TB, w, h int, a, b color.RGBA) ui.ImageSource {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			k := float64(x) / float64(w)
			px := color.RGBA{
				uint8(float64(a.R)*(1-k) + float64(b.R)*k),
				uint8(float64(a.G)*(1-k) + float64(b.G)*k),
				uint8(float64(a.B)*(1-k) + float64(b.B)*k), 255}
			if y > h/5 && y < h/5+h/10 {
				px = color.RGBA{255, 255, 255, 255}
			}
			img.Set(x, y, px)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	src, err := ui.DecodeImage(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return src
}

func TestTagCloseAndCheckable(t *testing.T) {
	closed, checked := 0, false
	tester := mount(func(c *ui.Context) {
		if Tag(c, "会议室", TagProps{Closable: true, Preset: "blue"}) {
			closed++
		}
		CheckableTag(c, "已签到", &checked)
	}, 400, 200)
	if err := tester.Click("关闭标签 会议室"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if closed != 1 {
		t.Fatalf("close clicks %d", closed)
	}
	if err := tester.Click("已签到"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !checked {
		t.Fatal("checkable tag did not check")
	}
}

func TestAvatarGroupFoldsAndBadges(t *testing.T) {
	items := []AvatarItem{{Text: "张"}, {Text: "李"}, {Text: "王"}, {Text: "赵"}, {Text: "钱"}}
	tester := mount(func(c *ui.Context) {
		AvatarGroup(c, items, AvatarGroupProps{MaxCount: 3})
		BadgeCount(c, func() { Avatar(c, "U", AvatarProps{Square: true, Size: Large}) }, BadgeProps{Count: 120})
		BadgeCount(c, nil, BadgeProps{Count: 7, Preset: "green"})
		StatusDot(c, "processing", "同步中")
		Ribbon(c, "推荐", func() { Card(c, func() { Muted(c, "卡片内容") }) })
	}, 600, 400)
	for _, want := range []string{"+2", "99+", "7", "同步中", "推荐"} {
		if !tester.HasText(want) {
			t.Fatalf("missing %q in %v", want, tester.Texts())
		}
	}
	if tester.HasText("赵") {
		t.Fatal("folded avatars show before hover")
	}
	x, y := center(t, tester, "+2")
	tester.Move(x, y)
	tester.Frame()
	tester.Advance(200 * time.Millisecond)
	if !tester.HasText("赵") || !tester.HasText("钱") {
		t.Fatalf("hover lists folded avatars: %v", tester.Texts())
	}
}

func TestCountdownTicksAndFinishes(t *testing.T) {
	var deadline time.Time
	finished := 0
	tester := ui.NewTester(func(c *ui.Context) {
		if deadline.IsZero() {
			deadline = c.Now().Add(3 * time.Second)
		}
		Page(c, Work, func() {
			Countdown(c, "距离开会", deadline, CountdownProps{Format: "mm:ss", OnFinish: func() { finished++ }})
			Statistic(c, "参会人数", 1128.5, StatisticProps{Precision: 1, Suffix: "人"})
		})
	}, 400, 200)
	tester.Frame()
	if !tester.HasText("00:03") || !tester.HasText("1,128.5") {
		t.Fatalf("start %v", tester.Texts())
	}
	tester.Advance(1100 * time.Millisecond)
	if !tester.HasText("00:01") && !tester.HasText("00:02") {
		t.Fatalf("after a second %v", tester.Texts())
	}
	tester.Advance(3 * time.Second)
	if !tester.HasText("00:00") || finished != 1 {
		t.Fatalf("finish %v, %d", tester.Texts(), finished)
	}
	tester.Advance(2 * time.Second)
	if finished != 1 {
		t.Fatal("OnFinish ran twice")
	}
}

func TestCalendarPicksDayAndMonth(t *testing.T) {
	day := time.Date(2026, 11, 6, 0, 0, 0, 0, time.Local)
	mode := CalendarMonth
	tester := mount(func(c *ui.Context) {
		Calendar(c, &day, CalendarProps{Mode: &mode, CellRender: func(d time.Time) {
			if d.Day() == 12 && d.Month() == 11 {
				Tag(c, "评审会", TagProps{Preset: "blue"})
			}
		}})
	}, 1100, 900)
	if !tester.HasText("评审会") {
		t.Fatal("cell render")
	}
	if err := tester.Click("2026-11-18"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if day.Day() != 18 {
		t.Fatalf("picked %v", day)
	}
	if err := tester.Click("年"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if mode != CalendarYear {
		t.Fatal("mode switch")
	}
	if err := tester.Click("2026-02"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if day.Month() != 2 || day.Day() != 18 {
		t.Fatalf("month pick %v", day)
	}
}

func TestCarouselArrowsDotsAutoplay(t *testing.T) {
	cur := 0
	tester := mount(func(c *ui.Context) {
		ui.Column(c).Width(400).Children(func() {
			Carousel(c, 3, func(i int) {
				ui.Column(c).Fill().Center().Background(Work.Primary).Children(func() {
					ui.Text(c, "幻灯片 "+itoa(i+1)).TextColor(Work.TextOnPrimary)
				})
			}, CarouselProps{Arrows: true, Autoplay: true, Current: &cur})
		})
	}, 500, 260)
	if err := tester.Click("下一张"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if cur != 1 {
		t.Fatalf("next: %d", cur)
	}
	if err := tester.Click("第 1 张"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if cur != 0 {
		t.Fatalf("dot: %d", cur)
	}
	if err := tester.Click("上一张"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if cur != 2 {
		t.Fatalf("previous wraps: %d", cur)
	}
	tester.Move(490, 250)
	tester.Advance(3100 * time.Millisecond)
	if cur != 0 {
		t.Fatalf("autoplay wraps to the first: %d", cur)
	}
}

func TestImagePreviewZoomRotateGroup(t *testing.T) {
	pics := []ui.ImageSource{
		testPicture(t, 160, 120, color.RGBA{22, 119, 255, 255}, color.RGBA{114, 46, 209, 255}),
		testPicture(t, 120, 160, color.RGBA{82, 196, 26, 255}, color.RGBA{19, 194, 194, 255}),
	}
	tester := mount(func(c *ui.Context) {
		ImageGroup(c, pics, ImageProps{Width: 80, Height: 60})
	}, 800, 600)
	r, ok := tester.Find("图片")
	if !ok {
		t.Fatal("no image")
	}
	tester.Click("图片")
	tester.Frame()
	tester.Advance(300 * time.Millisecond)
	if !tester.HasText("1 / 2") {
		t.Fatalf("preview counter: %v", tester.Texts())
	}
	_ = r
	tester.Wheel(400, 300, 0, -1)
	tester.Frame()
	tester.Pinch(400, 300, 2)
	tester.Frame()
	if err := tester.Click("向右旋转"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if err := tester.Click("缩小"); err != nil {
		t.Fatal("zoomed in, so 缩小 is enabled:", err)
	}
	tester.Key(0, ui.KeyRight)
	tester.Frame()
	if !tester.HasText("2 / 2") {
		t.Fatalf("right arrow: %v", tester.Texts())
	}
	tester.ClickAt(30, 580)
	tester.Frame()
	tester.Advance(300 * time.Millisecond)
	if tester.HasText("2 / 2") {
		t.Fatal("a click outside the image closes the preview")
	}
	tester.Click("图片")
	tester.Frame()
	tester.Advance(300 * time.Millisecond)
	tester.Key(0, ui.KeyEscape)
	tester.Frame()
	tester.Advance(300 * time.Millisecond)
	if tester.HasText("1 / 2") {
		t.Fatal("escape closes the preview")
	}
}

func TestTourStepsAndFinishes(t *testing.T) {
	open, cur, done := true, 0, 0
	var save ui.Element
	tester := mount(func(c *ui.Context) {
		ui.Row(c).Gap(12).Children(func() {
			Button(c, "上传")
			save = ui.Row(c).Children(func() { Button(c, "保存", ButtonProps{Type: Primary}) })
		})
		Tour(c, &open, &cur, []TourStep{
			{Title: "欢迎使用", Description: "这是Demo App"},
			{Title: "保存设置", Description: "修改后点击保存", Target: func() ui.Element { return save }, Placement: PlaceBottom},
		}, TourProps{OnFinish: func() { done++ }})
	}, 800, 500)
	tester.Advance(200 * time.Millisecond)
	if !tester.HasText("欢迎使用") {
		t.Fatalf("first step %v", tester.Texts())
	}
	if err := tester.Click("下一步"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	tester.Frame()
	if cur != 1 || !tester.HasText("修改后点击保存") {
		t.Fatalf("second step %d %v", cur, tester.Texts())
	}
	if err := tester.Click("结束导览"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if open || done != 1 {
		t.Fatalf("finish open=%v done=%d", open, done)
	}
}

func TestListPagesAndQRRefresh(t *testing.T) {
	page := 1
	refreshed := false
	tester := mount(func(c *ui.Context) {
		List(c, 23, func(i int) {
			ListItem(c, func() {
				ListItemMeta(c, ListMetaProps{Title: "终端 " + itoa(i+1), Description: "在线"})
			}, ListItemProps{Actions: []func(){func() { Button(c.Key(i), "编辑"+itoa(i), ButtonProps{Type: Link}) }}})
		}, ListProps{Page: &page, PageSize: 5, Header: func() { Muted(c, "终端列表") }})
		if QRCode(c, "https://example.com", QRCodeProps{Status: QRExpired, Size: 120}) {
			refreshed = true
		}
	}, 700, 900)
	if !tester.HasText("终端 5") || tester.HasText("终端 6") {
		t.Fatal("first page shows 1-5")
	}
	if err := tester.Click("3"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !tester.HasText("终端 11") || tester.HasText("终端 5") {
		t.Fatalf("page 3 %v", tester.Texts())
	}
	if err := tester.Click("点击刷新"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if !refreshed || !tester.HasText("二维码过期") {
		t.Fatal("refresh")
	}
}
