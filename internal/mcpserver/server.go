package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/imgencode"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/tools"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/version"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/window"
)

const (
	// ToolName is the public MCP tool name for screenshot capture.
	ToolName = "take_screenshot"
	// ToolDescription explains tool behavior to MCP clients.
	ToolDescription = "Take a screenshot of the user's screen and return it as an image"

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

	// PressKeyToolName sends key presses to the active application
	PressKeyToolName        = "press_key"
	PressKeyToolDescription = "Send a key press to the focused application window"

	// DefaultSSEPort keeps parity with the Python implementation.
	DefaultSSEPort = 3001
)

// Config controls MCP server metadata.
type Config struct {
	Name    string
	Version string
}

// NewServer creates and configures the MCP server with all tools.
func NewServer(service *tools.ScreenshotService, cfg Config) *sdkmcp.Server {
	if service == nil {
		service = tools.NewScreenshotService()
	}
	inputService := tools.NewInputService()
	if cfg.Name == "" {
		cfg.Name = "Screenshot MCP Server"
	}
	if cfg.Version == "" {
		cfg.Version = version.Version
	}

	server := sdkmcp.NewServer(
		&sdkmcp.Implementation{
			Name:    cfg.Name,
			Version: cfg.Version,
		},
		nil,
	)

	type screenshotArgs struct{}
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ToolName,
		Description: ToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ screenshotArgs) (*sdkmcp.CallToolResult, any, error) {
		data, err := service.TakeScreenshot(ctx)
		if err != nil {
			return nil, nil, err
		}
		return tools.ToolResultFromJPEG(data), nil, nil
	})

	// Add list_windows tool
	type listWindowsArgs struct{}
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ListWindowsToolName,
		Description: ListWindowsToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ listWindowsArgs) (*sdkmcp.CallToolResult, any, error) {
		windows, err := window.ListWindows(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list windows: %w", err)
		}

		jsonData, err := json.Marshal(map[string]interface{}{
			"windows": windows,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("marshal windows: %w", err)
		}

		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{
					Text: string(jsonData),
				},
			},
		}, nil, nil
	})

	// Add focus_window tool
	type focusWindowArgs struct {
		WindowID uint32 `json:"window_id"`
	}
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        FocusWindowToolName,
		Description: FocusWindowToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args focusWindowArgs) (*sdkmcp.CallToolResult, any, error) {
		if args.WindowID == 0 {
			return nil, nil, fmt.Errorf("window_id is required")
		}

		if err := window.FocusWindow(ctx, args.WindowID); err != nil {
			return nil, nil, fmt.Errorf("focus window: %w", err)
		}

		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{
					Text: fmt.Sprintf("Window %d focused successfully", args.WindowID),
				},
			},
		}, nil, nil
	})

	// Add take_window_screenshot tool
	type takeWindowScreenshotArgs struct {
		WindowID uint32 `json:"window_id"`
	}
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        TakeWindowScreenshotToolName,
		Description: TakeWindowScreenshotToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args takeWindowScreenshotArgs) (*sdkmcp.CallToolResult, any, error) {
		if args.WindowID == 0 {
			return nil, nil, fmt.Errorf("window_id is required")
		}

		data, metadata, err := window.TakeWindowScreenshot(ctx, args.WindowID, imgencode.DefaultOptions)
		if err != nil {
			return nil, nil, fmt.Errorf("take window screenshot: %w", err)
		}

		jsonData, err := json.Marshal(metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal metadata: %w", err)
		}

		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{
					Text: string(jsonData),
				},
				&sdkmcp.ImageContent{
					Data:     data,
					MIMEType: "image/jpeg",
				},
			},
		}, nil, nil
	})

	// Add click tool
	type clickArgs struct {
		WindowID uint32  `json:"window_id"`
		X        float64 `json:"x"`
		Y        float64 `json:"y"`
		Button   string  `json:"button,omitempty"`
		Clicks   int     `json:"clicks,omitempty"`
	}
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ClickToolName,
		Description: ClickToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args clickArgs) (*sdkmcp.CallToolResult, any, error) {
		if args.WindowID == 0 {
			return nil, nil, fmt.Errorf("window_id is required")
		}

		// Set defaults
		if args.Button == "" {
			args.Button = "left"
		}
		if args.Clicks == 0 {
			args.Clicks = 1
		}

		if err := window.Click(ctx, args.WindowID, args.X, args.Y, args.Button, args.Clicks); err != nil {
			return nil, nil, fmt.Errorf("click: %w", err)
		}

		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{
					Text: fmt.Sprintf("Clicked at (%.0f, %.0f) in window %d", args.X, args.Y, args.WindowID),
				},
			},
		}, nil, nil
	})

	// Add press_key tool
	type pressKeyArgs struct {
		WindowID   uint32   `json:"window_id"`
		Key        string   `json:"key"`
		Modifiers  []string `json:"modifiers,omitempty"`
		KeyPresses int      `json:"presses,omitempty"`
	}
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        PressKeyToolName,
		Description: PressKeyToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args pressKeyArgs) (*sdkmcp.CallToolResult, any, error) {
		if args.WindowID == 0 {
			return nil, nil, fmt.Errorf("window_id is required")
		}
		if args.Key == "" {
			return nil, nil, fmt.Errorf("key is required")
		}
		if args.KeyPresses == 0 {
			args.KeyPresses = 1
		}
		if args.KeyPresses < 0 {
			return nil, nil, fmt.Errorf("presses must be >= 0")
		}

		if err := window.FocusWindow(ctx, args.WindowID); err != nil {
			return nil, nil, fmt.Errorf("focus window: %w", err)
		}

		for i := 0; i < args.KeyPresses; i++ {
			if err := inputService.PressKey(ctx, args.Key, args.Modifiers); err != nil {
				return nil, nil, err
			}
		}

		return tools.ToolResultFromText(fmt.Sprintf("Pressed %q %d time(s) in window %d", args.Key, args.KeyPresses, args.WindowID)), nil, nil
	})

	return server
}

// RunStdio starts serving MCP over stdio.
func RunStdio(ctx context.Context, server *sdkmcp.Server) error {
	if server == nil {
		return fmt.Errorf("server is nil")
	}
	return server.Run(ctx, &sdkmcp.StdioTransport{})
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
		Addr:    fmt.Sprintf(":%d", port),
		Handler: NewSSEHTTPHandler(server),
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)

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
