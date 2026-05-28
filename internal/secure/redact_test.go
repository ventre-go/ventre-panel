package secure

import (
	"strings"
	"testing"
)

func TestMaskPassword(t *testing.T) {
	tests := []struct {
		input    string
		wantMask string
	}{
		{"", ""},
		{"a", "•"},
		{"secret", "••••••"},
		{"password123", "•••••••••••"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := MaskPassword(tt.input)
			if got != tt.wantMask {
				t.Errorf("MaskPassword(%q) = %q, want %q", tt.input, got, tt.wantMask)
			}
		})
	}
}

func TestRedactSecrets(t *testing.T) {
	secrets := []string{"sekret123", "admin"}

	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "no secrets",
			s:    "hello world",
			want: "hello world",
		},
		{
			name: "single secret redacted",
			s:    "password is sekret123 here",
			want: "password is (redacted) here",
		},
		{
			name: "multiple occurrences",
			s:    "admin:admin login failed",
			want: "(redacted):(redacted) login failed",
		},
		{
			name: "empty string",
			s:    "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactSecrets(tt.s, secrets)
			if got != tt.want {
				t.Errorf("RedactSecrets(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestSanitizeErrorForDisplay(t *testing.T) {
	secrets := []string{"mypassword"}
	errMsg := "auth failed for user root with password mypassword"
	got := SanitizeErrorForDisplay(errMsg, secrets)
	if strings.Contains(got, "mypassword") {
		t.Errorf("error message still contains password: %s", got)
	}
	if !strings.Contains(got, "(redacted)") {
		t.Errorf("expected (redacted) in sanitized error: %s", got)
	}
}

func TestRedactSecretsEmptySecret(t *testing.T) {
	secrets := []string{"", "good"}
	got := RedactSecrets("hello good world", secrets)
	if got != "hello (redacted) world" {
		t.Errorf("got %q, want %q", got, "hello (redacted) world")
	}
}
