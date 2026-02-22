// Package input exposes OS-specific input controllers used by the MCP layer.
package input

import "context"

// Controller provides OS-specific input injection primitives.
type Controller interface {
	// PressKey sends a key down+up event to the OS.
	//
	// key is a human-friendly name like "a", "enter", "tab", "escape", "left".
	// modifiers are optional strings like "shift", "control", "option", "command".
	PressKey(ctx context.Context, key string, modifiers []string) error

	// TypeText types a string of text character by character.
	// delayMs is the delay between keystrokes in milliseconds (0 for no delay).
	TypeText(ctx context.Context, text string, delayMs int) error

	// KeyDown sends a key down event (for hold actions).
	KeyDown(ctx context.Context, key string, modifiers []string) error

	// KeyUp sends a key up event (for hold actions).
	KeyUp(ctx context.Context, key string, modifiers []string) error
}
