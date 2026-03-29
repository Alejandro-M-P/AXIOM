// Package slots provides Bubbletea TUI components for slot selection.
package slots

import (
	"fmt"
	"strings"

	"axiom/internal/adapters/ui/theme"
	"axiom/internal/ports"
	"axiom/internal/slots"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WizardPhase represents the current phase of the wizard.
type WizardPhase int

const (
	PhaseSlotSelect     WizardPhase = iota // User picks DEV/DATA/SANDBOX
	PhaseCategoryWizard                    // Step through categories
	PhaseSummary                           // Final summary
)

// WizardModel is the Bubbletea model for wizard-style slot selection.
type WizardModel struct {
	presenter     ports.IPresenter
	phase         WizardPhase
	selectedSlot  string // "dev", "data", "sandbox"
	categoryIndex int    // Which subcategory we're on
	categories    []string
	accumulated   map[string]bool // Selected item IDs across all steps
	itemGroups    []ItemGroup     // Groups of items for current slot
	slotCursor    int             // Cursor for slot selection
	itemCursor    int             // Cursor for item selection within category
	width         int
	height        int
	done          bool
	canceled      bool
	allItems      []slots.SlotItem // All items passed to the wizard (stored for filtering)
}

// NewWizardModel creates a new wizard model.
func NewWizardModel(pres ports.IPresenter) *WizardModel {
	return &WizardModel{
		presenter:     pres,
		phase:         PhaseSlotSelect,
		selectedSlot:  "",
		categoryIndex: 0,
		categories:    []string{"ia", "languages", "tools", "data"},
		accumulated:   make(map[string]bool),
		itemGroups:    nil,
		slotCursor:    0,
		itemCursor:    0,
		done:          false,
		canceled:      false,
	}
}

// Init initializes the model.
func (m *WizardModel) Init() tea.Cmd {
	return nil
}

// Update handles user input.
func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// handleKey processes keyboard input.
func (m *WizardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.phase {
	case PhaseSlotSelect:
		return m.handleSlotSelectKey(msg)
	case PhaseCategoryWizard:
		return m.handleCategoryWizardKey(msg)
	case PhaseSummary:
		return m.handleSummaryKey(msg)
	}
	return m, nil
}

// handleSlotSelectKey handles keys in the slot selection phase.
func (m *WizardModel) handleSlotSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	slots := []string{"dev", "data", "sandbox"}

	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.canceled = true
		return m, tea.Quit

	case tea.KeyEnter:
		m.selectedSlot = slots[m.slotCursor]
		m.loadItemsForSlot()
		// If no categories (e.g., SANDBOX slot), skip to summary
		if len(m.categories) == 0 {
			m.phase = PhaseSummary
		} else {
			m.phase = PhaseCategoryWizard
			m.itemCursor = 0
		}
		return m, nil

	case tea.KeyDown, tea.KeyTab:
		if m.slotCursor < len(slots)-1 {
			m.slotCursor++
		}
		return m, nil

	case tea.KeyUp:
		if m.slotCursor > 0 {
			m.slotCursor--
		}
		return m, nil
	}
	return m, nil
}

// handleCategoryWizardKey handles keys in the category wizard phase.
func (m *WizardModel) handleCategoryWizardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	currentGroup := m.getCurrentGroup()
	if currentGroup == nil {
		return m, nil
	}
	totalItems := len(currentGroup.items)

	switch msg.Type {
	case tea.KeyEsc:
		m.canceled = true
		return m, tea.Quit

	case tea.KeyEnter:
		if m.categoryIndex < len(m.categories)-1 {
			m.categoryIndex++
			m.itemCursor = 0
		} else {
			m.phase = PhaseSummary
		}
		return m, nil

	case tea.KeySpace:
		itemID := ""
		if currentGroup != nil && m.itemCursor < len(currentGroup.items) {
			itemID = currentGroup.items[m.itemCursor].ID
		}
		if itemID != "" {
			m.accumulated[itemID] = !m.accumulated[itemID]
		}
		return m, nil

	case tea.KeyDown, tea.KeyTab:
		if m.itemCursor < totalItems-1 {
			m.itemCursor++
		}
		return m, nil

	case tea.KeyUp:
		if m.itemCursor > 0 {
			m.itemCursor--
		}
		return m, nil

	case tea.KeyEnd:
		m.itemCursor = totalItems - 1
		return m, nil

	case tea.KeyHome:
		m.itemCursor = 0
		return m, nil
	}
	return m, nil
}

