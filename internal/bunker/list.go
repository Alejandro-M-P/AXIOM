package bunker

import (
	"context"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/i18n"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// list maneja la listados de búnkeres.
func (m *Manager) list(ctx context.Context) error {
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

	return m.info(ctx, selected)
}

// info maneja la muestra de información de un búnker.
func (m *Manager) info(ctx context.Context, name string) error {
	if strings.TrimSpace(name) == "" {
		return m.list(ctx)
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

	m.ui.ShowLogo()
	m.ui.ShowCommandCard(
		"info",
		[]ports.Field{
			ports.NewField("fields.name", name),
			ports.NewField("fields.status", m.BunkerStatus(ctx, name)),
			ports.NewField("fields.image", imageName),
			ports.NewField("fields.gpu", hardware.Type),
			ports.NewField("fields.environment", humanPath(m.fs, bunkerEnvPath(cfg, name))),
			ports.NewField("fields.project", humanPath(m.fs, bunkerProjectPath(cfg, name))),
			ports.NewField("fields.size", bunkerEnvSize(m.fs, cfg, name)),
			ports.NewField("fields.last_activity", bunkerLastEntry(m.fs, cfg, name)),
			ports.NewField("fields.git_branch", defaultString(bunkerGitBranch(m.fs, cfg, name), "-")),
		},
		nil,
	)
	return nil
}

// bunkerStatus retorna el estado actual de un búnker.
func (m *Manager) bunkerStatus(ctx context.Context, name string) string {
	bunkers, err := m.runtime.ListBunkers(ctx)
	if err != nil {
		return i18n.Commands["bunker"]["status_stopped"]
	}

	for _, b := range bunkers {
		if b.Name == name {
			return b.Status
		}
	}
	return i18n.Commands["bunker"]["status_stopped"]
}
