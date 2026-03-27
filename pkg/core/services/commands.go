package bunker

import "strings"

// Run actúa como orquestador público del paquete.
// Cada comando de bunker se resuelve aquí y luego se delega a su implementación.
func (m *Manager) Run(command string, args []string) error {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "help":
		return m.Help()
	case "build":
		return m.Build()
	case "list":
		return m.List()
	case "create":
		return m.Create(firstArg(args))
	case "stop":
		return m.Stop()
	case "delete":
		return m.Delete(firstArg(args))
	case "delete-image":
		return m.DeleteImage()
	case "rebuild":
		return m.Rebuild()
	case "reset":
		return m.Reset()
	case "prune":
		return m.Prune()
	case "info":
		return m.Info(firstArg(args))
	default:
		return nil
	}
}

// KnownCommand verifica si un comando es conocido.
func KnownCommand(command string) bool {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "help", "build", "list", "create", "stop", "delete", "delete-image", "rebuild", "reset", "prune", "info":
		return true
	default:
		return false
	}
}

// firstArg retorna el primer elemento de args o empty string si no hay elementos.
func firstArg(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}
