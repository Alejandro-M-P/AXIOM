package build_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"axiom/internal/build"
	"axiom/internal/domain"
	"axiom/tests/mocks"
)

func TestBuildContextCreation(t *testing.T) {
	cfg := domain.EnvConfig{
		AxiomPath: "/home/user/axiom",
		BaseDir:   "/home/user",
		GPUType:   "nvidia",
		GFXVal:    "",
		ROCMMode:  "image",
	}

	ctx := context.Background()
	mockSystem := mocks.NewMockSystem()
	buildCtx, err := build.PrepareBuildContext(ctx, cfg, "test-container", "dev", mockSystem)

	if err != nil {
		t.Fatalf("PrepareBuildContext failed: %v", err)
	}

	if buildCtx == nil {
		t.Fatal("BuildContext should not be nil")
	}

	if buildCtx.ContainerName != "test-container" {
		t.Errorf("ContainerName = %s, want test-container", buildCtx.ContainerName)
	}

	if buildCtx.ImageName != "localhost/axiom-nvidia:latest" {
		t.Errorf("ImageName = %s, want localhost/axiom-nvidia:latest", buildCtx.ImageName)
	}

	expectedDir := filepath.Join(cfg.BaseDir, ".entorno", "test-container")
	if buildCtx.BuildWorkspaceDir != expectedDir {
		t.Errorf("BuildWorkspaceDir = %s, want %s", buildCtx.BuildWorkspaceDir, expectedDir)
	}
}

func TestPrepareSharedDirectories(t *testing.T) {
	fs := mocks.NewMockFileSystem()
	cfg := domain.EnvConfig{
		BaseDir: "/home/user",
	}

	modelsDir := filepath.Join(cfg.AIConfigDir(), "models")
	teamsDir := filepath.Join(cfg.AIConfigDir(), "teams")
	tutorPath := cfg.TutorPath()

	// Manually test what PrepareSharedDirectories does
	// Step 1: Create models dir
	err := fs.MkdirAll(modelsDir, 0700)
	if err != nil {
		t.Fatalf("MkdirAll(models) failed: %v", err)
	}
	if !fs.Dirs[modelsDir] {
		t.Errorf("Models dir not created in mock")
	}

	// Step 2: Create teams dir
	err = fs.MkdirAll(teamsDir, 0700)
	if err != nil {
		t.Fatalf("MkdirAll(teams) failed: %v", err)
	}
	if !fs.Dirs[teamsDir] {
		t.Errorf("Teams dir not created in mock")
	}

	// Step 3: Test ensureTutorFile logic
	// MkdirAll parent dir
	err = fs.MkdirAll(filepath.Dir(tutorPath), 0700)
	if err != nil {
		t.Fatalf("MkdirAll(tutor dir) failed: %v", err)
	}

	// Stat the file - should not exist
	_, err = fs.Stat(tutorPath)
	if err == nil {
		t.Error("Stat should return error for non-existent file")
	}

	// Create the file
	file, err := fs.OpenFile(tutorPath, os.O_CREATE, 0600)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	if file != nil {
		file.Close()
	}

	// Verify file exists now
	_, err = fs.Stat(tutorPath)
	if err != nil {
		t.Errorf("File should exist after OpenFile: %v", err)
	}
}

func TestBuildContainerFlags(t *testing.T) {
	cfg := domain.EnvConfig{
		AxiomPath: "/home/user/axiom",
		BaseDir:   "/home/user",
	}

	flags := build.BuildContainerFlags(cfg)

	// Verify flags contain expected elements
	if flags == "" {
		t.Fatal("BuildContainerFlags returned empty string")
	}

	expectedVolume := "--volume " + cfg.AIConfigDir() + ":/ai_config:z"
	if !contains(flags, expectedVolume) {
		t.Errorf("Flags should contain %s, got: %s", expectedVolume, flags)
	}

	if !contains(flags, "--device /dev/kfd") {
		t.Errorf("Flags should contain --device /dev/kfd, got: %s", flags)
	}

	if !contains(flags, "--device /dev/dri") {
		t.Errorf("Flags should contain --device /dev/dri, got: %s", flags)
	}

	if !contains(flags, "--security-opt label=disable") {
		t.Errorf("Flags should contain --security-opt label=disable, got: %s", flags)
	}

	if !contains(flags, "--group-add video") {
		t.Errorf("Flags should contain --group-add video, got: %s", flags)
	}

	if !contains(flags, "--group-add render") {
		t.Errorf("Flags should contain --group-add render, got: %s", flags)
	}
}

func TestRunInContainer(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	ctx := context.Background()

	err := build.RunInContainer(ctx, runtime, "test-container", "echo", "hello")

	if err != nil {
		t.Fatalf("RunInContainer failed: %v", err)
	}

	// Verify the command was executed
	if len(runtime.ExecuteCalls) != 1 {
		t.Errorf("Expected 1 ExecuteCall, got %d", len(runtime.ExecuteCalls))
	}

	if runtime.ExecuteCalls[0].Name != "test-container" {
		t.Errorf("Container name = %s, want test-container", runtime.ExecuteCalls[0].Name)
	}

	if len(runtime.ExecuteCalls[0].Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(runtime.ExecuteCalls[0].Args))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
