package bunker

const (
	LifecyclePending = "pending"
	LifecycleRunning = "running"
	LifecycleDone    = "done"
	LifecycleError   = "error"
)

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

// UI es el puerto que abstrae toda la presentación e interacción.
type UI interface {
	ShowLogo()
	ShowCommandCard(commandKey string, fields []Field, items []string)
	ShowWarning(title, subtitle string, fields []Field, items []string, footer string)
	ShowLog(logKey string, args ...any)
	AskString(promptKey string) (string, error)
	AskConfirm(promptKey string, args ...any) (bool, error)
	ShowHelp()
	GetText(key string, args ...any) string
	ClearScreen()
	RenderLifecycle(title, subtitle string, steps []LifecycleStep, taskTitle string, taskSteps []LifecycleStep)
	RenderLifecycleError(title string, steps []LifecycleStep, taskTitle string, taskSteps []LifecycleStep, err error, where string)
	AskConfirmInCard(commandKey string, fields []Field, items []string, promptKey string) (bool, error)
	AskDelete(name string, fields []Field) (confirm bool, reason string, deleteCode bool, err error)
	AskReset(fields []Field, items []string) (confirm bool, reason string, err error)
}