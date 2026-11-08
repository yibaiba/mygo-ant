package kit

import (
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// fakeDialog answers every open with paths and sends what it was asked.
func fakeDialog(paths ...string) (FileDialog, chan FileDialogOptions) {
	asked := make(chan FileDialogOptions, 4)
	return func(o FileDialogOptions) ([]string, error) {
		asked <- o
		return paths, nil
	}, asked
}

// settleUntil runs frames until ok holds, as the dialog answers on a
// goroutine of its own.
func settleUntil(t *testing.T, tester *ui.Tester, ok func() bool) {
	t.Helper()
	for i := 0; i < 200; i++ {
		tester.Frame()
		if ok() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("the picked files never arrived")
}

func asked(t *testing.T, ch chan FileDialogOptions) FileDialogOptions {
	t.Helper()
	select {
	case o := <-ch:
		return o
	case <-time.After(2 * time.Second):
		t.Fatal("the dialog was not opened")
	}
	return FileDialogOptions{}
}

func TestUploadOpensTheSetDialogAndAddsPicks(t *testing.T) {
	open, ch := fakeDialog("/fonts/a.ttf", "/notes.txt", "/fonts/b.OTF")
	SetFileDialog(open)
	t.Cleanup(func() { SetFileDialog(nil) })
	var files []string
	reported := false
	tester := mount(func(c *ui.Context) {
		if Upload(c, &files, UploadProps{Text: "拖入字体", Accept: []string{"ttf", "otf"}, Multiple: true}) {
			reported = true
		}
	}, 500, 300)
	if err := tester.Click("拖入字体"); err != nil {
		t.Fatal(err)
	}
	o := asked(t, ch)
	if o.Title != "拖入字体" || !o.Multiple || o.Directory || !slices.Equal(o.Accept, []string{"ttf", "otf"}) {
		t.Fatalf("dialog asked for %+v", o)
	}
	settleUntil(t, tester, func() bool { return len(files) > 0 })
	if !slices.Equal(files, []string{"/fonts/a.ttf", "/fonts/b.OTF"}) {
		t.Fatalf("files = %v: picks should be filtered by Accept as drops are", files)
	}
	if reported {
		t.Fatal("a click the kit opened a dialog for should not be reported")
	}
}

func TestUploadOnPickTakesThePicksOffTheBuild(t *testing.T) {
	open, ch := fakeDialog("/fonts/a.ttf")
	got := make(chan []string, 1)
	var files []string
	tester := mount(func(c *ui.Context) {
		Upload(c, &files, UploadProps{Text: "拖入字体", Pick: open, OnPick: func(p []string) { got <- p }})
	}, 500, 300)
	_ = tester.Click("拖入字体")
	asked(t, ch)
	select {
	case p := <-got:
		if !slices.Equal(p, []string{"/fonts/a.ttf"}) {
			t.Fatalf("OnPick got %v", p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("OnPick was not called")
	}
	tester.Frame()
	if len(files) != 0 {
		t.Fatal("with OnPick the caller adds the files")
	}
}

func TestPathPickerWritesThePickedPath(t *testing.T) {
	open, ch := fakeDialog("/opt/backup", "/ignored")
	value := ""
	changed := 0
	tester := mount(func(c *ui.Context) {
		if PathPicker(c, &value, "选择目录", PathPickerProps{Dialog: FileDialogOptions{Directory: true, Multiple: true}, Pick: open}) {
			changed++
		}
	}, 500, 200)
	_ = tester.Click("选择目录")
	o := asked(t, ch)
	if o.Title != "选择目录" || !o.Directory || o.Multiple {
		t.Fatalf("dialog asked for %+v: one folder, titled by the button", o)
	}
	settleUntil(t, tester, func() bool { return value != "" })
	if value != "/opt/backup" || changed != 1 {
		t.Fatalf("value %q, reported %d times", value, changed)
	}
}

func TestUploadListPicksAreAdded(t *testing.T) {
	open, ch := fakeDialog("/certs/a.pem", "/certs/b.crt", "/certs/c.key")
	var files []UploadFile
	var added []UploadFile
	picks := 0
	tester := mount(func(c *ui.Context) {
		ev := UploadList(c, &files, UploadListProps{Multiple: true, MaxCount: 1, Accept: []string{"pem", "crt"}, Pick: open, Text: "上传证书"})
		added = append(added, ev.Added...)
		if ev.Pick {
			picks++
		}
	}, 500, 300)
	_ = tester.Click("上传证书")
	if o := asked(t, ch); o.Multiple || o.Title != "上传证书" {
		t.Fatalf("MaxCount 1 should ask for one file: %+v", o)
	}
	settleUntil(t, tester, func() bool { return len(files) > 0 })
	if len(files) != 1 || files[0].Path != "/certs/b.crt" || len(added) != 1 || added[0].Path != "/certs/b.crt" {
		t.Fatalf("files %v, added %v: MaxCount 1 keeps the last accepted", files, added)
	}
	if picks != 0 {
		t.Fatal("Pick should stay false when the kit opened the dialog")
	}
}

func TestNoFileDialogReportsTheClick(t *testing.T) {
	SetFileDialog(nil)
	var files []string
	value := ""
	clicks := 0
	tester := mount(func(c *ui.Context) {
		if Upload(c, &files, UploadProps{Text: "拖入字体"}) {
			clicks++
		}
		if PathPicker(c, &value, "选择") {
			clicks++
		}
	}, 500, 400)
	_ = tester.Click("拖入字体")
	_ = tester.Click("选择")
	tester.Frame()
	if clicks != 2 {
		t.Fatalf("clicks = %d: without a dialog both clicks are the caller's", clicks)
	}
	if OpenFileDialog(nil, FileDialogOptions{}, func([]string) { t.Fatal("no dialog") }) {
		t.Fatal("OpenFileDialog should report that none is set")
	}
}

func TestSafeSaveFileName(t *testing.T) {
	if got := SafeSaveFileName(`/tmp/../diagnostic.zip`, "x.zip"); got != "diagnostic.zip" {
		t.Fatal(got)
	}
	if got := SafeSaveFileName("", ""); got != "diagnostic.zip" {
		t.Fatal(got)
	}
	if got := SafeSaveFileName(`bad:name*.zip`, "diagnostic.zip"); got != "badname.zip" {
		t.Fatal(got)
	}
}

func TestOpenSaveDialogUsesInjectedSave(t *testing.T) {
	asked := make(chan SaveDialogOptions, 1)
	SetSaveDialog(func(o SaveDialogOptions) (string, error) {
		asked <- o
		return "/tmp/out.zip", nil
	})
	t.Cleanup(func() { SetSaveDialog(nil) })
	got := make(chan string, 1)
	tester := mount(func(c *ui.Context) {
		if Button(c, "保存") {
			OpenSaveDialog(c, SaveDialogOptions{Title: "保存诊断包", DefaultName: "diagnostic.zip", Accept: []string{"zip"}}, func(path string) {
				got <- path
			})
		}
	}, 400, 120)
	if err := tester.Click("保存"); err != nil {
		t.Fatal(err)
	}
	select {
	case o := <-asked:
		if o.Title != "保存诊断包" || o.DefaultName != "diagnostic.zip" {
			t.Fatalf("%+v", o)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("save dialog not opened")
	}
	select {
	case p := <-got:
		if p != "/tmp/out.zip" {
			t.Fatal(p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("path not delivered")
	}
}
