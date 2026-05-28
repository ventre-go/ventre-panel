# Changelog

All notable changes to ventre-panel will be documented in this file.

## v0.1.0 - Unreleased

### Added
- Stateless batch SSH host input via dialogs and CSV import
- Password-based SSH authentication
- Batch SSH command execution with per-host results
- Batch single-file upload to multiple hosts
- Batch single-file download from multiple hosts
- Per-host result view with stdout, stderr, exit code, duration
- Configurable concurrency limit (1-32)
- Configurable command timeout
- Configurable known_hosts path
- Password masking in UI
- Password redaction in errors and results
- Password redaction for stdout/stderr/error result fields
- Chinese and English UI with runtime language switch
- Desktop Workbench layout (1440x900 default, 1280x800 minimum)

### Changed
- Shortened English button and tab labels to prevent text overflow
- Reworked toolbar grouping for stable bilingual two-row layout  
- Increased main window default and minimum logical size for English text
- Stabilized layout during language switching with split-offset preservation
- Adjusted fixed-size wrappers to accommodate longer English labels
- Simplified status bar summary text for height stability
- Desktop-optimized Workbench layout for modern HiDPI desktop environments
- Custom theme with larger fonts and controls for readability
- ventre-transport integration for all SSH operations
- Fyne-based cross-platform GUI
- Redesigned main window layout for Hosts -> Operation -> Results flow
- Top toolbar for language, timeout, concurrency, known_hosts, and security option
- Persistent in-memory hosts panel with host counts and table
- Add Host dialog instead of always-visible host form
- CSV Import dialog instead of always-visible import text area
- Operation panel split into Quick Command, Upload File, and Download File modes
- Bottom Status Bar with last-run summary and View Results
- Dedicated Result Inspector window for command, connection test, upload, and download results
- Result Inspector sizing tuned for long stdout/stderr/error viewing
- Removed password column from the hosts table
- Copy current result and copy all results with password redaction
- Improved desktop interaction flow and large-window space usage
- GUI updates from background operations marshalled through Fyne runtime context
- Localized Add Host, CSV Import, and file path validation messages
- Runtime operation buttons show running state while work is in progress
- Main window now opens at 1280x760 logical units, remains resizable, and is not full screen by default

### Security
- No host/password persistence (all state in-memory only)
- No language preference persistence
- SSH host key verification enabled by default
- Password masked in UI (password entry widget)
- Password redacted from all error and result display
- Password redacted from stdout/stderr if remote output echoes it
- Host key verification override default off with danger warning
- No local log files
- No direct SSH/SFTP imports (all through ventre-transport)
- Dialog validation errors are redacted and do not include source CSV rows

### Known Limitations
- No private key authentication in v0.1.0
- No directory transfer in v0.1.0 (single files only)
- No persistence or profiles (including language preference)
- No terminal emulator
- No host grouping
- No connection pooling
- No auto-retry
- No cancel (running operations cannot be cancelled)
- Windows/macOS builds not yet verified
