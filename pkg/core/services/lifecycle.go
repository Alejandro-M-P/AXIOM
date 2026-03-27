package bunker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"axiom/pkg/adapters/system/gpu"
)

type buildContext struct {
	config            EnvConfig
	gpuInfo           gpu.GPUInfo
	imageName         string
	buildWorkspaceDir string
}

// buildProgress mantiene el estado visual del build para la UI.
type buildProgress struct {
	ui        UI
	title     string
	subtitle  string
	steps     []LifecycleStep
	taskTitle string
	taskSteps []LifecycleStep
}

// Build ejecuta el flujo completo de construccion de la imagen base.
func (m *Manager) Build() error {
	ctx, err := m.prepareBuildContext()
	if err != nil {
		return err
	}

	progress := newBuildProgress(ctx, m.UI)
	progress.render()

	if err := progress.runStep(0, func() error {
		return m.prepareSharedDirectories(ctx)
	}); err != nil {
		progress.renderError(err, ctx.config.AIConfigDir())
		return err
	}

	if err := progress.runStep(1, func() error {
		return m.recreateBuildContainer(ctx)
	}); err != nil {
		progress.renderError(err, ctx.buildWorkspaceDir)
		return err
	}

	cleanupDone := false
	defer func() {
		if cleanupDone {
			return
		}
		_ = runCommandQuiet("distrobox-rm", m.buildContainerName, "--force", "--yes")
		_ = removePathWritable(ctx.buildWorkspaceDir)
	}()

	if err := progress.runStep(2, func() error {
		return m.installSystemBase(ctx, progress)
	}); err != nil {
		progress.renderError(err, ctx.buildWorkspaceDir)
		return err
	}

	if err := progress.runStep(3, func() error {
		return m.installDeveloperTools(ctx, progress)
	}); err != nil {
		progress.renderError(err, ctx.buildWorkspaceDir)
		return err
	}

	if err := progress.runStep(4, func() error {
		return m.installModelStack(ctx, progress)
	}); err != nil {
		progress.renderError(err, ctx.buildWorkspaceDir)
		return err
	}

	if err := progress.runStep(5, func() error {
		if err := m.exportBuildImage(ctx); err != nil {
			return err
		}
		if err := m.destroyBuildContainer(ctx); err != nil {
			return err
		}
		cleanupDone = true
		return nil
	}); err != nil {
		progress.renderError(err, ctx.buildWorkspaceDir)
		return err
	}

	progress.subtitle = m.UI.GetText("build.success_sub", ctx.imageName)
	progress.render()
	m.UI.ShowLog("build.success", ctx.imageName)
	return nil
}

func (m *Manager) prepareBuildContext() (buildContext, error) {
	cfg, err := m.LoadConfig()
	if err != nil {
		return buildContext{}, fmt.Errorf("env_read_error: %w", err)
	}

	hardware := resolveBuildGPU(cfg)
	imageName := baseImageName(hardware.Type)

	return buildContext{
		config:            cfg,
		gpuInfo:           hardware,
		imageName:         imageName,
		buildWorkspaceDir: cfg.BuildWorkspaceDir(m.buildContainerName),
	}, nil
}

func newBuildProgress(ctx buildContext, ui UI) *buildProgress {
	gpuModeText := ui.GetText("build.subtitle_host")
	if ctx.config.ROCMMode == "image" {
		gpuModeText = ui.GetText("build.subtitle_image")
	}

	return &buildProgress{
		ui:       ui,
		title:    ui.GetText("build.title", ctx.imageName),
		subtitle: ui.GetText("build.subtitle_base", ctx.gpuInfo.Type, gpuModeText),
		steps: []LifecycleStep{
			{Title: ui.GetText("step.prepare_dirs"), Detail: ctx.config.AIConfigDir(), Status: LifecyclePending},
			{Title: ui.GetText("step.recreate_container"), Detail: ctx.buildWorkspaceDir, Status: LifecyclePending},
			{Title: ui.GetText("step.install_base"), Detail: ui.GetText("detail.base_pkgs"), Status: LifecyclePending},
			{Title: ui.GetText("step.install_dev"), Detail: ui.GetText("detail.dev_tools"), Status: LifecyclePending},
			{Title: ui.GetText("step.install_ai"), Detail: ui.GetText("detail.ai_stack"), Status: LifecyclePending},
			{Title: ui.GetText("step.export_image"), Detail: ctx.imageName, Status: LifecyclePending},
		},
	}
}

