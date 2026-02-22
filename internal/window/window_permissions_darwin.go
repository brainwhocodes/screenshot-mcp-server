//go:build darwin
// +build darwin

package window

import "fmt"

// CheckPermissions checks if required permissions are granted.
func CheckPermissions() (screenRecording bool, accessibility bool) {
	screenRecording, accessibility = permissionState()
	return
}

// PermissionError indicates missing macOS permissions required by the tool.
type PermissionError struct {
	ToolName      string
	Screen        bool
	Accessibility bool
}

func (e *PermissionError) Error() string {
	if e.ToolName == "" {
		return "required macOS permissions are not granted"
	}
	return fmt.Sprintf("%s requires macOS permissions: screen recording=%t accessibility=%t. enable both in System Settings > Privacy & Security", e.ToolName, e.Screen, e.Accessibility)
}

// EnsureAutomationPermissions returns an explicit error when screen recording/accessibility are missing.
func EnsureAutomationPermissions(toolName string) error {
	screenRecording, accessibility := CheckPermissions()
	if screenRecording && accessibility {
		return nil
	}
	return &PermissionError{
		ToolName:      toolName,
		Screen:        screenRecording,
		Accessibility: accessibility,
	}
}
