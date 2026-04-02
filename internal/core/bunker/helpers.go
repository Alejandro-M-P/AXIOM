package bunker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/Alejandro-M-P/AXIOM/internal/config"
	"github.com/Alejandro-M-P/AXIOM/internal/i18n"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// EnvConfig is an alias to config.EnvConfig for convenience within the bunker package.
type EnvConfig = config.EnvConfig

// GPUInfo is an alias to config.GPUInfo for convenience within the bunker package.
type GPUInfo = config.GPUInfo

// sanitizeBunkerName valida y limpia el nombre de un búnker.
func sanitizeBunkerName(name string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(name))
	if clean == "." || clean == ".." || strings.Contains(clean, "/") || strings.Contains(clean, "\\") {
		return "", fmt.Errorf("errors.bunker.invalid_name")
	}
	return clean, nil
}

// formatBytes formatea bytes a formato legible (KB, MB, GB, etc.).
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf(i18n.Commands["bunker"]["bytes_format"], bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf(i18n.Commands["bunker"]["bytes_format_decimal"], float64(bytes)/float64(div), "KMGTPE"[exp])
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
func humanPath(fs ports.IFileSystem, path string) string {
	home, err := fs.UserHomeDir()
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
func removeProjectPath(fs ports.IFileSystem, path string) error {
	_, err := fs.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	info, err := fs.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("errors.bunker.not_dir")
	}
	return removePathWritable(fs, path)
}

// removePathWritable elimina un path haciéndolo escribible primero.
func removePathWritable(fs ports.IFileSystem, path string) error {
	_, err := fs.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	_ = fs.WalkDir(path, func(currentPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		mode := info.Mode()
		if mode&0200 == 0 {
			_ = fs.Chmod(currentPath, mode|0200)
		}
		return nil
	})
	return fs.RemoveAll(path)
}

// baseImageName retorna el nombre de la imagen base según el tipo de GPU.
func baseImageName(gpuType string) string {
	gpuType = strings.TrimSpace(gpuType)
	if gpuType == "" {
		gpuType = "generic"
	}
	return fmt.Sprintf(i18n.Commands["bunker"]["image_name"], gpuType)
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
func bunkerEnvSize(fs ports.IFileSystem, cfg EnvConfig, name string) string {
	path := config.BuildWorkspaceDir(cfg.BaseDir, name)
	if _, err := fs.Stat(path); err != nil {
		return "-"
	}
	var size int64
	err := fs.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
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
func bunkerGitBranch(fs ports.IFileSystem, cfg EnvConfig, name string) string {
	projectPath := filepath.Join(cfg.BaseDir, name)
	headPath := filepath.Join(projectPath, ".git", "HEAD")
	content, err := fs.ReadFile(headPath)
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
func bunkerLastEntry(fs ports.IFileSystem, cfg EnvConfig, name string) string {
	path := config.BuildWorkspaceDir(cfg.BaseDir, name)
	info, err := fs.Stat(path)
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
	return config.BuildWorkspaceDir(cfg.BaseDir, name)
}

// sshVolumeFlag retorna el flag de volumen SSH si hay un agent SSH activo.
// El socketPath viene del sistema (ISystem.SSHAgentSocket).
func sshVolumeFlag(socketPath string) string {
	if socketPath == "" {
		return ""
	}
	return fmt.Sprintf(i18n.Commands["bunker"]["volume_ssh"], socketPath, socketPath)
}

// ensureTutorFile asegura que el archivo tutor exista.
func ensureTutorFile(fs ports.IFileSystem, path string) error {
	if _, err := fs.Stat(path); err == nil {
		return nil
	}
	file, err := fs.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

// LoadConfig lee el archivo config.toml y retorna una configuración.
func LoadConfig(fs ports.IFileSystem, axiomPath string) (config.EnvConfig, error) {
	path := filepath.Join(axiomPath, "config.toml")

	data, err := fs.ReadFile(path)
	if err != nil {
		// Si el archivo no existe, retornamos config vacía (no es error)
		return config.EnvConfig{}, nil
	}

	var cfg struct {
		AxiomPath  string `toml:"axiom_path"`
		GitUser    string `toml:"git_user"`
		GitEmail   string `toml:"git_email"`
		GitToken   string `toml:"git_token"`
		AuthMode   string `toml:"auth_mode"`
		BaseDir    string `toml:"base_dir"`
		OllamaHost string `toml:"ollama_host"`
		ModelsDir  string `toml:"models_dir"`
		GpuType    string `toml:"gpu_type"`
		GfxVersion string `toml:"gfx_version"`
		RocmMode   string `toml:"rocm_mode"`
		Language   string `toml:"language"`
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return config.EnvConfig{}, err
	}

	result := config.EnvConfig{
		AxiomPath:  cfg.AxiomPath,
		GitUser:    cfg.GitUser,
		GitEmail:   cfg.GitEmail,
		GitToken:   cfg.GitToken,
		AuthMode:   cfg.AuthMode,
		BaseDir:    cfg.BaseDir,
		OllamaHost: cfg.OllamaHost,
		ModelsDir:  cfg.ModelsDir,
		GPUType:    cfg.GpuType,
		GFXVal:     cfg.GfxVersion,
		ROCMMode:   cfg.RocmMode,
		Language:   cfg.Language,
	}

	// Convertir BaseDir a ruta absoluta
	if result.BaseDir == "" || result.BaseDir == "." {
		cwd, err := os.Getwd()
		if err != nil {
			return result, fmt.Errorf("errors.bunker.failed_cwd: %w", err)
		}
		result.BaseDir = cwd
	} else {
		absPath, err := filepath.Abs(result.BaseDir)
		if err != nil {
			return result, fmt.Errorf("errors.bunker.failed_abs_path: %w", err)
		}
		result.BaseDir = absPath
	}

	return result, nil
}
