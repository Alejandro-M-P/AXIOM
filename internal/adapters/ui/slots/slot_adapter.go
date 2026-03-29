// Package slots provides Bubbletea TUI components for slot selection.
package slots

import (
	"fmt"

	"axiom/internal/ports"
	"axiom/internal/slots"
)

// SlotSelectorUI provides the TUI for slot selection.
// It wraps the slot manager and provides interactive selection.
type SlotSelectorUI struct {
	manager   *slots.SlotManager
	presenter ports.IPresenter
}

// NewSlotSelectorUI creates a new SlotSelectorUI.
func NewSlotSelectorUI(manager *slots.SlotManager, pres ports.IPresenter) *SlotSelectorUI {
	return &SlotSelectorUI{
		manager:   manager,
		presenter: pres,
	}
}

// RunSlotSelectorWithItems presents the interactive slot selector TUI and returns selected item IDs.
// It converts from slots.SlotItem to the UI's ItemGroup format.
func (u *SlotSelectorUI) RunSlotSelectorWithItems(category string, items []slots.SlotItem, preselected []string) ([]string, bool, error) {
	// Build ItemGroups from SlotItems
	groups := buildItemGroups(items)

	// Run the selector
	selectedIDs, confirmed, err := runSlotSelectorWithGroups(groups)
	if err != nil {
		return nil, false, err
	}

	if !confirmed {
		return nil, false, nil // User canceled
	}

	return selectedIDs, true, nil
}

// Execute implements slots.UISelectorRunner.
// It takes ItemGroup directly and returns selected IDs or error.
func (u *SlotSelectorUI) Execute(groups []ItemGroup) ([]string, error) {
	ids, _, err := runSlotSelectorWithGroups(groups)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		return []string{}, nil
	}
	return ids, nil
}

// runSlotSelectorWithGroups is an internal helper that converts the boolean return
// to match the UISelectorRunner interface.
func runSlotSelectorWithGroups(groups []ItemGroup) ([]string, bool, error) {
	result, confirmed, err := RunSlotSelector(groups)
	if err != nil {
		return nil, false, err
	}
	if !confirmed {
		return nil, false, nil
	}
	return result, true, nil
}

// RunWizard presents the wizard-style slot selector.
// Returns selected item IDs and whether user confirmed.
func (u *SlotSelectorUI) RunWizard(items []slots.SlotItem) ([]string, bool, error) {
	return RunWizard(items, u.presenter)
}

// RunWizardWithSlotItems presents the wizard-style slot selector and returns both
// selected item IDs and the selected slot (e.g., "dev", "data", "sandbox").
// This is useful for build operations where the slot determines the image name.
func (u *SlotSelectorUI) RunWizardWithSlotItems(items []slots.SlotItem) ([]string, string, bool, error) {
	return RunWizardWithSlot(items, u.presenter)
}

// RunWizardWithSlot implements ports.ISlotUI.
// It accepts any type and expects []slots.SlotItem.
func (u *SlotSelectorUI) RunWizardWithSlot(items any) ([]string, string, bool, error) {
	slotItems, ok := items.([]slots.SlotItem)
	if !ok {
		return nil, "", false, fmt.Errorf("invalid items type, expected []slots.SlotItem")
	}
	return u.RunWizardWithSlotItems(slotItems)
}

// buildItemGroups converts []slots.SlotItem to []ItemGroup for the UI.
// Items are grouped by their SubCategory field.
func buildItemGroups(items []slots.SlotItem) []ItemGroup {
	// Group items by subcategory
	groupsMap := make(map[string][]SlotItemDisplay)
	order := []string{"ia", "languages", "tools", "data"}

	for _, item := range items {
		display := SlotItemDisplay{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
		}
		groupsMap[item.SubCategory] = append(groupsMap[item.SubCategory], display)
	}

	// Build ordered result
	var result []ItemGroup
	for _, subcategory := range order {
		if items, ok := groupsMap[subcategory]; ok {
			title := getSubcategoryTitle(subcategory)
			result = append(result, ItemGroup{title: title, items: items})
		}
	}

	// Add any remaining subcategories not in predefined order
	for subcategory, items := range groupsMap {
		if !contains(order, subcategory) {
			title := getSubcategoryTitle(subcategory)
			result = append(result, ItemGroup{title: title, items: items})
		}
	}

	return result
}

var _ ports.ISlotUI = (*SlotSelectorUI)(nil)
