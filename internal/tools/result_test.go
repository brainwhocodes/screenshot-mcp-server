package tools

import (
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestToolResultFromJSONWithImage(t *testing.T) {
	result, err := ToolResultFromJSONWithImage(map[string]interface{}{
		"window_id": 12,
	}, []byte{1, 2, 3}, "image/png")
	if err != nil {
		t.Fatalf("ToolResultFromJSONWithImage() returned error: %v", err)
	}
	if result == nil || len(result.Content) != 2 {
		t.Fatalf("expected 2 content entries, got %d", len(result.Content))
	}

	text, ok := result.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("expected first content to be TextContent, got %T", result.Content[0])
	}
	if text.Text != `{"window_id":12}` {
		t.Fatalf("unexpected text payload: %q", text.Text)
	}

	image, ok := result.Content[1].(*sdkmcp.ImageContent)
	if !ok {
		t.Fatalf("expected second content to be ImageContent, got %T", result.Content[1])
	}
	if image.MIMEType != "image/png" {
		t.Fatalf("unexpected MIMEType: %q", image.MIMEType)
	}
	if len(image.Data) != 3 {
		t.Fatalf("unexpected image payload length: %d", len(image.Data))
	}
}
