package runtime

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Installer handles system package installation on Arch Linux.
// It lives in the adapter layer because it uses exec.CommandContext.
type Installer struct{}

// NewInstaller creates a new Installer.
func NewInstaller() *Installer {
	return &Installer{}
}

// InstallPackages installs a list of packages using pacman.
// Uses --needed to skip already-installed packages and --noconfirm for non-interactive mode.
func (i *Installer) InstallPackages(ctx context.Context, packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	// Sync repos first
	if err := i.exec(ctx, "sudo", "pacman", "-Sy", "--noconfirm"); err != nil {
		return fmt.Errorf("installer.sync_repos: %w", err)
	}

	// Install all packages in one invocation
	args := []string{"sudo", "pacman", "-S", "--needed", "--noconfirm"}
	args = append(args, packages...)
	if err := i.exec(ctx, args[0], args[1:]...); err != nil {
		return fmt.Errorf("installer.install_packages: %w", err)
	}

	return nil
}

// InstallHomebrew runs the official Homebrew install script.
func (i *Installer) InstallHomebrew(ctx context.Context) error {
	cmd := "NONINTERACTIVE=1 /bin/bash -c \"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
	if err := i.exec(ctx, "bash", "-c", cmd); err != nil {
		return fmt.Errorf("installer.install_homebrew: %w", err)
	}
	return nil
}

// exec runs a command with context support.
func (i *Installer) exec(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, string(output))
	}
	return nil
}
