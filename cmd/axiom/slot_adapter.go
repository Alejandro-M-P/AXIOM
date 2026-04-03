package main

import (
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/ui/slots"
	slotmanager "github.com/Alejandro-M-P/AXIOM/internal/core/slots"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// buildSlotAdapter implements ports.SlotManagerInterface for the build manager.
// It converts between ports types and slots types.
type buildSlotAdapter struct {
	manager *slotmanager.SlotManager
	ui      *slots.SlotSelectorUI
}

// uiRunnerAdapter implements slotmanager.UISelectorRunner by wrapping a SlotSelectorUI.
// This allows the SlotManager to use the TUI selector while keeping the UI dependency
// in the adapter layer.
type uiRunnerAdapter struct {
	ui      *slots.SlotSelectorUI
	manager *slotmanager.SlotManager
}

// RunSlotSelector implements slotmanager.UISelectorRunner.
// It calls the wizard-style selector instead of the flat selector.
func (a *uiRunnerAdapter) RunSlotSelector(groups []slotmanager.ItemGroup) ([]string, error) {
	// Get all items from registry to pass to wizard
	items := a.manager.Discover()

	// Call the wizard with all items
	selected, confirmed, err := a.ui.RunWizard(items)
	if err != nil {
		return nil, err
	}
	if !confirmed {
		return nil, nil // User cancelled
	}
	return selected, nil
}

// newBuildSlotAdapter creates a new build slot adapter.
func newBuildSlotAdapter(manager *slotmanager.SlotManager, pres ports.IPresenter) *buildSlotAdapter {
	ui := slots.NewSlotSelectorUI(manager, pres)

	// Create the UI runner adapter and inject it into the SlotManager
	runner := &uiRunnerAdapter{ui: ui, manager: manager}
	manager.SetUISelectorRunner(runner)

	return &buildSlotAdapter{
		manager: manager,
		ui:      ui,
	}
}

// HasSelection implements ports.SlotManagerInterface.
func (a *buildSlotAdapter) HasSelection() bool {
	return a.manager.HasSelection()
}

// GetAvailableItems returns all available items for a given category.
func (a *buildSlotAdapter) GetAvailableItems(category string) ([]slotmanager.SlotItem, error) {
	items := a.manager.GetByCategory(slotmanager.SlotCategory(category))
	return items, nil
}

// GetSelectedItems implements ports.SlotManagerInterface.
func (a *buildSlotAdapter) GetSelectedItems(category string) ([]ports.SlotItem, error) {
	items, err := a.manager.GetSelectedItems(category)
	if err != nil {
		return nil, err
	}
	result := make([]ports.SlotItem, len(items))
	for i, item := range items {
		result[i] = ports.SlotItem{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Category:    string(item.Category),
			Deps:        item.Deps,
		}
	}
	return result, nil
}

// RunSlotSelector implements ports.SlotManagerInterface.
func (a *buildSlotAdapter) RunSlotSelector(category string, items []ports.SlotItem, preselected []string) ([]string, bool, error) {
	slotItems := make([]slotmanager.SlotItem, len(items))
	for i, item := range items {
		slotItems[i] = slotmanager.SlotItem{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Category:    slotmanager.SlotCategory(item.Category),
			Deps:        item.Deps,
		}
	}
	return a.ui.RunSlotSelectorWithItems(category, slotItems, preselected)
}

// SaveSelection implements ports.SlotManagerInterface.
func (a *buildSlotAdapter) SaveSelection(selections []ports.SlotSelection) error {
	slotSelections := make([]slotmanager.SlotSelection, len(selections))
	for i, sel := range selections {
		slotSelections[i] = slotmanager.SlotSelection{
			Slot:     slotmanager.SlotCategory(sel.Slot),
			Selected: sel.Selected,
		}
	}
	return a.manager.SaveSelection(slotSelections)
}

// LoadSelection implements ports.SlotManagerInterface.
func (a *buildSlotAdapter) LoadSelection() ([]ports.SlotSelection, error) {
	selections, err := a.manager.LoadSelection()
	if err != nil {
		return nil, err
	}
	result := make([]ports.SlotSelection, len(selections))
	for i, sel := range selections {
		result[i] = ports.SlotSelection{
			Slot:     string(sel.Slot),
			Selected: sel.Selected,
		}
	}
	return result, nil
}

// Ensure buildSlotAdapter implements ports.SlotManagerInterface at compile time.
var _ ports.SlotManagerInterface = (*buildSlotAdapter)(nil)
