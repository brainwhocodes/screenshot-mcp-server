package mcpserver

import (
	"errors"
	"strings"
	"testing"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/window"
)

func TestAsToolExecutionErrorPreservesExistingToolError(t *testing.T) {
	toolErr := newFeatureUnavailable("wait_for_text", "vision not enabled")
	got := asToolExecutionError("tool-from-handler", toolErr)

	var gotTE ToolError
	if !errors.As(got, &gotTE) {
		t.Fatalf("expected ToolError, got %T", got)
	}

	if gotTE.ToolName != "wait_for_text" {
		t.Fatalf("tool name = %q, want %q", gotTE.ToolName, "wait_for_text")
	}
	if gotTE.Code != toolErrorCodeFeatureUnavailable {
		t.Fatalf("code = %q, want %q", gotTE.Code, toolErrorCodeFeatureUnavailable)
	}
}

func TestAsToolExecutionErrorMapsPermissionError(t *testing.T) {
	got := asToolExecutionError("take_window_screenshot", &window.PermissionError{
		ToolName:      "take_window_screenshot",
		Screen:        false,
		Accessibility: false,
	})

	var gotTE ToolError
	if !errors.As(got, &gotTE) {
		t.Fatalf("expected ToolError, got %T", got)
	}
	if gotTE.Code != toolErrorCodePermissionDenied {
		t.Fatalf("code = %q, want %q", gotTE.Code, toolErrorCodePermissionDenied)
	}
	if !strings.Contains(gotTE.Message, "required macOS permissions are not granted") {
		t.Fatalf("message missing permission phrase: %q", gotTE.Message)
	}
}

func TestAsToolExecutionErrorPassesThroughGenericError(t *testing.T) {
	const toolName = "take_screenshot"
	got := asToolExecutionError(toolName, errors.New("backend unavailable"))

	if !strings.Contains(got.Error(), toolName+": backend unavailable") {
		t.Fatalf("expected wrapped generic error, got %q", got)
	}
	if _, ok := got.(ToolError); ok {
		t.Fatal("did not expect ToolError for generic errors")
	}
}
