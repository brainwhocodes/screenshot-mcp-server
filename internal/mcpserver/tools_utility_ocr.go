package mcpserver

import (
	"context"
	"fmt"
	"time"
)

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
