package mcpserver

import "errors"

var (
	errFeatureUnavailable = errors.New("feature unavailable")
)

func newFeatureUnavailable(tool, reason string) error {
	return newToolError(tool, toolErrorCodeFeatureUnavailable, reason, errFeatureUnavailable)
}
