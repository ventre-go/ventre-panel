package ui

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/ventre-go/ventre-panel/internal/config"
	"github.com/ventre-go/ventre-panel/internal/i18n"
	"github.com/ventre-go/ventre-panel/internal/secure"
	"github.com/ventre-go/ventre-panel/internal/session"
	"github.com/ventre-go/ventre-panel/internal/validation"
)

// AppState holds the runtime state for the UI.
type AppState struct {
	mu     sync.Mutex
	mgr    *session.Manager
	opts   config.RuntimeOptions
	locale *i18n.Locale

	selectedResult int
	summary        runSummary
	currentOp      session.Operation
	splitOffset    float64
}

func (s *AppState) getSplitOffset() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.splitOffset <= 0 {
		return defaultSplitOffset
	}
	return s.splitOffset
}

func (s *AppState) setSplitOffset(v float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.splitOffset = v
}

type runSummary struct {
	Operation session.Operation
	Total     int
	Success   int
	Failure   int
	Running   int
}

type inputValidationError struct {
	key      string
	fallback string
	args     []any
}

func (e *inputValidationError) Error() string {
	return formatValidationMessage(e.fallback, e.args...)
}

type csvImportError struct {
	errs []error
}

func (e *csvImportError) Error() string {
	var messages []string
	for _, err := range e.errs {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "\n")
}

type panelRefreshers struct {
	hosts     func()
	operation func()
	results   func()
}

type minSizeLayout struct {
	min fyne.Size
}

type topLeadingLayout struct {
	maxWidth float32
}

type stableSizeLayout struct {
	min fyne.Size
}

type fixedHeightLayout struct {
	height float32
}

var (
	defaultWindowSize       = fyne.NewSize(1280, 760)
	minimumDashboardSize    = fyne.NewSize(1120, 680)
	hostsPanelMinSize       = fyne.NewSize(430, 0)
	resultInspectorSize     = fyne.NewSize(1200, 780)
	resultInspectorMinSize  = fyne.NewSize(960, 640)
	toolbarMinSize          = fyne.NewSize(0, 104)
	toolbarWarningHeight    = float32(22)
	operationContentMin     = fyne.NewSize(760, 430)
	operationCardMaxWidth   = float32(740)
	operationFieldWidth     = float32(680)
	operationBrowseWidth    = float32(110)
	operationBrowseInput    = float32(560)
	operationButtonMinSize  = fyne.NewSize(260, 42)
	viewResultsButtonSize   = fyne.NewSize(150, 40)
	statusBarMinSize        = fyne.NewSize(0, 56)
	knownHostsControlSize   = fyne.NewSize(400, 38)
	toolbarSelectSize       = fyne.NewSize(110, 38)
	languageSelectSize      = fyne.NewSize(180, 38)
	insecureCheckSize       = fyne.NewSize(260, 38)
	hostActionButtonSize    = fyne.NewSize(188, 38)
	hostWideButtonSize      = fyne.NewSize(388, 38)
	hostsEmptyStateSize     = fyne.NewSize(390, 86)
	hostTableColumnDefaults = []float32{56, 170, 72, 130, 140}
	defaultSplitOffset      = float64(0.39)
)

// currentMainSplit tracks the active HSplit for offset preservation across rebuilds.
var currentMainSplit *container.Split

func (l *minSizeLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	min := l.min
	for _, obj := range objects {
		if obj.Visible() {
			min = min.Max(obj.MinSize())
		}
	}
	return min
}

func (l *minSizeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, obj := range objects {
		if obj.Visible() {
			obj.Resize(size)
		}
	}
}

func (l *topLeadingLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var min fyne.Size
	for _, obj := range objects {
		if !obj.Visible() {
			continue
		}
		childMin := obj.MinSize()
		if l.maxWidth > 0 && childMin.Width > l.maxWidth {
			childMin.Width = l.maxWidth
		}
		min = min.Max(childMin)
	}
	return min
}

func (l *topLeadingLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	width := size.Width
	if l.maxWidth > 0 && width > l.maxWidth {
		width = l.maxWidth
	}
	for _, obj := range objects {
		if !obj.Visible() {
			continue
		}
		height := obj.MinSize().Height
		if height > size.Height {
			height = size.Height
		}
		obj.Move(fyne.NewPos(0, 0))
		obj.Resize(fyne.NewSize(width, height))
	}
}

func (l *stableSizeLayout) MinSize([]fyne.CanvasObject) fyne.Size {
	return l.min
}

func (l *stableSizeLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, obj := range objects {
		if obj.Visible() {
			obj.Resize(size)
		}
	}
}

func (l *fixedHeightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var width float32
	for _, obj := range objects {
		if obj.Visible() && obj.MinSize().Width > width {
			width = obj.MinSize().Width
		}
	}
	return fyne.NewSize(width, l.height)
}

func (l *fixedHeightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, obj := range objects {
		if obj.Visible() {
			obj.Resize(fyne.NewSize(size.Width, l.height))
		}
	}
}

func (r *panelRefreshers) refreshAll() {
	if r.hosts != nil {
		r.hosts()
	}
	if r.operation != nil {
		r.operation()
	}
	if r.results != nil {
		r.results()
	}
}

func (s *AppState) T(key string) string { return s.locale.T(key) }

func (s *AppState) setRunStarted(op session.Operation, total int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.summary = runSummary{Operation: op, Total: total, Running: total}
}

func (s *AppState) setRunFinished(op session.Operation, results []session.HostResult) {
	success, failure := resultCounts(results)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.summary = runSummary{
		Operation: op,
		Total:     len(results),
		Success:   success,
		Failure:   failure,
	}
}

func (s *AppState) clearRunSummary() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.summary = runSummary{}
	s.selectedResult = -1
}

func (s *AppState) runSummary() runSummary {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.summary
}

func (s *AppState) setSelectedResult(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.selectedResult = index
}

func (s *AppState) selectedResultIndex() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.selectedResult
}

func (s *AppState) setCurrentOperation(op session.Operation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentOp = op
}

func (s *AppState) currentOperation() session.Operation {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentOp == "" {
		return session.OpCommand
	}
	return s.currentOp
}

// BuildWindow constructs the main application window.
func BuildWindow(w fyne.Window) {
	state := &AppState{
		mgr:            session.NewManager(),
		opts:           config.Default(),
		locale:         i18n.NewLocale(),
		selectedResult: -1,
		currentOp:      session.OpCommand,
	}

	w.SetTitle(state.T("title"))
	w.SetPadded(true)
	setMainContent(w, state)
	configureMainWindow(w)
}

func configureMainWindow(w fyne.Window) {
	w.Resize(defaultWindowSize)
}

func setMainContent(w fyne.Window, state *AppState) {
	refs := &panelRefreshers{}
	content := container.New(&minSizeLayout{min: minimumDashboardSize}, buildDashboard(w, state, refs))
	w.SetContent(content)
	refs.refreshAll()
}

