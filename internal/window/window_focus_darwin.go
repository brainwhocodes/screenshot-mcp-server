//go:build darwin
// +build darwin

package window

import (
	"context"
)

// FocusWindow brings a window to the foreground.
func FocusWindow(ctx context.Context, windowID uint32) error {
	return focusWindow(ctx, windowID)
}

func findWindowByID(_ context.Context, windowID uint32) (*Window, error) {
	return findWindowByIDCached(windowID)
}
