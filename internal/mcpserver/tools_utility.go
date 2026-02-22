package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/safeexec"
)

// FeatureUnavailableError marks a tool as implemented as a placeholder.
type FeatureUnavailableError struct {
	Tool   string
	Reason string
}

func (e FeatureUnavailableError) Error() string {
	return fmt.Sprintf("%s is currently unavailable: %s", e.Tool, e.Reason)
}

func (e FeatureUnavailableError) Unwrap() error {
	return errFeatureUnavailable
}

var (
	errFeatureUnavailable = errors.New("feature unavailable")
)

func newFeatureUnavailable(tool, reason string) error {
	return FeatureUnavailableError{
		Tool:   tool,
		Reason: reason,
	}
}

// setClipboard sets the macOS clipboard content using pbcopy.
func setClipboard(ctx context.Context, text string) error {
	_, err := safeexec.RunCommandWithInput(ctx, []byte(text), "pbcopy")
	if err != nil {
		return fmt.Errorf("set clipboard: %w", err)
	}
	return nil
}

// getClipboard gets the macOS clipboard content using pbpaste.
func getClipboard(ctx context.Context) (string, error) {
	output, err := safeexec.RunCommand(ctx, "pbpaste")
	if err != nil {
		return "", fmt.Errorf("get clipboard: %w", err)
	}
	return string(output), nil
}

// waitForText waits for specific text to appear on screen using OCR.
// Placeholder implementation: Vision integration is not yet wired.
func waitForText(ctx context.Context, windowID uint32, text string, timeoutMs, pollIntervalMs int) (bool, error) {
	if text == "" {
		return false, fmt.Errorf("text is required")
	}
	if timeoutMs <= 0 {
		timeoutMs = 30000 // 30 seconds default
	}
	if pollIntervalMs <= 0 {
		pollIntervalMs = 1000 // 1 second default
	}
	if windowID == 0 {
		return false, fmt.Errorf("window_id is required")
	}

	deadline := time.After(time.Duration(timeoutMs) * time.Millisecond)
	ticker := time.NewTicker(time.Duration(pollIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("wait for text: %w", ctx.Err())
		case <-deadline:
			return false, fmt.Errorf("timeout waiting for text %q", text)
		case <-ticker.C:
			return false, newFeatureUnavailable(WaitForTextToolName, "Vision framework integration not implemented")
		}
	}
}

// restartApp quits and relaunches an application.
func restartApp(ctx context.Context, appName string) error {
	if appName == "" {
		return fmt.Errorf("app name is required")
	}
	if err := safeexec.ValidateCommandArg(appName); err != nil {
		return fmt.Errorf("invalid app name %q: %w", appName, err)
	}

	quotedApp := safeexec.QuoteAppleScriptString(appName)

	// Best-effort quit; if not running we continue with the launch attempt.
	_ = safeexec.RunAppleScript(ctx, fmt.Sprintf(`tell application "%s" to quit`, quotedApp))

	// Wait a moment for the app to quit.
	time.Sleep(500 * time.Millisecond)

	// Launch the app using osascript.
	err := safeexec.RunAppleScript(ctx, fmt.Sprintf(`tell application "%s" to activate`, quotedApp))
	if err != nil {
		return fmt.Errorf("launch app %q: %w", appName, err)
	}

	return nil
}

type recordingState struct {
	mu     sync.Mutex
	active map[string]bool
}

func newRecordingState() *recordingState {
	return &recordingState{
		active: make(map[string]bool),
	}
}

var recordingStateStore = newRecordingState()

func (state *recordingState) start(ctx context.Context, fps int, format string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("start recording: %w", err)
	}
	if format == "" {
		format = "mp4"
	}
	if fps <= 0 {
		return "", fmt.Errorf("fps must be > 0")
	}

	recordingID := fmt.Sprintf("rec_%d_%d_%s", time.Now().Unix(), fps, format)
	state.mu.Lock()
	state.active[recordingID] = true
	state.mu.Unlock()

	return recordingID, nil
}

func (state *recordingState) stop(recordingID string) (bool, error) {
	state.mu.Lock()
	defer state.mu.Unlock()

	if !state.active[recordingID] {
		return false, fmt.Errorf("recording %s not found or already stopped", recordingID)
	}

	delete(state.active, recordingID)
	return true, nil
}

// startRecording starts recording the screen.
// Placeholder implementation: full recording needs AVFoundation integration.
func startRecording(ctx context.Context, _ uint32, fps int, format string) (string, error) {
	recordingID, err := recordingStateStore.start(ctx, fps, format)
	if err != nil {
		return "", err
	}
	return recordingID, newFeatureUnavailable(StartRecordingToolName, "AVFoundation integration not implemented")
}

// stopRecording stops a screen recording.
func stopRecording(_ context.Context, recordingID string) error {
	if _, err := recordingStateStore.stop(recordingID); err != nil {
		return err
	}
	return newFeatureUnavailable(StopRecordingToolName, "AVFoundation integration not implemented")
}

// takeScreenshotWithCursor captures a screenshot including the mouse cursor.
func takeScreenshotWithCursor() error {
	return newFeatureUnavailable(TakeScreenshotWithCursorToolName, "cursor capture integration not implemented")
}
