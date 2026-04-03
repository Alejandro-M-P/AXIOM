package ports

import "context"

// IBuildProgress defines the port for tracking and rendering build progress.
// The UI layer implements this; the core only signals state changes.
type IBuildProgress interface {
	// StartStep marks a step as running and renders.
	StartStep(index int, title string, detail string)
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
	ExecuteBuild(ctx context.Context, items []BuildItem, containerName string, cfg BuildConfig, progress IBuildProgress) error
}
