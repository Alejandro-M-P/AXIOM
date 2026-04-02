package ports

import "context"

// ICommandRunner define el contrato para ejecutar comandos de shell.
// El core usa esta abstracción para no depender de exec.Command directamente.
type ICommandRunner interface {
	// RunShell ejecuta un comando en el shell del sistema.
	// Devuelve la salida combinada (stdout + stderr) y cualquier error.
	RunShell(ctx context.Context, cmd string) ([]byte, error)
}

// CommandExecutor is a port for executing shell commands with progress reporting.
// Used by slot implementations to run installation commands.
type CommandExecutor interface {
	Execute(ctx context.Context, msg string, name string, args ...string) error
}
