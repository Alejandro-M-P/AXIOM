// Package base provides OS detection and base tool installation functionality.
package base

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
)

// OSPreferences represents the structure of the os_preferences.toml file
type OSPreferences struct {
	Arch   OSConfig           `toml:"arch"`
	Ubuntu OSConfig           `toml:"ubuntu"`
	MacOS  OSConfig           `toml:"macos"`
	Tools  map[string]ToolDef `toml:"tools"`
}

// OSConfig represents configuration for a specific OS
type OSConfig struct {
	Name            string   `toml:"name"`
	PackageManager  string   `toml:"package_manager"`
	BaseTools       []string `toml:"base_tools"`
	InstallCommands []string `toml:"install_commands"`
}

// ToolDef represents installation commands for a tool across different OSes
type ToolDef struct {
	Arch   string `toml:"arch"`
	Ubuntu string `toml:"ubuntu"`
	MacOS  string `toml:"macos"`
}

// BaseInstaller handles automatic installation of base tools and package managers
type BaseInstaller struct {
	detector    *OSDetector
	preferences *OSPreferences
	execFunc    func(ctx context.Context, name string, arg ...string) *exec.Cmd
	osType      OSType
	osName      string
}

// NewBaseInstaller creates a new base installer with the default OS detector
func NewBaseInstaller(preferencesPath string) (*BaseInstaller, error) {
	prefs, err := loadPreferences(preferencesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OS preferences: %w", err)
	}

	detector := NewOSDetector()
	osType, osName, err := detector.Detect()
	if err != nil {
		// Don't fail if we can't detect OS, we'll handle it gracefully
		osType = OSUnknown
		osName = "Unknown"
	}

	return &BaseInstaller{
		detector:    detector,
		preferences: prefs,
		execFunc:    exec.CommandContext,
		osType:      osType,
		osName:      osName,
	}, nil
}

// NewBaseInstallerWithDeps creates a base installer with custom dependencies (for testing)
func NewBaseInstallerWithDeps(prefs *OSPreferences, detector *OSDetector, execFunc func(ctx context.Context, name string, arg ...string) *exec.Cmd) *BaseInstaller {
	osType, osName, _ := detector.Detect()
	return &BaseInstaller{
		detector:    detector,
		preferences: prefs,
		execFunc:    execFunc,
		osType:      osType,
		osName:      osName,
	}
}

// loadPreferences loads the OS preferences from a TOML file
func loadPreferences(path string) (*OSPreferences, error) {
	var prefs OSPreferences
	if _, err := toml.DecodeFile(path, &prefs); err != nil {
		return nil, err
	}
	return &prefs, nil
}

// GetOSType returns the detected OS type
func (b *BaseInstaller) GetOSType() OSType {
	return b.osType
}

// GetOSName returns the detected OS name
func (b *BaseInstaller) GetOSName() string {
	return b.osName
}

// IsToolInstalled checks if a tool (command) is available in the system
func (b *BaseInstaller) IsToolInstalled(tool string) bool {
	cmd := exec.Command("which", tool)
	err := cmd.Run()
	return err == nil
}

// InstallBaseTools installs the base tools defined in the OS preferences
// This should be called once at the beginning of the installation process
func (b *BaseInstaller) InstallBaseTools(ctx context.Context) error {
	if b.osType == OSUnknown {
		return fmt.Errorf("cannot install base tools: unknown operating system")
	}

	var config OSConfig
	switch b.osType {
	case OSArch:
		config = b.preferences.Arch
	case OSUbuntu:
		config = b.preferences.Ubuntu
	case OSMacOS:
		config = b.preferences.MacOS
	default:
		return fmt.Errorf("unsupported operating system: %s", b.osType)
	}

	// Execute install commands for the OS
	for _, cmdStr := range config.InstallCommands {
		if cmdStr == "" {
			continue
		}
		if err := b.executeCommand(ctx, cmdStr); err != nil {
			return fmt.Errorf("failed to install base tools: %w", err)
		}
	}

	return nil
}

// EnsureTool ensures a specific tool is installed
// Returns nil if already installed, installs it if not
func (b *BaseInstaller) EnsureTool(ctx context.Context, tool string) error {
	// Check if already installed
	if b.IsToolInstalled(tool) {
		return nil
	}

	// Get tool definition from preferences
	toolDef, exists := b.preferences.Tools[tool]
	if !exists {
		return fmt.Errorf("tool '%s' not found in preferences", tool)
	}

	// Get the install command for the current OS
	var cmdStr string
	switch b.osType {
	case OSArch:
		cmdStr = toolDef.Arch
	case OSUbuntu:
		cmdStr = toolDef.Ubuntu
	case OSMacOS:
		cmdStr = toolDef.MacOS
	default:
		return fmt.Errorf("unsupported operating system for tool installation: %s", b.osType)
	}

	if cmdStr == "" {
		return fmt.Errorf("no install command defined for tool '%s' on %s", tool, b.osType)
	}

	// Special handling for brew installation on macOS
	if tool == "brew" && b.osType == OSMacOS {
		return b.installBrew(ctx, cmdStr)
	}

	// Execute the install command
	if err := b.executeCommand(ctx, cmdStr); err != nil {
		return fmt.Errorf("failed to install %s: %w", tool, err)
	}

	return nil
}

// installBrew handles the special case of installing Homebrew on macOS
func (b *BaseInstaller) installBrew(ctx context.Context, installCmd string) error {
	// Check if already installed
	if b.IsToolInstalled("brew") {
		return nil
	}

	// Execute the brew install script
	cmd := b.execFunc(ctx, "bash", "-c", installCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install Homebrew: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// EnsureTools ensures multiple tools are installed
func (b *BaseInstaller) EnsureTools(ctx context.Context, tools []string) error {
	for _, tool := range tools {
		if err := b.EnsureTool(ctx, tool); err != nil {
			return err
		}
	}
	return nil
}

// DetectRequiredTools analyzes a command string and returns which base tools it requires
// This is used to automatically detect what needs to be installed
func (b *BaseInstaller) DetectRequiredTools(command string) []string {
	var required []string

	// Map of command patterns to required tools
	toolPatterns := map[string][]string{
		"npm ":     {"npm"},
		"npm\t":    {"npm"},
		"brew ":    {"brew"},
		"brew\t":   {"brew"},
		"pip ":     {"pip"},
		"pip3 ":    {"pip"},
		"pipx ":    {"pipx"},
		"cargo ":   {"cargo"},
		"pacman ":  {"pacman"},
		"apt ":     {"apt"},
		"apt-get ": {"apt"},
	}

	commandLower := strings.ToLower(command)

	for pattern, tools := range toolPatterns {
		if strings.Contains(commandLower, pattern) {
			// Check if tool is not already in the list
			for _, tool := range tools {
				if !b.contains(required, tool) {
					required = append(required, tool)
				}
			}
		}
	}

	return required
}

// contains checks if a string slice contains a value
func (b *BaseInstaller) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// executeCommand executes a shell command
func (b *BaseInstaller) executeCommand(ctx context.Context, cmdStr string) error {
	cmd := b.execFunc(ctx, "sh", "-c", cmdStr)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// GetPackageManager returns the package manager for the detected OS
func (b *BaseInstaller) GetPackageManager() string {
	return GetPackageManager(b.osType)
}
