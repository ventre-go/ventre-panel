package ui

import (
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
	fynecontainer "fyne.io/fyne/v2/container"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/ventre-go/ventre-panel/internal/config"
	"github.com/ventre-go/ventre-panel/internal/i18n"
	"github.com/ventre-go/ventre-panel/internal/session"
)

func newTestState() *AppState {
	return &AppState{
		mgr:            session.NewManager(),
		opts:           config.Default(),
		locale:         i18n.NewLocale(),
		selectedResult: -1,
		currentOp:      session.OpCommand,
	}
}

func walkCanvasObjects(obj fyne.CanvasObject, visit func(fyne.CanvasObject)) {
	if obj == nil {
		return
	}
	visit(obj)
	if c, ok := obj.(*fyne.Container); ok {
		for _, child := range c.Objects {
			walkCanvasObjects(child, visit)
		}
	}
	if s, ok := obj.(*fynecontainer.Split); ok {
		walkCanvasObjects(s.Leading, visit)
		walkCanvasObjects(s.Trailing, visit)
	}
	if tabs, ok := obj.(*fynecontainer.AppTabs); ok {
		for _, item := range tabs.Items {
			walkCanvasObjects(item.Content, visit)
		}
	}
	if card, ok := obj.(*widget.Card); ok {
		walkCanvasObjects(card.Content, visit)
	}
}

func findButtonByText(root fyne.CanvasObject, text string) *widget.Button {
	var found *widget.Button
	walkCanvasObjects(root, func(obj fyne.CanvasObject) {
		if found != nil {
			return
		}
		button, ok := obj.(*widget.Button)
		if ok && button.Text == text {
			found = button
		}
	})
	return found
}

func containsLabelText(root fyne.CanvasObject, text string) bool {
	found := false
	walkCanvasObjects(root, func(obj fyne.CanvasObject) {
		if found {
			return
		}
		label, ok := obj.(*widget.Label)
		if ok && label.Text == text {
			found = true
		}
	})
	return found
}

func firstTable(root fyne.CanvasObject) *widget.Table {
	var found *widget.Table
	walkCanvasObjects(root, func(obj fyne.CanvasObject) {
		if found != nil {
			return
		}
		table, ok := obj.(*widget.Table)
		if ok {
			found = table
		}
	})
	return found
}

func containsTopLeadingLayout(root fyne.CanvasObject) bool {
	found := false
	walkCanvasObjects(root, func(obj fyne.CanvasObject) {
		if found {
			return
		}
		c, ok := obj.(*fyne.Container)
		if !ok {
			return
		}
		_, found = c.Layout.(*topLeadingLayout)
	})
	return found
}

func firstOperationTabs(root fyne.CanvasObject) *fynecontainer.AppTabs {
	var found *fynecontainer.AppTabs
	walkCanvasObjects(root, func(obj fyne.CanvasObject) {
		if found != nil {
			return
		}
		tabs, ok := obj.(*fynecontainer.AppTabs)
		if ok && len(tabs.Items) == 3 {
			found = tabs
		}
	})
	return found
}

func setManagerResultsForTest(t *testing.T, mgr *session.Manager, results []session.HostResult) {
	t.Helper()
	field := reflect.ValueOf(mgr).Elem().FieldByName("results")
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(reflect.ValueOf(results))
}

func TestSetLocaleFromSelectionIgnoresCurrentLanguage(t *testing.T) {
	state := &AppState{
		mgr:    session.NewManager(),
		opts:   config.Default(),
		locale: i18n.NewLocale(),
	}

	if changed := setLocaleFromSelection(state, languageLabelZH); changed {
		t.Fatal("expected selecting the current language to be ignored")
	}
	if got := state.locale.Current(); got != i18n.ZH {
		t.Fatalf("expected locale to remain zh-CN, got %s", got)
	}
}

func TestSetLocaleFromSelectionChangesLanguage(t *testing.T) {
	state := &AppState{
		mgr:    session.NewManager(),
		opts:   config.Default(),
		locale: i18n.NewLocale(),
	}

	if changed := setLocaleFromSelection(state, languageLabelEN); !changed {
		t.Fatal("expected selecting a different language to report a change")
	}
	if got := state.locale.Current(); got != i18n.EN {
		t.Fatalf("expected locale to change to en-US, got %s", got)
	}
}

