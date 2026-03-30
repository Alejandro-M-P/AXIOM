package styles

import (
	_ "embed"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/theme"
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

// ============================================================
// ESTILOS RESPONSIVOS (Fullscreen TUI)
// ============================================================

// WindowStyleWithDimensions crea un estilo de ventana con dimensiones dinamicas.
// Ajusta el ancho y margen segun el tamano de la terminal disponible.
func WindowStyleWithDimensions(width, height int) lipgloss.Style {
	// Calcular ancho_optimo: 80% del ancho disponible, maximo 100, minimo 60
	optimalWidth := width * 80 / 100
	if optimalWidth > 100 {
		optimalWidth = 100
	}
	if optimalWidth < 60 {
		optimalWidth = 60
	}

	// Calcular margen horizontal
	margin := (width - optimalWidth) / 2
	if margin < 2 {
		margin = 2
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(defaultTheme.Primary).
		Padding(1, 4).
		Margin(1, margin).
		Width(optimalWidth)
}

// HeaderStyleWithDimensions crea un estilo de header dinamico.
func HeaderStyleWithDimensions(width int) lipgloss.Style {
	// Ajustar el ancho del header al contenedor
	maxWidth := width - 12 // Compensar por bordes y padding

	return lipgloss.NewStyle().
		Foreground(defaultTheme.Background).
		Background(defaultTheme.Primary).
		Padding(0, 1).
		Bold(true).
		MarginBottom(1).
		Width(maxWidth)
}

// ContentStyleWithDimensions crea un estilo de contenido responsivo.
func ContentStyleWithDimensions(width, height int) lipgloss.Style {
	optimalWidth := width * 80 / 100
	if optimalWidth > 100 {
		optimalWidth = 100
	}
	if optimalWidth < 60 {
		optimalWidth = 60
	}

	margin := (width - optimalWidth) / 2
	if margin < 2 {
		margin = 2
	}

	return lipgloss.NewStyle().
		Width(optimalWidth).
		Margin(0, margin)
}

// ButtonStyleWithDimensions crea un estilo de boton que se adapta al ancho disponible.
func ButtonStyleWithDimensions(width int, active bool) lipgloss.Style {
	base := lipgloss.NewStyle().
		Padding(0, 3).
		MarginRight(2)

	if active {
		return base.
			Foreground(defaultTheme.Text).
			Background(defaultTheme.Success).
			Bold(true)
	}

	return base.
		Foreground(defaultTheme.Text).
		Background(defaultTheme.Muted)
}

// FooterStyleWithDimensions crea un estilo de footer responsivo.
func FooterStyleWithDimensions(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(defaultTheme.Muted).
		Width(width - 4).
		MarginLeft(2)
}

// CalculateOptimalDimensions calcula las dimensiones optimas para el contenido.
// Devuelve (contentWidth, contentHeight) considerando margins y padding.
func CalculateOptimalDimensions(width, height int) (int, int) {
	// Usar 90% del espacio disponible para contenido
	contentWidth := width * 90 / 100
	if contentWidth > 120 {
		contentWidth = 120
	}
	if contentWidth < 50 {
		contentWidth = 50
	}

	contentHeight := height * 85 / 100
	if contentHeight > 40 {
		contentHeight = 40
	}
	if contentHeight < 20 {
		contentHeight = 20
	}

	return contentWidth, contentHeight
}
