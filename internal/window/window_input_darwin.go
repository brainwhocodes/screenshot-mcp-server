//go:build darwin
// +build darwin

package window

import (
	"context"
	"fmt"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/imgencode"
)

// MouseMove moves the mouse cursor to the specified coordinates.
// x, y are pixel coordinates in the screenshot image.
func MouseMove(ctx context.Context, windowID uint32, x, y float64) error {
	_, _, xPt, yPt, err := mapWindowInputPoint(ctx, windowID, x, y)
	if err != nil {
		return err
	}

	postMouseMoveEvent(xPt, yPt)
	return nil
}

// MouseDown sends a mouse down event at the specified coordinates.
func MouseDown(ctx context.Context, windowID uint32, x, y float64, button string) error {
	return postMouseButton(ctx, windowID, x, y, button, func(xPt, yPt float64, btn int) {
		postMouseDownEvent(xPt, yPt, btn)
	})
}

// ClickAt performs a mouse click at screen coordinates.
func ClickAt(ctx context.Context, x, y float64, button string, clicks int, coordSpace string) error {
	pointX, pointY, err := mapScreenPoint(ctx, x, y, coordSpace)
	if err != nil {
		return err
	}

	if clicks <= 0 {
		clicks = 1
	}

	btn := buttonToInt(button)
	postMouseClickEvent(pointX, pointY, btn, clicks)
	return nil
}

// MouseUp sends a mouse up event at the specified coordinates.
func MouseUp(ctx context.Context, windowID uint32, x, y float64, button string) error {
	return postMouseButton(ctx, windowID, x, y, button, func(xPt, yPt float64, btn int) {
		postMouseUpEvent(xPt, yPt, btn)
	})
}

// Drag performs a drag operation from one point to another.
func Drag(ctx context.Context, windowID uint32, fromX, fromY, toX, toY float64, button string) error {
	_, _, fromXPt, fromYPt, toXPt, toYPt, err := mapWindowDragPoints(ctx, windowID, fromX, fromY, toX, toY)
	if err != nil {
		return err
	}
	btn := buttonToInt(button)
	postMouseDownEvent(fromXPt, fromYPt, btn)
	postMouseMoveEvent(toXPt, toYPt)
	postMouseUpEvent(toXPt, toYPt, btn)
	return nil
}

// Scroll performs a scroll operation at the specified coordinates.
// deltaX and deltaY are in pixels (positive = right/down).
func Scroll(ctx context.Context, windowID uint32, x, y, deltaX, deltaY float64) error {
	_, _, xPt, yPt, err := mapWindowInputPoint(ctx, windowID, x, y)
	if err != nil {
		return err
	}
	postScrollEvent(xPt, yPt, deltaX, deltaY)
	return nil
}

func postMouseButton(
	ctx context.Context,
	windowID uint32,
	x, y float64,
	button string,
	send func(float64, float64, int),
) error {
	_, _, xPt, yPt, err := mapWindowInputPoint(ctx, windowID, x, y)
	if err != nil {
		return err
	}
	btn := buttonToInt(button)
	send(xPt, yPt, btn)
	return nil
}

func mapWindowInputPoint(ctx context.Context, windowID uint32, x, y float64) (*Window, *ScreenshotMetadata, float64, float64, error) {
	windows, err := ListWindows(ctx)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("list windows: %w", err)
	}

	var targetWindow *Window
	for i := range windows {
		if windows[i].WindowID == windowID {
			targetWindow = &windows[i]
			break
		}
	}

	if targetWindow == nil {
		return nil, nil, 0, 0, fmt.Errorf("window %d not found", windowID)
	}

	_, metadata, err := TakeWindowScreenshot(ctx, windowID, imgencode.Options{Quality: 1})
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("take screenshot for coordinate mapping: %w", err)
	}

	x = clampCoord(x, float64(metadata.ImageWidth))
	y = clampCoord(y, float64(metadata.ImageHeight))

	xPt := targetWindow.Bounds.X + (x / metadata.Scale)
	yPt := targetWindow.Bounds.Y + (y / metadata.Scale)

	return targetWindow, metadata, xPt, yPt, nil
}

func mapScreenPoint(_ context.Context, x, y float64, coordSpace string) (float64, float64, error) {
	if coordSpace == "" || coordSpace == "points" {
		return x, y, nil
	}
	if coordSpace != "pixels" {
		return 0, 0, fmt.Errorf("coord_space must be 'points' or 'pixels', got %q", coordSpace)
	}

	scale := scaleAtPoint(x, y)
	if scale <= 0 {
		scale = 1.0
	}
	return x / scale, y / scale, nil
}

func mapWindowDragPoints(
	ctx context.Context,
	windowID uint32,
	fromX, fromY, toX, toY float64,
) (*Window, *ScreenshotMetadata, float64, float64, float64, float64, error) {
	targetWindow, metadata, fromXPt, fromYPt, err := mapWindowInputPoint(ctx, windowID, fromX, fromY)
	if err != nil {
		return nil, nil, 0, 0, 0, 0, err
	}

	toX = clampCoord(toX, float64(metadata.ImageWidth))
	toY = clampCoord(toY, float64(metadata.ImageHeight))

	toXPt := targetWindow.Bounds.X + (toX / metadata.Scale)
	toYPt := targetWindow.Bounds.Y + (toY / metadata.Scale)

	return targetWindow, metadata, fromXPt, fromYPt, toXPt, toYPt, nil
}

// Click performs a mouse click at the specified coordinates.
// x, y are pixel coordinates in the screenshot image.
func Click(ctx context.Context, windowID uint32, x, y float64, button string, clicks int) error {
	targetWindow, _, xPt, yPt, err := mapWindowInputPoint(ctx, windowID, x, y)
	if err != nil {
		return err
	}

	// Verify point is within window bounds
	if xPt < targetWindow.Bounds.X || xPt >= targetWindow.Bounds.X+targetWindow.Bounds.Width ||
		yPt < targetWindow.Bounds.Y || yPt >= targetWindow.Bounds.Y+targetWindow.Bounds.Height {
		return fmt.Errorf("click coordinates outside window bounds")
	}

	btn := buttonToInt(button)
	postMouseClickEvent(xPt, yPt, btn, clicks)

	return nil
}

func clampCoord(val, maxValue float64) float64 {
	if val < 0 {
		return 0
	}
	if val >= maxValue {
		return maxValue - 1
	}
	return val
}

func buttonToInt(button string) int {
	switch button {
	case "right":
		return 1
	case "middle":
		return 2
	default:
		return 0
	}
}
