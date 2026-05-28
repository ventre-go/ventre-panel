# Usage Guide — ventre-panel v0.1.0

## Overview

ventre-panel is a stateless desktop GUI for batch SSH operations. It does not save any data between sessions.

The main window opens as a resizable Workbench at `1440 x 900` Fyne logical units. It is not full screen and is not maximized.

- **Top toolbar**: language, operation timeout, host concurrency, known_hosts path, and host key verification override
- **Left Hosts panel**: add/import/test/manage in-memory SSH hosts
- **Right Operation panel**: Quick Command, Upload File, and Download File
- **Bottom Status bar**: current/last operation summary and View Results

Full stdout, stderr, and error details are shown in the separate Result Inspector window, not permanently in the main Workbench.

## Adding Hosts

### Method 1: Manual Entry

1. In the left **Hosts** panel, click **Add Host**
2. Fill in: IP, Port (default 22), Username, Password
3. Click **Add**

The host appears in the table. The password is not shown in the host table.

### Method 2: CSV Import

1. In the left **Hosts** panel, click **Import Hosts**
2. Paste CSV-formatted hosts in the dialog:

```
ip,port,username,password
10.0.0.1,22,root,password1
10.0.0.2,22,root,password2
10.0.0.3,2222,ubuntu,password3
```

3. Click **Import**

CSV rules:
- Empty lines are skipped
- Port can be empty (defaults to 22)
- IP, username, and password are required
- Non-numeric port triggers an error

### Managing Hosts

- **Click** a host row in the table to toggle enabled/disabled
- Enabled rows are the selected hosts used by Quick Command, Upload File, and Download File
- **Test Enabled Hosts** runs an SSH connection test for enabled hosts
- **Remove Disabled Hosts** removes hosts that are currently disabled
- **Clear All** removes all hosts

## Executing Commands

1. Ensure you have enabled hosts in the left Hosts panel
2. Open **Quick Command** in the right Operation panel
3. Enter a shell command, e.g., `echo hello`
4. Set **Host concurrency** (1, 2, 4, 8, 16, 32) in the top toolbar — default 4
5. Set **Operation timeout(s)** (10, 30, 60, 120, 300) in the top toolbar — default 30
6. Click **Run on Selected Hosts**

Results update the bottom Status bar and open in the Result Inspector:
- Host IP:Port (username)
- Status (ok or exit=N or error)
- Exit code
- stdout
- stderr
- Error kind and description (if any)
- Duration

If stdout, stderr, or an error message contains the configured SSH password,
ventre-panel displays `(redacted)` instead of the password.

## Understanding Non-Zero Exit Code

A command that exits with a non-zero status (e.g., `exit 42`) is **not** a transport error.

Example: Running `sh -c 'echo fail >&2; exit 42'`

Results show:
- Status: `exit=42`
- ExitCode: 42
- Success: false
- stderr: `fail`
- No error kind (it's not an error, just a failed command)

Transport errors (like auth_failed, connect_failed) will have an ErrorKind displayed.

## Understanding ErrorKind

| ErrorKind | Display |
|-----------|---------|
| auth_failed | Authentication failed |
| connect_failed | Connection failed |
| host_key_failed | Host key verification failed |
| timeout | Operation timed out |
| already_exists | Target already exists |
| not_found | File or path not found |
| permission_denied | Permission denied |
| transfer_failed | File transfer failed |
| canceled | Canceled |
| closed | Connection closed |
| unsupported | Unsupported operation |

## Uploading a Single File

1. Ensure you have enabled hosts
2. Open **Upload File** in the right Operation panel
3. Enter **Local file path**
4. Enter **Remote destination path**
5. Check **Overwrite** if you want to overwrite existing files
6. Click **Upload to Selected Hosts**

Results show per-host upload status and duration.

Note: only single-file upload is supported. Directory upload is not supported.

## Downloading a Single File

1. Ensure you have enabled hosts
2. Open **Download File** in the right Operation panel
3. Enter **Remote file path** (must be the same path on all hosts)
4. Enter **Local output directory**
5. Check **Overwrite** if you want to overwrite existing files
6. Click **Download from Selected Hosts**

Files are saved to: `<output-dir>/<ip>/<filename>`

For example, downloading `/var/log/app.log` to `./downloads`:
- `./downloads/10.0.0.1/app.log`
- `./downloads/10.0.0.2/app.log`

This isolates files by host and prevents conflicts.

## Using known_hosts

By default, the system `~/.ssh/known_hosts` is used. To use a different file:

1. Use the **known_hosts** input in the top toolbar
2. Enter the custom known_hosts path
3. Leave empty to use the default

## Viewing Results

The bottom Status bar stays lightweight. It shows:

- Current operation type
- Total hosts in the current operation
- Success, failure, and running counts
- **View Results** to reopen the latest Result Inspector

The Result Inspector opens automatically after command execution, upload, download, or connection testing completes. It shows a per-host result table and detail tabs for stdout, stderr, and error message.

Select a result row to inspect stdout, stderr, and error details. **Copy Current Result** copies the selected result to the clipboard after password redaction. **Copy All Results** copies every in-memory result after the same redaction. **Clear Results** clears only in-memory results.

## Switching Language

Use the language selector in the top toolbar. Switching language updates the UI text without clearing hosts or results. The language setting is not saved and resets to zh-CN on restart.

## Why No Data Is Saved

ventre-panel is designed for ad-hoc, temporary batch operations. This means:

- You re-enter hosts each session (no saved hosts)
- You re-enter commands each time (no history)
- Passwords are never stored (no saved credentials)
- Configuration is not saved (fresh defaults each launch)

This is a security feature, not a missing feature.

## Why No Private Key Auth in v0.1.0

v0.1.0 focuses on the simplest possible SSH authentication: password. Private key support adds complexity around:

- Key file selection UI
- Passphrase handling
- Key format support (RSA, ECDSA, Ed25519)
- Security considerations for key storage

Private key auth is outside the v0.1.0 product boundary.

## Why No Directory Transfer in v0.1.0

v0.1.0 supports only single-file transfer. Directory transfer requires:

- Recursive file walking
- Directory structure preservation
- Progress tracking across multiple files
- Conflict handling per file

Directory transfer is outside the v0.1.0 product boundary.

## Top Toolbar Run Options

All run options are in-memory only and reset on app restart:

- **Operation timeout(s)**: Command, connection test, upload, and download timeout in seconds
- **Host concurrency**: Number of simultaneous host operations (1-32)
- **Known Hosts Path**: Custom known_hosts file path
- **Host key verification override**: Disable host key verification for this run (dangerous)
