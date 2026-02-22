package mcpserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArtifactPathAndFileCreation(t *testing.T) {
	// Ensure a clean global artifact state for deterministic test expectations.
	if err := SetAllowedRunDirectory(""); err != nil {
		t.Fatalf("reset allowlist: %v", err)
	}
	t.Cleanup(func() {
		_ = SetAllowedRunDirectory("")
	})

	tmpDir := t.TempDir()
	if err := SetAllowedRunDirectory(tmpDir); err != nil {
		t.Fatalf("set allowed directory: %v", err)
	}

	path1, err := ArtifactPath("window/scope", "jpg")
	if err != nil {
		t.Fatalf("artifact path for allowed dir: %v", err)
	}
	if filepath.Dir(path1) != filepath.Clean(tmpDir) {
		t.Fatalf("artifact path %q is not in allowlist dir %q", path1, tmpDir)
	}
	if !strings.Contains(filepath.Base(path1), "screenshot-mcp-server-window-scope-") {
		t.Fatalf("artifact name not sanitized as expected: %q", filepath.Base(path1))
	}

	file1, err := CreateArtifactFile("window/scope", "jpg")
	if err != nil {
		t.Fatalf("create artifact file: %v", err)
	}
	file1Name := file1.Name()
	if err := file1.Close(); err != nil {
		t.Fatalf("close first artifact file: %v", err)
	}
	defer func() {
		_ = os.Remove(file1Name)
	}()

	file2, err := CreateArtifactFile("window/scope", "jpg")
	if err != nil {
		t.Fatalf("create second artifact file: %v", err)
	}
	file2Name := file2.Name()
	if err := file2.Close(); err != nil {
		t.Fatalf("close second artifact file: %v", err)
	}
	defer func() {
		_ = os.Remove(file2Name)
	}()

	if file1Name == file2Name {
		t.Fatalf("expected unique artifact filenames, got %q twice", file1Name)
	}
	if filepath.Dir(file2Name) != filepath.Clean(tmpDir) {
		t.Fatalf("artifact path %q is not in allowlist dir %q", file2Name, tmpDir)
	}
}

func TestValidatePathAllowed(t *testing.T) {
	tmpDir := t.TempDir()
	otherDir := t.TempDir()

	if err := SetAllowedRunDirectory(tmpDir); err != nil {
		t.Fatalf("set allowed directory: %v", err)
	}
	t.Cleanup(func() {
		_ = SetAllowedRunDirectory("")
	})

	if err := ValidatePathAllowed(filepath.Join(tmpDir, "fixture.jpg")); err != nil {
		t.Fatalf("expected allowed path in dir to pass: %v", err)
	}
	if err := ValidatePathAllowed(filepath.Join(otherDir, "fixture.jpg")); err == nil {
		t.Fatal("expected denied path outside allowlist to fail")
	}

	if err := SetAllowedRunDirectory(""); err != nil {
		t.Fatalf("clear allowlist: %v", err)
	}
}
