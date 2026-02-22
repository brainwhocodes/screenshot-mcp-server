//go:build darwin
// +build darwin

package window

import (
	"context"
	"fmt"
	"image"
	"math"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/imgencode"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/screenshot"
)

// TakeWindowScreenshot captures a window and returns JPEG bytes with metadata
// Uses full screen capture and crop to window bounds
func TakeWindowScreenshot(ctx context.Context, windowID uint32, opts imgencode.Options) ([]byte, *ScreenshotMetadata, error) {
	encode := func(img image.Image) ([]byte, error) {
		return imgencode.EncodeJPEG(img, opts)
	}
	return captureWindowImage(ctx, windowID, encode)
}

// TakeWindowScreenshotImage captures a window and returns the raw cropped image with metadata.
func TakeWindowScreenshotImage(ctx context.Context, windowID uint32) (image.Image, *ScreenshotMetadata, error) {
	return captureWindowImageRaw(ctx, windowID)
}

func captureWindowImage(ctx context.Context, windowID uint32, encode func(image.Image) ([]byte, error)) ([]byte, *ScreenshotMetadata, error) {
	croppedImg, metadata, err := captureWindowImageRaw(ctx, windowID)
	if err != nil {
		return nil, nil, err
	}

	data, err := encode(croppedImg)
	if err != nil {
		return nil, nil, fmt.Errorf("encode screenshot: %w", err)
	}

	return data, metadata, nil
}

func captureWindowImageRaw(ctx context.Context, windowID uint32) (image.Image, *ScreenshotMetadata, error) {
	targetWindow, err := findWindowByID(ctx, windowID)
	if err != nil {
		return nil, nil, err
	}

	capturer := screenshot.NewCapturer()
	fullImg, err := capturer.Capture(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("capture screen: %w", err)
	}

	scale := getScaleForWindow(targetWindow.Bounds)
	cropRect := cropRectForWindow(targetWindow.Bounds, fullImg.Bounds(), scale)
	croppedImg := cropImage(fullImg, cropRect)

	return croppedImg, &ScreenshotMetadata{
		WindowID:    windowID,
		Bounds:      targetWindow.Bounds,
		ImageWidth:  cropRect.Dx(),
		ImageHeight: cropRect.Dy(),
		Scale:       scale,
	}, nil
}

// RegionMetadata contains metadata about a region screenshot
type RegionMetadata struct {
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	ImageWidth  int     `json:"image_width"`
	ImageHeight int     `json:"image_height"`
	Scale       float64 `json:"scale"`
	CoordSpace  string  `json:"coord_space"`
}

// TakeRegionScreenshot captures a region of the screen and returns JPEG bytes with metadata.
// coordSpace can be "points" (screen coordinates) or "pixels" (image coordinates).
// When coordSpace is "points", the region is specified in screen points (Quartz coordinates).
// When coordSpace is "pixels", the region is specified in pixel coordinates.
func TakeRegionScreenshot(ctx context.Context, x, y, width, height float64, coordSpace string, opts imgencode.Options) ([]byte, *RegionMetadata, error) {
	encode := func(img image.Image) ([]byte, error) {
		return imgencode.EncodeJPEG(img, opts)
	}
	return captureRegionScreenshot(ctx, x, y, width, height, coordSpace, encode)
}

// TakeWindowScreenshotPNG captures a window and returns PNG bytes (lossless).
func TakeWindowScreenshotPNG(ctx context.Context, windowID uint32) ([]byte, *ScreenshotMetadata, error) {
	encode := func(img image.Image) ([]byte, error) {
		return imgencode.EncodePNG(img)
	}
	return captureWindowImage(ctx, windowID, encode)
}

// TakeRegionScreenshotPNG captures a region and returns PNG bytes (lossless).
func TakeRegionScreenshotPNG(ctx context.Context, x, y, width, height float64, coordSpace string) ([]byte, *RegionMetadata, error) {
	encode := func(img image.Image) ([]byte, error) {
		return imgencode.EncodePNG(img)
	}
	return captureRegionScreenshot(ctx, x, y, width, height, coordSpace, encode)
}

func captureRegionScreenshot(ctx context.Context, x, y, width, height float64, coordSpace string, encode func(image.Image) ([]byte, error)) ([]byte, *RegionMetadata, error) {
	fullImg, err := screenshot.NewCapturer().Capture(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("capture screen: %w", err)
	}

	centerX := x + width/2
	centerY := y + height/2
	scale := scaleAtPoint(centerX, centerY)
	cropRect := cropRectForRegion(fullImg.Bounds(), x, y, width, height, scale, coordSpace)
	croppedImg := cropImage(fullImg, cropRect)

	data, err := encode(croppedImg)
	if err != nil {
		return nil, nil, fmt.Errorf("encode screenshot: %w", err)
	}

	metadata := &RegionMetadata{
		X:           float64(cropRect.Min.X) / scale,
		Y:           float64(cropRect.Min.Y) / scale,
		Width:       float64(cropRect.Dx()) / scale,
		Height:      float64(cropRect.Dy()) / scale,
		ImageWidth:  cropRect.Dx(),
		ImageHeight: cropRect.Dy(),
		Scale:       scale,
		CoordSpace:  coordSpace,
	}
	return data, metadata, nil
}

func cropRectForRegion(imgBounds image.Rectangle, x, y, width, height, scale float64, coordSpace string) image.Rectangle {
	var x1f, y1f, x2f, y2f float64
	if coordSpace == "pixels" {
		x1f, y1f, x2f, y2f = x, y, x+width, y+height
	} else {
		x1f, y1f, x2f, y2f = x*scale, y*scale, (x+width)*scale, (y+height)*scale
	}

	return clampRect(image.Rect(
		floatToIntBounded(x1f),
		floatToIntBounded(y1f),
		floatToIntBounded(x2f),
		floatToIntBounded(y2f),
	), imgBounds)
}

func getScaleForWindow(bounds Bounds) float64 {
	centerX := bounds.X + bounds.Width/2
	centerY := bounds.Y + bounds.Height/2
	return scaleAtPoint(centerX, centerY)
}

func cropRectForWindow(bounds Bounds, imgBounds image.Rectangle, scale float64) image.Rectangle {
	x1 := floatToIntBounded(bounds.X * scale)
	y1 := floatToIntBounded(bounds.Y * scale)
	x2 := floatToIntBounded((bounds.X + bounds.Width) * scale)
	y2 := floatToIntBounded((bounds.Y + bounds.Height) * scale)
	return clampRect(image.Rect(x1, y1, x2, y2), imgBounds)
}

func floatToIntBounded(value float64) int {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}

	maxInt := float64(int(^uint(0) >> 1))
	minInt := float64(-int(^uint(0)>>1) - 1)
	if value > maxInt {
		return int(^uint(0) >> 1)
	}
	if value < minInt {
		return -int(^uint(0)>>1) - 1
	}
	return int(value)
}

func clampRect(rect, bounds image.Rectangle) image.Rectangle {
	x1, y1, x2, y2 := rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y
	if x1 < bounds.Min.X {
		x1 = bounds.Min.X
	}
	if y1 < bounds.Min.Y {
		y1 = bounds.Min.Y
	}
	if x2 > bounds.Max.X {
		x2 = bounds.Max.X
	}
	if y2 > bounds.Max.Y {
		y2 = bounds.Max.Y
	}
	return image.Rect(x1, y1, x2, y2)
}

func cropImage(img image.Image, rect image.Rectangle) image.Image {
	return img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(rect)
}
