package mcpserver

import (
	"context"
	"fmt"
	"image"
	"math"
	"os"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/imgencode"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/window"
)

// FixtureComparisonResult extends ImageComparisonResult with assertion info.
type FixtureComparisonResult struct {
	ImageComparisonResult
	FixturePath string `json:"fixture_path"`
	WindowID    uint32 `json:"window_id"`
}

// MaskRegion defines a region to ignore during comparison.
type MaskRegion struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// compareImages compares two images and returns similarity metrics.
func compareImages(image1Path, image2Path string, threshold float64) (*ImageComparisonResult, error) {
	return compareImagesWithMasks(image1Path, image2Path, threshold, nil)
}

// compareImagesWithMasks compares two images with optional masked regions.
func compareImagesWithMasks(image1Path, image2Path string, threshold float64, maskRegions []MaskRegion) (*ImageComparisonResult, error) {
	if err := validateComparisonPaths(image1Path, image2Path); err != nil {
		return nil, err
	}

	img1, err := decodeImageFromPath(image1Path)
	if err != nil {
		return nil, err
	}
	img2, err := decodeImageFromPath(image2Path)
	if err != nil {
		return nil, err
	}

	intersection := imageIntersection(img1.Bounds(), img2.Bounds())
	if intersection.Empty() {
		return &ImageComparisonResult{
			Similarity:  0,
			Match:       false,
			DiffPixels:  0,
			TotalPixels: 0,
		}, nil
	}

	maskRects := buildMaskRects(maskRegions, intersection)
	diffPixels, totalPixels := comparePixelDiff(img1, img2, intersection, maskRects)
	if totalPixels == 0 {
		return &ImageComparisonResult{
			Similarity:  1,
			Match:       true,
			DiffPixels:  0,
			TotalPixels: 0,
		}, nil
	}

	similarity := 1.0 - float64(diffPixels)/float64(totalPixels)
	return &ImageComparisonResult{
		Similarity:  similarity,
		Match:       similarity >= threshold,
		DiffPixels:  diffPixels,
		TotalPixels: totalPixels,
	}, nil
}

func validateComparisonPaths(image1Path, image2Path string) error {
	if err := ValidatePathAllowed(image1Path); err != nil {
		return fmt.Errorf("image1 path not allowed: %w", err)
	}
	if err := ValidatePathAllowed(image2Path); err != nil {
		return fmt.Errorf("image2 path not allowed: %w", err)
	}
	return nil
}

func decodeImageFromPath(path string) (image.Image, error) {
	// #nosec G304 -- path is validated by ValidatePathAllowed before open.
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode image %q: %w", path, err)
	}
	return img, nil
}

func imageIntersection(a, b image.Rectangle) image.Rectangle {
	return a.Intersect(b)
}

func comparePixelDiff(img1, img2 image.Image, rect image.Rectangle, maskRects []image.Rectangle) (diffPixels int, totalPixels int) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if isMaskedPoint(x, y, maskRects) {
				continue
			}
			pixel1 := toRGBA8(img1.At(x, y))
			pixel2 := toRGBA8(img2.At(x, y))

			diff := abs(int(pixel1.R)-int(pixel2.R)) + abs(int(pixel1.G)-int(pixel2.G)) + abs(int(pixel1.B)-int(pixel2.B))
			if diff > 30 {
				diffPixels++
			}
			totalPixels++
		}
	}
	return diffPixels, totalPixels
}

func buildMaskRects(maskRegions []MaskRegion, imageBounds image.Rectangle) []image.Rectangle {
	masked := make([]image.Rectangle, 0, len(maskRegions))
	for _, region := range maskRegions {
		rect, ok := maskRegionToRect(region)
		if !ok {
			continue
		}
		rect = rect.Intersect(imageBounds)
		if rect.Empty() || rect.Dx() <= 0 || rect.Dy() <= 0 {
			continue
		}
		masked = append(masked, rect)
	}
	return masked
}

func maskRegionToRect(region MaskRegion) (image.Rectangle, bool) {
	xMin, ok := safeFloatToInt(region.X)
	if !ok {
		return image.Rectangle{}, false
	}
	yMin, ok := safeFloatToInt(region.Y)
	if !ok {
		return image.Rectangle{}, false
	}
	xMax, ok := safeFloatToInt(region.X + region.Width)
	if !ok {
		return image.Rectangle{}, false
	}
	yMax, ok := safeFloatToInt(region.Y + region.Height)
	if !ok {
		return image.Rectangle{}, false
	}
	return image.Rect(xMin, yMin, xMax, yMax), true
}

func safeFloatToInt(value float64) (int, bool) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, false
	}
	if value > float64(int(^uint(0)>>1)) || value < float64(-int(^uint(0)>>1)-1) {
		return 0, false
	}
	return int(value), true
}

func isMaskedPoint(x, y int, maskRects []image.Rectangle) bool {
	for _, rect := range maskRects {
		if x >= rect.Min.X && x < rect.Max.X && y >= rect.Min.Y && y < rect.Max.Y {
			return true
		}
	}
	return false
}

// assertScreenshotMatchesFixture compares a window screenshot to a golden fixture.
func assertScreenshotMatchesFixture(ctx context.Context, windowID uint32, fixturePath string, threshold float64, maskRegions []MaskRegion) (*FixtureComparisonResult, error) {
	if err := ValidatePathAllowed(fixturePath); err != nil {
		return nil, fmt.Errorf("fixture path not allowed: %w", err)
	}
	if windowID == 0 {
		return nil, fmt.Errorf("window_id is required")
	}
	threshold = defaultThreshold(threshold, defaultComparisonThreshold)
	if err := validateThreshold(threshold); err != nil {
		return nil, err
	}

	for _, region := range maskRegions {
		if region.Width < 0 || region.Height < 0 {
			return nil, fmt.Errorf("mask regions must have non-negative width and height")
		}
	}

	screenshotData, _, err := window.TakeWindowScreenshot(ctx, windowID, imgencode.DefaultOptions)
	if err != nil {
		return nil, fmt.Errorf("capture window screenshot: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "window-screenshot-*.jpg")
	if err != nil {
		return nil, fmt.Errorf("create temp screenshot: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("close temp screenshot file: %w", err)
	}
	if err := os.WriteFile(tmpFile.Name(), screenshotData, 0o600); err != nil {
		_ = os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("write temp screenshot: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	comparison, err := compareImagesWithMasks(tmpFile.Name(), fixturePath, threshold, maskRegions)
	if err != nil {
		return nil, err
	}

	return &FixtureComparisonResult{
		ImageComparisonResult: *comparison,
		FixturePath:           fixturePath,
		WindowID:              windowID,
	}, nil
}
