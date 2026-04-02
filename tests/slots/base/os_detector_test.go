package base_test

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/core/slots/base"
)

func TestOSDetector_Detect(t *testing.T) {
	detector := base.NewOSDetector()

	osType, osName, err := detector.Detect()

	// Should not return an error on supported platforms
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if err != nil {
			t.Errorf("Detect() returned error on %s: %v", runtime.GOOS, err)
		}

		// Check that we got a valid OS type
		if osType == base.OSUnknown {
			t.Errorf("Detect() returned OSUnknown on %s", runtime.GOOS)
		}

		if osName == "" {
			t.Error("Detect() returned empty osName")
		}
	} else {
		// Unsupported OS should return error
		if err == nil {
			t.Error("Detect() should return error on unsupported OS")
		}
	}
}

func TestOSDetector_isCommandAvailable(t *testing.T) {
	detector := base.NewOSDetector()

	// "echo" should be available on all Unix-like systems
	if runtime.GOOS != "windows" {
		if !detector.IsCommandAvailable("echo") {
			t.Error("IsCommandAvailable('echo') returned false, expected true")
		}
	}

	// A non-existent command should not be available
	if detector.IsCommandAvailable("this_command_definitely_does_not_exist_12345") {
		t.Error("IsCommandAvailable returned true for non-existent command")
	}
}

func TestGetPackageManager(t *testing.T) {
	tests := []struct {
		osType   base.OSType
		expected string
	}{
		{base.OSArch, "pacman"},
		{base.OSUbuntu, "apt"},
		{base.OSMacOS, "brew"},
		{base.OSUnknown, ""},
	}

	for _, tt := range tests {
		result := base.GetPackageManager(tt.osType)
		if result != tt.expected {
			t.Errorf("GetPackageManager(%s) = %s, expected %s", tt.osType, result, tt.expected)
		}
	}
}

func TestBaseInstaller_DetectRequiredTools(t *testing.T) {
	// Create a mock installer using the constructor with custom exec function
	prefs := &base.OSPreferences{
		Tools: map[string]base.ToolDef{
			"npm":   {Arch: "sudo pacman -S --noconfirm nodejs npm"},
			"brew":  {MacOS: "brew install"},
			"pip":   {Arch: "sudo pacman -S --noconfirm python-pip"},
			"cargo": {Arch: "sudo pacman -S --noconfirm rust"},
		},
	}

	// Create a minimal detector for testing
	detector := base.NewOSDetector()
	osType, _, _ := detector.Detect()

	installer := base.NewBaseInstallerWithDeps(prefs, detector, exec.CommandContext)
	_ = osType // Use the osType to verify we're working with a real OS

	tests := []struct {
		command  string
		expected []string
	}{
		{"npm install -g something", []string{"npm"}},
		{"brew install package", []string{"brew"}},
		{"pip install package", []string{"pip"}},
		{"cargo build", []string{"cargo"}},
		{"echo hello world", []string{}},
		{"pacman -S package", []string{"pacman"}}, // System package managers are detected but filtered later
	}

	for _, tt := range tests {
		result := installer.DetectRequiredTools(tt.command)

		if len(result) != len(tt.expected) {
			t.Errorf("DetectRequiredTools(%q) = %v, expected %v", tt.command, result, tt.expected)
			continue
		}

		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("DetectRequiredTools(%q)[%d] = %s, expected %s", tt.command, i, result[i], tt.expected[i])
			}
		}
	}
}

func TestIsBaseTool(t *testing.T) {
	baseTools := []string{"npm", "brew", "pip", "pipx", "cargo", "pacman", "apt", "git", "curl"}

	for _, tool := range baseTools {
		if !base.IsBaseTool(tool) {
			t.Errorf("IsBaseTool(%s) returned false, expected true", tool)
		}
	}

	if base.IsBaseTool("nodejs") {
		t.Error("IsBaseTool('nodejs') returned true, expected false")
	}

	if base.IsBaseTool("ollama") {
		t.Error("IsBaseTool('ollama') returned true, expected false")
	}
}

func TestBaseToolsToMap(t *testing.T) {
	m := base.BaseToolsToMap()

	expectedTools := []string{"npm", "brew", "pip", "pipx", "cargo", "pacman", "apt", "git", "curl"}
	for _, tool := range expectedTools {
		if !m[tool] {
			t.Errorf("BaseToolsToMap()[%s] = false, expected true", tool)
		}
	}
}

func TestOSType_String(t *testing.T) {
	tests := []struct {
		osType   base.OSType
		expected string
	}{
		{base.OSArch, "arch"},
		{base.OSUbuntu, "ubuntu"},
		{base.OSMacOS, "macos"},
		{base.OSUnknown, "unknown"},
	}

	for _, tt := range tests {
		result := tt.osType.String()
		if result != tt.expected {
			t.Errorf("OSType.String() = %s, expected %s", result, tt.expected)
		}
	}
}

// Mock exec function for testing
func mockExecCommand(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, arg...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	// This is a helper process for mocking exec.Command
	os.Exit(0)
}
