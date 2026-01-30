package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainUsageExit(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestMainHelper")
	cmd.Env = append(os.Environ(),
		"AMUX_MAIN_HELPER=1",
		"AMUX_MAIN_ARGS=",
	)
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected exit error")
	}
}

func TestMainUnknownExit(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestMainHelper")
	cmd.Env = append(os.Environ(),
		"AMUX_MAIN_HELPER=1",
		"AMUX_MAIN_ARGS=unknown",
	)
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected exit error")
	}
}

func TestMainGitExit(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestMainHelper")
	cmd.Env = append(os.Environ(),
		"AMUX_MAIN_HELPER=1",
		"AMUX_MAIN_ARGS=git unknown",
	)
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected exit error")
	}
}

func TestMainAgentExit(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestMainHelper")
	cmd.Env = append(os.Environ(),
		"AMUX_MAIN_HELPER=1",
		"AMUX_MAIN_ARGS=agent",
	)
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected exit error")
	}
}

func TestMainTestCommand(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := writeStubGoMain(binDir, "95.0"); err != nil {
		t.Fatalf("write stub go: %v", err)
	}
	if err := writeStubLintMain(binDir); err != nil {
		t.Fatalf("write stub lint: %v", err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestMainHelper")
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(),
		"AMUX_MAIN_HELPER=1",
		"AMUX_MAIN_ARGS=test --no-snapshot",
		"PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("main test command failed: %v", err)
	}
}

func writeStubGoMain(binDir string, coverage string) error {
	script := "#!/bin/sh\nif [ \"$1\" = \"tool\" ] && [ \"$2\" = \"cover\" ]; then\n  echo \"total: (statements) " + coverage + "%\"\n  exit 0\nfi\nfor arg in \"$@\"; do\n  case \"$arg\" in\n    -coverprofile=*)\n      path=\"${arg#-coverprofile=}\"\n      mkdir -p \"$(dirname \"$path\")\"\n      echo \"mode: set\" > \"$path\"\n      ;;\n  esac\ndone\nexit 0\n"
	path := filepath.Join(binDir, "go")
	return os.WriteFile(path, []byte(script), 0o755)
}

func writeStubLintMain(binDir string) error {
	script := "#!/bin/sh\nexit 0\n"
	path := filepath.Join(binDir, "golangci-lint")
	return os.WriteFile(path, []byte(script), 0o755)
}

func TestMainHelper(t *testing.T) {
	if os.Getenv("AMUX_MAIN_HELPER") != "1" {
		return
	}
	args := strings.TrimSpace(os.Getenv("AMUX_MAIN_ARGS"))
	if args == "" {
		os.Args = []string{os.Args[0]}
	} else {
		os.Args = append([]string{os.Args[0]}, strings.Split(args, " ")...)
	}
	main()
	os.Exit(0)
}