// handleSummaryKey handles keys in the summary phase.
func (m *WizardModel) handleSummaryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.canceled = true
		return m, tea.Quit

	case tea.KeyEnter:
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

// loadItemsForSlot loads the items for the selected slot into itemGroups.
func (m *WizardModel) loadItemsForSlot() {
	// Filter categories based on selected slot
	switch m.selectedSlot {
	case "dev":
		m.categories = []string{"ia", "languages", "tools"}
	case "data":
		m.categories = []string{"data"}
	case "sandbox":
		m.categories = []string{}
	default:
		m.categories = []string{"ia", "languages", "tools", "data"}
	}

	// Filter items by category matching the selected slot
	var filteredItems []slots.SlotItem
	for _, item := range m.allItems {
		if string(item.Category) == m.selectedSlot {
			filteredItems = append(filteredItems, item)
		}
	}

	// Group filtered items by subcategory
	groupMap := make(map[string][]SlotItemDisplay)
	for _, item := range filteredItems {
		groupMap[item.SubCategory] = append(groupMap[item.SubCategory], SlotItemDisplay{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
		})
	}

	// Build ordered groups based on filtered categories
	var groups []ItemGroup
	for _, subcat := range m.categories {
		if items, ok := groupMap[subcat]; ok {
			groups = append(groups, ItemGroup{
				title: m.getSubcategoryTitle(subcat),
				items: items,
			})
		}
	}

	// Add remaining categories not in predefined order (if any)
	for subcat, items := range groupMap {
		if !contains(m.categories, subcat) {
			groups = append(groups, ItemGroup{
				title: m.getSubcategoryTitle(subcat),
				items: items,
			})
		}
	}

	m.itemGroups = groups
}

// getCurrentGroup returns the current ItemGroup based on categoryIndex.
func (m *WizardModel) getCurrentGroup() *ItemGroup {
	if m.categoryIndex >= 0 && m.categoryIndex < len(m.itemGroups) {
		return &m.itemGroups[m.categoryIndex]
	}
	return nil
}

// View renders the wizard based on the current phase.
func (m *WizardModel) View() string {
	switch m.phase {
	case PhaseSlotSelect:
		return m.viewSlotSelect()
	case PhaseCategoryWizard:
		return m.viewCategoryWizard()
	case PhaseSummary:
		return m.viewSummary()
	}
	return ""
}

// viewSlotSelect renders the slot selection view.
func (m *WizardModel) viewSlotSelect() string {
	// Use theme
	t := theme.DefaultTheme()

	// Create header for slot selection phase
	header := theme.NewHeader(t, "Select Slot Type", "", "↑/↓: Navigate | Enter: Confirm | Esc: Cancel")

	// Define colors from theme
	accentColor := t.Primary
	dimColor := t.Muted
	selectedColor := t.Primary

	// Box style - fixed width container
	boxWidth := 78
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1, 2).
		Width(boxWidth)

	// Content width inside the box (accounting for border and padding)
	contentWidth := boxWidth - 6 // 2 for border + 4 for padding

	titleStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(t.Text)

	helpStyle := lipgloss.NewStyle().
		Foreground(dimColor)

	slotOptions := []struct {
		id             string
		nameKey        string
		descriptionKey string
	}{
		{"dev", "slot_wizard.dev", "slot_wizard.dev_desc"},
		{"data", "slot_wizard.data", "slot_wizard.data_desc"},
		{"sandbox", "slot_wizard.sandbox", "slot_wizard.sandbox_desc"},
	}

	var content strings.Builder

	// Add header at the start
	content.WriteString(header.View())
	content.WriteString("\n")

	// Title
	content.WriteString(titleStyle.Render(m.presenter.GetText("slot_wizard.title")))
	content.WriteString("\n")

	// Description
	descText := descStyle.Render(m.presenter.GetText("slot_wizard.slot_select_desc"))
	content.WriteString(descText)
	content.WriteString("\n\n")

	// Slot options - each rendered as a complete line
	for i, opt := range slotOptions {
		nameText := m.presenter.GetText(opt.nameKey)
		descText := m.presenter.GetText(opt.descriptionKey)

		var lineStyle lipgloss.Style
		var prefix string

		if i == m.slotCursor {
			prefix = "❯ "
			lineStyle = lipgloss.NewStyle().
				Foreground(selectedColor).
				Bold(true).
				Width(contentWidth)
		} else {
			prefix = "  "
			lineStyle = lipgloss.NewStyle().
				Foreground(dimColor).
				Width(contentWidth)
		}

		// Build the line with cursor prefix
		lineText := prefix + nameText + " - " + descText
		content.WriteString(lineStyle.Render(lineText))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(helpStyle.Render(m.presenter.GetText("slot_wizard.help_slot_select")))

	return boxStyle.Render(content.String())
}