func TestConfigureMainWindowUsesWorkbenchSizeWithoutFullscreenOrFixedSize(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("test")

	configureMainWindow(w)

	policy, ok := w.(interface {
		FullScreen() bool
		FixedSize() bool
	})
	if !ok {
		t.Fatal("test window does not expose fullscreen/fixed-size state")
	}
	if policy.FullScreen() {
		t.Fatal("expected main window not to open fullscreen")
	}
	if policy.FixedSize() {
		t.Fatal("expected main window to remain resizable")
	}

	size := w.Canvas().Size()
	if size.Width != 1440 || size.Height != 900 {
		t.Fatalf("expected default workbench size 1440x900, got %v", size)
	}
}

func TestDisplayOperationUsesLocalizedLabel(t *testing.T) {
	loc := i18n.NewLocale()

	if got := displayOperation(loc, session.OpTestConnection); got != "连接测试" {
		t.Fatalf("expected zh-CN test connection label, got %q", got)
	}

	loc.Set(i18n.EN)
	if got := displayOperation(loc, session.OpDownload); got != "File Download" {
		t.Fatalf("expected en-US download label, got %q", got)
	}
}

func TestHostCounts(t *testing.T) {
	total, enabled := hostCounts([]session.HostRow{
		{Enabled: true},
		{Enabled: false},
		{Enabled: true},
	})

	if total != 3 || enabled != 2 {
		t.Fatalf("expected total=3 enabled=2, got total=%d enabled=%d", total, enabled)
	}
}

func TestHostTableHeadersDoNotIncludePassword(t *testing.T) {
	loc := i18n.NewLocale()
	headers := strings.Join(hostTableHeaders(loc), "|")
	if strings.Contains(headers, loc.T("hosts.pass_placeholder")) {
		t.Fatalf("host table headers should not include password column: %q", headers)
	}
	if len(hostTableHeaders(loc)) != 5 {
		t.Fatalf("expected host table to have 5 columns, got %d", len(hostTableHeaders(loc)))
	}
}

func TestHostTableColumnWidthsAreReadable(t *testing.T) {
	widths := hostTableColumnWidths()
	want := []float32{60, 190, 80, 150, 150}
	if len(widths) != len(want) {
		t.Fatalf("expected %d column widths, got %d", len(want), len(widths))
	}
	for i := range want {
		if widths[i] != want[i] {
			t.Fatalf("column %d width = %.0f, want %.0f", i, widths[i], want[i])
		}
	}
}

func TestHostsPanelEmptyStateDoesNotRenderTable(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("hosts")
	state := newTestState()
	refs := &panelRefreshers{}

	panel := buildHostsPanel(w, state, refs)
	refs.hosts()

	if table := firstTable(panel); table != nil {
		t.Fatal("empty hosts panel must render standalone empty state instead of a normal table")
	}
	if !containsLabelText(panel, state.T("hosts.empty")) {
		t.Fatal("empty hosts panel did not show localized empty-state text")
	}
}

func TestHostsPanelRendersReadableTableOnlyWithHosts(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("hosts")
	state := newTestState()
	state.mgr.SetHosts([]session.HostRow{{
		Target:  session.HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "secret-pass"},
		Enabled: true,
	}})
	refs := &panelRefreshers{}

	panel := buildHostsPanel(w, state, refs)
	refs.hosts()

	table := firstTable(panel)
	if table == nil {
		t.Fatal("hosts panel with rows must render the host table")
	}
	rows, cols := table.Length()
	if rows != 2 || cols != 5 {
		t.Fatalf("expected table dimensions 2x5, got %dx%d", rows, cols)
	}
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			cell := table.CreateCell()
			table.UpdateCell(widget.TableCellID{Row: row, Col: col}, cell)
			label, ok := cell.(*widget.Label)
			if !ok {
				t.Fatalf("expected table cell label, got %T", cell)
			}
			if strings.Contains(label.Text, "secret-pass") {
				t.Fatalf("host table leaked password in cell %d,%d: %q", row, col, label.Text)
			}
		}
	}
}

