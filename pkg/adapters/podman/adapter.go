// Package podman implementa el adapter para operaciones de contenedores.
// Implementa la interfaz IContainerRuntime definida en pkg/core/ports.
package podman

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"axiom/pkg/core/domain"
	"axiom/pkg/core/ports"
)

const (
	defaultBuildContainerName = "axiom-build"
)

// PodmanAdapter implementa IContainerRuntime usando Podman y Distrobox.
type PodmanAdapter struct{}

// NewPodmanAdapter crea una nueva instancia del adapter.
func NewPodmanAdapter() *PodmanAdapter {
	return &PodmanAdapter{}
}

// EnsureDefaultBuildContainerName returns the default build container name.
func EnsureDefaultBuildContainerName() string {
	return defaultBuildContainerName
}

// CreateContainer crea un contenedor usando distrobox-create.
func (p *PodmanAdapter) CreateContainer(name, image, home string, flags string) error {
	args := []string{
		"distrobox-create",
		"--name", name,
		"--image", image,
		"--home", home,
		"--additional-flags", flags,
		"--yes",
	}
	return runCommand("distrobox-create", args...)
}

// StartContainer inicia un contenedor.
func (p *PodmanAdapter) StartContainer(name string) error {
	return runCommand("podman", "start", name)
}

// StopContainer detiene un contenedor en ejecución.
func (p *PodmanAdapter) StopContainer(name string) error {
	return runCommand("distrobox-stop", name, "--yes")
}

// RemoveContainer elimina un contenedor.
func (p *PodmanAdapter) RemoveContainer(name string, force bool) error {
	forceFlag := ""
	if force {
		forceFlag = "--force"
	}
	return runCommand("distrobox-rm", name, forceFlag, "--yes")
}

// ListContainers lista todos los contenedores.
func (p *PodmanAdapter) ListContainers() ([]domain.ContainerInfo, error) {
	output, err := runCommandOutput("podman", "ps", "-a", "--format", "json")
	if err != nil || strings.TrimSpace(output) == "" {
		return nil, nil
	}

	var containers []struct {
		Names []string `json:"Names"`
	}
	if err := json.Unmarshal([]byte(output), &containers); err != nil {
		return nil, err
	}

	var result []domain.ContainerInfo
	for _, c := range containers {
		for _, name := range c.Names {
			if name == "" || name == defaultBuildContainerName {
				continue
			}
			result = append(result, domain.ContainerInfo{
				Name:   name,
				Status: "unknown",
			})
		}
	}
	return result, nil
}

// ContainerExists verifica si un contenedor existe.
func (p *PodmanAdapter) ContainerExists(name string) (bool, error) {
	output, err := runCommandOutput("podman", "ps", "-a", "--format", "json")
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(output) == "" {
		return false, nil
	}

	var containers []struct {
		Names []string `json:"Names"`
	}
	if err := json.Unmarshal([]byte(output), &containers); err != nil {
		return false, err
	}

	for _, c := range containers {
		for _, cName := range c.Names {
			if cName == name {
				return true, nil
			}
		}
	}
	return false, nil
}

// ImageExists verifica si una imagen existe.
func (p *PodmanAdapter) ImageExists(image string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "podman", "image", "exists", image)
	return cmd.Run() == nil
}

// CommitContainer crea una imagen a partir de un contenedor.
func (p *PodmanAdapter) CommitContainer(containerName, imageName string) error {
	return runCommand("podman", "commit", containerName, imageName)
}

// RunCommand ejecuta un comando dentro de un contenedor, o en el host si el nombre está vacío.
func (p *PodmanAdapter) RunCommand(containerName string, args ...string) error {
	if containerName == "" {
		if len(args) == 0 {
			return nil
		}
		return runCommand(args[0], args[1:]...)
	}
	cmdArgs := append([]string{"-n", containerName, "--"}, args...)
	return runCommand("distrobox-enter", cmdArgs...)
}

func (p *PodmanAdapter) RunCommandWithInput(containerName, input string, args ...string) error {
	if containerName == "" {
		if len(args) == 0 {
			return nil
		}
		return runCommandWithInput(args[0], input, args[1:]...)
	}
	cmdArgs := append([]string{"-n", containerName, "--"}, args...)
	return runCommandWithInput("distrobox-enter", input, cmdArgs...)
}

func (p *PodmanAdapter) RunCommandOutput(containerName string, args ...string) (string, error) {
	if containerName == "" {
		if len(args) == 0 {
			return "", nil
		}
		return runCommandOutput(args[0], args[1:]...)
	}
	cmdArgs := append([]string{"-n", containerName, "--"}, args...)
	return runCommandOutput("distrobox-enter", cmdArgs...)
}

// DistroboxAdapter implementa IDistrobox.
type DistroboxAdapter struct{}

func NewDistroboxAdapter() *DistroboxAdapter {
	return &DistroboxAdapter{}
}

// Create crea un búnker usando distrobox-create.
func (d *DistroboxAdapter) Create(name, image, home string, flags string) error {
	args := []string{
		"distrobox-create",
		"--name", name,
		"--image", image,
		"--home", home,
		"--additional-flags", flags,
		"--yes",
	}
	return runCommand("distrobox-create", args...)
}

// Enter entra en un búnker de forma interactiva.
func (d *DistroboxAdapter) Enter(name string) error {
	cmd := exec.Command("distrobox-enter", name, "--", "bash", "--rcfile", "/dev/null", "-i")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Stop detiene un búnker.
func (d *DistroboxAdapter) Stop(name string) error {
	return runCommand("distrobox-stop", name, "--yes")
}

// Remove elimina un búnker.
func (d *DistroboxAdapter) Remove(name string, force bool) error {
	forceFlag := ""
	if force {
		forceFlag = "--force"
	}
	return runCommand("distrobox-rm", name, forceFlag, "--yes")
}

// VerifyInterface verifica que el adapter implementa la interfaz correctamente.
var _ ports.IContainerRuntime = (*PodmanAdapter)(nil)
var _ ports.IDistrobox = (*DistroboxAdapter)(nil)

// Funciones helper privadas.

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func runCommandWithInput(name, input string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func runCommandOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}