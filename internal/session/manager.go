package session

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/ventre-go/ventre-transport"
	transportssh "github.com/ventre-go/ventre-transport/ssh"

	"github.com/ventre-go/ventre-panel/internal/secure"
)

// Manager orchestrates batch operations across hosts using ventre-transport.
// It holds no persistent state.
type Manager struct {
	mu      sync.Mutex
	hosts   []HostRow
	results []HostResult
	options RunOptions
}

// RunOptions configures a batch execution run.
type RunOptions struct {
	Timeout               time.Duration
	Concurrency           int
	KnownHostsPath        string
	InsecureIgnoreHostKey bool
}

// NewManager creates a new session manager.
func NewManager() *Manager {
	return &Manager{}
}

// SetHosts replaces the current host list.
func (m *Manager) SetHosts(hosts []HostRow) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hosts = hosts
}

// Hosts returns the current host list.
func (m *Manager) Hosts() []HostRow {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]HostRow, len(m.hosts))
	copy(out, m.hosts)
	return out
}

// SetOptions configures the run options for the next operation.
func (m *Manager) SetOptions(opts RunOptions) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.options = opts
}

// Results returns all accumulated results.
func (m *Manager) Results() []HostResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]HostResult, len(m.results))
	copy(out, m.results)
	return out
}

// ClearResults removes all results from memory.
func (m *Manager) ClearResults() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results = nil
}

// EnabledHosts returns only the enabled hosts.
func (m *Manager) EnabledHosts() []HostRow {
	m.mu.Lock()
	defer m.mu.Unlock()
	var enabled []HostRow
	for _, h := range m.hosts {
		if h.Enabled {
			enabled = append(enabled, h)
		}
	}
	return enabled
}

func redactHostResult(result HostResult, h HostRow) HostResult {
	secrets := []string{h.Target.Password}
	result.Stdout = secure.RedactSecrets(result.Stdout, secrets)
	result.Stderr = secure.RedactSecrets(result.Stderr, secrets)
	result.ErrorMessage = secure.SanitizeErrorForDisplay(result.ErrorMessage, secrets)
	return result
}

// ExecCommands runs a command across all enabled hosts with the configured concurrency.
func (m *Manager) ExecCommands(ctx context.Context, command string) []HostResult {
	enabled := m.EnabledHosts()
	if len(enabled) == 0 {
		return nil
	}

	m.mu.Lock()
	opts := m.options
	m.mu.Unlock()

	return m.runConcurrently(ctx, enabled, opts, OpCommand, func(ctx context.Context, h HostRow) (HostResult, error) {
		return execOnHost(ctx, h, opts, command)
	})
}

// UploadFiles uploads a local file to all enabled hosts.
func (m *Manager) UploadFiles(ctx context.Context, localPath, remotePath string, overwrite bool) []HostResult {
	enabled := m.EnabledHosts()
	if len(enabled) == 0 {
		return nil
	}

	m.mu.Lock()
	opts := m.options
	m.mu.Unlock()

	return m.runConcurrently(ctx, enabled, opts, OpUpload, func(ctx context.Context, h HostRow) (HostResult, error) {
		return uploadToHost(ctx, h, opts, localPath, remotePath, overwrite)
	})
}

// DownloadFiles downloads a remote file from all enabled hosts.
func (m *Manager) DownloadFiles(ctx context.Context, remotePath, localDir string, overwrite bool) []HostResult {
	enabled := m.EnabledHosts()
	if len(enabled) == 0 {
		return nil
	}

	m.mu.Lock()
	opts := m.options
	m.mu.Unlock()

	return m.runConcurrently(ctx, enabled, opts, OpDownload, func(ctx context.Context, h HostRow) (HostResult, error) {
		return downloadFromHost(ctx, h, opts, remotePath, localDir, overwrite)
	})
}

