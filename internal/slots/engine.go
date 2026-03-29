// Package slots implements the slot-based installation system for AXIOM.
package slots

import (
	"context"
	"errors"
	"fmt"
)

// ErrCircularDependency indicates that a circular dependency was detected.
var ErrCircularDependency = errors.New("circular dependency detected")

// IInstallerEngine defines the interface for executing slot item installations.
// It handles dependency resolution and step execution.
type IInstallerEngine interface {
	// Execute runs the installation for the specified items in dependency order.
	// The Executor function is called for each item to perform the actual installation.
	Execute(items []SlotItem, exec Executor) error

	// ExecuteWithContext runs the installation with context for cancellation support.
	// The itemExec callback is called for each item with the item as parameter.
	ExecuteWithContext(ctx context.Context, items []SlotItem, itemExec func(ctx context.Context, item *SlotItem) error) error

	// ResolveDependencies returns the items in correct dependency order.
	// Items are sorted so that dependencies come before the items that depend on them.
	ResolveDependencies(items []SlotItem) ([]SlotItem, error)
}

// InstallerEngine implements IInstallerEngine.
type InstallerEngine struct {
	registry ISlotRegistry
}

// NewInstallerEngine creates a new InstallerEngine instance.
func NewInstallerEngine(registry ISlotRegistry) *InstallerEngine {
	return &InstallerEngine{
		registry: registry,
	}
}

// Execute runs the installation for the specified items in dependency order.
// Each item's own Executor closure is called to perform the actual installation.
// The exec parameter is a progress callback called before each item is installed.
func (e *InstallerEngine) Execute(items []SlotItem, exec Executor) error {
	if len(items) == 0 {
		return nil
	}

	// Resolve dependencies to get correct installation order
	orderedItems, err := e.ResolveDependencies(items)
	if err != nil {
		return err
	}

	// Execute each item in order
	for i := range orderedItems {
		item := &orderedItems[i]

		// Call progress callback if provided
		if exec != nil {
			if err := exec(context.Background()); err != nil {
				return fmt.Errorf("installation cancelled for %s: %w", item.ID, err)
			}
		}

		// Execute the item's installation logic
		if item.Executor != nil {
			if err := item.Executor(context.Background()); err != nil {
				return fmt.Errorf("failed to install %s: %w", item.ID, err)
			}
		}
	}

	return nil
}

// ResolveDependencies returns the items in correct dependency order using topological sort.
// It detects circular dependencies and returns an error if found.
func (e *InstallerEngine) ResolveDependencies(items []SlotItem) ([]SlotItem, error) {
	if len(items) == 0 {
		return nil, nil
	}

	// Build a map of item IDs for quick lookup
	itemMap := make(map[string]*SlotItem)
	for i := range items {
		itemMap[items[i].ID] = &items[i]
	}

	// Track visited nodes and nodes in current path (for cycle detection)
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var result []SlotItem

	// Visit function for topological sort (DFS)
	var visit func(id string) error
	visit = func(id string) error {
		// Check if item exists in our selection
		item, exists := itemMap[id]
		if !exists {
			// Item might be a dependency not in selection, try to get from registry
			registeredItem, err := e.registry.GetByID(id)
			if err != nil {
				return fmt.Errorf("unknown dependency: %s", id)
			}
			item = registeredItem
			itemMap[id] = item
		}

		// Check for circular dependency
		if inStack[id] {
			return fmt.Errorf("%w: %s", ErrCircularDependency, id)
		}

		// Skip if already visited
		if visited[id] {
			return nil
		}

		// Mark as in progress
		inStack[id] = true

		// Visit all dependencies first
		for _, depID := range item.Dependencies() {
			if err := visit(depID); err != nil {
				return err
			}
		}

		// Mark as done
		inStack[id] = false
		visited[id] = true

		// Add to result (dependencies come first due to post-order traversal)
		result = append(result, *item)

		return nil
	}

	// Visit all items in the selection
	for _, item := range items {
		if err := visit(item.ID); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// ExecuteWithContext runs the installation with a context for cancellation support.
// The itemExec callback is called for each item with the item as parameter.
func (e *InstallerEngine) ExecuteWithContext(ctx context.Context, items []SlotItem, itemExec func(ctx context.Context, item *SlotItem) error) error {
	if len(items) == 0 {
		return nil
	}

	orderedItems, err := e.ResolveDependencies(items)
	if err != nil {
		return err
	}

	for i := range orderedItems {
		item := &orderedItems[i]
		if err := itemExec(ctx, item); err != nil {
			return fmt.Errorf("failed to install %s: %w", item.ID, err)
		}
	}

	return nil
}
