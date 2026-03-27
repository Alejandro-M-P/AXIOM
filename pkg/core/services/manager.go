// Package bunker contiene la lógica de negocio pura.
// Utiliza inyección de dependencias para recibir implementaciones concretas de los puertos.
package bunker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"axiom/pkg/core/domain"
	"axiom/pkg/core/ports"
)

const defaultBuildContainerName = "axiom-build"

// Manager es el servicio principal que orquesta el ciclo de vida de los bunkers.
// Recibe las dependencias (adapters) en su constructor, permitiendo testing y flexibilidad.
type Manager struct {
	rootDir            string
	buildContainerName string

	// Dependencias inyectadas (puertos)
	Runtime ports.IContainerRuntime
	FS      ports.IFileSystem
	System  ports.ISystem
	UI      ports.IPresenter
}

// NewManager crea una nueva instancia del Manager con sus dependencias.
// Ejemplo de uso:
//
//	podmanAdapter := podmanadapter.NewPodmanAdapter()
//	fsAdapter := fsadapter.NewFSAdapter()
//	systemAdapter := systemadapter.NewSystemAdapter()
//	uiAdapter := uiadapter.NewConsoleUI()
//
//	manager := services.NewManager(rootDir, podmanAdapter, fsAdapter, systemAdapter, uiAdapter)
func NewManager(
	rootDir string,
	runtime ports.IContainerRuntime,
	fs ports.IFileSystem,
	system ports.ISystem,
	ui ports.IPresenter,
) *Manager {
	return &Manager{
		rootDir:            rootDir,
		buildContainerName: defaultBuildContainerName,
		Runtime:            runtime,
		FS:                 fs,
		System:             system,
		UI:                 ui,
	}
}

// LoadConfig carga la configuración desde el archivo .env
func (m *Manager) LoadConfig() (domain.EnvConfig, error) {
	return LoadEnvFile(m.FS, filepath.Join(m.rootDir, ".env"))
}

// LoadEnvFile parsea el archivo .env y retorna una configuración.
func LoadEnvFile(fs ports.IFileSystem, path string) (domain.EnvConfig, error) {
	file, err := fs.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return domain.EnvConfig{}, err
	}
	defer file.Close()

	cfg := domain.EnvConfig{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), "\"")

		switch key {
		case "AXIOM_PATH":
			cfg.AxiomPath = value
		case "AXIOM_GIT_USER":
			cfg.GitUser = value
		case "AXIOM_GIT_EMAIL":
			cfg.GitEmail = value
		case "AXIOM_GIT_TOKEN":
			cfg.GitToken = value
		case "AXIOM_AUTH_MODE":
			cfg.AuthMode = value
		case "AXIOM_BASE_DIR":
			cfg.BaseDir = value
		case "AXIOM_OLLAMA_HOST":
			cfg.OllamaHost = value
		case "AXIOM_MODELS_DIR":
			cfg.ModelsDir = value
		case "AXIOM_GPU_TYPE":
			cfg.GPUType = value
		case "AXIOM_GFX_VAL":
			cfg.GFXVal = value
		case "AXIOM_ROCM_MODE":
			cfg.ROCMMode = value
		case "AXIOM_LANGUAGE":
			cfg.Language = value
		}
	}
	if err := scanner.Err(); err != nil {
		return domain.EnvConfig{}, err
	}

	if cfg.AxiomPath == "" {
		cfg.AxiomPath = filepath.Dir(path)
	}
	if cfg.AuthMode == "" {
		cfg.AuthMode = "https"
	}
	if cfg.ROCMMode == "" {
		cfg.ROCMMode = "host"
	}
	if cfg.BaseDir == "" {
		return domain.EnvConfig{}, fmt.Errorf("missing_base_dir")
	}
	if cfg.ModelsDir == "" {
		cfg.ModelsDir = filepath.Join(cfg.BaseDir, "ai_config", "models")
	}
	if cfg.OllamaHost == "" {
		cfg.OllamaHost = "http://localhost:11434"
	}

	return cfg, nil
}

// GetBuildContainerName retorna el nombre del contenedor de build.
func (m *Manager) GetBuildContainerName() string {
	return m.buildContainerName
}

// SetBuildContainerName establece el nombre del contenedor de build.
func (m *Manager) SetBuildContainerName(name string) {
	m.buildContainerName = name
}
