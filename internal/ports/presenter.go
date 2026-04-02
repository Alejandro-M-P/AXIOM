package ports

import (
	"context"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/components"
)

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

// =====================
// INTERFACES ESPECÍFICAS
// =====================

// IBasicPresenter: operaciones fundamentales de presentación
type IBasicPresenter interface {
	ShowLogo()
	ShowLog(logKey string, args ...any)
	GetText(key string, args ...any) string
	ClearScreen()
	ShowHelp()
}

// ICardPresenter: renderizado de tarjetas y diálogos
type ICardPresenter interface {
	ShowCommandCard(commandKey string, fields []components.CardField, items []string)
	ShowWarning(title, subtitle string, fields []components.CardField, items []string, footer string)
	AskConfirmInCard(commandKey string, fields []components.CardField, items []string, promptKey string) (bool, error)
}

// ILifecyclePresenter: renderizado de lifecycle steps
type ILifecyclePresenter interface {
	RenderLifecycle(title, subtitle string, steps []LifecycleStep, taskTitle string, taskSteps []LifecycleStep)
	RenderLifecycleError(title string, steps []LifecycleStep, taskTitle string, taskSteps []LifecycleStep, err error, where string)
}

// IBunkerFlowPresenter: operaciones de flow de bunkers
type IBunkerFlowPresenter interface {
	AskDelete(name string, fields []components.CardField) (confirm bool, reason string, deleteCode bool, err error)
	AskReset(fields []components.CardField, items []string) (confirm bool, reason string, err error)
	AskCreateBunker(images []string) (name string, image string, confirmed bool, err error)
	AskSelectBunker(bunkers []string, statuses map[string]string, title, subtitle string) (selected string, confirmed bool, err error)
}

// IInteractivePresenter: input/output básico
type IInteractivePresenter interface {
	AskString(promptKey string) (string, error)
	AskConfirm(promptKey string, args ...any) (bool, error)
}

// IWizardPresenter: wizards de inicialización
type IWizardPresenter interface {
	RunHelpTUI() error
	RunInitWizard(ctx context.Context) error
	RunInitWizardResult(ctx context.Context) (bool, error)
	RunInitWizardWithParams(ctx context.Context, axiomPath string, envExists bool, lang string, homeDir string) (bool, error)
}

// IPresenter es el puerto que abstrae toda la presentación e interacción con el usuario.
// Las implementaciones pueden ser:
// - TUI (terminal interactiva con Bubbletea)
// - CLI (salida de línea de comando)
// - API (salida JSON para agentes)
// - Logger (solo logs)
type IPresenter interface {
	IBasicPresenter
	ICardPresenter
	ILifecyclePresenter
	IBunkerFlowPresenter
	IInteractivePresenter
	IWizardPresenter
}
