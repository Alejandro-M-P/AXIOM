package runtime_test

import (
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/adapters/runtime"
)

// ============================================================================
// TestPodmanCreateBunkerCommand
// ============================================================================

func TestPodmanCreateBunkerCommand(t *testing.T) {
	t.Run("returns correct command structure", func(t *testing.T) {
		cmd := runtime.Podman.CreateBunker("test-bunker", "fedora:latest", "/home/test", "--init")
		if len(cmd) == 0 {
			t.Fatal("expected non-empty command slice")
		}
		if cmd[0] != "distrobox-create" {
			t.Errorf("expected first element 'distrobox-create', got '%s'", cmd[0])
		}
	})

	t.Run("includes name flag", func(t *testing.T) {
		cmd := runtime.Podman.CreateBunker("my-bunker", "image", "/home", "")
		found := false
		for i, arg := range cmd {
			if arg == "--name" && i+1 < len(cmd) && cmd[i+1] == "my-bunker" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected --name my-bunker in command")
		}
	})

	t.Run("includes image flag", func(t *testing.T) {
		cmd := runtime.Podman.CreateBunker("bunker", "ubuntu:22.04", "/home", "")
		found := false
		for i, arg := range cmd {
			if arg == "--image" && i+1 < len(cmd) && cmd[i+1] == "ubuntu:22.04" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected --image ubuntu:22.04 in command")
		}
	})

	t.Run("includes home flag", func(t *testing.T) {
		cmd := runtime.Podman.CreateBunker("bunker", "img", "/custom/home", "")
		found := false
		for i, arg := range cmd {
			if arg == "--home" && i+1 < len(cmd) && cmd[i+1] == "/custom/home" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected --home /custom/home in command")
		}
	})

	t.Run("includes additional-flags", func(t *testing.T) {
		cmd := runtime.Podman.CreateBunker("bunker", "img", "/home", "--privileged")
		found := false
		for i, arg := range cmd {
			if arg == "--additional-flags" && i+1 < len(cmd) && cmd[i+1] == "--privileged" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected --additional-flags --privileged in command")
		}
	})

	t.Run("includes --yes flag", func(t *testing.T) {
		cmd := runtime.Podman.CreateBunker("bunker", "img", "/home", "")
		found := false
		for _, arg := range cmd {
			if arg == "--yes" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected --yes in command")
		}
	})

	t.Run("handles empty flags", func(t *testing.T) {
		cmd := runtime.Podman.CreateBunker("bunker", "img", "/home", "")
		if cmd == nil {
			t.Fatal("command should not be nil")
		}
		// Should still have --additional-flags even if empty
		found := false
		for i, arg := range cmd {
			if arg == "--additional-flags" {
				if i+1 < len(cmd) && cmd[i+1] == "" {
					found = true
				}
				break
			}
		}
		if !found {
			t.Error("expected --additional-flags with empty value")
		}
	})
}

// ============================================================================
// TestPodmanStartBunkerCommand
// ============================================================================

func TestPodmanStartBunkerCommand(t *testing.T) {
	t.Run("returns correct command structure", func(t *testing.T) {
		cmd := runtime.Podman.StartBunker("test-bunker")
		if len(cmd) == 0 {
			t.Fatal("expected non-empty command slice")
		}
		if cmd[0] != "podman" {
			t.Errorf("expected first element 'podman', got '%s'", cmd[0])
		}
	})

	t.Run("includes start subcommand", func(t *testing.T) {
		cmd := runtime.Podman.StartBunker("bunker")
		found := false
		for _, arg := range cmd {
			if arg == "start" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'start' subcommand")
		}
	})

	t.Run("includes bunker name", func(t *testing.T) {
		cmd := runtime.Podman.StartBunker("my-container")
		found := false
		for _, arg := range cmd {
			if arg == "my-container" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected bunker name 'my-container'")
		}
	})

	t.Run("command length is 3", func(t *testing.T) {
		cmd := runtime.Podman.StartBunker("bunker")
		if len(cmd) != 3 {
			t.Errorf("expected command length 3, got %d: %v", len(cmd), cmd)
		}
	})
}

