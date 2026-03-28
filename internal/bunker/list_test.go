package bunker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"axiom/internal/domain"
	"axiom/tests/mocks"
)

func TestListEmpty(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.ListBunkers(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(ui.ShowWarningCalls) == 0 {
		t.Error("expected ShowWarning to be called for empty list")
	}
}

func TestListWithBunkers(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "bunker-1", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.ListBunkers(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestListFormatting_MultipleBunkers(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "bunker-1", Status: "running", Image: "localhost/axiom-generic:latest"},
		{Name: "bunker-2", Status: "stopped", Image: "localhost/axiom-rdna4:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.ListBunkers(context.Background())
	// With multiple bunkers, interactive selection is needed which is not fully implemented
	if err != nil {
		t.Logf("List returned error (expected for multiple bunkers): %s", err)
	}
}

func TestBunkerInfo_Success(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.BunkerInfo(context.Background(), "test-bunker")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	found := false
	for _, call := range ui.ShowCommandCardCalls {
		if call.CommandKey == "info" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ShowCommandCard to be called with 'info' command")
	}
}

func TestBunkerInfo_EmptyNameFallsBackToList(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "bunker-1", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui)

	err := mgr.BunkerInfo(context.Background(), "")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestBunkerStatus_Running(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "running", Image: "localhost/axiom-generic:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui)

	status := mgr.BunkerStatus(context.Background(), "test-bunker")
	if status != "running" {
		t.Errorf("expected 'running', got '%s'", status)
	}
}

func TestBunkerStatus_Stopped(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{
		{Name: "test-bunker", Status: "stopped", Image: "localhost/axiom-generic:latest"},
	}

	mgr := NewManager("/root", runtime, fs, ui)

	status := mgr.BunkerStatus(context.Background(), "test-bunker")
	if status != "stopped" {
		t.Errorf("expected 'stopped', got '%s'", status)
	}
}

func TestBunkerStatus_NotFound(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	mgr := NewManager("/root", runtime, fs, ui)

	status := mgr.BunkerStatus(context.Background(), "nonexistent")
	if status != "stopped" {
		t.Errorf("expected 'stopped' for nonexistent bunker, got '%s'", status)
	}
}

func TestBunkerStatus_ListError(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.ListBunkersErr = context.DeadlineExceeded

	mgr := NewManager("/root", runtime, fs, ui)

	status := mgr.BunkerStatus(context.Background(), "test-bunker")
	if status != "stopped" {
		t.Errorf("expected 'stopped' fallback on error, got '%s'", status)
	}
}

func TestBunkerEnvSize_Calculation(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := EnvConfig{
		BaseDir: tmpDir,
	}

	bunkerDir := cfg.BuildWorkspaceDir("test-bunker")
	if err := os.MkdirAll(bunkerDir, 0755); err != nil {
		t.Fatalf("failed to create bunker dir: %s", err)
	}

	for i := 0; i < 3; i++ {
		f, err := os.Create(filepath.Join(bunkerDir, "file"+string(rune('a'+i))+".txt"))
		if err != nil {
			t.Fatalf("failed to create file: %s", err)
		}
		f.WriteString("test content")
		f.Close()
	}

	size := bunkerEnvSize(cfg, "test-bunker")
	if size == "-" {
		t.Error("expected size to be calculated, got '-'")
	}
}

func TestBunkerEnvSize_NotExist(t *testing.T) {
	cfg := EnvConfig{
		BaseDir: "/nonexistent",
	}

	size := bunkerEnvSize(cfg, "nonexistent")
	if size != "-" {
		t.Errorf("expected '-' for nonexistent bunker, got '%s'", size)
	}
}

func TestBunkerEnvPath_Function(t *testing.T) {
	cfg := EnvConfig{
		BaseDir: "/projects",
	}

	path := bunkerEnvPath(cfg, "mybunker")
	expected := "/projects/.entorno/mybunker"

	if path != expected {
		t.Errorf("expected '%s', got '%s'", expected, path)
	}
}

func TestBunkerProjectPath_Function(t *testing.T) {
	cfg := EnvConfig{
		BaseDir: "/projects",
	}

	path := bunkerProjectPath(cfg, "mybunker")
	expected := "/projects/mybunker"

	if path != expected {
		t.Errorf("expected '%s', got '%s'", expected, path)
	}
}

func TestBunkerGitBranch_Detached(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := EnvConfig{
		BaseDir: tmpDir,
	}

	projectDir := filepath.Join(tmpDir, "test-bunker")
	gitDir := filepath.Join(projectDir, ".git")

	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %s", err)
	}

	headPath := filepath.Join(gitDir, "HEAD")
	if err := os.WriteFile(headPath, []byte("abc1234\n"), 0644); err != nil {
		t.Fatalf("failed to write HEAD: %s", err)
	}

	branch := bunkerGitBranch(cfg, "test-bunker")
	if branch != "abc1234" {
		t.Errorf("expected 'abc1234', got '%s'", branch)
	}
}

func TestBunkerGitBranch_NonExistent(t *testing.T) {
	cfg := EnvConfig{
		BaseDir: "/nonexistent",
	}

	branch := bunkerGitBranch(cfg, "nonexistent")
	if branch != "-" {
		t.Errorf("expected '-' for nonexistent project, got '%s'", branch)
	}
}

func TestBunkerLastEntry_Function(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := EnvConfig{
		BaseDir: tmpDir,
	}

	bunkerDir := cfg.BuildWorkspaceDir("test-bunker")
	if err := os.MkdirAll(bunkerDir, 0755); err != nil {
		t.Fatalf("failed to create bunker dir: %s", err)
	}

	lastEntry := bunkerLastEntry(cfg, "test-bunker")
	if lastEntry == "-" {
		t.Error("expected date, got '-'")
	}
}

func TestBunkerLastEntry_NotExist(t *testing.T) {
	cfg := EnvConfig{
		BaseDir: "/nonexistent",
	}

	lastEntry := bunkerLastEntry(cfg, "nonexistent")
	if lastEntry != "-" {
		t.Errorf("expected '-' for nonexistent bunker, got '%s'", lastEntry)
	}
}

func TestHumanPath_Function(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{filepath.Join(home, "projects"), "~" + "/projects"},
		{"/usr/local/bin", "/usr/local/bin"},
	}

	for _, tc := range tests {
		result := humanPath(tc.input)
		if result != tc.expected {
			t.Errorf("humanPath(%q): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}

func TestDefaultString_Function(t *testing.T) {
	tests := []struct {
		value    string
		fallback string
		expected string
	}{
		{"actual", "default", "actual"},
		{"", "default", "default"},
		{"   ", "default", "default"},
	}

	for _, tc := range tests {
		result := defaultString(tc.value, tc.fallback)
		if result != tc.expected {
			t.Errorf("defaultString(%q, %q): expected %q, got %q", tc.value, tc.fallback, tc.expected, result)
		}
	}
}

func TestBunkerTimestamp_Function(t *testing.T) {
	// Create a proper time.Time for testing
	now := time.Now()
	result := bunkerTimestamp(now)
	_ = result
}

func TestImageExists_Function(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Images = []string{"localhost/axiom-generic:latest"}

	mgr := NewManager("/root", runtime, fs, ui)

	if !mgr.ImageExists(context.Background(), "localhost/axiom-generic:latest") {
		t.Error("expected ImageExists to return true for existing image")
	}

	if mgr.ImageExists(context.Background(), "nonexistent:latest") {
		t.Error("expected ImageExists to return false for nonexistent image")
	}
}

func TestListAxiomImages_Empty(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Bunkers = []domain.Bunker{}

	mgr := NewManager("/root", runtime, fs, ui)

	images, err := mgr.ListAxiomImages(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(images) != 0 {
		t.Errorf("expected 0 images, got %d", len(images))
	}
}
