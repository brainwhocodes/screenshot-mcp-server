package mcpserver

import (
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/safeexec"
	"github.com/brainwhocodes/screenshot_mcp_server/internal/tools"
	"github.com/brainwhocodes/screenshot_mcp_server/internal/window"
)

// waitForText waits for specific text to appear on screen using OCR.
func waitForText(ctx context.Context, service ScreenshotService, windowID uint32, text string, timeoutMs, pollIntervalMs int) (bool, error) {
	if text == "" {
		return false, fmt.Errorf("text is required")
	}
	if service == nil {
		service = tools.NewScreenshotService()
	}
	timeoutMs, pollIntervalMs = resolveTimeoutAndPoll(timeoutMs, pollIntervalMs)
	if _, err := exec.LookPath("tesseract"); err != nil {
		return false, fmt.Errorf("OCR dependency missing: tesseract is required for wait_for_text")
	}

	deadline := time.NewTimer(time.Duration(timeoutMs) * time.Millisecond)
	ticker := time.NewTicker(time.Duration(pollIntervalMs) * time.Millisecond)
	defer ticker.Stop()
	defer deadline.Stop()

	target := normalizeTextForMatch(text)

	for {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("wait for text: %w", ctx.Err())
		case <-deadline.C:
			return false, fmt.Errorf("timeout waiting for text %q", text)
		case <-ticker.C:
			found, err := imageContainsText(ctx, service, windowID, target)
			if err != nil {
				return false, fmt.Errorf("perform OCR: %w", err)
			}
			if found {
				return true, nil
			}
		}
	}
}

func imageContainsText(ctx context.Context, service ScreenshotService, windowID uint32, target string) (bool, error) {
	imageForOCR, err := captureImageForOCR(ctx, service, windowID)
	if err != nil {
		return false, fmt.Errorf("capture image for OCR: %w", err)
	}

	text, err := runOCR(ctx, imageForOCR)
	if err != nil {
		return false, fmt.Errorf("run OCR: %w", err)
	}

	return strings.Contains(normalizeTextForMatch(text), target), nil
}

func captureImageForOCR(ctx context.Context, service ScreenshotService, windowID uint32) (image.Image, error) {
	if service == nil {
		service = tools.NewScreenshotService()
	}
	if windowID == 0 {
		capturedImage, err := service.CaptureImage(ctx)
		if err != nil {
			return nil, fmt.Errorf("capture screen image: %w", err)
		}
		return capturedImage, nil
	}

	capturedImage, _, err := window.TakeWindowScreenshotImage(ctx, windowID)
	if err != nil {
		return nil, fmt.Errorf("capture window image: %w", err)
	}
	return capturedImage, nil
}

func runOCR(ctx context.Context, img image.Image) (string, error) {
	tmp, err := os.CreateTemp("", "wait-for-text-*.png")
	if err != nil {
		return "", fmt.Errorf("create temp image: %w", err)
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()
	defer func() {
		_ = tmp.Close()
	}()

	if err := png.Encode(tmp, img); err != nil {
		return "", fmt.Errorf("encode image for OCR: %w", err)
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		return "", fmt.Errorf("reset temp image: %w", err)
	}

	output, err := safeexec.RunCommandWithTimeout(ctx, 10*time.Second, "tesseract", tmp.Name(), "stdout")
	if err != nil {
		return "", fmt.Errorf("tesseract command failed: %w", err)
	}
	return string(output), nil
}

func normalizeTextForMatch(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}
