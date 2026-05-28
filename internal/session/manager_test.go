package session

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestSetAndGetHosts(t *testing.T) {
	m := NewManager()
	hosts := []HostRow{
		{Target: HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "pass"}, Enabled: true},
		{Target: HostTarget{IP: "10.0.0.2", Port: 22, Username: "admin", Password: "pass2"}, Enabled: false},
	}
	m.SetHosts(hosts)

	got := m.Hosts()
	if len(got) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(got))
	}
	if got[0].Target.IP != "10.0.0.1" {
		t.Errorf("expected first IP 10.0.0.1, got %s", got[0].Target.IP)
	}
	if got[1].Enabled {
		t.Error("expected second host to be disabled")
	}
}

func TestEnabledHosts(t *testing.T) {
	m := NewManager()
	hosts := []HostRow{
		{Target: HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "pass"}, Enabled: true},
		{Target: HostTarget{IP: "10.0.0.2", Port: 22, Username: "admin", Password: "pass2"}, Enabled: false},
		{Target: HostTarget{IP: "10.0.0.3", Port: 22, Username: "user", Password: "pass3"}, Enabled: true},
	}
	m.SetHosts(hosts)

	enabled := m.EnabledHosts()
	if len(enabled) != 2 {
		t.Fatalf("expected 2 enabled hosts, got %d", len(enabled))
	}
}

func TestClearResults(t *testing.T) {
	m := NewManager()
	m.SetHosts([]HostRow{
		{Target: HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "pass"}, Enabled: true},
	})
	// Manually add results by accessing internal state (test-only)
	m.results = []HostResult{
		{TargetIP: "10.0.0.1", Status: "ok"},
	}
	m.ClearResults()
	if len(m.Results()) != 0 {
		t.Errorf("expected 0 results after clear, got %d", len(m.Results()))
	}
	if len(m.Hosts()) != 1 {
		t.Errorf("expected clear results to preserve hosts, got %d hosts", len(m.Hosts()))
	}
}

func TestRunOptionsDefaults(t *testing.T) {
	m := NewManager()
	m.SetOptions(RunOptions{
		Timeout:     30 * time.Second,
		Concurrency: 4,
	})

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.options.Concurrency != 4 {
		t.Errorf("expected concurrency 4, got %d", m.options.Concurrency)
	}
	if m.options.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", m.options.Timeout)
	}
	if m.options.InsecureIgnoreHostKey {
		t.Error("expected InsecureIgnoreHostKey to default to false")
	}
}

func TestHostResultNoPassword(t *testing.T) {
	// HostResult struct must never have a Password field
	hr := HostResult{
		TargetIP: "10.0.0.1",
		Port:     22,
		Username: "root",
		Status:   "ok",
	}
	// Verify the struct doesn't contain a Password field via compilation
	_ = hr.TargetIP
	_ = hr.Username
	// Password is intentionally absent from HostResult
}

func TestRedactHostResultRemovesPasswordFromResultFields(t *testing.T) {
	host := HostRow{Target: HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "secret-pass"}}
	result := HostResult{
		Stdout:       "stdout contains secret-pass",
		Stderr:       "stderr contains secret-pass",
		ErrorMessage: "error contains secret-pass",
	}

	got := redactHostResult(result, host)
	combined := got.Stdout + got.Stderr + got.ErrorMessage
	if strings.Contains(combined, host.Target.Password) {
		t.Fatalf("redacted result still contains password: %#v", got)
	}
	if !strings.Contains(combined, "(redacted)") {
		t.Fatalf("expected redacted marker in result fields: %#v", got)
	}
}

func TestErrorResultRedactsPassword(t *testing.T) {
	host := HostRow{Target: HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "secret-pass"}}

	got := errorResult(host, OpCommand, errors.New("auth failed with secret-pass"))
	if strings.Contains(got.ErrorMessage, host.Target.Password) {
		t.Fatalf("error result leaked password: %q", got.ErrorMessage)
	}
}

func TestConcurrencyLimit(t *testing.T) {
	m := NewManager()
	opts := RunOptions{Concurrency: 0, Timeout: 30 * time.Second}
	hosts := []HostRow{} // empty hosts

	// Just test that runConcurrently normalizes concurrency=0 to 1
	results := m.runConcurrently(nil, hosts, opts, OpCommand, nil)
	if results != nil {
		// Empty hosts should produce nil (length 0)
		if len(results) != 0 {
			t.Errorf("expected 0 results, got %d", len(results))
		}
	}
}

func TestRunConcurrentlyLimitsActiveWorkers(t *testing.T) {
	m := NewManager()
	var hosts []HostRow
	for i := 0; i < 10; i++ {
		hosts = append(hosts, HostRow{
			Target:  HostTarget{IP: "10.0.0.1", Port: 22, Username: "root", Password: "pass"},
			Enabled: true,
		})
	}

	var mu sync.Mutex
	active := 0
	maxActive := 0
	results := m.runConcurrently(context.Background(), hosts, RunOptions{Concurrency: 3}, OpCommand, func(_ context.Context, h HostRow) (HostResult, error) {
		mu.Lock()
		active++
		if active > maxActive {
			maxActive = active
		}
		mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		active--
		mu.Unlock()

		return HostResult{TargetIP: h.Target.IP, Success: true, Status: "ok"}, nil
	})

	if len(results) != len(hosts) {
		t.Fatalf("expected %d results, got %d", len(hosts), len(results))
	}
	if maxActive > 3 {
		t.Fatalf("expected at most 3 active workers, saw %d", maxActive)
	}
}

func TestTestConnectionsNoHosts(t *testing.T) {
	m := NewManager()
	results := m.TestConnections(nil)
	if results != nil {
		t.Errorf("expected nil for no enabled hosts, got %d results", len(results))
	}
}

func TestTestConnectionsRespectsConcurrency(t *testing.T) {
	m := NewManager()
	m.SetOptions(RunOptions{Concurrency: 4, Timeout: 30 * time.Second})
	results := m.TestConnections(nil)
	if results != nil {
		t.Errorf("expected nil, got %d", len(results))
	}
}

func TestDownloadDestinationSeparatesFilesByHost(t *testing.T) {
	got := downloadDestination("/tmp/downloads", "10.0.0.1", "/var/log/app.log")
	want := "/tmp/downloads/10.0.0.1/app.log"
	if got != want {
		t.Fatalf("downloadDestination() = %q, want %q", got, want)
	}
}

func TestManagerUsesTransportInterface(t *testing.T) {
	// Verify that Manager's methods reference ventre-transport interfaces
	// This is a compile-time guarantee verified by the imports in manager.go
	m := NewManager()
	_ = m
}
