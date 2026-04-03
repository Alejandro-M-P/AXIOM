package runtime

// GetGPUDeviceFlags returns the GPU device flags required for container creation.
// These flags are runtime-specific (Podman, Docker, etc.) and depend on the GPU type.
func GetGPUDeviceFlags(gpuType string) []string {
	if gpuType == "" || gpuType == "generic" {
		return nil
	}

	// All GPU types need the same device access on Linux
	return []string{
		"--device", "/dev/kfd",
		"--device", "/dev/dri",
		"--security-opt", "label=disable",
		"--group-add", "video",
		"--group-add", "render",
	}
}
