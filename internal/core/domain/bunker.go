// Package domain contiene los modelos puros del negocio.
// No tiene dependencias externas - es la capa más interna de la Clean Architecture.
package domain

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

// BuildConfig contiene la configuración para crear un búnker.
type BuildConfig struct {
	Name           string
	Image          string
	ProjectDir     string
	EnvironmentDir string
	GPUType        string
	Flags          string
}
