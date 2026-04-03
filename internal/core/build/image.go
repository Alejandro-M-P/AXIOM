package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Alejandro-M-P/AXIOM/internal/config"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// BuildContext holds all the context needed for a build operation.
type BuildContext struct {
	Config            config.EnvConfig
	GPUInfo           ports.GPUInfo
	ImageName         string
	SlotName          string
	BuildWorkspaceDir string
	ContainerName     string
}

// PrepareBuildContext creates a BuildContext from the environment configuration.
// It generates container and workspace names based on the slot type.
func PrepareBuildContext(ctx context.Context, cfg config.EnvConfig, containerName, slotName string, system ports.ISystem, presenter ports.IPresenter) (*BuildContext, error) {
	gpuInfo, err := ResolveBuildGPU(ctx, cfg, system, presenter)
	if err != nil {
		return nil, fmt.Errorf("errors.build.image.gpu_resolution: %w", err)
	}

	// Generate container name based on slot: axiom-dev, axiom-data, axiom-sandbox
	// If containerName is provided (non-empty), use it; otherwise generate from slot
	actualContainerName := containerName
	if actualContainerName == "" {
		actualContainerName = fmt.Sprintf("axiom-%s", slotName)
	}

	// Use slot name for image name: axiom-dev, axiom-data, axiom-sandbox
	imageName := fmt.Sprintf("axiom-%s", slotName)

	return &BuildContext{
		Config:            cfg,
		GPUInfo:           gpuInfo,
		ImageName:         imageName,
		SlotName:          slotName,
		BuildWorkspaceDir: config.BuildWorkspaceDir(cfg.BaseDir, actualContainerName),
		ContainerName:     actualContainerName,
	}, nil
}

// PrepareSharedDirectories creates the necessary directories for the build.
func PrepareSharedDirectories(ctx context.Context, fs ports.IFileSystem, cfg config.EnvConfig) error {
	if err := fs.MkdirAll(filepath.Join(config.AIConfigDir(cfg.BaseDir), "models"), 0700); err != nil {
		return err
	}
	if err := fs.MkdirAll(filepath.Join(config.AIConfigDir(cfg.BaseDir), "teams"), 0700); err != nil {
		return err
	}
	if err := ensureTutorFile(fs, config.TutorPath(cfg.BaseDir)); err != nil {
		return err
	}
	return nil
}

// RecreateBuildContainer removes any existing build container and creates a fresh one.
func RecreateBuildContainer(ctx context.Context, runtime ports.IBunkerRuntime, fs ports.IFileSystem, containerName string, buildWorkspaceDir string, cfg config.EnvConfig, gpuType string) error {
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

	// Get volume flags from runtime (core doesn't know Podman flag format)
	aiConfigDir := config.AIConfigDir(cfg.BaseDir)
	configPath := filepath.Join(cfg.AxiomPath, "config.toml")
	volumeFlags, err := runtime.GetVolumeFlags(ctx, buildWorkspaceDir, containerName, aiConfigDir, configPath, gpuType, "")
	if err != nil {
		return fmt.Errorf("errors.build.get_volume_flags: %w", err)
	}

	// Get full create flags (volumes + GPU devices) from runtime
	flags, err := runtime.GetCreateFlags(ctx, containerName, "archlinux:latest", buildWorkspaceDir, volumeFlags, gpuType)
	if err != nil {
		return fmt.Errorf("errors.build.get_create_flags: %w", err)
	}

	// Create container using the base image
	return runtime.CreateBunker(ctx, containerName, "archlinux:latest", buildWorkspaceDir, flags)
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
	// El core NO usa os.IsNotExist - usa el método Exists del puerto
	if fs.Exists(path) {
		return nil
	}
	// O_CREATE = 64 en Go
	file, err := fs.CreateFile(path, 0600)
	if err != nil {
		return err
	}
	return file.Close()
}

// removePathWritable makes all files writable then removes the path.
func removePathWritable(fs ports.IFileSystem, path string) error {
	// El core NO usa os.IsNotExist - usa el método Exists del puerto
	if !fs.Exists(path) {
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
	return runtime.ExecuteWithInput(ctx, containerName, input, args...)
}
