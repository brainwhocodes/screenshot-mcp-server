package tools

import (
	"context"
	"errors"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestInputService_PressKey_NotConfigured(t *testing.T) {
	svc := &InputService{}
	if err := svc.PressKey(context.Background(), "a", nil); err == nil {
		t.Fatal("expected error for nil press function")
	}
}

func TestInputService_PressKey_CallsUnderlying(t *testing.T) {
	called := false
	var gotKey string
	var gotModifiers []string
	svc := &InputService{
		PressKeyFn: func(_ context.Context, key string, modifiers []string) error {
			called = true
			gotKey = key
			gotModifiers = append([]string(nil), modifiers...)
			return nil
		},
	}

	modifiers := []string{"command", "shift"}
	if err := svc.PressKey(context.Background(), "a", modifiers); err != nil {
		t.Fatalf("PressKey returned error: %v", err)
	}
	if !called {
		t.Fatal("expected PressKeyFn to be called")
	}
	if gotKey != "a" {
		t.Fatalf("unexpected key: got=%q want=%q", gotKey, "a")
	}
	if len(gotModifiers) != 2 || gotModifiers[0] != "command" || gotModifiers[1] != "shift" {
		t.Fatalf("unexpected modifiers: got=%v", gotModifiers)
	}
}

func TestInputService_PressKey_WrapsErrors(t *testing.T) {
	svc := &InputService{
		PressKeyFn: func(context.Context, string, []string) error {
			return errors.New("boom")
		},
	}

	err := svc.PressKey(context.Background(), "a", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() == "boom" {
		t.Fatalf("expected wrapped error, got %q", err.Error())
	}
}

func TestToolResultFromText(t *testing.T) {
	result := ToolResultFromText("hello")
	if result == nil {
		t.Fatal("expected result")
		return
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	text, ok := result.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}
	if text.Text != "hello" {
		t.Fatalf("unexpected text: got=%q want=%q", text.Text, "hello")
	}
}
