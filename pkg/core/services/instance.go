package bunker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"axiom/pkg/core/domain"
	"axiom/pkg/core/ports"
)

// Create crea o reusa un búnker y entra directamente dentro de él.
func (m *Manager) Create(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("missing_name")
	}

	cleanName, err := sanitizeBunkerName(name)
	if err != nil {
		return err
	}
	name = cleanName

	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("errors.env.read: %w", err)
	}

	projectDir := filepath.Join(cfg.BaseDir, name)
	envDir := cfg.BuildWorkspaceDir(name)
	rcPath := filepath.Join(envDir, ".bashrc")
	hardware := resolveBuildGPU(cfg)
	imageName := baseImageName(hardware.Type)
	sshMounted := sshVolumeFlag() != ""

	m.UI.ShowLogo()
	m.UI.ShowCommandCard(
		"create",
		[]ports.Field{
			{Label: "fields.name", Value: name},
			{Label: "fields.image", Value: imageName},
			{Label: "fields.project", Value: projectDir},
			{Label: "fields.environment", Value: envDir},
			{Label: "fields.gpu", Value: hardware.Type},
			{Label: "fields.ssh", Value: yesNo(sshMounted)},
		},
		nil,
	)

	sudoCmd := exec.Command("sudo", "-v")
	sudoCmd.Stdin = os.Stdin
	sudoCmd.Stdout = os.Stdout
	sudoCmd.Stderr = os.Stderr
	if err := sudoCmd.Run(); err != nil {
		return fmt.Errorf("access_denied")
	}

	if exists, err := m.distroboxExists(name); err != nil {
		return err
	} else if exists {
		prepareSSHAgent(cfg)
		m.UI.ShowWarning(
			"warnings.bunker_exists.title",
			"warnings.bunker_exists.desc",
			[]ports.Field{{Label: "fields.name", Value: name}, {Label: "fields.environment", Value: envDir}},
			nil,
			"warnings.bunker_exists.footer",
		)
		return enterBunker(name, rcPath)
	}

	if !m.Runtime.ImageExists(imageName) {
		available, _ := m.listAxiomImages()
		m.UI.ShowWarning(
			"warnings.missing_image.title",
			"warnings.missing_image.desc",
			[]ports.Field{{Label: "fields.expected", Value: imageName}},
			available,
			"warnings.missing_image.footer",
		)
		return fmt.Errorf("missing_image")
	}

	if err := os.MkdirAll(projectDir, 0700); err != nil {
		return err
	}
	if err := m.FS.MkdirAll(envDir, 0700); err != nil {
		return err
	}
	if err := m.FS.MkdirAll(filepath.Join(cfg.AIConfigDir(), "models"), 0700); err != nil {
		return err
	}
	if err := ensureTutorFile(m.FS, cfg.TutorPath()); err != nil {
		return err
	}

	flags := m.createContainerFlags(cfg, hardware.Type, name, projectDir)
	if err := m.Runtime.RunCommand("", "distrobox-create",
		"--name", name,
		"--image", imageName,
		"--home", envDir,
		"--additional-flags", flags,
		"--yes",
	); err != nil {
		return err
	}

	// Forzamos el arranque para que el entrypoint de distrobox inicie la configuración
	_ = m.Runtime.RunCommand("", "podman", "start", name)

	// Espera activa: comprobamos que el búnker esté 'running'
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	isReady := false
WaitLoop:
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout")
		case <-ticker.C:
			if m.bunkerStatus(name) == "running" {
				isReady = true
				break WaitLoop
			}
		}
	}
	if !isReady {
		return fmt.Errorf("unexpected")
	}
	// Pequeña gracia de tiempo para asegurar que el entrypoint termine de poblar ~/.entorno/
	time.Sleep(2 * time.Second)

	if err := m.Runtime.RunCommand("", "distrobox-enter", "-n", name, "--", "sudo", "pacman", "-Syu", "--noconfirm", "--needed"); err != nil {
		return err
	}

	gfxOverride := strings.TrimSpace(cfg.GFXVal)
	if gfxOverride == "" {
		gfxOverride = strings.TrimSpace(hardware.GfxVal)
	}
	if err := writeShellBootstrap(m.FS, cfg, name, envDir, gfxOverride); err != nil {
		return err
	}
	if err := writeStarshipConfig(m.FS, envDir); err != nil {
		return err
	}
	if err := copyTutorToAgents(cfg.TutorPath(), envDir); err != nil {
		return err
	}
	if err := writeOpencodeConfig(envDir); err != nil {
		return err
	}

	prepareSSHAgent(cfg)
	m.UI.ShowWarning(
		"warnings.bunker_ready.title",
		"warnings.bunker_ready.desc",
		[]ports.Field{{Label: "fields.name", Value: name}, {Label: "fields.image", Value: imageName}, {Label: "fields.environment", Value: envDir}},
		nil,
		"warnings.bunker_ready.footer",
	)
	return enterBunker(name, rcPath)
}

