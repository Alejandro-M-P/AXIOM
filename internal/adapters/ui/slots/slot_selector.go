// Package slots provides Bubbletea TUI components for slot selection.
package slots

import (
	"fmt"
	"os"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/components"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ItemGroup represents a group of slot items by subcategory.
type ItemGroup struct {
	title string
	items []SlotItemDisplay
}

// SlotItemDisplay represents a slot item shown in the selector.
type SlotItemDisplay struct {
	ID          string
	Name        string
	Description string
}

// SelectedSlots is the result message when user confirms selection.
type SelectedSlots struct {
	IDs []string
}

// SlotSelectorModel is the Bubbletea model for slot item selection.
// Embeds BaseModel for window size handling.
type SlotSelectorModel struct {
	ui.BaseModel   // Embed BaseModel for Width/Height from WindowSizeMsg
	groups         []ItemGroup
	selected       map[string]bool
	cursor         int
	done           bool
	canceled       bool
	firstItemIndex int // Index of first item in current group (for cursor mapping)
}

// NewSlotSelectorModel creates a new slot selector model.
func NewSlotSelectorModel(groups []ItemGroup) *SlotSelectorModel {
	selected := make(map[string]bool)
	for _, group := range groups {
		for _, item := range group.items {
			selected[item.ID] = false
		}
	}

	return &SlotSelectorModel{
		groups:   groups,
		selected: selected,
		cursor:   0,
		done:     false,
		canceled: false,
	}
}

// Init initializes the model.
// Calls BaseModel.Init() to request initial window size.
func (m *SlotSelectorModel) Init() tea.Cmd {
	return m.BaseModel.Init()
}

// Update handles user input.
func (m *SlotSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.BaseModel.Update(msg)
	}
	return m, nil
}

// handleKey processes keyboard input.
func (m *SlotSelectorModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalItems := countAllItems(m.groups)

	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.canceled = true
		return m, tea.Quit

	case tea.KeyEnter:
		m.done = true
		return m, tea.Quit

	case tea.KeySpace:
		// Toggle selection at cursor position
		itemID := getItemIDAtCursor(m.groups, m.cursor)
		if itemID != "" {
			m.selected[itemID] = !m.selected[itemID]
		}
		return m, nil

	case tea.KeyDown, tea.KeyTab:
		if m.cursor < totalItems-1 {
			m.cursor++
		}
		return m, nil

	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case tea.KeyEnd:
		m.cursor = totalItems - 1
		return m, nil

	case tea.KeyHome:
		m.cursor = 0
		return m, nil
	}
	return m, nil
}

// View renders the slot selector using CenteredContainer for fullscreen TUI.
func (m *SlotSelectorModel) View() string {
	var builder strings.Builder
	t := theme.DefaultTheme()

	// Header
	header := theme.NewHeader(t, "Slot Selection", "", "↑/↓: Navigate | Space: Toggle | Enter: Confirm | Esc: Cancel")
	builder.WriteString(header.View())
	builder.WriteString("\n")

	// Render groups
	cursor := 0
	for _, group := range m.groups {
		// Group header
		groupHeaderStyle := lipgloss.NewStyle().
			Foreground(t.Success).
			Bold(true).
			Padding(1, 0, 0, 0)
		builder.WriteString(groupHeaderStyle.Render("━━━ " + group.title + " ━━━\n"))

		// Items in group
		for _, item := range group.items {
			checked := "[ ]"
			checkboxStyle := lipgloss.NewStyle().Foreground(t.Muted)

			if m.selected[item.ID] {
				checked = "[x]"
				checkboxStyle = lipgloss.NewStyle().Foreground(t.Primary)
			}

			cursorPrefix := "  "
			cursorStyle := lipgloss.NewStyle()
			if cursor == m.cursor {
				cursorPrefix = "❯ "
				cursorStyle = lipgloss.NewStyle().Foreground(t.Text)
			} else {
				cursorStyle = lipgloss.NewStyle().Foreground(t.Muted)
			}

			// Item name with cursor
			itemName := fmt.Sprintf("%s%s %s", cursorPrefix, checkboxStyle.Render(checked), cursorStyle.Render(item.Name))

			// Description (dimmed, on same line)
			descStyle := lipgloss.NewStyle().
				Foreground(t.Muted).
				Italic(true)
			builder.WriteString(itemName + " - " + descStyle.Render(item.Description) + "\n")

			cursor++
		}
		builder.WriteString("\n")
	}

	// Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(t.Muted).
		Padding(1, 0, 0, 0)

	// Dynamic footer width based on terminal width
	footerWidth := m.Width
	if footerWidth < 60 {
		footerWidth = 60
	}
	footerLine := strings.Repeat("─", footerWidth-4)
	builder.WriteString(footerStyle.Render(footerLine + "\n"))

	// Selection summary
	selectedCount := countSelected(m.selected, m.groups)
	totalCount := countAllItems(m.groups)
	summaryStyle := lipgloss.NewStyle().Foreground(t.Success)
	builder.WriteString(summaryStyle.Render(fmt.Sprintf("Selected: %d/%d", selectedCount, totalCount)) + "    ")

	builder.WriteString(footerStyle.Render("Space: toggle  │  Enter: confirm  │  Esc: cancel\n"))

	// Use CenteredContainer for fullscreen centering
	centered := components.NewCenteredContainer(m.Width, m.Height)
	return centered.Render(builder.String())
}