func TestBuildHostRowFromInputDefaultsPort(t *testing.T) {
	row, err := buildHostRowFromInput("10.0.0.1", "", "root", "secret")
	if err != nil {
		t.Fatalf("expected valid row, got error: %v", err)
	}
	if row.Target.Port != 22 {
		t.Fatalf("expected default port 22, got %d", row.Target.Port)
	}
	if !row.Enabled {
		t.Fatal("expected added host to be enabled")
	}
}

func TestBuildHostRowFromInputRejectsInvalidPort(t *testing.T) {
	_, err := buildHostRowFromInput("10.0.0.1", "ssh", "root", "secret")
	if err == nil {
		t.Fatal("expected invalid port error")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("error leaked password: %q", err.Error())
	}
}

func TestBuildHostRowFromInputValidationMessagesAreLocalized(t *testing.T) {
	loc := i18n.NewLocale()
	_, err := buildHostRowFromInput("", "22", "root", "secret")
	if err == nil {
		t.Fatal("expected missing IP error")
	}
	if got := localizedErrorText(loc, err); got != "IP 不能为空" {
		t.Fatalf("expected zh-CN validation message, got %q", got)
	}

	loc.Set(i18n.EN)
	if got := localizedErrorText(loc, err); got != "IP is required" {
		t.Fatalf("expected en-US validation message, got %q", got)
	}
}

func TestBuildHostRowFromInputRejectsRequiredFields(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		port     string
		username string
		password string
		want     string
	}{
		{name: "empty IP", ip: "", port: "22", username: "root", password: "secret", want: "IP must not be empty"},
		{name: "empty username", ip: "10.0.0.1", port: "22", username: "", password: "secret", want: "username must not be empty"},
		{name: "empty password", ip: "10.0.0.1", port: "22", username: "root", password: "", want: "password must not be empty"},
		{name: "port out of range", ip: "10.0.0.1", port: "65536", username: "root", password: "secret", want: "port must be between 1 and 65535"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildHostRowFromInput(tt.ip, tt.port, tt.username, tt.password)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %q", tt.want, err.Error())
			}
			if strings.Contains(err.Error(), tt.password) && tt.password != "" {
				t.Fatalf("validation error leaked password: %q", err.Error())
			}
		})
	}
}

func TestBuildHostRowsFromCSVRejectsBatchWithoutPartialRows(t *testing.T) {
	rows, err := buildHostRowsFromCSV("10.0.0.1,22,root,secret\n10.0.0.2,nope,root,secret2")
	if err == nil {
		t.Fatal("expected CSV parse error")
	}
	if rows != nil {
		t.Fatalf("expected no rows on failed CSV import, got %#v", rows)
	}
	if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "secret2") {
		t.Fatalf("CSV error leaked password: %q", err.Error())
	}
}

func TestBuildHostRowsFromCSVValidationMessagesAreLocalizedAndRedacted(t *testing.T) {
	loc := i18n.NewLocale()
	_, err := buildHostRowsFromCSV("10.0.0.1,22,root,secret\n10.0.0.2,nope,root,secret2")
	if err == nil {
		t.Fatal("expected CSV parse error")
	}

	zh := localizedErrorText(loc, err)
	if !strings.Contains(zh, "第 2 行") || !strings.Contains(zh, "必须是数字") {
		t.Fatalf("expected localized zh-CN CSV line error, got %q", zh)
	}
	if strings.Contains(zh, "secret") || strings.Contains(zh, "secret2") {
		t.Fatalf("localized CSV error leaked password: %q", zh)
	}

	loc.Set(i18n.EN)
	en := localizedErrorText(loc, err)
	if !strings.Contains(en, "Line 2") || !strings.Contains(en, "must be a number") {
		t.Fatalf("expected localized en-US CSV line error, got %q", en)
	}
}

