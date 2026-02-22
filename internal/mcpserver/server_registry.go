package mcpserver

import (
	"context"
	"fmt"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/imgencode"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/tools"
)

type mouseButtonArgs struct {
	WindowID uint32  `json:"window_id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Button   string  `json:"button,omitempty"`
}

type mouseButtonAction func(context.Context, uint32, float64, float64, string) error

type keyActionHandler func(context.Context, *tools.InputService, string, []string) error

func registerScreenshotTools(server *sdkmcp.Server, service ScreenshotService, windowService WindowService) {
	registerTakeScreenshotTool(server, service, windowService)
	registerTakeScreenshotPNGTool(server, service, windowService)
}

func registerWindowDiscoveryTools(server *sdkmcp.Server, service ScreenshotService, windowService WindowService) {
	registerListWindowsTool(server, windowService)
	registerScreenshotHashTool(server, service, windowService)
}

func registerWindowTools(server *sdkmcp.Server, windowService WindowService) {
	registerFocusWindowTool(server, windowService)
	registerTakeWindowScreenshotTool(server, windowService)
	registerTakeWindowScreenshotPNGTool(server, windowService)
	registerTakeRegionScreenshotTool(server, windowService)
	registerTakeRegionScreenshotPNGTool(server, windowService)
	registerClickTool(server, windowService)
	registerMouseMoveTool(server, windowService)
	registerMouseButtonTool(server, MouseDownToolName, MouseDownToolDescription, "down", windowService.MouseDown, windowService)
	registerMouseButtonTool(server, MouseUpToolName, MouseUpToolDescription, "up", windowService.MouseUp, windowService)
	registerDragTool(server, windowService)
	registerScrollTool(server, windowService)
}

func registerInputTools(server *sdkmcp.Server, inputService *tools.InputService, windowService WindowService) {
	registerPressKeyTool(server, inputService, windowService)
	registerTypeTextTool(server, inputService, windowService)
	registerKeyActionTool(server, KeyDownToolName, KeyDownToolDescription, inputService, performKeyDown, windowService)
	registerKeyActionTool(server, KeyUpToolName, KeyUpToolDescription, inputService, performKeyUp, windowService)
}

func registerSystemTools(server *sdkmcp.Server, windowService WindowService) {
	registerWaitForPixelTool(server, windowService)
	registerWaitForRegionStableTool(server, windowService)
	registerLaunchAppTool(server, windowService)
	registerQuitAppTool(server, windowService)
	registerWaitForProcessTool(server, windowService)
	registerKillProcessTool(server, windowService)
	registerClipboardTools(server)
}

func registerImageUtilities(server *sdkmcp.Server, windowService WindowService) {
	registerWaitForImageMatchTool(server, windowService)
	registerFindImageMatchesTool(server, windowService)
	registerCompareImagesTool(server, windowService)
	registerAssertScreenshotMatchesFixtureTool(server, windowService)
}

func registerExperimentalTools(server *sdkmcp.Server, windowService WindowService) {
	registerWaitForTextTool(server, windowService)
	registerRestartAppTool(server, windowService)
	registerStartRecordingTool(server, windowService)
	registerStopRecordingTool(server, windowService)
	registerTakeScreenshotWithCursorTool(server, windowService)
}

func registerTakeScreenshotTool(server *sdkmcp.Server, service ScreenshotService, windowService WindowService) {
	type screenshotArgs struct{}
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ToolName,
		Description: ToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ screenshotArgs) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, ToolName); err != nil {
			return nil, nil, err
		}
		data, err := service.TakeScreenshot(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("take screenshot: %w", err)
		}
		return tools.ToolResultFromJPEG(data), nil, nil
	})
}

func registerTakeScreenshotPNGTool(server *sdkmcp.Server, service ScreenshotService, windowService WindowService) {
	type takeScreenshotPNGArgs struct{}
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        TakeScreenshotPNGToolName,
		Description: TakeScreenshotPNGToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ takeScreenshotPNGArgs) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, TakeScreenshotPNGToolName); err != nil {
			return nil, nil, err
		}
		data, err := service.TakeScreenshotPNG(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("take screenshot png: %w", err)
		}

		result := tools.ToolResultFromText("Screenshot captured (PNG).")
		result.Content = append(result.Content, &sdkmcp.ImageContent{
			Data:     data,
			MIMEType: "image/png",
		})
		return result, nil, nil
	})
}

func registerListWindowsTool(server *sdkmcp.Server, windowService WindowService) {
	type listWindowsArgs struct{}
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ListWindowsToolName,
		Description: ListWindowsToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ listWindowsArgs) (*sdkmcp.CallToolResult, any, error) {
		windows, err := windowService.ListWindows(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list windows: %w", err)
		}

		result, err := tools.ToolResultFromJSON(map[string]interface{}{
			"windows": windows,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("marshal windows: %w", err)
		}
		return result, nil, nil
	})
}

func registerScreenshotHashTool(server *sdkmcp.Server, service ScreenshotService, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ScreenshotHashToolName,
		Description: ScreenshotHashToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		Algorithm string `json:"algorithm,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, ScreenshotHashToolName); err != nil {
			return nil, nil, err
		}
		if args.Algorithm == "" {
			args.Algorithm = "perceptual"
		}

		img, err := service.CaptureImage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("capture screenshot: %w", err)
		}

		hash, err := computeImageHash(img, args.Algorithm)
		if err != nil {
			return nil, nil, fmt.Errorf("compute hash: %w", err)
		}

		result, err := tools.ToolResultFromJSON(map[string]interface{}{
			"hash":      hash,
			"algorithm": args.Algorithm,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("marshal hash: %w", err)
		}
		return result, nil, nil
	})
}

func registerFocusWindowTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        FocusWindowToolName,
		Description: FocusWindowToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID uint32 `json:"window_id"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, FocusWindowToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if err := windowService.FocusWindow(ctx, args.WindowID); err != nil {
			return nil, nil, fmt.Errorf("focus window: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Window %d focused successfully", args.WindowID)), nil, nil
	})
}

func registerTakeWindowScreenshotTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        TakeWindowScreenshotToolName,
		Description: TakeWindowScreenshotToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID uint32 `json:"window_id"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, TakeWindowScreenshotToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		data, metadata, err := windowService.TakeWindowScreenshot(ctx, args.WindowID, imgencode.DefaultOptions)
		if err != nil {
			return nil, nil, fmt.Errorf("take window screenshot: %w", err)
		}
		result, err := tools.ToolResultFromJSON(metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal window metadata: %w", err)
		}
		result.Content = append(result.Content, &sdkmcp.ImageContent{
			Data:     data,
			MIMEType: "image/jpeg",
		})
		return result, nil, nil
	})
}

func registerTakeWindowScreenshotPNGTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        TakeWindowScreenshotPNGToolName,
		Description: TakeWindowScreenshotPNGToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID uint32 `json:"window_id"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, TakeWindowScreenshotPNGToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		data, metadata, err := windowService.TakeWindowScreenshotPNG(ctx, args.WindowID)
		if err != nil {
			return nil, nil, fmt.Errorf("take window screenshot: %w", err)
		}
		result, err := tools.ToolResultFromJSON(metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal window metadata: %w", err)
		}
		result.Content = append(result.Content, &sdkmcp.ImageContent{
			Data:     data,
			MIMEType: "image/png",
		})
		return result, nil, nil
	})
}

func registerTakeRegionScreenshotTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        TakeRegionScreenshotToolName,
		Description: TakeRegionScreenshotToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		Width      float64 `json:"width"`
		Height     float64 `json:"height"`
		CoordSpace string  `json:"coord_space,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, TakeRegionScreenshotToolName); err != nil {
			return nil, nil, err
		}
		if err := validateRegionInput(args.Width, args.Height, args.CoordSpace); err != nil {
			return nil, nil, err
		}

		data, metadata, err := windowService.TakeRegionScreenshot(ctx, args.X, args.Y, args.Width, args.Height, args.CoordSpace, imgencode.DefaultOptions)
		if err != nil {
			return nil, nil, fmt.Errorf("take region screenshot: %w", err)
		}
		result, err := tools.ToolResultFromJSON(metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal metadata: %w", err)
		}
		result.Content = append(result.Content, &sdkmcp.ImageContent{
			Data:     data,
			MIMEType: "image/jpeg",
		})
		return result, nil, nil
	})
}

func registerTakeRegionScreenshotPNGTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        TakeRegionScreenshotPNGToolName,
		Description: TakeRegionScreenshotPNGToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		Width      float64 `json:"width"`
		Height     float64 `json:"height"`
		CoordSpace string  `json:"coord_space,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, TakeRegionScreenshotPNGToolName); err != nil {
			return nil, nil, err
		}
		if err := validateRegionInput(args.Width, args.Height, args.CoordSpace); err != nil {
			return nil, nil, err
		}

		data, metadata, err := windowService.TakeRegionScreenshotPNG(ctx, args.X, args.Y, args.Width, args.Height, args.CoordSpace)
		if err != nil {
			return nil, nil, fmt.Errorf("take region screenshot: %w", err)
		}
		result, err := tools.ToolResultFromJSON(metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal metadata: %w", err)
		}
		result.Content = append(result.Content, &sdkmcp.ImageContent{
			Data:     data,
			MIMEType: "image/png",
		})
		return result, nil, nil
	})
}

func registerClickTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ClickToolName,
		Description: ClickToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID uint32  `json:"window_id"`
		X        float64 `json:"x"`
		Y        float64 `json:"y"`
		Button   string  `json:"button,omitempty"`
		Clicks   int     `json:"clicks,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, ClickToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if args.Button == "" {
			args.Button = "left"
		}
		if args.Clicks == 0 {
			args.Clicks = 1
		}
		if err := windowService.Click(ctx, args.WindowID, args.X, args.Y, args.Button, args.Clicks); err != nil {
			return nil, nil, fmt.Errorf("click: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Clicked at (%.0f, %.0f) in window %d", args.X, args.Y, args.WindowID)), nil, nil
	})
}

func registerMouseMoveTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        MouseMoveToolName,
		Description: MouseMoveToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args mouseButtonArgs) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, MouseMoveToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if err := focusWindowAndHandleError(ctx, windowService, args.WindowID); err != nil {
			return nil, nil, fmt.Errorf("focus window: %w", err)
		}
		if err := windowService.MouseMove(ctx, args.WindowID, args.X, args.Y); err != nil {
			return nil, nil, fmt.Errorf("mouse move: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Mouse moved to (%.0f, %.0f) in window %d", args.X, args.Y, args.WindowID)), nil, nil
	})
}

func registerDragTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        DragToolName,
		Description: DragToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID uint32  `json:"window_id"`
		FromX    float64 `json:"from_x"`
		FromY    float64 `json:"from_y"`
		ToX      float64 `json:"to_x"`
		ToY      float64 `json:"to_y"`
		Button   string  `json:"button,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, DragToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if args.Button == "" {
			args.Button = "left"
		}
		if err := focusWindowAndHandleError(ctx, windowService, args.WindowID); err != nil {
			return nil, nil, fmt.Errorf("focus window: %w", err)
		}
		if err := windowService.Drag(ctx, args.WindowID, args.FromX, args.FromY, args.ToX, args.ToY, args.Button); err != nil {
			return nil, nil, fmt.Errorf("drag: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Dragged (%s) from (%.0f, %.0f) to (%.0f, %.0f) in window %d", args.Button, args.FromX, args.FromY, args.ToX, args.ToY, args.WindowID)), nil, nil
	})
}

func registerScrollTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        ScrollToolName,
		Description: ScrollToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID uint32  `json:"window_id"`
		X        float64 `json:"x"`
		Y        float64 `json:"y"`
		DeltaX   float64 `json:"delta_x"`
		DeltaY   float64 `json:"delta_y"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, ScrollToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if err := focusWindowAndHandleError(ctx, windowService, args.WindowID); err != nil {
			return nil, nil, fmt.Errorf("focus window: %w", err)
		}
		if err := windowService.Scroll(ctx, args.WindowID, args.X, args.Y, args.DeltaX, args.DeltaY); err != nil {
			return nil, nil, fmt.Errorf("scroll: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Scrolled (%.0f, %.0f) at (%.0f, %.0f) in window %d", args.DeltaX, args.DeltaY, args.X, args.Y, args.WindowID)), nil, nil
	})
}

