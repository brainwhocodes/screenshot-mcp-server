package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/tools"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/version"
)

const (
	// ToolName is the public MCP tool name for screenshot capture.
	ToolName = "take_screenshot"
	// ToolDescription explains tool behavior to MCP clients.
	ToolDescription = "Take a screenshot of the user's screen and return it as an image"
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
