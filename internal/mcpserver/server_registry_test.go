package mcpserver

import (
	"context"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/tools"
)

func TestNewServer_ToolNamesAreUnique(t *testing.T) {
	server := NewServer(tools.NewScreenshotService(), Config{})

	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	client := sdkmcp.NewClient(&sdkmcp.Implementation{
		Name:    "registry-test-client",
		Version: "0.1.0",
	}, nil)
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

	seen := make(map[string]struct{}, len(toolsResult.Tools))
	for _, tool := range toolsResult.Tools {
		toolName := tool.Name
		if _, exists := seen[toolName]; exists {
			t.Fatalf("duplicate tool name registered: %q", toolName)
		}
		seen[toolName] = struct{}{}
	}
}
