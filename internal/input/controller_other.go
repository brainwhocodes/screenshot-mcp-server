//go:build !darwin

package input

import (
	"context"
	"fmt"
)

type unsupportedController struct{}

// NewController returns an input controller that reports unsupported operations on non-macOS platforms.
func NewController() Controller {
	return unsupportedController{}
}

func (unsupportedController) PressKey(ctx context.Context, _ string, _ []string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}
	return fmt.Errorf("key presses are only supported on macOS (darwin)")
}

func (unsupportedController) TypeText(ctx context.Context, _ string, _ int) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}
	return fmt.Errorf("text typing is only supported on macOS (darwin)")
}

func (unsupportedController) KeyDown(ctx context.Context, _ string, _ []string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}
	return fmt.Errorf("key down is only supported on macOS (darwin)")
}

func (unsupportedController) KeyUp(ctx context.Context, _ string, _ []string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}
	return fmt.Errorf("key up is only supported on macOS (darwin)")
}
