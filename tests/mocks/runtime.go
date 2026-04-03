package mocks

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Alejandro-M-P/AXIOM/internal/core/domain"
)

// Compile-time check that MockRuntime implements ports.IBunkerRuntime
var _ interface {
	CreateBunker(ctx context.Context, name, image, home, flags string) error
	GetCreateFlags(ctx context.Context, name, image, home, volumeFlags, gpuType string) (string, error)
	GetVolumeFlags(ctx context.Context, projectDir, name, aiConfigDir, configPath, gpuType, sshSocket string) (string, error)
	StartBunker(ctx context.Context, name string) error
	StopBunker(ctx context.Context, name string) error
	RemoveBunker(ctx context.Context, name string, force bool) error
	ListBunkers(ctx context.Context) ([]domain.Bunker, error)
	BunkerExists(ctx context.Context, name string) (bool, error)
	ImageExists(ctx context.Context, image string) (bool, error)
	IsAxiomImage(image string) bool
	RemoveImage(ctx context.Context, image string, force bool) error
	CommitImage(ctx context.Context, containerName, imageName, author, message string) error
	ContainerState(ctx context.Context, name string) (string, error)
	StartContainer(ctx context.Context, name string) error
	EnterBunker(ctx context.Context, name, rcPath string) error
	ExecuteInBunker(ctx context.Context, name string, args ...string) error
} = (*MockRuntime)(nil)

// MockRuntime implements ports.IBunkerRuntime for testing.
type MockRuntime struct {
	mu sync.Mutex

	// Bunkers in the runtime
	Bunkers []domain.Bunker

	// Images available
	Images []string

	// Container states (name -> state)
	ContainerStates map[string]string

	// Errors to return on operations
	CreateBunkerErr   error
	StartBunkerErr    error
	StopBunkerErr     error
	RemoveBunkerErr   error
	ListBunkersErr    error
	BunkerExistsErr   error
	ImageExistsErr    error
	RemoveImageErr    error
	CommitImageErr    error
	ContainerStateErr error
	StartContainerErr error
	ExecuteErr        error

	// Track calls
	CreateBunkerCalls []CreateBunkerCall
	StartBunkerCalls  []string
	StopBunkerCalls   []string
	RemoveBunkerCalls []RemoveBunkerCall
	CommitImageCalls  []CommitImageCall
	ExecuteCalls      []ExecuteCall

	// Configuration
	ShouldCreateFail bool
	ShouldStartFail  bool
}

type CreateBunkerCall struct {
	Name  string
	Image string
	Home  string
	Flags string
}

type RemoveBunkerCall struct {
	Name  string
	Force bool
}

type CommitImageCall struct {
	ContainerName string
	ImageName     string
	Author        string
	Message       string
}

type ExecuteCall struct {
	Name  string
	Args  []string
	Input string
}

// NewMockRuntime creates a new MockRuntime with default values.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{
		Bunkers:           []domain.Bunker{},
		Images:            []string{"localhost/axiom-generic:latest"},
		ContainerStates:   map[string]string{},
		CreateBunkerCalls: []CreateBunkerCall{},
		StartBunkerCalls:  []string{},
		StopBunkerCalls:   []string{},
		RemoveBunkerCalls: []RemoveBunkerCall{},
		CommitImageCalls:  []CommitImageCall{},
		ExecuteCalls:      []ExecuteCall{},
	}
}

// CreateBunker implements ports.IBunkerRuntime.
func (m *MockRuntime) CreateBunker(ctx context.Context, name, image, home, flags string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CreateBunkerCalls = append(m.CreateBunkerCalls, CreateBunkerCall{
		Name:  name,
		Image: image,
		Home:  home,
		Flags: flags,
	})

	if m.CreateBunkerErr != nil {
		return m.CreateBunkerErr
	}
	if m.ShouldCreateFail {
		return &BunkerError{Op: "create", Name: name, Err: context.DeadlineExceeded}
	}

	// Add to bunkers if not exists
	for _, b := range m.Bunkers {
		if b.Name == name {
			return nil
		}
	}
	m.Bunkers = append(m.Bunkers, domain.Bunker{
		Name:   name,
		Status: "running",
		Image:  image,
	})
	return nil
}

// GetCreateFlags implements ports.IBunkerRuntime.
func (m *MockRuntime) GetCreateFlags(ctx context.Context, name, image, home, volumeFlags, gpuType string) (string, error) {
	gpuFlags := m.GetGPUDeviceFlags(ctx, gpuType)
	gpuFlagsStr := strings.Join(gpuFlags, " ")
	if volumeFlags != "" && gpuFlagsStr != "" {
		return volumeFlags + " " + gpuFlagsStr, nil
	}
	if volumeFlags != "" {
		return volumeFlags, nil
	}
	return gpuFlagsStr, nil
}

// GetVolumeFlags implements ports.IBunkerRuntime.
func (m *MockRuntime) GetVolumeFlags(ctx context.Context, projectDir, name, aiConfigDir, configPath, gpuType, sshSocket string) (string, error) {
	return fmt.Sprintf("--volume %s:/%s:z --volume %s:/ai_config:z --volume %s:/run/axiom/env:ro,z", projectDir, name, aiConfigDir, configPath), nil
}

