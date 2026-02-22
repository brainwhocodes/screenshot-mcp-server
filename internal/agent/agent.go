// Package agent implements the screenshot-driven automation loop and tool abstractions.
package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/imgencode"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/input"
	"github.com/codingthefuturewithai/screenshot_mcp_server/internal/window"
)

// Config defines how the automation agent runs.
type Config struct {
	Goal           string
	WindowTitle    string
	OwnerName      string
	MaxSteps       int
	MaxDuration    time.Duration
	RunDir         string
	SaveArtifacts  bool
	VisionClient   VisionClient
	ActionExecutor ActionExecutor
}

// VisionClient returns the next action from an LLM or policy engine.
type VisionClient interface {
	GetAction(ctx context.Context, screenshot []byte, goal string) (*Action, error)
}

// ActionExecutor executes normalized agent actions on the target window.
type ActionExecutor interface {
	FocusWindow(ctx context.Context, windowID uint32) error
	Click(ctx context.Context, windowID uint32, x, y float64, button string, clicks int) error
	PressKey(ctx context.Context, windowID uint32, key string, modifiers []string) error
}

// WindowActionExecutor executes actions directly via window package helpers.
type WindowActionExecutor struct{}

// FocusWindow focuses the requested window.
func (WindowActionExecutor) FocusWindow(ctx context.Context, windowID uint32) error {
	if err := window.FocusWindow(ctx, windowID); err != nil {
		return fmt.Errorf("focus window: %w", err)
	}
	return nil
}

// Click sends a click into the target window.
func (WindowActionExecutor) Click(ctx context.Context, windowID uint32, x, y float64, button string, clicks int) error {
	if err := window.Click(ctx, windowID, x, y, button, clicks); err != nil {
		return fmt.Errorf("click window: %w", err)
	}
	return nil
}

// PressKey presses and releases a key using the platform input controller.
func (WindowActionExecutor) PressKey(ctx context.Context, windowID uint32, key string, modifiers []string) error {
	controller := input.NewController()
	if err := controller.PressKey(ctx, key, modifiers); err != nil {
		return fmt.Errorf("press key on window %d: %w", windowID, err)
	}
	return nil
}

// Action is the normalized command returned by the vision model.
type Action struct {
	Action     string   `json:"action"`
	X          float64  `json:"x,omitempty"`
	Y          float64  `json:"y,omitempty"`
	Button     string   `json:"button,omitempty"`
	Clicks     int      `json:"clicks,omitempty"`
	Key        string   `json:"key,omitempty"`
	Modifiers  []string `json:"modifiers,omitempty"`
	Done       bool     `json:"done"`
	Why        string   `json:"why"`
	Confidence float64  `json:"confidence,omitempty"`
}

// StepArtifact captures one automation step for debugging.
type StepArtifact struct {
	Step           int           `json:"step"`
	Timestamp      time.Time     `json:"timestamp"`
	WindowID       uint32        `json:"window_id"`
	WindowBounds   window.Bounds `json:"window_bounds"`
	ImageWidth     int           `json:"image_width"`
	ImageHeight    int           `json:"image_height"`
	Scale          float64       `json:"scale"`
	LLMAction      *Action       `json:"llm_action"`
	Executed       bool          `json:"executed"`
	Error          string        `json:"error,omitempty"`
	ScreenshotPath string        `json:"screenshot_path,omitempty"`
}

// Agent drives the screenshot loop and vision model actions.
type Agent struct {
	config Config
}

// NewAgent creates a configured Agent with safe defaults.
func NewAgent(config Config) *Agent {
	if config.MaxSteps <= 0 {
		config.MaxSteps = 25
	}
	if config.MaxDuration <= 0 {
		config.MaxDuration = 2 * time.Minute
	}
	if config.ActionExecutor == nil {
		config.ActionExecutor = WindowActionExecutor{}
	}
	return &Agent{config: config}
}

// Run executes the full screenshot-driven automation loop.
func (a *Agent) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, a.config.MaxDuration)
	defer cancel()

	targetWindow, err := a.findTargetWindow(ctx)
	if err != nil {
		return err
	}

	if err := a.ensureArtifactDir(); err != nil {
		return err
	}

	return a.runSteps(ctx, targetWindow.WindowID)
}

func (a *Agent) runSteps(ctx context.Context, windowID uint32) error {
	for step := 1; step <= a.config.MaxSteps; step++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("agent run: %w", err)
		}

		done, err := a.runStep(ctx, windowID, step)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}

	return fmt.Errorf("max steps (%d) reached", a.config.MaxSteps)
}

