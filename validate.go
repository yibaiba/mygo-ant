package kit

import (
	"fmt"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
)

// NamePath is a Form field's name, as Ant Form.Item name: one segment
// ("user") or a nested path ("user", "name").
type NamePath []string

// NameOf reads a field name: a string ("user" or dotted "user.name"), a
// []string, or a NamePath. The empty name is no path.
func NameOf(name any) NamePath {
	switch v := name.(type) {
	case NamePath:
		return v
	case []string:
		return NamePath(v)
	case string:
		if v == "" {
			return nil
		}
		return NamePath(strings.Split(v, "."))
	}
	return nil
}

// NameKey is name as the Form stores it: segments joined by ".".
func NameKey(name any) string {
	p := NameOf(name)
	if len(p) == 0 {
		return ""
	}
	return strings.Join(p, ".")
}

// Rule checks a field's value and returns the error to show under it, ""
// when it passes. Any function of this shape is a custom rule.
type Rule func(value string) string

// Required fails a value that is empty or only spaces.
func Required(msg string) Rule {
	return func(v string) string { return pick(strings.TrimSpace(v) == "", msg, "") }
}

// Pattern fails a value re does not match; an empty value passes, as
// Ant's pattern leaves emptiness to Required.
func Pattern(re *regexp.Regexp, msg string) Rule {
	return func(v string) string { return pick(v != "" && !re.MatchString(v), msg, "") }
}

// MinLen fails a value of fewer than n characters; an empty value passes.
func MinLen(n int, msg string) Rule {
	return func(v string) string { return pick(v != "" && utf8.RuneCountInString(v) < n, msg, "") }
}

// IPv4 fails a value that is not a dotted IPv4 address; an empty value
// passes.
func IPv4(msg string) Rule {
	return func(v string) string {
		v = strings.TrimSpace(v)
		if v == "" {
			return ""
		}
		a, err := netip.ParseAddr(v)
		return pick(err != nil || !a.Is4(), msg, "")
	}
}

// IntRange fails a value that is not a whole number from min to max; an
// empty value passes.
func IntRange(min, max int, msg string) Rule {
	return func(v string) string {
		v = strings.TrimSpace(v)
		if v == "" {
			return ""
		}
		n, err := strconv.Atoi(v)
		return pick(err != nil || n < min || n > max, msg, "")
	}
}

// CheckRules runs rules in order and returns the first error, "" when
// value passes them all.
func CheckRules(value string, rules ...Rule) string {
	for _, r := range rules {
		if msg := r(value); msg != "" {
			return msg
		}
	}
	return ""
}

// ErrorStatus is the input status that shows err: StatusError when there
// is one.
func ErrorStatus(err string) Status { return pick(err == "", StatusNone, StatusError) }

// ValidateTrigger is when a Form shows a field's error, as Ant
// validateTrigger.
type ValidateTrigger uint8

// Validate triggers.
const (
	// ValidateChange shows an error after the value has been edited, the
	// default (onChange).
	ValidateChange ValidateTrigger = iota
	// ValidateBlur shows it after the field was focused and then left
	// (onBlur). FormItem with Form and Name calls Watch; Check sees the
	// blur the next frame.
	ValidateBlur
	// ValidateSubmit shows it only after Validate (onSubmit).
	ValidateSubmit
)

// Form validates the fields of one form, as Ant's Form: a field shows its
// error once it has been edited, and every field does after Validate.
// Keep it with the form's state; the zero Form is ready.
type Form struct {
	fields    map[string]*formField
	order     []string
	submitted bool
	frame     int
	deps      map[string][]string
	// ValidateTrigger is when Check returns a field's error; onChange
	// when zero. Set it once; it is not cleared by Reset.
	ValidateTrigger ValidateTrigger
}

type formField struct {
	initial string
	touched bool
	err     string
	// value is the value checked in frame, prev the one checked in the
	// frame before; external is an error SetError gave it.
	value, prev string
	frame       int
	external    string
	// unseen marks a field SetError made before it was checked.
	unseen bool
	// focused is the last Watch; blurred is set as it goes false.
	focused, blurred bool
}

// Begin starts a frame of the form, before its fields are checked: the
// fields Validate looks at are those checked since.
func (f *Form) Begin() {
	f.order = f.order[:0]
	f.frame++
}

// Match fails a value that differs from other, as a password check's
// "两次输入的密码不一致"; an empty value passes.
func Match(other, msg string) Rule {
	return func(v string) string { return pick(v != "" && v != other, msg, "") }
}

// Depends makes name re-check, and show its error, when one of deps
// changes after name was edited or given a value (Ant's dependencies).
// Check deps before name in the frame to see a change in the same frame.
// name and deps are a string or []string path (see NameOf).
func (f *Form) Depends(name any, deps ...any) {
	if f.deps == nil {
		f.deps = map[string][]string{}
	}
	keys := make([]string, 0, len(deps))
	for _, d := range deps {
		keys = append(keys, NameKey(d))
	}
	f.deps[NameKey(name)] = keys
}

