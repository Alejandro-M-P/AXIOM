package styles

import (
	_ "embed"

	"axiom/internal/adapters/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

//go:embed logo.txt
var logoASCII string

// defaultTheme cachea el tema para evitar recrearlo en cada acceso
var defaultTheme = theme.DefaultTheme()

// GetTheme devuelve el tema centralizado para uso externo
func GetTheme() *theme.Theme {
	return defaultTheme
}

var (
	// Paleta de colores Nord/Axiom - Obtenidas del Theme centralizado
	Cyan  = defaultTheme.Primary
	Green = defaultTheme.Success
	Red   = defaultTheme.Error
	Gray  = defaultTheme.Muted
	White = defaultTheme.Text
	Dark  = defaultTheme.Background

	// 1. EL CONTENEDOR (La "Cajita")
	WindowStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(defaultTheme.Primary).
			Padding(1, 4).
			Margin(1, 2).
			Width(80)

	// 2. TEXTOS Y TÍTULOS
	HeaderStyle = lipgloss.NewStyle().
			Foreground(defaultTheme.Background).
			Background(defaultTheme.Primary).
			Padding(0, 1).
			Bold(true).
			MarginBottom(1)

	ExampleStyle = lipgloss.NewStyle().
			Foreground(defaultTheme.Muted).
			Italic(true)

	GreenStyle = lipgloss.NewStyle().
			Foreground(defaultTheme.Success).
			Bold(true)

	// 3. BOTONES (Active/Inactive)
	ActiveButton = lipgloss.NewStyle().
			Foreground(defaultTheme.Text).
			Background(defaultTheme.Success).
			Padding(0, 3).
			MarginRight(2).
			Bold(true)

	InactiveButton = lipgloss.NewStyle().
			Foreground(defaultTheme.Text).
			Background(defaultTheme.Muted).
			Padding(0, 3).
			MarginRight(2)

	// 4. ESTILO DEL LOGO
	LogoStyle = lipgloss.NewStyle().Foreground(defaultTheme.Primary).Bold(true)
)

// GetLogo devuelve el logo ASCII para usarlo en cualquier comando
func GetLogo() string {
	return LogoStyle.Render("\n" + logoASCII + "\n")
}
