// Package mcpserver exposes MCP server wiring and tool registration for screenshot automation.
package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"image"
	"net/http"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/tools"
	"github.com/brainwhocodes/screenshot_mcp_server/internal/version"
)

// ScreenshotService defines MCP-facing screenshot capture dependencies.
type ScreenshotService interface {
	CaptureImage(context.Context) (image.Image, error)
	TakeScreenshot(context.Context) ([]byte, error)
	TakeScreenshotPNG(context.Context) ([]byte, error)
}

// Tool names and descriptions are exported for tests and integrations.
const (
	// ToolName is the public MCP tool name for screenshot capture.
	ToolName = "take_screenshot"
	// ToolDescription explains tool behavior to MCP clients.
	ToolDescription = "Take a screenshot of the user's screen and return it as an image"

	// TakeScreenshotPNGToolName captures full screen as PNG
	TakeScreenshotPNGToolName = "take_screenshot_png"
	// TakeScreenshotPNGToolDescription describes the full-screen PNG capture tool.
	TakeScreenshotPNGToolDescription = "Take a lossless PNG screenshot of the user's screen"

	// ListWindowsToolName lists all visible windows
	ListWindowsToolName        = "list_windows"
	ListWindowsToolDescription = "List all visible application windows with their metadata"

	// FocusWindowToolName brings a window to foreground
	FocusWindowToolName        = "focus_window"
	FocusWindowToolDescription = "Bring a window to the foreground and activate its application"

	// TakeWindowScreenshotToolName captures a specific window
	TakeWindowScreenshotToolName        = "take_window_screenshot"
	TakeWindowScreenshotToolDescription = "Take a screenshot of a specific window and return image with metadata"

	// ClickToolName performs mouse clicks
	ClickToolName        = "click"
	ClickToolDescription = "Click at specific coordinates within a window"

	// ClickScreenToolName performs mouse clicks by screen coordinates
	ClickScreenToolName        = "click_screen"
	ClickScreenToolDescription = "Click at screen coordinates without a window context"

	// PressKeyToolName sends key presses to the active application
	PressKeyToolName        = "press_key"
	PressKeyToolDescription = "Send a key press to the focused application window"

	// TypeTextToolName types text into the active application
	TypeTextToolName        = "type_text"
	TypeTextToolDescription = "Type text into the focused application window"

	// TakeRegionScreenshotToolName captures a region of the screen
	TakeRegionScreenshotToolName        = "take_region_screenshot"
	TakeRegionScreenshotToolDescription = "Take a screenshot of a specific region of the screen"

	// TakeWindowScreenshotPNGToolName captures a window as PNG (lossless)
	TakeWindowScreenshotPNGToolName        = "take_window_screenshot_png"
	TakeWindowScreenshotPNGToolDescription = "Take a lossless PNG screenshot of a specific window"

	// TakeRegionScreenshotPNGToolName captures a region as PNG (lossless)
	TakeRegionScreenshotPNGToolName        = "take_region_screenshot_png"
	TakeRegionScreenshotPNGToolDescription = "Take a lossless PNG screenshot of a specific region of the screen"

	// MouseMoveToolName moves the mouse cursor
	MouseMoveToolName        = "mouse_move"
	MouseMoveToolDescription = "Move the mouse cursor to specific coordinates within a window"

	// MouseDownToolName sends mouse down event
	MouseDownToolName        = "mouse_down"
	MouseDownToolDescription = "Send a mouse down event at specific coordinates within a window"

	// MouseUpToolName sends mouse up event
	MouseUpToolName        = "mouse_up"
	MouseUpToolDescription = "Send a mouse up event at specific coordinates within a window"

	// DragToolName performs drag operation
	DragToolName        = "drag"
	DragToolDescription = "Perform a drag operation from one point to another within a window"

	// ScrollToolName performs scroll operation
	ScrollToolName        = "scroll"
	ScrollToolDescription = "Perform a scroll operation at specific coordinates within a window"

	// KeyDownToolName sends key down event
	KeyDownToolName        = "key_down"
	KeyDownToolDescription = "Send a key down event (for hold actions)"

	// KeyUpToolName sends key up event
	KeyUpToolName        = "key_up"
	KeyUpToolDescription = "Send a key up event (for hold actions)"

	// WaitForPixelToolName waits for a pixel to match a color
	WaitForPixelToolName        = "wait_for_pixel"
	WaitForPixelToolDescription = "Wait until a pixel at specific coordinates matches an expected color"

	// WaitForRegionStableToolName waits for a region to stop changing
	WaitForRegionStableToolName        = "wait_for_region_stable"
	WaitForRegionStableToolDescription = "Wait until a region of the window stops changing"

	// LaunchAppToolName launches an application
	LaunchAppToolName        = "launch_app"
	LaunchAppToolDescription = "Launch an application by name or path"

	// QuitAppToolName quits an application
	QuitAppToolName        = "quit_app"
	QuitAppToolDescription = "Quit an application by name"

	// WaitForProcessToolName waits for a process
	WaitForProcessToolName        = "wait_for_process"
	WaitForProcessToolDescription = "Wait for a process with the given name to appear"

	// KillProcessToolName kills a process
	KillProcessToolName        = "kill_process"
	KillProcessToolDescription = "Kill a process by name"

	// ScreenshotHashToolName generates a hash of the screen content
	ScreenshotHashToolName        = "screenshot_hash"
	ScreenshotHashToolDescription = "Generate a stable hash of the current screen for change detection"

	// SetClipboardToolName sets clipboard content
	SetClipboardToolName        = "set_clipboard"
	SetClipboardToolDescription = "Set the clipboard to a text value"

	// GetClipboardToolName gets clipboard content
	GetClipboardToolName        = "get_clipboard"
	GetClipboardToolDescription = "Get the current clipboard text content"

	// WaitForImageMatchToolName waits for a template image to appear
	WaitForImageMatchToolName        = "wait_for_image_match"
	WaitForImageMatchToolDescription = "Wait until a template image appears on screen"

	// FindImageMatchesToolName finds all matches of a template
	FindImageMatchesToolName        = "find_image_matches"
	FindImageMatchesToolDescription = "Find all occurrences of a template image on screen"

	// CompareImagesToolName compares two images
	CompareImagesToolName        = "compare_images"
	CompareImagesToolDescription = "Compare two images and return similarity metrics"

	// AssertScreenshotMatchesFixtureToolName compares screenshot to fixture
	AssertScreenshotMatchesFixtureToolName        = "assert_screenshot_matches_fixture"
	AssertScreenshotMatchesFixtureToolDescription = "Compare a window screenshot to a golden fixture image"

	// WaitForTextToolName waits for text to appear using OCR
	WaitForTextToolName        = "wait_for_text"
	WaitForTextToolDescription = "Wait for specific text to appear on screen using OCR"

	// RestartAppToolName restarts an application
	RestartAppToolName        = "restart_app"
	RestartAppToolDescription = "Restart an application (quit and relaunch)"

	// StartRecordingToolName starts screen recording
	StartRecordingToolName        = "start_recording"
	StartRecordingToolDescription = "Start recording the screen to a video file"

	// StopRecordingToolName stops screen recording
	StopRecordingToolName        = "stop_recording"
	StopRecordingToolDescription = "Stop the current screen recording"

	// TakeScreenshotWithCursorToolName captures screenshot with cursor
	TakeScreenshotWithCursorToolName        = "take_screenshot_with_cursor"
	TakeScreenshotWithCursorToolDescription = "Take a screenshot including the mouse cursor"

	// DefaultSSEPort keeps parity with the Python implementation.
	DefaultSSEPort = 3001
)

