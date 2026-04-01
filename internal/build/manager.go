package build

import (
	"context"
	"fmt"

	"github.com/Alejandro-M-P/AXIOM/internal/domain"
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
func NewManager(runtime ports.IBunkerRuntime, fs ports.IFileSystem, ui ports.IPresenter, system ports.ISystem, buildContainer string, slotManager SlotManagerInterface) *Manager {
	return &Manager{
		runtime:        runtime,
		fs:             fs,
		ui:             ui,
		system:         system,
		buildContainer: buildContainer,
		slotManager:    slotManager,
	}
}

// Build executes the complete build flow for creating a base image.
// It integrates with the slot system to allow users to select which items to install.
func (m *Manager) Build(ctx context.Context, cfg domain.EnvConfig) error {
	// Load the slot selection that was saved by the router
	selections, err := m.slotManager.LoadSelection()
	if err != nil {
		return fmt.Errorf("errors.build.failed_load_selection: %w", err)
	}
	if len(selections) == 0 {
		return fmt.Errorf("errors.build.cancelled_no_selection")
	}

	// Determine slot name from selection
	slotName := "dev"
	if len(selections) > 0 && selections[0].Slot != "" {
		slotName = selections[0].Slot
	}

	// Prepare build context - use slot name for image
	// Pass empty containerName to auto-generate based on slot (axiom-dev, axiom-data, axiom-sandbox)
	var imageName string
	buildCtx, err := PrepareBuildContext(ctx, cfg, "", slotName, m.system)
	if err != nil {
		return err
	}

	// Initialize progress display
	progress := m.newBuildProgress(buildCtx, slotName)
	progress.render()

	// Check if the slot image already exists (AFTER Progress UI is shown)
	imageName = buildCtx.ImageName
	exists, err := m.runtime.ImageExists(ctx, imageName)
	if err != nil {
		m.ui.ShowLog("build.image_check_failed", err.Error())
	}
	if exists {
		confirm, err := m.ui.AskConfirmInCard(
			"build",
			[]ports.Field{
				{Label: m.ui.GetText("build.image_exists_title"), Value: imageName},
				{Label: m.ui.GetText("build.image_exists_warning"), Value: m.ui.GetText("build.image_exists_desc")},
			},
			nil,
			"build.image_exists_confirm",
		)
		if err != nil {
			return fmt.Errorf("errors.build.failed_ask_confirmation: %w", err)
		}
		if !confirm {
			return fmt.Errorf("errors.build.cancelled_image_exists")
		}
		// Delete existing image
		if err := m.runtime.RemoveImage(ctx, imageName, true); err != nil {
			return fmt.Errorf("errors.build.failed_remove_image: %w", err)
		}
	}

	// Step 0: Prepare shared directories
	if err := progress.RunStep(0, func() error {
		return PrepareSharedDirectories(ctx, m.fs, buildCtx.Config)
	}); err != nil {
		progress.renderErrorWithContext(err, buildCtx.Config.AIConfigDir())
		return err
	}

	// Step 1: Recreate build container
	containerName := buildCtx.ContainerName
	if err := progress.RunStep(1, func() error {
		return RecreateBuildContainer(ctx, m.runtime, m.fs, containerName, buildCtx.BuildWorkspaceDir, buildCtx.Config)
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
		_ = m.runtime.RemoveBunker(ctx, containerName, true)
		_ = removePathWritable(m.fs, buildCtx.BuildWorkspaceDir)
	}()

	// Step 2: Install system base (always needed)
	if err := progress.RunStep(2, func() error {
		return m.installSystemBase(ctx, buildCtx)
	}); err != nil {
		progress.renderErrorWithContext(err, buildCtx.BuildWorkspaceDir)
		return err
	}

	// Step 3-4: Install slot-specific items based on selection
	// DEV: install dev tools and AI stack
	// DATA: install data stack (databases)
	// SANDBOX: no additional installations
	switch slotName {
	case "dev":
		// Install dev tools
		if err := progress.RunStep(3, func() error {
			return m.installSlotItems(ctx, buildCtx, "dev", selections)
		}); err != nil {
			progress.renderErrorWithContext(err, buildCtx.BuildWorkspaceDir)
			return err
		}
		// Install AI stack
		if err := progress.RunStep(4, func() error {
			return m.installModelStack(ctx, buildCtx)
		}); err != nil {
			progress.renderErrorWithContext(err, buildCtx.BuildWorkspaceDir)
			return err
		}
	case "data":
		// Install data stack (databases)
		if err := progress.RunStep(3, func() error {
			return m.installSlotItems(ctx, buildCtx, "data", selections)
		}); err != nil {
			progress.renderErrorWithContext(err, buildCtx.BuildWorkspaceDir)
			return err
		}
		// No step 4 for data - will export directly
	case "sandbox":
		// Sandbox: only system base, skip steps 3 and 4
		// No additional installations
	}

	// Calculate export step index (varies by slot)
	exportStep := 3
	if slotName == "dev" {
		exportStep = 5
	} else if slotName == "data" {
		exportStep = 4
	}

	// Export image and cleanup (containerName already defined above)
	if err := progress.RunStep(exportStep, func() error {
		if err := m.exportBuildImage(ctx, buildCtx); err != nil {
			return err
		}
		if err := DestroyBuildContainer(ctx, m.runtime, m.fs, containerName, buildCtx.BuildWorkspaceDir); err != nil {
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

// runSlotSelectionUI presents the slot selection UI to the user.
func (m *Manager) runSlotSelectionUI() ([]SlotSelection, error) {
	// For now, run DEV slot selection first
	// In a full implementation, this would cycle through all categories
	items, err := m.slotManager.GetSelectedItems("dev")
	if err != nil {
		return nil, fmt.Errorf("errors.build.failed_get_dev_items: %w", err)
	}

	// Run the slot selector (no preselection from previous runs)
	selectedIDs, confirmed, err := m.slotManager.RunSlotSelector("dev", items, nil)
	if err != nil {
		return nil, fmt.Errorf("errors.build.slot_selector_failed: %w", err)
	}
	if !confirmed {
		return nil, fmt.Errorf("errors.build.slot_selection_cancelled")
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
		m.ui.ShowLog("build.no_items_selected", category)
		return nil
	}

	// Get items for this category
	items, err := m.slotManager.GetSelectedItems(category)
	if err != nil {
		return fmt.Errorf("errors.build.failed_get_items: %w", err)
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
	containerName := buildCtx.ContainerName
	exec := func(ctx context.Context, cmd string, args ...string) error {
		allArgs := []string{cmd}
		allArgs = append(allArgs, args...)
		return RunInContainer(ctx, m.runtime, containerName, allArgs...)
	}

	// Install each selected item
	for _, item := range selectedItems {
		m.ui.ShowLog("build.installing_item", item.Name)

		// For now, we just show the installation
		// In a full implementation, the item.Executor would be called
		// This is where the actual installation logic would go
		_ = exec // Will be used when items have actual executors
	}

	return nil
}

// Rebuild rebuilds an existing image after asking for confirmation.
func (m *Manager) Rebuild(ctx context.Context, cfg domain.EnvConfig) error {
	// Use empty containerName to auto-generate based on slot (axiom-dev, axiom-data, axiom-sandbox)
	buildCtx, err := PrepareBuildContext(ctx, cfg, "", "dev", m.system)
	if err != nil {
		return err
	}

	targetImage := buildCtx.ImageName

	confirm, _ := m.ui.AskConfirmInCard(
		"rebuild",
		[]ports.Field{
			{Label: m.ui.GetText("rebuild.image_label"), Value: targetImage},
			{Label: m.ui.GetText("rebuild.gpu_label"), Value: buildCtx.GPUInfo.Type},
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
// It generates different steps based on the slot type (dev, data, sandbox).
func (m *Manager) newBuildProgress(ctx *BuildContext, slotName string) *Progress {
	gpuModeText := m.ui.GetText("build.subtitle_host")
	if ctx.Config.ROCMMode == "image" {
		gpuModeText = m.ui.GetText("build.subtitle_image")
	}

	// Build steps dynamically based on slot type
	var steps []ports.LifecycleStep

	// Common steps: prepare dirs, recreate container, system base
	steps = append(steps, ports.LifecycleStep{
		Title: m.ui.GetText("step.prepare_dirs"), Detail: ctx.Config.AIConfigDir(), Status: ports.LifecyclePending})
	steps = append(steps, ports.LifecycleStep{
		Title: m.ui.GetText("step.recreate_container"), Detail: ctx.BuildWorkspaceDir, Status: ports.LifecyclePending})
	steps = append(steps, ports.LifecycleStep{
		Title: m.ui.GetText("step.install_base"), Detail: m.ui.GetText("detail.base_pkgs"), Status: ports.LifecyclePending})

	// Slot-specific steps
	switch slotName {
	case "dev":
		// Dev: install dev tools + AI stack
		steps = append(steps, ports.LifecycleStep{
			Title: m.ui.GetText("step.install_dev"), Detail: m.ui.GetText("detail.dev_tools"), Status: ports.LifecyclePending})
		steps = append(steps, ports.LifecycleStep{
			Title: m.ui.GetText("step.install_ai"), Detail: m.ui.GetText("detail.ai_stack"), Status: ports.LifecyclePending})
	case "data":
		// Data: install data stack (databases)
		steps = append(steps, ports.LifecycleStep{
			Title: m.ui.GetText("step.install_data"), Detail: m.ui.GetText("detail.data_stack"), Status: ports.LifecyclePending})
	case "sandbox":
		// Sandbox: only system base, no additional installations
		// No extra steps - just base system
	}

	// Always: export image
	steps = append(steps, ports.LifecycleStep{
		Title: m.ui.GetText("step.export_image"), Detail: ctx.ImageName, Status: ports.LifecyclePending})

	title := m.ui.GetText("build.title", ctx.ImageName)
	subtitle := m.ui.GetText("build.subtitle_base", ctx.GPUInfo.Type, gpuModeText)

	return NewProgress(m.ui, title, subtitle, steps)
}

// installSystemBase delegates to the exported function with proper executor.
func (m *Manager) installSystemBase(ctx context.Context, buildCtx *BuildContext) error {
	containerName := buildCtx.ContainerName
	exec := func(ctx context.Context, cmd string, args ...string) error {
		allArgs := []string{cmd}
		allArgs = append(allArgs, args...)
		return RunInContainer(ctx, m.runtime, containerName, allArgs...)
	}
	return InstallSystemBase(ctx, containerName, buildCtx, m.ui, exec)
}

// installDeveloperTools delegates to the exported function.
func (m *Manager) installDeveloperTools(ctx context.Context, buildCtx *BuildContext) error {
	containerName := buildCtx.ContainerName
	exec := func(ctx context.Context, cmd string, args ...string) error {
		allArgs := []string{cmd}
		allArgs = append(allArgs, args...)
		return RunInContainer(ctx, m.runtime, containerName, allArgs...)
	}
	return InstallDeveloperTools(ctx, containerName, buildCtx, m.ui, exec)
}

// installModelStack delegates to the exported function.
func (m *Manager) installModelStack(ctx context.Context, buildCtx *BuildContext) error {
	containerName := buildCtx.ContainerName
	exec := func(ctx context.Context, cmd string, args ...string) error {
		allArgs := []string{cmd}
		allArgs = append(allArgs, args...)
		return RunInContainer(ctx, m.runtime, containerName, allArgs...)
	}
	modelConfig := ModelStackConfig{GPUType: buildCtx.GPUInfo.Type}
	return InstallModelStack(ctx, containerName, buildCtx, modelConfig, m.ui, exec)
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
