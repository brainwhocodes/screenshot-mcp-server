//go:build !darwin
// +build !darwin

package window

import (
	"context"
	"errors"
	"fmt"
	"image"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/imgencode"
)

const unsupportedWindowToolsMessage = "window automation is only supported on macOS (darwin)"

// Window represents a desktop window.
type Window struct {
	WindowID   uint32 `json:"window_id"`
	OwnerName  string `json:"owner_name"`
	PID        int32  `json:"pid"`
	Title      string `json:"title"`
	Bounds     Bounds `json:"bounds"`
	IsOnScreen bool   `json:"is_on_screen"`
}

// Bounds represents window bounds in screen coordinates (points).
type Bounds struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ScreenshotMetadata contains metadata about a window screenshot.
type ScreenshotMetadata struct {
	WindowID    uint32  `json:"window_id"`
	Bounds      Bounds  `json:"bounds"`
	ImageWidth  int     `json:"image_width"`
	ImageHeight int     `json:"image_height"`
	Scale       float64 `json:"scale"`
}

// RegionMetadata contains metadata about region screenshots.
type RegionMetadata struct {
	ImageWidth  int     `json:"image_width"`
	ImageHeight int     `json:"image_height"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	Scale       float64 `json:"scale"`
}

// SupportsWindowTools reports whether this platform supports window-level automation features.
func SupportsWindowTools() bool {
	return false
}

// UnsupportedWindowToolsReason returns the platform-specific reason window automation is unavailable.
func UnsupportedWindowToolsReason() string {
	return unsupportedWindowToolsMessage
}

// IsTiny reports if a window is smaller than the minimum practical size.
func (w *Window) IsTiny() bool {
	return w.Bounds.Width < 50 || w.Bounds.Height < 50
}

// IsSystemWindow reports whether the window appears to be a system window.
func (w *Window) IsSystemWindow() bool {
	return true
}

// ListWindows returns an error on unsupported platforms.
func ListWindows(context.Context) ([]Window, error) {
	return nil, errors.New(unsupportedWindowToolsMessage)
}

// FocusWindow returns an error on unsupported platforms.
func FocusWindow(ctx context.Context, _ uint32) error {
	return fmt.Errorf("%w", unsupportedPlatformError("focus_window"))
}

// TakeWindowScreenshot returns an unsupported error on non-Darwin.
func TakeWindowScreenshot(context.Context, uint32, imgencode.Options) ([]byte, *ScreenshotMetadata, error) {
	return nil, nil, errors.New(unsupportedWindowToolsMessage)
}

// TakeWindowScreenshotImage returns an unsupported error on non-Darwin.
func TakeWindowScreenshotImage(context.Context, uint32) (image.Image, *ScreenshotMetadata, error) {
	return nil, nil, errors.New(unsupportedWindowToolsMessage)
}

// TakeWindowScreenshotPNG returns an unsupported error on non-Darwin.
func TakeWindowScreenshotPNG(context.Context, uint32) ([]byte, *ScreenshotMetadata, error) {
	return nil, nil, errors.New(unsupportedWindowToolsMessage)
}

// TakeRegionScreenshot returns an unsupported error on non-Darwin.
func TakeRegionScreenshot(context.Context, float64, float64, float64, float64, string, imgencode.Options) ([]byte, *RegionMetadata, error) {
	return nil, nil, errors.New(unsupportedWindowToolsMessage)
}

// TakeRegionScreenshotPNG returns an unsupported error on non-Darwin.
func TakeRegionScreenshotPNG(context.Context, float64, float64, float64, float64, string) ([]byte, *RegionMetadata, error) {
	return nil, nil, errors.New(unsupportedWindowToolsMessage)
}

// Click returns an unsupported error on non-Darwin.
func Click(context.Context, uint32, float64, float64, string, int) error {
	return fmt.Errorf("%w", unsupportedPlatformError("click"))
}

// MouseMove returns an unsupported error on non-Darwin.
func MouseMove(context.Context, uint32, float64, float64) error {
	return fmt.Errorf("%w", unsupportedPlatformError("mouse_move"))
}

// MouseDown returns an unsupported error on non-Darwin.
func MouseDown(context.Context, uint32, float64, float64, string) error {
	return fmt.Errorf("%w", unsupportedPlatformError("mouse_down"))
}

// MouseUp returns an unsupported error on non-Darwin.
func MouseUp(context.Context, uint32, float64, float64, string) error {
	return fmt.Errorf("%w", unsupportedPlatformError("mouse_up"))
}

// Drag returns an unsupported error on non-Darwin.
func Drag(context.Context, uint32, float64, float64, float64, float64, string) error {
	return fmt.Errorf("%w", unsupportedPlatformError("drag"))
}

// Scroll returns an unsupported error on non-Darwin.
func Scroll(context.Context, uint32, float64, float64, float64, float64) error {
	return fmt.Errorf("%w", unsupportedPlatformError("scroll"))
}

// CheckPermissions returns unsupported state on non-Darwin.
func CheckPermissions() (screenRecording bool, accessibility bool) {
	return false, false
}

// EnsureAutomationPermissions returns a platform-specific error on non-Darwin systems.
func EnsureAutomationPermissions(toolName string) error {
	return fmt.Errorf("%w", unsupportedPlatformError(toolName))
}

// WaitForPixel is unsupported on non-Darwin.
func WaitForPixel(context.Context, uint32, float64, float64, [4]uint8, int, int, int) error {
	return fmt.Errorf("%w", unsupportedPlatformError("wait_for_pixel"))
}

// WaitForRegionStable is unsupported on non-Darwin.
func WaitForRegionStable(context.Context, uint32, float64, float64, float64, float64, int, int, int) error {
	return fmt.Errorf("%w", unsupportedPlatformError("wait_for_region_stable"))
}

// unsupportedPlatformError makes stub errors consistent and easy to match in callers.
func unsupportedPlatformError(toolName string) error {
	return fmt.Errorf("%s: %s", toolName, unsupportedWindowToolsMessage)
}

// LaunchApp is unsupported on non-Darwin.
func LaunchApp(context.Context, string) error {
	return fmt.Errorf("%w", unsupportedPlatformError("launch_app"))
}

// QuitApp is unsupported on non-Darwin.
func QuitApp(context.Context, string) error {
	return fmt.Errorf("%w", unsupportedPlatformError("quit_app"))
}

// WaitForProcess is unsupported on non-Darwin.
func WaitForProcess(context.Context, string, int, int) error {
	return fmt.Errorf("%w", unsupportedPlatformError("wait_for_process"))
}

// KillProcess is unsupported on non-Darwin.
func KillProcess(context.Context, string) error {
	return fmt.Errorf("%w", unsupportedPlatformError("kill_process"))
}