// GetGPUDeviceFlags implements ports.IBunkerRuntime.
func (m *MockRuntime) GetGPUDeviceFlags(ctx context.Context, gpuType string) []string {
	if gpuType == "" || gpuType == "generic" {
		return nil
	}
	return []string{
		"--device", "/dev/kfd",
		"--device", "/dev/dri",
		"--security-opt", "label=disable",
		"--group-add", "video",
		"--group-add", "render",
	}
}

// StartBunker implements ports.IBunkerRuntime.
func (m *MockRuntime) StartBunker(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StartBunkerCalls = append(m.StartBunkerCalls, name)

	if m.StartBunkerErr != nil {
		return m.StartBunkerErr
	}
	if m.ShouldStartFail {
		return &BunkerError{Op: "start", Name: name, Err: context.DeadlineExceeded}
	}

	for i, b := range m.Bunkers {
		if b.Name == name {
			m.Bunkers[i].Status = "running"
			break
		}
	}
	return nil
}

// StopBunker implements ports.IBunkerRuntime.
func (m *MockRuntime) StopBunker(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.StopBunkerCalls = append(m.StopBunkerCalls, name)

	if m.StopBunkerErr != nil {
		return m.StopBunkerErr
	}

	for i, b := range m.Bunkers {
		if b.Name == name {
			m.Bunkers[i].Status = "stopped"
			break
		}
	}
	return nil
}

// RemoveBunker implements ports.IBunkerRuntime.
func (m *MockRuntime) RemoveBunker(ctx context.Context, name string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RemoveBunkerCalls = append(m.RemoveBunkerCalls, RemoveBunkerCall{
		Name:  name,
		Force: force,
	})

	if m.RemoveBunkerErr != nil {
		return m.RemoveBunkerErr
	}

	for i, b := range m.Bunkers {
		if b.Name == name {
			m.Bunkers = append(m.Bunkers[:i], m.Bunkers[i+1:]...)
			break
		}
	}
	return nil
}

// ListBunkers implements ports.IBunkerRuntime.
func (m *MockRuntime) ListBunkers(ctx context.Context) ([]domain.Bunker, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ListBunkersErr != nil {
		return nil, m.ListBunkersErr
	}
	return m.Bunkers, nil
}

// BunkerExists implements ports.IBunkerRuntime.
func (m *MockRuntime) BunkerExists(ctx context.Context, name string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.BunkerExistsErr != nil {
		return false, m.BunkerExistsErr
	}

	for _, b := range m.Bunkers {
		if b.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// ImageExists implements ports.IBunkerRuntime.
func (m *MockRuntime) ImageExists(ctx context.Context, image string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ImageExistsErr != nil {
		return false, m.ImageExistsErr
	}

	for _, img := range m.Images {
		if img == image {
			return true, nil
		}
	}
	return false, nil
}

// IsAxiomImage implements ports.IBunkerRuntime.
func (m *MockRuntime) IsAxiomImage(image string) bool {
	return strings.HasPrefix(image, "localhost/axiom-")
}

// RemoveImage implements ports.IBunkerRuntime.
func (m *MockRuntime) RemoveImage(ctx context.Context, image string, force bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.RemoveImageErr != nil {
		return m.RemoveImageErr
	}

	for i, img := range m.Images {
		if img == image {
			m.Images = append(m.Images[:i], m.Images[i+1:]...)
			break
		}
	}
	return nil
}

// CommitImage implements ports.IBunkerRuntime.
func (m *MockRuntime) CommitImage(ctx context.Context, containerName, imageName, author, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CommitImageCalls = append(m.CommitImageCalls, CommitImageCall{
		ContainerName: containerName,
		ImageName:     imageName,
		Author:        author,
		Message:       message,
	})

	if m.CommitImageErr != nil {
		return m.CommitImageErr
	}
	return nil
}

// ContainerState implements ports.IBunkerRuntime.
func (m *MockRuntime) ContainerState(ctx context.Context, name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ContainerStateErr != nil {
		return "", m.ContainerStateErr
	}

	if state, ok := m.ContainerStates[name]; ok {
		return state, nil
	}
	return "running", nil
}

// StartContainer implements ports.IBunkerRuntime.
func (m *MockRuntime) StartContainer(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.StartContainerErr != nil {
		return m.StartContainerErr
	}

	m.ContainerStates[name] = "running"
	return nil
}

// EnterBunker implements ports.IBunkerRuntime.
func (m *MockRuntime) EnterBunker(ctx context.Context, name, rcPath string) error {
	return nil
}

// ExecuteInBunker implements ports.IBunkerRuntime.
func (m *MockRuntime) ExecuteInBunker(ctx context.Context, name string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ExecuteCalls = append(m.ExecuteCalls, ExecuteCall{
		Name: name,
		Args: args,
	})

	if m.ExecuteErr != nil {
		return m.ExecuteErr
	}
	return nil
}

// ExecuteWithInput implements ports.IBunkerRuntime.
func (m *MockRuntime) ExecuteWithInput(ctx context.Context, name, input string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ExecuteCalls = append(m.ExecuteCalls, ExecuteCall{
		Name:  name,
		Args:  args,
		Input: input,
	})

	if m.ExecuteErr != nil {
		return m.ExecuteErr
	}
	return nil
}

// BunkerError represents a bunker operation error.
type BunkerError struct {
	Op   string
	Name string
	Err  error
}

func (e *BunkerError) Error() string {
	return e.Op + " error on " + e.Name + ": " + e.Err.Error()
}
