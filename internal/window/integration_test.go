//go:build darwin && integration

package window

import (
	"context"
	"testing"
)

// Integration tests require macOS and the integration build tag.
// Run with: go test -tags=integration ./internal/window/...

func TestIntegration_ListWindows(t *testing.T) {
	ctx := context.Background()

	windows, err := ListWindows(ctx)
	if err != nil {
		t.Fatalf("ListWindows failed: %v", err)
	}

	if len(windows) == 0 {
		t.Fatal("Expected at least one window, got none")
	}

	t.Logf("Found %d windows", len(windows))
	for i, w := range windows {
		t.Logf("  [%d] %q (%s) - %dx%d at (%.0f,%.0f)",
			i, w.Title, w.OwnerName,
			int(w.Bounds.Width), int(w.Bounds.Height),
			w.Bounds.X, w.Bounds.Y)
	}

	for _, w := range windows {
		if w.WindowID == 0 {
			t.Error("Window should have non-zero WindowID")
		}
		if w.OwnerName == "" {
			t.Error("Window should have OwnerName")
		}
		if w.Bounds.Width <= 0 || w.Bounds.Height <= 0 {
			t.Errorf("Window %d has invalid bounds: %v", w.WindowID, w.Bounds)
		}
	}
}

func TestIntegration_TakeWindowScreenshot(t *testing.T) {
	ctx := context.Background()

	windows, err := ListWindows(ctx)
	if err != nil {
		t.Fatalf("ListWindows failed: %v", err)
	}

	if len(windows) == 0 {
		t.Skip("No windows available for screenshot test")
	}

	targetWindow := windows[0]
	t.Logf("Taking screenshot of window %d (%s - %s)",
		targetWindow.WindowID, targetWindow.OwnerName, targetWindow.Title)

	data, metadata, err := TakeWindowScreenshot(ctx, targetWindow.WindowID, defaultTestOptions())
	if err != nil {
		t.Fatalf("TakeWindowScreenshot failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Screenshot data is empty")
	}

	if metadata == nil {
		t.Fatal("Screenshot metadata is nil")
	}

	t.Logf("Screenshot metadata: WindowID=%d, Bounds=%v, ImageSize=%dx%d, Scale=%.1f",
		metadata.WindowID, metadata.Bounds, metadata.ImageWidth, metadata.ImageHeight, metadata.Scale)

	if metadata.WindowID != targetWindow.WindowID {
		t.Errorf("Metadata WindowID = %d, want %d", metadata.WindowID, targetWindow.WindowID)
	}
	if metadata.ImageWidth <= 0 || metadata.ImageHeight <= 0 {
		t.Errorf("Invalid image dimensions: %dx%d", metadata.ImageWidth, metadata.ImageHeight)
	}
	if metadata.Scale < 1.0 {
		t.Errorf("Scale should be >= 1.0, got %.1f", metadata.Scale)
	}
}

func TestIntegration_FocusWindow(t *testing.T) {
	ctx := context.Background()

	windows, err := ListWindows(ctx)
	if err != nil {
		t.Fatalf("ListWindows failed: %v", err)
	}

	if len(windows) == 0 {
		t.Skip("No windows available for focus test")
	}

	targetWindow := windows[0]
	t.Logf("Focusing window %d (%s - %s)",
		targetWindow.WindowID, targetWindow.OwnerName, targetWindow.Title)

	err = FocusWindow(ctx, targetWindow.WindowID)
	if err != nil {
		t.Fatalf("FocusWindow failed: %v", err)
	}

	t.Log("Window focused successfully")
}

func TestIntegration_CheckPermissions(t *testing.T) {
	screenRecording, accessibility := CheckPermissions()

	t.Logf("Screen Recording permission: %v", screenRecording)
	t.Logf("Accessibility permission: %v", accessibility)

	if !screenRecording {
		t.Log("WARNING: Screen Recording permission not granted. Screenshots may fail.")
	}
	if !accessibility {
		t.Log("WARNING: Accessibility permission not granted. Click/type may fail.")
	}
}

func TestIntegration_Click(t *testing.T) {
	ctx := context.Background()

	windows, err := ListWindows(ctx)
	if err != nil {
		t.Fatalf("ListWindows failed: %v", err)
	}

	if len(windows) == 0 {
		t.Skip("No windows available for click test")
	}

	// Find a suitable window for clicking (preferably the test harness)
	var targetWindow *Window
	for i := range windows {
		if windows[i].Title == "UI Automation Test Target" {
			targetWindow = &windows[i]
			break
		}
	}
	if targetWindow == nil {
		targetWindow = &windows[0]
	}

	t.Logf("Testing click on window %d (%s - %s)",
		targetWindow.WindowID, targetWindow.OwnerName, targetWindow.Title)

	// First focus the window
	err = FocusWindow(ctx, targetWindow.WindowID)
	if err != nil {
		t.Fatalf("FocusWindow failed: %v", err)
	}

	// Take a screenshot to get current scale
	_, metadata, err := TakeWindowScreenshot(ctx, targetWindow.WindowID, defaultTestOptions())
	if err != nil {
		t.Fatalf("TakeWindowScreenshot failed: %v", err)
	}

	// Click at a safe location (center of window)
	centerX := float64(metadata.ImageWidth) / 2
	centerY := float64(metadata.ImageHeight) / 2

	t.Logf("Clicking at (%.0f, %.0f) in image coordinates", centerX, centerY)

	err = Click(ctx, targetWindow.WindowID, centerX, centerY, "left", 1)
	if err != nil {
		t.Fatalf("Click failed: %v", err)
	}

	t.Log("Click executed successfully")
}

func defaultTestOptions() struct{ Quality int } {
	return struct{ Quality int }{Quality: 60}
}
