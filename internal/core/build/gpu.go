package build

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/config"
	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// ResolveBuildGPU determines the GPU configuration for a build.
// If GPUType is explicitly set in config, it uses that; otherwise detects from hardware.
func ResolveBuildGPU(ctx context.Context, cfg config.EnvConfig, system ports.ISystem, presenter ports.IPresenter) (ports.GPUInfo, error) {
	result := ports.GPUInfo{}

	if cfg.GPUType != "" {
		result.Type = NormalizeGPUType(cfg.GPUType, cfg.GFXVal)
		result.Name = "gpu.forced_by_env"
		return result, nil
	}

	hw, err := system.DetectGPU()
	if err != nil {
		return result, fmt.Errorf("errors.build.gpu.detection: %w", err)
	}
	result.Type = NormalizeGPUType(hw.Type, "")
	result.Name = hw.Name

	if result.Name == "" {
		result.Name = "gpu.unknown"
	}

	return result, nil
}

// NormalizeGPUType converts various GPU type representations to a canonical form.
func NormalizeGPUType(gpuType, gfxVal string) string {
	gpuType = strings.ToLower(strings.TrimSpace(gpuType))
	gfxVal = strings.TrimSpace(gfxVal)

	switch gpuType {
	case "rdna4", "rdna3", "nvidia", "intel", "generic":
		return gpuType
	case "amd":
		major := gfxMajor(gfxVal)
		if major >= 12 {
			return "rdna4"
		}
		if major == 10 || major == 11 {
			return "rdna3"
		}
		return "amd"
	default:
		if major := gfxMajor(gfxVal); major >= 12 {
			return "rdna4"
		} else if major == 10 || major == 11 {
			return "rdna3"
		}
		if gpuType == "" {
			return "generic"
		}
		return gpuType
	}
}

// gfxMajor extracts the major version number from a gfx value.
func gfxMajor(gfxVal string) int {
	gfxVal = strings.TrimSpace(gfxVal)
	if gfxVal == "" {
		return 0
	}

	if strings.HasPrefix(strings.ToLower(gfxVal), "gfx") {
		gfxVal = strings.TrimPrefix(strings.ToLower(gfxVal), "gfx")
		if len(gfxVal) >= 2 {
			major, _ := strconv.Atoi(gfxVal[:2])
			return major
		}
	}

	parts := strings.Split(gfxVal, ".")
	major, _ := strconv.Atoi(parts[0])
	return major
}

// BaseImageName returns the image name for a given GPU type.
func BaseImageName(presenter ports.IPresenter, gpuType string) string {
	gpuType = strings.TrimSpace(gpuType)
	if gpuType == "" {
		gpuType = "generic"
	}
	return presenter.GetText("build.image_name", gpuType)
}

// OllamaArch returns the architecture suffix for Ollama downloads.
func OllamaArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("errors.build.gpu.unsupported_arch")
	}
}
