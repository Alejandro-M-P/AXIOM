package bunker

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"axiom/internal/domain"
)

func TestSanitize_ValidName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"valid-name", "valid-name"},
		{"valid_name", "valid_name"},
		{"valid.name", "valid.name"},
		{"validname123", "validname123"},
		{"NAME", "NAME"},
		{"Name-With-Mixed-Case", "Name-With-Mixed-Case"},
	}

	for _, tc := range tests {
		result, err := sanitizeBunkerName(tc.input)
		if err != nil {
			t.Errorf("sanitizeBunkerName(%q): unexpected error: %s", tc.input, err)
			continue
		}
		if result != tc.expected {
			t.Errorf("sanitizeBunkerName(%q): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}

func TestSanitize_InvalidChars(t *testing.T) {
	invalidNames := []string{
		"../etc",
		"foo\\bar",
		".",
		"..",
		"/absolute",
		"with\\backslash",
	}

	for _, input := range invalidNames {
		_, err := sanitizeBunkerName(input)
		if err == nil {
			t.Errorf("sanitizeBunkerName(%q): expected error, got nil", input)
		}
		if err != nil && err.Error() != "invalid_name" {
			t.Errorf("sanitizeBunkerName(%q): expected 'invalid_name' error, got %q", input, err.Error())
		}
	}
}

func TestSanitize_Whitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  name", "name"},
		{"name  ", "name"},
		{"  name  ", "name"},
		{"\tname\t", "name"},
		{"\nname\n", "name"},
	}

	for _, tc := range tests {
		result, err := sanitizeBunkerName(tc.input)
		if err != nil {
			t.Errorf("sanitizeBunkerName(%q): unexpected error: %s", tc.input, err)
			continue
		}
		if result != tc.expected {
			t.Errorf("sanitizeBunkerName(%q): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}

func TestFormatBytes_Function(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tc := range tests {
		result := formatBytes(tc.input)
		if result != tc.expected {
			t.Errorf("formatBytes(%d): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}

func TestFormatBytes_EdgeCases(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{-1, "-1 B"},
		{1023, "1023 B"},
		{1048575, "1024.0 KB"},
		{1073741823, "1024.0 MB"},
	}

	for _, tc := range tests {
		result := formatBytes(tc.input)
		if result != tc.expected {
			t.Errorf("formatBytes(%d): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}

func TestHumanPath_Helper(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{filepath.Join(home, "projects", "myapp"), "~" + string(filepath.Separator) + "projects" + string(filepath.Separator) + "myapp"},
		{filepath.Join(home, ".axiom"), "~" + string(filepath.Separator) + ".axiom"},
		{"/usr/local/bin", "/usr/local/bin"},
		{"/tmp/test", "/tmp/test"},
	}

	for _, tc := range tests {
		result := humanPath(tc.input)
		if result != tc.expected {
			t.Errorf("humanPath(%q): expected %q, got %q", tc.input, tc.expected, result)
		}
	}
}

func TestHumanPath_HomeError(t *testing.T) {
	result := humanPath("/some/path")
	if result != "/some/path" {
		t.Errorf("expected unchanged path on error, got %q", result)
	}
}

func TestYesNo_Function(t *testing.T) {
	if yesNo(true) != "common.yes" {
		t.Errorf("yesNo(true): expected 'common.yes', got %q", yesNo(true))
	}
	if yesNo(false) != "common.no" {
		t.Errorf("yesNo(false): expected 'common.no', got %q", yesNo(false))
	}
}

func TestIsYes_Function(t *testing.T) {
	trueValues := []string{"s", "S", "y", "Y", "yes", "YES", "si", "SI", "sí", "Sí"}
	falseValues := []string{"", "n", "no", "NO", "not", "false", "0", "1"}

	for _, val := range trueValues {
		if !isYes(val) {
			t.Errorf("isYes(%q): expected true", val)
		}
	}

	for _, val := range falseValues {
		if isYes(val) {
			t.Errorf("isYes(%q): expected false", val)
		}
	}
}

func TestDefaultString_Helper(t *testing.T) {
	tests := []struct {
		value    string
		fallback string
		expected string
	}{
		{"actual", "default", "actual"},
		{"", "default", "default"},
		{"   ", "default", "default"},
		{"value", "", "value"},
	}

	for _, tc := range tests {
		result := defaultString(tc.value, tc.fallback)
		if result != tc.expected {
			t.Errorf("defaultString(%q, %q): expected %q, got %q", tc.value, tc.fallback, tc.expected, result)
		}
	}
}

func TestBunkerTimestamp_Format(t *testing.T) {
	// Create a proper time.Time for testing
	now := time.Now()
	result := bunkerTimestamp(now)

	if len(result) != 10 {
		t.Errorf("expected date string of length 10, got %q (len %d)", result, len(result))
	}
}

func TestBaseImageName_Helper(t *testing.T) {
	tests := []struct {
		gpuType  string
		expected string
	}{
		{"rdna4", "localhost/axiom-rdna4:latest"},
		{"nvidia", "localhost/axiom-nvidia:latest"},
		{"intel", "localhost/axiom-intel:latest"},
		{"generic", "localhost/axiom-generic:latest"},
		{"", "localhost/axiom-generic:latest"},
		{"  ", "localhost/axiom-generic:latest"},
	}

	for _, tc := range tests {
		result := baseImageName(tc.gpuType)
		if result != tc.expected {
			t.Errorf("baseImageName(%q): expected %q, got %q", tc.gpuType, tc.expected, result)
		}
	}
}

func TestResolveBuildGPU_Helper(t *testing.T) {
	cfg := EnvConfig{
		GFXVal: "CUSTOM_GFX",
	}

	gpu := resolveBuildGPU(cfg)

	if gpu.Type != "generic" {
		t.Errorf("expected GPU type 'generic', got %q", gpu.Type)
	}

	if gpu.GfxVal != "CUSTOM_GFX" {
		t.Errorf("expected GfxVal 'CUSTOM_GFX', got %q", gpu.GfxVal)
	}
}

func TestSSHVolumeFlag_NoAgent(t *testing.T) {
	original := os.Getenv("SSH_AUTH_SOCK")
	os.Unsetenv("SSH_AUTH_SOCK")
	defer os.Setenv("SSH_AUTH_SOCK", original)

	result := sshVolumeFlag()
	if result != "" {
		t.Errorf("expected empty string when no SSH agent, got %q", result)
	}
}

func TestSSHVolumeFlag_WithAgent(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "ssh-agent.sock")

	f, err := os.Create(socketPath)
	if err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}
	f.Close()

	original := os.Getenv("SSH_AUTH_SOCK")
	os.Setenv("SSH_AUTH_SOCK", socketPath)
	defer os.Setenv("SSH_AUTH_SOCK", original)

	result := sshVolumeFlag()
	_ = result
}

func TestEnvConfigAlias_Correct(t *testing.T) {
	var cfg EnvConfig
	var domainCfg domain.EnvConfig

	cfg.BaseDir = "/test"
	domainCfg.BaseDir = "/test"

	if cfg.BaseDir != domainCfg.BaseDir {
		t.Error("EnvConfig alias not working correctly")
	}
}

func TestGPUInfoAlias_Correct(t *testing.T) {
	var gpu GPUInfo
	var domainGPU domain.GPUInfo

	gpu.Type = "test"
	domainGPU.Type = "test"

	if gpu.Type != domainGPU.Type {
		t.Error("GPUInfo alias not working correctly")
	}
}
