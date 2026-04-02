package build

import (
	"context"
	"fmt"

	"github.com/Alejandro-M-P/AXIOM/internal/core/slots"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// SlotInstaller implements ports.ISlotInstaller using the slot registry and engine.
type SlotInstaller struct {
	manager *slots.SlotManager
}

// NewSlotInstaller creates a new SlotInstaller.
func NewSlotInstaller(manager *slots.SlotManager) *SlotInstaller {
	return &SlotInstaller{manager: manager}
}

// Install implements ports.ISlotInstaller.
// It looks up the full SlotItem from the registry and executes installation.
func (i *SlotInstaller) Install(ctx context.Context, item ports.SlotItem, exec ports.ICommandRunner) error {
	// Convert ports.SlotItem to slots.SlotItem by looking up in registry
	// For simplicity, we assume the item exists in the registry.
	// We'll use the manager's Discover method to find it.
	allItems := i.manager.Discover()
	var fullItem *slots.SlotItem
	for _, si := range allItems {
		if si.ID == item.ID {
			fullItem = &si
			break
		}
	}
	if fullItem == nil {
		return fmt.Errorf("slot item not found in registry: %s", item.ID)
	}

	// Use the manager's engine to execute installation (requires converting to []slots.SlotItem)
	// The engine expects a progress callback; we'll pass nil.
	// Since we only have one item, we can call executeInstall directly (private).
	// Instead, we'll use the manager's Execute method which handles dependencies.
	// However, Execute expects []slots.SlotItem and a progress callback.
	// We'll create a simple progress callback that logs.
	items := []slots.SlotItem{*fullItem}
	return i.manager.Execute(items)
}

// Name implements ports.ISlotInstaller.
func (i *SlotInstaller) Name() string {
	return "default-slot-installer"
}
