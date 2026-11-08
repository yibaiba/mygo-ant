package kit

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func paths(files []UploadFile) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = filepath.Base(f.Path)
	}
	return out
}

func TestAddUploads(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "certs", "inner")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"certs/a.pem", "certs/inner/b.pem", "certs/inner/notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	folder := filepath.Join(dir, "certs")
	got, added := AddUploads(nil, []string{folder, "/x/c.pem"}, UploadListProps{Multiple: true, Directory: true, Accept: []string{"pem"}})
	if !slices.Equal(paths(got), []string{"a.pem", "b.pem", "c.pem"}) || len(added) != 3 {
		t.Fatalf("directory drop = %v", paths(got))
	}
	if got, _ := AddUploads(nil, []string{folder}, UploadListProps{Multiple: true}); len(got) != 0 {
		t.Fatal("a folder without Directory should be skipped")
	}
	again, added := AddUploads(got, []string{"/x/c.pem", "/x/d.pem"}, UploadListProps{Multiple: true, MaxCount: 4})
	if !slices.Equal(paths(again), []string{"a.pem", "b.pem", "c.pem", "d.pem"}) || len(added) != 1 {
		t.Fatalf("dedupe + max = %v added %v", paths(again), paths(added))
	}
	capped, added := AddUploads(again, []string{"/x/e.pem"}, UploadListProps{Multiple: true, MaxCount: 4})
	if len(capped) != 4 || len(added) != 0 {
		t.Fatal("MaxCount should cap the list")
	}
	one, _ := AddUploads(again, []string{"/x/f.pem", "/x/g.pem"}, UploadListProps{Multiple: true, MaxCount: 1})
	if !slices.Equal(paths(one), []string{"g.pem"}) {
		t.Fatalf("MaxCount 1 should replace with the last: %v", paths(one))
	}
	if !isImagePath("/a/B.PNG") || isImagePath("/a/b.pem") {
		t.Fatal("isImagePath")
	}
}

func TestFormListKeys(t *testing.T) {
	var k formListKeys
	k.sync(3)
	if !slices.Equal(k.ids, []string{"row-1", "row-2", "row-3"}) {
		t.Fatal(k.ids)
	}
	k.ids = append(k.ids[:1:1], k.ids[2:]...)
	k.sync(3)
	if !slices.Equal(k.ids, []string{"row-1", "row-3", "row-4"}) {
		t.Fatalf("keys after remove + add = %v", k.ids)
	}
}

func TestFormDependsChangedAndSetError(t *testing.T) {
	var f Form
	f.Depends("confirm", "password")
	pass, confirm := "secret1", ""
	frame := func() (string, string) {
		f.Begin()
		e1 := f.Check("password", pass, MinLen(6, "至少 6 位"))
		e2 := f.Check("confirm", confirm, Match(pass, "两次输入的密码不一致"))
		return e1, e2
	}
	frame()
	confirm = "secret1"
	if _, e := frame(); e != "" {
		t.Fatalf("matching confirm errs %q", e)
	}
	pass = "secret2"
	if _, e := frame(); e != "两次输入的密码不一致" {
		t.Fatalf("a changed dependency should re-check confirm, got %q", e)
	}
	if !f.Changed("password") || f.Value("password") != "secret2" {
		t.Fatal("Changed / Value")
	}
	frame()
	if f.Changed("password") {
		t.Fatal("Changed should clear in the next frame")
	}
	f.SetError("password", "服务器拒绝了该密码")
	if e, _ := frame(); e != "服务器拒绝了该密码" {
		t.Fatalf("SetError not shown: %q", e)
	}
	pass = "secret3"
	confirm = "secret3"
	if e, _ := frame(); e != "" {
		t.Fatalf("SetError should clear on change: %q", e)
	}
	var g Form
	g.SetError("host", "地址已被占用")
	g.Begin()
	if e := g.Check("host", "10.0.0.1"); e != "地址已被占用" || g.Touched("host") {
		t.Fatalf("an error set before the first check should show without touching: %q", e)
	}
}

func TestFormViewLayouts(t *testing.T) {
	layout := FormHorizontal
	name := ""
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			FormView(c, func() {
				FormItem(c, "名称", func() { Input(c, &name) }, FormItemProps{Required: true, Tooltip: "会议室名称"})
				FormItem(c, "备注", func() { Input(c, new(string)) })
			}, FormProps{FormConfig: FormConfig{Layout: layout, RequiredMark: MarkOptional}})
		})
	}, 800, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	label, _ := tester.Find("名称")
	if label.X < 100 {
		t.Fatalf("horizontal label should sit right in its column, at %v", label.X)
	}
	if !tester.HasText("（可选）") || !tester.HasText(":") {
		t.Fatal("optional mark and colon should show", tester.Texts())
	}
	layout = FormVertical
	tester.Frame()
	if l, _ := tester.Find("名称"); l.X > 40 || tester.HasText(":") {
		t.Fatalf("vertical label at %v, colon shown %v", l.X, tester.HasText(":"))
	}
	layout = FormInline
	tester.Frame()
	a, _ := tester.Find("名称")
	b, _ := tester.Find("备注")
	if a.Y != b.Y {
		t.Fatalf("inline items should share a line: %v vs %v", a.Y, b.Y)
	}
}