func buildDashboard(w fyne.Window, state *AppState, refs *panelRefreshers) fyne.CanvasObject {
	topBar := buildTopBar(w, state, refs)
	hostsPanel := buildHostsPanel(w, state, refs)
	operationPanel := buildOperationPanel(w, state, refs)
	statusBar := buildStatusBar(w, state, refs)

	mainSplit := container.NewHSplit(hostsPanel, operationPanel)
	mainSplit.Offset = state.getSplitOffset()
	currentMainSplit = mainSplit

	return container.NewBorder(topBar, statusBar, nil, nil, mainSplit)
}

// ---------- Shared helpers ----------

func displayErrorKind(locale *i18n.Locale, kind string) string {
	key := "err." + kind
	t := locale.T(key)
	if t != key {
		return t
	}
	return kind
}

func displayOperation(locale *i18n.Locale, op session.Operation) string {
	if op == "" {
		return locale.T("op.none")
	}
	key := "op." + string(op)
	t := locale.T(key)
	if t != key {
		return t
	}
	return string(op)
}

func configureSessionOptions(state *AppState) config.RuntimeOptions {
	state.mu.Lock()
	opts := state.opts
	state.mu.Unlock()

	state.mgr.SetOptions(session.RunOptions{
		Timeout:               time.Duration(opts.Timeout) * time.Second,
		Concurrency:           opts.Concurrency,
		KnownHostsPath:        opts.KnownHostsPath,
		InsecureIgnoreHostKey: opts.InsecureIgnoreHostKey,
	})
	return opts
}

func contextWithRuntimeTimeout(parent context.Context, opts config.RuntimeOptions) (context.Context, context.CancelFunc) {
	if opts.Timeout <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, time.Duration(opts.Timeout)*time.Second)
}

func fieldBlock(label string, field fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel(label), field)
}

func fixedSize(size fyne.Size, obj fyne.CanvasObject) fyne.CanvasObject {
	return container.NewGridWrap(size, obj)
}

func fixedWidth(width float32, obj fyne.CanvasObject) fyne.CanvasObject {
	return fixedSize(fyne.NewSize(width, obj.MinSize().Height), obj)
}

func toolbarGroupLabel(text string) *widget.Label {
	return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

func operationCard(content fyne.CanvasObject) fyne.CanvasObject {
	card := widget.NewCard("", "", container.NewPadded(content))
	return container.NewPadded(container.New(&topLeadingLayout{maxWidth: operationCardMaxWidth}, card))
}

func operationModePanel(content fyne.CanvasObject) fyne.CanvasObject {
	return container.New(&stableSizeLayout{min: operationContentMin}, operationCard(content))
}

func browseFieldRow(entry fyne.CanvasObject, browse fyne.CanvasObject) fyne.CanvasObject {
	return container.NewHBox(
		fixedWidth(operationBrowseInput, entry),
		fixedSize(fyne.NewSize(operationBrowseWidth, browse.MinSize().Height), browse),
	)
}

func hSpacer(width float32) fyne.CanvasObject {
	return fixedSize(fyne.NewSize(width, 1), widget.NewLabel(""))
}

func formatValidationMessage(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

func newInputValidationError(key, fallback string, args ...any) error {
	return &inputValidationError{key: key, fallback: fallback, args: args}
}

func localizedErrorText(locale *i18n.Locale, err error) string {
	if err == nil {
		return ""
	}

	var csvErr *csvImportError
	if errors.As(err, &csvErr) {
		var messages []string
		for _, item := range csvErr.errs {
			messages = append(messages, localizedErrorText(locale, item))
		}
		return strings.Join(messages, "\n")
	}

	var inputErr *inputValidationError
	if errors.As(err, &inputErr) {
		return formatValidationMessage(locale.T(inputErr.key), inputErr.args...)
	}

	var hostCSVErr validation.CSVError
	if errors.As(err, &hostCSVErr) {
		switch hostCSVErr.Code {
		case validation.CSVErrFields:
			return formatValidationMessage(locale.T("validation.csv_fields"), hostCSVErr.Line, hostCSVErr.Got)
		case validation.CSVErrIPRequired:
			return formatValidationMessage(locale.T("validation.csv_ip_required"), hostCSVErr.Line)
		case validation.CSVErrPortNumber:
			return formatValidationMessage(locale.T("validation.csv_port_number"), hostCSVErr.Line, hostCSVErr.Value)
		case validation.CSVErrPortRange:
			return formatValidationMessage(locale.T("validation.csv_port_range"), hostCSVErr.Line, hostCSVErr.Port)
		case validation.CSVErrUsernameRequired:
			return formatValidationMessage(locale.T("validation.csv_username_required"), hostCSVErr.Line)
		case validation.CSVErrPasswordRequired:
			return formatValidationMessage(locale.T("validation.csv_password_required"), hostCSVErr.Line)
		}
	}

	switch err.Error() {
	case "local file path must not be empty":
		return locale.T("validation.local_file_required")
	case "local file does not exist":
		return locale.T("validation.local_file_not_found")
	case "directory transfer is not supported; please select a single file":
		return locale.T("validation.directory_upload_unsupported")
	case "remote file path must not be empty":
		return locale.T("validation.remote_file_required")
	case "directory transfer is not supported; please specify a single file path":
		return locale.T("validation.directory_remote_unsupported")
	case "local directory path must not be empty":
		return locale.T("validation.local_dir_required")
	case "local output path must be a directory":
		return locale.T("validation.local_dir_expected")
	default:
		return err.Error()
	}
}

func showLocalizedError(w fyne.Window, locale *i18n.Locale, err error) {
	dialog.ShowError(fmt.Errorf("%s", localizedErrorText(locale, err)), w)
}

func wrappedLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	return label
}

func heading(text string) *widget.Label {
	return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

func hostTableHeaders(locale *i18n.Locale) []string {
	return []string{
		locale.T("hosts.enabled"),
		"IP",
		locale.T("hosts.port_placeholder"),
		locale.T("hosts.user_placeholder"),
		locale.T("hosts.status"),
	}
}

func hostTableColumnWidths() []float32 {
	out := make([]float32, len(hostTableColumnDefaults))
	copy(out, hostTableColumnDefaults)
	return out
}

func applyHostTableColumnWidths(table *widget.Table) {
	for i, width := range hostTableColumnWidths() {
		table.SetColumnWidth(i, width)
	}
}

func hostStatusText(locale *i18n.Locale, row session.HostRow) string {
	if !row.Enabled {
		return locale.T("hosts.status_disabled")
	}
	status := strings.TrimSpace(row.Status)
	switch status {
	case "":
		return locale.T("hosts.status_ready")
	case "ok":
		return locale.T("hosts.status_ready")
	case locale.T("hosts.testing"):
		return locale.T("hosts.status_testing")
	default:
		return locale.T("hosts.status_failed")
	}
}

func optionStrings(values []int) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, strconv.Itoa(value))
	}
	return out
}

func hostCounts(hosts []session.HostRow) (total int, enabled int) {
	for _, host := range hosts {
		total++
		if host.Enabled {
			enabled++
		}
	}
	return total, enabled
}

func resultCounts(results []session.HostResult) (success int, failure int) {
	for _, result := range results {
		if result.Success {
			success++
			continue
		}
		failure++
	}
	return success, failure
}

func enabledHostCount(state *AppState) int {
	return len(state.mgr.EnabledHosts())
}

