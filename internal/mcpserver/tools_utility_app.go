package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/safeexec"
)

// restartApp quits and relaunches an application.
func restartApp(ctx context.Context, appName string) error {
	if appName == "" {
		return fmt.Errorf("app name is required")
	}
	if err := safeexec.ValidateCommandArg(appName); err != nil {
		return fmt.Errorf("invalid app name %q: %w", appName, err)
	}

	quotedApp := safeexec.QuoteAppleScriptString(appName)

	// Best-effort quit; if not running we continue with the launch attempt.
	_ = safeexec.RunAppleScript(ctx, fmt.Sprintf(`tell application "%s" to quit`, quotedApp))

	// Wait a moment for the app to quit.
	time.Sleep(500 * time.Millisecond)

	// Launch the app using osascript.
	err := safeexec.RunAppleScript(ctx, fmt.Sprintf(`tell application "%s" to activate`, quotedApp))
	if err != nil {
		return fmt.Errorf("launch app %q: %w", appName, err)
	}

	return nil
}
