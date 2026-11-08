// Package kit is a MyGo component library: Ant Design v5's look,
// drawn on MyGo ui's headless bases.
//
// Layers: ui *Base (0) → Tokens/Page/ConfigProvider (1) → controls (2) →
// Form/Table/Dialog (3) → HomeBar/NavList/WindowButtons (4). See README.md.
//
// Every component is Name(c, required..., ...NameProps); a zero Props is
// the default look. Controls return whether they were clicked or changed,
// containers take the func that builds their children, and overlays with
// a trigger (Popconfirm, Dropdown) take a func that builds a kit button
// and returns its click. Page(c, Work, fn), Page(c, Home, fn) or
// Page(c, WorkDark, fn) picks the chrome; components read it with Of(c).
//
// Beyond Ant's controls (Button, Input, InputNumber, Select, DatePicker,
// TimePicker, RangePicker, Checkbox, Switch, Radio, RadioButtons,
// Segmented, Upload) it has forms with validation (Form, FormItem,
// FormRow, Rule), data display (Table with sorting and pages, Pagination,
// Descriptions, Collapse, Tag, Badge, Progress, Empty), feedback (Alert,
// Result, Spin, SpinOver, Skeleton, Messages), overlays (Dialog, Drawer,
// Confirm, Popconfirm, Tooltip, Dropdown, SplitButton), navigation
// (NavList, NavChips, Tabs, Steps, TopNav, HomeHeader, HomeBar) and
// layout (Space, Divider, GlassCard). Native MyGo is opt-in where it
// beats a CSS clone: ui.Table (Table Native), Element.Menu (Dropdown
// Native, AttachMenu), Dialog.Message / Notification (SetMessageBox,
// SetOSNotify), clipboard, TitleBar insets, Preferences (FollowDesktop),
// DesktopDark (OS appearance without mygo.Theme.IsDark), AutoFocus /
// onPressEnter (ui.AutoFocus, Submitted), and window chrome
// (WindowButtons). ConfigProvider.Dark is Ant's darkAlgorithm
// (DarkAlgorithm); Motion: Bool(false) is token.motion=false. The default
// kit look is painted Ant v5. See README.md for the token sources, each
// component's status and the known gaps.
package kit
