package bunker

import (
	"context"
	"testing"

	"github.com/Alejandro-M-P/AXIOM/internal/ports"
	"github.com/Alejandro-M-P/AXIOM/tests/mocks"
)

func TestNewManager(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
	if mgr.rootDir != "/root" {
		t.Errorf("expected rootDir '/root', got '%s'", mgr.rootDir)
	}
	if mgr.runtime != runtime {
		t.Error("runtime not set correctly")
	}
	if mgr.fs != fs {
		t.Error("fs not set correctly")
	}
	if mgr.ui != ui {
		t.Error("ui not set correctly")
	}
}

func TestNewManager_AllDependencies(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	mgr := NewManager("/test/root", runtime, fs, ui, mocks.NewMockSystem())

	if mgr.rootDir != "/test/root" {
		t.Errorf("rootDir mismatch: expected '/test/root', got '%s'", mgr.rootDir)
	}

	if mgr.runtime == nil {
		t.Error("runtime is nil")
	}
	if mgr.fs == nil {
		t.Error("fs is nil")
	}
	if mgr.ui == nil {
		t.Error("ui is nil")
	}
}

func TestManagerHasMutex(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	if mgr == nil {
		t.Fatal("Manager is nil")
	}
}

func TestManagerInterfaceCompliance(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	var _ interface {
		CreateBunker(ctx context.Context, name string) error
		DeleteBunker(ctx context.Context, name string, force, deleteImage bool) error
		DeleteBunkerImage(ctx context.Context) error
		ListBunkers(ctx context.Context) error
		BunkerInfo(ctx context.Context, name string) error
		StopBunker(ctx context.Context) error
		PruneBunkers(ctx context.Context) error
		BunkerStatus(ctx context.Context, name string) string
		ImageExists(ctx context.Context, image string) bool
		ListAxiomImages(ctx context.Context) ([]string, error)
		Help() error
		GetUI() ports.IPresenter
	} = NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())
}

func TestManager_GetUI(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	got := mgr.GetUI()
	if got == nil {
		t.Error("GetUI returned nil")
	}
	if got != ui {
		t.Error("GetUI did not return the expected presenter")
	}
}

func TestManagerAliases(t *testing.T) {
	runtime := mocks.NewMockRuntime()
	fs := mocks.NewMockFileSystem()
	ui := mocks.NewMockPresenter()

	mgr := NewManager("/root", runtime, fs, ui, mocks.NewMockSystem())

	_ = mgr.Create("test")
	_ = mgr.Delete("test")
	_ = mgr.List()
	_ = mgr.Stop()
	_ = mgr.Prune()
	_ = mgr.Info("test")
	_ = mgr.DeleteImage()
}
