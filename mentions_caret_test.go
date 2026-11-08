package kit

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCaretRectMeasuresPrefixAndLines(t *testing.T) {
	font := ui.Font{Family: Work.Font, Size: Work.FontSize, Lang: Work.Lang}
	a := caretRect("hello", 0, font, 0)
	b := caretRect("hello", 5, font, 0)
	if a.X != 0 || b.X <= a.X || b.Y != a.Y {
		t.Fatalf("caret 0 at %v, 5 at %v: should move right on one line", a, b)
	}
	if w := caretRect("你好", 2, font, 0).X; w < 2*Work.FontSize-1 || w > 2*Work.FontSize+1 {
		t.Fatalf("two CJK glyphs are %v wide, want about %v", w, 2*Work.FontSize)
	}
	nl := caretRect("ab\nab", 4, font, 0)
	first := caretRect("ab\nab", 1, font, 0)
	if nl.Y <= first.Y || nl.X != first.X {
		t.Fatalf("after a newline the caret should go a line down at the same x: %v vs %v", nl, first)
	}
	wrapped := caretRect("会议室会议室会议室会议室", 12, font, 5*Work.FontSize)
	if wrapped.Y < 2*a.H-1 {
		t.Fatalf("12 glyphs wrapped 5 to a line should end on the third line, at %v", wrapped)
	}
	if r := caretInView(ui.Rect{Y: 60, H: 20}, 44); r.Y != 24 {
		t.Fatalf("caretInView scrolled to %v, want 24", r.Y)
	}
	if r := caretInView(ui.Rect{Y: 10, H: 20}, 44); r.Y != 10 {
		t.Fatal("a caret in view should stay")
	}
}

func TestMentionsPopupFollowsCaret(t *testing.T) {
	value := ""
	opts := []Option{{Value: "alice", Label: "Alice"}, {Value: "bob", Label: "Bob"}}
	tester := mount(func(c *ui.Context) {
		ui.Box(c).Width(360).Children(func() { Mentions(c, &value, opts, MentionsProps{Rows: 3}) })
	}, 600, 400)
	tester.ClickAt(100, 30)
	tester.Type("@b")
	tester.Frame()
	tester.Frame()
	short, _ := tester.Find("Bob")
	caretA, _ := tester.TextCaret()
	tester.Key(0, ui.KeyEscape)
	tester.Type(" 会议纪要请同步给 @b")
	tester.Frame()
	tester.Frame()
	long, ok := tester.Find("Bob")
	if !ok {
		t.Fatal("popup closed")
	}
	caretB, _ := tester.TextCaret()
	if long.X-short.X < 100 {
		t.Fatalf("popup x %v → %v: should move with the caret (%v → %v)", short.X, long.X, caretA.X, caretB.X)
	}
	panelX := long.X - 4
	if d := panelX - caretB.X; d < -6 || d > 6 {
		t.Fatalf("panel starts at about %v, the caret is at %v", panelX, caretB.X)
	}
	tester.Type("\n\n@a")
	tester.Frame()
	tester.Frame()
	third, ok := tester.Find("Alice")
	caretC, _ := tester.TextCaret()
	if !ok || third.Y < caretC.Y+caretC.H {
		t.Fatalf("popup at %v should be under the caret %v on its new line", third, caretC)
	}
}

func TestGalleryMentionsShot(t *testing.T) {
	value := ""
	opts := []Option{{Value: "张三", Label: "张三"}, {Value: "张伟", Label: "张伟"}, {Value: "李四", Label: "李四"}}
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(24).Gap(8).Width(420).Children(func() {
				ui.Text(c, "Mentions：弹层跟着光标走（第二行、句中）").FontSize(14).TextColor(Work.TextSecondary)
				Mentions(c, &value, opts, MentionsProps{Rows: 3})
			})
		})
	}, 520, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.ClickAt(120, 70)
	tester.Type("会议纪要已整理完毕\n请 @张")
	tester.Frame()
	tester.Frame()
	shot(t, tester, "mygo-kit-gallery-mentions.png")
}
