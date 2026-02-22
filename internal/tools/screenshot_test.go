package tools

import (
	"bytes"
	"context"
	"errors"
	"image"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/imgencode"
	"github.com/brainwhocodes/screenshot_mcp_server/internal/testutil"
)

func TestTakeScreenshot_UsesFixtureWhenSet(t *testing.T) {
	fixturePath := testutil.WriteFixtureJPEG(t)
	t.Setenv(FixtureImagePathEnv, fixturePath)

	svc := &ScreenshotService{}
	data, err := svc.TakeScreenshot(context.Background())
	if err != nil {
		t.Fatalf("TakeScreenshot failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty fixture bytes")
	}
}

func TestTakeScreenshot_Success(t *testing.T) {
	want := []byte("jpeg-bytes")
	svc := &ScreenshotService{
		Capture: func(context.Context) (image.Image, error) {
			return image.NewRGBA(image.Rect(0, 0, 4, 4)), nil
		},
		Encode: func(_ image.Image, _ imgencode.Options) ([]byte, error) {
			return want, nil
		},
		Options: imgencode.DefaultOptions,
	}

	data, err := svc.TakeScreenshot(context.Background())
	if err != nil {
		t.Fatalf("TakeScreenshot failed: %v", err)
	}
	if !bytes.Equal(data, want) {
		t.Fatalf("unexpected bytes: got=%q want=%q", data, want)
	}
}

func TestTakeScreenshot_CaptureError(t *testing.T) {
	svc := &ScreenshotService{
		Capture: func(context.Context) (image.Image, error) {
			return nil, errors.New("capture failed")
		},
		Encode:  imgencode.EncodeJPEG,
		Options: imgencode.DefaultOptions,
	}

	if _, err := svc.TakeScreenshot(context.Background()); err == nil {
		t.Fatal("expected capture error")
	}
}

func TestToolResultFromJPEG(t *testing.T) {
	data := []byte{1, 2, 3}
	result := ToolResultFromJPEG(data)
	if result == nil {
		t.Fatal("expected result")
		return
	}
	if len(result.Content) != 2 {
		t.Fatalf("expected 2 content items, got %d", len(result.Content))
	}

	textContent, ok := result.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("expected text content first, got %T", result.Content[0])
	}
	if textContent.Text == "" {
		t.Fatal("expected non-empty text content")
	}

	imageContent, ok := result.Content[1].(*sdkmcp.ImageContent)
	if !ok {
		t.Fatalf("expected image content second, got %T", result.Content[1])
	}
	if imageContent.MIMEType != "image/jpeg" {
		t.Fatalf("unexpected mime type %q", imageContent.MIMEType)
	}
	if !bytes.Equal(imageContent.Data, data) {
		t.Fatalf("unexpected image data: got=%v want=%v", imageContent.Data, data)
	}
}
