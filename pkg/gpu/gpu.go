package gpu

import (
	
	"os/exec"
	"strings"
)

type GPUInfo struct {
	Type   string
	GfxVal string
	Name   string
}

func Detect() GPUInfo {
	info := GPUInfo{Type: "generic", GfxVal: "", Name: "Desconocida"}

	// 1. Usamos tu idea de lspci -mm -nn
	// Formato: "Slot" "Class" "Vendor" "Device" "Rev"
	out, err := exec.Command("lspci", "-mm", "-nn").Output()
	if err != nil {
		return info
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		l := strings.ToLower(line)
		if strings.Contains(l, "vga") || strings.Contains(l, "display") || strings.Contains(l, "3d") {
			// Extraemos el Vendor ID que está entre corchetes dentro de las comillas
			if strings.Contains(l, "[10de]") {
				info.Type = "nvidia"
			} else if strings.Contains(l, "[1002]") {
				info.Type = "amd"
				info.GfxVal = "11.0.0" // Valor seguro por defecto
			} else if strings.Contains(l, "[8086]") {
				info.Type = "intel"
			}
			info.Name = strings.Trim(line, "\"") // Limpiamos comillas
			break 
		}
	}

	// 2. Opcional: Refinar con glxinfo (Si hay sesión gráfica)
	render, err := exec.Command("sh", "-c", "glxinfo | grep 'Device:'").Output()
	if err == nil && len(render) > 0 {
		info.Name = strings.TrimPrefix(strings.TrimSpace(string(render)), "Device: ")
	}

	return info
}