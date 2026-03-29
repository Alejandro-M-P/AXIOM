package theme

import (
	"strings"
)

// HeaderComponent is a reusable header component for the TUI.
// It displays a logo, title, optional subtitle, and help text.
type HeaderComponent struct {
	theme    *Theme
	title    string
	subtitle string
	help     string
}

// NewHeader creates a new HeaderComponent with the given theme and content.
func NewHeader(theme *Theme, title, subtitle, help string) *HeaderComponent {
	return &HeaderComponent{
		theme:    theme,
		title:    title,
		subtitle: subtitle,
		help:     help,
	}
}

// WithHelp returns a new HeaderComponent with the help text updated.
// Useful for method chaining.
func (h *HeaderComponent) WithHelp(help string) *HeaderComponent {
	h.help = help
	return h
}

// View renders the header as a string.
// It includes the logo, title, subtitle, and help text.
func (h *HeaderComponent) View() string {
	var sb strings.Builder

	// Logo
	sb.WriteString(h.theme.GetLogo())
	sb.WriteString("\n\n")

	// Title
	sb.WriteString(h.theme.HeaderStyle.Render(h.title))
	sb.WriteString("\n")

	// Subtitle (if provided)
	if h.subtitle != "" {
		sb.WriteString(h.theme.SubtitleStyle.Render(h.subtitle))
		sb.WriteString("\n")
	}

	// Spacer
	sb.WriteString("\n")

	// Help text (if provided)
	if h.help != "" {
		sb.WriteString(h.theme.HelpStyle.Render(h.help))
		sb.WriteString("\n")
	}

	return sb.String()
}
