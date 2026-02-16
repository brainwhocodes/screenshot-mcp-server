package tools

import (
	"context"
	"fmt"
	"image"
	"os"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/imgencode"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/screenshot"
)

const (
	// FixtureImagePathEnv makes integration tests deterministic in headless environments.
	FixtureImagePathEnv = "SCREENSHOT_MCP_TEST_IMAGE_PATH"
)

// CaptureFunc captures a full-screen image.
type CaptureFunc func(context.Context) (image.Image, error)

// EncodeFunc encodes an image using the provided options.
type EncodeFunc func(image.Image, imgencode.Options) ([]byte, error)

// ScreenshotService wraps screenshot capture and JPEG encoding.
type ScreenshotService struct {
	Capture CaptureFunc
	Encode  EncodeFunc
	Options imgencode.Options
}

// NewScreenshotService returns the default screenshot service.
func NewScreenshotService() *ScreenshotService {
	capturer := screenshot.NewCapturer()
	return &ScreenshotService{
		Capture: capturer.Capture,
		Encode:  imgencode.EncodeJPEG,
		Options: imgencode.DefaultOptions,
	}
}

// TakeScreenshot returns JPEG bytes for the current screen.
func (s *ScreenshotService) TakeScreenshot(ctx context.Context) ([]byte, error) {
	if fixturePath := os.Getenv(FixtureImagePathEnv); fixturePath != "" {
		data, err := os.ReadFile(fixturePath)
		if err != nil {
			return nil, fmt.Errorf("read fixture image %q: %w", fixturePath, err)
		}
		return data, nil
	}

	if s == nil || s.Capture == nil || s.Encode == nil {
		return nil, fmt.Errorf("screenshot service is not configured")
	}

	img, err := s.Capture(ctx)
	if err != nil {
		return nil, fmt.Errorf("capture screenshot: %w", err)
	}

	data, err := s.Encode(img, s.Options)
	if err != nil {
		return nil, fmt.Errorf("encode screenshot: %w", err)
	}

	return data, nil
}

// ToolResultFromJPEG wraps bytes in MCP image content.
func ToolResultFromJPEG(data []byte) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{
				Text: "Screenshot captured.",
			},
			&sdkmcp.ImageContent{
				Data:     data,
				MIMEType: "image/jpeg",
			},
		},
	}
}
