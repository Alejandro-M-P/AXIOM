package bunker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
		return fmt.Errorf("no se pudo leer .env: %w", err)
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
		[]Field{
			{Label: "Nombre", Value: name},
			{Label: "Imagen", Value: imageName},
			{Label: "Proyecto", Value: projectDir},
			{Label: "Entorno", Value: envDir},
			{Label: "GPU", Value: hardware.Type},
			{Label: "SSH", Value: yesNo(sshMounted)},
		},
		nil,
	)

	if err := runCommandQuiet("sudo", "-v"); err != nil {
		return fmt.Errorf("access_denied")
	}

	if exists, err := distroboxExists(name); err != nil {
		return err
	} else if exists {
		prepareSSHAgent(cfg)
		m.UI.ShowWarning(
			"Búnker Existente",
			"Ya existe un contenedor con ese nombre; se abrirá directamente.",
			[]Field{{Label: "Nombre", Value: name}, {Label: "Entorno", Value: envDir}},
			nil,
			"Reutilizando configuración previa.",
		)
		return enterBunker(name, rcPath)
	}

	if !podmanImageExists(imageName) {
		available, _ := listAxiomImages()
		m.UI.ShowWarning(
			"Imagen Base No Disponible",
			"No se puede crear el búnker sin una imagen previa.",
			[]Field{{Label: "Esperada", Value: imageName}},
			available,
			"Construye primero la base con: axiom build",
		)
		return fmt.Errorf("missing_image")
	}

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(envDir, 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(cfg.AIConfigDir(), "models"), 0700); err != nil {
		return err
	}
	if err := ensureTutorFile(cfg.TutorPath()); err != nil {
		return err
	}

	flags := m.createContainerFlags(cfg, hardware.Type, name, projectDir)
	if err := runCommandQuiet(
		"distrobox-create",
		"--name", name,
		"--image", imageName,
		"--home", envDir,
		"--additional-flags", flags,
		"--yes",
	); err != nil {
		return err
	}

	// Forzamos el arranque para que el entrypoint de distrobox inicie la configuración
	_ = runCommandQuiet("podman", "start", name)

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
			if bunkerStatus(name) == "running" {
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

	if err := runCommandQuiet("distrobox-enter", "-n", name, "--", "sudo", "pacman", "-Syu", "--noconfirm", "--needed"); err != nil {
		return err
	}

	gfxOverride := strings.TrimSpace(cfg.GFXVal)
	if gfxOverride == "" {
		gfxOverride = strings.TrimSpace(hardware.GfxVal)
	}
	if err := writeShellBootstrap(cfg, name, envDir, gfxOverride); err != nil {
		return err
	}
	if err := writeStarshipConfig(envDir); err != nil {
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
		"Búnker Listo",
		"El contenedor ya está preparado y se abrirá ahora.",
		[]Field{{Label: "Nombre", Value: name}, {Label: "Imagen", Value: imageName}, {Label: "Entorno", Value: envDir}},
		nil,
		"Configuración aplicada: shell, starship y opencode.",
	)
	return enterBunker(name, rcPath)
}

