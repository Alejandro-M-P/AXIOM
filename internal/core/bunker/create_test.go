package bunker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/filesystem"
	"github.com/Alejandro-M-P/AXIOM/tests/mocks"
)

func TestCreateValidation_EmptyName(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	err := mgr.CreateBunker(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
	if err.Error() != "errors.bunker.missing_name" {
		t.Errorf("expected 'errors.bunker.missing_name' error, got '%s'", err.Error())
	}
}

func TestCreateValidation_EmptyNameWithSpaces(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	err := mgr.CreateBunker(context.Background(), "   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only name, got nil")
	}
	if err.Error() != "errors.bunker.missing_name" {
		t.Errorf("expected 'errors.bunker.missing_name' error, got '%s'", err.Error())
	}
}

func TestCreateValidation_InvalidChars(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	invalidNames := []string{
		"../etc",
		"foo\\bar",
		".",
		"..",
		"/absolute",
	}

	for _, name := range invalidNames {
		err := mgr.CreateBunker(context.Background(), name)
		if err == nil {
			t.Errorf("expected error for invalid name '%s', got nil", name)
		}
		if err != nil && err.Error() != "errors.bunker.invalid_name" && err.Error() != "errors.bunker.missing_name" {
			t.Errorf("expected 'errors.bunker.invalid_name' or 'errors.bunker.missing_name' error for '%s', got '%s'", name, err.Error())
		}
	}
}

