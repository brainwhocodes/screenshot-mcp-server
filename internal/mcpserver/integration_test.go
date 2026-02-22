//go:build darwin && integration

package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/tools"
)

// Integration tests require macOS and the integration build tag.
// Run with: go test -tags=integration ./internal/mcpserver/...

func TestIntegration_Server_ListWindows(t *testing.T) {
	server := NewServer(tools.NewScreenshotService(), Config{})

	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Run(ctx, serverTransport)
	}()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "integration-client"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	defer session.Close()

	result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: ListWindowsToolName})
	if err != nil {
		t.Fatalf("call list_windows: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	textContent, ok := result.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	var response struct {
		Windows []struct {
			WindowID  uint32 `json:"window_id"`
			OwnerName string `json:"owner_name"`
			Title     string `json:"title"`
			Bounds    struct {
				X      float64 `json:"x"`
				Y      float64 `json:"y"`
				Width  float64 `json:"width"`
				Height float64 `json:"height"`
			} `json:"bounds"`
		} `json:"windows"`
	}

	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("parse response: %v", err)
	}

	if len(response.Windows) == 0 {
		t.Fatal("expected at least one window")
	}

	t.Logf("Found %d windows", len(response.Windows))
	for i, w := range response.Windows {
		t.Logf("  [%d] WindowID=%d, Owner=%q, Title=%q, Bounds=(%.0f,%.0f %.0fx%.0f)",
			i, w.WindowID, w.OwnerName, w.Title,
			w.Bounds.X, w.Bounds.Y, w.Bounds.Width, w.Bounds.Height)
	}

	cancel()
	<-serverErr
}

func TestIntegration_Server_TakeWindowScreenshot(t *testing.T) {
	server := NewServer(tools.NewScreenshotService(), Config{})

	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Run(ctx, serverTransport)
	}()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "integration-client"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	defer session.Close()

	// First, list windows to get a window ID
	listResult, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: ListWindowsToolName})
	if err != nil {
		t.Fatalf("call list_windows: %v", err)
	}

	var listResponse struct {
		Windows []struct {
			WindowID uint32 `json:"window_id"`
		} `json:"windows"`
	}
	textContent, _ := listResult.Content[0].(*sdkmcp.TextContent)
	if err := json.Unmarshal([]byte(textContent.Text), &listResponse); err != nil {
		t.Fatalf("parse list response: %v", err)
	}

	if len(listResponse.Windows) == 0 {
		t.Skip("No windows available for screenshot test")
	}

	windowID := listResponse.Windows[0].WindowID

	// Take screenshot of the window
	screenshotResult, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name: TakeWindowScreenshotToolName,
		Arguments: map[string]interface{}{
			"window_id": windowID,
		},
	})
	if err != nil {
		t.Fatalf("call take_window_screenshot: %v", err)
	}

	if len(screenshotResult.Content) < 2 {
		t.Fatalf("expected at least 2 content items (metadata + image), got %d", len(screenshotResult.Content))
	}

	// Check metadata
	metaContent, ok := screenshotResult.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent for metadata, got %T", screenshotResult.Content[0])
	}

	var metadata struct {
		WindowID    uint32  `json:"window_id"`
		ImageWidth  int     `json:"image_width"`
		ImageHeight int     `json:"image_height"`
		Scale       float64 `json:"scale"`
	}
	if err := json.Unmarshal([]byte(metaContent.Text), &metadata); err != nil {
		t.Fatalf("parse metadata: %v", err)
	}

	t.Logf("Screenshot metadata: WindowID=%d, Size=%dx%d, Scale=%.1f",
		metadata.WindowID, metadata.ImageWidth, metadata.ImageHeight, metadata.Scale)

	// Check image
	imageContent, ok := screenshotResult.Content[1].(*sdkmcp.ImageContent)
	if !ok {
		t.Fatalf("expected ImageContent, got %T", screenshotResult.Content[1])
	}

	if imageContent.MIMEType != "image/jpeg" {
		t.Errorf("expected MIME type image/jpeg, got %s", imageContent.MIMEType)
	}

	if len(imageContent.Data) == 0 {
		t.Fatal("image data is empty")
	}

	// Verify it's valid JPEG
	if !bytes.HasPrefix(imageContent.Data, []byte{0xFF, 0xD8, 0xFF}) {
		t.Error("image data does not start with JPEG magic bytes")
	}

	t.Logf("Screenshot captured: %d bytes", len(imageContent.Data))

	cancel()
	<-serverErr
}

func TestIntegration_Server_FocusAndClick(t *testing.T) {
	server := NewServer(tools.NewScreenshotService(), Config{})

	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Run(ctx, serverTransport)
	}()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "integration-client"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect client: %v", err)
	}
	defer session.Close()

	// Get a window
	listResult, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: ListWindowsToolName})
	if err != nil {
		t.Fatalf("call list_windows: %v", err)
	}

	var listResponse struct {
		Windows []struct {
			WindowID uint32 `json:"window_id"`
		} `json:"windows"`
	}
	textContent, _ := listResult.Content[0].(*sdkmcp.TextContent)
	if err := json.Unmarshal([]byte(textContent.Text), &listResponse); err != nil {
		t.Fatalf("parse list response: %v", err)
	}

	if len(listResponse.Windows) == 0 {
		t.Skip("No windows available for click test")
	}

	windowID := listResponse.Windows[0].WindowID

	// Focus the window
	_, err = session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name: FocusWindowToolName,
		Arguments: map[string]interface{}{
			"window_id": windowID,
		},
	})
	if err != nil {
		t.Fatalf("call focus_window: %v", err)
	}

	t.Logf("Window %d focused", windowID)

	// Take screenshot to get dimensions
	screenshotResult, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name: TakeWindowScreenshotToolName,
		Arguments: map[string]interface{}{
			"window_id": windowID,
		},
	})
	if err != nil {
		t.Fatalf("call take_window_screenshot: %v", err)
	}

	metaContent, _ := screenshotResult.Content[0].(*sdkmcp.TextContent)
	var metadata struct {
		ImageWidth  int `json:"image_width"`
		ImageHeight int `json:"image_height"`
	}
	if err := json.Unmarshal([]byte(metaContent.Text), &metadata); err != nil {
		t.Fatalf("parse metadata: %v", err)
	}

	// Click in the center (harmless location)
	centerX := metadata.ImageWidth / 2
	centerY := metadata.ImageHeight / 2

	clickResult, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name: ClickToolName,
		Arguments: map[string]interface{}{
			"window_id": windowID,
			"x":         float64(centerX),
			"y":         float64(centerY),
		},
	})
	if err != nil {
		t.Fatalf("call click: %v", err)
	}

	clickText, _ := clickResult.Content[0].(*sdkmcp.TextContent)
	t.Logf("Click result: %s", clickText.Text)

	cancel()
	<-serverErr
}
