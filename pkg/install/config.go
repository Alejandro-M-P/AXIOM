package install

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config guarda los datos que recolectamos en el formulario
type Config struct {
	GitUser    string
	GitEmail   string
	GitToken   string
	AuthMode   string // ssh o https
	BaseDir    string
	ModelsDir  string
	GfxVersion string
	RocmMode   string // host o image
}

// Save escribe el archivo .env con permisos 600 (Seguridad AXIOM)
func (c Config) Save(axiomPath string) error {
	envPath := filepath.Join(axiomPath, ".env")
	
	// Contenido exacto de tu install.sh original
	content := fmt.Sprintf(`# ─── RUTAS AXIOM ─────────────────────────────────
AXIOM_PATH="%s"
# ─── IDENTIDAD GIT ───────────────────────────────
AXIOM_GIT_USER="%s"
AXIOM_GIT_EMAIL="%s"
AXIOM_GIT_TOKEN="%s"
AXIOM_AUTH_MODE="%s"
AXIOM_BASE_DIR="%s"
AXIOM_OLLAMA_HOST="http://localhost:11434"
AXIOM_MODELS_DIR="%s"
AXIOM_ROCM_MODE="%s"
AXIOM_GPU_TYPE=""
AXIOM_GFX_VAL="%s"`, 
		axiomPath, c.GitUser, c.GitEmail, c.GitToken, c.AuthMode, 
		c.BaseDir, c.ModelsDir, c.RocmMode, c.GfxVersion)

	// En Go, 0600 asegura que solo tú puedas leer tus tokens
	return os.WriteFile(envPath, []byte(content), 0600)
}