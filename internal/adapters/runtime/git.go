package runtime

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// GitAdapter implements ports.IGit by reading .git/HEAD directly
// through the IFileSystem abstraction.
type GitAdapter struct {
	fs ports.IFileSystem
}

// NewGitAdapter creates a new GitAdapter with the given filesystem.
func NewGitAdapter(fs ports.IFileSystem) *GitAdapter {
	return &GitAdapter{fs: fs}
}

var _ ports.IGit = (*GitAdapter)(nil)

// GetBranch returns the current git branch for a project path.
// Returns "-" if not a git repo or on error.
func (g *GitAdapter) GetBranch(projectPath string) string {
	headPath := projectPath + "/.git/HEAD"
	content, err := g.fs.ReadFile(headPath)
	if err != nil {
		return "-"
	}

	head := strings.TrimSpace(string(content))

	// Branch reference: "ref: refs/heads/main"
	if strings.HasPrefix(head, "ref: refs/heads/") {
		return strings.TrimPrefix(head, "ref: refs/heads/")
	}

	// Detached HEAD: short commit hash (e.g., "abc1234...")
	if len(head) >= 7 {
		return head[:7]
	}

	return "-"
}

// ConfigureUser sets git user.name and user.email for a specific repo.
func (g *GitAdapter) ConfigureUser(ctx context.Context, projectPath, name, email string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", projectPath, "config", "user.name", name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git config user.name: %w", err)
	}

	cmd = exec.CommandContext(ctx, "git", "-C", projectPath, "config", "user.email", email)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git config user.email: %w", err)
	}

	return nil
}
