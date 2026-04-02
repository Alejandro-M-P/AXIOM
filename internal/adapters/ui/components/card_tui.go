// Package components provides reusable UI components for the TUI.
package components

import (
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/theme"
	"github.com/Alejandro-M-P/AXIOM/internal/i18n"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CardTUI is a fullscreen component for displaying information cards.
type CardTUI struct {
	title    string
	subtitle string
	fields   []CardField
	items    []string
	footer   string
	width    int
	height   int
}

// CardField represents a label-value pair in a card.
type CardField struct {
	Label string
	Value string
}

// NewCardTUI creates a new fullscreen card component.
func NewCardTUI(title, subtitle string, fields []CardField, items []string, footer string) *CardTUI {
	return &CardTUI{
		title:    title,
		subtitle: subtitle,
		fields:   fields,
		items:    items,
		footer:   footer,
	}
}

// Init implements tea.Model.
func (c *CardTUI) Init() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{}
	}
}

// Update implements tea.Model.
func (c *CardTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyEsc:
			return c, tea.Quit
		}
	}
	return c, nil
}

// View implements tea.Model.
func (c *CardTUI) View() string {
	t := theme.DefaultTheme()

	// Build the card content
	var content strings.Builder

	// Title
	if c.title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(t.Background).
			Background(t.Primary).
			Padding(0, 1).
			Bold(true)
		content.WriteString(titleStyle.Render(c.title))
		content.WriteString("\n\n")
	}

	// Subtitle
	if c.subtitle != "" {
		subtitleStyle := lipgloss.NewStyle().
			Foreground(t.Muted).
			Italic(true)
		content.WriteString(subtitleStyle.Render(c.subtitle))
		content.WriteString("\n\n")
	}

	// Fields
	if len(c.fields) > 0 {
		for _, f := range c.fields {
			labelStyle := lipgloss.NewStyle().
				Foreground(t.Primary).
				Bold(true).
				Width(15)

			valueStyle := lipgloss.NewStyle().
				Foreground(t.Text)

			line := labelStyle.Render(f.Label) + " " + valueStyle.Render(f.Value)
			content.WriteString(line)
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Items
	if len(c.items) > 0 {
		for _, item := range c.items {
			itemStyle := lipgloss.NewStyle().
				Foreground(t.Secondary)
			content.WriteString(itemStyle.Render("• " + item))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Footer with escape button
	footerStyle := lipgloss.NewStyle().
		Foreground(t.Muted).
		Italic(true)
	content.WriteString(footerStyle.Render(i18n.GetWizardText("components", "card_escape")))

	// Use CenteredContainer for fullscreen
	centered := NewCenteredContainer(c.width, c.height)

	// Card style
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(2, 4)

	return centered.Render(cardStyle.Render(content.String()))
}

// RunCardTUI runs a fullscreen card and returns when the user exits.
func RunCardTUI(title, subtitle string, fields []CardField, items []string, footer string) error {
	model := NewCardTUI(title, subtitle, fields, items, footer)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
