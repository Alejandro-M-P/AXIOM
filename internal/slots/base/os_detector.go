// Package base provides OS detection and base tool installation functionality.
// This package handles automatic installation of system package managers and base tools
// without user intervention in the wizard.
package base

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// OSType represents the detected operating system type
type OSType string

const (
	OSArch    OSType = "arch"
	OSUbuntu  OSType = "ubuntu"
	OSMacOS   OSType = "macos"
	OSUnknown OSType = "unknown"
)

// OSDetector handles operating system detection
type OSDetector struct {
	execCommand func(name string, arg ...string) *exec.Cmd
}

// NewOSDetector creates a new OS detector with the default exec.Command
func NewOSDetector() *OSDetector {
	return &OSDetector{
		execCommand: exec.Command,
	}
}

// NewOSDetectorWithExec creates a new OS detector with a custom exec function (for testing)
func NewOSDetectorWithExec(execFunc func(name string, arg ...string) *exec.Cmd) *OSDetector {
	return &OSDetector{
		execCommand: execFunc,
	}
}

// Detect determines the current operating system
// Returns the OS type and a human-readable name
func (d *OSDetector) Detect() (OSType, string, error) {
	goos := runtime.GOOS

	switch goos {
	case "darwin":
		return OSMacOS, "macOS", nil
	case "linux":
		return d.detectLinux()
	default:
		return OSUnknown, fmt.Sprintf("errors.slots.base.os_detector.unknown_os: %s", goos), fmt.Errorf("errors.slots.base.unsupported_os: %s", goos)
	}
}

// detectLinux determines the specific Linux distribution
func (d *OSDetector) detectLinux() (OSType, string, error) {
	// Try /etc/os-release first (standard on most Linux distros)
	osType, name, err := d.detectFromOSRelease()
	if err == nil {
		return osType, name, nil
	}

	// Fallback: check for pacman (Arch)
	if d.isCommandAvailable("pacman") {
		return OSArch, "Arch Linux", nil
	}

	// Fallback: check for apt (Debian/Ubuntu)
	if d.isCommandAvailable("apt-get") || d.isCommandAvailable("apt") {
		return OSUbuntu, "Ubuntu/Debian", nil
	}

	return OSUnknown, "errors.slots.base.os_detector.unknown_linux", fmt.Errorf("errors.slots.base.os_detector.unknown_linux")
}

// detectFromOSRelease reads /etc/os-release to determine the distribution
func (d *OSDetector) detectFromOSRelease() (OSType, string, error) {
	cmd := d.execCommand("cat", "/etc/os-release")
	output, err := cmd.Output()
	if err != nil {
		return OSUnknown, "", err
	}

	content := string(output)
	contentLower := strings.ToLower(content)

	// Check for Arch Linux variants
	if strings.Contains(contentLower, "arch") {
		return OSArch, "Arch Linux", nil
	}

	// Check for Ubuntu
	if strings.Contains(contentLower, "ubuntu") {
		return OSUbuntu, "Ubuntu", nil
	}

	// Check for Debian
	if strings.Contains(contentLower, "debian") {
		return OSUbuntu, "Debian", nil
	}

	// Check ID_LIKE for derivatives
	if strings.Contains(contentLower, "id_like=arch") || strings.Contains(contentLower, "id_like=archlinux") {
		return OSArch, "Arch-based Linux", nil
	}

	if strings.Contains(contentLower, "id_like=debian") || strings.Contains(contentLower, "id_like=ubuntu") {
		return OSUbuntu, "Debian-based Linux", nil
	}

	return OSUnknown, "", fmt.Errorf("errors.slots.base.os_detector.os_not_recognized")
}

// isCommandAvailable checks if a command exists in the system PATH
func (d *OSDetector) isCommandAvailable(name string) bool {
	return d.IsCommandAvailable(name)
}

// IsCommandAvailable checks if a command exists in the system PATH (public version)
func (d *OSDetector) IsCommandAvailable(name string) bool {
	cmd := d.execCommand("which", name)
	err := cmd.Run()
	return err == nil
}

// GetPackageManager returns the package manager command for the detected OS
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

// String returns the string representation of the OS type
func (o OSType) String() string {
	return string(o)
}
