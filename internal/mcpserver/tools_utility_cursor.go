package mcpserver

// takeScreenshotWithCursor captures a screenshot including the mouse cursor.
func takeScreenshotWithCursor() error {
	return newFeatureUnavailable(TakeScreenshotWithCursorToolName, "cursor capture integration not implemented")
}
