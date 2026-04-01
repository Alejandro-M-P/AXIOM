package bunker

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// prune maneja la limpieza de búnkeres huérfanos.
// IMPORTANTE: Este método es thread-safe gracias al mutex del Manager.
func (m *Manager) prune(ctx context.Context) error {
	// Adquirir mutex para prevenir race conditions con otras operaciones de prune
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg, err := m.LoadConfig()
	if err != nil {
		return err
	}

	envBaseDir := filepath.Join(cfg.BaseDir, ".entorno")
	entries, err := m.fs.ReadDir(envBaseDir)
	if err != nil {
		return nil
	}

	var activeNames []string
	containers, err := m.runtime.ListBunkers(ctx)
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

	m.ui.ShowLogo()
	if len(orphans) == 0 {
		m.ui.ShowWarning(
			"warnings.prune_clean.title",
			"warnings.prune_clean.desc",
			nil,
			nil,
			"warnings.prune_clean.footer",
		)
		return nil
	}

	confirm, err := m.ui.AskConfirmInCard(
		"prune",
		[]ports.Field{{Label: "fields.orphans", Value: fmt.Sprintf("%d", len(orphans))}},
		orphans,
		"prune.confirm",
	)
	if err != nil || !confirm {
		return nil
	}

	m.ui.ShowLog("prune.cleaning")

	// Usar WaitGroup para esperara la eliminación de todos los directorios
	var wg sync.WaitGroup
	for _, h := range orphans {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			m.ui.ShowLog("prune.deleting_item", name)
			_ = removePathWritable(m.fs, filepath.Join(envBaseDir, name))
		}(h)
	}
	wg.Wait()

	m.ui.ShowWarning("warnings.prune_completed.title", "warnings.prune_completed.desc", nil, nil, "")
	return nil
}
