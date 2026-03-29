package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"axiom/internal/domain"
	"axiom/internal/ports"
)

// Manager orchestrates build operations.
type Manager struct {
	runtime        ports.IBunkerRuntime
	fs             ports.IFileSystem
	ui             ports.IPresenter
	buildContainer string
	slotManager    SlotManagerInterface
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
func NewManager(runtime ports.IBunkerRuntime, fs ports.IFileSystem, ui ports.IPresenter, buildContainer string, slotManager SlotManagerInterface) *Manager {
	return &Manager{
		runtime:        runtime,
		fs:             fs,
		ui:             ui,
		buildContainer: buildContainer,
		slotManager:    slotManager,
	}
}

// Build executes the complete build flow for creating a base image.
// It integrates with the slot system to allow users to select which items to install.
func (m *Manager) Build(ctx context.Context, cfg domain.EnvConfig) error {
	// Check if slot selection exists
	selections, err := m.slotManager.LoadSelection()
	if err != nil {
		m.ui.ShowLog("warn", "Failed to load slot selection:", err.Error())
		selections = []SlotSelection{}
	}

	// If selections exist, ask user whether to use them or reselect
	if len(selections) > 0 && m.hasValidSelection(selections) {
		// Count total selected items
		totalItems := 0
		for _, sel := range selections {
			totalItems += len(sel.Selected)
		}

		m.ui.ShowLogo()
		m.ui.ShowCommandCard(
			"build",
			[]ports.Field{
				{Label: "slot_wizard.previous_selection_title", Value: m.ui.GetText("slot_wizard.previous_selection_desc")},
				{Label: "slot_wizard.items_selected", Value: fmt.Sprintf("%d", totalItems)},
			},
			nil,
		)

		usePrevious, err := m.ui.AskConfirm("slot_wizard.use_previous")
		if err != nil {
			return fmt.Errorf("failed to ask confirmation: %w", err)
		}

		if !usePrevious {
			// User wants to reselect
			m.ui.ShowLog("info", "Running slot selector...")
			selections, err = m.runSlotSelectionUI()
			if err != nil {
				return fmt.Errorf("slot selection failed: %w", err)
			}
			if len(selections) == 0 {
				return fmt.Errorf("build cancelled: no slot selection made")
			}
		}
	} else {
		// No selections exist, run the slot selector
		m.ui.ShowLog("info", "No slot selection found. Running slot selector...")
		selections, err = m.runSlotSelectionUI()
		if err != nil {
			return fmt.Errorf("slot selection failed: %w", err)
		}
		if len(selections) == 0 {
			return fmt.Errorf("build cancelled: no slot selection made")
		}
	}

	// Prepare build context
	buildCtx, err := PrepareBuildContext(ctx, cfg, m.buildContainer)
	if err != nil {
		return err
	}

	// Initialize progress display
	progress := m.newBuildProgress(buildCtx)
	progress.render()

	// Step 0: Prepare shared directories
	if err := progress.RunStep(0, func() error {
		return PrepareSharedDirectories(ctx, m.fs, buildCtx.Config)
	}); err != nil {
		progress.renderErrorWithContext(err, buildCtx.Config.AIConfigDir())
		return err
	}

	// Step 1: Recreate build container
	if err := progress.RunStep(1, func() error {
		return RecreateBuildContainer(ctx, m.runtime, m.fs, m.buildContainer, buildCtx.BuildWorkspaceDir, buildCtx.Config)
	}); err != nil {
		progress.renderErrorWithContext(err, buildCtx.BuildWorkspaceDir)
		return err
	}

	// Track cleanup
	cleanupDone := false
	defer func() {
		if cleanupDone {
			return
		}
		_ = m.runtime.RemoveBunker(ctx, m.buildContainer, true)
		_ = removePathWritable(m.fs, buildCtx.BuildWorkspaceDir)
	}()

	// Step 2: Install system base
	if err := progress.RunStep(2, func() error {
		return m.installSystemBase(ctx, buildCtx)
	}); err != nil {
		progress.renderErrorWithContext(err, buildCtx.BuildWorkspaceDir)
		return err
	}

	// Step 3: Install developer tools (using slot selection)
	if err := progress.RunStep(3, func() error {
		return m.installSlotItems(ctx, buildCtx, "dev", selections)
	}); err != nil {
		progress.renderErrorWithContext(err, buildCtx.BuildWorkspaceDir)
		return err
	}

	// Step 4: Install data stack (using slot selection)
	if err := progress.RunStep(4, func() error {
		return m.installSlotItems(ctx, buildCtx, "data", selections)
	}); err != nil {
		progress.renderErrorWithContext(err, buildCtx.BuildWorkspaceDir)
		return err
	}

	// Step 5: Export image and cleanup
	if err := progress.RunStep(5, func() error {
		if err := m.exportBuildImage(ctx, buildCtx); err != nil {
			return err
		}
		if err := DestroyBuildContainer(ctx, m.runtime, m.fs, m.buildContainer, buildCtx.BuildWorkspaceDir); err != nil {
			return err
		}
		cleanupDone = true
		return nil
	}); err != nil {
		progress.renderErrorWithContext(err, buildCtx.BuildWorkspaceDir)
		return err
	}

	progress.subtitle = m.ui.GetText("build.success_sub", buildCtx.ImageName)
	progress.render()
	m.ui.ShowLog("build.success", buildCtx.ImageName)
	return nil
}

// hasValidSelection checks if there is at least one slot with selected items.
func (m *Manager) hasValidSelection(selections []SlotSelection) bool {
	for _, sel := range selections {
		if len(sel.Selected) > 0 {
			return true
		}
	}
	return false
}

// runSlotSelectionUI presents the slot selection UI to the user.
func (m *Manager) runSlotSelectionUI() ([]SlotSelection, error) {
	// For now, run DEV slot selection first
	// In a full implementation, this would cycle through all categories
	items, err := m.slotManager.GetSelectedItems("dev")
	if err != nil {
		return nil, fmt.Errorf("failed to get DEV slot items: %w", err)
	}

	// Get preselected items from config
	var preselected []string
	selections, _ := m.slotManager.LoadSelection()
	for _, sel := range selections {
		if sel.Slot == "dev" {
			preselected = sel.Selected
			break
		}
	}

	// Run the slot selector
	selectedIDs, confirmed, err := m.slotManager.RunSlotSelector("dev", items, preselected)
	if err != nil {
		return nil, fmt.Errorf("slot selector failed: %w", err)
	}
	if !confirmed {
		return nil, fmt.Errorf("slot selection cancelled")
	}

	// Return as SlotSelection
	return []SlotSelection{
		{Slot: "dev", Selected: selectedIDs},
	}, nil
}

// installSlotItems installs items from a specific slot category using the slot manager.
func (m *Manager) installSlotItems(ctx context.Context, buildCtx *BuildContext, category string, selections []SlotSelection) error {
	// Find selection for this category
	var categorySel SlotSelection
	found := false
	for _, sel := range selections {
		if sel.Slot == category {
			categorySel = sel
			found = true
			break
		}
	}
	if !found || len(categorySel.Selected) == 0 {
		m.ui.ShowLog("info", "No items selected for", category, "slot")
		return nil
	}

	// Get items for this category
	items, err := m.slotManager.GetSelectedItems(category)
	if err != nil {
		return fmt.Errorf("failed to get %s items: %w", category, err)
	}

	// Filter to only selected items
	var selectedItems []SlotItem
	for _, item := range items {
		for _, selID := range categorySel.Selected {
			if item.ID == selID {
				selectedItems = append(selectedItems, item)
				break
			}
		}
	}

	if len(selectedItems) == 0 {
		return nil
	}

	// Create exec function for running commands in container
	exec := func(ctx context.Context, cmd string, args ...string) error {
		allArgs := []string{cmd}
		allArgs = append(allArgs, args...)
		return RunInContainer(ctx, m.runtime, m.buildContainer, allArgs...)
	}

	// Install each selected item
	for _, item := range selectedItems {
		m.ui.ShowLog("info", "Installing:", item.Name)

		// For now, we just show the installation
		// In a full implementation, the item.Executor would be called
		// This is where the actual installation logic would go
		_ = exec // Will be used when items have actual executors
	}

	return nil
}

// Rebuild rebuilds an existing image after asking for confirmation.
func (m *Manager) Rebuild(ctx context.Context, cfg domain.EnvConfig) error {
	buildCtx, err := PrepareBuildContext(ctx, cfg, m.buildContainer)
	if err != nil {
		return err
	}

	targetImage := buildCtx.ImageName

	confirm, _ := m.ui.AskConfirmInCard(
		"rebuild",
		[]ports.Field{
			{Label: "Imagen", Value: targetImage},
			{Label: "GPU", Value: buildCtx.GPUInfo.Type},
		},
		nil,
		"rebuild.confirm",
	)
	if !confirm {
		return nil
	}

	steps := []ports.LifecycleStep{
		{Title: m.ui.GetText("rebuild.step_rm_image"), Detail: targetImage, Status: ports.LifecycleRunning},
	}
	m.ui.ClearScreen()
	m.ui.ShowLogo()
	m.ui.RenderLifecycle(m.ui.GetText("rebuild.title"), m.ui.GetText("rebuild.subtitle"), steps, "", nil)

	// Remove image via host command
	_ = m.runtime.RemoveImage(ctx, targetImage, true)
	steps[0].Status = ports.LifecycleDone
	m.ui.ClearScreen()
	m.ui.ShowLogo()
	m.ui.RenderLifecycle(m.ui.GetText("rebuild.title"), m.ui.GetText("rebuild.subtitle"), steps, "", nil)

	return m.Build(ctx, cfg)
}

// newBuildProgress creates a progress tracker for the build.
func (m *Manager) newBuildProgress(ctx *BuildContext) *Progress {
	gpuModeText := m.ui.GetText("build.subtitle_host")
	if ctx.Config.ROCMMode == "image" {
		gpuModeText = m.ui.GetText("build.subtitle_image")
	}

	steps := []ports.LifecycleStep{
		{Title: m.ui.GetText("step.prepare_dirs"), Detail: ctx.Config.AIConfigDir(), Status: ports.LifecyclePending},
		{Title: m.ui.GetText("step.recreate_container"), Detail: ctx.BuildWorkspaceDir, Status: ports.LifecyclePending},
		{Title: m.ui.GetText("step.install_base"), Detail: m.ui.GetText("detail.base_pkgs"), Status: ports.LifecyclePending},
		{Title: m.ui.GetText("step.install_dev"), Detail: m.ui.GetText("detail.dev_tools"), Status: ports.LifecyclePending},
		{Title: m.ui.GetText("step.install_ai"), Detail: m.ui.GetText("detail.ai_stack"), Status: ports.LifecyclePending},
		{Title: m.ui.GetText("step.export_image"), Detail: ctx.ImageName, Status: ports.LifecyclePending},
	}

	title := m.ui.GetText("build.title", ctx.ImageName)
	subtitle := m.ui.GetText("build.subtitle_base", ctx.GPUInfo.Type, gpuModeText)

	return NewProgress(m.ui, title, subtitle, steps)
}

// installSystemBase delegates to the exported function with proper executor.
func (m *Manager) installSystemBase(ctx context.Context, buildCtx *BuildContext) error {
	exec := func(ctx context.Context, cmd string, args ...string) error {
		allArgs := []string{cmd}
		allArgs = append(allArgs, args...)
		return RunInContainer(ctx, m.runtime, m.buildContainer, allArgs...)
	}
	return InstallSystemBase(ctx, m.buildContainer, buildCtx, m.ui, exec)
}

// installDeveloperTools delegates to the exported function.
func (m *Manager) installDeveloperTools(ctx context.Context, buildCtx *BuildContext) error {
	exec := func(ctx context.Context, cmd string, args ...string) error {
		allArgs := []string{cmd}
		allArgs = append(allArgs, args...)
		return RunInContainer(ctx, m.runtime, m.buildContainer, allArgs...)
	}
	return InstallDeveloperTools(ctx, m.buildContainer, buildCtx, m.ui, exec)
}

// installModelStack delegates to the exported function.
func (m *Manager) installModelStack(ctx context.Context, buildCtx *BuildContext) error {
	exec := func(ctx context.Context, cmd string, args ...string) error {
		allArgs := []string{cmd}
		allArgs = append(allArgs, args...)
		return RunInContainer(ctx, m.runtime, m.buildContainer, allArgs...)
	}
	modelConfig := ModelStackConfig{GPUType: buildCtx.GPUInfo.Type}
	return InstallModelStack(ctx, m.buildContainer, buildCtx, modelConfig, m.ui, exec)
}

// exportBuildImage commits the container to an image using podman commit.
// Note: This uses a host command since IBunkerRuntime doesn't have Commit method.
func (m *Manager) exportBuildImage(ctx context.Context, buildCtx *BuildContext) error {
	// Use podman commit to save the container as an image
	cmd := exec.CommandContext(ctx, "podman", "commit",
		"-a", "axiom",
		"-m", "AXIOM build image",
		m.buildContainer,
		buildCtx.ImageName,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("export_image: failed to commit container: %w", err)
	}

	return nil
}
