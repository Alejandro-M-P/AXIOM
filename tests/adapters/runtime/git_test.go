package runtime_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/filesystem"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/runtime"
)

func TestGitAdapter_GetBranch_ValidBranch(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %s", err)
	}

	headPath := filepath.Join(gitDir, "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/main\n"), 0644); err != nil {
		t.Fatalf("failed to write HEAD: %s", err)
	}

	fs := filesystem.NewFSAdapter()
	git := runtime.NewGitAdapter(fs)

	branch := git.GetBranch(tmpDir)
	if branch != "main" {
		t.Errorf("expected 'main', got '%s'", branch)
	}
}

func TestGitAdapter_GetBranch_DetachedHEAD(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %s", err)
	}

	headPath := filepath.Join(gitDir, "HEAD")
	if err := os.WriteFile(headPath, []byte("abc1234def5678\n"), 0644); err != nil {
		t.Fatalf("failed to write HEAD: %s", err)
	}

	fs := filesystem.NewFSAdapter()
	git := runtime.NewGitAdapter(fs)

	branch := git.GetBranch(tmpDir)
	if branch != "abc1234" {
		t.Errorf("expected 'abc1234', got '%s'", branch)
	}
}

func TestGitAdapter_GetBranch_NoRepo(t *testing.T) {
	tmpDir := t.TempDir()

	fs := filesystem.NewFSAdapter()
	git := runtime.NewGitAdapter(fs)

	branch := git.GetBranch(tmpDir)
	if branch != "-" {
		t.Errorf("expected '-' for no repo, got '%s'", branch)
	}
}

func TestGitAdapter_GetBranch_ShortHash(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %s", err)
	}

	headPath := filepath.Join(gitDir, "HEAD")
	if err := os.WriteFile(headPath, []byte("a1b2c3d\n"), 0644); err != nil {
		t.Fatalf("failed to write HEAD: %s", err)
	}

	fs := filesystem.NewFSAdapter()
	git := runtime.NewGitAdapter(fs)

	branch := git.GetBranch(tmpDir)
	if branch != "a1b2c3d" {
		t.Errorf("expected 'a1b2c3d', got '%s'", branch)
	}
}

func TestGitAdapter_GetBranch_TooShortHash(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %s", err)
	}

	headPath := filepath.Join(gitDir, "HEAD")
	if err := os.WriteFile(headPath, []byte("abc\n"), 0644); err != nil {
		t.Fatalf("failed to write HEAD: %s", err)
	}

	fs := filesystem.NewFSAdapter()
	git := runtime.NewGitAdapter(fs)

	branch := git.GetBranch(tmpDir)
	if branch != "-" {
		t.Errorf("expected '-' for too-short hash, got '%s'", branch)
	}
}

func TestGitAdapter_ImplementsIGit(t *testing.T) {
	fs := filesystem.NewFSAdapter()
	git := runtime.NewGitAdapter(fs)

	var _ interface{ GetBranch(string) string } = git
}

func TestGitAdapter_ConfigureUser_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a git repo so git config works
	if err := exec.Command("git", "-C", tmpDir, "init").Run(); err != nil {
		t.Fatalf("failed to init git repo: %s", err)
	}

	fs := filesystem.NewFSAdapter()
	git := runtime.NewGitAdapter(fs)

	err := git.ConfigureUser(context.Background(), tmpDir, "Test User", "test@example.com")
	if err != nil {
		t.Fatalf("ConfigureUser failed: %s", err)
	}

	// Verify the config was set
	nameOut, err := exec.Command("git", "-C", tmpDir, "config", "user.name").Output()
	if err != nil {
		t.Fatalf("failed to read user.name: %s", err)
	}
	if strings.TrimSpace(string(nameOut)) != "Test User" {
		t.Errorf("expected user.name 'Test User', got '%s'", strings.TrimSpace(string(nameOut)))
	}

	emailOut, err := exec.Command("git", "-C", tmpDir, "config", "user.email").Output()
	if err != nil {
		t.Fatalf("failed to read user.email: %s", err)
	}
	if strings.TrimSpace(string(emailOut)) != "test@example.com" {
		t.Errorf("expected user.email 'test@example.com', got '%s'", strings.TrimSpace(string(emailOut)))
	}
}

func TestGitAdapter_ConfigureUser_InvalidPath(t *testing.T) {
	fs := filesystem.NewFSAdapter()
	git := runtime.NewGitAdapter(fs)

	err := git.ConfigureUser(context.Background(), "/nonexistent/path", "User", "email@test.com")
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
}
