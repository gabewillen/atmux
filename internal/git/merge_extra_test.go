package git

import (
	"context"
	"fmt"
	"testing"
)

func TestMergeInvalidStrategyExtra(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, repoRoot string, args ...string) (ExecResult, error) {
		return ExecResult{}, nil
	}}
	_, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/tmp",
		AgentSlug:    "alpha",
		Strategy:     MergeStrategy("bad"),
		TargetBranch: "main",
	})
	if err == nil {
		t.Fatalf("expected invalid strategy error")
	}
}

func TestMergeBranchMissingExtra(t *testing.T) {
	runner := &Runner{Exec: func(ctx context.Context, repoRoot string, args ...string) (ExecResult, error) {
		if len(args) > 0 && args[0] == "show-ref" {
			return ExecResult{ExitCode: 128, Output: []byte("unknown revision")}, fmt.Errorf("missing")
		}
		return ExecResult{}, nil
	}}
	_, err := runner.Merge(context.Background(), MergeOptions{
		RepoRoot:     "/tmp",
		AgentSlug:    "alpha",
		Strategy:     StrategySquash,
		TargetBranch: "main",
	})
	if err == nil {
		t.Fatalf("expected branch missing error")
	}
}