// ============================================================================
// TestPodmanStopBunkerCommand
// ============================================================================

func TestPodmanStopBunkerCommand(t *testing.T) {
	t.Run("returns correct command structure", func(t *testing.T) {
		cmd := runtime.Podman.StopBunker("test-bunker")
		if len(cmd) == 0 {
			t.Fatal("expected non-empty command slice")
		}
		if cmd[0] != "distrobox-stop" {
			t.Errorf("expected first element 'distrobox-stop', got '%s'", cmd[0])
		}
	})

	t.Run("includes bunker name", func(t *testing.T) {
		cmd := runtime.Podman.StopBunker("my-bunker")
		found := false
		for _, arg := range cmd {
			if arg == "my-bunker" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected bunker name 'my-bunker'")
		}
	})

	t.Run("includes --yes flag", func(t *testing.T) {
		cmd := runtime.Podman.StopBunker("bunker")
		found := false
		for _, arg := range cmd {
			if arg == "--yes" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected --yes flag")
		}
	})
}

// ============================================================================
// TestPodmanRemoveBunkerCommand
// ============================================================================

func TestPodmanRemoveBunkerCommand(t *testing.T) {
	t.Run("without force returns correct command", func(t *testing.T) {
		cmd := runtime.Podman.RemoveBunker("test-bunker", false)
		if cmd[0] != "distrobox-rm" {
			t.Errorf("expected 'distrobox-rm', got '%s'", cmd[0])
		}
		foundYes := false
		for _, arg := range cmd {
			if arg == "--yes" {
				foundYes = true
				break
			}
		}
		if !foundYes {
			t.Error("expected --yes flag")
		}
		foundForce := false
		for _, arg := range cmd {
			if arg == "--force" {
				foundForce = true
				break
			}
		}
		if foundForce {
			t.Error("should not include --force when force=false")
		}
	})

	t.Run("with force includes --force flag", func(t *testing.T) {
		cmd := runtime.Podman.RemoveBunker("test-bunker", true)
		found := false
		for _, arg := range cmd {
			if arg == "--force" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected --force flag when force=true")
		}
	})

	t.Run("with force still includes --yes", func(t *testing.T) {
		cmd := runtime.Podman.RemoveBunker("bunker", true)
		found := false
		for _, arg := range cmd {
			if arg == "--yes" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected --yes flag even with force")
		}
	})

	t.Run("includes bunker name", func(t *testing.T) {
		cmd := runtime.Podman.RemoveBunker("my-container", true)
		found := false
		for _, arg := range cmd {
			if arg == "my-container" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected bunker name in command")
		}
	})
}

// ============================================================================
// TestPodmanListBunkersCommand
// ============================================================================

func TestPodmanListBunkersCommand(t *testing.T) {
	t.Run("returns correct command structure", func(t *testing.T) {
		cmd := runtime.Podman.ListBunkers()
		if len(cmd) == 0 {
			t.Fatal("expected non-empty command slice")
		}
		if cmd[0] != "podman" {
			t.Errorf("expected 'podman', got '%s'", cmd[0])
		}
	})

	t.Run("includes ps subcommand", func(t *testing.T) {
		cmd := runtime.Podman.ListBunkers()
		found := false
		for _, arg := range cmd {
			if arg == "ps" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'ps' subcommand")
		}
	})

	t.Run("includes -a flag for all containers", func(t *testing.T) {
		cmd := runtime.Podman.ListBunkers()
		found := false
		for _, arg := range cmd {
			if arg == "-a" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected -a flag")
		}
	})

	t.Run("includes json format", func(t *testing.T) {
		cmd := runtime.Podman.ListBunkers()
		found := false
		for i, arg := range cmd {
			if arg == "--format" && i+1 < len(cmd) && cmd[i+1] == "json" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected --format json")
		}
	})
}

// ============================================================================
// TestPodmanBunkerExistsCommand
// ============================================================================

