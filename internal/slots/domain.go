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

// Executor is a function type for progress callbacks during installation.
// It takes a context and returns an error if the execution should be cancelled.
type Executor func(ctx context.Context) error

// SlotItem represents a single installable unit within a slot.
// It contains all information needed to install and manage that item.
type SlotItem struct {
	ID           string       // Unique identifier (e.g., "ollama", "go")
	Name         string       // Human-readable name
	Description  string       // Description of what this item provides
	Category     SlotCategory // Which slot this item belongs to (dev, data, sandbox)
	SubCategory  string       // Subcategory for UI grouping (ia, languages, tools, data)
	Deps         []string     // IDs of items that must be installed first
	InstallCmd   string       // Command to install (from TOML)
	InstallSteps []string     // Multiple steps to install (from TOML, alternative to InstallCmd)
	IsBaseTool   bool         // If true, this is a base tool not shown in the wizard
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
