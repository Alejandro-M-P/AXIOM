package base

import "os/exec"

// OSType representa el tipo de sistema operativo
type OSType int

const (
	OSUnknown OSType = iota
	OSArch
	OSUbuntu
	OSMacOS
)

// String returns the string representation of OSType.
func (t OSType) String() string {
	switch t {
	case OSArch:
		return "arch"
	case OSUbuntu:
		return "ubuntu"
	case OSMacOS:
		return "macos"
	default:
		return "unknown"
	}
}

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

// IsCommandAvailable returns true if the given command is available in the system PATH.
func (d *OSDetector) IsCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
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