func (m *Manager) runConcurrently(ctx context.Context, hosts []HostRow, opts RunOptions, op Operation, fn func(context.Context, HostRow) (HostResult, error)) []HostResult {
	concurrency := opts.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}
	if concurrency > 32 {
		concurrency = 32
	}

	if concurrency > len(hosts) {
		concurrency = len(hosts)
	}

	jobs := make(chan int)
	results := make([]HostResult, len(hosts))
	var wg sync.WaitGroup

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for idx := range jobs {
				h := hosts[idx]
				result, err := fn(ctx, h)
				if err != nil {
					result = errorResult(h, op, err)
				}
				results[idx] = result
			}
		}()
	}

	for i := range hosts {
		jobs <- i
	}
	close(jobs)

	wg.Wait()

	m.mu.Lock()
	m.results = append(m.results, results...)
	m.mu.Unlock()

	return results
}

func execOnHost(ctx context.Context, h HostRow, opts RunOptions, command string) (HostResult, error) {
	t, err := transportssh.New(transportssh.Config{
		Host:                  h.Target.IP,
		Port:                  h.Target.Port,
		User:                  h.Target.Username,
		Password:              h.Target.Password,
		KnownHostsPath:        opts.KnownHostsPath,
		InsecureIgnoreHostKey: opts.InsecureIgnoreHostKey,
		Timeout:               opts.Timeout,
	})
	if err != nil {
		return HostResult{}, err
	}
	defer t.Close(ctx)

	startedAt := time.Now()
	result, err := t.Exec(ctx, transport.ExecRequest{
		Command: transport.Shell(command),
		Timeout: opts.Timeout,
	})
	finishedAt := time.Now()

	hr := HostResult{
		TargetIP:  h.Target.IP,
		Port:      h.Target.Port,
		Username:  h.Target.Username,
		Operation: OpCommand,
	}

	if result != nil {
		hr.Stdout = string(result.Stdout)
		hr.Stderr = string(result.Stderr)
		hr.ExitCode = result.ExitCode
		hr.Success = result.Success
		hr.StartedAt = result.StartedAt
		hr.FinishedAt = result.FinishedAt
		hr.Duration = result.Duration
	}

	if err != nil {
		kind := transport.KindOf(err)
		hr.ErrorKind = string(kind)
		hr.ErrorMessage = err.Error()
		hr.Status = "error"
	} else if !hr.Success {
		hr.Status = fmt.Sprintf("exit=%d", hr.ExitCode)
	} else {
		hr.Status = "ok"
	}

	if startedAt.IsZero() {
		hr.StartedAt = startedAt
	}
	if finishedAt.IsZero() {
		hr.FinishedAt = finishedAt
	}
	if hr.Duration == 0 {
		hr.Duration = finishedAt.Sub(startedAt)
	}

	return redactHostResult(hr, h), err
}

func uploadToHost(ctx context.Context, h HostRow, opts RunOptions, localPath, remotePath string, overwrite bool) (HostResult, error) {
	t, err := transportssh.New(transportssh.Config{
		Host:                  h.Target.IP,
		Port:                  h.Target.Port,
		User:                  h.Target.Username,
		Password:              h.Target.Password,
		KnownHostsPath:        opts.KnownHostsPath,
		InsecureIgnoreHostKey: opts.InsecureIgnoreHostKey,
		Timeout:               opts.Timeout,
	})
	if err != nil {
		return HostResult{}, err
	}
	defer t.Close(ctx)

	startedAt := time.Now()
	result, err := t.Upload(ctx, transport.UploadRequest{
		Source:        localPath,
		Destination:   remotePath,
		CreateParents: true,
		Overwrite:     overwrite,
	})
	finishedAt := time.Now()

	hr := HostResult{
		TargetIP:  h.Target.IP,
		Port:      h.Target.Port,
		Username:  h.Target.Username,
		Operation: OpUpload,
	}

	if result != nil {
		hr.StartedAt = result.StartedAt
		hr.FinishedAt = result.FinishedAt
		hr.Duration = result.Duration
		hr.Success = true
		hr.Status = "ok"
	}

	if err != nil {
		kind := transport.KindOf(err)
		hr.ErrorKind = string(kind)
		hr.ErrorMessage = err.Error()
		hr.Status = "error"
		hr.Success = false
	}

	if startedAt.IsZero() {
		hr.StartedAt = startedAt
	}
	if finishedAt.IsZero() {
		hr.FinishedAt = finishedAt
	}
	if hr.Duration == 0 {
		hr.Duration = finishedAt.Sub(startedAt)
	}

	return redactHostResult(hr, h), err
}

