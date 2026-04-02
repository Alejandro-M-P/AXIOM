// Package slots implements the slot-based installation system for AXIOM.
package slots

import (
	"context"
	"errors"
	"fmt"

	"github.com/Alejandro-M-P/AXIOM/internal/core/slots/base"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// ErrCircularDependency indicates that a circular dependency was detected.
var ErrCircularDependency = errors.New("engine.circular_dependency")

// IInstallerEngine defines the interface for executing slot item installations.
// It handles dependency resolution and step execution.
type IInstallerEngine interface {
	// Execute runs the installation for the specified items in dependency order.
	// Commands are read from TOML configuration (InstallCmd or InstallSteps).
	// The progress callback is called before each item installation.
	Execute(items []SlotItem, progress Executor) error

	// ExecuteWithContext runs the installation with context for cancellation support.
	// The itemExec callback is called for each item with the item as parameter.
	ExecuteWithContext(ctx context.Context, items []SlotItem, itemExec func(ctx context.Context, item *SlotItem) error) error

	// ResolveDependencies returns the items in correct dependency order.
	// Items are sorted so that dependencies come before the items that depend on them.
	ResolveDependencies(items []SlotItem) ([]SlotItem, error)
}

// InstallerEngine implements IInstallerEngine.
type InstallerEngine struct {
	registry      ISlotRegistry
	baseInstaller *base.BaseInstaller
	analyzer      *base.SlotCommandAnalyzer
	runner        ports.ICommandRunner
}

// NewInstallerEngine creates a new InstallerEngine instance.
func NewInstallerEngine(registry ISlotRegistry, runner ports.ICommandRunner) *InstallerEngine {
	return &InstallerEngine{
		registry: registry,
		runner:   runner,
	}
}

// NewInstallerEngineWithBase creates a new InstallerEngine with base tools support.
// This enables automatic installation of package managers and base tools.
func NewInstallerEngineWithBase(registry ISlotRegistry, preferencesPath string, runner ports.ICommandRunner) (*InstallerEngine, error) {
	installer, err := base.NewBaseInstaller(preferencesPath)
	if err != nil {
		return nil, fmt.Errorf("engine.base_tools_failed: %w", err)
	}

	return &InstallerEngine{
		registry:      registry,
		baseInstaller: installer,
		analyzer:      base.NewSlotCommandAnalyzer(installer),
		runner:        runner,
	}, nil
}

// PrepareEnvironment installs base tools and prepares the system for slot installation.
// This should be called once before executing slot installations.
func (e *InstallerEngine) PrepareEnvironment(ctx context.Context) error {
	if e.analyzer == nil {
		// Base tools not enabled, skip
		return nil
	}
	return e.analyzer.PrepareEnvironment(ctx)
}

// GetBaseInstaller returns the base installer if configured.
func (e *InstallerEngine) GetBaseInstaller() *base.BaseInstaller {
	return e.baseInstaller
}

// Execute runs the installation for the specified items in dependency order.
// Installation commands are read from TOML (InstallCmd or InstallSteps).
// The exec parameter is a progress callback called before each item is installed.
func (e *InstallerEngine) Execute(items []SlotItem, progress Executor) error {
	ctx := context.Background()
	return e.ExecuteWithContext(ctx, items, func(ctx context.Context, item *SlotItem) error {
		// Call progress callback if provided
		if progress != nil {
			if err := progress(ctx); err != nil {
				return fmt.Errorf("engine.cancelled: %s: %w", item.ID, err)
			}
		}
		return e.executeInstall(ctx, item)
	})
}

// executeInstall runs the installation commands for a single item.
// Uses InstallCmd if present, otherwise executes InstallSteps sequentially.
// Automatically ensures base tools are installed before executing commands.
func (e *InstallerEngine) executeInstall(ctx context.Context, item *SlotItem) error {
	// Skip if no install commands defined
	if item.InstallCmd == "" && len(item.InstallSteps) == 0 {
		return nil
	}

	// Single command mode
	if item.InstallCmd != "" {
		// Ensure base tools are installed before executing
		if e.analyzer != nil {
			if err := e.analyzer.AnalyzeAndInstallRequirements(ctx, item.InstallCmd); err != nil {
				return fmt.Errorf("engine.base_tools_failed: %s: %w", item.ID, err)
			}
		}

		if _, err := e.runner.RunShell(ctx, item.InstallCmd); err != nil {
			return fmt.Errorf("engine.install_failed: %s: %w", item.ID, err)
		}
		return nil
	}

	// Multi-step mode
	for _, step := range item.InstallSteps {
		// Ensure base tools are installed before each step
		if e.analyzer != nil {
			if err := e.analyzer.AnalyzeAndInstallRequirements(ctx, step); err != nil {
				return fmt.Errorf("engine.base_tools_failed: %s: %w", item.ID, err)
			}
		}

		if _, err := e.runner.RunShell(ctx, step); err != nil {
			return fmt.Errorf("engine.step_failed: %s: %w", item.ID, err)
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
				return fmt.Errorf("engine.unknown_dependency: %s", id)
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
			return fmt.Errorf("engine.install_failed: %s: %w", item.ID, err)
		}
	}

	return nil
}
