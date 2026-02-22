package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/safeexec"
	"github.com/brainwhocodes/screenshot_mcp_server/internal/tools"
	"github.com/brainwhocodes/screenshot_mcp_server/internal/window"
)

type recordingState struct {
	mu     sync.Mutex
	active map[string]*recordingSession
}

func newRecordingState() *recordingState {
	return &recordingState{
		active: make(map[string]*recordingSession),
	}
}

type recordingSession struct {
	id              string
	windowID        uint32
	fps             int
	format          string
	outputPath      string
	frameDir        string
	service         ScreenshotService
	done            chan struct{}
	cancel          context.CancelFunc
	recordingError  error
	recordingFrames int
	frameFiles      []string
}

func (state *recordingState) start(ctx context.Context, service ScreenshotService, windowID uint32, fps int, format string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("start recording: %w", err)
	}
	if service == nil {
		service = tools.NewScreenshotService()
	}
	format = strings.TrimSpace(strings.ToLower(format))
	if format == "" {
		format = "mp4"
	}
	if fps <= 0 {
		return "", fmt.Errorf("fps must be > 0")
	}
	if _, err := normalizeRecordingFormat(format); err != nil {
		return "", err
	}

	recordingID := fmt.Sprintf("rec_%d_%d_%s", time.Now().UnixNano(), fps, format)
	frameDir, err := os.MkdirTemp("", "screenshot-recording-"+recordingID+"-")
	if err != nil {
		return "", fmt.Errorf("create recording directory: %w", err)
	}
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.%s", recordingID, format))

	recordingCtx, cancel := context.WithCancel(ctx)
	session := &recordingSession{
		id:         recordingID,
		windowID:   windowID,
		fps:        fps,
		format:     format,
		outputPath: outputPath,
		frameDir:   frameDir,
		service:    service,
		done:       make(chan struct{}),
		cancel:     cancel,
	}
	state.mu.Lock()
	state.active[recordingID] = session
	state.mu.Unlock()

	go runRecordingSession(recordingCtx, session)

	return recordingID, nil
}

func (state *recordingState) stop(ctx context.Context, recordingID string) (string, string, error) {
	state.mu.Lock()
	session, exists := state.active[recordingID]
	if !exists {
		state.mu.Unlock()
		return "", "", fmt.Errorf("recording %s not found or already stopped", recordingID)
	}
	delete(state.active, recordingID)
	state.mu.Unlock()

	session.cancel()
	<-session.done

	defer func() {
		_ = os.RemoveAll(session.frameDir)
	}()

	if session.recordingError != nil && !errors.Is(session.recordingError, context.Canceled) {
		return "", "", session.recordingError
	}
	if session.recordingFrames == 0 {
		return "", "", fmt.Errorf("no frames captured for recording %s", recordingID)
	}

	recordingPath, warning := session.outputPath, ""
	videoPath, warning, err := finalizeRecordingSession(ctx, session)
	if err != nil {
		return "", "", err
	}
	if videoPath != "" {
		recordingPath = videoPath
	}
	return recordingPath, warning, nil
}

func normalizeRecordingFormat(format string) (string, error) {
	switch format {
	case "gif", "mp4":
		return format, nil
	default:
		return "", fmt.Errorf("unsupported recording format %q; supported formats are mp4, gif", format)
	}
}

func runRecordingSession(ctx context.Context, session *recordingSession) {
	defer close(session.done)

	ticker := time.NewTicker(time.Second / time.Duration(session.fps))
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}

		if err := captureAndPersistFrame(ctx, session); err != nil {
			session.recordingError = err
			return
		}
		session.recordingFrames++

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func finalizeRecordingSession(ctx context.Context, session *recordingSession) (string, string, error) {
	switch session.format {
	case "mp4":
		if err := encodeRecordingToMP4(ctx, session); err != nil {
			fallbackPath := strings.TrimSuffix(session.outputPath, ".mp4") + ".gif"
			if encodeErr := encodeRecordingToGIF(session, fallbackPath); encodeErr != nil {
				return "", "", err
			}
			return fallbackPath, "ffmpeg not installed or failed; encoded as animated GIF", nil
		}
		return session.outputPath, "", nil
	case "gif":
		if err := encodeRecordingToGIF(session, session.outputPath); err != nil {
			return "", "", err
		}
		return session.outputPath, "", nil
	default:
		return "", "", fmt.Errorf("unsupported output format %q", session.format)
	}
}

