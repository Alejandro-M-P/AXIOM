package podman_test

import (
	"testing"
)

// ============= MOCK =============

// MockRuntime simula un Podman real para testing
type MockRuntime struct {
	Containers     []ContainerInfo
	Images         []string
	CommandsCalled []string // Para verificar qué comandos se ejecutaron
	Err            error
}

type ContainerInfo struct {
	Name   string
	Status string
	Image  string
}

// Métodos del mock
func (m *MockRuntime) CreateContainer(name, image, home, flags string) error {
	m.CommandsCalled = append(m.CommandsCalled, "CreateContainer:"+name)
	return m.Err
}

func (m *MockRuntime) StartContainer(name string) error {
	m.CommandsCalled = append(m.CommandsCalled, "StartContainer:"+name)
	return m.Err
}

func (m *MockRuntime) StopContainer(name string) error {
	m.CommandsCalled = append(m.CommandsCalled, "StopContainer:"+name)
	return m.Err
}

func (m *MockRuntime) RemoveContainer(name string, force bool) error {
	m.CommandsCalled = append(m.CommandsCalled, "RemoveContainer:"+name)
	return m.Err
}

func (m *MockRuntime) ListContainers() ([]ContainerInfo, error) {
	m.CommandsCalled = append(m.CommandsCalled, "ListContainers")
	return m.Containers, m.Err
}

func (m *MockRuntime) ContainerExists(name string) (bool, error) {
	m.CommandsCalled = append(m.CommandsCalled, "ContainerExists:"+name)
	for _, c := range m.Containers {
		if c.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockRuntime) ImageExists(image string) bool {
	m.CommandsCalled = append(m.CommandsCalled, "ImageExists:"+image)
	for _, img := range m.Images {
		if img == image {
			return true
		}
	}
	return false
}

func (m *MockRuntime) CommitContainer(name, image string) error {
	m.CommandsCalled = append(m.CommandsCalled, "CommitContainer:"+name)
	return m.Err
}

func (m *MockRuntime) RunCommand(containerName string, args ...string) error {
	m.CommandsCalled = append(m.CommandsCalled, "RunCommand:"+containerName)
	return m.Err
}

func (m *MockRuntime) RunCommandWithInput(containerName, input string, args ...string) error {
	m.CommandsCalled = append(m.CommandsCalled, "RunCommandWithInput:"+containerName)
	return m.Err
}

func (m *MockRuntime) RunCommandOutput(containerName string, args ...string) (string, error) {
	m.CommandsCalled = append(m.CommandsCalled, "RunCommandOutput:"+containerName)
	return "", m.Err
}

// ============= TESTS =============

// Test: Crear un contenedor
func TestCreateContainer(t *testing.T) {
	mock := &MockRuntime{
		Containers: []ContainerInfo{},
	}

	// Simula crear un contenedor
	err := mock.CreateContainer("mi-bunker", "ubuntu:22.04", "/home/user", "")

	// Verifica que no hubo error
	if err != nil {
		t.Error("No debería dar error al crear contenedor")
	}

	// Verifica que se ejecutó el comando correcto
	if len(mock.CommandsCalled) == 0 {
		t.Error("Debería haber llamado a CreateContainer")
	}
}

// Test: Verificar si existe un contenedor
func TestContainerExists_Found(t *testing.T) {
	mock := &MockRuntime{
		Containers: []ContainerInfo{
			{Name: "bunker1", Status: "running"},
			{Name: "bunker2", Status: "stopped"},
		},
	}

	// Pregunta si existe bunker1
	exists, _ := mock.ContainerExists("bunker1")

	if !exists {
		t.Error("bunker1 debería existir")
	}
}

func TestContainerExists_NotFound(t *testing.T) {
	mock := &MockRuntime{
		Containers: []ContainerInfo{
			{Name: "bunker1", Status: "running"},
		},
	}

	// Pregunta si existe un contenedor que NO existe
	exists, _ := mock.ContainerExists("bunker-inexistente")

	if exists {
		t.Error("bunker-inexistente NO debería existir")
	}
}

// Test: Verificar si existe una imagen
func TestImageExists(t *testing.T) {
	mock := &MockRuntime{
		Images: []string{
			"localhost/axiom-rdna4:latest",
			"localhost/axiom-nvidia:latest",
		},
	}

	// La imagen existe
	if !mock.ImageExists("localhost/axiom-rdna4:latest") {
		t.Error("La imagen debería existir")
	}

	// La imagen NO existe
	if mock.ImageExists("localhost/imagen-falsa:latest") {
		t.Error("La imagen NO debería existir")
	}
}

// Test: Listar contenedores
func TestListContainers(t *testing.T) {
	mock := &MockRuntime{
		Containers: []ContainerInfo{
			{Name: "bunker1", Status: "running", Image: "ubuntu:22.04"},
			{Name: "bunker2", Status: "stopped", Image: "debian:11"},
		},
	}

	containers, _ := mock.ListContainers()

	if len(containers) != 2 {
		t.Error("Deberían haber 2 contenedores")
	}

	if containers[0].Name != "bunker1" {
		t.Error("El primer contenedor debería llamarse bunker1")
	}
}

// Test: Iniciar un contenedor
func TestStartContainer(t *testing.T) {
	mock := &MockRuntime{
		Containers: []ContainerInfo{
			{Name: "bunker1", Status: "stopped"},
		},
	}

	err := mock.StartContainer("bunker1")

	if err != nil {
		t.Error("No debería dar error al iniciar")
	}

	// Verifica que se ejecutó el comando
	if len(mock.CommandsCalled) == 0 {
		t.Error("Debería haber llamado a StartContainer")
	}
}

// Test: Detener un contenedor
func TestStopContainer(t *testing.T) {
	mock := &MockRuntime{}

	err := mock.StopContainer("bunker1")

	if err != nil {
		t.Error("No debería dar error al detener")
	}
}

// Test: Eliminar un contenedor
func TestRemoveContainer(t *testing.T) {
	mock := &MockRuntime{}

	err := mock.RemoveContainer("bunker1", true)

	if err != nil {
		t.Error("No debería dar error al eliminar")
	}
}

// Test: Simular error
func TestSimulateError(t *testing.T) {
	mock := &MockRuntime{
		Err: fmt.Errorf("error simulado"),
	}

	err := mock.CreateContainer("test", "image", "/home", "")

	if err == nil {
		t.Error("Debería devolver error")
	}
}

import "fmt"
