package bunker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"axiom/pkg/ui/styles"
)

const defaultBuildContainerName = "axiom-build"

type EnvConfig struct {
	AxiomPath  string
	GitUser    string
	GitEmail   string
	GitToken   string
	AuthMode   string
	BaseDir    string
	OllamaHost string
	ModelsDir  string
	GPUType    string
	GFXVal     string
	ROCMMode   string
}

// Manager centraliza el ciclo de vida del paquete bunker.
// main.go solo debería resolver el comando y delegar aquí.
type Manager struct {
	rootDir            string
	buildContainerName string
}

func NewManager(rootDir string) *Manager {
	return &Manager{
		rootDir:            rootDir,
		buildContainerName: defaultBuildContainerName,
	}
}

// Run actúa como orquestador público del paquete.
// Cada comando de bunker se resuelve aquí y luego se delega a su implementación.
func (m *Manager) Run(command string, args []string) error {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "help":
		return m.Help()
	case "build":
		return m.Build()
	case "list":
		return m.List()
	case "create":
		return m.Create(firstArg(args))
	case "delete", "eliminar":
		return m.Delete(firstArg(args))
	case "delete-image", "image-delete", "prune-images":
		return m.DeleteImage()
	default:
		return fmt.Errorf("comando bunker no soportado todavía: %s", command)
	}
}

func KnownCommand(command string) bool {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case "help", "build", "list", "create", "delete", "eliminar", "delete-image", "image-delete", "prune-images":
		return true
	default:
		return false
	}
}

// Help muestra los comandos disponibles del orquestador bunker.
func (m *Manager) Help() error {
	fmt.Println(styles.GetLogo())
	fmt.Println(styles.RenderBunkerCard(
		"AXIOM Help",
		"Comandos disponibles del búnker en la versión Go actual.",
		[]styles.BunkerDetail{
			{Label: "Build", Value: "Construye la imagen base: axiom build"},
			{Label: "List", Value: "Lista los búnkeres detectados: axiom list"},
			{Label: "Create", Value: "Crea o abre un búnker: axiom create <nombre>"},
			{Label: "Delete", Value: "Elimina un búnker por nombre o selector: axiom delete [nombre]"},
			{Label: "Eliminar", Value: "Alias en español de delete: axiom eliminar [nombre]"},
			{Label: "Delete Image", Value: "Elimina la imagen base activa: axiom delete-image"},
		},
		[]string{
			"Si no pasas nombre en delete/eliminar, se abre un selector con flechas.",
			"Al borrar un búnker puedes decidir si también se elimina el código del proyecto.",
			"delete-image también muestra las imágenes de AXIOM detectadas antes y después.",
		},
		"Siguiente paso: seguir portando info, stop, rebuild, prune y reset.",
	))
	return nil
}

func (m *Manager) LoadConfig() (EnvConfig, error) {
	return LoadEnvFile(filepath.Join(m.rootDir, ".env"))
}

// LoadEnvFile parsea el .env mínimo que necesita el orquestador Go.
// No intenta ser un parser genérico: solo mapea las claves que AXIOM usa realmente.
func LoadEnvFile(path string) (EnvConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return EnvConfig{}, err
	}
	defer file.Close()

	cfg := EnvConfig{}
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
		}
	}
	if err := scanner.Err(); err != nil {
		return EnvConfig{}, err
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
		return EnvConfig{}, fmt.Errorf("AXIOM_BASE_DIR no está definido en %s", path)
	}
	if cfg.ModelsDir == "" {
		cfg.ModelsDir = filepath.Join(cfg.BaseDir, "ai_config", "models")
	}
	if cfg.OllamaHost == "" {
		cfg.OllamaHost = "http://localhost:11434"
	}

	return cfg, nil
}

func (c EnvConfig) BuildWorkspaceDir(containerName string) string {
	return filepath.Join(c.BaseDir, ".entorno", containerName)
}

func (c EnvConfig) AIConfigDir() string {
	return filepath.Join(c.BaseDir, "ai_config")
}

func (c EnvConfig) TutorPath() string {
	return filepath.Join(c.AIConfigDir(), "teams", "tutor.md")
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return strings.TrimSpace(args[0])
}
