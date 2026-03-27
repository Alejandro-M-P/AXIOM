// Package ports define las interfaces (contratos) que el Core necesita
// para comunicarse con el mundo exterior.
// Este paquete no tiene implementaciones - solo define los contratos.
package ports

import "axiom/pkg/core/domain"

// IContainerRuntime define el contrato para operaciones de contenedores.
// Las implementaciones pueden usar Podman, Docker, u otro runtime.
type IContainerRuntime interface {
	// CreateContainer crea un nuevo contenedor con la imagen especificada.
	CreateContainer(name, image, home string, flags string) error

	// StartContainer inicia un contenedor existente.
	StartContainer(name string) error

	// StopContainer detiene un contenedor en ejecución.
	StopContainer(name string) error

	// RemoveContainer elimina un contenedor.
	RemoveContainer(name string, force bool) error

	// ListContainers lista todos los contenedores.
	ListContainers() ([]domain.ContainerInfo, error)

	// ContainerExists verifica si un contenedor existe.
	ContainerExists(name string) (bool, error)

	// ImageExists verifica si una imagen existe.
	ImageExists(image string) bool

	// CommitContainer crea una imagen a partir de un contenedor.
	CommitContainer(name, image string) error

	// RunCommand ejecuta un comando dentro de un contenedor.
	RunCommand(containerName string, args ...string) error

	// RunCommandWithInput ejecuta un comando con entrada stdin.
	RunCommandWithInput(containerName, input string, args ...string) error

	// RunCommandOutput ejecuta un comando y retorna la salida.
	RunCommandOutput(containerName string, args ...string) (string, error)
}

// IDistrobox proporciona operaciones específicas de Distrobox.
type IDistrobox interface {
	// Create crea un búnker usando distrobox-create.
	Create(name, image, home string, flags string) error

	// Enter entra en un búnker de forma interactiva.
	Enter(name string) error

	// Stop detiene un búnker.
	Stop(name string) error

	// Remove elimina un búnker.
	Remove(name string, force bool) error
}
