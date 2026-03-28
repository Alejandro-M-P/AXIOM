package build

import (
	"context"
	"fmt"
	"strings"

	"axiom/internal/ports"
)

// ModelStackConfig holds configuration for the AI/ML model stack installation.
type ModelStackConfig struct {
	GPUType string
}

// InstallSystemBase installs the base system packages in the container.
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

	progress := NewProgress(ui, "",
		ui.GetText("group.base"),
		[]ports.LifecycleStep{
			{Title: ui.GetText("task.sync_repos"), Detail: ui.GetText("detail.sync_cmd"), Status: ports.LifecyclePending},
			{Title: ui.GetText("task.install_pkgs"), Detail: ui.GetText("detail.pkgs_count", len(packages)), Status: ports.LifecyclePending},
		})

	// Step 0: Sync repos
	progress.StartStep(0)
	if err := exec(ctx, "sudo", "pacman", "-Sy", "--noconfirm"); err != nil {
		progress.FailStep(err)
		return err
	}
	progress.FinishStep()

	// Step 1: Install packages
	progress.StartStep(1)
	args := []string{"sudo", "pacman", "-S", "--needed", "--noconfirm"}
	args = append(args, packages...)
	if err := exec(ctx, args[0], args[1:]...); err != nil {
		progress.FailStep(err)
		return err
	}
	progress.FinishStep()

	// Step 2: Install Homebrew
	progress.StartStep(2)
	cmd := "NONINTERACTIVE=1 /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
	if err := exec(ctx, "bash", "-c", cmd); err != nil {
		progress.FailStep(err)
		return err
	}
	progress.FinishStep()

	return nil
}

// InstallDeveloperTools installs development tools (opencode, engram, gentle-ai).
func InstallDeveloperTools(ctx context.Context, containerName string, cfg *BuildContext, ui ports.IPresenter, exec func(context.Context, string, ...string) error) error {
	progress := NewProgress(ui, "",
		ui.GetText("group.dev"),
		[]ports.LifecycleStep{
			{Title: ui.GetText("task.install_opencode"), Detail: ui.GetText("detail.npm_global"), Status: ports.LifecyclePending},
			{Title: "Configurando Homebrew Tap", Detail: "Gentleman-Programming/homebrew-tap", Status: ports.LifecyclePending},
			{Title: "Instalando herramientas de desarrollo", Detail: "brew install engram gentle-ai", Status: ports.LifecyclePending},
		})

	// Step 0: Install opencode via npm
	progress.StartStep(0)
	if err := exec(ctx, "sudo", "npm", "install", "-g", "opencode-ai"); err != nil {
		progress.FailStep(err)
		return err
	}
	progress.FinishStep()

	brewPath := "/home/linuxbrew/.linuxbrew/bin/brew"

	// Step 1: Add Homebrew tap
	progress.StartStep(1)
	if err := exec(ctx, brewPath, "tap", "Gentleman-Programming/homebrew-tap"); err != nil {
		progress.FailStep(err)
		return err
	}
	progress.FinishStep()

	// Step 2: Install engram and gentle-ai
	progress.StartStep(2)
	if err := exec(ctx, brewPath, "install", "engram", "gentle-ai"); err != nil {
		progress.FailStep(err)
		return err
	}
	progress.FinishStep()

	return nil
}

// InstallModelStack installs Ollama and configures the AI stack.
func InstallModelStack(ctx context.Context, containerName string, cfg *BuildContext, modelConfig ModelStackConfig, ui ports.IPresenter, exec func(context.Context, string, ...string) error) error {
	progress := NewProgress(ui, "",
		ui.GetText("group.ai"),
		[]ports.LifecycleStep{
			{Title: ui.GetText("task.install_ollama"), Detail: cfg.GPUInfo.Type, Status: ports.LifecyclePending},
			{Title: ui.GetText("task.clean_caches"), Detail: ui.GetText("detail.tmp_pacman"), Status: ports.LifecyclePending},
		})

	// Step 0: Install Ollama
	progress.StartStep(0)
	if err := installOllama(ctx, cfg.GPUInfo.Type, exec); err != nil {
		progress.FailStep(err)
		return err
	}
	progress.FinishStep()

	// Step 1: Clean build caches
	progress.StartStep(1)
	if err := cleanBuildCaches(ctx, exec); err != nil {
		progress.FailStep(err)
		return err
	}
	progress.FinishStep()

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