func TestBuildHostRowsFromCSVParsesValidRows(t *testing.T) {
	rows, err := buildHostRowsFromCSV("10.0.0.1,22,root,secret\n10.0.0.2,,ubuntu,secret2")
	if err != nil {
		t.Fatalf("expected valid CSV, got error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[1].Target.Port != 22 {
		t.Fatalf("expected empty CSV port to default to 22, got %d", rows[1].Target.Port)
	}
}

func TestSelectedHostsCountMatchesEnabledHosts(t *testing.T) {
	_, enabled := hostCounts([]session.HostRow{
		{Enabled: true},
		{Enabled: false},
		{Enabled: true},
	})
	if enabled != 2 {
		t.Fatalf("expected selected/enabled host count 2, got %d", enabled)
	}
}

func TestResultCopyRedactsHostPasswords(t *testing.T) {
	loc := i18n.NewLocale()
	hosts := []session.HostRow{
		{Target: session.HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "secret-pass"}},
	}
	result := session.HostResult{
		TargetIP:     "10.0.0.1",
		Port:         22,
		Username:     "root",
		Operation:    session.OpCommand,
		Status:       "ok",
		Success:      true,
		ExitCode:     0,
		Duration:     time.Second,
		Stdout:       "stdout secret-pass",
		Stderr:       "stderr secret-pass",
		ErrorMessage: "error secret-pass",
	}

	text := formatResultForCopy(loc, result, hosts)
	if strings.Contains(text, "secret-pass") {
		t.Fatalf("copied result leaked password: %q", text)
	}
	if !strings.Contains(text, "(redacted)") {
		t.Fatalf("expected copied result to include redaction marker: %q", text)
	}
}

func TestAllResultsCopyRedactsHostPasswords(t *testing.T) {
	loc := i18n.NewLocale()
	hosts := []session.HostRow{
		{Target: session.HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "secret-pass"}},
	}
	results := []session.HostResult{
		{
			TargetIP:     "10.0.0.1",
			Port:         22,
			Username:     "root",
			Operation:    session.OpCommand,
			Status:       "ok",
			Success:      true,
			ExitCode:     0,
			Duration:     time.Second,
			Stdout:       "stdout secret-pass",
			ErrorMessage: "error secret-pass",
		},
		{
			TargetIP:  "10.0.0.1",
			Port:      22,
			Username:  "root",
			Operation: session.OpUpload,
			Status:    "error",
			Stderr:    "stderr secret-pass",
		},
	}

	text := formatAllResultsForCopy(loc, results, hosts)
	if strings.Contains(text, "secret-pass") {
		t.Fatalf("copied all results leaked password: %q", text)
	}
	if strings.Count(text, "(redacted)") < 3 {
		t.Fatalf("expected copied all results to redact every occurrence: %q", text)
	}
}

func TestResultDetailsRedactsHostPasswords(t *testing.T) {
	loc := i18n.NewLocale()
	hosts := []session.HostRow{
		{Target: session.HostTarget{Password: "secret-pass"}},
	}
	result := session.HostResult{
		Stdout:       "stdout secret-pass",
		Stderr:       "stderr secret-pass",
		ErrorKind:    "auth_failed",
		ErrorMessage: "error secret-pass",
	}

	stdout, stderr, errText := resultDetails(loc, result, hosts)
	combined := stdout + stderr + errText
	if strings.Contains(combined, "secret-pass") {
		t.Fatalf("result details leaked password: %q", combined)
	}
}

func TestResultInspectorContentRedactsHostPasswords(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("inspector")
	state := newTestState()
	state.mgr.SetHosts([]session.HostRow{{
		Target:  session.HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "secret-pass"},
		Enabled: true,
	}})
	results := []session.HostResult{{
		TargetIP:     "10.0.0.1",
		Port:         22,
		Username:     "root",
		Operation:    session.OpCommand,
		Status:       "ok",
		Success:      true,
		Stdout:       "stdout secret-pass",
		Stderr:       "stderr secret-pass",
		ErrorMessage: "error secret-pass",
	}}
	setManagerResultsForTest(t, state.mgr, results)
	state.setRunFinished(session.OpCommand, results)
	state.setSelectedResult(0)

	content := buildResultInspectorContent(w, state, nil)
	var text strings.Builder
	walkCanvasObjects(content, func(obj fyne.CanvasObject) {
		switch typed := obj.(type) {
		case *widget.Label:
			text.WriteString(typed.Text)
		case *widget.Entry:
			text.WriteString(typed.Text)
		}
	})
	if strings.Contains(text.String(), "secret-pass") {
		t.Fatalf("result inspector content leaked password: %q", text.String())
	}
}

func TestResultCounts(t *testing.T) {
	success, failure := resultCounts([]session.HostResult{
		{Success: true},
		{Success: false},
		{Success: true},
	})
	if success != 2 || failure != 1 {
		t.Fatalf("expected success=2 failure=1, got success=%d failure=%d", success, failure)
	}
}

