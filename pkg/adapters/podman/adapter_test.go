package podman

import (
	"testing"
)

// MockRuntime implementa IContainerRuntime para testing
type MockRuntime struct {
	Containers []ContainerInfo
	Images     []string
	Err        error
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

func (m *MockRuntime) ListContainers() ([]ContainerInfo, error) {
	return m.Containers, m.Err
}

func (m *MockRuntime) ContainerExists(name string) (bool, error) {
	for _, c := range m.Containers {
		if c.Name == name {
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

// Tests de ejemplo

func TestContainerExists(t *testing.T) {
	mock := &MockRuntime{
		Containers: []ContainerInfo{
			{Name: "bunker1", Status: "running"},
			{Name: "bunker2", Status: "stopped"},
		},
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

func TestImageExists(t *testing.T) {
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
