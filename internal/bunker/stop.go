package bunker

import (
	"context"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// stop maneja la detención de un búnker.
func (m *Manager) stop(ctx context.Context) error {
	cfg, err := m.LoadConfig()
	if err != nil {
		return err
	}

	names, err := m.listBunkerNames(ctx, cfg)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		m.ui.ShowLogo()
		m.ui.ShowWarning(
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
		if m.BunkerStatus(ctx, name) == "running" {
			activeNames = append(activeNames, name)
		}
	}

	if len(activeNames) == 0 {
		m.ui.ShowLogo()
		m.ui.ShowWarning(
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

	if err := m.runtime.StopBunker(ctx, selected); err != nil {
		return err
	}

	m.ui.ShowLogo()
	m.ui.ShowCommandCard(
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
