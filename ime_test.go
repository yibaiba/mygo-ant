package kit

import (
	"testing"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
)

func TestTakeCommittedSkipsWhileComposing(t *testing.T) {
	if takeCommitted(true, true) {
		t.Fatal("composing change must wait")
	}
	if !takeCommitted(true, false) {
		t.Fatal("committed change")
	}
	if takeCommitted(false, false) || takeCommitted(false, true) {
		t.Fatal("no change")
	}
}

func TestInputMaxLengthWaitsForCompositionEnd(t *testing.T) {
	value := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Padding(20).Width(280).Children(func() {
				Input(c, &value, InputProps{Prefix: "名称", MaxLength: 2, Width: 220})
			})
		})
	}, 400, 120)
	tester.SetLocale("zh-CN")
	tester.Frame()
	r, ok := tester.Find("名称")
	if !ok {
		t.Fatal("no field")
	}
	tester.ClickAt(r.X+r.W+40, r.Y+r.H/2)
	tester.Compose("你好世界", 4)
	tester.Frame()
	if utf8.RuneCountInString(value) > 2 {
		t.Fatalf("clamped preedit into %q", value)
	}
	tester.Type("你好世界")
	tester.Frame()
	if value != "你好" {
		t.Fatalf("after commit %q, want 你好", value)
	}
}

func TestOTPDoesNotAdvanceWhileComposing(t *testing.T) {
	code := ""
	tester := mount(func(c *ui.Context) {
		OTP(c, &code, OTPProps{Length: 4})
	}, 400, 200)
	x, y := center(t, tester, "验证码第 1 位")
	tester.ClickAt(x, y)
	tester.Compose("12", 1)
	tester.Frame()
	if tester.Focused("验证码第 2 位") {
		t.Fatalf("OTP advanced during compose, code=%q", code)
	}
	tester.Compose("", 0)
	tester.Type("1")
	tester.Frame()
	if code != "1" {
		t.Fatalf("after commit %q", code)
	}
	if !tester.Focused("验证码第 2 位") {
		t.Fatal("did not advance after composition ended")
	}
}

func TestMentionsSkipsCaretWhileComposing(t *testing.T) {
	value := ""
	opts := Opts("alice", "bob")
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Box(c).Width(320).Padding(16).Children(func() {
				Mentions(c, &value, opts, MentionsProps{Rows: 2})
			})
		})
	}, 400, 240)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.ClickAt(40, 40)
	tester.Compose("zhang", 5)
	tester.Frame()
	if tester.HasText("alice") || tester.HasText("bob") {
		t.Fatalf("mentions popup treated preedit as committed: %v", tester.Texts())
	}
}
