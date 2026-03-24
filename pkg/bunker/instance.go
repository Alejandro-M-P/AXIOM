package bunker

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"axiom/pkg/ui/styles"
)

// Create crea o reusa un búnker y entra directamente dentro de él.
func (m *Manager) Create(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("uso: axiom create [nombre]")
	}

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

	fmt.Println(styles.GetLogo())
	fmt.Println(styles.RenderBunkerCard(
		"Crear Búnker",
		"Preparando entorno de trabajo desde la imagen base",
		[]styles.BunkerDetail{
			{Label: "Nombre", Value: name},
			{Label: "Imagen", Value: imageName},
			{Label: "Proyecto", Value: projectDir},
			{Label: "Entorno", Value: envDir},
			{Label: "GPU", Value: hardware.Type},
			{Label: "SSH", Value: yesNo(sshMounted)},
		},
		nil,
		"El búnker reutilizará la imagen base ya construida.",
	))

	if err := runCommandQuiet("sudo", "-v"); err != nil {
		return fmt.Errorf("acceso denegado: %w", err)
	}

	if exists, err := distroboxExists(name); err != nil {
		return err
	} else if exists {
		prepareSSHAgent(cfg)
		fmt.Println(styles.RenderBunkerCard(
			"Búnker Existente",
			"Ya existe un contenedor con ese nombre; se abrirá directamente.",
			[]styles.BunkerDetail{{Label: "Nombre", Value: name}, {Label: "Entorno", Value: envDir}},
			nil,
			"Reutilizando configuración previa.",
		))
		return enterBunker(name, rcPath)
	}

	if !podmanImageExists(imageName) {
		available, _ := listAxiomImages()
		return fmt.Errorf("no se encontró la imagen base %s. Ejecuta: axiom build\n%s", imageName, styles.RenderBunkerWarning(
			"Imagen Base No Disponible",
			"No se puede crear el búnker sin una imagen previa.",
			[]styles.BunkerDetail{{Label: "Esperada", Value: imageName}},
			available,
			"Construye primero la base con: axiom build",
		))
	}

	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(cfg.AIConfigDir(), "models"), 0755); err != nil {
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
	fmt.Println(styles.RenderBunkerCard(
		"Búnker Listo",
		"El contenedor ya está preparado y se abrirá ahora.",
		[]styles.BunkerDetail{{Label: "Nombre", Value: name}, {Label: "Imagen", Value: imageName}, {Label: "Entorno", Value: envDir}},
		nil,
		"Configuración aplicada: shell, starship y opencode.",
	))
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
		fmt.Println(styles.GetLogo())
		fmt.Println(styles.RenderBunkerCard(
			"Stop Búnker",
			"No hay búnkeres creados.",
			nil,
			nil,
			"Usa axiom create <nombre> para crear uno nuevo.",
		))
		return nil
	}

	hardware := resolveBuildGPU(cfg)
	imageName := baseImageName(hardware.Type)
	var rows []styles.BunkerRow
	for _, name := range names {
		if bunkerStatus(name) != "running" {
			continue
		}
		rows = append(rows, styles.BunkerRow{
			Name:        name,
			Status:      bunkerStatus(name),
			Size:        bunkerEnvSize(cfg, name),
			LastEntry:   bunkerLastEntry(cfg, name),
			GitBranch:   bunkerGitBranch(cfg, name),
			Image:       imageName,
			GPU:         hardware.Type,
			ProjectPath: humanPath(bunkerProjectPath(cfg, name)),
			EnvPath:     humanPath(bunkerEnvPath(cfg, name)),
		})
	}

	if len(rows) == 0 {
		fmt.Println(styles.GetLogo())
		fmt.Println(styles.RenderBunkerCard(
			"Stop Búnker",
			"No hay búnkeres activos ahora mismo.",
			nil,
			nil,
			"Todos los búnkeres detectados ya están parados.",
		))
		return nil
	}

	selected, err := selectBunkerRow(rows)
	if err != nil {
		return err
	}

	if selected.Status == "stopped" {
		fmt.Println(styles.GetLogo())
		fmt.Println(styles.RenderBunkerCard(
			selected.Name,
			"El búnker ya está parado.",
			[]styles.BunkerDetail{{Label: "Estado", Value: selected.Status}, {Label: "Entorno", Value: selected.EnvPath}},
			nil,
			"No fue necesario hacer cambios.",
		))
		return nil
	}

	if err := runCommandQuiet("distrobox-stop", selected.Name, "--yes"); err != nil {
		return err
	}

	fmt.Println(styles.GetLogo())
	fmt.Println(styles.RenderBunkerCard(
		selected.Name,
		"El búnker se ha parado correctamente.",
		[]styles.BunkerDetail{
			{Label: "Estado", Value: "stopped"},
			{Label: "Imagen", Value: selected.Image},
			{Label: "GPU", Value: selected.GPU},
			{Label: "Entorno", Value: selected.EnvPath},
			{Label: "Proyecto", Value: selected.ProjectPath},
		},
		nil,
		"El entorno sigue intacto y puedes volver a abrirlo con axiom create <nombre>.",
	))
	return nil
}

