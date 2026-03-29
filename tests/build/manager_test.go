package build_test

import (
	"testing"

	"axiom/internal/build"
	"axiom/tests/mocks"
)

func TestNewManager(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()
	buildContainer := "test-container"

	mgr := build.NewManager(runtime, fs, ui, buildContainer, nil)

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
}

func TestManagerFields(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()
	buildContainer := "test-container"

	mgr := build.NewManager(runtime, fs, ui, buildContainer, nil)

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
}
