package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveHistoryPathUsesProjectRoot(t *testing.T) {
	repoDir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(repoDir, "cmd", "weavelens"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	if err := os.Chdir(filepath.Join(repoDir, "cmd", "weavelens")); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	want := filepath.Join(repoDir, historyFile)
	if got := resolveHistoryPath(); got != want {
		t.Fatalf("resolveHistoryPath() = %q, want %q", got, want)
	}
}

func TestResolveHistoryPathRespectsEnvOverride(t *testing.T) {
	want := filepath.Join(t.TempDir(), ".custom-scans.json")
	t.Setenv("WEAVELENS_HISTORY_FILE", want)

	if got := resolveHistoryPath(); got != want {
		t.Fatalf("resolveHistoryPath() = %q, want %q", got, want)
	}
}
