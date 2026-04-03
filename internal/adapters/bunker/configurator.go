package bunker

import (
	"fmt"
	"path/filepath"

	"github.com/Alejandro-M-P/AXIOM/internal/config"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

type BunkerConfiguratorAdapter struct {
	fs        ports.IFileSystem
	assetsDir string
	presenter ports.IPresenter
}

func NewBunkerConfiguratorAdapter(fs ports.IFileSystem, assetsDir string, presenter ports.IPresenter) *BunkerConfiguratorAdapter {
	return &BunkerConfiguratorAdapter{fs: fs, assetsDir: assetsDir, presenter: presenter}
}

func (c *BunkerConfiguratorAdapter) WriteShellBootstrap(cfg config.EnvConfig, name, envDir, gfxOverride string) error {
	shellConfig := filepath.Join(envDir, ".bashrc")
	content := fmt.Sprintf(c.presenter.GetText("bunker_bootstrap.shell_header"), name)
	if gfxOverride != "" {
		content += fmt.Sprintf(c.presenter.GetText("bunker_bootstrap.gfx_export"), gfxOverride)
	}
	return c.fs.WriteFile(shellConfig, []byte(content), 0644)
}

func (c *BunkerConfiguratorAdapter) WriteStarshipConfig(envDir string, starshipAssetPath string) error {
	starshipConfig := filepath.Join(envDir, ".config", "starship.toml")

	data, err := c.fs.ReadFile(starshipAssetPath)
	if err != nil {
		return err
	}

	if err := c.fs.MkdirAll(filepath.Dir(starshipConfig), 0755); err != nil {
		return err
	}
	return c.fs.WriteFile(starshipConfig, data, 0644)
}

func (c *BunkerConfiguratorAdapter) CopyTutorToAgents(tutorPath, envDir string) error {
	if _, err := c.fs.Stat(tutorPath); err != nil {
		return nil
	}

	data, err := c.fs.ReadFile(tutorPath)
	if err != nil {
		return err
	}

	teamsDir := filepath.Join(envDir, "teams")
	if err := c.fs.MkdirAll(teamsDir, 700); err != nil {
		return err
	}

	destPath := filepath.Join(teamsDir, "tutor.md")
	return c.fs.WriteFile(destPath, data, 700)
}

func (c *BunkerConfiguratorAdapter) WriteOpencodeConfig(envDir string) error {
	opencodeConfig := filepath.Join(envDir, ".config", "opencode", "config.yaml")
	content := `axiom:
  enable: true
  workspace: /projects
`

	dir := filepath.Dir(opencodeConfig)
	if err := c.fs.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return c.fs.WriteFile(opencodeConfig, []byte(content), 0644)
}
