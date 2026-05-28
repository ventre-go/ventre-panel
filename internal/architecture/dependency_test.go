package architecture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoDirectSSHorSFTPImports(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	forbidden := []string{
		`"golang.org/x/crypto/` + `ssh"`,
		`"github.com/pkg/` + `sftp"`,
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".agents", ".codex", ".claude", "dist", "fyne-cross":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source := string(data)
		for _, importPath := range forbidden {
			if strings.Contains(source, importPath) {
				t.Fatalf("forbidden direct SSH/SFTP import %s found in %s", importPath, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVentreTransportDependencyIsDeclared(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "github.com/ventre-go/ventre-transport") {
		t.Fatal("expected go.mod to declare github.com/ventre-go/ventre-transport")
	}
}

func TestApplicationCodeDoesNotUsePersistentFyneAPIs(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	forbidden := []string{
		".Preferences(",
		".Storage(",
		"app.NewWithID(",
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".agents", ".codex", ".claude", "dist", "fyne-cross":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source := string(data)
		for _, token := range forbidden {
			if strings.Contains(source, token) {
				t.Fatalf("persistent Fyne API token %q found in %s", token, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestApplicationCodeDoesNotExposePrivateKeyAuth(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	forbidden := []string{
		"PrivateKey",
		"privateKey",
		"key file",
		"keyPath",
	}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".agents", ".codex", ".claude", "dist", "fyne-cross":
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		source := string(data)
		for _, token := range forbidden {
			if strings.Contains(source, token) {
				t.Fatalf("private-key auth token %q found in %s", token, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