// Delete elimina un búnker y permite decidir si también se borra el código del proyecto.
func (m *Manager) Delete(name string) error {
	_ = strings.TrimSpace(name)

	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("no se pudo leer .env: %w", err)
	}

	names, err := listBunkerNames(cfg)
	if err != nil {
		return err
	}
	selected, err := selectBunkerName(names)
	if err != nil {
		return err
	}
	name = selected

	envDir := cfg.BuildWorkspaceDir(name)
	projectDir := filepath.Join(cfg.BaseDir, name)

	fmt.Println(styles.GetLogo())
	fmt.Println(styles.RenderBunkerCard(
		"Eliminar Búnker",
		"Se eliminará el contenedor y podrás decidir si también borrar el código.",
		[]styles.BunkerDetail{{Label: "Nombre", Value: name}, {Label: "Entorno", Value: envDir}, {Label: "Proyecto", Value: projectDir}},
		nil,
		"",
	))

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("📝 Razón técnica obligatoria: ")
	reason, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("cancelado: se requiere justificación")
	}

	fmt.Print("🗂️  ¿Borrar también el código del proyecto? (s/N): ")
	deleteProject, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	deleteCode := isYes(deleteProject)


	fmt.Printf("❗ ¿Confirmas el borrado del búnker '%s'? (s/N): ", name)
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !isYes(confirm) {
		fmt.Println(styles.RenderBunkerWarning(
			"Operación Cancelada",
			"No se realizaron cambios.",
			[]styles.BunkerDetail{{Label: "Búnker", Value: name}},
			nil,
			"El entorno y el código permanecen intactos.",
		))
		return nil
	}

	steps := []styles.LifecycleStep{
		{Title: "Eliminar contenedor", Detail: name, Status: styles.LifecyclePending},
		{Title: "Eliminar entorno local", Detail: envDir, Status: styles.LifecyclePending},
	}
	if deleteCode {
		steps = append(steps, styles.LifecycleStep{Title: "Eliminar código del proyecto", Detail: projectDir, Status: styles.LifecyclePending})
	}

	renderDelete := func(current []styles.LifecycleStep) {
		fmt.Print("\033[H\033[2J")
		fmt.Println(styles.GetLogo())
		fmt.Println(styles.RenderLifecycle("Eliminando Búnker", "Desinstalando recursos del búnker", current))
	}

	renderDelete(steps)
	steps[0].Status = styles.LifecycleRunning
	renderDelete(steps)
	if err := runCommandQuiet("distrobox-rm", name, "--force", "--yes"); err != nil {
		steps[0].Status = styles.LifecycleError
		renderDelete(steps)
		return err
	}
	steps[0].Status = styles.LifecycleDone

	steps[1].Status = styles.LifecycleRunning
	renderDelete(steps)
	if err := removePathWritable(envDir); err != nil {
		steps[1].Status = styles.LifecycleError
		renderDelete(steps)
		return err
	}
	steps[1].Status = styles.LifecycleDone

	if deleteCode {
		steps[2].Status = styles.LifecycleRunning
		renderDelete(steps)
		if err := removeProjectPath(projectDir); err != nil {
			steps[2].Status = styles.LifecycleError
			renderDelete(steps)
			return err
		}
		steps[2].Status = styles.LifecycleDone
	}

	renderDelete(steps)
	fmt.Println(styles.RenderBunkerCard(
		"Búnker Eliminado",
		"La desinstalación terminó correctamente.",
		[]styles.BunkerDetail{{Label: "Nombre", Value: name}, {Label: "Entorno", Value: envDir}, {Label: "Código borrado", Value: yesNo(deleteCode)}},
		nil,
		"",
	))
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

	fmt.Println(styles.GetLogo())
	fmt.Println(styles.RenderBunkerCard(
		"Eliminar Imagen Base",
		"Se eliminará la imagen asociada a la GPU activa y se listan las disponibles.",
		[]styles.BunkerDetail{{Label: "Objetivo", Value: targetImage}, {Label: "GPU", Value: hardware.Type}},
		images,
		"Usa axiom build para reconstruirla cuando la necesites.",
	))

	if err := runCommandQuiet("podman", "rmi", targetImage, "--force"); err != nil {
		if !podmanImageExists(targetImage) {
			return fmt.Errorf("no se encontró la imagen %s", targetImage)
		}
		return err
	}

	remaining, _ := listAxiomImages()
	fmt.Println(styles.RenderBunkerCard(
		"Imagen Eliminada",
		"La imagen base se eliminó correctamente.",
		[]styles.BunkerDetail{{Label: "Eliminada", Value: targetImage}},
		remaining,
		"Estas son las imágenes de AXIOM que siguen disponibles.",
	))
	return nil
}

