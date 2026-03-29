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

// bunkerSelectorModel es el modelo para seleccionar un bunker de forma interactiva.
type bunkerSelectorModel struct {
	bunkers  []string
	statuses map[string]string // nombre -> estado
	cursor   int
	done     bool
	canceled bool
	width    int
	height   int
	title    string
	subtitle string
}

// newBunkerSelectorModel crea un nuevo selector de bunkers.
func newBunkerSelectorModel(bunkers []string, statuses map[string]string, title, subtitle string) bunkerSelectorModel {
	return bunkerSelectorModel{
		bunkers:  bunkers,
		statuses: statuses,
		cursor:   0,
		title:    title,
		subtitle: subtitle,
	}
}

// Init inicializa el modelo.
func (m bunkerSelectorModel) Init() tea.Cmd {
	return nil
}

// Update maneja los mensajes.
func (m bunkerSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit

		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit

		case tea.KeyUp, tea.KeyLeft:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown, tea.KeyRight:
			if m.cursor < len(m.bunkers)-1 {
				m.cursor++
			}

		case tea.KeyHome:
			m.cursor = 0

		case tea.KeyEnd:
			m.cursor = len(m.bunkers) - 1
		}
	}

	return m, nil
}

// View renderiza el selector.
func (m bunkerSelectorModel) View() string {
	width := m.width
	height := m.height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	t := theme.DefaultTheme()

	// Calcular dimensiones del contenido
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
	header := theme.NewHeader(t, m.title, m.subtitle, "↑/↓: Navigate | Enter: Select | Esc: Cancel")
	content.WriteString(header.View())
	content.WriteString("\n\n")

	// Lista de bunkers
	for i, bunker := range m.bunkers {
		prefix := "  "
		status := m.statuses[bunker]
		if status == "" {
			status = "stopped"
		}

		var nameStyle lipgloss.Style
		if i == m.cursor {
			prefix = "❯ "
			nameStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
		} else {
			nameStyle = lipgloss.NewStyle().Foreground(t.Text)
		}

		// Estado con color
		var statusStyle lipgloss.Style
		switch status {
		case "running", "Started":
			statusStyle = lipgloss.NewStyle().Foreground(t.Success)
		case "exited", "stopped", "Stopped":
			statusStyle = lipgloss.NewStyle().Foreground(t.Muted)
		default:
			statusStyle = lipgloss.NewStyle().Foreground(t.Muted)
		}

		statusText := statusStyle.Render(fmt.Sprintf("[%s]", status))
		content.WriteString(nameStyle.Render(prefix+bunker) + "  " + statusText + "\n")
	}

	// Footer
	content.WriteString("\n")
	footerStyle := lipgloss.NewStyle().Foreground(t.Muted)
	content.WriteString(footerStyle.Render(strings.Repeat("─", contentWidth-4)))
	content.WriteString("\n")
	content.WriteString(footerStyle.Render(fmt.Sprintf("Selected: %d/%d", m.cursor+1, len(m.bunkers))))

	return styles.GetLogo() + "\n" + windowStyle.Render(content.String())
}

// AskSelectBunker ejecuta el selector interactivo de bunkers.
// Retorna el nombre seleccionado, confirmado, y error.
func (c *ConsoleUI) AskSelectBunker(bunkers []string, statuses map[string]string, title, subtitle string) (string, bool, error) {
	if len(bunkers) == 0 {
		return "", false, fmt.Errorf("no hay búnkeres disponibles")
	}

	model := newBunkerSelectorModel(bunkers, statuses, title, subtitle)
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	)

	finalModel, err := p.Run()

	// Cleanup terminal
	fmt.Print("\033[?1049l")
	fmt.Print("\033[?25h")
	os.Stdout.Sync()

	if err != nil {
		return "", false, fmt.Errorf("failed to run bunker selector: %w", err)
	}

	resultModel, ok := finalModel.(bunkerSelectorModel)
	if !ok {
		return "", false, fmt.Errorf("unexpected model type: %T", finalModel)
	}

	if resultModel.canceled {
		return "", false, nil
	}

	if !resultModel.done {
		return "", false, nil
	}

	return resultModel.bunkers[resultModel.cursor], true, nil
}
