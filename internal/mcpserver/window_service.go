package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"image"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/imgencode"
	"github.com/brainwhocodes/screenshot_mcp_server/internal/window"
)

// WindowService represents the host operations required by MCP window tools.
type WindowService interface {
	SupportsWindowTools() bool
	EnsureAutomationPermissions(toolName string) error
	ListWindows(context.Context) ([]window.Window, error)
	FocusWindow(context.Context, uint32) error
	TakeWindowScreenshot(context.Context, uint32, imgencode.Options) ([]byte, *window.ScreenshotMetadata, error)
	TakeWindowScreenshotImage(context.Context, uint32) (image.Image, *window.ScreenshotMetadata, error)
	TakeWindowScreenshotPNG(context.Context, uint32) ([]byte, *window.ScreenshotMetadata, error)
	TakeRegionScreenshot(context.Context, float64, float64, float64, float64, string, imgencode.Options) ([]byte, *window.RegionMetadata, error)
	TakeRegionScreenshotPNG(context.Context, float64, float64, float64, float64, string) ([]byte, *window.RegionMetadata, error)
	Click(context.Context, uint32, float64, float64, string, int) error
	MouseMove(context.Context, uint32, float64, float64) error
	MouseDown(context.Context, uint32, float64, float64, string) error
	MouseUp(context.Context, uint32, float64, float64, string) error
	Drag(context.Context, uint32, float64, float64, float64, float64, string) error
	Scroll(context.Context, uint32, float64, float64, float64, float64) error
	WaitForPixel(context.Context, uint32, float64, float64, [4]uint8, int, int, int) error
	WaitForRegionStable(context.Context, uint32, float64, float64, float64, float64, int, int, int) error
	LaunchApp(context.Context, string) error
	QuitApp(context.Context, string) error
	WaitForProcess(context.Context, string, int, int) error
	KillProcess(context.Context, string) error
}

type defaultWindowService struct{}

func (defaultWindowService) SupportsWindowTools() bool {
	return window.SupportsWindowTools()
}

func (defaultWindowService) EnsureAutomationPermissions(toolName string) error {
	if !window.SupportsWindowTools() {
		return nil
	}
	if err := window.EnsureAutomationPermissions(toolName); err != nil {
		return wrapWindowServiceError("ensure automation permissions", err)
	}
	return nil
}

func (defaultWindowService) ListWindows(ctx context.Context) ([]window.Window, error) {
	windows, err := window.ListWindows(ctx)
	if err != nil {
		return nil, wrapWindowServiceError("list windows", err)
	}
	return windows, nil
}

func (defaultWindowService) FocusWindow(ctx context.Context, windowID uint32) error {
	if err := window.FocusWindow(ctx, windowID); err != nil {
		return wrapWindowServiceError("focus window", err)
	}
	return nil
}

func (defaultWindowService) TakeWindowScreenshot(ctx context.Context, windowID uint32, options imgencode.Options) ([]byte, *window.ScreenshotMetadata, error) {
	data, metadata, err := window.TakeWindowScreenshot(ctx, windowID, options)
	if err != nil {
		return nil, nil, wrapWindowServiceError("take window screenshot", err)
	}
	return data, metadata, nil
}

func (defaultWindowService) TakeWindowScreenshotImage(ctx context.Context, windowID uint32) (image.Image, *window.ScreenshotMetadata, error) {
	screenshot, metadata, err := window.TakeWindowScreenshotImage(ctx, windowID)
	if err != nil {
		return nil, nil, wrapWindowServiceError("take window screenshot", err)
	}
	return screenshot, metadata, nil
}

func (defaultWindowService) TakeWindowScreenshotPNG(ctx context.Context, windowID uint32) ([]byte, *window.ScreenshotMetadata, error) {
	data, metadata, err := window.TakeWindowScreenshotPNG(ctx, windowID)
	if err != nil {
		return nil, nil, wrapWindowServiceError("take window screenshot", err)
	}
	return data, metadata, nil
}

func (defaultWindowService) TakeRegionScreenshot(ctx context.Context, x, y, width, height float64, coordSpace string, options imgencode.Options) ([]byte, *window.RegionMetadata, error) {
	data, metadata, err := window.TakeRegionScreenshot(ctx, x, y, width, height, coordSpace, options)
	if err != nil {
		return nil, nil, wrapWindowServiceError("take region screenshot", err)
	}
	return data, metadata, nil
}

