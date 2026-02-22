// Package safeexec centralizes validated command execution helpers.
package safeexec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	defaultCommandTimeout     = 10 * time.Second
	defaultAppleScriptTimeout = 5 * time.Second
)

var errNilContext = errors.New("context is required")

// RunCommand executes a command with a bounded context and returns stdout/stderr.
func RunCommand(ctx context.Context, command string, args ...string) ([]byte, error) {
	return RunCommandWithInput(ctx, nil, command, args...)
}

// RunCommandWithInput executes a command with stdin input and a bounded context.
func RunCommandWithInput(ctx context.Context, input []byte, command string, args ...string) ([]byte, error) {
	return RunCommandWithInputAndTimeout(ctx, input, defaultCommandTimeout, command, args...)
}

// RunCommandWithTimeout executes a command with an explicit context timeout and returns stdout/stderr.
func RunCommandWithTimeout(ctx context.Context, timeout time.Duration, command string, args ...string) ([]byte, error) {
	return RunCommandWithInputAndTimeout(ctx, nil, timeout, command, args...)
}

// RunCommandWithInputAndTimeout executes a command with stdin input and an explicit timeout.
func RunCommandWithInputAndTimeout(ctx context.Context, input []byte, timeout time.Duration, command string, args ...string) ([]byte, error) {
	if err := ValidateCommandArg(command); err != nil {
		return nil, fmt.Errorf("invalid command %q: %w", command, err)
	}
	for _, arg := range args {
		if err := ValidateCommandArg(arg); err != nil {
			return nil, fmt.Errorf("invalid arg %q: %w", arg, err)
		}
	}

	ctxWithTimeout, cancel, err := contextWithTimeout(ctx, timeout)
	if err != nil {
		return nil, err
	}
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, command, args...)
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, wrapCommandError(command, err)
	}
	return output, nil
}

// RunCommandNoOutput runs a command and discards stdout/stderr.
func RunCommandNoOutput(ctx context.Context, command string, args ...string) error {
	_, err := RunCommand(ctx, command, args...)
	return err
}

func validateCommand(command string, args ...string) error {
	if err := ValidateCommandArg(command); err != nil {
		return err
	}

	for _, arg := range args {
		if err := ValidateCommandArg(arg); err != nil {
			return err
		}
	}

	return nil
}

// CommandContext returns a validated exec.CommandContext with timeout handling.
func CommandContext(ctx context.Context, command string, args ...string) (*exec.Cmd, context.CancelFunc, error) {
	if err := validateCommand(command, args...); err != nil {
		return nil, func() {}, fmt.Errorf("invalid command invocation: %w", err)
	}

	ctxWithTimeout, cancel, err := contextWithTimeout(ctx, defaultCommandTimeout)
	if err != nil {
		return nil, func() {}, err
	}
	return exec.CommandContext(ctxWithTimeout, command, args...), cancel, nil
}

func wrapCommandError(command string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("command %q failed: %w", command, err)
}

// ValidateCommandArg provides a narrow validation for shell-like command injection vectors.
func ValidateCommandArg(arg string) error {
	if strings.TrimSpace(arg) == "" {
		return errors.New("empty argument")
	}
	if strings.ContainsRune(arg, '\x00') {
		return errors.New("contains null byte")
	}
	return nil
}

func contextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc, error) {
	if ctx == nil {
		return nil, func() {}, errNilContext
	}
	if timeout <= 0 {
		timeout = defaultCommandTimeout
	}

	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining <= timeout {
			return ctx, func() {}, nil
		}
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	return ctxWithTimeout, cancel, nil
}

// QuoteAppleScriptString escapes quotes and backslashes for safe AppleScript string interpolation.
func QuoteAppleScriptString(value string) string {
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(escaped, `"`, `\"`)
}

// RunAppleScript executes a short AppleScript command via osascript.
func RunAppleScript(ctx context.Context, script string) error {
	if script == "" {
		return fmt.Errorf("script is required")
	}
	output, err := RunCommandWithTimeout(ctx, defaultAppleScriptTimeout, "osascript", "-e", script)
	if err != nil {
		return fmt.Errorf("apple script failed: %w, output: %s", err, string(output))
	}
	return nil
}