// Stop detiene un búnker activo sin borrar su entorno ni el proyecto.
func (m *Manager) Stop() error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("errors.env.read: %w", err)
	}

	names, err := m.listBunkerNames(cfg)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		m.UI.ShowLogo()
		m.UI.ShowWarning(
			"warnings.no_bunkers.title",
			"warnings.no_bunkers.desc",
			nil,
			nil,
			"warnings.no_bunkers.footer",
		)
		return nil
	}

	var activeNames []string
	for _, name := range names {
		if m.bunkerStatus(name) == "running" {
			activeNames = append(activeNames, name)
		}
	}

	if len(activeNames) == 0 {
		m.UI.ShowLogo()
		m.UI.ShowWarning(
			"warnings.none_active.title",
			"warnings.none_active.desc",
			nil,
			nil,
			"warnings.none_active.footer",
		)
		return nil
	}

	selected, err := selectBunkerInteractive("prompts.select_active.title", "prompts.select_active.desc", activeNames)
	if err != nil {
		return err
	}

	if err := m.Runtime.RunCommand("", "distrobox-stop", selected, "--yes"); err != nil {
		return err
	}

	m.UI.ShowLogo()
	m.UI.ShowCommandCard(
		"stop",
		[]ports.Field{
			{Label: "fields.name", Value: selected},
			{Label: "fields.status", Value: "stopped"},
			{Label: "fields.environment", Value: humanPath(bunkerEnvPath(cfg, selected))},
		},
		nil,
	)
	return nil
}

// Delete elimina un búnker y permite decidir si también se borra el código del proyecto.
func (m *Manager) Delete(name string) error {
	name = strings.TrimSpace(name)

	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("errors.env.read: %w", err)
	}

	if name == "" {
		names, err := m.listBunkerNames(cfg)
		if err != nil {
			return err
		}
		selected, err := selectBunkerInteractive("prompts.delete_bunker.title", "prompts.delete_bunker.desc", names)
		if err != nil {
			return err
		}
		name = selected
	}

	cleanName, err := sanitizeBunkerName(name)
	if err != nil {
		return err
	}
	name = cleanName

	envDir := cfg.BuildWorkspaceDir(name)
	projectDir := filepath.Join(cfg.BaseDir, name)

	confirm, reason, deleteCode, err := m.UI.AskDelete(name, []ports.Field{
		{Label: "fields.name", Value: name},
		{Label: "fields.environment", Value: envDir},
		{Label: "fields.project", Value: projectDir},
	})
	if err != nil {
		return err
	}
	if !confirm {
		return nil
	}

	_ = appendTutorLog(m.FS, cfg.TutorPath(), fmt.Sprintf("logs.tutor.bunker_deleted", name, reason))

	m.UI.ShowLog("delete.cleaning")

	if err := m.Runtime.RunCommand("", "distrobox-rm", name, "--force", "--yes"); err != nil {
		return err
	}

	if err := removePathWritable(m.FS, envDir); err != nil {
		return err
	}

	if deleteCode {
		if err := removeProjectPath(m.FS, projectDir); err != nil {
			return err
		}
	}

	m.UI.ShowWarning(
		"warnings.bunker_deleted.title",
		"warnings.bunker_deleted.desc",
		[]ports.Field{{Label: "fields.name", Value: name}, {Label: "fields.environment", Value: envDir}, {Label: "fields.code_deleted", Value: yesNo(deleteCode)}},
		nil,
		"",
	)
	return nil
}

