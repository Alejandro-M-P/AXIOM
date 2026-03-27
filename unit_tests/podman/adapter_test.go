package podman_test

import (
	"testing"
)

// MockRuntime es un mock para testing que implementa IContainerRuntime
// Nota: Los tests reales requieren resolver primero los conflictos de tipos
type MockRuntime struct {
	ContainerNames []string
	Images         []string
	Err            error
}

func (m *MockRuntime) CreateContainer(name, image, home, flags string) error {
	return m.Err
}

func (m *MockRuntime) StartContainer(name string) error {
	return m.Err
}

func (m *MockRuntime) StopContainer(name string) error {
	return m.Err
}

func (m *MockRuntime) RemoveContainer(name string, force bool) error {
	return m.Err
}

func (m *MockRuntime) ListContainers() (interface{}, error) {
	return nil, m.Err
}

func (m *MockRuntime) ContainerExists(name string) (bool, error) {
	for _, n := range m.ContainerNames {
		if n == name {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockRuntime) ImageExists(image string) bool {
	for _, img := range m.Images {
		if img == image {
			return true
		}
	}
	return false
}

func (m *MockRuntime) CommitContainer(name, image string) error {
	return m.Err
}

func (m *MockRuntime) RunCommand(containerName string, args ...string) error {
	return m.Err
}

func (m *MockRuntime) RunCommandWithInput(containerName, input string, args ...string) error {
	return m.Err
}

func (m *MockRuntime) RunCommandOutput(containerName string, args ...string) (string, error) {
	return "", m.Err
}

// Tests básicos

func TestMockContainerExists(t *testing.T) {
	mock := &MockRuntime{
		ContainerNames: []string{"bunker1", "bunker2"},
	}

	exists, _ := mock.ContainerExists("bunker1")
	if !exists {
		t.Error("Expected bunker1 to exist")
	}

	exists, _ = mock.ContainerExists("bunker3")
	if exists {
		t.Error("Expected bunker3 to not exist")
	}
}

func TestMockImageExists(t *testing.T) {
	mock := &MockRuntime{
		Images: []string{
			"localhost/axiom-rdna4:latest",
			"localhost/axiom-nvidia:latest",
		},
	}

	if !mock.ImageExists("localhost/axiom-rdna4:latest") {
		t.Error("Expected image to exist")
	}

	if mock.ImageExists("localhost/axiom-fake:latest") {
		t.Error("Expected image to not exist")
	}
}
