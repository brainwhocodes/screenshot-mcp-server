// Package main implements the agent CLI entrypoint.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brainwhocodes/screenshot_mcp_server/internal/agent"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stderr))
}

func run(args []string, stderr io.Writer) int {
	parsed, err := parseAgentArgs(args)
	if err != nil {
		return handleRunParseError(stderr, err)
	}

	ag := agent.NewAgent(agent.Config{
		Goal:          parsed.goal,
		WindowTitle:   parsed.windowTitle,
		OwnerName:     parsed.ownerName,
		MaxSteps:      parsed.maxSteps,
		MaxDuration:   parsed.maxDuration,
		RunDir:        parsed.runDir,
		SaveArtifacts: parsed.saveArtifacts,
		VisionClient:  parsed.visionClient,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	return runAgent(ctx, ag, stderr, parsed)
}

func handleRunParseError(stderr io.Writer, err error) int {
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	_ = writeStderrLine(stderr, fmt.Sprintf("Error: %v", err))
	return 2
}

func runAgent(ctx context.Context, ag *agent.Agent, stderr io.Writer, parsed parsedAgentArgs) int {
	if err := printAgentStartup(stderr, parsed.goal, parsed.runDir); err != nil {
		_ = writeStderrLine(stderr, fmt.Sprintf("Error: %v", err))
		return 2
	}

	if err := ag.Run(ctx); err != nil {
		return formatRunError(stderr, err)
	}

	if err := writeStderrLine(stderr, "Agent completed successfully"); err != nil {
		return 2
	}
	return 0
}

type parsedAgentArgs struct {
	goal          string
	windowTitle   string
	ownerName     string
	maxSteps      int
	maxDuration   time.Duration
	runDir        string
	saveArtifacts bool
	visionClient  agent.VisionClient
}

func parseAgentArgs(args []string) (parsedAgentArgs, error) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	goal := fs.String("goal", "", "Goal for the automation agent (required)")
	windowTitle := fs.String("window", "", "Target window title")
	ownerName := fs.String("app", "", "Target application name")
	maxSteps := fs.Int("max-steps", 25, "Maximum number of steps")
	maxDuration := fs.Duration("timeout", 2*time.Minute, "Maximum duration")
	runDir := fs.String("run-dir", "", "Directory to save run artifacts")
	apiKey := fs.String("api-key", "", "OpenAI API key (or set OPENAI_API_KEY env)")
	model := fs.String("model", "gpt-4o", "OpenAI model to use")
	dryRun := fs.Bool("dry-run", false, "Run without executing actions (for testing)")

	if err := fs.Parse(args); err != nil {
		return parsedAgentArgs{}, fmt.Errorf("parse flags: %w", err)
	}
	if *goal == "" {
		return parsedAgentArgs{}, fmt.Errorf("-goal is required")
	}

	key := *apiKey
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	if key == "" && !*dryRun {
		return parsedAgentArgs{}, fmt.Errorf("OpenAI API key required (set -api-key or OPENAI_API_KEY)")
	}

	visionClient := buildVisionClient(*dryRun, key, *model)

	saveArtifacts := *runDir != ""
	if !saveArtifacts {
		tmpDir, err := os.MkdirTemp("", "agent-run-*")
		if err != nil {
			return parsedAgentArgs{}, fmt.Errorf("create temp dir: %w", err)
		}
		*runDir = tmpDir
		saveArtifacts = true
	}

	return parsedAgentArgs{
		goal:          *goal,
		windowTitle:   *windowTitle,
		ownerName:     *ownerName,
		maxSteps:      *maxSteps,
		maxDuration:   *maxDuration,
		runDir:        *runDir,
		saveArtifacts: saveArtifacts,
		visionClient:  visionClient,
	}, nil
}

func buildVisionClient(dryRun bool, key, model string) agent.VisionClient {
	if dryRun {
		return &agent.MockVisionClient{
			GetActionFunc: func(_ context.Context, _ []byte, _ string) (*agent.Action, error) {
				return &agent.Action{
					Action: "noop",
					Done:   true,
					Why:    "dry run mode - no actions executed",
				}, nil
			},
		}
	}
	return agent.NewOpenAIVisionClient(key, model)
}

func printAgentStartup(stderr io.Writer, goal, runDir string) error {
	if err := writeStderrf(stderr, "Starting agent with goal: %s\n", goal); err != nil {
		return err
	}
	return writeStderrf(stderr, "Run directory: %s\n", runDir)
}

func formatRunError(stderr io.Writer, err error) int {
	switch err {
	case context.DeadlineExceeded:
		_ = writeStderrLine(stderr, "Agent timed out")
		return 1
	case context.Canceled:
		_ = writeStderrLine(stderr, "Agent cancelled")
		return 130
	default:
		_ = writeStderrf(stderr, "Agent error: %v\n", err)
		return 1
	}
}

func writeStderrf(stderr io.Writer, format string, args ...any) error {
	if _, err := fmt.Fprintf(stderr, format, args...); err != nil {
		return fmt.Errorf("write stderr: %w", err)
	}
	return nil
}

func writeStderrLine(stderr io.Writer, text string) error {
	return writeStderrf(stderr, "%s\n", text)
}
