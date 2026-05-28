package config

import (
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := Default()

	if opts.Timeout != 30 {
		t.Errorf("expected default timeout 30, got %d", opts.Timeout)
	}
	if opts.Concurrency != 4 {
		t.Errorf("expected default concurrency 4, got %d", opts.Concurrency)
	}
	if opts.InsecureIgnoreHostKey {
		t.Error("InsecureIgnoreHostKey must default to false")
	}
	if opts.KnownHostsPath != "" {
		t.Errorf("expected empty KnownHostsPath, got %s", opts.KnownHostsPath)
	}
}

func TestConcurrencyOptions(t *testing.T) {
	opts := ConcurrencyOptions()
	expected := []int{1, 2, 4, 8, 16, 32}
	if len(opts) != len(expected) {
		t.Errorf("expected %d options, got %d", len(expected), len(opts))
	}
	for i, v := range expected {
		if opts[i] != v {
			t.Errorf("expected option %d to be %d, got %d", i, v, opts[i])
		}
	}
}

func TestRuntimeOptionsMutable(t *testing.T) {
	opts := Default()
	opts.InsecureIgnoreHostKey = true
	opts.Concurrency = 8

	// Verify that modifying a returned struct doesn't affect a new Default()
	opts2 := Default()
	if opts2.InsecureIgnoreHostKey {
		t.Error("Default() should still return false after modifying a copy")
	}
	if opts2.Concurrency != 4 {
		t.Error("Default() should still return 4 after modifying a copy")
	}
}
