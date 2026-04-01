// Package components provides reusable UI components for the TUI.
package components

import (
	"fmt"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfirmDialog is a reusable confirmation dialog component.
// Displays a centered dialog with a message and Yes/No buttons.
type ConfirmDialog struct {
	title     string
	message   string
	yesLabel  string
	noLabel   string
	width     int
	height    int
	cursor    int // 0 = yes, 1 = no
	confirmed bool
	canceled  bool
}

// NewConfirmDialog creates a new confirmation dialog.
func NewConfirmDialog(title, message string) *ConfirmDialog {
	return &ConfirmDialog{
		title:    title,
		message:  message,
		yesLabel: "YES",
		noLabel:  "NO",
		cursor:   0,
	}
}

// WithYesLabel sets the label for the Yes button.
func (d *ConfirmDialog) WithYesLabel(label string) *ConfirmDialog {
	d.yesLabel = label
	return d
}

// WithNoLabel sets the label for the No button.
func (d *ConfirmDialog) WithNoLabel(label string) *ConfirmDialog {
	d.noLabel = label
	return d
}

// SetDimensions sets the dialog dimensions for centering.
func (d *ConfirmDialog) SetDimensions(width, height int) {
	d.width = width
	d.height = height
}

// Init initializes the dialog.
func (d *ConfirmDialog) Init() tea.Cmd {
	return func() tea.Msg {
		return tea.WindowSizeMsg{}
	}
}

// Update handles user input.
func (d *ConfirmDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			d.canceled = true
			return d, tea.Quit

		case tea.KeyEnter:
			d.confirmed = (d.cursor == 0)
			return d, tea.Quit

		case tea.KeyLeft, tea.KeyRight, tea.KeyUp, tea.KeyDown:
			// Toggle cursor between Yes and No
			if d.cursor == 0 {
				d.cursor = 1
			} else {
				d.cursor = 0
			}
		}
	}
	return d, nil
}

// View renders the confirmation dialog using CenteredContainer.
func (d *ConfirmDialog) View() string {
	t := theme.DefaultTheme()

	// Render buttons
	yesButton := lipgloss.NewStyle().
		Foreground(t.Text).
		Background(t.Muted).
		Padding(0, 3).
		MarginRight(2).
		Render(d.yesLabel)

	noButton := lipgloss.NewStyle().
		Foreground(t.Text).
		Background(t.Muted).
		Padding(0, 3).
		Render(d.noLabel)

	if d.cursor == 0 {
		yesButton = lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.Success).
			Padding(0, 3).
			MarginRight(2).
			Bold(true).
			Render(d.yesLabel)
	} else {
		noButton = lipgloss.NewStyle().
			Foreground(t.Text).
			Background(t.Success).
			Padding(0, 3).
			Bold(true).
			Render(d.noLabel)
	}

	// Build the dialog content
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(t.Background).
		Background(t.Primary).
		Padding(0, 1).
		Bold(true)
	content.WriteString(titleStyle.Render(d.title))
	content.WriteString("\n\n")

	// Message
	messageStyle := lipgloss.NewStyle().
		Foreground(t.Text)
	content.WriteString(messageStyle.Render(d.message))
	content.WriteString("\n\n")

	// Buttons
	content.WriteString(yesButton)
	content.WriteString(noButton)
	content.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(t.Muted)
	content.WriteString(helpStyle.Render("[Enter] Confirm  |  [Esc] Cancel"))

	// Use CenteredContainer for fullscreen centering
	centered := NewCenteredContainer(d.width, d.height)

	// Dialog box style
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Primary).
		Padding(2, 4)

	return centered.Render(dialogStyle.Render(content.String()))
}

// IsConfirmed returns true if user selected Yes.
func (d *ConfirmDialog) IsConfirmed() bool {
	return d.confirmed
}

// IsCanceled returns true if user pressed Esc or Ctrl+C.
func (d *ConfirmDialog) IsCanceled() bool {
	return d.canceled
}

// RunConfirmDialog runs a confirmation dialog and returns the result.
func RunConfirmDialog(title, message string) (bool, error) {
	model := NewConfirmDialog(title, message)
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("errors.ui.confirm_dialog_failed: %w", err)
	}

	resultModel, ok := finalModel.(*ConfirmDialog)
	if !ok {
		return false, fmt.Errorf("errors.ui.unexpected_model")
	}

	if resultModel.canceled {
		return false, nil
	}

	return resultModel.confirmed, nil
}
