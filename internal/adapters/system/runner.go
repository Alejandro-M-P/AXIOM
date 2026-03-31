package system

import (
	"context"
	"os/exec"
)

// ShellRunner implementa ports.ICommandRunner.
// Es el único lugar fuera de adapters/runtime y slots/base donde
// exec.CommandContext puede existir (Regla 1 — Golden Rules).
type ShellRunner struct{}

// NewShellRunner crea una nueva instancia de ShellRunner.
func NewShellRunner() *ShellRunner {
	return &ShellRunner{}
}

// RunShell ejecuta un comando en el shell del sistema.
// Devuelve la salida combinada (stdout + stderr) y cualquier error.
func (r *ShellRunner) RunShell(ctx context.Context, cmd string) ([]byte, error) {
	return exec.CommandContext(ctx, "sh", "-c", cmd).CombinedOutput()
}

// Verificación de compilación: ShellRunner implementa ports.ICommandRunner
var _ interface {
	RunShell(ctx context.Context, cmd string) ([]byte, error)
} = (*ShellRunner)(nil)
