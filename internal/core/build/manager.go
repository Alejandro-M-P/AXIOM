package build

import (
	"context"
	"fmt"

	"github.com/Alejandro-M-P/AXIOM/internal/config"
	"github.com/Alejandro-M-P/AXIOM/internal/core/slots"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// Manager orchestrates build operations.
type Manager struct {
	runtime        ports.IBunkerRuntime
	fs             ports.IFileSystem
	ui             ports.IPresenter
	system         ports.ISystem
	buildContainer string
	slotManager    SlotManagerInterface
	installer      ports.IBuildInstaller
}

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

// NewManager creates a new build manager.
func NewManager(runtime ports.IBunkerRuntime, fs ports.IFileSystem, ui ports.IPresenter, system ports.ISystem, buildContainer string, slotManager SlotManagerInterface, installer ports.IBuildInstaller) *Manager {
	return &Manager{
		runtime:        runtime,
		fs:             fs,
		ui:             ui,
		system:         system,
		buildContainer: buildContainer,
		slotManager:    slotManager,
		installer:      installer,
	}
}

// GetUI returns the presenter for use by the router (adapter layer).
func (m *Manager) GetUI() ports.IPresenter {
	return m.ui
}

// Build generates and returns a BuildPlan for the complete build flow.
// The caller (adapter) is responsible for executing the plan with progress tracking.
// Pre-flight checks (slot selection, image existence) are performed before plan generation.
func (m *Manager) Build(ctx context.Context, cfg config.EnvConfig) (*BuildPlan, error) {
	// Load the slot selection that was saved by the router
	selections, err := m.slotManager.LoadSelection()
	if err != nil {
		return nil, fmt.Errorf("errors.build.failed_load_selection: %w", err)
	}
	if len(selections) == 0 {
		return nil, fmt.Errorf("errors.build.cancelled_no_selection")
	}

	// Determine slot name from selection
	slotName := "dev"
	if len(selections) > 0 && selections[0].Slot != "" {
		slotName = selections[0].Slot
	}

	// Prepare build context - use slot name for image
	buildCtx, err := PrepareBuildContext(ctx, cfg, "", slotName, m.system, m.ui)
	if err != nil {
		return nil, err
	}

	// Pre-flight: Check if the slot image already exists
	imageName := buildCtx.ImageName
	exists, err := m.runtime.ImageExists(ctx, imageName)
	if err != nil {
		m.ui.ShowLog("build.image_check_failed", err.Error())
	}
	if exists {
		confirm, err := m.ui.AskConfirmInCard(
			"build",
			[]ports.Field{
				ports.NewField(m.ui.GetText("build.image_exists_title"), imageName),
				ports.NewField(m.ui.GetText("build.image_exists_warning"), m.ui.GetText("build.image_exists_desc")),
			},
			nil,
			"build.image_exists_confirm",
		)
		if err != nil {
			return nil, fmt.Errorf("errors.build.failed_ask_confirmation: %w", err)
		}
		if !confirm {
			return nil, fmt.Errorf("errors.build.cancelled_image_exists")
		}
		// Delete existing image
		if err := m.runtime.RemoveImage(ctx, imageName, true); err != nil {
			return nil, fmt.Errorf("errors.build.failed_remove_image: %w", err)
		}
	}

	// Build the execution plan
	plan := m.makeBuildPlan(ctx, buildCtx, slotName, selections)
	return plan, nil
}

// makeBuildPlan creates a BuildPlan with steps based on slot type.
// The core defines WHAT to do; the adapter decides HOW to execute and render it.
func (m *Manager) makeBuildPlan(ctx context.Context, buildCtx *BuildContext, slotName string, selections []SlotSelection) *BuildPlan {
	gpuModeText := m.ui.GetText("build.subtitle_host")
	if buildCtx.Config.ROCMMode == "image" {
		gpuModeText = m.ui.GetText("build.subtitle_image")
	}

	containerName := buildCtx.ContainerName
	cleanupDone := false

	// Collect BuildItems from slot selections for the installer
	buildItems := m.collectBuildItems(selections)

	var steps []BuildStep

	// Common step 0: Prepare shared directories
	steps = append(steps, BuildStep{
		Title:  m.ui.GetText("step.prepare_dirs"),
		Detail: config.AIConfigDir(buildCtx.Config.BaseDir),
		Exec: func(ctx context.Context) error {
			return PrepareSharedDirectories(ctx, m.fs, buildCtx.Config)
		},
	})

	// Common step 1: Recreate build container
	steps = append(steps, BuildStep{
		Title:  m.ui.GetText("step.recreate_container"),
		Detail: buildCtx.BuildWorkspaceDir,
		Exec: func(ctx context.Context) error {
			return RecreateBuildContainer(ctx, m.runtime, m.fs, containerName, buildCtx.BuildWorkspaceDir, buildCtx.Config, buildCtx.GPUInfo.Type)
		},
	})

	// Common step 2: Install everything via the adapter's BuildInstaller
	steps = append(steps, BuildStep{
		Title:  m.ui.GetText("step.install_base"),
		Detail: m.ui.GetText("detail.base_pkgs"),
		Exec: func(ctx context.Context) error {
			bcfg := ports.BuildConfig{GPUType: buildCtx.GPUInfo.Type}
			return m.installer.ExecuteBuild(ctx, buildItems, containerName, bcfg, nil, m.slotManager)
		},
	})

	// Slot-specific display steps (no installation logic - installer handles it)
	switch slotName {
	case "dev":
		steps = append(steps, BuildStep{
			Title:  m.ui.GetText("step.install_dev"),
			Detail: m.ui.GetText("detail.dev_tools"),
			Exec:   func(ctx context.Context) error { return nil },
		})
		steps = append(steps, BuildStep{
			Title:  m.ui.GetText("step.install_ai"),
			Detail: m.ui.GetText("detail.ai_stack"),
			Exec:   func(ctx context.Context) error { return nil },
		})
	case "data":
		steps = append(steps, BuildStep{
			Title:  m.ui.GetText("step.install_data"),
			Detail: m.ui.GetText("detail.data_stack"),
			Exec:   func(ctx context.Context) error { return nil },
		})
	case "sandbox":
		// Sandbox: only system base, no additional installations
	}

	// Final step: Export image and cleanup
	steps = append(steps, BuildStep{
		Title:  m.ui.GetText("step.export_image"),
		Detail: buildCtx.ImageName,
		Exec: func(ctx context.Context) error {
			if err := m.exportBuildImage(ctx, buildCtx); err != nil {
				return err
			}
			if err := DestroyBuildContainer(ctx, m.runtime, m.fs, containerName, buildCtx.BuildWorkspaceDir); err != nil {
				return err
			}
			cleanupDone = true
			return nil
		},
	})

	title := m.ui.GetText("build.title", buildCtx.ImageName)
	subtitle := m.ui.GetText("build.subtitle_base", buildCtx.GPUInfo.Type, gpuModeText)

	return &BuildPlan{
		Title:    title,
		Subtitle: subtitle,
		Steps:    steps,
		Cleanup: func() {
			if !cleanupDone {
				_ = m.runtime.RemoveBunker(ctx, containerName, true)
				_ = removePathWritable(m.fs, buildCtx.BuildWorkspaceDir)
			}
		},
		OnSuccess: func() {
			m.ui.ShowLog("build.success", buildCtx.ImageName)
		},
	}
}

// collectBuildItems converts slot selections into BuildItems for the installer.
func (m *Manager) collectBuildItems(selections []SlotSelection) []ports.BuildItem {
	var items []ports.BuildItem
	for _, sel := range selections {
		slotItems, err := m.slotManager.GetSelectedItems(sel.Slot)
		if err != nil {
			continue
		}
		for _, item := range slotItems {
			// Check if this item was selected
			selected := false
			for _, selID := range sel.Selected {
				if item.ID == selID {
					selected = true
					break
				}
			}
			if !selected {
				continue
			}
			items = append(items, ports.BuildItem{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				Category:    item.Category,
				Deps:        item.Deps,
				NeedsOllama: item.ID == "ollama" || item.Category == "ia",
			})
		}
	}
	return items
}

// Rebuild rebuilds an existing image after asking for confirmation.
// Returns a BuildPlan for the caller to execute with progress tracking.
func (m *Manager) Rebuild(ctx context.Context, cfg config.EnvConfig) (*BuildPlan, error) {
	// Use empty containerName to auto-generate based on slot (axiom-dev, axiom-data, axiom-sandbox)
	buildCtx, err := PrepareBuildContext(ctx, cfg, "", "dev", m.system, m.ui)
	if err != nil {
		return nil, err
	}

	targetImage := buildCtx.ImageName

	confirm, _ := m.ui.AskConfirmInCard(
		"rebuild",
		[]ports.Field{
			ports.NewField(m.ui.GetText("rebuild.image_label"), targetImage),
			ports.NewField(m.ui.GetText("rebuild.gpu_label"), buildCtx.GPUInfo.Type),
		},
		nil,
		"rebuild.confirm",
	)
	if !confirm {
		return nil, nil
	}

	// Remove existing image
	_ = m.runtime.RemoveImage(ctx, targetImage, true)

	// Return the build plan (same as Build)
	return m.Build(ctx, cfg)
}

// SaveSlotSelection saves slot selection for the build process.
func (m *Manager) SaveSlotSelection(selectedSlot string, selectedIDs []string) error {
	selections := []SlotSelection{{Slot: selectedSlot, Selected: selectedIDs}}
	return m.slotManager.SaveSelection(selections)
}

// exportBuildImage commits the container to an image using the runtime.
func (m *Manager) exportBuildImage(ctx context.Context, buildCtx *BuildContext) error {
	containerName := buildCtx.ContainerName

	// Get container state
	state, err := m.runtime.ContainerState(ctx, containerName)
	if err != nil {
		return fmt.Errorf("errors.build.export_failed_check_status: %w", err)
	}
	if state == "" {
		return fmt.Errorf("errors.build.export_container_not_found: %s", containerName)
	}

	// Check if container is running, if not start it
	if state != "running" {
		if err := m.runtime.StartContainer(ctx, containerName); err != nil {
			return fmt.Errorf("errors.build.export_failed_start_container: %w", err)
		}
	}

	// Commit the container to an image
	author := m.ui.GetText("build.commit.author")
	message := m.ui.GetText("build.commit.message")
	if err := m.runtime.CommitImage(ctx, containerName, buildCtx.ImageName, author, message); err != nil {
		return fmt.Errorf("errors.build.export_failed_commit: %w", err)
	}

	return nil
}
