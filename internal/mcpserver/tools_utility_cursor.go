package mcpserver

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/safeexec"
	"github.com/brainwhocodes/screenshot_mcp_server/internal/tools"
)

// takeScreenshotWithCursor captures a screenshot including the mouse cursor where possible.
func takeScreenshotWithCursor(ctx context.Context, service ScreenshotService) ([]byte, bool, error) {
	data, capturedWithCursor, err := captureScreenshotWithCursorTool(ctx)
	if err == nil {
		return data, capturedWithCursor, nil
	}

	if service == nil {
		service = tools.NewScreenshotService()
	}
	data, err = service.TakeScreenshot(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("take screenshot fallback: %w", err)
	}
	return data, false, nil
}

func captureScreenshotWithCursorTool(ctx context.Context) ([]byte, bool, error) {
	if _, err := exec.LookPath("screencapture"); err != nil {
		return nil, false, fmt.Errorf("screencapture unavailable: %w", err)
	}

	tmp, err := os.CreateTemp("", "screenshot-with-cursor-*.jpg")
	if err != nil {
		return nil, false, fmt.Errorf("create temp screenshot: %w", err)
	}
	tmpPath := tmp.Name()
	if closeErr := tmp.Close(); closeErr != nil {
		_ = os.Remove(tmpPath)
		return nil, false, fmt.Errorf("close temp screenshot file: %w", closeErr)
	}
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := safeexec.RunCommandWithTimeout(ctx, 8*time.Second, "screencapture", "-C", "-x", "-t", "jpg", tmpPath); err != nil {
		return nil, false, fmt.Errorf("screencapture command failed: %w", err)
	}

	// #nosec G304 -- path is generated via os.CreateTemp.
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, false, fmt.Errorf("read screenshot file: %w", err)
	}
	return data, true, nil
}
