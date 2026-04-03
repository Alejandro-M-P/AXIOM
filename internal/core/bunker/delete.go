package bunker

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/config"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// delete maneja la eliminación de un búnker.
func (m *Manager) delete(ctx context.Context, name string, force, deleteImage bool) error {
	name = strings.TrimSpace(name)

	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("errors.env.read: %w", err)
	}

	if name == "" {
		names, err := m.listBunkerNames(ctx, cfg)
		if err != nil {
			return err
		}
		selected, err := selectBunkerInteractive(m.ui, "prompts.delete_bunker.title", "prompts.delete_bunker.desc", names)
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

	envDir := config.BuildWorkspaceDir(cfg.BaseDir, name)
	projectDir := filepath.Join(cfg.BaseDir, name)

	confirm, reason, deleteCode, err := m.ui.AskDelete(name, []ports.Field{
		ports.NewField("fields.name", name),
		ports.NewField("fields.environment", envDir),
		ports.NewField("fields.project", projectDir),
	})
	if err != nil {
		return err
	}
	if !confirm {
		return nil
	}

	_ = appendTutorLog(fmt.Sprintf("logs.tutor.bunker_deleted %s %s", name, reason))

	m.ui.ShowLog("delete.cleaning")

	if err := m.runtime.RemoveBunker(ctx, name, force); err != nil {
		return err
	}

	if err := removePathWritable(m.fs, envDir); err != nil {
		return err
	}

	if deleteCode {
		if err := removeProjectPath(m.fs, projectDir); err != nil {
			return err
		}
	}

	m.ui.ShowWarning(
		"warnings.bunker_deleted.title",
		"warnings.bunker_deleted.desc",
		[]ports.Field{
			ports.NewField("fields.name", name),
			ports.NewField("fields.environment", envDir),
			ports.NewField("fields.code_deleted", yesNo(deleteCode)),
		},
		nil,
		"",
	)
	return nil
}

// deleteImage maneja la eliminación de una imagen de búnker.
func (m *Manager) deleteImage(ctx context.Context) error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("errors.env.read: %w", err)
	}

	hardware := resolveBuildGPU(cfg)
	targetImage := baseImageName(m.ui, hardware.Type)
	images, err := m.listAxiomImages(ctx)
	if err != nil {
		return err
	}

	confirm, err := m.ui.AskConfirmInCard(
		"delete-image",
		[]ports.Field{
			ports.NewField("fields.target", targetImage),
			ports.NewField("fields.gpu", hardware.Type),
		},
		images,
		"delete-image.confirm",
	)
	if err != nil || !confirm {
		return nil
	}

	if err := m.runtime.RemoveImage(ctx, targetImage, true); err != nil {
		if !m.ImageExists(ctx, targetImage) {
			return fmt.Errorf("errors.bunker.image_not_found: %s", targetImage)
		}
		return err
	}

	remaining, _ := m.listAxiomImages(ctx)
	m.ui.ShowWarning(
		"warnings.image_deleted.title",
		"warnings.image_deleted.desc",
		[]ports.Field{
			ports.NewField("fields.deleted", targetImage),
		},
		remaining,
		"warnings.image_deleted.footer",
	)
	return nil
}

// listAxiomImages lista las imágenes de AXIOM disponibles.
// Nota: Esta función usa ListBunkers ya que ExecuteInBunker no retorna output.
// Para obtener output de comandos, se necesitaría un método diferente en el runtime.
func (m *Manager) listAxiomImages(ctx context.Context) ([]string, error) {
	bunkers, err := m.runtime.ListBunkers(ctx)
	if err != nil {
		return nil, err
	}

	var images []string
	seen := make(map[string]bool)
	for _, b := range bunkers {
		if m.runtime.IsAxiomImage(b.Image) && !seen[b.Image] {
			seen[b.Image] = true
			images = append(images, b.Image)
		}
	}
	sort.Strings(images)
	return images, nil
}

// listBunkerNames lista los nombres de todos los búnkeres.
func (m *Manager) listBunkerNames(ctx context.Context, cfg EnvConfig) ([]string, error) {
	seen := make(map[string]struct{})
	var names []string

	containers, err := m.runtime.ListBunkers(ctx)
	if err == nil {
		for _, c := range containers {
			if c.Name == "" || c.Name == DefaultBuildContainerName {
				continue
			}
			if _, ok := seen[c.Name]; ok {
				continue
			}
			seen[c.Name] = struct{}{}
			names = append(names, c.Name)
		}
	}

	envBaseDir := filepath.Join(cfg.BaseDir, ".entorno")
	entries, readErr := m.fs.ReadDir(envBaseDir)
	if readErr == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := strings.TrimSpace(entry.Name())
			if name == "" || name == DefaultBuildContainerName {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}

	sort.Strings(names)
	return names, nil
}

// selectBunkerInteractive permite seleccionar un búnker de forma interactiva utilizando el presenter.
func selectBunkerInteractive(ui ports.IPresenter, title, action string, names []string) (string, error) {
	if len(names) == 0 {
		return "", fmt.Errorf("errors.bunker.no_bunkers_available")
	}
	if len(names) == 1 {
		return names[0], nil
	}

	// Convertir nombres a map de statuses (todos como "available" por ahora)
	statuses := make(map[string]string)
	for _, name := range names {
		statuses[name] = ui.GetText("common.available") // o un estado más específico si lo necesitamos
	}

	selected, confirmed, err := ui.AskSelectBunker(names, statuses, title, action)
	if err != nil {
		return "", err
	}
	if !confirmed {
		return "", fmt.Errorf("errors.bunker.selection_cancelled")
	}
	return selected, nil
}

// listAxiomPodmanImages lista imágenes usando comando podman directamente.
// Esta es una implementación alternativa que usa parsing de JSON.
func listAxiomPodmanImages(ctx context.Context, runtime ports.IBunkerRuntime) ([]string, error) {
	// Intentar ejecutar comando directamente - esto puede no funcionar
	// dependiendo de la implementación del runtime
	var images []string
	seen := make(map[string]bool)

	bunkers, err := runtime.ListBunkers(ctx)
	if err != nil {
		return nil, err
	}

	for _, b := range bunkers {
		if runtime.IsAxiomImage(b.Image) && !seen[b.Image] {
			seen[b.Image] = true
			images = append(images, b.Image)
		}
	}

	_ = images // Placeholder para evitar unused error
	return images, nil
}
