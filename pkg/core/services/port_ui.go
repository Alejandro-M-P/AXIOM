package bunker

import (
	"axiom/pkg/core/ports"
)

const (
	LifecyclePending = "pending"
	LifecycleRunning = "running"
	LifecycleDone    = "done"
	LifecycleError   = "error"
)

// LifecycleStep es un alias de ports.LifecycleStep para compatibilidad.
type LifecycleStep = ports.LifecycleStep

// UI es el puerto que abstrae toda la presentación e interacción.
type UI interface {
	ShowLogo()
	ShowCommandCard(commandKey string, fields []ports.Field, items []string)
	ShowWarning(title, subtitle string, fields []ports.Field, items []string, footer string)
	ShowLog(logKey string, args ...any)
	AskString(promptKey string) (string, error)
	AskConfirm(promptKey string, args ...any) (bool, error)
	ShowHelp()
	GetText(key string, args ...any) string
	ClearScreen()
	RenderLifecycle(title, subtitle string, steps []ports.LifecycleStep, taskTitle string, taskSteps []ports.LifecycleStep)
	RenderLifecycleError(title string, steps []ports.LifecycleStep, taskTitle string, taskSteps []ports.LifecycleStep, err error, where string)
	AskConfirmInCard(commandKey string, fields []ports.Field, items []string, promptKey string) (bool, error)
	AskDelete(name string, fields []ports.Field) (confirm bool, reason string, deleteCode bool, err error)
	AskReset(fields []ports.Field, items []string) (confirm bool, reason string, err error)
}
