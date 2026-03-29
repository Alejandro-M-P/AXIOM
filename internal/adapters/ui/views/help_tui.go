package ui

import (
	"fmt"
	"os"
	"strings"

	"axiom/internal/adapters/ui/styles"
	"axiom/internal/adapters/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// helpTUIModel es el modelo para mostrar ayuda en TUI.
type helpTUIModel struct {
	commands []commandHelp
	cursor   int
	done     bool
	width    int
	height   int
}

// commandHelp representa un comando en la ayuda.
type commandHelp struct {
	name        string
	description string
	usage       string
}

// newHelpTUIModel crea un nuevo modelo de ayuda.
func newHelpTUIModel() helpTUIModel {
	commands := []commandHelp{
		{"init", "Initialize AXIOM configuration", "axiom init"},
		{"create", "Create a new bunker", "axiom create"},
		{"build", "Build environment with slots", "axiom build"},
		{"list", "List and select bunkers", "axiom list"},
		{"info", "Show bunker information", "axiom info <name>"},
		{"delete", "Delete a bunker", "axiom delete [name]"},
		{"stop", "Stop a running bunker", "axiom stop [name]"},
		{"prune", "Clean up orphaned bunkers", "axiom prune"},
		{"slots", "Discover available slots", "axiom slots"},
		{"help", "Show this help", "axiom help"},
	}
	return helpTUIModel{
		commands: commands,
		cursor:   0,
	}
}

// Init inicializa el modelo.
func (m helpTUIModel) Init() tea.Cmd {
	return nil
}

// Update maneja los mensajes.
func (m helpTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC, tea.KeyEnter:
			m.done = true
			return m, tea.Quit

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.cursor < len(m.commands)-1 {
				m.cursor++
			}

		case tea.KeyHome:
			m.cursor = 0

		case tea.KeyEnd:
			m.cursor = len(m.commands) - 1
		}
	}

	return m, nil
}

// View renderiza la ayuda.
func (m helpTUIModel) View() string {
	width := m.width
	height := m.height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	t := theme.DefaultTheme()

	// Calcular dimensiones
	contentWidth := width * 80 / 100
	if contentWidth > 80 {
		contentWidth = 80
	}
	if contentWidth < 50 {
		contentWidth = 50
	}
	margin := (width - contentWidth) / 2
	if margin < 2 {
		margin = 2
	}

	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(1, 2).
		Margin(1, margin).
		Width(contentWidth)

	var content strings.Builder

	// Header
	header := theme.NewHeader(t, "AXIOM Help", "", "↑/↓: Navigate | Esc: Exit")
	content.WriteString(header.View())
	content.WriteString("\n")

	// Título
	titleStyle := lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Width(contentWidth - 4)
	content.WriteString(titleStyle.Render("\nAvailable Commands:\n"))

	// Lista de comandos
	for i, cmd := range m.commands {
		prefix := "  "
		cmdStyle := lipgloss.NewStyle().Foreground(t.Text)
		descStyle := lipgloss.NewStyle().Foreground(t.Muted)
		usageStyle := lipgloss.NewStyle().Foreground(t.Primary).Italic(true)

		if i == m.cursor {
			prefix = "❯ "
			cmdStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
		}

		line := fmt.Sprintf("%s%s  -  %s", prefix, cmdStyle.Render(cmd.name), descStyle.Render(cmd.description))
		content.WriteString(line + "\n")
		content.WriteString(usageStyle.Render("    "+cmd.usage) + "\n\n")
	}

	// Footer
	footerStyle := lipgloss.NewStyle().Foreground(t.Muted)
	content.WriteString(footerStyle.Render("\n" + strings.Repeat("─", contentWidth-4)))
	content.WriteString("\n")
	content.WriteString(footerStyle.Render("Press Esc to exit"))

	return styles.GetLogo() + "\n" + windowStyle.Render(content.String())
}

// RunHelpTUI ejecuta la ayuda en modo TUI.
func (c *ConsoleUI) RunHelpTUI() error {
	model := newHelpTUIModel()
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	)

	_, err := p.Run()

	// Cleanup terminal
	fmt.Print("\033[?1049l")
	fmt.Print("\033[?25h")
	os.Stdout.Sync()

	return err
}
