package screenshot

import (
	"context"
	"fmt"
	"image"
	"image/draw"

	"github.com/kbinani/screenshot"
)

// Capturer captures the full virtual screen as an image.
type Capturer interface {
	Capture(context.Context) (image.Image, error)
}

// SystemCapturer captures the active displays from the local machine.
type SystemCapturer struct{}

// NewCapturer returns the default screen capturer.
func NewCapturer() Capturer {
	return SystemCapturer{}
}

// Capture captures all active displays into one image.
func (SystemCapturer) Capture(_ context.Context) (image.Image, error) {
	displayCount := screenshot.NumActiveDisplays()
	if displayCount <= 0 {
		return nil, fmt.Errorf("no active displays available")
	}

	unionBounds := screenshot.GetDisplayBounds(0)
	for i := 1; i < displayCount; i++ {
		unionBounds = unionRect(unionBounds, screenshot.GetDisplayBounds(i))
	}

	canvas := image.NewRGBA(image.Rect(0, 0, unionBounds.Dx(), unionBounds.Dy()))
	for i := 0; i < displayCount; i++ {
		displayBounds := screenshot.GetDisplayBounds(i)
		captured, err := screenshot.CaptureRect(displayBounds)
		if err != nil {
			return nil, fmt.Errorf("capture display %d: %w", i, err)
		}

		dst := image.Rect(
			displayBounds.Min.X-unionBounds.Min.X,
			displayBounds.Min.Y-unionBounds.Min.Y,
			displayBounds.Max.X-unionBounds.Min.X,
			displayBounds.Max.Y-unionBounds.Min.Y,
		)
		draw.Draw(canvas, dst, captured, captured.Bounds().Min, draw.Src)
	}

	return canvas, nil
}

func unionRect(a, b image.Rectangle) image.Rectangle {
	minX := a.Min.X
	if b.Min.X < minX {
		minX = b.Min.X
	}
	minY := a.Min.Y
	if b.Min.Y < minY {
		minY = b.Min.Y
	}
	maxX := a.Max.X
	if b.Max.X > maxX {
		maxX = b.Max.X
	}
	maxY := a.Max.Y
	if b.Max.Y > maxY {
		maxY = b.Max.Y
	}
	return image.Rect(minX, minY, maxX, maxY)
}
