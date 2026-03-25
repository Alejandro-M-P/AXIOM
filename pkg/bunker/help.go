package bunker

// Help muestra los comandos disponibles del orquestador bunker.
func (m *Manager) Help() error {
	m.UI.ShowLogo()
	m.UI.ShowHelp()
	return nil
}
