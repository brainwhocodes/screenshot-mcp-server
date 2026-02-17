package input

import "context"

// Controller provides OS-specific input injection primitives.
type Controller interface {
	// PressKey sends a key down+up event to the OS.
	//
	// key is a human-friendly name like "a", "enter", "tab", "escape", "left".
	// modifiers are optional strings like "shift", "control", "option", "command".
	PressKey(ctx context.Context, key string, modifiers []string) error
}