// DeleteImage elimina la imagen base correspondiente a la GPU configurada/detectada.
func (m *Manager) DeleteImage() error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("errors.env.read: %w", err)
	}

	hardware := resolveBuildGPU(cfg)
	targetImage := baseImageName(hardware.Type)
	images, err := m.listAxiomImages()
	if err != nil {
		return err
	}

	confirm, err := m.UI.AskConfirmInCard(
		"delete-image",
		[]ports.Field{{Label: "fields.target", Value: targetImage}, {Label: "fields.gpu", Value: hardware.Type}},
		images,
		"delete-image.confirm",
	)
	if err != nil || !confirm {
		return nil
	}

	if err := m.Runtime.RunCommand("", "podman", "rmi", targetImage, "--force"); err != nil {
		if !m.Runtime.ImageExists(targetImage) {
			return fmt.Errorf("errors.bunker.image_not_found: %s", targetImage)
		}
		return err
	}

	remaining, _ := m.listAxiomImages()
	m.UI.ShowWarning(
		"warnings.image_deleted.title",
		"warnings.image_deleted.desc",
		[]ports.Field{{Label: "fields.deleted", Value: targetImage}},
		remaining,
		"warnings.image_deleted.footer",
	)
	return nil
}

func (m *Manager) createContainerFlags(cfg domain.EnvConfig, gpuType, name, projectDir string) string {
	parts := []string{
		fmt.Sprintf("--volume %s:/%s:z", projectDir, name),
		fmt.Sprintf("--volume %s:/ai_config:z", cfg.AIConfigDir()),
		fmt.Sprintf("--volume %s:/run/axiom/env:ro,z", filepath.Join(cfg.AxiomPath, ".env")),
		"--device /dev/kfd",
		"--device /dev/dri",
		"--security-opt label=disable",
		"--group-add video",
		"--group-add render",
	}

	if cfg.ROCMMode == "host" {
		parts = append(parts, hostGPUVolumeFlags(gpuType)...)
	}
	if sshFlag := sshVolumeFlag(); sshFlag != "" {
		parts = append(parts, sshFlag)
	}

	return strings.Join(parts, " ")
}

func (m *Manager) distroboxExists(name string) (bool, error) {
	return m.Runtime.ContainerExists(name)
}

func (m *Manager) listAxiomImages() ([]string, error) {
	output, err := m.Runtime.RunCommandOutput("", "podman", "images", "--format", "json")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(output) == "" {
		return nil, nil
	}

	var imagesData []struct {
		Names []string `json:"Names"`
	}
	if err := json.Unmarshal([]byte(output), &imagesData); err != nil {
		return nil, err
	}

	var images []string
	for _, img := range imagesData {
		for _, n := range img.Names {
			if strings.HasPrefix(n, "localhost/axiom-") {
				images = append(images, n)
			}
		}
	}
	sort.Strings(images)
	return images, nil
}

