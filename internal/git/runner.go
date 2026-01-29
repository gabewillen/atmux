package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var (
	// ErrRepoRequired is returned when a repo root is required.
	ErrRepoRequired = errors.New("repo root required")
)

// ExecResult captures command output and exit status.
type ExecResult struct {
	Output   []byte
	ExitCode int
}

// ExecFunc executes a git command in the provided repo root.
type ExecFunc func(ctx context.Context, repoRoot string, args ...string) (ExecResult, error)

// Runner executes git commands.
type Runner struct {
	Exec ExecFunc
}

// NewRunner constructs a Runner using the default executor.
func NewRunner() *Runner {
	return &Runner{Exec: defaultExec}
}

func (r *Runner) run(ctx context.Context, repoRoot string, args ...string) (ExecResult, error) {
	if r == nil || r.Exec == nil {
		return ExecResult{}, fmt.Errorf("git runner: %w", ErrRepoRequired)
	}
	return r.Exec(ctx, repoRoot, args...)
}

func defaultExec(ctx context.Context, repoRoot string, args ...string) (ExecResult, error) {
	if strings.TrimSpace(repoRoot) == "" {
		return ExecResult{}, fmt.Errorf("git exec: %w", ErrRepoRequired)
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoRoot
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return ExecResult{Output: buf.Bytes(), ExitCode: exitCode}, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
		}
		return ExecResult{Output: buf.Bytes(), ExitCode: exitCode}, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return ExecResult{Output: buf.Bytes(), ExitCode: exitCode}, nil
}
