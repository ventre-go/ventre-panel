package validation

import (
	"os"
	"testing"
)

func TestValidateLocalFilePath_NotDirectory(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "ventre-test-file-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = ValidateLocalFilePath(tmpFile.Name())
	if err != nil {
		t.Errorf("expected no error for regular file, got: %v", err)
	}
}

func TestValidateLocalFilePath_Directory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ventre-test-dir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	err = ValidateLocalFilePath(tmpDir)
	if err == nil {
		t.Error("expected error for directory path, got nil")
	}
}

func TestValidateLocalFilePath_Empty(t *testing.T) {
	err := ValidateLocalFilePath("")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

func TestValidateLocalFilePath_NonExistent(t *testing.T) {
	err := ValidateLocalFilePath("/tmp/ventre-nonexistent-file-12345.txt")
	if err == nil {
		t.Error("expected error for non-existent local upload file, got nil")
	}
}

func TestValidateRemoteFilePath_DirectorySuffix(t *testing.T) {
	err := ValidateRemoteFilePath("/var/log/")
	if err == nil {
		t.Error("expected error for remote path ending with /, got nil")
	}
}

func TestValidateRemoteFilePath_File(t *testing.T) {
	err := ValidateRemoteFilePath("/var/log/app.log")
	if err != nil {
		t.Errorf("expected no error for file path, got: %v", err)
	}
}

func TestValidateRemoteFilePath_Empty(t *testing.T) {
	err := ValidateRemoteFilePath("")
	if err == nil {
		t.Error("expected error for empty remote path, got nil")
	}
}

func TestValidateLocalDirPath_Empty(t *testing.T) {
	err := ValidateLocalDirPath("")
	if err == nil {
		t.Error("expected error for empty path, got nil")
	}
}

func TestValidateLocalDirPath_RejectsExistingFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ventre-output-file-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	err = ValidateLocalDirPath(tmpFile.Name())
	if err == nil {
		t.Error("expected error for existing file output path, got nil")
	}
}

func TestValidateLocalDirPath_AllowsExistingDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ventre-output-dir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := ValidateLocalDirPath(tmpDir); err != nil {
		t.Fatalf("expected existing directory to be valid, got %v", err)
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		cmd     string
		wantErr bool
	}{
		{"echo hello", false},
		{"", true},
		{"   ", true},
	}
	for _, tt := range tests {
		t.Run("cmd="+tt.cmd, func(t *testing.T) {
			err := ValidateCommand(tt.cmd)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}
