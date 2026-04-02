package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
)

// Compile-time checks
var _ ports.ISystem = (*MockSystem)(nil)
var _ ports.IDependencyChecker = (*MockDependencyChecker)(nil)

// MockSystem implements ports.ISystem for testing.
type MockSystem struct {
	mu sync.Mutex

	// Track calls
	DetectGPUCalls       int
	CheckDepsCalls       int
	RefreshSudoCalls     int
	PrepareSSHAgentCalls int

	// Return values
	GPUInfo            ports.GPUInfo
	CheckDepsErr       error
	RefreshSudoErr     error
	PrepareSSHAgentErr error
}

// NewMockSystem creates a new MockSystem with default values.
func NewMockSystem() *MockSystem {
	return &MockSystem{
		GPUInfo: ports.GPUInfo{
			Type: "rdna4",
			Name: "AMD Radeon RX 7700 XT",
		},
	}
}

func (m *MockSystem) DetectGPU() (ports.GPUInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DetectGPUCalls++
	return m.GPUInfo, nil
}

func (m *MockSystem) CheckDeps() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CheckDepsCalls++
	return m.CheckDepsErr
}

func (m *MockSystem) RefreshSudo(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RefreshSudoCalls++
	return m.RefreshSudoErr
}

func (m *MockSystem) UserHomeDir() (string, error) {
	return "/home/test", nil
}

func (m *MockSystem) SSHKeyPath() (string, error) {
	return "/home/test/.ssh/id_ed25519", nil
}

func (m *MockSystem) SSHAgentSocket() (string, error) {
	return "", nil
}

func (m *MockSystem) PrepareSSHAgent(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PrepareSSHAgentCalls++
	return m.PrepareSSHAgentErr
}

// GetCommandPath retorna la ruta de un comando (mock).
func (m *MockSystem) GetCommandPath(name string) (string, error) {
	return "", fmt.Errorf("not implemented in mock")
}

// MockDependencyChecker implements ports.IDependencyChecker for testing.
type MockDependencyChecker struct {
	mu sync.Mutex

	// Storage
	Commands map[string]string // command name -> path

	// Track calls
	HasCommandCalls     []string
	GetCommandPathCalls []string

	// Configuration
	HasCommandResult bool
	GetCommandErr    error
}

// NewMockDependencyChecker creates a new MockDependencyChecker with default values.
func NewMockDependencyChecker() *MockDependencyChecker {
	return &MockDependencyChecker{
		Commands:            make(map[string]string),
		HasCommandCalls:     []string{},
		GetCommandPathCalls: []string{},
	}
}

func (m *MockDependencyChecker) HasCommand(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.HasCommandCalls = append(m.HasCommandCalls, name)
	return m.HasCommandResult
}

func (m *MockDependencyChecker) GetCommandPath(name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.GetCommandPathCalls = append(m.GetCommandPathCalls, name)
	if path, ok := m.Commands[name]; ok {
		return path, nil
	}
	return "", m.GetCommandErr
}
