package kit

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

type galleryForm struct {
	tok                      *Tokens
	files, drag, pics, cards []UploadFile
	circles, full            []UploadFile
	name, room, note, pass   string
	confirm, kw, status      string
	hosts                    []galleryHost
	form                     Form
	seats                    string
	agree                    bool
}

type galleryHost struct{ IP, Port string }

// gradientPNG writes a w×h picture shading from a to b, for thumbnails.
func gradientPNG(t *testing.T, path string, a, b color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, 160, 120))
	for y := 0; y < 120; y++ {
		for x := 0; x < 160; x++ {
			k := float64(x+y) / 280
			mix := func(p, q uint8) uint8 { return uint8(float64(p)*(1-k) + float64(q)*k) }
			img.Set(x, y, color.RGBA{mix(a.R, b.R), mix(a.G, b.G), mix(a.B, b.B), 255})
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

func newGalleryForm(t *testing.T) *galleryForm {
	dir := t.TempDir()
	p1, p2, p3 := filepath.Join(dir, "会议室.png"), filepath.Join(dir, "签到表.png"), filepath.Join(dir, "logo.png")
	gradientPNG(t, p1, color.RGBA{30, 89, 215, 255}, color.RGBA{135, 208, 104, 255})
	gradientPNG(t, p2, color.RGBA{250, 173, 20, 255}, color.RGBA{255, 77, 79, 255})
	gradientPNG(t, p3, color.RGBA{114, 46, 209, 255}, color.RGBA{19, 194, 194, 255})
	return &galleryForm{
		tok: Work,
		files: []UploadFile{
			{Path: "/data/议程.docx"},
			{Path: "/data/会议纪要.pdf", Status: UploadUploading, Percent: 45},
			{Path: "/data/附件.zip", Status: UploadError, Error: "上传失败：网络中断"},
		},
		drag: []UploadFile{{Path: "/certs/server.pem"}},
		pics: []UploadFile{
			{Path: p1},
			{Path: "/data/大合照.jpg", Status: UploadUploading, Percent: 70},
			{Path: "/data/损坏.png", Status: UploadError, Error: "文件损坏"},
		},
		cards: []UploadFile{
			{Path: p1}, {Path: p2},
			{Path: "/data/上传中.png", Status: UploadUploading, Percent: 60},
			{Path: "/data/失败.png", Status: UploadError, Error: "超过 2MB"},
		},
		circles: []UploadFile{{Path: p3}},
		full:    []UploadFile{{Path: p1}, {Path: p2}},
		name:    "第三会议室",
		pass:    "secret1",
		confirm: "secret2",
		hosts:   []galleryHost{{"172.30.0.9", "36678"}, {"172.30.0.10", "36678"}},
		seats:   "24",
	}
}

func (g *galleryForm) uploads(c *ui.Context) {
	tok := g.tok
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 上传").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(16).Gap(16).Children(func() {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(16).Children(func() {
				panel(c, tok, "文本列表 listType=text", func() {
					UploadList(c, &g.files, UploadListProps{Multiple: true})
				})
				panel(c, tok, "拖拽上传 Dragger + 目录", func() {
					UploadList(c, &g.drag, UploadListProps{Dragger: true, Directory: true, Multiple: true, Hint: "支持拖入整个文件夹，只保留 .pem / .crt", Accept: []string{"pem", "crt"}})
				})
				panel(c, tok, "图片列表 listType=picture", func() {
					UploadList(c, &g.pics, UploadListProps{ListType: ListPicture, Multiple: true})
				})
			})
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(16).Children(func() {
				panel(c, tok, "照片墙 picture-card（悬停显示预览 / 删除）", func() {
					UploadList(c, &g.cards, UploadListProps{ListType: ListPictureCard, Multiple: true})
				})
				panel(c, tok, "圆形 picture-circle", func() {
					UploadList(c, &g.circles, UploadListProps{ListType: ListPictureCircle, Multiple: true})
				})
				panel(c, tok, "maxCount=2（已满，隐藏上传）/ 禁用", func() {
					UploadList(c, &g.full, UploadListProps{ListType: ListPictureCard, Multiple: true, MaxCount: 2})
					UploadList(c, new([]UploadFile), UploadListProps{Disabled: true})
				})
			})
		})
	})
}