func TestPodmanBunkerExistsCommand(t *testing.T) {
	t.Run("returns correct command structure", func(t *testing.T) {
		cmd := runtime.Podman.BunkerExists("test-bunker")
		if len(cmd) == 0 {
			t.Fatal("expected non-empty command slice")
		}
		if cmd[0] != "podman" {
			t.Errorf("expected 'podman', got '%s'", cmd[0])
		}
	})

	t.Run("uses ps command like ListBunkers", func(t *testing.T) {
		// BunkerExists uses the same command as ListBunkers - it checks by listing
		listCmd := runtime.Podman.ListBunkers()
		existsCmd := runtime.Podman.BunkerExists("bunker")
		if len(listCmd) != len(existsCmd) {
			t.Errorf("expected same length as ListBunkers, got %d vs %d", len(existsCmd), len(listCmd))
		}
	})

	t.Run("command does not use exists subcommand", func(t *testing.T) {
		// Note: unlike ImageExists which uses 'podman image exists',
		// BunkerExists uses 'podman ps -a --format json' for listing
		cmd := runtime.Podman.BunkerExists("bunker")
		for _, arg := range cmd {
			if arg == "exists" {
				t.Error("BunkerExists should not use 'exists' subcommand")
				break
			}
		}
	})
}

// ============================================================================
// TestPodmanImageExistsCommand
// ============================================================================

func TestPodmanImageExistsCommand(t *testing.T) {
	t.Run("returns correct command structure", func(t *testing.T) {
		cmd := runtime.Podman.ImageExists("fedora:latest")
		if len(cmd) == 0 {
			t.Fatal("expected non-empty command slice")
		}
		if cmd[0] != "podman" {
			t.Errorf("expected 'podman', got '%s'", cmd[0])
		}
	})

	t.Run("includes image subcommand", func(t *testing.T) {
		cmd := runtime.Podman.ImageExists("image")
		found := false
		for _, arg := range cmd {
			if arg == "image" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'image' subcommand")
		}
	})

	t.Run("includes exists subcommand", func(t *testing.T) {
		cmd := runtime.Podman.ImageExists("image")
		found := false
		for _, arg := range cmd {
			if arg == "exists" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'exists' subcommand")
		}
	})

	t.Run("includes image name", func(t *testing.T) {
		cmd := runtime.Podman.ImageExists("ubuntu:22.04")
		found := false
		for _, arg := range cmd {
			if arg == "ubuntu:22.04" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected image name 'ubuntu:22.04'")
		}
	})
}

// ============================================================================
// TestPodmanRemoveImageCommand
// ============================================================================

func TestPodmanRemoveImageCommand(t *testing.T) {
	t.Run("without force returns correct command", func(t *testing.T) {
		cmd := runtime.Podman.RemoveImage("fedora:latest", false)
		if cmd[0] != "podman" {
			t.Errorf("expected 'podman', got '%s'", cmd[0])
		}
		if cmd[1] != "rmi" {
			t.Errorf("expected 'rmi', got '%s'", cmd[1])
		}
		foundForce := false
		for _, arg := range cmd {
			if arg == "--force" {
				foundForce = true
				break
			}
		}
		if foundForce {
			t.Error("should not include --force when force=false")
		}
	})

	t.Run("with force includes --force flag", func(t *testing.T) {
		cmd := runtime.Podman.RemoveImage("fedora:latest", true)
		found := false
		for _, arg := range cmd {
			if arg == "--force" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected --force flag when force=true")
		}
	})

	t.Run("includes image name", func(t *testing.T) {
		cmd := runtime.Podman.RemoveImage("my-image:latest", true)
		found := false
		for _, arg := range cmd {
			if arg == "my-image:latest" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected image name 'my-image:latest'")
		}
	})
}

// ============================================================================
// TestPodmanEnterBunkerCommand
// ============================================================================

