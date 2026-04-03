package ports

import "context"

// SlotManagerInterface defines the contract for slot operations during build.
// This avoids importing the slots package directly to prevent circular dependencies.
type SlotManagerInterface interface {
	// HasSelection returns true if slot selections exist for any category.
	HasSelection() bool

	// GetSelectedItems returns the selected slot items for a given category.
	GetSelectedItems(category string) ([]SlotItem, error)

	// RunSlotSelector presents the slot selection UI and returns selected item IDs.
	RunSlotSelector(category string, items []SlotItem, preselected []string) ([]string, bool, error)

	// SaveSelection persists the user's slot selections.
	SaveSelection(selections []SlotSelection) error

	// LoadSelection reads the user's slot selections.
	LoadSelection() ([]SlotSelection, error)
}

// SlotItem represents a single installable unit within a slot.
type SlotItem struct {
	ID          string
	Name        string
	Description string
	Category    string
	Deps        []string
}

// SlotSelection represents a user's selection for a particular slot.
type SlotSelection struct {
	Slot     string   `toml:"slot"`
	Selected []string `toml:"selected"`
}

// IBuildProgress defines the port for tracking and rendering build progress.
// The UI layer implements this; the core only signals state changes.
type IBuildProgress interface {
	// StartStep marks a step as running and renders. titleKey is an i18n key, titleParams are passed to GetText.
	StartStep(index int, titleKey string, titleParams []string, detail string)
	// FinishStep marks the current step as done and renders.
	FinishStep()
	// FailStep marks the current step as failed with an error.
	FailStep(err error)
	// Render forces a re-render of the current state.
	Render()
}

// BuildItem represents a single installable unit during a build.
// The core creates these from slot selections; the adapter executes them.
type BuildItem struct {
	ID          string
	Name        string
	Description string
	Category    string
	Deps        []string // System packages needed by this item
	NeedsOllama bool     // Whether this item requires Ollama installed
}

// BuildConfig holds runtime configuration needed by the installer.
type BuildConfig struct {
	GPUType string // Canonical GPU type (e.g. "nvidia", "rdna3", "generic")
}

// IBuildInstaller defines the port for executing a complete build.
// The adapter layer implements this with exec.Command, pacman, distrobox, etc.
// The core is blind to HOW installation happens.
type IBuildInstaller interface {
	// ExecuteBuild installs base packages, slot dependencies, runs install commands,
	// installs Ollama if needed, and cleans caches. All execution happens in the
	// container referenced by the installer's runtime.
	ExecuteBuild(ctx context.Context, items []BuildItem, containerName string, cfg BuildConfig, progress IBuildProgress, slotManager SlotManagerInterface) error
}