// Config controls MCP server metadata.
type Config struct {
	Name              string
	Version           string
	ExperimentalTools bool
	InputService      *tools.InputService
	WindowService     WindowService
}

// NewServer creates and configures the MCP server with all tools.
func NewServer(service ScreenshotService, cfg Config) *sdkmcp.Server {
	if service == nil {
		service = tools.NewScreenshotService()
	}
	windowService := cfg.WindowService
	if windowService == nil {
		windowService = newWindowService()
	}
	inputService := cfg.InputService
	if inputService == nil {
		inputService = tools.NewInputService()
	}
	if cfg.Name == "" {
		cfg.Name = "Screenshot MCP Server"
	}
	if cfg.Version == "" {
		cfg.Version = version.Version
	}
	recordingState := newRecordingState()

	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{
			Name:    cfg.Name,
			Version: cfg.Version,
		},
		nil,
	)

	registerScreenshotTools(server, service, windowService)
	if windowService.SupportsWindowTools() {
		registerWindowDiscoveryTools(server, windowService)
		registerWindowTools(server, windowService)
		registerInputTools(server, inputService, windowService)
		registerSystemTools(server, windowService)
		registerImageUtilities(server, windowService)
		if cfg.ExperimentalTools {
			registerExperimentalTools(server, service, windowService, recordingState)
		}
	}

	return server
}

// RunStdio starts serving MCP over stdio.
func RunStdio(ctx context.Context, server *sdkmcp.Server) error {
	if server == nil {
		return fmt.Errorf("server is nil")
	}
	if err := server.Run(ctx, &sdkmcp.StdioTransport{}); err != nil {
		return fmt.Errorf("run stdio server: %w", err)
	}
	return nil
}

// NewSSEHTTPHandler returns an HTTP handler for MCP SSE transport.
func NewSSEHTTPHandler(server *sdkmcp.Server) http.Handler {
	return sdkmcp.NewSSEHandler(func(_ *http.Request) *sdkmcp.Server {
		return server
	}, nil)
}

// ListenAndServeSSE starts an HTTP server that serves MCP SSE transport.
func ListenAndServeSSE(ctx context.Context, server *sdkmcp.Server, port int) error {
	if server == nil {
		return fmt.Errorf("server is nil")
	}
	if port <= 0 {
		return fmt.Errorf("invalid port %d", port)
	}

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           NewSSEHTTPHandler(server),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("shutdown http server: %w", err)
		}

		err := <-errCh
		if errors.Is(err, http.ErrServerClosed) || err == nil {
			return nil
		}
		return err
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) || err == nil {
			return nil
		}
		return err
	}
}
