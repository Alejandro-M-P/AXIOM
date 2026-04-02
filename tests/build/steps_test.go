package build_test

import (
	"context"
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/build"
	"github.com/Alejandro-M-P/AXIOM/internal/config"
	"github.com/Alejandro-M-P/AXIOM/tests/mocks"
)

func TestInstallSystemBase(t *testing.T) {
	ui := mocks.NewMockPresenter()
	cfg := config.EnvConfig{
		AxiomPath: "/home/user/axiom",
		BaseDir:   "/home/user",
		ROCMMode:  "",
	}
	containerName := "test-container"
	buildCtx := &build.BuildContext{
		Config:        cfg,
		GPUInfo:       &config.GPUInfo{Type: "generic"},
		ContainerName: containerName,
	}

	execCalled := false
	exec := func(ctx context.Context, cmd string, args ...string) error {
		execCalled = true
		return nil
	}

	err := build.InstallSystemBase(context.Background(), containerName, buildCtx, ui, exec)

	if err != nil {
		t.Fatalf("InstallSystemBase failed: %v", err)
	}

	if !execCalled {
		t.Error("exec function was not called")
	}
}

func TestInstallDeveloperTools(t *testing.T) {
	ui := mocks.NewMockPresenter()
	cfg := config.EnvConfig{
		AxiomPath: "/home/user/axiom",
		BaseDir:   "/home/user",
	}
	containerName := "test-container"
	buildCtx := &build.BuildContext{
		Config:        cfg,
		GPUInfo:       &config.GPUInfo{Type: "generic"},
		ContainerName: containerName,
	}

	execCalled := false
	exec := func(ctx context.Context, cmd string, args ...string) error {
		execCalled = true
		return nil
	}

	err := build.InstallDeveloperTools(context.Background(), containerName, buildCtx, ui, exec)

	if err != nil {
		t.Fatalf("InstallDeveloperTools failed: %v", err)
	}

	if !execCalled {
		t.Error("exec function was not called")
	}
}

func TestInstallModelStack(t *testing.T) {
	ui := mocks.NewMockPresenter()
	cfg := config.EnvConfig{
		AxiomPath: "/home/user/axiom",
		BaseDir:   "/home/user",
	}
	containerName := "test-container"
	buildCtx := &build.BuildContext{
		Config:        cfg,
		GPUInfo:       &config.GPUInfo{Type: "nvidia"},
		ContainerName: containerName,
	}
	modelConfig := build.ModelStackConfig{GPUType: "nvidia"}

	execCalled := false
	exec := func(ctx context.Context, cmd string, args ...string) error {
		execCalled = true
		return nil
	}

	err := build.InstallModelStack(context.Background(), containerName, buildCtx, modelConfig, ui, exec)

	if err != nil {
		t.Fatalf("InstallModelStack failed: %v", err)
	}

	if !execCalled {
		t.Error("exec function was not called")
	}
}

func TestInstallOllama(t *testing.T) {
	// Test that ollama installation works with different GPU types
	testCases := []struct {
		name    string
		gpuType string
	}{
		{"NVIDIA GPU", "nvidia"},
		{"AMD GPU", "amd"},
		{"RDNA GPU", "rdna3"},
		{"Generic GPU", "generic"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			execCalled := false
			exec := func(ctx context.Context, cmd string, args ...string) error {
				execCalled = true
				return nil
			}

			// We can't directly test installOllama since it's not exported,
			// but we verify the exec function gets called through InstallModelStack
			cfg := config.EnvConfig{
				AxiomPath: "/home/user/axiom",
				BaseDir:   "/home/user",
			}
			containerName := "test-container"
			buildCtx := &build.BuildContext{
				Config:        cfg,
				GPUInfo:       &config.GPUInfo{Type: tc.gpuType},
				ContainerName: containerName,
			}
			modelConfig := build.ModelStackConfig{GPUType: tc.gpuType}
			ui := mocks.NewMockPresenter()

			err := build.InstallModelStack(context.Background(), containerName, buildCtx, modelConfig, ui, exec)

			if err != nil {
				t.Fatalf("InstallModelStack failed for GPU type %s: %v", tc.gpuType, err)
			}

			if !execCalled {
				t.Error("exec function was not called")
			}
		})
	}
}
