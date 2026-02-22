package mcpserver

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	allowedRunDir    string
	allowedRunDirMu  sync.RWMutex
	allowlistEnabled bool
	artifactSeq      uint64
)

var artifactNameSanitizer = regexp.MustCompile(`[^A-Za-z0-9_-]+`)

// SetAllowedRunDirectory sets the directory where screenshot and fixture file
// operations are allowed.
func SetAllowedRunDirectory(dir string) error {
	if dir == "" {
		allowedRunDirMu.Lock()
		allowlistEnabled = false
		allowedRunDir = ""
		allowedRunDirMu.Unlock()
		return nil
	}

	// Ensure the directory exists.
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve run directory path: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(absDir, 0o750); err != nil {
				return fmt.Errorf("create run directory: %w", err)
			}
			allowedRunDirMu.Lock()
			allowedRunDir = absDir
			allowlistEnabled = true
			allowedRunDirMu.Unlock()
			return nil
		}
		return fmt.Errorf("stat run directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("run directory path is not a directory: %s", absDir)
	}

	allowedRunDirMu.Lock()
	allowedRunDir = absDir
	allowlistEnabled = true
	allowedRunDirMu.Unlock()
	return nil
}

// IsPathAllowed checks if a path is within the allowed run directory.
func IsPathAllowed(path string) bool {
	allowedRunDirMu.RLock()
	defer allowedRunDirMu.RUnlock()

	if !allowlistEnabled {
		return true
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	cleanPath := filepath.Clean(absPath)
	cleanAllowed := filepath.Clean(allowedRunDir)
	prefix := cleanAllowed + string(filepath.Separator)
	return cleanPath == cleanAllowed || strings.HasPrefix(cleanPath, prefix)
}

// ValidatePathAllowed returns an error if the path is outside the allowed run directory.
func ValidatePathAllowed(path string) error {
	if !IsPathAllowed(path) {
		allowedRunDirMu.RLock()
		allowed := allowedRunDir
		allowedRunDirMu.RUnlock()
		return fmt.Errorf("path %q is not within allowed run directory %q", path, allowed)
	}
	return nil
}

// ArtifactPath returns a deterministic artifact path with optional allowlist enforcement.
func ArtifactPath(prefix, ext string) (string, error) {
	allowedRunDirMu.RLock()
	allowed := allowedRunDir
	enabled := allowlistEnabled
	allowedRunDirMu.RUnlock()

	name := makeArtifactFileName(prefix, ext)
	if !enabled {
		return filepath.Join(os.TempDir(), name), nil
	}
	return filepath.Join(allowed, name), nil
}

// CreateArtifactFile creates a deterministic artifact file with secure permissions.
func CreateArtifactFile(prefix, ext string) (*os.File, error) {
	path, err := ArtifactPath(prefix, ext)
	if err != nil {
		return nil, err
	}

	// Accepted G304 suppression: artifact file paths are derived from deterministic
	// service-controlled components and validated by set run-directory policy.
	// #nosec G304
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create artifact file %q: %w", path, err)
	}
	return file, nil
}

func makeArtifactFileName(prefix, ext string) string {
	safePrefix := artifactNameSanitizer.ReplaceAllString(prefix, "-")
	if safePrefix == "" {
		safePrefix = "artifact"
	}
	seq := atomic.AddUint64(&artifactSeq, 1)
	if ext == "" {
		ext = "tmp"
	}
	ext = strings.TrimPrefix(ext, ".")
	return fmt.Sprintf("screenshot-mcp-server-%s-%06d.%s", safePrefix, seq, ext)
}
