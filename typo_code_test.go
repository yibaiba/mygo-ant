package kit

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCodeSpansPadAndLocateRuns(t *testing.T) {
	spans := []ui.Span{{Text: "运行 "}, {Text: "go test"}, {Text: " 再按 "}, {Text: "Ctrl"}}
	out, runs := codeSpans(spans, []TypoProps{{}, {Code: true}, {}, {Keyboard: true}})
	if len(runs) != 2 || runs[0].kbd || !runs[1].kbd {
		t.Fatalf("runs = %+v", runs)
	}
	if out[1].Text != codePad+"go test"+codeEnd || runs[0].start != 3 || runs[0].end != 3+10 {
		t.Fatalf("code run %+v over %q", runs[0], out[1].Text)
	}
	if runs[1].start != runs[0].end+4 || runs[1].end != runs[1].start+7 {
		t.Fatalf("keyboard run %+v", runs[1])
	}
}

func TestCodeTextStaysOnOneLineInARow(t *testing.T) {
	var code, plain ui.Rect
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Row(c).Gap(12).Children(func() {
				code = TypoText(c, "go test ./...", TypoProps{Code: true}).Bounds()
				plain = TypoText(c, "go").Bounds()
			})
		})
	}, 400, 200)
	tester.Frame()
	tester.Frame()
	if code.H > plain.H+0.5 {
		t.Fatalf("code text is %v high next to %v: its end pad wrapped", code.H, plain.H)
	}
}

func TestRichParagraphBoxesMatchLayout(t *testing.T) {
	var b ui.Rect
	runs := []TypoRun{
		{Text: "升级前请先在终端执行 "},
		{Text: "systemctl stop example-server.service", Props: TypoProps{Code: true}},
		{Text: "，确认停止后按 "},
		{Text: "Enter", Props: TypoProps{Keyboard: true}},
		{Text: " 继续。"},
	}
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ui.Column(c).Width(320).Padding(10).Children(func() {
				b = RichParagraph(c, runs...).Bounds()
			})
		})
	}, 400, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	tester.Frame()
	spans := make([]ui.Span, len(runs))
	kinds := make([]TypoProps, len(runs))
	for i, r := range runs {
		spans[i] = typoSpan(Work, r.Text, r.Props, Work.Text, Work.FontSize)
		kinds[i] = r.Props
	}
	spans, codes := codeSpans(spans, kinds)
	l := ui.ShapeRichText(spans, typoFont(Work, Work.FontSize), b.W)
	w, h := l.Size()
	if h < b.H-0.5 || h > b.H+0.5 || w > b.W+0.5 {
		t.Fatalf("shaped %vx%v, laid out %vx%v: the boxes would not sit on the text", w, h, b.W, b.H)
	}
	if rects := l.SelectionRects(ui.TextInputRange{Start: codes[0].start, End: codes[0].end}); len(rects) < 2 {
		t.Fatalf("the long command should wrap over lines, boxed on each: %v", rects)
	}
}

func TestGalleryTypographyCodeShot(t *testing.T) {
	for _, mode := range []struct {
		tok  *Tokens
		name string
	}{{Work, "mygo-kit-gallery-typography-code.png"}, {WorkDark, "mygo-kit-gallery-dark-typography-code.png"}} {
		tok := mode.tok
		tester := ui.NewTester(func(c *ui.Context) {
			Page(c, tok, func() {
				ui.Column(c).Padding(24).Gap(12).Width(440).Children(func() {
					ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
						TypoText(c, "go test ./...", TypoProps{Code: true})
						TypoText(c, "Ctrl", TypoProps{Keyboard: true})
						TypoText(c, "+")
						TypoText(c, "Shift", TypoProps{Keyboard: true})
						TypoText(c, "强调代码", TypoProps{Code: true, Strong: true})
					})
					Paragraph(c, "systemctl restart example-server.service && journalctl -u example-server -f", TypoProps{Code: true})
					RichParagraph(c,
						TypoRun{Text: "升级前请先在终端执行 "},
						TypoRun{Text: "systemctl stop example-server.service", Props: TypoProps{Code: true}},
						TypoRun{Text: "，确认服务停止后按 "},
						TypoRun{Text: "Enter", Props: TypoProps{Keyboard: true}},
						TypoRun{Text: " 继续；如需中断请按 "},
						TypoRun{Text: "Ctrl", Props: TypoProps{Keyboard: true}},
						TypoRun{Text: " + "},
						TypoRun{Text: "C", Props: TypoProps{Keyboard: true}},
						TypoRun{Text: "。"},
					)
				})
			})
		}, 500, 300)
		tester.SetLocale("zh-CN")
		tester.Frame()
		tester.Frame()
		shot(t, tester, mode.name)
	}
}
