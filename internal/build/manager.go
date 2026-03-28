package build

import (
	"context"
	"fmt"

	"axiom/internal/domain"
	"axiom/internal/ports"
)

// Manager orchestrates build operations.
type Manager struct {
	runtime        ports.IBunkerRuntime
	fs             ports.IFileSystem
	ui             ports.IPresenter
	buildContainer string
}

// NewManager creates a new build manager.
func NewManager(runtime ports.IBunkerRuntime, fs ports.IFileSystem, ui ports.IPresenter, buildContainer string) *Manager {
	return &Manager{
		runtime:        runtime,
		fs:             fs,
		ui:             ui,
		buildContainer: buildContainer,
	}
}

// Build executes the complete build flow for creating a base image.
func (m *Manager) Build(ctx context.Context, cfg domain.EnvConfig) error {
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

	// Step 3: Install developer tools
	if err := progress.RunStep(3, func() error {
		return m.installDeveloperTools(ctx, buildCtx)
	}); err != nil {
		progress.renderErrorWithContext(err, buildCtx.BuildWorkspaceDir)
		return err
	}

	// Step 4: Install model stack
	if err := progress.RunStep(4, func() error {
		return m.installModelStack(ctx, buildCtx)
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

// exportBuildImage commits the container to an image.
// Note: This requires a host command since IBunkerRuntime doesn't have Commit.
func (m *Manager) exportBuildImage(ctx context.Context, buildCtx *BuildContext) error {
	// TODO: This needs a host command approach since IBunkerRuntime doesn't have CommitContainer
	// For now, we'll use a placeholder that indicates this needs implementation
	return fmt.Errorf("export_image: IBunkerRuntime does not support image commit - requires host command")
}
