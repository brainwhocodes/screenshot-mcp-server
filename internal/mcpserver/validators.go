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

func resolveTimeoutAndPoll(timeoutMs, pollMs int) (int, int) {
	if timeoutMs <= 0 {
		timeoutMs = defaultImageWaitTimeoutMs
	}
	if pollMs <= 0 {
		pollMs = defaultImageWaitPollMs
	}
	return timeoutMs, pollMs
}
