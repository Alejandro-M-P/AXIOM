package mocks

import "context"

// MockGit implements ports.IGit for testing.
type MockGit struct {
	Branch           string
	ConfigureUserErr error
	LastProjectPath  string
	LastUserName     string
	LastUserEmail    string
}

// NewMockGit creates a new MockGit that returns "-" by default.
func NewMockGit() *MockGit {
	return &MockGit{Branch: "-"}
}

// GetBranch returns the configured branch string.
func (m *MockGit) GetBranch(projectPath string) string {
	return m.Branch
}

// ConfigureUser records the call parameters and returns the configured error.
func (m *MockGit) ConfigureUser(ctx context.Context, projectPath, name, email string) error {
	m.LastProjectPath = projectPath
	m.LastUserName = name
	m.LastUserEmail = email
	return m.ConfigureUserErr
}