func (g *galleryForm) forms(c *ui.Context) {
	tok := g.tok
	f := &g.form
	Page(c, tok, func() {
		ui.Row(c).FillWidth().Height(40).Background(tok.Brand).Padding(0, 16).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "组件库 · 表单布局").FontSize(16).FontWeight(600).TextColor(tok.TextOnPrimary)
		})
		ui.Row(c).Fill().AlignItems(ui.Start).Padding(16).Gap(16).Children(func() {
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(16).Children(func() {
				panel(c, tok, "水平布局 labelCol=6 wrapperCol=18", func() {
					f.Begin()
					FormView(c, func() {
						nameErr := f.Check("name", g.name, Required("请输入会议室名称"))
						FormItem(c, "会议室名称", func() {
							Input(c, &g.name, InputProps{Status: ErrorStatus(nameErr)})
						}, FormItemProps{Required: true, Tooltip: "显示在签到屏上", Error: nameErr, Extra: "最多 20 个字"})
						FormItem(c, "容纳人数", func() {
							InputNumber(c, &g.seats, InputNumberProps{Min: 1, Max: 200})
						}, FormItemProps{Warning: "超过 20 人建议开启分组签到"})
						passErr := f.Check("pass", g.pass, MinLen(6, "至少 6 位"))
						FormItem(c, "管理密码", func() {
							Input(c, &g.pass, InputProps{Password: true, Status: ErrorStatus(passErr)})
						}, FormItemProps{Required: true, Error: passErr})
						f.Depends("confirm", "pass")
						confirmErr := f.Check("confirm", g.confirm, Match(g.pass, "两次输入的密码不一致"))
						FormItem(c, "确认密码", func() {
							Input(c, &g.confirm, InputProps{Password: true, Status: ErrorStatus(confirmErr)})
						}, FormItemProps{Required: true, Error: confirmErr, Help: "dependencies: 管理密码"})
						FormItem(c, "", func() {
							ui.Row(c).Gap(8).Children(func() {
								Button(c, "保存", ButtonProps{Type: Primary})
								Button(c, "重置")
							})
						})
					}, FormProps{FormConfig: FormConfig{Layout: FormHorizontal}})
					f.Validate(c)
				})
				panel(c, tok, "动态增减 Form.List", func() {
					FormView(c, func() {
						FormList(c, &g.hosts, func() galleryHost { return galleryHost{Port: "36678"} }, func(i int, h *galleryHost, remove func()) {
							ui.Row(c).FillWidth().Gap(8).AlignItems(ui.Center).Children(func() {
								ui.Box(c).Grow(1).Children(func() { Input(c, &h.IP, InputProps{Placeholder: "地址"}) })
								ui.Box(c).Width(120).Children(func() { Input(c, &h.Port, InputProps{Placeholder: "端口"}) })
								if FormListRemove(c, "删除") {
									remove()
								}
							})
						}, FormListProps{Add: "添加服务器", Max: 5})
					})
				})
			})
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(16).Children(func() {
				panel(c, tok, "行内布局 layout=inline", func() {
					FormView(c, func() {
						FormItem(c, "关键字", func() { Input(c, &g.kw, InputProps{Placeholder: "会议名称", Width: 160}) })
						FormItem(c, "状态", func() {
							Select(c, &g.status, Opts("进行中", "已结束"), SelectProps{Width: 120, Placeholder: "全部"})
						})
						FormItem(c, "", func() { Button(c, "查询", ButtonProps{Type: Primary}) })
					}, FormProps{FormConfig: FormConfig{Layout: FormInline}})
				})
				panel(c, tok, "requiredMark=optional · labelAlign=left · 无冒号", func() {
					FormView(c, func() {
						FormItem(c, "联系人", func() { Input(c, &g.room) }, FormItemProps{Required: true})
						FormItem(c, "备注", func() { Input(c, &g.note) }, FormItemProps{Help: "选填项在标签后标注“（可选）”"})
					}, FormProps{FormConfig: FormConfig{Layout: FormHorizontal, LabelWidth: 110, LabelLeft: true, NoColon: true, RequiredMark: MarkOptional}})
				})
				panel(c, tok, "垂直布局 · 禁用表单 · size=small", func() {
					FormView(c, func() {
						FormRow(c, func() {
							FormItem(c, "会议室名称", func() { Input(c, &g.name) }, FormItemProps{Min: 160, Required: true})
							FormItem(c, "人数", func() { InputNumber(c, &g.seats) }, FormItemProps{Min: 120})
						})
						FormItem(c, "", func() { Checkbox(c, &g.agree, "同意会议室使用规范") })
						Button(c, "提交", ButtonProps{Type: Primary})
					}, FormProps{Disabled: true, Size: Small})
				})
			})
		})
	})
}

func TestGalleryFormShots(t *testing.T) {
	g := newGalleryForm(t)
	up := ui.NewTester(g.uploads, 1440, 860)
	up.SetLocale("zh-CN")
	up.Frame()
	up.Frame()
	if r, ok := up.Find("签到表.png"); ok {
		up.Move(r.X+r.W/2, r.Y+r.H/2)
	}
	up.Frame()
	for _, want := range []string{"议程.docx", "上传中", "server.pem"} {
		if !up.HasText(want) {
			t.Fatalf("upload gallery missing %q", want)
		}
	}
	shot(t, up, "mygo-kit-gallery-upload.png")

	fg := newGalleryForm(t)
	fm := ui.NewTester(fg.forms, 1440, 900)
	fm.SetLocale("zh-CN")
	fm.Frame()
	fm.Frame()
	fm.Frame()
	for _, want := range []string{"两次输入的密码不一致", "添加服务器", "（可选）"} {
		if !fm.HasText(want) {
			t.Fatalf("form gallery missing %q", want)
		}
	}
	shot(t, fm, "mygo-kit-gallery-form.png")

	dg := newGalleryForm(t)
	dg.tok = WorkDark
	dark := ui.NewTester(dg.forms, 1440, 900)
	dark.SetLocale("zh-CN")
	dark.Frame()
	dark.Frame()
	dark.Frame()
	shot(t, dark, "mygo-kit-gallery-dark-form.png")
}
