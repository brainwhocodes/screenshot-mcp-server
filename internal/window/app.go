//go:build darwin

// Package window provides macOS window discovery and input helpers used by MCP tools.
package window

import (
	"context"
	"fmt"
	"time"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/safeexec"
)

const processCommandTimeout = 5 * time.Second

// LaunchApp launches an application by name or path.
// appName can be the app name (e.g., "Safari") or path to the app bundle.
func LaunchApp(ctx context.Context, appName string) error {
	if appName == "" {
		return fmt.Errorf("app name is required")
	}

	if err := safeexec.ValidateCommandArg(appName); err != nil {
		return fmt.Errorf("invalid app name %q: %w", appName, err)
	}
	output, err := safeexec.RunCommandWithInput(ctx, nil, "open", "-a", appName)
	if err != nil {
		return fmt.Errorf("launch app %q: %w, output: %s", appName, err, string(output))
	}

	return nil
}

// QuitApp quits an application by name using AppleScript.
func QuitApp(ctx context.Context, appName string) error {
	if appName == "" {
		return fmt.Errorf("app name is required")
	}
	if err := safeexec.ValidateCommandArg(appName); err != nil {
		return fmt.Errorf("invalid app name %q: %w", appName, err)
	}

	quotedAppName := safeexec.QuoteAppleScriptString(appName)
	script := fmt.Sprintf(`tell application "%s" to quit`, quotedAppName)
	if err := safeexec.RunAppleScript(ctx, script); err != nil {
		return fmt.Errorf("quit app %q: %w", appName, err)
	}

	return nil
}

// WaitForProcess waits for a process with the given name to appear.
// timeoutMs is the maximum time to wait in milliseconds.
// pollIntervalMs is how often to check in milliseconds.
func WaitForProcess(ctx context.Context, processName string, timeoutMs, pollIntervalMs int) error {
	if pollIntervalMs <= 0 {
		pollIntervalMs = 100
	}
	if timeoutMs <= 0 {
		timeoutMs = 5000
	}

	timeout := time.After(time.Duration(timeoutMs) * time.Millisecond)
	ticker := time.NewTicker(time.Duration(pollIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for process: %w", ctx.Err())
		case <-timeout:
			return fmt.Errorf("timeout waiting for process %q", processName)
		case <-ticker.C:
			if isProcessRunning(ctx, processName) {
				return nil
			}
		}
	}
}

// KillProcess kills a process by name using pkill.
func KillProcess(ctx context.Context, processName string) error {
	if processName == "" {
		return fmt.Errorf("process name is required")
	}
	if err := safeexec.ValidateCommandArg(processName); err != nil {
		return fmt.Errorf("invalid process name %q: %w", processName, err)
	}

	output, err := safeexec.RunCommandWithTimeout(ctx, processCommandTimeout, "pkill", "-x", processName)
	if err != nil {
		return fmt.Errorf("kill process %q: %w, output: %s", processName, err, string(output))
	}

	return nil
}

func isProcessRunning(ctx context.Context, processName string) bool {
	_, err := safeexec.RunCommandWithTimeout(ctx, processCommandTimeout, "pgrep", "-x", processName)
	return err == nil
}
