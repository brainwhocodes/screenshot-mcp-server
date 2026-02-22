package mcpserver

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type recordingState struct {
	mu     sync.Mutex
	active map[string]bool
}

func newRecordingState() *recordingState {
	return &recordingState{
		active: make(map[string]bool),
	}
}

func (state *recordingState) start(ctx context.Context, fps int, format string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("start recording: %w", err)
	}
	if format == "" {
		format = "mp4"
	}
	if fps <= 0 {
		return "", fmt.Errorf("fps must be > 0")
	}

	recordingID := fmt.Sprintf("rec_%d_%d_%s", time.Now().Unix(), fps, format)
	state.mu.Lock()
	state.active[recordingID] = true
	state.mu.Unlock()

	return recordingID, nil
}

func (state *recordingState) stop(recordingID string) (bool, error) {
	state.mu.Lock()
	defer state.mu.Unlock()

	if !state.active[recordingID] {
		return false, fmt.Errorf("recording %s not found or already stopped", recordingID)
	}

	delete(state.active, recordingID)
	return true, nil
}

// startRecording starts recording the screen.
// Placeholder implementation: full recording needs AVFoundation integration.
func startRecording(ctx context.Context, state *recordingState, _ uint32, fps int, format string) (string, error) {
	recordingID, err := state.start(ctx, fps, format)
	if err != nil {
		return "", err
	}
	return recordingID, newFeatureUnavailable(StartRecordingToolName, "AVFoundation integration not implemented")
}

// stopRecording stops a screen recording.
func stopRecording(_ context.Context, state *recordingState, recordingID string) error {
	if _, err := state.stop(recordingID); err != nil {
		return err
	}
	return newFeatureUnavailable(StopRecordingToolName, "AVFoundation integration not implemented")
}
