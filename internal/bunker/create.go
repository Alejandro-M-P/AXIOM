package bunker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// BunkerFlags contiene las flags para crear un búnker.
type BunkerFlags struct {
	GPUType    string
	ProjectDir string
	HomeDir    string
}

// create maneja la lógica de creación de un búnker.
func (m *Manager) create(ctx context.Context, name string) error {
	return m.createWithImage(ctx, name, "")
}

// createWithImage maneja la lógica de creación de un búnker con imagen específica.
func (m *Manager) createWithImage(ctx context.Context, name, image string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("errors.bunker.missing_name")
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

	// Use provided image or auto-detect based on GPU
	imageName := strings.TrimSpace(image)
	if imageName == "" {
		imageName = baseImageName(hardware.Type)
	}

	// Get SSH agent socket for volume mounting
	sshSocket, _ := m.system.SSHAgentSocket()
	sshMounted := sshVolumeFlag(sshSocket) != ""

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

	if err := m.system.RefreshSudo(ctx); err != nil {
		return fmt.Errorf("errors.bunker.access_denied")
	}

	exists, err := m.bunkerExists(name)
	if err != nil {
		return err
	}
	if exists {
		if strings.ToLower(strings.TrimSpace(cfg.AuthMode)) == "ssh" {
			_ = m.system.PrepareSSHAgent(ctx)
		}
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
		return m.runtime.EnterBunker(ctx, name, rcPath)
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
		return fmt.Errorf("errors.bunker.missing_image")
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
	if err := ensureTutorFile(m.fs, cfg.TutorPath()); err != nil {
		return err
	}

	flags := m.createContainerFlags(cfg, hardware.Type, name, projectDir, sshSocket)
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
			return fmt.Errorf("errors.bunker.timeout")
		case <-ticker.C:
			if m.BunkerStatus(ctx, name) == "running" {
				isReady = true
				break WaitLoop
			}
		}
	}
	if !isReady {
		return fmt.Errorf("errors.bunker.unexpected")
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

	if strings.ToLower(strings.TrimSpace(cfg.AuthMode)) == "ssh" {
		_ = m.system.PrepareSSHAgent(ctx)
	}
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
	return m.runtime.EnterBunker(ctx, name, rcPath)
}

// bunkerExists verifica si un búnker existe.
func (m *Manager) bunkerExists(name string) (bool, error) {
	return m.runtime.BunkerExists(context.Background(), name)
}

// createContainerFlags genera los flags para crear el contenedor.
func (m *Manager) createContainerFlags(cfg EnvConfig, gpuType, name, projectDir, sshSocket string) string {
	// 1. Obtener volume flags del presenter (usa i18n)
	volFlags, err := m.ui.GetBunkerVolumeFlags(
		projectDir,
		name,
		cfg.AIConfigDir(),
		cfg.AxiomPath+"/config.toml",
		gpuType,
		sshSocket,
	)
	volumeStr := ""
	if err == nil && len(volFlags) > 0 {
		// Unir todos los volume flags
		for _, v := range volFlags {
			if v != "" {
				volumeStr += v + " "
			}
		}
	}

	// 2. Pasar al runtime que añade los device flags
	flags, err := m.runtime.GetCreateFlags(
		context.Background(),
		name,
		"", // image - no se usa en GetCreateFlags
		"", // home - no se usa en GetCreateFlags
		strings.TrimSpace(volumeStr),
	)
	if err != nil {
		// Si falla, devolver solo los volume flags.
		// Los device flags son responsabilidad del runtime, no del core.
		return strings.TrimSpace(volumeStr)
	}
	return flags
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
