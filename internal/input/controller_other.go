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
		return ctx.Err()
	default:
	}
	return fmt.Errorf("key presses are only supported on macOS (darwin)")
}