// countAllItems returns the total number of items across all groups.
func countAllItems(groups []ItemGroup) int {
	count := 0
	for _, group := range groups {
		count += len(group.items)
	}
	return count
}

// countSelected returns the number of selected items.
func countSelected(selected map[string]bool, groups []ItemGroup) int {
	count := 0
	for _, group := range groups {
		for _, item := range group.items {
			if selected[item.ID] {
				count++
			}
		}
	}
	return count
}

// getItemIDAtCursor returns the item ID at the given cursor position.
func getItemIDAtCursor(groups []ItemGroup, cursor int) string {
	curr := 0
	for _, group := range groups {
		for _, item := range group.items {
			if curr == cursor {
				return item.ID
			}
			curr++
		}
	}
	return ""
}

// GetSelectedIDs returns the IDs of all selected items.
func (m *SlotSelectorModel) GetSelectedIDs() []string {
	var ids []string
	for _, group := range m.groups {
		for _, item := range group.items {
			if m.selected[item.ID] {
				ids = append(ids, item.ID)
			}
		}
	}
	return ids
}

// IsDone returns true if the user has confirmed their selection.
func (m *SlotSelectorModel) IsDone() bool {
	return m.done
}

// IsCanceled returns true if the user cancelled the selection.
func (m *SlotSelectorModel) IsCanceled() bool {
	return m.canceled
}

// RunSlotSelector runs the interactive slot selector TUI.
// Returns the selected item IDs, whether the user confirmed (true) or cancelled (false), and any error.
func RunSlotSelector(groups []ItemGroup) ([]string, bool, error) {
	model := NewSlotSelectorModel(groups)
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	)

	finalModel, err := p.Run()

	// Ensure terminal is cleaned up properly even if there's an error
	// This prevents the "corrupted text" issue when running multiple times
	cleanupTerminal()

	if err != nil {
		return nil, false, fmt.Errorf("failed to run slot selector: %w", err)
	}

	resultModel, ok := finalModel.(*SlotSelectorModel)
	if !ok {
		return nil, false, fmt.Errorf("unexpected model type: %T", finalModel)
	}

	if resultModel.canceled {
		return nil, false, nil
	}

	return resultModel.GetSelectedIDs(), true, nil
}

// SlotUIAdapter provides a simple interface for running slot selection.
type SlotUIAdapter struct {
	groups []ItemGroup
}

// NewSlotUIAdapter creates a new SlotUIAdapter.
func NewSlotUIAdapter(groups []ItemGroup) *SlotUIAdapter {
	return &SlotUIAdapter{groups: groups}
}

// Run executes the slot selection UI and returns selected IDs, confirmed status, and error.
func (a *SlotUIAdapter) Run() ([]string, bool, error) {
	return RunSlotSelector(a.groups)
}

// Builder helps construct ItemGroup lists.
type Builder struct {
	groups map[string][]SlotItemDisplay
}

// NewBuilder creates a new ItemGroup builder.
func NewBuilder() *Builder {
	return &Builder{
		groups: make(map[string][]SlotItemDisplay),
	}
}

// AddItem adds a slot item to a subcategory group.
func (b *Builder) AddItem(subcategory, id, name, description string) *Builder {
	b.groups[subcategory] = append(b.groups[subcategory], SlotItemDisplay{
		ID:          id,
		Name:        name,
		Description: description,
	})
	return b
}

// Build constructs the final list of ItemGroup.
func (b *Builder) Build() []ItemGroup {
	// Define the order of subcategories
	order := []string{"ia", "languages", "tools", "data"}

	var result []ItemGroup
	for _, subcategory := range order {
		if items, ok := b.groups[subcategory]; ok {
			title := getSubcategoryTitle(subcategory)
			result = append(result, ItemGroup{title: title, items: items})
		}
	}

	// Add any remaining subcategories not in the predefined order
	for subcategory, items := range b.groups {
		if !contains(order, subcategory) {
			title := getSubcategoryTitle(subcategory)
			result = append(result, ItemGroup{title: title, items: items})
		}
	}

	return result
}

// contains checks if a string slice contains a value.
func contains(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// getSubcategoryTitle returns a human-readable title for a subcategory.
func getSubcategoryTitle(subcategory string) string {
	switch subcategory {
	case "ia":
		return "AI / LLM Models"
	case "languages":
		return "Programming Languages"
	case "tools":
		return "Developer Tools"
	case "data":
		return "Data Stores"
	default:
		return strings.Title(subcategory)
	}
}
