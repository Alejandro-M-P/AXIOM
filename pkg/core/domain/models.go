// Package domain contiene los modelos puros del negocio.
// No tiene dependencias externas - es la capa más interna de la Clean Architecture.
package domain

import "path/filepath"

// EnvConfig representa la configuración del entorno AXIOM.
// Se carga desde el archivo .env en la raíz del proyecto.
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

// BuildWorkspaceDir retorna la ruta al directorio de trabajo del búnker.
func (c EnvConfig) BuildWorkspaceDir(containerName string) string {
	return filepath.Join(c.BaseDir, ".entorno", containerName)
}

// AIConfigDir retorna la ruta al directorio de configuración de AI.
func (c EnvConfig) AIConfigDir() string {
	return filepath.Join(c.BaseDir, "ai_config")
}

// TutorPath retorna la ruta al archivo tutor.md.
func (c EnvConfig) TutorPath() string {
	return filepath.Join(c.AIConfigDir(), "teams", "tutor.md")
}

// Bunker representa un búnker (contenedor de desarrollo) en el sistema.
type Bunker struct {
	Name            string
	Status          string // "running", "stopped", "created", "exited"
	Image           string
	GPUType         string
	ProjectPath     string
	EnvironmentPath string
	Size            string
	LastActivity    string
	GitBranch       string
}

// ContainerInfo representa información de un contenedor.
type ContainerInfo struct {
	Name   string
	Status string
	Image  string
}

// GPUInfo representa la información de la GPU del sistema.
type GPUInfo struct {
	Type       string
	GfxVal     string
	Name       string
	RawGfx     string
	PCIAddress string
	VendorID   string
	DeviceID   string
}

// BuildConfig contiene la configuración para crear un búnker.
type BuildConfig struct {
	Name           string
	Image          string
	ProjectDir     string
	EnvironmentDir string
	GPUType        string
	Flags          string
}

// Field representa un campo para mostrar en la UI.
type Field struct {
	Label string
	Value string
}