func hostActionLabel(locale *i18n.Locale, zeroKey, oneKey, manyKey string, count int) string {
	switch count {
	case 0:
		return locale.T(zeroKey)
	case 1:
		return locale.T(oneKey)
	default:
		return formatValidationMessage(locale.T(manyKey), count)
	}
}

func commandRunLabel(locale *i18n.Locale, count int) string {
	return hostActionLabel(locale, "cmd.run", "cmd.run_one", "cmd.run_many", count)
}

func uploadRunLabel(locale *i18n.Locale, count int) string {
	return hostActionLabel(locale, "transfer.upload_btn", "transfer.upload_btn_one", "transfer.upload_btn_many", count)
}

func downloadRunLabel(locale *i18n.Locale, count int) string {
	return hostActionLabel(locale, "transfer.download_btn", "transfer.download_btn_one", "transfer.download_btn_many", count)
}

func buildHostRowFromInput(ip, portText, username, password string) (session.HostRow, error) {
	ip = strings.TrimSpace(ip)
	portText = strings.TrimSpace(portText)
	username = strings.TrimSpace(username)
	passwordForValidation := strings.TrimSpace(password)
	if portText == "" {
		portText = "22"
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return session.HostRow{}, newInputValidationError("validation.port_number", "port must be a number")
	}
	switch {
	case ip == "":
		return session.HostRow{}, newInputValidationError("validation.ip_required", "IP must not be empty")
	case port < 1 || port > 65535:
		return session.HostRow{}, newInputValidationError("validation.port_range", "port must be between 1 and 65535")
	case username == "":
		return session.HostRow{}, newInputValidationError("validation.username_required", "username must not be empty")
	case passwordForValidation == "":
		return session.HostRow{}, newInputValidationError("validation.password_required", "password must not be empty")
	}
	return session.HostRow{
		Target: session.HostTarget{
			IP:       ip,
			Port:     port,
			Username: username,
			Password: password,
		},
		Enabled: true,
	}, nil
}

func buildHostRowsFromCSV(input string) ([]session.HostRow, error) {
	if strings.TrimSpace(input) == "" {
		return nil, newInputValidationError("validation.csv_empty", "CSV input must not be empty")
	}
	parsed, errs := validation.ParseHostsCSV(input)
	if len(errs) > 0 {
		return nil, &csvImportError{errs: errs}
	}
	rows := make([]session.HostRow, 0, len(parsed))
	for _, parsedHost := range parsed {
		rows = append(rows, session.HostRow{
			Target: session.HostTarget{
				IP:       parsedHost.IP,
				Port:     parsedHost.Port,
				Username: parsedHost.Username,
				Password: parsedHost.Password,
			},
			Enabled: true,
		})
	}
	return rows, nil
}

func redactionSecrets(hosts []session.HostRow) []string {
	var secrets []string
	for _, host := range hosts {
		if host.Target.Password != "" {
			secrets = append(secrets, host.Target.Password)
		}
	}
	return secrets
}

func redactTextForHosts(text string, hosts []session.HostRow) string {
	return secure.RedactSecrets(text, redactionSecrets(hosts))
}

func resultHost(result session.HostResult) string {
	return fmt.Sprintf("%s:%d", result.TargetIP, result.Port)
}

func resultExitCode(result session.HostResult) string {
	if result.Operation != session.OpCommand || result.ExitCode < 0 {
		return "-"
	}
	return strconv.Itoa(result.ExitCode)
}

func resultErrorKind(locale *i18n.Locale, result session.HostResult) string {
	if result.ErrorKind == "" {
		return "-"
	}
	return displayErrorKind(locale, result.ErrorKind)
}

func formatResultForCopy(locale *i18n.Locale, result session.HostResult, hosts []session.HostRow) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: %s\n", locale.T("results.host"), resultHost(result)))
	sb.WriteString(fmt.Sprintf("%s: %s\n", locale.T("results.operation"), displayOperation(locale, result.Operation)))
	sb.WriteString(fmt.Sprintf("%s: %s\n", locale.T("cmd.status"), result.Status))
	sb.WriteString(fmt.Sprintf("%s: %s\n", locale.T("cmd.exit_code"), resultExitCode(result)))
	sb.WriteString(fmt.Sprintf("%s: %v\n", locale.T("cmd.success"), result.Success))
	sb.WriteString(fmt.Sprintf("%s: %s\n", locale.T("cmd.duration"), result.Duration.Round(time.Millisecond)))
	if result.ErrorKind != "" {
		sb.WriteString(fmt.Sprintf("%s: %s\n", locale.T("cmd.error"), resultErrorKind(locale, result)))
	}
	if result.ErrorMessage != "" {
		sb.WriteString(fmt.Sprintf("%s:\n%s\n", locale.T("results.error_message"), result.ErrorMessage))
	}
	if result.Stdout != "" {
		sb.WriteString(fmt.Sprintf("%s:\n%s\n", locale.T("cmd.stdout"), result.Stdout))
	}
	if result.Stderr != "" {
		sb.WriteString(fmt.Sprintf("%s:\n%s\n", locale.T("cmd.stderr"), result.Stderr))
	}
	return redactTextForHosts(sb.String(), hosts)
}

func formatAllResultsForCopy(locale *i18n.Locale, results []session.HostResult, hosts []session.HostRow) string {
	var sb strings.Builder
	for i, result := range results {
		if i > 0 {
			sb.WriteString("\n---\n")
		}
		sb.WriteString(formatResultForCopy(locale, result, hosts))
	}
	return redactTextForHosts(sb.String(), hosts)
}

func resultDetails(locale *i18n.Locale, result session.HostResult, hosts []session.HostRow) (stdout string, stderr string, errText string) {
	stdout = result.Stdout
	if stdout == "" {
		stdout = locale.T("results.empty_output")
	}
	stderr = result.Stderr
	if stderr == "" {
		stderr = locale.T("results.empty_output")
	}
	if result.ErrorKind == "" && result.ErrorMessage == "" {
		errText = locale.T("results.no_error")
	} else {
		var sb strings.Builder
		if result.ErrorKind != "" {
			sb.WriteString(fmt.Sprintf("%s: %s\n", locale.T("cmd.error"), resultErrorKind(locale, result)))
		}
		if result.ErrorMessage != "" {
			sb.WriteString(result.ErrorMessage)
		}
		errText = sb.String()
	}
	return redactTextForHosts(stdout, hosts), redactTextForHosts(stderr, hosts), redactTextForHosts(errText, hosts)
}

const (
	languageLabelZH = "中文 (zh-CN)"
	languageLabelEN = "English (en-US)"
)

func selectedLanguageLabel(lang i18n.Lang) string {
	if lang == i18n.EN {
		return languageLabelEN
	}
	return languageLabelZH
}

func languageFromSelection(selection string) i18n.Lang {
	if selection == languageLabelEN {
		return i18n.EN
	}
	return i18n.ZH
}

func setLocaleFromSelection(state *AppState, selection string) bool {
	newLang := languageFromSelection(selection)

	state.mu.Lock()
	defer state.mu.Unlock()
	if state.locale.Current() == newLang {
		return false
	}
	state.locale.Set(newLang)
	return true
}