func TestNonZeroExitCodeDisplaysAsCommandResultNotTransportError(t *testing.T) {
	loc := i18n.NewLocale()
	result := session.HostResult{
		Operation: session.OpCommand,
		Status:    "exit=42",
		ExitCode:  42,
		Success:   false,
		Stderr:    "fail\n",
	}

	if got := resultExitCode(result); got != "42" {
		t.Fatalf("expected exit code display 42, got %q", got)
	}
	if got := resultErrorKind(loc, result); got != "-" {
		t.Fatalf("expected no transport error kind for non-zero exit, got %q", got)
	}
	_, stderr, errText := resultDetails(loc, result, nil)
	if !strings.Contains(stderr, "fail") {
		t.Fatalf("expected stderr to be preserved, got %q", stderr)
	}
	if errText != loc.T("results.no_error") {
		t.Fatalf("expected no error message for non-zero exit, got %q", errText)
	}
}

func TestStatusSummaryText(t *testing.T) {
	loc := i18n.NewLocale()
	if got := statusSummaryText(loc, runSummary{}, false); got != "尚无结果。" {
		t.Fatalf("unexpected empty status text: %q", got)
	}

	running := statusSummaryText(loc, runSummary{Operation: session.OpCommand, Total: 3, Running: 3}, true)
	if !strings.Contains(running, "正在运行") || !strings.Contains(running, "3") {
		t.Fatalf("unexpected running status text: %q", running)
	}

	loc.Set(i18n.EN)
	done := statusSummaryText(loc, runSummary{Operation: session.OpDownload, Total: 3, Success: 2, Failure: 1}, true)
	if done != "Last: download | total 3 | success 2 | failed 1 | running 0" {
		t.Fatalf("unexpected done status text: %q", done)
	}
}

func TestTopBarUsesGroupedReadableControls(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("toolbar")
	state := newTestState()
	refs := &panelRefreshers{}

	topBar := buildTopBar(w, state, refs)

	for _, key := range []string{"options.run_group", "options.ssh_group", "options.ui_group"} {
		if !containsLabelText(topBar, state.T(key)) {
			t.Fatalf("top toolbar missing group label %q", state.T(key))
		}
	}
	if topBar.MinSize().Height < 100 {
		t.Fatalf("expected stable toolbar min height to be at least 100, got %.0f", topBar.MinSize().Height)
	}
}

func TestTopBarHeightIsStableAcrossWarningAndLanguage(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("toolbar")
	refs := &panelRefreshers{}

	state := newTestState()
	off := buildTopBar(w, state, refs).MinSize()

	state.opts.InsecureIgnoreHostKey = true
	on := buildTopBar(w, state, refs).MinSize()
	if off.Height != on.Height {
		t.Fatalf("toolbar height changed after insecure warning toggle: off=%.0f on=%.0f", off.Height, on.Height)
	}

	state.locale.Set(i18n.EN)
	en := buildTopBar(w, state, refs).MinSize()
	if off.Height != en.Height {
		t.Fatalf("toolbar height changed after language switch: zh=%.0f en=%.0f", off.Height, en.Height)
	}
}

func TestOperationPanelsUseTopLeadingCards(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("operation")
	state := newTestState()
	refs := &panelRefreshers{}
	register := func(func(int)) {}

	panels := []fyne.CanvasObject{
		buildQuickCommandPanel(w, state, refs, register),
		buildUploadPanel(w, state, refs, register),
		buildDownloadPanel(w, state, refs, register),
	}
	for i, panel := range panels {
		if !containsTopLeadingLayout(panel) {
			t.Fatalf("operation panel %d should use a top-leading card layout", i)
		}
	}
}

