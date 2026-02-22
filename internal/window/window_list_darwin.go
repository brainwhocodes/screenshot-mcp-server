//go:build darwin
// +build darwin

package window

import (
	"context"
)

// ListWindows returns all visible windows.
func ListWindows(_ context.Context) ([]Window, error) {
	return listWindows()
}
