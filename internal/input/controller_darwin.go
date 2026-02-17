//go:build darwin

package input

/*
#cgo LDFLAGS: -framework ApplicationServices

#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdbool.h>

CGEventRef create_keyboard_event(CGKeyCode keyCode, bool keyDown) {
	return CGEventCreateKeyboardEvent(NULL, keyCode, keyDown);
}
*/
import "C"

import (
	"context"
	"fmt"
	"strings"
)

type darwinController struct{}

// NewController returns the macOS input controller.
func NewController() Controller {
	return darwinController{}
}

func (darwinController) PressKey(ctx context.Context, key string, modifiers []string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if C.AXIsProcessTrusted() == 0 {
		return fmt.Errorf("accessibility permission required (System Settings → Privacy & Security → Accessibility)")
	}

	normalizedKey := normalizeToken(key)
	if normalizedKey == "" {
		return fmt.Errorf("key is required")
	}

	keyCode, ok := darwinKeyCodes[normalizedKey]
	if !ok {
		return fmt.Errorf("unsupported key %q", key)
	}

	flags, err := darwinModifierFlags(modifiers)
	if err != nil {
		return err
	}

	if err := darwinPostKeyEvent(keyCode, flags, true); err != nil {
		return err
	}
	if err := darwinPostKeyEvent(keyCode, flags, false); err != nil {
		return err
	}

	return nil
}

func normalizeToken(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}

func darwinModifierFlags(raw []string) (C.CGEventFlags, error) {
	var flags C.CGEventFlags
	for _, item := range raw {
		switch normalizeToken(item) {
		case "":
			continue
		case "shift":
			flags |= C.kCGEventFlagMaskShift
		case "control", "ctrl":
			flags |= C.kCGEventFlagMaskControl
		case "option", "alt":
			flags |= C.kCGEventFlagMaskAlternate
		case "command", "cmd", "meta":
			flags |= C.kCGEventFlagMaskCommand
		case "fn":
			flags |= C.kCGEventFlagMaskSecondaryFn
		default:
			return 0, fmt.Errorf("unsupported modifier %q (supported: shift, control, option, command, fn)", item)
		}
	}
	return flags, nil
}

func darwinPostKeyEvent(keyCode uint16, flags C.CGEventFlags, keyDown bool) error {
	event := C.create_keyboard_event(C.CGKeyCode(keyCode), C.bool(keyDown))
	if event == 0 {
		return fmt.Errorf("failed to create keyboard event")
	}
	defer C.CFRelease(C.CFTypeRef(event))

	if flags != 0 {
		C.CGEventSetFlags(event, flags)
	}

	C.CGEventPost(C.kCGHIDEventTap, event)
	return nil
}

var darwinKeyCodes = map[string]uint16{
	// Letters
	"a": 0x00,
	"b": 0x0B,
	"c": 0x08,
	"d": 0x02,
	"e": 0x0E,
	"f": 0x03,
	"g": 0x05,
	"h": 0x04,
	"i": 0x22,
	"j": 0x26,
	"k": 0x28,
	"l": 0x25,
	"m": 0x2E,
	"n": 0x2D,
	"o": 0x1F,
	"p": 0x23,
	"q": 0x0C,
	"r": 0x0F,
	"s": 0x01,
	"t": 0x11,
	"u": 0x20,
	"v": 0x09,
	"w": 0x0D,
	"x": 0x07,
	"y": 0x10,
	"z": 0x06,

	// Digits
	"0": 0x1D,
	"1": 0x12,
	"2": 0x13,
	"3": 0x14,
	"4": 0x15,
	"5": 0x17,
	"6": 0x16,
	"7": 0x1A,
	"8": 0x1C,
	"9": 0x19,

	// Whitespace / control
	"space":  0x31,
	"tab":    0x30,
	"enter":  0x24,
	"return": 0x24,
	"escape": 0x35,
	"esc":    0x35,

	// Editing/navigation
	"backspace":     0x33,
	"delete":        0x33,
	"forwarddelete": 0x75,
	"home":          0x73,
	"end":           0x77,
	"pageup":        0x74,
	"pagedown":      0x79,

	// Arrows
	"left":  0x7B,
	"right": 0x7C,
	"down":  0x7D,
	"up":    0x7E,

	// Function keys
	"f1":  0x7A,
	"f2":  0x78,
	"f3":  0x63,
	"f4":  0x76,
	"f5":  0x60,
	"f6":  0x61,
	"f7":  0x62,
	"f8":  0x64,
	"f9":  0x65,
	"f10": 0x6D,
	"f11": 0x67,
	"f12": 0x6F,

	// Punctuation (US layout)
	"-":            0x1B,
	"minus":        0x1B,
	"=":            0x18,
	"equal":        0x18,
	"[":            0x21,
	"leftbracket":  0x21,
	"]":            0x1E,
	"rightbracket": 0x1E,
	"\\":           0x2A,
	"backslash":    0x2A,
	";":            0x29,
	"semicolon":    0x29,
	"'":            0x27,
	"quote":        0x27,
	",":            0x2B,
	"comma":        0x2B,
	".":            0x2F,
	"period":       0x2F,
	"/":            0x2C,
	"slash":        0x2C,
	"`":            0x32,
	"grave":        0x32,
}
