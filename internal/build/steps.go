package build

import (
	"context"
	"fmt"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// ModelStackConfig holds configuration for the AI/ML model stack installation.
type ModelStackConfig struct {
	GPUType string
}

// InstallSystemBase installs the base system packages in the container.
// NOTE: This function should be called within a Progress.RunStep() from the build manager.
// It logs progress instead of creating a nested Progress to avoid UI conflicts.
func InstallSystemBase(ctx context.Context, containerName string, cfg *BuildContext, ui ports.IPresenter, exec func(context.Context, string, ...string) error) error {
	packages := []string{"base-devel", "git", "curl", "jq", "wget", "nodejs", "npm", "go", "fzf", "starship"}
	if cfg.Config.ROCMMode == "image" {
		switch {
		case cfg.GPUInfo.Type == "nvidia":
			packages = append(packages, "nvidia-utils", "cuda")
		case strings.HasPrefix(cfg.GPUInfo.Type, "rdna"), cfg.GPUInfo.Type == "amd":
			packages = append(packages, "rocm-hip-sdk")
		case cfg.GPUInfo.Type == "intel":
			packages = append(packages, "intel-compute-runtime", "onevpl-intel-gpu")
		}
	}

	// Sync repos
	ui.ShowLog("info", ui.GetText("task.sync_repos"))
	if err := exec(ctx, "sudo", "pacman", "-Sy", "--noconfirm"); err != nil {
		return fmt.Errorf("errors.build.steps.failed_sync_repos: %w", err)
	}

	// Install packages
	ui.ShowLog("info", ui.GetText("task.install_pkgs"), ui.GetText("detail.pkgs_count", len(packages)))
	args := []string{"sudo", "pacman", "-S", "--needed", "--noconfirm"}
	args = append(args, packages...)
	if err := exec(ctx, args[0], args[1:]...); err != nil {
		return fmt.Errorf("errors.build.steps.failed_install_pkgs: %w", err)
	}

	// Install Homebrew
	ui.ShowLog("info", ui.GetText("task.install_homebrew"))
	cmd := "NONINTERACTIVE=1 /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
	if err := exec(ctx, "bash", "-c", cmd); err != nil {
		return fmt.Errorf("errors.build.steps.failed_install_homebrew: %w", err)
	}

	return nil
}

// InstallDeveloperTools installs development tools (opencode, engram, gentle-ai).
// NOTE: This function should be called within a Progress.RunStep() from the build manager.
// It logs progress instead of creating a nested Progress to avoid UI conflicts.
func InstallDeveloperTools(ctx context.Context, containerName string, cfg *BuildContext, ui ports.IPresenter, exec func(context.Context, string, ...string) error) error {
	// Install opencode via npm
	ui.ShowLog("info", ui.GetText("task.install_opencode"))
	if err := exec(ctx, "sudo", "npm", "install", "-g", "opencode-ai"); err != nil {
		return fmt.Errorf("errors.build.steps.failed_install_opencode: %w", err)
	}

	brewPath := "/home/linuxbrew/.linuxbrew/bin/brew"

	// Add Homebrew tap
	ui.ShowLog("info", ui.GetText("task.configure_brew_tap", "Gentleman-Programming/homebrew-tap"))
	if err := exec(ctx, brewPath, "tap", "Gentleman-Programming/homebrew-tap"); err != nil {
		return fmt.Errorf("errors.build.steps.failed_add_brew_tap: %w", err)
	}

	// Install engram and gentle-ai
	ui.ShowLog("info", ui.GetText("task.install_dev_tools", "engram gentle-ai"))
	if err := exec(ctx, brewPath, "install", "engram", "gentle-ai"); err != nil {
		return fmt.Errorf("errors.build.steps.failed_install_dev_tools: %w", err)
	}

	return nil
}

// InstallModelStack installs Ollama and configures the AI stack.
// NOTE: This function should be called within a Progress.RunStep() from the build manager.
// It logs progress instead of creating a nested Progress to avoid UI conflicts.
func InstallModelStack(ctx context.Context, containerName string, cfg *BuildContext, modelConfig ModelStackConfig, ui ports.IPresenter, exec func(context.Context, string, ...string) error) error {
	// Install Ollama
	ui.ShowLog("info", ui.GetText("task.install_ollama"), cfg.GPUInfo.Type)
	if err := installOllama(ctx, cfg.GPUInfo.Type, exec); err != nil {
		return fmt.Errorf("errors.build.steps.failed_install_ollama: %w", err)
	}

	// Clean build caches
	ui.ShowLog("info", ui.GetText("task.clean_caches"))
	if err := cleanBuildCaches(ctx, exec); err != nil {
		return fmt.Errorf("errors.build.steps.failed_clean_caches: %w", err)
	}

	return nil
}

// installOllama downloads and installs Ollama for the specified GPU type.
func installOllama(ctx context.Context, gpuType string, exec func(context.Context, string, ...string) error) error {
	arch, err := OllamaArch()
	if err != nil {
		return err
	}

	mainURL := fmt.Sprintf("https://ollama.com/download/ollama-linux-%s.tar.zst", arch)
	if err := exec(ctx, "curl", "-fsSL", mainURL, "-o", "/tmp/ollama.tar.zst"); err != nil {
		return err
	}
	if err := exec(ctx, "sudo", "tar", "--zstd", "-xf", "/tmp/ollama.tar.zst", "-C", "/usr"); err != nil {
		return err
	}

	if strings.HasPrefix(gpuType, "rdna") || gpuType == "amd" {
		rocmURL := fmt.Sprintf("https://ollama.com/download/ollama-linux-%s-rocm.tar.zst", arch)
		if err := exec(ctx, "curl", "-fsSL", rocmURL, "-o", "/tmp/ollama-rocm.tar.zst"); err != nil {
			return err
		}
		if err := exec(ctx, "sudo", "tar", "--zstd", "-xf", "/tmp/ollama-rocm.tar.zst", "-C", "/usr"); err != nil {
			return err
		}
	}

	return nil
}

// cleanBuildCaches removes temporary files and package manager cache.
func cleanBuildCaches(ctx context.Context, exec func(context.Context, string, ...string) error) error {
	if err := exec(ctx, "sudo", "pacman", "-Scc", "--noconfirm"); err != nil {
		return err
	}

	// Note: Getting HOME environment variable requires a different approach
	// For now, just clean known paths
	_ = exec(ctx, "sudo", "rm", "-rf", "/tmp/agent-teams", "/tmp/ollama.tar.zst", "/tmp/ollama-rocm.tar.zst", "/var/cache/pacman/pkg")
	return nil
}
