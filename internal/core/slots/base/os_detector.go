package base

// OSType representa el tipo de sistema operativo
type OSType int

const (
	OSUnknown OSType = iota
	OSArch
	OSUbuntu
	OSMacOS
)

// OSDetector detecta el sistema operativo
type OSDetector struct{}

// NewOSDetector crea un nuevo detector
func NewOSDetector() *OSDetector {
	return &OSDetector{}
}

// Detect retorna el tipo de OS (siempre Arch Linux - AXIOM solo corre en Arch)
func (d *OSDetector) Detect() (OSType, string, error) {
	return OSArch, "Arch Linux", nil
}

// GetPackageManager returns the package manager name for the given OS type
func GetPackageManager(osType OSType) string {
	switch osType {
	case OSArch:
		return "pacman"
	case OSUbuntu:
		return "apt"
	case OSMacOS:
		return "brew"
	default:
		return ""
	}
}