// ---------- Top Toolbar ----------

func buildTopBar(w fyne.Window, state *AppState, refs *panelRefreshers) fyne.CanvasObject {
	loc := state.locale

	concurrencySelect := widget.NewSelect(optionStrings(config.ConcurrencyOptions()), func(s string) {
		v, _ := strconv.Atoi(s)
		state.mu.Lock()
		state.opts.Concurrency = v
		state.mu.Unlock()
	})
	concurrencySelect.SetSelected(strconv.Itoa(state.opts.Concurrency))

	timeoutSelect := widget.NewSelect([]string{"10", "30", "60", "120", "300"}, func(s string) {
		v, _ := strconv.Atoi(s)
		state.mu.Lock()
		state.opts.Timeout = v
		state.mu.Unlock()
	})
	timeoutSelect.SetSelected(strconv.Itoa(state.opts.Timeout))

	knownHostsEntry := widget.NewEntry()
	knownHostsEntry.SetPlaceHolder(loc.T("options.known_hosts_ph"))
	knownHostsEntry.SetText(state.opts.KnownHostsPath)
	knownHostsEntry.OnChanged = func(s string) {
		state.mu.Lock()
		state.opts.KnownHostsPath = s
		state.mu.Unlock()
	}

	warning := widget.NewLabel("")
	warning.Wrapping = fyne.TextTruncate
	warning.Truncation = fyne.TextTruncateEllipsis
	warning.Importance = widget.DangerImportance
	if state.opts.InsecureIgnoreHostKey {
		warning.SetText(loc.T("options.insecure_warn_short"))
	}
	warningSlot := container.New(&fixedHeightLayout{height: toolbarWarningHeight}, warning)

	insecureCheck := widget.NewCheck(loc.T("options.insecure_short"), nil)
	insecureCheck.SetChecked(state.opts.InsecureIgnoreHostKey)
	insecureCheck.OnChanged = func(v bool) {
		state.mu.Lock()
		state.opts.InsecureIgnoreHostKey = v
		state.mu.Unlock()
		if v {
			warning.SetText(loc.T("options.insecure_warn_short"))
		} else {
			warning.SetText("")
		}
		warning.Refresh()
	}

	langSelect := widget.NewSelect([]string{languageLabelZH, languageLabelEN}, nil)
	langSelect.SetSelected(selectedLanguageLabel(state.locale.Current()))
	langSelect.OnChanged = func(s string) {
		if !setLocaleFromSelection(state, s) {
			return
		}
		w.SetTitle(state.T("title"))
		if currentMainSplit != nil {
			state.setSplitOffset(float64(currentMainSplit.Offset))
		}
		currentSize := w.Canvas().Size()
		setMainContent(w, state)
		if currentSize.Width > 0 && currentSize.Height > 0 {
			w.Resize(currentSize)
		}
	}

	runControls := container.NewHBox(
		toolbarGroupLabel(loc.T("options.run_group")),
		hSpacer(8),
		widget.NewLabel(loc.T("cmd.timeout")),
		hSpacer(4),
		fixedSize(toolbarSelectSize, timeoutSelect),
		hSpacer(14),
		widget.NewLabel(loc.T("cmd.concurrency")),
		hSpacer(4),
		fixedSize(toolbarSelectSize, concurrencySelect),
	)
	sshControls := container.NewHBox(
		toolbarGroupLabel(loc.T("options.ssh_group")),
		hSpacer(8),
		widget.NewLabel(loc.T("options.known_hosts_short")),
		hSpacer(4),
		fixedSize(knownHostsControlSize, knownHostsEntry),
		hSpacer(14),
		fixedSize(insecureCheckSize, insecureCheck),
	)
	uiControls := container.NewHBox(
		toolbarGroupLabel(loc.T("options.ui_group")),
		hSpacer(8),
		widget.NewLabel(loc.T("options.lang")),
		hSpacer(4),
		fixedSize(languageSelectSize, langSelect),
	)

	firstRow := container.NewBorder(nil, nil, runControls, uiControls, nil)
	secondRow := container.NewBorder(nil, nil, sshControls, nil, nil)
	toolbar := container.NewVBox(firstRow, secondRow, warningSlot)

	return container.New(&stableSizeLayout{min: toolbarMinSize}, container.NewPadded(toolbar))
}

// ---------- Hosts Panel ----------