func TestOperationTabsSelectionMatchesContent(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("operation")
	state := newTestState()
	refs := &panelRefreshers{}

	panel := buildOperationPanel(w, state, refs)
	tabs := firstOperationTabs(panel)
	if tabs == nil {
		t.Fatal("operation tabs not found")
	}

	tabs.SelectIndex(1)
	if state.currentOperation() != session.OpUpload {
		t.Fatalf("expected current operation upload, got %s", state.currentOperation())
	}
	if !containsLabelText(tabs.Selected().Content, state.T("operation.upload_file")) {
		t.Fatal("upload tab did not expose upload content")
	}
	if containsLabelText(tabs.Selected().Content, state.T("operation.download_file")) {
		t.Fatal("upload tab content contains download heading")
	}

	tabs.SelectIndex(2)
	if state.currentOperation() != session.OpDownload {
		t.Fatalf("expected current operation download, got %s", state.currentOperation())
	}
	if !containsLabelText(tabs.Selected().Content, state.T("operation.download_file")) {
		t.Fatal("download tab did not expose download content")
	}
	if containsLabelText(tabs.Selected().Content, state.T("operation.upload_file")) {
		t.Fatal("download tab content contains upload heading")
	}

	tabs.SelectIndex(0)
	if state.currentOperation() != session.OpCommand {
		t.Fatalf("expected current operation command, got %s", state.currentOperation())
	}
	if !containsLabelText(tabs.Selected().Content, state.T("cmd.label")) {
		t.Fatal("quick command tab did not expose command content")
	}
}

func TestOperationTabSelectionSurvivesPanelRebuild(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("operation")
	state := newTestState()
	state.setCurrentOperation(session.OpDownload)
	refs := &panelRefreshers{}

	panel := buildOperationPanel(w, state, refs)
	tabs := firstOperationTabs(panel)
	if tabs == nil {
		t.Fatal("operation tabs not found")
	}
	if tabs.SelectedIndex() != 2 {
		t.Fatalf("expected rebuilt panel to select Download tab, got index %d", tabs.SelectedIndex())
	}
	if !containsLabelText(tabs.Selected().Content, state.T("operation.download_file")) {
		t.Fatal("rebuilt Download tab did not show download content")
	}
}

func TestOperationButtonTapKeepsCurrentOperation(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("operation")
	state := newTestState()
	refs := &panelRefreshers{}
	register := func(func(int)) {}

	uploadPanel := buildUploadPanel(w, state, refs, register)
	uploadBtn := findButtonByText(uploadPanel, state.T("transfer.upload_btn"))
	if uploadBtn == nil {
		t.Fatal("upload button not found")
	}
	uploadBtn.OnTapped()
	if state.currentOperation() != session.OpUpload {
		t.Fatalf("expected upload button to preserve upload operation, got %s", state.currentOperation())
	}

	downloadPanel := buildDownloadPanel(w, state, refs, register)
	downloadBtn := findButtonByText(downloadPanel, state.T("transfer.download_btn"))
	if downloadBtn == nil {
		t.Fatal("download button not found")
	}
	downloadBtn.OnTapped()
	if state.currentOperation() != session.OpDownload {
		t.Fatalf("expected download button to preserve download operation, got %s", state.currentOperation())
	}
}

func TestRefreshStatusBarViewDisablesAndEnablesViewResults(t *testing.T) {
	loc := i18n.NewLocale()
	label := widget.NewLabel("")
	button := widget.NewButton(loc.T("results.view"), nil)

	refreshStatusBarView(loc, runSummary{}, false, label, button)
	if !button.Disabled() {
		t.Fatal("expected View Results to be disabled with no results")
	}
	if label.Text != loc.T("status.no_results") {
		t.Fatalf("unexpected no-results status text: %q", label.Text)
	}

	refreshStatusBarView(loc, runSummary{Operation: session.OpDownload, Total: 3, Success: 2, Failure: 1}, true, label, button)
	if button.Disabled() {
		t.Fatal("expected View Results to be enabled with results")
	}
	if !strings.Contains(label.Text, "运行中 0") {
		t.Fatalf("expected completed status summary to include running 0, got %q", label.Text)
	}
}

func TestStatusBarMinSizeIsStableAcrossResultStates(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("status")
	state := newTestState()
	refs := &panelRefreshers{}

	statusBar := buildStatusBar(w, state, refs)
	noResults := statusBar.MinSize()

	results := []session.HostResult{{TargetIP: "10.0.0.1", Operation: session.OpDownload, Success: true}}
	setManagerResultsForTest(t, state.mgr, results)
	state.setRunFinished(session.OpDownload, results)
	refs.results()
	withResults := statusBar.MinSize()

	if noResults.Height != withResults.Height {
		t.Fatalf("status bar height changed across result states: no=%.0f with=%.0f", noResults.Height, withResults.Height)
	}
}

