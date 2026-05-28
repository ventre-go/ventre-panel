# ventre-panel

A stateless cross-platform desktop client for batch SSH command execution and file transfer, powered by [ventre-transport](https://github.com/ventre-go/ventre-transport).

English: A stateless cross-platform desktop client for batch SSH command execution and file transfer, powered by ventre-transport.

中文: 一个无状态跨平台桌面客户端，用于批量 SSH 命令执行和文件传输，底层由 ventre-transport 驱动。

## What It Is

ventre-panel is a lightweight GUI tool that lets you:

- Input a list of SSH hosts temporarily
- Execute shell commands in batch across enabled hosts
- Upload a single file to enabled hosts
- Download a single file from enabled hosts
- View per-host structured results

## What It Is NOT

- SSH config manager
- Terminal emulator
- SFTP file manager
- Task orchestration system
- Workflow engine
- Host asset management
- Password manager
- AI agent
- Monitoring system

## Platform Support

| Platform | Status |
|----------|--------|
| Linux    | Build tested (x86-64) |
| Windows  | Target support (not yet verified) |
| macOS    | Target support (not yet verified) |

## v0.1.0 Features

- Batch SSH host input via Add Host and CSV Import dialogs
- Password-based SSH authentication (only auth method in v0.1.0)
- Batch shell command execution
- Batch single-file upload
- Batch single-file download
- Persistent in-memory host panel
- Operation panel with Quick Command, Upload File, and Download File modes
- Bottom status bar with last-run summary and View Results
- Dedicated Result Inspector window for per-host result tables and stdout/stderr/error details
- Copy current or all results with password redaction
- Host concurrency limit (1-32)
- Password redaction in UI and results
- SSH host key verification enabled by default
- Chinese and English UI with runtime language switch
- Desktop-optimized Workbench layout (1440x900 default, 1280x800 minimum content size)
- All state is in-memory only — nothing is persisted, including language preference

## No Persistence

ventre-panel does NOT save:

- Hosts
- Passwords
- Command history
- File path history
- Execution results
- Configuration
- Logs
- UI state

When you close the app, everything is gone. This is intentional.

## Host Fields

Each SSH host has exactly 4 fields:

- IP
- Port (defaults to 22)
- Username
- Password

No other fields are supported in v0.1.0.

## v0.1.0 Limitations

- Only password authentication (no private key auth)
- Only single-file upload (no directory transfer)
- Only single-file download (no directory transfer)
- No host management or grouping
- No configuration persistence
- No history
- No cancel (running operations cannot be cancelled)
- No retry
- No connection pooling
- No multi-hop SSH

## CSV Host Format

Paste hosts in CSV format:

```
ip,port,username,password
10.0.0.1,22,root,password1
10.0.0.2,22,root,password2
10.0.0.3,2222,ubuntu,password3
```

Empty port defaults to 22.

## Workbench Main Window

The main window is a lightweight Workbench. It opens at `1440 x 900` Fyne logical units, is not full screen, is not maximized, and remains user-resizable. The minimum usable content size is `1280 x 800` logical units. Fyne and the operating system handle HiDPI scaling for 2K/4K displays.

- **Top toolbar**: language, operation timeout, host concurrency, known_hosts path, and host key verification override
- **Left Hosts panel**: in-memory host table without a password column, host counts, Add Host dialog, CSV Import dialog, connection test, delete disabled hosts, clear all
- **Right Operation panel**: Quick Command, Upload File, and Download File modes
- **Bottom Status bar**: current/last operation summary and View Results

The Add Host and CSV Import controls open dialogs. Their input is never saved to disk. The main window does not permanently display stdout, stderr, or error details.

## Result Inspector

Command, upload, download, and connection-test results are viewed in a dedicated Result Inspector window. It opens at `1200 x 780` logical units with a `960 x 640` minimum content size and can be resized by the user.

The Result Inspector shows:

- Operation summary
- Per-host result table
- Stdout, stderr, and error message detail tabs
- Copy Current Result and Copy All Results actions with password redaction
- Clear Results for in-memory results only

Closing the Result Inspector does not clear results. The Workbench Status bar keeps the summary, and **View Results** can reopen the latest in-memory results.

## Command Execution Example

1. Add or import hosts in the left Hosts panel
2. Test enabled hosts if needed
3. Enter: `echo hello`
4. Set host concurrency and operation timeout in the top toolbar
5. Click "Run on Selected Hosts"
6. Results appear in the Status bar summary and open in the Result Inspector with per-host stdout, stderr, exit code

## Non-Zero Exit Code Semantics

A non-zero exit code is NOT a transport error. ventre-transport distinguishes:

- **Transport error** (`err != nil`): auth failed, connect failed, timeout, host key failure
- **Command failure** (`err == nil`, `result.Success == false`): command ran but exited non-zero

The UI displays both cases clearly.

## File Upload Example

1. Add or import hosts in the left Hosts panel
2. Open Upload File in the Operation panel
3. Enter local file path
4. Enter remote destination path
5. Configure overwrite
6. Click "Upload to Selected Hosts"

## File Download Example

1. Add or import hosts in the left Hosts panel
2. Open Download File in the Operation panel
3. Enter remote file path
4. Enter local output directory
5. Files are saved as: `<output-dir>/<ip>/<filename>`
6. Configure overwrite
7. Click "Download from Selected Hosts"

## Directory Transfer

Directory transfer is not supported in v0.1.0. The UI explicitly notes this. Attempting to transfer a directory will result in an "unsupported" error.

## known_hosts

By default, ventre-panel uses the SSH backend default known_hosts behavior. You can specify a custom known_hosts path in the top toolbar. This setting applies only to the current run and is not persisted.

## Host Key Verification Override

In the top toolbar, there is an advanced option "Skip SSH host key verification for this run". This is:

- **Default: OFF**
- **Marked with a danger warning about MITM risk**
- **Not persisted** — resets on app restart
- Only for trusted, isolated networks

Do not enable this unless you understand the security implications.

## Language / 语言

ventre-panel supports Chinese (zh-CN) and English (en-US).

- **Default**: Chinese (zh-CN)
- **Switch**: top toolbar language selector
- **Switch is instant** — hosts and results are preserved
- **Language preference is NOT saved** — resets to zh-CN on restart

## Security

See [SECURITY.md](SECURITY.md) for full security documentation.

Key points:
- Passwords exist only in memory
- Passwords are masked in the UI
- Passwords are redacted from errors and results
- No local log files
- No persistence (including language preference)
- Host key verification enabled by default

## Build

### Native Build (Linux)

**Prerequisites:**
- Go 1.26+
- Fyne system dependencies: `libgl1-mesa-dev libxxf86vm-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev`

```bash
go build -o dist/ventre-panel ./cmd/ventre-panel
```

If you are building from a source archive or copied directory without valid Git
metadata, use:

```bash
go build -buildvcs=false -o dist/ventre-panel ./cmd/ventre-panel
```

### Cross-Platform Build (via fyne-cross)

fyne-cross uses Docker/Podman containers to cross-compile Fyne apps for all platforms.

**Prerequisites:**
- Docker or Podman (with podman-docker)
- `go install github.com/fyne-io/fyne-cross@latest`

```bash
# Linux
fyne-cross linux -app-id com.ventre-go.ventre-panel

# Windows
fyne-cross windows -app-id com.ventre-go.ventre-panel

# macOS (Intel + Apple Silicon)
fyne-cross darwin -app-id com.ventre-go.ventre-panel
```

Output packages are placed in the `fyne-cross/` directory.

### Manual Cross-Compilation (advanced)

Manual cross-compilation requires platform-specific C toolchains:

- **Windows**: mingw-w64 (`x86_64-w64-mingw32-gcc`)
- **macOS**: osxcross or macOS SDK

These are complex to set up. Use `fyne-cross` unless you have specific requirements.

### Run Tests

```bash
go test ./...
go test -race ./...
go vet ./...
```

## Dependency

ventre-panel depends on [github.com/ventre-go/ventre-transport](https://github.com/ventre-go/ventre-transport) for all SSH operations. It never directly imports `golang.org/x/crypto/ssh` or `github.com/pkg/sftp`.

## Known Limitations (v0.1.0)

- Only password auth (no private key)
- No directory transfer
- No persistence or profiles
- No terminal emulator
- No multi-hop SSH
- No connection pooling
- No auto-retry
- No cancel (running operations cannot be cancelled mid-flight)
- Windows/macOS not yet verified
- No results export

## License

See [LICENSE](LICENSE).
