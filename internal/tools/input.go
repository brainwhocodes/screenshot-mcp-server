// Package tools provides MCP-facing service abstractions.
package tools

import (
	"context"
	"fmt"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/input"
)

// PressKeyFunc sends a key press to the OS.
type PressKeyFunc func(context.Context, string, []string) error

// TypeTextFunc types text to the OS.
type TypeTextFunc func(context.Context, string, int) error

// KeyDownFunc sends a key down event to the OS.
type KeyDownFunc func(context.Context, string, []string) error

// KeyUpFunc sends a key up event to the OS.
type KeyUpFunc func(context.Context, string, []string) error

// InputService wraps OS input injection primitives.
type InputService struct {
	PressKeyFn PressKeyFunc
	TypeTextFn TypeTextFunc
	KeyDownFn  KeyDownFunc
	KeyUpFn    KeyUpFunc
}

// NewInputService returns the default input service.
func NewInputService() *InputService {
	controller := input.NewController()
	return &InputService{
		PressKeyFn: controller.PressKey,
		TypeTextFn: controller.TypeText,
		KeyDownFn:  controller.KeyDown,
		KeyUpFn:    controller.KeyUp,
	}
}

// PressKey sends a key down+up event.
func (s *InputService) PressKey(ctx context.Context, key string, modifiers []string) error {
	if s == nil || s.PressKeyFn == nil {
		return fmt.Errorf("input service is not configured")
	}
	if err := s.PressKeyFn(ctx, key, modifiers); err != nil {
		return fmt.Errorf("press key: %w", err)
	}
	return nil
}

// TypeText types a string of text character by character.
func (s *InputService) TypeText(ctx context.Context, text string, delayMs int) error {
	if s == nil || s.TypeTextFn == nil {
		return fmt.Errorf("input service is not configured")
	}
	if err := s.TypeTextFn(ctx, text, delayMs); err != nil {
		return fmt.Errorf("type text: %w", err)
	}
	return nil
}

// KeyDown sends a key down event (for hold actions).
func (s *InputService) KeyDown(ctx context.Context, key string, modifiers []string) error {
	if s == nil || s.KeyDownFn == nil {
		return fmt.Errorf("input service is not configured")
	}
	if err := s.KeyDownFn(ctx, key, modifiers); err != nil {
		return fmt.Errorf("key down: %w", err)
	}
	return nil
}

// KeyUp sends a key up event (for hold actions).
func (s *InputService) KeyUp(ctx context.Context, key string, modifiers []string) error {
	if s == nil || s.KeyUpFn == nil {
		return fmt.Errorf("input service is not configured")
	}
	if err := s.KeyUpFn(ctx, key, modifiers); err != nil {
		return fmt.Errorf("key up: %w", err)
	}
	return nil
}