func buildHostsPanel(w fyne.Window, state *AppState, refs *panelRefreshers) fyne.CanvasObject {
	loc := state.locale
	hostRows := state.mgr.Hosts()

	statsLabel := widget.NewLabel("")
	emptyState := buildHostsEmptyState(loc)
	hostTable := widget.NewTable(
		func() (int, int) {
			return len(hostRows) + 1, len(hostTableHeaders(loc))
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Truncation = fyne.TextTruncateEllipsis
			return label
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			label.TextStyle = fyne.TextStyle{Bold: id.Row == 0}
			if id.Row == 0 {
				label.SetText(hostTableHeaders(loc)[id.Col])
				return
			}
			row := hostRows[id.Row-1]
			enabled := loc.T("common.yes")
			if !row.Enabled {
				enabled = loc.T("common.no")
			}
			values := []string{
				enabled,
				row.Target.IP,
				strconv.Itoa(row.Target.Port),
				row.Target.Username,
				hostStatusText(loc, row),
			}
			label.SetText(values[id.Col])
		},
	)

	applyHostTableColumnWidths(hostTable)
	hostTable.SetRowHeight(0, 34)
	body := container.NewMax()
	updateBody := func() {
		if len(hostRows) == 0 {
			body.Objects = []fyne.CanvasObject{emptyState}
		} else {
			body.Objects = []fyne.CanvasObject{hostTable}
		}
		body.Refresh()
	}
	updateBody()

	refs.hosts = func() {
		hostRows = state.mgr.Hosts()
		total, enabled := hostCounts(hostRows)
		statsLabel.SetText(fmt.Sprintf("%s: %d    %s: %d    %s: %d", loc.T("hosts.total"), total, loc.T("hosts.enabled_count"), enabled, loc.T("hosts.selected_count"), enabled))
		for i := 1; i <= len(hostRows); i++ {
			hostTable.SetRowHeight(i, 38)
		}
		updateBody()
		hostTable.Refresh()
		statsLabel.Refresh()
		if refs.operation != nil {
			refs.operation()
		}
	}

	hostTable.OnSelected = func(id widget.TableCellID) {
		if id.Row <= 0 || id.Row > len(hostRows) {
			return
		}
		current := state.mgr.Hosts()
		idx := id.Row - 1
		if idx < len(current) {
			current[idx].Enabled = !current[idx].Enabled
			state.mu.Lock()
			state.mgr.SetHosts(current)
			state.mu.Unlock()
			refs.hosts()
		}
	}

	addBtn := widget.NewButtonWithIcon(loc.T("hosts.add"), theme.ContentAddIcon(), func() {
		showAddHostDialog(w, state, refs.refreshAll)
	})
	addBtn.Importance = widget.HighImportance

	importBtn := widget.NewButtonWithIcon(loc.T("hosts.import"), theme.UploadIcon(), func() {
		showImportHostsDialog(w, state, refs.refreshAll)
	})

	var testConnBtn *widget.Button
	testConnBtn = widget.NewButtonWithIcon(loc.T("hosts.test_conn"), theme.SearchIcon(), func() {
		enabled := state.mgr.EnabledHosts()
		if len(enabled) == 0 {
			dialog.ShowError(fmt.Errorf("%s", loc.T("cmd.no_hosts")), w)
			return
		}

		opts := configureSessionOptions(state)
		state.setRunStarted(session.OpTestConnection, len(enabled))
		if refs.results != nil {
			refs.results()
		}

		testConnBtn.SetText(loc.T("hosts.testing"))
		testConnBtn.Disable()
		current := state.mgr.Hosts()
		for i := range current {
			if current[i].Enabled {
				current[i].Status = loc.T("hosts.testing")
			}
		}
		state.mu.Lock()
		state.mgr.SetHosts(current)
		state.mu.Unlock()
		refs.hosts()

		go func(opts config.RuntimeOptions) {
			ctx, cancel := contextWithRuntimeTimeout(context.Background(), opts)
			defer cancel()
			results := state.mgr.TestConnections(ctx)

			fyne.Do(func() {
				current := state.mgr.Hosts()
				for _, result := range results {
					for i := range current {
						if current[i].Target.IP == result.TargetIP && current[i].Target.Port == result.Port && current[i].Target.Username == result.Username {
							if result.Success {
								current[i].Status = "ok"
							} else {
								current[i].Status = result.ErrorKind
							}
							break
						}
					}
				}
				state.mu.Lock()
				state.mgr.SetHosts(current)
				state.mu.Unlock()
				state.setRunFinished(session.OpTestConnection, results)
				testConnBtn.SetText(loc.T("hosts.test_conn"))
				testConnBtn.Enable()
				refs.refreshAll()
				showResultInspector(w, state, refs.refreshAll)
			})
		}(opts)
	})
	testConnBtn.Importance = widget.HighImportance

	removeBtn := widget.NewButtonWithIcon(loc.T("hosts.remove"), theme.ContentRemoveIcon(), func() {
		current := state.mgr.Hosts()
		var newRows []session.HostRow
		for _, host := range current {
			if host.Enabled {
				newRows = append(newRows, host)
			}
		}
		state.mu.Lock()
		state.mgr.SetHosts(newRows)
		state.mu.Unlock()
		refs.refreshAll()
	})

	clearBtn := widget.NewButtonWithIcon(loc.T("hosts.clear"), theme.ContentClearIcon(), func() {
		state.mu.Lock()
		state.mgr.SetHosts(nil)
		state.mu.Unlock()
		refs.refreshAll()
	})
	clearBtn.Importance = widget.WarningImportance

	primaryButtons := container.NewHBox(
		fixedSize(hostActionButtonSize, addBtn),
		fixedSize(hostActionButtonSize, importBtn),
	)
	testButtons := fixedSize(hostWideButtonSize, testConnBtn)
	cleanupButtons := container.NewHBox(
		fixedSize(hostActionButtonSize, removeBtn),
		fixedSize(hostActionButtonSize, clearBtn),
	)

	content := container.NewBorder(
		container.NewVBox(
			heading(loc.T("panel.hosts")),
			statsLabel,
			primaryButtons,
			testButtons,
			widget.NewSeparator(),
		),
		container.NewVBox(widget.NewSeparator(), cleanupButtons),
		nil, nil,
		body,
	)
	return container.New(&stableSizeLayout{min: hostsPanelMinSize}, content)
}

func buildHostsEmptyState(locale *i18n.Locale) fyne.CanvasObject {
	label := wrappedLabel(locale.T("hosts.empty"))
	label.Alignment = fyne.TextAlignCenter
	return container.NewCenter(fixedSize(hostsEmptyStateSize, container.NewPadded(label)))
}

func showAddHostDialog(w fyne.Window, state *AppState, onAdded func()) {
	loc := state.locale

	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder(loc.T("hosts.ip_placeholder"))
	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder(loc.T("hosts.port_placeholder"))
	portEntry.SetText("22")
	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder(loc.T("hosts.user_placeholder"))
	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder(loc.T("hosts.pass_placeholder"))

	errorLabel := wrappedLabel("")
	errorLabel.Importance = widget.DangerImportance

	form := container.NewVBox(
		fieldBlock(loc.T("hosts.ip_placeholder"), ipEntry),
		fieldBlock(loc.T("hosts.port_placeholder"), portEntry),
		fieldBlock(loc.T("hosts.user_placeholder"), userEntry),
		fieldBlock(loc.T("hosts.pass_placeholder"), passEntry),
		errorLabel,
	)

	d := dialog.NewCustomWithoutButtons(loc.T("dialog.add_host.title"), form, w)
	cancelBtn := widget.NewButtonWithIcon(loc.T("dialog.cancel"), theme.CancelIcon(), func() {
		d.Hide()
	})
	addBtn := widget.NewButtonWithIcon(loc.T("dialog.add"), theme.ConfirmIcon(), func() {
		row, err := buildHostRowFromInput(ipEntry.Text, portEntry.Text, userEntry.Text, passEntry.Text)
		if err != nil {
			errorLabel.SetText(localizedErrorText(loc, err))
			errorLabel.Refresh()
			return
		}
		current := state.mgr.Hosts()
		current = append(current, row)
		state.mu.Lock()
		state.mgr.SetHosts(current)
		state.mu.Unlock()
		if onAdded != nil {
			onAdded()
		}
		d.Hide()
	})
	addBtn.Importance = widget.HighImportance
	d.SetButtons([]fyne.CanvasObject{cancelBtn, addBtn})
	d.Resize(fyne.NewSize(460, 340))
	d.Show()
}

func showImportHostsDialog(w fyne.Window, state *AppState, onImported func()) {
	loc := state.locale

	csvEntry := widget.NewMultiLineEntry()
	csvEntry.SetPlaceHolder(loc.T("hosts.csv_placeholder"))
	csvEntry.SetMinRowsVisible(8)

	errorLabel := wrappedLabel("")
	errorLabel.Importance = widget.DangerImportance

	content := container.NewVBox(
		wrappedLabel(loc.T("dialog.import_hosts.help")),
		csvEntry,
		errorLabel,
	)

	d := dialog.NewCustomWithoutButtons(loc.T("dialog.import_hosts.title"), content, w)
	cancelBtn := widget.NewButtonWithIcon(loc.T("dialog.cancel"), theme.CancelIcon(), func() {
		d.Hide()
	})
	importBtn := widget.NewButtonWithIcon(loc.T("dialog.import"), theme.ConfirmIcon(), func() {
		rows, err := buildHostRowsFromCSV(csvEntry.Text)
		if err != nil {
			errorLabel.SetText(localizedErrorText(loc, err))
			errorLabel.Refresh()
			return
		}
		current := state.mgr.Hosts()
		current = append(current, rows...)
		state.mu.Lock()
		state.mgr.SetHosts(current)
		state.mu.Unlock()
		if onImported != nil {
			onImported()
		}
		d.Hide()
	})
	importBtn.Importance = widget.HighImportance
	d.SetButtons([]fyne.CanvasObject{cancelBtn, importBtn})
	d.Resize(fyne.NewSize(680, 500))
	d.Show()
}

