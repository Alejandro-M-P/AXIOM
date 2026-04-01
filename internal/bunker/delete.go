package bunker

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

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

	confirm, reason, deleteCode, err := m.ui.AskDelete(name, []ports.Field{
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
			{Label: "fields.name", Value: name},
			{Label: "fields.environment", Value: envDir},
			{Label: "fields.code_deleted", Value: yesNo(deleteCode)},
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
	targetImage := baseImageName(hardware.Type)
	images, err := m.listAxiomImages(ctx)
	if err != nil {
		return err
	}

	confirm, err := m.ui.AskConfirmInCard(
		"delete-image",
		[]ports.Field{
			{Label: "fields.target", Value: targetImage},
			{Label: "fields.gpu", Value: hardware.Type},
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
			{Label: "fields.deleted", Value: targetImage},
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
		if strings.HasPrefix(b.Image, "localhost/axiom-") && !seen[b.Image] {
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
			if c.Name == "" || c.Name == defaultBuildContainerName {
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
			if name == "" || name == defaultBuildContainerName {
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

// selectBunkerInteractive permite seleccionar un búnker de forma interactiva.
// Placeholder - en la implementación real se usaría la UI interactiva.
// Por ahora retornamos error ya que esta función requiere implementación de UI.
func selectBunkerInteractive(title, action string, names []string) (string, error) {
	if len(names) == 0 {
		return "", fmt.Errorf("no hay búnkeres disponibles")
	}
	// En una implementación real, esto invocaría la UI de selección
	// Por ahora retornamos el primero si solo hay uno
	if len(names) == 1 {
		return names[0], nil
	}
	return "", fmt.Errorf("selección interactiva no implementada")
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
		if strings.HasPrefix(b.Image, "localhost/axiom-") && !seen[b.Image] {
			seen[b.Image] = true
			images = append(images, b.Image)
		}
	}

	_ = images // Placeholder para evitar unused error
	return images, nil
}
