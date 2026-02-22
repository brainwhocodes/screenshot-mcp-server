package agent

import (
	"context"
	"encoding/json"
	"testing"
)

func TestAgent_NewAgentDefaults(t *testing.T) {
	cfg := Config{
		Goal: "test goal",
		VisionClient: &MockVisionClient{
			GetActionFunc: func(_ context.Context, _ []byte, _ string) (*Action, error) {
				return &Action{Action: "noop", Done: true}, nil
			},
		},
	}
	ag := NewAgent(cfg)
	if ag.config.MaxSteps != 25 {
		t.Errorf("expected MaxSteps 25, got %d", ag.config.MaxSteps)
	}
}

func TestAgent_MockVisionClient(t *testing.T) {
	client := &MockVisionClient{}
	action, err := client.GetAction(context.Background(), []byte("test"), "goal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action.Action != "noop" {
		t.Errorf("expected noop, got %s", action.Action)
	}
	if !action.Done {
		t.Error("expected done=true")
	}
}

func TestAction_JSONRoundTrip(t *testing.T) {
	original := Action{
		Action:     "click",
		X:          100.5,
		Y:          200.5,
		Button:     "left",
		Clicks:     1,
		Done:       false,
		Why:        "clicking the button",
		Confidence: 0.95,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed Action
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if parsed.Action != original.Action {
		t.Errorf("action: got %q want %q", parsed.Action, original.Action)
	}
	if parsed.X != original.X {
		t.Errorf("x: got %f want %f", parsed.X, original.X)
	}
	if parsed.Y != original.Y {
		t.Errorf("y: got %f want %f", parsed.Y, original.Y)
	}
	if parsed.Button != original.Button {
		t.Errorf("button: got %q want %q", parsed.Button, original.Button)
	}
	if parsed.Clicks != original.Clicks {
		t.Errorf("clicks: got %d want %d", parsed.Clicks, original.Clicks)
	}
	if parsed.Done != original.Done {
		t.Errorf("done: got %v want %v", parsed.Done, original.Done)
	}
	if parsed.Why != original.Why {
		t.Errorf("why: got %q want %q", parsed.Why, original.Why)
	}
	if parsed.Confidence != original.Confidence {
		t.Errorf("confidence: got %f want %f", parsed.Confidence, original.Confidence)
	}
}

func TestEncodeImageBase64(t *testing.T) {
	input := []byte("hello world")
	encoded := EncodeImageBase64(input)
	if encoded == "" {
		t.Fatal("expected non-empty base64")
	}
}
