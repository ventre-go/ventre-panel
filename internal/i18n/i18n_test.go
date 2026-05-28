package i18n

import (
	"strings"
	"testing"
)

func TestNewLocaleDefaultsToZH(t *testing.T) {
	l := NewLocale()
	if l.Current() != ZH {
		t.Errorf("expected default language zh-CN, got %s", l.Current())
	}
}

func TestZHTranslations(t *testing.T) {
	l := NewLocale()
	l.Set(ZH)

	tests := []struct{ key, want string }{
		{"title", "ventre-panel — 批量 SSH 面板"},
		{"tab.hosts", "主机"},
		{"tab.command", "命令"},
		{"tab.transfer", "传输"},
		{"tab.results", "历史结果"},
		{"tab.run_options", "运行选项"},
		{"hosts.add", "添加主机"},
		{"cmd.run", "在选中主机上执行"},
		{"transfer.upload_btn", "上传到选中主机"},
		{"transfer.unsupported", "当前版本仅支持单文件上传和单文件下载，不支持目录传输。"},
		{"results.no_results", "本次运行暂无结果。"},
		{"op.test_connection", "连接测试"},
		{"err.auth_failed", "认证失败"},
		{"err.host_key_failed", "主机密钥校验失败"},
		{"err.timeout", "操作超时"},
		{"err.not_found", "文件或路径不存在"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := l.T(tt.key)
			if got != tt.want {
				t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestENTranslations(t *testing.T) {
	l := NewLocale()
	l.Set(EN)

	tests := []struct{ key, want string }{
		{"title", "ventre-panel — Batch SSH Panel"},
		{"tab.hosts", "Hosts"},
		{"tab.command", "Command"},
		{"tab.transfer", "Transfer"},
		{"tab.results", "Result History"},
		{"hosts.add", "Add Host"},
		{"cmd.run", "Run on Selected Hosts"},
		{"transfer.upload_btn", "Upload to Selected Hosts"},
		{"op.test_connection", "Connection Test"},
		{"err.auth_failed", "Authentication failed"},
		{"err.host_key_failed", "Host key verification failed"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := l.T(tt.key)
			if got != tt.want {
				t.Errorf("T(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestMissingKeyFallback(t *testing.T) {
	l := NewLocale()

	// Custom key not in any translation map
	got := l.T("nonexistent.key.xyz")
	if got != "nonexistent.key.xyz" {
		t.Errorf("expected fallback to key itself, got %q", got)
	}
}

func TestLanguageSwitch(t *testing.T) {
	l := NewLocale()

	// Start in zh-CN
	if l.Current() != ZH {
		t.Errorf("expected zh-CN, got %s", l.Current())
	}
	if l.T("tab.hosts") != "主机" {
		t.Errorf("expected Chinese, got %s", l.T("tab.hosts"))
	}

	// Switch to en-US
	l.Set(EN)
	if l.Current() != EN {
		t.Errorf("expected en-US, got %s", l.Current())
	}
	if l.T("tab.hosts") != "Hosts" {
		t.Errorf("expected English, got %s", l.T("tab.hosts"))
	}

	// Switch back to zh-CN
	l.Set(ZH)
	if l.T("tab.hosts") != "主机" {
		t.Errorf("expected Chinese after switch back, got %s", l.T("tab.hosts"))
	}
}

func TestErrorKindChinese(t *testing.T) {
	l := NewLocale()
	l.Set(ZH)

	kindToZH := map[string]string{
		"auth_failed":       "认证失败",
		"connect_failed":    "连接失败",
		"host_key_failed":   "主机密钥校验失败",
		"timeout":           "操作超时",
		"canceled":          "已取消",
		"command_failed":    "命令执行失败（非零退出码）",
		"not_found":         "文件或路径不存在",
		"already_exists":    "目标已存在",
		"permission_denied": "权限不足",
		"transfer_failed":   "文件传输失败",
		"invalid_request":   "请求无效",
		"unsupported":       "不支持的操作",
		"closed":            "连接已关闭",
		"internal":          "内部错误",
	}

	for kind, want := range kindToZH {
		key := "err." + kind
		got := l.T(key)
		if got != want {
			t.Errorf("T(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestErrorKindEnglish(t *testing.T) {
	l := NewLocale()
	l.Set(EN)

	kindToEN := map[string]string{
		"auth_failed":       "Authentication failed",
		"connect_failed":    "Connection failed",
		"host_key_failed":   "Host key verification failed",
		"timeout":           "Operation timed out",
		"canceled":          "Canceled",
		"command_failed":    "Command failed (non-zero exit code)",
		"not_found":         "File or path not found",
		"already_exists":    "Target already exists",
		"permission_denied": "Permission denied",
		"transfer_failed":   "File transfer failed",
		"invalid_request":   "Invalid request",
		"unsupported":       "Unsupported operation",
		"closed":            "Connection closed",
		"internal":          "Internal error",
	}

	for kind, want := range kindToEN {
		key := "err." + kind
		got := l.T(key)
		if got != want {
			t.Errorf("T(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestDirectoryUnsupportedBilingual(t *testing.T) {
	l := NewLocale()

	l.Set(ZH)
	zhText := l.T("transfer.unsupported")
	if zhText == "" {
		t.Error("zh-CN transfer.unsupported is empty")
	}
	if zhText == "transfer.unsupported" {
		t.Error("zh-CN transfer.unsupported not translated")
	}

	l.Set(EN)
	enText := l.T("transfer.unsupported")
	if enText == "" {
		t.Error("en-US transfer.unsupported is empty")
	}
	if enText == "transfer.unsupported" {
		t.Error("en-US transfer.unsupported not translated")
	}
}

func TestInsecureWarningBilingual(t *testing.T) {
	l := NewLocale()

	l.Set(ZH)
	zhText := l.T("options.insecure_warn")
	if zhText == "" || zhText == "options.insecure_warn" {
		t.Fatalf("zh-CN insecure warning not translated: %q", zhText)
	}

	l.Set(EN)
	enText := l.T("options.insecure_warn")
	if enText == "" || enText == "options.insecure_warn" {
		t.Fatalf("en-US insecure warning not translated: %q", enText)
	}

	enShort := l.T("options.insecure_warn_short")
	if enShort == "" || enShort == "options.insecure_warn_short" {
		t.Fatalf("en-US short insecure warning not translated: %q", enShort)
	}
	if !strings.Contains(enShort, "Danger") || !strings.Contains(enShort, "OFF") {
		t.Fatalf("en-US short insecure warning must clearly communicate danger: %q", enShort)
	}
}

func TestPanelAndDialogKeysBilingual(t *testing.T) {
	keys := []string{
		"panel.hosts",
		"panel.operation",
		"panel.results",
		"operation.quick_command",
		"operation.file_transfer",
		"operation.upload_file",
		"operation.download_file",
		"dialog.add_host.title",
		"dialog.import_hosts.title",
		"dialog.import_hosts.help",
		"dialog.close",
		"hosts.empty",
		"hosts.total",
		"hosts.enabled_count",
		"hosts.selected_count",
		"results.current_operation",
		"results.total_hosts",
		"results.success_count",
		"results.failure_count",
		"results.running_count",
		"results.error_kind",
		"results.error_message",
		"results.copy",
		"results.view",
		"results.copy_current",
		"results.copy_all",
		"transfer.browse",
		"transfer.download_note",
		"cmd.nonzero_note",
		"cmd.run_one",
		"cmd.run_many",
		"transfer.upload_btn_one",
		"transfer.upload_btn_many",
		"transfer.download_btn_one",
		"transfer.download_btn_many",
		"options.insecure_warn_short",
		"options.run_group",
		"options.ssh_group",
		"options.ui_group",
		"status.no_results",
		"status.running",
		"status.last_operation",
		"status.total",
		"status.success",
		"status.failed",
		"inspector.title",
		"validation.ip_required",
		"validation.port_number",
		"validation.port_range",
		"validation.username_required",
		"validation.password_required",
		"validation.csv_empty",
		"validation.csv_fields",
		"validation.csv_ip_required",
		"validation.csv_port_number",
		"validation.csv_port_range",
		"validation.csv_username_required",
		"validation.csv_password_required",
		"validation.local_file_required",
		"validation.local_file_not_found",
		"validation.remote_file_required",
		"validation.local_dir_required",
		"validation.local_dir_expected",
		"validation.directory_upload_unsupported",
		"validation.directory_remote_unsupported",
	}

	l := NewLocale()
	for _, lang := range []Lang{ZH, EN} {
		l.Set(lang)
		for _, key := range keys {
			t.Run(string(lang)+"/"+key, func(t *testing.T) {
				got := l.T(key)
				if got == "" || got == key {
					t.Fatalf("missing translation for %s in %s: %q", key, lang, got)
				}
			})
		}
	}
}

func TestTranslationKeyParity(t *testing.T) {
	for key := range zhMessages {
		if _, ok := enMessages[key]; !ok {
			t.Fatalf("missing en-US translation for key %q", key)
		}
	}
	for key := range enMessages {
		if _, ok := zhMessages[key]; !ok {
			t.Fatalf("missing zh-CN translation for key %q", key)
		}
	}
}
