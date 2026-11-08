package kit

import (
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestMessagesShowMoreThanThree(t *testing.T) {
	var m MessageAPI
	for i := range 6 {
		m.Info("消息 " + strconv.Itoa(i))
	}
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {})
		Messages(c, MessagesProps{API: &m})
	}, 600, 500)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Advance(300 * time.Millisecond)
	prev := float32(-1)
	for i := range 6 {
		r, ok := tester.Find("消息 " + strconv.Itoa(i))
		if !ok {
			t.Fatalf("message %d not shown: %v", i, tester.Texts())
		}
		if r.Y <= prev {
			t.Fatalf("message %d at y=%v, above the one before at %v: oldest should be on top", i, r.Y, prev)
		}
		prev = r.Y
	}
}

func TestMessagesMaxCountAndExpiryOrder(t *testing.T) {
	var m MessageAPI
	closed := []string{}
	open := func(key string, d time.Duration) {
		m.Open(MessageConfig{Key: key, Content: key, Duration: d, OnClose: func() { closed = append(closed, key) }})
	}
	open("a", 2*time.Second)
	open("b", time.Second)
	open("c", 0)
	open("d", -1)
	max := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {})
		Messages(c, MessagesProps{API: &m, MaxCount: max})
	}, 600, 400)
	tester.Frame()
	if got := m.Keys(); !slices.Equal(got, []string{"a", "b", "c", "d"}) {
		t.Fatalf("keys = %v", got)
	}
	tester.Move(5, 395)
	tester.Advance(1100 * time.Millisecond)
	tester.Frame()
	if !slices.Equal(closed, []string{"b"}) {
		t.Fatalf("after 1.1s closed = %v, want b", closed)
	}
	tester.Advance(1000 * time.Millisecond)
	tester.Frame()
	if !slices.Equal(closed, []string{"b", "a"}) {
		t.Fatalf("after 2.1s closed = %v, want b then a", closed)
	}
	tester.Advance(1000 * time.Millisecond)
	tester.Frame()
	if !slices.Equal(closed, []string{"b", "a", "c"}) || !slices.Equal(m.Keys(), []string{"d"}) {
		t.Fatalf("after 3.1s closed = %v, showing %v: the default is 3s, -1 stays", closed, m.Keys())
	}
	open("e", -1)
	open("f", -1)
	open("d", -1)
	max = 2
	tester.Frame()
	if got := m.Keys(); !slices.Equal(got, []string{"e", "f"}) {
		t.Fatalf("MaxCount 2 kept %v, want the newest e, f (d was updated in place, oldest)", got)
	}
	m.Destroy("")
	tester.Frame()
	if len(m.Keys()) != 0 || closed[len(closed)-1] != "f" {
		t.Fatalf("destroy all left %v, closed %v", m.Keys(), closed)
	}
}

func TestMessagesPauseOnHover(t *testing.T) {
	var m MessageAPI
	m.Open(MessageConfig{Key: "k", Content: "悬停暂停", Duration: time.Second})
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {})
		Messages(c, MessagesProps{API: &m})
	}, 600, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Advance(300 * time.Millisecond)
	r, _ := tester.Find("悬停暂停")
	tester.Move(r.X+r.W/2, r.Y+r.H/2)
	tester.Frame()
	tester.Advance(3 * time.Second)
	tester.Frame()
	if len(m.Keys()) != 1 {
		t.Fatal("a hovered message should not close")
	}
	tester.Move(5, 395)
	tester.Frame()
	tester.Advance(1200 * time.Millisecond)
	tester.Frame()
	if len(m.Keys()) != 0 {
		t.Fatal("the message should close once the pointer leaves")
	}
}

func TestMessagesDrainToastsInOrder(t *testing.T) {
	sent := false
	closed := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {})
		if !sent {
			for i := range 5 {
				c.AddToast(ui.Toast{Title: "提示 " + strconv.Itoa(i), Type: "success", Timeout: time.Second, OnClose: func() { closed++ }})
			}
			sent = true
		}
		Messages(c)
	}, 600, 500)
	tester.SetLocale("zh-CN")
	for range 4 {
		tester.Frame()
	}
	tester.Advance(300 * time.Millisecond)
	prev := float32(-1)
	for i := range 5 {
		r, ok := tester.Find("提示 " + strconv.Itoa(i))
		if !ok {
			t.Fatalf("toast %d not shown: %v", i, tester.Texts())
		}
		if r.Y <= prev {
			t.Fatalf("toast %d at %v, not under the one before (%v)", i, r.Y, prev)
		}
		prev = r.Y
	}
	if closed != 0 {
		t.Fatalf("draining ran OnClose %d times", closed)
	}
	tester.Move(5, 495)
	tester.Advance(2 * time.Second)
	tester.Frame()
	if closed != 5 {
		t.Fatalf("OnClose ran %d times after the toasts' time, want 5", closed)
	}
}

func TestGalleryMessagesShot(t *testing.T) {
	var m MessageAPI
	m.Success("服务器安装成功")
	m.Info("正在检查更新")
	m.Warning("磁盘剩余空间不足 10%")
	m.Error("网络连接失败")
	m.Loading("正在上传离线包…")
	m.Success("所有服务已启动")
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Fill().Padding(24).Children(func() {
				ui.Text(c, "message：六条同时显示（不再受 ui 三条上限），最早的在上").FontSize(14).TextColor(Work.TextSecondary)
			})
		})
		Messages(c, MessagesProps{API: &m, Top: 56})
	}, 900, 420)
	tester.SetLocale("zh-CN")
	tester.Frame()
	shot(t, tester, "mygo-kit-gallery-messages.png")
}