// viewCategoryWizard renders the category wizard step view.
func (m *WizardModel) viewCategoryWizard() string {
	// Use theme
	t := theme.DefaultTheme()

	// Get current category for header subtitle
	currentGroup := m.getCurrentGroup()
	categoryName := ""
	if currentGroup != nil {
		categoryName = currentGroup.title
	}

	// Create header for category wizard phase
	header := theme.NewHeader(t, "Select Items", categoryName, "↑/↓: Navigate | Space: Toggle | Enter: Confirm | Esc: Cancel")

	// Define colors from theme
	accentColor := t.Primary
	dimColor := t.Muted
	selectedColor := t.Primary
	mutedColor := t.Muted
	textColor := t.Text
	greenColor := t.Success

	// Box style - fixed width container
	boxWidth := 78
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1, 2).
		Width(boxWidth)

	// Content width inside the box (accounting for border and padding)
	contentWidth := boxWidth - 6 // 2 for border + 4 for padding

	titleStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)

	headerStyle := lipgloss.NewStyle().
		Foreground(greenColor).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(dimColor)

	var content strings.Builder

	// Add header at the start
	content.WriteString(header.View())
	content.WriteString("\n")

	// Step indicator as title
	stepIndicator := m.presenter.GetText("slot_wizard.step_indicator", m.categoryIndex+1, len(m.categories))
	content.WriteString(titleStyle.Render(stepIndicator))
	content.WriteString("\n\n")

	currentGroup = m.getCurrentGroup()
	if currentGroup != nil {
		// Group header
		content.WriteString(headerStyle.Render(currentGroup.title))
		content.WriteString("\n\n")

		// Items list
		for i, item := range currentGroup.items {
			checked := "[ ]"
			checkboxFg := mutedColor

			if m.accumulated[item.ID] {
				checked = "[x]"
				checkboxFg = selectedColor
			}

			var lineStyle lipgloss.Style
			var lineText string

			if i == m.itemCursor {
				// Selected item: cursor prefix INSIDE the styled render
				checkboxStyle := lipgloss.NewStyle().Foreground(checkboxFg)
				itemStyle := lipgloss.NewStyle().Foreground(textColor)
				lineStyle = lipgloss.NewStyle().
					Foreground(selectedColor).
					Bold(true).
					Width(contentWidth)
				lineText = "❯ " + checkboxStyle.Render(checked) + " " + itemStyle.Render(item.Name) + " - " + item.Description
			} else {
				// Unselected item
				checkboxStyle := lipgloss.NewStyle().Foreground(checkboxFg)
				itemStyle := lipgloss.NewStyle().Foreground(mutedColor)
				lineStyle = lipgloss.NewStyle().
					Foreground(mutedColor).
					Width(contentWidth)
				lineText = "  " + checkboxStyle.Render(checked) + " " + itemStyle.Render(item.Name) + " - " + item.Description
			}

			content.WriteString(lineStyle.Render(lineText))
			content.WriteString("\n")
		}

		content.WriteString("\n")

		// Selected count
		selectedCount := m.countCurrentSelections()
		totalCount := len(currentGroup.items)
		countStyle := lipgloss.NewStyle().Foreground(greenColor)
		countText := m.presenter.GetText("slot_wizard.selected_count", selectedCount, totalCount)
		content.WriteString(countStyle.Render(countText))
		content.WriteString("\n")
	}

	// Help text
	content.WriteString(helpStyle.Render(m.presenter.GetText("slot_wizard.help_category")))

	return boxStyle.Render(content.String())
}

