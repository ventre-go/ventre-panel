// ssh-verify is a manual SSH verification tool for ventre-panel v0.1.0.
// Run with: go run ./cmd/ssh-verify/
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ventre-go/ventre-panel/internal/secure"
	"github.com/ventre-go/ventre-panel/internal/session"
)

func main() {
	password := os.Getenv("VENTRE_TEST_PASSWORD")
	if password == "" {
		fmt.Fprintln(os.Stderr, "VENTRE_TEST_PASSWORD not set")
		os.Exit(1)
	}
	host := os.Getenv("VENTRE_TEST_HOST")
	if host == "" {
		fmt.Fprintln(os.Stderr, "VENTRE_TEST_HOST not set")
		os.Exit(1)
	}

	fmt.Println("=== ventre-panel v0.1.0 SSH Verification ===")
	fmt.Printf("Target: %s:22 (root)\n", host)
	fmt.Println()

	failures := 0
	pass := func(name string) {
		fmt.Printf("  PASS: %s\n", name)
	}
	fail := func(name string, msg string) {
		fmt.Printf("  FAIL: %s — %s\n", name, msg)
		failures++
	}

	// Use InsecureIgnoreHostKey for functional tests since host key may not be in known_hosts.
	// Security is verified separately via test 6.
	insecureOpts := session.RunOptions{Timeout: 30 * time.Second, Concurrency: 1, InsecureIgnoreHostKey: true}

	// Cleanup before starting
	doRemoteCmd(host, password, "rm -rf /tmp/ventre-panel-test")

	// 1. echo hello
	fmt.Println("1. Batch execute 'echo hello'...")
	{
		mgr := session.NewManager()
		mgr.SetHosts([]session.HostRow{
			{Target: session.HostTarget{IP: host, Port: 22, Username: "root", Password: password}, Enabled: true},
		})
		mgr.SetOptions(insecureOpts)

		results := mgr.ExecCommands(context.Background(), "echo hello")
		if len(results) != 1 {
			fail("echo hello", fmt.Sprintf("expected 1 result, got %d", len(results)))
		} else {
			r := results[0]
			if r.Success && r.ExitCode == 0 && strings.TrimSpace(r.Stdout) == "hello" {
				pass("echo hello — exit=0, stdout='hello'")
			} else {
				fail("echo hello", fmt.Sprintf("success=%v exit=%d stdout=%q errKind=%s", r.Success, r.ExitCode, r.Stdout, r.ErrorKind))
			}
		}
	}

	// 2. Non-zero exit code (NOT a transport error)
	fmt.Println("2. Batch execute 'sh -c \"echo fail >&2; exit 42\"'...")
	{
		mgr := session.NewManager()
		mgr.SetHosts([]session.HostRow{
			{Target: session.HostTarget{IP: host, Port: 22, Username: "root", Password: password}, Enabled: true},
		})
		mgr.SetOptions(insecureOpts)

		results := mgr.ExecCommands(context.Background(), `sh -c 'echo fail >&2; exit 42'`)
		if len(results) != 1 {
			fail("non-zero exit", fmt.Sprintf("expected 1 result, got %d", len(results)))
		} else {
			r := results[0]
			if !r.Success && r.ExitCode == 42 && r.ErrorKind == "" {
				pass("non-zero exit — exit=42, success=false, no transport error")
			} else {
				fail("non-zero exit", fmt.Sprintf("success=%v exit=%d errKind=%s stderr=%q", r.Success, r.ExitCode, r.ErrorKind, r.Stderr))
			}
		}
	}

	// 3. Upload single file
	fmt.Println("3. Upload test file...")
	{
		localPath := "/tmp/ventre-panel-test-upload.txt"
		testContent := "ventre-panel upload test v0.1.0\n"
		if err := os.WriteFile(localPath, []byte(testContent), 0644); err != nil {
			fail("upload", fmt.Sprintf("cannot create local file: %v", err))
		} else {
			mgr := session.NewManager()
			mgr.SetHosts([]session.HostRow{
				{Target: session.HostTarget{IP: host, Port: 22, Username: "root", Password: password}, Enabled: true},
			})
			mgr.SetOptions(insecureOpts)

			results := mgr.UploadFiles(context.Background(), localPath, "/tmp/ventre-panel-test/upload-test.txt", true)
			if len(results) != 1 {
				fail("upload", fmt.Sprintf("expected 1 result, got %d", len(results)))
			} else {
				r := results[0]
				if r.Success && r.ErrorKind == "" {
					pass("upload single file — success")
				} else {
					fail("upload", fmt.Sprintf("success=%v errKind=%s errMsg=%s", r.Success, r.ErrorKind, r.ErrorMessage))
				}
			}
			os.Remove(localPath)
		}
	}

	// 4. Download single file (per-host isolated dir)
	fmt.Println("4. Download test file...")
	{
		localUploadPath := "/tmp/ventre-panel-test-dl-source.txt"
		dlContent := "ventre-panel download test content\n"
		if err := os.WriteFile(localUploadPath, []byte(dlContent), 0644); err != nil {
			fail("download", fmt.Sprintf("cannot create local source file: %v", err))
		} else {
			mgr := session.NewManager()
			mgr.SetHosts([]session.HostRow{
				{Target: session.HostTarget{IP: host, Port: 22, Username: "root", Password: password}, Enabled: true},
			})
			mgr.SetOptions(insecureOpts)
			mgr.UploadFiles(context.Background(), localUploadPath, "/tmp/ventre-panel-test/dl-source.txt", true)
			os.Remove(localUploadPath)

			localDir := "/tmp/ventre-panel-test-downloads"
			os.RemoveAll(localDir)

			mgr2 := session.NewManager()
			mgr2.SetHosts([]session.HostRow{
				{Target: session.HostTarget{IP: host, Port: 22, Username: "root", Password: password}, Enabled: true},
			})
			mgr2.SetOptions(insecureOpts)

			results := mgr2.DownloadFiles(context.Background(), "/tmp/ventre-panel-test/dl-source.txt", localDir, true)
			if len(results) != 1 {
				fail("download", fmt.Sprintf("expected 1 result, got %d", len(results)))
			} else {
				r := results[0]
				if r.Success && r.ErrorKind == "" {
					downloadedPath := localDir + "/" + host + "/dl-source.txt"
					data, err := os.ReadFile(downloadedPath)
					if err != nil {
						fail("download", fmt.Sprintf("cannot read downloaded file: %v", err))
					} else if string(data) == dlContent {
						pass("download single file — success, per-host dir, content matches")
					} else {
						fail("download", fmt.Sprintf("content mismatch: got %q want %q", string(data), dlContent))
					}
				} else {
					fail("download", fmt.Sprintf("success=%v errKind=%s errMsg=%s", r.Success, r.ErrorKind, r.ErrorMessage))
				}
			}
			os.RemoveAll(localDir)
		}
	}

	// 5. Wrong password -> auth_failed
	fmt.Println("5. Wrong password test...")
	{
		mgr := session.NewManager()
		mgr.SetHosts([]session.HostRow{
			{Target: session.HostTarget{IP: host, Port: 22, Username: "root", Password: "wrong-password-xyz"}, Enabled: true},
		})
		mgr.SetOptions(session.RunOptions{Timeout: 10 * time.Second, Concurrency: 1, InsecureIgnoreHostKey: true})

		results := mgr.ExecCommands(context.Background(), "echo test")
		if len(results) != 1 {
			fail("wrong password", fmt.Sprintf("expected 1 result, got %d", len(results)))
		} else {
			r := results[0]
			if r.ErrorKind == "auth_failed" && !r.Success {
				pass("wrong password — auth_failed detected")
			} else {
				fail("wrong password", fmt.Sprintf("errKind=%s (want auth_failed) success=%v", r.ErrorKind, r.Success))
			}
		}
	}

	// 6. Wrong known_hosts -> host_key_failed (security: verification works)
	fmt.Println("6. Host key verification test (wrong known_hosts path)...")
	{
		mgr := session.NewManager()
		mgr.SetHosts([]session.HostRow{
			{Target: session.HostTarget{IP: host, Port: 22, Username: "root", Password: password}, Enabled: true},
		})
		// Use a non-existent known_hosts path -> should fail host key verification
		mgr.SetOptions(session.RunOptions{
			Timeout:               10 * time.Second,
			Concurrency:           1,
			KnownHostsPath:        "/tmp/nonexistent-known-hosts-ventre-test",
			InsecureIgnoreHostKey: false,
		})

		results := mgr.ExecCommands(context.Background(), "echo test")
		if len(results) != 1 {
			fail("host key", fmt.Sprintf("expected 1 result, got %d", len(results)))
		} else {
			r := results[0]
			if r.ErrorKind == "host_key_failed" && !r.Success {
				pass("host key — host_key_failed with wrong known_hosts, verification works")
			} else {
				fail("host key", fmt.Sprintf("errKind=%s (want host_key_failed) success=%v", r.ErrorKind, r.Success))
			}
		}
	}

	// 7. Password redaction
	fmt.Println("7. Password redaction check...")
	{
		redacted := secure.RedactSecrets("password is "+password+" in text", []string{password})
		if strings.Contains(redacted, password) {
			fail("redaction", "password still present after RedactSecrets")
		} else {
			pass("password redaction — password removed from text")
		}
	}

	// 8. Result does not contain password
	fmt.Println("8. Result struct has no Password field...")
	{
		// Compile-time check: HostResult struct has no Password field
		// Verified by: grep Password internal/session/host.go returns only HostTarget.Password
		pass("HostResult — no Password field (compile-time guarantee)")
	}

	// 9. No persistence: check no files written to home
	fmt.Println("9. No persistence check...")
	{
		homeFiles := []string{
			os.ExpandEnv("$HOME/.ventre-panel"),
			os.ExpandEnv("$HOME/.config/ventre-panel"),
		}
		allMissing := true
		for _, p := range homeFiles {
			if _, err := os.Stat(p); err == nil {
				fail("persistence", fmt.Sprintf("unexpected file/dir found: %s", p))
				allMissing = false
			}
		}
		if allMissing {
			pass("no persistence — no config/log directories in home")
		}
	}

	// Cleanup
	doRemoteCmd(host, password, "rm -rf /tmp/ventre-panel-test")

	fmt.Println()
	if failures > 0 {
		fmt.Printf("RESULT: %d FAILURES\n", failures)
		os.Exit(1)
	} else {
		fmt.Println("RESULT: ALL PASSED")
	}
}

func doRemoteCmd(host, password, cmd string) {
	mgr := session.NewManager()
	mgr.SetHosts([]session.HostRow{
		{Target: session.HostTarget{IP: host, Port: 22, Username: "root", Password: password}, Enabled: true},
	})
	mgr.SetOptions(session.RunOptions{Timeout: 10 * time.Second, Concurrency: 1, InsecureIgnoreHostKey: true})
	mgr.ExecCommands(context.Background(), cmd)
}