// Value is the value name had when last checked. name is a string or
// []string path (see NameOf).
func (f *Form) Value(name any) string {
	if st, ok := f.fields[NameKey(name)]; ok {
		return st.value
	}
	return ""
}

// Changed reports whether name's value changed since the frame before,
// as Ant's shouldUpdate comparing the previous and current values.
func (f *Form) Changed(name any) bool {
	st, ok := f.fields[NameKey(name)]
	return ok && st.value != st.prev
}

// Touched reports whether name has been edited.
func (f *Form) Touched(name any) bool {
	st, ok := f.fields[NameKey(name)]
	return ok && st.touched
}

// SetError shows msg under name, as Ant's setFields with errors for a
// server's answer, until its value changes; "" clears it.
func (f *Form) SetError(name any, msg string) {
	if f.fields == nil {
		f.fields = map[string]*formField{}
	}
	key := NameKey(name)
	st, ok := f.fields[key]
	if !ok {
		st = &formField{unseen: true}
		f.fields[key] = st
	}
	st.external = msg
}

// Values is the checked fields as a nested map: "user.name" becomes
// {"user":{"name": "…"}}. Sibling keys share a map.
func (f *Form) Values() map[string]any {
	out := map[string]any{}
	for name, st := range f.fields {
		if name == "" {
			continue
		}
		nestSet(out, strings.Split(name, "."), st.value)
	}
	return out
}

// SetValues writes a nested map (or flat dotted keys) into the fields,
// as Ant's setFieldsValue: it does not touch them, and it clears a
// SetError on each key it sets.
func (f *Form) SetValues(v map[string]any) {
	if v == nil {
		return
	}
	if f.fields == nil {
		f.fields = map[string]*formField{}
	}
	flat := map[string]string{}
	flattenValues("", v, flat)
	for key, val := range flat {
		st, ok := f.fields[key]
		if !ok {
			st = &formField{}
			f.fields[key] = st
		}
		st.initial, st.value, st.prev, st.external, st.unseen = val, val, val, "", false
	}
}

func nestSet(m map[string]any, path []string, v string) {
	if len(path) == 1 {
		m[path[0]] = v
		return
	}
	child, _ := m[path[0]].(map[string]any)
	if child == nil {
		child = map[string]any{}
		m[path[0]] = child
	}
	nestSet(child, path[1:], v)
}

func flattenValues(prefix string, v any, out map[string]string) {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flattenValues(key, child, out)
		}
	case nil:
		return
	default:
		if prefix != "" {
			out[prefix] = fmt.Sprint(t)
		}
	}
}

// Watch records whether name has the keyboard focus this frame, for
// ValidateBlur. FormItem with Form and Name calls it. It does nothing
// until Check has created the field, so it does not count as an edit.
// name is a string or []string path (see NameOf).
func (f *Form) Watch(name any, focused bool) {
	st, ok := f.fields[NameKey(name)]
	if !ok {
		return
	}
	if st.focused && !focused {
		st.blurred = true
	}
	st.focused = focused
}

// Check checks the field name holding value against rules and returns the
// error to show under it now: "" until the field has been edited or the
// form validated. name is a string ("user" or "user.name") or []string
// {"user","name"}; old string names keep working.
func (f *Form) Check(name any, value string, rules ...Rule) string {
	if f.fields == nil {
		f.fields = map[string]*formField{}
	}
	key := NameKey(name)
	st, ok := f.fields[key]
	if !ok {
		st = &formField{initial: value, value: value, prev: value, frame: f.frame}
		f.fields[key] = st
	} else if st.unseen {
		st.initial, st.value, st.prev, st.frame, st.unseen = value, value, value, f.frame, false
	}
	if st.frame != f.frame {
		st.prev, st.frame = st.value, f.frame
	}
	if value != st.value && st.external != "" {
		st.external = ""
	}
	st.value = value
	st.touched = st.touched || value != st.initial
	for _, d := range f.deps[key] {
		if f.Changed(d) && (st.touched || value != "") {
			st.touched = true
		}
	}
	st.err = CheckRules(value, rules...)
	if st.err == "" {
		st.err = st.external
	}
	f.order = append(f.order, key)
	show := st.external != "" || f.submitted
	switch f.ValidateTrigger {
	case ValidateSubmit:
	case ValidateBlur:
		show = show || st.blurred
	default:
		show = show || st.touched
	}
	if show {
		return st.err
	}
	return ""
}

// Validate shows every field's error from now on and returns the first
// error of the fields checked since Begin, "" when the form passes. Call
// it on submit, after the fields are checked in the frame.
func (f *Form) Validate(c *ui.Context) string {
	err, shown := f.validate()
	if err != "" && shown {
		c.Invalidate()
	}
	return err
}

// validate is Validate without the frame: it also reports whether it
// newly shows the errors.
func (f *Form) validate() (string, bool) {
	shown := !f.submitted
	f.submitted = true
	for _, name := range f.order {
		if err := f.fields[name].err; err != "" {
			return err, shown
		}
	}
	return "", shown
}

// Reset hides the errors again, as after a successful save: the values
// now become the fields' untouched values.
func (f *Form) Reset() {
	f.fields, f.submitted = nil, false
}