// viewSummary renders the final summary view.
func (m *WizardModel) viewSummary() string {
	// Use theme
	t := theme.DefaultTheme()

	// Create header for summary phase
	header := theme.NewHeader(t, "Summary", "", "↑/↓: Navigate | Space: Toggle | Enter: Confirm | Esc: Cancel")

	// Define colors from theme
	accentColor := t.Primary
	dimColor := t.Muted
	textColor := t.Text
	greenColor := t.Success

	// Box style - fixed width container
	boxWidth := 78
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(1, 2).
		Width(boxWidth)

	// Content width inside the box (accounting for border and padding)
	contentWidth := boxWidth - 6 // 2 for border + 4 for padding

	titleStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(t.Text)

	headerStyle := lipgloss.NewStyle().
		Foreground(greenColor).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Foreground(dimColor)

	var content strings.Builder

	// Add header at the start
	content.WriteString(header.View())
	content.WriteString("\n")

	// Title
	content.WriteString(titleStyle.Render(m.presenter.GetText("slot_wizard.summary")))
	content.WriteString("\n")

	// Description
	content.WriteString(descStyle.Render(m.presenter.GetText("slot_wizard.summary_desc")))
	content.WriteString("\n\n")

	// Slot label and name - properly styled
	slotLabel := m.presenter.GetText("slot_wizard.slot_label")
	slotLabelStyle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true)
	slotValueStyle := lipgloss.NewStyle().
		Foreground(textColor)
	slotLine := slotLabelStyle.Render(slotLabel) + slotValueStyle.Render(": "+m.selectedSlot)
	content.WriteString(slotLine)
	content.WriteString("\n\n")

	// Group selections by category
	for _, group := range m.itemGroups {
		hasSelected := false
		for _, item := range group.items {
			if m.accumulated[item.ID] {
				hasSelected = true
				break
			}
		}

		if !hasSelected {
			continue
		}

		// Group header
		content.WriteString(headerStyle.Render(group.title))
		content.WriteString("\n")

		// Selected items - cursor/check prefix INSIDE styled render
		for _, item := range group.items {
			if m.accumulated[item.ID] {
				lineStyle := lipgloss.NewStyle().
					Foreground(textColor).
					Width(contentWidth)
				lineText := "  ✓ " + item.Name
				content.WriteString(lineStyle.Render(lineText))
				content.WriteString("\n")
			}
		}
		content.WriteString("\n")
	}

	// Total selected count
	totalSelected := m.countTotalSelections()
	summaryStyle := lipgloss.NewStyle().Foreground(greenColor)
	content.WriteString(summaryStyle.Render(m.presenter.GetText("slot_wizard.total_selected", totalSelected)))
	content.WriteString("\n")

	// Help text
	content.WriteString(helpStyle.Render(m.presenter.GetText("slot_wizard.help_summary")))

	return boxStyle.Render(content.String())
}

// countCurrentSelections counts selected items in the current category.
func (m *WizardModel) countCurrentSelections() int {
	currentGroup := m.getCurrentGroup()
	if currentGroup == nil {
		return 0
	}
	count := 0
	for _, item := range currentGroup.items {
		if m.accumulated[item.ID] {
			count++
		}
	}
	return count
}

// countTotalSelections counts all accumulated selections.
func (m *WizardModel) countTotalSelections() int {
	count := 0
	for _, selected := range m.accumulated {
		if selected {
			count++
		}
	}
	return count
}

// getSubcategoryTitle returns a human-readable title for a subcategory using i18n.
func (m *WizardModel) getSubcategoryTitle(subcategory string) string {
	switch subcategory {
	case "ia":
		return m.presenter.GetText("slot_wizard.step_ai")
	case "languages":
		return m.presenter.GetText("slot_wizard.step_languages")
	case "tools":
		return m.presenter.GetText("slot_wizard.step_tools")
	case "data":
		return m.presenter.GetText("slot_wizard.step_data")
	default:
		return strings.Title(subcategory)
	}
}

// GetSelectedIDs returns the IDs of all selected items.
func (m *WizardModel) GetSelectedIDs() []string {
	var ids []string
	for id, selected := range m.accumulated {
		if selected {
			ids = append(ids, id)
		}
	}
	return ids
}

// IsDone returns true if the user has confirmed their selection.
func (m *WizardModel) IsDone() bool {
	return m.done
}

// IsCanceled returns true if the user cancelled the selection.
func (m *WizardModel) IsCanceled() bool {
	return m.canceled
}

// RunWizard runs the interactive wizard-style slot selector TUI.
// Returns the selected item IDs, whether the user confirmed (true) or cancelled (false), and any error.
func RunWizard(items []slots.SlotItem, pres ports.IPresenter) ([]string, bool, error) {
	model := NewWizardModel(pres)

	// Store all items for later filtering when slot is selected
	model.allItems = items

	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return nil, false, fmt.Errorf("failed to run wizard selector: %w", err)
	}

	resultModel, ok := finalModel.(*WizardModel)
	if !ok {
		return nil, false, fmt.Errorf("unexpected model type: %T", finalModel)
	}

	if resultModel.canceled {
		return nil, false, nil
	}

	return resultModel.GetSelectedIDs(), true, nil
}
