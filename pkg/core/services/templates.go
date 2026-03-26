package bunker

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed assets/starship.toml
var starshipTemplate string

//go:embed assets/opencode.json
var opencodeTemplate string

type opencodeConfig struct {
	Schema     string                      `json:"$schema,omitempty"`
	Agent      map[string]opencodeAgent    `json:"agent,omitempty"`
	MCP        map[string]opencodeMCP      `json:"mcp,omitempty"`
	Permission opencodePermission          `json:"permission,omitempty"`
	Provider   map[string]opencodeProvider `json:"provider,omitempty"`
}

type opencodeAgent struct {
	Description string          `json:"description,omitempty"`
	Mode        string          `json:"mode,omitempty"`
	Prompt      string          `json:"prompt,omitempty"`
	Tools       map[string]bool `json:"tools,omitempty"`
}

type opencodeMCP struct {
	Enabled bool     `json:"enabled,omitempty"`
	Type    string   `json:"type,omitempty"`
	URL     string   `json:"url,omitempty"`
	Command []string `json:"command,omitempty"`
}

type opencodePermission struct {
	Bash map[string]string `json:"bash,omitempty"`
	Read map[string]string `json:"read,omitempty"`
}

type opencodeProvider struct {
	NPM     string                     `json:"npm,omitempty"`
	Options map[string]string          `json:"options,omitempty"`
	Models  map[string]map[string]bool `json:"models,omitempty"`
}

// writeShellBootstrap genera el único archivo de arranque interactivo que necesita el búnker.
// No depende de core.sh ni git.sh: la preparación queda autocontenida en Go.
func writeShellBootstrap(cfg EnvConfig, name, envDir, gfxOverride string) error {
	path := filepath.Join(envDir, ".bashrc")
	return os.WriteFile(path, []byte(renderShellBootstrap(cfg, name, gfxOverride)), 0600)
}

func renderShellBootstrap(cfg EnvConfig, name, gfxOverride string) string {
	var lines []string
	lines = append(lines,
		"eval \"$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)\"",
		fmt.Sprintf("export AXIOM_PATH=%q", cfg.AxiomPath),
		fmt.Sprintf("export AXIOM_GIT_USER=%q", cfg.GitUser),
		fmt.Sprintf("export AXIOM_GIT_EMAIL=%q", cfg.GitEmail),
		fmt.Sprintf("export AXIOM_AUTH_MODE=%q", defaultString(cfg.AuthMode, "https")),
		fmt.Sprintf("export SSH_AUTH_SOCK=%q", os.Getenv("SSH_AUTH_SOCK")),
		"export OLLAMA_MODELS=\"/ai_config/models\"",
	)
	if strings.TrimSpace(gfxOverride) != "" {
		lines = append(lines, fmt.Sprintf("export HSA_OVERRIDE_GFX_VERSION=%s", strings.TrimSpace(gfxOverride)))
	}
	lines = append(lines,
		"eval \"$(starship init bash)\"",
		fmt.Sprintf("cd /%s", name),
		"Archive=\"$HOME/.axiom_done\"",
		"if [ ! -f \"$Archive\" ]; then",
		"    if command -v gentle-ai >/dev/null 2>&1; then",
		"        gentle-ai",
		"    else",
		"        echo \"gentle-ai no encontrado, omitiendo arranque inicial.\"",
		"    fi",
		"    echo done > \"$Archive\"",
		"fi",
	)
	return strings.Join(lines, "\n") + "\n"
}

func writeStarshipConfig(envDir string) error {
	configDir := filepath.Join(envDir, ".config")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir, "starship.toml"), []byte(starshipTemplate), 0600)
}

func copyTutorToAgents(tutorPath, envDir string) error {
	data, err := os.ReadFile(tutorPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	configDir := filepath.Join(envDir, ".config", "opencode")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir, "AGENTS.md"), data, 0600)
}

func writeOpencodeConfig(envDir string) error {
	configDir := filepath.Join(envDir, ".config", "opencode")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}
	path := filepath.Join(configDir, "opencode.json")

	cfg, err := defaultOpencodeConfig()
	if err != nil {
		return err
	}
	if existing, readErr := os.ReadFile(path); readErr == nil {
		var current opencodeConfig
		if json.Unmarshal(existing, &current) == nil {
			mergeOpencodeConfig(&current, cfg)
			cfg = current
		}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}

func defaultOpencodeConfig() (opencodeConfig, error) {
	var cfg opencodeConfig
	if err := json.Unmarshal([]byte(opencodeTemplate), &cfg); err != nil {
		return opencodeConfig{}, fmt.Errorf("no se pudo parsear assets/opencode.json: %w", err)
	}
	return cfg, nil
}

func mergeOpencodeConfig(current *opencodeConfig, defaults opencodeConfig) {
	if current.Agent == nil {
		current.Agent = map[string]opencodeAgent{}
	}
	for key, value := range defaults.Agent {
		current.Agent[key] = value
	}
	if current.MCP == nil {
		current.MCP = defaults.MCP
	}
	if len(current.Permission.Bash) == 0 {
		current.Permission.Bash = defaults.Permission.Bash
	}
	if len(current.Permission.Read) == 0 {
		current.Permission.Read = defaults.Permission.Read
	}
	if current.Provider == nil {
		current.Provider = defaults.Provider
	}
	if strings.TrimSpace(current.Schema) == "" {
		current.Schema = defaults.Schema
	}
}