func encodeRecordingToGIF(session *recordingSession, outputPath string) error {
	if len(session.frameFiles) == 0 {
		return fmt.Errorf("no frames captured for recording %s", session.id)
	}

	animated := &gif.GIF{
		Image: make([]*image.Paletted, 0, len(session.frameFiles)),
	}
	for _, framePath := range session.frameFiles {
		frame, err := loadPNGImage(framePath)
		if err != nil {
			return err
		}

		paletted := image.NewPaletted(frame.Bounds(), palette.Plan9)
		draw.FloydSteinberg.Draw(paletted, frame.Bounds(), frame, image.Point{})
		delay := int(math.Round(100.0 / float64(session.fps)))
		if delay < 1 {
			delay = 1
		}
		animated.Image = append(animated.Image, paletted)
		animated.Delay = append(animated.Delay, delay)
		animated.Disposal = append(animated.Disposal, gif.DisposalBackground)
	}

	// #nosec G304 -- outputPath is derived from a generated temp path.
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create recording file: %w", err)
	}
	defer func() {
		_ = out.Close()
	}()

	if err := gif.EncodeAll(out, animated); err != nil {
		return fmt.Errorf("encode gif: %w", err)
	}
	return nil
}

func encodeRecordingToMP4(ctx context.Context, session *recordingSession) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not installed")
	}

	inputPattern := filepath.Join(session.frameDir, "frame_%06d.png")
	outputPath := session.outputPath
	args := []string{
		"-y",
		"-framerate",
		fmt.Sprintf("%d", session.fps),
		"-i",
		inputPattern,
		"-c:v",
		"libx264",
		"-pix_fmt",
		"yuv420p",
		"-movflags",
		"+faststart",
		outputPath,
	}
	if _, err := safeexec.RunCommandWithTimeout(ctx, 60*time.Second, "ffmpeg", args...); err != nil {
		return fmt.Errorf("ffmpeg encode failed: %w", err)
	}

	return nil
}

func captureAndPersistFrame(ctx context.Context, session *recordingSession) error {
	img, err := captureImageForRecording(ctx, session.service, session.windowID)
	if err != nil {
		return fmt.Errorf("capture frame: %w", err)
	}

	framePath := filepath.Join(session.frameDir, fmt.Sprintf("frame_%06d.png", session.recordingFrames))
	// #nosec G304 -- framePath is derived from a generated temp directory.
	out, err := os.Create(framePath)
	if err != nil {
		return fmt.Errorf("create frame file: %w", err)
	}
	defer func() {
		_ = out.Close()
	}()
	if err := png.Encode(out, img); err != nil {
		return fmt.Errorf("encode frame: %w", err)
	}

	session.frameFiles = append(session.frameFiles, framePath)
	return nil
}

func captureImageForRecording(ctx context.Context, service ScreenshotService, windowID uint32) (image.Image, error) {
	if windowID == 0 {
		capturedImage, err := service.CaptureImage(ctx)
		if err != nil {
			return nil, fmt.Errorf("capture screen image: %w", err)
		}
		return capturedImage, nil
	}

	croppedImage, _, err := window.TakeWindowScreenshotImage(ctx, windowID)
	if err != nil {
		return nil, fmt.Errorf("capture window image: %w", err)
	}
	return croppedImage, nil
}

func loadPNGImage(path string) (image.Image, error) {
	// #nosec G304 -- path comes from generated temp files.
	in, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open frame image %q: %w", path, err)
	}
	defer func() {
		_ = in.Close()
	}()

	decoded, err := png.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("decode frame image %q: %w", path, err)
	}
	return decoded, nil
}

// startRecording starts screen capture into a background recording job.
func startRecording(ctx context.Context, state *recordingState, windowID uint32, fps int, format string) (string, error) {
	recordingID, err := state.start(ctx, nil, windowID, fps, format)
	if err != nil {
		return "", err
	}
	return recordingID, nil
}

// stopRecording stops a background recording job and returns the output path.
func stopRecording(ctx context.Context, state *recordingState, recordingID string) (string, string, error) {
	videoPath, warning, err := state.stop(ctx, recordingID)
	if err != nil {
		return "", "", err
	}
	return videoPath, warning, nil
}
