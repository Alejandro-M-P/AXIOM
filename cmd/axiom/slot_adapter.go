package main

import (
	"axiom/internal/adapters/ui/slots"
	"axiom/internal/build"
	slotmanager "axiom/internal/slots"
)

// buildSlotAdapter implements build.SlotManagerInterface for the build manager.
// It converts between build types and slots types.
type buildSlotAdapter struct {
	manager *slotmanager.SlotManager
	ui      *slots.SlotSelectorUI
}

// uiRunnerAdapter implements slotmanager.UISelectorRunner by wrapping a SlotSelectorUI.
// This allows the SlotManager to use the TUI selector while keeping the UI dependency
// in the adapter layer.
type uiRunnerAdapter struct {
	ui *slots.SlotSelectorUI
}

// RunSlotSelector implements slotmanager.UISelectorRunner.
func (a *uiRunnerAdapter) RunSlotSelector(groups []slotmanager.ItemGroup) ([]string, error) {
	// Use the Builder to construct ItemGroups since fields are private
	builder := slots.NewBuilder()
	for _, g := range groups {
		for _, item := range g.Items {
			// Builder.AddItem takes (subcategory, id, name, description)
			// We use the group title as subcategory key
			builder.AddItem(g.Title, item.ID, item.Name, item.Description)
		}
	}
	uiGroups := builder.Build()
	return a.ui.Execute(uiGroups)
}

// newBuildSlotAdapter creates a new build slot adapter.
func newBuildSlotAdapter(manager *slotmanager.SlotManager) *buildSlotAdapter {
	ui := slots.NewSlotSelectorUI(manager)

	// Create the UI runner adapter and inject it into the SlotManager
	runner := &uiRunnerAdapter{ui: ui}
	manager.SetUISelectorRunner(runner)

	return &buildSlotAdapter{
		manager: manager,
		ui:      ui,
	}
}

// HasSelection implements build.SlotManagerInterface.
func (a *buildSlotAdapter) HasSelection() bool {
	return a.manager.HasSelection()
}

// GetAvailableItems returns all available items for a given category.
func (a *buildSlotAdapter) GetAvailableItems(category string) ([]slotmanager.SlotItem, error) {
	items := a.manager.GetByCategory(slotmanager.SlotCategory(category))
	return items, nil
}

// GetSelectedItems implements build.SlotManagerInterface.
func (a *buildSlotAdapter) GetSelectedItems(category string) ([]build.SlotItem, error) {
	items, err := a.manager.GetSelectedItems(category)
	if err != nil {
		return nil, err
	}
	result := make([]build.SlotItem, len(items))
	for i, item := range items {
		result[i] = build.SlotItem{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			Category:    string(item.Category),
			Deps:        item.Deps,
		}
	}
	return result, nil
}

// RunSlotSelector implements build.SlotManagerInterface.
func (a *buildSlotAdapter) RunSlotSelector(category string, items []build.SlotItem, preselected []string) ([]string, bool, error) {
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

// SaveSelection implements build.SlotManagerInterface.
func (a *buildSlotAdapter) SaveSelection(selections []build.SlotSelection) error {
	slotSelections := make([]slotmanager.SlotSelection, len(selections))
	for i, sel := range selections {
		slotSelections[i] = slotmanager.SlotSelection{
			Slot:     slotmanager.SlotCategory(sel.Slot),
			Selected: sel.Selected,
		}
	}
	return a.manager.SaveSelection(slotSelections)
}

// LoadSelection implements build.SlotManagerInterface.
func (a *buildSlotAdapter) LoadSelection() ([]build.SlotSelection, error) {
	selections, err := a.manager.LoadSelection()
	if err != nil {
		return nil, err
	}
	result := make([]build.SlotSelection, len(selections))
	for i, sel := range selections {
		result[i] = build.SlotSelection{
			Slot:     string(sel.Slot),
			Selected: sel.Selected,
		}
	}
	return result, nil
}

// Ensure buildSlotAdapter implements build.SlotManagerInterface at compile time.
var _ build.SlotManagerInterface = (*buildSlotAdapter)(nil)
