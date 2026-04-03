// Package ports define las interfaces (contratos) que el Core necesita
// para comunicarse con el mundo exterior.
// Este paquete no tiene implementaciones - solo define los contratos.
package ports

import (
	"context"

	"github.com/Alejandro-M-P/AXIOM/internal/core/domain"
)

// IBunkerRuntime define el contrato para operaciones de bunkers (contenedores AXIOM).
// Las implementaciones pueden usar Podman/Distrobox, Docker, u otro runtime.
// TODO: Cambiar a IRuntime cuando se soporte múltiples runtimes
type IBunkerRuntime interface {
	// CreateBunker crea un nuevo bunker con la imagen especificada.
	CreateBunker(ctx context.Context, name, image, home string, flags string) error

	// GetCreateFlags generates the flags for creating a bunker.
	// volumeFlags comes from GetVolumeFlags - the runtime adds device flags based on gpuType.
	GetCreateFlags(ctx context.Context, name, image, home, volumeFlags, gpuType string) (string, error)

	// GetVolumeFlags returns the volume mount flags for container creation.
	// These are technical flags (--volume ...), not user-visible text.
	GetVolumeFlags(ctx context.Context, projectDir, name, aiConfigDir, configPath, gpuType, sshSocket string) (string, error)

	// GetGPUDeviceFlags returns the GPU device flags required for container creation.
	// Returns nil for generic/no GPU. These are runtime-specific (Podman vs Docker).
	GetGPUDeviceFlags(ctx context.Context, gpuType string) []string

	// StartBunker inicia un bunker existente.
	StartBunker(ctx context.Context, name string) error

	// StopBunker detiene un bunker en ejecución.
	StopBunker(ctx context.Context, name string) error

	// RemoveBunker elimina un bunker.
	RemoveBunker(ctx context.Context, name string, force bool) error

	// ListBunkers lista todos los bunkers.
	ListBunkers(ctx context.Context) ([]domain.Bunker, error)

	// BunkerExists verifica si un bunker existe.
	BunkerExists(ctx context.Context, name string) (bool, error)

	// ImageExists verifica si una imagen existe.
	ImageExists(ctx context.Context, image string) (bool, error)

	// IsAxiomImage devuelve true si la imagen pertenece a AXIOM.
	IsAxiomImage(image string) bool

	// RemoveImage elimina una imagen.
	RemoveImage(ctx context.Context, image string, force bool) error

	// CommitImage hace commit de un contenedor a una imagen.
	// author y message son textos visibles para el usuario (traducibles).
	CommitImage(ctx context.Context, containerName, imageName, author, message string) error

	// ContainerState devuelve el estado de un contenedor (running, exited, etc.).
	ContainerState(ctx context.Context, name string) (string, error)

	// StartContainer inicia un contenedor existente.
	StartContainer(ctx context.Context, name string) error

	// EnterBunker entra en un bunker de forma interactiva (TTY).
	// rcPath es la ruta al archivo rcfile de bash (puede ser vacío).
	EnterBunker(ctx context.Context, name, rcPath string) error

	// ExecuteInBunker ejecuta un comando dentro de un bunker.
	ExecuteInBunker(ctx context.Context, name string, args ...string) error

	// ExecuteWithInput ejecuta un comando dentro de un bunker pasando input por stdin.
	ExecuteWithInput(ctx context.Context, name, input string, args ...string) error
}

// NOTE: IDistrobox fue eliminada. Toda la funcionalidad de Distrobox está
// integrada en IBunkerRuntime. El adapter de Podman usa distrobox-create,
// distrobox-enter, distrobox-stop, distrobox-rm internamente, pero el
// Core no lo sabe - solo conoce IBunkerRuntime.
