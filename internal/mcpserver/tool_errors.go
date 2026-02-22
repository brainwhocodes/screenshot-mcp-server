package mcpserver

import (
	"errors"
	"fmt"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/window"
)

// ToolErrorCode classifies MCP tool execution errors.
type ToolErrorCode string

const (
	toolErrorCodePermissionDenied   ToolErrorCode = "permission_denied"
	toolErrorCodeFeatureUnavailable ToolErrorCode = "feature_unavailable"
)

const permissionDeniedMessage = "required macOS permissions are not granted"

// ToolError carries machine-actionable error context for MCP tool failures.
type ToolError struct {
	ToolName string
	Code     ToolErrorCode
	Message  string
	Cause    error
}

func (e ToolError) Error() string {
	if e.ToolName == "" {
		if e.Cause == nil {
			return e.Message
		}
		return fmt.Sprintf("%s: %s", e.Message, e.Cause)
	}

	if e.Cause == nil {
		return fmt.Sprintf("%s: %s (%s)", e.ToolName, e.Message, e.Code)
	}
	return fmt.Sprintf("%s: %s (%s): %s", e.ToolName, e.Message, e.Code, e.Cause)
}

// Unwrap enables errors.Is/As checks through wrapped tool errors.
func (e ToolError) Unwrap() error {
	return e.Cause
}

func asToolError(err error) (ToolError, bool) {
	var te ToolError
	if errors.As(err, &te) {
		return te, true
	}
	var ptr *ToolError
	if errors.As(err, &ptr) && ptr != nil {
		return *ptr, true
	}
	return ToolError{}, false
}

func asPermissionError(err error) bool {
	var permissionErr *window.PermissionError
	return errors.As(err, &permissionErr)
}

func asFeatureUnavailableError(err error) bool {
	return errors.Is(err, errFeatureUnavailable)
}

func asToolExecutionError(toolName string, err error) error {
	if err == nil {
		return nil
	}

	if te, ok := asToolError(err); ok {
		if te.ToolName == "" {
			te.ToolName = toolName
		}
		return te
	}

	if asPermissionError(err) {
		return newToolError(toolName, toolErrorCodePermissionDenied, permissionDeniedMessage, err)
	}

	if asFeatureUnavailableError(err) {
		return newToolError(toolName, toolErrorCodeFeatureUnavailable, err.Error(), err)
	}

	return fmt.Errorf("%s: %w", toolName, err)
}

func newToolError(toolName string, code ToolErrorCode, message string, cause error) error {
	return ToolError{
		ToolName: toolName,
		Code:     code,
		Message:  message,
		Cause:    cause,
	}
}