func downloadFromHost(ctx context.Context, h HostRow, opts RunOptions, remotePath, localDir string, overwrite bool) (HostResult, error) {
	t, err := transportssh.New(transportssh.Config{
		Host:                  h.Target.IP,
		Port:                  h.Target.Port,
		User:                  h.Target.Username,
		Password:              h.Target.Password,
		KnownHostsPath:        opts.KnownHostsPath,
		InsecureIgnoreHostKey: opts.InsecureIgnoreHostKey,
		Timeout:               opts.Timeout,
	})
	if err != nil {
		return HostResult{}, err
	}
	defer t.Close(ctx)

	localDest := downloadDestination(localDir, h.Target.IP, remotePath)

	startedAt := time.Now()
	result, err := t.Download(ctx, transport.DownloadRequest{
		Source:        remotePath,
		Destination:   localDest,
		CreateParents: true,
		Overwrite:     overwrite,
	})
	finishedAt := time.Now()

	hr := HostResult{
		TargetIP:  h.Target.IP,
		Port:      h.Target.Port,
		Username:  h.Target.Username,
		Operation: OpDownload,
	}

	if result != nil {
		hr.StartedAt = result.StartedAt
		hr.FinishedAt = result.FinishedAt
		hr.Duration = result.Duration
		hr.Success = true
		hr.Status = "ok"
	}

	if err != nil {
		kind := transport.KindOf(err)
		hr.ErrorKind = string(kind)
		hr.ErrorMessage = err.Error()
		hr.Status = "error"
		hr.Success = false
	}

	if startedAt.IsZero() {
		hr.StartedAt = startedAt
	}
	if finishedAt.IsZero() {
		hr.FinishedAt = finishedAt
	}
	if hr.Duration == 0 {
		hr.Duration = finishedAt.Sub(startedAt)
	}

	return redactHostResult(hr, h), err
}

func downloadDestination(localDir, targetIP, remotePath string) string {
	if localDir == "" {
		return localDir
	}
	return filepath.Join(localDir, targetIP, path.Base(remotePath))
}

// TestConnections tests SSH connectivity for all enabled hosts without executing any command.
func (m *Manager) TestConnections(ctx context.Context) []HostResult {
	enabled := m.EnabledHosts()
	if len(enabled) == 0 {
		return nil
	}

	m.mu.Lock()
	opts := m.options
	m.mu.Unlock()

	return m.runConcurrently(ctx, enabled, opts, OpTestConnection, func(ctx context.Context, h HostRow) (HostResult, error) {
		return testConnection(ctx, h, opts)
	})
}

func testConnection(ctx context.Context, h HostRow, opts RunOptions) (HostResult, error) {
	t, err := transportssh.New(transportssh.Config{
		Host:                  h.Target.IP,
		Port:                  h.Target.Port,
		User:                  h.Target.Username,
		Password:              h.Target.Password,
		KnownHostsPath:        opts.KnownHostsPath,
		InsecureIgnoreHostKey: opts.InsecureIgnoreHostKey,
		Timeout:               opts.Timeout,
	})
	if err != nil {
		return HostResult{}, err
	}

	startedAt := time.Now()
	// Just connect and immediately close — verifies auth and connectivity
	closeErr := t.Close(ctx)
	finishedAt := time.Now()

	hr := HostResult{
		TargetIP:   h.Target.IP,
		Port:       h.Target.Port,
		Username:   h.Target.Username,
		Operation:  OpTestConnection,
		Success:    true,
		Status:     "ok",
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		Duration:   finishedAt.Sub(startedAt),
	}

	if closeErr != nil {
		kind := transport.KindOf(closeErr)
		hr.ErrorKind = string(kind)
		hr.ErrorMessage = closeErr.Error()
		hr.Status = "error"
		hr.Success = false
	}

	return redactHostResult(hr, h), nil
}

func errorResult(h HostRow, op Operation, err error) HostResult {
	kind := transport.KindOf(err)
	return redactHostResult(HostResult{
		TargetIP:     h.Target.IP,
		Port:         h.Target.Port,
		Username:     h.Target.Username,
		Operation:    op,
		Status:       "error",
		ErrorKind:    string(kind),
		ErrorMessage: err.Error(),
		Success:      false,
		ExitCode:     -1,
	}, h)
}
