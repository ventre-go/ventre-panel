// Package config holds runtime options that apply to the current session only.
// Nothing in this package is persisted.
package config

// RuntimeOptions holds global settings for the current run.
// These are never saved to disk.
type RuntimeOptions struct {
	Timeout               int    // seconds, 0 means use transport default
	Concurrency           int    // 1-32, default 4
	KnownHostsPath        string // empty means use transport default (~/.ssh/known_hosts)
	InsecureIgnoreHostKey bool   // default false, must show red warning in UI
}

// Default returns safe default runtime options.
func Default() RuntimeOptions {
	return RuntimeOptions{
		Timeout:               30,
		Concurrency:           4,
		KnownHostsPath:        "",
		InsecureIgnoreHostKey: false,
	}
}

// ConcurrencyOptions returns the allowed concurrency values for the UI.
func ConcurrencyOptions() []int {
	return []int{1, 2, 4, 8, 16, 32}
}
