package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/runtime"
	"github.com/Alejandro-M-P/AXIOM/internal/adapters/system/gpu"
	"github.com/Alejandro-M-P/AXIOM/internal/domain"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// SystemAdapter implementa ports.ISystem
type SystemAdapter struct{}

func NewSystemAdapter() *SystemAdapter {
	return &SystemAdapter{}
}

func (s *SystemAdapter) DetectGPU() domain.GPUInfo {
	info := gpu.Detect()
	return domain.GPUInfo{
		Type:       info.Type,
		GfxVal:     info.GfxVal,
		Name:       info.Name,
		RawGfx:     info.RawGfx,
		PCIAddress: info.PCIAddress,
		VendorID:   info.VendorID,
		DeviceID:   info.DeviceID,
	}
}

func (s *SystemAdapter) CheckDeps() error {
	for _, dep := range runtime.RequiredDeps {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("errors.system.dependency_missing: %s", dep)
		}
	}
	return nil
}

func (s *SystemAdapter) RefreshSudo(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "sudo", "-v")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *SystemAdapter) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (s *SystemAdapter) SSHKeyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", "id_ed25519"), nil
}

func (s *SystemAdapter) SSHAgentSocket() (string, error) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return "", nil
	}
	info, err := os.Stat(sock)
	if err != nil || info.Mode()&os.ModeSocket == 0 {
		return "", nil
	}
	return sock, nil
}

func (s *SystemAdapter) PrepareSSHAgent(ctx context.Context) error {
	keyPath, err := s.SSHKeyPath()
	if err != nil {
		return err
	}

	ctx1, cancel1 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel1()
	if exec.CommandContext(ctx1, "ssh-add", "-l").Run() == nil {
		return nil
	}

	ctx2, cancel2 := context.WithTimeout(ctx, 10*time.Second)
	defer cancel2()
	return exec.CommandContext(ctx2, "ssh-add", keyPath).Run()
}

// GetCommandPath retorna la ruta absoluta de un comando en el PATH.
// Implementa ports.IDependencyChecker.
func (s *SystemAdapter) GetCommandPath(name string) (string, error) {
	return exec.LookPath(name)
}

var _ ports.ISystem = (*SystemAdapter)(nil)

// PrepareFS crea la estructura de carpetas necesaria para los búnkeres
func PrepareFS(axiomPath, baseDir string) error {
	// Equivalente a mkdir -p "$DIR/lib"
	if err := os.MkdirAll(filepath.Join(axiomPath, "lib"), 0700); err != nil {
		return err
	}

	// Jerarquía de búnkeres: ai_config/models, ai_config/teams, .entorno
	subDirs := []string{"ai_config/models", "ai_config/teams", ".entorno"}
	for _, sd := range subDirs {
		if err := os.MkdirAll(filepath.Join(baseDir, sd), 0700); err != nil {
			return err
		}
	}
	return nil
}

// CreateWrapper genera el acceso directo 'axiom' en ~/.local/bin
func CreateWrapper(axiomPath string) error {
	home, _ := os.UserHomeDir()
	binPath := filepath.Join(home, ".local/bin")
	os.MkdirAll(binPath, 0700)

	target := filepath.Join(binPath, "axiom")
	// Wrapper que exporta la ruta y lanza axiom.sh
	content := fmt.Sprintf("#!/bin/bash\nexport AXIOM_PATH=\"%s\"\nbash \"$AXIOM_PATH/axiom.sh\" \"$@\"\n", axiomPath)

	return os.WriteFile(target, []byte(content), 0755)
}
