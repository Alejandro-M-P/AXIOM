package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/Alejandro-M-P/AXIOM/internal/domain"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// PodmanAdapter implementa IContainerRuntime usando Podman/Distrobox
type PodmanAdapter struct {
	cmds CommandSet
}

// NewPodmanAdapter crea un adapter de Podman
func NewPodmanAdapter() *PodmanAdapter {
	return &PodmanAdapter{cmds: Podman}
}

// NewPodmanAdapterWithCommands crea un adapter con CommandSet custom (para testing)
func NewPodmanAdapterWithCommands(cmds CommandSet) *PodmanAdapter {
	return &PodmanAdapter{cmds: cmds}
}

var _ ports.IBunkerRuntime = (*PodmanAdapter)(nil)

func (a *PodmanAdapter) CreateBunker(ctx context.Context, name, image, home, flags string) error {
	args := a.cmds.CreateBunker(name, image, home, flags)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("runtime.create_failed: %w", err)
	}
	return nil
}

// GetCreateFlags genera los flags para crear un bunker.
// volumeFlags viene del presenter (que usa i18n) - solo añade device flags aquí.
func (a *PodmanAdapter) GetCreateFlags(ctx context.Context, name, image, home, volumeFlags string) (string, error) {
	// Device flags - centralized in commands.go
	if volumeFlags != "" {
		return volumeFlags + " " + DeviceFlags, nil
	}
	return DeviceFlags, nil
}

func (a *PodmanAdapter) StartBunker(ctx context.Context, name string) error {
	args := a.cmds.StartBunker(name)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("runtime.start_failed: %w", err)
	}
	return nil
}

func (a *PodmanAdapter) StopBunker(ctx context.Context, name string) error {
	args := a.cmds.StopBunker(name)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("runtime.stop_failed: %w", err)
	}
	return nil
}

func (a *PodmanAdapter) RemoveBunker(ctx context.Context, name string, force bool) error {
	args := a.cmds.RemoveBunker(name, force)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("runtime.remove_failed: %w", err)
	}
	return nil
}

func (a *PodmanAdapter) ListBunkers(ctx context.Context) ([]domain.Bunker, error) {
	args := a.cmds.ListBunkers()
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("runtime.list_failed: %w", err)
	}

	var containers []struct {
		ID     string   `json:"Id"`
		Names  []string `json:"Names"`
		Image  string   `json:"Image"`
		Status string   `json:"Status"`
		State  string   `json:"State"`
	}

	if err := json.Unmarshal(output, &containers); err != nil {
		return nil, fmt.Errorf("runtime.parse_failed: %w", err)
	}

	result := make([]domain.Bunker, 0, len(containers))
	for _, c := range containers {
		if len(c.Names) > 0 {
			result = append(result, domain.Bunker{
				Name:   c.Names[0],
				Image:  c.Image,
				Status: c.Status,
			})
		}
	}

	return result, nil
}

func (a *PodmanAdapter) BunkerExists(ctx context.Context, name string) (bool, error) {
	bunkers, err := a.ListBunkers(ctx)
	if err != nil {
		return false, err
	}
	for _, b := range bunkers {
		if b.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (a *PodmanAdapter) ImageExists(ctx context.Context, image string) (bool, error) {
	args := a.cmds.ImageExists(image)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (a *PodmanAdapter) RemoveImage(ctx context.Context, image string, force bool) error {
	args := a.cmds.RemoveImage(image, force)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("runtime.remove_image_failed: %w", err)
	}
	return nil
}

func (a *PodmanAdapter) CommitImage(ctx context.Context, containerName, imageName, author, message string) error {
	args := a.cmds.CommitImage(containerName, imageName, author, message)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("runtime.commit_failed: %w", err)
	}
	return nil
}

func (a *PodmanAdapter) ContainerState(ctx context.Context, name string) (string, error) {
	args := a.cmds.ContainerState(name)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("runtime.state_failed: %w", err)
	}
	return string(output), nil
}

func (a *PodmanAdapter) StartContainer(ctx context.Context, name string) error {
	// StartBunker already does this - delegate to it
	return a.StartBunker(ctx, name)
}

func (a *PodmanAdapter) EnterBunker(ctx context.Context, name, rcPath string) error {
	args := a.cmds.EnterBunker(name, rcPath)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *PodmanAdapter) ExecuteInBunker(ctx context.Context, name string, args ...string) error {
	cmdArgs := a.cmds.ExecuteInBunker(name, args...)
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("runtime.execute_failed: %w", err)
	}
	return nil
}
