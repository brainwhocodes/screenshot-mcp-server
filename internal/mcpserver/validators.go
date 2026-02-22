package mcpserver

import "fmt"

const (
	defaultImageMatchThreshold = 0.8
	defaultComparisonThreshold = 0.95
	defaultImageWaitTimeoutMs  = 30_000
	defaultImageWaitPollMs     = 500
)

func defaultThreshold(value, fallback float64) float64 {
	if value == 0 {
		return fallback
	}
	return value
}

func validateThreshold(value float64) error {
	if value < 0 || value > 1 {
		return fmt.Errorf("threshold must be between 0.00 and 1.00")
	}
	return nil
}

func validateWindowID(windowID uint32) error {
	if windowID == 0 {
		return fmt.Errorf("window_id is required")
	}
	return nil
}

func validatePositiveDimensions(width, height float64) error {
	if width <= 0 {
		return fmt.Errorf("width is required and must be > 0")
	}
	if height <= 0 {
		return fmt.Errorf("height is required and must be > 0")
	}
	return nil
}

func normalizeCoordSpace(coordSpace string) (string, error) {
	if coordSpace == "" {
		return "points", nil
	}
	if coordSpace != "points" && coordSpace != "pixels" {
		return "", fmt.Errorf("coord_space must be 'points' or 'pixels', got %q", coordSpace)
	}
	return coordSpace, nil
}

func validateRegionInput(width, height float64, coordSpace string) error {
	if err := validatePositiveDimensions(width, height); err != nil {
		return err
	}
	_, err := normalizeCoordSpace(coordSpace)
	return err
}

func validateMaskRegions(maskRegions []MaskRegion) error {
	for _, region := range maskRegions {
		if region.Width < 0 || region.Height < 0 {
			return fmt.Errorf("mask regions must have non-negative width and height")
		}
	}
	return nil
}

func resolveTimeoutAndPoll(timeoutMs, pollMs int) (int, int) {
	if timeoutMs <= 0 {
		timeoutMs = defaultImageWaitTimeoutMs
	}
	if pollMs <= 0 {
		pollMs = defaultImageWaitPollMs
	}
	return timeoutMs, pollMs
}
