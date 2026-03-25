package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colores específicos para errores (brillantes y claros)
	FailBorderColor = lipgloss.Color("#FF5C5C") // Rojo coral brillante
	FailTextColor   = lipgloss.Color("#FF7676") // Tono más suave para el texto del título
	FailCommandColor = lipgloss.Color("#EBCB8B") // Amarillo cálido para que el comando destaque

	ErrorWindowStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(FailBorderColor).
		Padding(1, 2).
		Margin(1, 2).
		Width(84)

	ErrorTitleStyle = lipgloss.NewStyle().
		Foreground(FailTextColor).
		Bold(true)

	ErrorCommandStyle = lipgloss.NewStyle().
		Foreground(FailCommandColor).
		Bold(true).
		MarginTop(1).
		MarginBottom(1)

	ErrorDescStyle = lipgloss.NewStyle().
		Foreground(White).
		MarginBottom(1)

	ActionIconStyle = lipgloss.NewStyle().
		Foreground(Green).
		MarginRight(1)

	ActionTextStyle = lipgloss.NewStyle().
		Foreground(Cyan).
		Bold(true).
		Italic(true)
)

// RenderErrorCard dibuja una tarjeta de error estandarizada
func RenderErrorCard(command, title, description, action string) string {
	var lines []string

	lines = append(lines, ErrorTitleStyle.Render("🚨 "+strings.ToUpper(title)))

	if command != "" {
		lines = append(lines, ErrorCommandStyle.Render(fmt.Sprintf("Comando ejecutado: axiom %s", command)))
	}

	lines = append(lines, ErrorDescStyle.Render(description))

	if action != "" {
		lines = append(lines, ActionIconStyle.Render("💡 Sugerencia:")+ActionTextStyle.Render(action))
	}

	return ErrorWindowStyle.Render(strings.Join(lines, "\n"))
}