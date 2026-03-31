package ports

import (
	"context"

	"github.com/Alejandro-M-P/AXIOM/internal/domain"
)

// ISystem define el contrato para operaciones del sistema.
// Incluye detección de hardware, verificación de dependencias, etc.
type ISystem interface {
	// DetectGPU detecta la GPU del sistema.
	DetectGPU() domain.GPUInfo

	// CheckDeps verifica que las dependencias necesarias estén instaladas.
	CheckDeps() error

	// RefreshSudo renueva el ticket de sudo.
	RefreshSudo(ctx context.Context) error

	// UserHomeDir retorna el directorio home del usuario.
	UserHomeDir() (string, error)

	// SSHKeyPath retorna la ruta de la clave SSH por defecto.
	SSHKeyPath() (string, error)

	// SSHAgentSocket retorna la ruta del socket del agente SSH si está activo.
	SSHAgentSocket() (string, error)

	// PrepareSSHAgent prepara el agent SSH con la clave por defecto.
	PrepareSSHAgent(ctx context.Context) error

	// GetCommandPath retorna la ruta absoluta de un comando en el PATH.
	GetCommandPath(name string) (string, error)
}

// IDependencyChecker verifica la disponibilidad de herramientas del sistema.
type IDependencyChecker interface {
	// HasCommand verifica si un comando está disponible en el PATH.
	HasCommand(name string) bool

	// GetCommandPath retorna la ruta absoluta de un comando.
	GetCommandPath(name string) (string, error)
}