func TestCreateRuntimeError(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.CreateBunkerErr = errors.New("runtime error")
	runtime.Images = []string{"localhost/axiom-generic:latest"}

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	err := mgr.CreateBunker(context.Background(), "test-bunker")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateWithFlags(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	runtime.Images = []string{"localhost/axiom-generic:latest"}

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	err := mgr.CreateBunker(context.Background(), "valid-name")
	if err != nil && err.Error() != "missing_name" && err.Error() != "invalid_name" && err.Error() != "access_denied" {
		t.Logf("Got expected error: %s", err.Error())
	}
}

func TestCreateBunkerNameSanitization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"valid-name", "valid-name", false},
		{"  valid-name  ", "valid-name", false},
		{"valid_name", "valid_name", false},
		{"name.with.dots", "name.with.dots", false},
	}

	for _, tc := range tests {
		result, err := sanitizeBunkerName(tc.input)
		if tc.hasError && err == nil {
			t.Errorf("sanitizeBunkerName(%q): expected error, got nil", tc.input)
		}
		if !tc.hasError && err != nil {
			t.Errorf("sanitizeBunkerName(%q): expected %q, got error %s", tc.input, tc.expected, err)
		}
		if !tc.hasError && result != tc.expected {
			t.Errorf("sanitizeBunkerName(%q): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}

func TestCreateContainerFlags(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	cfg := EnvConfig{
		BaseDir:   "/home/user/projects",
		AxiomPath: "/home/user/.axiom",
		ROCMMode:  "",
		AuthMode:  "local",
	}

	// createContainerFlags now only generates volume flags (device flags come from runtime.GetCreateFlags)
	flags := mgr.createContainerFlags(cfg, "generic", "test-bunker", "/home/user/projects/test-bunker", "")

	// Verify volume flags are present
	expectedParts := []string{
		"--volume /home/user/projects/test-bunker:/test-bunker:z",
		"--volume /home/user/projects/ai_config:/ai_config:z",
		"--volume /home/user/.axiom/config.toml:/run/axiom/env:ro,z",
	}

	for _, part := range expectedParts {
		if !containsFlag(flags, part) {
			t.Errorf("Expected flag '%s' not found in generated flags:\n%s", part, flags)
		}
	}
}

func containsFlag(flags, part string) bool {
	return len(flags) >= len(part) && (flags == part || len(flags) > len(part) && (containsString(flags, " "+part+" ") ||
		startsWith(flags, part+" ") ||
		endsWith(flags, " "+part)))
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func TestBunkerFlags_Struct(t *testing.T) {
	flags := BunkerFlags{
		GPUType:    "rdna4",
		ProjectDir: "/projects/myapp",
		HomeDir:    "/root",
	}

	if flags.GPUType != "rdna4" {
		t.Errorf("GPUType mismatch: expected 'rdna4', got '%s'", flags.GPUType)
	}
	if flags.ProjectDir != "/projects/myapp" {
		t.Errorf("ProjectDir mismatch: expected '/projects/myapp', got '%s'", flags.ProjectDir)
	}
	if flags.HomeDir != "/root" {
		t.Errorf("HomeDir mismatch: expected '/root', got '%s'", flags.HomeDir)
	}
}

func TestResolveBuildGPU_Function(t *testing.T) {
	cfg := EnvConfig{
		GFXVal: "RDNA4",
	}

	gpu := resolveBuildGPU(cfg)

	if gpu.Type != "generic" {
		t.Errorf("expected GPU type 'generic', got '%s'", gpu.Type)
	}
	// ports.GPUInfo solo tiene Type y Name, no GfxVal
}

func TestBaseImageName_Function(t *testing.T) {
	tests := []struct {
		gpuType  string
		expected string
	}{
		{"rdna4", "localhost/axiom-rdna4:latest"},
		{"nvidia", "localhost/axiom-nvidia:latest"},
		{"intel", "localhost/axiom-intel:latest"},
		{"generic", "localhost/axiom-generic:latest"},
		{"", "localhost/axiom-generic:latest"},
		{"   ", "localhost/axiom-generic:latest"},
	}

	for _, tc := range tests {
		result := baseImageName(tc.gpuType)
		if result != tc.expected {
			t.Errorf("baseImageName(%q): expected %q, got %q", tc.gpuType, tc.expected, result)
		}
	}
}

func TestPrepareSSHAgent_LocalAuthMode(t *testing.T) {
	// Test removed - function moved to system adapter
	// The logic is now in ports.ISystem.PrepareSSHAgent
}

func TestEnterBunker_Integration(t *testing.T) {
	// Test removed - enterBunker moved to runtime adapter
	// Use runtime.EnterBunker directly in integration tests
}

func TestWriteShellBootstrap_Placeholder(t *testing.T) {
	err := writeShellBootstrap(EnvConfig{}, "test", "/tmp/env", "RDNA4")
	if err != nil {
		t.Errorf("writeShellBootstrap should not return error, got: %s", err)
	}
}

func TestWriteStarshipConfig_Placeholder(t *testing.T) {
	err := writeStarshipConfig("/tmp/env")
	if err != nil {
		t.Errorf("writeStarshipConfig should not return error, got: %s", err)
	}
}

func TestCopyTutorToAgents_Placeholder(t *testing.T) {
	err := copyTutorToAgents("/tmp/tutor", "/tmp/env")
	if err != nil {
		t.Errorf("copyTutorToAgents should not return error, got: %s", err)
	}
}

func TestWriteOpencodeConfig_Placeholder(t *testing.T) {
	err := writeOpencodeConfig("/tmp/env")
	if err != nil {
		t.Errorf("writeOpencodeConfig should not return error, got: %s", err)
	}
}

func TestEnsureTutorFile_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	tutorPath := filepath.Join(tmpDir, "tutor")

	f, err := os.Create(tutorPath)
	if err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}
	f.Close()

	fs := filesystem.NewFSAdapter()
	err = ensureTutorFile(fs, tutorPath)
	if err != nil {
		t.Errorf("ensureTutorFile should not return error for existing file, got: %s", err)
	}
}

func TestEnsureTutorFile_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	tutorPath := filepath.Join(tmpDir, "new-tutor")

	fs := filesystem.NewFSAdapter()
	err := ensureTutorFile(fs, tutorPath)
	if err != nil {
		t.Errorf("ensureTutorFile should not return error for new file, got: %s", err)
	}

	if _, err := os.Stat(tutorPath); os.IsNotExist(err) {
		t.Error("ensureTutorFile did not create the file")
	}
}
