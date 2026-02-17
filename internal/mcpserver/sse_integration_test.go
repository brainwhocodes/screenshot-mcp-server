package mcpserver

import (
	"bytes"
	"context"
	"image/jpeg"
	"net/http/httptest"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/testutil"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/tools"
)

func TestServerIntegration_SSE(t *testing.T) {
	fixturePath := testutil.WriteFixtureJPEG(t)
	t.Setenv(tools.FixtureImagePathEnv, fixturePath)

	server := NewServer(tools.NewScreenshotService(), Config{})
	httpServer := httptest.NewServer(NewSSEHTTPHandler(server))
	defer httpServer.Close()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "sse-client"}, nil)
	ctx := context.Background()

	session, err := client.Connect(ctx, &sdkmcp.SSEClientTransport{Endpoint: httpServer.URL}, nil)
	if err != nil {
		t.Fatalf("connect over sse: %v", err)
	}
	defer session.Close()

	toolsResult, err := session.ListTools(ctx, &sdkmcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("list tools over sse: %v", err)
	}
	if !containsTool(toolsResult.Tools, PressKeyToolName) {
		t.Fatalf("expected tool %q in list", PressKeyToolName)
	}

	callResult, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: ToolName})
	if err != nil {
		t.Fatalf("call tool over sse: %v", err)
	}

	for _, content := range callResult.Content {
		imageContent, ok := content.(*sdkmcp.ImageContent)
		if !ok {
			continue
		}
		if imageContent.MIMEType != "image/jpeg" {
			t.Fatalf("unexpected mime type %q", imageContent.MIMEType)
		}
		if _, err := jpeg.Decode(bytes.NewReader(imageContent.Data)); err != nil {
			t.Fatalf("decode jpeg from sse response: %v", err)
		}
		return
	}

	t.Fatalf("no image content found; got %d content items", len(callResult.Content))
}
