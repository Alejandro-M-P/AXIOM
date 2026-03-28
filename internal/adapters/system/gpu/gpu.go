package gpu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	gfxNameRe    = regexp.MustCompile(`\bgfx[0-9a-zA-Z]+\b`)
	pciVendorRe  = regexp.MustCompile(`\[([0-9a-fA-F]{4}):([0-9a-fA-F]{4})\]`)
	whitespaceRe = regexp.MustCompile(`\s+`)
)

type GPUInfo struct {
	Type       string
	GfxVal     string
	Name       string
	RawGfx     string
	PCIAddress string
	VendorID   string
	DeviceID   string
}

func Detect() GPUInfo {
	info := GPUInfo{
		Type:   "generic",
		Name:   "Desconocida",
		GfxVal: "",
	}

	fillFromLSPCI(&info)
	fillFromSysfs(&info)

	if info.Type == "amd" {
		info.RawGfx, info.GfxVal = detectAMDGfx()
	}

	return info
}

func fillFromLSPCI(info *GPUInfo) {
	out, err := runCommand("lspci", "-Dnn")
	if err != nil || strings.TrimSpace(out) == "" {
		return
	}

	for _, line := range strings.Split(out, "\n") {
		if !isGPULine(line) {
			continue
		}

		info.PCIAddress = firstField(line)
		info.Name = cleanLSPCIName(line)

		if match := pciVendorRe.FindStringSubmatch(line); len(match) == 3 {
			info.VendorID = strings.ToLower(match[1])
			info.DeviceID = strings.ToLower(match[2])
		}

		if info.Type == "generic" {
			info.Type = typeFromVendor(info.VendorID, line)
		}

		return
	}
}

func fillFromSysfs(info *GPUInfo) {
	entries, err := filepath.Glob("/sys/class/drm/card*/device")
	if err != nil {
		return
	}

	for _, base := range entries {
		classValue := readTrimmed(filepath.Join(base, "class"))
		if !isGPUClass(classValue) {
			continue
		}

		if info.VendorID == "" {
			info.VendorID = normalizeHex(readTrimmed(filepath.Join(base, "vendor")))
		}
		if info.DeviceID == "" {
			info.DeviceID = normalizeHex(readTrimmed(filepath.Join(base, "device")))
		}
		if info.PCIAddress == "" {
			info.PCIAddress = filepath.Base(base)
		}
		if info.Name == "" || info.Name == "Desconocida" {
			if name := readTrimmed(filepath.Join(base, "product_name")); name != "" {
				info.Name = name
			}
		}
		if info.Type == "generic" {
			info.Type = typeFromVendor(info.VendorID, info.Name)
		}

		return
	}
}

func detectAMDGfx() (string, string) {
	if raw := detectRawGfxFromCommand("rocm_agent_enumerator", "-name"); raw != "" {
		return raw, normalizeGfxOverride(raw)
	}
	if raw := detectRawGfxFromCommand("rocminfo"); raw != "" {
		return raw, normalizeGfxOverride(raw)
	}
	if raw := detectRawGfxFromKFD(); raw != "" {
		return raw, normalizeGfxOverride(raw)
	}
	return "", ""
}

func detectRawGfxFromCommand(name string, args ...string) string {
	out, err := runCommand(name, args...)
	if err != nil {
		return ""
	}
	return firstGfxName(out)
}

func detectRawGfxFromKFD() string {
	entries, err := filepath.Glob("/sys/class/kfd/kfd/topology/nodes/*/properties")
	if err != nil {
		return ""
	}

	for _, path := range entries {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(data), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 2 || fields[0] != "gfx_target_version" {
				continue
			}

			raw := strings.TrimSpace(fields[1])
			if raw == "" {
				continue
			}

			if strings.HasPrefix(strings.ToLower(raw), "gfx") {
				return strings.ToLower(raw)
			}

			if normalized := gfxNameFromTargetVersion(raw); normalized != "" {
				return normalized
			}
		}
	}

	return ""
}

func firstGfxName(input string) string {
	match := gfxNameRe.FindString(strings.ToLower(input))
	return strings.TrimSpace(match)
}

func normalizeGfxOverride(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return ""
	}

	if strings.HasPrefix(raw, "gfx") {
		raw = strings.TrimPrefix(raw, "gfx")
	}

	if strings.Contains(raw, ".") {
		return raw
	}

	if len(raw) < 3 {
		return raw
	}

	digits := make([]int, 0, len(raw))
	for _, r := range raw {
		if r < '0' || r > '9' {
			return raw
		}
		digits = append(digits, int(r-'0'))
	}

	majorDigits := 2
	if len(digits) == 3 {
		majorDigits = 1
	}
	if len(digits) < majorDigits+2 {
		return raw
	}

	major := digitsToInt(digits[:majorDigits])
	minor := digitsToInt(digits[majorDigits : majorDigits+1])
	patch := digitsToInt(digits[majorDigits+1 : majorDigits+2])

	if len(digits) > majorDigits+2 {
		patch = digitsToInt(digits[majorDigits+1:])
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func gfxNameFromTargetVersion(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) != 6 {
		return ""
	}

	major, err1 := strconv.Atoi(raw[0:2])
	minor, err2 := strconv.Atoi(raw[2:4])
	patch, err3 := strconv.Atoi(raw[4:6])
	if err1 != nil || err2 != nil || err3 != nil {
		return ""
	}

	return fmt.Sprintf("gfx%d%d%d", major, minor, patch)
}

func typeFromVendor(vendorID, fallback string) string {
	switch strings.ToLower(vendorID) {
	case "10de":
		return "nvidia"
	case "1002":
		return "amd"
	case "8086":
		return "intel"
	}

	fallback = strings.ToLower(fallback)
	switch {
	case strings.Contains(fallback, "nvidia"):
		return "nvidia"
	case strings.Contains(fallback, "amd"), strings.Contains(fallback, "radeon"):
		return "amd"
	case strings.Contains(fallback, "intel"):
		return "intel"
	default:
		return "generic"
	}
}

func cleanLSPCIName(line string) string {
	line = strings.TrimSpace(strings.ReplaceAll(line, "\"", ""))
	if line == "" {
		return "Desconocida"
	}

	if idx := strings.Index(line, " "); idx >= 0 {
		line = strings.TrimSpace(line[idx+1:])
	}
	line = pciVendorRe.ReplaceAllString(line, "")
	line = whitespaceRe.ReplaceAllString(strings.TrimSpace(line), " ")

	return strings.TrimSpace(line)
}

func isGPULine(line string) bool {
	line = strings.ToLower(line)
	return strings.Contains(line, " vga ") ||
		strings.Contains(line, " 3d controller ") ||
		strings.Contains(line, " display controller ")
}

func isGPUClass(classValue string) bool {
	classValue = strings.ToLower(strings.TrimSpace(classValue))
	return strings.HasPrefix(classValue, "0x03")
}

func normalizeHex(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	return strings.TrimPrefix(value, "0x")
}

func readTrimmed(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func firstField(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func digitsToInt(digits []int) int {
	value := 0
	for _, digit := range digits {
		value = value*10 + digit
	}
	return value
}

func runCommand(name string, args ...string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}

	out, err := exec.Command(path, args...).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