// ---------- Operation Panel ----------

func buildOperationPanel(w fyne.Window, state *AppState, refs *panelRefreshers) fyne.CanvasObject {
	loc := state.locale

	headerSelectedLabel := widget.NewLabel("")
	var modeRefreshers []func(int)
	registerModeRefresher := func(f func(int)) {
		modeRefreshers = append(modeRefreshers, f)
	}
	refs.operation = func() {
		enabled := enabledHostCount(state)
		text := fmt.Sprintf("%s: %d", loc.T("hosts.selected_count"), enabled)
		headerSelectedLabel.SetText(text)
		headerSelectedLabel.Refresh()
		for _, refresh := range modeRefreshers {
			refresh(enabled)
		}
	}

	commandTab := container.NewTabItem(loc.T("operation.quick_command"), buildQuickCommandPanel(w, state, refs, registerModeRefresher))
	uploadTab := container.NewTabItem(loc.T("operation.upload_file"), buildUploadPanel(w, state, refs, registerModeRefresher))
	downloadTab := container.NewTabItem(loc.T("operation.download_file"), buildDownloadPanel(w, state, refs, registerModeRefresher))
	tabs := container.NewAppTabs(commandTab, uploadTab, downloadTab)
	tabs.SetTabLocation(container.TabLocationTop)
	tabs.OnSelected = func(item *container.TabItem) {
		switch item {
		case uploadTab:
			state.setCurrentOperation(session.OpUpload)
		case downloadTab:
			state.setCurrentOperation(session.OpDownload)
		default:
			state.setCurrentOperation(session.OpCommand)
		}
	}
	switch state.currentOperation() {
	case session.OpUpload:
		tabs.Select(uploadTab)
	case session.OpDownload:
		tabs.Select(downloadTab)
	default:
		tabs.Select(commandTab)
	}

	return container.NewBorder(
		container.NewVBox(heading(loc.T("panel.operation")), headerSelectedLabel, widget.NewSeparator()),
		nil, nil, nil,
		tabs,
	)
}

func buildQuickCommandPanel(w fyne.Window, state *AppState, refs *panelRefreshers, registerRefresh func(func(int))) fyne.CanvasObject {
	loc := state.locale

	selectedLabel := widget.NewLabel("")
	cmdEntry := widget.NewMultiLineEntry()
	cmdEntry.SetPlaceHolder(loc.T("cmd.placeholder"))
	cmdEntry.SetMinRowsVisible(5)

	explain := wrappedLabel(loc.T("cmd.nonzero_note"))
	explain.Importance = widget.WarningImportance

	running := false
	var runBtn *widget.Button
	update := func(count int) {
		selectedLabel.SetText(fmt.Sprintf("%s: %d", loc.T("hosts.selected_count"), count))
		if running {
			runBtn.SetText(loc.T("cmd.running"))
			runBtn.Disable()
		} else {
			runBtn.SetText(commandRunLabel(loc, count))
			if count == 0 || strings.TrimSpace(cmdEntry.Text) == "" {
				runBtn.Disable()
			} else {
				runBtn.Enable()
			}
		}
		selectedLabel.Refresh()
		runBtn.Refresh()
	}
	runBtn = widget.NewButtonWithIcon(loc.T("cmd.run"), theme.MediaPlayIcon(), func() {
		state.setCurrentOperation(session.OpCommand)
		cmd := strings.TrimSpace(cmdEntry.Text)
		if err := validation.ValidateCommand(cmd); err != nil {
			dialog.ShowError(fmt.Errorf("%s", loc.T("cmd.empty")), w)
			return
		}
		enabled := state.mgr.EnabledHosts()
		if len(enabled) == 0 {
			dialog.ShowError(fmt.Errorf("%s", loc.T("cmd.no_hosts")), w)
			return
		}

		opts := configureSessionOptions(state)
		state.setRunStarted(session.OpCommand, len(enabled))
		if refs.results != nil {
			refs.results()
		}

		running = true
		update(len(enabled))
		go func(opts config.RuntimeOptions) {
			ctx, cancel := contextWithRuntimeTimeout(context.Background(), opts)
			defer cancel()
			results := state.mgr.ExecCommands(ctx, cmd)

			fyne.Do(func() {
				state.setRunFinished(session.OpCommand, results)
				running = false
				update(enabledHostCount(state))
				refs.refreshAll()
				showResultInspector(w, state, refs.refreshAll)
			})
		}(opts)
	})
	runBtn.Importance = widget.HighImportance
	cmdEntry.OnChanged = func(_ string) {
		update(enabledHostCount(state))
	}
	registerRefresh(update)

	content := container.NewVBox(
		heading(loc.T("cmd.label")),
		fixedWidth(operationFieldWidth, cmdEntry),
		selectedLabel,
		container.NewHBox(fixedSize(operationButtonMinSize, runBtn)),
		explain,
	)
	return operationModePanel(content)
}

func buildUploadPanel(w fyne.Window, state *AppState, refs *panelRefreshers, registerRefresh func(func(int))) fyne.CanvasObject {
	loc := state.locale

	selectedLabel := widget.NewLabel("")
	uploadLocal := widget.NewEntry()
	uploadLocal.SetPlaceHolder(loc.T("transfer.upload_local_ph"))
	browseUpload := widget.NewButtonWithIcon(loc.T("transfer.browse"), theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()
			uploadLocal.SetText(reader.URI().Path())
		}, w)
	})
	uploadLocalRow := browseFieldRow(uploadLocal, browseUpload)

	uploadRemote := widget.NewEntry()
	uploadRemote.SetPlaceHolder(loc.T("transfer.upload_remote_ph"))
	uploadOverwrite := widget.NewCheck(loc.T("transfer.overwrite"), nil)

	unsupportedNote := wrappedLabel(loc.T("transfer.unsupported"))
	unsupportedNote.Importance = widget.WarningImportance

	running := false
	var uploadBtn *widget.Button
	update := func(count int) {
		selectedLabel.SetText(fmt.Sprintf("%s: %d", loc.T("hosts.selected_count"), count))
		if running {
			uploadBtn.SetText(loc.T("transfer.uploading"))
			uploadBtn.Disable()
		} else {
			uploadBtn.SetText(uploadRunLabel(loc, count))
			if count == 0 || strings.TrimSpace(uploadLocal.Text) == "" || strings.TrimSpace(uploadRemote.Text) == "" {
				uploadBtn.Disable()
			} else {
				uploadBtn.Enable()
			}
		}
		selectedLabel.Refresh()
		uploadBtn.Refresh()
	}
	uploadBtn = widget.NewButtonWithIcon(loc.T("transfer.upload_btn"), theme.UploadIcon(), func() {
		state.setCurrentOperation(session.OpUpload)
		localPath := strings.TrimSpace(uploadLocal.Text)
		remotePath := strings.TrimSpace(uploadRemote.Text)
		if err := validation.ValidateLocalFilePath(localPath); err != nil {
			showLocalizedError(w, loc, err)
			return
		}
		if err := validation.ValidateRemoteFilePath(remotePath); err != nil {
			showLocalizedError(w, loc, err)
			return
		}
		enabled := state.mgr.EnabledHosts()
		if len(enabled) == 0 {
			dialog.ShowError(fmt.Errorf("%s", loc.T("cmd.no_hosts")), w)
			return
		}

		opts := configureSessionOptions(state)
		overwrite := uploadOverwrite.Checked
		state.setRunStarted(session.OpUpload, len(enabled))
		if refs.results != nil {
			refs.results()
		}

		running = true
		update(len(enabled))
		go func(opts config.RuntimeOptions, overwrite bool) {
			ctx, cancel := contextWithRuntimeTimeout(context.Background(), opts)
			defer cancel()
			results := state.mgr.UploadFiles(ctx, localPath, remotePath, overwrite)
			fyne.Do(func() {
				state.setRunFinished(session.OpUpload, results)
				running = false
				update(enabledHostCount(state))
				refs.refreshAll()
				showResultInspector(w, state, refs.refreshAll)
			})
		}(opts, overwrite)
	})
	uploadBtn.Importance = widget.HighImportance
	uploadLocal.OnChanged = func(_ string) { update(enabledHostCount(state)) }
	uploadRemote.OnChanged = func(_ string) { update(enabledHostCount(state)) }
	registerRefresh(update)

	content := container.NewVBox(
		heading(loc.T("operation.upload_file")),
		fieldBlock(loc.T("transfer.upload_local"), uploadLocalRow),
		fieldBlock(loc.T("transfer.upload_remote"), fixedWidth(operationFieldWidth, uploadRemote)),
		uploadOverwrite,
		selectedLabel,
		container.NewHBox(fixedSize(operationButtonMinSize, uploadBtn)),
		unsupportedNote,
	)
	return operationModePanel(content)
}

