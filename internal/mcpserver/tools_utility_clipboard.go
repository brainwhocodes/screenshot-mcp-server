package mcpserver

import (
	"context"
	"fmt"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/safeexec"
)

// setClipboard sets the macOS clipboard content using pbcopy.
func setClipboard(ctx context.Context, text string) error {
	_, err := safeexec.RunCommandWithInput(ctx, []byte(text), "pbcopy")
	if err != nil {
		return fmt.Errorf("set clipboard: %w", err)
	}
	return nil
}

// getClipboard gets the macOS clipboard content using pbpaste.
func getClipboard(ctx context.Context) (string, error) {
	output, err := safeexec.RunCommand(ctx, "pbpaste")
	if err != nil {
		return "", fmt.Errorf("get clipboard: %w", err)
	}
	return string(output), nil
}