func TestFormViewDisablesAndFormListAddsRemoves(t *testing.T) {
	type host struct{ IP, Port string }
	hosts := []host{{"10.0.0.1", "8080"}}
	off := false
	saved := 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			FormView(c, func() {
				FormList(c, &hosts, func() host { return host{} }, func(i int, h *host, remove func()) {
					ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Center).Children(func() {
						Input(c, &h.IP, InputProps{Placeholder: "地址 " + itoa(i+1)})
						if FormListRemove(c, "删除第 "+itoa(i+1)+" 行") {
							remove()
						}
					})
				}, FormListProps{Add: "添加地址", Max: 3})
				if Button(c, "保存") {
					saved++
				}
			}, FormProps{Disabled: off})
		})
	}, 600, 400)
	tester.SetLocale("zh-CN")
	tester.Frame()
	_ = tester.Click("添加地址")
	tester.Frame()
	_ = tester.Click("添加地址")
	tester.Frame()
	if len(hosts) != 3 || tester.HasText("添加地址") {
		t.Fatalf("hosts %d; add should hide at Max", len(hosts))
	}
	if err := tester.Click("删除第 1 行"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if len(hosts) != 2 || hosts[0].IP != "" {
		t.Fatalf("remove should drop the first row: %+v", hosts)
	}
	off = true
	tester.Frame()
	_ = tester.Click("保存")
	_ = tester.Click("删除第 1 行")
	tester.Frame()
	if saved != 0 || len(hosts) != 2 {
		t.Fatal("a disabled form should take no clicks")
	}
}

func writePNG(t *testing.T, path string, c color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	for y := 0; y < 48; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestUploadListDropsRemovesAndPreviews(t *testing.T) {
	dir := t.TempDir()
	pic := filepath.Join(dir, "room.png")
	writePNG(t, pic, color.RGBA{30, 89, 215, 255})
	var files []UploadFile
	listType := ListText
	var last UploadEvent
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			ev := UploadList(c, &files, UploadListProps{ListType: listType, Multiple: true, MaxCount: 3})
			if ev.Pick || ev.Added != nil || ev.Removed != nil || ev.Preview != nil {
				last = ev
			}
		})
	}, 600, 500)
	tester.SetLocale("zh-CN")
	tester.Frame()
	_ = tester.Click("上传")
	tester.Frame()
	if !last.Pick {
		t.Fatal("the button should report Pick")
	}
	r, _ := tester.Find("上传")
	tester.DropFiles(r.X+5, r.Y+5, pic, "/tmp/agenda.pdf")
	tester.Frame()
	if len(files) != 2 || len(last.Added) != 2 {
		t.Fatalf("drop added %d", len(files))
	}
	files[1].Status, files[1].Percent = UploadUploading, 40
	files = append(files, UploadFile{Path: "/tmp/bad.doc", Status: UploadError, Error: "网络错误"})
	tester.Frame()
	if tester.HasText("上传") {
		t.Fatal("the trigger should hide at MaxCount")
	}
	if err := tester.Click("预览文件 room.png"); err != nil {
		t.Fatal(err)
	}
	tester.Advance(time.Second)
	if _, open := tester.Find("关闭预览"); last.Preview == nil || last.Preview.Path != pic || !open {
		t.Fatalf("preview %+v", last.Preview)
	}
	tester.Key(0, ui.KeyEscape)
	tester.Advance(time.Second)
	listType = ListPictureCard
	tester.Frame()
	r, _ = tester.Find("上传中")
	if r.W == 0 {
		t.Fatal("an uploading card should say so")
	}
	if err := tester.Click("移除 bad.doc"); err == nil {
		tester.Frame()
	} else {
		c, _ := tester.Find("bad.doc")
		tester.Move(c.X+c.W/2, c.Y+c.H/2)
		tester.Frame()
		if err := tester.Click("移除 bad.doc"); err != nil {
			t.Fatal(err)
		}
		tester.Frame()
	}
	if len(files) != 2 || len(last.Removed) != 1 || last.Removed[0].Path != "/tmp/bad.doc" {
		t.Fatalf("remove: %v", paths(files))
	}
}

func TestUploadActionPostsMultipart(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "agenda.pdf")
	if err := os.WriteFile(src, []byte("hello-upload"), 0o644); err != nil {
		t.Fatal(err)
	}
	gotName, gotField := "", ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gotField = r.FormValue("room")
		f, hdr, err := r.FormFile("file")
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer f.Close()
		gotName = hdr.Filename
		io.Copy(io.Discard, f)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	var files []UploadFile
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			UploadList(c, &files, UploadListProps{Multiple: true, Action: srv.URL, Data: map[string]string{"room": "一号"}})
		})
	}, 500, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	r, _ := tester.Find("上传")
	tester.DropFiles(r.X+5, r.Y+5, src)
	tester.Frame()
	if len(files) != 1 {
		t.Fatalf("should add the file: %+v", files)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		tester.Frame()
		if len(files) == 1 && files[0].Status == UploadDone && files[0].Percent == 100 {
			if gotName != "agenda.pdf" || gotField != "一号" {
				t.Fatalf("posted name=%q field=%q", gotName, gotField)
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("upload did not finish: %+v", files)
}

func TestUploadCustomRequestWritesProgress(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "note.txt")
	if err := os.WriteFile(src, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var files []UploadFile
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			UploadList(c, &files, UploadListProps{
				CustomRequest: func(req UploadRequest) {
					req.OnProgress(40)
					req.OnSuccess(nil)
				},
			})
		})
	}, 500, 300)
	tester.SetLocale("zh-CN")
	tester.Frame()
	r, _ := tester.Find("上传")
	tester.DropFiles(r.X+5, r.Y+5, src)
	tester.Frame()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		tester.Frame()
		if len(files) == 1 && files[0].Status == UploadDone {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("customRequest did not finish: %+v", files)
}
