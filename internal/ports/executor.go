package ports

import "context"

// ICommandRunner define el contrato para ejecutar comandos de shell.
// El core usa esta abstracción para no depender de exec.Command directamente.
type ICommandRunner interface {
	// RunShell ejecuta un comando en el shell del sistema.
	// Devuelve la salida combinada (stdout + stderr) y cualquier error.
	RunShell(ctx context.Context, cmd string) ([]byte, error)
}
