package validation

import (
	"fmt"
	"os"
	"strings"
)

// ValidateCommand checks that a command is non-empty.
func ValidateCommand(cmd string) error {
	if strings.TrimSpace(cmd) == "" {
		return fmt.Errorf("command must not be empty")
	}
	return nil
}

// ValidateLocalFilePath checks that a local file path is non-empty and not a directory.
func ValidateLocalFilePath(p string) error {
	p = strings.TrimSpace(p)
	if p == "" {
		return fmt.Errorf("local file path must not be empty")
	}
	info, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("local file does not exist")
	}
	if info.IsDir() {
		return fmt.Errorf("directory transfer is not supported; please select a single file")
	}
	return nil
}

// ValidateRemoteFilePath checks that a remote file path is non-empty.
func ValidateRemoteFilePath(p string) error {
	p = strings.TrimSpace(p)
	if p == "" {
		return fmt.Errorf("remote file path must not be empty")
	}
	if strings.HasSuffix(p, "/") {
		return fmt.Errorf("directory transfer is not supported; please specify a single file path")
	}
	return nil
}

// ValidateLocalDirPath checks that a local directory path is non-empty.
func ValidateLocalDirPath(p string) error {
	p = strings.TrimSpace(p)
	if p == "" {
		return fmt.Errorf("local directory path must not be empty")
	}
	if info, err := os.Stat(p); err == nil && !info.IsDir() {
		return fmt.Errorf("local output path must be a directory")
	}
	return nil
}
