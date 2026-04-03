// Package base provides base tool installation functionality for AXIOM slots.
// AXIOM runs exclusively on Arch Linux — no OS detection needed.
package base

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
)

// OSPreferences represents the structure of the os_preferences.toml file.
// Only Arch Linux is supported — other fields are ignored.
type OSPreferences struct {
	Arch  OSConfig           `toml:"arch"`
	Tools map[string]ToolDef `toml:"tools"`
}

// OSConfig represents configuration for Arch Linux.
type OSConfig struct {
	Name            string   `toml:"name"`
	PackageManager  string   `toml:"package_manager"`
	BaseTools       []string `toml:"base_tools"`
	InstallCommands []string `toml:"install_commands"`
}

// ToolDef represents installation command for a tool on Arch Linux.
type ToolDef struct {
	Arch string `toml:"arch"`
}

// BaseInstaller handles automatic installation of base tools and package managers.
type BaseInstaller struct {
	preferences *OSPreferences
	execFunc    func(ctx context.Context, name string, arg ...string) *exec.Cmd
}

// NewBaseInstaller creates a new base installer with the default preferences.
func NewBaseInstaller(preferencesPath string) (*BaseInstaller, error) {
	prefs, err := loadPreferences(preferencesPath)
	if err != nil {
		return nil, fmt.Errorf("errors.slots.base.failed_load_prefs: %w", err)
	}

	return &BaseInstaller{
		preferences: prefs,
		execFunc:    exec.CommandContext,
	}, nil
}

// NewBaseInstallerWithDeps creates a base installer with custom dependencies (for testing).
func NewBaseInstallerWithDeps(prefs *OSPreferences, execFunc func(ctx context.Context, name string, arg ...string) *exec.Cmd) *BaseInstaller {
	return &BaseInstaller{
		preferences: prefs,
		execFunc:    execFunc,
	}
}

// loadPreferences loads the OS preferences from a TOML file.
func loadPreferences(path string) (*OSPreferences, error) {
	var prefs OSPreferences
	if _, err := toml.DecodeFile(path, &prefs); err != nil {
		return nil, err
	}
	return &prefs, nil
}

// GetPackageManager returns the package manager for Arch Linux.
func (b *BaseInstaller) GetPackageManager() string {
	return "pacman"
}

// IsToolInstalled checks if a tool (command) is available in the system.
func (b *BaseInstaller) IsToolInstalled(tool string) bool {
	if b.execFunc != nil {
		cmd := b.execFunc(context.Background(), "which", tool)
		return cmd.Run() == nil
	}
	cmd := exec.Command("which", tool)
	return cmd.Run() == nil
}

// IsCommandAvailable is an alias for IsToolInstalled.
func (b *BaseInstaller) IsCommandAvailable(tool string) bool {
	return b.IsToolInstalled(tool)
}

// InstallBaseTools installs the base tools defined in the Arch Linux preferences.
func (b *BaseInstaller) InstallBaseTools(ctx context.Context) error {
	config := b.preferences.Arch

	for _, cmdStr := range config.InstallCommands {
		if cmdStr == "" {
			continue
		}
		if err := b.executeCommand(ctx, cmdStr); err != nil {
			return fmt.Errorf("errors.slots.base.failed_install_base_tools: %w", err)
		}
	}

	return nil
}

// EnsureTool ensures a specific tool is installed.
// Returns nil if already installed, installs it if not.
func (b *BaseInstaller) EnsureTool(ctx context.Context, tool string) error {
	if b.IsToolInstalled(tool) {
		return nil
	}

	toolDef, exists := b.preferences.Tools[tool]
	if !exists {
		return fmt.Errorf("errors.slots.base.tool_not_found: %s", tool)
	}

	cmdStr := toolDef.Arch
	if cmdStr == "" {
		return fmt.Errorf("errors.slots.base.no_install_cmd")
	}

	if err := b.executeCommand(ctx, cmdStr); err != nil {
		return fmt.Errorf("errors.slots.base.failed_install_tool: %s: %w", tool, err)
	}

	return nil
}

// EnsureTools ensures multiple tools are installed.
func (b *BaseInstaller) EnsureTools(ctx context.Context, tools []string) error {
	for _, tool := range tools {
		if err := b.EnsureTool(ctx, tool); err != nil {
			return err
		}
	}
	return nil
}

// DetectRequiredTools analyzes a command string and returns which base tools it requires.
func (b *BaseInstaller) DetectRequiredTools(command string) []string {
	var required []string

	toolPatterns := map[string][]string{
		"npm ":    {"npm"},
		"npm\t":   {"npm"},
		"pacman ": {"pacman"},
	}

	commandLower := strings.ToLower(command)

	for pattern, tools := range toolPatterns {
		if strings.Contains(commandLower, pattern) {
			for _, tool := range tools {
				if !b.contains(required, tool) {
					required = append(required, tool)
				}
			}
		}
	}

	return required
}

func (b *BaseInstaller) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// executeCommand executes a shell command.
func (b *BaseInstaller) executeCommand(ctx context.Context, cmdStr string) error {
	cmd := b.execFunc(ctx, "sh", "-c", cmdStr)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("errors.slots.base.command_failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}
