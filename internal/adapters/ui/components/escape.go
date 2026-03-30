package components

import (
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// EscapeButton es un boton estilizado para mostrar la opcion de salir con ESC.
// Se muestra usualmente en el footer de las pantallas TUI.
type EscapeButton struct {
	text  string
	theme *theme.Theme
}

// NewEscapeButton crea un nuevo EscapeButton con el texto por defecto.
func NewEscapeButton() *EscapeButton {
	return &EscapeButton{
		text:  "Salir", // default, will be overridden by caller
		theme: theme.DefaultTheme(),
	}
}

// WithText establece un texto personalizado para el boton.
func (b *EscapeButton) WithText(text string) *EscapeButton {
	if text != "" {
		b.text = text
	}
	return b
}

// Render devuelve el boton estilizado para mostrar en la UI.
// Usa los estilos existentes del theme de AXIOM para mantener consistencia.
func (b *EscapeButton) Render() string {
	// Estilo para el indicador [ESC]
	escStyle := lipgloss.NewStyle().
		Foreground(b.theme.Primary).
		Bold(true)

	// Estilo para el texto
	textStyle := lipgloss.NewStyle().
		Foreground(b.theme.Muted)

	return escStyle.Render("[ESC]") + " " + textStyle.Render(b.text)
}

// RenderAlt devuelve una version alternativa del boton con "q" como atajo.
func (b *EscapeButton) RenderAlt() string {
	escStyle := lipgloss.NewStyle().
		Foreground(b.theme.Primary).
		Bold(true)

	textStyle := lipgloss.NewStyle().
		Foreground(b.theme.Muted)

	return escStyle.Render("[q]") + " " + textStyle.Render(b.text)
}

// RenderCompact devuelve una version compacta en una sola linea.
func (b *EscapeButton) RenderCompact() string {
	style := lipgloss.NewStyle().
		Foreground(b.theme.Muted).
		Italic(true)

	return style.Render("ESC " + b.text)
}
