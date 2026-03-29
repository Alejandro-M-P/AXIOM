// Package slots implements the slot-based installation system for AXIOM.
// It allows users to compose development environments by selecting individual
// installation items grouped into categories (slots).
package slots

import (
	"context"
)

// SlotCategory represents the category/type of a slot.
type SlotCategory string

// Slot category constants.
const (
	SlotDEV     SlotCategory = "dev"
	SlotDATA    SlotCategory = "data"
	SlotSANDBOX SlotCategory = "sandbox"
)

// Executor is a function type that executes an installation step.
// It takes a context and returns an error if the execution fails.
type Executor func(ctx context.Context) error

// SlotItem represents a single installable unit within a slot.
// It contains all information needed to install and manage that item.
type SlotItem struct {
	ID          string       // Unique identifier (e.g., "ollama", "go")
	Name        string       // Human-readable name
	Description string       // Description of what this item provides
	Category    SlotCategory // Which slot this item belongs to (dev, data, sandbox)
	SubCategory string       // Subcategory for UI grouping (ia, languages, tools, data)
	Deps        []string     // IDs of items that must be installed first
	Executor    Executor     // The actual installation logic
}

// SlotSelection represents a user's selection for a particular slot.
// This is persisted to configuration files.
type SlotSelection struct {
	Slot     SlotCategory `toml:"slot"`
	Selected []string     `toml:"selected"` // Item IDs selected by user
}

// Dependencies returns the list of dependencies for this item.
func (s *SlotItem) Dependencies() []string {
	return s.Deps
}

// GetCategory returns the slot category for this item.
func (s *SlotItem) GetCategory() SlotCategory {
	return s.Category
}