func TestPodmanEnterBunkerCommand(t *testing.T) {
	t.Run("returns correct command structure", func(t *testing.T) {
		cmd := runtime.Podman.EnterBunker("test-bunker")
		if len(cmd) == 0 {
			t.Fatal("expected non-empty command slice")
		}
		if cmd[0] != "distrobox-enter" {
			t.Errorf("expected 'distrobox-enter', got '%s'", cmd[0])
		}
	})

	t.Run("includes bunker name", func(t *testing.T) {
		cmd := runtime.Podman.EnterBunker("my-bunker")
		found := false
		for _, arg := range cmd {
			if arg == "my-bunker" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected bunker name 'my-bunker'")
		}
	})

	t.Run("uses bash login shell", func(t *testing.T) {
		cmd := runtime.Podman.EnterBunker("bunker")
		found := false
		for _, arg := range cmd {
			if arg == "bash" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'bash' command")
		}
	})

	t.Run("uses --rcfile flag", func(t *testing.T) {
		cmd := runtime.Podman.EnterBunker("bunker")
		found := false
		for _, arg := range cmd {
			if arg == "--rcfile" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected '--rcfile' flag")
		}
	})

	t.Run("uses -i for interactive mode", func(t *testing.T) {
		cmd := runtime.Podman.EnterBunker("bunker")
		found := false
		for _, arg := range cmd {
			if arg == "-i" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected '-i' flag for interactive mode")
		}
	})
}

// ============================================================================
// TestPodmanExecuteInBunkerCommand
// ============================================================================

func TestPodmanExecuteInBunkerCommand(t *testing.T) {
	t.Run("returns correct command structure", func(t *testing.T) {
		cmd := runtime.Podman.ExecuteInBunker("test-bunker", "ls")
		if len(cmd) == 0 {
			t.Fatal("expected non-empty command slice")
		}
		if cmd[0] != "distrobox-enter" {
			t.Errorf("expected 'distrobox-enter', got '%s'", cmd[0])
		}
	})

	t.Run("includes -n flag for name", func(t *testing.T) {
		cmd := runtime.Podman.ExecuteInBunker("my-bunker", "ls")
		found := false
		for i, arg := range cmd {
			if arg == "-n" && i+1 < len(cmd) && cmd[i+1] == "my-bunker" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected -n bunker-name in command")
		}
	})

	t.Run("appends single argument", func(t *testing.T) {
		cmd := runtime.Podman.ExecuteInBunker("bunker", "ls", "-la")
		// Should have: distrobox-enter, -n, name, --, ls, -la
		found := false
		for i := 0; i < len(cmd); i++ {
			if cmd[i] == "--" && i+1 < len(cmd) && cmd[i+1] == "ls" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected command to contain '-- ls'")
		}
	})

	t.Run("appends multiple arguments", func(t *testing.T) {
		cmd := runtime.Podman.ExecuteInBunker("bunker", "git", "status")
		// Check that both git and status are in command after --
		foundGit := false
		foundStatus := false
		afterDash := false
		for _, arg := range cmd {
			if arg == "--" {
				afterDash = true
				continue
			}
			if afterDash {
				if arg == "git" {
					foundGit = true
				}
				if arg == "status" {
					foundStatus = true
				}
			}
		}
		if !foundGit {
			t.Error("expected 'git' after --")
		}
		if !foundStatus {
			t.Error("expected 'status' after --")
		}
	})

	t.Run("handles zero extra arguments", func(t *testing.T) {
		cmd := runtime.Podman.ExecuteInBunker("bunker")
		// Should still have distrobox-enter -n bunker --
		if len(cmd) < 4 {
			t.Errorf("expected at least 4 elements, got %d: %v", len(cmd), cmd)
		}
		// The -- should still be there
		foundDash := false
		for _, arg := range cmd {
			if arg == "--" {
				foundDash = true
				break
			}
		}
		if !foundDash {
			t.Error("expected '--' even with no extra arguments")
		}
	})

	t.Run("handles command with spaces in args", func(t *testing.T) {
		cmd := runtime.Podman.ExecuteInBunker("bunker", "echo", "hello world")
		found := false
		for i := 0; i < len(cmd); i++ {
			if cmd[i] == "echo" && i+1 < len(cmd) && cmd[i+1] == "hello world" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected 'echo hello world' after --")
		}
	})
}