func (p *buildProgress) runStep(index int, fn func() error) error {
	p.taskTitle = ""
	p.taskSteps = nil
	for i := range p.steps {
		if i < index && p.steps[i].Status != LifecycleDone {
			p.steps[i].Status = LifecycleDone
		}
		if i == index {
			p.steps[i].Status = LifecycleRunning
		}
	}
	p.render()

	if err := fn(); err != nil {
		p.steps[index].Status = LifecycleError
		return err
	}

	p.steps[index].Status = LifecycleDone
	p.taskTitle = ""
	p.taskSteps = nil
	p.render()
	return nil
}

func (p *buildProgress) startTaskGroup(title string, steps []LifecycleStep) {
	p.taskTitle = title
	p.taskSteps = steps
	p.render()
}

func (p *buildProgress) runTask(index int, fn func() error) error {
	for i := range p.taskSteps {
		if i < index && p.taskSteps[i].Status != LifecycleDone {
			p.taskSteps[i].Status = LifecycleDone
		}
		if i == index {
			p.taskSteps[i].Status = LifecycleRunning
		}
	}
	p.render()

	if err := fn(); err != nil {
		if index >= 0 && index < len(p.taskSteps) {
			p.taskSteps[index].Status = LifecycleError
		}
		return err
	}

	if index >= 0 && index < len(p.taskSteps) {
		p.taskSteps[index].Status = LifecycleDone
	}
	p.render()
	return nil
}

func (p *buildProgress) render() {
	p.ui.ClearScreen()
	p.ui.ShowLogo()
	p.ui.RenderLifecycle(p.title, p.subtitle, p.steps, p.taskTitle, p.taskSteps)
}

func (p *buildProgress) renderError(err error, where string) {
	p.ui.ClearScreen()
	p.ui.ShowLogo()
	p.ui.RenderLifecycleError(p.title, p.steps, p.taskTitle, p.taskSteps, err, where)
}

func (m *Manager) prepareSharedDirectories(ctx buildContext) error {
	if err := os.MkdirAll(filepath.Join(ctx.config.AIConfigDir(), "models"), 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(ctx.config.AIConfigDir(), "teams"), 0700); err != nil {
		return err
	}
	if err := ensureTutorFile(ctx.config.TutorPath()); err != nil {
		return err
	}
	return runCommandQuiet("sudo", "chown", "-R", currentUserGroup(), ctx.config.AIConfigDir())
}

func (m *Manager) recreateBuildContainer(ctx buildContext) error {
	_ = runCommandQuiet("distrobox-rm", m.buildContainerName, "--force", "--yes")
	if err := removePathWritable(ctx.buildWorkspaceDir); err != nil {
		return err
	}
	if err := os.MkdirAll(ctx.buildWorkspaceDir, 0700); err != nil {
		return err
	}

	flags := m.buildContainerFlags(ctx)
	return runCommandQuiet(
		"distrobox-create",
		"--name", m.buildContainerName,
		"--image", "archlinux:latest",
		"--home", ctx.buildWorkspaceDir,
		"--additional-flags", flags,
		"--yes",
	)
}

func (m *Manager) buildContainerFlags(ctx buildContext) string {
	return fmt.Sprintf(
		"--volume %s:/ai_config:z --volume %s:/run/axiom/env:ro,z --device /dev/kfd --device /dev/dri --security-opt label=disable --group-add video --group-add render",
		ctx.config.AIConfigDir(),
		filepath.Join(ctx.config.AxiomPath, ".env"),
	)
}

