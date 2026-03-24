package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Paleta de colores Nord/Axiom
	Cyan   = lipgloss.Color("#88c0d0")
	Green  = lipgloss.Color("#a3be8c")
	Red    = lipgloss.Color("#bf616a")
	Gray   = lipgloss.Color("#4c566a")
	White  = lipgloss.Color("#eceff4")
	Dark   = lipgloss.Color("#2e3440")

	// 1. EL CONTENEDOR (La "Cajita")
	WindowStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Cyan).
			Padding(1, 4).
			Margin(1, 2).
			Width(80)

	// 2. TEXTOS Y TÍTULOS
	HeaderStyle = lipgloss.NewStyle().
			Foreground(Dark).
			Background(Cyan).
			Padding(0, 1).
			Bold(true).
			MarginBottom(1)

	ExampleStyle = lipgloss.NewStyle().
			Foreground(Gray).
			Italic(true)

	GreenStyle = lipgloss.NewStyle().
			Foreground(Green).
			Bold(true)

	// 3. BOTONES (Active/Inactive)
	ActiveButton = lipgloss.NewStyle().
			Foreground(White).
			Background(Green).
			Padding(0, 3).
			MarginRight(2).
			Bold(true)

	InactiveButton = lipgloss.NewStyle().
			Foreground(White).
			Background(Gray).
			Padding(0, 3).
			MarginRight(2)

	// 4. ESTILO DEL LOGO
	LogoStyle = lipgloss.NewStyle().Foreground(Cyan).Bold(true)
)

// GetLogo devuelve el logo ASCII para usarlo en cualquier comando
func GetLogo() string {
	ascii := `
  █████╗ ██╗  ██╗██╗ ██████╗ ███╗   ███╗
 ██╔══██╗╚██╗██╔╝██║██╔═══██╗████╗ ████║
 ███████║ ╚███╔╝ ██║██║   ██║██╔████╔██║
 ██╔══██║ ██╔██╗ ██║██║   ██║██║╚██╔╝██║
 ██║  ██║██╔╝ ██╗██║╚██████╔╝██║ ╚═╝ ██║
 ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚═╝     ╚═╝
`
	return LogoStyle.Render(ascii)
}