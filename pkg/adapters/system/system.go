package install

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"axiom/pkg/adapters/system/gpu"
	"axiom/pkg/core/domain"
	"axiom/pkg/core/ports"
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
	deps := []string{"distrobox", "podman", "jq"}
	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("errors.system.dependency_missing: %s", dep)
		}
	}
	return nil
}

var _ ports.ISystem = (*SystemAdapter)(nil)

// CheckDeps verifica las dependencias críticas del sistema
func CheckDeps() error {
	deps := []string{"distrobox", "podman", "jq"}
	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			return fmt.Errorf("errors.system.dependency_missing: %s", dep)
		}
	}
	return nil
}

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

func IsInstalled(axiomPath string) bool {
	envPath := filepath.Join(axiomPath, ".env")
	_, err := os.Stat(envPath)
	return !os.IsNotExist(err)
}
