package gpu

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type GPUInfo struct {
	Type   string
	GfxVal string
	Name   string
}

func Detect() GPUInfo {
	info := GPUInfo{Type: "generic", GfxVal: "1100", Name: "Desconocida"}

	// 1. Sacamos el nombre real con lspci (Tu comando favorito)
	out, _ := exec.Command("lspci", "-vnnn").Output()
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "VGA") || strings.Contains(line, "3D controller") {
			info.Name = strings.TrimSpace(strings.ReplaceAll(line, "\"", ""))
			l := strings.ToLower(info.Name)
			if strings.Contains(l, "nvidia") { info.Type = "nvidia"; info.GfxVal = "" }
			if strings.Contains(l, "amd") || strings.Contains(l, "radeon") { info.Type = "amd" }
			break
		}
	}

	// 2. 🎯 CONSULTA DIRECTA AL HARDWARE (Kernel KFD)
	if info.Type == "amd" {
    for i := 0; i < 10; i++ {
        path := fmt.Sprintf("/sys/class/kfd/kfd/topology/nodes/%d/properties", i)
        data, err := os.ReadFile(path)
        if err == nil {
            content := string(data)
            if strings.Contains(content, "gfx_target_version") {
                lines := strings.Split(content, "\n")
                for _, l := range lines {
                    if strings.Contains(l, "gfx_target_version") {
                        fields := strings.Fields(l)
                        if len(fields) >= 2 {
                            v := fields[1] // Recibimos "120001"
                            
                            // 🧬 LÓGICA DE VERSIÓN:
                            // Los 2 primeros: Mayor (12)
                            // Los 2 siguientes: Menor (00)
                            // Los 2 últimos: Parche (01)
                            if len(v) >= 6 {
                                major := v[0:2]
                                minor := strings.TrimPrefix(v[2:4], "0") 
                                if minor == "" { minor = "0" }
                                patch := strings.TrimPrefix(v[4:6], "0")
                                if patch == "" { patch = "0" }
                                
                                info.GfxVal = fmt.Sprintf("%s.%s.%s", major, minor, patch)
                            } else {
                                info.GfxVal = v
                            }
                            break
                        }
                    }
                }
            }
        }
    }
	}

	return info

}