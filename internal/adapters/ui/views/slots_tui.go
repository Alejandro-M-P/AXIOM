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

// slotsTUIModel es el modelo para mostrar slots disponibles en TUI.
type slotsTUIModel struct {
	slots  []slotDisplay
	cursor int
	done   bool
	width  int
	height int
}

// slotDisplay representa un slot para mostrar.
type slotDisplay struct {
	name        string
	description string
	category    string
}

// newSlotsTUIModel crea un nuevo modelo de slots.
func newSlotsTUIModel() slotsTUIModel {
	return slotsTUIModel{
		slots:  []slotDisplay{},
		cursor: 0,
	}
}

// Init inicializa el modelo.
func (m slotsTUIModel) Init() tea.Cmd {
	return nil
}

// Update maneja los mensajes.
func (m slotsTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.slots)-1 {
				m.cursor++
			}

		case tea.KeyHome:
			m.cursor = 0

		case tea.KeyEnd:
			m.cursor = len(m.slots) - 1
		}
	}

	return m, nil
}

// View renderiza los slots.
func (m slotsTUIModel) View() string {
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
	header := theme.NewHeader(t, "Available Slots", "", "↑/↓: Navigate | Esc: Exit")
	content.WriteString(header.View())
	content.WriteString("\n")

	// Título
	titleStyle := lipgloss.NewStyle().
		Foreground(t.Primary).
		Bold(true).
		Width(contentWidth - 4)
	content.WriteString(titleStyle.Render("\nSlots discovered in your project:\n\n"))

	if len(m.slots) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(t.Muted).Italic(true)
		content.WriteString(emptyStyle.Render("  No slots discovered. Run 'axiom init' first.") + "\n")
	}

	// Lista de slots
	for i, slot := range m.slots {
		prefix := "  "
		slotStyle := lipgloss.NewStyle().Foreground(t.Text)
		catStyle := lipgloss.NewStyle().Foreground(t.Success)
		descStyle := lipgloss.NewStyle().Foreground(t.Muted)

		if i == m.cursor {
			prefix = "❯ "
			slotStyle = lipgloss.NewStyle().Foreground(t.Primary).Bold(true)
		}

		line := fmt.Sprintf("%s%s  [%s]", prefix, slotStyle.Render(slot.name), catStyle.Render(slot.category))
		content.WriteString(line + "\n")
		content.WriteString(descStyle.Render("    "+slot.description) + "\n\n")
	}

	// Footer
	footerStyle := lipgloss.NewStyle().Foreground(t.Muted)
	content.WriteString(footerStyle.Render("\n" + strings.Repeat("─", contentWidth-4)))
	content.WriteString("\n")
	content.WriteString(footerStyle.Render(fmt.Sprintf("Total: %d slots | Press Esc to exit", len(m.slots))))

	return styles.GetLogo() + "\n" + windowStyle.Render(content.String())
}

// RunSlotsTUI ejecuta la vista de slots en modo TUI.
func (c *ConsoleUI) RunSlotsTUI(slots []slotDisplay) error {
	model := newSlotsTUIModel()
	model.slots = slots
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
