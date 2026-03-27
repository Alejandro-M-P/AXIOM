package ports

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

// Field representa un par clave-valor genérico (Ej: "GPU" -> "rdna4")
type Field struct {
	Label string
	Value string
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
}
