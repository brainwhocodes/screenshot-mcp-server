//go:build darwin

package window

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"time"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/screenshot"
)

// WaitForPixel waits until the pixel at (x,y) matches the expected RGBA within tolerance.
// x, y are pixel coordinates in the window screenshot.
// rgba is the expected color as [4]uint8{R, G, B, A}.
// tolerance is the maximum difference per channel (0-255).
// timeoutMs is the maximum time to wait in milliseconds.
// pollIntervalMs is how often to check in milliseconds.
func WaitForPixel(ctx context.Context, windowID uint32, x, y float64, rgba [4]uint8, tolerance int, timeoutMs, pollIntervalMs int) error {
	if pollIntervalMs <= 0 {
		pollIntervalMs = 100
	}
	if timeoutMs <= 0 {
		timeoutMs = 5000
	}

	timeout := time.After(time.Duration(timeoutMs) * time.Millisecond)
	ticker := time.NewTicker(time.Duration(pollIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for pixel: %w", ctx.Err())
		case <-timeout:
			return fmt.Errorf("timeout waiting for pixel at (%.0f, %.0f) to match %v within tolerance %d", x, y, rgba, tolerance)
		case <-ticker.C:
			img, metadata, err := captureWindowImageForWait(ctx, windowID)
			if err != nil {
				continue
			}

			px := int(clampCoord(x, float64(metadata.ImageWidth)))
			py := int(clampCoord(y, float64(metadata.ImageHeight)))

			c := img.At(px, py)
			r, g, b, a := c.RGBA()

			// Convert from 16-bit to 8-bit
			r8 := int(r >> 8)
			g8 := int(g >> 8)
			b8 := int(b >> 8)
			a8 := int(a >> 8)

			if colorMatch(r8, g8, b8, a8, rgba, tolerance) {
				return nil
			}
		}
	}
}

// WaitForRegionStable waits until a region of the window stops changing.
// x, y, width, height define the region in pixel coordinates.
// stableCount is the number of consecutive identical screenshots required.
// timeoutMs is the maximum time to wait in milliseconds.
// pollIntervalMs is how often to check in milliseconds.
func WaitForRegionStable(ctx context.Context, windowID uint32, x, y, width, height float64, stableCount, timeoutMs, pollIntervalMs int) error {
	if pollIntervalMs <= 0 {
		pollIntervalMs = 100
	}
	if timeoutMs <= 0 {
		timeoutMs = 5000
	}
	if stableCount <= 0 {
		stableCount = 2
	}

	timeout := time.After(time.Duration(timeoutMs) * time.Millisecond)
	ticker := time.NewTicker(time.Duration(pollIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	var prevHash uint64
	consecutiveMatches := 0

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for region stability: %w", ctx.Err())
		case <-timeout:
			return fmt.Errorf("timeout waiting for region (%.0f, %.0f, %.0f, %.0f) to stabilize", x, y, width, height)
		case <-ticker.C:
			img, metadata, err := captureWindowImageForWait(ctx, windowID)
			if err != nil {
				continue
			}

			px := int(clampCoord(x, float64(metadata.ImageWidth)))
			py := int(clampCoord(y, float64(metadata.ImageHeight)))
			px2 := int(clampCoord(x+width, float64(metadata.ImageWidth)))
			py2 := int(clampCoord(y+height, float64(metadata.ImageHeight)))

			if px2 <= px || py2 <= py {
				continue
			}

			region := image.Rect(px, py, px2, py2)
			hash := imageHash(img, region)

			if hash == prevHash {
				consecutiveMatches++
				if consecutiveMatches >= stableCount {
					return nil
				}
			} else {
				consecutiveMatches = 0
			}
			prevHash = hash
		}
	}
}

func colorMatch(r, g, b, a int, expected [4]uint8, tolerance int) bool {
	return absDiff(int(r), int(expected[0])) <= tolerance &&
		absDiff(int(g), int(expected[1])) <= tolerance &&
		absDiff(int(b), int(expected[2])) <= tolerance &&
		absDiff(int(a), int(expected[3])) <= tolerance
}

func absDiff(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

func imageHash(img image.Image, region image.Rectangle) uint64 {
	bounds := img.Bounds()
	if region.Empty() || !bounds.Overlaps(region) {
		return 0
	}
	region = region.Intersect(bounds)

	var hash uint64
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			hash = hash*31 + uint64(r>>8) + uint64(g>>8)<<8 + uint64(b>>8)<<16 + uint64(a>>8)<<24
		}
	}
	return hash
}

func captureWindowImageForWait(ctx context.Context, windowID uint32) (image.Image, *ScreenshotMetadata, error) {
	targetWindow, err := findWindowByID(ctx, windowID)
	if err != nil {
		return nil, nil, fmt.Errorf("find window %d: %w", windowID, err)
	}

	capturer := screenshot.NewCapturer()
	fullImage, err := capturer.Capture(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("capture screen: %w", err)
	}

	scale := getScaleForWindow(targetWindow.Bounds)
	cropRect := cropRectForWindow(targetWindow.Bounds, fullImage.Bounds(), scale)
	croppedImage := cropImage(fullImage, cropRect)
	cropped := image.NewRGBA(croppedImage.Bounds())
	draw.Draw(cropped, croppedImage.Bounds(), croppedImage, croppedImage.Bounds().Min, draw.Src)

	return cropped, &ScreenshotMetadata{
		WindowID:    windowID,
		Bounds:      targetWindow.Bounds,
		ImageWidth:  cropRect.Dx(),
		ImageHeight: cropRect.Dy(),
		Scale:       scale,
	}, nil
}
