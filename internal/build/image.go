package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/domain"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// BuildContext holds all the context needed for a build operation.
type BuildContext struct {
	Config            domain.EnvConfig
	GPUInfo           *domain.GPUInfo
	ImageName         string
	SlotName          string
	BuildWorkspaceDir string
	ContainerName     string
}

// PrepareBuildContext creates a BuildContext from the environment configuration.
// It generates container and workspace names based on the slot type.
func PrepareBuildContext(ctx context.Context, cfg domain.EnvConfig, containerName, slotName string, system ports.ISystem) (*BuildContext, error) {
	gpuInfo, err := ResolveBuildGPU(ctx, cfg, system)
	if err != nil {
		return nil, fmt.Errorf("gpu_resolution: %w", err)
	}

	// Generate container name based on slot: axiom-dev, axiom-data, axiom-sandbox
	// If containerName is provided (non-empty), use it; otherwise generate from slot
	actualContainerName := containerName
	if actualContainerName == "" {
		actualContainerName = fmt.Sprintf("axiom-%s", slotName)
	}

	// Use slot name for image name: axiom-dev, axiom-data, axiom-sandbox
	imageName := fmt.Sprintf("localhost/axiom-%s:latest", slotName)

	return &BuildContext{
		Config:            cfg,
		GPUInfo:           gpuInfo,
		ImageName:         imageName,
		SlotName:          slotName,
		BuildWorkspaceDir: cfg.BuildWorkspaceDir(actualContainerName),
		ContainerName:     actualContainerName,
	}, nil
}

// PrepareSharedDirectories creates the necessary directories for the build.
func PrepareSharedDirectories(ctx context.Context, fs ports.IFileSystem, cfg domain.EnvConfig) error {
	if err := fs.MkdirAll(filepath.Join(cfg.AIConfigDir(), "models"), 0700); err != nil {
		return err
	}
	if err := fs.MkdirAll(filepath.Join(cfg.AIConfigDir(), "teams"), 0700); err != nil {
		return err
	}
	if err := ensureTutorFile(fs, cfg.TutorPath()); err != nil {
		return err
	}
	return nil
}

// RecreateBuildContainer removes any existing build container and creates a fresh one.
func RecreateBuildContainer(ctx context.Context, runtime ports.IBunkerRuntime, fs ports.IFileSystem, containerName string, buildWorkspaceDir string, cfg domain.EnvConfig) error {
	// Remove existing container if any
	_ = runtime.RemoveBunker(ctx, containerName, true)

	// Clean up build workspace
	if err := removePathWritable(fs, buildWorkspaceDir); err != nil {
		return err
	}

	// Create workspace directory
	if err := fs.MkdirAll(buildWorkspaceDir, 0700); err != nil {
		return err
	}

	// Build container flags
	flags := BuildContainerFlags(cfg)

	// Create container using the base image
	return runtime.CreateBunker(ctx, containerName, "archlinux:latest", buildWorkspaceDir, flags)
}

// BuildContainerFlags returns the docker/podman flags needed for the build container.
func BuildContainerFlags(cfg domain.EnvConfig) string {
	return fmt.Sprintf(
		"--volume %s:/ai_config:z --volume %s:/run/axiom/env:ro,z --device /dev/kfd --device /dev/dri --security-opt label=disable --group-add video --group-add render",
		cfg.AIConfigDir(),
		filepath.Join(cfg.AxiomPath, "config.toml"),
	)
}

// ExportBuildImage commits the build container to an image using the runtime.
// author and message are visible to the user (translatable via i18n).
func ExportBuildImage(ctx context.Context, runtime ports.IBunkerRuntime, containerName string, imageName string, author, message string) error {
	return runtime.CommitImage(ctx, containerName, imageName, author, message)
}

// DestroyBuildContainer removes the build container and cleans up workspace.
func DestroyBuildContainer(ctx context.Context, runtime ports.IBunkerRuntime, fs ports.IFileSystem, containerName string, buildWorkspaceDir string) error {
	if err := runtime.RemoveBunker(ctx, containerName, true); err != nil {
		return err
	}
	return removePathWritable(fs, buildWorkspaceDir)
}

// EnsureTutorFile ensures the tutor.md file exists.
func ensureTutorFile(fs ports.IFileSystem, path string) error {
	if err := fs.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if _, err := fs.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	file, err := fs.OpenFile(path, os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	return file.Close()
}

// removePathWritable makes all files writable then removes the path.
func removePathWritable(fs ports.IFileSystem, path string) error {
	if _, err := fs.Stat(path); os.IsNotExist(err) {
		return nil
	}
	_ = fs.WalkDir(path, func(currentPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		mode := info.Mode()
		if mode&0200 == 0 {
			_ = fs.Chmod(currentPath, mode|0200)
		}
		return nil
	})
	return fs.RemoveAll(path)
}

// RunInContainer executes a command in the build container via distrobox-enter.
func RunInContainer(ctx context.Context, runtime ports.IBunkerRuntime, containerName string, args ...string) error {
	return runtime.ExecuteInBunker(ctx, containerName, args...)
}

// RunInContainerWithInput executes a command with stdin input.
func RunInContainerWithInput(ctx context.Context, runtime ports.IBunkerRuntime, containerName string, input string, args ...string) error {
	// ExecuteInBunker doesn't support input - this would need a different approach
	// For now, combine input into a single command
	cmd := fmt.Sprintf("printf '%s' | %s", input, strings.Join(args, " "))
	return runtime.ExecuteInBunker(ctx, containerName, "bash", "-c", cmd)
}
