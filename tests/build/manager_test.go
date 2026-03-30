package build_test

import (
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/build"
	"github.com/Alejandro-M-P/AXIOM/tests/mocks"
)

func TestNewManager(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()
	mockSystem := mocks.NewMockSystem()
	buildContainer := "test-container"

	mgr := build.NewManager(runtime, fs, ui, mockSystem, buildContainer, nil)

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestManagerFields(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()
	mockSystem := mocks.NewMockSystem()
	buildContainer := "test-container"

	mgr := build.NewManager(runtime, fs, ui, mockSystem, buildContainer, nil)

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
}