func registerPressKeyTool(server *sdkmcp.Server, inputService *tools.InputService, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        PressKeyToolName,
		Description: PressKeyToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID   uint32   `json:"window_id"`
		Key        string   `json:"key"`
		Modifiers  []string `json:"modifiers,omitempty"`
		KeyPresses int      `json:"presses,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, PressKeyToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
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
		if err := focusWindowAndHandleError(ctx, windowService, args.WindowID); err != nil {
			return nil, nil, err
		}
		for i := 0; i < args.KeyPresses; i++ {
			if err := inputService.PressKey(ctx, args.Key, args.Modifiers); err != nil {
				return nil, nil, fmt.Errorf("press key: %w", err)
			}
		}
		return tools.ToolResultFromText(fmt.Sprintf("Pressed %q %d time(s) in window %d", args.Key, args.KeyPresses, args.WindowID)), nil, nil
	})
}

func registerTypeTextTool(server *sdkmcp.Server, inputService *tools.InputService, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        TypeTextToolName,
		Description: TypeTextToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID uint32 `json:"window_id"`
		Text     string `json:"text"`
		DelayMs  int    `json:"delay_ms,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, TypeTextToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if args.Text == "" {
			return nil, nil, fmt.Errorf("text is required")
		}
		if err := focusWindowAndHandleError(ctx, windowService, args.WindowID); err != nil {
			return nil, nil, err
		}
		if err := inputService.TypeText(ctx, args.Text, args.DelayMs); err != nil {
			return nil, nil, fmt.Errorf("type text: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Typed %q in window %d", args.Text, args.WindowID)), nil, nil
	})
}

func registerKeyActionTool(server *sdkmcp.Server, toolName, description string, inputService *tools.InputService, handler keyActionHandler, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        toolName,
		Description: description,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID  uint32   `json:"window_id"`
		Key       string   `json:"key"`
		Modifiers []string `json:"modifiers,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, toolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if args.Key == "" {
			return nil, nil, fmt.Errorf("key is required")
		}
		if err := focusWindowAndHandleError(ctx, windowService, args.WindowID); err != nil {
			return nil, nil, err
		}
		if err := handler(ctx, inputService, args.Key, args.Modifiers); err != nil {
			return nil, nil, err
		}
		return tools.ToolResultFromText(fmt.Sprintf("%s %q in window %d", toolName, args.Key, args.WindowID)), nil, nil
	})
}

func performKeyDown(ctx context.Context, inputService *tools.InputService, key string, modifiers []string) error {
	if err := inputService.KeyDown(ctx, key, modifiers); err != nil {
		return fmt.Errorf("key down: %w", err)
	}
	return nil
}

func performKeyUp(ctx context.Context, inputService *tools.InputService, key string, modifiers []string) error {
	if err := inputService.KeyUp(ctx, key, modifiers); err != nil {
		return fmt.Errorf("key up: %w", err)
	}
	return nil
}

func registerWaitForPixelTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        WaitForPixelToolName,
		Description: WaitForPixelToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID       uint32   `json:"window_id"`
		X              float64  `json:"x"`
		Y              float64  `json:"y"`
		RGBA           [4]uint8 `json:"rgba"`
		Tolerance      int      `json:"tolerance,omitempty"`
		TimeoutMs      int      `json:"timeout_ms,omitempty"`
		PollIntervalMs int      `json:"poll_interval_ms,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, WaitForPixelToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if err := windowService.WaitForPixel(ctx, args.WindowID, args.X, args.Y, args.RGBA, args.Tolerance, args.TimeoutMs, args.PollIntervalMs); err != nil {
			return nil, nil, fmt.Errorf("wait for pixel: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Pixel at (%.0f, %.0f) matched color %v in window %d", args.X, args.Y, args.RGBA, args.WindowID)), nil, nil
	})
}

func registerWaitForRegionStableTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        WaitForRegionStableToolName,
		Description: WaitForRegionStableToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID       uint32  `json:"window_id"`
		X              float64 `json:"x"`
		Y              float64 `json:"y"`
		Width          float64 `json:"width"`
		Height         float64 `json:"height"`
		StableCount    int     `json:"stable_count,omitempty"`
		TimeoutMs      int     `json:"timeout_ms,omitempty"`
		PollIntervalMs int     `json:"poll_interval_ms,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, WaitForRegionStableToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if err := validateRegionInput(args.Width, args.Height, "points"); err != nil {
			return nil, nil, err
		}
		if err := windowService.WaitForRegionStable(ctx, args.WindowID, args.X, args.Y, args.Width, args.Height, args.StableCount, args.TimeoutMs, args.PollIntervalMs); err != nil {
			return nil, nil, fmt.Errorf("wait for region stable: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Region (%.0f, %.0f, %.0f, %.0f) stabilized in window %d", args.X, args.Y, args.Width, args.Height, args.WindowID)), nil, nil
	})
}

func registerLaunchAppTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        LaunchAppToolName,
		Description: LaunchAppToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		AppName string `json:"app_name"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if args.AppName == "" {
			return nil, nil, fmt.Errorf("app_name is required")
		}
		if err := windowService.LaunchApp(ctx, args.AppName); err != nil {
			return nil, nil, fmt.Errorf("launch app: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Launched app %q", args.AppName)), nil, nil
	})
}

func registerQuitAppTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        QuitAppToolName,
		Description: QuitAppToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		AppName string `json:"app_name"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if args.AppName == "" {
			return nil, nil, fmt.Errorf("app_name is required")
		}
		if err := windowService.QuitApp(ctx, args.AppName); err != nil {
			return nil, nil, fmt.Errorf("quit app: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Quit app %q", args.AppName)), nil, nil
	})
}

func registerWaitForProcessTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        WaitForProcessToolName,
		Description: WaitForProcessToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		ProcessName    string `json:"process_name"`
		TimeoutMs      int    `json:"timeout_ms,omitempty"`
		PollIntervalMs int    `json:"poll_interval_ms,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if args.ProcessName == "" {
			return nil, nil, fmt.Errorf("process_name is required")
		}
		if err := windowService.WaitForProcess(ctx, args.ProcessName, args.TimeoutMs, args.PollIntervalMs); err != nil {
			return nil, nil, fmt.Errorf("wait for process: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Process %q is running", args.ProcessName)), nil, nil
	})
}

func registerKillProcessTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        KillProcessToolName,
		Description: KillProcessToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		ProcessName string `json:"process_name"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if args.ProcessName == "" {
			return nil, nil, fmt.Errorf("process_name is required")
		}
		if err := windowService.KillProcess(ctx, args.ProcessName); err != nil {
			return nil, nil, fmt.Errorf("kill process: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Killed process %q", args.ProcessName)), nil, nil
	})
}

func registerClipboardTools(server *sdkmcp.Server) {
	registerSetClipboardTool(server)
	registerGetClipboardTool(server)
}

func registerSetClipboardTool(server *sdkmcp.Server) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        SetClipboardToolName,
		Description: SetClipboardToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		Text string `json:"text"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if args.Text == "" {
			return nil, nil, fmt.Errorf("text is required")
		}
		if err := setClipboard(ctx, args.Text); err != nil {
			return nil, nil, fmt.Errorf("set clipboard: %w", err)
		}
		return tools.ToolResultFromText("Clipboard set successfully"), nil, nil
	})
}

func registerGetClipboardTool(server *sdkmcp.Server) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        GetClipboardToolName,
		Description: GetClipboardToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, _ struct{}) (*sdkmcp.CallToolResult, any, error) {
		text, err := getClipboard(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("get clipboard: %w", err)
		}
		return tools.ToolResultFromText(text), nil, nil
	})
}

func registerWaitForImageMatchTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        WaitForImageMatchToolName,
		Description: WaitForImageMatchToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID       uint32  `json:"window_id,omitempty"`
		TemplateImage  string  `json:"template_image"`
		Threshold      float64 `json:"threshold,omitempty"`
		TimeoutMs      int     `json:"timeout_ms,omitempty"`
		PollIntervalMs int     `json:"poll_interval_ms,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, WaitForImageMatchToolName); err != nil {
			return nil, nil, err
		}
		if args.TemplateImage == "" {
			return nil, nil, fmt.Errorf("template_image is required")
		}
		args.Threshold = defaultThreshold(args.Threshold, defaultImageMatchThreshold)
		if err := validateThreshold(args.Threshold); err != nil {
			return nil, nil, err
		}
		args.TimeoutMs, args.PollIntervalMs = resolveTimeoutAndPoll(args.TimeoutMs, args.PollIntervalMs)
		coords, err := waitForImageMatch(ctx, args.WindowID, args.TemplateImage, args.Threshold, args.TimeoutMs, args.PollIntervalMs)
		if err != nil {
			return nil, nil, fmt.Errorf("wait for image match: %w", err)
		}
		result, err := tools.ToolResultFromJSON(map[string]interface{}{
			"found": true,
			"x":     coords.X,
			"y":     coords.Y,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("marshal result: %w", err)
		}
		return result, nil, nil
	})
}

func registerFindImageMatchesTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        FindImageMatchesToolName,
		Description: FindImageMatchesToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID      uint32  `json:"window_id,omitempty"`
		TemplateImage string  `json:"template_image"`
		Threshold     float64 `json:"threshold,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, FindImageMatchesToolName); err != nil {
			return nil, nil, err
		}
		if args.TemplateImage == "" {
			return nil, nil, fmt.Errorf("template_image is required")
		}
		args.Threshold = defaultThreshold(args.Threshold, defaultImageMatchThreshold)
		if err := validateThreshold(args.Threshold); err != nil {
			return nil, nil, err
		}
		matches, err := findImageMatches(ctx, args.WindowID, args.TemplateImage, args.Threshold)
		if err != nil {
			return nil, nil, fmt.Errorf("find image matches: %w", err)
		}
		result, err := tools.ToolResultFromJSON(map[string]interface{}{
			"matches": matches,
			"count":   len(matches),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("marshal result: %w", err)
		}
		return result, nil, nil
	})
}

func registerCompareImagesTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        CompareImagesToolName,
		Description: CompareImagesToolDescription,
	}, func(_ context.Context, _ *sdkmcp.CallToolRequest, args struct {
		Image1    string  `json:"image1"`
		Image2    string  `json:"image2"`
		Threshold float64 `json:"threshold,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, CompareImagesToolName); err != nil {
			return nil, nil, err
		}
		if args.Image1 == "" || args.Image2 == "" {
			return nil, nil, fmt.Errorf("both image1 and image2 are required")
		}
		args.Threshold = defaultThreshold(args.Threshold, defaultComparisonThreshold)
		if err := validateThreshold(args.Threshold); err != nil {
			return nil, nil, err
		}
		result, err := compareImages(args.Image1, args.Image2, args.Threshold)
		if err != nil {
			return nil, nil, fmt.Errorf("compare images: %w", err)
		}
		resultJSON, err := tools.ToolResultFromJSON(result)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal result: %w", err)
		}
		return resultJSON, nil, nil
	})
}

func registerAssertScreenshotMatchesFixtureTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        AssertScreenshotMatchesFixtureToolName,
		Description: AssertScreenshotMatchesFixtureToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID    uint32       `json:"window_id"`
		FixturePath string       `json:"fixture_path"`
		Threshold   float64      `json:"threshold,omitempty"`
		MaskRegions []MaskRegion `json:"mask_regions,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, AssertScreenshotMatchesFixtureToolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if args.FixturePath == "" {
			return nil, nil, fmt.Errorf("fixture_path is required")
		}
		args.Threshold = defaultThreshold(args.Threshold, defaultComparisonThreshold)
		if err := validateThreshold(args.Threshold); err != nil {
			return nil, nil, err
		}
		result, err := assertScreenshotMatchesFixture(ctx, args.WindowID, args.FixturePath, args.Threshold, args.MaskRegions)
		if err != nil {
			return nil, nil, fmt.Errorf("assert screenshot matches fixture: %w", err)
		}
		resultJSON, err := tools.ToolResultFromJSON(result)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal result: %w", err)
		}
		return resultJSON, nil, nil
	})
}

func registerWaitForTextTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        WaitForTextToolName,
		Description: WaitForTextToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID       uint32 `json:"window_id,omitempty"`
		Text           string `json:"text"`
		TimeoutMs      int    `json:"timeout_ms,omitempty"`
		PollIntervalMs int    `json:"poll_interval_ms,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if args.Text == "" {
			return nil, nil, fmt.Errorf("text is required")
		}
		if err := ensureWindowPermissions(windowService, WaitForTextToolName); err != nil {
			return nil, nil, err
		}
		found, err := waitForText(ctx, args.WindowID, args.Text, args.TimeoutMs, args.PollIntervalMs)
		if err != nil {
			return nil, nil, fmt.Errorf("wait for text: %w", err)
		}
		result, err := tools.ToolResultFromJSON(map[string]interface{}{"text": args.Text, "found": found})
		if err != nil {
			return nil, nil, fmt.Errorf("marshal result: %w", err)
		}
		return result, nil, nil
	})
}

func registerRestartAppTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        RestartAppToolName,
		Description: RestartAppToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		AppName string `json:"app_name"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if args.AppName == "" {
			return nil, nil, fmt.Errorf("app_name is required")
		}
		if err := ensureWindowPermissions(windowService, RestartAppToolName); err != nil {
			return nil, nil, err
		}
		if err := restartApp(ctx, args.AppName); err != nil {
			return nil, nil, fmt.Errorf("restart app: %w", err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("App %q restarted successfully", args.AppName)), nil, nil
	})
}

func registerStartRecordingTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        StartRecordingToolName,
		Description: StartRecordingToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		WindowID uint32 `json:"window_id,omitempty"`
		FPS      int    `json:"fps,omitempty"`
		Format   string `json:"format,omitempty"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, StartRecordingToolName); err != nil {
			return nil, nil, err
		}
		if args.FPS == 0 {
			args.FPS = 30
		}
		if args.Format == "" {
			args.Format = "mp4"
		}
		recordingID, err := startRecording(ctx, args.WindowID, args.FPS, args.Format)
		if err != nil {
			return nil, nil, fmt.Errorf("start recording: %w", err)
		}
		result, err := tools.ToolResultFromJSON(map[string]string{
			"recording_id": recordingID,
			"status":       "started",
		})
		if err != nil {
			return nil, nil, fmt.Errorf("marshal result: %w", err)
		}
		return result, nil, nil
	})
}

func registerStopRecordingTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        StopRecordingToolName,
		Description: StopRecordingToolDescription,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args struct {
		RecordingID string `json:"recording_id"`
	}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, StopRecordingToolName); err != nil {
			return nil, nil, err
		}
		if args.RecordingID == "" {
			return nil, nil, fmt.Errorf("recording_id is required")
		}
		if err := stopRecording(ctx, args.RecordingID); err != nil {
			return nil, nil, fmt.Errorf("stop recording: %w", err)
		}
		videoPath := ""
		result, err := tools.ToolResultFromJSON(map[string]string{
			"recording_id": args.RecordingID,
			"video_path":   videoPath,
			"status":       "stopped",
		})
		if err != nil {
			return nil, nil, fmt.Errorf("stop recording: %w", err)
		}
		return result, nil, nil
	})
}

func registerTakeScreenshotWithCursorTool(server *sdkmcp.Server, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        TakeScreenshotWithCursorToolName,
		Description: TakeScreenshotWithCursorToolDescription,
	}, func(_ context.Context, _ *sdkmcp.CallToolRequest, _ struct{}) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, TakeScreenshotWithCursorToolName); err != nil {
			return nil, nil, err
		}
		if err := takeScreenshotWithCursor(); err != nil {
			return nil, nil, fmt.Errorf("take screenshot with cursor: %w", err)
		}
		result := tools.ToolResultFromText("Cursor-aware screenshots are not yet implemented")
		return result, nil, nil
	})
}

func registerMouseButtonTool(server *sdkmcp.Server, toolName, description string, verb string, action mouseButtonAction, windowService WindowService) {
	sdkmcp.AddTool(server, &sdkmcp.Tool{
		Name:        toolName,
		Description: description,
	}, func(ctx context.Context, _ *sdkmcp.CallToolRequest, args mouseButtonArgs) (*sdkmcp.CallToolResult, any, error) {
		if err := ensureWindowPermissions(windowService, toolName); err != nil {
			return nil, nil, err
		}
		if err := validateWindowID(args.WindowID); err != nil {
			return nil, nil, err
		}
		if args.Button == "" {
			args.Button = "left"
		}
		if err := focusWindowAndHandleError(ctx, windowService, args.WindowID); err != nil {
			return nil, nil, fmt.Errorf("focus window: %w", err)
		}
		if err := action(ctx, args.WindowID, args.X, args.Y, args.Button); err != nil {
			return nil, nil, fmt.Errorf("mouse %s: %w", verb, err)
		}
		return tools.ToolResultFromText(fmt.Sprintf("Mouse %s (%s) at (%.0f, %.0f) in window %d", verb, args.Button, args.X, args.Y, args.WindowID)), nil, nil
	})
}

func focusWindowAndHandleError(ctx context.Context, windowService WindowService, windowID uint32) error {
	if err := windowService.FocusWindow(ctx, windowID); err != nil {
		return fmt.Errorf("focus window: %w", err)
	}
	return nil
}

func ensureWindowPermissions(windowService WindowService, toolName string) error {
	if err := windowService.EnsureAutomationPermissions(toolName); err != nil {
		return fmt.Errorf("%s: %w", toolName, err)
	}
	return nil
}
