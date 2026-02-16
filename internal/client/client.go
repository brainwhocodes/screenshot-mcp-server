package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/mcpserver"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/version"
)

const (
	// ServerCommandEnv allows overriding the stdio server command.
	ServerCommandEnv = "SCREENSHOT_MCP_SERVER_COMMAND"
	// ServerArgsEnv allows overriding the stdio server args.
	ServerArgsEnv = "SCREENSHOT_MCP_SERVER_ARGS"
)

// Config controls how the client connects to the screenshot server.
type Config struct {
	ServerCommand string
	ServerArgs    []string
}

// DefaultConfig returns the default client configuration.
func DefaultConfig() Config {
	cmd := strings.TrimSpace(os.Getenv(ServerCommandEnv))
	if cmd == "" {
		cmd = "screenshot_mcp_server"
	}

	args := strings.Fields(os.Getenv(ServerArgsEnv))
	if len(args) == 0 {
		args = []string{"server"}
	}
	return Config{
		ServerCommand: cmd,
		ServerArgs:    args,
	}
}

// TakeScreenshotToFile launches the server over stdio and writes JPEG output to outputPath.
func TakeScreenshotToFile(ctx context.Context, outputPath string, cfg Config) error {
	data, err := TakeScreenshot(ctx, cfg)
	if err != nil {
		return err
	}
	return writeOutput(outputPath, data)
}

// TakeScreenshot launches the stdio server and returns screenshot JPEG bytes.
func TakeScreenshot(ctx context.Context, cfg Config) ([]byte, error) {
	if strings.TrimSpace(cfg.ServerCommand) == "" {
		cfg = DefaultConfig()
	}

	cmd := exec.CommandContext(ctx, cfg.ServerCommand, cfg.ServerArgs...)
	transport := &sdkmcp.CommandTransport{Command: cmd}
	return TakeScreenshotWithTransport(ctx, transport)
}

// TakeScreenshotWithTransport returns screenshot bytes using the provided transport.
func TakeScreenshotWithTransport(ctx context.Context, transport sdkmcp.Transport) ([]byte, error) {
	if transport == nil {
		return nil, fmt.Errorf("transport is nil")
	}

	c := sdkmcp.NewClient(
		&sdkmcp.Implementation{
			Name:    "screenshot_mcp_server-client",
			Version: version.Version,
		},
		nil,
	)

	session, err := c.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to server: %w", err)
	}
	defer session.Close()

	result, err := session.CallTool(ctx, &sdkmcp.CallToolParams{Name: mcpserver.ToolName})
	if err != nil {
		return nil, fmt.Errorf("call %q tool: %w", mcpserver.ToolName, err)
	}

	return ExtractJPEG(result)
}

// TakeScreenshotToFileWithTransport writes screenshot JPEG bytes using the provided transport.
func TakeScreenshotToFileWithTransport(ctx context.Context, outputPath string, transport sdkmcp.Transport) error {
	data, err := TakeScreenshotWithTransport(ctx, transport)
	if err != nil {
		return err
	}
	return writeOutput(outputPath, data)
}

// ExtractJPEG extracts JPEG bytes from a tool result.
func ExtractJPEG(result *sdkmcp.CallToolResult) ([]byte, error) {
	if result == nil {
		return nil, fmt.Errorf("result is nil")
	}
	if result.IsError {
		if err := result.GetError(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("tool returned error")
	}

	for _, content := range result.Content {
		imageContent, ok := content.(*sdkmcp.ImageContent)
		if !ok {
			continue
		}
		if len(imageContent.Data) == 0 {
			return nil, fmt.Errorf("image content is empty")
		}
		if imageContent.MIMEType != "" && imageContent.MIMEType != "image/jpeg" {
			return nil, fmt.Errorf("unexpected mime type %q", imageContent.MIMEType)
		}
		return imageContent.Data, nil
	}

	return nil, errors.New("no image content returned by tool")
}

func writeOutput(outputPath string, data []byte) error {
	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
}
