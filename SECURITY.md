# Security

## Password Handling

- Passwords exist **only in memory** during the application runtime.
- Passwords are **masked** in the UI (password fields show `••••`).
- Passwords are **not displayed** in the main hosts table.
- Passwords are **never written to disk**, logs, config files, or results.
- Passwords are **redacted** from error messages and result fields, including
  stdout/stderr if remote output contains the configured password.
- Passwords are redacted again before a selected result is copied to the
  clipboard.
- Passwords are **not persisted** across application restarts.

## No Persistence Policy

ventre-panel intentionally does not persist any data:

- No hosts saved to disk
- No passwords saved (not to keychain, not to config)
- No command history
- No file path history
- No execution results saved
- No language preference saved
- No application preferences saved
- No log files written

All state exists only in memory and is lost when the application exits.

Add Host dialog input, CSV Import dialog input, file paths, Result Inspector
content, and result details are runtime UI state only. They are not written to
preferences, config files, logs, or any local database.

CSV Import validation reports row numbers and redacted field-level errors. It
does not display or log the raw CSV row, so passwords are not echoed back in
validation messages.

## Fyne Preferences and Storage

ventre-panel creates the Fyne application without a persistent application ID
and does not call Fyne Preferences or Storage APIs for application data. Hosts,
passwords, paths, command text, results, language selection, and window state
are not saved by ventre-panel.

## No Local Logs

ventre-panel does **not** write log files. All error and result display is in-memory and in-UI only.

## Language Preference Not Persisted

The language setting (zh-CN / en-US) is **not saved** to disk. It resets to the default (zh-CN) on every app restart. This is consistent with the stateless design of ventre-panel.

## Host Key Verification

SSH host key verification is **enabled by default**. The application uses:

- The system's `known_hosts` file (`~/.ssh/known_hosts`) by default
- A custom path if specified in the top toolbar

This prevents man-in-the-middle attacks on SSH connections.

## Host Key Verification Override

An advanced option exists to disable host key verification for the current run. This option:

- Is **disabled by default**
- Is **clearly marked with a danger warning** about MITM risk
- Is **not persisted** — it resets on every app restart
- Should **only** be used in trusted, isolated networks

Enabling this option makes SSH connections vulnerable to man-in-the-middle attacks.

## Clipboard Risk

The Result Inspector can copy the selected result or all in-memory results to
the clipboard. Before copying, ventre-panel applies the same password redaction
used for in-UI result display.
Clipboard content can still include host IPs, usernames, commands, paths, and
remote command output, so users should treat copied results as sensitive.

## Screenshots Risk

Users should be aware that screenshots of the application may contain:

- Host IP addresses
- Usernames
- Command text
- File paths
- Execution output

Passwords are masked in the UI, but other sensitive information (like command output) may be visible in screenshots.

## File Transfer Safety

- Upload and download paths are user-specified
- Overwrite behavior is user-controlled
- Downloads are isolated by host IP to prevent file conflicts

## No Private Key Support in v0.1.0

v0.1.0 only supports password-based authentication. Private key authentication is not available. This means:

- Users must use password auth
- Private key files are not read
- SSH agent is not used

Private key support is outside the v0.1.0 product boundary.

## Vulnerability Reporting

TBD — please check the repository for the latest security contact information.
