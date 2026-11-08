# mygo-ant

Ant Design v5 look for [MyGo](https://github.com/egoist/mygo) native UI. Package `kit` paints on `ui` headless bases (`ButtonBase`, `SelectBase`, `TextInputBase`, …). It depends only on `mygo/ui` and the Go standard library.

Uses official `github.com/egoist/mygo` v0.8.3. On LoongArch ABI 1.0 (or if you need the Linux GTK DRI/font fixes), you may optionally:

```
replace github.com/egoist/mygo => github.com/yibaiba/mygo v0.8.3-loong64.1
```

```go
import "github.com/yibaiba/mygo-ant"

kit.Page(c, kit.Work, func() {
	if kit.Button(c, "Save", kit.ButtonProps{Type: kit.Primary}) {
		save()
	}
})
```

`Page(c, kit.Work, fn)`, `Page(c, kit.Home, fn)` or `Page(c, kit.WorkDark, fn)` picks chrome. Components read tokens with `kit.Of(c)`.

See [doc.go](doc.go) for the component list. MIT license.
