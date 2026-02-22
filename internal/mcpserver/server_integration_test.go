package mcpserver

import (
	"bytes"
	"context"
	"errors"
	"image/jpeg"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/testutil"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/tools"
)

func TestServerIntegration_InMemory(t *testing.T) {
	fixturePath := testutil.WriteFixtureJPEG(t)
	t.Setenv(tools.FixtureImagePathEnv, fixturePath)

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
	defer func() {
		_ = session.Close()
	}()

	toolsResult, err := session.ListTools(ctx, &sdkmcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if !containsTool(toolsResult.Tools, ToolName) {
		t.Fatalf("expected tool %q in list", ToolName)
	}
	if !containsTool(toolsResult.Tools, PressKeyToolName) {
		t.Fatalf("expected tool %q in list", PressKeyToolName)
	}

	callResult, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: ToolName})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}

	data := extractImage(t, callResult)
	if _, err := jpeg.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("decode jpeg: %v", err)
	}

	cancel()
	if err := <-serverErr; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("server run error: %v", err)
	}
}

func containsTool(list []*sdkmcp.Tool, name string) bool {
	for _, tool := range list {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func extractImage(t *testing.T, result *sdkmcp.CallToolResult) []byte {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("empty content")
	}

	for _, content := range result.Content {
		imageContent, ok := content.(*sdkmcp.ImageContent)
		if !ok {
			continue
		}
		if imageContent.MIMEType != "image/jpeg" {
			t.Fatalf("unexpected mime type %q", imageContent.MIMEType)
		}
		if len(imageContent.Data) == 0 {
			t.Fatal("empty image data")
		}
		return imageContent.Data
	}

	t.Fatalf("no image content found; got %d content items", len(result.Content))
	return nil
}
