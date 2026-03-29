package bunker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"axiom/internal/domain"
	"axiom/internal/ports"
)

// EnvConfig alias para domain.EnvConfig para evitar imports múltiples
type EnvConfig = domain.EnvConfig

// GPUInfo alias para domain.GPUInfo
type GPUInfo = domain.GPUInfo

// sanitizeBunkerName valida y limpia el nombre de un búnker.
func sanitizeBunkerName(name string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(name))
	if clean == "." || clean == ".." || strings.Contains(clean, "/") || strings.Contains(clean, "\\") {
		return "", fmt.Errorf("invalid_name")
	}
	return clean, nil
}

// formatBytes formatea bytes a formato legible (KB, MB, GB, etc.).
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// yesNo convierte un booleano a string de sí/no internacionalizado.
func yesNo(value bool) string {
	if value {
		return "common.yes"
	}
	return "common.no"
}

// isYes verifica si un string representa "sí".
func isYes(value string) bool {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	return trimmed == "s" || trimmed == "y" || trimmed == "si" || trimmed == "sí" || trimmed == "yes"
}

// humanPath convierte un path a formato legible con ~ para el home.
func humanPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return strings.Replace(path, home, "~", 1)
	}
	return path
}

// defaultString retorna el valor por defecto si el valor está vacío.
func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// bunkerTimestamp formatea un time.Time a string de fecha.
func bunkerTimestamp(t time.Time) string {
	return t.Format("2006-01-02")
}

// appendTutorLog agrega una línea al log del tutor.
func appendTutorLog(line string) error {
	// Placeholder - en la implementación real se escribiría al archivo
	return nil
}

// removeProjectPath elimina el path del proyecto si es un directorio.
func removeProjectPath(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not_dir")
	}
	return removePathWritable(path)
}

// removePathWritable elimina un path haciéndolo escribible primero.
func removePathWritable(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	_ = filepath.WalkDir(path, func(currentPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		mode := info.Mode()
		if mode&0200 == 0 {
			_ = os.Chmod(currentPath, mode|0200)
		}
		return nil
	})
	return os.RemoveAll(path)
}

// baseImageName retorna el nombre de la imagen base según el tipo de GPU.
func baseImageName(gpuType string) string {
	gpuType = strings.TrimSpace(gpuType)
	if gpuType == "" {
		gpuType = "generic"
	}
	return fmt.Sprintf("localhost/axiom-%s:latest", gpuType)
}

// resolveBuildGPU detecta la GPU para builds.
// Placeholder - en producción se usaría gpu.Detect() del paquete adapters.
func resolveBuildGPU(cfg EnvConfig) GPUInfo {
	return GPUInfo{
		Type:   "generic",
		GfxVal: cfg.GFXVal,
		Name:   "Detected",
	}
}

// bunkerEnvSize calcula el tamaño del directorio de entorno.
func bunkerEnvSize(cfg EnvConfig, name string) string {
	path := cfg.BuildWorkspaceDir(name)
	if _, err := os.Stat(path); err != nil {
		return "-"
	}
	var size int64
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				size += info.Size()
			}
		}
		return nil
	})
	if err != nil || size == 0 {
		return "-"
	}
	return formatBytes(size)
}

// bunkerGitBranch obtiene la rama git del proyecto.
func bunkerGitBranch(cfg EnvConfig, name string) string {
	projectPath := filepath.Join(cfg.BaseDir, name)
	headPath := filepath.Join(projectPath, ".git", "HEAD")
	content, err := os.ReadFile(headPath)
	if err != nil {
		return "-"
	}
	head := strings.TrimSpace(string(content))
	if strings.HasPrefix(head, "ref: refs/heads/") {
		return strings.TrimPrefix(head, "ref: refs/heads/")
	}
	if len(head) >= 7 {
		return head[:7]
	}
	return "-"
}

// bunkerLastEntry obtiene la última fecha de modificación del entorno.
func bunkerLastEntry(cfg EnvConfig, name string) string {
	path := cfg.BuildWorkspaceDir(name)
	info, err := os.Stat(path)
	if err != nil {
		return "-"
	}
	return info.ModTime().Format("2006-01-02")
}

// bunkerProjectPath retorna el path del proyecto.
func bunkerProjectPath(cfg EnvConfig, name string) string {
	return filepath.Join(cfg.BaseDir, name)
}

// bunkerEnvPath retorna el path del entorno.
func bunkerEnvPath(cfg EnvConfig, name string) string {
	return cfg.BuildWorkspaceDir(name)
}

// sshVolumeFlag retorna el flag de volumen SSH si hay un agent SSH activo.
func sshVolumeFlag() string {
	sock := strings.TrimSpace(os.Getenv("SSH_AUTH_SOCK"))
	if sock == "" {
		return ""
	}
	info, err := os.Stat(sock)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		return ""
	}
	return fmt.Sprintf("--volume %s:%s", sock, sock)
}

// ensureTutorFile asegura que el archivo tutor exista.
func ensureTutorFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}

// LoadEnvFile parsea el archivo .env y retorna una configuración.
func LoadEnvFile(fs interface{}, path string) (EnvConfig, error) {
	fileSystem, ok := fs.(ports.IFileSystem)
	if !ok {
		return EnvConfig{}, fmt.Errorf("invalid filesystem interface")
	}

	data, err := fileSystem.ReadFile(path)
	if err != nil {
		// Si el archivo no existe, retornamos config vacía (no es error)
		return EnvConfig{}, nil
	}

	cfg := EnvConfig{}
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Ignorar líneas vacías y comentarios
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parsear formato KEY="value" o KEY=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remover comillas si existen
		if len(value) >= 2 && strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = value[1 : len(value)-1]
		}

		// Mapear las variables de entorno a los campos de EnvConfig
		switch key {
		case "AXIOM_PATH":
			cfg.AxiomPath = value
		case "AXIOM_BASE_DIR":
			cfg.BaseDir = value
		case "AXIOM_GIT_USER":
			cfg.GitUser = value
		case "AXIOM_GIT_EMAIL":
			cfg.GitEmail = value
		case "AXIOM_GIT_TOKEN":
			cfg.GitToken = value
		case "AXIOM_AUTH_MODE":
			cfg.AuthMode = value
		case "AXIOM_GPU_TYPE":
			cfg.GPUType = value
		case "AXIOM_GFX_VAL":
			cfg.GFXVal = value
		case "AXIOM_ROCM_MODE":
			cfg.ROCMMode = value
		}
	}

	// Convertir BaseDir a ruta absoluta
	if cfg.BaseDir == "" || cfg.BaseDir == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return cfg, fmt.Errorf("failed to get current working directory: %w", err)
		}
		cfg.BaseDir = cwd
	} else {
		absPath, err := filepath.Abs(cfg.BaseDir)
		if err != nil {
			return cfg, fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		cfg.BaseDir = absPath
	}

	return cfg, nil
}