// Stop detiene un búnker activo sin borrar su entorno ni el proyecto.
func (m *Manager) Stop() error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("no se pudo leer .env: %w", err)
	}

	names, err := listBunkerNames(cfg)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		m.UI.ShowLogo()
		m.UI.ShowWarning(
			"Sin búnkeres",
			"No hay búnkeres creados.",
			nil,
			nil,
			"Usa axiom create <nombre> para crear uno nuevo.",
		)
		return nil
	}

	var activeNames []string
	for _, name := range names {
		if bunkerStatus(name) == "running" {
			activeNames = append(activeNames, name)
		}
	}

	if len(activeNames) == 0 {
		m.UI.ShowLogo()
		m.UI.ShowWarning(
			"Ninguno activo",
			"No hay búnkeres activos ahora mismo.",
			nil,
			nil,
			"Todos los búnkeres detectados ya están parados.",
		)
		return nil
	}

	selected, err := selectBunkerInteractive("Búnkeres Activos", "Pulsa Enter para detener este búnker", activeNames)
	if err != nil {
		return err
	}

	if err := runCommandQuiet("distrobox-stop", selected, "--yes"); err != nil {
		return err
	}

	m.UI.ShowLogo()
	m.UI.ShowCommandCard(
		"stop",
		[]Field{
			{Label: "Nombre", Value: selected},
			{Label: "Estado", Value: "stopped"},
			{Label: "Entorno", Value: humanPath(bunkerEnvPath(cfg, selected))},
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
		return fmt.Errorf("no se pudo leer .env: %w", err)
	}

	if name == "" {
		names, err := listBunkerNames(cfg)
		if err != nil {
			return err
		}
		selected, err := selectBunkerInteractive("Eliminar Búnker", "Pulsa Enter para eliminar este búnker", names)
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

	confirm, reason, deleteCode, err := m.UI.AskDelete(name, []Field{
		{Label: "Nombre", Value: name},
		{Label: "Entorno", Value: envDir},
		{Label: "Proyecto", Value: projectDir},
	})
	if err != nil {
		return err
	}
	if !confirm {
		return nil
	}

	_ = appendTutorLog(cfg.TutorPath(), fmt.Sprintf("- Búnker '%s' eliminado (Razón: %s)", name, reason))

	m.UI.ShowLog("delete.cleaning")

	if err := runCommandQuiet("distrobox-rm", name, "--force", "--yes"); err != nil {
		return err
	}

	if err := removePathWritable(envDir); err != nil {
		return err
	}

	if deleteCode {
		if err := removeProjectPath(projectDir); err != nil {
			return err
		}
	}

	m.UI.ShowWarning(
		"Búnker Eliminado",
		"La desinstalación terminó correctamente.",
		[]Field{{Label: "Nombre", Value: name}, {Label: "Entorno", Value: envDir}, {Label: "Código borrado", Value: yesNo(deleteCode)}},
		nil,
		"",
	)
	return nil
}

// DeleteImage elimina la imagen base correspondiente a la GPU configurada/detectada.
func (m *Manager) DeleteImage() error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("no se pudo leer .env: %w", err)
	}

	hardware := resolveBuildGPU(cfg)
	targetImage := baseImageName(hardware.Type)
	images, err := listAxiomImages()
	if err != nil {
		return err
	}

	confirm, err := m.UI.AskConfirmInCard(
		"delete-image",
		[]Field{{Label: "Objetivo", Value: targetImage}, {Label: "GPU", Value: hardware.Type}},
		images,
		"delete-image.confirm",
	)
	if err != nil || !confirm {
		return nil
	}

	if err := runCommandQuiet("podman", "rmi", targetImage, "--force"); err != nil {
		if !podmanImageExists(targetImage) {
			return fmt.Errorf("no se encontró la imagen %s", targetImage)
		}
		return err
	}

	remaining, _ := listAxiomImages()
	m.UI.ShowWarning(
		"Imagen Eliminada",
		"La imagen base se eliminó correctamente.",
		[]Field{{Label: "Eliminada", Value: targetImage}},
		remaining,
		"Estas son las imágenes de AXIOM que siguen disponibles.",
	)
	return nil
}

func (m *Manager) createContainerFlags(cfg EnvConfig, gpuType, name, projectDir string) string {
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

func distroboxExists(name string) (bool, error) {
	output, err := runCommandOutputQuiet("distrobox-list", "--no-color")
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) > 1 && strings.TrimSpace(parts[1]) == name {
				return true, nil
			}
		}
	}
	return false, nil
}

func listAxiomImages() ([]string, error) {
	output, err := runCommandOutputQuiet("podman", "images", "--format", "{{.Repository}}:{{.Tag}}")
	if err != nil {
		return nil, err
	}
	var images []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "localhost/axiom-") {
			images = append(images, line)
		}
	}
	sort.Strings(images)
	return images, nil
}

func podmanImageExists(image string) bool {
	cmd := exec.Command("podman", "image", "exists", image)
	return cmd.Run() == nil
}