func buildDownloadPanel(w fyne.Window, state *AppState, refs *panelRefreshers, registerRefresh func(func(int))) fyne.CanvasObject {
	loc := state.locale

	selectedLabel := widget.NewLabel("")
	downloadRemote := widget.NewEntry()
	downloadRemote.SetPlaceHolder(loc.T("transfer.download_remote_ph"))

	downloadLocal := widget.NewEntry()
	downloadLocal.SetPlaceHolder(loc.T("transfer.download_local_ph"))
	browseDownload := widget.NewButtonWithIcon(loc.T("transfer.browse"), theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			downloadLocal.SetText(uri.Path())
		}, w)
	})
	downloadLocalRow := browseFieldRow(downloadLocal, browseDownload)

	downloadOverwrite := widget.NewCheck(loc.T("transfer.overwrite"), nil)
	downloadNote := wrappedLabel(loc.T("transfer.download_note"))
	downloadNote.Importance = widget.WarningImportance

	running := false
	var downloadBtn *widget.Button
	update := func(count int) {
		selectedLabel.SetText(fmt.Sprintf("%s: %d", loc.T("hosts.selected_count"), count))
		if running {
			downloadBtn.SetText(loc.T("transfer.downloading"))
			downloadBtn.Disable()
		} else {
			downloadBtn.SetText(downloadRunLabel(loc, count))
			if count == 0 || strings.TrimSpace(downloadRemote.Text) == "" || strings.TrimSpace(downloadLocal.Text) == "" {
				downloadBtn.Disable()
			} else {
				downloadBtn.Enable()
			}
		}
		selectedLabel.Refresh()
		downloadBtn.Refresh()
	}
	downloadBtn = widget.NewButtonWithIcon(loc.T("transfer.download_btn"), theme.DownloadIcon(), func() {
		state.setCurrentOperation(session.OpDownload)
		remotePath := strings.TrimSpace(downloadRemote.Text)
		localDir := strings.TrimSpace(downloadLocal.Text)
		if err := validation.ValidateRemoteFilePath(remotePath); err != nil {
			showLocalizedError(w, loc, err)
			return
		}
		if err := validation.ValidateLocalDirPath(localDir); err != nil {
			showLocalizedError(w, loc, err)
			return
		}
		enabled := state.mgr.EnabledHosts()
		if len(enabled) == 0 {
			dialog.ShowError(fmt.Errorf("%s", loc.T("cmd.no_hosts")), w)
			return
		}

		opts := configureSessionOptions(state)
		overwrite := downloadOverwrite.Checked
		state.setRunStarted(session.OpDownload, len(enabled))
		if refs.results != nil {
			refs.results()
		}

		running = true
		update(len(enabled))
		go func(opts config.RuntimeOptions, overwrite bool) {
			ctx, cancel := contextWithRuntimeTimeout(context.Background(), opts)
			defer cancel()
			results := state.mgr.DownloadFiles(ctx, remotePath, localDir, overwrite)
			fyne.Do(func() {
				state.setRunFinished(session.OpDownload, results)
				running = false
				update(enabledHostCount(state))
				refs.refreshAll()
				showResultInspector(w, state, refs.refreshAll)
			})
		}(opts, overwrite)
	})
	downloadBtn.Importance = widget.HighImportance
	downloadRemote.OnChanged = func(_ string) { update(enabledHostCount(state)) }
	downloadLocal.OnChanged = func(_ string) { update(enabledHostCount(state)) }
	registerRefresh(update)

	content := container.NewVBox(
		heading(loc.T("operation.download_file")),
		fieldBlock(loc.T("transfer.download_remote"), fixedWidth(operationFieldWidth, downloadRemote)),
		fieldBlock(loc.T("transfer.download_local"), downloadLocalRow),
		downloadOverwrite,
		selectedLabel,
		container.NewHBox(fixedSize(operationButtonMinSize, downloadBtn)),
		downloadNote,
	)
	return operationModePanel(content)
}

// ---------- Status Bar and Result Inspector ----------

func statusSummaryText(locale *i18n.Locale, summary runSummary, hasResults bool) string {
	sep := locale.T("status.separator")
	if summary.Running > 0 {
		return fmt.Sprintf(
			"%s%s%s%s%s %d%s%s %d",
			locale.T("status.running"),
			locale.T("status.colon"),
			displayStatusOperation(locale, summary.Operation),
			sep,
			locale.T("status.total"),
			summary.Total,
			sep,
			locale.T("status.running_count"),
			summary.Running,
		)
	}
	if !hasResults {
		return locale.T("status.no_results")
	}
	return fmt.Sprintf(
		"%s%s%s%s%s %d%s%s %d%s%s %d%s%s %d",
		locale.T("status.last_operation"),
		locale.T("status.colon"),
		displayStatusOperation(locale, summary.Operation),
		sep,
		locale.T("status.total"),
		summary.Total,
		sep,
		locale.T("status.success"),
		summary.Success,
		sep,
		locale.T("status.failed"),
		summary.Failure,
		sep,
		locale.T("status.running_count"),
		summary.Running,
	)
}

func displayStatusOperation(locale *i18n.Locale, op session.Operation) string {
	if op == "" {
		return locale.T("op.none")
	}
	key := "status.op." + string(op)
	t := locale.T(key)
	if t != key {
		return t
	}
	return displayOperation(locale, op)
}