func (m *Manager) createContainerFlags(cfg EnvConfig, gpuType, name, projectDir string) string {
	parts := []string{
		fmt.Sprintf("--volume %s:/%s", projectDir, name),
		fmt.Sprintf("--volume %s:/ai_config", cfg.AIConfigDir()),
		fmt.Sprintf("--volume %s:/run/axiom/env:ro", filepath.Join(cfg.AxiomPath, ".env")),
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
		flags = append(flags, fmt.Sprintf("--volume %s:%s:ro", path, path))
	}

	switch gpuType {
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
		return fmt.Errorf("la ruta del proyecto no es un directorio: %s", path)
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

type bunkerSummary struct {
	Name        string
	Status      string
	Size        string
	LastEntry   string
	GitBranch   string
	Image       string
	GPU         string
	ProjectPath string
	EnvPath     string
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

	fmt.Println(styles.GetLogo())
	if len(names) == 0 {
		fmt.Println(styles.RenderBunkerCard(
			"Búnkeres",
			"Estado actual del sistema AXIOM",
			nil,
			nil,
			"No hay búnkeres creados todavía. Usa: axiom create <nombre>",
		))
		return nil
	}

	hardware := resolveBuildGPU(cfg)
	imageName := baseImageName(hardware.Type)

	var rows []styles.BunkerRow
	for _, name := range names {
		summary := bunkerSummary{
			Name:        name,
			Status:      bunkerStatus(name),
			Size:        bunkerEnvSize(cfg, name),
			LastEntry:   bunkerLastEntry(cfg, name),
			GitBranch:   bunkerGitBranch(cfg, name),
			Image:       imageName,
			GPU:         hardware.Type,
			ProjectPath: humanPath(bunkerProjectPath(cfg, name)),
			EnvPath:     humanPath(bunkerEnvPath(cfg, name)),
		}
		rows = append(rows, styles.BunkerRow{
			Name:        summary.Name,
			Status:      summary.Status,
			Size:        summary.Size,
			LastEntry:   summary.LastEntry,
			GitBranch:   summary.GitBranch,
			Image:       summary.Image,
			GPU:         summary.GPU,
			ProjectPath: summary.ProjectPath,
			EnvPath:     summary.EnvPath,
		})
	}

	selected, err := selectBunkerRow(rows)
	if err != nil {
		return err
	}

	fmt.Println(styles.GetLogo())
	fmt.Println(styles.RenderBunkerCard(
		selected.Name,
		"Ficha del búnker seleccionado.",
		[]styles.BunkerDetail{
			{Label: "Estado", Value: selected.Status},
			{Label: "Imagen", Value: selected.Image},
			{Label: "GPU", Value: selected.GPU},
			{Label: "Entorno", Value: selected.EnvPath},
			{Label: "Proyecto", Value: selected.ProjectPath},
			{Label: "Tamaño", Value: selected.Size},
			{Label: "Última actividad", Value: selected.LastEntry},
			{Label: "Rama git", Value: defaultString(selected.GitBranch, "-")},
		},
		nil,
		"Usa axiom delete para eliminar este búnker o axiom create para entrar si ya existe.",
	))
	return nil
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