func prepareSSHAgent(cfg domain.EnvConfig) {
	if strings.ToLower(strings.TrimSpace(cfg.AuthMode)) != "ssh" {
		return
	}
	ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	if exec.CommandContext(ctx1, "ssh-add", "-l").Run() == nil {
		return
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	_ = exec.CommandContext(ctx2, "ssh-add", filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519")).Run()
}

func sshVolumeFlag() string {
	sock := strings.TrimSpace(os.Getenv("SSH_AUTH_SOCK"))
	if sock == "" {
		return ""
	}
	info, err := os.Stat(sock)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		return ""
	}
	return fmt.Sprintf("--volume %s:%s", sock, sock)
}

func hostGPUVolumeFlags(gpuType string) []string {
	var flags []string
	addPath := func(path string) {
		realPath, err := filepath.EvalSymlinks(path)
		if err == nil && realPath != path {
			// Resolver symlink: montamos el real apuntando al nombre original esperado
			flags = append(flags, fmt.Sprintf("--volume %s:%s:ro", realPath, path))
			flags = append(flags, fmt.Sprintf("--volume %s:%s:ro", realPath, realPath))
		} else {
			flags = append(flags, fmt.Sprintf("--volume %s:%s:ro", path, path))
		}
	}

	switch strings.ToLower(strings.TrimSpace(gpuType)) {
	case "rdna3", "rdna4", "amd", "generic":
		for _, path := range []string{"/usr/lib/rocm", "/usr/lib64/rocm", "/opt/rocm"} {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				addPath(path)
			}
		}
		for _, binary := range []string{"rocminfo", "rocm-smi"} {
			if resolved, err := exec.LookPath(binary); err == nil {
				addPath(resolved)
			}
		}
	case "nvidia":
		// Inyección nativa CDI (Container Device Interface) para NVIDIA moderno
		flags = append(flags, "--device=nvidia.com/gpu=all")
		for _, path := range []string{"/usr/lib/x86_64-linux-gnu/libcuda.so.1", "/usr/local/cuda"} {
			if _, err := os.Stat(path); err == nil {
				addPath(path)
			}
		}
	case "intel":
		for _, path := range []string{"/usr/lib/intel-opencl", "/usr/lib/x86_64-linux-gnu/intel-opencl"} {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				addPath(path)
			}
		}
	}
	return flags
}

func enterBunker(name, rcPath string) error {
	cmd := exec.Command("distrobox-enter", name, "--", "bash", "--rcfile", rcPath, "-i")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func appendTutorLog(fs ports.IFileSystem, path, line string) error {
	if err := ensureTutorFile(fs, path); err != nil {
		return err
	}
	file, err := fs.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = fmt.Fprintln(file, line)
	return err
}

func removeProjectPath(fs ports.IFileSystem, path string) error {
	info, err := fs.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not_dir")
	}
	return removePathWritable(fs, path)
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func yesNo(value bool) string {
	if value {
		return "common.yes"
	}
	return "common.no"
}

func isYes(value string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	return trimmed == "s" || trimmed == "y" || trimmed == "si" || trimmed == "sí" || trimmed == "yes"
}

// List muestra el estado de los búnkeres detectados en el sistema.
func (m *Manager) List() error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("errors.env.read: %w", err)
	}

	names, err := m.listBunkerNames(cfg)
	if err != nil {
		return err
	}

	if len(names) == 0 {
		m.UI.ShowLogo()
		m.UI.ShowWarning(
			"warnings.no_bunkers_list.title",
			"warnings.no_bunkers_list.desc",
			nil,
			nil,
			"warnings.no_bunkers_list.footer",
		)
		return nil
	}

	selected, err := selectBunkerInteractive("prompts.select_available.title", "prompts.select_available.desc", names)
	if err != nil {
		return err
	}

	return m.Info(selected)
}

func (m *Manager) bunkerStatus(name string) string {
	containers, err := m.Runtime.ListContainers()
	if err != nil {
		return "stopped"
	}

	for _, c := range containers {
		if c.Name == name {
			return c.Status
		}
	}
	return "stopped"
}

func bunkerEnvSize(cfg domain.EnvConfig, name string) string {
	path := cfg.BuildWorkspaceDir(name)
	if _, err := os.Stat(path); err != nil {
		return "-"
	}
	var size int64
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				size += info.Size()
			}
		}
		return nil
	})
	if err != nil || size == 0 {
		return "-"
	}
	return formatBytes(size)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func bunkerGitBranch(cfg domain.EnvConfig, name string) string {
	projectPath := filepath.Join(cfg.BaseDir, name)
	headPath := filepath.Join(projectPath, ".git", "HEAD")
	content, err := os.ReadFile(headPath)
	if err != nil {
		return "-"
	}
	head := strings.TrimSpace(string(content))
	if strings.HasPrefix(head, "ref: refs/heads/") {
		return strings.TrimPrefix(head, "ref: refs/heads/")
	}
	if len(head) >= 7 {
		return head[:7] // fallback en caso de "detached HEAD" (ej. commit exacto)
	}
	return "-"
}

