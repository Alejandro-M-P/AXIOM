// Package base provides OS detection and base tool installation functionality.
package base

import (
	"context"
	"fmt"
	"os/exec"
)

// Executor is a function type for executing commands, compatible with slots.Executor
type Executor func(ctx context.Context, msg string, name string, args ...string) error

// InstallationStep represents a step in the installation process
type InstallationStep struct {
	Description string
	Command     string
}

// SlotCommandAnalyzer analyzes slot commands to detect required base tools
type SlotCommandAnalyzer struct {
	installer *BaseInstaller
}

// NewSlotCommandAnalyzer creates a new analyzer with the given installer
func NewSlotCommandAnalyzer(installer *BaseInstaller) *SlotCommandAnalyzer {
	return &SlotCommandAnalyzer{
		installer: installer,
	}
}

// AnalyzeAndInstallRequirements analyzes a slot command and installs required base tools
// This should be called before executing the slot command
func (a *SlotCommandAnalyzer) AnalyzeAndInstallRequirements(ctx context.Context, command string) error {
	// Detect which tools are required
	requiredTools := a.installer.DetectRequiredTools(command)

	if len(requiredTools) == 0 {
		// No special tools required
		return nil
	}

	// Install each required tool
	for _, tool := range requiredTools {
		// Skip if tool is a system package manager (pacman, apt) - these should be pre-installed
		if tool == "pacman" || tool == "apt" {
			continue
		}

		if err := a.installer.EnsureTool(ctx, tool); err != nil {
			return fmt.Errorf("failed to ensure tool '%s': %w", tool, err)
		}
	}

	return nil
}

// PrepareEnvironment prepares the environment by installing base tools and package managers
// This should be called once at the start of the slot installation process
func (a *SlotCommandAnalyzer) PrepareEnvironment(ctx context.Context) error {
	return a.installer.InstallBaseTools(ctx)
}

// GetInstaller returns the underlying BaseInstaller
func (a *SlotCommandAnalyzer) GetInstaller() *BaseInstaller {
	return a.installer
}

// IsBaseTool checks if a tool is considered a base tool (not shown in wizard)
// Base tools are those managed by this package
func IsBaseTool(tool string) bool {
	baseTools := []string{"npm", "brew", "pip", "pipx", "cargo", "pacman", "apt", "git", "curl"}
	for _, bt := range baseTools {
		if bt == tool {
			return true
		}
	}
	return false
}

// BaseToolsToMap returns a map of base tools for quick lookup
func BaseToolsToMap() map[string]bool {
	baseTools := []string{"npm", "brew", "pip", "pipx", "cargo", "pacman", "apt", "git", "curl", "base-devel", "build-essential"}
	m := make(map[string]bool)
	for _, bt := range baseTools {
		m[bt] = true
	}
	return m
}

// DetectOSInfo detects and returns OS information
func DetectOSInfo() (OSType, string, error) {
	detector := NewOSDetector()
	return detector.Detect()
}

// DefaultPreferencesPath returns the default path to the OS preferences file
func DefaultPreferencesPath() string {
	return "configs/os_preferences.toml"
}

// ExecuteWithBaseTools wraps command execution with automatic base tool installation
func ExecuteWithBaseTools(ctx context.Context, command string, execFunc func(ctx context.Context, name string, arg ...string) *exec.Cmd) error {
	// Create installer with default preferences
	installer, err := NewBaseInstaller(DefaultPreferencesPath())
	if err != nil {
		// If we can't load preferences, just execute the command
		cmd := execFunc(ctx, "sh", "-c", command)
		return cmd.Run()
	}

	analyzer := NewSlotCommandAnalyzer(installer)

	// Ensure required tools are installed
	if err := analyzer.AnalyzeAndInstallRequirements(ctx, command); err != nil {
		return fmt.Errorf("failed to install required base tools: %w", err)
	}

	// Execute the command
	cmd := execFunc(ctx, "sh", "-c", command)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}
