package tools

import (
	"context"
	"fmt"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/input"
)

// PressKeyFunc sends a key press to the OS.
type PressKeyFunc func(context.Context, string, []string) error

// InputService wraps OS input injection primitives.
type InputService struct {
	PressKeyFn PressKeyFunc
}

// NewInputService returns the default input service.
func NewInputService() *InputService {
	controller := input.NewController()
	return &InputService{
		PressKeyFn: controller.PressKey,
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
