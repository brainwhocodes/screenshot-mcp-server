package main

import (
	"io"
	"os"
	"testing"
)

func TestAgentCLI_Build(t *testing.T) {
	if os.Getenv("TEST_AGENT_CLI") == "" {
		t.Skip("Skipping agent CLI test (set TEST_AGENT_CLI=1 to run)")
	}
}

func TestAgentCLI_Help(t *testing.T) {
	args := []string{"-help"}
	code := run(args, io.Discard)
	if code != 0 {
		t.Errorf("expected exit code 0 for -help, got %d", code)
	}
}

func TestAgentCLI_MissingGoal(t *testing.T) {
	args := []string{}
	code := run(args, io.Discard)
	if code != 2 {
		t.Errorf("expected exit code 2 for missing goal, got %d", code)
	}
}
