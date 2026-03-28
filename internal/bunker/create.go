package bunker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"axiom/internal/ports"
)

// BunkerFlags contiene las flags para crear un búnker.
type BunkerFlags struct {
	GPUType    string
	ProjectDir string
	HomeDir    string
}

// create maneja la lógica de creación de un búnker.
func (m *Manager) create(ctx context.Context, name string) error {
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

	m.ui.ShowLogo()
	m.ui.ShowCommandCard(
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

	exists, err := m.bunkerExists(name)
	if err != nil {
		return err
	}
	if exists {
		prepareSSHAgent(cfg)
		m.ui.ShowWarning(
			"warnings.bunker_exists.title",
			"warnings.bunker_exists.desc",
			[]ports.Field{
				{Label: "fields.name", Value: name},
				{Label: "fields.environment", Value: envDir},
			},
			nil,
			"warnings.bunker_exists.footer",
		)
		return enterBunker(name, rcPath)
	}

	if !m.ImageExists(ctx, imageName) {
		available, _ := m.ListAxiomImages(ctx)
		m.ui.ShowWarning(
			"warnings.missing_image.title",
			"warnings.missing_image.desc",
			[]ports.Field{
				{Label: "fields.expected", Value: imageName},
			},
			available,
			"warnings.missing_image.footer",
		)
		return fmt.Errorf("missing_image")
	}

	if err := os.MkdirAll(projectDir, 0700); err != nil {
		return err
	}
	if err := m.fs.MkdirAll(envDir, 0700); err != nil {
		return err
	}
	if err := m.fs.MkdirAll(filepath.Join(cfg.AIConfigDir(), "models"), 0700); err != nil {
		return err
	}
	if err := ensureTutorFile(cfg.TutorPath()); err != nil {
		return err
	}

	flags := m.createContainerFlags(cfg, hardware.Type, name, projectDir)
	if err := m.runtime.CreateBunker(ctx, name, imageName, envDir, flags); err != nil {
		return err
	}

	// Forzamos el arranque para que el entrypoint de distrobox inicie la configuración
	_ = m.runtime.StartBunker(ctx, name)

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
			if m.BunkerStatus(ctx, name) == "running" {
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

	if err := m.runtime.ExecuteInBunker(ctx, name, "sudo", "pacman", "-Syu", "--noconfirm", "--needed"); err != nil {
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
	m.ui.ShowWarning(
		"warnings.bunker_ready.title",
		"warnings.bunker_ready.desc",
		[]ports.Field{
			{Label: "fields.name", Value: name},
			{Label: "fields.image", Value: imageName},
			{Label: "fields.environment", Value: envDir},
		},
		nil,
		"warnings.bunker_ready.footer",
	)
	return enterBunker(name, rcPath)
}

// bunkerExists verifica si un búnker existe.
func (m *Manager) bunkerExists(name string) (bool, error) {
	return m.runtime.BunkerExists(context.Background(), name)
}

// createContainerFlags genera los flags para crear el contenedor.
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

// hostGPUVolumeFlags retorna los flags de volumen para GPU en modo host.
func hostGPUVolumeFlags(gpuType string) []string {
	var flags []string
	addPath := func(path string) {
		realPath, err := filepath.EvalSymlinks(path)
		if err == nil && realPath != path {
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

// prepareSSHAgent prepara el agent SSH si está configurado.
func prepareSSHAgent(cfg EnvConfig) {
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

// enterBunker entra en un búnker de forma interactiva.
func enterBunker(name, rcPath string) error {
	cmd := exec.Command("distrobox-enter", name, "--", "bash", "--rcfile", rcPath, "-i")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Placeholder functions - implementar según necesidad
func writeShellBootstrap(cfg EnvConfig, name, envDir, gfxOverride string) error {
	return nil
}

func writeStarshipConfig(envDir string) error {
	return nil
}

func copyTutorToAgents(tutorPath, envDir string) error {
	return nil
}

func writeOpencodeConfig(envDir string) error {
	return nil
}
