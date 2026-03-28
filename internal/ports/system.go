package ports

import "axiom/internal/domain"

// ISystem define el contrato para operaciones del sistema.
// Incluye detección de hardware, verificación de dependencias, etc.
type ISystem interface {
	// DetectGPU detecta la GPU del sistema.
	DetectGPU() domain.GPUInfo

	// CheckDeps verifica que las dependencias necesarias estén instaladas.
	CheckDeps() error
}

// IDependencyChecker verifica la disponibilidad de herramientas del sistema.
type IDependencyChecker interface {
	// HasCommand verifica si un comando está disponible en el PATH.
	HasCommand(name string) bool

	// GetCommandPath retorna la ruta absoluta de un comando.
	GetCommandPath(name string) (string, error)
}
