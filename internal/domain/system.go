// Package domain contiene los modelos puros del negocio.
// No tiene dependencias externas - es la capa más interna de la Clean Architecture.
package domain

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
	Language   string
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