func (m *Manager) installSystemBase(ctx buildContext, progress *buildProgress) error {
	packages := []string{"base-devel", "git", "curl", "jq", "wget", "nodejs", "npm", "go", "fzf", "starship"}
	if ctx.config.ROCMMode == "image" {
		switch {
		case ctx.gpuInfo.Type == "nvidia":
			packages = append(packages, "nvidia-utils", "cuda")
		case strings.HasPrefix(ctx.gpuInfo.Type, "rdna"), ctx.gpuInfo.Type == "amd":
			packages = append(packages, "rocm-hip-sdk")
		case ctx.gpuInfo.Type == "intel":
			packages = append(packages, "intel-compute-runtime", "onevpl-intel-gpu")
		}
	}

	progress.startTaskGroup(m.UI.GetText("group.base"), []LifecycleStep{
		{Title: m.UI.GetText("task.sync_repos"), Detail: m.UI.GetText("detail.sync_cmd"), Status: LifecyclePending},
		{Title: m.UI.GetText("task.install_pkgs"), Detail: m.UI.GetText("detail.pkgs_count", len(packages)), Status: LifecyclePending},
	})
	if err := progress.runTask(0, func() error {
		return m.runInContainer("sudo", "pacman", "-Sy", "--noconfirm")
	}); err != nil {
		return err
	}
	if err := progress.runTask(1, func() error {
		args := append([]string{"sudo", "pacman", "-S", "--needed", "--noconfirm"}, packages...)
		return m.runInContainer(args...)
	}); err != nil {
		return err
	}
	if err := progress.runTask(2, func() error {
		cmd := "NONINTERACTIVE=1 /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
		return m.runInContainer("bash", "-c", cmd)
	}); err != nil {
		return err
	}
	return nil
}

func (m *Manager) installDeveloperTools(ctx buildContext, progress *buildProgress) error {
	progress.startTaskGroup(m.UI.GetText("group.dev"), []LifecycleStep{
		{Title: m.UI.GetText("task.install_opencode"), Detail: m.UI.GetText("detail.npm_global"), Status: LifecyclePending},
		{Title: "Configurando Homebrew Tap", Detail: "Gentleman-Programming/homebrew-tap", Status: LifecyclePending},
		{Title: "Instalando herramientas de desarrollo", Detail: "brew install engram gentle-ai", Status: LifecyclePending},
	})

	if err := progress.runTask(0, func() error {
		return m.runInContainer("sudo", "npm", "install", "-g", "opencode-ai")
	}); err != nil {
		return err
	}

	brewPath := "/home/linuxbrew/.linuxbrew/bin/brew"

	if err := progress.runTask(1, func() error {
		return m.runInContainer(brewPath, "tap", "Gentleman-Programming/homebrew-tap")
	}); err != nil {
		return err
	}

	if err := progress.runTask(2, func() error {
		return m.runInContainer(brewPath, "install", "engram", "gentle-ai")
	}); err != nil {
		return err
	}

	return nil
}

func (m *Manager) installModelStack(ctx buildContext, progress *buildProgress) error {
	progress.startTaskGroup(m.UI.GetText("group.ai"), []LifecycleStep{
		{Title: m.UI.GetText("task.install_ollama"), Detail: ctx.gpuInfo.Type, Status: LifecyclePending},
		{Title: m.UI.GetText("task.clean_caches"), Detail: m.UI.GetText("detail.tmp_pacman"), Status: LifecyclePending},
	})
	if err := progress.runTask(0, func() error {
		return m.installOllama(ctx.gpuInfo.Type)
	}); err != nil {
		return err
	}
	if err := progress.runTask(1, func() error {
		return m.cleanBuildCaches()
	}); err != nil {
		return err
	}
	return nil
}

func (m *Manager) installOllama(gpuType string) error {
	arch, err := ollamaArch()
	if err != nil {
		return err
	}

	mainURL := fmt.Sprintf("https://ollama.com/download/ollama-linux-%s.tar.zst", arch)
	if err := m.runInContainer("curl", "-fsSL", mainURL, "-o", "/tmp/ollama.tar.zst"); err != nil {
		return err
	}
	if err := m.runInContainer("sudo", "tar", "--zstd", "-xf", "/tmp/ollama.tar.zst", "-C", "/usr"); err != nil {
		return err
	}

	if strings.HasPrefix(gpuType, "rdna") || gpuType == "amd" {
		rocmURL := fmt.Sprintf("https://ollama.com/download/ollama-linux-%s-rocm.tar.zst", arch)
		if err := m.runInContainer("curl", "-fsSL", rocmURL, "-o", "/tmp/ollama-rocm.tar.zst"); err != nil {
			return err
		}
		if err := m.runInContainer("sudo", "tar", "--zstd", "-xf", "/tmp/ollama-rocm.tar.zst", "-C", "/usr"); err != nil {
			return err
		}
	}

	return nil
}

func ollamaArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported_arch")
	}
}

func (m *Manager) cleanBuildCaches() error {
	if err := m.runInContainer("sudo", "pacman", "-Scc", "--noconfirm"); err != nil {
		return err
	}

	homeDir, err := m.runInContainerOutput("printenv", "HOME")
	if err == nil && strings.TrimSpace(homeDir) != "" {
		homeDir = strings.TrimSpace(homeDir)
		cacheDir := filepath.ToSlash(filepath.Join(homeDir, ".cache"))
		cacheGoDir := filepath.ToSlash(filepath.Join(cacheDir, "go"))
		_ = m.runInContainer("chmod", "-R", "+w", cacheGoDir, cacheDir)
		_ = m.runInContainer("sudo", "rm", "-rf", "/tmp/agent-teams", "/tmp/ollama.tar.zst", "/tmp/ollama-rocm.tar.zst", cacheGoDir, "/var/cache/pacman/pkg")
		return nil
	}

	_ = m.runInContainer("sudo", "rm", "-rf", "/tmp/agent-teams", "/tmp/ollama.tar.zst", "/tmp/ollama-rocm.tar.zst", "/var/cache/pacman/pkg")
	return nil
}

func (m *Manager) exportBuildImage(ctx buildContext) error {
	return runCommandQuiet("podman", "commit", m.buildContainerName, ctx.imageName)
}

func (m *Manager) destroyBuildContainer(ctx buildContext) error {
	if err := runCommandQuiet("distrobox-rm", m.buildContainerName, "--force", "--yes"); err != nil {
		return err
	}
	return removePathWritable(ctx.buildWorkspaceDir)
}

func (m *Manager) runInContainer(args ...string) error {
	containerArgs := append([]string{"-n", m.buildContainerName, "--"}, args...)
	return runCommandQuiet("distrobox-enter", containerArgs...)
}

func (m *Manager) runInteractiveInContainer(input string, args ...string) error {
	containerArgs := append([]string{"-n", m.buildContainerName, "--"}, args...)
	return m.Runtime.RunCommandWithInput("", input, "distrobox-enter", containerArgs...)
}

func (m *Manager) runInContainerOutput(args ...string) (string, error) {
	return m.Runtime.RunCommandOutput(m.buildContainerName, args...)
}

func resolveBuildGPU(cfg EnvConfig) gpu.GPUInfo {
	if cfg.GPUType != "" {
		return gpu.GPUInfo{
			Type:   normalizeGPUType(cfg.GPUType, cfg.GFXVal),
			GfxVal: cfg.GFXVal,
			Name:   "Forzada por .env",
		}
	}

	hw := gpu.Detect()
	hw.Type = normalizeGPUType(hw.Type, hw.GfxVal)
	if hw.Name == "" {
		hw.Name = "Desconocida"
	}
	return hw
}

func normalizeGPUType(gpuType, gfxVal string) string {
	gpuType = strings.ToLower(strings.TrimSpace(gpuType))
	gfxVal = strings.TrimSpace(gfxVal)

	switch gpuType {
	case "rdna4", "rdna3", "nvidia", "intel", "generic":
		return gpuType
	case "amd":
		major := gfxMajor(gfxVal)
		if major >= 12 {
			return "rdna4"
		}
		if major == 10 || major == 11 {
			return "rdna3"
		}
		return "amd"
	default:
		if major := gfxMajor(gfxVal); major >= 12 {
			return "rdna4"
		} else if major == 10 || major == 11 {
			return "rdna3"
		}
		if gpuType == "" {
			return "generic"
		}
		return gpuType
	}
}

func gfxMajor(gfxVal string) int {
	gfxVal = strings.TrimSpace(gfxVal)
	if gfxVal == "" {
		return 0
	}

	if strings.HasPrefix(strings.ToLower(gfxVal), "gfx") {
		gfxVal = strings.TrimPrefix(strings.ToLower(gfxVal), "gfx")
		if len(gfxVal) >= 2 {
			major, _ := strconv.Atoi(gfxVal[:2])
			return major
		}
	}

	parts := strings.Split(gfxVal, ".")
	major, _ := strconv.Atoi(parts[0])
	return major
}

