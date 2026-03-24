package install

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config guarda los datos que recolectamos en el formulario TUI.
// Los campos deben coincidir con lo que capturamos en pkg/ui/form.go.
type Config struct {
	GitUser    string
	GitEmail   string
	GitToken   string
	AuthMode   string // "ssh" o "https"
	BaseDir    string
	ModelsDir  string
	GfxVersion string // Valor GFX para AMD (ej: gfx1100)
	GpuType    string // amd, nvidia o intel
	RocmMode   string // host o image
}

// Save escribe el archivo .env con permisos 600 (Seguridad AXIOM).
// Este archivo será el cerebro que lea el comando 'axiom build'.
func (c Config) Save(axiomPath string) error {
	envPath := filepath.Join(axiomPath, ".env")

	// Usamos un template estrictamente ordenado para evitar desfases en el archivo.
	content := fmt.Sprintf(`# ─── RUTAS AXIOM ─────────────────────────────────
AXIOM_PATH="%s"

# ─── IDENTIDAD GIT ───────────────────────────────
AXIOM_GIT_USER="%s"
AXIOM_GIT_EMAIL="%s"
AXIOM_GIT_TOKEN="%s"
AXIOM_AUTH_MODE="%s"

# ─── DIRECTORIOS DE TRABAJO ──────────────────────
AXIOM_BASE_DIR="%s"
AXIOM_OLLAMA_HOST="http://localhost:11434"
AXIOM_MODELS_DIR="%s"

# ─── HARDWARE & DRIVERS ─────────────────────────
AXIOM_GPU_TYPE="%s"
AXIOM_GFX_VAL="%s"
AXIOM_ROCM_MODE="%s"`,
		axiomPath,    // 1. AXIOM_PATH
		c.GitUser,    // 2. AXIOM_GIT_USER
		c.GitEmail,   // 3. AXIOM_GIT_EMAIL
		c.GitToken,   // 4. AXIOM_GIT_TOKEN
		c.AuthMode,   // 5. AXIOM_AUTH_MODE
		c.BaseDir,    // 6. AXIOM_BASE_DIR
		c.ModelsDir,  // 7. AXIOM_MODELS_DIR
		c.GpuType,    // 8. AXIOM_GPU_TYPE
		c.GfxVersion, // 9. AXIOM_GFX_VAL
		c.RocmMode,   // 10. AXIOM_ROCM_MODE
	)

	// El permiso 0600 (rw-------) es obligatorio para proteger el GIT_TOKEN.
	return os.WriteFile(envPath, []byte(content), 0600)
}