// Package slots implements the slot-based installation system for AXIOM.
package slots

import (
	"fmt"
)

// ISlotRegistry defines the interface for discovering and accessing slot items.
type ISlotRegistry interface {
	// Discover returns all available slot items across all categories.
	Discover() []SlotItem

	// GetByCategory returns all items belonging to the specified slot category.
	GetByCategory(category SlotCategory) []SlotItem

	// GetByID returns a single slot item by its unique identifier.
	// Returns an error if the item is not found.
	GetByID(id string) (*SlotItem, error)
}

// Registry implements ISlotRegistry using a package-level registration system.
// Items are registered via Register() method or through init() functions.
type Registry struct {
	itemsByID       map[string]*SlotItem
	itemsByCategory map[SlotCategory][]*SlotItem
}

// globalRegistry is the singleton registry instance.
var globalRegistry *Registry

// init registers the global registry.
func init() {
	globalRegistry = NewRegistry()
}

// NewRegistry creates a new empty registry instance.
func NewRegistry() *Registry {
	return &Registry{
		itemsByID:       make(map[string]*SlotItem),
		itemsByCategory: make(map[SlotCategory][]*SlotItem),
	}
}

// Register adds a slot item to the registry.
// If an item with the same ID already exists, it will be overwritten.
func (r *Registry) Register(item *SlotItem) {
	r.itemsByID[item.ID] = item
	r.itemsByCategory[item.Category] = append(r.itemsByCategory[item.Category], item)
}

// Discover returns all registered slot items.
func (r *Registry) Discover() []SlotItem {
	result := make([]SlotItem, 0, len(r.itemsByID))
	for _, item := range r.itemsByID {
		result = append(result, *item)
	}
	return result
}

// GetByCategory returns all items belonging to the specified slot category.
func (r *Registry) GetByCategory(category SlotCategory) []SlotItem {
	items := r.itemsByCategory[category]
	result := make([]SlotItem, len(items))
	for i, item := range items {
		result[i] = *item
	}
	return result
}

// GetByID returns a single slot item by its unique identifier.
// Returns an error if the item is not found.
func (r *Registry) GetByID(id string) (*SlotItem, error) {
	item, exists := r.itemsByID[id]
	if !exists {
		return nil, fmt.Errorf("slot item not found: %s", id)
	}
	return item, nil
}

// GetRegistry returns the global registry instance.
func GetRegistry() ISlotRegistry {
	return globalRegistry
}

// RegisterItem adds an item to the global registry.
// This is a convenience function for external packages to register items.
func RegisterItem(item *SlotItem) {
	globalRegistry.Register(item)
}
