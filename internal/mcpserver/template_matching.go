package mcpserver

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"math"
	"os"
	"time"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/imgencode"
	screencap "github.com/codingthefuturewithai/screenshot_mcp_server/internal/screenshot"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/window"
)

// waitForImageMatch waits for a template image to appear on screen.
func waitForImageMatch(ctx context.Context, windowID uint32, templatePath string, threshold float64, timeoutMs, pollIntervalMs int) (*Point, error) {
	threshold = defaultThreshold(threshold, defaultImageMatchThreshold)
	if err := validateThreshold(threshold); err != nil {
		return nil, fmt.Errorf("wait for image match: %w", err)
	}
	timeoutMs, pollIntervalMs = resolveTimeoutAndPoll(timeoutMs, pollIntervalMs, defaultImageWaitTimeoutMs, defaultImageWaitPollMs)

	deadline := time.Now().Add(time.Duration(timeoutMs) * time.Millisecond)
	for time.Now().Before(deadline) {
		matches, err := findImageMatches(ctx, windowID, templatePath, threshold)
		if err == nil && len(matches) > 0 {
			return &Point{X: matches[0].X, Y: matches[0].Y}, nil
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for image match: %w", ctx.Err())
		case <-time.After(time.Duration(pollIntervalMs) * time.Millisecond):
			// Continue polling.
		}
	}

	return nil, fmt.Errorf("template image not found within timeout")
}

// findImageMatches finds all occurrences of a template image using normalized cross-correlation.
func findImageMatches(ctx context.Context, windowID uint32, templatePath string, threshold float64) ([]ImageMatch, error) {
	if err := ValidatePathAllowed(templatePath); err != nil {
		return nil, fmt.Errorf("template path not allowed: %w", err)
	}

	// #nosec G304 -- templatePath is validated by path allowlist before open.
	templateFile, err := os.Open(templatePath)
	if err != nil {
		return nil, fmt.Errorf("open template image: %w", err)
	}
	defer func() { _ = templateFile.Close() }()

	templateImg, _, err := image.Decode(templateFile)
	if err != nil {
		return nil, fmt.Errorf("decode template image: %w", err)
	}

	var sourceImage image.Image
	switch windowID {
	case 0:
		capturer := screencap.NewCapturer()
		src, err := capturer.Capture(ctx)
		if err != nil {
			return nil, fmt.Errorf("capture screenshot: %w", err)
		}
		sourceImage = src
	default:
		screenshotData, _, err := window.TakeWindowScreenshot(ctx, windowID, imgencode.DefaultOptions)
		if err != nil {
			return nil, fmt.Errorf("capture window screenshot: %w", err)
		}
		sourceImage, err = decodeJPEGImage(screenshotData)
		if err != nil {
			return nil, fmt.Errorf("decode window screenshot: %w", err)
		}
	}

	return performTemplateMatching(sourceImage, templateImg, threshold), nil
}

func performTemplateMatching(screenshot, template image.Image, threshold float64) []ImageMatch {
	var matches []ImageMatch

	screenBounds := screenshot.Bounds()
	templateBounds := template.Bounds()

	templateWidth := templateBounds.Max.X - templateBounds.Min.X
	templateHeight := templateBounds.Max.Y - templateBounds.Min.Y
	if templateWidth <= 0 || templateHeight <= 0 {
		return matches
	}

	templateGray := make([]float64, templateWidth*templateHeight)
	var templateMean float64
	for y := 0; y < templateHeight; y++ {
		for x := 0; x < templateWidth; x++ {
			r, g, b, _ := template.At(templateBounds.Min.X+x, templateBounds.Min.Y+y).RGBA()
			gray := float64((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))) / 65535.0
			templateGray[y*templateWidth+x] = gray
			templateMean += gray
		}
	}
	templateMean /= float64(templateWidth * templateHeight)
	for i := range templateGray {
		templateGray[i] -= templateMean
	}

	var templateStdDev float64
	for _, v := range templateGray {
		templateStdDev += v * v
	}
	templateStdDev = math.Sqrt(templateStdDev)
	if templateStdDev == 0 {
		return matches
	}

	maxY := screenBounds.Max.Y - templateHeight
	maxX := screenBounds.Max.X - templateWidth
	for y := screenBounds.Min.Y; y <= maxY; y += 2 {
		for x := screenBounds.Min.X; x <= maxX; x += 2 {
			score := computeNCC(screenshot, x, y, templateWidth, templateHeight, templateGray, templateStdDev)
			if score >= threshold {
				matches = append(matches, ImageMatch{
					X:      float64(x),
					Y:      float64(y),
					Width:  float64(templateWidth),
					Height: float64(templateHeight),
					Score:  score,
				})
			}
		}
	}

	return matches
}

func computeNCC(img image.Image, startX, startY, width, height int, templateGray []float64, templateStdDev float64) float64 {
	var regionMean float64
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(startX+x, startY+y).RGBA()
			gray := float64((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))) / 65535.0
			regionMean += gray
		}
	}
	regionMean /= float64(width * height)

	var numerator, regionStdDev float64
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(startX+x, startY+y).RGBA()
			gray := float64((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))) / 65535.0
			centered := gray - regionMean
			templateVal := templateGray[y*width+x]
			numerator += centered * templateVal
			regionStdDev += centered * centered
		}
	}

	regionStdDev = math.Sqrt(regionStdDev)
	if regionStdDev == 0 || templateStdDev == 0 {
		return 0
	}

	return numerator / (regionStdDev * templateStdDev)
}

func decodeJPEGImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return img, nil
}