func (a *Agent) runStep(ctx context.Context, windowID uint32, step int) (bool, error) {
	artifact := StepArtifact{
		Step:         step,
		Timestamp:    time.Now(),
		WindowID:     windowID,
		WindowBounds: window.Bounds{},
	}

	action, err := a.takeStep(ctx, windowID, artifact.Step, &artifact, a.config.Goal)
	if err != nil {
		return false, err
	}
	artifact.LLMAction = action

	if action.Done {
		a.saveArtifact(artifact)
		return true, nil
	}
	if action.Confidence > 0 && action.Confidence < 0.5 {
		return false, fmt.Errorf("confidence too low: %.2f", action.Confidence)
	}

	execErr := a.executeAction(ctx, windowID, action)
	if execErr != nil {
		artifact.Error = execErr.Error()
	} else {
		artifact.Executed = true
	}

	a.saveArtifact(artifact)

	if execErr != nil {
		return false, fmt.Errorf("execute action: %w", execErr)
	}

	time.Sleep(100 * time.Millisecond)
	return false, nil
}

func (a *Agent) findTargetWindow(ctx context.Context) (*window.Window, error) {
	windows, err := window.ListWindows(ctx)
	if err != nil {
		return nil, fmt.Errorf("list windows: %w", err)
	}

	for i := range windows {
		if a.config.WindowTitle != "" && windows[i].Title == a.config.WindowTitle {
			return &windows[i], nil
		}
		if a.config.OwnerName != "" && windows[i].OwnerName == a.config.OwnerName {
			return &windows[i], nil
		}
	}

	if len(windows) == 0 {
		return nil, fmt.Errorf("no windows found")
	}
	return &windows[0], nil
}

func (a *Agent) ensureArtifactDir() error {
	if a.config.SaveArtifacts && a.config.RunDir != "" {
		if err := os.MkdirAll(a.config.RunDir, 0o750); err != nil {
			return fmt.Errorf("create run dir: %w", err)
		}
	}
	return nil
}

func (a *Agent) takeStep(ctx context.Context, windowID uint32, step int, artifact *StepArtifact, goal string) (*Action, error) {
	if err := a.config.ActionExecutor.FocusWindow(ctx, windowID); err != nil {
		return nil, fmt.Errorf("focus window: %w", err)
	}

	data, metadata, err := window.TakeWindowScreenshot(ctx, windowID, imgencode.DefaultOptions)
	if err != nil {
		return nil, fmt.Errorf("take screenshot: %w", err)
	}

	artifact.WindowBounds = metadata.Bounds
	artifact.ImageWidth = metadata.ImageWidth
	artifact.ImageHeight = metadata.ImageHeight
	artifact.Scale = metadata.Scale
	artifact.Step = step

	if a.config.SaveArtifacts && a.config.RunDir != "" {
		screenshotPath := filepath.Join(a.config.RunDir, fmt.Sprintf("step_%04d.jpg", step))
		if err := os.WriteFile(screenshotPath, data, 0o600); err != nil {
			return nil, fmt.Errorf("save screenshot: %w", err)
		}
		artifact.ScreenshotPath = screenshotPath
	}

	action, err := a.config.VisionClient.GetAction(ctx, data, goal)
	if err != nil {
		return nil, fmt.Errorf("get action from vision model: %w", err)
	}
	return action, nil
}

func (a *Agent) executeAction(ctx context.Context, windowID uint32, action *Action) error {
	switch action.Action {
	case "click":
		button := action.Button
		if button == "" {
			button = "left"
		}
		clicks := action.Clicks
		if clicks == 0 {
			clicks = 1
		}
		if err := a.config.ActionExecutor.Click(ctx, windowID, action.X, action.Y, button, clicks); err != nil {
			return fmt.Errorf("click action: %w", err)
		}
		return nil
	case "press_key":
		if err := a.config.ActionExecutor.PressKey(ctx, windowID, action.Key, action.Modifiers); err != nil {
			return fmt.Errorf("press key action: %w", err)
		}
		return nil
	case "noop":
		return nil
	default:
		return fmt.Errorf("unknown action: %s", action.Action)
	}
}

func (a *Agent) saveArtifact(artifact StepArtifact) {
	if a.config.RunDir == "" {
		return
	}
	artifactPath := filepath.Join(a.config.RunDir, fmt.Sprintf("step_%04d.json", artifact.Step))
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}
	if err := os.WriteFile(artifactPath, data, 0o600); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "save artifact: %v\n", err)
	}
}

// MockVisionClient provides a test-friendly VisionClient implementation.
type MockVisionClient struct {
	GetActionFunc func(ctx context.Context, screenshot []byte, goal string) (*Action, error)
}

// GetAction returns the configured action if a mock function is set, or a noop action.
func (m *MockVisionClient) GetAction(ctx context.Context, screenshot []byte, goal string) (*Action, error) {
	if m.GetActionFunc != nil {
		return m.GetActionFunc(ctx, screenshot, goal)
	}
	return &Action{Action: "noop", Done: true, Why: "mock client default"}, nil
}

// EncodeImageBase64 returns the base64 representation of image bytes.
func EncodeImageBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
