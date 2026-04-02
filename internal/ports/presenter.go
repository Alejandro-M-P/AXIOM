package ports

import "context"

// Constantes para el estado de los pasos del lifecycle.
const (
	LifecyclePending = "pending"
	LifecycleRunning = "running"
	LifecycleDone    = "done"
	LifecycleError   = "error"
)

// LifecycleStep representa un paso en el proceso de lifecycle.
type LifecycleStep struct {
	Title  string
	Detail string
	Status string
}

// IPresenter es el puerto que abstrae toda la presentación e interacción con el usuario.
// Las implementaciones pueden ser:
// - TUI (terminal interactiva con Bubbletea)
// - CLI (salida de línea de comando)
// - API (salida JSON para agentes)
// - Logger (solo logs)
type IPresenter interface {
	// ShowLogo muestra el logo de AXIOM.
	ShowLogo()

	// ShowCommandCard muestra una tarjeta de comando con campos.
	ShowCommandCard(commandKey string, fields []Field, items []string)

	// ShowWarning muestra una advertencia.
	ShowWarning(title, subtitle string, fields []Field, items []string, footer string)

	// ShowLog muestra un mensaje de log.
	ShowLog(logKey string, args ...any)

	// AskString solicita una cadena al usuario.
	AskString(promptKey string) (string, error)

	// AskConfirm solicita confirmación al usuario.
	AskConfirm(promptKey string, args ...any) (bool, error)

	// ShowHelp muestra la ayuda.
	ShowHelp()

	// GetText obtiene texto internacionalizado.
	GetText(key string, args ...any) string

	// ClearScreen limpia la pantalla.
	ClearScreen()

	// RenderLifecycle renderiza el progreso del lifecycle.
	RenderLifecycle(title, subtitle string, steps []LifecycleStep, taskTitle string, taskSteps []LifecycleStep)

	// RenderLifecycleError renderiza un error en el lifecycle.
	RenderLifecycleError(title string, steps []LifecycleStep, taskTitle string, taskSteps []LifecycleStep, err error, where string)

	// AskConfirmInCard solicita confirmación dentro de una tarjeta.
	AskConfirmInCard(commandKey string, fields []Field, items []string, promptKey string) (bool, error)

	// AskDelete solicita confirmación de eliminación.
	AskDelete(name string, fields []Field) (confirm bool, reason string, deleteCode bool, err error)

	// AskReset solicita confirmación de reset.
	AskReset(fields []Field, items []string) (confirm bool, reason string, err error)

	// AskCreateBunker solicita información para crear un nuevo bunker.
	AskCreateBunker(images []string) (name string, image string, confirmed bool, err error)

	// AskSelectBunker solicita selección de un bunker de forma interactiva.
	AskSelectBunker(bunkers []string, statuses map[string]string, title, subtitle string) (selected string, confirmed bool, err error)

	// RunHelpTUI ejecuta la ayuda en modo TUI interactivo.
	RunHelpTUI() error

	// RunInitWizard ejecuta el wizard de inicialización.
	RunInitWizard(ctx context.Context) error

	// RunInitWizardResult ejecuta el wizard de inicialización y devuelve si completó exitosamente.
	RunInitWizardResult(ctx context.Context) (bool, error)

	// RunInitWizardWithParams ejecuta el wizard de inicialización con parámetros específicos.
	RunInitWizardWithParams(ctx context.Context, axiomPath string, envExists bool, lang string, homeDir string) (bool, error)
}
