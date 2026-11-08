package kit

import (
	"regexp"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestRules(t *testing.T) {
	noHan := regexp.MustCompile(`^[^\x{4e00}-\x{9fa5}]+$`)
	cases := []struct {
		name  string
		rule  Rule
		value string
		fails bool
	}{
		{"required empty", Required("必填"), "", true},
		{"required spaces", Required("必填"), "  ", true},
		{"required ok", Required("必填"), "a", false},
		{"pattern empty passes", Pattern(noHan, "无中文"), "", false},
		{"pattern han", Pattern(noHan, "无中文"), "密码abc", true},
		{"pattern ok", Pattern(noHan, "无中文"), "abc123!", false},
		{"minlen short", MinLen(6, "至少 6 位"), "abc", true},
		{"minlen runes", MinLen(3, "至少 3 位"), "中文字", false},
		{"ipv4 ok", IPv4("IP"), " 172.30.0.9 ", false},
		{"ipv4 bad octet", IPv4("IP"), "172.30.0.256", true},
		{"ipv4 v6", IPv4("IP"), "::1", true},
		{"ipv4 name", IPv4("IP"), "server.local", true},
		{"ipv4 empty passes", IPv4("IP"), "", false},
		{"range ok", IntRange(1, 65535, "端口"), "3306", false},
		{"range zero", IntRange(1, 65535, "端口"), "0", true},
		{"range text", IntRange(1, 65535, "端口"), "33a", true},
		{"range high", IntRange(1, 65535, "端口"), "70000", true},
	}
	for _, tc := range cases {
		if got := tc.rule(tc.value) != ""; got != tc.fails {
			t.Errorf("%s: fails = %v", tc.name, got)
		}
	}
	if got := CheckRules("", Required("请输入端口"), IntRange(1, 65535, "端口范围")); got != "请输入端口" {
		t.Errorf("first error = %q", got)
	}
	if got := CheckRules("0", Required("请输入端口"), IntRange(1, 65535, "端口范围")); got != "端口范围" {
		t.Errorf("second rule error = %q", got)
	}
	custom := func(v string) string { return pick(v == "admin", "不能用 admin", "") }
	if got := CheckRules("admin", Required("x"), custom); got != "不能用 admin" {
		t.Errorf("custom rule = %q", got)
	}
}

func TestFormShowsErrorsAfterEditOrValidate(t *testing.T) {
	var f Form
	host, port := "", "3306"
	frame := func() (string, string) {
		f.Begin()
		return f.Check("主机", host, Required("请输入地址")), f.Check("端口", port, Required("请输入端口"), IntRange(1, 65535, "端口范围"))
	}
	if h, p := frame(); h != "" || p != "" {
		t.Fatalf("untouched fields show %q %q", h, p)
	}
	port = "0"
	if h, p := frame(); h != "" || p != "端口范围" {
		t.Fatalf("after editing the port: %q %q", h, p)
	}
	port = "3306"
	frame()
	if err, shown := f.validate(); err != "请输入地址" || !shown {
		t.Fatalf("Validate = %q", err)
	}
	if h, _ := frame(); h != "请输入地址" {
		t.Fatalf("after Validate the untouched host shows %q", h)
	}
	host = "db.local"
	if h, p := frame(); h != "" || p != "" {
		t.Fatalf("fixed fields still show %q %q", h, p)
	}
	if err := errOf(f.validate()); err != "" {
		t.Fatalf("valid form = %q", err)
	}
	f.Begin()
	f.Check("端口", "0")
	if err := errOf(f.validate()); err != "" {
		t.Fatalf("a field not checked this frame failed the form: %q", err)
	}
	f.Reset()
	host = ""
	if h, _ := frame(); h != "" {
		t.Fatalf("after Reset: %q", h)
	}
}

func TestFormItemShowsValidationOnSubmit(t *testing.T) {
	var f Form
	host, saved := "", 0
	tester := ui.NewTester(func(c *ui.Context) {
		Page(c, Work, func() {
			f.Begin()
			err := f.Check("主机", host, Required("请输入数据库地址"))
			FormItem(c, "主机", func() { Input(c, &host, InputProps{Status: ErrorStatus(err)}) }, FormItemProps{Required: true, Error: err})
			if Button(c, "保存", ButtonProps{Type: Primary}) && f.Validate(c) == "" {
				saved++
			}
		})
	}, 400, 240)
	tester.SetLocale("zh-CN")
	tester.Frame()
	if tester.HasText("请输入数据库地址") {
		t.Fatal("error before submit")
	}
	if err := tester.Click("保存"); err != nil {
		t.Fatal(err)
	}
	tester.Frame()
	if saved != 0 || !tester.HasText("请输入数据库地址") {
		t.Fatalf("submit with an empty host: saved %d, error shown %v", saved, tester.HasText("请输入数据库地址"))
	}
	tester.Key(0, ui.KeyTab)
	tester.Type("db")
	tester.Frame()
	if tester.HasText("请输入数据库地址") {
		t.Fatal("error stayed after typing")
	}
	_ = tester.Click("保存")
	tester.Frame()
	if saved != 1 {
		t.Fatalf("saved = %d", saved)
	}
}

func errOf(s string, _ bool) string { return s }

func TestNameOfAndFormNestedValues(t *testing.T) {
	if NameKey("host") != "host" || NameKey("user.name") != "user.name" {
		t.Fatal(NameKey("user.name"))
	}
	if NameKey([]string{"user", "name"}) != "user.name" || NameKey(NamePath{"a", "b", "c"}) != "a.b.c" {
		t.Fatal(NameOf([]string{"user", "name"}))
	}
	if NameKey("") != "" || NameKey(nil) != "" {
		t.Fatal("empty name")
	}
	var f Form
	f.Begin()
	if e := f.Check([]string{"user", "name"}, "Ada", Required("请输入姓名")); e != "" {
		t.Fatalf("untouched nested: %q", e)
	}
	f.Check("user.email", "ada@corp.cn")
	f.Check("host", "db.local")
	got := f.Values()
	user, _ := got["user"].(map[string]any)
	if user["name"] != "Ada" || user["email"] != "ada@corp.cn" || got["host"] != "db.local" {
		t.Fatalf("Values = %#v", got)
	}
	if f.Value([]string{"user", "name"}) != "Ada" || f.Value("user.email") != "ada@corp.cn" {
		t.Fatal("Value path")
	}
	f.SetValues(map[string]any{"user": map[string]any{"name": "Bob"}})
	if f.Value("user.name") != "Bob" || f.Value("user.email") != "ada@corp.cn" {
		t.Fatalf("SetValues should merge: name=%q email=%q", f.Value("user.name"), f.Value("user.email"))
	}
	f.Depends([]string{"user", "name"}, "host")
	f.Begin()
	f.Check("host", "other")
	f.Check([]string{"user", "name"}, "Bob", Required("请输入姓名"))
	if !f.Changed("host") {
		t.Fatal("host should have changed")
	}
	f.SetError([]string{"user", "email"}, "已被占用")
	f.Begin()
	if e := f.Check("user.email", "ada@corp.cn"); e != "已被占用" {
		t.Fatalf("nested SetError = %q", e)
	}
}
