package bunker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"axiom/pkg/gpu"
	"axiom/pkg/ui/styles"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

type buildContext struct {
	config            EnvConfig
	gpuInfo           gpu.GPUInfo
	imageName         string
	buildWorkspaceDir string
}

// buildProgress mantiene el estado visual del build para la UI.
type buildProgress struct {
	title    string
	subtitle string
	steps    []styles.LifecycleStep
}

// Build ejecuta el flujo completo de construccion de la imagen base.
func (m *Manager) Build() error {
	ctx, err := m.prepareBuildContext()
	if err != nil {
		return err
	}

	progress := newBuildProgress(ctx)
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
		// Si alguna fase falla, intentamos desmontar el contenedor temporal.
		if cleanupDone {
			return
		}
		_ = runCommandQuiet("distrobox-rm", m.buildContainerName, "--force", "--yes")
		_ = removePathWritable(ctx.buildWorkspaceDir)
	}()

	if err := progress.runStep(2, func() error {
		return m.installSystemBase(ctx)
	}); err != nil {
		progress.renderError(err, ctx.buildWorkspaceDir)
		return err
	}

	if err := progress.runStep(3, func() error {
		return m.installDeveloperTools(ctx)
	}); err != nil {
		progress.renderError(err, ctx.buildWorkspaceDir)
		return err
	}

	if err := progress.runStep(4, func() error {
		return m.installModelStack(ctx)
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

	progress.subtitle = fmt.Sprintf("Imagen lista: %s", ctx.imageName)
	progress.render()
	fmt.Printf("\n✅ Imagen %s lista. Usa: axiom create [nombre]\n", ctx.imageName)
	return nil
}

// prepareBuildContext reune la configuracion y los valores derivados que el build necesita.
func (m *Manager) prepareBuildContext() (buildContext, error) {
	cfg, err := m.LoadConfig()
	if err != nil {
		return buildContext{}, fmt.Errorf("no se pudo leer .env: %w", err)
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

// newBuildProgress define las fases funcionales que vera el usuario durante el build.
func newBuildProgress(ctx buildContext) *buildProgress {
	gpuModeText := "Drivers desde host"
	if ctx.config.ROCMMode == "image" {
		gpuModeText = "Drivers dentro de la imagen"
	}

	return &buildProgress{
		title: fmt.Sprintf("Construyendo %s", ctx.imageName),
		subtitle: fmt.Sprintf("GPU: %s | Modo: %s", ctx.gpuInfo.Type, gpuModeText),
		steps: []styles.LifecycleStep{
			{Title: "Preparar directorios", Detail: ctx.config.AIConfigDir(), Status: styles.LifecyclePending},
			{Title: "Recrear contenedor temporal", Detail: ctx.buildWorkspaceDir, Status: styles.LifecyclePending},
			{Title: "Instalar sistema base", Detail: "pacman + paquetes GPU", Status: styles.LifecyclePending},
			{Title: "Instalar herramientas dev", Detail: "OpenCode, Engram, gentle-ai", Status: styles.LifecyclePending},
			{Title: "Preparar stack IA", Detail: "Ollama + agent-teams-lite", Status: styles.LifecyclePending},
			{Title: "Empaquetar imagen", Detail: ctx.imageName, Status: styles.LifecyclePending},
		},
	}
}

// runStep actualiza el estado visual antes y despues de ejecutar la accion real.
func (p *buildProgress) runStep(index int, fn func() error) error {
	for i := range p.steps {
		if i < index && p.steps[i].Status != styles.LifecycleDone {
			p.steps[i].Status = styles.LifecycleDone
		}
		if i == index {
			p.steps[i].Status = styles.LifecycleRunning
		}
	}
	p.render()

	if err := fn(); err != nil {
		p.steps[index].Status = styles.LifecycleError
		return err
	}

	p.steps[index].Status = styles.LifecycleDone
	p.render()
	return nil
}

// render repinta toda la pantalla para mantener una sensacion de flujo guiado.
func (p *buildProgress) render() {
	fmt.Print("\033[H\033[2J")
	fmt.Println(styles.GetLogo())
	fmt.Println(styles.RenderLifecycle(p.title, p.subtitle, p.steps))
}

// renderError conserva el paso fallido y anade contexto util para depuracion.
func (p *buildProgress) renderError(err error, where string) {
	fmt.Print("\033[H\033[2J")
	fmt.Println(styles.GetLogo())
	fmt.Println(styles.RenderLifecycleError(p.title, p.steps, err, where))
}

// prepareSharedDirectories crea los directorios persistentes compartidos por los bunker.
func (m *Manager) prepareSharedDirectories(ctx buildContext) error {
	if err := os.MkdirAll(filepath.Join(ctx.config.AIConfigDir(), "models"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(ctx.config.AIConfigDir(), "teams"), 0755); err != nil {
		return err
	}
	if err := ensureTutorFile(ctx.config.TutorPath()); err != nil {
		return err
	}
	return runCommandQuiet("sudo", "chown", "-R", currentUserGroup(), ctx.config.AIConfigDir())
}

// recreateBuildContainer fuerza un build limpio eliminando cualquier contenedor temporal previo.
func (m *Manager) recreateBuildContainer(ctx buildContext) error {
	_ = runCommandQuiet("distrobox-rm", m.buildContainerName, "--force", "--yes")
	if err := removePathWritable(ctx.buildWorkspaceDir); err != nil {
		return err
	}
	if err := os.MkdirAll(ctx.buildWorkspaceDir, 0755); err != nil {
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

// buildContainerFlags concentra dispositivos, mounts y permisos del contenedor temporal.
func (m *Manager) buildContainerFlags(ctx buildContext) string {
	return fmt.Sprintf(
		"--volume %s:/ai_config --volume %s:/run/axiom/env:ro --device /dev/kfd --device /dev/dri --security-opt label=disable --group-add video --group-add render",
		ctx.config.AIConfigDir(),
		filepath.Join(ctx.config.AxiomPath, ".env"),
	)
}

// installSystemBase instala el sistema Arch minimo y, si toca, los paquetes GPU internos.
func (m *Manager) installSystemBase(ctx buildContext) error {
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

	args := append([]string{"sudo", "pacman", "-Sy", "--needed", "--noconfirm"}, packages...)
	return m.runInContainer(args...)
}

// installDeveloperTools instala herramientas de trabajo del bunker fuera del sistema base.
func (m *Manager) installDeveloperTools(ctx buildContext) error {
	if err := m.runInContainer("sudo", "npm", "install", "-g", "opencode-ai"); err != nil {
		return err
	}
	if err := m.runInContainer("go", "install", "github.com/Gentleman-Programming/engram/cmd/engram@latest"); err != nil {
		return err
	}

	gopath, err := m.runInContainerOutput("go", "env", "GOPATH")
	if err != nil {
		return err
	}
	engramPath := filepath.ToSlash(filepath.Join(strings.TrimSpace(gopath), "bin", "engram"))
	if err := m.runInContainer("sudo", "cp", "-f", engramPath, "/usr/local/bin/"); err != nil {
		return err
	}

	version, err := latestGentleAIVersion(ctx.config.GitToken)
	if err != nil {
		version = "0.1.0"
	}
	assetURL := fmt.Sprintf(
		"https://github.com/Gentleman-Programming/gentle-ai/releases/download/v%s/gentle-ai_%s_linux_amd64.tar.gz",
		version,
		version,
	)
	if err := m.runInContainer("curl", "-fsSL", assetURL, "-o", "/tmp/gentle-ai.tar.gz"); err != nil {
		return err
	}
	if err := m.runInContainer("tar", "-xzf", "/tmp/gentle-ai.tar.gz", "-C", "/tmp"); err != nil {
		return err
	}
	if err := m.runInContainer("sudo", "mv", "/tmp/gentle-ai", "/usr/local/bin/gentle-ai"); err != nil {
		return err
	}
	if err := m.runInContainer("sudo", "chmod", "+x", "/usr/local/bin/gentle-ai"); err != nil {
		return err
	}
	return m.runInContainer("rm", "-f", "/tmp/gentle-ai.tar.gz")
}

func latestGentleAIVersion(token string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/Gentleman-Programming/gentle-ai/releases/latest", nil)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("github releases respondio %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	version := strings.TrimSpace(strings.TrimPrefix(release.TagName, "v"))
	if version == "" {
		return "", fmt.Errorf("tag_name vacio")
	}
	return version, nil
}

// installModelStack prepara el runtime de modelos y su configuracion inicial.
func (m *Manager) installModelStack(ctx buildContext) error {
	if err := m.installOllama(ctx.gpuInfo.Type); err != nil {
		return err
	}
	if err := m.installAgentTeamsLite(); err != nil {
		return err
	}
	return m.cleanBuildCaches()
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
		return "", fmt.Errorf("arquitectura no soportada para Ollama: %s", runtime.GOARCH)
	}
}

// installAgentTeamsLite levanta Ollama de forma temporal para ejecutar el setup inicial.
func (m *Manager) installAgentTeamsLite() error {
	ollamaCmd := exec.Command("distrobox-enter", "-n", m.buildContainerName, "--", "ollama", "serve")
	logFile, err := os.OpenFile("/tmp/ollama-build.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	ollamaCmd.Stdout = logFile
	ollamaCmd.Stderr = logFile
	if err := ollamaCmd.Start(); err != nil {
		return err
	}

	stopped := false
	defer func() {
		if stopped || ollamaCmd.Process == nil {
			return
		}
		_ = ollamaCmd.Process.Kill()
		_, _ = ollamaCmd.Process.Wait()
	}()

	if err := m.waitForOllama(); err != nil {
		return err
	}
	_ = m.runInContainer("rm", "-rf", "/tmp/agent-teams")
	if err := m.runInContainer("git", "clone", "https://github.com/Gentleman-Programming/agent-teams-lite.git", "/tmp/agent-teams"); err != nil {
		return err
	}
	if err := m.runInteractiveInContainer("1\n", "/tmp/agent-teams/scripts/setup.sh", "--all"); err != nil {
		return err
	}

	if ollamaCmd.Process != nil {
		_ = ollamaCmd.Process.Kill()
		_, _ = ollamaCmd.Process.Wait()
		stopped = true
	}
	return nil
}

// waitForOllama hace polling del endpoint local hasta que Ollama responde.
func (m *Manager) waitForOllama() error {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := m.runInContainerOutput("curl", "-s", "http://localhost:11434/"); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("ollama no arranco en 60s")
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

// exportBuildImage congela el contenedor temporal en una imagen reutilizable.
func (m *Manager) exportBuildImage(ctx buildContext) error {
	return runCommandQuiet("podman", "commit", m.buildContainerName, ctx.imageName)
}

// destroyBuildContainer elimina el contenedor temporal y su home asociado.
func (m *Manager) destroyBuildContainer(ctx buildContext) error {
	if err := runCommandQuiet("distrobox-rm", m.buildContainerName, "--force", "--yes"); err != nil {
		return err
	}
	return removePathWritable(ctx.buildWorkspaceDir)
}

// runInContainer ejecuta comandos dentro del contenedor temporal en modo silencioso.
func (m *Manager) runInContainer(args ...string) error {
	containerArgs := append([]string{"-n", m.buildContainerName, "--"}, args...)
	return runCommandQuiet("distrobox-enter", containerArgs...)
}

// runInteractiveInContainer permite responder prompts puntuales sin exponer ruido en la UI.
func (m *Manager) runInteractiveInContainer(input string, args ...string) error {
	containerArgs := append([]string{"-n", m.buildContainerName, "--"}, args...)
	return runCommandWithInput("distrobox-enter", input, containerArgs...)
}

func (m *Manager) runInContainerOutput(args ...string) (string, error) {
	containerArgs := append([]string{"-n", m.buildContainerName, "--"}, args...)
	return runCommandOutputQuiet("distrobox-enter", containerArgs...)
}

// resolveBuildGPU prioriza overrides del .env y cae a autodeteccion cuando no existen.
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

// normalizeGPUType convierte tipos amplios como amd en familias mas utiles para el build.
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
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE, 0644)
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

// runCommandQuiet captura stdout/stderr y solo los expone si el comando falla.
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
