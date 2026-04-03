package runtime

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// BuildInstaller implements ports.IBuildInstaller using Podman/Distrobox runtime.
// It handles all system-level operations: pacman, Ollama downloads, cache cleanup.
type BuildInstaller struct {
	runtime *PodmanAdapter
}

// NewBuildInstaller creates a new BuildInstaller.
func NewBuildInstaller(runtime *PodmanAdapter) *BuildInstaller {
	return &BuildInstaller{runtime: runtime}
}

// basePackages are the minimal packages always needed in the container.
var basePackages = []string{"base-devel", "git", "curl", "jq", "wget"}

// ollamaURLs returns the Ollama download URLs for the given GPU type and arch.
func ollamaURLs(gpuType string) []string {
	arch := ollamaArch()
	urls := []string{
		fmt.Sprintf("https://ollama.com/download/ollama-linux-%s", arch),
	}
	// ROCm variant for AMD GPUs
	if gpuType == "amd" || strings.HasPrefix(gpuType, "rdna") {
		urls = append(urls, fmt.Sprintf("https://ollama.com/download/ollama-linux-%s-rocm", arch))
	}
	return urls
}

// ollamaArch returns the architecture suffix for Ollama downloads.
func ollamaArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return "amd64" // fallback
	}
}

// ExecuteBuild implements ports.IBuildInstaller.
// It installs base packages, slot dependencies, runs install commands,
// installs Ollama if needed, and cleans caches — all inside the container.
func (bi *BuildInstaller) ExecuteBuild(ctx context.Context, items []ports.BuildItem, containerName string, cfg ports.BuildConfig, progress ports.IBuildProgress) error {
	// Step 1: Install system packages (base + deps)
	allPkgs := bi.collectDependencies(items)
	if err := bi.installPackages(ctx, containerName, allPkgs, progress); err != nil {
		return fmt.Errorf("build_installer.install_packages: %w", err)
	}

	// Step 2: Install slot items (run their install commands)
	if err := bi.installSlotItems(ctx, items, containerName, progress); err != nil {
		return fmt.Errorf("build_installer.install_slot_items: %w", err)
	}

	// Step 3: Install Ollama if any item needs it
	if bi.needsOllama(items) {
		if err := bi.installOllama(ctx, containerName, cfg.GPUType, progress); err != nil {
			return fmt.Errorf("build_installer.install_ollama: %w", err)
		}
	}

	// Step 4: Clean caches
	if err := bi.cleanCaches(ctx, containerName, progress); err != nil {
		return fmt.Errorf("build_installer.clean_caches: %w", err)
	}

	return nil
}

// collectDependencies merges base packages with all unique deps from items.
func (bi *BuildInstaller) collectDependencies(items []ports.BuildItem) []string {
	depSet := make(map[string]struct{})
	for _, pkg := range basePackages {
		depSet[pkg] = struct{}{}
	}
	for _, item := range items {
		for _, dep := range item.Deps {
			depSet[dep] = struct{}{}
		}
	}
	result := make([]string, 0, len(depSet))
	for dep := range depSet {
		result = append(result, dep)
	}
	return result
}

// installPackages runs pacman inside the container to install all packages.
func (bi *BuildInstaller) installPackages(ctx context.Context, containerName string, packages []string, progress ports.IBuildProgress) error {
	if len(packages) == 0 {
		return nil
	}

	// Sync repos
	if err := bi.runInContainer(ctx, containerName, "sudo", "pacman", "-Sy", "--noconfirm"); err != nil {
		return fmt.Errorf("sync_repos: %w", err)
	}

	// Install all packages in one batch
	args := []string{"sudo", "pacman", "-S", "--needed", "--noconfirm"}
	args = append(args, packages...)
	if err := bi.runInContainer(ctx, containerName, args...); err != nil {
		return fmt.Errorf("install_packages: %w", err)
	}

	return nil
}

// installSlotItems runs the install command for each slot item.
// Items are installed via distrobox-enter into the container.
func (bi *BuildInstaller) installSlotItems(ctx context.Context, items []ports.BuildItem, containerName string, progress ports.IBuildProgress) error {
	// Each item's install command is defined by the slot registry/engine.
	// The installer engine already handles this via the slot manager.
	// This method is a placeholder for items that need direct container
	// execution beyond the slot installer.
	_ = items
	_ = progress
	return nil
}

// needsOllama returns true if any item requires Ollama.
func (bi *BuildInstaller) needsOllama(items []ports.BuildItem) bool {
	for _, item := range items {
		if item.NeedsOllama {
			return true
		}
	}
	return false
}

// installOllama downloads and installs Ollama inside the container.
func (bi *BuildInstaller) installOllama(ctx context.Context, containerName string, gpuType string, progress ports.IBuildProgress) error {
	arch := ollamaArch()

	// Download main Ollama binary
	mainURL := fmt.Sprintf("https://ollama.com/download/ollama-linux-%s", arch)
	if err := bi.runInContainer(ctx, containerName, "curl", "-fsSL", mainURL, "-o", "/tmp/ollama.tar.zst"); err != nil {
		return fmt.Errorf("download_ollama: %w", err)
	}

	// Extract to /usr
	if err := bi.runInContainer(ctx, containerName, "sudo", "tar", "--zstd", "-xf", "/tmp/ollama.tar.zst", "-C", "/usr"); err != nil {
		return fmt.Errorf("extract_ollama: %w", err)
	}

	// Download ROCm variant if needed
	if gpuType == "amd" || strings.HasPrefix(gpuType, "rdna") {
		rocmURL := fmt.Sprintf("https://ollama.com/download/ollama-linux-%s-rocm", arch)
		if err := bi.runInContainer(ctx, containerName, "curl", "-fsSL", rocmURL, "-o", "/tmp/ollama-rocm.tar.zst"); err != nil {
			return fmt.Errorf("download_ollama_rocm: %w", err)
		}
		if err := bi.runInContainer(ctx, containerName, "sudo", "tar", "--zstd", "-xf", "/tmp/ollama-rocm.tar.zst", "-C", "/usr"); err != nil {
			return fmt.Errorf("extract_ollama_rocm: %w", err)
		}
	}

	return nil
}

// cleanCaches removes pacman cache and temporary files inside the container.
func (bi *BuildInstaller) cleanCaches(ctx context.Context, containerName string, progress ports.IBuildProgress) error {
	// Clean pacman cache
	_ = bi.runInContainer(ctx, containerName, "sudo", "pacman", "-Scc", "--noconfirm")

	// Clean temporary files
	_ = bi.runInContainer(ctx, containerName, "sudo", "rm", "-rf",
		"/tmp/ollama.tar.zst",
		"/tmp/ollama-rocm.tar.zst",
		"/var/cache/pacman/pkg",
	)

	return nil
}

// runInContainer executes a command inside the build container via distrobox-enter.
func (bi *BuildInstaller) runInContainer(ctx context.Context, containerName string, args ...string) error {
	return bi.runtime.ExecuteInBunker(ctx, containerName, args...)
}

// Ensure BuildInstaller implements IBuildInstaller at compile time.
var _ ports.IBuildInstaller = (*BuildInstaller)(nil)