func baseImageName(gpuType string) string {
	gpuType = strings.TrimSpace(gpuType)
	if gpuType == "" {
		gpuType = "generic"
	}
	return fmt.Sprintf("localhost/axiom-%s:latest", gpuType)
}

func ensureTutorFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	return file.Close()
}

func currentUserGroup() string {
	user := os.Getenv("USER")
	if user == "" {
		user = "root"
	}
	return user + ":" + user
}

func removePathWritable(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	_ = filepath.Walk(path, func(currentPath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		mode := info.Mode()
		if mode&0200 == 0 {
			_ = os.Chmod(currentPath, mode|0200)
		}
		return nil
	})
	return os.RemoveAll(path)
}

func runCommandQuiet(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func runCommandWithInput(name, input string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func runCommandOutputQuiet(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return strings.TrimSpace(string(output)), nil
}

func (m *Manager) Rebuild() error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return err
	}

	hardware := resolveBuildGPU(cfg)
	targetImage := baseImageName(hardware.Type)

	confirm, _ := m.UI.AskConfirmInCard(
		"rebuild",
		[]Field{
			{Label: "Imagen", Value: targetImage},
			{Label: "GPU", Value: hardware.Type},
		},
		nil,
		"rebuild.confirm",
	)
	if !confirm {
		return nil
	}

	steps := []LifecycleStep{
		{Title: m.UI.GetText("rebuild.step_rm_image"), Detail: targetImage, Status: LifecycleRunning},
	}
	m.UI.ClearScreen()
	m.UI.ShowLogo()
	m.UI.RenderLifecycle(m.UI.GetText("rebuild.title"), m.UI.GetText("rebuild.subtitle"), steps, "", nil)

	_ = runCommandQuiet("podman", "rmi", targetImage, "--force")
	steps[0].Status = LifecycleDone
	m.UI.ClearScreen()
	m.UI.ShowLogo()
	m.UI.RenderLifecycle(m.UI.GetText("rebuild.title"), m.UI.GetText("rebuild.subtitle"), steps, "", nil)

	return m.Build()
}

func (m *Manager) Reset() error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return err
	}

	hardware := resolveBuildGPU(cfg)
	targetImage := baseImageName(hardware.Type)
	names, err := m.listBunkerNames(cfg)
	if err != nil {
		names = []string{}
	}

	confirm, reason, err := m.UI.AskReset([]Field{
		{Label: "Búnkeres", Value: fmt.Sprintf("%d detectados", len(names))},
		{Label: "Imagen", Value: targetImage},
	}, names)
	if err != nil {
		return err
	}
	if !confirm {
		return nil
	}

	_ = appendTutorLog(cfg.TutorPath(), m.UI.GetText("reset.log_reason", strings.TrimSpace(reason)))

	steps := make([]LifecycleStep, 0, len(names)+1)
	for _, name := range names {
		steps = append(steps, LifecycleStep{Title: m.UI.GetText("reset.step_rm_bunker"), Detail: name, Status: LifecyclePending})
	}
	steps = append(steps, LifecycleStep{Title: m.UI.GetText("reset.step_rm_image"), Detail: targetImage, Status: LifecyclePending})

	renderReset := func(current []LifecycleStep) {
		m.UI.ClearScreen()
		m.UI.ShowLogo()
		m.UI.RenderLifecycle(m.UI.GetText("reset.title"), m.UI.GetText("reset.subtitle"), current, "", nil)
	}

	renderReset(steps)

	for i, name := range names {
		steps[i].Status = LifecycleRunning
		renderReset(steps)
		_ = runCommandQuiet("distrobox-rm", name, "--force", "--yes")
		_ = removePathWritable(cfg.BuildWorkspaceDir(name))
		steps[i].Status = LifecycleDone
		renderReset(steps)
	}

	lastIdx := len(steps) - 1
	steps[lastIdx].Status = LifecycleRunning
	renderReset(steps)
	_ = runCommandQuiet("podman", "rmi", targetImage, "--force")
	steps[lastIdx].Status = LifecycleDone
	renderReset(steps)

	m.UI.ShowWarning(m.UI.GetText("reset.success_title"), m.UI.GetText("reset.success_desc"), nil, nil, m.UI.GetText("reset.success_footer"))
	return nil
}
