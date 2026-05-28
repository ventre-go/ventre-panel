# Release Checklist — ventre-panel v0.1.0

## Pre-Release Verification

### Code Quality
- [ ] `gofmt -l .` returns no output
- [ ] `go test ./...` passes all tests
- [ ] `go test -race ./...` passes all tests (or environment limitation documented)
- [ ] `go vet ./...` returns no errors

### Build
- [ ] `go build -o dist/ventre-panel ./cmd/ventre-panel` succeeds on current platform
- [ ] If building outside a valid Git checkout, `go build -buildvcs=false -o dist/ventre-panel ./cmd/ventre-panel` succeeds
- [ ] Binary can be executed and displays GUI

### Dependency Check
- [ ] `go.mod` depends on `github.com/ventre-go/ventre-transport`
- [ ] No direct import of `golang.org/x/crypto/ssh`
- [ ] No direct import of `github.com/pkg/sftp`
- [ ] No local `replace` directive in `go.mod` (use tagged version)

### Manual SSH Tests
- [ ] Input at least 2 SSH hosts (or same host with different records)
- [ ] Batch execute `echo hello` — verify stdout per host
- [ ] Batch execute `sh -c 'echo fail >&2; exit 42'` — verify ExitCode=42, Success=false, stderr visible
- [ ] Upload a small file to `/tmp/ventre-panel-test/...` on each host
- [ ] Download the same file to local `<dir>/<ip>/` subdirectories
- [ ] Verify downloaded file content matches original
- [ ] Test with wrong password — verify `auth_failed` display
- [ ] Test with wrong known_hosts — verify `host_key_failed` display

### No Persistence Check
- [ ] Close and reopen app — verify host list is empty
- [ ] Close and reopen app — verify command field is empty
- [ ] Close and reopen app — verify results are empty
- [ ] Close and reopen app — verify top toolbar run options are back to defaults
- [ ] Check home directory — no `~/.ventre-panel` directory created
- [ ] Check home directory — no config files created

### Password Redaction Check
- [ ] Password masked in UI (password fields show dots/bullets)
- [ ] Password not visible in results display
- [ ] Password not visible in error messages
- [ ] Password not visible in stdout/stderr output

### v0.1.0 Feature Verification
- [ ] Private key auth is NOT present in UI
- [ ] Directory transfer is explicitly marked unsupported
- [ ] Only 4 host fields: IP, Port, Username, Password
- [ ] Host key verification override defaults to OFF with danger warning
- [ ] Concurrency limit works (test with 1 vs 4)

### Documentation Review
- [ ] README.md is accurate and complete
- [ ] SECURITY.md is accurate and complete
- [ ] CHANGELOG.md contains v0.1.0 Unreleased entries
- [ ] docs/usage.md exists and is complete
- [ ] Known limitations are honestly documented

### Final Checks
- [ ] No Blocker issues remain
- [ ] No unexplained High issues remain
- [ ] CI passes on GitHub (if configured)

## Post-Release (do NOT do automatically)

- Tag: `git tag v0.1.0`
- Push tag: `git push origin v0.1.0`
- Create GitHub Release with changelog