func bunkerLastEntry(cfg domain.EnvConfig, name string) string {
	path := cfg.BuildWorkspaceDir(name)
	info, err := os.Stat(path)
	if err != nil {
		return "-"
	}
	return info.ModTime().Format("2006-01-02")
}

func bunkerProjectPath(cfg domain.EnvConfig, name string) string {
	return filepath.Join(cfg.BaseDir, name)
}

func bunkerEnvPath(cfg domain.EnvConfig, name string) string {
	return cfg.BuildWorkspaceDir(name)
}

func humanPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return strings.Replace(path, home, "~", 1)
	}
	return path
}

func bunkerTimestamp(t time.Time) string {
	return t.Format("2006-01-02")
}

func (m *Manager) Prune() error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return err
	}

	envBaseDir := filepath.Join(cfg.BaseDir, ".entorno")
	entries, err := os.ReadDir(envBaseDir)
	if err != nil {
		return nil
	}

	var activeNames []string
	containers, err := m.Runtime.ListContainers()
	if err == nil {
		for _, c := range containers {
			if c.Name != "" {
				activeNames = append(activeNames, c.Name)
			}
		}
	}

	activeMap := make(map[string]bool)
	for _, n := range activeNames {
		activeMap[n] = true
	}

	var orphans []string
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == defaultBuildContainerName {
			continue
		}
		if !activeMap[entry.Name()] {
			orphans = append(orphans, entry.Name())
		}
	}

	m.UI.ShowLogo()
	if len(orphans) == 0 {
		m.UI.ShowWarning(
			"warnings.prune_clean.title",
			"warnings.prune_clean.desc",
			nil,
			nil,
			"warnings.prune_clean.footer",
		)
		return nil
	}

	confirm, err := m.UI.AskConfirmInCard(
		"prune",
		[]ports.Field{{Label: "fields.orphans", Value: fmt.Sprintf("%d", len(orphans))}},
		orphans,
		"prune.confirm",
	)
	if err != nil || !confirm {
		return nil
	}

	m.UI.ShowLog("prune.cleaning")

	var wg sync.WaitGroup
	for _, h := range orphans {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			m.UI.ShowLog("prune.deleting_item", name)
			_ = removePathWritable(m.FS, filepath.Join(envBaseDir, name))
		}(h)
	}
	wg.Wait()

	m.UI.ShowWarning("warnings.prune_completed.title", "warnings.prune_completed.desc", nil, nil, "")
	return nil
}

func (m *Manager) Info(name string) error {
	if strings.TrimSpace(name) == "" {
		return m.List()
	}

	cleanName, err := sanitizeBunkerName(name)
	if err != nil {
		return err
	}
	name = cleanName

	cfg, err := m.LoadConfig()
	if err != nil {
		return err
	}

	hardware := resolveBuildGPU(cfg)
	imageName := baseImageName(hardware.Type)

	m.UI.ShowLogo()
	m.UI.ShowCommandCard(
		"info",
		[]ports.Field{
			{Label: "fields.name", Value: name},
			{Label: "fields.status", Value: m.bunkerStatus(name)},
			{Label: "fields.image", Value: imageName},
			{Label: "fields.gpu", Value: hardware.Type},
			{Label: "fields.environment", Value: humanPath(bunkerEnvPath(cfg, name))},
			{Label: "fields.project", Value: humanPath(bunkerProjectPath(cfg, name))},
			{Label: "fields.size", Value: bunkerEnvSize(cfg, name)},
			{Label: "fields.last_activity", Value: bunkerLastEntry(cfg, name)},
			{Label: "fields.git_branch", Value: defaultString(bunkerGitBranch(cfg, name), "-")},
		},
		nil,
	)
	return nil
}

func sanitizeBunkerName(name string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(name))
	if clean == "." || clean == ".." || strings.Contains(clean, "/") || strings.Contains(clean, "\\") {
		return "", fmt.Errorf("invalid_name")
	}
	return clean, nil
}
