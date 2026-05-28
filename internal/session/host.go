// Package session holds the in-memory session state for the current run.
// Nothing is persisted.
package session

import "time"

// HostTarget is the minimum set of fields needed to connect to an SSH host.
type HostTarget struct {
	IP       string
	Port     int
	Username string
	Password string
}

// HostRow is a HostTarget with runtime UI state.
type HostRow struct {
	Target  HostTarget
	Enabled bool
	Status  string
}

// Operation describes what kind of operation was performed.
type Operation string

const (
	OpCommand        Operation = "command"
	OpTestConnection Operation = "test_connection"
	OpUpload         Operation = "upload"
	OpDownload       Operation = "download"
)

// HostResult holds the result of an operation on a single host.
// It must never contain a Password.
type HostResult struct {
	TargetIP string
	Port     int
	Username string

	Operation Operation
	Status    string

	ExitCode int
	Success  bool

	Stdout string
	Stderr string

	ErrorKind    string
	ErrorMessage string

	StartedAt  time.Time
	FinishedAt time.Time
	Duration   time.Duration
}
