package kit

import (
	"runtime"
	"strings"
	"testing"
)

func TestFontListPutsSystemUIFirst(t *testing.T) {
	list := FontList()
	if list == "" {
		t.Fatal("empty FontList")
	}
	if strings.HasPrefix(list, ProductFont) {
		t.Fatalf("WenQuanYi must not be first: %q", list)
	}
	first := strings.TrimSpace(strings.Split(list, ",")[0])
	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(first, "Segoe") {
			t.Fatalf("windows first family %q", first)
		}
	case "darwin":
		if !strings.Contains(first, "SF Pro") && !strings.Contains(first, "PingFang") {
			t.Fatalf("darwin first family %q", first)
		}
	default:
		if first != "system-ui" {
			t.Fatalf("linux first family %q", first)
		}
	}
	if Work.Font != list {
		t.Fatalf("Work.Font %q != FontList %q", Work.Font, list)
	}
}
