package theme

import (
	_ "embed"

	"github.com/charmbracelet/lipgloss"
)

//go:embed logo.txt
var LogoASCII string

type Theme struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Success    lipgloss.Color
	Error      lipgloss.Color
	Warning    lipgloss.Color
	Muted      lipgloss.Color
	Text       lipgloss.Color
	Background lipgloss.Color

	HeaderStyle     lipgloss.Style
	TitleStyle      lipgloss.Style
	SubtitleStyle   lipgloss.Style
	HelpStyle       lipgloss.Style
	SelectedStyle   lipgloss.Style
	UnselectedStyle lipgloss.Style
	CursorStyle     lipgloss.Style
	LogoStyle       lipgloss.Style
}

func DefaultTheme() *Theme {
	t := &Theme{
		Primary:    lipgloss.Color("#88c0d0"),
		Secondary:  lipgloss.Color("#a3be8c"),
		Success:    lipgloss.Color("#a3be8c"),
		Error:      lipgloss.Color("#bf616a"),
		Warning:    lipgloss.Color("#ebcb8b"),
		Muted:      lipgloss.Color("#4c566a"),
		Text:       lipgloss.Color("#eceff4"),
		Background: lipgloss.Color("#2e3440"),
	}

	t.LogoStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	t.HeaderStyle = lipgloss.NewStyle().Foreground(t.Background).Background(t.Primary).Padding(0, 1).Bold(true)
	t.TitleStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	t.SubtitleStyle = lipgloss.NewStyle().Foreground(t.Muted).Italic(true)
	t.HelpStyle = lipgloss.NewStyle().Foreground(t.Muted)
	t.SelectedStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
	t.UnselectedStyle = lipgloss.NewStyle().Foreground(t.Muted)
	t.CursorStyle = lipgloss.NewStyle().Foreground(t.Text).Bold(true)

	return t
}

func (t *Theme) GetLogo() string {
	return t.LogoStyle.Render(LogoASCII)
}
