// Package slots provides Bubbletea TUI components for slot selection.
package slots

import (
	"fmt"
	"os"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/theme"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
	"github.com/Alejandro-M-P/AXIOM/internal/slots"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WizardPhase represents the current phase of the wizard.
type WizardPhase int

const (
	PhaseSlotSelect WizardPhase = iota // User picks DEV/DATA/SANDBOX
	PhaseItemWizard                    // Step through items one category at a time
	PhaseSummary                       // Final summary
)

// ItemWizardPhase represents the current step within the item wizard.
type ItemWizardPhase int

const (
	ItemPhaseNone ItemWizardPhase = iota
	// For DEV (paso a paso)
	ItemPhaseAI
	ItemPhaseLanguages
	ItemPhaseTools
	// For DATA (todas juntas)
	ItemPhaseDataAll // Muestra todas las DB en una pantalla
)

// WizardModel is the Bubbletea model for wizard-style slot selection.
type WizardModel struct {
	presenter      ports.IPresenter
	phase          WizardPhase
	selectedSlot   string            // "dev", "data", "sandbox"
	itemPhase      ItemWizardPhase   // Current step within PhaseItemWizard
	itemPhaseOrder []ItemWizardPhase // Ordered list of phases for current slot
	accumulated    map[string]bool   // Selected item IDs across all steps
	currentItems   []SlotItemDisplay // Items for the current step
	slotCursor     int               // Cursor for slot selection
	itemCursor     int               // Cursor for item selection within category
	width          int
	height         int
	done           bool
	canceled       bool
	allItems       []slots.SlotItem // All items passed to the wizard (stored for filtering)
}

// NewWizardModel creates a new wizard model.
func NewWizardModel(pres ports.IPresenter) *WizardModel {
	return &WizardModel{
		presenter:      pres,
		phase:          PhaseSlotSelect,
		selectedSlot:   "",
		itemPhase:      ItemPhaseNone,
		itemPhaseOrder: nil,
		accumulated:    make(map[string]bool),
		currentItems:   nil,
		slotCursor:     0,
		itemCursor:     0,
		done:           false,
		canceled:       false,
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
	case PhaseItemWizard:
		return m.handleItemWizardKey(msg)
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
		m.setupItemWizard()
		// If no phases (e.g., SANDBOX slot), skip to summary
		if len(m.itemPhaseOrder) == 0 {
			m.phase = PhaseSummary
		} else {
			m.phase = PhaseItemWizard
			m.itemPhase = m.itemPhaseOrder[0]
			m.loadCurrentPhaseItems()
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

// handleItemWizardKey handles keys in the item wizard phase.
func (m *WizardModel) handleItemWizardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalItems := len(m.currentItems)

	switch msg.Type {
	case tea.KeyEsc:
		// Go back to previous step or slot selection
		currentIndex := m.getCurrentPhaseIndex()
		if currentIndex > 0 {
			// Go back to previous phase
			m.itemPhase = m.itemPhaseOrder[currentIndex-1]
			m.loadCurrentPhaseItems()
			m.itemCursor = 0
		} else {
			// Go back to slot selection
			m.phase = PhaseSlotSelect
			m.itemPhase = ItemPhaseNone
			m.itemPhaseOrder = nil
			m.currentItems = nil
		}
		return m, nil

	case tea.KeyEnter:
		currentIndex := m.getCurrentPhaseIndex()
		if currentIndex < len(m.itemPhaseOrder)-1 {
			// Go to next phase
			m.itemPhase = m.itemPhaseOrder[currentIndex+1]
			m.loadCurrentPhaseItems()
			m.itemCursor = 0
		} else {
			// Go to summary
			m.phase = PhaseSummary
		}
		return m, nil

	case tea.KeySpace:
		if m.itemCursor < totalItems {
			itemID := m.currentItems[m.itemCursor].ID
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
		if totalItems > 0 {
			m.itemCursor = totalItems - 1
		}
		return m, nil

	case tea.KeyHome:
		m.itemCursor = 0
		return m, nil
	}
	return m, nil
}

// getCurrentPhaseIndex returns the index of the current phase in itemPhaseOrder.
func (m *WizardModel) getCurrentPhaseIndex() int {
	for i, phase := range m.itemPhaseOrder {
		if phase == m.itemPhase {
			return i
		}
	}
	return -1
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

// setupItemWizard configures the item wizard phases based on selected slot.
func (m *WizardModel) setupItemWizard() {
	switch m.selectedSlot {
	case "dev":
		m.itemPhaseOrder = []ItemWizardPhase{
			ItemPhaseAI,
			ItemPhaseLanguages,
			ItemPhaseTools,
		}
	case "data":
		// Solo una fase con todas las DB
		m.itemPhaseOrder = []ItemWizardPhase{
			ItemPhaseDataAll,
		}
	case "sandbox":
		m.itemPhaseOrder = []ItemWizardPhase{}
	}
}

// loadCurrentPhaseItems loads items for the current item phase.
func (m *WizardModel) loadCurrentPhaseItems() {
	m.currentItems = nil

	for _, item := range m.allItems {
		if !m.matchesCurrentPhase(item) {
			continue
		}

		// Get name and description from i18n
		name := m.presenter.GetText("slots." + item.ID + ".name")
		description := m.presenter.GetText("slots." + item.ID + ".description")

		m.currentItems = append(m.currentItems, SlotItemDisplay{
			ID:          item.ID,
			Name:        name,
			Description: description,
		})
	}
}

// matchesCurrentPhase checks if an item matches the current wizard phase.
// Base tools (IsBaseTool=true) are excluded from the wizard.
func (m *WizardModel) matchesCurrentPhase(item slots.SlotItem) bool {
	// Skip base tools - they are installed automatically, not selected by user
	if item.IsBaseTool {
		return false
	}

	// First check if item belongs to the selected slot
	if string(item.Category) != m.selectedSlot {
		return false
	}

	switch m.itemPhase {
	case ItemPhaseAI:
		return item.SubCategory == "ia"
	case ItemPhaseLanguages:
		return item.SubCategory == "languages"
	case ItemPhaseTools:
		return item.SubCategory == "tools"
	case ItemPhaseDataAll:
		// Mostrar TODOS los items de DATA (postgres, mysql, mongodb, redis, sqlite)
		return item.Category == "data"
	}

	return false
}

// View renders the wizard based on the current phase.
func (m *WizardModel) View() string {
	switch m.phase {
	case PhaseSlotSelect:
		return m.viewSlotSelect()
	case PhaseItemWizard:
		return m.viewItemWizard()
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

// viewItemWizard renders the item wizard step view (one category at a time).
func (m *WizardModel) viewItemWizard() string {
	// Use theme
	t := theme.DefaultTheme()

	// Get current phase title and step info
	phaseTitle := m.getPhaseTitle()
	currentStep := m.getCurrentPhaseIndex() + 1
	totalSteps := len(m.itemPhaseOrder)

	// Create header for item wizard phase
	header := theme.NewHeader(t, "Select Items", phaseTitle, "Space: Toggle | Enter: Next | Esc: Back")

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
	stepText := m.presenter.GetText("wizard.steps.step_format", currentStep, totalSteps)
	content.WriteString(titleStyle.Render(stepText))
	content.WriteString("\n\n")

	// Category header
	content.WriteString(headerStyle.Render(phaseTitle))
	content.WriteString("\n\n")

	// Items list
	for i, item := range m.currentItems {
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

	if len(m.currentItems) > 0 {
		content.WriteString("\n")

		// Selected count
		selectedCount := m.countCurrentSelections()
		totalCount := len(m.currentItems)
		countStyle := lipgloss.NewStyle().Foreground(greenColor)
		countText := m.presenter.GetText("wizard.steps.selected_format", selectedCount, totalCount)
		content.WriteString(countStyle.Render(countText))
		content.WriteString("\n")
	}

	// Help text
	content.WriteString(helpStyle.Render(m.presenter.GetText("slot_wizard.help_category")))

	return boxStyle.Render(content.String())
}

// getPhaseTitle returns the display title for the current item phase.
func (m *WizardModel) getPhaseTitle() string {
	switch m.itemPhase {
	case ItemPhaseAI:
		return m.presenter.GetText("slot_wizard.step_ai")
	case ItemPhaseLanguages:
		return m.presenter.GetText("slot_wizard.step_languages")
	case ItemPhaseTools:
		return m.presenter.GetText("slot_wizard.step_tools")
	case ItemPhaseDataAll:
		return m.presenter.GetText("slot_wizard.data_all_title")
	}
	return ""
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

	// Show all selected items
	if m.countTotalSelections() > 0 {
		content.WriteString(headerStyle.Render(m.presenter.GetText("slot_wizard.selected_items")))
		content.WriteString("\n")

		// Selected items
		for _, item := range m.allItems {
			if m.accumulated[item.ID] {
				lineStyle := lipgloss.NewStyle().
					Foreground(textColor).
					Width(contentWidth)
				// Get name from i18n
				itemName := m.presenter.GetText("slots." + item.ID + ".name")
				lineText := "  ✓ " + itemName
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
	count := 0
	for _, item := range m.currentItems {
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

// GetSelectedSlot returns the slot that was selected in the wizard (e.g., "dev", "data", "sandbox").
func (m *WizardModel) GetSelectedSlot() string {
	return m.selectedSlot
}

// RunWizard runs the interactive wizard-style slot selector TUI.
// Returns the selected item IDs, whether the user confirmed (true) or cancelled (false), and any error.
func RunWizard(items []slots.SlotItem, pres ports.IPresenter) ([]string, bool, error) {
	model := NewWizardModel(pres)

	// Store all items for later filtering when slot is selected
	model.allItems = items

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
		return nil, false, fmt.Errorf("errors.ui.wizard_selector_failed: %w", err)
	}

	resultModel, ok := finalModel.(*WizardModel)
	if !ok {
		return nil, false, fmt.Errorf("errors.ui.unexpected_model")
	}

	if resultModel.canceled {
		return nil, false, nil
	}

	return resultModel.GetSelectedIDs(), true, nil
}

// WizardResult holds the result of a wizard selection.
type WizardResult struct {
	SelectedIDs  []string
	SelectedSlot string
}

// RunWizardWithSlot runs the wizard and returns both selected items AND the selected slot.
// This is useful for build operations where the slot determines the image name.
func RunWizardWithSlot(items []slots.SlotItem, pres ports.IPresenter) ([]string, string, bool, error) {
	model := NewWizardModel(pres)

	// Store all items for later filtering when slot is selected
	model.allItems = items

	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
	)

	finalModel, err := p.Run()

	// Ensure terminal is cleaned up properly even if there's an error
	cleanupTerminal()

	if err != nil {
		return nil, "", false, fmt.Errorf("errors.ui.wizard_selector_failed: %w", err)
	}

	resultModel, ok := finalModel.(*WizardModel)
	if !ok {
		return nil, "", false, fmt.Errorf("errors.ui.unexpected_model")
	}

	if resultModel.canceled {
		return nil, "", false, nil
	}

	return resultModel.GetSelectedIDs(), resultModel.GetSelectedSlot(), true, nil
}

// cleanupTerminal forces exit from alternate screen mode and flushes stdout
// to prevent corrupted display when running Bubble Tea programs multiple times
func cleanupTerminal() {
	// Exit alternate screen mode
	fmt.Print("\033[?1049l")
	// Reset cursor visibility (show cursor)
	fmt.Print("\033[?25h")
	// Flush stdout to ensure sequences are sent
	os.Stdout.Sync()
}