func TestViewResultsButtonOpensResultInspector(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("status")
	state := newTestState()
	results := []session.HostResult{{
		TargetIP:  "10.0.0.1",
		Port:      22,
		Username:  "root",
		Operation: session.OpDownload,
		Status:    "ok",
		Success:   true,
	}}
	setManagerResultsForTest(t, state.mgr, results)
	state.setRunFinished(session.OpDownload, results)
	refs := &panelRefreshers{}

	statusBar := buildStatusBar(w, state, refs)
	w.SetContent(statusBar)
	refs.results()

	viewButton := findButtonByText(statusBar, state.T("results.view"))
	if viewButton == nil {
		t.Fatal("could not find View Results button")
	}
	if viewButton.Disabled() {
		t.Fatal("View Results button should be enabled when results exist")
	}

	before := len(app.Driver().AllWindows())
	fynetest.Tap(viewButton)
	after := len(app.Driver().AllWindows())
	if after != before+1 {
		t.Fatalf("expected View Results to open Result Inspector window, windows before=%d after=%d", before, after)
	}
}

func TestResultInspectorClosePreservesResults(t *testing.T) {
	app := fynetest.NewTempApp(t)
	w := app.NewWindow("status")
	state := newTestState()
	results := []session.HostResult{{
		TargetIP:  "10.0.0.1",
		Port:      22,
		Username:  "root",
		Operation: session.OpCommand,
		Status:    "ok",
		Success:   true,
	}}
	setManagerResultsForTest(t, state.mgr, results)
	state.setRunFinished(session.OpCommand, results)

	before := len(app.Driver().AllWindows())
	showResultInspector(w, state, nil)
	windows := app.Driver().AllWindows()
	if len(windows) != before+1 {
		t.Fatalf("expected inspector window to open, windows before=%d after=%d", before, len(windows))
	}
	windows[len(windows)-1].Close()
	if got := len(state.mgr.Results()); got != 1 {
		t.Fatalf("closing inspector cleared results, got %d", got)
	}
}

func TestLanguageSwitchPreservesHostsAndRunSummary(t *testing.T) {
	state := &AppState{
		mgr:    session.NewManager(),
		opts:   config.Default(),
		locale: i18n.NewLocale(),
	}
	state.mgr.SetHosts([]session.HostRow{{Target: session.HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "secret"}, Enabled: true}})
	results := []session.HostResult{{TargetIP: "10.0.0.1", Operation: session.OpCommand, Success: true}}
	setManagerResultsForTest(t, state.mgr, results)
	state.setRunFinished(session.OpCommand, results)

	if changed := setLocaleFromSelection(state, languageLabelEN); !changed {
		t.Fatal("expected locale switch")
	}

	if got := len(state.mgr.Hosts()); got != 1 {
		t.Fatalf("expected hosts to be preserved, got %d", got)
	}
	summary := state.runSummary()
	if summary.Operation != session.OpCommand || summary.Success != 1 {
		t.Fatalf("expected run summary to be preserved, got %#v", summary)
	}
	if got := len(state.mgr.Results()); got != 1 {
		t.Fatalf("expected results to be preserved, got %d", got)
	}
}

func TestLanguageSwitchDoesNotPersistToNewState(t *testing.T) {
	state := &AppState{
		mgr:    session.NewManager(),
		opts:   config.Default(),
		locale: i18n.NewLocale(),
	}
	if changed := setLocaleFromSelection(state, languageLabelEN); !changed {
		t.Fatal("expected locale switch")
	}

	newState := &AppState{
		mgr:    session.NewManager(),
		opts:   config.Default(),
		locale: i18n.NewLocale(),
	}
	if got := newState.locale.Current(); got != i18n.ZH {
		t.Fatalf("expected fresh state to use default zh-CN, got %s", got)
	}
}

func TestLanguageSwitchPreservesCurrentOperation(t *testing.T) {
	state := newTestState()
	state.setCurrentOperation(session.OpDownload)
	if changed := setLocaleFromSelection(state, languageLabelEN); !changed {
		t.Fatal("expected locale switch")
	}
	if got := state.currentOperation(); got != session.OpDownload {
		t.Fatalf("expected language switch to preserve current operation, got %s", got)
	}
}