func (defaultWindowService) TakeRegionScreenshotPNG(ctx context.Context, x, y, width, height float64, coordSpace string) ([]byte, *window.RegionMetadata, error) {
	data, metadata, err := window.TakeRegionScreenshotPNG(ctx, x, y, width, height, coordSpace)
	if err != nil {
		return nil, nil, wrapWindowServiceError("take region screenshot", err)
	}
	return data, metadata, nil
}

func (defaultWindowService) Click(ctx context.Context, windowID uint32, x, y float64, button string, clicks int) error {
	if err := window.Click(ctx, windowID, x, y, button, clicks); err != nil {
		return wrapWindowServiceError("click", err)
	}
	return nil
}

func (defaultWindowService) MouseMove(ctx context.Context, windowID uint32, x, y float64) error {
	if err := window.MouseMove(ctx, windowID, x, y); err != nil {
		return wrapWindowServiceError("mouse move", err)
	}
	return nil
}

func (defaultWindowService) MouseDown(ctx context.Context, windowID uint32, x, y float64, button string) error {
	if err := window.MouseDown(ctx, windowID, x, y, button); err != nil {
		return wrapWindowServiceError("mouse down", err)
	}
	return nil
}

func (defaultWindowService) MouseUp(ctx context.Context, windowID uint32, x, y float64, button string) error {
	if err := window.MouseUp(ctx, windowID, x, y, button); err != nil {
		return wrapWindowServiceError("mouse up", err)
	}
	return nil
}

func (defaultWindowService) Drag(ctx context.Context, windowID uint32, fromX, fromY, toX, toY float64, button string) error {
	if err := window.Drag(ctx, windowID, fromX, fromY, toX, toY, button); err != nil {
		return wrapWindowServiceError("drag", err)
	}
	return nil
}

func (defaultWindowService) Scroll(ctx context.Context, windowID uint32, x, y, deltaX, deltaY float64) error {
	if err := window.Scroll(ctx, windowID, x, y, deltaX, deltaY); err != nil {
		return wrapWindowServiceError("scroll", err)
	}
	return nil
}

func (defaultWindowService) WaitForPixel(ctx context.Context, windowID uint32, x, y float64, rgba [4]uint8, tolerance, timeoutMs, pollMs int) error {
	if err := window.WaitForPixel(ctx, windowID, x, y, rgba, tolerance, timeoutMs, pollMs); err != nil {
		return wrapWindowServiceError("wait for pixel", err)
	}
	return nil
}

func (defaultWindowService) WaitForRegionStable(ctx context.Context, windowID uint32, x, y, width, height float64, stableCount, timeoutMs, pollMs int) error {
	if err := window.WaitForRegionStable(ctx, windowID, x, y, width, height, stableCount, timeoutMs, pollMs); err != nil {
		return wrapWindowServiceError("wait for region stable", err)
	}
	return nil
}

func (defaultWindowService) LaunchApp(ctx context.Context, appName string) error {
	if err := window.LaunchApp(ctx, appName); err != nil {
		return wrapWindowServiceError("launch app", err)
	}
	return nil
}

func (defaultWindowService) QuitApp(ctx context.Context, appName string) error {
	if err := window.QuitApp(ctx, appName); err != nil {
		return wrapWindowServiceError("quit app", err)
	}
	return nil
}

func (defaultWindowService) WaitForProcess(ctx context.Context, processName string, timeoutMs, pollMs int) error {
	if err := window.WaitForProcess(ctx, processName, timeoutMs, pollMs); err != nil {
		return wrapWindowServiceError("wait for process", err)
	}
	return nil
}

func (defaultWindowService) KillProcess(ctx context.Context, processName string) error {
	if err := window.KillProcess(ctx, processName); err != nil {
		return wrapWindowServiceError("kill process", err)
	}
	return nil
}

func wrapWindowServiceError(operation string, err error) error {
	// Keep request lifecycle and placeholder errors visible to callers.
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func newWindowService() WindowService {
	return defaultWindowService{}
}
