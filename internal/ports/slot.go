package ports

import "context"

// SlotItemDisplay representa un item de slot para mostrar en la UI.
// Solo contiene información visual, sin dependencias ni categoría.
type SlotItemDisplay struct {
	ID          string
	Name        string
	Description string
}

// ISlotInstaller define el contrato para instaladores de slots.
// Cada tipo de slot (dev/data/sandbox) puede tener su instalador.
type ISlotInstaller interface {
	// Install ejecuta la instalación del item usando el command runner.
	Install(ctx context.Context, item SlotItem, exec ICommandRunner) error
	// Name retorna el nombre del instalador (ej: "dev-installer").
	Name() string
}

// IPackageInstaller define el contrato para instalar paquetes del sistema.
// El adaptador de runtime (Arch Linux) implementa esto con pacman/brew.
type IPackageInstaller interface {
	// InstallPackages instala una lista de paquetes del sistema.
	InstallPackages(ctx context.Context, packages []string) error
	// InstallHomebrew instala Homebrew en el sistema.
	InstallHomebrew(ctx context.Context) error
}
