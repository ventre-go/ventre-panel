package validation

import (
	"testing"
)

func TestValidateHostTarget(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		port        int
		username    string
		password    string
		wantErr     bool
		errContains string
	}{
		{"valid", "10.0.0.1", 22, "root", "secret", false, ""},
		{"empty IP", "", 22, "root", "secret", true, "IP must not be empty"},
		{"port too low", "10.0.0.1", 0, "root", "secret", true, "port must be between 1 and 65535"},
		{"port too high", "10.0.0.1", 99999, "root", "secret", true, "port must be between 1 and 65535"},
		{"empty username", "10.0.0.1", 22, "", "secret", true, "username must not be empty"},
		{"empty password", "10.0.0.1", 22, "root", "", true, "password must not be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostTarget(tt.ip, tt.port, tt.username, tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestParseHostsCSV(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantHosts int
		wantErrs  int
	}{
		{
			name:      "single valid line",
			input:     "10.0.0.1,22,root,secret",
			wantHosts: 1,
			wantErrs:  0,
		},
		{
			name:      "default port",
			input:     "10.0.0.1,,root,secret",
			wantHosts: 1,
			wantErrs:  0,
		},
		{
			name:      "multiple hosts",
			input:     "10.0.0.1,22,root,pass1\n10.0.0.2,22,root,pass2\n10.0.0.3,2222,ubuntu,pass3",
			wantHosts: 3,
			wantErrs:  0,
		},
		{
			name:      "skip empty lines",
			input:     "10.0.0.1,22,root,pass1\n\n\n10.0.0.2,22,root,pass2",
			wantHosts: 2,
			wantErrs:  0,
		},
		{
			name:      "empty IP error",
			input:     ",22,root,pass",
			wantHosts: 0,
			wantErrs:  1,
		},
		{
			name:      "non-numeric port error",
			input:     "10.0.0.1,abc,root,pass",
			wantHosts: 0,
			wantErrs:  1,
		},
		{
			name:      "empty username error",
			input:     "10.0.0.1,22,,pass",
			wantHosts: 0,
			wantErrs:  1,
		},
		{
			name:      "empty password error",
			input:     "10.0.0.1,22,root,",
			wantHosts: 0,
			wantErrs:  1,
		},
		{
			name:      "insufficient fields",
			input:     "10.0.0.1",
			wantHosts: 0,
			wantErrs:  1,
		},
		{
			name:      "too many fields",
			input:     "10.0.0.1,22,root,pass,extra",
			wantHosts: 0,
			wantErrs:  1,
		},
		{
			name:      "port out of range",
			input:     "10.0.0.1,99999,root,pass",
			wantHosts: 0,
			wantErrs:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hosts, errs := ParseHostsCSV(tt.input)
			if len(hosts) != tt.wantHosts {
				t.Errorf("got %d hosts, want %d", len(hosts), tt.wantHosts)
			}
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d: %v", len(errs), tt.wantErrs, errs)
			}
		})
	}
}

func TestParseHostsCSVDefaultPort(t *testing.T) {
	hosts, errs := ParseHostsCSV("10.0.0.1,,root,pass")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].Port != 22 {
		t.Errorf("expected default port 22, got %d", hosts[0].Port)
	}
}

func TestParseHostsCSVQuotedPasswordWithComma(t *testing.T) {
	hosts, errs := ParseHostsCSV(`10.0.0.1,22,root,"pa,ss"`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].Password != "pa,ss" {
		t.Fatalf("expected quoted password with comma to parse, got %q", hosts[0].Password)
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		port    string
		wantErr bool
	}{
		{"", false},
		{"22", false},
		{"65535", false},
		{"0", true},
		{"65536", true},
		{"abc", true},
		{"-1", true},
	}

	for _, tt := range tests {
		t.Run("port="+tt.port, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
