package ports

import "context"

// IGit defines the contract for Git operations.
// The core uses this port to query Git state without knowing implementation details.
type IGit interface {
	// GetBranch returns the current git branch for a project path.
	// Returns "-" if not a git repo or on error.
	GetBranch(projectPath string) string
	// ConfigureUser sets git user.name and user.email for a specific repo.
	ConfigureUser(ctx context.Context, projectPath, name, email string) error
}