func refreshStatusBarView(locale *i18n.Locale, summary runSummary, hasResults bool, statusLabel *widget.Label, viewBtn *widget.Button) {
	statusLabel.SetText(statusSummaryText(locale, summary, hasResults))
	if hasResults {
		viewBtn.Enable()
	} else {
		viewBtn.Disable()
	}
	statusLabel.Refresh()
	viewBtn.Refresh()
}

func buildStatusBar(w fyne.Window, state *AppState, refs *panelRefreshers) fyne.CanvasObject {
	loc := state.locale
	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextTruncate
	statusLabel.Truncation = fyne.TextTruncateEllipsis
	viewBtn := widget.NewButtonWithIcon(loc.T("results.view"), theme.VisibilityIcon(), func() {
		if len(state.mgr.Results()) == 0 {
			dialog.ShowInformation(loc.T("results.view"), loc.T("results.no_results"), w)
			return
		}
		showResultInspector(w, state, refs.refreshAll)
	})

	refs.results = func() {
		results := state.mgr.Results()
		refreshStatusBarView(loc, state.runSummary(), len(results) > 0, statusLabel, viewBtn)
	}
	refs.results()

	bar := container.NewBorder(
		widget.NewSeparator(),
		nil, nil,
		fixedSize(viewResultsButtonSize, viewBtn),
		container.NewPadded(statusLabel),
	)
	return container.New(&stableSizeLayout{min: statusBarMinSize}, bar)
}

func showResultInspector(owner fyne.Window, state *AppState, onChanged func()) {
	results := state.mgr.Results()
	if len(results) == 0 {
		dialog.ShowInformation(state.T("results.view"), state.T("results.no_results"), owner)
		return
	}
	app := fyne.CurrentApp()
	if app == nil {
		dialog.ShowInformation(state.T("results.view"), state.T("results.no_results"), owner)
		return
	}

	inspector := app.NewWindow(state.T("inspector.title"))
	inspector.SetPadded(true)
	content := container.New(&minSizeLayout{min: resultInspectorMinSize}, buildResultInspectorContent(inspector, state, onChanged))
	inspector.SetContent(content)
	inspector.Resize(resultInspectorSize)
	inspector.Show()
}

func buildResultInspectorContent(w fyne.Window, state *AppState, onChanged func()) fyne.CanvasObject {
	loc := state.locale
	resultRows := state.mgr.Results()

	summaryLabel := wrappedLabel("")
	stdoutEntry := widget.NewMultiLineEntry()
	stdoutEntry.SetMinRowsVisible(12)
	stdoutEntry.Disable()
	stderrEntry := widget.NewMultiLineEntry()
	stderrEntry.SetMinRowsVisible(12)
	stderrEntry.Disable()
	errorEntry := widget.NewMultiLineEntry()
	errorEntry.SetMinRowsVisible(12)
	errorEntry.Disable()

	updateDetails := func() {
		idx := state.selectedResultIndex()
		if idx < 0 || idx >= len(resultRows) {
			stdoutEntry.SetText(loc.T("results.no_selection"))
			stderrEntry.SetText(loc.T("results.no_selection"))
			errorEntry.SetText(loc.T("results.no_selection"))
			return
		}
		stdout, stderr, errText := resultDetails(loc, resultRows[idx], state.mgr.Hosts())
		stdoutEntry.SetText(stdout)
		stderrEntry.SetText(stderr)
		errorEntry.SetText(errText)
	}

	resultTable := widget.NewTable(
		func() (int, int) {
			return len(resultRows) + 1, 6
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Truncation = fyne.TextTruncateEllipsis
			return label
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			label.TextStyle = fyne.TextStyle{Bold: id.Row == 0}
			if id.Row == 0 {
				headers := []string{
					loc.T("results.host"),
					loc.T("results.operation"),
					loc.T("cmd.status"),
					loc.T("cmd.exit_code"),
					loc.T("cmd.duration"),
					loc.T("results.error_kind"),
				}
				label.SetText(headers[id.Col])
				return
			}
			result := resultRows[id.Row-1]
			values := []string{
				resultHost(result),
				displayOperation(loc, result.Operation),
				result.Status,
				resultExitCode(result),
				result.Duration.Round(time.Millisecond).String(),
				resultErrorKind(loc, result),
			}
			label.SetText(values[id.Col])
		},
	)
	resultTable.SetColumnWidth(0, 190)
	resultTable.SetColumnWidth(1, 130)
	resultTable.SetColumnWidth(2, 120)
	resultTable.SetColumnWidth(3, 90)
	resultTable.SetColumnWidth(4, 120)
	resultTable.SetColumnWidth(5, 210)
	resultTable.SetRowHeight(0, 34)
	for i := 1; i <= len(resultRows); i++ {
		resultTable.SetRowHeight(i, 36)
	}
	resultTable.OnSelected = func(id widget.TableCellID) {
		if id.Row <= 0 {
			return
		}
		state.setSelectedResult(id.Row - 1)
		updateDetails()
	}

	summary := state.runSummary()
	summaryLabel.SetText(statusSummaryText(loc, summary, len(resultRows) > 0))
	updateDetails()

	copyCurrentBtn := widget.NewButtonWithIcon(loc.T("results.copy_current"), theme.ContentCopyIcon(), func() {
		idx := state.selectedResultIndex()
		if idx < 0 || idx >= len(resultRows) {
			dialog.ShowInformation(loc.T("results.copy_current"), loc.T("results.no_selection"), w)
			return
		}
		text := formatResultForCopy(loc, resultRows[idx], state.mgr.Hosts())
		if app := fyne.CurrentApp(); app != nil {
			app.Clipboard().SetContent(text)
			return
		}
		w.Clipboard().SetContent(text)
	})
	copyAllBtn := widget.NewButtonWithIcon(loc.T("results.copy_all"), theme.ContentCopyIcon(), func() {
		text := formatAllResultsForCopy(loc, resultRows, state.mgr.Hosts())
		if app := fyne.CurrentApp(); app != nil {
			app.Clipboard().SetContent(text)
			return
		}
		w.Clipboard().SetContent(text)
	})
	clearBtn := widget.NewButtonWithIcon(loc.T("results.clear"), theme.ContentClearIcon(), func() {
		state.mgr.ClearResults()
		state.clearRunSummary()
		if onChanged != nil {
			onChanged()
		}
		w.Close()
	})
	closeBtn := widget.NewButtonWithIcon(loc.T("dialog.close"), theme.CancelIcon(), func() {
		w.Close()
	})

	detailTabs := container.NewAppTabs(
		container.NewTabItem(loc.T("cmd.stdout"), stdoutEntry),
		container.NewTabItem(loc.T("cmd.stderr"), stderrEntry),
		container.NewTabItem(loc.T("results.error_message"), errorEntry),
	)
	detailTabs.SetTabLocation(container.TabLocationTop)

	tableAndDetails := container.NewVSplit(resultTable, detailTabs)
	tableAndDetails.Offset = 0.34

	return container.NewBorder(
		container.NewVBox(heading(loc.T("inspector.title")), summaryLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), container.NewHBox(copyCurrentBtn, copyAllBtn, clearBtn, layout.NewSpacer(), closeBtn)),
		nil, nil,
		tableAndDetails,
	)
}