// ============================================================================
// TestCommandSetStructure
// ============================================================================

func TestCommandSetStructure(t *testing.T) {
	t.Run("Podman variable is initialized", func(t *testing.T) {
		// Podman is a CommandSet struct value, not a pointer
		// It should be a zero-value or initialized struct
		cmds := runtime.Podman
		_ = cmds.CreateBunker // Verify the field exists
	})

	t.Run("all function fields are assigned", func(t *testing.T) {
		cmds := runtime.Podman

		if cmds.CreateBunker == nil {
			t.Error("CreateBunker is nil")
		}
		if cmds.StartBunker == nil {
			t.Error("StartBunker is nil")
		}
		if cmds.StopBunker == nil {
			t.Error("StopBunker is nil")
		}
		if cmds.RemoveBunker == nil {
			t.Error("RemoveBunker is nil")
		}
		if cmds.ListBunkers == nil {
			t.Error("ListBunkers is nil")
		}
		if cmds.BunkerExists == nil {
			t.Error("BunkerExists is nil")
		}
		if cmds.ImageExists == nil {
			t.Error("ImageExists is nil")
		}
		if cmds.RemoveImage == nil {
			t.Error("RemoveImage is nil")
		}
		if cmds.EnterBunker == nil {
			t.Error("EnterBunker is nil")
		}
		if cmds.ExecuteInBunker == nil {
			t.Error("ExecuteInBunker is nil")
		}
	})

	t.Run("all functions return non-nil slices", func(t *testing.T) {
		cmds := runtime.Podman

		if cmds.CreateBunker("test", "img", "/home", "") == nil {
			t.Error("CreateBunker should return non-nil slice")
		}
		if cmds.StartBunker("test") == nil {
			t.Error("StartBunker should return non-nil slice")
		}
		if cmds.StopBunker("test") == nil {
			t.Error("StopBunker should return non-nil slice")
		}
		if cmds.RemoveBunker("test", false) == nil {
			t.Error("RemoveBunker should return non-nil slice")
		}
		if cmds.ListBunkers() == nil {
			t.Error("ListBunkers should return non-nil slice")
		}
		if cmds.BunkerExists("test") == nil {
			t.Error("BunkerExists should return non-nil slice")
		}
		if cmds.ImageExists("test") == nil {
			t.Error("ImageExists should return non-nil slice")
		}
		if cmds.RemoveImage("test", false) == nil {
			t.Error("RemoveImage should return non-nil slice")
		}
		if cmds.EnterBunker("test") == nil {
			t.Error("EnterBunker should return non-nil slice")
		}
		if cmds.ExecuteInBunker("test") == nil {
			t.Error("ExecuteInBunker should return non-nil slice")
		}
	})

	t.Run("CommandSet struct has correct fields", func(t *testing.T) {
		// This test verifies the structure matches the interface expectations
		cmds := runtime.CommandSet{}

		// These should all compile - verifying field names and types
		cmds.CreateBunker = func(name, image, home, flags string) []string {
			return []string{"test"}
		}
		cmds.StartBunker = func(name string) []string {
			return []string{"test"}
		}
		cmds.StopBunker = func(name string) []string {
			return []string{"test"}
		}
		cmds.RemoveBunker = func(name string, force bool) []string {
			return []string{"test"}
		}
		cmds.ListBunkers = func() []string {
			return []string{"test"}
		}
		cmds.BunkerExists = func(name string) []string {
			return []string{"test"}
		}
		cmds.ImageExists = func(image string) []string {
			return []string{"test"}
		}
		cmds.RemoveImage = func(image string, force bool) []string {
			return []string{"test"}
		}
		cmds.EnterBunker = func(name string) []string {
			return []string{"test"}
		}
		cmds.ExecuteInBunker = func(name string, args ...string) []string {
			return []string{"test"}
		}

		// All assignments succeeded
	})
}
