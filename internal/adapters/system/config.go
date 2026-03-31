package system

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/Alejandro-M-P/AXIOM/internal/domain"
	"github.com/BurntSushi/toml"
)

// Config guarda los datos que recolectamos en el formulario TUI.
// Se persiste como config.toml en la raíz del proyecto AXIOM.
type Config struct {
	// --- RUTAS ---
	AxiomPath string `toml:"axiom_path"`

	// --- IDENTIDAD Y AUTENTICACIÓN ---
	GitUser  string `toml:"git_user"`
	GitEmail string `toml:"git_email"`
	GitToken string `toml:"git_token"`
	AuthMode string `toml:"auth_mode"` // "ssh" o "https"

	// --- DIRECTORIOS CORE ---
	BaseDir    string `toml:"base_dir"`
	OllamaHost string `toml:"ollama_host"`
	ModelsDir  string `toml:"models_dir"`

	// --- HARDWARE ---
	GfxVersion string `toml:"gfx_version"` // Valor GFX para AMD (ej: gfx1100)
	GpuType    string `toml:"gpu_type"`    // amd, nvidia o intel
	RocmMode   string `toml:"rocm_mode"`   // host o image

	// --- FUTURO: ECOSISTEMA Y TUI (WIP) ---
	Language    string `toml:"language"`     // "es" o "en"
	Theme       string `toml:"theme"`        // Tema de la TUI
	CatalogPath string `toml:"catalog_path"` // Ruta al catalog.toml
	LogLevel    string `toml:"log_level"`    // "debug", "info", "warn", "error"
}

// ToEnvConfig convierte Config del adapter a EnvConfig del dominio.
func (c Config) ToEnvConfig() domain.EnvConfig {
	return domain.EnvConfig{
		AxiomPath:  c.AxiomPath,
		GitUser:    c.GitUser,
		GitEmail:   c.GitEmail,
		GitToken:   c.GitToken,
		AuthMode:   c.AuthMode,
		BaseDir:    c.BaseDir,
		OllamaHost: c.OllamaHost,
		ModelsDir:  c.ModelsDir,
		GPUType:    c.GpuType,
		GFXVal:     c.GfxVersion,
		ROCMMode:   c.RocmMode,
		Language:   c.Language,
	}
}

// ConfigPath retorna la ruta del archivo de configuración.
func ConfigPath(axiomPath string) string {
	return filepath.Join(axiomPath, "config.toml")
}

// Save escribe el archivo config.toml con permisos 600 (Seguridad AXIOM).
func (c Config) Save(axiomPath string) error {
	path := ConfigPath(axiomPath)

	// Aplicar defaults antes de serializar
	cfg := c
	if cfg.Language == "" {
		cfg.Language = "es"
	}
	if cfg.Theme == "" {
		cfg.Theme = "default"
	}
	if cfg.CatalogPath == "" {
		cfg.CatalogPath = filepath.Join(axiomPath, "catalog.toml")
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.OllamaHost == "" {
		cfg.OllamaHost = "http://localhost:11434"
	}

	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	return os.WriteFile(path, buf.Bytes(), 0600)
}

// LoadConfig lee el archivo config.toml y retorna un Config.
func LoadConfig(fs interface{}, axiomPath string) (Config, error) {
	path := ConfigPath(axiomPath)

	fileSystem, ok := fs.(interface {
		ReadFile(path string) ([]byte, error)
		Exists(path string) bool
	})
	if !ok {
		return Config{}, nil
	}

	if !fileSystem.Exists(path) {
		return Config{}, nil
	}

	data, err := fileSystem.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
