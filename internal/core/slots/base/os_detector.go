package base

import (
	"fmt"
	"os/exec"
)

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

// Detect retorna el tipo de OS, nombre, y error
func (d *OSDetector) Detect() (OSType, string, error) {
	// Check for Arch Linux
	if _, err := exec.LookPath("pacman"); err == nil {
		return OSArch, "Arch Linux", nil
	}
	// Check for Ubuntu/Debian
	if _, err := exec.LookPath("apt"); err == nil {
		return OSUbuntu, "Ubuntu/Debian", nil
	}
	// Check for macOS
	if _, err := exec.LookPath("brew"); err == nil {
		return OSMacOS, "macOS", nil
	}
	return OSUnknown, "Unknown", fmt.Errorf("unsupported OS")
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
