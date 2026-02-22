package client

import (
	"bytes"
	"context"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/mcpserver"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/testutil"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/tools"
)

func TestExtractJPEG_NoImageContent(t *testing.T) {
	_, err := ExtractJPEG(&sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{Text: "hello"},
		},
	})
	if err == nil {
		t.Fatal("expected error for missing image content")
	}
}

func TestTakeScreenshotWithTransport_InMemory(t *testing.T) {
	fixturePath := testutil.WriteFixtureJPEG(t)
	t.Setenv(tools.FixtureImagePathEnv, fixturePath)

	server := mcpserver.NewServer(tools.NewScreenshotService(), mcpserver.Config{})
	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	data, err := TakeScreenshotWithTransport(ctx, clientTransport)
	if err != nil {
		t.Fatalf("TakeScreenshotWithTransport failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty image data")
	}
	if _, err := jpeg.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("decode returned image: %v", err)
	}
}

func TestTakeScreenshotToFileWithTransport_InMemory(t *testing.T) {
	fixturePath := testutil.WriteFixtureJPEG(t)
	t.Setenv(tools.FixtureImagePathEnv, fixturePath)

	server := mcpserver.NewServer(tools.NewScreenshotService(), mcpserver.Config{})
	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	outputPath := filepath.Join(t.TempDir(), "output.jpg")
	if err := TakeScreenshotToFileWithTransport(ctx, outputPath, clientTransport); err != nil {
		t.Fatalf("TakeScreenshotToFileWithTransport failed: %v", err)
	}

	// Accepted G304 suppression: test output path is managed by t.TempDir().
	// #nosec G304
	written, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if len(written) == 0 {
		t.Fatal("expected non-empty file")
	}
	if _, err := jpeg.Decode(bytes.NewReader(written)); err != nil {
		t.Fatalf("decode output file: %v", err)
	}
}
