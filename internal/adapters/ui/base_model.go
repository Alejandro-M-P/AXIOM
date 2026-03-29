package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// BaseModel es un struct embebible que proporciona funcionalidad común
// para manejar el tamaño de la ventana en aplicaciones Bubbletea.
// Puede ser embebido en cualquier modelo para obtener soporte automático
// de redimensionamiento de ventana.
type BaseModel struct {
	Width  int
	Height int
}

// Init envía un mensaje WindowSizeMsg inicial para obtener las dimensiones
// de la terminal al iniciar. Esto asegura que el modelo tenga las medidas
// correctas desde el primer render.
func (m *BaseModel) Init() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{}
	}
}

// Update maneja tea.WindowSizeMsg para actualizar las dimensiones del modelo.
// Cualquier modelo que embeba BaseModel debe llamar a este método en su Update
// para mantener sincronizado el tamaño de la ventana.
func (m *BaseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return nil, nil
}
