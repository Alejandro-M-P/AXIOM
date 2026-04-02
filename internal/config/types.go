package config

// EnvConfig represents the configuration of the AXIOM environment.
// Loaded from the config.toml file at the project root.
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

// GPUInfo represents GPU hardware information.
type GPUInfo struct {
	Type       string
	GfxVal     string
	Name       string
	RawGfx     string
	PCIAddress string
	VendorID   string
	DeviceID   string
}