func prepareSSHAgent(cfg EnvConfig) {
	if strings.ToLower(strings.TrimSpace(cfg.AuthMode)) != "ssh" {
		return
	}
	if exec.Command("ssh-add", "-l").Run() == nil {
		return
	}
	_ = exec.Command("ssh-add", filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519")).Run()
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

func appendTutorLog(path, line string) error {
	if err := ensureTutorFile(path); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = fmt.Fprintln(file, line)
	return err
}

func removeProjectPath(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not_dir")
	}
	return removePathWritable(path)
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func yesNo(value bool) string {
	if value {
		return "sí"
	}
	return "no"
}

func isYes(value string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	return trimmed == "s" || trimmed == "y" || trimmed == "si" || trimmed == "sí" || trimmed == "yes"
}

// List muestra el estado de los búnkeres detectados en el sistema.
func (m *Manager) List() error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("no se pudo leer .env: %w", err)
	}

	names, err := listBunkerNames(cfg)
	if err != nil {
		return err
	}

	if len(names) == 0 {
		m.UI.ShowLogo()
		m.UI.ShowWarning(
			"Sin búnkeres",
			"No se ha detectado ningún entorno.",
			nil,
			nil,
			"No hay búnkeres creados todavía. Usa: axiom create <nombre>",
		)
		return nil
	}

	selected, err := selectBunkerInteractive("Búnkeres Disponibles", "Pulsa Enter para ver la ficha técnica", names)
	if err != nil {
		return err
	}

	return m.Info(selected)
}

func bunkerStatus(name string) string {
	output, err := runCommandOutputQuiet("podman", "ps", "--format", "{{.Names}}")
	if err != nil {
		return "stopped"
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == name {
			return "running"
		}
	}
	return "stopped"
}

func bunkerEnvSize(cfg EnvConfig, name string) string {
	path := cfg.BuildWorkspaceDir(name)
	if _, err := os.Stat(path); err != nil {
		return "-"
	}
	cmd := exec.Command("du", "-sh", path)
	output, err := cmd.Output()
	if err != nil {
		return "-"
	}
	fields := strings.Fields(strings.TrimSpace(string(output)))
	if len(fields) == 0 {
		return "-"
	}
	return fields[0]
}

func bunkerGitBranch(cfg EnvConfig, name string) string {
	projectPath := filepath.Join(cfg.BaseDir, name)
	if _, err := os.Stat(filepath.Join(projectPath, ".git")); err != nil {
		return "-"
	}
	output, err := runCommandOutputQuiet("git", "-C", projectPath, "branch", "--show-current")
	if err != nil || strings.TrimSpace(output) == "" {
		return "-"
	}
	return strings.TrimSpace(output)
}

func bunkerLastEntry(cfg EnvConfig, name string) string {
	path := cfg.BuildWorkspaceDir(name)
	info, err := os.Stat(path)
	if err != nil {
		return "-"
	}
	return info.ModTime().Format("2006-01-02")
}

func bunkerProjectPath(cfg EnvConfig, name string) string {
	return filepath.Join(cfg.BaseDir, name)
}

func bunkerEnvPath(cfg EnvConfig, name string) string {
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
	output, err := runCommandOutputQuiet("distrobox-list", "--no-color")
	if err == nil {
		for _, line := range strings.Split(output, "\n") {
			if !strings.Contains(line, "|") {
				continue
			}
			parts := strings.Split(line, "|")
			if len(parts) < 2 {
				continue
			}
			name := strings.TrimSpace(parts[1])
			if name != "" {
				activeNames = append(activeNames, name)
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
			"Todo limpio",
			"No hay entornos huérfanos.",
			nil,
			nil,
			"Todo está limpio.",
		)
		return nil
	}

	confirm, err := m.UI.AskConfirmInCard(
		"prune",
		[]Field{{Label: "Huérfanos", Value: fmt.Sprintf("%d detectados", len(orphans))}},
		orphans,
		"prune.confirm",
	)
	if err != nil || !confirm {
		return nil
	}

	m.UI.ShowLog("prune.cleaning")

	for _, h := range orphans {
		m.UI.ShowLog("prune.deleting_item", h)
		_ = removePathWritable(filepath.Join(envBaseDir, h))
	}

	m.UI.ShowWarning("Prune Completado", "Se liberó el espacio de los entornos huérfanos.", nil, nil, "")
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
		[]Field{
			{Label: "Nombre", Value: name},
			{Label: "Estado", Value: bunkerStatus(name)},
			{Label: "Imagen", Value: imageName},
			{Label: "GPU", Value: hardware.Type},
			{Label: "Entorno", Value: humanPath(bunkerEnvPath(cfg, name))},
			{Label: "Proyecto", Value: humanPath(bunkerProjectPath(cfg, name))},
			{Label: "Tamaño", Value: bunkerEnvSize(cfg, name)},
			{Label: "Última actividad", Value: bunkerLastEntry(cfg, name)},
			{Label: "Rama git", Value: defaultString(bunkerGitBranch(cfg, name), "-")},
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
