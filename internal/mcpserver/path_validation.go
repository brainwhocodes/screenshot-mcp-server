package mcpserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	allowedRunDir    string
	allowedRunDirMu  sync.RWMutex
	allowlistEnabled bool
)

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
